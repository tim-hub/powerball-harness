#!/usr/bin/env bash
# advisor-check-cache.sh — look up a cached advisor decision by the triple
# (task_id, reason_code, error_signature) in
# .claude/state/advisor/history.jsonl.
#
# Referenced by harness/agents/advisor.md's "Duplicate Suppression" step: the
# advisor MUST call this script first and return the cached decision when HIT;
# only on MISS should it proceed to load `context_sources` and reason fresh.
# This preserves the cache-first ordering that Phase 73.4 is the unit test for.
#
# Usage:
#   advisor-check-cache.sh --task <id> --reason <code> --sig <normalized_sig>
#
# Output:
#   HIT  → most recent matching decision JSON on stdout (one line, unchanged);
#          exit 0
#   MISS → empty stdout; exit 1
#   usage error → stderr; exit 2

set -euo pipefail

PROJECT_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
HISTORY_FILE="$PROJECT_ROOT/.claude/state/advisor/history.jsonl"

TASK_ID=""
REASON=""
SIG=""

while [[ $# -gt 0 ]]; do
    case "$1" in
        --task)    TASK_ID="${2:-}"; shift 2 ;;
        --reason)  REASON="${2:-}";  shift 2 ;;
        --sig)     SIG="${2:-}";     shift 2 ;;
        -h|--help)
            sed -n '1,19p' "$0" | sed 's/^# \{0,1\}//'
            exit 0
            ;;
        *)
            echo "unknown argument: $1" >&2
            exit 2
            ;;
    esac
done

[[ -n "$TASK_ID" ]] || { echo "--task is required" >&2; exit 2; }
[[ -n "$REASON"  ]] || { echo "--reason is required" >&2; exit 2; }
[[ -n "$SIG"     ]] || { echo "--sig is required" >&2; exit 2; }

if [[ ! -f "$HISTORY_FILE" ]]; then
    exit 1   # MISS — no history yet
fi

# Prefer jq for authoritative JSON parsing. Fall back to grep when jq is
# unavailable (minimal CI images); the fallback is best-effort on compact
# JSON and will correctly MISS if fields are present in unusual order.
if command -v jq >/dev/null 2>&1; then
    match=$(
        jq -c \
           --arg t "$TASK_ID" \
           --arg r "$REASON" \
           --arg s "$SIG" \
           'select(.task_id == $t and .reason_code == $r and .error_signature == $s)' \
           "$HISTORY_FILE" 2>/dev/null | tail -1
    )
else
    pattern="\"task_id\":\"$TASK_ID\".*\"reason_code\":\"$REASON\".*\"error_signature\":\"$SIG\""
    match=$(grep -E "$pattern" "$HISTORY_FILE" 2>/dev/null | tail -1 || true)
fi

if [[ -z "$match" ]]; then
    exit 1   # MISS
fi

printf '%s\n' "$match"
exit 0   # HIT
