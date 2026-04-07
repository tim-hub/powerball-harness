#!/bin/bash
# fix-symlinks.sh
# Windows 環境で壊れた symlink / plain-text link projection を検出し、実体コピーで自動修復する
#
# 用途: session-init.sh から呼び出し
# 動作:
#   - codex/.codex/skills/ および opencode/skills/ 内の harness-* skill mirror を検証
#   - skills/ (SSOT) から実体コピーで修復する
#   - 修復件数を stdout に出力（JSON 形式）
#
# 出力:
#   {"fixed": N, "checked": M, "details": ["codex/harness-work", ...]}

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PLUGIN_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

SKILLS_DIR="$PLUGIN_ROOT/skills"

# harness skill 一覧
HARNESS_SKILLS=("harness-plan" "harness-work" "harness-review" "harness-setup" "harness-release" "harness-sync")

# ミラー先（skills/ は SSOT なのでチェック対象外）
MIRROR_ROOTS=(
  "codex/.codex/skills"
  "opencode/skills"
)

FIXED=0
CHECKED=0
FIXED_NAMES=()

for mirror_root in "${MIRROR_ROOTS[@]}"; do
  mirror_dir="$PLUGIN_ROOT/$mirror_root"
  [ -d "$mirror_dir" ] || continue

  for skill in "${HARNESS_SKILLS[@]}"; do
    CHECKED=$((CHECKED + 1))
    mirror_path="$mirror_dir/$skill"
    source_path="$SKILLS_DIR/$skill"

    # ソースが存在しない場合はスキップ
    [ -d "$source_path" ] || continue

    # 正常: ディレクトリとして存在 → スキップ
    if [ -d "$mirror_path" ] && [ ! -L "$mirror_path" ]; then
      continue
    fi

    # 壊れた plain-text link: 通常ファイルとして存在（Windows git clone で発生）
    if [ -f "$mirror_path" ]; then
      rm -f "$mirror_path"
      cp -r "$source_path" "$mirror_path"
      FIXED=$((FIXED + 1))
      FIXED_NAMES+=("${mirror_root##*/}/$skill")
      continue
    fi

    # symlink の場合も実体コピーに置換
    if [ -L "$mirror_path" ]; then
      rm -f "$mirror_path"
      cp -r "$source_path" "$mirror_path"
      FIXED=$((FIXED + 1))
      FIXED_NAMES+=("${mirror_root##*/}/$skill")
      continue
    fi

    # 存在しない場合もコピー
    if [ ! -e "$mirror_path" ]; then
      cp -r "$source_path" "$mirror_path"
      FIXED=$((FIXED + 1))
      FIXED_NAMES+=("${mirror_root##*/}/$skill")
    fi
  done
done

# JSON 出力
NAMES_JSON="[]"
if [ ${#FIXED_NAMES[@]} -gt 0 ]; then
  NAMES_JSON="["
  for i in "${!FIXED_NAMES[@]}"; do
    [ "$i" -gt 0 ] && NAMES_JSON+=","
    NAMES_JSON+="\"${FIXED_NAMES[$i]}\""
  done
  NAMES_JSON+="]"
fi

echo "{\"fixed\":${FIXED},\"checked\":${CHECKED},\"details\":${NAMES_JSON}}"
