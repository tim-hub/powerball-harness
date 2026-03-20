#!/bin/bash
# show-failures.sh
# StopFailure ログの集計サマリーを表示
#
# stop-failures.jsonl を読み込み、エラーコード別集計・直近5件・推奨アクションを出力。
# harness-sync --show-failures から呼び出される。スタンドアロンでも実行可能。
#
# Usage: bash scripts/show-failures.sh [--days N] [--json]
#   --days N  集計対象の日数（デフォルト: 30）
#   --json    JSON 形式で出力（パイプライン用）

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PLUGIN_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# path-utils.sh の読み込み
if [ -f "${SCRIPT_DIR}/path-utils.sh" ]; then
  source "${SCRIPT_DIR}/path-utils.sh"
fi

# プロジェクトルートを検出
if declare -F detect_project_root > /dev/null 2>&1; then
  PROJECT_ROOT="${PROJECT_ROOT:-$(detect_project_root 2>/dev/null || pwd)}"
else
  PROJECT_ROOT="${PROJECT_ROOT:-$(pwd)}"
fi

# ステートディレクトリ（CLAUDE_PLUGIN_DATA 対応）
if [ -n "${CLAUDE_PLUGIN_DATA:-}" ]; then
  _project_hash="$(printf '%s' "${PROJECT_ROOT}" | { shasum -a 256 2>/dev/null || sha256sum 2>/dev/null || echo "default  -"; } | cut -c1-12)"
  [ -z "${_project_hash}" ] && _project_hash="default"
  STATE_DIR="${CLAUDE_PLUGIN_DATA}/projects/${_project_hash}"
else
  STATE_DIR="${PROJECT_ROOT}/.claude/state"
fi
LOG_FILE="${STATE_DIR}/stop-failures.jsonl"

# === 引数パース ===
DAYS=30
JSON_OUTPUT=false

while [ $# -gt 0 ]; do
  case "$1" in
    --days)
      if [ $# -lt 2 ] || ! [[ ${2} =~ ^[0-9]+$ ]]; then
        echo "エラー: --days には正の整数を指定してください" >&2
        exit 1
      fi
      DAYS="$2"; shift 2 ;;
    --json) JSON_OUTPUT=true; shift ;;
    *) shift ;;
  esac
done

# === ログファイル存在チェック ===
if [ ! -f "${LOG_FILE}" ] || [ ! -s "${LOG_FILE}" ]; then
  if [ "${JSON_OUTPUT}" = true ]; then
    echo '{"total":0,"entries":[],"summary":"No StopFailure events recorded."}'
  else
    echo "StopFailure ログがありません（${LOG_FILE}）"
    echo ""
    echo "これは良いニュースです — API エラーによるセッション停止失敗が発生していません。"
  fi
  exit 0
fi

# === jq が必要 ===
if ! command -v jq > /dev/null 2>&1; then
  # jq なしの場合は行数だけ表示
  LINE_COUNT=$(wc -l < "${LOG_FILE}" | tr -d ' ')
  echo "StopFailure ログ: ${LINE_COUNT} 件（詳細表示には jq が必要です）"
  echo "ログファイル: ${LOG_FILE}"
  exit 0
fi

# === 集計 ===
CUTOFF_DATE=$(date -u -v-${DAYS}d +"%Y-%m-%dT" 2>/dev/null || date -u -d "${DAYS} days ago" +"%Y-%m-%dT" 2>/dev/null || echo "")

# 全件 or 期間フィルタ
if [ -n "${CUTOFF_DATE}" ]; then
  FILTERED=$(jq -c "select(.timestamp >= \"${CUTOFF_DATE}\")" "${LOG_FILE}" 2>/dev/null || cat "${LOG_FILE}")
else
  FILTERED=$(cat "${LOG_FILE}")
fi

TOTAL=$(echo "${FILTERED}" | grep -c '^{' 2>/dev/null || echo "0")

if [ "${TOTAL}" -eq 0 ]; then
  if [ "${JSON_OUTPUT}" = true ]; then
    echo '{"total":0,"entries":[],"summary":"No events in the specified period."}'
  else
    echo "直近 ${DAYS} 日間の StopFailure イベント: 0 件"
  fi
  exit 0
fi

# エラーコード別集計
COUNT_429=$(echo "${FILTERED}" | jq -r 'select(.error_code == "429") | .error_code' 2>/dev/null | wc -l | tr -d ' ')
COUNT_401=$(echo "${FILTERED}" | jq -r 'select(.error_code == "401") | .error_code' 2>/dev/null | wc -l | tr -d ' ')
COUNT_500=$(echo "${FILTERED}" | jq -r 'select(.error_code == "500") | .error_code' 2>/dev/null | wc -l | tr -d ' ')
COUNT_OTHER=$(( TOTAL - COUNT_429 - COUNT_401 - COUNT_500 ))
[ "${COUNT_OTHER}" -lt 0 ] && COUNT_OTHER=0

# 直近 5 件
RECENT=$(echo "${FILTERED}" | tail -5 | jq -r '[.timestamp, .error_code, .session_id, .message] | join(" | ")' 2>/dev/null || echo "(parse error)")

# === 出力 ===
if [ "${JSON_OUTPUT}" = true ]; then
  jq -nc \
    --argjson total "${TOTAL}" \
    --argjson c429 "${COUNT_429}" \
    --argjson c401 "${COUNT_401}" \
    --argjson c500 "${COUNT_500}" \
    --argjson cother "${COUNT_OTHER}" \
    --argjson days "${DAYS}" \
    '{
      total: $total,
      period_days: $days,
      by_code: { "429": $c429, "401": $c401, "500": $c500, other: $cother }
    }'
else
  echo "StopFailure サマリー（直近 ${DAYS} 日）"
  echo "========================================"
  echo ""
  echo "合計: ${TOTAL} 件"
  echo ""
  echo "エラー分布:"
  [ "${COUNT_429}" -gt 0 ] && echo "  429 (Rate Limit): ${COUNT_429} 回"
  [ "${COUNT_401}" -gt 0 ] && echo "  401 (Auth):       ${COUNT_401} 回"
  [ "${COUNT_500}" -gt 0 ] && echo "  500 (Server):     ${COUNT_500} 回"
  [ "${COUNT_OTHER}" -gt 0 ] && echo "  その他:           ${COUNT_OTHER} 回"
  [ "${TOTAL}" -eq 0 ] && echo "  （イベントなし）"
  echo ""
  echo "直近 5 件:"
  echo "${RECENT}" | while IFS= read -r line; do
    [ -n "${line}" ] && echo "  ${line}"
  done
  echo ""

  # 推奨アクション
  if [ "${COUNT_429}" -ge 5 ]; then
    echo "推奨: 429 エラーが多発しています。Breezing の並列 Worker 数を削減してください。"
  elif [ "${COUNT_429}" -ge 1 ]; then
    echo "情報: 429 エラーが ${COUNT_429} 回発生しています。頻発するなら Worker 数の調整を検討してください。"
  fi
  if [ "${COUNT_401}" -ge 1 ]; then
    echo "推奨: 認証エラーが発生しています。claude auth login で認証を更新してください。"
  fi
fi
