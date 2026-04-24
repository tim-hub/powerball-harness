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
TEMPLATE_DIR="$SCRIPT_DIR/../templates"

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

if [ -f "$SCRIPT_DIR/config-utils.sh" ]; then
  # shellcheck source=./config-utils.sh
  source "$SCRIPT_DIR/config-utils.sh"
fi

# ===== 共通ヘルパー =====
output_json() {
  local message="$1"
  cat <<EOF
{"hookSpecificOutput":{"hookEventName":"Setup","additionalContext":"$message"}}
EOF
}

setup_locale() {
  if type get_harness_locale >/dev/null 2>&1; then
    get_harness_locale
  else
    case "${CLAUDE_CODE_HARNESS_LANG:-en}" in
      ja) printf '%s\n' "ja" ;;
      *) printf '%s\n' "en" ;;
    esac
  fi
}

template_for_locale() {
  local relative_path="$1"
  local locale="$2"
  local localized_path="$TEMPLATE_DIR/locales/$locale/$relative_path"

  if [ "$locale" = "ja" ] && [ -f "$localized_path" ]; then
    printf '%s\n' "$localized_path"
    return 0
  fi

  printf '%s\n' "$TEMPLATE_DIR/$relative_path"
}

escape_sed_repl() {
  printf '%s' "$1" | sed -e 's/\\/\\\\/g' -e 's/[\/&|]/\\&/g'
}

render_template() {
  local template_path="$1"
  local dest_path="$2"
  local locale="$3"
  local project_name setup_date project_esc date_esc locale_esc

  project_name="$(basename "$(pwd)")"
  setup_date="$(date +"%Y-%m-%d")"
  project_esc="$(escape_sed_repl "$project_name")"
  date_esc="$(escape_sed_repl "$setup_date")"
  locale_esc="$(escape_sed_repl "$locale")"

  sed \
    -e "s|{{PROJECT_NAME}}|$project_esc|g" \
    -e "s|{{DATE}}|$date_esc|g" \
    -e "s|{{LANGUAGE}}|$locale_esc|g" \
    "$template_path" > "$dest_path"
}

msg() {
  local key="$1"
  local locale="$2"

  case "$locale:$key" in
    ja:cache_synced) printf '%s\n' "プラグインキャッシュ同期完了" ;;
    ja:state_ready) printf '%s\n' "状態ディレクトリ初期化完了" ;;
    ja:config_created) printf '%s\n' "設定ファイル生成完了" ;;
    ja:agents_created) printf '%s\n' "AGENTS.md 生成完了" ;;
    ja:claude_created) printf '%s\n' "CLAUDE.md 生成完了" ;;
    ja:plans_created) printf '%s\n' "Plans.md 生成完了" ;;
    ja:already_initialized) printf '%s\n' "[Setup:init] ハーネスは既に初期化済みです" ;;
    ja:cache_maintenance) printf '%s\n' "キャッシュ同期完了" ;;
    ja:old_sessions_removed) printf '%s\n' "古いセッションアーカイブ削除" ;;
    ja:maintenance_done) printf '%s\n' "[Setup:maintenance] メンテナンス完了（変更なし）" ;;
    ja:template_updates) printf '%s\n' "テンプレート更新あり" ;;
    ja:config_warning) printf '%s\n' "警告: 設定ファイルの構文エラー" ;;
    ja:unknown_mode) printf '%s\n' "[Setup] 不明なモード" ;;
    *:cache_synced) printf '%s\n' "Plugin cache synced" ;;
    *:state_ready) printf '%s\n' "State directory initialized" ;;
    *:config_created) printf '%s\n' "Config file generated" ;;
    *:agents_created) printf '%s\n' "AGENTS.md generated" ;;
    *:claude_created) printf '%s\n' "CLAUDE.md generated" ;;
    *:plans_created) printf '%s\n' "Plans.md generated" ;;
    *:already_initialized) printf '%s\n' "[Setup:init] Harness is already initialized" ;;
    *:cache_maintenance) printf '%s\n' "Cache synced" ;;
    *:old_sessions_removed) printf '%s\n' "Old session archives removed" ;;
    *:maintenance_done) printf '%s\n' "[Setup:maintenance] Maintenance complete (no changes)" ;;
    *:template_updates) printf '%s\n' "Template updates available" ;;
    *:config_warning) printf '%s\n' "Warning: config file has a syntax error" ;;
    *:unknown_mode) printf '%s\n' "[Setup] Unknown mode" ;;
  esac
}

join_messages() {
  local joined=""
  local entry

  for entry in "$@"; do
    if [ -z "$joined" ]; then
      joined="$entry"
    else
      joined="$joined, $entry"
    fi
  done

  printf '%s\n' "$joined"
}

# ===== Init モード: 初回セットアップ =====
run_init() {
  local messages=()
  local locale
  locale="$(setup_locale)"

  # 1. プラグインキャッシュの同期
  if [ -f "$SCRIPT_DIR/sync-plugin-cache.sh" ]; then
    bash "$SCRIPT_DIR/sync-plugin-cache.sh" >/dev/null 2>&1 || true
    messages+=("$(msg cache_synced "$locale")")
  fi

  # 2. 状態ディレクトリの初期化
  STATE_DIR=".claude/state"
  mkdir -p "$STATE_DIR"

  # 3. デフォルト設定ファイルの生成（存在しない場合）
  CONFIG_FILE=".claude-code-harness.config.yaml"
  if [ ! -f "$CONFIG_FILE" ]; then
    local config_template
    config_template="$(template_for_locale ".claude-code-harness.config.yaml.template" "$locale")"
    if [ -f "$config_template" ]; then
      render_template "$config_template" "$CONFIG_FILE" "$locale"
      messages+=("$(msg config_created "$locale")")
    fi
  fi

  # 4. AGENTS.md / CLAUDE.md の生成（存在しない場合）
  if [ ! -f "AGENTS.md" ]; then
    local agents_template
    agents_template="$(template_for_locale "AGENTS.md.template" "$locale")"
    if [ -f "$agents_template" ]; then
      render_template "$agents_template" "AGENTS.md" "$locale"
      messages+=("$(msg agents_created "$locale")")
    fi
  fi

  if [ ! -f "CLAUDE.md" ]; then
    local claude_template
    claude_template="$(template_for_locale "CLAUDE.md.template" "$locale")"
    if [ -f "$claude_template" ]; then
      render_template "$claude_template" "CLAUDE.md" "$locale"
      messages+=("$(msg claude_created "$locale")")
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

    local plans_template
    plans_template="$(template_for_locale "Plans.md.template" "$locale")"
    if [ -f "$plans_template" ]; then
      render_template "$plans_template" "$PLANS_PATH" "$locale"
      messages+=("$(msg plans_created "$locale")")
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
    output_json "$(msg already_initialized "$locale")"
  else
    local msg_str
    msg_str="$(join_messages "${messages[@]}")"
    output_json "[Setup:init] $msg_str"
  fi
}

# ===== Maintenance モード: メンテナンス処理 =====
run_maintenance() {
  local messages=()
  local locale
  locale="$(setup_locale)"

  # 1. プラグインキャッシュの同期
  if [ -f "$SCRIPT_DIR/sync-plugin-cache.sh" ]; then
    bash "$SCRIPT_DIR/sync-plugin-cache.sh" >/dev/null 2>&1 || true
    messages+=("$(msg cache_maintenance "$locale")")
  fi

  # 2. 古いセッションファイルのクリーンアップ
  STATE_DIR=".claude/state"
  ARCHIVE_DIR="$STATE_DIR/sessions"

  if [ -d "$ARCHIVE_DIR" ]; then
    # 7日以上前のセッションアーカイブを削除
    find "$ARCHIVE_DIR" -name "session-*.json" -mtime +7 -delete 2>/dev/null || true
    messages+=("$(msg old_sessions_removed "$locale")")
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
        messages+=("$(msg template_updates "$locale"): ${UPDATES_COUNT}")
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
        messages+=("$(msg config_warning "$locale")")
      fi
    fi
  fi

  # 結果出力
  if [ ${#messages[@]} -eq 0 ]; then
    output_json "$(msg maintenance_done "$locale")"
  else
    local msg_str
    msg_str="$(join_messages "${messages[@]}")"
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
    locale="$(setup_locale)"
    output_json "$(msg unknown_mode "$locale"): $MODE"
    exit 1
    ;;
esac
