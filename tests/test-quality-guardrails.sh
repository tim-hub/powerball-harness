#!/bin/bash
# test-quality-guardrails.sh
# Validation tests for test tampering prevention (3-layer defense strategy)
#
# Test targets:
# - Layer 1: Rules template existence and structure
# - Layer 2: Skills quality guardrail integration
# - harness-init deployment configuration

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PLUGIN_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Counters
PASSED=0
FAILED=0
TOTAL=0

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Test function
assert_file_exists() {
  local file="$1"
  local description="$2"
  TOTAL=$((TOTAL + 1))

  if [ -f "$PLUGIN_ROOT/$file" ]; then
    echo -e "${GREEN}✓${NC} $description"
    PASSED=$((PASSED + 1))
    return 0
  else
    echo -e "${RED}✗${NC} $description"
    echo "  Expected file: $file"
    FAILED=$((FAILED + 1))
    return 1
  fi
}

assert_file_contains() {
  local file="$1"
  local pattern="$2"
  local description="$3"
  TOTAL=$((TOTAL + 1))

  if [ ! -f "$PLUGIN_ROOT/$file" ]; then
    echo -e "${RED}✗${NC} $description (file not found)"
    FAILED=$((FAILED + 1))
    return 1
  fi

  if grep -qE "$pattern" "$PLUGIN_ROOT/$file"; then
    echo -e "${GREEN}✓${NC} $description"
    PASSED=$((PASSED + 1))
    return 0
  else
    echo -e "${RED}✗${NC} $description"
    echo "  Expected pattern: $pattern"
    echo "  File: $file"
    FAILED=$((FAILED + 1))
    return 1
  fi
}

assert_json_key_exists() {
  local file="$1"
  local key="$2"
  local description="$3"
  TOTAL=$((TOTAL + 1))

  if ! command -v jq >/dev/null 2>&1; then
    echo -e "${YELLOW}⚠${NC} $description (jq not available, skipped)"
    return 0
  fi

  if [ ! -f "$PLUGIN_ROOT/$file" ]; then
    echo -e "${RED}✗${NC} $description (file not found)"
    FAILED=$((FAILED + 1))
    return 1
  fi

  if jq -e "$key" "$PLUGIN_ROOT/$file" >/dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} $description"
    PASSED=$((PASSED + 1))
    return 0
  else
    echo -e "${RED}✗${NC} $description"
    echo "  Expected key: $key"
    echo "  File: $file"
    FAILED=$((FAILED + 1))
    return 1
  fi
}

echo "=================================================="
echo "Test tampering prevention (3-layer defense strategy) validation"
echo "=================================================="
echo ""

# ============================================
# Layer 1: Rules template validation
# ============================================
echo "## Layer 1: Rules templates"
echo ""

# Check existence of test quality rule template
assert_file_exists \
  "harness/templates/rules/test-quality.md.template" \
  "test-quality.md.template exists"

# Check existence of implementation quality rule template
assert_file_exists \
  "harness/templates/rules/implementation-quality.md.template" \
  "implementation-quality.md.template exists"

# Required content in test-quality.md
assert_file_contains \
  "harness/templates/rules/test-quality.md.template" \
  "it.skip|test.skip" \
  "test-quality.md contains skip prohibition pattern"

assert_file_contains \
  "harness/templates/rules/test-quality.md.template" \
  "eslint|lint|disable" \
  "test-quality.md contains lint configuration tampering prohibition"

assert_file_contains \
  "harness/templates/rules/test-quality.md.template" \
  "_harness_template" \
  "test-quality.md contains frontmatter metadata"

# Required content in implementation-quality.md
assert_file_contains \
  "harness/templates/rules/implementation-quality.md.template" \
  "ハードコード|hardcode" \
  "implementation-quality.md contains hardcoding prohibition"

assert_file_contains \
  "harness/templates/rules/implementation-quality.md.template" \
  "スタブ|stub" \
  "implementation-quality.md contains stub prohibition"

assert_file_contains \
  "harness/templates/rules/implementation-quality.md.template" \
  "_harness_template" \
  "implementation-quality.md contains frontmatter metadata"

# Registration in template-registry.json
assert_json_key_exists \
  "harness/templates/template-registry.json" \
  '.templates["rules/test-quality.md.template"]' \
  "test-quality.md is registered in template-registry.json"

assert_json_key_exists \
  "harness/templates/template-registry.json" \
  '.templates["rules/implementation-quality.md.template"]' \
  "implementation-quality.md is registered in template-registry.json"

echo ""

# ============================================
# Layer 2: Skills quality guardrail validation
# ============================================
echo "## Layer 2: Skills quality guardrails"
echo ""

# impl skill quality guardrails
assert_file_contains \
  "harness/skills/impl/SKILL.md" \
  "品質ガードレール|Quality Guardrails" \
  "impl/SKILL.md has quality guardrail section"

assert_file_contains \
  "harness/skills/impl/SKILL.md" \
  "禁止パターン|Prohibited|禁止" \
  "impl/SKILL.md has prohibited patterns defined"

assert_file_contains \
  "harness/skills/impl/SKILL.md" \
  "purpose-driven|Purpose-Driven|目的駆動" \
  "impl/SKILL.md has Purpose-Driven Implementation principle"

# verify skill quality guardrails
assert_file_contains \
  "harness/skills/verify/SKILL.md" \
  "品質ガードレール|Quality Guardrails" \
  "verify/SKILL.md has quality guardrail section"

assert_file_contains \
  "harness/skills/verify/SKILL.md" \
  "改ざん禁止|Tampering Prohibited|禁止" \
  "verify/SKILL.md has tampering prohibition patterns defined"

assert_file_contains \
  "harness/skills/verify/SKILL.md" \
  "承認リクエスト|Approval Request" \
  "verify/SKILL.md has approval request format"

echo ""

# ============================================
# harness-init integration validation
# ============================================
echo "## harness-init integration"
echo ""

# harness-init quality rule deployment configuration (after skill migration)
assert_file_contains \
  "harness/skills/harness-init/SKILL.md" \
  "setup|Setup|Environment" \
  "harness-init contains setup functionality"

# Confirm quality rule files exist
assert_file_contains \
  ".claude/rules/test-quality.md" \
  "テスト改ざん|Test Tampering|禁止" \
  "test-quality.md has test tampering prevention rules"

assert_file_contains \
  ".claude/rules/implementation-quality.md" \
  "形骸化|stub|placeholder" \
  "implementation-quality.md has hollow implementation prohibition rules"

echo ""

# ============================================
# Documentation validation
# ============================================
echo "## Documentation"
echo ""

# CLAUDE.md test tampering prevention section
assert_file_contains \
  "CLAUDE.md" \
  "テスト改ざん防止|Test Tampering Prevention" \
  "CLAUDE.md has test tampering prevention section"

assert_file_contains \
  "CLAUDE.md" \
  "3層防御|3-layer|第1層|第2層|第3層" \
  "CLAUDE.md has 3-layer defense strategy description"

# README.md quality assurance references
assert_file_contains \
  "README.md" \
  "Test tampering|Quality|品質" \
  "README.md has quality assurance references"

# Design documentation
assert_file_exists \
  "docs/QUALITY_GUARD_DESIGN.md" \
  "Layer 3 Hooks design document exists"

echo ""

# ============================================
# Results summary
# ============================================
echo "=================================================="
echo "Test Results"
echo "=================================================="
echo ""
echo "Total: $TOTAL"
echo -e "Passed: ${GREEN}$PASSED${NC}"
echo -e "Failed: ${RED}$FAILED${NC}"
echo ""

if [ $FAILED -eq 0 ]; then
  echo -e "${GREEN}✓ All tests passed${NC}"
  exit 0
else
  echo -e "${RED}✗ $FAILED test(s) failed${NC}"
  exit 1
fi
