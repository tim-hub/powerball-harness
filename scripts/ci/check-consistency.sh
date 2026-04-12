#!/bin/bash
# check-consistency.sh
# Plugin consistency check
#
# Usage: ./scripts/ci/check-consistency.sh
# Exit codes:
#   0 - All checks passed
#   1 - Inconsistencies found

set -euo pipefail

PLUGIN_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
ERRORS=0

echo "🔍 claude-code-harness Consistency Check"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# ================================
# 1. Template file existence check
# ================================
echo ""
echo "📁 [1/12] Checking template files..."

REQUIRED_TEMPLATES=(
  "templates/AGENTS.md.template"
  "templates/CLAUDE.md.template"
  "templates/Plans.md.template"
  "templates/.claude-code-harness-version.template"
  "templates/.claude-code-harness.config.yaml.template"
  "templates/cursor/commands/start-session.md"
  "templates/cursor/commands/project-overview.md"
  "templates/cursor/commands/plan-with-cc.md"
  "templates/cursor/commands/handoff-to-claude.md"
  "templates/cursor/commands/review-cc-work.md"
  "templates/claude/settings.security.json.template"
  "templates/claude/settings.local.json.template"
  "templates/rules/workflow.md.template"
  "templates/rules/coding-standards.md.template"
  "templates/rules/plans-management.md.template"
  "templates/rules/testing.md.template"
  "templates/rules/ui-debugging-agent-browser.md.template"
)

for template in "${REQUIRED_TEMPLATES[@]}"; do
  if [ ! -f "$PLUGIN_ROOT/$template" ]; then
    echo "  Missing: $template"
    ERRORS=$((ERRORS + 1))
  else
    echo "  ✅ $template"
  fi
done

# ================================
# 2. Command <-> Skill consistency
# ================================
echo ""
echo "🔗 [2/12] Command/skill reference consistency..."

# Check if templates referenced by commands exist
check_command_references() {
  local cmd_file="$1"
  local cmd_name=$(basename "$cmd_file" .md)

  # Extract template references
  local refs=$(grep -oE 'templates/[a-zA-Z0-9/_.-]+' "$cmd_file" 2>/dev/null || true)

  for ref in $refs; do
    if [ ! -e "$PLUGIN_ROOT/$ref" ] && [ ! -e "$PLUGIN_ROOT/${ref}.template" ]; then
      echo "  $cmd_name: referenced target does not exist: $ref"
      ERRORS=$((ERRORS + 1))
    fi
  done
}

for cmd in "$PLUGIN_ROOT/commands"/*.md; do
  check_command_references "$cmd"
done
echo "  Command reference check complete"

# ================================
# 3. Version number consistency
# ================================
echo ""
echo "🏷️ [3/12] Version number consistency..."

VERSION_FILE="$PLUGIN_ROOT/VERSION"
PLUGIN_JSON="$PLUGIN_ROOT/.claude-plugin/plugin.json"

if [ -f "$VERSION_FILE" ] && [ -f "$PLUGIN_JSON" ]; then
  FILE_VERSION=$(cat "$VERSION_FILE" | tr -d '[:space:]')
  JSON_VERSION=$(grep '"version"' "$PLUGIN_JSON" | head -1 | sed 's/.*: *"\([^"]*\)".*/\1/')

  if [ "$FILE_VERSION" != "$JSON_VERSION" ]; then
    echo "  Version mismatch: VERSION=$FILE_VERSION, plugin.json=$JSON_VERSION"
    ERRORS=$((ERRORS + 1))
  else
    echo "  VERSION and plugin.json match: $FILE_VERSION"
  fi
fi

LATEST_RELEASE_URL="https://github.com/tim-hub/powerball-harness/releases/latest"
LATEST_RELEASE_BADGE="https://img.shields.io/github/v/release/tim-hub/powerball-harness?display_name=tag&sort=semver"

# ================================
# 4. Expected skill file structure
# ================================
echo ""
echo "📋 [4/12] Expected skill file structure..."

# 2agent config integrated into harness-setup (v3)
# Check existence of skills-v3/harness-setup/SKILL.md
SETUP_V3="$PLUGIN_ROOT/skills-v3/harness-setup/SKILL.md"
if [ -f "$SETUP_V3" ]; then
  echo "  skills-v3/harness-setup/SKILL.md exists (includes 2-agent config)"
else
  echo "  skills-v3/harness-setup/SKILL.md not found"
  ERRORS=$((ERRORS + 1))
fi

# ================================
# 5. Hooks config consistency
# ================================
echo ""
echo "🪝 [5/12] Hooks configuration consistency..."

HOOKS_JSON="$PLUGIN_ROOT/hooks/hooks.json"
if [ -f "$HOOKS_JSON" ]; then
  # Check script references in hooks.json
  SCRIPT_REFS=$(grep -oE '\$\{CLAUDE_PLUGIN_ROOT\}/scripts/[a-zA-Z0-9_./-]+' "$HOOKS_JSON" 2>/dev/null || true)

  for ref in $SCRIPT_REFS; do
    script_name=$(echo "$ref" | sed 's|\${CLAUDE_PLUGIN_ROOT}/scripts/||')
    if [ ! -f "$PLUGIN_ROOT/scripts/$script_name" ]; then
      echo "  hooks.json: script does not exist: scripts/$script_name"
      ERRORS=$((ERRORS + 1))
    else
      echo "  ✅ scripts/$script_name"
    fi
  done
fi

# ================================
# 6. /start-task deprecation regression check
# ================================
echo ""
echo "🚫 [6/12] /start-task deprecation regression check..."

# Operational files (excluding CHANGELOG etc. history)
START_TASK_TARGETS=(
  "commands/"
  "skills/"
  "workflows/"
  "profiles/"
  "templates/"
  "scripts/"
  "DEVELOPMENT_FLOW_GUIDE.md"
  "IMPLEMENTATION_GUIDE.md"
  "README.md"
)

START_TASK_FOUND=0
for target in "${START_TASK_TARGETS[@]}"; do
  if [ -e "$PLUGIN_ROOT/$target" ]; then
    # Search for /start-task references (excluding history/explanation context)
    # Exclusion patterns: deleted/deprecated/Removed (history), equivalent/integrated/legacy/absorbed (migration), improvement/usage (CHANGELOG)
    # Exclude history/migration context: deleted, deprecated, removed, equivalent, integrated, legacy, absorbed, improvement, usage
    REFS=$(grep -rn "/start-task" "$PLUGIN_ROOT/$target" 2>/dev/null \
      | grep -v "deleted" | grep -v "deprecated" | grep -v "Removed" \
      | grep -v "equivalent" | grep -v "integrated" | grep -v "legacy" | grep -v "absorbed" \
      | grep -v "improvement" | grep -v "usage" | grep -v "CHANGELOG" \
      | grep -v "check-consistency.sh" \
      || true)
    if [ -n "$REFS" ]; then
      echo "  /start-task reference remains: $target"
      echo "$REFS" | head -3 | sed 's/^/      /'
      START_TASK_FOUND=$((START_TASK_FOUND + 1))
    fi
  fi
done

if [ $START_TASK_FOUND -eq 0 ]; then
  echo "  No /start-task references (operational)"
else
  ERRORS=$((ERRORS + START_TASK_FOUND))
fi

# ================================
# 7. docs/ normalization regression check
# ================================
echo ""
echo "📁 [7/12] docs/ normalization regression check..."

# Check root references to proposal.md / priority_matrix.md
DOCS_TARGETS=(
  "commands/"
  "skills/"
)

DOCS_ISSUES=0
for target in "${DOCS_TARGETS[@]}"; do
  if [ -d "$PLUGIN_ROOT/$target" ]; then
    # Search for references to root-level proposal.md / technical-spec.md / priority_matrix.md
    # Detect those missing docs/ prefix
    REFS=$(grep -rn "proposal.md\|technical-spec.md\|priority_matrix.md" "$PLUGIN_ROOT/$target" 2>/dev/null | grep -v "docs/" | grep -v "\.template" || true)
    if [ -n "$REFS" ]; then
      echo "  Reference without docs/ prefix: $target"
      echo "$REFS" | head -3 | sed 's/^/      /'
      DOCS_ISSUES=$((DOCS_ISSUES + 1))
    fi
  fi
done

if [ $DOCS_ISSUES -eq 0 ]; then
  echo "  docs/ normalization OK"
else
  ERRORS=$((ERRORS + DOCS_ISSUES))
fi

# ================================
# 8. bypassPermissions baseline regression check
# ================================
echo ""
echo "🔓 [8/12] bypassPermissions regression check..."

BYPASS_ISSUES=0

# Check 1: disableBypassPermissionsMode not re-introduced in templates
SECURITY_TEMPLATE="$PLUGIN_ROOT/templates/claude/settings.security.json.template"
if [ -f "$SECURITY_TEMPLATE" ]; then
  if grep -q "disableBypassPermissionsMode" "$SECURITY_TEMPLATE"; then
    echo "  disableBypassPermissionsMode remains in settings.security.json.template"
    echo "      This setting should be removed for bypassPermissions operation"
    BYPASS_ISSUES=$((BYPASS_ISSUES + 1))
  else
    echo "  ✅ No disableBypassPermissionsMode"
  fi
fi

# Check 2: Edit/Write not in permissions.ask
if [ -f "$SECURITY_TEMPLATE" ]; then
  if grep -q '"Edit' "$SECURITY_TEMPLATE" || grep -q '"Write' "$SECURITY_TEMPLATE"; then
    echo "  Edit/Write found in ask of settings.security.json.template"
    echo "      Edit/Write should not be in ask for bypassPermissions operation"
    BYPASS_ISSUES=$((BYPASS_ISSUES + 1))
  else
    echo "  ✅ No Edit/Write in ask"
  fi
fi

# Check 2.5: Bash permission syntax regression check (prefix requires :*)
if [ -f "$SECURITY_TEMPLATE" ]; then
  # Portable regex: use [(] / [*] instead of escaping to avoid BSD grep issues.
  if grep -nEq 'Bash[(][^)]*[^:][*]' "$SECURITY_TEMPLATE"; then
    echo "  Invalid Bash permission syntax in settings.security.json.template"
    echo "      Use :* for prefix matching (e.g., Bash(git status:*))"
    grep -nE 'Bash[(][^)]*[^:][*]' "$SECURITY_TEMPLATE" | head -3 | sed 's/^/      /'
    BYPASS_ISSUES=$((BYPASS_ISSUES + 1))
  else
    echo "  ✅ Bash permission syntax OK (:*)"
  fi
fi

# Check 3: settings.local.json.template exists with documented permission mode as defaultMode
# NOTE: shipped default maintains bypassPermissions; Auto Mode is follow-up rollout for teammate execution path
LOCAL_TEMPLATE="$PLUGIN_ROOT/templates/claude/settings.local.json.template"
if [ -f "$LOCAL_TEMPLATE" ]; then
  if grep -q '"defaultMode"[[:space:]]*:[[:space:]]*"bypassPermissions"' "$LOCAL_TEMPLATE"; then
    mode_val=$(grep '"defaultMode"' "$LOCAL_TEMPLATE" | head -1 | sed 's/.*: *"\([^"]*\)".*/\1/')
    echo "  ✅ settings.local.json.template: defaultMode=${mode_val}"
  else
    echo "  defaultMode=bypassPermissions not found in settings.local.json.template"
    BYPASS_ISSUES=$((BYPASS_ISSUES + 1))
  fi
else
  echo "  settings.local.json.template does not exist"
  BYPASS_ISSUES=$((BYPASS_ISSUES + 1))
fi

if [ $BYPASS_ISSUES -eq 0 ]; then
  echo "  bypassPermissions operation OK"
else
  ERRORS=$((ERRORS + BYPASS_ISSUES))
fi

# ================================
# 9. ccp-* skill deprecation regression check
# ================================
echo ""
echo "🚫 [9/12] ccp-* skill deprecation regression check..."

CCP_ISSUES=0

# Check 1: ccp- not in skill names
CCP_NAMES=$(grep -rn "^name: ccp-" "$PLUGIN_ROOT/skills/" 2>/dev/null || true)
if [ -n "$CCP_NAMES" ]; then
  echo "  name: ccp-* remains in skills"
  echo "$CCP_NAMES" | head -3 | sed 's/^/      /'
  CCP_ISSUES=$((CCP_ISSUES + 1))
else
  echo "  No name: ccp-* in skills"
fi

# Check 2: ccp- not in workflow skill references
CCP_WORKFLOWS=$(grep -rn "skill: ccp-" "$PLUGIN_ROOT/workflows/" 2>/dev/null || true)
if [ -n "$CCP_WORKFLOWS" ]; then
  echo "  skill: ccp-* remains in workflows"
  echo "$CCP_WORKFLOWS" | head -3 | sed 's/^/      /'
  CCP_ISSUES=$((CCP_ISSUES + 1))
else
  echo "  No skill: ccp-* in workflows"
fi

# Check 3: No ccp-* directories remaining
CCP_DIRS=$(find "$PLUGIN_ROOT/skills" -type d -name "ccp-*" 2>/dev/null || true)
if [ -n "$CCP_DIRS" ]; then
  echo "  ccp-* directories remain"
  echo "$CCP_DIRS" | head -3 | sed 's/^/      /'
  CCP_ISSUES=$((CCP_ISSUES + 1))
else
  echo "  No ccp-* directories"
fi

if [ $CCP_ISSUES -eq 0 ]; then
  echo "  ccp-* skill deprecation OK"
else
  ERRORS=$((ERRORS + CCP_ISSUES))
fi

# ================================
# 10. v3 skill mirror check
# ================================
echo ""
echo "📦 [10/12] v3 skill mirror check..."

V3_SKILLS_DIR="$PLUGIN_ROOT/skills-v3"
CLAUDE_MIRROR="$PLUGIN_ROOT/skills"
CODEX_MIRROR="$PLUGIN_ROOT/codex/.codex/skills"
OPENCODE_MIRROR="$PLUGIN_ROOT/opencode/skills"
MIRROR_ISSUES=0

# v3 core skills (5-verb harness- prefix) mirror check
V3_CORE_SKILLS="harness-plan harness-work harness-review harness-release harness-setup"
V3_AUX_SKILLS="harness-sync"

if [ -d "$V3_SKILLS_DIR" ]; then
  for skill in $V3_CORE_SKILLS; do
    src="$V3_SKILLS_DIR/$skill"
    for mirror_name in claude codex opencode; do
      case "$mirror_name" in
        claude) mirror_root="$CLAUDE_MIRROR" ;;
        codex) mirror_root="$CODEX_MIRROR" ;;
        opencode) mirror_root="$OPENCODE_MIRROR" ;;
      esac

      if [ ! -d "$mirror_root" ]; then
        continue
      fi

      mirror_path="$mirror_root/$skill"
      if [ ! -d "$mirror_path" ]; then
        echo "  $mirror_name: $skill does not exist as directory"
        MIRROR_ISSUES=$((MIRROR_ISSUES + 1))
        continue
      fi

      if [ -L "$mirror_path" ]; then
        echo "  $mirror_name: $skill is still a symlink"
        MIRROR_ISSUES=$((MIRROR_ISSUES + 1))
        continue
      fi

      if diff -qr "$src" "$mirror_path" >/dev/null 2>&1; then
        echo "  ✅ $mirror_name: $skill mirror is in sync"
      else
        echo "  $mirror_name: $skill mirror is out of sync with skills-v3"
        MIRROR_ISSUES=$((MIRROR_ISSUES + 1))
      fi
    done
  done

  for skill in $V3_AUX_SKILLS; do
    src="$V3_SKILLS_DIR/$skill"
    for mirror_name in claude codex opencode; do
      case "$mirror_name" in
        claude) mirror_root="$CLAUDE_MIRROR" ;;
        codex) mirror_root="$CODEX_MIRROR" ;;
        opencode) mirror_root="$OPENCODE_MIRROR" ;;
      esac

      if [ ! -d "$mirror_root" ]; then
        continue
      fi

      mirror_path="$mirror_root/$skill"
      if [ ! -d "$mirror_path" ]; then
        echo "  $mirror_name: $skill does not exist as directory"
        MIRROR_ISSUES=$((MIRROR_ISSUES + 1))
        continue
      fi

      if [ -L "$mirror_path" ]; then
        echo "  $mirror_name: $skill is still a symlink"
        MIRROR_ISSUES=$((MIRROR_ISSUES + 1))
        continue
      fi

      if diff -qr "$src" "$mirror_path" >/dev/null 2>&1; then
        echo "  ✅ $mirror_name: $skill mirror is in sync"
      else
        echo "  ❌ $mirror_name: $skill mirror drifted from skills-v3/$skill"
        MIRROR_ISSUES=$((MIRROR_ISSUES + 1))
      fi
    done
  done
else
  echo "  skills-v3/ does not exist (skipped)"
fi

if [ $MIRROR_ISSUES -gt 0 ]; then
  ERRORS=$((ERRORS + MIRROR_ISSUES))
fi

# breezing alias must match skills-v3 in both public mirror and codex mirror
for mirror_entry in "claude:$CLAUDE_MIRROR/breezing" "codex:$CODEX_MIRROR/breezing"; do
  mirror_name="${mirror_entry%%:*}"
  mirror_path="${mirror_entry#*:}"
  if [ ! -d "$mirror_path" ]; then
    echo "  $mirror_name: breezing does not exist as directory"
    ERRORS=$((ERRORS + 1))
    continue
  fi

  if [ -L "$mirror_path" ]; then
    echo "  $mirror_name: breezing is still a symlink"
    ERRORS=$((ERRORS + 1))
    continue
  fi

  if diff -qr "$V3_SKILLS_DIR/breezing" "$mirror_path" >/dev/null 2>&1; then
    echo "  ✅ $mirror_name: breezing mirror is in sync"
  else
    echo "  $mirror_name: breezing mirror is out of sync with skills-v3"
    ERRORS=$((ERRORS + 1))
  fi
done

# ================================
# 11. CHANGELOG format validation
# ================================
echo ""
echo "📝 [11/12] CHANGELOG format validation..."

CHANGELOG_ISSUES=0

for changelog in "$PLUGIN_ROOT/CHANGELOG.md" "$PLUGIN_ROOT/CHANGELOG_ja.md"; do
  if [ ! -f "$changelog" ]; then
    continue
  fi

  cl_name=$(basename "$changelog")

  # Check 1: Keep a Changelog header (## [x.y.z] - YYYY-MM-DD format)
  BAD_DATES=$(grep -nE '^\#\# \[[0-9]' "$changelog" | grep -vE '[0-9]{4}-[0-9]{2}-[0-9]{2}' | grep -v "Unreleased" || true)
  if [ -n "$BAD_DATES" ]; then
    echo "  $cl_name: entries with non-ISO 8601 dates"
    echo "$BAD_DATES" | head -3 | sed 's/^/      /'
    CHANGELOG_ISSUES=$((CHANGELOG_ISSUES + 1))
  fi

  # Check 2: non-standard section headings (outside Keep a Changelog 1.1.0 six types)
  NON_STANDARD=$(grep -nE '^\#\#\# ' "$changelog" \
    | grep -viE '(Added|Changed|Deprecated|Removed|Fixed|Security|What.*Changed|What.*Changed.*for.*You)' \
    | grep -viE '(Internal|Breaking|Migration|Summary|Before)' \
    || true)
  if [ -n "$NON_STANDARD" ]; then
    echo "  $cl_name: non-standard section headings (review recommended)"
    echo "$NON_STANDARD" | head -3 | sed 's/^/      /'
    # Warning only (not an error)
  fi

  # Check 3: Does [Unreleased] section exist
  if ! grep -q '^\#\# \[Unreleased\]' "$changelog"; then
    echo "  $cl_name: [Unreleased] section not found"
    CHANGELOG_ISSUES=$((CHANGELOG_ISSUES + 1))
  fi
done

if [ $CHANGELOG_ISSUES -eq 0 ]; then
  echo "  CHANGELOG format OK"
else
  ERRORS=$((ERRORS + CHANGELOG_ISSUES))
fi

# ================================
# 12. README claim drift check
# ================================
echo ""
echo "📚 [12/12] README claim drift check..."

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
    echo "  ${label}: file does not exist: $file_path"
    README_ISSUES=$((README_ISSUES + 1))
    return
  fi

  if grep -qF "$needle" "$file_path"; then
    echo "  ✅ ${label}"
  else
    echo "  ${label}: required string not found"
    README_ISSUES=$((README_ISSUES + 1))
  fi
}

check_absent_string() {
  local file_path="$1"
  local needle="$2"
  local label="$3"

  if [ ! -f "$file_path" ]; then
    echo "  ${label}: file does not exist: $file_path"
    README_ISSUES=$((README_ISSUES + 1))
    return
  fi

  if grep -qF "$needle" "$file_path"; then
    echo "  ${label}: stale claim remains"
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
    echo "  ${label}: file does not exist"
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
check_fixed_string "$README_EN" "docs/CURSOR_INTEGRATION.md" "README.md cursor doc link"
check_fixed_string "$README_EN" "docs/evidence/work-all.md" "README.md work-all evidence link"
check_fixed_string "$README_EN" "docs/distribution-scope.md" "README.md distribution scope link"
check_fixed_string "$README_EN" "5 verb skills" "README.md 5 verb skills message"
check_fixed_string "$README_EN" "TypeScript guardrail engine" "README.md TypeScript guardrail engine message"
check_absent_string "$README_EN" "Production-ready code." "README.md stale production-ready wording"

check_fixed_string "$SCOPE_DOC" '| `commands/` | Compatibility-retained |' "distribution-scope commands classification"
check_fixed_string "$SCOPE_DOC" '| `mcp-server/` | Development-only and distribution-excluded |' "distribution-scope mcp-server classification"
check_fixed_string "$RUBRIC_DOC" "| Static evidence |" "benchmark-rubric static evidence"
check_fixed_string "$RUBRIC_DOC" "| Executed evidence |" "benchmark-rubric executed evidence"
check_fixed_string "$POSITIONING_DOC" "runtime enforcement" "positioning-notes runtime enforcement"

if [ $README_ISSUES -eq 0 ]; then
  echo "  README claim drift check OK"
else
  ERRORS=$((ERRORS + README_ISSUES))
fi

# ================================
# Result summary
# ================================
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [ $ERRORS -eq 0 ]; then
  echo "All checks passed"
  exit 0
else
  echo "$ERRORS issues found"
  exit 1
fi
