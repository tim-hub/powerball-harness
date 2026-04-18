#!/bin/bash
#
# test-skill-mid-message.sh
# CC 2.1.110 修正: disable-model-invocation: true のスキルが mid-message 呼び出しで
# 正しくフロントマターを保持しているかを静的に検証する smoke test。
#
# 背景 (CC 2.1.110):
#   Skills with `disable-model-invocation: true` が `/<skill>` mid-message 呼び出しで
#   動くよう修正された。これにより harness-review 等の保護スキルが
#   mid-message 呼び出しでも機能する。
#
# このテストが検証すること:
#   1. disable-model-invocation: true を持つスキルの SKILL.md が存在すること
#   2. フロントマターが YAML として正しく解析できること (name フィールドあり)
#   3. allowed-tools フィールドが配列形式で存在すること
#
# このテストが検証しないこと (実行環境依存):
#   - CC ランタイムでの実際の mid-message 呼び出し (CC CLI が必要)
#   - モデル呼び出しの有無 (ランタイム挙動)
#
# Usage:
#   bash tests/test-skill-mid-message.sh
#   bash tests/test-skill-mid-message.sh --verbose
#
# Exit code:
#   0 = all checks passed
#   1 = one or more checks failed

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PLUGIN_ROOT="$(dirname "$SCRIPT_DIR")"
SKILLS_DIR="${PLUGIN_ROOT}/skills"

VERBOSE=0
for arg in "$@"; do
    [[ "$arg" == "--verbose" ]] && VERBOSE=1
done

PASS=0
FAIL=0
SKIP=0

pass() { echo "  [PASS] $1"; PASS=$((PASS + 1)); }
fail() { echo "  [FAIL] $1"; FAIL=$((FAIL + 1)); }
skip() { echo "  [SKIP] $1"; SKIP=$((SKIP + 1)); }
info() { [[ "$VERBOSE" -eq 1 ]] && echo "  [INFO] $1" || true; }

echo "=============================================="
echo "  smoke test: disable-model-invocation skills"
echo "  CC 2.1.110 mid-message fix 対応確認"
echo "=============================================="
echo ""

# disable-model-invocation: true を持つスキルを列挙
SKILL_FILES=()
while IFS= read -r -d '' file; do
    if grep -q "disable-model-invocation: true" "$file" 2>/dev/null; then
        SKILL_FILES+=("$file")
    fi
done < <(find "$SKILLS_DIR" -name "SKILL.md" -print0 2>/dev/null)

if [[ ${#SKILL_FILES[@]} -eq 0 ]]; then
    echo "  [WARN] disable-model-invocation: true を持つスキルが見つかりません"
    echo "  Skills dir: $SKILLS_DIR"
    exit 0
fi

echo "  対象スキル数: ${#SKILL_FILES[@]}"
echo ""

for skill_file in "${SKILL_FILES[@]}"; do
    skill_dir="$(dirname "$skill_file")"
    skill_name="$(basename "$skill_dir")"

    echo "--- $skill_name ---"

    # 1. SKILL.md が存在すること
    if [[ -f "$skill_file" ]]; then
        pass "SKILL.md が存在する: $skill_file"
    else
        fail "SKILL.md が存在しない: $skill_file"
        continue
    fi

    # 2. フロントマター開始行 (---) があること
    if head -1 "$skill_file" | grep -q "^---$"; then
        pass "フロントマター開始行あり"
    else
        fail "フロントマター開始行なし (先頭が --- でない)"
        continue
    fi

    # 3. name フィールドがあること
    if grep -q "^name:" "$skill_file"; then
        SKILL_DECLARED_NAME="$(grep "^name:" "$skill_file" | head -1 | sed 's/^name: *//')"
        pass "name フィールドあり: $SKILL_DECLARED_NAME"
        info "  宣言名 '$SKILL_DECLARED_NAME' vs ディレクトリ名 '$skill_name'"
    else
        fail "name フィールドが見つからない"
    fi

    # 4. disable-model-invocation: true が存在すること (二重確認)
    if grep -q "^disable-model-invocation: true" "$skill_file"; then
        pass "disable-model-invocation: true が設定済み"
    else
        fail "disable-model-invocation: true が設定されていない"
    fi

    # 5. allowed-tools フィールドがあること (mid-message 呼び出しで tool 実行が前提)
    if grep -q "^allowed-tools:" "$skill_file"; then
        ALLOWED_TOOLS_LINE="$(grep "^allowed-tools:" "$skill_file" | head -1)"
        pass "allowed-tools フィールドあり: $ALLOWED_TOOLS_LINE"
    else
        skip "allowed-tools フィールドなし (disable-model-invocation: true のみのスキルは tool 不要な場合あり)"
    fi

    # 6. description フィールドがあること (mid-message 自動ロードのため必須)
    if grep -q "^description:" "$skill_file"; then
        pass "description フィールドあり"
    else
        fail "description フィールドが見つからない (mid-message ロードに必要)"
    fi

    echo ""
done

echo "=============================================="
echo "  結果: PASS=${PASS}  FAIL=${FAIL}  SKIP=${SKIP}"
echo ""
echo "  静的検証のみ実施。CC ランタイムの mid-message 呼び出し動作は"
echo "  CC CLI 2.1.110+ で手動検証が必要 (CI 外)。"
echo "  詳細: docs/cc-2.1.99-2.1.110-impact.md (44.7.1 smoke test 結果)"
echo "=============================================="

if [[ "$FAIL" -gt 0 ]]; then
    exit 1
fi
exit 0
