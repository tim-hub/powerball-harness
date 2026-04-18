#!/usr/bin/env bash
# test-advisor-integration.sh — end-to-end flow combining the Phase 73.2
# scoped loader and Phase 73.4 cache-check helper. Exercises the advisor's
# expected control flow on a repeated_failure with context_sources=[trace]:
#
#   1. Check cache via advisor-check-cache.sh  (expect MISS)
#   2. On MISS, load context via advisor-load-context.sh
#   3. Verify trace content supports the expected CORRECTION decision
#      pattern — per DoD, a "fix was a single-file rename" trace should
#      surface the rename approach so the advisor can recommend a local fix
#   4. Seed a decision in history.jsonl, re-check cache (expect HIT, with
#      the decision JSON unchanged)
#
# This is a shell-level integration test of the two helpers composing
# correctly; it does not invoke the Opus advisor LLM itself.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
LOADER="$REPO_ROOT/harness/scripts/advisor-load-context.sh"
CHECKER="$REPO_ROOT/harness/scripts/advisor-check-cache.sh"

TEST_DIR="$(mktemp -d "${TMPDIR:-/tmp}/advisor-integration-test.XXXXXX")"
trap 'rm -rf "$TEST_DIR"' EXIT
cd "$TEST_DIR"

git init -q
git config user.email t@example.com
git config user.name t
mkdir -p .claude/state/traces .claude/state/advisor

fail() { echo "FAIL: $*"; exit 1; }

# --- Fixture: a trace showing "fix was a single-file rename" (DoD example) ---
cat > .claude/state/traces/99.1.jsonl <<'EOF'
{"schema":"trace.v1","ts":"2026-04-18T10:00:00Z","task_id":"99.1","event_type":"task_start","agent":"worker","attempt_n":1,"payload":{"description":"rename legacy API"}}
{"schema":"trace.v1","ts":"2026-04-18T10:01:00Z","task_id":"99.1","event_type":"tool_call","agent":"worker","attempt_n":1,"payload":{"tool":"Bash","args_summary":"cmd=go build ./..."}}
{"schema":"trace.v1","ts":"2026-04-18T10:01:30Z","task_id":"99.1","event_type":"error","agent":"worker","attempt_n":1,"payload":{"error_signature":"undefined: oldfuncname","raw_error":"main.go: undefined: oldfuncname"}}
{"schema":"trace.v1","ts":"2026-04-18T10:02:00Z","task_id":"99.1","event_type":"fix_attempt","agent":"worker","attempt_n":2,"payload":{"prior_error_signature":"undefined: oldfuncname","approach":"rename oldfuncname to newfuncname in caller file"}}
{"schema":"trace.v1","ts":"2026-04-18T10:02:30Z","task_id":"99.1","event_type":"tool_call","agent":"worker","attempt_n":2,"payload":{"tool":"Edit","args_summary":"file_path=main.go"}}
{"schema":"trace.v1","ts":"2026-04-18T10:03:00Z","task_id":"99.1","event_type":"outcome","agent":"worker","attempt_n":2,"payload":{"status":"success","notes":"rename resolved the undefined symbol"}}
EOF
git add .
git commit -q -m "seed (99.1): single-file rename fixture"

# --- Step 1: cache check should MISS (no history yet) ---
if bash "$CHECKER" --task 99.1 --reason repeated_failure --sig "undefined: oldfuncname" 2>/dev/null; then
    fail "step 1 cache should MISS; history.jsonl not seeded"
fi

# --- Step 2: load context for [trace] ---
context=$(bash "$LOADER" --task 99.1 --sources trace)

# --- Step 3: verify the trace content enables a rename-aware CORRECTION ---
echo "$context" | grep -q '"event_type":"fix_attempt"' \
    || fail "loaded trace must include fix_attempt event"
echo "$context" | grep -q 'rename oldfuncname to newfuncname' \
    || fail "loaded trace must show the rename approach (per DoD)"
echo "$context" | grep -q '"event_type":"outcome"' \
    || fail "loaded trace must include the successful outcome event"

# --- Step 4: seed a decision, re-check cache, expect HIT with same JSON ---
cat > .claude/state/advisor/history.jsonl <<'EOF'
{"task_id":"99.1","reason_code":"repeated_failure","error_signature":"undefined: oldfuncname","decision":"CORRECTION","rationale":"trace shows single-file rename fix pattern","suggested_approach":"apply the rename to any remaining call sites"}
EOF
out=$(bash "$CHECKER" --task 99.1 --reason repeated_failure --sig "undefined: oldfuncname")
echo "$out" | grep -q '"decision":"CORRECTION"' || fail "cached decision should be CORRECTION"
echo "$out" | grep -q 'rename' || fail "cached rationale should mention rename"

# --- Step 5: a different error signature must MISS even now ---
if bash "$CHECKER" --task 99.1 --reason repeated_failure --sig "different error" 2>/dev/null; then
    fail "a different error_signature must MISS even with seeded history"
fi

echo "PASS test-advisor-integration.sh"
