#!/bin/bash
# test-update-notification.sh
# Validation tests for the update notification feature for existing users
#
# Test targets:
# - session-init.sh new rule detection
# - session-init.sh old hook configuration detection
# - template-tracker.sh needsInstall reporting
# - harness-update.md hooks detection logic

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PLUGIN_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Counters
PASSED=0
FAILED=0
TOTAL=0

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Test function
assert_file_contains() {
  local file="$1"
  local pattern="$2"
  local description="$3"
  TOTAL=$((TOTAL + 1))

  if [ ! -f "$PLUGIN_ROOT/$file" ]; then
    echo -e "${RED}✗${NC} $description (file not found)"
    FAILED=$((FAILED + 1))
    return 1
  fi

  if grep -qE "$pattern" "$PLUGIN_ROOT/$file"; then
    echo -e "${GREEN}✓${NC} $description"
    PASSED=$((PASSED + 1))
    return 0
  else
    echo -e "${RED}✗${NC} $description"
    echo "  Expected pattern: $pattern"
    echo "  File: $file"
    FAILED=$((FAILED + 1))
    return 1
  fi
}

assert_script_runs() {
  local script="$1"
  local description="$2"
  TOTAL=$((TOTAL + 1))

  if [ ! -f "$PLUGIN_ROOT/$script" ]; then
    echo -e "${RED}✗${NC} $description (script not found)"
    FAILED=$((FAILED + 1))
    return 1
  fi

  if bash -n "$PLUGIN_ROOT/$script" 2>/dev/null; then
    echo -e "${GREEN}✓${NC} $description"
    PASSED=$((PASSED + 1))
    return 0
  else
    echo -e "${RED}✗${NC} $description (syntax error)"
    FAILED=$((FAILED + 1))
    return 1
  fi
}

echo "=================================================="
echo "Update notification feature validation for existing users"
echo "=================================================="
echo ""

# ============================================
# session-init.sh validation
# ============================================
echo "## session-init.sh"
echo ""

assert_script_runs \
  "scripts/session-init.sh" \
  "session-init.sh syntax is correct"

assert_file_contains \
  "scripts/session-init.sh" \
  "QUALITY_RULES.*test-quality.md.*implementation-quality.md" \
  "quality protection rules check logic exists"

assert_file_contains \
  "scripts/session-init.sh" \
  "MISSING_RULES_INFO" \
  "missing rules notification variable exists"

assert_file_contains \
  "scripts/session-init.sh" \
  "OLD_HOOKS_INFO" \
  "old hook configuration detection variable exists"

assert_file_contains \
  "scripts/session-init.sh" \
  "jq.*\.hooks" \
  "hooks section detection logic exists"

assert_file_contains \
  "scripts/session-init.sh" \
  "INSTALLS_COUNT" \
  "new install count processing exists"

echo ""

# ============================================
# template-tracker.sh validation
# ============================================
echo "## template-tracker.sh"
echo ""

assert_script_runs \
  "scripts/template-tracker.sh" \
  "template-tracker.sh syntax is correct"

assert_file_contains \
  "scripts/template-tracker.sh" \
  "installs_details" \
  "install details tracking variable exists"

assert_file_contains \
  "scripts/template-tracker.sh" \
  "installsCount" \
  "installsCount output exists"

assert_file_contains \
  "scripts/template-tracker.sh" \
  "installs_count" \
  "install count tracking exists"

echo ""

# ============================================
# harness-update.md validation
# ============================================
echo "## harness-update.md"
echo ""

# harness-update skill validation (after v2.17.0+ skill migration)
assert_file_contains \
  "skills/harness-update/SKILL.md" \
  "hook|Hook|plugin" \
  "harness-update has hook-related description"

assert_file_contains \
  "skills/harness-update/SKILL.md" \
  "Breaking Changes|breaking-changes|deprecated" \
  "harness-update has breaking change detection"

assert_file_contains \
  "skills/harness-update/SKILL.md" \
  "backup|Backup" \
  "harness-update has backup functionality"

assert_file_contains \
  "skills/harness-update/SKILL.md" \
  "verification|Verification|検証" \
  "harness-update has verification functionality"

echo ""

# ============================================
# Results summary
# ============================================
echo "=================================================="
echo "Test Results"
echo "=================================================="
echo ""
echo "Total: $TOTAL"
echo -e "Passed: ${GREEN}$PASSED${NC}"
echo -e "Failed: ${RED}$FAILED${NC}"
echo ""

if [ $FAILED -eq 0 ]; then
  echo -e "${GREEN}✓ All tests passed${NC}"
  exit 0
else
  echo -e "${RED}✗ $FAILED test(s) failed${NC}"
  exit 1
fi
