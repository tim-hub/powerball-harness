#!/bin/bash
# config-utils.sh
# ハーネス設定ファイルからの値取得ユーティリティ
#
# Usage: source "${SCRIPT_DIR}/config-utils.sh"
#        plans_path=$(get_plans_file_path)

# 設定ファイルのデフォルトパス
CONFIG_FILE="${CONFIG_FILE:-.claude-code-harness.config.yaml}"

# plansDirectory 設定を取得（デフォルト: "."）
get_plans_directory() {
  local default="."

  if [ ! -f "$CONFIG_FILE" ]; then
    echo "$default"
    return 0
  fi

  local value=""

  # jq が利用可能な場合は yq 相当の処理
  if command -v yq >/dev/null 2>&1; then
    value=$(yq -r '.plansDirectory // empty' "$CONFIG_FILE" 2>/dev/null)
  elif command -v python3 >/dev/null 2>&1; then
    # Python で YAML パース
    value=$(python3 - "$CONFIG_FILE" <<'PY' 2>/dev/null
import sys
try:
    import yaml
    with open(sys.argv[1]) as f:
        data = yaml.safe_load(f) or {}
    print(data.get('plansDirectory', ''))
except:
    print('')
PY
)
  else
    # フォールバック: grep + sed で簡易パース
    value=$(grep "^plansDirectory:" "$CONFIG_FILE" 2>/dev/null | sed 's/^plansDirectory:[[:space:]]*//' | tr -d '"' | tr -d "'" || echo "")
  fi

  # 空の場合はデフォルト値を返す
  if [ -z "$value" ]; then
    echo "$default"
  else
    echo "$value"
  fi
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
