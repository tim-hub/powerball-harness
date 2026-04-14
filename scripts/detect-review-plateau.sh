#!/bin/bash
# detect-review-plateau.sh
# レビュー修正ループで行き詰まっているかを判定し、Lead に戦略転換（pivot）を促す。
#
# Usage: ./scripts/detect-review-plateau.sh <task_id> [--calibration-file <path>]
#
# Exit codes:
#   0 = PIVOT_NOT_REQUIRED
#   1 = INSUFFICIENT_DATA
#   2 = PIVOT_REQUIRED
#
# Output (stdout):
#   STATUS: PIVOT_REQUIRED | PIVOT_NOT_REQUIRED | INSUFFICIENT_DATA
#   ENTRIES: <N>
#   JACCARD_AVG: <0.XX>  (N>=3 のみ)
#   REASON: <説明>

set -euo pipefail

if ! command -v jq >/dev/null 2>&1; then
  echo "jq is required" >&2
  exit 1
fi

# --- 引数解析 ---
TASK_ID=""
CALIBRATION_FILE=".claude/state/review-calibration.jsonl"

_positional=()
while [ "$#" -gt 0 ]; do
  case "$1" in
    --calibration-file)
      shift
      CALIBRATION_FILE="${1:-}"
      ;;
    --*)
      # 未知のオプションは無視
      ;;
    *)
      _positional+=("$1")
      ;;
  esac
  shift
done

if [ "${#_positional[@]}" -ge 1 ]; then
  TASK_ID="${_positional[0]}"
fi

if [ -z "$TASK_ID" ]; then
  echo "Usage: scripts/detect-review-plateau.sh <task_id> [--calibration-file <path>]" >&2
  exit 1
fi

if [ ! -f "$CALIBRATION_FILE" ]; then
  echo "STATUS: INSUFFICIENT_DATA"
  echo "ENTRIES: 0"
  echo "REASON: calibration file not found: $CALIBRATION_FILE"
  exit 1
fi

# --- 同一 task_id のエントリを抽出（直近 N 件）---
ENTRIES_JSON="$(jq -c --arg tid "$TASK_ID" \
  'select(.task.id == $tid)' \
  "$CALIBRATION_FILE" 2>/dev/null | tail -3)"

ENTRY_COUNT="$(echo "$ENTRIES_JSON" | grep -c '"schema_version"' 2>/dev/null || echo 0)"

# 全件数も取得（件数チェックは全件で行う）
ALL_ENTRIES_JSON="$(jq -c --arg tid "$TASK_ID" \
  'select(.task.id == $tid)' \
  "$CALIBRATION_FILE" 2>/dev/null)"

TOTAL_COUNT="$(echo "$ALL_ENTRIES_JSON" | grep -c '"schema_version"' 2>/dev/null || echo 0)"

# --- N < 3: INSUFFICIENT_DATA ---
if [ "$TOTAL_COUNT" -lt 3 ]; then
  echo "STATUS: INSUFFICIENT_DATA"
  echo "ENTRIES: $TOTAL_COUNT"
  echo "REASON: not enough calibration entries (need >= 3, found $TOTAL_COUNT)"
  exit 1
fi

# --- N >= 3: 直近 3 件を分析 ---

# 各エントリからファイル集合を抽出する関数
# 優先順位:
#   1. review_result_snapshot.files_changed[]
#   2. gaps[].location から ':' 区切り先頭
#   3. どちらもなければ空集合
extract_files() {
  local entry="$1"
  local files=""

  # 1. review_result_snapshot.files_changed
  files="$(echo "$entry" | jq -r '
    (.review_result_snapshot.files_changed // []) | .[]
  ' 2>/dev/null)"

  if [ -z "$files" ]; then
    # 2. gaps[].location から ':' 区切り先頭
    files="$(echo "$entry" | jq -r '
      (.gaps // [])
      | map(select(.location != null and .location != ""))
      | map(.location | split(":")[0])
      | .[]
    ' 2>/dev/null)"
  fi

  echo "$files"
}

# 直近 3 件のエントリを配列に格納
ENTRY1="$(echo "$ENTRIES_JSON" | sed -n '1p')"
ENTRY2="$(echo "$ENTRIES_JSON" | sed -n '2p')"
ENTRY3="$(echo "$ENTRIES_JSON" | sed -n '3p')"

# ファイル集合を抽出（重複排除・ソート済み）
FILES1="$(extract_files "$ENTRY1" | sort -u)"
FILES2="$(extract_files "$ENTRY2" | sort -u)"
FILES3="$(extract_files "$ENTRY3" | sort -u)"

# Jaccard 類似度計算関数
# |A ∩ B| / |A ∪ B|
jaccard() {
  local set_a="$1"
  local set_b="$2"

  # どちらも空なら類似度 1.0（同じ空集合）
  if [ -z "$set_a" ] && [ -z "$set_b" ]; then
    echo "1.0"
    return
  fi

  # 片方のみ空なら類似度 0.0
  if [ -z "$set_a" ] || [ -z "$set_b" ]; then
    echo "0.0"
    return
  fi

  # 共通要素数（intersection）
  local intersection
  intersection="$(comm -12 \
    <(echo "$set_a" | sort -u) \
    <(echo "$set_b" | sort -u) \
    | wc -l | tr -d ' ')"

  # 和集合数（union = |A| + |B| - |intersection|）
  local count_a count_b union
  count_a="$(echo "$set_a" | sort -u | wc -l | tr -d ' ')"
  count_b="$(echo "$set_b" | sort -u | wc -l | tr -d ' ')"
  union=$(( count_a + count_b - intersection ))

  if [ "$union" -eq 0 ]; then
    echo "1.0"
    return
  fi

  # 小数点計算（bash は整数のみなので awk を使用）
  awk "BEGIN { printf \"%.4f\", $intersection / $union }"
}

# 3 ペアの Jaccard 類似度を計算
J12="$(jaccard "$FILES1" "$FILES2")"
J13="$(jaccard "$FILES1" "$FILES3")"
J23="$(jaccard "$FILES2" "$FILES3")"

# 平均 Jaccard
JACCARD_AVG="$(awk "BEGIN { printf \"%.4f\", ($J12 + $J13 + $J23) / 3 }")"

# 条件 (b): 全ペアの平均 Jaccard > 0.7
THRESHOLD="0.7"
IS_PLATEAU="$(awk "BEGIN { print ($JACCARD_AVG > $THRESHOLD) ? \"yes\" : \"no\" }")"

if [ "$IS_PLATEAU" = "yes" ]; then
  echo "STATUS: PIVOT_REQUIRED"
  echo "ENTRIES: $TOTAL_COUNT"
  echo "JACCARD_AVG: $JACCARD_AVG"
  echo "REASON: review iterations >= 3 and file-set similarity (Jaccard avg $JACCARD_AVG) > $THRESHOLD — stuck in same files"
  exit 2
else
  echo "STATUS: PIVOT_NOT_REQUIRED"
  echo "ENTRIES: $TOTAL_COUNT"
  echo "JACCARD_AVG: $JACCARD_AVG"
  echo "REASON: file-set similarity (Jaccard avg $JACCARD_AVG) <= $THRESHOLD — review is making progress"
  exit 0
fi
