#!/bin/bash
# setup-hook.sh
# Setup Hook: Setup processing on claude --init / --maintenance
#
# Usage:
#   setup-hook.sh init        # Initial setup
#   setup-hook.sh maintenance # Maintenance processing
#
# Output: Outputs hookSpecificOutput in JSON format

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MODE="${1:-init}"

# ===== SIMPLE mode detection =====
SIMPLE_MODE="false"
if [ -f "$SCRIPT_DIR/check-simple-mode.sh" ]; then
  # shellcheck source=./check-simple-mode.sh
  source "$SCRIPT_DIR/check-simple-mode.sh"
  if is_simple_mode; then
    SIMPLE_MODE="true"
    echo -e "\033[1;33m[WARNING]\033[0m CLAUDE_CODE_SIMPLE mode detected — skills/agents/memory disabled" >&2
  fi
fi

# Read JSON input from stdin (Claude Code v2.1.10+)
INPUT=""
if [ ! -t 0 ]; then
  INPUT=$(cat 2>/dev/null || true)
fi

# ===== Common helpers =====
output_json() {
  local message="$1"
  cat <<EOF
{"hookSpecificOutput":{"hookEventName":"Setup","additionalContext":"$message"}}
EOF
}

# ===== Init mode: initial setup =====
run_init() {
  local messages=()

  # 1. Sync plugin cache
  if [ -f "$SCRIPT_DIR/sync-plugin-cache.sh" ]; then
    bash "$SCRIPT_DIR/sync-plugin-cache.sh" >/dev/null 2>&1 || true
    messages+=("Plugin cache sync complete")
  fi

  # 2. Initialize state directory
  STATE_DIR=".claude/state"
  mkdir -p "$STATE_DIR"

  # 3. Generate default config file (if not exists)
  CONFIG_FILE=".claude-code-harness.config.yaml"
  if [ ! -f "$CONFIG_FILE" ]; then
    if [ -f "$SCRIPT_DIR/../templates/.claude-code-harness.config.yaml.template" ]; then
      cp "$SCRIPT_DIR/../templates/.claude-code-harness.config.yaml.template" "$CONFIG_FILE"
      messages+=("Config file generated")
    fi
  fi

  # 4. Generate CLAUDE.md (if not exists)
  if [ ! -f "CLAUDE.md" ]; then
    if [ -f "$SCRIPT_DIR/../templates/CLAUDE.md.template" ]; then
      cp "$SCRIPT_DIR/../templates/CLAUDE.md.template" "CLAUDE.md"
      messages+=("CLAUDE.md generated")
    fi
  fi

  # 5. Generate Plans.md (if not exists)
  # Consider plansDirectory setting
  if [ -f "$SCRIPT_DIR/config-utils.sh" ]; then
    source "$SCRIPT_DIR/config-utils.sh"
    PLANS_PATH=$(get_plans_file_path)
  else
    PLANS_PATH="Plans.md"
  fi

  if [ ! -f "$PLANS_PATH" ]; then
    # Create directory if it does not exist
    PLANS_DIR=$(dirname "$PLANS_PATH")
    [ "$PLANS_DIR" != "." ] && mkdir -p "$PLANS_DIR"

    if [ -f "$SCRIPT_DIR/../templates/Plans.md.template" ]; then
      cp "$SCRIPT_DIR/../templates/Plans.md.template" "$PLANS_PATH"
      messages+=("Plans.md generated")
    fi
  fi

  # 6. Initialize template tracker
  if [ -f "$SCRIPT_DIR/template-tracker.sh" ]; then
    bash "$SCRIPT_DIR/template-tracker.sh" init >/dev/null 2>&1 || true
  fi

  # Add SIMPLE mode warning
  if [ "$SIMPLE_MODE" = "true" ]; then
    messages+=("WARNING: CLAUDE_CODE_SIMPLE mode — skills/agents/memory disabled, hooks only")
  fi

  # Output result
  if [ ${#messages[@]} -eq 0 ]; then
    output_json "[Setup:init] Harness is already initialized"
  else
    local msg_str
    msg_str=$(IFS=', '; echo "${messages[*]}")
    output_json "[Setup:init] $msg_str"
  fi
}

# ===== Maintenance mode: maintenance processing =====
run_maintenance() {
  local messages=()

  # 1. Sync plugin cache
  if [ -f "$SCRIPT_DIR/sync-plugin-cache.sh" ]; then
    bash "$SCRIPT_DIR/sync-plugin-cache.sh" >/dev/null 2>&1 || true
    messages+=("Cache sync complete")
  fi

  # 2. Clean up old session files
  STATE_DIR=".claude/state"
  ARCHIVE_DIR="$STATE_DIR/sessions"

  if [ -d "$ARCHIVE_DIR" ]; then
    # Delete session archives older than 7 days
    find "$ARCHIVE_DIR" -name "session-*.json" -mtime +7 -delete 2>/dev/null || true
    messages+=("Old session archives deleted")
  fi

  # 3. Clean up temporary files
  if [ -d "$STATE_DIR" ]; then
    # Delete .tmp files
    find "$STATE_DIR" -name "*.tmp" -delete 2>/dev/null || true
  fi

  # 4. Check for template updates
  if [ -f "$SCRIPT_DIR/template-tracker.sh" ]; then
    CHECK_RESULT=$(bash "$SCRIPT_DIR/template-tracker.sh" check 2>/dev/null || echo '{"needsCheck": false}')
    if command -v jq >/dev/null 2>&1; then
      NEEDS_UPDATE=$(echo "$CHECK_RESULT" | jq -r '.needsCheck // false')
      if [ "$NEEDS_UPDATE" = "true" ]; then
        UPDATES_COUNT=$(echo "$CHECK_RESULT" | jq -r '.updatesCount // 0')
        messages+=("Template updates available: ${UPDATES_COUNT} item(s)")
      fi
    fi
  fi

  # 5. Add SIMPLE mode warning
  if [ "$SIMPLE_MODE" = "true" ]; then
    messages+=("WARNING: CLAUDE_CODE_SIMPLE mode — skills/agents/memory disabled, hooks only")
  fi

  # 6. Validate config file
  CONFIG_FILE=".claude-code-harness.config.yaml"
  if [ -f "$CONFIG_FILE" ]; then
    # Basic YAML syntax check
    if command -v python3 >/dev/null 2>&1; then
      if ! python3 -c "import yaml; yaml.safe_load(open('$CONFIG_FILE'))" 2>/dev/null; then
        messages+=("Warning: config file syntax error")
      fi
    fi
  fi

  # Output result
  if [ ${#messages[@]} -eq 0 ]; then
    output_json "[Setup:maintenance] Maintenance complete (no changes)"
  else
    local msg_str
    msg_str=$(IFS=', '; echo "${messages[*]}")
    output_json "[Setup:maintenance] $msg_str"
  fi
}

# ===== Main processing =====
case "$MODE" in
  init)
    run_init
    ;;
  maintenance)
    run_maintenance
    ;;
  *)
    output_json "[Setup] Unknown mode: $MODE"
    exit 1
    ;;
esac
