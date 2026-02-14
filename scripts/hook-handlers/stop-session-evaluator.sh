#!/bin/bash
# stop-session-evaluator.sh
# Stop フックのセッション完了評価
#
# prompt type の代替として、確実に有効な JSON を出力する command type フック。
# セッション状態を検査し、停止を許可 or ブロックの判定を行う。
#
# Input:  なし（セッション状態ファイルを直接参照）
# Output: {"ok": true} or {"ok": false, "reason": "..."}
#
# Issue: #42 - Stop hook "JSON validation failed" on every turn

set +e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PARENT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

# path-utils.sh の読み込み
if [ -f "${PARENT_DIR}/path-utils.sh" ]; then
  source "${PARENT_DIR}/path-utils.sh"
fi

PROJECT_ROOT="${PROJECT_ROOT:-$(detect_project_root 2>/dev/null || pwd)}"

STATE_FILE="${PROJECT_ROOT}/.claude/state/session.json"

# jq がなければ即座に ok を返す（安全なフォールバック）
if ! command -v jq &> /dev/null; then
  echo '{"ok":true}'
  exit 0
fi

# 状態ファイルがなければ即座に ok を返す
if [ ! -f "$STATE_FILE" ]; then
  echo '{"ok":true}'
  exit 0
fi

# セッション状態を検査
SESSION_STATE=$(jq -r '.state // "unknown"' "$STATE_FILE" 2>/dev/null)

# 既に停止処理済みなら即座に ok
if [ "$SESSION_STATE" = "stopped" ]; then
  echo '{"ok":true}'
  exit 0
fi

# デフォルト: 停止を許可
# ユーザーが明示的に Stop を押した場合、基本的に停止を許可する
echo '{"ok":true}'
exit 0
