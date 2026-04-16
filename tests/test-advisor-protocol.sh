#!/bin/bash
# test-advisor-protocol.sh
# Verifies advisor agent read-only enforcement and run-advisor-consultation.sh behavior.

set -euo pipefail
export TMPDIR=/tmp  # Force /tmp for sandboxed execution (sandbox blocks /var/folders)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
ADVISOR_MD="${PROJECT_ROOT}/harness/agents/advisor.md"
CONSULT_SCRIPT="${PROJECT_ROOT}/harness/scripts/run-advisor-consultation.sh"

PASS=0
FAIL=0

pass() { echo "PASS: $1"; PASS=$((PASS + 1)); }
fail() { echo "FAIL: $1"; FAIL=$((FAIL + 1)); }

# ---------------------------------------------------------------------------
# Test 1 — Agent has read-only allowed-tools
# ---------------------------------------------------------------------------
ALLOWED_LINE="$(grep '^allowed-tools:' "${ADVISOR_MD}" | head -1)"

for FORBIDDEN in Write Edit Bash Task; do
  if echo "${ALLOWED_LINE}" | grep -q "${FORBIDDEN}"; then
    fail "Test 1: '${FORBIDDEN}' must NOT appear in advisor allowed-tools (found in: ${ALLOWED_LINE})"
  fi
done

for REQUIRED in Read Grep Glob; do
  if echo "${ALLOWED_LINE}" | grep -q "${REQUIRED}"; then
    pass "Test 1: '${REQUIRED}' appears in advisor allowed-tools"
  else
    fail "Test 1: '${REQUIRED}' must appear in advisor allowed-tools (line: ${ALLOWED_LINE})"
  fi
done

# ---------------------------------------------------------------------------
# Test 2 — Response schema contains PLAN/CORRECTION/STOP
# ---------------------------------------------------------------------------
MATCH_COUNT="$(grep -c 'PLAN\|CORRECTION\|STOP' "${ADVISOR_MD}" || true)"
if [ "${MATCH_COUNT}" -ge 3 ]; then
  pass "Test 2: advisor.md contains >= 3 matches for PLAN/CORRECTION/STOP (found: ${MATCH_COUNT})"
else
  fail "Test 2: expected >= 3 matches for PLAN/CORRECTION/STOP, got ${MATCH_COUNT}"
fi

# ---------------------------------------------------------------------------
# Test 3 — Script respects max_consults_per_task
# ---------------------------------------------------------------------------
TMP_DIR="$(mktemp -d "/tmp/harness-test-advisor.XXXXXX")"
trap 'rm -rf "${TMP_DIR}"' EXIT

# Set up fake project structure with state directory
mkdir -p "${TMP_DIR}/.claude/state/advisor"

# Write 3 history entries for the same task_id (matching max_consults_per_task)
for i in 1 2 3; do
  echo "{\"task_id\":\"test-task-1\", \"reason_code\":\"repeated_failure\", \"timestamp\":\"2026-01-0${i}T00:00:00Z\"}" \
    >> "${TMP_DIR}/.claude/state/advisor/history.jsonl"
done

# Initialize a git repo so git rev-parse --show-toplevel works inside the script
git -C "${TMP_DIR}" init -q
git -C "${TMP_DIR}" config user.name "Harness Test"
git -C "${TMP_DIR}" config user.email "harness-test@example.com"

# Create a patched copy of the script inside harness/scripts/ to preserve
# the PLUGIN_DIR="${SCRIPT_DIR}/.." -> harness/ path resolution, and
# override PROJECT_ROOT to point to TMP_DIR.
mkdir -p "${TMP_DIR}/harness/scripts"
cp "${PROJECT_ROOT}/harness/.claude-code-harness.config.yaml" \
   "${TMP_DIR}/harness/.claude-code-harness.config.yaml"
PATCHED_SCRIPT="${TMP_DIR}/harness/scripts/run-advisor-consultation.sh"
sed \
  "s|PROJECT_ROOT=\"\$(git -C \"\${SCRIPT_DIR}\" rev-parse --show-toplevel)\"|PROJECT_ROOT=\"${TMP_DIR}\"|" \
  "${CONSULT_SCRIPT}" > "${PATCHED_SCRIPT}"
chmod +x "${PATCHED_SCRIPT}"

CONSULT_OUTPUT="$(bash "${PATCHED_SCRIPT}" \
  --task-id "test-task-1" \
  --reason-code "repeated_failure" \
  --error-sig "test error" 2>&1)" || true

CONSULT_EXIT=$?
if [ "${CONSULT_EXIT}" -eq 0 ] && echo "${CONSULT_OUTPUT}" | grep -qi "max consultations reached"; then
  pass "Test 3: script exits 0 and outputs 'max consultations reached' when limit is hit"
else
  fail "Test 3: expected exit 0 with 'max consultations reached', got exit=${CONSULT_EXIT}, output=${CONSULT_OUTPUT}"
fi

# ---------------------------------------------------------------------------
# Test 4 — BASH_SOURCE is used (not \$0) for path resolution
# ---------------------------------------------------------------------------
if grep -q 'BASH_SOURCE' "${CONSULT_SCRIPT}"; then
  pass "Test 4: BASH_SOURCE is present in run-advisor-consultation.sh"
else
  fail "Test 4: BASH_SOURCE not found in run-advisor-consultation.sh"
fi

# Ensure \$0 does not appear in path resolution lines (SCRIPT_DIR / PROJECT_ROOT / PLUGIN_DIR)
if grep -E '^\s*(SCRIPT_DIR|PROJECT_ROOT|PLUGIN_DIR)=' "${CONSULT_SCRIPT}" | grep -q '\$0'; then
  fail "Test 4: \$0 used in path resolution — should use BASH_SOURCE instead"
else
  pass "Test 4: \$0 not used in path resolution lines"
fi

# ---------------------------------------------------------------------------
# Test 5 — --help flag exits 0
# ---------------------------------------------------------------------------
if bash "${CONSULT_SCRIPT}" --help >/dev/null 2>&1; then
  pass "Test 5: --help exits 0"
else
  fail "Test 5: --help did not exit 0"
fi

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
echo ""
echo "Results: ${PASS} passed, ${FAIL} failed"

if [ "${FAIL}" -gt 0 ]; then
  exit 1
fi

echo "test-advisor-protocol: ok"
