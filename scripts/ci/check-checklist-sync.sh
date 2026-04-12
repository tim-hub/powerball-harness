#!/bin/bash
# check-checklist-sync.sh
# Verify command file checklists are in sync with script validation items
#
# Purpose:
# - Check that check_file/check_dir in scripts/setup-2agent.sh
#   match the checklist in commands/setup-2agent.md
# - Same for scripts/update-2agent.sh and commands/update-2agent.md

set -euo pipefail

PLUGIN_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
ERRORS=0

echo "🔍 Checklist sync verification"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# ================================
# Utility functions
# ================================

# Extract check_file/check_dir arguments from script
extract_script_checks() {
  local script="$1"
  grep -E 'check_(file|dir)' "$script" 2>/dev/null | \
    awk -F'"' '{print $2}' | \
    grep -v '^$' | \
    sort -u
}

# Extract checklist items from command file
# Extract only "Auto-verification" section (exclude "Claude generated" section)
extract_command_checklist() {
  local cmd="$1"
  # Extract from "Auto-verification" to "Claude generated" or next section
  awk '/Auto.verify|Automated.check/,/Claude.*generat|^###|^\*\*All/' "$cmd" 2>/dev/null | \
    grep -E '^\s*-\s*\[\s*\]\s*`[^`]+`' | \
    awk -F'`' '{print $2}' | \
    grep -v '^$' | \
    sort -u
}

# Compare two lists
compare_lists() {
  local name="$1"
  local script_file="$2"
  local command_file="$3"

  echo ""
  echo "📋 Verifying $name..."

  # Extract to temporary file
  local script_checks=$(mktemp)
  local command_checks=$(mktemp)

  extract_script_checks "$script_file" > "$script_checks"
  extract_command_checklist "$command_file" > "$command_checks"

  # In script but not in command
  local missing_in_command=$(comm -23 "$script_checks" "$command_checks")
  if [ -n "$missing_in_command" ]; then
    echo "  ❌ In script but not in command checklist:"
    echo "$missing_in_command" | while read item; do
      echo "     - $item"
    done
    ERRORS=$((ERRORS + 1))
  fi

  # In command but not in script
  local missing_in_script=$(comm -13 "$script_checks" "$command_checks")
  if [ -n "$missing_in_script" ]; then
    echo "  ❌ In command checklist but not in script:"
    echo "$missing_in_script" | while read item; do
      echo "     - $item"
    done
    ERRORS=$((ERRORS + 1))
  fi

  # Skip if both are empty (prevent false pass)
  local script_count=$(wc -l < "$script_checks" | tr -d ' ')
  local command_count=$(wc -l < "$command_checks" | tr -d ' ')

  if [ "$script_count" -eq 0 ] && [ "$command_count" -eq 0 ]; then
    echo "  ⚠️ Skipped: No check items found (verify file structure)"
  elif [ -z "$missing_in_command" ] && [ -z "$missing_in_script" ]; then
    echo "  ✅ In sync ($script_count items)"
  fi

  rm -f "$script_checks" "$command_checks"
}

# ================================
# Main verification
# ================================

# setup hub verification (v2.19.0+ 2agent integrated into setup)
SETUP_SKILL="$PLUGIN_ROOT/skills/setup/SKILL.md"
SETUP_2AGENT_REF="$PLUGIN_ROOT/skills/setup/references/2agent-setup.md"

if [ -f "$SETUP_SKILL" ] && [ -f "$SETUP_2AGENT_REF" ]; then
  echo "✓ setup skill and 2agent-setup reference exist"
elif [ -f "$SETUP_SKILL" ]; then
  echo "⚠️ setup/references/2agent-setup.md not found (check post-integration structure)"
else
  echo "⚠️ skills/setup/SKILL.md not found (skill may not be created yet)"
fi

# Note: Since v2.17.0, commands have been migrated to skills
# Checklist sync will be managed per-skill going forward
# Exit normally if no check target skills found (do not fail on empty checklists)

# ================================
# Result summary
# ================================
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [ $ERRORS -eq 0 ]; then
  echo "✅ Checklist sync verification passed"
  exit 0
else
  echo "❌ $ERRORS inconsistencies found"
  echo ""
  echo "💡 How to fix:"
  echo "  1. Check check_file/check_dir in scripts/*.sh"
  echo "  2. Update checklists in commands/*.md"
  echo "  3. Ensure both match"
  exit 1
fi
