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

# Check if Plans.md exists
plans_file_exists() {
  local plans_path
  plans_path=$(get_plans_file_path)
  [ -f "$plans_path" ]
}
