#!/usr/bin/env bash
# suggest-compact.sh
# PostToolUse(Edit|Write) hook: increments a per-session tool-call counter and
# suggests /compact at strategic checkpoints.
#
# Adapted from affaan-m/everything-claude-code strategic-compact/suggest-compact.sh
# See: https://github.com/affaan-m/everything-claude-code/tree/main/skills/strategic-compact
#
# Harness adaptations vs upstream:
#   - Counter at .claude/state/compact-counter-<session_id>.json (not /tmp/)
#   - Emits {"systemMessage":"..."} on stdout at threshold (reaches model context,
#     not just stderr operator breadcrumbs)
#   - Env vars HARNESS_COMPACT_THRESHOLD / HARNESS_COMPACT_INTERVAL
#   - Suppression: skips when role=worker AND Plans.md has cc:WIP
#   - Project root via git rev-parse per .claude/rules/path-conventions.md
#   - Always exits 0 — never blocks tool execution
#
# Hook config (harness/hooks/hooks.json):
#   PostToolUse / matcher "Edit|Write" / command:
#   bash "${CLAUDE_PLUGIN_ROOT}/skills/harness-compact/scripts/suggest-compact.sh"

set -euo pipefail

THRESHOLD="${HARNESS_COMPACT_THRESHOLD:-50}"
INTERVAL="${HARNESS_COMPACT_INTERVAL:-25}"
SESSION_ID="${CLAUDE_SESSION_ID:-default}"

# project-root: user's git repository root
PROJECT_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
STATE_DIR="${PROJECT_ROOT}/.claude/state"
COUNTER_FILE="${STATE_DIR}/compact-counter-${SESSION_ID}.json"
SESSION_FILE="${STATE_DIR}/session.json"     # project-root: session state
PLANS_FILE="${PROJECT_ROOT}/Plans.md"        # project-root: Plans.md

# --- Suppression: worker session with cc:WIP tasks ---
# Mirrors the PreCompact role-gate in pre-compact-save.js so we don't nag a
# Worker that is already blocked from compacting.
if [ -f "${SESSION_FILE}" ] && command -v python3 >/dev/null 2>&1; then
  _role=$(python3 -c "
import json, sys
try:
    d = json.load(open(sys.argv[1]))
    print((d.get('role') or d.get('agent_role') or '').lower().strip())
except Exception:
    print('')
" "${SESSION_FILE}" 2>/dev/null || true)
  if [ "${_role}" = "worker" ] && grep -q "cc:WIP" "${PLANS_FILE}" 2>/dev/null; then
    echo "[HarnessCompact] suppressed (worker session with cc:WIP tasks)" >&2
    exit 0
  fi
fi

# --- Read + increment counter ---
mkdir -p "${STATE_DIR}" 2>/dev/null || true

_count=1
if [ -f "${COUNTER_FILE}" ] && command -v python3 >/dev/null 2>&1; then
  _count=$(python3 -c "
import json, sys
try:
    d = json.load(open(sys.argv[1]))
    print(int(d.get('count', 0)) + 1)
except Exception:
    print(1)
" "${COUNTER_FILE}" 2>/dev/null || echo 1)
fi

# Write updated counter (best-effort; never block on failure)
if command -v python3 >/dev/null 2>&1; then
  python3 -c "
import json, sys
d = {'count': int(sys.argv[1]), 'threshold': int(sys.argv[2]), 'interval': int(sys.argv[3])}
with open(sys.argv[4], 'w') as f:
    json.dump(d, f)
" "${_count}" "${THRESHOLD}" "${INTERVAL}" "${COUNTER_FILE}" 2>/dev/null || true
fi

# --- Decide whether to suggest ---
_suggest=false
if [ "${_count}" -eq "${THRESHOLD}" ]; then
  _suggest=true
elif [ "${_count}" -gt "${THRESHOLD}" ]; then
  _remainder=$(( (_count - THRESHOLD) % INTERVAL ))
  [ "${_remainder}" -eq 0 ] && _suggest=true
fi

if [ "${_suggest}" = "true" ]; then
  _msg="${_count} tool calls this session — consider \`/compact\` if transitioning phases or completing a milestone. Compacting now preserves the handoff artifact and re-injects Plans.md context after resume. Decision guide: harness/skills/harness-compact/references/decision-framework.md"
  echo "[HarnessCompact] ${_count} tool calls — strategic compact checkpoint" >&2
  if command -v python3 >/dev/null 2>&1; then
    python3 -c "import json,sys; print(json.dumps({'systemMessage':sys.argv[1]}))" \
      "${_msg}" 2>/dev/null || true
  fi
fi

exit 0
