#!/bin/bash
# teammate-idle.sh
# TeammateIdle フックハンドラ
# Teammate がアイドル状態になった時にタイムラインに記録する
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
  echo '{"decision":"approve","reason":"TeammateIdle: no payload"}'
  exit 0
fi

# === フィールド抽出 ===
TEAMMATE_NAME=""
TEAM_NAME=""

if command -v jq >/dev/null 2>&1; then
  TEAMMATE_NAME="$(echo "${INPUT}" | jq -r '.teammate_name // .agent_name // ""' 2>/dev/null)"
  TEAM_NAME="$(echo "${INPUT}" | jq -r '.team_name // ""' 2>/dev/null)"
elif command -v python3 >/dev/null 2>&1; then
  _parsed="$(echo "${INPUT}" | python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    print(d.get('teammate_name', d.get('agent_name', '')))
    print(d.get('team_name', ''))
except:
    print('')
    print('')
" 2>/dev/null)"
  TEAMMATE_NAME="$(echo "${_parsed}" | head -1)"
  TEAM_NAME="$(echo "${_parsed}" | tail -1)"
fi

# === タイムライン記録 ===
ensure_state_dir

# JSON 文字列内の特殊文字をエスケープ
TEAMMATE_NAME_ESCAPED="${TEAMMATE_NAME//\"/\\\"}"
TEAM_NAME_ESCAPED="${TEAM_NAME//\"/\\\"}"

log_entry="{\"event\":\"teammate_idle\",\"teammate\":\"${TEAMMATE_NAME_ESCAPED}\",\"team\":\"${TEAM_NAME_ESCAPED}\",\"timestamp\":\"$(get_timestamp)\"}"

echo "${log_entry}" >> "${TIMELINE_FILE}" 2>/dev/null || true

# === レスポンス ===
echo '{"decision":"approve","reason":"TeammateIdle tracked"}'
exit 0
