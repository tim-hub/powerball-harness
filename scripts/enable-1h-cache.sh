#!/usr/bin/env bash
# enable-1h-cache.sh
# ENABLE_PROMPT_CACHING_1H=1 を .env.local に追記する（冪等）。
# CC v2.1.108+ の 1 時間 prompt cache を Harness 長時間セッションで opt-in するためのスクリプト。
#
# 使い方:
#   bash scripts/enable-1h-cache.sh
#
# 効果:
#   - プロジェクトルートの .env.local に ENABLE_PROMPT_CACHING_1H=1 を追記する
#   - すでに設定済みの場合は何もしない（冪等）
#   - .env.local が存在しない場合は新規作成する
#
# 選択基準:
#   - セッション長が 30 分を超える見込みなら 1h cache を選ぶ
#   - 30 分以内の短いやり取りが続くだけなら既定の 5 分 cache で十分
#
# 注意:
#   - .env.local はリポジトリにコミットしない（.gitignore 対象推奨）
#   - グローバル設定は変更しない。このプロジェクトのセッションにのみ適用される

set -euo pipefail

REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || echo "$(cd "$(dirname "$0")/.." && pwd)")"
ENV_LOCAL="${REPO_ROOT}/.env.local"
KEY="ENABLE_PROMPT_CACHING_1H"
VALUE="1"
ENTRY="${KEY}=${VALUE}"

# すでに有効な設定行が存在するか確認（コメント行は無視）
if grep -qE "^${KEY}=${VALUE}$" "${ENV_LOCAL}" 2>/dev/null; then
  echo "[enable-1h-cache] ${ENTRY} はすでに ${ENV_LOCAL} に設定されています（変更なし）。"
  exit 0
fi

# 既存ファイルに同じキーで別の値がある場合は上書きせず警告して終了
if grep -qE "^${KEY}=" "${ENV_LOCAL}" 2>/dev/null; then
  existing_val=$(grep -E "^${KEY}=" "${ENV_LOCAL}" | tail -1)
  echo "[enable-1h-cache] 警告: ${ENV_LOCAL} に既存の設定 '${existing_val}' があります。" >&2
  echo "[enable-1h-cache] 手動で確認してから再実行してください。" >&2
  exit 1
fi

# .env.local に追記（ファイルが存在しない場合は新規作成）
{
  echo ""
  echo "# CC v2.1.108+ の 1 時間 prompt cache（30 分超のセッションで推奨）"
  echo "${ENTRY}"
} >> "${ENV_LOCAL}"

echo "[enable-1h-cache] ${ENTRY} を ${ENV_LOCAL} に追記しました。"
echo "[enable-1h-cache] 次回の長時間セッション（30 分超）から有効になります。"
