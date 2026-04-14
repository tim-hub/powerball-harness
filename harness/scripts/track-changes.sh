#!/bin/bash
# track-changes.sh
# Track file changes and update the state file
#
# Usage: Auto-executed from the PostToolUse hook
# Input: stdin JSON (Claude Code hooks) / compat: $1=tool_name, $2=file_path
#
# Cross-platform: Supports Windows (Git Bash/MSYS2/Cygwin/WSL), macOS, Linux

set +e

# Load cross-platform path utilities
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [ -f "$SCRIPT_DIR/path-utils.sh" ]; then
  # shellcheck source=./path-utils.sh
  source "$SCRIPT_DIR/path-utils.sh"
else
  # Fallback: minimal normalize_path and is_path_under
  normalize_path() {
    local p="$1"
    p="${p//\\//}"
    echo "$p"
  }
  is_path_under() {
    local child="$1"
    local parent="$2"
    child="$(normalize_path "$child")"
    parent="$(normalize_path "$parent")"
    [[ "$parent" != */ ]] && parent="${parent}/"
    [[ "${child}/" == "${parent}"* ]] || [ "$child" = "${parent%/}" ]
  }
fi

INPUT=""
if [ ! -t 0 ]; then
  INPUT="$(cat 2>/dev/null)"
fi

TOOL_NAME="${1:-}"
FILE_PATH="${2:-}"
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

  [ -z "$TOOL_NAME" ] && TOOL_NAME="${TOOL_NAME_FROM_STDIN:-}"
  [ -z "$FILE_PATH" ] && FILE_PATH="${FILE_PATH_FROM_STDIN:-}"
  CWD="${CWD_FROM_STDIN:-}"
fi

TOOL_NAME="${TOOL_NAME:-unknown}"

# Normalize to project-relative path when possible (cross-platform)
if [ -n "$CWD" ] && [ -n "$FILE_PATH" ]; then
  NORM_FILE_PATH="$(normalize_path "$FILE_PATH")"
  NORM_CWD="$(normalize_path "$CWD")"

  if is_path_under "$NORM_FILE_PATH" "$NORM_CWD"; then
    # Remove the CWD prefix to get relative path
    cwd_with_slash="${NORM_CWD%/}/"
    if [[ "$NORM_FILE_PATH" == "$cwd_with_slash"* ]]; then
      FILE_PATH="${NORM_FILE_PATH#$cwd_with_slash}"
    fi
  fi
fi
STATE_FILE=".claude/state/session.json"
CURRENT_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ)

# Skip if the state file does not exist
if [ ! -f "$STATE_FILE" ]; then
  exit 0
fi

# Skip if file path is empty
if [ -z "$FILE_PATH" ]; then
  exit 0
fi

# Detect changes to important files
IMPORTANT_FILES="Plans.md CLAUDE.md AGENTS.md"
IS_IMPORTANT="false"

for important in $IMPORTANT_FILES; do
  if [[ "$FILE_PATH" == *"$important"* ]]; then
    IS_IMPORTANT="true"
    break
  fi
done

# Detect test files
if [[ "$FILE_PATH" == *".test."* ]] || [[ "$FILE_PATH" == *".spec."* ]] || [[ "$FILE_PATH" == *"__tests__"* ]]; then
  IS_IMPORTANT="true"
fi

# Record changes (use jq if available, otherwise skip)
if command -v jq &> /dev/null; then
  # Add a new change entry
  TEMP_FILE=$(mktemp /tmp/harness-tmp.XXXXXX 2>/dev/null) || {
    # Silently skip on mktemp failure (do not interrupt the PostToolUse hook)
    exit 0
  }
  # Guarantee cleanup
  trap 'rm -f "$TEMP_FILE"' EXIT

  if jq --arg file "$FILE_PATH" \
        --arg action "$TOOL_NAME" \
        --arg timestamp "$CURRENT_TIME" \
        --arg important "$IS_IMPORTANT" \
        '.changes_this_session += [{
          "file": $file,
          "action": $action,
          "timestamp": $timestamp,
          "important": ($important == "true")
        }]' "$STATE_FILE" > "$TEMP_FILE" 2>/dev/null; then
    mv "$TEMP_FILE" "$STATE_FILE" 2>/dev/null || true
  fi
fi

# Notify when important files change
if [ "$IS_IMPORTANT" = "true" ]; then
  case "$FILE_PATH" in
    *Plans.md*)
      echo "📋 Plans.md has been updated"
      ;;
    *CLAUDE.md*)
      echo "📝 CLAUDE.md has been updated"
      ;;
    *AGENTS.md*)
      echo "📝 AGENTS.md has been updated"
      ;;
    *.test.*|*.spec.*|*__tests__*)
      echo "🧪 Test file updated: $(basename "$FILE_PATH")"
      ;;
  esac
fi

# ==============================================================================
# Reset review_status in Work mode
# ==============================================================================
# When a Write/Edit occurs during /work, reset review_status to pending
# This ensures a re-review is always required after a code change
# Backward compatibility: prefer work-active.json, fall back to ultrawork-active.json
# ==============================================================================
WORK_FILE=".claude/state/work-active.json"
if [ ! -f "$WORK_FILE" ]; then
  WORK_FILE=".claude/state/ultrawork-active.json"
fi
if [ -f "$WORK_FILE" ] && command -v jq >/dev/null 2>&1; then
  CURRENT_STATUS=$(jq -r '.review_status // "pending"' "$WORK_FILE" 2>/dev/null)

  # Reset to pending only when status is passed or failed
  if [ "$CURRENT_STATUS" = "passed" ] || [ "$CURRENT_STATUS" = "failed" ]; then
    TEMP_UW=$(mktemp /tmp/harness-tmp.XXXXXX 2>/dev/null)
    if [ -n "$TEMP_UW" ]; then
      if jq '.review_status = "pending"' "$WORK_FILE" > "$TEMP_UW" 2>/dev/null; then
        mv "$TEMP_UW" "$WORK_FILE" 2>/dev/null || rm -f "$TEMP_UW"
        echo "⚠️ work: code change detected → resetting review_status to pending (re-review required)" >&2
      else
        rm -f "$TEMP_UW"
      fi
    fi
  fi
fi

exit 0
