#!/usr/bin/env bash
# advisor-load-context.sh — fetch scoped raw context for the advisor agent
#
# Referenced by harness/agents/advisor.md as the scoped loader invoked when
# the caller passes `context_sources` on a cache miss (Phase 73.2).
#
# Usage:
#   advisor-load-context.sh --task <task_id> --sources <csv>
#
# Flags:
#   --task <id>              Plans.md task id (e.g. "73.4")
#   --sources <csv>          Comma-separated: trace,git_diff,session_log,patterns
#   --per-source-cap <N>     Per-source byte cap (default 10240 = 10 KiB)
#
# Output: markdown sections, one per requested source. Each section begins with
# `## source: <name> (size: <N> bytes)`. Missing files / empty matches produce
# a terse placeholder rather than an error — the advisor must proceed with
# whatever context is available.
#
# Exit codes:
#   0  success (even when some sources are empty)
#   2  usage error (missing/invalid flags)

set -euo pipefail

PROJECT_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"

TASK_ID=""
SOURCES=""
PER_SOURCE_CAP="${PER_SOURCE_CAP:-10240}"

while [[ $# -gt 0 ]]; do
    case "$1" in
        --task)
            TASK_ID="${2:-}"
            shift 2
            ;;
        --sources)
            SOURCES="${2:-}"
            shift 2
            ;;
        --per-source-cap)
            PER_SOURCE_CAP="${2:-}"
            shift 2
            ;;
        -h|--help)
            sed -n '1,25p' "$0" | sed 's/^# \{0,1\}//'
            exit 0
            ;;
        *)
            echo "unknown argument: $1" >&2
            exit 2
            ;;
    esac
done

[[ -n "$TASK_ID"  ]] || { echo "--task is required" >&2; exit 2; }
[[ -n "$SOURCES"  ]] || { echo "--sources is required" >&2; exit 2; }

# trim_content applies a head+tail truncation to stdin so the advisor sees
# both the task-start context and the most recent events, with an explicit
# marker showing what was dropped. Simple truncation would lose whichever end
# the advisor cares most about (typically the recent end).
trim_content() {
    local cap="$1"
    local content
    content=$(cat)
    local size=${#content}
    if (( size <= cap )); then
        printf '%s' "$content"
        return 0
    fi
    # Reserve ~40 bytes for the marker; split the remainder in half.
    local half=$(( (cap - 40) / 2 ))
    if (( half <= 0 )); then
        # Cap is tiny; just hard-truncate head.
        printf '%s' "${content:0:$cap}"
        return 0
    fi
    local head_part="${content:0:$half}"
    local tail_part="${content: -$half}"
    local dropped=$(( size - (half * 2) ))
    printf '%s\n\n… [%d bytes trimmed] …\n\n%s' "$head_part" "$dropped" "$tail_part"
}

# load_trace: Phase 72 per-task trace — the primary causal signal for
# repeated_failure reasoning.
load_trace() {
    local path="$PROJECT_ROOT/.claude/state/traces/$TASK_ID.jsonl"
    if [[ ! -f "$path" ]]; then
        echo "(no trace file at .claude/state/traces/$TASK_ID.jsonl)"
        return 0
    fi
    cat "$path"
}

# load_git_diff: diff from the task-start commit (the first commit whose
# message mentions "($TASK_ID)") through HEAD. Leverages this project's
# commit-message convention of embedding task ids in parens. Falls back to
# working-tree diff when no such commit exists yet.
load_git_diff() {
    local start_commit
    start_commit=$(
        git -C "$PROJECT_ROOT" log --grep="($TASK_ID)" --format=%H 2>/dev/null | tail -1
    )
    if [[ -z "$start_commit" ]]; then
        # Task hasn't been committed yet — show current working-tree diff.
        git -C "$PROJECT_ROOT" diff 2>/dev/null || echo "(no working-tree diff)"
        return 0
    fi
    # Include the task-start commit itself by diffing from its parent.
    git -C "$PROJECT_ROOT" diff "${start_commit}~1..HEAD" 2>/dev/null \
        || git -C "$PROJECT_ROOT" diff "$start_commit..HEAD" 2>/dev/null \
        || echo "(could not diff from task-start commit $start_commit)"
}

# load_session_log: grep session-log entries mentioning $TASK_ID, with 3
# lines of context on each side. Catches cross-session details the worker
# may have lost — primary signal for plateau_before_escalation.
load_session_log() {
    local path="$PROJECT_ROOT/.claude/memory/session-log.md"
    if [[ ! -f "$path" ]]; then
        echo "(no session log)"
        return 0
    fi
    # grep -C 3 for context; -F for literal match (task ids have dots).
    local matches
    matches=$(grep -n -C 3 -F "$TASK_ID" "$path" 2>/dev/null || true)
    if [[ -z "$matches" ]]; then
        echo "(no session-log mentions of $TASK_ID)"
        return 0
    fi
    printf '%s' "$matches"
}

# load_patterns: H2 headings list only. Full patterns.md is ~400 lines and
# would blow any reasonable per-source cap. The advisor can Read the file
# directly if it needs a specific pattern's detail.
load_patterns() {
    local path="$PROJECT_ROOT/.claude/memory/patterns.md"
    if [[ ! -f "$path" ]]; then
        echo "(no patterns.md)"
        return 0
    fi
    local headings
    headings=$(grep -E "^## P[0-9]+:" "$path" 2>/dev/null || true)
    if [[ -z "$headings" ]]; then
        echo "(no P-pattern headings found)"
        return 0
    fi
    echo "Pattern catalog (headings only — Read patterns.md for details):"
    echo ""
    printf '%s' "$headings"
}

# main loop: iterate over requested sources, emitting one markdown section
# per source with size metadata in the heading.
IFS=',' read -ra SRC_ARRAY <<< "$SOURCES"
for raw_src in "${SRC_ARRAY[@]}"; do
    # Strip whitespace (tolerate "trace, git_diff" spelled with spaces)
    src=$(echo "$raw_src" | tr -d '[:space:]')
    [[ -z "$src" ]] && continue

    case "$src" in
        trace|git_diff|session_log|patterns)
            ;;
        *)
            printf '## source: %s (unknown — skipped)\n\n' "$src"
            continue
            ;;
    esac

    tmp=$(mktemp "${TMPDIR:-/tmp}/advisor-ctx.XXXXXX")
    trap 'rm -f "$tmp"' EXIT
    case "$src" in
        trace)       load_trace       > "$tmp" ;;
        git_diff)    load_git_diff    > "$tmp" ;;
        session_log) load_session_log > "$tmp" ;;
        patterns)    load_patterns    > "$tmp" ;;
    esac

    size=$(wc -c < "$tmp" | tr -d ' ')
    if (( size > PER_SOURCE_CAP )); then
        printf '## source: %s (size: %d bytes, trimmed to %d)\n\n' "$src" "$size" "$PER_SOURCE_CAP"
    else
        printf '## source: %s (size: %d bytes)\n\n' "$src" "$size"
    fi
    trim_content "$PER_SOURCE_CAP" < "$tmp"
    printf '\n\n'
    rm -f "$tmp"
    trap - EXIT
done
