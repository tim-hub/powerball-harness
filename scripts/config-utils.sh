#!/bin/bash
# config-utils.sh
# ハーネス設定ファイルからの値取得ユーティリティ
#
# Usage: source "${SCRIPT_DIR}/config-utils.sh"
#        plans_path=$(get_plans_file_path)

# 設定ファイルのデフォルトパス
CONFIG_FILE="${CONFIG_FILE:-.claude-code-harness.config.yaml}"

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

  # yq が利用可能な場合
  if command -v yq >/dev/null 2>&1; then
    value=$(yq -r '.plansDirectory // empty' "$CONFIG_FILE" 2>/dev/null)
  fi

  # yq で取得できなかった場合、Python を試行
  if [ -z "$value" ] && command -v python3 >/dev/null 2>&1; then
    # Python で YAML パース（pyyaml がない場合は空を返す）
    value=$(python3 - "$CONFIG_FILE" <<'PY' 2>/dev/null
import sys
try:
    import yaml
    with open(sys.argv[1]) as f:
        data = yaml.safe_load(f) or {}
    print(data.get('plansDirectory', ''))
except ImportError:
    # pyyaml not installed - return empty to trigger grep fallback
    pass
except:
    pass
PY
)
  fi

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
