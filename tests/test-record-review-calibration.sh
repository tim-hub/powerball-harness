#!/bin/bash
# test-record-review-calibration.sh
# record-review-calibration.sh の smoke テスト
#
# テスト一覧:
#   1. arg parsing: 入力ファイルなしで exit 1
#   2. arg parsing: 存在しないファイルで exit 3
#   3. arg parsing: calibration なしで正常終了（exit 0、出力なし）
#   4. arg parsing: 無効な label で exit 4
#   5. arg parsing: --review-result フラグが positional を汚染しない
#   6. critical_issues[] と gaps[severity:critical] の両方をカウント
#   7. findings[severity:high] 2件 → major_count = 2
#   8. gaps[severity:major] 1件 + findings[severity:high] 1件 → major_count = 2

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
SCRIPT="${PROJECT_ROOT}/scripts/record-review-calibration.sh"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

pass=0
fail=0

assert_eq() {
  local desc="$1" expected="$2" actual="$3"
  if [ "$expected" = "$actual" ]; then
    echo "  PASS: $desc"
    pass=$((pass + 1))
  else
    echo "  FAIL: $desc"
    echo "        expected: $expected"
    echo "        actual:   $actual"
    fail=$((fail + 1))
  fi
}

# ---- 共通入力ファイルの準備 ----

# calibration 付き最小入力
cat > "${TMP_DIR}/with-cal.json" <<'EOF'
{
  "schema_version": "review-result.v1",
  "verdict": "REQUEST_CHANGES",
  "reviewer_profile": "static",
  "calibration": {
    "label": "false_positive",
    "source": "manual",
    "notes": "テスト用",
    "prompt_hint": "",
    "few_shot_ready": true
  },
  "gaps": []
}
EOF

# calibration なし入力
cat > "${TMP_DIR}/no-cal.json" <<'EOF'
{
  "schema_version": "review-result.v1",
  "verdict": "APPROVE",
  "reviewer_profile": "static"
}
EOF

# 無効 label 入力
cat > "${TMP_DIR}/bad-label.json" <<'EOF'
{
  "schema_version": "review-result.v1",
  "verdict": "REQUEST_CHANGES",
  "reviewer_profile": "static",
  "calibration": {
    "label": "unknown_label",
    "source": "manual",
    "notes": "",
    "prompt_hint": "",
    "few_shot_ready": true
  }
}
EOF

# ---- テスト 1: 入力ファイルなしで exit 1 ----
actual_exit=0
"$SCRIPT" 2>/dev/null || actual_exit=$?
assert_eq "test-1: 引数なし → exit 1" "1" "$actual_exit"

# ---- テスト 2: 存在しないファイルで exit 3 ----
actual_exit=0
"$SCRIPT" "${TMP_DIR}/nonexistent.json" "${TMP_DIR}/out.jsonl" 2>/dev/null || actual_exit=$?
assert_eq "test-2: 存在しないファイル → exit 3" "3" "$actual_exit"

# ---- テスト 3: calibration なしで exit 0 ----
actual_exit=0
"$SCRIPT" "${TMP_DIR}/no-cal.json" "${TMP_DIR}/out3.jsonl" 2>/dev/null || actual_exit=$?
assert_eq "test-3: calibration なし → exit 0" "0" "$actual_exit"
# 出力ファイルが作られないこと
if [ ! -f "${TMP_DIR}/out3.jsonl" ]; then
  echo "  PASS: test-3b: 出力ファイルが作られない"
  pass=$((pass + 1))
else
  echo "  FAIL: test-3b: 出力ファイルが作られてしまった"
  fail=$((fail + 1))
fi

# ---- テスト 4: 無効 label で exit 4 ----
actual_exit=0
"$SCRIPT" "${TMP_DIR}/bad-label.json" "${TMP_DIR}/out4.jsonl" 2>/dev/null || actual_exit=$?
assert_eq "test-4: 無効 label → exit 4" "4" "$actual_exit"

# ---- テスト 5: --review-result フラグが positional を汚染しない ----
# --review-result を入力の前に置いても INPUT_FILE が正しく認識されること
actual_exit=0
"$SCRIPT" --review-result "${TMP_DIR}/with-cal.json" "${TMP_DIR}/with-cal.json" "${TMP_DIR}/out5.jsonl" 2>/dev/null || actual_exit=$?
assert_eq "test-5: --review-result フラグが positional を汚染しない → exit 0" "0" "$actual_exit"

# ---- テスト 6: critical_issues[] と gaps[severity:critical] の両方をカウント ----
cat > "${TMP_DIR}/dual-critical.json" <<'EOF'
{
  "schema_version": "review-result.v1",
  "verdict": "REQUEST_CHANGES",
  "reviewer_profile": "static",
  "calibration": {
    "label": "missed_bug",
    "source": "manual",
    "notes": "",
    "prompt_hint": "",
    "few_shot_ready": true
  },
  "critical_issues": [
    "旧形式 critical issue"
  ],
  "gaps": [
    {"severity": "critical", "issue": "normalized critical gap"},
    {"severity": "major",    "issue": "normalized major gap"}
  ]
}
EOF

OUT6="${TMP_DIR}/out6.jsonl"
"$SCRIPT" "${TMP_DIR}/dual-critical.json" "$OUT6" >/dev/null
actual_critical="$(jq -r '.critical_count' "$OUT6")"
actual_major="$(jq -r '.major_count' "$OUT6")"
assert_eq "test-6a: critical_issues[1] + gaps[critical][1] → critical_count = 2" "2" "$actual_critical"
assert_eq "test-6b: gaps[major][1] → major_count = 1" "1" "$actual_major"

# ---- テスト 7: findings[severity:high] 2件 → major_count = 2 ----
cat > "${TMP_DIR}/high-findings.json" <<'EOF'
{
  "schema_version": "review-result.v1",
  "verdict": "REQUEST_CHANGES",
  "reviewer_profile": "static",
  "calibration": {
    "label": "false_negative",
    "source": "manual",
    "notes": "",
    "prompt_hint": "",
    "few_shot_ready": true
  },
  "findings": [
    {"severity": "high",   "title": "blocking finding 1"},
    {"severity": "high",   "title": "blocking finding 2"},
    {"severity": "medium", "title": "non-blocking finding"}
  ]
}
EOF

OUT7="${TMP_DIR}/out7.jsonl"
"$SCRIPT" "${TMP_DIR}/high-findings.json" "$OUT7" >/dev/null
actual_major7="$(jq -r '.major_count' "$OUT7")"
actual_critical7="$(jq -r '.critical_count' "$OUT7")"
assert_eq "test-7a: findings[high][2] → major_count = 2" "2" "$actual_major7"
assert_eq "test-7b: findings[medium] → critical_count = 0" "0" "$actual_critical7"

# ---- テスト 8: gaps[major] 1件 + findings[high] 1件 → major_count = 2（各ソースの合算） ----
cat > "${TMP_DIR}/mixed-major.json" <<'EOF'
{
  "schema_version": "review-result.v1",
  "verdict": "REQUEST_CHANGES",
  "reviewer_profile": "static",
  "calibration": {
    "label": "overstrict_rule",
    "source": "manual",
    "notes": "",
    "prompt_hint": "",
    "few_shot_ready": true
  },
  "gaps": [
    {"severity": "major", "issue": "normalized major from gaps"}
  ],
  "findings": [
    {"severity": "high", "title": "raw high from companion"}
  ]
}
EOF

OUT8="${TMP_DIR}/out8.jsonl"
"$SCRIPT" "${TMP_DIR}/mixed-major.json" "$OUT8" >/dev/null
actual_major8="$(jq -r '.major_count' "$OUT8")"
assert_eq "test-8: gaps[major][1] + findings[high][1] → major_count = 2" "2" "$actual_major8"

# ---- 結果集計 ----
echo ""
echo "test-record-review-calibration: ${pass} passed, ${fail} failed"
if [ "$fail" -gt 0 ]; then
  exit 1
fi
