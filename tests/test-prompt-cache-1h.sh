#!/usr/bin/env bash
# test-prompt-cache-1h.sh
# enable-1h-cache.sh の動作検証テスト
#
# テスト内容:
#   1. 新規 env.local への追記（ENABLE_PROMPT_CACHING_1H=1 が書き込まれること）
#   2. 冪等性（2 回実行しても同じ行が 1 行だけ存在すること）
#   3. 既存の別キー行への干渉なし（他のキーが維持されること）
#   4. 既存の同キー・別値がある場合は警告して exit 1 すること
#   5. env.local が ENABLE_PROMPT_CACHING_1H=1 の行を持つと env に伝播すること

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
TARGET_SCRIPT="${ROOT_DIR}/scripts/enable-1h-cache.sh"

PASS_COUNT=0
FAIL_COUNT=0

pass_test() {
  echo "PASS: $1"
  PASS_COUNT=$((PASS_COUNT + 1))
}

fail_test() {
  echo "FAIL: $1"
  FAIL_COUNT=$((FAIL_COUNT + 1))
}

# テスト用一時ディレクトリ（git リポジトリを模倣するため git init が必要）
setup_tmp_repo() {
  local tmp_dir
  tmp_dir="$(mktemp -d)"
  git -C "${tmp_dir}" init -q
  echo "${tmp_dir}"
}

cleanup_tmp() {
  local dir="$1"
  rm -rf "${dir}"
}

# ---------- テスト 1: スクリプトが存在し実行可能であること ----------
echo "--- Test 1: スクリプトの存在と実行権限 ---"
if [[ -f "${TARGET_SCRIPT}" ]]; then
  pass_test "enable-1h-cache.sh が存在する"
else
  fail_test "enable-1h-cache.sh が存在しない (path: ${TARGET_SCRIPT})"
fi

if [[ -x "${TARGET_SCRIPT}" ]]; then
  pass_test "enable-1h-cache.sh が実行可能"
else
  fail_test "enable-1h-cache.sh に実行権限がない"
fi

# ---------- テスト 2: 新規 env.local への追記 ----------
echo "--- Test 2: 新規 env.local への追記 ---"
TMP_REPO="$(setup_tmp_repo)"

# env.local が存在しない状態で実行
if (cd "${TMP_REPO}" && bash "${TARGET_SCRIPT}" > /dev/null 2>&1); then
  if [[ -f "${TMP_REPO}/env.local" ]]; then
    pass_test "env.local が新規作成された"
  else
    fail_test "env.local が作成されなかった"
  fi

  if grep -qE "^export ENABLE_PROMPT_CACHING_1H=1$" "${TMP_REPO}/env.local"; then
    pass_test "ENABLE_PROMPT_CACHING_1H=1 が env.local に書き込まれた"
  else
    fail_test "ENABLE_PROMPT_CACHING_1H=1 が env.local に見つからない"
  fi
else
  fail_test "スクリプト実行が失敗した（新規 env.local）"
fi

cleanup_tmp "${TMP_REPO}"

# ---------- テスト 3: 冪等性（2 回実行） ----------
echo "--- Test 3: 冪等性 ---"
TMP_REPO="$(setup_tmp_repo)"

(cd "${TMP_REPO}" && bash "${TARGET_SCRIPT}" > /dev/null 2>&1)
(cd "${TMP_REPO}" && bash "${TARGET_SCRIPT}" > /dev/null 2>&1)

COUNT=$(grep -cE "^export ENABLE_PROMPT_CACHING_1H=1$" "${TMP_REPO}/env.local" 2>/dev/null || echo "0")
if [[ "${COUNT}" -eq 1 ]]; then
  pass_test "2 回実行後も ENABLE_PROMPT_CACHING_1H=1 は 1 行だけ（冪等）"
else
  fail_test "冪等性違反: ENABLE_PROMPT_CACHING_1H=1 が ${COUNT} 行存在する"
fi

cleanup_tmp "${TMP_REPO}"

# ---------- テスト 4: 既存の他キー行への干渉なし ----------
echo "--- Test 4: 既存の他キー行への干渉なし ---"
TMP_REPO="$(setup_tmp_repo)"
echo "SOME_OTHER_KEY=hello" > "${TMP_REPO}/env.local"

(cd "${TMP_REPO}" && bash "${TARGET_SCRIPT}" > /dev/null 2>&1)

if grep -qE "^SOME_OTHER_KEY=hello$" "${TMP_REPO}/env.local"; then
  pass_test "既存キー SOME_OTHER_KEY が維持された"
else
  fail_test "既存キー SOME_OTHER_KEY が消えた"
fi

if grep -qE "^export ENABLE_PROMPT_CACHING_1H=1$" "${TMP_REPO}/env.local"; then
  pass_test "ENABLE_PROMPT_CACHING_1H=1 が追記された"
else
  fail_test "ENABLE_PROMPT_CACHING_1H=1 が追記されなかった"
fi

cleanup_tmp "${TMP_REPO}"

# ---------- テスト 5: 既存の同キー・別値がある場合は exit 1 ----------
echo "--- Test 5: 同キー・別値の場合は exit 1 ---"
TMP_REPO="$(setup_tmp_repo)"
echo "ENABLE_PROMPT_CACHING_1H=0" > "${TMP_REPO}/env.local"

if (cd "${TMP_REPO}" && bash "${TARGET_SCRIPT}" > /dev/null 2>&1); then
  fail_test "同キー・別値でも exit 0 になった（exit 1 期待）"
else
  pass_test "同キー・別値の場合に exit 1 が返った"
fi

cleanup_tmp "${TMP_REPO}"

# ---------- テスト 6: env への伝播シミュレーション ----------
# env.local の値を source した場合、ENABLE_PROMPT_CACHING_1H が設定されること
echo "--- Test 6: env.local を source した場合の env 伝播 ---"
TMP_REPO="$(setup_tmp_repo)"
(cd "${TMP_REPO}" && bash "${TARGET_SCRIPT}" > /dev/null 2>&1)

# env.local を source して変数が設定されるか確認
SOURCED_VALUE=$(bash -c "source '${TMP_REPO}/env.local' 2>/dev/null; echo \"\${ENABLE_PROMPT_CACHING_1H:-UNSET}\"")
if [[ "${SOURCED_VALUE}" == "1" ]]; then
  pass_test "env.local を source すると ENABLE_PROMPT_CACHING_1H=1 が環境変数に設定される"
else
  fail_test "env.local source 後の ENABLE_PROMPT_CACHING_1H が期待値 '1' でなく '${SOURCED_VALUE}'"
fi

# Critical: source した env.local が subprocess (claude 等) にも env として伝播するか
# `export KEY=VALUE` 形式でないと subprocess には継承されない (shell-local 変数のまま)
CHILD_VALUE=$(bash -c "source '${TMP_REPO}/env.local' 2>/dev/null; bash -c 'echo \"\${ENABLE_PROMPT_CACHING_1H:-UNSET}\"'")
if [[ "${CHILD_VALUE}" == "1" ]]; then
  pass_test "env.local source 後、subprocess (子 bash) にも ENABLE_PROMPT_CACHING_1H=1 が伝播 (export 確認)"
else
  fail_test "env.local source 後の subprocess で ENABLE_PROMPT_CACHING_1H が期待値 '1' でなく '${CHILD_VALUE}' — export 抜け"
fi

cleanup_tmp "${TMP_REPO}"

# ---------- 結果サマリ ----------
echo ""
echo "========================================"
echo "Results: ${PASS_COUNT} passed, ${FAIL_COUNT} failed"
echo "========================================"

if [[ "${FAIL_COUNT}" -gt 0 ]]; then
  exit 1
fi

exit 0
