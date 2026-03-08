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
TOTAL_TASKS=0
COMPLETED_COUNT=0

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
AGENT_ID=""
AGENT_TYPE=""
REQUEST_CONTINUE=""
STOP_REASON=""

if command -v jq >/dev/null 2>&1; then
  TEAMMATE_NAME="$(printf '%s' "${INPUT}" | jq -r '.teammate_name // .agent_name // ""' 2>/dev/null || true)"
  TASK_ID="$(printf '%s' "${INPUT}" | jq -r '.task_id // ""' 2>/dev/null || true)"
  TASK_SUBJECT="$(printf '%s' "${INPUT}" | jq -r '.task_subject // .subject // ""' 2>/dev/null || true)"
  TASK_DESCRIPTION="$(printf '%s' "${INPUT}" | jq -r '(.task_description // .description // "" | tostring)[0:100]' 2>/dev/null || true)"
  AGENT_ID="$(printf '%s' "${INPUT}" | jq -r '.agent_id // ""' 2>/dev/null || true)"
  AGENT_TYPE="$(printf '%s' "${INPUT}" | jq -r '.agent_type // ""' 2>/dev/null || true)"
  REQUEST_CONTINUE="$(printf '%s' "${INPUT}" | jq -r '(if has("continue") then (.continue | tostring) else "" end)' 2>/dev/null || true)"
  STOP_REASON="$(printf '%s' "${INPUT}" | jq -r '.stopReason // .stop_reason // ""' 2>/dev/null || true)"
elif command -v python3 >/dev/null 2>&1; then
  _parsed="$(printf '%s' "${INPUT}" | python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    print(d.get('teammate_name', d.get('agent_name', '')))
    print(d.get('task_id', ''))
    print(d.get('task_subject', d.get('subject', '')))
    print(str(d.get('task_description', d.get('description', '')))[:100])
    print(d.get('agent_id', ''))
    print(d.get('agent_type', ''))
    cont = d.get('continue', '')
    print(str(cont).lower() if isinstance(cont, bool) else str(cont))
    print(d.get('stopReason', d.get('stop_reason', '')))
except:
    print('')
    print('')
    print('')
    print('')
    print('')
    print('')
    print('')
    print('')
" 2>/dev/null)"
  TEAMMATE_NAME="$(echo "${_parsed}" | sed -n '1p')"
  TASK_ID="$(echo "${_parsed}" | sed -n '2p')"
  TASK_SUBJECT="$(echo "${_parsed}" | sed -n '3p')"
  TASK_DESCRIPTION="$(echo "${_parsed}" | sed -n '4p')"
  AGENT_ID="$(echo "${_parsed}" | sed -n '5p')"
  AGENT_TYPE="$(echo "${_parsed}" | sed -n '6p')"
  REQUEST_CONTINUE="$(echo "${_parsed}" | sed -n '7p')"
  STOP_REASON="$(echo "${_parsed}" | sed -n '8p')"
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
    --arg agent_id "${AGENT_ID}" \
    --arg agent_type "${AGENT_TYPE}" \
    --arg timestamp "${TS}" \
    '{event:$event, teammate:$teammate, task_id:$task_id, subject:$subject, description:$description, agent_id:$agent_id, agent_type:$agent_type, timestamp:$timestamp}')"
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
    'agent_id': sys.argv[5],
    'agent_type': sys.argv[6],
    'timestamp': sys.argv[7]
}, ensure_ascii=False))
" "${TEAMMATE_NAME}" "${TASK_ID}" "${TASK_SUBJECT}" "${TASK_DESCRIPTION}" "${AGENT_ID}" "${AGENT_TYPE}" "${TS}" 2>/dev/null)" || log_entry=""
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
  # セッション ID を取得（シグナルのセッションスコープ用）
  SESSION_ID=""
  if command -v jq >/dev/null 2>&1; then
    SESSION_ID="$(jq -r '.session_id // ""' "${BREEZING_ACTIVE}" 2>/dev/null)" || SESSION_ID=""
  elif command -v python3 >/dev/null 2>&1; then
    SESSION_ID="$(python3 -c "
import json, sys
try:
    d = json.load(open(sys.argv[1]))
    print(d.get('session_id', ''))
except:
    print('')
" "${BREEZING_ACTIVE}" 2>/dev/null)" || SESSION_ID=""
  fi

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
      TOTAL_TASKS="$(printf '%s' "${_batch_info}" | cut -f1)"
      CURRENT_BATCH_IDS="$(printf '%s' "${_batch_info}" | cut -f2)"
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
  # grep -F: 固定文字列マッチ（メタ文字リスク回避）
  COMPLETED_COUNT=0
  if [ -f "${TIMELINE_FILE}" ] && [ -n "${CURRENT_BATCH_IDS}" ]; then
    IFS=',' read -ra _batch_id_arr <<< "${CURRENT_BATCH_IDS}"
    for _bid in "${_batch_id_arr[@]}"; do
      _bid="$(printf '%s' "${_bid}" | tr -d '[:space:]')"
      if [ -n "${_bid}" ]; then
        if grep -Fq "\"task_id\":\"${_bid}\"" "${TIMELINE_FILE}" 2>/dev/null; then
          COMPLETED_COUNT=$(( COMPLETED_COUNT + 1 ))
        fi
      fi
    done
    unset _batch_id_arr _bid
  elif [ -f "${TIMELINE_FILE}" ]; then
    # フォールバック: バッチ ID が取得できない場合は全体カウント
    COMPLETED_COUNT="$(grep -Fc '"event":"task_completed"' "${TIMELINE_FILE}" 2>/dev/null)" || COMPLETED_COUNT=0
  fi

  # シグナル dedup ヘルパー: セッションスコープで重複チェック
  # SESSION_ID があればセッション単位（同一レコード内で AND 判定）、なければグローバルで dedup
  _signal_exists() {
    local sig_type="$1"
    if [ -n "${SESSION_ID}" ] && [ -f "${SIGNALS_FILE}" ]; then
      # 同一行に signal と session_id の両方を含むレコードが存在するか
      grep -F "\"${sig_type}\"" "${SIGNALS_FILE}" 2>/dev/null | grep -Fq "\"session_id\":\"${SESSION_ID}\"" 2>/dev/null
    else
      grep -Fq "\"${sig_type}\"" "${SIGNALS_FILE}" 2>/dev/null
    fi
  }

  # シグナル JSON 構築ヘルパー
  _build_signal_json() {
    local sig_type="$1" completed="$2" total="$3" ts="$4"
    if command -v jq >/dev/null 2>&1; then
      jq -nc \
        --arg signal "${sig_type}" \
        --arg session_id "${SESSION_ID}" \
        --arg completed "${completed}" \
        --arg total "${total}" \
        --arg timestamp "${ts}" \
        '{signal:$signal, session_id:$session_id, completed:$completed, total:$total, timestamp:$timestamp}'
    elif command -v python3 >/dev/null 2>&1; then
      python3 -c "
import json, sys
print(json.dumps({
    'signal': sys.argv[1],
    'session_id': sys.argv[2],
    'completed': sys.argv[3],
    'total': sys.argv[4],
    'timestamp': sys.argv[5]
}, ensure_ascii=False))
" "${sig_type}" "${SESSION_ID}" "${completed}" "${total}" "${ts}" 2>/dev/null
    fi
  }

  # 50% 完了シグナル: 部分レビュー推奨
  # -ge で閾値飛び越え（同時完了等）に対応、既発行チェックで重複防止
  # 切り上げ計算で閾値が早すぎるトリガーを防止
  # HALF>1 ガード: バッチサイズ 1-2 では部分レビューは不要（全体レビューで十分）
  if [ "${TOTAL_TASKS}" -gt 0 ] 2>/dev/null; then
    HALF=$(( (TOTAL_TASKS + 1) / 2 ))
    if [ "${COMPLETED_COUNT}" -ge "${HALF}" ] && [ "${HALF}" -gt 1 ] 2>/dev/null; then
      if ! _signal_exists "partial_review_recommended"; then
        SIGNAL_ENTRY="$(_build_signal_json "partial_review_recommended" "${COMPLETED_COUNT}" "${TOTAL_TASKS}" "${TS}")" || SIGNAL_ENTRY=""
        if [ -n "${SIGNAL_ENTRY}" ]; then
          printf '%s\n' "${SIGNAL_ENTRY}" >> "${SIGNALS_FILE}" 2>/dev/null || true
        fi
      fi
    fi

    # 60% 完了シグナル: 次バッチ登録推奨（Progressive Batch 用）
    # 切り上げ: (n * 60 + 99) / 100 で端数切り捨てによる早期トリガーを防止
    SIXTY_PCT=$(( (TOTAL_TASKS * 60 + 99) / 100 ))
    if [ "${COMPLETED_COUNT}" -ge "${SIXTY_PCT}" ] && [ "${SIXTY_PCT}" -gt 0 ] 2>/dev/null; then
      if ! _signal_exists "next_batch_recommended"; then
        BATCH_SIGNAL="$(_build_signal_json "next_batch_recommended" "${COMPLETED_COUNT}" "${TOTAL_TASKS}" "${TS}")" || BATCH_SIGNAL=""
        if [ -n "${BATCH_SIGNAL}" ]; then
          printf '%s\n' "${BATCH_SIGNAL}" >> "${SIGNALS_FILE}" 2>/dev/null || true
        fi
      fi
    fi
  fi
fi

# === テスト結果参照ロジック ===
# auto-test-runner の結果ファイルを確認し、未実行/失敗時は exit 2 でエスカレーション
TEST_RESULT_FILE="${STATE_DIR}/test-result.json"
QUALITY_GATE_FILE="${STATE_DIR}/task-quality-gate.json"

# タスク ID 別の失敗カウントを更新
# $1: task_id, $2: "increment"|"reset"
update_failure_count() {
  local tid="$1"
  local action="$2"
  [ -z "${tid}" ] && return 0

  if command -v jq >/dev/null 2>&1; then
    local current_count=0
    if [ -f "${QUALITY_GATE_FILE}" ]; then
      current_count="$(jq -r --arg tid "${tid}" '.[$tid].failure_count // 0' "${QUALITY_GATE_FILE}" 2>/dev/null)" || current_count=0
    fi
    local new_count=0
    if [ "${action}" = "increment" ]; then
      new_count=$(( current_count + 1 ))
    fi
    # ファイル更新（存在しない場合は新規作成）
    local existing="{}"
    if [ -f "${QUALITY_GATE_FILE}" ]; then
      existing="$(cat "${QUALITY_GATE_FILE}" 2>/dev/null)" || existing="{}"
    fi
    jq -n \
      --argjson existing "${existing}" \
      --arg tid "${tid}" \
      --argjson count "${new_count}" \
      --arg ts "${TS}" \
      --arg action "${action}" \
      '$existing * {($tid): {"failure_count": $count, "last_action": $action, "updated_at": $ts}}' \
      > "${QUALITY_GATE_FILE}.tmp" 2>/dev/null && \
      mv "${QUALITY_GATE_FILE}.tmp" "${QUALITY_GATE_FILE}" 2>/dev/null || true
    echo "${new_count}"
  elif command -v python3 >/dev/null 2>&1; then
    python3 -c "
import json, sys
tid, action, ts = sys.argv[1], sys.argv[2], sys.argv[3]
path = sys.argv[4]
try:
    data = json.load(open(path))
except Exception:
    data = {}
entry = data.get(tid, {})
count = entry.get('failure_count', 0)
if action == 'increment':
    count += 1
else:
    count = 0
data[tid] = {'failure_count': count, 'last_action': action, 'updated_at': ts}
json.dump(data, open(path, 'w'), ensure_ascii=False, indent=2)
print(count)
" "${tid}" "${action}" "${TS}" "${QUALITY_GATE_FILE}" 2>/dev/null || echo "0"
  else
    echo "0"
  fi
}

check_test_result() {
  # 結果ファイルが存在しない場合: テスト未実行
  if [ ! -f "${TEST_RESULT_FILE}" ]; then
    return 0  # 未実行はスキップ（テストが不要なプロジェクトも存在するため）
  fi

  # 結果ファイルの読み取り
  local status=""
  if command -v jq >/dev/null 2>&1; then
    status="$(jq -r '.status // ""' "${TEST_RESULT_FILE}" 2>/dev/null)" || status=""
  elif command -v python3 >/dev/null 2>&1; then
    status="$(python3 -c "
import json, sys
try:
    d = json.load(open(sys.argv[1]))
    print(d.get('status', ''))
except:
    print('')
" "${TEST_RESULT_FILE}" 2>/dev/null)" || status=""
  fi

  # status=failed の場合は失敗として検知
  if [ "${status}" = "failed" ]; then
    return 1
  fi

  return 0
}

if ! check_test_result; then
  # テスト失敗: 失敗カウントをインクリメント
  FAIL_COUNT="$(update_failure_count "${TASK_ID}" "increment")"

  # タイムラインにも記録
  FAIL_ENTRY=""
  if command -v jq >/dev/null 2>&1; then
    FAIL_ENTRY="$(jq -nc \
      --arg event "test_result_failed" \
      --arg teammate "${TEAMMATE_NAME}" \
      --arg task_id "${TASK_ID}" \
      --arg subject "${TASK_SUBJECT}" \
      --arg timestamp "${TS}" \
      --arg failure_count "${FAIL_COUNT}" \
      '{event:$event, teammate:$teammate, task_id:$task_id, subject:$subject, timestamp:$timestamp, failure_count:$failure_count}')"
  fi
  if [ -n "${FAIL_ENTRY}" ]; then
    echo "${FAIL_ENTRY}" >> "${TIMELINE_FILE}" 2>/dev/null || true
  fi

  # 3回連続失敗でエスカレーション（D21 自動化: exit 0 + stderr 出力）
  ESCALATION_THRESHOLD=3
  if [ "${FAIL_COUNT}" -ge "${ESCALATION_THRESHOLD}" ] 2>/dev/null; then
    # テスト結果から原因分類・推奨アクション・試行履歴を収集
    _last_cmd=""
    _last_output=""
    _failure_category="unknown"
    _recommended_action=""

    if [ -f "${TEST_RESULT_FILE}" ]; then
      if command -v jq >/dev/null 2>&1; then
        _last_cmd="$(jq -r '.command // ""' "${TEST_RESULT_FILE}" 2>/dev/null)" || _last_cmd=""
        _last_output="$(jq -r '.output // ""' "${TEST_RESULT_FILE}" 2>/dev/null | head -20)" || _last_output=""
      elif command -v python3 >/dev/null 2>&1; then
        _parsed_result="$(python3 -c "
import json, sys
try:
    d = json.load(open(sys.argv[1]))
    print(d.get('command', ''))
    print('---OUTPUT---')
    out = d.get('output', '')
    print('\n'.join(str(out).split('\n')[:20]))
except Exception as e:
    print('')
    print('---OUTPUT---')
" "${TEST_RESULT_FILE}" 2>/dev/null)"
        _last_cmd="$(echo "${_parsed_result}" | head -1)"
        _last_output="$(echo "${_parsed_result}" | tail -n +3)"
      fi
    fi

    # 原因分類（テスト出力のキーワードから推定）
    if echo "${_last_output}" | grep -qi "syntax\|SyntaxError\|parse error\|unexpected token"; then
      _failure_category="syntax_error"
      _recommended_action="構文エラーを修正してください。コードの文法を確認してください。"
    elif echo "${_last_output}" | grep -qi "cannot find module\|module not found\|import.*error\|ModuleNotFoundError"; then
      _failure_category="import_error"
      _recommended_action="モジュール/インポートエラーを修正してください。依存関係を確認してください（npm install / pip install）。"
    elif echo "${_last_output}" | grep -qi "type.*error\|TypeError\|is not assignable\|Property.*does not exist"; then
      _failure_category="type_error"
      _recommended_action="型エラーを修正してください。型定義と実装の不一致を確認してください。"
    elif echo "${_last_output}" | grep -qi "assertion\|AssertionError\|expect.*received\|toBe\|toEqual\|FAIL\|FAILED"; then
      _failure_category="assertion_error"
      _recommended_action="テストアサーションが失敗しています。期待値と実際の値の差分を確認してください。"
    elif echo "${_last_output}" | grep -qi "timeout\|Timeout\|ETIMEDOUT\|timed out"; then
      _failure_category="timeout"
      _recommended_action="タイムアウトが発生しました。非同期処理やネットワーク依存を確認してください。"
    elif echo "${_last_output}" | grep -qi "permission\|EACCES\|EPERM\|access denied"; then
      _failure_category="permission_error"
      _recommended_action="権限エラーが発生しています。ファイルのパーミッションを確認してください。"
    else
      _failure_category="runtime_error"
      _recommended_action="ランタイムエラーが発生しています。テスト出力を詳しく確認してください。"
    fi

    # 試行履歴を quality-gate ファイルから取得
    _history_summary=""
    if [ -f "${QUALITY_GATE_FILE}" ]; then
      if command -v jq >/dev/null 2>&1; then
        _history_summary="$(jq -r --arg tid "${TASK_ID}" '
          .[$tid] |
          "  失敗カウント: \(.failure_count // 0)\n  最終更新: \(.updated_at // "不明")"
        ' "${QUALITY_GATE_FILE}" 2>/dev/null)" || _history_summary=""
      fi
    fi

    # エスカレーションレポートを stderr に出力
    {
      echo ""
      echo "=========================================="
      echo "[ESCALATION] 3回連続失敗を検知 - 自動修正ループを停止"
      echo "=========================================="
      echo "  タスク ID  : ${TASK_ID}"
      echo "  タスク名   : ${TASK_SUBJECT}"
      echo "  担当者     : ${TEAMMATE_NAME}"
      echo "  連続失敗数 : ${FAIL_COUNT}"
      echo "  検知時刻   : ${TS}"
      echo "------------------------------------------"
      echo "  [原因分類]"
      echo "  カテゴリ   : ${_failure_category}"
      echo ""
      echo "  [推奨アクション]"
      echo "  ${_recommended_action}"
      echo ""
      if [ -n "${_last_cmd}" ]; then
        echo "  [最後に実行したコマンド]"
        echo "  ${_last_cmd}"
        echo ""
      fi
      if [ -n "${_last_output}" ]; then
        echo "  [テスト出力（最大20行）]"
        echo "${_last_output}" | while IFS= read -r _line; do
          echo "    ${_line}"
        done
        echo ""
      fi
      if [ -n "${_history_summary}" ]; then
        echo "  [試行履歴]"
        echo "${_history_summary}"
        echo ""
      fi
      echo "  詳細ファイル: ${QUALITY_GATE_FILE}"
      echo "  手動対応が必要です。エスカレーション記録後、ループを終了します。"
      echo "=========================================="
      echo ""
    } >&2 2>/dev/null || true

    # エスカレーション記録をタイムラインに追記
    ESC_ENTRY=""
    if command -v jq >/dev/null 2>&1; then
      ESC_ENTRY="$(jq -nc \
        --arg event "escalation_triggered" \
        --arg teammate "${TEAMMATE_NAME}" \
        --arg task_id "${TASK_ID}" \
        --arg subject "${TASK_SUBJECT}" \
        --arg timestamp "${TS}" \
        --arg failure_count "${FAIL_COUNT}" \
        '{event:$event, teammate:$teammate, task_id:$task_id, subject:$subject, timestamp:$timestamp, failure_count:$failure_count}')"
    fi
    if [ -n "${ESC_ENTRY}" ]; then
      echo "${ESC_ENTRY}" >> "${TIMELINE_FILE}" 2>/dev/null || true
    fi

    # exit 0 でフックを承認（自動修正ループを止めた後の復旧は手動）
    echo '{"decision":"approve","reason":"TaskCompleted: 3-strike escalation triggered - manual intervention required"}'
    exit 0
  fi

  echo '{"decision":"block","reason":"TaskCompleted: test result shows failure - escalation required"}'
  exit 2
else
  # テスト成功またはテスト未実行: 失敗カウントをリセット
  if [ -n "${TASK_ID}" ]; then
    update_failure_count "${TASK_ID}" "reset" >/dev/null 2>&1 || true
  fi
fi

# === レスポンス ===
if [ "${REQUEST_CONTINUE}" = "false" ] || [ -n "${STOP_REASON}" ]; then
  FINAL_STOP_REASON="${STOP_REASON:-TaskCompleted requested stop}"
  if command -v jq >/dev/null 2>&1; then
    jq -nc --arg reason "${FINAL_STOP_REASON}" '{"continue": false, "stopReason": $reason}'
  else
    printf '{"continue": false, "stopReason": "%s"}\n' "${FINAL_STOP_REASON//\"/\\\"}"
  fi
  exit 0
fi

if [ "${TOTAL_TASKS}" -gt 0 ] 2>/dev/null && [ "${COMPLETED_COUNT}" -ge "${TOTAL_TASKS}" ] 2>/dev/null; then
  if command -v jq >/dev/null 2>&1; then
    jq -nc --arg reason "all_tasks_completed" '{"continue": false, "stopReason": $reason}'
  else
    echo '{"continue": false, "stopReason": "all_tasks_completed"}'
  fi
  exit 0
fi

# プログレスサマリー付きレスポンス
if [ "${TOTAL_TASKS}" -gt 0 ] 2>/dev/null && [ -n "${TASK_SUBJECT:-}" ]; then
  PROGRESS_MSG="📊 Progress: Task ${COMPLETED_COUNT}/${TOTAL_TASKS} 完了 — \"${TASK_SUBJECT}\""
  if command -v jq >/dev/null 2>&1; then
    jq -nc \
      --arg reason "TaskCompleted tracked" \
      --arg msg "${PROGRESS_MSG}" \
      '{"decision":"approve","reason":$reason,"systemMessage":$msg}'
  else
    # jq がない場合のフォールバック（特殊文字をエスケープ）
    _escaped_msg="${PROGRESS_MSG//\"/\\\"}"
    printf '{"decision":"approve","reason":"TaskCompleted tracked","systemMessage":"%s"}\n' "${_escaped_msg}"
  fi
else
  echo '{"decision":"approve","reason":"TaskCompleted tracked"}'
fi
exit 0
