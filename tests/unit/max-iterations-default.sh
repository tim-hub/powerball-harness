#!/bin/bash
# max-iterations-default.sh
# review.max_iterations のデフォルト値と手動上書きの動作を検証するユニットテスト
#
# Usage: ./tests/unit/max-iterations-default.sh
#
# 設計方針（Finding 4 対応後）:
#   generate-sprint-contract.sh の detectMaxIterations() は HTML コメントマーカーのみを受け付ける。
#   記法: <!-- max_iterations: 15 -->（Markdown として表示されないため例示テキストと区別可能）
#   素のテキスト「max_iterations: 15」は意図的に無視する（自己参照バグ防止）。
#   範囲ガード: 1-30 の範囲外は profile default にフォールバック + stderr 警告。

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

PASS=0
FAIL=0

check() {
  local label="$1"
  local expected="$2"
  local actual="$3"
  if [ "$actual" = "$expected" ]; then
    echo "  PASS: $label (got $actual)"
    PASS=$((PASS + 1))
  else
    echo "  FAIL: $label — expected $expected, got $actual" >&2
    FAIL=$((FAIL + 1))
  fi
}

# ベースの Plans.md を用意
cat > "${TMP_DIR}/Plans.md" <<'EOF'
| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| T-static | static タスク | test を実装する | - | cc:TODO |
| T-browser | browser タスク | browser で UI フローを確認する | - | cc:TODO |
| T-html-comment | HTML コメントタスク | <!-- max_iterations: 15 --> を DoD に記載 | - | cc:TODO |
| T-out-of-range | 範囲外タスク | <!-- max_iterations: 100 --> は無効 | - | cc:TODO |
| T-plain-text | 素のテキストタスク | max_iterations: 15 とだけ書いてある | - | cc:TODO |
EOF

echo "=== Case (i): contract なし相当（static profile）→ max_iterations=3 ==="
OUT_I="${TMP_DIR}/out-i.json"
(cd "${TMP_DIR}" && "${PROJECT_ROOT}/scripts/generate-sprint-contract.sh" "T-static" "${TMP_DIR}/Plans.md" "${OUT_I}" >/dev/null)
ACTUAL_I="$(jq -r '.review.max_iterations' "${OUT_I}")"
check "static profile → 3" "3" "${ACTUAL_I}"

echo "=== Case (ii): browser profile → max_iterations=5 ==="
OUT_II="${TMP_DIR}/out-ii.json"
(cd "${TMP_DIR}" && HARNESS_BROWSER_REVIEW_DISABLE_PLAYWRIGHT=1 \
  "${PROJECT_ROOT}/scripts/generate-sprint-contract.sh" "T-browser" "${TMP_DIR}/Plans.md" "${OUT_II}" >/dev/null)
ACTUAL_II="$(jq -r '.review.max_iterations' "${OUT_II}")"
check "browser profile → 5" "5" "${ACTUAL_II}"

echo "=== Case (iii): HTML コメント <!-- max_iterations: 15 --> → 15（明示指定） ==="
OUT_III="${TMP_DIR}/out-iii.json"
(cd "${TMP_DIR}" && "${PROJECT_ROOT}/scripts/generate-sprint-contract.sh" "T-html-comment" "${TMP_DIR}/Plans.md" "${OUT_III}" >/dev/null)
ACTUAL_III="$(jq -r '.review.max_iterations' "${OUT_III}")"
check "HTML コメントマーカーで max_iterations=15 の明示指定" "15" "${ACTUAL_III}"

echo "=== Case (iv): HTML コメント <!-- max_iterations: 100 --> は範囲外 → profile default にフォールバック + stderr 警告 ==="
OUT_IV="${TMP_DIR}/out-iv.json"
STDERR_IV="${TMP_DIR}/stderr-iv.txt"
(cd "${TMP_DIR}" && "${PROJECT_ROOT}/scripts/generate-sprint-contract.sh" "T-out-of-range" "${TMP_DIR}/Plans.md" "${OUT_IV}" 2>"${STDERR_IV}" >/dev/null)
ACTUAL_IV="$(jq -r '.review.max_iterations' "${OUT_IV}")"
check "範囲外 max_iterations=100 → profile default の 3" "3" "${ACTUAL_IV}"
# stderr に警告が出力されているか確認
if grep -q "out of range" "${STDERR_IV}"; then
  echo "  PASS: stderr に範囲外警告が出力されている"
  PASS=$((PASS + 1))
else
  echo "  FAIL: stderr に範囲外警告が出力されていない（内容: $(cat "${STDERR_IV}")）" >&2
  FAIL=$((FAIL + 1))
fi

echo "=== Case (v): 素のテキスト「max_iterations: 15」（HTML コメントなし）→ profile default のまま ==="
OUT_V="${TMP_DIR}/out-v.json"
(cd "${TMP_DIR}" && "${PROJECT_ROOT}/scripts/generate-sprint-contract.sh" "T-plain-text" "${TMP_DIR}/Plans.md" "${OUT_V}" >/dev/null)
ACTUAL_V="$(jq -r '.review.max_iterations' "${OUT_V}")"
check "素のテキスト max_iterations: 15 は無視 → profile default の 3" "3" "${ACTUAL_V}"

echo ""
echo "Results: ${PASS} passed, ${FAIL} failed"

if [ "${FAIL}" -gt 0 ]; then
  exit 1
fi
exit 0
