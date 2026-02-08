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
echo "📁 [1/5] テンプレートファイルの存在確認..."

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
echo "🔗 [2/5] コマンド ↔ スキル の参照整合性..."

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
echo "🏷️ [3/5] バージョン番号の一貫性..."

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

# ================================
# 4. スキルの期待ファイル構成
# ================================
echo ""
echo "📋 [4/5] スキル定義の期待ファイル構成..."

# update-2agent-files の REQUIRED_FILES とテンプレートの同期
# 2agent は setup hub に統合済み（skills/_archived/2agent/）
# 代わりに setup/references/2agent-setup.md を確認
SETUP_2AGENT="$PLUGIN_ROOT/skills/setup/references/2agent-setup.md"
if [ -f "$SETUP_2AGENT" ]; then
  echo "  ✅ setup/references/2agent-setup.md が存在"
else
  echo "  ❌ setup/references/2agent-setup.md が見つかりません"
  ERRORS=$((ERRORS + 1))
fi

# ================================
# 5. Hooks 設定の整合性
# ================================
echo ""
echo "🪝 [5/5] Hooks 設定の整合性..."

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
echo "🚫 [6/9] /start-task 廃止の回帰チェック..."

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
echo "📁 [7/9] docs/ 正規化の回帰チェック..."

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
echo "🔓 [8/9] bypassPermissions 前提運用の回帰チェック..."

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

# Check 2: permissions.ask に Edit / Write が入っていないこと
if [ -f "$SECURITY_TEMPLATE" ]; then
  if grep -q '"Edit' "$SECURITY_TEMPLATE" || grep -q '"Write' "$SECURITY_TEMPLATE"; then
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

# Check 3: settings.local.json.template が存在し、defaultMode が bypassPermissions であること
LOCAL_TEMPLATE="$PLUGIN_ROOT/templates/claude/settings.local.json.template"
if [ -f "$LOCAL_TEMPLATE" ]; then
  if grep -q '"defaultMode"[[:space:]]*:[[:space:]]*"bypassPermissions"' "$LOCAL_TEMPLATE"; then
    echo "  ✅ settings.local.json.template: defaultMode=bypassPermissions"
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
echo "🚫 [9/9] ccp-* スキル廃止の回帰チェック..."

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
