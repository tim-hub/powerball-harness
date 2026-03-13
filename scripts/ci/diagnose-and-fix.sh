#!/bin/bash
# diagnose-and-fix.sh
# CI エラーを診断し、修正案を提案または自動修正するスクリプト
#
# Usage:
#   ./scripts/ci/diagnose-and-fix.sh          # 診断のみ
#   ./scripts/ci/diagnose-and-fix.sh --fix    # 自動修正も実行
#
# このスクリプトは CI 失敗時に Claude が実行し、修正案を得るために使用します。

set -euo pipefail

PLUGIN_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$PLUGIN_ROOT"

AUTO_FIX=false
if [ "$1" = "--fix" ]; then
  AUTO_FIX=true
fi

ISSUES_FOUND=0
FIXES_APPLIED=0

echo "🔧 CI 診断＆修正ツール"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# ================================
# 1. バージョン同期チェック
# ================================
check_version_sync() {
  echo "📋 [1/5] バージョン同期チェック..."

  local file_version=$(cat VERSION 2>/dev/null | tr -d '[:space:]')
  local json_version=$(grep '"version"' .claude-plugin/plugin.json | head -1 | sed 's/.*: *"\([^"]*\)".*/\1/')

  if [ "$file_version" != "$json_version" ]; then
    echo "  ❌ VERSION ($file_version) と plugin.json ($json_version) が不一致"
    ISSUES_FOUND=$((ISSUES_FOUND + 1))

    if [ "$AUTO_FIX" = true ]; then
      echo "  🔧 修正中: plugin.json を $file_version に更新..."
      sed -i.bak "s/\"version\": \"$json_version\"/\"version\": \"$file_version\"/" .claude-plugin/plugin.json
      rm -f .claude-plugin/plugin.json.bak
      FIXES_APPLIED=$((FIXES_APPLIED + 1))
      echo "  ✅ 修正完了"
    else
      echo "  💡 修正案: plugin.json の version を \"$file_version\" に変更"
    fi
  else
    echo "  ✅ 同期済み (v$file_version)"
  fi
}

# ================================
# 2. チェックリスト同期チェック
# ================================
check_checklist_sync() {
  echo ""
  echo "📋 [2/5] チェックリスト同期チェック..."

  if ./scripts/ci/check-checklist-sync.sh >/dev/null 2>&1; then
    echo "  ✅ 同期済み"
  else
    echo "  ❌ スクリプトとコマンドのチェックリストが不一致"
    ISSUES_FOUND=$((ISSUES_FOUND + 1))

    echo "  💡 修正案:"
    echo "     1. scripts/*.sh の check_file/check_dir を確認"
    echo "     2. commands/*.md のチェックリストを手動で更新"
    echo "     (自動修正非対応 - 手動で確認が必要)"
  fi
}

# ================================
# 3. テンプレート存在チェック
# ================================
check_templates() {
  echo ""
  echo "📋 [3/5] テンプレート存在チェック..."

  local missing=()
  local templates=(
    "templates/AGENTS.md.template"
    "templates/CLAUDE.md.template"
    "templates/Plans.md.template"
    "templates/.claude-code-harness-version.template"
    "templates/cursor/commands/start-session.md"
    "templates/cursor/commands/handoff-to-claude.md"
    "templates/cursor/commands/review-cc-work.md"
    "templates/rules/workflow.md.template"
    "templates/rules/coding-standards.md.template"
  )

  for t in "${templates[@]}"; do
    if [ ! -f "$t" ]; then
      missing+=("$t")
    fi
  done

  if [ ${#missing[@]} -eq 0 ]; then
    echo "  ✅ 全テンプレート存在"
  else
    echo "  ❌ 不足テンプレート:"
    for m in "${missing[@]}"; do
      echo "     - $m"
    done
    ISSUES_FOUND=$((ISSUES_FOUND + 1))
    echo "  💡 修正案: 不足ファイルを作成"
  fi
}

# ================================
# 4. Hooks 整合性チェック
# ================================
check_hooks() {
  echo ""
  echo "📋 [4/5] Hooks 整合性チェック..."

  if ! jq empty hooks/hooks.json 2>/dev/null; then
    echo "  ❌ hooks.json が無効な JSON"
    ISSUES_FOUND=$((ISSUES_FOUND + 1))
    echo "  💡 修正案: hooks/hooks.json の JSON 構文を確認"
    return
  fi

  local missing_scripts=()
  local script_refs=$(grep -oE 'scripts/[a-zA-Z0-9_.-]+' hooks/hooks.json 2>/dev/null || true)

  for ref in $script_refs; do
    if [ ! -f "$ref" ]; then
      missing_scripts+=("$ref")
    fi
  done

  if [ ${#missing_scripts[@]} -eq 0 ]; then
    echo "  ✅ Hooks 設定正常"
  else
    echo "  ❌ 参照スクリプト不足:"
    for s in "${missing_scripts[@]}"; do
      echo "     - $s"
    done
    ISSUES_FOUND=$((ISSUES_FOUND + 1))
    echo "  💡 修正案: 不足スクリプトを作成、または hooks.json から参照を削除"
  fi
}

# ================================
# 5. リリースメタデータチェック
# ================================
check_version_bump() {
  echo ""
  echo "📋 [5/5] リリースメタデータチェック..."

  local check_log
  check_log="$(mktemp)"

  if bash ./scripts/ci/check-version-bump.sh >"$check_log" 2>&1; then
    sed 's/^/  /' "$check_log"
    rm -f "$check_log"
    return
  fi

  sed 's/^/  /' "$check_log"
  rm -f "$check_log"

  if [ "$AUTO_FIX" = true ] && ! bash ./scripts/sync-version.sh check >/dev/null 2>&1; then
    echo "  🔧 修正中: plugin.json を VERSION に同期..."
    bash ./scripts/sync-version.sh sync
    FIXES_APPLIED=$((FIXES_APPLIED + 1))

    if bash ./scripts/ci/check-version-bump.sh >/dev/null 2>&1; then
      echo "  ✅ plugin.json 同期で release metadata 整合を回復"
      return
    fi
  fi

  ISSUES_FOUND=$((ISSUES_FOUND + 1))
  echo "  💡 修正方針:"
  echo "     - 通常 PR では VERSION を変更しない"
  echo "     - release 時だけ VERSION / plugin.json / CHANGELOG release entry を一緒に更新する"
}

# ================================
# メイン実行
# ================================

check_version_sync
check_checklist_sync
check_templates
check_hooks
check_version_bump

# ================================
# 結果サマリー
# ================================

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [ $ISSUES_FOUND -eq 0 ]; then
  echo "✅ 問題は見つかりませんでした"
  exit 0
fi

echo "📊 結果サマリー:"
echo "  - 検出された問題: $ISSUES_FOUND 個"

if [ "$AUTO_FIX" = true ]; then
  echo "  - 自動修正: $FIXES_APPLIED 個"
  if [ $FIXES_APPLIED -gt 0 ]; then
    echo ""
    echo "💡 次のステップ:"
    echo "  1. 修正内容を確認: git diff"
    echo "  2. CHANGELOG.md を更新"
    echo "  3. コミット＆プッシュ"
  fi
else
  echo ""
  echo "💡 自動修正を実行するには:"
  echo "  ./scripts/ci/diagnose-and-fix.sh --fix"
fi

exit 1
