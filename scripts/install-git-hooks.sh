#!/usr/bin/env bash
#
# install-git-hooks.sh
# Repo-managed git hooks installer (uses core.hooksPath).
#
# Usage:
#   ./scripts/install-git-hooks.sh
#
# Windows:
#   Requires Git for Windows (includes Git Bash).
#   Run from Git Bash, WSL, or PowerShell.
#

set -euo pipefail

ROOT="$(git rev-parse --show-toplevel 2>/dev/null || true)"
if [ -z "$ROOT" ]; then
  echo "エラー: Git リポジトリではありません"
  exit 1
fi

cd "$ROOT"

if [ ! -d ".githooks" ]; then
  echo "エラー: .githooks/ ディレクトリが見つかりません"
  exit 1
fi

chmod +x .githooks/pre-commit 2>/dev/null || true

git config core.hooksPath .githooks

echo ""
echo "=== Git Hooks 有効化完了 ==="
echo ""
echo "  core.hooksPath = .githooks"
echo ""
echo "  pre-commit:"
echo "    - release metadata を編集したときに VERSION と plugin.json を同期"
echo "    - 通常のコード変更では version を自動 bump しない"
echo ""

# Windows 向け注意事項
if [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "cygwin" ]] || [[ -n "${WINDIR:-}" ]]; then
  echo "  [Windows 注意]"
  echo "    Git hooks は Git Bash（Git for Windows に付属）で実行されます。"
  echo "    hooks が動作しない場合は Git for Windows をインストールしてください:"
  echo "    https://gitforwindows.org/"
  echo ""
fi

echo "完了！"
