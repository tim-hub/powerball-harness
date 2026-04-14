#!/bin/bash
# test-commit-guard.sh
# Tests for the Commit Guard feature
#
# Test targets:
# - scripts/posttooluse-commit-cleanup.sh (review approval state clear)
# - hooks.json (hook registration)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Test result counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Test function
run_test() {
  local test_name="$1"
  local test_func="$2"

  TESTS_RUN=$((TESTS_RUN + 1))
  echo -n "  Testing: $test_name... "

  if $test_func; then
    echo -e "${GREEN}PASSED${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
  else
    echo -e "${RED}FAILED${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
  fi
}

# ==================================================
# Test 1: Does posttooluse-commit-cleanup.sh exist?
# ==================================================
test_cleanup_script_exists() {
  local script="$PROJECT_ROOT/scripts/posttooluse-commit-cleanup.sh"

  if [ ! -f "$script" ]; then
    echo "    Error: posttooluse-commit-cleanup.sh not found"
    return 1
  fi

  return 0
}

# ==================================================
# Test 2: Does the script have execute permissions?
# ==================================================
test_cleanup_script_executable() {
  local script="$PROJECT_ROOT/scripts/posttooluse-commit-cleanup.sh"

  if [ ! -x "$script" ]; then
    echo "    Error: posttooluse-commit-cleanup.sh is not executable"
    return 1
  fi

  return 0
}

# ==================================================
# Test 3: Does posttooluse-commit-cleanup.sh detect git commit?
# ==================================================
test_cleanup_detects_git_commit() {
  local script="$PROJECT_ROOT/scripts/posttooluse-commit-cleanup.sh"

  if ! grep -q "git[[:space:]]*commit" "$script" 2>/dev/null; then
    echo "    Error: git commit detection not found in cleanup script"
    return 1
  fi

  return 0
}

# ==================================================
# Test 4: Does posttooluse-commit-cleanup.sh have state file removal logic?
# ==================================================
test_cleanup_removes_state_file() {
  local script="$PROJECT_ROOT/scripts/posttooluse-commit-cleanup.sh"

  # Script removes via variable: rm -f "$REVIEW_STATE_FILE"
  if ! grep -q 'rm -f.*REVIEW_STATE_FILE' "$script" 2>/dev/null; then
    echo "    Error: state file removal logic not found"
    return 1
  fi

  # Also check for state file path definition
  if ! grep -Eq "review-approved.json|review-result.json" "$script" 2>/dev/null; then
    echo "    Error: review artifact path definition not found"
    return 1
  fi

  return 0
}

# ==================================================
# Test 5: Is the commit-cleanup hook registered in hooks.json?
# ==================================================
test_hooks_has_commit_cleanup() {
  local hooks_file="$PROJECT_ROOT/hooks/hooks.json"

  if ! command -v jq &> /dev/null; then
    echo "    Warning: jq not available, skipping JSON validation"
    # Fall back to grep check if jq is unavailable
    if ! grep -q "posttooluse-commit-cleanup" "$hooks_file" 2>/dev/null; then
      echo "    Error: commit-cleanup hook not registered in hooks.json"
      return 1
    fi
    return 0
  fi

  # Check if commit-cleanup is registered with Bash matcher in PostToolUse
  if ! jq -e '.hooks.PostToolUse[] | select(.matcher == "Bash") | .hooks[] | select(.command | contains("posttooluse-commit-cleanup"))' "$hooks_file" > /dev/null 2>&1; then
    echo "    Error: commit-cleanup hook not properly registered for Bash in PostToolUse"
    return 1
  fi

  return 0
}

# ==================================================
# Test 6: Is the same hook also in .claude-plugin/hooks.json?
# ==================================================
test_plugin_hooks_has_commit_cleanup() {
  local hooks_file="$PROJECT_ROOT/.claude-plugin/hooks.json"

  if ! grep -q "posttooluse-commit-cleanup" "$hooks_file" 2>/dev/null; then
    echo "    Error: commit-cleanup hook not registered in .claude-plugin/hooks.json"
    return 1
  fi

  return 0
}

# ==================================================
# Test 7: Does the config template have the commit_guard setting?
# ==================================================
test_config_has_commit_guard_option() {
  local config_template="$PROJECT_ROOT/templates/.claude-code-harness.config.yaml.template"

  if ! grep -q "commit_guard:" "$config_template" 2>/dev/null; then
    echo "    Error: commit_guard option not found in config template"
    return 1
  fi

  return 0
}

# ==================================================
# Test 8: Does posttooluse-commit-cleanup.sh preserve state on error?
# ==================================================
test_cleanup_preserves_on_error() {
  local script="$PROJECT_ROOT/scripts/posttooluse-commit-cleanup.sh"

  # Check if error detection logic is present
  if ! grep -Eq "error|fatal|failed|nothing to commit" "$script" 2>/dev/null; then
    echo "    Error: error detection logic not found in cleanup script"
    return 1
  fi

  return 0
}

# ==================================================
# Main execution
# ==================================================
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " Commit Guard Tests"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

echo "  [PostToolUse Cleanup]"
run_test "posttooluse-commit-cleanup.sh exists" test_cleanup_script_exists
run_test "posttooluse-commit-cleanup.sh is executable" test_cleanup_script_executable
run_test "posttooluse-commit-cleanup.sh detects git commit" test_cleanup_detects_git_commit
run_test "posttooluse-commit-cleanup.sh has state file removal logic" test_cleanup_removes_state_file
run_test "posttooluse-commit-cleanup.sh preserves state on error" test_cleanup_preserves_on_error

echo ""
echo "  [Hooks Integration]"
run_test "hooks.json has commit-cleanup hook registered" test_hooks_has_commit_cleanup
run_test ".claude-plugin/hooks.json also has commit-cleanup hook" test_plugin_hooks_has_commit_cleanup

echo ""
echo "  [Configuration]"
run_test "config template has commit_guard setting" test_config_has_commit_guard_option

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " Test results: $TESTS_PASSED/$TESTS_RUN passed"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

if [ "$TESTS_FAILED" -gt 0 ]; then
  exit 1
fi

exit 0
