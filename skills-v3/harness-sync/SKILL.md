---
name: harness-sync
description: "Use this skill whenever the user asks to sync progress, check what's done, see current status, asks 'where am I', 'how far along', or runs /harness-sync. Also supports --snapshot for progress snapshots. Do NOT load for: creating new plans (use harness-plan), code implementation (use harness-work), code review (use harness-review), or release. Syncs progress between Plans.md and actual implementation — detects drift, updates task markers, and runs retrospectives."
allowed-tools: ["Read", "Edit", "Bash", "Grep", "Glob"]
argument-hint: "[--snapshot|--no-retro]"
effort: medium
---

# Harness Sync

Compares Plans.md against implementation status, detecting and updating discrepancies.
Standalone version of the former `sync-status` and `harness-plan sync` subcommand.

## Quick Reference

| User Input | Behavior |
|------------|------|
| `harness-sync` | Progress sync + retrospective (default ON) |
| `harness-sync --no-retro` | Progress sync only (skip retrospective) |
| `harness-sync --snapshot` | Save snapshot (point-in-time progress record) |
| "where am I?" / "check progress" | Same as above |

## Options

| Option | Description | Default |
|----------|------|----------|
| `--snapshot` | Save current progress as a snapshot | false |
| `--no-retro` | Skip retrospective | false (runs by default) |

## Step 0: Plans.md Validation

Verify the existence and format of Plans.md. If issues are found, provide guidance and stop immediately.

| State | Guidance |
|------|------|
| Plans.md does not exist | `Plans.md not found. Please create one with harness-plan create.` -> **Stop** |
| Header lacks DoD / Depends columns (v1 format) | `Plans.md is in the legacy format (3 columns). Please regenerate with harness-plan create for v2 (5 columns). Existing tasks will be automatically carried over.` -> **Stop** |
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

# Project info
PROJECT=$(tail -1 .claude/state/agent-trace.jsonl 2>/dev/null | \
  jq -r '.metadata.project')
```

**Cross-reference points**:

| Check Item | Detection Method |
|------------|----------|
| File edits not in Plans.md | Agent Trace vs task descriptions |
| Files differing from task description | Expected files vs actual edits |
| Tasks with no recent edits | Agent Trace timeline vs WIP duration |

## Step 2: Drift Detection

| Check Item | Detection Method |
|------------|----------|
| Completed but still `cc:WIP` | Commit history vs markers |
| Started but still `cc:TODO` | Changed files vs markers |
| `cc:done` but uncommitted | git status vs markers |

## Step 3: Plans.md Update Proposal

When drift is detected, propose and execute updates:

```
Plans.md updates needed

| Task | Current | After | Reason |
|------|------|--------|------|
| XX   | cc:WIP | cc:done | Already committed |
| YY   | cc:TODO | cc:WIP | Files already edited |

Apply updates? (yes / no)
```

## Step 4: Progress Summary Output

```markdown
## Progress Summary

**Project**: {{project_name}}

| Status | Count |
|----------|------|
| Not started (cc:TODO) | {{count}} |
| In progress (cc:WIP) | {{count}} |
| Done (cc:done) | {{count}} |
| PM confirmed (pm:confirmed) | {{count}} |

**Progress**: {{percent}}%

### Recently Edited Files (Agent Trace)
- {{file1}}
- {{file2}}
```

## Step 4.5: Snapshot Save (`--snapshot` specified)

When `--snapshot` is specified, save the current progress state as a timestamped snapshot.

### Save Location

Saved in JSON format under the `.claude/state/snapshots/` directory:

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

If a previous snapshot exists, display the diff:

```markdown
## Snapshot Diff

| Metric | Previous ({{prev_time}}) | Current | Change |
|------|---------------------|------|------|
| Progress rate | {{prev}}% | {{current}}% | +{{diff}}%pt |
| Completed tasks | {{prev_done}} | {{current_done}} | +{{diff_done}} |
| WIP tasks | {{prev_wip}} | {{current_wip}} | {{diff_wip}} |
```

> **Design intent**: Snapshots are for manual use when the user wants to "record the current state."
> They are separate from the automatic progress feed during breezing (26.2.3).

## Step 5: Next Action Proposal

```
What to do next

**Priority 1**: {{task}}
- Reason: {{pending request / waiting for unblock}}

**Recommended**: harness-work, harness-review
```

## Anomaly Detection

| Situation | Warning |
|------|------|
| Multiple `cc:WIP` | Multiple tasks in progress simultaneously |
| Unprocessed `pm:requested` | Process PM's request first |
| Large drift | Task management is falling behind |
| WIP without updates for 3+ days | Check if blocked |

## Step 6: Retrospective (default ON)

Automatically runs a retrospective when 1 or more `cc:done` tasks exist.
Can be explicitly skipped with `--no-retro`.

### Step R1: Collect Completed Tasks

```bash
# Extract cc:done / pm:confirmed from Plans.md
grep -E 'cc:done|pm:confirmed' Plans.md

# Recent completion commit history
git log --oneline --since="7 days ago"

# Change volume
git diff --stat HEAD~10
```

### Step R2: Retrospective 4 Items

| Item | Analysis Method |
|------|---------|
| **Estimation accuracy** | Infer expected file count from Plans.md task descriptions -> Compare with actual changed file count from `git diff --stat` |
| **Block causes** | Aggregate reason patterns for tasks with `blocked` markers (technical/external dependency/unclear spec) |
| **Quality marker accuracy** | Did tasks tagged with `[feature:security]` etc. actually encounter related issues? |
| **Scope variation** | Task count at initial Plans.md commit vs current task count (additions/removals) |

### Step R3: Retrospective Summary Output

```markdown
## Retrospective Summary

**Period**: {{start_date}} - {{end_date}}

| Metric | Value |
|------|-----|
| Completed tasks | {{count}} |
| Blocks occurred | {{blocked_count}} |
| Scope variation | +{{added}} / -{{removed}} |
| Estimation accuracy | Expected {{est}} files -> Actual {{actual}} files |

### Learnings
- {{1-2 lines of learnings}}

### Actions for Next Time
- {{1-2 lines of improvement actions}}
```

### Step R4: Record to harness-mem

Record retrospective results to harness-mem so they can be referenced during the next `create`.
Save location: Under the corresponding agent memory in `.claude/agent-memory/`.

## Related Skills

- `harness-plan` — Plan creation and task management
- `harness-work` — Task implementation
- `harness-review` — Code review
