#!/usr/bin/env bash
# test-detect-review-plateau.sh
# Unit tests for harness/scripts/detect-review-plateau.sh
#
# Exit code: 0 = all tests passed, 1 = at least one test failed

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
DETECT_SCRIPT="${REPO_ROOT}/harness/scripts/detect-review-plateau.sh"

PASS=0
FAIL=0

pass() { echo "  PASS: $1"; PASS=$((PASS + 1)); }
fail() { echo "  FAIL: $1"; FAIL=$((FAIL + 1)); }

run_detect() {
  bash "${DETECT_SCRIPT}" "$@" 2>/dev/null
}

# ── Setup ──────────────────────────────────────────────────────────────────────
TMP="$(mktemp -d)"
CALIBRATION_FILE="${TMP}/review-calibration.jsonl"
trap 'rm -rf "${TMP}"' EXIT

# Helper: build a calibration entry for a given task_id with an optional files_changed list
make_entry() {
  local task_id="$1"
  shift
  local files_json="[]"
  if [ $# -gt 0 ]; then
    files_json="[$(printf '"%s",' "$@" | sed 's/,$//')]"
  fi
  printf '{"task":{"id":"%s"},"review_result_snapshot":{"files_changed":%s}}\n' \
    "${task_id}" "${files_json}"
}

# ── Test: missing task_id argument ────────────────────────────────────────────
echo "Test group: argument validation"

output="$(bash "${DETECT_SCRIPT}" 2>&1 || true)"
if echo "${output}" | grep -q "Usage"; then
  pass "prints usage when no task_id given"
else
  fail "expected Usage message; got: ${output}"
fi

# ── Test: --help ───────────────────────────────────────────────────────────────
output="$(bash "${DETECT_SCRIPT}" --help 2>&1)"
if echo "${output}" | grep -q "detect-review-plateau"; then
  pass "--help shows usage"
else
  fail "--help did not show usage"
fi

# ── Test: calibration file not found ──────────────────────────────────────────
echo "Test group: missing calibration file"

output="$(run_detect "42.1" --calibration-file "/nonexistent/path.jsonl" || true)"
exit_code="$(bash "${DETECT_SCRIPT}" "42.1" --calibration-file "/nonexistent/path.jsonl" 2>/dev/null; echo $?) " || true
actual_exit=0
bash "${DETECT_SCRIPT}" "42.1" --calibration-file "/nonexistent/path.jsonl" > /dev/null 2>&1 || actual_exit=$?

if [ "${actual_exit}" -eq 1 ]; then
  pass "exits 1 when calibration file not found"
else
  fail "expected exit 1 for missing calibration file; got exit ${actual_exit}"
fi

if echo "${output}" | grep -q "INSUFFICIENT_DATA"; then
  pass "outputs INSUFFICIENT_DATA when calibration file not found"
else
  fail "expected INSUFFICIENT_DATA; got: ${output}"
fi

# ── Test: fewer than 3 entries → INSUFFICIENT_DATA ────────────────────────────
echo "Test group: fewer than 3 entries"

printf '' > "${CALIBRATION_FILE}"
make_entry "task42" "foo.sh" >> "${CALIBRATION_FILE}"
make_entry "task42" "bar.sh" >> "${CALIBRATION_FILE}"

actual_exit=0
output="$(bash "${DETECT_SCRIPT}" "task42" --calibration-file "${CALIBRATION_FILE}" 2>/dev/null)" || actual_exit=$?

if [ "${actual_exit}" -eq 1 ]; then
  pass "exits 1 with 2 entries (INSUFFICIENT_DATA)"
else
  fail "expected exit 1 with 2 entries; got exit ${actual_exit}"
fi

if echo "${output}" | grep -q "INSUFFICIENT_DATA"; then
  pass "STATUS is INSUFFICIENT_DATA with 2 entries"
else
  fail "expected INSUFFICIENT_DATA; got: ${output}"
fi

if echo "${output}" | grep -q "ENTRIES: 2"; then
  pass "ENTRIES count is 2"
else
  fail "expected ENTRIES: 2; got: ${output}"
fi

# ── Test: 3 entries with different files → PIVOT_NOT_REQUIRED ─────────────────
echo "Test group: 3 entries with low similarity (PIVOT_NOT_REQUIRED)"

printf '' > "${CALIBRATION_FILE}"
make_entry "task-prog" "a.sh" "b.sh" >> "${CALIBRATION_FILE}"
make_entry "task-prog" "c.sh" "d.sh" >> "${CALIBRATION_FILE}"
make_entry "task-prog" "e.sh" "f.sh" >> "${CALIBRATION_FILE}"

actual_exit=0
output="$(bash "${DETECT_SCRIPT}" "task-prog" --calibration-file "${CALIBRATION_FILE}" 2>/dev/null)" || actual_exit=$?

if [ "${actual_exit}" -eq 0 ]; then
  pass "exits 0 when review is making progress (low Jaccard)"
else
  fail "expected exit 0 for progress; got exit ${actual_exit}"
fi

if echo "${output}" | grep -q "PIVOT_NOT_REQUIRED"; then
  pass "STATUS is PIVOT_NOT_REQUIRED for low-similarity entries"
else
  fail "expected PIVOT_NOT_REQUIRED; got: ${output}"
fi

if echo "${output}" | grep -q "JACCARD_AVG:"; then
  pass "JACCARD_AVG is present in output"
else
  fail "expected JACCARD_AVG in output; got: ${output}"
fi

# ── Test: 3 entries with identical files → PIVOT_REQUIRED ─────────────────────
echo "Test group: 3 entries with identical files (PIVOT_REQUIRED)"

printf '' > "${CALIBRATION_FILE}"
make_entry "task-stuck" "scripts/foo.sh" "scripts/bar.sh" >> "${CALIBRATION_FILE}"
make_entry "task-stuck" "scripts/foo.sh" "scripts/bar.sh" >> "${CALIBRATION_FILE}"
make_entry "task-stuck" "scripts/foo.sh" "scripts/bar.sh" >> "${CALIBRATION_FILE}"

actual_exit=0
output="$(bash "${DETECT_SCRIPT}" "task-stuck" --calibration-file "${CALIBRATION_FILE}" 2>/dev/null)" || actual_exit=$?

if [ "${actual_exit}" -eq 2 ]; then
  pass "exits 2 when pivot is required (high Jaccard)"
else
  fail "expected exit 2 for plateau; got exit ${actual_exit}"
fi

if echo "${output}" | grep -q "PIVOT_REQUIRED"; then
  pass "STATUS is PIVOT_REQUIRED for identical-file entries"
else
  fail "expected PIVOT_REQUIRED; got: ${output}"
fi

# ── Test: only entries for a different task_id → INSUFFICIENT_DATA ─────────────
echo "Test group: entries for a different task_id"

printf '' > "${CALIBRATION_FILE}"
make_entry "other-task" "x.sh" >> "${CALIBRATION_FILE}"
make_entry "other-task" "y.sh" >> "${CALIBRATION_FILE}"
make_entry "other-task" "z.sh" >> "${CALIBRATION_FILE}"

actual_exit=0
output="$(bash "${DETECT_SCRIPT}" "my-task" --calibration-file "${CALIBRATION_FILE}" 2>/dev/null)" || actual_exit=$?

if [ "${actual_exit}" -eq 1 ]; then
  pass "exits 1 when no entries match the given task_id"
else
  fail "expected exit 1 for mismatched task_id; got exit ${actual_exit}"
fi

if echo "${output}" | grep -q "INSUFFICIENT_DATA"; then
  pass "STATUS is INSUFFICIENT_DATA when task_id doesn't match"
else
  fail "expected INSUFFICIENT_DATA for mismatched task_id; got: ${output}"
fi

# ── Test: gaps[].location fallback ────────────────────────────────────────────
echo "Test group: gaps[].location fallback for file extraction"

printf '' > "${CALIBRATION_FILE}"
printf '{"task":{"id":"gaps-task"},"gaps":[{"location":"src/a.go:42"},{"location":"src/a.go:99"}]}\n' >> "${CALIBRATION_FILE}"
printf '{"task":{"id":"gaps-task"},"gaps":[{"location":"src/a.go:10"},{"location":"src/b.go:5"}]}\n' >> "${CALIBRATION_FILE}"
printf '{"task":{"id":"gaps-task"},"gaps":[{"location":"src/a.go:77"},{"location":"src/b.go:3"}]}\n' >> "${CALIBRATION_FILE}"

output="$(bash "${DETECT_SCRIPT}" "gaps-task" --calibration-file "${CALIBRATION_FILE}" 2>/dev/null)" || true

if echo "${output}" | grep -q "JACCARD_AVG:"; then
  pass "gap-based file extraction works (JACCARD_AVG present)"
else
  fail "expected JACCARD_AVG when using gaps fallback; got: ${output}"
fi

# ── Summary ────────────────────────────────────────────────────────────────────
echo ""
echo "Results: ${PASS} passed, ${FAIL} failed"

if [ "${FAIL}" -gt 0 ]; then
  exit 1
fi
exit 0
