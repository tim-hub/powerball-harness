#!/bin/bash
# session-auto-broadcast.sh
# Automatic broadcast on file changes
#
# Called from PostToolUse (Write|Edit)
# Auto-notify on changes to important files (API, type definitions, etc.)
#
# Input: JSON from stdin (includes tool_input)
# Output: JSON (hookSpecificOutput)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# ===== Configuration =====
# Patterns for auto-broadcast targets
AUTO_BROADCAST_PATTERNS=(
  "src/api/"
  "src/types/"
  "src/interfaces/"
  "api/"
  "types/"
  "schema.prisma"
  "openapi"
  "swagger"
  ".graphql"
)

# Configuration file path
CONFIG_FILE=".claude/sessions/auto-broadcast.json"

# ===== Read JSON input from stdin =====
INPUT=""
if [ -t 0 ]; then
  : # No input when stdin is a TTY
else
  INPUT=$(cat 2>/dev/null || true)
fi

# ===== Extract file path =====
FILE_PATH=""
if [ -n "$INPUT" ] && command -v jq >/dev/null 2>&1; then
  FILE_PATH="$(echo "$INPUT" | jq -r '.tool_input.file_path // .tool_input.path // empty' 2>/dev/null)"
fi

# Exit if no file path
if [ -z "$FILE_PATH" ]; then
  echo '{"hookSpecificOutput":{"hookEventName":"PostToolUse","additionalContext":""}}'
  exit 0
fi

# ===== Check if auto-broadcast is enabled =====
AUTO_BROADCAST_ENABLED="true"
if [ -f "$CONFIG_FILE" ] && command -v jq >/dev/null 2>&1; then
  AUTO_BROADCAST_ENABLED="$(jq -r '.enabled // true' "$CONFIG_FILE" 2>/dev/null)"
fi

if [ "$AUTO_BROADCAST_ENABLED" != "true" ]; then
  echo '{"hookSpecificOutput":{"hookEventName":"PostToolUse","additionalContext":""}}'
  exit 0
fi

# ===== Pattern matching =====
should_broadcast="false"
matched_pattern=""

for pattern in "${AUTO_BROADCAST_PATTERNS[@]}"; do
  if [[ "$FILE_PATH" == *"$pattern"* ]]; then
    should_broadcast="true"
    matched_pattern="$pattern"
    break
  fi
done

# Check custom patterns too
if [ "$should_broadcast" = "false" ] && [ -f "$CONFIG_FILE" ] && command -v jq >/dev/null 2>&1; then
  CUSTOM_PATTERNS=$(jq -r '.patterns // [] | .[]' "$CONFIG_FILE" 2>/dev/null)
  while IFS= read -r pattern; do
    if [ -n "$pattern" ] && [[ "$FILE_PATH" == *"$pattern"* ]]; then
      should_broadcast="true"
      matched_pattern="$pattern"
      break
    fi
  done <<< "$CUSTOM_PATTERNS"
fi

# ===== Execute broadcast =====
if [ "$should_broadcast" = "true" ]; then
  # Extract filename
  FILE_NAME=$(basename "$FILE_PATH")

  # Execute broadcast
  bash "$SCRIPT_DIR/session-broadcast.sh" --auto "$FILE_PATH" "Matched pattern '$matched_pattern'" >/dev/null 2>/dev/null || true

  # Output notification message
  cat <<EOF
{"hookSpecificOutput":{"hookEventName":"PostToolUse","additionalContext":"📢 Auto-broadcast: Notified other sessions of changes to ${FILE_NAME}"}}
EOF
else
  echo '{"hookSpecificOutput":{"hookEventName":"PostToolUse","additionalContext":""}}'
fi

exit 0
