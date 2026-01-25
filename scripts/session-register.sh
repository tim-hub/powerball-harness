#!/bin/bash
# session-register.sh
# セッションを active.json に登録する（出力なし）
#
# 使用方法:
#   ./session-register.sh [session_id]
#
# session_id を省略した場合は .claude/state/session.json から取得
# hook から呼び出される際は出力を抑制し、JSON 出力と混ざらないようにする

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# ===== 設定 =====
SESSIONS_DIR=".claude/sessions"
ACTIVE_FILE="${SESSIONS_DIR}/active.json"
SESSION_FILE=".claude/state/session.json"
STALE_THRESHOLD=3600  # 1時間経過したセッションは stale とみなす

# ===== ヘルパー関数 =====
get_session_id_from_file() {
  if [ -f "$SESSION_FILE" ] && command -v jq >/dev/null 2>&1; then
    jq -r '.session_id // empty' "$SESSION_FILE" 2>/dev/null
  fi
}

get_current_timestamp() {
  date +%s
}

# ===== メイン処理 =====
main() {
  # セッションID を取得（引数優先、なければファイルから）
  local session_id="${1:-}"
  if [ -z "$session_id" ]; then
    session_id=$(get_session_id_from_file)
  fi

  # セッションID がない場合は何もしない（エラーも出さない）
  if [ -z "$session_id" ] || [ "$session_id" = "null" ]; then
    exit 0
  fi

  # jq がない場合は何もしない
  if ! command -v jq >/dev/null 2>&1; then
    exit 0
  fi

  # ディレクトリ作成
  mkdir -p "$SESSIONS_DIR"

  local current_time=$(get_current_timestamp)
  local short_id="${session_id:0:12}"

  # active.json を読み込み（存在しない場合は空オブジェクト）
  local session_data="{}"
  if [ -f "$ACTIVE_FILE" ]; then
    session_data=$(cat "$ACTIVE_FILE" 2>/dev/null || echo "{}")
  fi

  # 一時ファイル用のクリーンアップ設定
  local tmp_file=""
  cleanup_tmp() { [ -n "$tmp_file" ] && [ -f "$tmp_file" ] && rm -f "$tmp_file"; }
  trap cleanup_tmp EXIT

  # セッションを登録/更新
  tmp_file=$(mktemp)
  echo "$session_data" | jq \
    --arg id "$session_id" \
    --arg short "$short_id" \
    --arg time "$current_time" \
    --arg pid "$$" \
    '.[$id] = {
      "short_id": $short,
      "last_seen": ($time | tonumber),
      "pid": $pid,
      "status": "active"
    }' > "$tmp_file" && mv "$tmp_file" "$ACTIVE_FILE"

  # 古いセッションをクリーンアップ（24時間以上経過したもの）
  local cleanup_threshold=$((current_time - 86400))
  tmp_file=$(mktemp)
  jq --arg threshold "$cleanup_threshold" \
    'to_entries | map(select(.value.last_seen > ($threshold | tonumber))) | from_entries' \
    "$ACTIVE_FILE" > "$tmp_file" && mv "$tmp_file" "$ACTIVE_FILE"
}

main "$@"
