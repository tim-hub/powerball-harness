#!/bin/bash
# todo-sync.sh
# Bidirectional sync between TodoWrite and Plans.md
#
# Called from PostToolUse hook, reflects TodoWrite state changes in Plans.md
#
# Mapping:
#   TodoWrite status  → Plans.md marker
#   pending          → cc:TODO
#   in_progress      → cc:WIP
#   completed        → cc:done

set +e  # Do not stop on error

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Read JSON input from stdin
INPUT=""
if [ ! -t 0 ]; then
  INPUT=$(cat 2>/dev/null || true)
fi

# jq is required
if ! command -v jq >/dev/null 2>&1; then
  exit 0
fi

# Exit if no input
if [ -z "$INPUT" ]; then
  exit 0
fi

# Parse TodoWrite tool output
TOOL_NAME=$(echo "$INPUT" | jq -r '.tool_name // empty' 2>/dev/null)

# Ignore if not TodoWrite
if [ "$TOOL_NAME" != "TodoWrite" ]; then
  exit 0
fi

# Get Plans.md path
if [ -f "${SCRIPT_DIR}/config-utils.sh" ]; then
  source "${SCRIPT_DIR}/config-utils.sh"
  PLANS_FILE=$(get_plans_file_path)
else
  PLANS_FILE="Plans.md"
fi

# Exit if Plans.md does not exist
if [ ! -f "$PLANS_FILE" ]; then
  exit 0
fi

# State directory
STATE_DIR=".claude/state"
mkdir -p "$STATE_DIR"
SYNC_STATE_FILE="${STATE_DIR}/todo-sync-state.json"

# Get TodoWrite todos array
TODOS=$(echo "$INPUT" | jq -r '.tool_input.todos // []' 2>/dev/null)

if [ -z "$TODOS" ] || [ "$TODOS" = "null" ] || [ "$TODOS" = "[]" ]; then
  exit 0
fi

# Save sync state
echo "$TODOS" | jq '{
  synced_at: (now | todate),
  todos: .
}' > "$SYNC_STATE_FILE" 2>/dev/null

# Update task state in Plans.md
# Note: Maintaining Plans.md format while updating is complex,
# so we only log here and leave the actual update to Claude Code

# Record in event log
EVENT_LOG="${STATE_DIR}/session.events.jsonl"
NOW=$(date -u +%Y-%m-%dT%H:%M:%SZ)

PENDING_COUNT=$(echo "$TODOS" | jq '[.[] | select(.status == "pending")] | length' 2>/dev/null || echo "0")
WIP_COUNT=$(echo "$TODOS" | jq '[.[] | select(.status == "in_progress")] | length' 2>/dev/null || echo "0")
DONE_COUNT=$(echo "$TODOS" | jq '[.[] | select(.status == "completed")] | length' 2>/dev/null || echo "0")

if [ -f "$EVENT_LOG" ]; then
  echo "{\"type\":\"todo.sync\",\"ts\":\"$NOW\",\"data\":{\"pending\":$PENDING_COUNT,\"in_progress\":$WIP_COUNT,\"completed\":$DONE_COUNT}}" >> "$EVENT_LOG"
fi

# ===== All-complete detection and warning in Work mode =====
WORK_WARNING=""
WORK_FILE="${STATE_DIR}/work-active.json"
# Backward compatibility: try ultrawork-active.json if work-active.json doesn't exist
if [ ! -f "$WORK_FILE" ]; then
  WORK_FILE="${STATE_DIR}/ultrawork-active.json"
fi
TOTAL_COUNT=$((PENDING_COUNT + WIP_COUNT + DONE_COUNT))

# All tasks complete (pending=0, WIP=0, completed>0) and in Work mode
if [ "$PENDING_COUNT" -eq 0 ] && [ "$WIP_COUNT" -eq 0 ] && [ "$DONE_COUNT" -gt 0 ]; then
  if [ -f "$WORK_FILE" ]; then
    REVIEW_STATUS=$(jq -r '.review_status // "pending"' "$WORK_FILE" 2>/dev/null)

    if [ "$REVIEW_STATUS" != "passed" ]; then
      WORK_WARNING="\n\n⚠️ **Pre-completion check for work**: review_status=${REVIEW_STATUS}\n→ Please obtain APPROVE via /harness-review before completing"
    fi
  fi
fi

# Output sync info as additionalContext
OUTPUT="[TodoSync] Synced with Plans.md: TODO=$PENDING_COUNT, WIP=$WIP_COUNT, done=$DONE_COUNT${WORK_WARNING}"

if command -v jq >/dev/null 2>&1; then
  jq -nc --arg ctx "$OUTPUT" \
    '{hookSpecificOutput:{additionalContext:$ctx}}'
else
  cat <<EOF
{"hookSpecificOutput":{"additionalContext":"[TodoSync] Synced with Plans.md: TODO=$PENDING_COUNT, WIP=$WIP_COUNT, done=$DONE_COUNT"}}
EOF
fi
