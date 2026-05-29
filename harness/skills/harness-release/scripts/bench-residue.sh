#!/usr/bin/env bash
# bench-residue.sh — Benchmark the migration residue scanner
# Compares batched rg (new default) vs sequential grep (legacy).
#
# Usage:
#   bash harness/skills/harness-release/scripts/bench-residue.sh
#
# Set RESIDUE_SCANNER_BACKEND=rg|grep|auto to override the tested backends.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCANNER="${SCRIPT_DIR}/check-residue.py"

echo "=== Residue Scanner Benchmark ==="
echo "Scanner : ${SCANNER}"
echo "Date    : $(date)"
echo ""

bench() {
    local label="$1" backend="$2"
    echo "--- ${label} ---"
    { time RESIDUE_SCANNER_BACKEND="${backend}" python3 "${SCANNER}" > /dev/null 2>&1; } 2>&1
    echo ""
}

if command -v rg > /dev/null 2>&1; then
    RG_VERSION=$(rg --version | head -1)
    echo "rg      : ${RG_VERSION}"
    echo ""
    bench "rg  (batched, new default)" "rg"
else
    echo "rg not installed — skipping rg benchmark."
    echo "Install ripgrep: brew install ripgrep  OR  cargo install ripgrep"
    echo ""
fi

bench "grep (sequential, legacy)" "grep"

echo "Tip: the 'real' row is wall-clock time. rg is typically 10-30x faster"
echo "     because it does one filesystem scan vs one scan per deleted concept."
