#!/bin/bash
# auto-cleanup-hook.sh
# PostToolUse Hook: Automatically checks file size after writing to Plans.md, etc.
#
# Input: JSON from stdin (tool_name, tool_input, etc.)
# Output: Feedback via additionalContext

set +e

# Read input JSON (Claude Code hooks pass JSON via stdin)
INPUT=""
if [ ! -t 0 ]; then
  INPUT="$(cat 2>/dev/null)"
fi

# Get file_path / cwd from stdin JSON (fall back to python3 if jq is unavailable)
FILE_PATH=""
CWD=""
if [ -n "$INPUT" ]; then
  if command -v jq >/dev/null 2>&1; then
    FILE_PATH="$(printf '%s' "$INPUT" | jq -r '.tool_input.file_path // .tool_response.filePath // empty' 2>/dev/null)"
    CWD="$(printf '%s' "$INPUT" | jq -r '.cwd // empty' 2>/dev/null)"
  elif command -v python3 >/dev/null 2>&1; then
    eval "$(printf '%s' "$INPUT" | python3 -c '
import json, shlex, sys
try:
    data = json.load(sys.stdin)
except Exception:
    data = {}
cwd = data.get("cwd") or ""
tool_input = data.get("tool_input") or {}
tool_response = data.get("tool_response") or {}
file_path = tool_input.get("file_path") or tool_response.get("filePath") or ""
print(f"CWD_FROM_STDIN={shlex.quote(cwd)}")
print(f"FILE_PATH_FROM_STDIN={shlex.quote(file_path)}")
' 2>/dev/null)"
    FILE_PATH="${FILE_PATH_FROM_STDIN:-}"
    CWD="${CWD_FROM_STDIN:-}"
  fi
fi

# Exit if file_path is empty
if [ -z "$FILE_PATH" ]; then
  exit 0
fi

# Normalize to project-relative path when possible (also works with absolute paths, but matching is more reliable)
if [ -n "$CWD" ] && [[ "$FILE_PATH" == "$CWD/"* ]]; then
  FILE_PATH="${FILE_PATH#$CWD/}"
fi

# Default thresholds
PLANS_MAX_LINES=${PLANS_MAX_LINES:-200}
SESSION_LOG_MAX_LINES=${SESSION_LOG_MAX_LINES:-500}
CLAUDE_MD_MAX_LINES=${CLAUDE_MD_MAX_LINES:-100}

# Variable to store feedback
FEEDBACK=""

# Check Plans.md
if [[ "$FILE_PATH" == *"Plans.md"* ]] || [[ "$FILE_PATH" == *"plans.md"* ]]; then
  if [ -f "$FILE_PATH" ]; then
    lines=$(wc -l < "$FILE_PATH" | tr -d ' ')
    if [ "$lines" -gt "$PLANS_MAX_LINES" ]; then
      FEEDBACK="⚠️ Plans.md has ${lines} lines (limit: ${PLANS_MAX_LINES}). It is recommended to archive old tasks with /harness-plan archive."
    fi

    # SSOT sync check when Plans.md cleanup (archive move) is detected
    # Confirm that /harness-remember sync has been run before editing the archive section
    if grep -q "📦 Archive\|## Archive\|Archive" "$FILE_PATH" 2>/dev/null; then
      # Resolve repository root for consistent state directory lookup
      CWD="${CWD:-$(pwd)}"  # Fallback to pwd if empty
      REPO_ROOT=$(git -C "$CWD" rev-parse --show-toplevel 2>/dev/null) || REPO_ROOT="$CWD"
      STATE_DIR="${REPO_ROOT}/.claude/state"

      SSOT_FLAG="${STATE_DIR}/.ssot-synced-this-session"

      if [ ! -f "$SSOT_FLAG" ]; then
        # If flag is missing, add a warning prompting SSOT sync
        SSOT_WARNING="**Please run /harness-remember sync before cleaning up Plans.md** - Important decisions and learnings may not yet be reflected in SSOT (decisions.md/patterns.md)."

        if [ -n "$FEEDBACK" ]; then
          FEEDBACK="${FEEDBACK} | ${SSOT_WARNING}"
        else
          FEEDBACK="⚠️ ${SSOT_WARNING}"
        fi
      fi
    fi
  fi
fi

# Check session-log.md
if [[ "$FILE_PATH" == *"session-log.md"* ]]; then
  if [ -f "$FILE_PATH" ]; then
    lines=$(wc -l < "$FILE_PATH" | tr -d ' ')
    if [ "$lines" -gt "$SESSION_LOG_MAX_LINES" ]; then
      FEEDBACK="⚠️ session-log.md has ${lines} lines (limit: ${SESSION_LOG_MAX_LINES}). Run /maintenance --prune-logs to prune older months."
    fi
  fi
fi

# Check CLAUDE.md
if [[ "$FILE_PATH" == *"CLAUDE.md"* ]] || [[ "$FILE_PATH" == *"claude.md"* ]]; then
  if [ -f "$FILE_PATH" ]; then
    lines=$(wc -l < "$FILE_PATH" | tr -d ' ')
    if [ "$lines" -gt "$CLAUDE_MD_MAX_LINES" ]; then
      FEEDBACK="⚠️ CLAUDE.md is ${lines} lines. Consider splitting it into .claude/rules/, or moving content to docs/ and referencing it as @docs/filename.md."
    fi
  fi
fi

# Output JSON if there is feedback
if [ -n "$FEEDBACK" ]; then
  echo "{\"hookSpecificOutput\": {\"hookEventName\": \"PostToolUse\", \"additionalContext\": \"$FEEDBACK\"}}"
fi

# Always exit with success
exit 0
