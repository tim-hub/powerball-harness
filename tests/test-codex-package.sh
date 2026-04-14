#!/usr/bin/env bash
#
# test-codex-package.sh
# Validate Codex/OpenCode setup templates and harness-setup scripts
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

# Test 1: Codex template files exist
log_test "Codex template files exist"
required_codex_files=(
  "templates/codex/config.toml"
  "templates/codex/rules/harness.rules"
  "templates/codex/.codexignore"
  "templates/codex/AGENTS.md"
)
codex_files_ok=true
for file in "${required_codex_files[@]}"; do
  if [ -f "$file" ]; then
    echo "  ok: $file"
  else
    echo "  missing: $file"
    codex_files_ok=false
  fi
done
if $codex_files_ok; then
  log_pass "Codex template files present"
else
  log_fail "Missing Codex template files"
fi

# Test 1.5: execpolicy rules examples are consistent
log_test "Execpolicy rules examples are valid"
if command -v python3 >/dev/null 2>&1; then
  RULES_FILE="templates/codex/rules/harness.rules"
  if [ ! -f "$RULES_FILE" ]; then
    log_fail "harness.rules not found at $RULES_FILE"
  elif python3 -c "
import shlex, sys, re
text = open('$RULES_FILE').read()
pattern_re = re.compile(r'pattern\s*=\s*\[([^\]]+)\]')
match_re = re.compile(r'\b(not_match|match)\s*=\s*\[([^\]]*)\]', re.DOTALL)
errors = []
for m in match_re.finditer(text):
    field = m.group(1)
    should_match = field == 'match'
    examples = [e.strip().strip('\"') for e in m.group(2).split(',') if e.strip().strip('\"')]
    for ex in examples:
        tokens = shlex.split(ex) if ex else []
        if not tokens and ex:
            errors.append(f'bad example: {ex!r}')
if errors:
    print('ERROR:', errors)
    sys.exit(1)
print('ok')
" 2>/dev/null; then
    log_pass "Rules examples are consistent"
  else
    log_fail "Rules examples invalid (Codex may ignore custom rules)"
  fi
else
  echo "  skipped: python3 not found"
  log_pass "Rules examples check skipped"
fi

# Test 1.6: Codex config.toml has multi_agent + harness roles
log_test "templates/codex/config.toml has multi_agent + harness roles"
config_ok=true
if ! grep -q "multi_agent = true" "templates/codex/config.toml"; then
  echo "  missing: multi_agent = true"
  config_ok=false
fi
for role in "implementer" "reviewer" "claude_implementer" "claude_reviewer"; do
  if ! grep -q "\[agents\.${role}\]" "templates/codex/config.toml"; then
    echo "  missing: [agents.${role}]"
    config_ok=false
  fi
done
if $config_ok; then
  log_pass "config.toml has required multi-agent defaults"
else
  log_fail "config.toml missing required multi-agent defaults"
fi

# Test 2: OpenCode template files exist
log_test "OpenCode template files exist"
required_oc_files=(
  "templates/opencode/opencode.json"
  "templates/opencode/AGENTS.md"
)
oc_files_ok=true
for file in "${required_oc_files[@]}"; do
  if [ -f "$file" ]; then
    echo "  ok: $file"
  else
    echo "  missing: $file"
    oc_files_ok=false
  fi
done
if [ ! -d "templates/opencode/commands" ]; then
  echo "  missing: templates/opencode/commands/"
  oc_files_ok=false
fi
if $oc_files_ok; then
  log_pass "OpenCode template files present"
else
  log_fail "Missing OpenCode template files"
fi

# Test 3: Codex-native skill overrides exist
log_test "Codex-native skill overrides in templates/codex-skills/"
if [ -d "templates/codex-skills" ]; then
  override_count=$(find "templates/codex-skills" -name "SKILL.md" | wc -l | tr -d ' ')
  if [ "$override_count" -gt 0 ]; then
    log_pass "templates/codex-skills/ has $override_count skill override(s)"
  else
    log_fail "templates/codex-skills/ exists but has no SKILL.md files"
  fi
else
  log_fail "templates/codex-skills/ not found"
fi

# Test 4: harness-setup setup scripts exist
log_test "harness-setup scripts exist"
setup_scripts_ok=true
for script in "skills/harness-setup/scripts/setup-codex.sh" "skills/harness-setup/scripts/setup-opencode.sh"; do
  if [ -f "$script" ]; then
    echo "  ok: $script"
  else
    echo "  missing: $script"
    setup_scripts_ok=false
  fi
done
if $setup_scripts_ok; then
  log_pass "harness-setup scripts present"
else
  log_fail "Missing harness-setup scripts"
fi

# Test 4.1: setup-codex.sh avoids stale notify config
log_test "setup-codex.sh avoids stale notify config"
if grep -q '^\[notify\]' "skills/harness-setup/scripts/setup-codex.sh" 2>/dev/null; then
  echo "  stale [notify] section remains in setup-codex.sh"
  log_fail "setup-codex.sh still injects invalid notify config"
else
  log_pass "setup-codex.sh does not inject stale notify config"
fi

# Test 5: Core Harness skills exist in skills/ SSOT
log_test "Core Harness skills exist in skills/"
required_skills=(
  "skills/harness-plan"
  "skills/harness-sync"
  "skills/harness-work"
  "skills/harness-review"
  "skills/harness-release"
  "skills/harness-setup"
  "skills/breezing"
)
skills_ok=true
for dir in "${required_skills[@]}"; do
  if [ -d "$dir" ] && [ -f "$dir/SKILL.md" ]; then
    echo "  ok: $dir"
  else
    echo "  missing: $dir"
    skills_ok=false
  fi
done
if $skills_ok; then
  log_pass "Core Harness skills present in skills/"
else
  log_fail "Missing core Harness skills"
fi

# Test 6: All skills/ have required description frontmatter
log_test "All skills/ have required description frontmatter"
skill_frontmatter_ok=true
while IFS= read -r skill_file; do
  if ! grep -q '^description:' "$skill_file"; then
    echo "  missing description: $skill_file"
    skill_frontmatter_ok=false
  fi
done < <(find skills -name SKILL.md | sort)
if $skill_frontmatter_ok; then
  log_pass "All skills have description frontmatter"
else
  log_fail "Some skills are missing description frontmatter"
fi

# Test 7: Codex-native breezing uses spawn_agent API
log_test "Codex-native breezing uses spawn_agent API"
if grep -q 'spawn_agent' "templates/codex-skills/breezing/SKILL.md" 2>/dev/null; then
  log_pass "templates/codex-skills/breezing uses spawn_agent"
else
  log_fail "templates/codex-skills/breezing missing spawn_agent (should be Codex-native)"
fi

# Summary
if [ "$FAILED" -eq 0 ]; then
  echo "All tests passed: $PASSED"
  exit 0
fi

echo "Tests failed: $FAILED (passed: $PASSED)"
exit 1
