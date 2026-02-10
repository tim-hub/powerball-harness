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

if command -v jq >/dev/null 2>&1; then
  _jq_parsed="$(echo "${INPUT}" | jq -r '[
    (.teammate_name // .agent_name // ""),
    (.task_id // ""),
    (.task_subject // .subject // "")
  ] | @tsv' 2>/dev/null)"
  if [ -n "${_jq_parsed}" ]; then
    IFS=$'\t' read -r TEAMMATE_NAME TASK_ID TASK_SUBJECT <<< "${_jq_parsed}"
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
except:
    print('')
    print('')
    print('')
" 2>/dev/null)"
  TEAMMATE_NAME="$(echo "${_parsed}" | sed -n '1p')"
  TASK_ID="$(echo "${_parsed}" | sed -n '2p')"
  TASK_SUBJECT="$(echo "${_parsed}" | sed -n '3p')"
fi

# === タイムライン記録 ===
ensure_state_dir

# JSON 文字列内の特殊文字をエスケープ
TEAMMATE_NAME_ESCAPED="${TEAMMATE_NAME//\"/\\\"}"
TASK_ID_ESCAPED="${TASK_ID//\"/\\\"}"
TASK_SUBJECT_ESCAPED="${TASK_SUBJECT//\"/\\\"}"

log_entry="{\"event\":\"task_completed\",\"teammate\":\"${TEAMMATE_NAME_ESCAPED}\",\"task_id\":\"${TASK_ID_ESCAPED}\",\"subject\":\"${TASK_SUBJECT_ESCAPED}\",\"timestamp\":\"$(get_timestamp)\"}"

echo "${log_entry}" >> "${TIMELINE_FILE}" 2>/dev/null || true

# === レスポンス ===
echo '{"decision":"approve","reason":"TaskCompleted tracked"}'
exit 0
