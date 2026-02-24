#!/bin/bash
# setup-hook.sh
# Setup Hook: claude --init / --maintenance 時のセットアップ処理
#
# Usage:
#   setup-hook.sh init        # 初回セットアップ
#   setup-hook.sh maintenance # メンテナンス処理
#
# 出力: JSON形式で hookSpecificOutput を出力

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MODE="${1:-init}"

# ===== SIMPLE モード検出 =====
SIMPLE_MODE="false"
if [ -f "$SCRIPT_DIR/check-simple-mode.sh" ]; then
  # shellcheck source=./check-simple-mode.sh
  source "$SCRIPT_DIR/check-simple-mode.sh"
  if is_simple_mode; then
    SIMPLE_MODE="true"
    echo -e "\033[1;33m[WARNING]\033[0m CLAUDE_CODE_SIMPLE mode detected — skills/agents/memory disabled" >&2
  fi
fi

# stdin から JSON 入力を読み取り（Claude Code v2.1.10+）
INPUT=""
if [ ! -t 0 ]; then
  INPUT=$(cat 2>/dev/null || true)
fi

# ===== 共通ヘルパー =====
output_json() {
  local message="$1"
  cat <<EOF
{"hookSpecificOutput":{"hookEventName":"Setup","additionalContext":"$message"}}
EOF
}

# ===== Init モード: 初回セットアップ =====
run_init() {
  local messages=()

  # 1. プラグインキャッシュの同期
  if [ -f "$SCRIPT_DIR/sync-plugin-cache.sh" ]; then
    bash "$SCRIPT_DIR/sync-plugin-cache.sh" >/dev/null 2>&1 || true
    messages+=("プラグインキャッシュ同期完了")
  fi

  # 2. 状態ディレクトリの初期化
  STATE_DIR=".claude/state"
  mkdir -p "$STATE_DIR"

  # 3. デフォルト設定ファイルの生成（存在しない場合）
  CONFIG_FILE=".claude-code-harness.config.yaml"
  if [ ! -f "$CONFIG_FILE" ]; then
    if [ -f "$SCRIPT_DIR/../templates/.claude-code-harness.config.yaml.template" ]; then
      cp "$SCRIPT_DIR/../templates/.claude-code-harness.config.yaml.template" "$CONFIG_FILE"
      messages+=("設定ファイル生成完了")
    fi
  fi

  # 4. CLAUDE.md の生成（存在しない場合）
  if [ ! -f "CLAUDE.md" ]; then
    if [ -f "$SCRIPT_DIR/../templates/CLAUDE.md.template" ]; then
      cp "$SCRIPT_DIR/../templates/CLAUDE.md.template" "CLAUDE.md"
      messages+=("CLAUDE.md 生成完了")
    fi
  fi

  # 5. Plans.md の生成（存在しない場合）
  # plansDirectory 設定を考慮
  if [ -f "$SCRIPT_DIR/config-utils.sh" ]; then
    source "$SCRIPT_DIR/config-utils.sh"
    PLANS_PATH=$(get_plans_file_path)
  else
    PLANS_PATH="Plans.md"
  fi

  if [ ! -f "$PLANS_PATH" ]; then
    # ディレクトリが存在しない場合は作成
    PLANS_DIR=$(dirname "$PLANS_PATH")
    [ "$PLANS_DIR" != "." ] && mkdir -p "$PLANS_DIR"

    if [ -f "$SCRIPT_DIR/../templates/Plans.md.template" ]; then
      cp "$SCRIPT_DIR/../templates/Plans.md.template" "$PLANS_PATH"
      messages+=("Plans.md 生成完了")
    fi
  fi

  # 6. テンプレートトラッカーの初期化
  if [ -f "$SCRIPT_DIR/template-tracker.sh" ]; then
    bash "$SCRIPT_DIR/template-tracker.sh" init >/dev/null 2>&1 || true
  fi

  # SIMPLE モード警告を追加
  if [ "$SIMPLE_MODE" = "true" ]; then
    messages+=("WARNING: CLAUDE_CODE_SIMPLE mode — skills/agents/memory disabled, hooks only")
  fi

  # 結果出力
  if [ ${#messages[@]} -eq 0 ]; then
    output_json "[Setup:init] ハーネスは既に初期化済みです"
  else
    local msg_str
    msg_str=$(IFS=', '; echo "${messages[*]}")
    output_json "[Setup:init] $msg_str"
  fi
}

# ===== Maintenance モード: メンテナンス処理 =====
run_maintenance() {
  local messages=()

  # 1. プラグインキャッシュの同期
  if [ -f "$SCRIPT_DIR/sync-plugin-cache.sh" ]; then
    bash "$SCRIPT_DIR/sync-plugin-cache.sh" >/dev/null 2>&1 || true
    messages+=("キャッシュ同期完了")
  fi

  # 2. 古いセッションファイルのクリーンアップ
  STATE_DIR=".claude/state"
  ARCHIVE_DIR="$STATE_DIR/sessions"

  if [ -d "$ARCHIVE_DIR" ]; then
    # 7日以上前のセッションアーカイブを削除
    find "$ARCHIVE_DIR" -name "session-*.json" -mtime +7 -delete 2>/dev/null || true
    messages+=("古いセッションアーカイブ削除")
  fi

  # 3. 一時ファイルのクリーンアップ
  if [ -d "$STATE_DIR" ]; then
    # .tmp ファイルを削除
    find "$STATE_DIR" -name "*.tmp" -delete 2>/dev/null || true
  fi

  # 4. テンプレート更新チェック
  if [ -f "$SCRIPT_DIR/template-tracker.sh" ]; then
    CHECK_RESULT=$(bash "$SCRIPT_DIR/template-tracker.sh" check 2>/dev/null || echo '{"needsCheck": false}')
    if command -v jq >/dev/null 2>&1; then
      NEEDS_UPDATE=$(echo "$CHECK_RESULT" | jq -r '.needsCheck // false')
      if [ "$NEEDS_UPDATE" = "true" ]; then
        UPDATES_COUNT=$(echo "$CHECK_RESULT" | jq -r '.updatesCount // 0')
        messages+=("テンプレート更新あり: ${UPDATES_COUNT}件")
      fi
    fi
  fi

  # 5. SIMPLE モード警告を追加
  if [ "$SIMPLE_MODE" = "true" ]; then
    messages+=("WARNING: CLAUDE_CODE_SIMPLE mode — skills/agents/memory disabled, hooks only")
  fi

  # 6. 設定ファイルの検証
  CONFIG_FILE=".claude-code-harness.config.yaml"
  if [ -f "$CONFIG_FILE" ]; then
    # 基本的な YAML 構文チェック
    if command -v python3 >/dev/null 2>&1; then
      if ! python3 -c "import yaml; yaml.safe_load(open('$CONFIG_FILE'))" 2>/dev/null; then
        messages+=("警告: 設定ファイルの構文エラー")
      fi
    fi
  fi

  # 結果出力
  if [ ${#messages[@]} -eq 0 ]; then
    output_json "[Setup:maintenance] メンテナンス完了（変更なし）"
  else
    local msg_str
    msg_str=$(IFS=', '; echo "${messages[*]}")
    output_json "[Setup:maintenance] $msg_str"
  fi
}

# ===== メイン処理 =====
case "$MODE" in
  init)
    run_init
    ;;
  maintenance)
    run_maintenance
    ;;
  *)
    output_json "[Setup] 不明なモード: $MODE"
    exit 1
    ;;
esac
