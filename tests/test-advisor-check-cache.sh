#!/usr/bin/env bash
# test-advisor-check-cache.sh — unit test for the cache-check helper
# (Phase 73.4). Verifies the advisor's cache-first ordering contract:
# the helper returns HIT (exit 0 + decision JSON) or MISS (exit 1) based
# on whether (task_id, reason_code, error_signature) matches any line in
# .claude/state/advisor/history.jsonl.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
CHECKER="$REPO_ROOT/harness/scripts/advisor-check-cache.sh"

[[ -f "$CHECKER" ]] || { echo "FAIL: checker not found at $CHECKER"; exit 1; }

TEST_DIR="$(mktemp -d "${TMPDIR:-/tmp}/advisor-cache-test.XXXXXX")"
trap 'rm -rf "$TEST_DIR"' EXIT
cd "$TEST_DIR"

git init -q
git config user.email t@example.com
git config user.name t
mkdir -p .claude/state/advisor

fail() { echo "FAIL: $*"; exit 1; }

# --- MISS: no history file at all ---
if bash "$CHECKER" --task 1.1 --reason repeated_failure --sig "compile error" 2>/dev/null; then
    fail "no history file should exit 1 (MISS)"
fi

# --- MISS: empty history file ---
touch .claude/state/advisor/history.jsonl
if bash "$CHECKER" --task 1.1 --reason repeated_failure --sig "compile error" 2>/dev/null; then
    fail "empty history file should MISS"
fi

# --- HIT: one matching entry ---
cat > .claude/state/advisor/history.jsonl <<'EOF'
{"task_id":"1.1","reason_code":"repeated_failure","error_signature":"compile error","decision":"CORRECTION","rationale":"simple fix","suggested_approach":"apply the patch"}
{"task_id":"2.1","reason_code":"high_risk_preflight","error_signature":"migration risk","decision":"STOP","rationale":"human needed","suggested_approach":null}
EOF
out=$(bash "$CHECKER" --task 1.1 --reason repeated_failure --sig "compile error")
line_count=$(printf '%s\n' "$out" | sed '/^$/d' | wc -l | tr -d ' ')
[[ "$line_count" == "1" ]] || fail "HIT should return exactly 1 line, got $line_count"
printf '%s' "$out" | grep -q '"decision":"CORRECTION"' \
    || fail "HIT output should contain the cached decision (CORRECTION)"
printf '%s' "$out" | grep -q '"suggested_approach":"apply the patch"' \
    || fail "HIT should preserve the full cached JSON including suggested_approach"

# --- MISS: different error_signature (partial-field match should still miss) ---
if bash "$CHECKER" --task 1.1 --reason repeated_failure --sig "OTHER ERROR" 2>/dev/null; then
    fail "different sig should MISS even with matching task/reason"
fi

# --- MISS: different reason_code ---
if bash "$CHECKER" --task 1.1 --reason plateau_before_escalation --sig "compile error" 2>/dev/null; then
    fail "different reason should MISS"
fi

# --- HIT: multiple matching entries returns the most recent (last) ---
cat >> .claude/state/advisor/history.jsonl <<'EOF'
{"task_id":"1.1","reason_code":"repeated_failure","error_signature":"compile error","decision":"PLAN","rationale":"new data changed the call","suggested_approach":"replan from scratch"}
EOF
out=$(bash "$CHECKER" --task 1.1 --reason repeated_failure --sig "compile error")
printf '%s' "$out" | grep -q '"decision":"PLAN"' \
    || fail "multiple matches should return most recent entry (expected PLAN)"

# --- Usage errors ---
if bash "$CHECKER" --task 1.1 --reason x 2>/dev/null; then
    fail "missing --sig should exit 2"
fi
if bash "$CHECKER" --reason x --sig y 2>/dev/null; then
    fail "missing --task should exit 2"
fi
if bash "$CHECKER" --unknown flag 2>/dev/null; then
    fail "unknown flag should exit 2"
fi

echo "PASS test-advisor-check-cache.sh"
