#!/usr/bin/env bash
# run-advisor-consultation.sh
# Check advisor config, gate on consult count, write last-request.json,
# and print the advisor prompt block to stdout.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"  # skill-local: harness/scripts/
PROJECT_ROOT="$(git -C "${SCRIPT_DIR}" rev-parse --show-toplevel)"  # project-root
PLUGIN_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"  # plugin-local: harness/

CONFIG="${PLUGIN_DIR}/.claude-code-harness.config.yaml"
STATE_DIR="${PROJECT_ROOT}/.claude/state/advisor"

# ---------------------------------------------------------------------------
# Defaults
# ---------------------------------------------------------------------------
TASK_ID=""
REASON_CODE="unknown"
ERROR_SIG=""
RETRY_COUNT=0

# ---------------------------------------------------------------------------
# Argument parsing
# ---------------------------------------------------------------------------
while [[ $# -gt 0 ]]; do
  case "$1" in
    --help|-h)
      cat <<'EOF'
Usage: run-advisor-consultation.sh [OPTIONS]

Request an Opus advisor consultation for a failing task.

Options:
  --task-id     <id>      Task identifier (e.g. "62.11")
  --reason-code <code>    Reason for consultation (e.g. "type_error", "test_failure")
  --error-sig   <sig>     Short error signature / fingerprint
  --retry-count <n>       Number of retries already attempted
  --help                  Print this help and exit

Exit codes:
  0  Advisor prompt printed (or advisor disabled / max consults reached)
  1  Missing required argument
  2  Config file not found
EOF
      exit 0
      ;;
    --task-id)
      TASK_ID="${2:-}"
      shift 2
      ;;
    --reason-code)
      REASON_CODE="${2:-}"
      shift 2
      ;;
    --error-sig)
      ERROR_SIG="${2:-}"
      shift 2
      ;;
    --retry-count)
      RETRY_COUNT="${2:-0}"
      shift 2
      ;;
    *)
      echo "Unknown argument: $1" >&2
      exit 1
      ;;
  esac
done

if [ -z "$TASK_ID" ]; then
  echo "Error: --task-id is required" >&2
  exit 1
fi

# ---------------------------------------------------------------------------
# Read advisor config (grep/awk only — no yq dependency)
# ---------------------------------------------------------------------------
if [ ! -f "$CONFIG" ]; then
  echo "Config file not found: $CONFIG" >&2
  exit 2
fi

ADVISOR_ENABLED=$(awk '/^advisor:/{f=1; next} f && /^[^ ]/{f=0} f && /enabled:/{print $2}' "$CONFIG" | head -1)
MAX_CONSULTS=$(grep 'max_consults_per_task:' "$CONFIG" | awk '{print $2}')

# ---------------------------------------------------------------------------
# Exit early if advisor is disabled
# ---------------------------------------------------------------------------
if [ "${ADVISOR_ENABLED}" != "true" ]; then
  echo "advisor disabled" >&2
  exit 0
fi

# ---------------------------------------------------------------------------
# Check consult count for this task
# ---------------------------------------------------------------------------
HISTORY_FILE="${STATE_DIR}/history.jsonl"
CONSULT_COUNT=0
if [ -f "$HISTORY_FILE" ]; then
  CONSULT_COUNT=$(grep -c "\"task_id\":\"${TASK_ID}\"" "$HISTORY_FILE" 2>/dev/null || true)
fi

if [ "$CONSULT_COUNT" -ge "$MAX_CONSULTS" ]; then
  echo "max consultations reached (${CONSULT_COUNT}/${MAX_CONSULTS}) for task ${TASK_ID}" >&2
  exit 0
fi

# ---------------------------------------------------------------------------
# Ensure state directory exists
# ---------------------------------------------------------------------------
mkdir -p "$STATE_DIR"

# ---------------------------------------------------------------------------
# Write last-request.json
# ---------------------------------------------------------------------------
TIMESTAMP="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
REQUEST_FILE="${STATE_DIR}/last-request.json"

cat > "$REQUEST_FILE" <<EOF
{"task_id": "${TASK_ID}", "reason_code": "${REASON_CODE}", "error_signature": "${ERROR_SIG}", "retry_count": ${RETRY_COUNT}, "timestamp": "${TIMESTAMP}"}
EOF

# Append to history so the count check above stays accurate for subsequent calls
echo "{\"task_id\":\"${TASK_ID}\", \"reason_code\":\"${REASON_CODE}\", \"timestamp\":\"${TIMESTAMP}\"}" >> "$HISTORY_FILE"

# ---------------------------------------------------------------------------
# Print advisor prompt block to stdout
# ---------------------------------------------------------------------------
cat <<EOF

========== ADVISOR CONSULTATION REQUEST ==========
Please invoke the Opus advisor subagent with the following request:

  Task ID    : ${TASK_ID}
  Reason     : ${REASON_CODE}
  Error sig  : ${ERROR_SIG:-"(none)"}
  Retry count: ${RETRY_COUNT}
  Timestamp  : ${TIMESTAMP}

Request file: ${REQUEST_FILE}

Advisor subagent invocation:
  Agent(
    subagent_type = "advisor",
    prompt = "$(cat "${REQUEST_FILE}")"
  )

Consult $((CONSULT_COUNT + 1)) of ${MAX_CONSULTS} for this task.
==================================================

EOF
