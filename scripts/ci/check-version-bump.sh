#!/bin/bash
# check-version-bump.sh
# release metadata policy check
#
# 目的:
# - 通常 PR では VERSION bump を要求しない
# - VERSION を更新した場合だけ、plugin.json / CHANGELOG release entry が揃っていることを確認する
#
# 使用方法:
# - PR の場合: GITHUB_BASE_REF 環境変数を設定
# - Push の場合: 前のコミットと比較

set -euo pipefail

echo "🏷️ リリースメタデータチェック"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [ -n "${GITHUB_BASE_REF:-}" ]; then
  BASE="origin/$GITHUB_BASE_REF"
  DIFF_TARGET="HEAD"
  echo "📌 PR モード: $BASE と比較"
elif [ -n "${GITHUB_EVENT_NAME:-}" ] && [ "$GITHUB_EVENT_NAME" = "push" ]; then
  BASE="HEAD~1"
  DIFF_TARGET="HEAD"
  echo "📌 Push モード: 前のコミットと比較"
else
  BASE="origin/main"
  DIFF_TARGET=""
  echo "📌 ローカルモード: $BASE と比較"
fi

if ! git rev-parse "$BASE" >/dev/null 2>&1; then
  echo "⚠️ 比較対象 ($BASE) が見つかりません。スキップします。"
  exit 0
fi

semver_gt() {
  local left="$1"
  local right="$2"
  local l_major=0 l_minor=0 l_patch=0
  local r_major=0 r_minor=0 r_patch=0

  IFS='.' read -r l_major l_minor l_patch <<< "$left"
  IFS='.' read -r r_major r_minor r_patch <<< "$right"

  for value in "$l_major" "$l_minor" "$l_patch" "$r_major" "$r_minor" "$r_patch"; do
    if ! [[ "$value" =~ ^[0-9]+$ ]]; then
      return 1
    fi
  done

  if ((10#$l_major > 10#$r_major)); then
    return 0
  fi
  if ((10#$l_major < 10#$r_major)); then
    return 1
  fi
  if ((10#$l_minor > 10#$r_minor)); then
    return 0
  fi
  if ((10#$l_minor < 10#$r_minor)); then
    return 1
  fi
  if ((10#$l_patch > 10#$r_patch)); then
    return 0
  fi

  return 1
}

echo ""
echo "🔍 変更ファイルをチェック..."

RELEASE_METADATA_FILES="VERSION .claude-plugin/plugin.json CHANGELOG.md"
if [ -n "$DIFF_TARGET" ]; then
  CHANGED_RELEASE_METADATA=$(git diff --name-only "$BASE" "$DIFF_TARGET" -- $RELEASE_METADATA_FILES 2>/dev/null | grep -v "^$" || true)
else
  CHANGED_RELEASE_METADATA=$(git diff --name-only "$BASE" -- $RELEASE_METADATA_FILES 2>/dev/null | grep -v "^$" || true)
fi

if [ -z "$CHANGED_RELEASE_METADATA" ]; then
  echo "  ✅ release metadata 変更なし（通常 PR / 通常 push として許容）"
  exit 0
fi

echo "  📝 変更された release metadata:"
echo "$CHANGED_RELEASE_METADATA" | head -10 | while read -r file; do
  echo "     - $file"
done
CHANGED_COUNT=$(echo "$CHANGED_RELEASE_METADATA" | wc -l | tr -d ' ')
if [ "$CHANGED_COUNT" -gt 10 ]; then
  echo "     ... 他 $((CHANGED_COUNT - 10)) ファイル"
fi

echo ""
echo "🔍 バージョン変更をチェック..."

CURRENT_VERSION=$(cat VERSION 2>/dev/null | tr -d '[:space:]')
BASE_VERSION=$(git show "$BASE:VERSION" 2>/dev/null | tr -d '[:space:]' || echo "")

echo "  ベース: v${BASE_VERSION:-なし}"
echo "  現在:   v${CURRENT_VERSION}"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [ -z "$BASE_VERSION" ]; then
  echo "✅ 新規プロジェクト（release metadata チェックスキップ）"
  exit 0
fi

if [ "$CURRENT_VERSION" = "$BASE_VERSION" ]; then
  echo "✅ VERSION は未変更です。通常 PR / 通常 push では version bump は不要です。"

  if bash ./scripts/sync-version.sh check >/dev/null 2>&1; then
    echo "✅ plugin.json も VERSION と一致しています。"
    exit 0
  fi

  echo "❌ VERSION は未変更ですが plugin.json と不一致です。"
  bash ./scripts/sync-version.sh check
  exit 1
fi

if ! semver_gt "$CURRENT_VERSION" "$BASE_VERSION"; then
  echo "❌ VERSION は更新されていますが、SemVer として増分になっていません。"
  echo "   ベース: $BASE_VERSION"
  echo "   現在:   $CURRENT_VERSION"
  exit 1
fi

echo "✅ release 用の VERSION 更新を検出: $BASE_VERSION → $CURRENT_VERSION"

ERRORS=0

if bash ./scripts/sync-version.sh check >/dev/null 2>&1; then
  echo "✅ plugin.json の version も一致しています。"
else
  echo "❌ plugin.json の version が VERSION と一致していません。"
  bash ./scripts/sync-version.sh check || true
  ERRORS=$((ERRORS + 1))
fi

if grep -Eq "^## \[$CURRENT_VERSION\] - [0-9]{4}-[0-9]{2}-[0-9]{2}$" CHANGELOG.md; then
  echo "✅ CHANGELOG.md に v$CURRENT_VERSION の release entry があります。"
else
  echo "❌ CHANGELOG.md に v$CURRENT_VERSION の release entry がありません。"
  ERRORS=$((ERRORS + 1))
fi

if grep -Eq "^\[$CURRENT_VERSION\]: https://github.com/Chachamaru127/claude-code-harness/compare/v" CHANGELOG.md; then
  echo "✅ CHANGELOG compare link があります。"
else
  echo "❌ CHANGELOG compare link がありません。"
  ERRORS=$((ERRORS + 1))
fi

if [ "$ERRORS" -gt 0 ]; then
  echo ""
  echo "💡 修正方針:"
  echo "  - 通常 PR では VERSION を変更しない"
  echo "  - release を切るときだけ VERSION / plugin.json / CHANGELOG release entry を一緒に更新する"
  exit 1
fi

echo "✅ release metadata チェックOK"
exit 0
