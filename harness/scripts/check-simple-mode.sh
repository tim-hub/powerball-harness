#!/bin/bash
# check-simple-mode.sh
# Utility for detecting CLAUDE_CODE_SIMPLE mode
#
# Usage:
#   source scripts/check-simple-mode.sh
#   if is_simple_mode; then echo "SIMPLE mode"; fi
#
# Environment:
#   CLAUDE_CODE_SIMPLE=1  → skills/memory/agents are disabled (CC v2.1.50+)
#
# Returns:
#   0 (true)  if SIMPLE mode is active
#   1 (false) if normal mode

# Determine whether SIMPLE mode is active.
# When CLAUDE_CODE_SIMPLE=1, Claude Code strips skills/memory/agents.
is_simple_mode() {
  [ "${CLAUDE_CODE_SIMPLE:-0}" = "1" ]
}

# Generate a warning message for SIMPLE mode (English only)
# Args: $1 = lang (en, or omit for English)
# Output: warning message string
simple_mode_warning() {
  local lang="${1:-en}"

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
WARNING: CLAUDE_CODE_SIMPLE mode detected (CC v2.1.50+)
- Skills DISABLED: /work, /breezing, /plan-with-agent, /harness-review unavailable
- Agents DISABLED: task-worker, code-reviewer, parallel execution unavailable
- Memory DISABLED: project memory and cross-session learning unavailable
- Hooks ACTIVE: safety guards and session management continue to operate
→ See docs/SIMPLE_MODE_COMPATIBILITY.md for details
MSG
  fi
}
