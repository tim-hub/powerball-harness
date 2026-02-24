#!/bin/bash
# check-simple-mode.sh
# CLAUDE_CODE_SIMPLE モード検出ユーティリティ
#
# Usage:
#   source scripts/check-simple-mode.sh
#   if is_simple_mode; then echo "SIMPLE mode"; fi
#
# Environment:
#   CLAUDE_CODE_SIMPLE=1  → skills/memory/agents が無効化される (CC v2.1.50+)
#
# Returns:
#   0 (true)  if SIMPLE mode is active
#   1 (false) if normal mode

# SIMPLE モード判定
# CLAUDE_CODE_SIMPLE=1 の場合、Claude Code は skills/memory/agents をストリップする
is_simple_mode() {
  [ "${CLAUDE_CODE_SIMPLE:-0}" = "1" ]
}

# SIMPLE モードの警告メッセージを生成（日本語/英語）
# Args: $1 = lang (ja|en)
# Output: 警告メッセージ文字列
simple_mode_warning() {
  local lang="${1:-ja}"

  if [ "$lang" = "en" ]; then
    cat <<'MSG'
WARNING: CLAUDE_CODE_SIMPLE mode detected (CC v2.1.50+)
- Skills DISABLED: /work, /breezing, /plan-with-agent, /harness-review unavailable
- Agents DISABLED: task-worker, code-reviewer, parallel execution unavailable
- Memory DISABLED: project memory and cross-session learning unavailable
- Hooks ACTIVE: safety guards and session management continue to operate
→ See docs/SIMPLE_MODE_COMPATIBILITY.md for details
MSG
  else
    cat <<'MSG'
警告: CLAUDE_CODE_SIMPLE モードが検出されました (CC v2.1.50+)
- スキル無効: /work, /breezing, /plan-with-agent, /harness-review は使用不可
- エージェント無効: task-worker, code-reviewer, 並列実行は使用不可
- メモリ無効: プロジェクトメモリ・セッション間学習は使用不可
- フック有効: 安全ガード・セッション管理は引き続き動作
→ 詳細は docs/SIMPLE_MODE_COMPATIBILITY.md を参照
MSG
  fi
}
