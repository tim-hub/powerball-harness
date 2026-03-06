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

# Test 1.5: execpolicy rules examples are consistent (prevents Codex startup parse errors)
log_test "Execpolicy rules examples are valid"
if command -v python3 >/dev/null 2>&1; then
  if python3 - <<'PY'
from __future__ import annotations

import shlex
import sys
from pathlib import Path


def _matches_prefix(pattern: list[object], tokens: list[str]) -> bool:
    if len(tokens) < len(pattern):
        return False

    for i, pe in enumerate(pattern):
        t = tokens[i]
        if isinstance(pe, str):
            if t != pe:
                return False
        elif isinstance(pe, (list, tuple)):
            if t not in pe:
                return False
        else:
            raise TypeError(f"Unsupported pattern element at index {i}: {pe!r}")
    return True


def _load_rules(path: Path) -> list[dict[str, object]]:
    rules: list[dict[str, object]] = []

    def prefix_rule(**kwargs):  # type: ignore[no-redef]
        rules.append(kwargs)

    g = {"prefix_rule": prefix_rule}
    code = path.read_text(encoding="utf-8")
    exec(compile(code, str(path), "exec"), g, {})
    return rules


def _validate(path: Path) -> list[str]:
    errs: list[str] = []
    rules = _load_rules(path)
    if not rules:
        return [f"{path}: no prefix_rule() found"]

    for idx, rule in enumerate(rules):
        pattern = rule.get("pattern")
        if not isinstance(pattern, list):
            errs.append(f"{path}: rule {idx} missing/invalid pattern: {pattern!r}")
            continue

        for field, should_match in (("match", True), ("not_match", False)):
            examples = rule.get(field, [])
            if examples is None:
                continue
            if not isinstance(examples, list):
                errs.append(f"{path}: rule {idx} {field} is not a list: {examples!r}")
                continue

            for ex in examples:
                if not isinstance(ex, str):
                    errs.append(f"{path}: rule {idx} {field} example is not str: {ex!r}")
                    continue
                tokens = shlex.split(ex)
                ok = _matches_prefix(pattern, tokens)
                if ok != should_match:
                    verdict = "matches" if ok else "does not match"
                    errs.append(
                        f"{path}: rule {idx} {field} example {ex!r} {verdict} pattern {pattern!r}"
                    )
    return errs


errors: list[str] = []
for p in [Path("codex/.codex/rules/harness.rules")]:
    errors.extend(_validate(p))

if errors:
    print("ERROR: execpolicy rules examples invalid:")
    for e in errors:
        print("  -", e)
    sys.exit(1)

print("ok")
PY
  then
    log_pass "Rules examples are consistent"
  else
    log_fail "Rules examples invalid (Codex may ignore custom rules)"
  fi
else
  echo "  skipped: python3 not found"
  log_pass "Rules examples check skipped"
fi

# Test 1.6: path-based skill bundle is present
log_test "Codex path-based core skills exist"
required_skill_dirs=(
  "codex/.codex/skills/harness-plan"
  "codex/.codex/skills/harness-work"
  "codex/.codex/skills/harness-review"
  "codex/.codex/skills/harness-release"
  "codex/.codex/skills/harness-setup"
  "codex/.codex/skills/breezing"
)
skills_ok=true
for dir in "${required_skill_dirs[@]}"; do
  if [ -e "$dir" ]; then
    echo "  ok: $dir"
  else
    echo "  missing: $dir"
    skills_ok=false
  fi
done
if $skills_ok; then
  log_pass "Path-based core skills are present"
else
  log_fail "Missing path-based core skills"
fi

log_test "Non-breezing Codex skills are CLI-only"
cli_only_targets=(
  "codex/.codex/skills/harness-work/SKILL.md"
  "codex/.codex/skills/harness-review/SKILL.md"
  "codex/.codex/skills/routing-rules.md"
)
forbidden_cli_terms=(
  "Codex MCP"
  "claude mcp add --scope user codex"
  "claude mcp list | grep -i codex"
  "@openai/codex-cli"
  "MCP server connection error"
  "all MCP calls"
)
cli_only_ok=true
for pat in "${forbidden_cli_terms[@]}"; do
  if rg -n --fixed-strings "$pat" "${cli_only_targets[@]}" >/tmp/codex-cli-only.$$ 2>/dev/null; then
    echo "  forbidden CLI-only pattern found: $pat"
    head -5 /tmp/codex-cli-only.$$ | sed 's/^/    /'
    cli_only_ok=false
  fi
done
rm -f /tmp/codex-cli-only.$$ || true
if $cli_only_ok; then
  log_pass "CLI-only vocabulary checks passed for non-breezing Codex skills"
else
  log_fail "CLI-only vocabulary check failed for non-breezing Codex skills"
fi

log_test "Codex docs point at harness-* workflow surfaces"
workflow_surface_ok=true
if ! rg -q --fixed-strings '$harness-work' "codex/README.md"; then
  echo "  missing: \$harness-work in codex/README.md"
  workflow_surface_ok=false
fi
if ! rg -q --fixed-strings '$harness-review' "codex/README.md"; then
  echo "  missing: \$harness-review in codex/README.md"
  workflow_surface_ok=false
fi
if ! rg -q --fixed-strings -- '--codex' "codex/.codex/skills/harness-work/SKILL.md"; then
  echo "  missing: --codex in harness-work/SKILL.md"
  workflow_surface_ok=false
fi
if ! rg -q --fixed-strings '/harness-work' "codex/.codex/skills/breezing/SKILL.md"; then
  echo "  missing: /harness-work alias note in breezing/SKILL.md"
  workflow_surface_ok=false
fi
if $workflow_surface_ok; then
  log_pass "Harness workflow surfaces are documented for Codex"
else
  log_fail "Harness workflow surface checks failed"
fi

log_test "codex/.codex/config.toml has multi_agent + harness roles"
config_ok=true
if ! rg -q --fixed-strings "multi_agent = true" "codex/.codex/config.toml"; then
  echo "  missing: multi_agent = true"
  config_ok=false
fi
for role in "implementer" "reviewer" "claude_implementer" "claude_reviewer"; do
  if ! rg -q --fixed-strings "[agents.${role}]" "codex/.codex/config.toml"; then
    echo "  missing: [agents.${role}]"
    config_ok=false
  fi
done
if $config_ok; then
  log_pass "config.toml has required multi-agent defaults"
else
  log_fail "config.toml missing required multi-agent defaults"
fi

# Test 1.7: setup scripts should not create duplicate skill listings
log_test "Codex setup scripts guard against duplicate skill listings"
scripts_ok=true
setup_scripts=(
  "scripts/setup-codex.sh"
  "scripts/codex-setup-local.sh"
)

for script in "${setup_scripts[@]}"; do
  if rg -q --fixed-strings '${target}.backup.' "$script"; then
    echo "  legacy in-place backup naming remains: $script"
    scripts_ok=false
  fi
  if ! rg -q --fixed-strings 'should_skip_sync_entry' "$script"; then
    echo "  missing should_skip_sync_entry: $script"
    scripts_ok=false
  fi
  if ! rg -q --fixed-strings 'cleanup_legacy_skill_entries' "$script"; then
    echo "  missing cleanup_legacy_skill_entries: $script"
    scripts_ok=false
  fi
  if ! rg -q --fixed-strings 'extract_skill_frontmatter_name' "$script"; then
    echo "  missing extract_skill_frontmatter_name: $script"
    scripts_ok=false
  fi
  if ! rg -q --fixed-strings 'cleanup_legacy_skill_name_duplicates' "$script"; then
    echo "  missing cleanup_legacy_skill_name_duplicates: $script"
    scripts_ok=false
  fi
  if ! rg -q --fixed-strings '_archived|*.backup.*' "$script"; then
    echo "  missing legacy skip rule (_archived|*.backup.*): $script"
    scripts_ok=false
  fi
  if ! rg -q --fixed-strings '/backups/' "$script"; then
    echo "  missing external backup root (/backups/): $script"
    scripts_ok=false
  fi
done

if $scripts_ok; then
  log_pass "Setup script duplicate-skill guards are present"
else
  log_fail "Setup script duplicate-skill guards are missing"
fi

# Test 1.8: codex-setup-local should cleanup duplicate frontmatter names
log_test "codex-setup-local cleans duplicate frontmatter skill aliases"
if PROJECT_ROOT="$PROJECT_ROOT" bash -lc '
set -euo pipefail
project_root="$PROJECT_ROOT"
tmp_home="$(mktemp -d)"
trap "rm -rf \"$tmp_home\"" EXIT

export HOME="$tmp_home"
export CODEX_HOME="$tmp_home/.codex"
mkdir -p "$CODEX_HOME/skills/legacy-harness-plan/references"
cat > "$CODEX_HOME/skills/legacy-harness-plan/SKILL.md" <<'"'"'EOF'"'"'
---
name: harness-plan
description: duplicate frontmatter name for migration test
allowed-tools: ["Read"]
---
EOF

CLAUDE_PLUGIN_ROOT="$project_root" bash "$project_root/scripts/codex-setup-local.sh" --user >/dev/null

test -f "$CODEX_HOME/skills/harness-plan/SKILL.md"
if [ -d "$CODEX_HOME/skills/legacy-harness-plan" ]; then
  echo "  duplicate alias directory still exists"
  exit 1
fi
if ! find "$CODEX_HOME/backups/codex-setup-local" -type d -name "legacy-harness-plan.*" | grep -q .; then
  echo "  duplicate alias backup not found"
  exit 1
fi
'; then
  log_pass "Duplicate frontmatter alias cleanup works"
else
  log_fail "Duplicate frontmatter alias cleanup failed"
fi

# Test 2: skills directory parity
log_test "Skills parity by SKILL name"
if [ -d "opencode/skills" ] && [ -d "codex/.codex/skills" ]; then
  get_skill_names() {
    local root="$1"
    find "$root" -mindepth 1 -maxdepth 1 -type d | while IFS= read -r d; do
      local dirname
      dirname="$(basename "$d")"
      # Skip dev/test/unsupported skills (matches build-opencode.js logic)
      case "$dirname" in
        test-*|x-*|breezing|_archived|harness-ui) continue ;;
      esac
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
  skill_name="$(basename "$skill_dir")"
  case "$skill_name" in
    _archived|harness-ui)
      # distribution-excluded buckets are allowed without SKILL.md
      continue
      ;;
  esac
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
