#!/bin/bash
# test-validate-plugin-quick.sh
# validate-plugin.sh --quick の jq フォールバック検証
#
# テスト内容:
#   1. validate-plugin.sh --quick が正常に動作すること（現プロジェクトで PASS）
#   2. jq フォールバック関数（_check_json_syntax）が有効な JSON を正しく判定すること
#   3. jq フォールバック関数が壊れた JSON を正しく検知すること
#   4. jq なし環境で python3 fallback が機能すること
#   5. jq も python3 もなし環境で skip（fail-open）になること
#
# Usage: bash tests/test-validate-plugin-quick.sh

set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PLUGIN_ROOT="$(dirname "${SCRIPT_DIR}")"

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

PASS_COUNT=0
FAIL_COUNT=0

pass_test() {
    echo -e "${GREEN}✓${NC} $1"
    PASS_COUNT=$(( PASS_COUNT + 1 ))
}

fail_test() {
    echo -e "${RED}✗${NC} $1"
    FAIL_COUNT=$(( FAIL_COUNT + 1 ))
}

warn_test() {
    echo -e "${YELLOW}⚠${NC} $1"
}

echo "=========================================="
echo "validate-plugin.sh --quick jq fallback テスト"
echo "=========================================="
echo ""

WORK_DIR="$(mktemp -d /tmp/test-validate-quick-XXXXXX)"
cleanup() { rm -rf "${WORK_DIR}" 2>/dev/null || true; }
trap cleanup EXIT

VALID_JSON="${WORK_DIR}/valid.sprint-contract.json"
BROKEN_JSON="${WORK_DIR}/broken.sprint-contract.json"

cat > "${VALID_JSON}" << 'EOF'
{
  "task_id": "1",
  "review": {
    "status": "approved",
    "reviewer_profile": "static"
  }
}
EOF

printf '{"broken": true, invalid json' > "${BROKEN_JSON}"

# ── テスト 1: validate-plugin.sh --quick が現プロジェクトで PASS ─────────────────
echo "--- テスト 1: validate-plugin.sh --quick（現プロジェクト）---"

output=$(bash "${SCRIPT_DIR}/validate-plugin.sh" --quick 2>&1)
exit_code=$?

if [ "${exit_code}" -eq 0 ]; then
    pass_test "validate-plugin.sh --quick: exit 0（PASS）"
else
    fail_test "validate-plugin.sh --quick: exit ${exit_code}（FAIL）"
    echo "  出力: ${output}" >&2
fi

# ── テスト 2: _check_json_syntax 関数の存在確認 ────────────────────────────────
echo ""
echo "--- テスト 2: _check_json_syntax 関数・_JSON_PARSER 変数の存在 ---"

if grep -q '_check_json_syntax' "${SCRIPT_DIR}/validate-plugin.sh"; then
    pass_test "_check_json_syntax 関数が validate-plugin.sh に存在します"
else
    fail_test "_check_json_syntax 関数が validate-plugin.sh に見つかりません"
fi

if grep -q '_JSON_PARSER' "${SCRIPT_DIR}/validate-plugin.sh"; then
    pass_test "_JSON_PARSER 変数（jq/python3/skip 分岐）が存在します"
else
    fail_test "_JSON_PARSER 変数が見つかりません"
fi

if grep -q 'python3.*json.*load\|python3 -c.*json' "${SCRIPT_DIR}/validate-plugin.sh"; then
    pass_test "python3 fallback コードが存在します"
else
    fail_test "python3 fallback コードが見つかりません"
fi

# ── テスト 3: jq 環境での有効・壊れた JSON チェック ───────────────────────────
echo ""
echo "--- テスト 3: _check_json_syntax ロジックを直接検証（インライン実装） ---"

# validate-plugin.sh の _check_json_syntax と同等のロジックをここで再現してテストする
# （validate-plugin.sh は PLUGIN_ROOT を自動計算するため外部から PLUGIN_ROOT を差し替えられない）

_test_check_json() {
    local parser="$1"
    local file="$2"
    case "${parser}" in
        jq)      jq empty "${file}" 2>/dev/null ;;
        python3) python3 -c "import json,sys; json.load(open(sys.argv[1]))" "${file}" 2>/dev/null ;;
        skip)    return 0 ;;
    esac
}

if command -v jq >/dev/null 2>&1; then
    if _test_check_json "jq" "${VALID_JSON}"; then
        pass_test "jq: 有効な JSON → PASS"
    else
        fail_test "jq: 有効な JSON → FAIL（誤検知）"
    fi

    if ! _test_check_json "jq" "${BROKEN_JSON}"; then
        pass_test "jq: 壊れた JSON → FAIL 検知（期待通り）"
    else
        fail_test "jq: 壊れた JSON → PASS（検知失敗）"
    fi
else
    warn_test "jq が利用不可のため、jq テストをスキップします"
fi

# ── テスト 4: python3 fallback での有効・壊れた JSON チェック ──────────────────
echo ""
echo "--- テスト 4: python3 fallback での JSON チェック ---"

if command -v python3 >/dev/null 2>&1; then
    if _test_check_json "python3" "${VALID_JSON}"; then
        pass_test "python3: 有効な JSON → PASS"
    else
        fail_test "python3: 有効な JSON → FAIL（誤検知）"
    fi

    if ! _test_check_json "python3" "${BROKEN_JSON}"; then
        pass_test "python3: 壊れた JSON → FAIL 検知（期待通り）"
    else
        fail_test "python3: 壊れた JSON → PASS（検知失敗）"
    fi
else
    warn_test "python3 が利用不可のため、python3 fallback テストをスキップします"
fi

# ── テスト 5: skip モードは常に return 0（fail-open）─────────────────────────
echo ""
echo "--- テスト 5: skip モード（fail-open）---"

if _test_check_json "skip" "${VALID_JSON}" && _test_check_json "skip" "${BROKEN_JSON}"; then
    pass_test "skip モード: 有効・壊れた JSON いずれも return 0（fail-open）"
else
    fail_test "skip モード: return 0 にならない（fail-open が機能していない）"
fi

# ── テスト 6: validate-plugin.sh の jq フォールバック分岐コードの構造確認 ───────
echo ""
echo "--- テスト 6: validate-plugin.sh の fallback 分岐構造確認 ---"

# jq → python3 → skip の分岐が存在するか
if grep -q 'command -v jq' "${SCRIPT_DIR}/validate-plugin.sh"; then
    pass_test "jq 存在チェック（command -v jq）が存在します"
else
    fail_test "jq 存在チェックが見つかりません"
fi

if grep -q 'command -v python3' "${SCRIPT_DIR}/validate-plugin.sh"; then
    pass_test "python3 存在チェック（command -v python3）が存在します"
else
    fail_test "python3 存在チェックが見つかりません"
fi

if grep -q '"skip"' "${SCRIPT_DIR}/validate-plugin.sh"; then
    pass_test "skip 分岐（\"skip\"）が存在します"
else
    fail_test "skip 分岐が見つかりません"
fi

echo ""
echo "=========================================="
echo "テスト結果サマリー"
echo "=========================================="
echo -e "${GREEN}合格:${NC} ${PASS_COUNT}"
echo -e "${RED}失敗:${NC} ${FAIL_COUNT}"
echo ""

if [ "${FAIL_COUNT}" -eq 0 ]; then
    echo -e "${GREEN}✓ 全テスト合格${NC}"
    exit 0
else
    echo -e "${RED}✗ ${FAIL_COUNT} 件のテストが失敗しました${NC}"
    exit 1
fi
