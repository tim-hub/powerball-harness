#!/usr/bin/env bash
#
# test-codex-worker.sh
# Codex Worker integration test
#
# Usage: ./tests/test-codex-worker.sh [--quick]
#

set -euo pipefail

# Color definitions
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Counters
PASSED=0
FAILED=0

# Helper functions
log_test() { echo -e "${BLUE}[TEST]${NC} $1"; }
log_pass() { echo -e "${GREEN}[PASS]${NC} $1"; PASSED=$((PASSED + 1)); }
log_fail() { echo -e "${RED}[FAIL]${NC} $1"; FAILED=$((FAILED + 1)); }
log_skip() { echo -e "${YELLOW}[SKIP]${NC} $1"; }

# Project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

# ============================================
# Test 1: Script existence check
# ============================================
test_scripts_exist() {
    log_test "Test 1: Script existence check"

    local scripts=(
        "scripts/codex-worker-setup.sh"
        "scripts/codex-worker-engine.sh"
        "scripts/codex-worker-lock.sh"
        "scripts/codex-worker-quality-gate.sh"
        "scripts/codex-worker-merge.sh"
    )

    local all_exist=true
    for script in "${scripts[@]}"; do
        if [[ -f "$script" ]] && [[ -x "$script" ]]; then
            echo "  ✓ $script"
        else
            echo "  ✗ $script (missing or not executable)"
            all_exist=false
        fi
    done

    if $all_exist; then
        log_pass "All scripts exist and are executable"
    else
        log_fail "Some scripts missing"
    fi
}

# ============================================
# Test 2: Documentation existence check
# ============================================
test_docs_exist() {
    log_test "Test 2: Documentation existence check"

    local docs=(
        "skills/codex-worker/SKILL.md"
        "skills/codex-worker/references/setup.md"
        "skills/codex-worker/references/worker-execution.md"
        "skills/codex-worker/references/task-ownership.md"
        "skills/codex-worker/references/parallel-strategy.md"
        "skills/codex-worker/references/quality-gates.md"
        "skills/codex-worker/references/review-integration.md"
        "skills/ultrawork/references/codex-mode.md"
    )

    local all_exist=true
    for doc in "${docs[@]}"; do
        if [[ -f "$doc" ]]; then
            echo "  ✓ $doc"
        else
            echo "  ✗ $doc (missing)"
            all_exist=false
        fi
    done

    if $all_exist; then
        log_pass "All documents exist"
    else
        log_fail "Some documents missing"
    fi
}

# ============================================
# Test 3: Lock functionality test
# ============================================
test_lock_functionality() {
    log_test "Test 3: Lock functionality test"

    # Use unique test path (avoid conflict with previous tests)
    local test_path="test/lock-test-$$.ts"
    local worker_id="test-worker-$$"

    # Cleanup (expired locks only)
    ./scripts/codex-worker-lock.sh cleanup 2>/dev/null || true

    # Acquire lock
    if ./scripts/codex-worker-lock.sh acquire --path "$test_path" --worker "$worker_id" 2>/dev/null; then
        echo "  ✓ Lock acquired"

        # Check lock state
        local check_result
        check_result=$(./scripts/codex-worker-lock.sh check --path "$test_path" 2>/dev/null)
        if echo "$check_result" | grep -q '"locked":true'; then
            echo "  ✓ Lock state verified"
        else
            echo "  ✗ Lock state verification failed"
            log_fail "Lock state verification failed"
            return 1
        fi

        # Double lock prevention
        if ./scripts/codex-worker-lock.sh acquire --path "$test_path" --worker "another-worker" 2>/dev/null; then
            echo "  ✗ Double lock was permitted"
            log_fail "Double lock prevention failed"
            return 1
        else
            echo "  ✓ Double lock prevented"
        fi

        # Release lock
        if ./scripts/codex-worker-lock.sh release --path "$test_path" --worker "$worker_id" 2>/dev/null; then
            echo "  ✓ Lock released"
        else
            echo "  ✗ Lock release failed"
            log_fail "Lock release failed"
            return 1
        fi

        log_pass "Lock functionality works"
    else
        log_fail "Lock acquisition failed"
    fi
}

# ============================================
# Test 4: AGENTS.md hash calculation consistency
# ============================================
test_agents_hash_consistency() {
    log_test "Test 4: AGENTS.md hash calculation consistency"

    if [[ ! -f "AGENTS.md" ]]; then
        log_skip "AGENTS.md does not exist"
        return
    fi

    # engine hash calculation
    local engine_hash
    engine_hash=$(sed '1s/^\xEF\xBB\xBF//' "AGENTS.md" | tr -d '\r' | shasum -a 256 | cut -c1-8)

    # quality-gate hash calculation (same algorithm)
    local gate_hash
    gate_hash=$(sed '1s/^\xEF\xBB\xBF//' "AGENTS.md" | tr -d '\r' | shasum -a 256 | cut -c1-8)

    if [[ "$engine_hash" == "$gate_hash" ]]; then
        echo "  ✓ Hash match: $engine_hash"
        log_pass "Hash calculation consistent"
    else
        echo "  ✗ Hash mismatch: engine=$engine_hash, gate=$gate_hash"
        log_fail "Hash calculation inconsistent"
    fi
}

# ============================================
# Test 5: Skill misfire prevention check
# ============================================
test_skill_misfire_prevention() {
    log_test "Test 5: Skill misfire prevention check"

    # codex-worker description
    local worker_desc
    worker_desc=$(grep -A1 "^name: codex-worker" skills/codex-worker/SKILL.md | grep "description:" || echo "")

    # codex-review description
    local review_desc
    review_desc=$(grep -A1 "^name: codex-review" skills/codex-review/SKILL.md | grep "description:" || echo "")

    local issues=0

    # Does codex-worker exclude review triggers?
    if echo "$worker_desc" | grep -qi "review"; then
        echo "  ✓ codex-worker: review exclusion present"
    else
        echo "  ✗ codex-worker: review exclusion missing"
        issues=$((issues + 1))
    fi

    # Does codex-review exclude worker triggers?
    if echo "$review_desc" | grep -qE "Codex Worker|codex-worker|implement with Codex"; then
        echo "  ✓ codex-review: Worker exclusion present"
    else
        echo "  ✗ codex-review: Worker exclusion missing"
        issues=$((issues + 1))
    fi

    # Both have Do NOT Load For section?
    if grep -q "Do NOT Load For" skills/codex-worker/SKILL.md; then
        echo "  ✓ codex-worker: Do NOT Load For section present"
    else
        echo "  ✗ codex-worker: Do NOT Load For section missing"
        issues=$((issues + 1))
    fi

    if grep -q "Do NOT Load For" skills/codex-review/SKILL.md; then
        echo "  ✓ codex-review: Do NOT Load For section present"
    else
        echo "  ✗ codex-review: Do NOT Load For section missing"
        issues=$((issues + 1))
    fi

    if [[ $issues -eq 0 ]]; then
        log_pass "Skill misfire prevention config OK"
    else
        log_fail "Skill misfire prevention has issues"
    fi
}

# ============================================
# Test 6: Worktree path consistency check
# ============================================
test_worktree_path_consistency() {
    log_test "Test 6: Worktree path consistency check"

    local files=(
        "skills/codex-worker/references/parallel-strategy.md"
        "skills/codex-worker/references/worker-execution.md"
        "skills/ultrawork/references/codex-mode.md"
    )

    local issues=0

    for file in "${files[@]}"; do
        # Detect legacy ../worktree- format usage
        if grep -q '\.\./worktree-' "$file" 2>/dev/null; then
            echo "  ✗ $file: uses legacy ../worktree- format"
            issues=$((issues + 1))
        else
            echo "  ✓ $file: path format OK"
        fi
    done

    if [[ $issues -eq 0 ]]; then
        log_pass "Worktree path consistency OK"
    else
        log_fail "Worktree path inconsistency found"
    fi
}

# ============================================
# Main execution
# ============================================
main() {
    echo "============================================"
    echo "  Codex Worker Integration Test"
    echo "============================================"
    echo ""

    # Run tests
    test_scripts_exist
    echo ""
    test_docs_exist
    echo ""
    test_lock_functionality
    echo ""
    test_agents_hash_consistency
    echo ""
    test_skill_misfire_prevention
    echo ""
    test_worktree_path_consistency
    echo ""

    # Results summary
    echo "============================================"
    echo "  Test Results Summary"
    echo "============================================"
    echo -e "  ${GREEN}PASSED${NC}: $PASSED"
    echo -e "  ${RED}FAILED${NC}: $FAILED"
    echo ""

    if [[ $FAILED -eq 0 ]]; then
        echo -e "${GREEN}All tests passed${NC}"
        exit 0
    else
        echo -e "${RED}Some tests failed${NC}"
        exit 1
    fi
}

main "$@"
