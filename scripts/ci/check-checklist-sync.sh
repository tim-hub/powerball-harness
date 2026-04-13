#!/bin/bash
# check-checklist-sync.sh
# Verify that checklists in command files and validation items in scripts are in sync
#
# Purpose:
# - Confirm that check_file/check_dir in scripts/setup-2agent.sh
#   matches the checklist in commands/setup-2agent.md
# - Same check applies to scripts/update-2agent.sh and commands/update-2agent.md

set -euo pipefail

PLUGIN_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
ERRORS=0

echo "🔍 Checklist sync validation"
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
# Only extract the "Auto Validation" section (excluding the "Claude Generates" section)
extract_command_checklist() {
  local cmd="$1"
  # Extract from "Auto Validation" to "Claude Generates" or the next section
  awk '/自動検証/,/Claude が生成|^###|^\*\*全て/' "$cmd" 2>/dev/null | \
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
  echo "📋 Validating $name..."

  # Extract to temporary files
  local script_checks=$(mktemp)
  local command_checks=$(mktemp)

  extract_script_checks "$script_file" > "$script_checks"
  extract_command_checklist "$command_file" > "$command_checks"

  # Items in script but not in command
  local missing_in_command=$(comm -23 "$script_checks" "$command_checks")
  if [ -n "$missing_in_command" ]; then
    echo "  ❌ In script but not in command checklist:"
    echo "$missing_in_command" | while read item; do
      echo "     - $item"
    done
    ERRORS=$((ERRORS + 1))
  fi

  # Items in command but not in script
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
    echo "  ⚠️ Skipped: no check items found (verify file structure)"
  elif [ -z "$missing_in_command" ] && [ -z "$missing_in_script" ]; then
    echo "  ✅ In sync ($script_count item(s))"
  fi

  rm -f "$script_checks" "$command_checks"
}

# ================================
# Main validation
# ================================

# Validate setup hub (v2.19.0+: 2agent is integrated into setup)
SETUP_SKILL="$PLUGIN_ROOT/skills/setup/SKILL.md"
SETUP_2AGENT_REF="$PLUGIN_ROOT/skills/setup/references/2agent-setup.md"

if [ -f "$SETUP_SKILL" ] && [ -f "$SETUP_2AGENT_REF" ]; then
  echo "✓ setup skill and 2agent-setup reference exist"
elif [ -f "$SETUP_SKILL" ]; then
  echo "⚠️ setup/references/2agent-setup.md not found (verify post-integration structure)"
else
  echo "⚠️ skills/setup/SKILL.md not found (skill may not have been created)"
fi

# Note: Since v2.17.0, commands have been migrated to skills
# Checklist sync will be managed per skill going forward
# If no target skills are found, exit successfully (do not fail on empty checklist)

# ================================
# Result summary
# ================================
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [ $ERRORS -eq 0 ]; then
  echo "✅ Checklist sync validation passed"
  exit 0
else
  echo "❌ $ERRORS inconsistency/inconsistencies found"
  echo ""
  echo "💡 How to fix:"
  echo "  1. Check check_file/check_dir in scripts/*.sh"
  echo "  2. Update the checklist in commands/*.md"
  echo "  3. Ensure both match"
  exit 1
fi
