#!/bin/bash
# Plugin validation test for VibeCoder
# This script verifies that claude-code-harness is correctly configured

set -u
set -o pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PLUGIN_ROOT="$(dirname "$SCRIPT_DIR")"
HARNESS_ROOT="$PLUGIN_ROOT/harness"

echo "=========================================="
echo "Claude harness - Plugin validation test"
echo "=========================================="
echo ""

# Color output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

PASS_COUNT=0
FAIL_COUNT=0
WARN_COUNT=0

# Record test results
pass_test() {
    echo -e "${GREEN}✓${NC} $1"
    PASS_COUNT=$((PASS_COUNT + 1))
}

fail_test() {
    echo -e "${RED}✗${NC} $1"
    FAIL_COUNT=$((FAIL_COUNT + 1))
}

warn_test() {
    echo -e "${YELLOW}⚠${NC} $1"
    WARN_COUNT=$((WARN_COUNT + 1))
}

json_is_valid() {
    local file="$1"
    python3 - <<'PY' "$file" >/dev/null 2>&1
import json, sys
path = sys.argv[1]
with open(path, "r", encoding="utf-8") as f:
    json.load(f)
PY
}

json_has_key() {
    local file="$1"
    local key="$2"
    python3 - <<'PY' "$file" "$key" >/dev/null 2>&1
import json, sys
path, key = sys.argv[1], sys.argv[2]
with open(path, "r", encoding="utf-8") as f:
    data = json.load(f)
if key not in data:
    raise SystemExit(1)
PY
}

has_frontmatter_description() {
    local file="$1"
    # Check if frontmatter exists and contains a description:
    awk '
      NR==1 { if ($0 != "---") exit 1 }
      NR>1 && $0=="---" { exit 2 }  # End of frontmatter without description
      NR>1 && $0 ~ /^description:/ { exit 0 }
      NR>50 { exit 1 }              # Safety limit
    ' "$file"
}

echo "1. Plugin structure validation"
echo "----------------------------------------"

# Check marketplace.json existence
if [ -f "$PLUGIN_ROOT/.claude-plugin/marketplace.json" ]; then
    pass_test "marketplace.json exists"
else
    fail_test "marketplace.json not found"
    exit 1
fi

# Validate marketplace.json
if json_is_valid "$PLUGIN_ROOT/.claude-plugin/marketplace.json"; then
    pass_test "marketplace.json is valid JSON"
else
    fail_test "marketplace.json is invalid JSON"
    exit 1
fi

# Check required top-level fields
REQUIRED_FIELDS=("name" "plugins")
for field in "${REQUIRED_FIELDS[@]}"; do
    if json_has_key "$PLUGIN_ROOT/.claude-plugin/marketplace.json" "$field"; then
        pass_test "marketplace.json has $field field"
    else
        fail_test "marketplace.json is missing $field field"
    fi
done

echo ""
echo "2. Command validation (legacy)"
echo "----------------------------------------"

# v2.17.0+: Commands have been migrated to Skills
# Only validate if commands/ directory exists (backward compatibility)
if [ -d "$PLUGIN_ROOT/commands" ]; then
    CMD_COUNT=$(find "$PLUGIN_ROOT/commands" -name "*.md" -type f | wc -l | tr -d ' ')
    pass_test "commands/ has ${CMD_COUNT} command files (legacy)"

    # Show subdirectory structure
    for subdir in "$PLUGIN_ROOT/commands"/*/; do
        if [ -d "$subdir" ]; then
            subdir_name=$(basename "$subdir")
            subdir_count=$(find "$subdir" -name "*.md" -type f | wc -l | tr -d ' ')
            if [ "$subdir_count" -gt 0 ]; then
                pass_test "  └─ ${subdir_name}/ has ${subdir_count} commands"
            else
                warn_test "  └─ ${subdir_name}/ is empty (no command files)"
            fi
        fi
    done

    # Check frontmatter description existence (improves discoverability in SlashCommand tool / /help)
    MISSING_DESC=0
    while IFS= read -r cmd_file; do
        if has_frontmatter_description "$cmd_file"; then
            pass_test "frontmatter description: $(basename "$cmd_file")"
        else
            warn_test "frontmatter description not found: $(basename "$cmd_file")"
            MISSING_DESC=$((MISSING_DESC + 1))
        fi
    done < <(find "$PLUGIN_ROOT/commands" -name "*.md" -type f | sort)
else
    # v2.17.0+: commands/ not needed since migration to Skills
    pass_test "commands/ migrated to Skills (v2.17.0+)"
fi

echo ""
echo "3. Skill validation"
echo "----------------------------------------"

# Check for skills directory
if [ -d "$HARNESS_ROOT/skills" ]; then
    SKILL_COUNT=$(find "$HARNESS_ROOT/skills" -name "SKILL.md" | wc -l)
    pass_test "$SKILL_COUNT skills defined"

    # Check skill frontmatter (sample)
    SKILLS_WITH_DESCRIPTION=0
    SKILLS_WITH_ALLOWED_TOOLS=0

    find "$HARNESS_ROOT/skills" -name "SKILL.md" | while read -r skill_file; do
        if grep -q "^description:" "$skill_file"; then
            ((SKILLS_WITH_DESCRIPTION++))
        fi
        if grep -q "^allowed-tools:" "$skill_file"; then
            ((SKILLS_WITH_ALLOWED_TOOLS++))
        fi
    done

    if [ $SKILL_COUNT -gt 0 ]; then
        pass_test "Skill files are properly placed"
    fi
else
    warn_test "skills directory not found"
fi

echo ""
echo "4. Agent validation"
echo "----------------------------------------"

if [ -d "$HARNESS_ROOT/agents" ]; then
    AGENT_COUNT=$(find "$HARNESS_ROOT/agents" -name "*.md" | wc -l)
    if [ $AGENT_COUNT -gt 0 ]; then
        pass_test "$AGENT_COUNT agents defined"
    else
        warn_test "No agents defined"
    fi
else
    warn_test "agents directory not found"
fi

echo ""
echo "5. Hook validation"
echo "----------------------------------------"

if [ -f "$HARNESS_ROOT/hooks/hooks.json" ]; then
    if json_is_valid "$HARNESS_ROOT/hooks/hooks.json"; then
        pass_test "hooks.json is valid JSON"

        pass_test "hooks.json is readable"
    else
        fail_test "hooks.json is invalid JSON"
    fi
else
    warn_test "hooks.json not found"
fi

POST_TOOL_FAILURE="$HARNESS_ROOT/scripts/hook-handlers/post-tool-failure.sh"
if [ -f "$POST_TOOL_FAILURE" ]; then
    tmp_dir="$(mktemp -d "/tmp/harness-test.XXXXXX")"
    target_file="$tmp_dir/target.txt"
    mkdir -p "$tmp_dir/.claude/state"
    printf 'SAFE\n' > "$target_file"
    ln -s "$target_file" "$tmp_dir/.claude/state/tool-failure-counter.txt"

    hook_output="$(printf '{"tool_name":"Bash","error":"boom"}' | PROJECT_ROOT="$tmp_dir" bash "$POST_TOOL_FAILURE" 2>/dev/null || true)"
    target_after="$(cat "$target_file" 2>/dev/null || true)"

    if [ "$hook_output" = "{}" ] && [ "$target_after" = "SAFE" ]; then
        pass_test "post-tool-failure.sh does not overwrite symlink state file"
    else
        fail_test "post-tool-failure.sh symlink defense is insufficient"
    fi

    rm -rf "$tmp_dir"
fi

MEMORY_WRAPPERS=(
    "$HARNESS_ROOT/scripts/lib/harness-mem-bridge.sh"
    "$HARNESS_ROOT/scripts/hook-handlers/memory-bridge.sh"
    "$HARNESS_ROOT/scripts/hook-handlers/memory-session-start.sh"
    "$HARNESS_ROOT/scripts/hook-handlers/memory-user-prompt.sh"
    "$HARNESS_ROOT/scripts/hook-handlers/memory-post-tool-use.sh"
    "$HARNESS_ROOT/scripts/hook-handlers/memory-stop.sh"
    "$HARNESS_ROOT/scripts/hook-handlers/memory-codex-notify.sh"
)
for wrapper in "${MEMORY_WRAPPERS[@]}"; do
    if [ -f "$wrapper" ]; then
        pass_test "memory wrapper exists: $(basename "$wrapper")"
    else
        fail_test "memory wrapper not found: $wrapper"
    fi
done

if bash "$PLUGIN_ROOT/tests/test-memory-hook-wiring.sh" >/dev/null 2>&1; then
    pass_test "memory hook wiring is valid"
else
    fail_test "memory hook wiring consistency is broken"
fi

if bash "$PLUGIN_ROOT/tests/test-sync-plugin-cache.sh" >/dev/null 2>&1; then
    pass_test "sync-plugin-cache can sync memory wrappers to distribution cache"
else
    fail_test "sync-plugin-cache cannot sync memory wrappers to distribution cache"
fi

if bash "$PLUGIN_ROOT/tests/test-runtime-reactive-hooks.sh" >/dev/null 2>&1; then
    pass_test "reactive hook runtime (TaskCreated/FileChanged/CwdChanged) is working"
else
    fail_test "reactive hook runtime (TaskCreated/FileChanged/CwdChanged) has issues"
fi

if bash "$PLUGIN_ROOT/tests/test-claude-upstream-integration.sh" >/dev/null 2>&1; then
    pass_test "Claude Code 2.1.80-2.1.86 integration points are wired"
else
    fail_test "Claude Code 2.1.80-2.1.86 integration points have gaps"
fi

echo ""
echo "6. Script validation"
echo "----------------------------------------"

if [ -d "$HARNESS_ROOT/scripts" ]; then
    SCRIPT_COUNT=$(find "$HARNESS_ROOT/scripts" -name "*.sh" -type f | wc -l)
    if [ $SCRIPT_COUNT -gt 0 ]; then
        pass_test "$SCRIPT_COUNT script(s) exist"

        # Check execute permissions (GNU/BSD compatible: use -perm -111)
        EXECUTABLE_COUNT=$(find "$HARNESS_ROOT/scripts" -name "*.sh" -type f -perm -111 | wc -l | tr -d ' ')
        if [ $EXECUTABLE_COUNT -eq $SCRIPT_COUNT ]; then
            pass_test "All scripts have execute permissions"
        else
            warn_test "Some scripts do not have execute permissions ($EXECUTABLE_COUNT/$SCRIPT_COUNT)"
        fi
    else
        warn_test "No scripts found"
    fi
else
    warn_test "scripts directory not found"
fi

echo ""
echo "7. Documentation validation"
echo "----------------------------------------"

if [ -f "$PLUGIN_ROOT/README.md" ]; then
    README_SIZE=$(wc -c < "$PLUGIN_ROOT/README.md")
    if [ $README_SIZE -gt 1000 ]; then
        pass_test "README.md exists (${README_SIZE} bytes)"
    else
        warn_test "README.md is too brief (${README_SIZE} bytes)"
    fi
else
    fail_test "README.md not found"
fi

if [ -f "$PLUGIN_ROOT/IMPLEMENTATION_GUIDE.md" ]; then
    pass_test "IMPLEMENTATION_GUIDE.md exists"
else
    warn_test "IMPLEMENTATION_GUIDE.md not found (recommended)"
fi

echo ""
echo "8. Claude Code plugin validation (v2.1.77+)"
echo "----------------------------------------"

# Run only if the claude command is available
if command -v claude > /dev/null 2>&1; then
    # Check subcommand existence (plugin validate not available below v2.1.77)
    if claude plugin validate --help > /dev/null 2>&1; then
        if claude plugin validate "$PLUGIN_ROOT/.claude-plugin/marketplace.json" > /dev/null 2>&1; then
            pass_test "claude plugin validate passed"
        else
            fail_test "claude plugin validate detected errors (CC v2.1.77+ required)"
        fi
    else
        warn_test "claude plugin validate not supported (recommend updating to CC v2.1.77+)"
    fi
else
    warn_test "claude command not installed (skipping claude plugin validate)"
fi

echo ""
echo "9. Hardening parity validation"
echo "----------------------------------------"

HARDENING_DOC="$PLUGIN_ROOT/docs/hardening-parity.md"
HARDENING_CONTRACT="$HARNESS_ROOT/scripts/lib/codex-hardening-contract.txt"
if [ -f "$HARDENING_DOC" ]; then
    pass_test "hardening parity document exists"
else
    fail_test "docs/hardening-parity.md not found"
fi

if [ -f "$HARDENING_CONTRACT" ] && grep -q "HARNESS_HARDENING_CONTRACT_V1" "$HARDENING_CONTRACT"; then
    pass_test "Codex hardening contract template exists"
else
    fail_test "scripts/lib/codex-hardening-contract.txt not found"
fi

if grep -q "docs/hardening-parity.md" "$PLUGIN_ROOT/README.md"; then
    pass_test "README.md links to hardening parity document"
else
    fail_test "README.md has no link to hardening parity document"
fi

RULES_FILE="$PLUGIN_ROOT/go/internal/guardrail/rules.go"
RULE_IDS=(
    "R10:no-git-bypass-flags"
    "R11:no-reset-hard-protected-branch"
    "R12:deny-direct-push-protected-branch"
    "R13:warn-protected-review-paths"
)
for rule_id in "${RULE_IDS[@]}"; do
    if grep -q "$rule_id" "$RULES_FILE"; then
        pass_test "guardrail rule: $rule_id"
    else
        fail_test "guardrail rule not found: $rule_id"
    fi
done

CODEX_WRAPPER="$HARNESS_ROOT/scripts/codex/codex-exec-wrapper.sh"
if grep -q "codex-hardening-contract.txt" "$CODEX_WRAPPER"; then
    pass_test "Codex wrapper references the hardening contract template"
else
    fail_test "Codex wrapper does not reference the hardening contract template"
fi

CODEX_ENGINE="$HARNESS_ROOT/scripts/codex-worker-engine.sh"
if grep -q "codex-hardening-contract.txt" "$CODEX_ENGINE"; then
    pass_test "Codex worker engine references the hardening contract template"
else
    fail_test "Codex worker engine does not reference the hardening contract template"
fi

CODEX_GATE="$HARNESS_ROOT/scripts/codex-worker-quality-gate.sh"
if grep -q "gate_hardening()" "$CODEX_GATE" && grep -q '"hardening"' "$CODEX_GATE"; then
    pass_test "Codex quality gate has hardening parity check"
else
    fail_test "Codex quality gate has no hardening parity check"
fi

echo ""
echo "10. Migration residue check"
echo "----------------------------------------"

if bash "$PLUGIN_ROOT/local-scripts/check-residue.sh" > /dev/null 2>&1; then
    pass_test "No migration residue detected (local-scripts/check-residue.sh clean)"
else
    fail_test "Migration residue found — run 'bash local-scripts/check-residue.sh' to see details"
fi

echo ""
echo "11. Skill description format check"
echo "----------------------------------------"

AUDIT_SCRIPT="$PLUGIN_ROOT/local-scripts/audit-skill-descriptions.sh"
if [ ! -x "$AUDIT_SCRIPT" ]; then
    # Older branches may not have the audit script — warn instead of fail so CI
    # stays green during rollback or bisect.
    warn_test "audit-skill-descriptions.sh not found or not executable"
else
    # Capture violations (stdout only). Each line is tab-separated:
    #   <file>\t<kind>\t<snippet>
    # The script's summary goes to stderr and is dropped here.
    AUDIT_STDOUT="$(bash "$AUDIT_SCRIPT" 2>/dev/null)"
    AUDIT_EXIT=$?
    SKILL_COUNT="$(find "$HARNESS_ROOT/skills" "$HARNESS_ROOT/templates/codex-skills" -name SKILL.md 2>/dev/null | wc -l | tr -d ' ')"
    if [ "$AUDIT_EXIT" -eq 0 ]; then
        pass_test "All ${SKILL_COUNT} SKILL.md descriptions conform"
    else
        # One fail_test per violation — so per-file feedback survives into the
        # summary. DoD: deleting any 'Use when ' prefix must show that file's
        # path in the failure output.
        while IFS=$'\t' read -r v_file v_kind _v_snippet; do
            [ -n "$v_file" ] && fail_test "SKILL description: $v_file ($v_kind)"
        done <<<"$AUDIT_STDOUT"
    fi
fi

echo ""
echo "12. Skill path convention check"
echo "----------------------------------------"

# Flag bare relative script paths in bash code blocks inside SKILL.md files.
# Anchored paths (${CLAUDE_SKILL_DIR}/..., ${CLAUDE_PLUGIN_ROOT}/...) are fine.
# Bare paths like `bash skills/foo/scripts/bar.sh` break when CWD differs.
PATH_VIOLATIONS=""
while IFS= read -r skill_md; do
    # Look for bash invocations with bare skill-relative paths
    bad_lines=$(grep -nE '^\s*bash\s+"?(\./)?skills/' "$skill_md" 2>/dev/null || true)
    if [ -n "$bad_lines" ]; then
        PATH_VIOLATIONS="${PATH_VIOLATIONS}${skill_md}:\n${bad_lines}\n"
    fi
done < <(find "$HARNESS_ROOT/skills" -name "SKILL.md" 2>/dev/null | sort)

if [ -z "$PATH_VIOLATIONS" ]; then
    pass_test "No bare relative script paths in bash code blocks"
else
    echo -e "${PATH_VIOLATIONS}"
    fail_test "Bare relative script paths found — use \${CLAUDE_SKILL_DIR}/scripts/ instead"
fi

echo ""
echo "=========================================="
echo "Test Results Summary"
echo "=========================================="
echo -e "${GREEN}Passed:${NC} $PASS_COUNT"
echo -e "${YELLOW}Warnings:${NC} $WARN_COUNT"
echo -e "${RED}Failed:${NC} $FAIL_COUNT"
echo ""

if [ $FAIL_COUNT -eq 0 ]; then
    echo -e "${GREEN}✓ All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}✗ $FAIL_COUNT test(s) failed${NC}"
    exit 1
fi
