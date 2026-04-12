#!/bin/bash
# validate-plugin-v3.sh
# Harness v3 プラグイン構造バリデーター
#
# Usage: ./tests/validate-plugin-v3.sh
# Exit codes:
#   0 - All checks passed
#   1 - Failures found

set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PLUGIN_ROOT="$(dirname "$SCRIPT_DIR")"

echo "=========================================="
echo "Claude Harness v3 — プラグイン検証テスト"
echo "=========================================="
echo ""

# カラー出力
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

PASS_COUNT=0
FAIL_COUNT=0
WARN_COUNT=0

pass_test() { echo -e "${GREEN}✓${NC} $1"; PASS_COUNT=$((PASS_COUNT + 1)); }
fail_test() { echo -e "${RED}✗${NC} $1"; FAIL_COUNT=$((FAIL_COUNT + 1)); }
warn_test() { echo -e "${YELLOW}⚠${NC} $1"; WARN_COUNT=$((WARN_COUNT + 1)); }

# ============================================================
# [1] v3 コア構造チェック
# ============================================================
echo "📁 [1/6] v3 コア構造チェック..."

V3_REQUIRED_FILES=(
  "core/package.json"
  "core/tsconfig.json"
  "core/src/index.ts"
  "core/src/types.ts"
  "core/src/guardrails/rules.ts"
  "core/src/guardrails/pre-tool.ts"
  "core/src/guardrails/post-tool.ts"
  "core/src/guardrails/permission.ts"
  "core/src/guardrails/tampering.ts"
  "core/src/engine/lifecycle.ts"
)

for f in "${V3_REQUIRED_FILES[@]}"; do
  if [ -f "$PLUGIN_ROOT/$f" ]; then
    pass_test "$f"
  else
    fail_test "$f (存在しない)"
  fi
done

# ============================================================
# [2] 5動詞スキルチェック
# ============================================================
echo ""
echo "🎯 [2/6] 5動詞スキルチェック..."

V3_SKILLS=(harness-plan harness-work harness-review harness-release harness-setup)
AUX_V3_SKILLS=(harness-sync)

for skill in "${V3_SKILLS[@]}"; do
  skill_dir="$PLUGIN_ROOT/skills/$skill"
  skill_md="$skill_dir/SKILL.md"

  if [ ! -d "$skill_dir" ]; then
    fail_test "skills/$skill/ (ディレクトリなし)"
    continue
  fi

  if [ ! -f "$skill_md" ]; then
    fail_test "skills/$skill/SKILL.md (なし)"
    continue
  fi

  # frontmatter の name: チェック
  if grep -q "^name: $skill$" "$skill_md"; then
    pass_test "skills/$skill/SKILL.md (name: $skill)"
  else
    fail_test "skills/$skill/SKILL.md (name: フィールドが '$skill' でない)"
  fi
done

echo ""
echo "🧭 [2.5/6] 補助 workflow surface チェック..."

for skill in "${AUX_V3_SKILLS[@]}"; do
  skill_dir="$PLUGIN_ROOT/skills/$skill"
  skill_md="$skill_dir/SKILL.md"

  if [ ! -d "$skill_dir" ]; then
    fail_test "skills/$skill/ (ディレクトリなし)"
    continue
  fi

  if [ ! -f "$skill_md" ]; then
    fail_test "skills/$skill/SKILL.md (なし)"
    continue
  fi

  if grep -q "^name: $skill$" "$skill_md"; then
    pass_test "skills/$skill/SKILL.md (name: $skill)"
  else
    fail_test "skills/$skill/SKILL.md (name: フィールドが '$skill' でない)"
  fi
done

# ============================================================
# [3] Public mirror bundle チェック
# ============================================================
echo ""
echo "📦 [3/6] Public mirror bundle チェック..."

MIRRORS=(
  "codex/.codex/skills"
  "opencode/skills"
)

for mirror_dir in "${MIRRORS[@]}"; do
  if [ ! -d "$PLUGIN_ROOT/$mirror_dir" ]; then
    warn_test "$mirror_dir (存在しない、スキップ)"
    continue
  fi

  for skill in "${V3_SKILLS[@]}"; do
    source_dir="$PLUGIN_ROOT/skills/$skill"
    mirror_path="$PLUGIN_ROOT/$mirror_dir/$skill"

    if [ ! -d "$mirror_path" ]; then
      fail_test "$mirror_dir/$skill (ディレクトリなし)"
      continue
    fi

    if [ -L "$mirror_path" ]; then
      fail_test "$mirror_dir/$skill (symlink のまま)"
      continue
    fi

    if diff -qr "$source_dir" "$mirror_path" >/dev/null 2>&1; then
      pass_test "$mirror_dir/$skill (skills/$skill と同期)"
    else
      fail_test "$mirror_dir/$skill (skills/$skill と差分あり)"
    fi
  done

  for skill in "${AUX_V3_SKILLS[@]}"; do
    source_dir="$PLUGIN_ROOT/skills/$skill"
    mirror_path="$PLUGIN_ROOT/$mirror_dir/$skill"

    if [ ! -d "$mirror_path" ]; then
      fail_test "$mirror_dir/$skill (ディレクトリなし)"
      continue
    fi

    if [ -L "$mirror_path" ]; then
      fail_test "$mirror_dir/$skill (symlink のまま)"
      continue
    fi

    if diff -qr "$source_dir" "$mirror_path" >/dev/null 2>&1; then
      pass_test "$mirror_dir/$skill (skills/$skill と同期)"
    else
      fail_test "$mirror_dir/$skill (skills/$skill と差分あり)"
    fi
  done
done

# ============================================================
# [4] 3エージェントチェック
# ============================================================
echo ""
echo "🤖 [4/6] 3エージェントチェック..."

V3_AGENTS=(worker reviewer scaffolder)

for agent in "${V3_AGENTS[@]}"; do
  agent_file="$PLUGIN_ROOT/agents/$agent.md"
  if [ -f "$agent_file" ]; then
    # name: フィールド確認
    if grep -q "^name: $agent$" "$agent_file"; then
      pass_test "agents/$agent.md (name: $agent)"
    else
      fail_test "agents/$agent.md (name: フィールドが '$agent' でない)"
    fi
  else
    fail_test "agents/$agent.md (存在しない)"
  fi
done

# team-composition.md
if [ -f "$PLUGIN_ROOT/agents/team-composition.md" ]; then
  pass_test "agents/team-composition.md"
else
  warn_test "agents/team-composition.md (なし)"
fi

# ============================================================
# [5] TypeScript 型チェック
# ============================================================
echo ""
echo "🔷 [5/6] TypeScript 型チェック..."

CORE_DIR="$PLUGIN_ROOT/core"

if [ ! -d "$CORE_DIR/node_modules" ]; then
  warn_test "core/node_modules なし — npm ci が必要 (スキップ)"
else
  if cd "$CORE_DIR" && npm run typecheck --silent 2>/dev/null; then
    pass_test "core/ TypeScript 型チェック通過"
  else
    fail_test "core/ TypeScript 型チェック失敗"
  fi
  cd "$PLUGIN_ROOT"
fi

# ============================================================
# [6] hooks シム チェック
# ============================================================
echo ""
echo "🪝 [6/6] hooks シムチェック..."

HOOK_FILES=(
  "hooks/pre-tool.sh"
  "hooks/post-tool.sh"
  "hooks/session.sh"
  "hooks/hooks.json"
)

for f in "${HOOK_FILES[@]}"; do
  if [ -f "$PLUGIN_ROOT/$f" ]; then
    pass_test "$f"
  else
    fail_test "$f (存在しない)"
  fi
done

for f in \
  "scripts/lib/harness-mem-bridge.sh" \
  "scripts/hook-handlers/memory-bridge.sh" \
  "scripts/hook-handlers/memory-session-start.sh" \
  "scripts/hook-handlers/memory-user-prompt.sh" \
  "scripts/hook-handlers/memory-post-tool-use.sh" \
  "scripts/hook-handlers/memory-stop.sh" \
  "scripts/hook-handlers/memory-codex-notify.sh"
do
  if [ -f "$PLUGIN_ROOT/$f" ]; then
    pass_test "$f"
  else
    fail_test "$f (存在しない)"
  fi
done

# ============================================================
# [7] Hardening parity チェック
# ============================================================
echo ""
echo "🛡️ [7/7] Hardening parity チェック..."

if [ -f "$PLUGIN_ROOT/docs/hardening-parity.md" ]; then
  pass_test "docs/hardening-parity.md"
else
  fail_test "docs/hardening-parity.md (存在しない)"
fi

if [ -f "$PLUGIN_ROOT/scripts/lib/codex-hardening-contract.txt" ] && grep -q 'HARNESS_HARDENING_CONTRACT_V1' "$PLUGIN_ROOT/scripts/lib/codex-hardening-contract.txt"; then
  pass_test "scripts/lib/codex-hardening-contract.txt"
else
  fail_test "scripts/lib/codex-hardening-contract.txt (存在しない)"
fi

if grep -q 'docs/hardening-parity.md' "$PLUGIN_ROOT/README.md"; then
  pass_test "README.md → hardening parity リンク"
else
  fail_test "README.md に hardening parity リンクがない"
fi

for rule_id in \
  "R10:no-git-bypass-flags" \
  "R11:no-reset-hard-protected-branch" \
  "R12:warn-direct-push-protected-branch" \
  "R13:warn-protected-review-paths"
do
  if grep -q "$rule_id" "$PLUGIN_ROOT/core/src/guardrails/rules.ts"; then
    pass_test "core/src/guardrails/rules.ts ($rule_id)"
  else
    fail_test "core/src/guardrails/rules.ts ($rule_id がない)"
  fi
done

if grep -q 'codex-hardening-contract.txt' "$PLUGIN_ROOT/scripts/codex/codex-exec-wrapper.sh"; then
  pass_test "codex-exec-wrapper.sh hardening contract template"
else
  fail_test "codex-exec-wrapper.sh が hardening contract template を参照していない"
fi

if grep -q 'codex-hardening-contract.txt' "$PLUGIN_ROOT/scripts/codex-worker-engine.sh"; then
  pass_test "codex-worker-engine.sh hardening contract template"
else
  fail_test "codex-worker-engine.sh が hardening contract template を参照していない"
fi

if grep -q 'gate_hardening()' "$PLUGIN_ROOT/scripts/codex-worker-quality-gate.sh"; then
  pass_test "codex-worker-quality-gate.sh hardening gate"
else
  fail_test "codex-worker-quality-gate.sh に hardening gate がない"
fi

# ============================================================
# サマリー
# ============================================================
echo ""
echo "=========================================="
echo "結果サマリー"
echo "=========================================="
echo -e "${GREEN}✓ 通過${NC}: $PASS_COUNT"
echo -e "${RED}✗ 失敗${NC}: $FAIL_COUNT"
echo -e "${YELLOW}⚠ 警告${NC}: $WARN_COUNT"
echo ""

if [ "$FAIL_COUNT" -gt 0 ]; then
  echo -e "${RED}❌ バリデーション失敗: $FAIL_COUNT 件のエラーがあります${NC}"
  exit 1
else
  echo -e "${GREEN}✅ バリデーション通過${NC}"
  exit 0
fi
