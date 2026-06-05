#!/usr/bin/env bash
# Guard the harness-plan planning quality contract introduced in Phase 100 (task 100.7).

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

failures=0

err() { echo "test-harness-plan-quality: FAIL: $*" >&2; failures=$((failures + 1)); }

assert_file()     { [ -f "$1" ] || err "missing file: $1"; }
assert_contains() { grep -qF "$2" "$1" || err "$1 missing: $2"; }
assert_absent()   { ! grep -qF "$2" "$1" || err "$1 should not contain: $2"; }

skill="skills/harness-plan/SKILL.md"
create_ref="skills/harness-plan/references/create.md"
quality_ref="skills/harness-plan/references/planning-quality.md"

# --- File existence ---
assert_file "$skill"
assert_file "$create_ref"
assert_file "$quality_ref"

# --- SKILL.md has planning-quality pointer ---
assert_contains "$skill" "planning-quality.md"

# --- create.md has Planning Quality Gate step ---
assert_contains "$create_ref" "Step 1.5"
assert_contains "$create_ref" "Planning Quality Gate"
assert_contains "$create_ref" "planning-quality.md"
assert_contains "$create_ref" "Applicability"
assert_contains "$create_ref" "Subagent Debate"
assert_contains "$create_ref" "Quality Contract Output"

# --- planning-quality.md key content checks ---
assert_contains "$quality_ref" "This is not a standalone subcommand"
assert_contains "$quality_ref" "WebSearch"
assert_contains "$quality_ref" "Do not assume direct access to the harness-mem database"
assert_contains "$quality_ref" "Product / Strategy"
assert_contains "$quality_ref" "Architecture / Implementation"
assert_contains "$quality_ref" "QA / Regression"
assert_contains "$quality_ref" "Skeptic"
assert_contains "$quality_ref" "Implementation Feasibility"
assert_contains "$quality_ref" "Regression Safety"
assert_contains "$quality_ref" "Core to this product"
assert_contains "$quality_ref" "Evidence Strength"
assert_contains "$quality_ref" "Required status is blocked"
assert_contains "$quality_ref" "Add a spike or spec task first"
assert_contains "$quality_ref" "Quality Contract Output"

# --- No Japanese in planning-quality.md ---
if python3 -c "import sys, re; data=open(sys.argv[1]).read(); sys.exit(0 if re.search(r'[぀-ヿ一-鿿]', data) else 1)" "$quality_ref" 2>/dev/null; then
  err "Japanese characters found in $quality_ref"
fi

if [ "$failures" -gt 0 ]; then
  exit 1
fi

echo "test-harness-plan-quality: ok"
