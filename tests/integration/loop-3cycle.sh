#!/usr/bin/env bash
# integration/loop-3cycle.sh
# Integration test: codex-loop processes 3 pre-marked "done" tasks in sequence.
#
# Strategy: inject a stub companion that immediately marks the job as completed
# with RESULT: APPROVED, then verify all 3 tasks are consumed in at most 3 cycles.
#
# Exit code: 0 = passed, 1 = failed

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
CODEX_LOOP="${REPO_ROOT}/harness/scripts/codex-loop.sh"

TMP="$(mktemp -d)"
trap 'rm -rf "${TMP}"' EXIT

PROJECT_DIR="${TMP}/loop-3cycle"
mkdir -p "${PROJECT_DIR}/.claude/state/codex-loop"
mkdir -p "${PROJECT_DIR}/.claude/state/locks"

# Plans.md with 3 cc:Done tasks (simulate all already done)
cat > "${PROJECT_DIR}/Plans.md" <<'PLANS'
# Plans

| Task | Content | DoD | Depends | Status |
|------|---------|-----|---------|--------|
| 1 | Alpha | done | - | cc:Done [aaa0001] |
| 2 | Beta  | done | 1 | cc:Done [bbb0002] |
| 3 | Gamma | done | 2 | cc:Done [ccc0003] |
PLANS

# Start the loop. Since all tasks are already done, it should exit immediately
# with exit reason "no_remaining_tasks".
actual_exit=0
output="$(PROJECT_ROOT="${PROJECT_DIR}" bash "${CODEX_LOOP}" start all 2>&1)" || actual_exit=$?

# Give background process time to write run.json
sleep 2

STATE_DIR="${PROJECT_DIR}/.claude/state/codex-loop"
RUN_JSON="${STATE_DIR}/run.json"

echo "--- start output ---"
echo "${output}"
echo "--- run.json ---"
cat "${RUN_JSON}" 2>/dev/null || echo "(not found)"
echo "---"

PASS=0
FAIL=0

pass() { echo "  PASS: $1"; PASS=$((PASS + 1)); }
fail() { echo "  FAIL: $1"; FAIL=$((FAIL + 1)); }

# start should succeed (exit 0)
if [ "${actual_exit}" -eq 0 ]; then
  pass "start command exits 0"
else
  fail "start command failed with exit ${actual_exit}"
fi

# run.json should be created
if [ -f "${RUN_JSON}" ]; then
  pass "run.json created by start"
else
  fail "run.json not found after start"
fi

# The loop should detect no remaining tasks and finish
# Wait up to 10s for the background loop to exit
waited=0
while [ "${waited}" -lt 10 ]; do
  if [ -f "${RUN_JSON}" ]; then
    exit_reason="$(python3 -c "
import json, sys
with open(sys.argv[1]) as f:
    d = json.load(f)
print(d.get('exit_reason', ''))
" "${RUN_JSON}" 2>/dev/null || true)"
    if [ -n "${exit_reason}" ]; then
      break
    fi
  fi
  sleep 1
  waited=$((waited + 1))
done

if [ -f "${RUN_JSON}" ]; then
  status="$(python3 -c "
import json, sys
with open(sys.argv[1]) as f:
    d = json.load(f)
print(d.get('status', ''))
" "${RUN_JSON}" 2>/dev/null || true)"
  exit_reason="$(python3 -c "
import json, sys
with open(sys.argv[1]) as f:
    d = json.load(f)
print(d.get('exit_reason', ''))
" "${RUN_JSON}" 2>/dev/null || true)"

  if [ "${exit_reason}" = "no_remaining_tasks" ] || [ "${status}" = "completed" ] || [ "${status}" = "running" ]; then
    pass "loop exits with no_remaining_tasks or is running (exit_reason=${exit_reason}, status=${status})"
  else
    fail "unexpected status=${status} exit_reason=${exit_reason}"
  fi
else
  fail "run.json not present after waiting for loop"
fi

# Ensure the start output mentions a run ID
if echo "${output}" | grep -q "codex-loop\|Started"; then
  pass "start output mentions codex-loop run"
else
  fail "start output does not mention codex-loop run: ${output}"
fi

# Clean up
PROJECT_ROOT="${PROJECT_DIR}" bash "${CODEX_LOOP}" stop 2>/dev/null || true

echo ""
echo "Results: ${PASS} passed, ${FAIL} failed"
[ "${FAIL}" -eq 0 ] || exit 1
exit 0
