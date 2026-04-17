#!/usr/bin/env bash
# integration/loop-max-cycles.sh
# Integration test: verify that when cycle_count >= max_cycles, the loop
# exits with reason "max_cycles" and does not start another cycle.
#
# Strategy: write a run.json with cycle_count == max_cycles already reached,
# then call the `run` subcommand (foreground) and verify it exits with
# a max_cycles finalization.
#
# Exit code: 0 = passed, 1 = failed

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
CODEX_LOOP="${REPO_ROOT}/harness/scripts/codex-loop.sh"

TMP="$(mktemp -d)"
trap 'rm -rf "${TMP}"' EXIT

PROJECT_DIR="${TMP}/max-cycles-test"
STATE_DIR="${PROJECT_DIR}/.claude/state/codex-loop"
LOCKS_DIR="${PROJECT_DIR}/.claude/state/locks"
LOCK_DIR="${LOCKS_DIR}/codex-loop.lock.d"

mkdir -p "${STATE_DIR}" "${LOCKS_DIR}"

PASS=0
FAIL=0

pass() { echo "  PASS: $1"; PASS=$((PASS + 1)); }
fail() { echo "  FAIL: $1"; FAIL=$((FAIL + 1)); }

# ── Setup: Plans.md ───────────────────────────────────────────────────────────
cat > "${PROJECT_DIR}/Plans.md" <<'PLANS'
# Plans

| Task | Content | DoD | Depends | Status |
|------|---------|-----|---------|--------|
| 1 | Task one | done | - | cc:TODO |
| 2 | Task two | done | - | cc:TODO |
PLANS

RUN_JSON="${STATE_DIR}/run.json"

# ── Test: start with --max-cycles 0 is handled ────────────────────────────────
echo "Test group: max_cycles boundary"

# Write a run.json that has already reached max_cycles
cat > "${RUN_JSON}" <<EOF
{
  "schema_version": "codex-loop-run.v1",
  "run_id": "codex-loop-maxtest-001",
  "selection": "all",
  "max_cycles": 3,
  "pacing": "worker",
  "delay_seconds": 270,
  "cycle_count": 3,
  "consultations": 0,
  "last_decision": null,
  "last_trigger": null,
  "last_model": null,
  "consulted_trigger_hashes": [],
  "task_consultations": {},
  "status": "running",
  "pid": null,
  "started_at": "2025-01-01T00:00:00Z",
  "updated_at": "2025-01-01T00:00:00Z",
  "project_root": "${PROJECT_DIR}",
  "plans_file": "${PROJECT_DIR}/Plans.md"
}
EOF

# Acquire lock so `run` doesn't refuse to start
mkdir -p "${LOCK_DIR}"
printf '%s\n' "$$" > "${LOCK_DIR}/pid"

# Call `run` directly (foreground); it should see cycle_count >= max_cycles and exit 0
actual_exit=0
timeout 10 bash "${CODEX_LOOP}" run --run-id "codex-loop-maxtest-001" 2>/dev/null || actual_exit=$?

# Exit 0 means finalize_run was called and exited cleanly
if [ "${actual_exit}" -eq 0 ]; then
  pass "run exits 0 when cycle_count >= max_cycles"
else
  fail "expected exit 0 when max_cycles reached; got ${actual_exit}"
fi

# Check the exit_reason in run.json
if [ -f "${RUN_JSON}" ]; then
  exit_reason="$(python3 -c "
import json, sys
with open(sys.argv[1]) as f:
    d = json.load(f)
print(d.get('exit_reason', ''))
" "${RUN_JSON}" 2>/dev/null || true)"
  final_status="$(python3 -c "
import json, sys
with open(sys.argv[1]) as f:
    d = json.load(f)
print(d.get('status', ''))
" "${RUN_JSON}" 2>/dev/null || true)"

  if [ "${exit_reason}" = "max_cycles" ]; then
    pass "exit_reason is max_cycles"
  else
    fail "expected exit_reason=max_cycles; got '${exit_reason}'"
  fi

  if [ "${final_status}" = "completed" ]; then
    pass "status is completed after max_cycles"
  else
    fail "expected status=completed; got '${final_status}'"
  fi

  # pid should be cleared
  final_pid="$(python3 -c "
import json, sys
with open(sys.argv[1]) as f:
    d = json.load(f)
print(d.get('pid'))
" "${RUN_JSON}" 2>/dev/null || true)"
  if [ "${final_pid}" = "None" ] || [ -z "${final_pid}" ]; then
    pass "pid is cleared after finalization"
  else
    fail "expected pid=null after finalization; got '${final_pid}'"
  fi
else
  fail "run.json not found after run command"
fi

# ── Test: start respects --max-cycles flag ────────────────────────────────────
echo "Test group: --max-cycles flag"

PROJECT_B="${TMP}/proj-b"
mkdir -p "${PROJECT_B}"
cp "${PROJECT_DIR}/Plans.md" "${PROJECT_B}/Plans.md"

actual_exit=0
output="$(PROJECT_ROOT="${PROJECT_B}" bash "${CODEX_LOOP}" start all --max-cycles 5 2>&1)" || actual_exit=$?

if [ "${actual_exit}" -eq 0 ]; then
  pass "start with --max-cycles 5 exits 0"
else
  fail "start with --max-cycles failed; exit ${actual_exit}"
fi

sleep 1
RUN_JSON_B="${PROJECT_B}/.claude/state/codex-loop/run.json"
if [ -f "${RUN_JSON_B}" ]; then
  stored_max="$(python3 -c "
import json, sys
with open(sys.argv[1]) as f:
    d = json.load(f)
print(d.get('max_cycles', 0))
" "${RUN_JSON_B}" 2>/dev/null || true)"
  if [ "${stored_max}" = "5" ]; then
    pass "max_cycles=5 stored in run.json"
  else
    fail "expected max_cycles=5 in run.json; got '${stored_max}'"
  fi
else
  fail "run.json not created after start with --max-cycles"
fi

PROJECT_ROOT="${PROJECT_B}" bash "${CODEX_LOOP}" stop 2>/dev/null || true

echo ""
echo "Results: ${PASS} passed, ${FAIL} failed"
[ "${FAIL}" -eq 0 ] || exit 1
exit 0
