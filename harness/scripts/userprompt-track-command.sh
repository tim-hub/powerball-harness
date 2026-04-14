#!/bin/bash
# userprompt-track-command.sh
# Detect slash commands on UserPromptSubmit and record usage
# + Create pending state for Skill-required commands
#
# Usage: Auto-executed from the UserPromptSubmit hook
# Input: stdin JSON (Claude Code hooks)
# Output: JSON (continue)

set +e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
STATE_DIR=".claude/state"
PENDING_DIR="${STATE_DIR}/pending-skills"
RECORD_USAGE="$SCRIPT_DIR/record-usage.js"

# List of Skill-required commands
# These commands are expected to use the Skill tool
SKILL_REQUIRED_COMMANDS="work|harness-review|validate|plan-with-agent"

# Extract value from JSON (jq preferred)
json_get() {
  local json="$1"
  local key="$2"
  local default="${3:-}"

  if command -v jq >/dev/null 2>&1; then
    echo "$json" | jq -r "$key // \"$default\"" 2>/dev/null || echo "$default"
  else
    echo "$default"
  fi
}

# Read JSON input from stdin
INPUT=""
if [ ! -t 0 ]; then
  INPUT="$(cat 2>/dev/null)"
fi

[ -z "$INPUT" ] && { echo '{"continue":true}'; exit 0; }

# Extract prompt
PROMPT=$(json_get "$INPUT" ".prompt" "")

# Skip empty prompts
[ -z "$PROMPT" ] && { echo '{"continue":true}'; exit 0; }

# Detect slash commands (line starts with /xxx)
# For multi-line input, only the first line is checked
FIRST_LINE=$(echo "$PROMPT" | head -n1)

if [[ "$FIRST_LINE" =~ ^/([a-zA-Z0-9_:/-]+) ]]; then
  RAW_COMMAND="${BASH_REMATCH[1]}"

  # Normalize command name (remove plugin prefix)
  # /claude-code-harness:core:work → work
  # /claude-code-harness/work → work
  # /work → work
  COMMAND_NAME="$RAW_COMMAND"
  # claude-code-harness:xxx:yyy → yyy (last segment)
  if [[ "$COMMAND_NAME" =~ ^claude-code-harness[:/] ]]; then
    COMMAND_NAME=$(echo "$COMMAND_NAME" | sed 's|.*[:/]||')
  fi

  # Record command usage
  if [ -f "$RECORD_USAGE" ] && [ -n "$COMMAND_NAME" ]; then
    node "$RECORD_USAGE" command "$COMMAND_NAME" >/dev/null 2>&1 || true
  fi

  # Check whether this is a Skill-required command
  if echo "$COMMAND_NAME" | grep -qiE "^($SKILL_REQUIRED_COMMANDS)$"; then
    # Permission hardening: prompt_preview contains user input,
    # restrict file permissions to owner-only (rwx------/rw-------)
    OLD_UMASK=$(umask)
    umask 077

    # Create pending directory (symlink bypass protection)
    if [ -L "$PENDING_DIR" ] || [ -L "$(dirname "$PENDING_DIR")" ]; then
      echo "[track-command] Warning: symlink detected in state path, skipping" >&2
      umask "$OLD_UMASK"
    else
    mkdir -p "$PENDING_DIR"

    # Create pending file (with timestamp)
    PENDING_FILE="${PENDING_DIR}/${COMMAND_NAME}.pending"
    # Security: refuse if pending file is a symlink
    if [ -L "$PENDING_FILE" ]; then
      echo "[track-command] Warning: symlink detected at $PENDING_FILE, skipping" >&2
      umask "$OLD_UMASK"
    else
    cat > "$PENDING_FILE" <<EOF
{
  "command": "$COMMAND_NAME",
  "started_at": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "prompt_preview": "$(echo "$PROMPT" | head -c 200 | tr '\n' ' ' | sed 's/"/\\"/g')"
}
EOF

    # Restore original umask
    umask "$OLD_UMASK"
    fi  # end symlink check for PENDING_FILE
    fi  # end symlink check for PENDING_DIR
  fi
fi

echo '{"continue":true}'
exit 0
