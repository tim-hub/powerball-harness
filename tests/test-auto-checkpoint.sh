#!/usr/bin/env bash
# test-auto-checkpoint.sh
# Unit tests for harness/scripts/auto-checkpoint.sh
#
# Exit code: 0 = all tests passed, 1 = at least one test failed

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
CHECKPOINT_SCRIPT="${REPO_ROOT}/harness/scripts/auto-checkpoint.sh"

PASS=0
FAIL=0

pass() { echo "  PASS: $1"; PASS=$((PASS + 1)); }
fail() { echo "  FAIL: $1"; FAIL=$((FAIL + 1)); }

# ── Setup ──────────────────────────────────────────────────────────────────────
TMP="$(mktemp -d)"
trap 'rm -rf "${TMP}"' EXIT

# Temporary PROJECT_ROOT for isolation
PROJECT_DIR="${TMP}/project"
mkdir -p "${PROJECT_DIR}/.claude/state"

SPRINT_CONTRACT="${TMP}/sprint-contract.json"
REVIEW_RESULT="${TMP}/review-result.json"

printf '{"task":{"id":"42.1"},"checks":[]}' > "${SPRINT_CONTRACT}"
printf '{"schema_version":"review-result.v1","verdict":"APPROVE"}' > "${REVIEW_RESULT}"

# ── Test: --help flag ─────────────────────────────────────────────────────────
echo "Test group: --help flag"

output="$(bash "${CHECKPOINT_SCRIPT}" --help 2>&1 || true)"
if echo "${output}" | grep -q "task_id"; then
  pass "--help shows usage with task_id"
else
  fail "--help did not show usage; got: ${output}"
fi

# ── Test: missing arguments ────────────────────────────────────────────────────
echo "Test group: argument validation"

actual_exit=0
bash "${CHECKPOINT_SCRIPT}" 2>/dev/null || actual_exit=$?
if [ "${actual_exit}" -eq 1 ]; then
  pass "exits 1 with no arguments"
else
  fail "expected exit 1 with no arguments; got ${actual_exit}"
fi

actual_exit=0
bash "${CHECKPOINT_SCRIPT}" "42.1" "abc1234" 2>/dev/null || actual_exit=$?
if [ "${actual_exit}" -eq 1 ]; then
  pass "exits 1 with only 2 arguments"
else
  fail "expected exit 1 with 2 arguments; got ${actual_exit}"
fi

# ── Test: HARNESS_MEM_DISABLE=1 (skip API call, still writes audit record) ────
echo "Test group: HARNESS_MEM_DISABLE=1"

actual_exit=0
PROJECT_ROOT="${PROJECT_DIR}" \
HARNESS_MEM_DISABLE=1 \
  bash "${CHECKPOINT_SCRIPT}" "42.1" "abc1234" "${SPRINT_CONTRACT}" "${REVIEW_RESULT}" \
  2>/dev/null || actual_exit=$?

CHECKPOINT_EVENTS="${PROJECT_DIR}/.claude/state/checkpoint-events.jsonl"
SESSION_EVENTS="${PROJECT_DIR}/.claude/state/session-events.jsonl"

# With HARNESS_MEM_DISABLE=1, API fails → exit 1, but audit record is written
if [ "${actual_exit}" -eq 1 ]; then
  pass "exits 1 when HARNESS_MEM_DISABLE=1 (api failure)"
else
  fail "expected exit 1 with HARNESS_MEM_DISABLE=1; got ${actual_exit}"
fi

if [ -f "${CHECKPOINT_EVENTS}" ]; then
  pass "checkpoint-events.jsonl created"
else
  fail "checkpoint-events.jsonl not found after run"
fi

if grep -q '"type":"checkpoint"' "${CHECKPOINT_EVENTS}" 2>/dev/null; then
  pass "checkpoint-events.jsonl contains type=checkpoint record"
else
  fail "checkpoint-events.jsonl missing checkpoint record"
fi

if grep -q '"task":"42.1"' "${CHECKPOINT_EVENTS}" 2>/dev/null; then
  pass "checkpoint record has correct task_id"
else
  fail "checkpoint record missing task_id=42.1"
fi

if grep -q '"commit":"abc1234"' "${CHECKPOINT_EVENTS}" 2>/dev/null; then
  pass "checkpoint record has correct commit hash"
else
  fail "checkpoint record missing commit=abc1234"
fi

if grep -q '"status":"failed"' "${CHECKPOINT_EVENTS}" 2>/dev/null; then
  pass "checkpoint record has status=failed when API disabled"
else
  fail "expected status=failed in checkpoint record"
fi

# session-events.jsonl should also have a failure record
if [ -f "${SESSION_EVENTS}" ]; then
  if grep -q '"type":"checkpoint_failed"' "${SESSION_EVENTS}" 2>/dev/null; then
    pass "session-events.jsonl contains checkpoint_failed record"
  else
    fail "session-events.jsonl missing checkpoint_failed record"
  fi
else
  fail "session-events.jsonl not found after API failure"
fi

# ── Test: harness-mem-client not found ────────────────────────────────────────
echo "Test group: harness-mem-client not found"

PROJECT_DIR2="${TMP}/project2"
mkdir -p "${PROJECT_DIR2}/.claude/state"

actual_exit=0
PROJECT_ROOT="${PROJECT_DIR2}" \
HARNESS_MEM_CLIENT="/nonexistent/harness-mem-client.sh" \
  bash "${CHECKPOINT_SCRIPT}" "99.1" "def5678" "${SPRINT_CONTRACT}" "${REVIEW_RESULT}" \
  2>/dev/null || actual_exit=$?

if [ "${actual_exit}" -eq 1 ]; then
  pass "exits 1 when harness-mem-client not found"
else
  fail "expected exit 1 when harness-mem-client missing; got ${actual_exit}"
fi

CHECKPOINT_EVENTS2="${PROJECT_DIR2}/.claude/state/checkpoint-events.jsonl"
if [ -f "${CHECKPOINT_EVENTS2}" ]; then
  pass "checkpoint-events.jsonl created even when client missing"
else
  fail "checkpoint-events.jsonl not created when client missing"
fi

# ── Test: successful API response (mock client) ────────────────────────────────
echo "Test group: mock harness-mem-client success"

PROJECT_DIR3="${TMP}/project3"
mkdir -p "${PROJECT_DIR3}/.claude/state"

MOCK_CLIENT="${TMP}/mock-mem-client.sh"
cat > "${MOCK_CLIENT}" <<'MOCK'
#!/usr/bin/env bash
# Mock: always succeed
printf '{"ok":true,"id":"checkpoint-mock-001"}\n'
exit 0
MOCK
chmod +x "${MOCK_CLIENT}"

actual_exit=0
PROJECT_ROOT="${PROJECT_DIR3}" \
HARNESS_MEM_CLIENT="${MOCK_CLIENT}" \
  bash "${CHECKPOINT_SCRIPT}" "10.1" "fff0000" "${SPRINT_CONTRACT}" "${REVIEW_RESULT}" \
  2>/dev/null || actual_exit=$?

if [ "${actual_exit}" -eq 0 ]; then
  pass "exits 0 on successful API response"
else
  fail "expected exit 0 for successful API; got ${actual_exit}"
fi

CHECKPOINT_EVENTS3="${PROJECT_DIR3}/.claude/state/checkpoint-events.jsonl"
if [ -f "${CHECKPOINT_EVENTS3}" ]; then
  pass "checkpoint-events.jsonl created on success"
else
  fail "checkpoint-events.jsonl not created on success"
fi

if grep -q '"status":"ok"' "${CHECKPOINT_EVENTS3}" 2>/dev/null; then
  pass "checkpoint record has status=ok on success"
else
  fail "expected status=ok in checkpoint record; got: $(cat "${CHECKPOINT_EVENTS3}" 2>/dev/null)"
fi

# ── Test: mock client returns ok:false ────────────────────────────────────────
echo "Test group: mock harness-mem-client returns ok:false"

PROJECT_DIR4="${TMP}/project4"
mkdir -p "${PROJECT_DIR4}/.claude/state"

MOCK_CLIENT_FAIL="${TMP}/mock-mem-client-fail.sh"
cat > "${MOCK_CLIENT_FAIL}" <<'MOCK'
#!/usr/bin/env bash
# Mock: API error response
printf '{"ok":false,"error":"internal_server_error"}\n'
exit 0
MOCK
chmod +x "${MOCK_CLIENT_FAIL}"

actual_exit=0
PROJECT_ROOT="${PROJECT_DIR4}" \
HARNESS_MEM_CLIENT="${MOCK_CLIENT_FAIL}" \
  bash "${CHECKPOINT_SCRIPT}" "77.2" "aaa1111" "${SPRINT_CONTRACT}" "${REVIEW_RESULT}" \
  2>/dev/null || actual_exit=$?

if [ "${actual_exit}" -eq 1 ]; then
  pass "exits 1 when API returns ok:false"
else
  fail "expected exit 1 when API returns ok:false; got ${actual_exit}"
fi

CHECKPOINT_EVENTS4="${PROJECT_DIR4}/.claude/state/checkpoint-events.jsonl"
if grep -q '"status":"failed"' "${CHECKPOINT_EVENTS4}" 2>/dev/null; then
  pass "checkpoint record has status=failed when API returns ok:false"
else
  fail "expected status=failed when API returns ok:false"
fi

# ── Test: sprint_contract path not found → uses empty JSON ───────────────────
echo "Test group: missing sprint_contract path (graceful degradation)"

PROJECT_DIR5="${TMP}/project5"
mkdir -p "${PROJECT_DIR5}/.claude/state"

actual_exit=0
PROJECT_ROOT="${PROJECT_DIR5}" \
HARNESS_MEM_DISABLE=1 \
  bash "${CHECKPOINT_SCRIPT}" "5.1" "ccc2222" "/nonexistent/contract.json" "${REVIEW_RESULT}" \
  2>/dev/null || actual_exit=$?

# Should still write the audit record (exit 1 due to HARNESS_MEM_DISABLE)
CHECKPOINT_EVENTS5="${PROJECT_DIR5}/.claude/state/checkpoint-events.jsonl"
if [ -f "${CHECKPOINT_EVENTS5}" ]; then
  pass "checkpoint-events.jsonl created even with missing contract path"
else
  fail "checkpoint-events.jsonl not created with missing contract path"
fi

# ── Test: lock files created ───────────────────────────────────────────────────
echo "Test group: lock directory creation"

LOCKS_DIR="${PROJECT_DIR}/.claude/state/locks"
if [ -d "${LOCKS_DIR}" ]; then
  pass "locks directory created"
else
  fail "locks directory not created"
fi

# ── Summary ────────────────────────────────────────────────────────────────────
echo ""
echo "Results: ${PASS} passed, ${FAIL} failed"

if [ "${FAIL}" -gt 0 ]; then
  exit 1
fi
exit 0
