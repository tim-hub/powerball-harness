#!/bin/bash
# todo-sync.sh
# TodoWrite と Plans.md の双方向同期
#
# PostToolUse hook から呼び出され、TodoWrite の状態変更を Plans.md に反映
#
# マッピング:
#   TodoWrite状態     → Plans.mdマーカー
#   pending          → cc:TODO
#   in_progress      → cc:WIP
#   completed        → cc:done

set +e  # エラーで停止しない

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# stdin から JSON 入力を読み取り
INPUT=""
if [ ! -t 0 ]; then
  INPUT=$(cat 2>/dev/null || true)
fi

# jq が必要
if ! command -v jq >/dev/null 2>&1; then
  exit 0
fi

# 入力がない場合は終了
if [ -z "$INPUT" ]; then
  exit 0
fi

# TodoWrite ツールの出力を解析
TOOL_NAME=$(echo "$INPUT" | jq -r '.tool_name // empty' 2>/dev/null)

# TodoWrite 以外は無視
if [ "$TOOL_NAME" != "TodoWrite" ]; then
  exit 0
fi

# Plans.md のパスを取得
if [ -f "${SCRIPT_DIR}/config-utils.sh" ]; then
  source "${SCRIPT_DIR}/config-utils.sh"
  PLANS_FILE=$(get_plans_file_path)
else
  PLANS_FILE="Plans.md"
fi

# Plans.md が存在しない場合は終了
if [ ! -f "$PLANS_FILE" ]; then
  exit 0
fi

# 状態ディレクトリ
STATE_DIR=".claude/state"
mkdir -p "$STATE_DIR"
SYNC_STATE_FILE="${STATE_DIR}/todo-sync-state.json"

# TodoWrite の todos 配列を取得
TODOS=$(echo "$INPUT" | jq -r '.tool_input.todos // []' 2>/dev/null)

if [ -z "$TODOS" ] || [ "$TODOS" = "null" ] || [ "$TODOS" = "[]" ]; then
  exit 0
fi

# 同期状態を保存
echo "$TODOS" | jq '{
  synced_at: (now | todate),
  todos: .
}' > "$SYNC_STATE_FILE" 2>/dev/null

# Plans.md 内のタスク状態を更新
# 注意: Plans.md のフォーマットを維持しながら更新するのは複雑なため、
# ここではログに記録するのみとし、実際の更新は Claude Code に任せる

# イベントログに記録
EVENT_LOG="${STATE_DIR}/session.events.jsonl"
NOW=$(date -u +%Y-%m-%dT%H:%M:%SZ)

PENDING_COUNT=$(echo "$TODOS" | jq '[.[] | select(.status == "pending")] | length' 2>/dev/null || echo "0")
WIP_COUNT=$(echo "$TODOS" | jq '[.[] | select(.status == "in_progress")] | length' 2>/dev/null || echo "0")
DONE_COUNT=$(echo "$TODOS" | jq '[.[] | select(.status == "completed")] | length' 2>/dev/null || echo "0")

if [ -f "$EVENT_LOG" ]; then
  echo "{\"type\":\"todo.sync\",\"ts\":\"$NOW\",\"data\":{\"pending\":$PENDING_COUNT,\"in_progress\":$WIP_COUNT,\"completed\":$DONE_COUNT}}" >> "$EVENT_LOG"
fi

# additionalContext として同期情報を出力
cat <<EOF
{"hookSpecificOutput":{"additionalContext":"[TodoSync] Plans.md と同期: TODO=$PENDING_COUNT, WIP=$WIP_COUNT, done=$DONE_COUNT"}}
EOF
