#!/bin/bash
# plans-format-check.sh
# Check the format of Plans.md and warn / suggest migration if old format is detected

set -uo pipefail

PLANS_FILE="${1:-Plans.md}"

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

# If Plans.md does not exist
if [ ! -f "$PLANS_FILE" ]; then
  output_json "skip" "Plans.md not found" "false"
  exit 0
fi

# Format check
ISSUES=()
MIGRATION_NEEDED=false

# 1. Check for deprecated markers (cursor:WIP, cursor:done)
if grep -qE 'cursor:(WIP|done)' "$PLANS_FILE" 2>/dev/null; then
  MIGRATION_NEEDED=true
  ISSUES+=("\"cursor:WIP and cursor:done are deprecated. Please migrate to pm:pending / pm:confirmed.\"")
fi

# 2. Check for marker legend section
if ! grep -qE '## Marker Legend' "$PLANS_FILE" 2>/dev/null; then
  ISSUES+=("\"Marker legend section is missing. Recommended to add from the template.\"")
fi

# 3. Check for valid harness markers
# New format: cc:TODO, cc:WIP, cc:WORK, cc:DONE, cc:done, cc:blocked, pm:pending, pm:confirmed, cursor:pending, cursor:confirmed
if ! grep -qE 'cc:(TODO|WIP|WORK|DONE|done|blocked)|pm:(pending|confirmed)|cursor:(pending|confirmed)' "$PLANS_FILE" 2>/dev/null; then
  # Also check old format (cursor:WIP/done)
  if ! grep -qE 'cursor:(WIP|done)' "$PLANS_FILE" 2>/dev/null; then
    ISSUES+=("\"Harness markers (cc:TODO, cc:WIP, etc.) not found.\"")
  fi
fi

# Output result
if [ ${#ISSUES[@]} -eq 0 ]; then
  output_json "ok" "Plans.md format is up to date" "false"
else
  ISSUES_JSON=$(printf '%s,' "${ISSUES[@]}" | sed 's/,$//')
  if [ "$MIGRATION_NEEDED" = true ]; then
    output_json "migration_required" "Old format detected in Plans.md. Migration available via /harness-update." "true" "[$ISSUES_JSON]"
  else
    output_json "warning" "Plans.md has items that could be improved" "false" "[$ISSUES_JSON]"
  fi
fi
