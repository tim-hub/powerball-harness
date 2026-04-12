#!/bin/bash
# diagnose-and-fix.sh
# Script to diagnose CI errors and suggest or auto-fix them
#
# Usage:
#   ./scripts/ci/diagnose-and-fix.sh          # Diagnose only
#   ./scripts/ci/diagnose-and-fix.sh --fix    # Also run auto-fix
#
# This script is executed by Claude on CI failure to obtain fix suggestions.

set -euo pipefail

PLUGIN_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$PLUGIN_ROOT"

AUTO_FIX=false
if [ "$1" = "--fix" ]; then
  AUTO_FIX=true
fi

ISSUES_FOUND=0
FIXES_APPLIED=0

echo "🔧 CI Diagnosis & Fix Tool"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# ================================
# 1. Version sync check
# ================================
check_version_sync() {
  echo "📋 [1/5] Version sync check..."

  local file_version=$(cat VERSION 2>/dev/null | tr -d '[:space:]')
  local json_version=$(grep '"version"' .claude-plugin/plugin.json | head -1 | sed 's/.*: *"\([^"]*\)".*/\1/')

  if [ "$file_version" != "$json_version" ]; then
    echo "  ❌ VERSION ($file_version) and plugin.json ($json_version) mismatch"
    ISSUES_FOUND=$((ISSUES_FOUND + 1))

    if [ "$AUTO_FIX" = true ]; then
      echo "  🔧 Fixing: Updating plugin.json to $file_version..."
      sed -i.bak "s/\"version\": \"$json_version\"/\"version\": \"$file_version\"/" .claude-plugin/plugin.json
      rm -f .claude-plugin/plugin.json.bak
      FIXES_APPLIED=$((FIXES_APPLIED + 1))
      echo "  ✅ Fix applied"
    else
      echo "  💡 Fix suggestion: Change plugin.json version to \"$file_version\""
    fi
  else
    echo "  ✅ In sync (v$file_version)"
  fi
}

# ================================
# 2. Checklist sync check
# ================================
check_checklist_sync() {
  echo ""
  echo "📋 [2/5] Checklist sync check..."

  if ./scripts/ci/check-checklist-sync.sh >/dev/null 2>&1; then
    echo "  ✅ In sync"
  else
    echo "  ❌ Script and command checklists mismatch"
    ISSUES_FOUND=$((ISSUES_FOUND + 1))

    echo "  💡 Fix suggestion:"
    echo "     1. Check check_file/check_dir in scripts/*.sh"
    echo "     2. Manually update checklists in commands/*.md"
    echo "     (Auto-fix not supported - manual verification required)"
  fi
}

# ================================
# 3. Template existence check
# ================================
check_templates() {
  echo ""
  echo "📋 [3/5] Template existence check..."

  local missing=()
  local templates=(
    "templates/AGENTS.md.template"
    "templates/CLAUDE.md.template"
    "templates/Plans.md.template"
    "templates/.claude-code-harness-version.template"
    "templates/cursor/commands/start-session.md"
    "templates/cursor/commands/handoff-to-claude.md"
    "templates/cursor/commands/review-cc-work.md"
    "templates/rules/workflow.md.template"
    "templates/rules/coding-standards.md.template"
  )

  for t in "${templates[@]}"; do
    if [ ! -f "$t" ]; then
      missing+=("$t")
    fi
  done

  if [ ${#missing[@]} -eq 0 ]; then
    echo "  ✅ All templates exist"
  else
    echo "  ❌ Missing templates:"
    for m in "${missing[@]}"; do
      echo "     - $m"
    done
    ISSUES_FOUND=$((ISSUES_FOUND + 1))
    echo "  💡 Fix suggestion: Create missing files"
  fi
}

# ================================
# 4. Hooks consistency check
# ================================
check_hooks() {
  echo ""
  echo "📋 [4/5] Hooks consistency check..."

  if ! jq empty hooks/hooks.json 2>/dev/null; then
    echo "  ❌ hooks.json is invalid JSON"
    ISSUES_FOUND=$((ISSUES_FOUND + 1))
    echo "  💡 Fix suggestion: Check JSON syntax of hooks/hooks.json"
    return
  fi

  local missing_scripts=()
  local script_refs=$(grep -oE 'scripts/[a-zA-Z0-9_.-]+' hooks/hooks.json 2>/dev/null || true)

  for ref in $script_refs; do
    if [ ! -f "$ref" ]; then
      missing_scripts+=("$ref")
    fi
  done

  if [ ${#missing_scripts[@]} -eq 0 ]; then
    echo "  ✅ Hooks config OK"
  else
    echo "  ❌ Missing referenced scripts:"
    for s in "${missing_scripts[@]}"; do
      echo "     - $s"
    done
    ISSUES_FOUND=$((ISSUES_FOUND + 1))
    echo "  💡 Fix suggestion: Create missing scripts or remove references from hooks.json"
  fi
}

# ================================
# 5. Release metadata check
# ================================
check_version_bump() {
  echo ""
  echo "📋 [5/5] Release metadata check..."

  local check_log
  check_log="$(mktemp)"

  if bash ./scripts/ci/check-version-bump.sh >"$check_log" 2>&1; then
    sed 's/^/  /' "$check_log"
    rm -f "$check_log"
    return
  fi

  sed 's/^/  /' "$check_log"
  rm -f "$check_log"

  if [ "$AUTO_FIX" = true ] && ! bash ./scripts/sync-version.sh check >/dev/null 2>&1; then
    echo "  🔧 Fixing: Syncing plugin.json to VERSION..."
    bash ./scripts/sync-version.sh sync
    FIXES_APPLIED=$((FIXES_APPLIED + 1))

    if bash ./scripts/ci/check-version-bump.sh >/dev/null 2>&1; then
      echo "  ✅ Release metadata consistency restored via plugin.json sync"
      return
    fi
  fi

  ISSUES_FOUND=$((ISSUES_FOUND + 1))
  echo "  💡 Fix strategy:"
  echo "     - Do not change VERSION in normal PRs"
  echo "     - Update VERSION / plugin.json / CHANGELOG release entry together only when releasing"
}

# ================================
# Main execution
# ================================

check_version_sync
check_checklist_sync
check_templates
check_hooks
check_version_bump

# ================================
# Result summary
# ================================

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [ $ISSUES_FOUND -eq 0 ]; then
  echo "✅ No issues found"
  exit 0
fi

echo "📊 Result summary:"
echo "  - Issues found: $ISSUES_FOUND"

if [ "$AUTO_FIX" = true ]; then
  echo "  - Auto-fixes applied: $FIXES_APPLIED"
  if [ $FIXES_APPLIED -gt 0 ]; then
    echo ""
    echo "💡 Next steps:"
    echo "  1. Review fixes: git diff"
    echo "  2. Update CHANGELOG.md"
    echo "  3. Commit & push"
  fi
else
  echo ""
  echo "💡 To run auto-fix:"
  echo "  ./scripts/ci/diagnose-and-fix.sh --fix"
fi

exit 1
