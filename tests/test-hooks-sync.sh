#!/bin/bash
# test-hooks-sync.sh
# Validates hooks.json structure and content (Phase 52+: source at harness/hooks/hooks.json)

set -euo pipefail
export TMPDIR=/tmp  # Force /tmp for sandboxed execution (sandbox blocks /var/folders)

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Test result counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
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

HOOKS_FILE="$PROJECT_ROOT/harness/hooks/hooks.json"

# ==================================================
# Test 1: Does harness/hooks/hooks.json exist?
# ==================================================
test_hooks_file_exists() {
  if [ ! -f "$HOOKS_FILE" ]; then
    echo "    Error: harness/hooks/hooks.json not found"
    return 1
  fi
  return 0
}

# ==================================================
# Test 2: Is the JSON valid?
# ==================================================
test_valid_json() {
  if ! jq empty "$HOOKS_FILE" 2>/dev/null; then
    echo "    Error: harness/hooks/hooks.json is not valid JSON"
    return 1
  fi
  return 0
}

# ==================================================
# Test 3: Do required hook events exist?
# ==================================================
test_required_hook_events() {
  local required_events=("PreToolUse" "SessionStart" "Stop" "PostToolUse")
  local missing=""

  for event in "${required_events[@]}"; do
    if ! jq -e ".hooks.$event" "$HOOKS_FILE" > /dev/null 2>&1; then
      missing="${missing}$event, "
    fi
  done

  if [ -n "$missing" ]; then
    echo "    Error: Missing required events: ${missing%, }"
    return 1
  fi
  return 0
}

# ==================================================
# Test 4: Are there no forbidden patterns (improper use of type: "prompt")?
# ==================================================
test_no_forbidden_prompt_usage() {
  local forbidden_events=("PreToolUse" "PostToolUse" "UserPromptSubmit")
  local violations=""

  for event in "${forbidden_events[@]}"; do
    if jq -e ".hooks.$event[]?.hooks[]? | select(.type == \"prompt\")" "$HOOKS_FILE" > /dev/null 2>&1; then
      violations="${violations}$event, "
    fi
  done

  if [ -n "$violations" ]; then
    echo "    Error: type: prompt should not be used in: ${violations%, }"
    echo "    (Stop and SubagentStop are the only valid events for prompt type)"
    return 1
  fi
  return 0
}

# ==================================================
# Test 5: Does marketplace.json point to harness/?
# ==================================================
test_marketplace_manifest() {
  local marketplace_file="$PROJECT_ROOT/.claude-plugin/marketplace.json"
  if [ ! -f "$marketplace_file" ]; then
    echo "    Error: .claude-plugin/marketplace.json not found"
    return 1
  fi
  if ! jq -e '.plugins[] | select(.source == "./harness/")' "$marketplace_file" > /dev/null 2>&1; then
    echo "    Error: marketplace.json should have a plugin with source: ./harness/"
    return 1
  fi
  return 0
}

# ==================================================
# Main execution
# ==================================================
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " Hooks sync tests"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

if ! command -v jq &> /dev/null; then
  echo -e "${RED}Error: jq is required but not installed${NC}"
  exit 1
fi

run_test "harness/hooks/hooks.json exists" test_hooks_file_exists
run_test "JSON is valid" test_valid_json
run_test "Required hook events exist" test_required_hook_events
run_test "No forbidden prompt usage" test_no_forbidden_prompt_usage
run_test "marketplace.json points to harness/" test_marketplace_manifest

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " Test results: $TESTS_PASSED/$TESTS_RUN passed"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

if [ "$TESTS_FAILED" -gt 0 ]; then
  exit 1
fi
exit 0
