#!/usr/bin/env bash
# integration/loop-plans-concurrent.sh
# Integration test: concurrent safety of Plans.md access.
#
# This test verifies that the flock guard in auto-checkpoint.sh prevents
# interleaved writes to checkpoint-events.jsonl when multiple processes
# call the checkpoint script simultaneously.
#
# Strategy:
# 1. Launch N parallel checkpoint calls with HARNESS_MEM_DISABLE=1.
# 2. Wait for all to complete.
# 3. Verify that checkpoint-events.jsonl has exactly N well-formed JSON lines
#    (no interleaved/corrupted output).
#
# Exit code: 0 = passed, 1 = failed

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
CHECKPOINT_SCRIPT="${REPO_ROOT}/harness/scripts/auto-checkpoint.sh"

TMP="$(mktemp -d)"
trap 'rm -rf "${TMP}"' EXIT

PROJECT_DIR="${TMP}/concurrent-test"
mkdir -p "${PROJECT_DIR}/.claude/state"

SPRINT_CONTRACT="${TMP}/contract.json"
REVIEW_RESULT="${TMP}/review-result.json"

printf '{"task":{"id":"concurrent"},"checks":[]}' > "${SPRINT_CONTRACT}"
printf '{"schema_version":"review-result.v1","verdict":"APPROVE"}' > "${REVIEW_RESULT}"

PASS=0
FAIL=0

pass() { echo "  PASS: $1"; PASS=$((PASS + 1)); }
fail() { echo "  FAIL: $1"; FAIL=$((FAIL + 1)); }

# ── Test: N concurrent checkpoints produce N valid JSONL lines ────────────────
echo "Test group: concurrent checkpoint writes"

PARALLEL_COUNT=5

pids=()
for i in $(seq 1 ${PARALLEL_COUNT}); do
  (
    PROJECT_ROOT="${PROJECT_DIR}" \
    HARNESS_MEM_DISABLE=1 \
    CHECKPOINT_LOCK_TIMEOUT=15 \
      bash "${CHECKPOINT_SCRIPT}" \
        "task-${i}" \
        "commit${i}abc" \
        "${SPRINT_CONTRACT}" \
        "${REVIEW_RESULT}" \
      2>/dev/null
    # Exit 1 is expected (HARNESS_MEM_DISABLE=1) — that's fine
    true
  ) &
  pids+=($!)
done

# Wait for all background processes
for pid in "${pids[@]}"; do
  wait "${pid}" 2>/dev/null || true
done

CHECKPOINT_EVENTS="${PROJECT_DIR}/.claude/state/checkpoint-events.jsonl"

if [ -f "${CHECKPOINT_EVENTS}" ]; then
  pass "checkpoint-events.jsonl created by concurrent writers"
else
  fail "checkpoint-events.jsonl not found after concurrent writes"
fi

# Count lines
line_count=0
if [ -f "${CHECKPOINT_EVENTS}" ]; then
  line_count="$(wc -l < "${CHECKPOINT_EVENTS}" | tr -d ' ')"
fi

if [ "${line_count}" -eq "${PARALLEL_COUNT}" ]; then
  pass "exactly ${PARALLEL_COUNT} lines in checkpoint-events.jsonl (no missing writes)"
else
  fail "expected ${PARALLEL_COUNT} lines; got ${line_count}"
fi

# Verify each line is valid JSON
invalid_count=0
if [ -f "${CHECKPOINT_EVENTS}" ]; then
  while IFS= read -r line; do
    if ! echo "${line}" | python3 -c "import json,sys; json.loads(sys.stdin.read())" 2>/dev/null; then
      invalid_count=$((invalid_count + 1))
    fi
  done < "${CHECKPOINT_EVENTS}"
fi

if [ "${invalid_count}" -eq 0 ]; then
  pass "all ${line_count} JSONL lines are valid JSON"
else
  fail "${invalid_count}/${line_count} JSONL lines are invalid JSON (possible interleaving)"
fi

# Verify each task appears exactly once
for i in $(seq 1 ${PARALLEL_COUNT}); do
  task_count=0
  if [ -f "${CHECKPOINT_EVENTS}" ]; then
    task_count="$(grep -c "\"task-${i}\"" "${CHECKPOINT_EVENTS}" 2>/dev/null || true)"
  fi
  if [ "${task_count}" -eq 1 ]; then
    pass "task-${i} appears exactly once in checkpoint-events.jsonl"
  else
    fail "task-${i} appears ${task_count} times (expected 1)"
  fi
done

# ── Test: session-events.jsonl also has N entries (failure records) ────────────
SESSION_EVENTS="${PROJECT_DIR}/.claude/state/session-events.jsonl"
if [ -f "${SESSION_EVENTS}" ]; then
  session_count="$(wc -l < "${SESSION_EVENTS}" | tr -d ' ')"
  if [ "${session_count}" -eq "${PARALLEL_COUNT}" ]; then
    pass "session-events.jsonl has ${PARALLEL_COUNT} entries (one per checkpoint failure)"
  else
    fail "expected ${PARALLEL_COUNT} session events; got ${session_count}"
  fi
else
  fail "session-events.jsonl not found after concurrent checkpoint failures"
fi

echo ""
echo "Results: ${PASS} passed, ${FAIL} failed"
[ "${FAIL}" -eq 0 ] || exit 1
exit 0
