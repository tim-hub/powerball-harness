---
name: maintenance
description: "Performs periodic cleanup — session log pruning, stale state files, orphaned worktrees, cache purge, trace archival. Use when performing routine housekeeping."
when_to_use: "prune logs, clear state, purge cache, clean worktrees, archive traces, housekeeping, maintenance, cleanup"
allowed-tools: ["Read", "Write", "Edit", "Bash", "Glob", "Grep"]
argument-hint: "[--prune-logs|--clear-state|--purge-cache|--clean-worktrees|--archive-traces|--all]"
user-invocable: true
---

# Maintenance

Periodic housekeeping skill for Harness projects. Handles session log pruning, stale state
file removal, plugin cache purging, and orphaned worktree cleanup.

## Quick Reference

| Subcommand | What It Does |
|------------|-------------|
| `--prune-logs` | Remove session log entries older than 90 days |
| `--clear-state` | Remove stale files from `.claude/state/` |
| `--purge-cache` | Clear the plugin cache from `~/.claude/plugins/cache/` |
| `--clean-worktrees` | Remove git worktrees with no associated branch |
| `--archive-traces` | Move per-task trace files for completed tasks older than 30 days to `.claude/memory/archive/traces/YYYY-MM/` |
| `--all` | Run all of the above in sequence |

## When to Use

Run `/maintenance` when any of the following apply:

- `session-log.md` exceeds 500 lines and the auto-cleanup hook has warned you
- `.claude/state/` contains `loop-active.json` but no loop is running
- Old sprint contracts from completed phases are cluttering `.claude/state/contracts/`
- `git worktree list` shows entries for branches that have been deleted
- The plugin cache has grown stale after a Harness version upgrade

## Operations

### --prune-logs

Scans `.claude/memory/session-log.md` and removes entries older than 90 days.
Preserves entries from the current month and the previous month unconditionally.

**Will delete**: Session log H2 sections (`## YYYY-MM`) older than 90 days.

**Will not touch**: `decisions.md`, `patterns.md`, or any archive files.

### --clear-state

Scans `.claude/state/` and removes files that are safe to delete:

- `loop-active.json` — only if no Breezing/harness-work loop is currently running
- Sprint contract files (`*.sprint-contract.json`) for tasks marked `cc:Done` in Plans.md
- Stale `review-result.json` files older than 7 days

**Will not touch**: Active loop state, pending fix proposals, or the `contracts/` directory itself.

### --purge-cache

Removes the plugin cache directory at `~/.claude/plugins/cache/`. The cache is rebuilt
automatically on next plugin load. Use after a major Harness version upgrade if stale
cached files are causing unexpected behavior.

### --archive-traces

Runs `bash "${CLAUDE_SKILL_DIR}/scripts/archive-traces.sh"` to move per-task execution
traces (schema: `trace.v1` — see `.claude/memory/schemas/trace.v1.md`) out of the active
`.claude/state/traces/` directory and into dated archive buckets once they are no longer
actively consumed.

**Eligibility** (both must hold):
1. The trace file's task is marked `cc:Done` in Plans.md
2. The file's mtime is older than `RETENTION_DAYS` (default: 30)

**Target layout**: `.claude/memory/archive/traces/YYYY-MM/<task_id>.jsonl`, bucketed by
each file's mtime so archives stay roughly chronological.

**Environment knobs** (rarely needed):
- `RETENTION_DAYS=N` — override the 30-day default
- `DRY_RUN=1` — print planned moves without executing them
- `VERBOSE=1` — log skip reasons for every trace file considered

**Idempotent**: a second run finds nothing to do because archived files are no longer
in `.claude/state/traces/`. Safe to invoke from cron.

**Destructive**: prompts for confirmation before deleting.

### --clean-worktrees

Lists all git worktrees (`git worktree list`) and removes entries where the associated
branch no longer exists in the repository.

Safe: uses `git worktree remove --force` only on worktrees confirmed to have no live branch.

## Safety Guarantees

- All operations print a dry-run summary before making changes
- Destructive operations (`--purge-cache`) require explicit confirmation
- Operations never touch files outside their defined scope
- The main working tree is never modified

## Execution Details

See [`${CLAUDE_SKILL_DIR}/references/cleanup.md`](${CLAUDE_SKILL_DIR}/references/cleanup.md)
for step-by-step execution procedures for each subcommand.

## Related Skills

- `harness-plan archive` — Archive completed phases from Plans.md (not the same as maintenance)
- `memory` — SSOT sync and memory management
- `harness-work` — Implementation; run maintenance before long breezing sessions
