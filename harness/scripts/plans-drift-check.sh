#!/bin/bash
# plans-drift-check.sh — Lightweight Plans.md stale-marker detector
# Runs at harness-work entry before mode selection
#
# Exit 0: no stale markers (safe to proceed)
# Exit 1: stale markers detected (prompt user to confirm)
#
# Usage: bash harness/scripts/plans-drift-check.sh [--quiet]
set -euo pipefail

PROJECT_ROOT="$(git rev-parse --show-toplevel)"  # project-root

PLANS_FILE="${PROJECT_ROOT}/Plans.md"
QUIET="${1:-}"

if [ ! -f "${PLANS_FILE}" ]; then
  exit 0  # No Plans.md — no drift possible
fi

STALE_COUNT=0
STALE_REPORT=""

# --- Check cc:WIP and cc:TODO tasks for staleness ---
# Parse each line looking for task rows with task IDs (e.g. 3, 12.1, 65.1)
while IFS= read -r line; do
  # Match task ID patterns like | 3 | or | 12.1 | or | 65.1 |
  # Task rows in Plans.md v2 format: | N | Task Name | Content | DoD | Depends | Status |
  if [[ "${line}" =~ \|[[:space:]]*([0-9]+(\.[0-9]+)*)[[:space:]]*\| ]]; then
    task_id="${BASH_REMATCH[1]}"

    # Check cc:WIP tasks — potentially abandoned if no recent commit references them
    if echo "${line}" | grep -q 'cc:WIP'; then
      if ! git -C "${PROJECT_ROOT}" log --oneline -10 | grep -qF "${task_id}"; then
        STALE_REPORT="${STALE_REPORT}\n  [WIP] ${task_id} — no recent commit mentions this task (possibly abandoned WIP)"
        STALE_COUNT=$((STALE_COUNT + 1))
      fi
    fi

    # Check cc:TODO tasks — potentially already done if recent commit mentions the task number
    if echo "${line}" | grep -q 'cc:TODO'; then
      if git -C "${PROJECT_ROOT}" log --oneline -20 | grep -qF "${task_id}"; then
        STALE_REPORT="${STALE_REPORT}\n  [TODO] ${task_id} — recent commit mentions this task (may already be implemented)"
        STALE_COUNT=$((STALE_COUNT + 1))
      fi
    fi
  fi
done < "${PLANS_FILE}"

if [ "${STALE_COUNT}" -eq 0 ]; then
  if [ "${QUIET}" != "--quiet" ]; then
    echo "Plans.md drift check: OK (no stale markers)"
  fi
  exit 0
fi

echo ""
echo "Plans.md drift detected — ${STALE_COUNT} potentially stale marker(s):"
printf "%b\n" "${STALE_REPORT}"
echo ""
echo "Run /harness-plan sync to review and correct markers."
echo ""

exit 1
