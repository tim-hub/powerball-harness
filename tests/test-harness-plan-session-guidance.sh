#!/bin/bash
# test-harness-plan-session-guidance.sh
# harness-plan create 完了時のセッション起動案内が維持されているかを検証する。

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PLUGIN_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

SKILL_FILE="$PLUGIN_ROOT/skills/harness-plan/SKILL.md"
CREATE_REF="$PLUGIN_ROOT/skills/harness-plan/references/create.md"
LONGRUN_SCRIPT="$PLUGIN_ROOT/scripts/claude-longrun.sh"

pass() {
  echo "PASS: $1"
}

fail() {
  echo "FAIL: $1" >&2
  exit 1
}

require_contains() {
  local file="$1"
  local pattern="$2"
  local label="$3"

  if grep -Fq "$pattern" "$file"; then
    pass "$label"
  else
    fail "$label ($file に '$pattern' がありません)"
  fi
}

echo "=== harness-plan session guidance test ==="

[ -f "$SKILL_FILE" ] || fail "harness-plan の SKILL.md が見つかりません"
[ -f "$CREATE_REF" ] || fail "harness-plan create reference が見つかりません"
[ -f "$LONGRUN_SCRIPT" ] || fail "claude-longrun.sh が見つかりません"

require_contains "$SKILL_FILE" "### create 完了時のセッション起動案内（必須）" "SKILL.md に必須案内セクションがある"
require_contains "$SKILL_FILE" "新しいセッションの起動コマンド:" "SKILL.md に起動コマンド文言がある"
require_contains "$SKILL_FILE" "起動後の最初の入力:" "SKILL.md に最初の入力文言がある"
require_contains "$SKILL_FILE" "bash scripts/claude-longrun.sh" "SKILL.md に長時間用起動コマンドがある"

require_contains "$CREATE_REF" "## Step 7: セッション起動コマンドと最初の入力を必ず案内する" "create reference に案内ステップがある"
require_contains "$CREATE_REF" "次の一歩:" "create reference に出力例がある"
require_contains "$CREATE_REF" "/breezing all" "create reference に breezing 導線がある"
require_contains "$CREATE_REF" "/harness-loop all" "create reference に harness-loop 導線がある"

require_contains "$LONGRUN_SCRIPT" "export ENABLE_PROMPT_CACHING_1H=1" "claude-longrun.sh が 1 時間キャッシュを有効化する"
require_contains "$LONGRUN_SCRIPT" 'exec claude "$@"' "claude-longrun.sh が claude を起動する"

echo "All session guidance checks passed."
