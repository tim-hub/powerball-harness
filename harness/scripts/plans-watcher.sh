#!/bin/bash
# plans-watcher.sh - Monitor changes to Plans.md and generate notifications to PM (compatible: cursor:*)
# Called from PostToolUse hook

set +e  # Do not stop on error

# Get the changed file (stdin JSON preferred / compatible: $1,$2)
INPUT=""
if [ ! -t 0 ]; then
  INPUT="$(cat 2>/dev/null)"
fi

CHANGED_FILE="${1:-}"
TOOL_NAME="${2:-}"
CWD=""

if [ -n "$INPUT" ]; then
  if command -v jq >/dev/null 2>&1; then
    TOOL_NAME_FROM_STDIN="$(printf '%s' "$INPUT" | jq -r '.tool_name // empty' 2>/dev/null)"
    FILE_PATH_FROM_STDIN="$(printf '%s' "$INPUT" | jq -r '.tool_input.file_path // .tool_response.filePath // empty' 2>/dev/null)"
    CWD_FROM_STDIN="$(printf '%s' "$INPUT" | jq -r '.cwd // empty' 2>/dev/null)"
  elif command -v python3 >/dev/null 2>&1; then
    eval "$(printf '%s' "$INPUT" | python3 -c '
import json, shlex, sys
try:
    data = json.load(sys.stdin)
except Exception:
    data = {}
tool_name = data.get("tool_name") or ""
cwd = data.get("cwd") or ""
tool_input = data.get("tool_input") or {}
tool_response = data.get("tool_response") or {}
file_path = tool_input.get("file_path") or tool_response.get("filePath") or ""
print(f"TOOL_NAME_FROM_STDIN={shlex.quote(tool_name)}")
print(f"CWD_FROM_STDIN={shlex.quote(cwd)}")
print(f"FILE_PATH_FROM_STDIN={shlex.quote(file_path)}")
' 2>/dev/null)"
  fi

  [ -z "$CHANGED_FILE" ] && CHANGED_FILE="${FILE_PATH_FROM_STDIN:-}"
  [ -z "$TOOL_NAME" ] && TOOL_NAME="${TOOL_NAME_FROM_STDIN:-}"
  CWD="${CWD_FROM_STDIN:-}"
fi

# Normalize to project-relative path if possible
if [ -n "$CWD" ] && [ -n "$CHANGED_FILE" ] && [[ "$CHANGED_FILE" == "$CWD/"* ]]; then
  CHANGED_FILE="${CHANGED_FILE#$CWD/}"
fi

# Plans.md path (respecting plansDirectory setting)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [ -f "${SCRIPT_DIR}/config-utils.sh" ]; then
  source "${SCRIPT_DIR}/config-utils.sh"
  PLANS_FILE=$(get_plans_file_path)
  plans_file_exists || PLANS_FILE=""
else
  # Fallback: legacy search logic
  find_plans_file() {
      for f in Plans.md plans.md PLANS.md PLANS.MD; do
          if [ -f "$f" ]; then
              echo "$f"
              return 0
          fi
      done
      return 1
  }
  PLANS_FILE=$(find_plans_file)
fi

# Skip changes to files other than Plans.md
if [ -z "$PLANS_FILE" ]; then
    exit 0
fi

case "$CHANGED_FILE" in
    "$PLANS_FILE"|*/"$PLANS_FILE") ;;
    *) exit 0 ;;
esac

# State directory
STATE_DIR=".claude/state"
mkdir -p "$STATE_DIR"

# Get previous state
PREV_STATE_FILE="${STATE_DIR}/plans-state.json"

# Count markers
count_markers() {
    local marker=$1
    local count=0
    if [ -f "$PLANS_FILE" ]; then
        count=$(grep -c "$marker" "$PLANS_FILE" 2>/dev/null || true)
        [ -z "$count" ] && count=0
    fi
    echo "$count"
}

# Get current state (pm:* is canonical. cursor:* is treated as alias for compatibility)
PM_PENDING=$(( $(count_markers "pm:pending") + $(count_markers "cursor:pending") ))
CC_TODO=$(count_markers "cc:TODO")
CC_WIP=$(count_markers "cc:WIP")
CC_DONE=$(count_markers "cc:done")
PM_CONFIRMED=$(( $(count_markers "pm:confirmed") + $(count_markers "cursor:confirmed") ))

# Detect new tasks
NEW_TASKS=""
if [ -f "$PREV_STATE_FILE" ]; then
    PREV_PM_PENDING=$(jq -r '.pm_pending // 0' "$PREV_STATE_FILE" 2>/dev/null || echo "0")
    if [ "$PM_PENDING" -gt "$PREV_PM_PENDING" ] 2>/dev/null; then
        NEW_TASKS="pm:pending"
    fi
fi

# Detect completed tasks
COMPLETED_TASKS=""
if [ -f "$PREV_STATE_FILE" ]; then
    PREV_CC_DONE=$(jq -r '.cc_done // 0' "$PREV_STATE_FILE" 2>/dev/null || echo "0")
    if [ "$CC_DONE" -gt "$PREV_CC_DONE" ] 2>/dev/null; then
        COMPLETED_TASKS="cc:done"
    fi
fi

# Save state
cat > "$PREV_STATE_FILE" << EOF
{
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "pm_pending": $PM_PENDING,
  "cc_todo": $CC_TODO,
  "cc_wip": $CC_WIP,
  "cc_done": $CC_DONE,
  "pm_confirmed": $PM_CONFIRMED
}
EOF

# Generate notification
generate_notification() {
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "📋 Plans.md change detected"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

    if [ -n "$NEW_TASKS" ]; then
        echo "🆕 New tasks: PM has submitted a request"
        echo "   → Run /sync-status to review the status, then /work to start"
    fi

    if [ -n "$COMPLETED_TASKS" ]; then
        echo "✅ Tasks completed: ready to report to PM"
        echo "   → Report via /handoff-to-pm-claude (or /handoff-to-cursor)"
    fi

    echo ""
    echo "📊 Current status:"
    echo "   pm:pending     : $PM_PENDING item(s) (compat: cursor:pending)"
    echo "   cc:TODO        : $CC_TODO item(s)"
    echo "   cc:WIP         : $CC_WIP item(s)"
    echo "   cc:done        : $CC_DONE item(s)"
    echo "   pm:confirmed   : $PM_CONFIRMED item(s) (compat: cursor:confirmed)"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
}

# Only notify if there are changes
if [ -n "$NEW_TASKS" ] || [ -n "$COMPLETED_TASKS" ]; then
    generate_notification
fi

# Generate PM notification file (for two-role workflow coordination)
if [ -n "$NEW_TASKS" ] || [ -n "$COMPLETED_TASKS" ]; then
    PM_NOTIFICATION_FILE="${STATE_DIR}/pm-notification.md"
    CURSOR_NOTIFICATION_FILE="${STATE_DIR}/cursor-notification.md" # compatibility alias
    cat > "$PM_NOTIFICATION_FILE" << EOF
# Notification to PM

**Generated**: $(date +"%Y-%m-%d %H:%M:%S")

## Status Changes

EOF

    if [ -n "$NEW_TASKS" ]; then
        echo "### 🆕 New Tasks" >> "$PM_NOTIFICATION_FILE"
        echo "" >> "$PM_NOTIFICATION_FILE"
        echo "New tasks have been requested by PM (pm:pending / compat: cursor:pending)." >> "$PM_NOTIFICATION_FILE"
        echo "" >> "$PM_NOTIFICATION_FILE"
    fi

    if [ -n "$COMPLETED_TASKS" ]; then
        echo "### ✅ Completed Tasks" >> "$PM_NOTIFICATION_FILE"
        echo "" >> "$PM_NOTIFICATION_FILE"
        echo "Impl Claude has completed tasks. Please review (cc:done)." >> "$PM_NOTIFICATION_FILE"
        echo "" >> "$PM_NOTIFICATION_FILE"
    fi

    echo "---" >> "$PM_NOTIFICATION_FILE"
    echo "" >> "$PM_NOTIFICATION_FILE"
    echo "**Next action**: PM Claude reviews and re-requests if needed (/handoff-to-impl-claude)." >> "$PM_NOTIFICATION_FILE"

    # Compatibility: also output same content to legacy filename
    cp -f "$PM_NOTIFICATION_FILE" "$CURSOR_NOTIFICATION_FILE" 2>/dev/null || true
fi
