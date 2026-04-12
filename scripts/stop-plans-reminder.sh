#!/bin/bash
# stop-plans-reminder.sh
# Stop Hook: Plans.md marker update reminder
#
# Claude Code 2.1.1 compat: implemented with command type instead of prompt type
# Output: JSON format {"decision": "approve", "reason": "...", "systemMessage": "..."}

set -euo pipefail

# Decision variables
NEED_REMINDER="false"
REASON=""
MESSAGE=""

# Check for changes
HAS_CHANGES="false"

# Git uncommitted changes
if [ -d ".git" ]; then
  GIT_UNCOMMITTED=$(git status --porcelain 2>/dev/null | wc -l | tr -d ' ' || echo "0")
  if [ "$GIT_UNCOMMITTED" -gt 0 ]; then
    HAS_CHANGES="true"
  fi
fi

# Changes during session
if [ -f ".claude/state/session.json" ] && command -v jq >/dev/null 2>&1; then
  SESSION_CHANGES=$(jq '.changes_this_session // 0' .claude/state/session.json 2>/dev/null || echo "0")
  if [ "$SESSION_CHANGES" != "0" ] && [ "$SESSION_CHANGES" != "null" ]; then
    HAS_CHANGES="true"
  fi
fi

# Check Plans.md only when there are changes
if [ "$HAS_CHANGES" = "true" ] && [ -f "Plans.md" ]; then
  PM_PENDING=$(( $(grep -c "pm:requested" Plans.md 2>/dev/null || echo "0") + $(grep -c "cursor:requested" Plans.md 2>/dev/null || echo "0") ))
  CC_WIP=$(grep -c "cc:WIP" Plans.md 2>/dev/null || echo "0")
  CC_DONE=$(grep -c "cc:done" Plans.md 2>/dev/null || echo "0")

  # When there are PM requests
  if [ "$PM_PENDING" -gt 0 ]; then
    NEED_REMINDER="true"
    REASON="pm_pending_tasks > 0"
    MESSAGE="Plans.md: ${PM_PENDING} pm:requested item(s) found. Update to cc:WIP when starting work, and cc:done when done"
  fi

  # When there are WIP tasks
  if [ "$CC_WIP" -gt 0 ]; then
    NEED_REMINDER="true"
    REASON="cc_wip_tasks > 0"
    MESSAGE="Plans.md: ${CC_WIP} cc:WIP item(s) found. Update to cc:done when completed"
  fi

  # When there are Done tasks (awaiting PM confirmation)
  if [ "$CC_DONE" -gt 0 ]; then
    NEED_REMINDER="true"
    REASON="cc_done_tasks > 0"
    MESSAGE="Plans.md: ${CC_DONE} cc:done item(s) found. Update to pm:confirmed after PM confirmation"
  fi
fi

# JSON output
if [ "$NEED_REMINDER" = "true" ]; then
  cat << EOF
{"decision": "approve", "reason": "$REASON", "systemMessage": "$MESSAGE"}
EOF
else
  cat << EOF
{"decision": "approve", "reason": "No reminder needed", "systemMessage": ""}
EOF
fi
