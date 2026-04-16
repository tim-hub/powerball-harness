#!/usr/bin/env bash
# test-advisor-protocol.sh
# Validates the advisor agent and run-advisor-consultation.sh contract.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"  # project-root: tests/
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"  # project-root

# ---------------------------------------------------------------------------
# Check A — advisor agent has read-only allowed-tools (no Write/Edit/Bash/Task)
# ---------------------------------------------------------------------------
TOOLS_LINE=$(grep 'allowed-tools:' "${PROJECT_ROOT}/harness/agents/advisor.md")
for FORBIDDEN in Write Edit Bash Task Agent; do
  if echo "$TOOLS_LINE" | grep -q "\"$FORBIDDEN\""; then
    echo "FAIL: advisor allowed-tools contains forbidden tool: $FORBIDDEN" >&2
    exit 1
  fi
done
echo "PASS: advisor allowed-tools are read-only"

# ---------------------------------------------------------------------------
# Check B — response schema contains PLAN/CORRECTION/STOP
# ---------------------------------------------------------------------------
for DECISION in PLAN CORRECTION STOP; do
  if ! grep -q "$DECISION" "${PROJECT_ROOT}/harness/agents/advisor.md"; then
    echo "FAIL: advisor.md missing decision type: $DECISION" >&2
    exit 1
  fi
done
echo "PASS: response schema contains PLAN/CORRECTION/STOP"

# ---------------------------------------------------------------------------
# Check C — run-advisor-consultation.sh --help exits 0
# ---------------------------------------------------------------------------
if ! bash "${PROJECT_ROOT}/harness/scripts/run-advisor-consultation.sh" --help > /dev/null 2>&1; then
  echo "FAIL: run-advisor-consultation.sh --help returned non-zero" >&2
  exit 1
fi
echo "PASS: --help exits 0"

# ---------------------------------------------------------------------------
# Check D — script uses BASH_SOURCE not $0
# ---------------------------------------------------------------------------
if grep -q 'dirname "\$0"' "${PROJECT_ROOT}/harness/scripts/run-advisor-consultation.sh"; then
  echo "FAIL: script uses \$0 instead of BASH_SOURCE" >&2
  exit 1
fi
if ! grep -q 'BASH_SOURCE' "${PROJECT_ROOT}/harness/scripts/run-advisor-consultation.sh"; then
  echo "FAIL: script does not use BASH_SOURCE" >&2
  exit 1
fi
echo "PASS: script uses BASH_SOURCE for path resolution"

echo "All advisor protocol checks passed."
