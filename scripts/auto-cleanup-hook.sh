#!/bin/bash
# auto-cleanup-hook.sh
# PostToolUse Hook: Plans.md 等への書き込み後に自動でサイズチェック
#
# 入力: stdin から JSON（tool_name, tool_input 等）
# 出力: additionalContext でフィードバック

set +e

# 入力JSONを読み取り（Claude Code hooks は stdin で JSON を渡す）
INPUT=""
if [ ! -t 0 ]; then
  INPUT="$(cat 2>/dev/null)"
fi

# stdin JSON から file_path / cwd を取得（jq がなければ python3 を試す）
FILE_PATH=""
CWD=""
if [ -n "$INPUT" ]; then
  if command -v jq >/dev/null 2>&1; then
    FILE_PATH="$(echo "$INPUT" | jq -r '.tool_input.file_path // .tool_response.filePath // empty' 2>/dev/null)"
    CWD="$(echo "$INPUT" | jq -r '.cwd // empty' 2>/dev/null)"
  elif command -v python3 >/dev/null 2>&1; then
    eval "$(echo "$INPUT" | python3 - <<'PY' 2>/dev/null
import json, shlex, sys
try:
    data = json.load(sys.stdin)
except Exception:
    data = {}
cwd = data.get("cwd") or ""
tool_input = data.get("tool_input") or {}
tool_response = data.get("tool_response") or {}
file_path = tool_input.get("file_path") or tool_response.get("filePath") or ""
print(f"CWD_FROM_STDIN={shlex.quote(cwd)}")
print(f"FILE_PATH_FROM_STDIN={shlex.quote(file_path)}")
PY
)"
    FILE_PATH="${FILE_PATH_FROM_STDIN:-}"
    CWD="${CWD_FROM_STDIN:-}"
  fi
fi

# file_path が空なら終了
if [ -z "$FILE_PATH" ]; then
  exit 0
fi

# 可能ならプロジェクト相対パスへ正規化（絶対パスでも動作するが判定が安定する）
if [ -n "$CWD" ] && [[ "$FILE_PATH" == "$CWD/"* ]]; then
  FILE_PATH="${FILE_PATH#$CWD/}"
fi

# デフォルト閾値
PLANS_MAX_LINES=${PLANS_MAX_LINES:-200}
SESSION_LOG_MAX_LINES=${SESSION_LOG_MAX_LINES:-500}
CLAUDE_MD_MAX_LINES=${CLAUDE_MD_MAX_LINES:-100}

# フィードバックを格納する変数
FEEDBACK=""

# Plans.md のチェック
if [[ "$FILE_PATH" == *"Plans.md"* ]] || [[ "$FILE_PATH" == *"plans.md"* ]]; then
  if [ -f "$FILE_PATH" ]; then
    lines=$(wc -l < "$FILE_PATH" | tr -d ' ')
    if [ "$lines" -gt "$PLANS_MAX_LINES" ]; then
      FEEDBACK="⚠️ Plans.md が ${lines} 行です（上限: ${PLANS_MAX_LINES}行）。/maintenance で古いタスクをアーカイブすることを推奨します。"
    fi

    # Plans.md クリーンアップ（アーカイブ移動）検知時の SSOT 同期チェック
    # アーカイブセクションへの編集がある場合、/memory sync の事前実行を確認
    if grep -q "📦 アーカイブ\|## アーカイブ\|Archive" "$FILE_PATH" 2>/dev/null; then
      # Resolve repository root for consistent state directory lookup
      CWD="${CWD:-$(pwd)}"  # Fallback to pwd if empty
      REPO_ROOT=$(git -C "$CWD" rev-parse --show-toplevel 2>/dev/null) || REPO_ROOT="$CWD"
      STATE_DIR="${REPO_ROOT}/.claude/state"

      SSOT_FLAG="${STATE_DIR}/.ssot-synced-this-session"

      if [ ! -f "$SSOT_FLAG" ]; then
        # フラグがない場合、SSOT 同期を促す警告を追加
        SSOT_WARNING="**Plans.md クリーンアップ前に /memory sync を実行してください** - 重要な決定や学習事項が SSOT (decisions.md/patterns.md) に反映されていない可能性があります。"

        if [ -n "$FEEDBACK" ]; then
          FEEDBACK="${FEEDBACK} | ${SSOT_WARNING}"
        else
          FEEDBACK="⚠️ ${SSOT_WARNING}"
        fi
      fi
    fi
  fi
fi

# session-log.md のチェック
if [[ "$FILE_PATH" == *"session-log.md"* ]]; then
  if [ -f "$FILE_PATH" ]; then
    lines=$(wc -l < "$FILE_PATH" | tr -d ' ')
    if [ "$lines" -gt "$SESSION_LOG_MAX_LINES" ]; then
      FEEDBACK="⚠️ session-log.md が ${lines} 行です（上限: ${SESSION_LOG_MAX_LINES}行）。/maintenance で月別に分割することを推奨します。"
    fi
  fi
fi

# CLAUDE.md のチェック
if [[ "$FILE_PATH" == *"CLAUDE.md"* ]] || [[ "$FILE_PATH" == *"claude.md"* ]]; then
  if [ -f "$FILE_PATH" ]; then
    lines=$(wc -l < "$FILE_PATH" | tr -d ' ')
    if [ "$lines" -gt "$CLAUDE_MD_MAX_LINES" ]; then
      FEEDBACK="⚠️ CLAUDE.md が ${lines} 行です。.claude/rules/ への分割、または docs/ に移動して @docs/filename.md で参照することを検討してください。"
    fi
  fi
fi

# フィードバックがあれば JSON で出力
if [ -n "$FEEDBACK" ]; then
  echo "{\"hookSpecificOutput\": {\"hookEventName\": \"PostToolUse\", \"additionalContext\": \"$FEEDBACK\"}}"
fi

# 常に成功で終了
exit 0
