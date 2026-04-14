# sync Subcommand -- Progress Sync Flow

Compares implementation status against Plans.md, detects discrepancies, and updates accordingly.

## Step 0: Plans.md Validation

Verify the existence and format of Plans.md. If there are issues, provide guidance and stop immediately.

| State | Guidance |
|-------|----------|
| Plans.md does not exist | `Plans.md not found. Create one with /harness-plan create.` -> **Stop** |
| Header lacks DoD / Depends columns (v1 format) | `Plans.md is in the old format (3 columns). Regenerate as v2 (5 columns) with /harness-plan create. Existing tasks will be carried over automatically.` -> **Stop** |
| v2 format (5 columns) | Proceed to Step 1 |

## Step 1: Collect Current State (Parallel)

```bash
# Plans.md state
cat Plans.md

# Git change status
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
# Recent edited file list
RECENT_FILES=$(tail -20 .claude/state/agent-trace.jsonl 2>/dev/null | \
  jq -r '.files[].path' | sort -u)

# Project info
PROJECT=$(tail -1 .claude/state/agent-trace.jsonl 2>/dev/null | \
  jq -r '.metadata.project')
```

**Cross-reference points**:

| Check Item | Detection Method |
|-----------|-----------------|
| File edits not in Plans.md | Agent Trace vs task descriptions |
| Files differing from task description | Expected files vs actual edits |
| Tasks with no recent edits | Agent Trace timeline vs WIP duration |

## Step 2: Discrepancy Detection

| Check Item | Detection Method |
|-----------|-----------------|
| Completed but still `cc:WIP` | Commit history vs marker |
| Started but still `cc:TODO` | Changed files vs marker |
| `cc:done` but uncommitted | git status vs marker |

### Artifact Hash Backward Compatibility

Recognizes both `cc:done [a1b2c3d]` format (with commit hash) and `cc:done` (without hash).

**Matching rules**:
- `cc:done` -> Treated as done without hash
- `cc:done [xxxxxxx]` -> Treated as done with hash. Retains the 7-character short hash
- When hash is present, can verify commit existence by cross-referencing with `git log --oneline`

> **Backward compatibility**: The hashless format remains valid. Does not break existing Plans.md files.

## Step 3: Plans.md Update Proposal

When discrepancies are detected, propose and execute:

```
Plans.md update needed

| Task | Current | New | Reason |
|------|---------|-----|--------|
| XX   | cc:WIP | cc:done | Already committed |
| YY   | cc:TODO | cc:WIP | Files already edited |

Proceed with update? (yes / no)
```

## Step 4: Progress Summary Output

```markdown
## Progress Summary

**Project**: {{project_name}}

| Status | Count |
|--------|-------|
| Not Started (cc:TODO) | {{count}} |
| In Progress (cc:WIP) | {{count}} |
| Completed (cc:done) | {{count}} |
| PM Confirmed (pm:confirmed) | {{count}} |

**Progress**: {{percent}}%

### Recently Edited Files (Agent Trace)
- {{file1}}
- {{file2}}
```

## Step 5: Next Action Suggestion

```
Next steps

**Priority 1**: {{task}}
- Reason: {{requested / waiting for unblock}}

**Recommended**: harness-work, harness-review
```

## Anomaly Detection

| Situation | Warning |
|-----------|---------|
| Multiple `cc:WIP` | Multiple tasks in progress simultaneously |
| Unprocessed `pm:requested` | Process PM requests first |
| Large discrepancy | Task management is falling behind |
| WIP with no updates for 3+ days | Check if blocked |

## Step 6: Retrospective (ON by default)

When `sync` runs, if there is at least 1 `cc:done` task, a retrospective is automatically executed.
Can be explicitly skipped with `--no-retro`.

### Step R1: Collect Completed Tasks

```bash
# Extract cc:done / pm:confirmed tasks from Plans.md
grep -E 'cc:done|pm:confirmed' Plans.md

# Recent completion commit history
git log --oneline --since="7 days ago"

# Change scale
git diff --stat HEAD~10
```

### Step R2: Four Retrospective Items

| Item | Analysis Method |
|------|----------------|
| **Estimation accuracy** | Infer expected file count from Plans.md task descriptions -> Compare with actual changed file count from `git diff --stat` |
| **Block causes** | Aggregate reason patterns for tasks with `blocked` marker (technical / external dependency / unclear spec) |
| **Quality marker accuracy** | Check whether tasks tagged `[feature:security]` etc. actually had related issues |
| **Scope changes** | Task count at initial Plans.md commit vs current task count (additions/deletions) |

### Step R3: Retrospective Summary Output

```markdown
## Retrospective Summary

**Period**: {{start_date}} -- {{end_date}}

| Metric | Value |
|--------|-------|
| Completed tasks | {{count}} |
| Block occurrences | {{blocked_count}} |
| Scope changes | +{{added}} / -{{removed}} |
| Estimation accuracy | Expected {{est}} files -> Actual {{actual}} files |

### Learnings
- {{1-2 line learning}}

### Action items for next time
- {{1-2 line improvement action}}
```

### Step R4: Record to harness-mem

Record retrospective results to harness-mem so they can be referenced in future `create` runs.
Destination: Agent memory under `.claude/agent-memory/` for the corresponding agent.
