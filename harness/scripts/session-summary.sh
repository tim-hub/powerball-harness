#!/bin/bash
# session-summary.sh
# Generate a summary at session end
#
# Usage: Called automatically from Stop hook

set +e

STATE_FILE=".claude/state/session.json"
EVENT_LOG_FILE=".claude/state/session.events.jsonl"
ARCHIVE_DIR=".claude/state/sessions"
CURRENT_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ)

# Source plans.json helpers (SSOT: .claude/harness/plans.json)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [ -f "${SCRIPT_DIR}/config-utils.sh" ]; then
  # shellcheck source=./config-utils.sh
  source "${SCRIPT_DIR}/config-utils.sh"
fi

# Skip if state file does not exist
if [ ! -f "$STATE_FILE" ]; then
  exit 0
fi

# Skip if jq is not available
if ! command -v jq &> /dev/null; then
  exit 0
fi

# Skip if already recorded to memory (prevents double-execution of Stop hook)
ALREADY_LOGGED=$(jq -r '.memory_logged // false' "$STATE_FILE" 2>/dev/null)
if [ "$ALREADY_LOGGED" = "true" ]; then
  exit 0
fi

# Retrieve session information
SESSION_START=$(jq -r '.started_at' "$STATE_FILE")
CHANGES_COUNT=$(jq '.changes_this_session | length' "$STATE_FILE")
IMPORTANT_CHANGES=$(jq '[.changes_this_session[] | select(.important == true)] | length' "$STATE_FILE")

# Git info
GIT_COMMITS=0
if [ -d ".git" ]; then
  # Approximate commit count since session start
  GIT_COMMITS=$(git log --oneline --since="$SESSION_START" 2>/dev/null | wc -l | tr -d ' ' || echo "0")
fi

# Task status from plans.json (SSOT: .claude/harness/plans.json)
COMPLETED_TASKS=0
WIP_TASK_TITLE=""
if declare -F plans_json_exists >/dev/null 2>&1 && plans_json_exists; then
  COMPLETED_TASKS=$(plans_count_status "cc:done")
  # Get current WIP task title (first match)
  WIP_TASK_TITLE=$(plans_wip_names 1 2>/dev/null | head -1 || true)
fi

# Retrieve recently edited file info from Agent Trace
AGENT_TRACE_FILE=".claude/state/agent-trace.jsonl"
RECENT_EDITS=""
RECENT_PROJECT=""
if [ -f "$AGENT_TRACE_FILE" ]; then
  # Extract edited files from last 10 trace entries
  RECENT_EDITS=$(tail -10 "$AGENT_TRACE_FILE" 2>/dev/null | jq -r '.files[].path' 2>/dev/null | sort -u | head -5 || true)
  # Get the latest project information
  RECENT_PROJECT=$(tail -1 "$AGENT_TRACE_FILE" 2>/dev/null | jq -r '.metadata.project // empty' 2>/dev/null || true)
fi

# Calculate session duration
START_EPOCH=$(date -j -f "%Y-%m-%dT%H:%M:%SZ" "$SESSION_START" "+%s" 2>/dev/null || date -d "$SESSION_START" "+%s" 2>/dev/null || echo "0")
NOW_EPOCH=$(date +%s)
DURATION_MINUTES=$(( (NOW_EPOCH - START_EPOCH) / 60 ))

# Print summary (only when there are changes)
if [ "$CHANGES_COUNT" -gt 0 ] || [ "$GIT_COMMITS" -gt 0 ] || [ -n "$RECENT_EDITS" ]; then
  echo ""
  echo "📊 Session Summary"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

  # Project name (from Agent Trace)
  if [ -n "$RECENT_PROJECT" ]; then
    echo "📁 Project: ${RECENT_PROJECT}"
  fi

  # Current task (WIP)
  if [ -n "$WIP_TASK_TITLE" ]; then
    echo "🎯 Current task: ${WIP_TASK_TITLE}"
  fi

  if [ "$COMPLETED_TASKS" -gt 0 ]; then
    echo "✅ Completed tasks: ${COMPLETED_TASKS}"
  fi

  echo "📝 Changed files: ${CHANGES_COUNT}"

  if [ "$IMPORTANT_CHANGES" -gt 0 ]; then
    echo "⚠️ Important changes: ${IMPORTANT_CHANGES}"
  fi

  if [ "$GIT_COMMITS" -gt 0 ]; then
    echo "💾 Commits: ${GIT_COMMITS}"
  fi

  if [ "$DURATION_MINUTES" -gt 0 ]; then
    echo "⏱️ Session duration: ${DURATION_MINUTES}m"
  fi

  # Recently edited files (from Agent Trace)
  if [ -n "$RECENT_EDITS" ]; then
    echo ""
    echo "📄 Recent edits:"
    echo "$RECENT_EDITS" | while read -r f; do
      [ -n "$f" ] && echo "   - $f"
    done
  fi

  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo ""
fi

# Record session end time and logged flag in state file
append_event() {
  local event_type="$1"
  local event_state="$2"
  local event_time="$3"

  # Initialize event log
  mkdir -p ".claude/state" 2>/dev/null || true
  touch "$EVENT_LOG_FILE" 2>/dev/null || true

  if command -v jq >/dev/null 2>&1; then
    local seq
    local event_id
    seq=$(jq -r '.event_seq // 0' "$STATE_FILE" 2>/dev/null)
    seq=$((seq + 1))
    event_id=$(printf "event-%06d" "$seq")

    jq --arg state "$event_state" \
       --arg updated_at "$event_time" \
       --arg event_id "$event_id" \
       --argjson event_seq "$seq" \
       '.state = $state | .updated_at = $updated_at | .last_event_id = $event_id | .event_seq = $event_seq' \
       "$STATE_FILE" > "${STATE_FILE}.tmp" && mv "${STATE_FILE}.tmp" "$STATE_FILE"

    echo "{\"id\":\"$event_id\",\"type\":\"$event_type\",\"ts\":\"$event_time\",\"state\":\"$event_state\"}" >> "$EVENT_LOG_FILE"
  fi
}

append_event "session.stop" "stopped" "$CURRENT_TIME"

if command -v jq >/dev/null 2>&1; then
  jq --arg ended_at "$CURRENT_TIME" \
     --arg duration "$DURATION_MINUTES" \
     '. + {ended_at: $ended_at, duration_minutes: ($duration | tonumber), memory_logged: true}' \
     "$STATE_FILE" > "${STATE_FILE}.tmp" && mv "${STATE_FILE}.tmp" "$STATE_FILE"
fi

# Save to archive (for resume/fork)
if [ -f "$STATE_FILE" ]; then
  mkdir -p "$ARCHIVE_DIR" 2>/dev/null || true
  if command -v jq >/dev/null 2>&1; then
    ARCHIVE_ID=$(jq -r '.session_id // empty' "$STATE_FILE" 2>/dev/null)
    if [ -n "$ARCHIVE_ID" ]; then
      cp "$STATE_FILE" "$ARCHIVE_DIR/${ARCHIVE_ID}.json" 2>/dev/null || true
      if [ -f "$EVENT_LOG_FILE" ]; then
        cp "$EVENT_LOG_FILE" "$ARCHIVE_DIR/${ARCHIVE_ID}.events.jsonl" 2>/dev/null || true
      fi
    fi
  fi
fi

exit 0
