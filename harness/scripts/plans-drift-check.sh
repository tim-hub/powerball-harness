#!/bin/bash
# plans-drift-check.sh — Lightweight stale-marker detector over plans.json
# Runs at harness-work entry before mode selection.
#
# Exit 0: no stale markers (safe to proceed)
# Exit 1: stale markers detected (prompt user to confirm)
#
# SSOT: .claude/harness/plans.json
# Usage: bash harness/scripts/plans-drift-check.sh [--quiet]
set -euo pipefail

PROJECT_ROOT="$(git rev-parse --show-toplevel)"  # project-root
PLANS_JSON="${PROJECT_ROOT}/.claude/harness/plans.json"
QUIET="${1:-}"

# No plans.json or no jq — no drift possible / cannot inspect
if [ ! -f "${PLANS_JSON}" ] || ! command -v jq >/dev/null 2>&1; then
  exit 0
fi

STALE_COUNT=0
STALE_REPORT=""

# cc:WIP tasks — potentially abandoned if no recent commit references them
while IFS= read -r task_id; do
  [ -z "${task_id}" ] && continue
  if ! git -C "${PROJECT_ROOT}" log --oneline -10 | grep -qF "${task_id}"; then
    STALE_REPORT="${STALE_REPORT}\n  [WIP] ${task_id} — no recent commit mentions this task (possibly abandoned WIP)"
    STALE_COUNT=$((STALE_COUNT + 1))
  fi
done < <(jq -r '.phases[].tasks[]? | select(.status=="cc:WIP") | .id' "${PLANS_JSON}" 2>/dev/null)

# cc:TODO tasks — potentially already done if a recent commit mentions the task id
while IFS= read -r task_id; do
  [ -z "${task_id}" ] && continue
  if git -C "${PROJECT_ROOT}" log --oneline -20 | grep -qF "${task_id}"; then
    STALE_REPORT="${STALE_REPORT}\n  [TODO] ${task_id} — recent commit mentions this task (may already be implemented)"
    STALE_COUNT=$((STALE_COUNT + 1))
  fi
done < <(jq -r '.phases[].tasks[]? | select(.status=="cc:TODO") | .id' "${PLANS_JSON}" 2>/dev/null)

if [ "${STALE_COUNT}" -eq 0 ]; then
  if [ "${QUIET}" != "--quiet" ]; then
    echo "plans.json drift check: OK (no stale markers)"
  fi
  exit 0
fi

echo ""
echo "plans.json drift detected — ${STALE_COUNT} potentially stale marker(s):"
printf "%b\n" "${STALE_REPORT}"
echo ""
echo "Run /harness-plan sync to review and correct markers."
echo ""

exit 1
