#!/bin/bash
# config-utils.sh
# ハーネス設定ファイルからの値取得ユーティリティ
#
# Usage: source "${SCRIPT_DIR}/config-utils.sh"
#        plans_path=$(get_plans_file_path)

# 設定ファイルのデフォルトパス
CONFIG_FILE="${CONFIG_FILE:-.claude-code-harness.config.yaml}"
PROJECT_ROOT="${PROJECT_ROOT:-$(pwd)}"

yaml_get_value() {
  local query="$1"
  local file="${2:-$CONFIG_FILE}"
  local value=""

  if [ ! -f "$file" ]; then
    return 0
  fi

  if command -v yq >/dev/null 2>&1; then
    value="$(yq -r "${query} // empty" "$file" 2>/dev/null || true)"
  fi

  if [ -z "$value" ] && command -v python3 >/dev/null 2>&1; then
    value="$(python3 - "$file" "$query" <<'PY' 2>/dev/null
import sys

try:
    import yaml
except ImportError:
    raise SystemExit(0)

path = sys.argv[2].strip()
if path.startswith("."):
    path = path[1:]
if path.startswith("."):
    path = path[1:]
keys = [part for part in path.split(".") if part]

with open(sys.argv[1], "r", encoding="utf-8") as fh:
    data = yaml.safe_load(fh) or {}

value = data
for key in keys:
    if isinstance(value, dict) and key in value:
        value = value[key]
    else:
        value = None
        break

if value is None:
    raise SystemExit(0)
if isinstance(value, bool):
    print("true" if value else "false")
else:
    print(value)
PY
)"
  fi

  printf '%s\n' "$value"
}

# plansDirectory の検証（セキュリティ）
# 絶対パス、親ディレクトリ参照、symlink脱出を拒否
validate_plans_directory() {
  local value="$1"
  local default="."

  # 空の場合はデフォルト
  [ -z "$value" ] && echo "$default" && return 0

  # Security: 絶対パスを拒否
  case "$value" in
    /*) echo "$default" && return 0 ;;
  esac

  # Security: 親ディレクトリ参照 (..) を拒否
  case "$value" in
    *..*)  echo "$default" && return 0 ;;
  esac

  # Security: symlink脱出を検出（realpathが利用可能な場合）
  if command -v realpath >/dev/null 2>&1 && [ -e "$value" ]; then
    local project_root
    local resolved_path
    project_root=$(realpath "." 2>/dev/null) || project_root=$(pwd)
    resolved_path=$(realpath "$value" 2>/dev/null)

    if [ -n "$resolved_path" ]; then
      # 解決されたパスがプロジェクトルート内にあるか確認
      case "$resolved_path" in
        "$project_root"/*) ;; # OK: プロジェクト内
        "$project_root") ;;   # OK: プロジェクトルート自体
        *) echo "$default" && return 0 ;; # NG: プロジェクト外
      esac
    fi
  fi

  echo "$value"
}

# plansDirectory 設定を取得（デフォルト: "."）
get_plans_directory() {
  local default="."

  if [ ! -f "$CONFIG_FILE" ]; then
    echo "$default"
    return 0
  fi

  local value=""
  value="$(yaml_get_value '.plansDirectory' "$CONFIG_FILE")"

  # yq/Python で取得できなかった場合、grep + sed でフォールバック
  if [ -z "$value" ]; then
    value=$(grep "^plansDirectory:" "$CONFIG_FILE" 2>/dev/null | sed 's/^plansDirectory:[[:space:]]*//' | tr -d '"' | tr -d "'" || echo "")
  fi

  # 検証してから返す
  validate_plans_directory "$value"
}

# Plans.md のフルパスを取得
get_plans_file_path() {
  local plans_dir
  plans_dir=$(get_plans_directory)

  # ディレクトリ内で Plans.md を検索（大文字小文字を区別しない）
  for f in Plans.md plans.md PLANS.md PLANS.MD; do
    local full_path="${plans_dir}/${f}"
    # "." の場合は "./" を省略
    [ "$plans_dir" = "." ] && full_path="$f"

    if [ -f "$full_path" ]; then
      echo "$full_path"
      return 0
    fi
  done

  # 見つからない場合はデフォルトパスを返す
  local default_path="${plans_dir}/Plans.md"
  [ "$plans_dir" = "." ] && default_path="Plans.md"
  echo "$default_path"
}

# Plans.md が存在するかチェック
plans_file_exists() {
  local plans_path
  plans_path=$(get_plans_file_path)
  [ -f "$plans_path" ]
}

normalize_boolean() {
  local value="$1"
  local default="$2"
  case "${value}" in
    true|TRUE|True|yes|YES|Yes|on|ON|On|1) echo "true" ;;
    false|FALSE|False|no|NO|No|off|OFF|Off|0) echo "false" ;;
    *) echo "$default" ;;
  esac
}

normalize_integer() {
  local value="$1"
  local default="$2"
  case "${value}" in
    ''|*[!0-9]*) echo "$default" ;;
    *) echo "$value" ;;
  esac
}

advisor_config_value() {
  local key="$1"
  local default="$2"
  local value=""
  value="$(yaml_get_value ".advisor.${key}" "$CONFIG_FILE")"
  if [ -z "$value" ]; then
    printf '%s\n' "$default"
    return 0
  fi
  printf '%s\n' "$value"
}

get_advisor_enabled() {
  normalize_boolean "$(advisor_config_value "enabled" "true")" "true"
}

get_advisor_mode() {
  advisor_config_value "mode" "on-demand"
}

get_advisor_max_consults_per_task() {
  normalize_integer "$(advisor_config_value "max_consults_per_task" "3")" "3"
}

get_advisor_retry_threshold() {
  normalize_integer "$(advisor_config_value "retry_threshold" "2")" "2"
}

get_advisor_consult_before_user_escalation() {
  normalize_boolean "$(advisor_config_value "consult_before_user_escalation" "true")" "true"
}

get_advisor_claude_model() {
  advisor_config_value "claude_model" "opus"
}

get_advisor_codex_model() {
  advisor_config_value "codex_model" "gpt-5.4"
}

get_advisor_state_dir() {
  printf '%s\n' "${PROJECT_ROOT}/.claude/state/advisor"
}

get_advisor_history_file() {
  printf '%s/history.jsonl\n' "$(get_advisor_state_dir)"
}

get_advisor_last_request_file() {
  printf '%s/last-request.json\n' "$(get_advisor_state_dir)"
}

get_advisor_last_response_file() {
  printf '%s/last-response.json\n' "$(get_advisor_state_dir)"
}

ensure_advisor_state_files() {
  local state_dir history_file request_file response_file
  state_dir="$(get_advisor_state_dir)"
  history_file="$(get_advisor_history_file)"
  request_file="$(get_advisor_last_request_file)"
  response_file="$(get_advisor_last_response_file)"

  mkdir -p "${state_dir}"
  [ -f "${history_file}" ] || : > "${history_file}"
  [ -f "${request_file}" ] || printf '{}\n' > "${request_file}"
  [ -f "${response_file}" ] || printf '{}\n' > "${response_file}"
}
