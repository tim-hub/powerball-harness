#!/bin/bash
# test-session-control.sh
# Minimal tests for session-control resume/fork

set -euo pipefail
export TMPDIR=/tmp  # Force /tmp for sandboxed execution (sandbox blocks /var/folders)

if ! command -v jq >/dev/null 2>&1 && ! command -v python3 >/dev/null 2>&1; then
  echo "SKIP: jq or python3 required"
  exit 0
fi

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SCRIPT="$ROOT_DIR/scripts/session-control.sh"

TMP_DIR="$(mktemp -d "/tmp/harness-test.XXXXXX")"
cleanup() { rm -rf "$TMP_DIR"; }
trap cleanup EXIT

mkdir -p "$TMP_DIR/.claude/state/sessions"

BASE_SESSION_ID="session-123"
BASE_SESSION_FILE="$TMP_DIR/.claude/state/sessions/${BASE_SESSION_ID}.json"
BASE_EVENTS_FILE="$TMP_DIR/.claude/state/sessions/${BASE_SESSION_ID}.events.jsonl"

cat > "$BASE_SESSION_FILE" <<EOF
{
  "session_id": "$BASE_SESSION_ID",
  "state": "stopped",
  "event_seq": 1,
  "last_event_id": "event-000001"
}
EOF

echo "{\"id\":\"event-000001\",\"type\":\"session.stop\",\"ts\":\"2026-01-01T00:00:00Z\",\"state\":\"stopped\"}" > "$BASE_EVENTS_FILE"

pushd "$TMP_DIR" >/dev/null

# Resume by id
"$SCRIPT" --resume "$BASE_SESSION_ID"

if command -v jq >/dev/null 2>&1; then
  resumed_id=$(jq -r '.session_id' .claude/state/session.json)
  if [ "$resumed_id" != "$BASE_SESSION_ID" ]; then
    echo "FAIL: resume session_id mismatch"
    exit 1
  fi
else
  resumed_id=$(python3 - <<'PY'
import json
print(json.load(open(".claude/state/session.json")).get("session_id"))
PY
)
  if [ "$resumed_id" != "$BASE_SESSION_ID" ]; then
    echo "FAIL: resume session_id mismatch"
    exit 1
  fi
fi

# Fork current
"$SCRIPT" --fork current --reason "test fork"

if command -v jq >/dev/null 2>&1; then
  fork_parent=$(jq -r '.parent_session_id' .claude/state/session.json)
  if [ -z "$fork_parent" ]; then
    echo "FAIL: fork parent_session_id empty"
    exit 1
  fi
else
  fork_parent=$(python3 - <<'PY'
import json
print(json.load(open(".claude/state/session.json")).get("parent_session_id"))
PY
)
  if [ -z "$fork_parent" ]; then
    echo "FAIL: fork parent_session_id empty"
    exit 1
  fi
fi

popd >/dev/null

echo "OK"
