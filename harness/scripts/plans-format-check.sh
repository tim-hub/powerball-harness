#!/bin/bash
# plans-format-check.sh
# SSOT is .claude/harness/plans.json (validated by the Go plan-cli structs).
# This check no longer asserts the legacy Plans.md markdown table format. Instead:
#   - If a legacy Plans.md exists, advise running `harness plan-cli migrate`.
#   - Otherwise, validate that plans.json parses cleanly (jq empty check) and exit 0.

set -uo pipefail

# Source plans.json helpers (SSOT: .claude/harness/plans.json)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [ -f "${SCRIPT_DIR}/config-utils.sh" ]; then
  # shellcheck source=./config-utils.sh
  source "${SCRIPT_DIR}/config-utils.sh"
fi

# Function for JSON output
output_json() {
  local status="$1"
  local message="$2"
  local migration_needed="${3:-false}"
  local issues="${4:-[]}"

  cat <<EOF
{
  "status": "$status",
  "message": "$message",
  "migration_needed": $migration_needed,
  "issues": $issues,
  "hookSpecificOutput": {
    "hookEventName": "SessionStart",
    "additionalContext": "$message"
  }
}
EOF
}

# Legacy Plans.md migration bridge: if the markdown still exists, advise migrating.
if declare -F plans_file_exists >/dev/null 2>&1 && plans_file_exists; then
  output_json "migration_required" \
    "Legacy Plans.md detected. plans.json is now the source of truth — run \`harness plan-cli migrate\` to convert." \
    "true"
  exit 0
fi

# Resolve plans.json path
if declare -F get_plans_json_path >/dev/null 2>&1; then
  PLANS_JSON="$(get_plans_json_path)"
else
  PLANS_JSON=".claude/harness/plans.json"
fi

# If plans.json does not exist, there is nothing to validate.
if [ ! -f "$PLANS_JSON" ]; then
  output_json "skip" "plans.json not found" "false"
  exit 0
fi

# Validate that plans.json parses cleanly.
if command -v jq >/dev/null 2>&1; then
  if jq empty "$PLANS_JSON" >/dev/null 2>&1; then
    output_json "ok" "plans.json parses cleanly" "false"
  else
    output_json "warning" "plans.json is not valid JSON — run \`harness plan-cli\` to inspect/repair." "false"
  fi
else
  # No jq available — cannot validate, but plans.json is the SSOT, so do not fail.
  output_json "skip" "jq not available; skipped plans.json validation" "false"
fi

exit 0
