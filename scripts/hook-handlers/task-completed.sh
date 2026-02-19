#!/bin/bash
# task-completed.sh
# TaskCompleted フックハンドラ
# タスクが完了した時にタイムラインに記録する
#
# Input: stdin JSON from Claude Code hooks
# Output: JSON to approve the event

set -euo pipefail

# === 設定 ===
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PARENT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

# path-utils.sh の読み込み
if [ -f "${PARENT_DIR}/path-utils.sh" ]; then
  source "${PARENT_DIR}/path-utils.sh"
fi

# プロジェクトルートを検出
PROJECT_ROOT="${PROJECT_ROOT:-$(detect_project_root 2>/dev/null || pwd)}"

# タイムラインファイル
STATE_DIR="${PROJECT_ROOT}/.claude/state"
TIMELINE_FILE="${STATE_DIR}/breezing-timeline.jsonl"

# === ユーティリティ関数 ===

ensure_state_dir() {
  mkdir -p "${STATE_DIR}" 2>/dev/null || true
  chmod 700 "${STATE_DIR}" 2>/dev/null || true
}

# JSONL ローテーション（500 行超過時に 400 行に切り詰め）
rotate_jsonl() {
  local file="$1"
  local _lines
  _lines="$(wc -l < "${file}" 2>/dev/null)" || _lines=0
  if [ "${_lines}" -gt 500 ] 2>/dev/null; then
    tail -400 "${file}" > "${file}.tmp" 2>/dev/null && \
      mv "${file}.tmp" "${file}" 2>/dev/null || true
  fi
}

get_timestamp() {
  date -u +"%Y-%m-%dT%H:%M:%SZ"
}

# === stdin から JSON ペイロードを読み取り ===
INPUT=""
if [ ! -t 0 ]; then
  INPUT="$(cat 2>/dev/null)"
fi

# ペイロードが空の場合はスキップ
if [ -z "${INPUT}" ]; then
  echo '{"decision":"approve","reason":"TaskCompleted: no payload"}'
  exit 0
fi

# === フィールド抽出 ===
TEAMMATE_NAME=""
TASK_ID=""
TASK_SUBJECT=""
TASK_DESCRIPTION=""

if command -v jq >/dev/null 2>&1; then
  _jq_parsed="$(echo "${INPUT}" | jq -r '[
    (.teammate_name // .agent_name // ""),
    (.task_id // ""),
    (.task_subject // .subject // ""),
    ((.task_description // .description // "" | tostring)[0:100])
  ] | @tsv' 2>/dev/null)"
  if [ -n "${_jq_parsed}" ]; then
    IFS=$'\t' read -r TEAMMATE_NAME TASK_ID TASK_SUBJECT TASK_DESCRIPTION <<< "${_jq_parsed}"
  fi
  unset _jq_parsed
elif command -v python3 >/dev/null 2>&1; then
  _parsed="$(echo "${INPUT}" | python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    print(d.get('teammate_name', d.get('agent_name', '')))
    print(d.get('task_id', ''))
    print(d.get('task_subject', d.get('subject', '')))
    print(str(d.get('task_description', d.get('description', '')))[:100])
except:
    print('')
    print('')
    print('')
    print('')
" 2>/dev/null)"
  TEAMMATE_NAME="$(echo "${_parsed}" | sed -n '1p')"
  TASK_ID="$(echo "${_parsed}" | sed -n '2p')"
  TASK_SUBJECT="$(echo "${_parsed}" | sed -n '3p')"
  TASK_DESCRIPTION="$(echo "${_parsed}" | sed -n '4p')"
fi

# === タイムライン記録（jq -nc で安全な JSON 構築） ===
ensure_state_dir
TS="$(get_timestamp)"

if command -v jq >/dev/null 2>&1; then
  log_entry="$(jq -nc \
    --arg event "task_completed" \
    --arg teammate "${TEAMMATE_NAME}" \
    --arg task_id "${TASK_ID}" \
    --arg subject "${TASK_SUBJECT}" \
    --arg description "${TASK_DESCRIPTION}" \
    --arg timestamp "${TS}" \
    '{event:$event, teammate:$teammate, task_id:$task_id, subject:$subject, description:$description, timestamp:$timestamp}')"
else
  # フォールバック: python3 で安全にエスケープ
  log_entry="$(python3 -c "
import json, sys
print(json.dumps({
    'event': 'task_completed',
    'teammate': sys.argv[1],
    'task_id': sys.argv[2],
    'subject': sys.argv[3],
    'description': sys.argv[4],
    'timestamp': sys.argv[5]
}, ensure_ascii=False))
" "${TEAMMATE_NAME}" "${TASK_ID}" "${TASK_SUBJECT}" "${TASK_DESCRIPTION}" "${TS}" 2>/dev/null)" || log_entry=""
fi

if [ -n "${log_entry}" ]; then
  echo "${log_entry}" >> "${TIMELINE_FILE}" 2>/dev/null || true
  rotate_jsonl "${TIMELINE_FILE}"
fi

# === シグナル生成（動的オーケストレーション） ===
SIGNALS_FILE="${STATE_DIR}/breezing-signals.jsonl"
BREEZING_ACTIVE="${STATE_DIR}/breezing-active.json"

# breezing セッションがアクティブな場合のみシグナルを生成
if [ -f "${BREEZING_ACTIVE}" ]; then
  # breezing-active.json から現在バッチのタスク ID リストと合計タスク数を取得
  TOTAL_TASKS=0
  CURRENT_BATCH_IDS=""
  if command -v jq >/dev/null 2>&1; then
    # Progressive Batching: 現在の in_progress バッチから取得
    _batch_info="$(jq -r '
      (.batching.batches // [] | map(select(.status == "in_progress")) | .[0].task_ids // []) as $ids |
      ($ids | length) as $len |
      ($ids | join(",")) as $csv |
      "\($len)\t\($csv)"
    ' "${BREEZING_ACTIVE}" 2>/dev/null)" || _batch_info=""
    if [ -n "${_batch_info}" ]; then
      TOTAL_TASKS="$(echo "${_batch_info}" | cut -f1)"
      CURRENT_BATCH_IDS="$(echo "${_batch_info}" | cut -f2)"
    fi
    # batching でない場合は plans_md_mapping のキー数で推定
    if [ "${TOTAL_TASKS}" = "0" ] || [ "${TOTAL_TASKS}" = "null" ] || [ -z "${TOTAL_TASKS}" ]; then
      TOTAL_TASKS="$(jq -r '.plans_md_mapping // {} | keys | length' "${BREEZING_ACTIVE}" 2>/dev/null)" || TOTAL_TASKS=0
      # 非バッチモード: plans_md_mapping のキーをバッチ ID として使用
      CURRENT_BATCH_IDS="$(jq -r '.plans_md_mapping // {} | keys | join(",")' "${BREEZING_ACTIVE}" 2>/dev/null)" || CURRENT_BATCH_IDS=""
    fi
    unset _batch_info
  fi

  # 現在バッチのタスクのみで完了数をカウント（前バッチの完了を除外）
  # 各タスク ID につき存在すれば +1（リテイク等による重複カウントを防止）
  COMPLETED_COUNT=0
  if [ -f "${TIMELINE_FILE}" ] && [ -n "${CURRENT_BATCH_IDS}" ]; then
    IFS=',' read -ra _batch_id_arr <<< "${CURRENT_BATCH_IDS}"
    for _bid in "${_batch_id_arr[@]}"; do
      _bid="$(echo "${_bid}" | tr -d '[:space:]')"
      if [ -n "${_bid}" ]; then
        if grep -q "\"task_id\":\"${_bid}\"" "${TIMELINE_FILE}" 2>/dev/null; then
          COMPLETED_COUNT=$(( COMPLETED_COUNT + 1 ))
        fi
      fi
    done
    unset _batch_id_arr _bid
  elif [ -f "${TIMELINE_FILE}" ]; then
    # フォールバック: バッチ ID が取得できない場合は全体カウント
    COMPLETED_COUNT="$(grep -c '"event":"task_completed"' "${TIMELINE_FILE}" 2>/dev/null)" || COMPLETED_COUNT=0
  fi

  # 50% 完了シグナル: 部分レビュー推奨
  # -ge で閾値飛び越え（同時完了等）に対応、既発行チェックで重複防止
  # 切り上げ計算で閾値が早すぎるトリガーを防止
  if [ "${TOTAL_TASKS}" -gt 0 ] 2>/dev/null; then
    HALF=$(( (TOTAL_TASKS + 1) / 2 ))
    if [ "${COMPLETED_COUNT}" -ge "${HALF}" ] && [ "${HALF}" -gt 1 ] 2>/dev/null; then
      # 既にシグナル発行済みか確認（重複防止）
      if ! grep -q '"partial_review_recommended"' "${SIGNALS_FILE}" 2>/dev/null; then
        SIGNAL_ENTRY=""
        if command -v jq >/dev/null 2>&1; then
          SIGNAL_ENTRY="$(jq -nc \
            --arg signal "partial_review_recommended" \
            --arg completed "${COMPLETED_COUNT}" \
            --arg total "${TOTAL_TASKS}" \
            --arg timestamp "${TS}" \
            '{signal:$signal, completed:$completed, total:$total, timestamp:$timestamp}')"
        elif command -v python3 >/dev/null 2>&1; then
          SIGNAL_ENTRY="$(python3 -c "
import json, sys
print(json.dumps({
    'signal': 'partial_review_recommended',
    'completed': sys.argv[1],
    'total': sys.argv[2],
    'timestamp': sys.argv[3]
}, ensure_ascii=False))
" "${COMPLETED_COUNT}" "${TOTAL_TASKS}" "${TS}" 2>/dev/null)" || SIGNAL_ENTRY=""
        fi
        if [ -n "${SIGNAL_ENTRY}" ]; then
          echo "${SIGNAL_ENTRY}" >> "${SIGNALS_FILE}" 2>/dev/null || true
        fi
      fi
    fi

    # 60% 完了シグナル: 次バッチ登録推奨（Progressive Batch 用）
    # 切り上げ: (n * 60 + 99) / 100 で端数切り捨てによる早期トリガーを防止
    SIXTY_PCT=$(( (TOTAL_TASKS * 60 + 99) / 100 ))
    if [ "${COMPLETED_COUNT}" -ge "${SIXTY_PCT}" ] && [ "${SIXTY_PCT}" -gt 0 ] 2>/dev/null; then
      # 既にシグナル発行済みか確認（重複防止）
      if ! grep -q '"next_batch_recommended"' "${SIGNALS_FILE}" 2>/dev/null; then
        BATCH_SIGNAL=""
        if command -v jq >/dev/null 2>&1; then
          BATCH_SIGNAL="$(jq -nc \
            --arg signal "next_batch_recommended" \
            --arg completed "${COMPLETED_COUNT}" \
            --arg total "${TOTAL_TASKS}" \
            --arg timestamp "${TS}" \
            '{signal:$signal, completed:$completed, total:$total, timestamp:$timestamp}')"
        elif command -v python3 >/dev/null 2>&1; then
          BATCH_SIGNAL="$(python3 -c "
import json, sys
print(json.dumps({
    'signal': 'next_batch_recommended',
    'completed': sys.argv[1],
    'total': sys.argv[2],
    'timestamp': sys.argv[3]
}, ensure_ascii=False))
" "${COMPLETED_COUNT}" "${TOTAL_TASKS}" "${TS}" 2>/dev/null)" || BATCH_SIGNAL=""
        fi
        if [ -n "${BATCH_SIGNAL}" ]; then
          echo "${BATCH_SIGNAL}" >> "${SIGNALS_FILE}" 2>/dev/null || true
        fi
      fi
    fi
  fi
fi

# === レスポンス ===
echo '{"decision":"approve","reason":"TaskCompleted tracked"}'
exit 0
