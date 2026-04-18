#!/usr/bin/env bash
# run-advisor-consultation.sh
# Codex advisor consultation wrapper.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
CONFIG_UTILS="${PROJECT_ROOT}/scripts/config-utils.sh"
COMPANION_BIN="${CODEX_ADVISOR_COMPANION:-${PROJECT_ROOT}/scripts/codex-companion.sh}"
SCHEMA_FILE="${PROJECT_ROOT}/scripts/lib/advisor-response.schema.json"

usage() {
  cat <<'EOF'
Usage:
  scripts/run-advisor-consultation.sh --request-file <path> [--response-file <path>] [--model <model>] [--timeout-sec <n>]
EOF
}

timestamp_utc() {
  date -u +"%Y-%m-%dT%H:%M:%SZ"
}

append_history_line() {
  local history_file="$1"
  local payload="$2"
  printf '%s\n' "${payload}" >> "${history_file}"
}

REQUEST_FILE=""
RESPONSE_FILE=""
MODEL=""
TIMEOUT_SEC="180"

while [ $# -gt 0 ]; do
  case "$1" in
    --request-file)
      REQUEST_FILE="${2:-}"
      shift 2
      ;;
    --response-file)
      RESPONSE_FILE="${2:-}"
      shift 2
      ;;
    --model)
      MODEL="${2:-}"
      shift 2
      ;;
    --timeout-sec)
      TIMEOUT_SEC="${2:-}"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

[ -n "${REQUEST_FILE}" ] || {
  echo "--request-file is required" >&2
  exit 2
}
[ -f "${REQUEST_FILE}" ] || {
  echo "request file not found: ${REQUEST_FILE}" >&2
  exit 1
}
[ -f "${SCHEMA_FILE}" ] || {
  echo "advisor response schema not found: ${SCHEMA_FILE}" >&2
  exit 1
}

# shellcheck disable=SC1090
PROJECT_ROOT="${PROJECT_ROOT}" CONFIG_FILE="${PROJECT_ROOT}/.claude-code-harness.config.yaml" source "${CONFIG_UTILS}"
ensure_advisor_state_files

if [ -z "${MODEL}" ]; then
  MODEL="$(get_advisor_codex_model)"
fi

if [ -z "${RESPONSE_FILE}" ]; then
  RESPONSE_FILE="${PROJECT_ROOT}/.claude/state/advisor/last-response.json"
fi

HISTORY_FILE="$(get_advisor_history_file)"
LAST_REQUEST_FILE="$(get_advisor_last_request_file)"
LAST_RESPONSE_FILE="$(get_advisor_last_response_file)"

REQUEST_JSON="$(cat "${REQUEST_FILE}")"
printf '%s\n' "${REQUEST_JSON}" > "${LAST_REQUEST_FILE}"

PROMPT_FILE="$(mktemp)"
RAW_OUTPUT_FILE="$(mktemp)"
STDERR_FILE="$(mktemp)"
trap 'rm -f "${PROMPT_FILE}" "${RAW_OUTPUT_FILE}" "${STDERR_FILE}"' EXIT

cat > "${PROMPT_FILE}" <<EOF
You are the Harness advisor. You are not the executor.

Rules:
- Return JSON only.
- Follow advisor-response.v1 exactly.
- Do not use tools.
- Do not write code.
- Do not address the end user.
- If the executor should continue with a new plan, use decision PLAN.
- If the executor should keep the plan but correct a local mistake, use decision CORRECTION.
- If the executor must stop and escalate, use decision STOP and provide stop_reason.

Request JSON:
${REQUEST_JSON}
EOF

set +e
python3 - "${PROMPT_FILE}" "${RAW_OUTPUT_FILE}" "${STDERR_FILE}" "${TIMEOUT_SEC}" "${COMPANION_BIN}" "${MODEL}" "${SCHEMA_FILE}" <<'PY'
import pathlib
import subprocess
import sys

prompt_file = pathlib.Path(sys.argv[1])
stdout_file = pathlib.Path(sys.argv[2])
stderr_file = pathlib.Path(sys.argv[3])
timeout_sec = int(sys.argv[4])
companion = sys.argv[5]
model = sys.argv[6]
schema = sys.argv[7]

prompt = prompt_file.read_text(encoding="utf-8")

def _to_text(value):
    if value is None:
        return ""
    if isinstance(value, bytes):
        return value.decode("utf-8", errors="replace")
    return value

try:
    completed = subprocess.run(
        ["bash", companion, "task", "--model", model, "--output-schema", schema],
        input=prompt,
        text=True,
        capture_output=True,
        timeout=timeout_sec,
        check=False,
    )
except subprocess.TimeoutExpired as exc:
    # TimeoutExpired.stdout/stderr can be bytes even when text=True, because
    # the exception carries the raw partial buffers captured before timeout.
    stdout_file.write_text(_to_text(exc.stdout), encoding="utf-8")
    stderr_file.write_text(_to_text(exc.stderr) + "\nTIMEOUT\n", encoding="utf-8")
    raise SystemExit(124)

stdout_file.write_text(_to_text(completed.stdout), encoding="utf-8")
stderr_file.write_text(_to_text(completed.stderr), encoding="utf-8")
raise SystemExit(completed.returncode)
PY
COMMAND_EXIT=$?
set -e

if [ "${COMMAND_EXIT}" -eq 124 ]; then
  append_history_line "${HISTORY_FILE}" "$(python3 - "${REQUEST_FILE}" "${MODEL}" "${TIMEOUT_SEC}" <<'PY'
import json
import pathlib
import sys
from datetime import datetime, timezone

request = json.loads(pathlib.Path(sys.argv[1]).read_text(encoding="utf-8"))
print(json.dumps({
    "timestamp": datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ"),
    "status": "timeout",
    "request": request,
    "model": sys.argv[2],
    "timeout_sec": int(sys.argv[3])
}, ensure_ascii=False))
PY
)"
  echo "advisor consultation timed out after ${TIMEOUT_SEC}s" >&2
  exit 124
fi

if [ "${COMMAND_EXIT}" -ne 0 ]; then
  append_history_line "${HISTORY_FILE}" "$(python3 - "${REQUEST_FILE}" "${MODEL}" "${STDERR_FILE}" <<'PY'
import json
import pathlib
import sys
from datetime import datetime, timezone

request = json.loads(pathlib.Path(sys.argv[1]).read_text(encoding="utf-8"))
stderr_text = pathlib.Path(sys.argv[3]).read_text(encoding="utf-8")
print(json.dumps({
    "timestamp": datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ"),
    "status": "error",
    "request": request,
    "model": sys.argv[2],
    "stderr": stderr_text.strip()
}, ensure_ascii=False))
PY
)"
  cat "${STDERR_FILE}" >&2
  exit "${COMMAND_EXIT}"
fi

python3 - "${RAW_OUTPUT_FILE}" "${REQUEST_FILE}" "${RESPONSE_FILE}" "${LAST_RESPONSE_FILE}" "${HISTORY_FILE}" "${MODEL}" <<'PY'
import json
import pathlib
import sys
from datetime import datetime, timezone

raw_output_path = pathlib.Path(sys.argv[1])
request_path = pathlib.Path(sys.argv[2])
response_path = pathlib.Path(sys.argv[3])
last_response_path = pathlib.Path(sys.argv[4])
history_path = pathlib.Path(sys.argv[5])
model = sys.argv[6]

raw_text = raw_output_path.read_text(encoding="utf-8").strip()
if not raw_text:
    raise SystemExit("advisor returned empty output")

try:
    payload = json.loads(raw_text)
except json.JSONDecodeError as exc:
    raise SystemExit(f"advisor returned invalid JSON: {exc}")

required = {"schema_version", "decision", "summary", "executor_instructions", "confidence", "stop_reason"}
missing = sorted(required - payload.keys())
if missing:
    raise SystemExit(f"advisor response missing keys: {', '.join(missing)}")

if payload["schema_version"] != "advisor-response.v1":
    raise SystemExit("advisor response schema_version mismatch")
if payload["decision"] not in {"PLAN", "CORRECTION", "STOP"}:
    raise SystemExit("advisor response decision must be PLAN, CORRECTION, or STOP")
if not isinstance(payload["executor_instructions"], list):
    raise SystemExit("advisor response executor_instructions must be an array")
if payload["decision"] == "STOP" and not payload["stop_reason"]:
    raise SystemExit("STOP decision requires stop_reason")

rendered = json.dumps(payload, ensure_ascii=False, indent=2) + "\n"
response_path.write_text(rendered, encoding="utf-8")
last_response_path.write_text(rendered, encoding="utf-8")

request = json.loads(request_path.read_text(encoding="utf-8"))
history_entry = {
    "timestamp": datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ"),
    "status": "ok",
    "model": model,
    "request": request,
    "response": payload
}
with history_path.open("a", encoding="utf-8") as fh:
    fh.write(json.dumps(history_entry, ensure_ascii=False) + "\n")
print(rendered, end="")
PY
