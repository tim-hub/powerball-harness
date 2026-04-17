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
| Header lacks DoD / Depends columns (v1 format) | `Plans.md is in the old format (3 columns). Please regenerate as v2 (5 columns) with harness-plan create. Existing tasks will be automatically carried over.` → **Stop** |
| v2 format (5 columns) | Proceed to Step 1 |

## Step 1: Gather Current State (parallel)

```bash
# Plans.md state
cat Plans.md

# Git change state
git status
git diff --stat HEAD~3

# Recent commit history
git log --oneline -10

# Agent trace (recently edited files)
tail -20 .claude/state/agent-trace.jsonl 2>/dev/null | jq -r '.files[].path' | sort -u
```

## Step 1.5: Agent Trace Analysis

Retrieve recent edit history from Agent Trace and cross-reference with Plans.md tasks:

```bash
# Recent edited files list
RECENT_FILES=$(tail -20 .claude/state/agent-trace.jsonl 2>/dev/null | \
  jq -r '.files[].path' | sort -u)

# Project information
PROJECT=$(tail -1 .claude/state/agent-trace.jsonl 2>/dev/null | \
  jq -r '.metadata.project')
```

**Cross-reference points**:

| Check Item | Detection Method |
|------------|-----------------|
| File edits not in Plans.md | Agent Trace vs task descriptions |
| Files differing from task descriptions | Expected files vs actual edits |
| Tasks with no edits for a long time | Agent Trace timeline vs WIP duration |

## Step 2: Drift Detection

| Check Item | Detection Method |
|------------|-----------------|
| Completed but still `cc:WIP` | Commit history vs markers |
| Started but still `cc:TODO` | Changed files vs markers |
| `cc:Done` but uncommitted | git status vs markers |

## Step 3: Plans.md Update Proposal

When drift is detected, propose and execute updates:

```
Plans.md updates needed

| Task | Current | Proposed | Reason |
|------|---------|----------|--------|
| XX   | cc:WIP  | cc:Done  | Already committed |
| YY   | cc:TODO | cc:WIP   | Files already edited |

Proceed with updates? (yes / no)
```

## Step 4: Progress Summary Output

```markdown
## Progress Summary

**Project**: {{project_name}}

| Status | Count |
|--------|-------|
| Not Started (cc:TODO) | {{count}} |
| In Progress (cc:WIP) | {{count}} |
| Done (cc:Done) | {{count}} |
| PM Confirmed (pm:confirmed) | {{count}} |

**Progress Rate**: {{percent}}%

### Recently Edited Files (Agent Trace)
- {{file1}}
- {{file2}}
```

## Step 4.5: Snapshot Save (`--snapshot` specified)

When `--snapshot` is specified, save the current progress state as a timestamped snapshot.

### Storage Location

Save in JSON format to the `.claude/state/snapshots/` directory:

```bash
SNAPSHOT_DIR="${PROJECT_ROOT}/.claude/state/snapshots"
mkdir -p "${SNAPSHOT_DIR}"
SNAPSHOT_FILE="${SNAPSHOT_DIR}/progress-$(date -u +%Y%m%dT%H%M%SZ).json"
```

### Snapshot Contents

```json
{
  "timestamp": "2026-03-08T10:30:00Z",
  "phase": "Phase 26",
  "progress": {
    "total": 16,
    "todo": 5,
    "wip": 3,
    "done": 6,
    "confirmed": 2
  },
  "progress_rate": 50,
  "recent_commits": ["abc1234 feat: ...", "def5678 fix: ..."],
  "recent_files": ["skills/harness-work/SKILL.md", "..."],
  "notes": ""
}
```

### Diff Comparison

When a previous snapshot exists, display the diff:

```markdown
## Snapshot Diff

| Metric | Previous ({{prev_time}}) | Current | Change |
|--------|--------------------------|---------|--------|
| Progress Rate | {{prev}}% | {{current}}% | +{{diff}}%pt |
| Done Tasks | {{prev_done}} | {{current_done}} | +{{diff_done}} |
| WIP Tasks | {{prev_wip}} | {{current_wip}} | {{diff_wip}} |
```

> **Design Intent**: Snapshots are for manual use when the user wants to record the current state.
> This is separate from the automatic progress feed during breezing (26.2.3).

## Step 5: Next Action Proposal

```
Next steps

**Priority 1**: {{task}}
- Reason: {{in progress / waiting to unblock}}

**Recommended**: harness-work, harness-review
```

## Anomaly Detection

| Situation | Warning |
|-----------|---------|
| Multiple `cc:WIP` | Multiple tasks in progress simultaneously |
| Unprocessed `pm:requested` | Process PM requests first |
| Large drift | Task management is not keeping up |
| WIP with no updates for 3+ days | Check if blocked |

## Step 6: Retrospective (ON by default)

Automatically runs a retrospective when 1 or more `cc:Done` tasks exist.
Can be explicitly skipped with `--no-retro`.

### Step R1: Collect Completed Tasks

```bash
# Extract cc:Done / pm:confirmed tasks from Plans.md
grep -E 'cc:Done|pm:confirmed' Plans.md

# Recent completion commit history
git log --oneline --since="7 days ago"

# Change scale
git diff --stat HEAD~10
```

### Step R2: Four Retrospective Items

| Item | Analysis Method |
|------|-----------------|
| **Estimation Accuracy** | Infer expected file count from Plans.md task descriptions → compare with actual changed file count from `git diff --stat` |
| **Block Causes** | Aggregate reason patterns for tasks with `blocked` markers (technical/external dependency/unclear specs) |
| **Quality Marker Accuracy** | Check whether tasks tagged with `[feature:security]` etc. actually encountered related issues |
| **Scope Changes** | Compare task count from initial Plans.md commit vs current (additions/deletions) |

### Step R3: Retrospective Summary Output

```markdown
## Retrospective Summary

**Period**: {{start_date}} - {{end_date}}

| Metric | Value |
|--------|-------|
| Completed Tasks | {{count}} |
| Blocks Occurred | {{blocked_count}} |
| Scope Changes | +{{added}} / -{{removed}} |
| Estimation Accuracy | Expected {{est}} files → Actual {{actual}} files |

### Learnings
- {{1-2 line learning}}

### Improvements for Next Time
- {{1-2 line improvement action}}
```

### Step R4: Recording to harness-mem

Record retrospective results to harness-mem for reference in future `create` invocations.
Storage location: the corresponding agent memory under `.claude/agent-memory/`.

## Related Skills

- `harness-plan` — Plan creation and task management
- `harness-work` — Task implementation
- `harness-review` — Code review
