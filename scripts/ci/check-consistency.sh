#!/bin/bash
# check-consistency.sh
# プラグインの整合性チェック
#
# Usage: ./scripts/ci/check-consistency.sh
# Exit codes:
#   0 - All checks passed
#   1 - Inconsistencies found

set -euo pipefail

PLUGIN_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
ERRORS=0

echo "🔍 claude-code-harness 整合性チェック"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# ================================
# 1. テンプレートファイルの存在確認
# ================================
echo ""
echo "📁 [1/13] テンプレートファイルの存在確認..."

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
    echo "  ❌ 不足: $template"
    ERRORS=$((ERRORS + 1))
  else
    echo "  ✅ $template"
  fi
done

# ================================
# 2. コマンド ↔ スキル の整合性
# ================================
echo ""
echo "🔗 [2/13] コマンド ↔ スキル の参照整合性..."

# コマンドが参照するテンプレートが存在するか
check_command_references() {
  local cmd_file="$1"
  local cmd_name=$(basename "$cmd_file" .md)

  # テンプレートへの参照を抽出
  local refs=$(grep -oE 'templates/[a-zA-Z0-9/_.-]+' "$cmd_file" 2>/dev/null || true)

  for ref in $refs; do
    if [ ! -e "$PLUGIN_ROOT/$ref" ] && [ ! -e "$PLUGIN_ROOT/${ref}.template" ]; then
      echo "  ❌ $cmd_name: 参照先が存在しない: $ref"
      ERRORS=$((ERRORS + 1))
    fi
  done
}

for cmd in "$PLUGIN_ROOT/commands"/*.md; do
  check_command_references "$cmd"
done
echo "  ✅ コマンド参照チェック完了"

# ================================
# 3. バージョン番号の一貫性
# ================================
echo ""
echo "🏷️ [3/13] バージョン番号の一貫性..."

VERSION_FILE="$PLUGIN_ROOT/VERSION"
PLUGIN_JSON="$PLUGIN_ROOT/.claude-plugin/plugin.json"

if [ -f "$VERSION_FILE" ] && [ -f "$PLUGIN_JSON" ]; then
  FILE_VERSION=$(cat "$VERSION_FILE" | tr -d '[:space:]')
  JSON_VERSION=$(grep '"version"' "$PLUGIN_JSON" | head -1 | sed 's/.*: *"\([^"]*\)".*/\1/')

  if [ "$FILE_VERSION" != "$JSON_VERSION" ]; then
    echo "  ❌ バージョン不一致: VERSION=$FILE_VERSION, plugin.json=$JSON_VERSION"
    ERRORS=$((ERRORS + 1))
  else
    echo "  ✅ VERSION と plugin.json が一致: $FILE_VERSION"
  fi
fi

LATEST_RELEASE_URL="https://github.com/Chachamaru127/claude-code-harness/releases/latest"
LATEST_RELEASE_BADGE="https://img.shields.io/github/v/release/Chachamaru127/claude-code-harness?display_name=tag&sort=semver"

# ================================
# 4. スキルの期待ファイル構成
# ================================
echo ""
echo "📋 [4/13] スキル定義の期待ファイル構成..."

# 2agent 設定は harness-setup に統合済み
# skills/harness-setup/SKILL.md の存在を確認
SETUP_SKILL="$PLUGIN_ROOT/skills/harness-setup/SKILL.md"
if [ -f "$SETUP_SKILL" ]; then
  echo "  ✅ skills/harness-setup/SKILL.md が存在（2agent 設定を包含）"
else
  echo "  ❌ skills/harness-setup/SKILL.md が見つかりません"
  ERRORS=$((ERRORS + 1))
fi

# ================================
# 5. Hooks 設定の整合性
# ================================
echo ""
echo "🪝 [5/13] Hooks 設定の整合性..."

HOOKS_JSON="$PLUGIN_ROOT/hooks/hooks.json"
if [ -f "$HOOKS_JSON" ]; then
  # hooks.json 内のスクリプト参照を確認
  SCRIPT_REFS=$(grep -oE '\$\{CLAUDE_PLUGIN_ROOT\}/scripts/[a-zA-Z0-9_./-]+' "$HOOKS_JSON" 2>/dev/null || true)

  for ref in $SCRIPT_REFS; do
    script_name=$(echo "$ref" | sed 's|\${CLAUDE_PLUGIN_ROOT}/scripts/||')
    if [ ! -f "$PLUGIN_ROOT/scripts/$script_name" ]; then
      echo "  ❌ hooks.json: スクリプトが存在しない: scripts/$script_name"
      ERRORS=$((ERRORS + 1))
    else
      echo "  ✅ scripts/$script_name"
    fi
  done
fi

# ================================
# 6. /start-task 廃止の回帰チェック
# ================================
echo ""
echo "🚫 [6/13] /start-task 廃止の回帰チェック..."

# 運用導線ファイル（CHANGELOG等の履歴は除外）
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
    # /start-task への参照を検索（履歴・説明文脈は除外）
    # 除外パターン: 削除/廃止/Removed（履歴）, 相当/統合/従来/吸収（移行説明）, 改善/使い分け（CHANGELOG）
    REFS=$(grep -rn "/start-task" "$PLUGIN_ROOT/$target" 2>/dev/null \
      | grep -v "削除" | grep -v "廃止" | grep -v "Removed" \
      | grep -v "相当" | grep -v "統合" | grep -v "従来" | grep -v "吸収" \
      | grep -v "改善" | grep -v "使い分け" | grep -v "CHANGELOG" \
      | grep -v "check-consistency.sh" \
      || true)
    if [ -n "$REFS" ]; then
      echo "  ❌ /start-task 参照が残存: $target"
      echo "$REFS" | head -3 | sed 's/^/      /'
      START_TASK_FOUND=$((START_TASK_FOUND + 1))
    fi
  fi
done

if [ $START_TASK_FOUND -eq 0 ]; then
  echo "  ✅ /start-task 参照なし（運用導線）"
else
  ERRORS=$((ERRORS + START_TASK_FOUND))
fi

# ================================
# 7. docs/ 正規化の回帰チェック
# ================================
echo ""
echo "📁 [7/13] docs/ 正規化の回帰チェック..."

# proposal.md / priority_matrix.md のルート参照をチェック
DOCS_TARGETS=(
  "commands/"
  "skills/"
)

DOCS_ISSUES=0
for target in "${DOCS_TARGETS[@]}"; do
  if [ -d "$PLUGIN_ROOT/$target" ]; then
    # ルート直下の proposal.md / technical-spec.md / priority_matrix.md への参照を検索
    # docs/ プレフィックスがないものを検出
    REFS=$(grep -rn "proposal.md\|technical-spec.md\|priority_matrix.md" "$PLUGIN_ROOT/$target" 2>/dev/null | grep -v "docs/" | grep -v "\.template" || true)
    if [ -n "$REFS" ]; then
      echo "  ❌ docs/ プレフィックスなしの参照: $target"
      echo "$REFS" | head -3 | sed 's/^/      /'
      DOCS_ISSUES=$((DOCS_ISSUES + 1))
    fi
  fi
done

if [ $DOCS_ISSUES -eq 0 ]; then
  echo "  ✅ docs/ 正規化OK"
else
  ERRORS=$((ERRORS + DOCS_ISSUES))
fi

# ================================
# 8. bypassPermissions 前提運用の回帰チェック
# ================================
echo ""
echo "🔓 [8/13] bypassPermissions 前提運用の回帰チェック..."

BYPASS_ISSUES=0

# Check 1: disableBypassPermissionsMode が templates に戻っていないこと
SECURITY_TEMPLATE="$PLUGIN_ROOT/templates/claude/settings.security.json.template"
if [ -f "$SECURITY_TEMPLATE" ]; then
  if grep -q "disableBypassPermissionsMode" "$SECURITY_TEMPLATE"; then
    echo "  ❌ settings.security.json.template に disableBypassPermissionsMode が残存"
    echo "      bypassPermissions 前提運用のため、この設定は削除してください"
    BYPASS_ISSUES=$((BYPASS_ISSUES + 1))
  else
    echo "  ✅ disableBypassPermissionsMode なし"
  fi
fi

# Check 2: permissions.ask セクションに Edit / Write が入っていないこと
# NOTE: deny セクションの Edit/Write は二重防御として正当。ask のみをチェック
if [ -f "$SECURITY_TEMPLATE" ]; then
  # ask セクションのみ抽出して Edit/Write を検索
  ASK_EDIT_WRITE=$(sed -n '/"ask"/,/\]/p' "$SECURITY_TEMPLATE" | grep -E '"(Edit|Write|MultiEdit)' || true)
  if [ -n "$ASK_EDIT_WRITE" ]; then
    echo "  ❌ settings.security.json.template の ask に Edit/Write が含まれている"
    echo "      bypassPermissions 前提運用のため、Edit/Write は ask に入れないでください"
    BYPASS_ISSUES=$((BYPASS_ISSUES + 1))
  else
    echo "  ✅ ask に Edit/Write なし"
  fi
fi

# Check 2.5: Bash パーミッション構文の回帰チェック（prefix は :* 必須）
if [ -f "$SECURITY_TEMPLATE" ]; then
  # Portable regex: use [(] / [*] instead of escaping to avoid BSD grep issues.
  if grep -nEq 'Bash[(][^)]*[^:][*]' "$SECURITY_TEMPLATE"; then
    echo "  ❌ settings.security.json.template に不正な Bash パーミッション構文が含まれています"
    echo "      prefix マッチングは :* を使用してください（例: Bash(git status:*)）"
    grep -nE 'Bash[(][^)]*[^:][*]' "$SECURITY_TEMPLATE" | head -3 | sed 's/^/      /'
    BYPASS_ISSUES=$((BYPASS_ISSUES + 1))
  else
    echo "  ✅ Bash パーミッション構文OK (:*)"
  fi
fi

# Check 3: settings.local.json.template が存在し、defaultMode が documented な permission mode であること
# NOTE: shipped default は bypassPermissions を維持し、Auto Mode は teammate 実行経路の follow-up rollout として扱う
LOCAL_TEMPLATE="$PLUGIN_ROOT/templates/claude/settings.local.json.template"
if [ -f "$LOCAL_TEMPLATE" ]; then
  if grep -q '"defaultMode"[[:space:]]*:[[:space:]]*"bypassPermissions"' "$LOCAL_TEMPLATE"; then
    mode_val=$(grep '"defaultMode"' "$LOCAL_TEMPLATE" | head -1 | sed 's/.*: *"\([^"]*\)".*/\1/')
    echo "  ✅ settings.local.json.template: defaultMode=${mode_val}"
  else
    echo "  ❌ settings.local.json.template に defaultMode=bypassPermissions がありません"
    BYPASS_ISSUES=$((BYPASS_ISSUES + 1))
  fi
else
  echo "  ❌ settings.local.json.template が存在しません"
  BYPASS_ISSUES=$((BYPASS_ISSUES + 1))
fi

if [ $BYPASS_ISSUES -eq 0 ]; then
  echo "  ✅ bypassPermissions 前提運用OK"
else
  ERRORS=$((ERRORS + BYPASS_ISSUES))
fi

# ================================
# 9. ccp-* スキル廃止の回帰チェック
# ================================
echo ""
echo "🚫 [9/13] ccp-* スキル廃止の回帰チェック..."

CCP_ISSUES=0

# Check 1: skills の name: に ccp- が含まれていないこと
CCP_NAMES=$(grep -rn "^name: ccp-" "$PLUGIN_ROOT/skills/" 2>/dev/null || true)
if [ -n "$CCP_NAMES" ]; then
  echo "  ❌ skills に name: ccp-* が残存"
  echo "$CCP_NAMES" | head -3 | sed 's/^/      /'
  CCP_ISSUES=$((CCP_ISSUES + 1))
else
  echo "  ✅ skills に name: ccp-* なし"
fi

# Check 2: workflows の skill: に ccp- が含まれていないこと
CCP_WORKFLOWS=$(grep -rn "skill: ccp-" "$PLUGIN_ROOT/workflows/" 2>/dev/null || true)
if [ -n "$CCP_WORKFLOWS" ]; then
  echo "  ❌ workflows に skill: ccp-* が残存"
  echo "$CCP_WORKFLOWS" | head -3 | sed 's/^/      /'
  CCP_ISSUES=$((CCP_ISSUES + 1))
else
  echo "  ✅ workflows に skill: ccp-* なし"
fi

# Check 3: ccp-* ディレクトリが残っていないこと
CCP_DIRS=$(find "$PLUGIN_ROOT/skills" -type d -name "ccp-*" 2>/dev/null || true)
if [ -n "$CCP_DIRS" ]; then
  echo "  ❌ ccp-* ディレクトリが残存"
  echo "$CCP_DIRS" | head -3 | sed 's/^/      /'
  CCP_ISSUES=$((CCP_ISSUES + 1))
else
  echo "  ✅ ccp-* ディレクトリなし"
fi

if [ $CCP_ISSUES -eq 0 ]; then
  echo "  ✅ ccp-* スキル廃止OK"
else
  ERRORS=$((ERRORS + CCP_ISSUES))
fi

# ================================
# 10. スキル Mirror チェック
# ================================
echo ""
echo "📦 [10/13] スキル Mirror チェック..."

SKILLS_DIR="$PLUGIN_ROOT/skills"
CODEX_MIRROR="$PLUGIN_ROOT/codex/.codex/skills"
OPENCODE_MIRROR="$PLUGIN_ROOT/opencode/skills"
MIRROR_ISSUES=0

# コアスキル（5動詞 harness- prefix + aux）の mirror チェック
# SSOT: skills/ → ミラー先: codex/.codex/skills/, opencode/skills/
# NOTE: mirror 側には disable-model-invocation: true が追加されている（自動発動抑制）
#       この差異は意図的なため、比較時に除外する
HARNESS_SKILLS="harness-plan harness-work harness-review harness-release harness-setup harness-sync harness-loop"

# mirror 比較用ヘルパー: disable-model-invocation 行を除外してファイル単位で diff
# mirror 固有の設定（自動発動抑制）は意図的な差異のため許容する
diff_mirror() {
  local src_dir="$1"
  local mirror_dir="$2"

  # ファイル一覧を比較（ファイル構成の一致を確認）
  local src_files mirror_files
  src_files="$(cd "$src_dir" && find . -type f | sort)"
  mirror_files="$(cd "$mirror_dir" && find . -type f | sort)"
  if [ "$src_files" != "$mirror_files" ]; then
    return 1
  fi

  # 各ファイルを個別に比較（disable-model-invocation 行のみ除外）
  local f compared=0
  while IFS= read -r f; do
    [ -z "$f" ] && continue
    if ! diff -q \
      <(grep -v '^disable-model-invocation:' "$src_dir/$f") \
      <(grep -v '^disable-model-invocation:' "$mirror_dir/$f") \
      >/dev/null 2>&1; then
      return 1
    fi
    compared=$((compared + 1))
  done <<< "$src_files"

  # ファイル比較が1件も実行されなかった場合は安全側に倒す
  [ "$compared" -gt 0 ]
}

for skill in $HARNESS_SKILLS; do
  src="$SKILLS_DIR/$skill"
  if [ ! -d "$src" ]; then
    echo "  ❌ skills/$skill が存在しません（SSOT 欠落）"
    MIRROR_ISSUES=$((MIRROR_ISSUES + 1))
    continue
  fi

  for mirror_name in codex opencode; do
    case "$mirror_name" in
      codex) mirror_root="$CODEX_MIRROR" ;;
      opencode) mirror_root="$OPENCODE_MIRROR" ;;
    esac

    if [ ! -d "$mirror_root" ]; then
      continue
    fi

    mirror_path="$mirror_root/$skill"
    if [ ! -d "$mirror_path" ]; then
      echo "  ❌ $mirror_name: $skill がディレクトリとして存在しません"
      MIRROR_ISSUES=$((MIRROR_ISSUES + 1))
      continue
    fi

    if [ -L "$mirror_path" ]; then
      echo "  ❌ $mirror_name: $skill が symlink のままです"
      MIRROR_ISSUES=$((MIRROR_ISSUES + 1))
      continue
    fi

    if diff_mirror "$src" "$mirror_path"; then
      echo "  ✅ $mirror_name: $skill mirror is in sync"
    else
      echo "  ❌ $mirror_name: $skill mirror が skills/ と不一致"
      MIRROR_ISSUES=$((MIRROR_ISSUES + 1))
    fi
  done
done

if [ $MIRROR_ISSUES -gt 0 ]; then
  ERRORS=$((ERRORS + MIRROR_ISSUES))
fi

# breezing alias は codex mirror のみ（skills/ に SSOT あり）
BREEZING_SRC="$SKILLS_DIR/breezing"
if [ -d "$BREEZING_SRC" ]; then
  BREEZING_CODEX="$CODEX_MIRROR/breezing"
  if [ ! -d "$BREEZING_CODEX" ]; then
    echo "  ❌ codex: breezing がディレクトリとして存在しません"
    ERRORS=$((ERRORS + 1))
  elif [ -L "$BREEZING_CODEX" ]; then
    echo "  ❌ codex: breezing が symlink のままです"
    ERRORS=$((ERRORS + 1))
  elif diff_mirror "$BREEZING_SRC" "$BREEZING_CODEX"; then
    echo "  ✅ codex: breezing mirror is in sync"
  else
    echo "  ❌ codex: breezing mirror が skills/ と不一致"
    ERRORS=$((ERRORS + 1))
  fi
else
  echo "  ❌ skills/breezing が存在しません（SSOT 欠落）"
  ERRORS=$((ERRORS + 1))
fi

# ================================
# 11. CHANGELOG フォーマット検証
# ================================
echo ""
echo "📝 [11/13] CHANGELOG フォーマット検証..."

CHANGELOG_ISSUES=0

for changelog in "$PLUGIN_ROOT/CHANGELOG.md" "$PLUGIN_ROOT/CHANGELOG_ja.md"; do
  if [ ! -f "$changelog" ]; then
    continue
  fi

  cl_name=$(basename "$changelog")

  # Check 1: Keep a Changelog ヘッダー（## [x.y.z] - YYYY-MM-DD 形式）
  BAD_DATES=$(grep -nE '^\#\# \[[0-9]' "$changelog" | grep -vE '[0-9]{4}-[0-9]{2}-[0-9]{2}' | grep -v "Unreleased" || true)
  if [ -n "$BAD_DATES" ]; then
    echo "  ❌ $cl_name: ISO 8601 日付でないエントリ"
    echo "$BAD_DATES" | head -3 | sed 's/^/      /'
    CHANGELOG_ISSUES=$((CHANGELOG_ISSUES + 1))
  fi

  # Check 2: 非標準セクション見出し（Keep a Changelog 1.1.0 の 6 種以外）
  NON_STANDARD=$(grep -nE '^\#\#\# ' "$changelog" \
    | grep -viE '(Added|Changed|Deprecated|Removed|Fixed|Security|What.*Changed|あなたにとって)' \
    | grep -viE '(Internal|Breaking|Migration|Summary|Before)' \
    || true)
  if [ -n "$NON_STANDARD" ]; then
    echo "  ⚠️ $cl_name: 非標準セクション見出し（確認推奨）"
    echo "$NON_STANDARD" | head -3 | sed 's/^/      /'
    # 警告のみ（エラーにはしない）
  fi

  # Check 3: [Unreleased] セクションが存在するか
  if ! grep -q '^\#\# \[Unreleased\]' "$changelog"; then
    echo "  ❌ $cl_name: [Unreleased] セクションがありません"
    CHANGELOG_ISSUES=$((CHANGELOG_ISSUES + 1))
  fi
done

if [ $CHANGELOG_ISSUES -eq 0 ]; then
  echo "  ✅ CHANGELOG フォーマットOK"
else
  ERRORS=$((ERRORS + CHANGELOG_ISSUES))
fi

# ================================
# 12. README claim drift チェック
# ================================
echo ""
echo "📚 [12/13] README claim drift チェック..."

README_ISSUES=0
README_EN="$PLUGIN_ROOT/README.md"
README_JA="$PLUGIN_ROOT/README_ja.md"
SCOPE_DOC="$PLUGIN_ROOT/docs/distribution-scope.md"
RUBRIC_DOC="$PLUGIN_ROOT/docs/benchmark-rubric.md"
POSITIONING_DOC="$PLUGIN_ROOT/docs/positioning-notes.md"
WORK_ALL_DOC="$PLUGIN_ROOT/docs/evidence/work-all.md"

check_fixed_string() {
  local file_path="$1"
  local needle="$2"
  local label="$3"

  if [ ! -f "$file_path" ]; then
    echo "  ❌ ${label}: ファイルが存在しません: $file_path"
    README_ISSUES=$((README_ISSUES + 1))
    return
  fi

  if grep -qF "$needle" "$file_path"; then
    echo "  ✅ ${label}"
  else
    echo "  ❌ ${label}: 必須文字列が見つかりません"
    README_ISSUES=$((README_ISSUES + 1))
  fi
}

check_absent_string() {
  local file_path="$1"
  local needle="$2"
  local label="$3"

  if [ ! -f "$file_path" ]; then
    echo "  ❌ ${label}: ファイルが存在しません: $file_path"
    README_ISSUES=$((README_ISSUES + 1))
    return
  fi

  if grep -qF "$needle" "$file_path"; then
    echo "  ❌ ${label}: 古い claim が残っています"
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
    echo "  ❌ ${label}: ファイルが存在しません"
    README_ISSUES=$((README_ISSUES + 1))
  fi
}

check_fixed_string "$README_EN" "$LATEST_RELEASE_URL" "README.md latest release link"
check_fixed_string "$README_JA" "$LATEST_RELEASE_URL" "README_ja.md latest release link"
check_fixed_string "$README_EN" "$LATEST_RELEASE_BADGE" "README.md latest release badge"
check_fixed_string "$README_JA" "$LATEST_RELEASE_BADGE" "README_ja.md latest release badge"

check_exists "$SCOPE_DOC" "distribution-scope.md"
check_exists "$RUBRIC_DOC" "benchmark-rubric.md"
check_exists "$POSITIONING_DOC" "positioning-notes.md"
check_exists "$WORK_ALL_DOC" "work-all evidence doc"

check_fixed_string "$README_EN" "docs/CLAUDE_CODE_COMPATIBILITY.md" "README.md compatibility doc link"
check_fixed_string "$README_EN" "docs/CURSOR_INTEGRATION.md" "README.md cursor doc link"
check_fixed_string "$README_EN" "docs/evidence/work-all.md" "README.md work-all evidence link"
check_fixed_string "$README_EN" "docs/distribution-scope.md" "README.md distribution scope link"
check_fixed_string "$README_EN" "5 verb skills" "README.md 5 verb skills message"
check_fixed_string "$README_EN" "Go-native guardrail engine" "README.md Go-native guardrail engine message"
check_absent_string "$README_EN" "Production-ready code." "README.md stale production-ready wording"

check_fixed_string "$README_JA" "docs/CLAUDE_CODE_COMPATIBILITY.md" "README_ja.md compatibility doc link"
check_fixed_string "$README_JA" "docs/CURSOR_INTEGRATION.md" "README_ja.md cursor doc link"
check_fixed_string "$README_JA" "docs/evidence/work-all.md" "README_ja.md work-all evidence link"
check_fixed_string "$README_JA" "docs/distribution-scope.md" "README_ja.md distribution scope link"
check_fixed_string "$README_JA" "5動詞スキル" "README_ja.md 5動詞スキル message"
check_fixed_string "$README_JA" "Go ネイティブガードレールエンジン" "README_ja.md Go ネイティブガードレールエンジン message"
check_absent_string "$README_JA" "本番品質のコード。" "README_ja.md stale production-ready wording"

check_fixed_string "$SCOPE_DOC" '| `commands/` | Compatibility-retained |' "distribution-scope commands classification"
check_fixed_string "$SCOPE_DOC" '| `mcp-server/` | Development-only and distribution-excluded |' "distribution-scope mcp-server classification"
check_fixed_string "$RUBRIC_DOC" "| Static evidence |" "benchmark-rubric static evidence"
check_fixed_string "$RUBRIC_DOC" "| Executed evidence |" "benchmark-rubric executed evidence"
check_fixed_string "$POSITIONING_DOC" "runtime enforcement" "positioning-notes runtime enforcement"

if [ $README_ISSUES -eq 0 ]; then
  echo "  ✅ README claim drift チェックOK"
else
  ERRORS=$((ERRORS + README_ISSUES))
fi

# ================================
# 13. EN/JA ビジュアル同期チェック
# ================================
echo ""
echo "🎨 [13/13] EN/JA ビジュアル同期チェック..."

VISUAL_EN_DIR="$PLUGIN_ROOT/assets/readme-visuals-en/generated"
VISUAL_JA_DIR="$PLUGIN_ROOT/assets/readme-visuals-ja/generated"
VISUAL_ISSUES=0

if [ -d "$VISUAL_EN_DIR" ] && [ -d "$VISUAL_JA_DIR" ]; then
  # EN にあるファイルが JA にも存在し、viewBox サイズが一致することを確認
  for en_svg in "$VISUAL_EN_DIR"/*.svg; do
    [ ! -f "$en_svg" ] && continue
    svg_name=$(basename "$en_svg")
    ja_svg="$VISUAL_JA_DIR/$svg_name"

    if [ ! -f "$ja_svg" ]; then
      echo "  ❌ JA 版が欠落: $svg_name"
      VISUAL_ISSUES=$((VISUAL_ISSUES + 1))
      continue
    fi

    # viewBox の高さを比較（構造の大幅な乖離を検出）
    en_viewbox=$(grep -o 'viewBox="[^"]*"' "$en_svg" | head -1)
    ja_viewbox=$(grep -o 'viewBox="[^"]*"' "$ja_svg" | head -1)
    if [ "$en_viewbox" != "$ja_viewbox" ]; then
      echo "  ⚠️ viewBox 不一致: $svg_name (EN: $en_viewbox / JA: $ja_viewbox)"
      # 警告のみ（日本語は文字幅が異なるため高さ差は許容）
    fi

    # テーブル行数を比較（<rect y= の数で簡易判定）
    en_rows=$(grep -c '<rect y=' "$en_svg" 2>/dev/null || echo 0)
    ja_rows=$(grep -c '<rect y=' "$ja_svg" 2>/dev/null || echo 0)
    if [ "$en_rows" != "$ja_rows" ]; then
      echo "  ❌ 行数不一致: $svg_name (EN: ${en_rows}行 / JA: ${ja_rows}行)"
      VISUAL_ISSUES=$((VISUAL_ISSUES + 1))
    else
      echo "  ✅ $svg_name (${en_rows}行)"
    fi
  done
else
  echo "  ⚠️ EN/JA ビジュアルディレクトリが見つかりません（スキップ）"
fi

if [ $VISUAL_ISSUES -gt 0 ]; then
  ERRORS=$((ERRORS + VISUAL_ISSUES))
fi

# ================================
# 結果サマリー
# ================================
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [ $ERRORS -eq 0 ]; then
  echo "✅ すべてのチェックに合格しました"
  exit 0
else
  echo "❌ $ERRORS 個の問題が見つかりました"
  exit 1
fi
