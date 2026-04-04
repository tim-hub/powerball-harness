#!/usr/bin/env bash
# codex-companion.sh — Proxy to official codex-plugin-cc companion
#
# 公式プラグイン openai/codex-plugin-cc の codex-companion.mjs を
# 動的に発見して呼び出す。Harness のスキル・エージェントは
# raw `codex exec` ではなく、このプロキシ経由で Codex を呼び出す。
#
# Usage:
#   bash scripts/codex-companion.sh task --write "Fix the bug"
#   bash scripts/codex-companion.sh review --base HEAD~3
#   bash scripts/codex-companion.sh setup --json
#   bash scripts/codex-companion.sh status
#   bash scripts/codex-companion.sh result <job-id>
#   bash scripts/codex-companion.sh cancel <job-id>
#
# Subcommands: task, review, adversarial-review, setup, status, result, cancel
#
# Effort 伝播:
#   task サブコマンド実行時に calculate-effort.sh で effort を計算し、
#   --effort フラグで companion に渡す。calculate-effort.sh がない場合は
#   環境変数 CODEX_EFFORT（未設定時: medium）にフォールバックする。

set -euo pipefail

# 公式プラグインの companion を検索
# Claude/Codex どちらの plugin ディレクトリでも見つかるようにし、
# cache と marketplace 配下の両方を対象にする。
PLUGIN_DIRS=()
[ -d "${HOME}/.claude/plugins" ] && PLUGIN_DIRS+=("${HOME}/.claude/plugins")
[ -d "${HOME}/.codex/plugins" ] && PLUGIN_DIRS+=("${HOME}/.codex/plugins")

COMPANION=""
if [ "${#PLUGIN_DIRS[@]}" -gt 0 ]; then
  # パスからバージョンセグメントを抽出し数値比較（macOS BSD sort 互換）
  COMPANION=$(find "${PLUGIN_DIRS[@]}" -name "codex-companion.mjs" \
    \( -path "*/openai-codex/*" -o -path "*/codex-plugin-cc/*" -o -path "*/plugins/codex/*" \) \
    2>/dev/null \
    | awk -F/ '{version="0.0.0"; for(i=1;i<=NF;i++){if($i~/^[0-9]+\.[0-9]+(\.[0-9]+)?$/){version=$i}} print version,$0}' \
    | sort -t. -k1,1n -k2,2n -k3,3n \
    | tail -1 \
    | cut -d' ' -f2-)
fi

if [ -z "$COMPANION" ]; then
  echo "ERROR: codex-plugin-cc が見つかりません。" >&2
  echo "インストール: plugin marketplace add openai/codex-plugin-cc" >&2
  echo "または: /codex:setup を実行してください" >&2
  exit 1
fi

# ---- Effort 伝播（task サブコマンドのみ）----
# task サブコマンドの場合、タスク説明から effort を計算して --effort フラグで渡す。
# calculate-effort.sh が存在しない場合は CODEX_EFFORT 環境変数（デフォルト: medium）を使う。
SUBCOMMAND="${1:-}"
if [ "$SUBCOMMAND" = "task" ]; then
  SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
  EFFORT_SCRIPT="${SCRIPT_DIR}/calculate-effort.sh"

  # 既に --effort フラグが指定されている場合、または --resume-last の場合はスキップ
  # --resume-last は継続プロンプト（「続きをやって」等）が入るため effort 計算が不正確になる
  EFFORT_ALREADY_SET=0
  for arg in "$@"; do
    if [ "$arg" = "--effort" ] || echo "$arg" | grep -qE '^--effort='; then
      EFFORT_ALREADY_SET=1
      break
    fi
    if [ "$arg" = "--resume-last" ] || [ "$arg" = "--resume" ]; then
      EFFORT_ALREADY_SET=1
      break
    fi
  done

  if [ "$EFFORT_ALREADY_SET" -eq 0 ]; then
    # タスク説明を引数から抽出（最後の非フラグ引数）
    # Boolean フラグ（値を取らない）: --write, --resume-last, --json, --full-auto, --ephemeral, --oss, --skip-git-repo-check
    # 値付きフラグ（次の引数を消費）: --base, --effort, --model, -m, -i, --image, -c, --config, -C, --cd, --add-dir, --output-schema, -o, --output-last-message, --color, --enable, --disable, --local-provider
    # 未知の --* フラグ → 安全側で値付き（次引数を消費）として扱う
    TASK_DESC=""
    EXPECT_VALUE=""
    for arg in "${@:2}"; do
      if [ -n "$EXPECT_VALUE" ]; then
        # 前のフラグの値なのでスキップ
        EXPECT_VALUE=""
        continue
      fi
      case "$arg" in
        --write|--resume-last|--json|--full-auto|--ephemeral|--oss|--skip-git-repo-check|--dangerously-bypass-approvals-and-sandbox|--background|--resume|--fresh)
          # 値を取らない boolean フラグ → スキップするだけ
          ;;
        --base|--effort|--model|-m|-i|--image|-c|--config|-C|--cd|--add-dir|--output-schema|-o|--output-last-message|--color|--enable|--disable|--local-provider)
          # 明示的に値を取るフラグ
          EXPECT_VALUE="$arg"
          ;;
        --*)
          # 未知のフラグ → 安全側で値付きとして扱う（誤って次引数を TASK_DESC にしない）
          EXPECT_VALUE="$arg"
          ;;
        *)
          # 非フラグ引数 = タスク説明
          TASK_DESC="$arg"
          ;;
      esac
    done

    # effort を計算
    COMPUTED_EFFORT=""
    if [ -f "$EFFORT_SCRIPT" ]; then
      if [ -n "$TASK_DESC" ]; then
        COMPUTED_EFFORT=$(bash "$EFFORT_SCRIPT" "$TASK_DESC" 2>/dev/null || true)
      elif [ ! -t 0 ]; then
        # stdin が利用可能（パイプ）: 内容を読み取って effort を計算
        STDIN_CONTENT=$(cat)
        if [ -n "$STDIN_CONTENT" ]; then
          COMPUTED_EFFORT=$(echo "$STDIN_CONTENT" | bash "$EFFORT_SCRIPT" 2>/dev/null || true)
          # stdin を再セットアップ（here-string 経由で companion に渡す）
          exec node "$COMPANION" "$@" --effort "${COMPUTED_EFFORT:-medium}" <<< "$STDIN_CONTENT"
        fi
        # stdin が空の場合（</dev/null 等）はフォールスルーして通常フローへ
      fi
    fi

    # フォールバック: 環境変数 CODEX_EFFORT → medium
    if [ -z "$COMPUTED_EFFORT" ]; then
      COMPUTED_EFFORT="${CODEX_EFFORT:-medium}"
    fi

    # companion がサポートする effort レベルのみ渡す
    case "$COMPUTED_EFFORT" in
      none|minimal|low|medium|high|xhigh) ;;
      *) COMPUTED_EFFORT="medium" ;;
    esac

    exec node "$COMPANION" "$@" --effort "$COMPUTED_EFFORT"
  fi
fi

exec node "$COMPANION" "$@"
