---
name: harness-sync
description: "Detects drift between Plans.md markers and actual implementation, then corrects them. Use when checking sync status or running a retrospective."
when_to_use: "sync status, drift check, where am I, retrospective, markers out of date, plans out of sync, snapshot"
allowed-tools: ["Read", "Edit", "Bash", "Grep", "Glob"]
argument-hint: "[--snapshot|--no-retro]"
effort: medium
---

# Harness Sync

Compares Plans.md against actual implementation status, detecting and updating discrepancies.
Standalone version of the former `sync-status` and `harness-plan sync` subcommands.

## Quick Reference

| User Input | Subcommand | Behavior |
|------------|------------|----------|
| `harness-sync` | (default) | Progress sync + retrospective (ON by default) |
| `harness-sync --no-retro` | `--no-retro` | Progress sync only (skip retrospective) |
| `harness-sync --snapshot` | `--snapshot` | Save snapshot (point-in-time progress record) |
| "Where am I?" / "Check progress" | (default) | Same as default sync |

## Options

| Option | Description | Default |
|--------|-------------|---------|
| `--snapshot` | Save current progress as a snapshot | false |
| `--no-retro` | Skip retrospective | false (runs by default) |

## Step 0: Plans.md Validation

Verify Plans.md existence and format. If there are issues, provide guidance and stop immediately.

| State | Guidance |
|-------|----------|
| Plans.md does not exist | `Plans.md not found. Please create one with harness-plan create.` → **Stop** |
| Header lacks DoD / Depends columns (v1 format) | `Plans.md is in the old format (3 columns). Please regenerate as v2 (5 columns) with harness-plan create.` → **Stop** |
| v2 format (5 columns) | Proceed to Step 1 |

## Step 1: Gather Current State (parallel)

```bash
cat Plans.md
git status
git diff --stat HEAD~3
git log --oneline -10
tail -20 .claude/state/agent-trace.jsonl 2>/dev/null | jq -r '.files[].path' | sort -u
```

See [`${CLAUDE_SKILL_DIR}/references/sync-details.md`](${CLAUDE_SKILL_DIR}/references/sync-details.md) for Agent Trace cross-reference analysis (Step 1.5).

## Step 2: Drift Detection

| Check Item | Detection Method |
|------------|-----------------|
| Completed but still `cc:WIP` | Commit history vs markers |
| Started but still `cc:TODO` | Changed files vs markers |
| `cc:Done` but uncommitted | git status vs markers |

Full update proposal, progress summary, snapshot format, next-action proposal, and retrospective flow:
[`${CLAUDE_SKILL_DIR}/references/sync-details.md`](${CLAUDE_SKILL_DIR}/references/sync-details.md)

## Anomaly Detection

| Situation | Warning |
|-----------|---------|
| Multiple `cc:WIP` | Multiple tasks in progress simultaneously |
| Unprocessed `pm:requested` | Process PM requests first |
| Large drift | Task management is not keeping up |
| WIP with no updates for 3+ days | Check if blocked |

## Related Skills

- `harness-plan` — Plan creation and task management
- `harness-work` — Task implementation
- `harness-review` — Code review
