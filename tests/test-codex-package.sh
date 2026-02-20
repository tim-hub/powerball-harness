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

# Test 1.6: codex multi-agent migration guardrails
log_test "Codex multi-agent migration vocabulary checks"
targets=(
  "codex/.codex/skills/breezing"
  "codex/.codex/skills/work"
)
forbidden=(
  "delegate mode"
  "TaskCreate"
  "subagent_type"
  "CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS"
  ".claude/state"
)

forbidden_hit=false
for pat in "${forbidden[@]}"; do
  if rg -n --fixed-strings "$pat" "${targets[@]}" >/tmp/codex-forbidden.$$ 2>/dev/null; then
    echo "  forbidden pattern found: $pat"
    head -5 /tmp/codex-forbidden.$$ | sed 's/^/    /'
    forbidden_hit=true
  fi
done
rm -f /tmp/codex-forbidden.$$ || true

if $forbidden_hit; then
  log_fail "Forbidden legacy vocabulary remains in Codex skills"
else
  log_pass "No forbidden legacy vocabulary in Codex skills"
fi

log_test "Codex native multi-agent keywords present"
required_keywords=(
  "spawn_agent"
  "wait"
  "send_input"
  "resume_agent"
  "close_agent"
)
missing_keyword=false
for kw in "${required_keywords[@]}"; do
  if rg -q --fixed-strings "$kw" "${targets[@]}"; then
    echo "  ok: $kw"
  else
    echo "  missing: $kw"
    missing_keyword=true
  fi
done
if $missing_keyword; then
  log_fail "Missing codex native multi-agent keywords"
else
  log_pass "Codex native multi-agent keywords present"
fi

log_test "--claude review routing is fixed to Claude"
claude_review_check=true
if ! rg -q --fixed-strings '| `review_engine` | `codex` | `claude` |' "codex/.codex/skills/breezing/SKILL.md"; then
  echo "  missing review_engine matrix in breezing/SKILL.md"
  claude_review_check=false
fi
if ! rg -q --fixed-strings '| `review_engine` | `codex` | `claude` |' "codex/.codex/skills/work/SKILL.md"; then
  echo "  missing review_engine matrix in work/SKILL.md"
  claude_review_check=false
fi
if ! rg -q --fixed-strings -- "--claude + --codex-review" "codex/.codex/skills/breezing" "codex/.codex/skills/work"; then
  echo "  missing conflict rule: --claude + --codex-review"
  claude_review_check=false
fi
if $claude_review_check; then
  log_pass "--claude review routing checks passed"
else
  log_fail "--claude review routing checks failed"
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
