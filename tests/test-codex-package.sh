#!/usr/bin/env bash
#
# test-codex-package.sh
# Validate Codex CLI package contents
#
# Usage: ./tests/test-codex-package.sh
#

set -euo pipefail

PASSED=0
FAILED=0

log_test() { echo "[TEST] $1"; }
log_pass() { echo "[PASS] $1"; PASSED=$((PASSED + 1)); }
log_fail() { echo "[FAIL] $1"; FAILED=$((FAILED + 1)); }

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

# Test 1: required files
log_test "Required files exist"
required_files=(
  "codex/AGENTS.md"
  "codex/README.md"
  "codex/.codex/rules/harness.rules"
)
all_exist=true
for file in "${required_files[@]}"; do
  if [ -f "$file" ]; then
    echo "  ok: $file"
  else
    echo "  missing: $file"
    all_exist=false
  fi
done
if $all_exist; then
  log_pass "Required files present"
else
  log_fail "Missing required files"
fi

# Test 2: skills directory parity
log_test "Skills parity by SKILL name"
if [ -d "opencode/skills" ] && [ -d "codex/.codex/skills" ]; then
  get_skill_names() {
    local root="$1"
    find "$root" -mindepth 1 -maxdepth 1 -type d | while IFS= read -r d; do
      if [ -f "$d/SKILL.md" ]; then
        sed -n 's/^name:[[:space:]]*//p' "$d/SKILL.md" | head -n 1 | tr -d '\"'
      fi
    done | sort
  }

  source_list=$(get_skill_names opencode/skills)
  target_list=$(get_skill_names codex/.codex/skills)

  if diff -u <(echo "$source_list") <(echo "$target_list") >/dev/null; then
    log_pass "Skill names match"
  else
    echo "[DETAIL] opencode vs codex skill names differ"
    diff -u <(echo "$source_list") <(echo "$target_list") || true
    log_fail "Skill names mismatch"
  fi
else
  log_fail "Skills directories missing"
fi

# Test 3: SKILL.md exists for each Codex skill
log_test "Each Codex skill has SKILL.md"
missing_skill=false
while IFS= read -r skill_dir; do
  if [ ! -f "$skill_dir/SKILL.md" ]; then
    echo "  missing: $skill_dir/SKILL.md"
    missing_skill=true
  fi
done < <(find codex/.codex/skills -mindepth 1 -maxdepth 1 -type d | sort)

if $missing_skill; then
  log_fail "Missing SKILL.md"
else
  log_pass "All skills have SKILL.md"
fi

# Summary
if [ "$FAILED" -eq 0 ]; then
  echo "All tests passed: $PASSED"
  exit 0
fi

echo "Tests failed: $FAILED (passed: $PASSED)"
exit 1
