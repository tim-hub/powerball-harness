#!/bin/bash
# session-summary.sh
# Generate summary at session end
#
# Usage: Auto-executed from Stop hook

set +e

STATE_FILE=".claude/state/session.json"
MEMORY_DIR=".claude/memory"
SESSION_LOG_FILE="${MEMORY_DIR}/session-log.md"
EVENT_LOG_FILE=".claude/state/session.events.jsonl"
ARCHIVE_DIR=".claude/state/sessions"
CURRENT_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ)

# Skip if no state file
if [ ! -f "$STATE_FILE" ]; then
  exit 0
fi

# Skip if jq not available
if ! command -v jq &> /dev/null; then
  exit 0
fi

# Skip if already logged to memory (prevent double Stop hook execution)
ALREADY_LOGGED=$(jq -r '.memory_logged // false' "$STATE_FILE" 2>/dev/null)
if [ "$ALREADY_LOGGED" = "true" ]; then
  exit 0
fi

# Get session information
SESSION_ID=$(jq -r '.session_id // "unknown"' "$STATE_FILE")
SESSION_START=$(jq -r '.started_at' "$STATE_FILE")
PROJECT_NAME=$(jq -r '.project_name // empty' "$STATE_FILE")
GIT_BRANCH=$(jq -r '.git.branch // empty' "$STATE_FILE")
CHANGES_COUNT=$(jq '.changes_this_session | length' "$STATE_FILE")
IMPORTANT_CHANGES=$(jq '[.changes_this_session[] | select(.important == true)] | length' "$STATE_FILE")

# Git info
GIT_COMMITS=0
if [ -d ".git" ]; then
  # Commit count since session start (approximate)
  GIT_COMMITS=$(git log --oneline --since="$SESSION_START" 2>/dev/null | wc -l | tr -d ' ' || echo "0")
fi

# Plans.md task status
COMPLETED_TASKS=0
WIP_TASK_TITLE=""
if [ -f "Plans.md" ]; then
  COMPLETED_TASKS=$(grep -c "cc:done" Plans.md 2>/dev/null || echo "0")
  # Get current WIP task title (first one)
  WIP_TASK_TITLE=$(grep -E "^\s*-\s*\[.\]\s*\*\*.*\`cc:WIP\`" Plans.md 2>/dev/null | head -1 | sed 's/.*\*\*\(.*\)\*\*.*/\1/' || true)
fi

# Get recent edit file info from Agent Trace
AGENT_TRACE_FILE=".claude/state/agent-trace.jsonl"
RECENT_EDITS=""
RECENT_PROJECT=""
if [ -f "$AGENT_TRACE_FILE" ]; then
  # Extract edit files from last 10 traces
  RECENT_EDITS=$(tail -10 "$AGENT_TRACE_FILE" 2>/dev/null | jq -r '.files[].path' 2>/dev/null | sort -u | head -5 || true)
  # Get latest project info
  RECENT_PROJECT=$(tail -1 "$AGENT_TRACE_FILE" 2>/dev/null | jq -r '.metadata.project // empty' 2>/dev/null || true)
fi

# Calculate session duration
START_EPOCH=$(date -j -f "%Y-%m-%dT%H:%M:%SZ" "$SESSION_START" "+%s" 2>/dev/null || date -d "$SESSION_START" "+%s" 2>/dev/null || echo "0")
NOW_EPOCH=$(date +%s)
DURATION_MINUTES=$(( (NOW_EPOCH - START_EPOCH) / 60 ))

# Summary output (only when changes exist)
if [ "$CHANGES_COUNT" -gt 0 ] || [ "$GIT_COMMITS" -gt 0 ] || [ -n "$RECENT_EDITS" ]; then
  echo ""
  echo "Session Summary"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

  # Project name (from Agent Trace)
  if [ -n "$RECENT_PROJECT" ]; then
    echo "Project: ${RECENT_PROJECT}"
  fi

  # Current task (WIP)
  if [ -n "$WIP_TASK_TITLE" ]; then
    echo "Current task: ${WIP_TASK_TITLE}"
  fi

  if [ "$COMPLETED_TASKS" -gt 0 ]; then
    echo "Completed tasks: ${COMPLETED_TASKS}"
  fi

  echo "Changed files: ${CHANGES_COUNT}"

  if [ "$IMPORTANT_CHANGES" -gt 0 ]; then
    echo "Important changes: ${IMPORTANT_CHANGES}"
  fi

  if [ "$GIT_COMMITS" -gt 0 ]; then
    echo "Commits: ${GIT_COMMITS}"
  fi

  if [ "$DURATION_MINUTES" -gt 0 ]; then
    echo "Session duration: ${DURATION_MINUTES} min"
  fi

  # Recent edits (from Agent Trace)
  if [ -n "$RECENT_EDITS" ]; then
    echo ""
    echo "Recent edits:"
    echo "$RECENT_EDITS" | while read -r f; do
      [ -n "$f" ] && echo "   - $f"
    done
  fi

  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo ""
fi

# ================================
# Auto-append to `.claude/memory/session-log.md` (create if needed)
# ================================

# Even without changes, we may want a record that the session started,
# so write log if session start is available (empty sessions are OK)
if [ -n "$SESSION_START" ] && [ "$SESSION_START" != "null" ]; then
  mkdir -p "$MEMORY_DIR" 2>/dev/null || true

  if [ ! -f "$SESSION_LOG_FILE" ]; then
    cat > "$SESSION_LOG_FILE" << 'EOF'
# Session Log

Per-session work log (primarily for local use).
Promote important decisions to `.claude/memory/decisions.md` and reusable solutions to `.claude/memory/patterns.md`.

## Index

- (Add as needed)

---
EOF
  fi

  # Changed files list (deduplicated)
  CHANGED_FILES=$(jq -r '.changes_this_session[]?.file' "$STATE_FILE" 2>/dev/null | awk 'NF' | awk '!seen[$0]++')
  IMPORTANT_FILES=$(jq -r '.changes_this_session[]? | select(.important == true) | .file' "$STATE_FILE" 2>/dev/null | awk 'NF' | awk '!seen[$0]++')

  # WIP tasks (extract briefly if present)
  WIP_TASKS=""
  if [ -f "Plans.md" ]; then
    WIP_TASKS=$(grep -n "cc:WIP\|pm:requested\|cursor:requested" Plans.md 2>/dev/null | head -20 || true)
  fi

  {
    echo ""
    echo "## Session: ${CURRENT_TIME}"
    echo ""
    echo "- session_id: \`${SESSION_ID}\`"
    [ -n "$PROJECT_NAME" ] && echo "- project: \`${PROJECT_NAME}\`"
    [ -n "$GIT_BRANCH" ] && echo "- branch: \`${GIT_BRANCH}\`"
    echo "- started_at: \`${SESSION_START}\`"
    echo "- ended_at: \`${CURRENT_TIME}\`"
    [ "$DURATION_MINUTES" -gt 0 ] && echo "- duration_minutes: ${DURATION_MINUTES}"
    echo "- changes: ${CHANGES_COUNT}"
    [ "$IMPORTANT_CHANGES" -gt 0 ] && echo "- important_changes: ${IMPORTANT_CHANGES}"
    [ "$GIT_COMMITS" -gt 0 ] && echo "- commits: ${GIT_COMMITS}"
    echo ""
    echo "### Changed Files"
    if [ -n "$CHANGED_FILES" ]; then
      echo "$CHANGED_FILES" | while read -r f; do
        [ -n "$f" ] && echo "- \`$f\`"
      done
    else
      echo "- (none)"
    fi
    echo ""
    echo "### Important Changes (important=true)"
    if [ -n "$IMPORTANT_FILES" ]; then
      echo "$IMPORTANT_FILES" | while read -r f; do
        [ -n "$f" ] && echo "- \`$f\`"
      done
    else
      echo "- (none)"
    fi
    echo ""
    echo "### Handoff to Next Session (optional)"
    if [ -n "$WIP_TASKS" ]; then
      echo ""
      echo "**Plans.md WIP/pending (excerpt)**:"
      echo ""
      echo '```'
      echo "$WIP_TASKS"
      echo '```'
    else
      echo "- (Add as needed)"
    fi
    echo ""
    echo "---"
  } >> "$SESSION_LOG_FILE" 2>/dev/null || true
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

# Archive for resume/fork
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
