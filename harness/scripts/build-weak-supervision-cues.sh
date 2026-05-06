#!/usr/bin/env bash
# Build compact advisor context from .claude/state/elicitation/events.jsonl.
# Reads the local elicitation ledger and emits weak-supervision cues for the
# advisor request file provided via --request-file.

set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  bash harness/scripts/build-weak-supervision-cues.sh --request-file <advisor-request.json> [--project-root <dir>]
EOF
}

REQUEST_FILE=""
# project-root: user's git repository root
PROJECT_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"

while [ $# -gt 0 ]; do
  case "$1" in
    --request-file)
      REQUEST_FILE="${2:-}"
      shift 2
      ;;
    --project-root)
      PROJECT_ROOT="${2:-}"
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

LEDGER_FILE="${PROJECT_ROOT}/.claude/state/elicitation/events.jsonl"

python3 - "${REQUEST_FILE}" "${LEDGER_FILE}" <<'PY'
import json
import pathlib
import sys

request_path = pathlib.Path(sys.argv[1])
ledger_path = pathlib.Path(sys.argv[2])
request = json.loads(request_path.read_text(encoding="utf-8"))
reason = request.get("reason_code", "")
task_id = request.get("task_id", "")

eligible_reasons = {
    "retry-threshold",
    "pivot-required",
    "needs-spike",
    "security-sensitive",
    "state-migration",
    "advisor-required",
}
if reason not in eligible_reasons:
    raise SystemExit(0)

if not ledger_path.exists():
    raise SystemExit(0)

interesting = {"weak_label", "judge_verdict", "eval_result", "counterexample"}
hits = []
for line in ledger_path.read_text(encoding="utf-8").splitlines():
    if not line.strip():
        continue
    try:
        event = json.loads(line)
    except json.JSONDecodeError:
        continue
    if event.get("schema_version") != "elicitation-event.v1":
        continue
    if event.get("event_kind") not in interesting:
        continue
    event_task = event.get("task_id", "")
    if event_task and task_id and event_task != task_id:
        continue
    hits.append(event)

if not hits:
    raise SystemExit(0)

print("Weak-supervision cues from local elicitation ledger:")
for event in hits[-5:]:
    bits = [
        f"kind={event.get('event_kind', 'unknown')}",
        f"run={event.get('run_id', 'unknown')}",
    ]
    if event.get("task_id"):
        bits.append(f"task={event['task_id']}")
    if event.get("rubric_id"):
        bits.append(f"rubric={event['rubric_id']}")
    if event.get("verdict"):
        bits.append(f"verdict={event['verdict']}")
    if event.get("reward_score") is not None:
        bits.append(f"reward={event['reward_score']}")
    if event.get("evidence_refs"):
        bits.append("evidence=" + ",".join(event["evidence_refs"][:3]))
    if event.get("message"):
        bits.append("note=" + event["message"][:140])
    print("- " + "; ".join(bits))
PY
