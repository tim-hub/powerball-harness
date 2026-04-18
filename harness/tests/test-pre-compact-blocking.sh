#\!/usr/bin/env bash
# test-pre-compact-blocking.sh — Tests role-based compaction blocking in pre-compact-save.js
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
HOOK="$SCRIPT_DIR/../scripts/hook-handlers/pre-compact-save.js"
PASS=0; FAIL=0

run_hook() {
  local role="${1:-}" wip="${2:-false}"
  local tmpdir
  tmpdir="$(mktemp -d "$TMPDIR/pre-compact-test-XXXXXX")"
  mkdir -p "$tmpdir/.claude/state"

  if [ "$wip" = "true" ]; then
    printf '| 1.1 | Fix the bug | Tests pass | - | cc:WIP |\n' > "$tmpdir/Plans.md"
  fi

  local output
  if [ -n "$role" ]; then
    output=$(cd "$tmpdir" && HARNESS_SESSION_ROLE="$role" node "$HOOK" 2>/dev/null || true)
  else
    output=$(cd "$tmpdir" && node "$HOOK" 2>/dev/null || true)
  fi
  rm -rf "$tmpdir"
  echo "$output"
}

assert_continue() {
  local label="$1" output="$2" expected="$3"
  local actual
  actual=$(printf '%s' "$output" | python3 -c "import json,sys; d=json.load(sys.stdin); print(d.get('continue','?'))" 2>/dev/null || echo "parse_error")
  if [ "$actual" = "$expected" ]; then
    echo "  PASS [$label]: continue=$actual"
    PASS=$((PASS+1))
  else
    echo "  FAIL [$label]: expected continue=$expected, got continue=$actual"
    printf '       output: %s\n' "$output"
    FAIL=$((FAIL+1))
  fi
}

echo "=== pre-compact-save.js role-based blocking (4 permutations) ==="

# 1. worker + WIP → BLOCK (continue: False in Python json → false in JSON)
out=$(run_hook "worker" "true")
assert_continue "worker + WIP → BLOCK" "$out" "False"

# 2. worker + no WIP → PASS (continue: True)
out=$(run_hook "worker" "false")
assert_continue "worker + no WIP → PASS" "$out" "True"

# 3. reviewer + WIP → PASS (continue: True)
out=$(run_hook "reviewer" "true")
assert_continue "reviewer + WIP → PASS" "$out" "True"

# 4. unknown role + WIP → PASS/warn (continue: True)
out=$(run_hook "" "true")
assert_continue "unknown role + WIP → PASS (warn)" "$out" "True"

echo ""
echo "Results: $PASS passed, $FAIL failed"
[ "$FAIL" -eq 0 ] && exit 0 || exit 1
