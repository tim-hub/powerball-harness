#!/usr/bin/env bash
# show-stats.sh
# harness-compact stats subcommand: prints current session compact state.
# Output: deterministic key=value lines (one per stat).

set -euo pipefail

SESSION_ID="${CLAUDE_SESSION_ID:-default}"
THRESHOLD="${HARNESS_COMPACT_THRESHOLD:-50}"
INTERVAL="${HARNESS_COMPACT_INTERVAL:-25}"

# project-root: user's git repository root
PROJECT_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
STATE_DIR="${PROJECT_ROOT}/.claude/state"
COUNTER_FILE="${STATE_DIR}/compact-counter-${SESSION_ID}.json"
SESSION_FILE="${STATE_DIR}/session.json"  # project-root: session state
PLANS_FILE="${PROJECT_ROOT}/Plans.md"     # project-root: Plans.md

# --- Read counter ---
_count=0
if [ -f "${COUNTER_FILE}" ] && command -v python3 >/dev/null 2>&1; then
  _count=$(python3 -c "
import json, sys
try:
    d = json.load(open(sys.argv[1]))
    print(int(d.get('count', 0)))
except Exception:
    print(0)
" "${COUNTER_FILE}" 2>/dev/null || echo 0)
fi

# --- Next reminder ---
_next_reminder_at="${THRESHOLD}"
if [ "${_count}" -ge "${THRESHOLD}" ]; then
  _elapsed=$(( _count - THRESHOLD ))
  _intervals=$(( _elapsed / INTERVAL ))
  _next_reminder_at=$(( THRESHOLD + (_intervals + 1) * INTERVAL ))
fi

# --- Session age ---
_session_age_minutes=0
if [ -f "${SESSION_FILE}" ] && command -v python3 >/dev/null 2>&1; then
  _session_age_minutes=$(python3 -c "
import json, sys
from datetime import datetime, timezone
try:
    d = json.load(open(sys.argv[1]))
    started = d.get('started_at') or d.get('start_time') or ''
    if started:
        dt = datetime.fromisoformat(started.replace('Z', '+00:00'))
        age = (datetime.now(timezone.utc) - dt).total_seconds() / 60
        print(int(max(0, age)))
    else:
        print(0)
except Exception:
    print(0)
" "${SESSION_FILE}" 2>/dev/null || echo 0)
fi

# --- WIP task count ---
_wip_task_count=0
if [ -f "${PLANS_FILE}" ]; then
  _wip_task_count=$(grep -c "cc:WIP" "${PLANS_FILE}" 2>/dev/null || echo 0)
fi

# --- Output (deterministic key=value) ---
printf "count=%d\n"              "${_count}"
printf "threshold=%d\n"          "${THRESHOLD}"
printf "next_reminder_at=%d\n"   "${_next_reminder_at}"
printf "session_age_minutes=%d\n" "${_session_age_minutes}"
printf "wip_task_count=%d\n"     "${_wip_task_count}"
