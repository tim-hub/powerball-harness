#!/bin/bash
# check-consistency.sh
# Plugin consistency check
#
# Usage: ./local-scripts/check-consistency.sh
# Exit codes:
#   0 - All checks passed
#   1 - Inconsistencies found

set -euo pipefail

PLUGIN_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
HARNESS_ROOT="$PLUGIN_ROOT/harness"
ERRORS=0

echo "🔍 claude-code-harness consistency check"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# ================================
# 1. Template file existence check
# ================================
echo ""
echo "📁 [1/13] Checking template file existence..."

REQUIRED_TEMPLATES=(
  "harness/templates/AGENTS.md.template"
  "harness/templates/CLAUDE.md.template"
  "harness/templates/Plans.md.template"
  "harness/templates/.claude-code-harness-version.template"
  "harness/templates/.claude-code-harness.config.yaml.template"
  "harness/templates/cursor/commands/start-session.md"
  "harness/templates/cursor/commands/project-overview.md"
  "harness/templates/cursor/commands/plan-with-cc.md"
  "harness/templates/cursor/commands/handoff-to-claude.md"
  "harness/templates/cursor/commands/review-cc-work.md"
  "harness/templates/claude/settings.security.json.template"
  "harness/templates/claude/settings.local.json.template"
  "harness/templates/rules/workflow.md.template"
  "harness/templates/rules/coding-standards.md.template"
  "harness/templates/rules/plans-management.md.template"
  "harness/templates/rules/testing.md.template"
  "harness/templates/rules/ui-debugging-agent-browser.md.template"
)

for template in "${REQUIRED_TEMPLATES[@]}"; do
  if [ ! -f "$PLUGIN_ROOT/$template" ]; then
    echo "  ❌ Missing: $template"
    ERRORS=$((ERRORS + 1))
  else
    echo "  ✅ $template"
  fi
done

# ================================
# 2. Command ↔ Skill consistency
# ================================
echo ""
echo "🔗 [2/13] Command ↔ Skill reference consistency..."

# Check if templates referenced by commands exist
check_command_references() {
  local cmd_file="$1"
  local cmd_name=$(basename "$cmd_file" .md)

  # Extract references to templates (handle both old templates/ and new harness/templates/)
  local refs=$(grep -oE '(harness/)?templates/[a-zA-Z0-9/_.-]+' "$cmd_file" 2>/dev/null || true)

  for ref in $refs; do
    # Normalize: add harness/ prefix if missing
    [[ "$ref" != harness/* ]] && ref="harness/$ref"
    if [ ! -e "$PLUGIN_ROOT/$ref" ] && [ ! -e "$PLUGIN_ROOT/${ref}.template" ]; then
      echo "  ❌ $cmd_name: Referenced path does not exist: $ref"
      ERRORS=$((ERRORS + 1))
    fi
  done
}

for cmd in "$PLUGIN_ROOT/commands"/*.md; do
  check_command_references "$cmd"
done
echo "  ✅ Command reference check complete"

# ================================
# 3. Version number consistency
# ================================
echo ""
echo "🏷️ [3/13] Version number consistency..."

VERSION_FILE="$HARNESS_ROOT/VERSION"
HARNESS_TOML="$HARNESS_ROOT/harness.toml"

if [ -f "$VERSION_FILE" ]; then
  FILE_VERSION=$(cat "$VERSION_FILE" | tr -d '[:space:]')

  if [ -f "$HARNESS_TOML" ]; then
    TOML_VERSION=$(grep '^version' "$HARNESS_TOML" | head -1 | sed 's/.*= "\([^"]*\)".*/\1/' || true)
    if [ "$FILE_VERSION" != "$TOML_VERSION" ]; then
      echo "  ❌ Version mismatch: VERSION=$FILE_VERSION, harness.toml=$TOML_VERSION"
      ERRORS=$((ERRORS + 1))
    else
      echo "  ✅ VERSION and harness.toml match: $FILE_VERSION"
    fi
  else
    echo "  ✅ VERSION: $FILE_VERSION (harness.toml not found, skipped)"
  fi
fi

LATEST_RELEASE_URL="https://github.com/tim-hub/powerball-harness/releases/latest"
LATEST_RELEASE_BADGE="https://img.shields.io/github/v/release/tim-hub/powerball-harness?display_name=tag&sort=semver"

# ================================
# 4. Expected skill file structure
# ================================
echo ""
echo "📋 [4/13] Expected skill definition file structure..."

# 2agent config is integrated into harness-setup
# Check for existence of skills/harness-setup/SKILL.md
SETUP_SKILL="$HARNESS_ROOT/skills/harness-setup/SKILL.md"
if [ -f "$SETUP_SKILL" ]; then
  echo "  ✅ harness/skills/harness-setup/SKILL.md exists (includes 2agent config)"
else
  echo "  ❌ harness/skills/harness-setup/SKILL.md not found"
  ERRORS=$((ERRORS + 1))
fi

# ================================
# 5. Hooks configuration consistency
# ================================
echo ""
echo "🪝 [5/13] Hooks configuration consistency..."

HOOKS_JSON="$HARNESS_ROOT/hooks/hooks.json"
if [ -f "$HOOKS_JSON" ]; then
  # Check script references in hooks.json
  SCRIPT_REFS=$(grep -oE '\$\{CLAUDE_PLUGIN_ROOT\}/scripts/[a-zA-Z0-9_./-]+' "$HOOKS_JSON" 2>/dev/null || true)

  for ref in $SCRIPT_REFS; do
    script_name=$(echo "$ref" | sed 's|\${CLAUDE_PLUGIN_ROOT}/scripts/||')
    if [ ! -f "$HARNESS_ROOT/scripts/$script_name" ]; then
      echo "  ❌ hooks.json: Script does not exist: harness/scripts/$script_name"
      ERRORS=$((ERRORS + 1))
    else
      echo "  ✅ harness/scripts/$script_name"
    fi
  done
fi

# ================================
# 6. /start-task deprecation regression check
# ================================
echo ""
echo "🚫 [6/13] /start-task deprecation regression check..."

# Operational flow files (excluding history like CHANGELOG)
START_TASK_TARGETS=(
  "harness/skills/"
  "harness/workflows/"
  "harness/templates/"
  "harness/scripts/"
  "README.md"
)

START_TASK_FOUND=0
for target in "${START_TASK_TARGETS[@]}"; do
  if [ -e "$PLUGIN_ROOT/$target" ]; then
    # Search for references to /start-task (excluding historical/explanatory context)
    # Exclusion patterns: deleted/deprecated/Removed (history), equivalent/integrated/legacy/absorbed (migration notes), improvements/usage distinction (CHANGELOG)
    REFS=$(grep -rn "/start-task" "$PLUGIN_ROOT/$target" 2>/dev/null \
      | grep -v "削除" | grep -v "廃止" | grep -v "Removed" \
      | grep -v "相当" | grep -v "統合" | grep -v "従来" | grep -v "吸収" \
      | grep -v "改善" | grep -v "使い分け" | grep -v "CHANGELOG" \
      | grep -v "check-consistency.sh" \
      || true)
    if [ -n "$REFS" ]; then
      echo "  ❌ /start-task reference still present: $target"
      echo "$REFS" | head -3 | sed 's/^/      /'
      START_TASK_FOUND=$((START_TASK_FOUND + 1))
    fi
  fi
done

if [ $START_TASK_FOUND -eq 0 ]; then
  echo "  ✅ No /start-task references (operational flow)"
else
  ERRORS=$((ERRORS + START_TASK_FOUND))
fi

# ================================
# 7. docs/ normalization regression check
# ================================
echo ""
echo "📁 [7/13] docs/ normalization regression check..."

# Check root-level references to proposal.md / priority_matrix.md
DOCS_TARGETS=(
  "harness/skills/"
)

DOCS_ISSUES=0
for target in "${DOCS_TARGETS[@]}"; do
  if [ -d "$PLUGIN_ROOT/$target" ]; then
    # Search for root-level references to proposal.md / technical-spec.md / priority_matrix.md
    # Detect those without docs/ prefix
    REFS=$(grep -rn "proposal.md\|technical-spec.md\|priority_matrix.md" "$PLUGIN_ROOT/$target" 2>/dev/null | grep -v "docs/" | grep -v "\.template" || true)
    if [ -n "$REFS" ]; then
      echo "  ❌ Reference without docs/ prefix: $target"
      echo "$REFS" | head -3 | sed 's/^/      /'
      DOCS_ISSUES=$((DOCS_ISSUES + 1))
    fi
  fi
done

if [ $DOCS_ISSUES -eq 0 ]; then
  echo "  ✅ docs/ normalization OK"
else
  ERRORS=$((ERRORS + DOCS_ISSUES))
fi

# ================================
# 8. bypassPermissions assumption regression check
# ================================
echo ""
echo "🔓 [8/13] bypassPermissions assumption regression check..."

BYPASS_ISSUES=0

SECURITY_TEMPLATE="$HARNESS_ROOT/templates/claude/settings.security.json.template"
if [ -f "$SECURITY_TEMPLATE" ]; then
  if grep -q "disableBypassPermissionsMode" "$SECURITY_TEMPLATE"; then
    echo "  ❌ disableBypassPermissionsMode still present in settings.security.json.template"
    echo "      Please remove this setting as bypassPermissions is assumed"
    BYPASS_ISSUES=$((BYPASS_ISSUES + 1))
  else
    echo "  ✅ disableBypassPermissionsMode not present"
  fi
fi

# Check 2: Edit / Write must not be in the permissions.ask section
# NOTE: Edit/Write in the deny section is legitimate as a double-defense. Only check ask section
if [ -f "$SECURITY_TEMPLATE" ]; then
  ASK_EDIT_WRITE=$(sed -n '/"ask"/,/\]/p' "$SECURITY_TEMPLATE" | grep -E '"(Edit|Write|MultiEdit)' || true)
  if [ -n "$ASK_EDIT_WRITE" ]; then
    echo "  ❌ settings.security.json.template ask section contains Edit/Write"
    echo "      Do not include Edit/Write in ask, as bypassPermissions is assumed"
    BYPASS_ISSUES=$((BYPASS_ISSUES + 1))
  else
    echo "  ✅ ask section has no Edit/Write"
  fi
fi

if [ -f "$SECURITY_TEMPLATE" ]; then
  # Portable regex: use [(] / [*] instead of escaping to avoid BSD grep issues.
  if grep -nEq 'Bash[(][^)]*[^:][*]' "$SECURITY_TEMPLATE"; then
    echo "  ❌ settings.security.json.template contains invalid Bash permission syntax"
    echo "      Use :* for prefix matching (e.g., Bash(git status:*))"
    grep -nE 'Bash[(][^)]*[^:][*]' "$SECURITY_TEMPLATE" | head -3 | sed 's/^/      /'
    BYPASS_ISSUES=$((BYPASS_ISSUES + 1))
  else
    echo "  ✅ Bash permission syntax OK (:*)"
  fi
fi

# Check 3: settings.local.json.template must exist and defaultMode must be a documented permission mode
# NOTE: shipped default keeps bypassPermissions, Auto Mode is treated as a follow-up rollout for the teammate execution path
LOCAL_TEMPLATE="$HARNESS_ROOT/templates/claude/settings.local.json.template"
if [ -f "$LOCAL_TEMPLATE" ]; then
  if grep -q '"defaultMode"[[:space:]]*:[[:space:]]*"bypassPermissions"' "$LOCAL_TEMPLATE"; then
    mode_val=$(grep '"defaultMode"' "$LOCAL_TEMPLATE" | head -1 | sed 's/.*: *"\([^"]*\)".*/\1/')
    echo "  ✅ settings.local.json.template: defaultMode=${mode_val}"
  else
    echo "  ❌ settings.local.json.template does not have defaultMode=bypassPermissions"
    BYPASS_ISSUES=$((BYPASS_ISSUES + 1))
  fi
else
  echo "  ❌ settings.local.json.template does not exist"
  BYPASS_ISSUES=$((BYPASS_ISSUES + 1))
fi

if [ $BYPASS_ISSUES -eq 0 ]; then
  echo "  ✅ bypassPermissions assumption OK"
else
  ERRORS=$((ERRORS + BYPASS_ISSUES))
fi

# ================================
# 9. ccp-* skill deprecation regression check
# ================================
echo ""
echo "🚫 [9/13] ccp-* skill deprecation regression check..."

CCP_ISSUES=0

# Check 1: skills must not have name: starting with ccp-
CCP_NAMES=$(grep -rn "^name: ccp-" "$HARNESS_ROOT/skills/" 2>/dev/null || true)
if [ -n "$CCP_NAMES" ]; then
  echo "  ❌ ccp-* name: still present in skills"
  echo "$CCP_NAMES" | head -3 | sed 's/^/      /'
  CCP_ISSUES=$((CCP_ISSUES + 1))
else
  echo "  ✅ No name: ccp-* in skills"
fi

# Check 2: workflows must not have skill: starting with ccp-
CCP_WORKFLOWS=$(grep -rn "skill: ccp-" "$HARNESS_ROOT/workflows/" 2>/dev/null || true)
if [ -n "$CCP_WORKFLOWS" ]; then
  echo "  ❌ ccp-* skill: still present in workflows"
  echo "$CCP_WORKFLOWS" | head -3 | sed 's/^/      /'
  CCP_ISSUES=$((CCP_ISSUES + 1))
else
  echo "  ✅ No skill: ccp-* in workflows"
fi

# Check 3: ccp-* directories must not remain
CCP_DIRS=$(find "$HARNESS_ROOT/skills" -type d -name "ccp-*" 2>/dev/null || true)
if [ -n "$CCP_DIRS" ]; then
  echo "  ❌ ccp-* directories still present"
  echo "$CCP_DIRS" | head -3 | sed 's/^/      /'
  CCP_ISSUES=$((CCP_ISSUES + 1))
else
  echo "  ✅ No ccp-* directories"
fi

if [ $CCP_ISSUES -eq 0 ]; then
  echo "  ✅ ccp-* skill deprecation OK"
else
  ERRORS=$((ERRORS + CCP_ISSUES))
fi

# ================================
# 10. Template existence check
# ================================
echo ""
echo "📦 [10/13] Template existence check..."

TEMPLATE_ISSUES=0

# Codex template
for f in "templates/codex/config.toml" "templates/codex/rules/harness.rules" "templates/codex/.codexignore" "templates/codex/AGENTS.md"; do
  if [ ! -f "$HARNESS_ROOT/$f" ]; then
    echo "  ❌ harness/$f not found"
    TEMPLATE_ISSUES=$((TEMPLATE_ISSUES + 1))
  fi
done

# OpenCode template
for f in "templates/opencode/opencode.json" "templates/opencode/AGENTS.md"; do
  if [ ! -f "$HARNESS_ROOT/$f" ]; then
    echo "  ❌ harness/$f not found"
    TEMPLATE_ISSUES=$((TEMPLATE_ISSUES + 1))
  fi
done
if [ ! -d "$HARNESS_ROOT/templates/opencode/commands" ]; then
  echo "  ❌ harness/templates/opencode/commands/ not found"
  TEMPLATE_ISSUES=$((TEMPLATE_ISSUES + 1))
fi

# Codex-native skill overrides
if [ ! -d "$HARNESS_ROOT/templates/codex-skills" ]; then
  echo "  ❌ harness/templates/codex-skills/ not found"
  TEMPLATE_ISSUES=$((TEMPLATE_ISSUES + 1))
fi

# Setup scripts
for script in "skills/harness-setup/scripts/setup-codex.sh" "skills/harness-setup/scripts/setup-opencode.sh"; do
  if [ ! -f "$HARNESS_ROOT/$script" ]; then
    echo "  ❌ harness/$script not found"
    TEMPLATE_ISSUES=$((TEMPLATE_ISSUES + 1))
  fi
done

if [ $TEMPLATE_ISSUES -eq 0 ]; then
  echo "  ✅ All templates and setup scripts present"
else
  ERRORS=$((ERRORS + TEMPLATE_ISSUES))
fi

# ================================
# 11. CHANGELOG format validation
# ================================
echo ""
echo "📝 [11/13] CHANGELOG format validation..."

CHANGELOG_ISSUES=0

for changelog in "$PLUGIN_ROOT/CHANGELOG.md" "$PLUGIN_ROOT/CHANGELOG_ja.md"; do
  if [ ! -f "$changelog" ]; then
    continue
  fi

  cl_name=$(basename "$changelog")

  # Check 1: Keep a Changelog header (## [x.y.z] - YYYY-MM-DD format)
  BAD_DATES=$(grep -nE '^\#\# \[[0-9]' "$changelog" | grep -vE '[0-9]{4}-[0-9]{2}-[0-9]{2}' | grep -v "Unreleased" || true)
  if [ -n "$BAD_DATES" ]; then
    echo "  ❌ $cl_name: Entry with non-ISO 8601 date"
    echo "$BAD_DATES" | head -3 | sed 's/^/      /'
    CHANGELOG_ISSUES=$((CHANGELOG_ISSUES + 1))
  fi

  # Check 2: Non-standard section headings (other than the 6 types in Keep a Changelog 1.1.0)
  NON_STANDARD=$(grep -nE '^\#\#\# ' "$changelog" \
    | grep -viE '(Added|Changed|Deprecated|Removed|Fixed|Security|What.*Changed|あなたにとって)' \
    | grep -viE '(Internal|Breaking|Migration|Summary|Before)' \
    || true)
  if [ -n "$NON_STANDARD" ]; then
    echo "  ⚠️ $cl_name: Non-standard section headings (review recommended)"
    echo "$NON_STANDARD" | head -3 | sed 's/^/      /'
    # Warning only (not an error)
  fi

  # Check 3: [Unreleased] section must exist
  if ! grep -q '^\#\# \[Unreleased\]' "$changelog"; then
    echo "  ❌ $cl_name: [Unreleased] section is missing"
    CHANGELOG_ISSUES=$((CHANGELOG_ISSUES + 1))
  fi
done

if [ $CHANGELOG_ISSUES -eq 0 ]; then
  echo "  ✅ CHANGELOG format OK"
else
  ERRORS=$((ERRORS + CHANGELOG_ISSUES))
fi

# ================================
# 12. README claim drift check
# ================================
echo ""
echo "📚 [12/13] README claim drift check..."

README_ISSUES=0
README_EN="$PLUGIN_ROOT/README.md"
SCOPE_DOC="$PLUGIN_ROOT/docs/distribution-scope.md"
RUBRIC_DOC="$PLUGIN_ROOT/docs/benchmark-rubric.md"
POSITIONING_DOC="$PLUGIN_ROOT/docs/positioning-notes.md"
WORK_ALL_DOC="$PLUGIN_ROOT/docs/evidence/work-all.md"

check_fixed_string() {
  local file_path="$1"
  local needle="$2"
  local label="$3"

  if [ ! -f "$file_path" ]; then
    echo "  ❌ ${label}: File does not exist: $file_path"
    README_ISSUES=$((README_ISSUES + 1))
    return
  fi

  if grep -qF "$needle" "$file_path"; then
    echo "  ✅ ${label}"
  else
    echo "  ❌ ${label}: Required string not found"
    README_ISSUES=$((README_ISSUES + 1))
  fi
}

check_absent_string() {
  local file_path="$1"
  local needle="$2"
  local label="$3"

  if [ ! -f "$file_path" ]; then
    echo "  ❌ ${label}: File does not exist: $file_path"
    README_ISSUES=$((README_ISSUES + 1))
    return
  fi

  if grep -qF "$needle" "$file_path"; then
    echo "  ❌ ${label}: Stale claim still present"
    README_ISSUES=$((README_ISSUES + 1))
  else
    echo "  ✅ ${label}"
  fi
}

check_exists() {
  local file_path="$1"
  local label="$2"

  if [ -f "$file_path" ]; then
    echo "  ✅ ${label}"
  else
    echo "  ❌ ${label}: File does not exist"
    README_ISSUES=$((README_ISSUES + 1))
  fi
}

check_fixed_string "$README_EN" "$LATEST_RELEASE_URL" "README.md latest release link"
check_fixed_string "$README_EN" "$LATEST_RELEASE_BADGE" "README.md latest release badge"

check_exists "$SCOPE_DOC" "distribution-scope.md"
check_exists "$RUBRIC_DOC" "benchmark-rubric.md"
check_exists "$POSITIONING_DOC" "positioning-notes.md"
check_exists "$WORK_ALL_DOC" "work-all evidence doc"

check_fixed_string "$README_EN" "docs/CLAUDE_CODE_COMPATIBILITY.md" "README.md compatibility doc link"
check_fixed_string "$README_EN" "docs/evidence/work-all.md" "README.md work-all evidence link"
check_fixed_string "$README_EN" "Go-native guardrail engine" "README.md Go-native guardrail engine message"
check_absent_string "$README_EN" "Production-ready code." "README.md stale production-ready wording"


check_fixed_string "$SCOPE_DOC" '| `commands/` | Compatibility-retained |' "distribution-scope commands classification"
check_fixed_string "$SCOPE_DOC" '| `mcp-server/` | Development-only and distribution-excluded |' "distribution-scope mcp-server classification"
check_fixed_string "$RUBRIC_DOC" "| Static evidence |" "benchmark-rubric static evidence"
check_fixed_string "$RUBRIC_DOC" "| Executed evidence |" "benchmark-rubric executed evidence"
check_fixed_string "$POSITIONING_DOC" "runtime enforcement" "positioning-notes runtime enforcement"

if [ $README_ISSUES -eq 0 ]; then
  echo "  ✅ README claim drift check OK"
else
  ERRORS=$((ERRORS + README_ISSUES))
fi

# ================================
# 13. EN/JA visual sync check
# ================================
echo ""
echo "🎨 [13/13] EN/JA visual sync check..."

VISUAL_EN_DIR="$PLUGIN_ROOT/docs/assets/readme-visuals-en/generated"
VISUAL_JA_DIR="$PLUGIN_ROOT/docs/assets/readme-visuals-ja/generated"
VISUAL_ISSUES=0

if [ -d "$VISUAL_EN_DIR" ] && [ -d "$VISUAL_JA_DIR" ]; then
  # Verify that files present in EN also exist in JA and that viewBox sizes match
  for en_svg in "$VISUAL_EN_DIR"/*.svg; do
    [ ! -f "$en_svg" ] && continue
    svg_name=$(basename "$en_svg")
    ja_svg="$VISUAL_JA_DIR/$svg_name"

    if [ ! -f "$ja_svg" ]; then
      echo "  ❌ JA version missing: $svg_name"
      VISUAL_ISSUES=$((VISUAL_ISSUES + 1))
      continue
    fi

    # Compare viewBox heights (detect significant structural divergence)
    en_viewbox=$(grep -o 'viewBox="[^"]*"' "$en_svg" | head -1)
    ja_viewbox=$(grep -o 'viewBox="[^"]*"' "$ja_svg" | head -1)
    if [ "$en_viewbox" != "$ja_viewbox" ]; then
      echo "  ⚠️ viewBox mismatch: $svg_name (EN: $en_viewbox / JA: $ja_viewbox)"
      # Warning only (height differences are acceptable since Japanese characters have different widths)
    fi

    # Compare table row counts (simple check using count of <rect y=)
    en_rows=$(grep -c '<rect y=' "$en_svg" 2>/dev/null || echo 0)
    ja_rows=$(grep -c '<rect y=' "$ja_svg" 2>/dev/null || echo 0)
    if [ "$en_rows" != "$ja_rows" ]; then
      echo "  ❌ Row count mismatch: $svg_name (EN: ${en_rows} rows / JA: ${ja_rows} rows)"
      VISUAL_ISSUES=$((VISUAL_ISSUES + 1))
    else
      echo "  ✅ $svg_name (${en_rows} rows)"
    fi
  done
else
  echo "  ⚠️ EN/JA visual directories not found (skipping)"
fi

if [ $VISUAL_ISSUES -gt 0 ]; then
  ERRORS=$((ERRORS + VISUAL_ISSUES))
fi

# ================================
# Result summary
# ================================
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [ $ERRORS -eq 0 ]; then
  echo "✅ All checks passed"
  exit 0
else
  echo "❌ $ERRORS issue(s) found"
  exit 1
fi
