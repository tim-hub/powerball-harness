#!/usr/bin/env bash
set -euo pipefail

# 長時間タスク向けに Claude Code を 1 時間 prompt cache 付きで起動する。
# 既定値を変えず、必要なセッションだけ opt-in するための薄いラッパー。

if ! command -v claude >/dev/null 2>&1; then
  echo "ERROR: 'claude' command が見つかりません。" >&2
  echo "Claude Code CLI をインストールしてから再実行してください。" >&2
  exit 1
fi

export ENABLE_PROMPT_CACHING_1H=1

exec claude "$@"
