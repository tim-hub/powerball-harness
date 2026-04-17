#!/usr/bin/env bash
# archive-traces.sh — move per-task trace files for completed-and-aged tasks
# from .claude/state/traces/ to .claude/memory/archive/traces/YYYY-MM/.
#
# Eligibility rules (all must hold):
#   1. The trace file's task is marked `cc:Done` in Plans.md
#   2. The file's mtime is older than RETENTION_DAYS (default: 30)
#
# Idempotent: a second run finds no eligible files because already-archived
# traces are no longer in .claude/state/traces/.
#
# Environment:
#   RETENTION_DAYS    Days since last modification to keep active (default 30)
#   DRY_RUN=1         Print actions without moving files
#   VERBOSE=1         Log skip reasons for each file
#
# Exit codes:
#   0  Success (even if nothing to archive)
#   2  Missing project root / cannot locate Plans.md
#
# Used by: skills/maintenance (--archive-traces subcommand). See SKILL.md.

set -euo pipefail

# project-root: user's repo, not the skill directory
PROJECT_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
TRACES_DIR="${PROJECT_ROOT}/.claude/state/traces"
PLANS_FILE="${PROJECT_ROOT}/Plans.md"
ARCHIVE_ROOT="${PROJECT_ROOT}/.claude/memory/archive/traces"
RETENTION_DAYS="${RETENTION_DAYS:-30}"
DRY_RUN="${DRY_RUN:-0}"
VERBOSE="${VERBOSE:-0}"

log() {
    [[ "$VERBOSE" == "1" ]] && echo "[archive-traces] $*" >&2
    return 0
}

log_always() {
    echo "[archive-traces] $*" >&2
}

# mtime_epoch returns the modification time of $1 as Unix seconds.
# Handles both BSD (macOS) and GNU (Linux) stat.
mtime_epoch() {
    if stat -f "%m" "$1" 2>/dev/null; then
        return 0
    fi
    stat -c "%Y" "$1"
}

# format_ym converts an epoch timestamp to "YYYY-MM".
# Handles both BSD (-r) and GNU (-d @) date.
format_ym() {
    local epoch="$1"
    if date -r "$epoch" +"%Y-%m" 2>/dev/null; then
        return 0
    fi
    date -d "@$epoch" +"%Y-%m"
}

# task_is_done checks whether Plans.md has a row for $1 with a cc:Done marker.
# Uses grep with an anchored regex to avoid substring false matches.
task_is_done() {
    local task_id="$1"
    # Escape dots in the task id for regex safety.
    local escaped
    escaped=$(printf '%s' "$task_id" | sed 's/\./\\./g')
    grep -qE "^\|[[:space:]]*${escaped}[[:space:]]*\|.*\bcc:Done\b" "$PLANS_FILE"
}

# resolve_task_id derives the Plans.md task id from a trace filename.
# Rotated files (72.1.1.jsonl) map to their base task (72.1); un-rotated
# files (72.1.jsonl) map to themselves (72.1).
resolve_task_id() {
    local filename="$1" # without the .jsonl suffix
    # If filename ends in .<digits> AND stripping that yields a Plans.md task,
    # use the stripped form. Otherwise use the full filename.
    if [[ "$filename" =~ ^(.+)\.([0-9]+)$ ]]; then
        local stripped="${BASH_REMATCH[1]}"
        if task_is_done "$stripped" 2>/dev/null; then
            echo "$stripped"
            return 0
        fi
    fi
    echo "$filename"
}

if [[ ! -d "$TRACES_DIR" ]]; then
    log_always "no traces directory ($TRACES_DIR) — nothing to do"
    exit 0
fi

if [[ ! -f "$PLANS_FILE" ]]; then
    log_always "no Plans.md at $PLANS_FILE — cannot determine cc:Done tasks"
    exit 2
fi

archived=0
skipped=0

shopt -s nullglob
for trace_file in "$TRACES_DIR"/*.jsonl; do
    filename="$(basename "$trace_file" .jsonl)"
    task_id="$(resolve_task_id "$filename")"

    if ! task_is_done "$task_id"; then
        log "skip $filename: task $task_id not cc:Done"
        skipped=$((skipped + 1))
        continue
    fi

    # find -mtime +N matches files modified more than N days ago.
    if ! find "$trace_file" -mtime +"$RETENTION_DAYS" -print -quit 2>/dev/null | grep -q .; then
        log "skip $filename: within ${RETENTION_DAYS}-day retention"
        skipped=$((skipped + 1))
        continue
    fi

    epoch="$(mtime_epoch "$trace_file")"
    ym="$(format_ym "$epoch")"
    target_dir="${ARCHIVE_ROOT}/${ym}"
    target_path="${target_dir}/${filename}.jsonl"

    if [[ "$DRY_RUN" == "1" ]]; then
        echo "DRY-RUN would move: $trace_file -> $target_path"
        continue
    fi

    mkdir -p "$target_dir"
    if [[ -e "$target_path" ]]; then
        log_always "conflict: $target_path already exists — leaving $trace_file in place"
        skipped=$((skipped + 1))
        continue
    fi
    mv "$trace_file" "$target_path"
    log_always "archived $filename -> ${ym}/"
    archived=$((archived + 1))
done

log_always "done: archived=$archived skipped=$skipped"
exit 0
