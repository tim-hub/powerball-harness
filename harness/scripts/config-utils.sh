#!/bin/bash
# config-utils.sh
# Utility for retrieving values from the harness configuration file
#
# Usage: source "${SCRIPT_DIR}/config-utils.sh"
#        plans_path=$(get_plans_file_path)

# Default path to the configuration file
CONFIG_FILE="${CONFIG_FILE:-.claude-code-harness.config.yaml}"

# Validate plansDirectory (security)
# Reject absolute paths, parent directory references, and symlink escapes
validate_plans_directory() {
  local value="$1"
  local default="."

  # Return default if empty
  [ -z "$value" ] && echo "$default" && return 0

  # Security: Reject absolute paths
  case "$value" in
    /*) echo "$default" && return 0 ;;
  esac

  # Security: Reject parent directory references (..)
  case "$value" in
    *..*)  echo "$default" && return 0 ;;
  esac

  # Security: Detect symlink escapes (when realpath is available)
  if command -v realpath >/dev/null 2>&1 && [ -e "$value" ]; then
    local project_root
    local resolved_path
    project_root=$(realpath "." 2>/dev/null) || project_root=$(pwd)
    resolved_path=$(realpath "$value" 2>/dev/null)

    if [ -n "$resolved_path" ]; then
      # Confirm the resolved path is within the project root
      case "$resolved_path" in
        "$project_root"/*) ;; # OK: inside project
        "$project_root") ;;   # OK: project root itself
        *) echo "$default" && return 0 ;; # NG: outside project
      esac
    fi
  fi

  echo "$value"
}

# Get plansDirectory setting (default: ".")
get_plans_directory() {
  local default="."

  if [ ! -f "$CONFIG_FILE" ]; then
    echo "$default"
    return 0
  fi

  local value=""

  # If yq is available
  if command -v yq >/dev/null 2>&1; then
    value=$(yq -r '.plansDirectory // empty' "$CONFIG_FILE" 2>/dev/null)
  fi

  # If not retrievable via yq, try Python
  if [ -z "$value" ] && command -v python3 >/dev/null 2>&1; then
    # Parse YAML with Python (returns empty if pyyaml is not installed)
    value=$(python3 - "$CONFIG_FILE" <<'PY' 2>/dev/null
import sys
try:
    import yaml
    with open(sys.argv[1]) as f:
        data = yaml.safe_load(f) or {}
    print(data.get('plansDirectory', ''))
except ImportError:
    # pyyaml not installed - return empty to trigger grep fallback
    pass
except:
    pass
PY
)
  fi

  # If not retrievable via yq/Python, fall back to grep + sed
  if [ -z "$value" ]; then
    value=$(grep "^plansDirectory:" "$CONFIG_FILE" 2>/dev/null | sed 's/^plansDirectory:[[:space:]]*//' | tr -d '"' | tr -d "'" || echo "")
  fi

  # Validate before returning
  validate_plans_directory "$value"
}

# Get the full path to Plans.md
get_plans_file_path() {
  local plans_dir
  plans_dir=$(get_plans_directory)

  # Search for Plans.md in the directory (case-insensitive)
  for f in Plans.md plans.md PLANS.md PLANS.MD; do
    local full_path="${plans_dir}/${f}"
    # When ".", omit the "./" prefix
    [ "$plans_dir" = "." ] && full_path="$f"

    if [ -f "$full_path" ]; then
      echo "$full_path"
      return 0
    fi
  done

  # If not found, return the default path
  local default_path="${plans_dir}/Plans.md"
  [ "$plans_dir" = "." ] && default_path="Plans.md"
  echo "$default_path"
}

# Check if Plans.md exists (legacy — used only by the Plans.md → plans.json
# migration bridge; new code should use plans_json_exists).
plans_file_exists() {
  local plans_path
  plans_path=$(get_plans_file_path)
  [ -f "$plans_path" ]
}

# ---------------------------------------------------------------------------
# plans.json SSOT query helpers
#
# `.claude/harness/plans.json` is the single source of truth for tasks/phases.
# These helpers read it directly with jq (fast path for always-on hooks).
# `harness plan-cli` remains the read/write interface for agents and the web UI.
# All helpers degrade gracefully (empty/zero) when plans.json or jq is absent,
# so hooks never hard-fail on a fresh project.
# ---------------------------------------------------------------------------

# Path to plans.json (honors plansDirectory; canonical location is
# <plansDir>/.claude/harness/plans.json, default ./.claude/harness/plans.json).
get_plans_json_path() {
  local plans_dir
  plans_dir=$(get_plans_directory)
  if [ "$plans_dir" = "." ]; then
    echo ".claude/harness/plans.json"
  else
    echo "${plans_dir}/.claude/harness/plans.json"
  fi
}

# Exit 0 if plans.json exists.
plans_json_exists() {
  [ -f "$(get_plans_json_path)" ]
}

# Run a jq filter over plans.json. Echoes empty and returns non-zero when
# jq is unavailable or plans.json is missing.
_plans_jq() {
  local filter="$1"
  local pj
  pj="$(get_plans_json_path)"
  command -v jq >/dev/null 2>&1 || return 1
  [ -f "$pj" ] || return 1
  jq -r "$filter" "$pj" 2>/dev/null
}

# Count tasks with a given status (e.g. cc:WIP, cc:TODO, cc:done).
# Note: local var is named want_status, not status — `status` is a read-only
# special variable in zsh/fish and breaks if this lib is sourced there.
plans_count_status() {
  local want_status="$1"
  local n
  n="$(_plans_jq "[.phases[].tasks[]? | select(.status==\"${want_status}\")] | length")"
  echo "${n:-0}"
}

# Exit 0 if at least one task is cc:WIP.
plans_has_wip() {
  [ "$(plans_count_status 'cc:WIP')" -gt 0 ] 2>/dev/null
}

# Print names of cc:WIP tasks, one per line (optional arg: max count).
plans_wip_names() {
  local limit="${1:-0}"
  local names
  names="$(_plans_jq '[.phases[].tasks[]? | select(.status=="cc:WIP") | .name] | .[]')"
  if [ "$limit" -gt 0 ] 2>/dev/null; then
    echo "$names" | head -n "$limit"
  else
    echo "$names"
  fi
}

# Exit 0 if any cc:WIP task carries the skip:tdd quality marker.
plans_wip_has_skip_tdd() {
  local n
  n="$(_plans_jq '[.phases[].tasks[]? | select(.status=="cc:WIP") | select((.qualityMarkers // []) | index("skip:tdd"))] | length')"
  [ "${n:-0}" -gt 0 ] 2>/dev/null
}
