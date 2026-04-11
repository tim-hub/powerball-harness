# sync Subcommand — Progress Sync Flow

Cross-reference implementation status with Plans.md, detecting and updating discrepancies.

## Step 0: Plans.md Validation

Verify the existence and format of Plans.md. If issues are found, provide guidance and stop immediately.

| State | Guidance |
|------|------|
| Plans.md does not exist | `Plans.md not found. Please create one with /harness-plan create.` -> **Stop** |
| Header lacks DoD / Depends columns (v1 format) | `Plans.md is in the legacy format (3 columns). Please regenerate with /harness-plan create for v2 (5 columns). Existing tasks will be automatically carried over.` -> **Stop** |
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

### Artifact Hash Backward Compatibility

Both `cc:done [a1b2c3d]` format (with commit hash) and `cc:done` (without hash) are recognized.

**Matching rules**:
- `cc:done` -> Treated as done without hash
- `cc:done [xxxxxxx]` -> Treated as done with hash. 7-character short hash is preserved
- When hash is present, commit existence can be verified against `git log --oneline`

> **Backward compatibility**: The hashless format remains valid. Existing Plans.md files are not broken.

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

During `sync` execution, automatically runs a retrospective when 1 or more `cc:done` tasks exist.
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
