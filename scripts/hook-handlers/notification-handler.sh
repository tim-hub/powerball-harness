#!/usr/bin/env bash
# notification-handler.sh
# Notification フックハンドラ
# Claude Code が通知を発行する際に発火
# permission_prompt, idle_prompt, auth_success 等のイベントを記録
#
# Input: stdin JSON from Claude Code hooks
# Output: JSON to approve the event
# Hook event: Notification

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

# ログファイル
STATE_DIR="${PROJECT_ROOT}/.claude/state"
LOG_FILE="${STATE_DIR}/notification-events.jsonl"

# === ユーティリティ関数 ===

ensure_state_dir() {
  local state_parent
  state_parent="$(dirname "${STATE_DIR}")"

  # Security: refuse symlinked state paths to avoid overwriting arbitrary files.
  if [ -L "${state_parent}" ] || [ -L "${STATE_DIR}" ]; then
    return 1
  fi

  mkdir -p "${STATE_DIR}" 2>/dev/null || true
  chmod 700 "${STATE_DIR}" 2>/dev/null || true

  [ -d "${STATE_DIR}" ] || return 1
  [ ! -L "${STATE_DIR}" ] || return 1
  return 0
}

# JSONL ローテーション（500 行超過時に 400 行に切り詰め）
rotate_jsonl() {
  local file="$1"

  # Security: refuse symlinked log or tmp files
  if [ -L "${file}" ] || [ -L "${file}.tmp" ]; then
    return 1
  fi

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
  exit 0
fi

# === フィールド抽出 ===
NOTIFICATION_TYPE=""
SESSION_ID=""
AGENT_TYPE=""

if command -v jq >/dev/null 2>&1; then
  NOTIFICATION_TYPE="$(printf '%s' "${INPUT}" | jq -r '.notification_type // .type // .matcher // ""' 2>/dev/null || true)"
  SESSION_ID="$(printf '%s' "${INPUT}" | jq -r '.session_id // ""' 2>/dev/null || true)"
  AGENT_TYPE="$(printf '%s' "${INPUT}" | jq -r '.agent_type // ""' 2>/dev/null || true)"
elif command -v python3 >/dev/null 2>&1; then
  _parsed="$(printf '%s' "${INPUT}" | python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    print(d.get('notification_type', d.get('type', d.get('matcher', ''))))
    print(d.get('session_id', ''))
    print(d.get('agent_type', ''))
except:
    print('')
    print('')
    print('')
" 2>/dev/null)"
  NOTIFICATION_TYPE="$(echo "${_parsed}" | sed -n '1p')"
  SESSION_ID="$(echo "${_parsed}" | sed -n '2p')"
  AGENT_TYPE="$(echo "${_parsed}" | sed -n '3p')"
fi

# === ログ記録 ===
if ! ensure_state_dir; then
  exit 0
fi
TS="$(get_timestamp)"

log_entry=""
if command -v jq >/dev/null 2>&1; then
  log_entry="$(jq -nc \
    --arg event "notification" \
    --arg notification_type "${NOTIFICATION_TYPE}" \
    --arg session_id "${SESSION_ID}" \
    --arg agent_type "${AGENT_TYPE}" \
    --arg timestamp "${TS}" \
    '{event:$event, notification_type:$notification_type, session_id:$session_id, agent_type:$agent_type, timestamp:$timestamp}')"
elif command -v python3 >/dev/null 2>&1; then
  log_entry="$(python3 -c "
import json, sys
print(json.dumps({
    'event': 'notification',
    'notification_type': sys.argv[1],
    'session_id': sys.argv[2],
    'agent_type': sys.argv[3],
    'timestamp': sys.argv[4]
}, ensure_ascii=False))
" "${NOTIFICATION_TYPE}" "${SESSION_ID}" "${AGENT_TYPE}" "${TS}" 2>/dev/null)" || log_entry=""
fi

if [ -n "${log_entry}" ]; then
  # Security: refuse symlinked log file
  if [ -L "${LOG_FILE}" ]; then
    exit 0
  fi
  echo "${log_entry}" >> "${LOG_FILE}" 2>/dev/null || true
  rotate_jsonl "${LOG_FILE}"
fi

# === Breezing 中の permission_prompt 検出 ===
# Breezing のバックグラウンド Worker では permission prompt に応答不能
# ログに記録することで事後分析を可能にする
if [ "${NOTIFICATION_TYPE}" = "permission_prompt" ] && [ -n "${AGENT_TYPE}" ]; then
  echo "Notification: permission_prompt for agent_type=${AGENT_TYPE}" >&2
fi

exit 0
