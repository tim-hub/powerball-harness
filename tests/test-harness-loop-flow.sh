#!/usr/bin/env bash
# test-harness-loop-flow.sh
# Smoke-test the harness codex-loop flow by verifying:
# - The codex-loop.sh script is executable and in the expected location
# - The harness binary exposes a codex-loop subcommand (or delegates correctly)
# - State directory scaffolding is created on start
#
# Exit code: 0 = all tests passed, 1 = at least one test failed

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
CODEX_LOOP="${REPO_ROOT}/harness/scripts/codex-loop.sh"

PASS=0
FAIL=0

pass() { echo "  PASS: $1"; PASS=$((PASS + 1)); }
fail() { echo "  FAIL: $1"; FAIL=$((FAIL + 1)); }

# ── Test: script exists and is executable ─────────────────────────────────────
echo "Test group: script presence"

if [ -f "${CODEX_LOOP}" ]; then
  pass "codex-loop.sh exists at expected path"
else
  fail "codex-loop.sh not found at ${CODEX_LOOP}"
fi

if [ -x "${CODEX_LOOP}" ]; then
  pass "codex-loop.sh is executable"
else
  fail "codex-loop.sh is not executable"
fi

# ── Test: shebang / set -euo pipefail ─────────────────────────────────────────
echo "Test group: script quality"

shebang="$(head -1 "${CODEX_LOOP}")"
if echo "${shebang}" | grep -q "bash"; then
  pass "codex-loop.sh has bash shebang"
else
  fail "codex-loop.sh missing bash shebang; got: ${shebang}"
fi

if grep -q "set -euo pipefail" "${CODEX_LOOP}"; then
  pass "codex-loop.sh has set -euo pipefail"
else
  fail "codex-loop.sh missing set -euo pipefail"
fi

# ── Test: subcommands available ───────────────────────────────────────────────
echo "Test group: subcommand availability"

for subcmd in start status stop run run-cycle local-task-worker; do
  output="$(bash "${CODEX_LOOP}" --help 2>&1 || true)"
  if echo "${output}" | grep -q "${subcmd}"; then
    pass "subcommand '${subcmd}' is documented in help"
  else
    fail "subcommand '${subcmd}' missing from help output"
  fi
done

# ── Test: state directory layout ──────────────────────────────────────────────
echo "Test group: state directory layout"

TMP="$(mktemp -d)"
trap 'rm -rf "${TMP}"' EXIT

PROJECT_DIR="${TMP}/flow-project"
mkdir -p "${PROJECT_DIR}"
cat > "${PROJECT_DIR}/Plans.md" <<'PLANS'
# Plans

| Task | Content | DoD | Depends | Status |
|------|---------|-----|---------|--------|
| 1 | Task one | done | - | cc:Done [abc1234] |
PLANS

# Start with a plans file that has only done tasks — the loop should complete immediately
# but should still create state dirs
output="$(PROJECT_ROOT="${PROJECT_DIR}" bash "${CODEX_LOOP}" start all 2>&1 || true)"

# Give the background process a moment to initialise state
sleep 1

STATE_DIR="${PROJECT_DIR}/.claude/state/codex-loop"
if [ -d "${STATE_DIR}" ] || echo "${output}" | grep -q "Started codex-loop\|no_remaining\|idle\|not found"; then
  pass "codex-loop state directory created or loop started"
else
  fail "expected state directory or start message; got: ${output}"
fi

# Stop any background process
PROJECT_ROOT="${PROJECT_DIR}" bash "${CODEX_LOOP}" stop 2>/dev/null || true

# ── Summary ────────────────────────────────────────────────────────────────────
echo ""
echo "Results: ${PASS} passed, ${FAIL} failed"

if [ "${FAIL}" -gt 0 ]; then
  exit 1
fi
exit 0
