#!/bin/bash
# test-commit-guard.sh
# Commit Guard 機能のテスト
#
# テスト対象:
# - scripts/pretooluse-guard.sh (git commit ブロックロジック)
# - scripts/posttooluse-commit-cleanup.sh (レビュー承認状態クリア)
# - hooks.json (フック登録)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# テスト結果カウンター
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# カラー出力
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# テスト関数
run_test() {
  local test_name="$1"
  local test_func="$2"

  TESTS_RUN=$((TESTS_RUN + 1))
  echo -n "  Testing: $test_name... "

  if $test_func; then
    echo -e "${GREEN}PASSED${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
  else
    echo -e "${RED}FAILED${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
  fi
}

# ==================================================
# Test 1: posttooluse-commit-cleanup.sh が存在するか
# ==================================================
test_cleanup_script_exists() {
  local script="$PROJECT_ROOT/scripts/posttooluse-commit-cleanup.sh"

  if [ ! -f "$script" ]; then
    echo "    Error: posttooluse-commit-cleanup.sh not found"
    return 1
  fi

  return 0
}

# ==================================================
# Test 2: スクリプトに実行権限があるか
# ==================================================
test_cleanup_script_executable() {
  local script="$PROJECT_ROOT/scripts/posttooluse-commit-cleanup.sh"

  if [ ! -x "$script" ]; then
    echo "    Error: posttooluse-commit-cleanup.sh is not executable"
    return 1
  fi

  return 0
}

# ==================================================
# Test 3: pretooluse-guard.sh に git commit 検出ロジックがあるか
# ==================================================
test_pretooluse_has_commit_guard() {
  local script="$PROJECT_ROOT/scripts/pretooluse-guard.sh"

  if ! grep -q "git[[:space:]]*commit" "$script" 2>/dev/null; then
    echo "    Error: git commit detection not found in pretooluse-guard.sh"
    return 1
  fi

  if ! grep -Eq "review-approved.json|review-result.json" "$script" 2>/dev/null; then
    echo "    Error: review artifact check not found in pretooluse-guard.sh"
    return 1
  fi

  return 0
}

# ==================================================
# Test 4: pretooluse-guard.sh にブロックメッセージがあるか
# ==================================================
test_pretooluse_has_block_message() {
  local script="$PROJECT_ROOT/scripts/pretooluse-guard.sh"

  if ! grep -q "deny_git_commit_no_review" "$script" 2>/dev/null; then
    echo "    Error: deny_git_commit_no_review message not found"
    return 1
  fi

  return 0
}

# ==================================================
# Test 5: posttooluse-commit-cleanup.sh に git commit 検出があるか
# ==================================================
test_cleanup_detects_git_commit() {
  local script="$PROJECT_ROOT/scripts/posttooluse-commit-cleanup.sh"

  if ! grep -q "git[[:space:]]*commit" "$script" 2>/dev/null; then
    echo "    Error: git commit detection not found in cleanup script"
    return 1
  fi

  return 0
}

# ==================================================
# Test 6: posttooluse-commit-cleanup.sh に状態ファイル削除ロジックがあるか
# ==================================================
test_cleanup_removes_state_file() {
  local script="$PROJECT_ROOT/scripts/posttooluse-commit-cleanup.sh"

  # スクリプトは変数経由で削除: rm -f "$REVIEW_STATE_FILE"
  if ! grep -q 'rm -f.*REVIEW_STATE_FILE' "$script" 2>/dev/null; then
    echo "    Error: state file removal logic not found"
    return 1
  fi

  # 状態ファイルパスの定義も確認
  if ! grep -Eq "review-approved.json|review-result.json" "$script" 2>/dev/null; then
    echo "    Error: review artifact path definition not found"
    return 1
  fi

  return 0
}

# ==================================================
# Test 7: hooks.json に commit-cleanup フックが登録されているか
# ==================================================
test_hooks_has_commit_cleanup() {
  local hooks_file="$PROJECT_ROOT/hooks/hooks.json"

  if ! command -v jq &> /dev/null; then
    echo "    Warning: jq not available, skipping JSON validation"
    # jq がなくても grep で確認
    if ! grep -q "posttooluse-commit-cleanup" "$hooks_file" 2>/dev/null; then
      echo "    Error: commit-cleanup hook not registered in hooks.json"
      return 1
    fi
    return 0
  fi

  # PostToolUse に Bash マッチャーで commit-cleanup が登録されているか
  if ! jq -e '.hooks.PostToolUse[] | select(.matcher == "Bash") | .hooks[] | select(.command | contains("posttooluse-commit-cleanup"))' "$hooks_file" > /dev/null 2>&1; then
    echo "    Error: commit-cleanup hook not properly registered for Bash in PostToolUse"
    return 1
  fi

  return 0
}

# ==================================================
# Test 8: .claude-plugin/hooks.json にも同じフックがあるか
# ==================================================
test_plugin_hooks_has_commit_cleanup() {
  local hooks_file="$PROJECT_ROOT/.claude-plugin/hooks.json"

  if ! grep -q "posttooluse-commit-cleanup" "$hooks_file" 2>/dev/null; then
    echo "    Error: commit-cleanup hook not registered in .claude-plugin/hooks.json"
    return 1
  fi

  return 0
}

# ==================================================
# Test 9: config テンプレートに commit_guard 設定があるか
# ==================================================
test_config_has_commit_guard_option() {
  local config_template="$PROJECT_ROOT/templates/.claude-code-harness.config.yaml.template"

  if ! grep -q "commit_guard:" "$config_template" 2>/dev/null; then
    echo "    Error: commit_guard option not found in config template"
    return 1
  fi

  return 0
}

# ==================================================
# Test 10: posttooluse-commit-cleanup.sh がエラー時に状態を保持するか
# ==================================================
test_cleanup_preserves_on_error() {
  local script="$PROJECT_ROOT/scripts/posttooluse-commit-cleanup.sh"

  # エラーパターン検出ロジックがあるか確認
  if ! grep -Eq "error|fatal|failed|nothing to commit" "$script" 2>/dev/null; then
    echo "    Error: error detection logic not found in cleanup script"
    return 1
  fi

  return 0
}

# ==================================================
# メイン実行
# ==================================================
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " Commit Guard テスト"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

echo "  [PreToolUse Guard]"
run_test "pretooluse-guard.sh に git commit 検出ロジックがある" test_pretooluse_has_commit_guard
run_test "pretooluse-guard.sh にブロックメッセージがある" test_pretooluse_has_block_message

echo ""
echo "  [PostToolUse Cleanup]"
run_test "posttooluse-commit-cleanup.sh が存在する" test_cleanup_script_exists
run_test "posttooluse-commit-cleanup.sh に実行権限がある" test_cleanup_script_executable
run_test "posttooluse-commit-cleanup.sh に git commit 検出がある" test_cleanup_detects_git_commit
run_test "posttooluse-commit-cleanup.sh に状態ファイル削除ロジックがある" test_cleanup_removes_state_file
run_test "posttooluse-commit-cleanup.sh がエラー時に状態を保持する" test_cleanup_preserves_on_error

echo ""
echo "  [Hooks Integration]"
run_test "hooks.json に commit-cleanup フックが登録されている" test_hooks_has_commit_cleanup
run_test ".claude-plugin/hooks.json にも commit-cleanup フックがある" test_plugin_hooks_has_commit_cleanup

echo ""
echo "  [Configuration]"
run_test "config テンプレートに commit_guard 設定がある" test_config_has_commit_guard_option

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " テスト結果: $TESTS_PASSED/$TESTS_RUN passed"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

if [ "$TESTS_FAILED" -gt 0 ]; then
  exit 1
fi

exit 0
