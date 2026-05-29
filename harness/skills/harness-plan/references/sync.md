# sync Subcommand -- Progress Sync Flow

Compares implementation status against the plan (`harness plan-cli list`/`get`), detects discrepancies, and updates accordingly via `harness plan-cli update`. The SSOT is `.claude/harness/plans.json`.

## Step 0: Plan Validation

Verify the plan exists and is readable. If there are issues, provide guidance and stop immediately.

| State | Guidance |
|-------|----------|
| Plan is empty / no phases (`harness plan-cli list` returns nothing) | `No plan found. Create one with /harness-plan create.` -> **Stop** |
| Legacy `Plans.md` exists but `.claude/harness/plans.json` does not | `Legacy Plans.md detected. Run \`harness plan-cli migrate\` to convert it to the JSON SSOT.` -> **Stop** |
| `.claude/harness/plans.json` present | Proceed to Step 1 |

## Step 1: Collect Current State (Parallel)

```bash
# Plan state (SSOT is .claude/harness/plans.json)
harness plan-cli list

# Git change status
git status
git diff --stat HEAD~3

# Recent commit history
git log --oneline -10

# Agent trace (recently edited files)
tail -20 .claude/state/agent-trace.jsonl 2>/dev/null | jq -r '.files[].path' | sort -u
```

## Step 1.5: Agent Trace Analysis

Retrieve recent edit history from Agent Trace and cross-reference with plan tasks (`harness plan-cli list`):

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
| File edits not in any plan task | Agent Trace vs task descriptions |
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

> **Backward compatibility**: The hashless format remains valid. Does not break existing plans.

## Step 3: Plan Update Proposal

When discrepancies are detected, propose and (on approval) execute each change with `harness plan-cli update <task-id> --status <status>` (add `--hash` for `cc:done`, `--reason` for `blocked`):

```
Plan update needed

| Task | Current | New | Reason |
|------|---------|-----|--------|
| XX   | cc:WIP | cc:done | Already committed |
| YY   | cc:TODO | cc:WIP | Files already edited |

Proceed with update? (yes / no)
```

On "yes", apply each row, e.g.:

```bash
harness plan-cli update XX --status cc:done --hash <7char>
harness plan-cli update YY --status cc:WIP
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

## Step 4.5: Snapshot Save (`--snapshot` flag)

When `--snapshot` is specified, save the current progress as a timestamped JSON record:

```bash
SNAPSHOT_DIR="${PROJECT_ROOT}/.claude/state/snapshots"
mkdir -p "${SNAPSHOT_DIR}"
SNAPSHOT_FILE="${SNAPSHOT_DIR}/progress-$(date -u +%Y%m%dT%H%M%SZ).json"
```

Snapshot schema:

```json
{
  "timestamp": "2026-03-08T10:30:00Z",
  "phase": "Phase N",
  "progress": { "total": 16, "todo": 5, "wip": 3, "done": 6, "confirmed": 2 },
  "progress_rate": 50,
  "recent_commits": ["abc1234 feat: ..."],
  "recent_files": ["skills/harness-work/SKILL.md"],
  "notes": ""
}
```

When a previous snapshot exists, display a diff table (Previous vs Current: progress rate, done count, WIP count).

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
# List cc:done / pm:confirmed tasks from the plan
harness plan-cli list --status cc:done
harness plan-cli list --status pm:confirmed

# Recent completion commit history
git log --oneline --since="7 days ago"

# Change scale
git diff --stat HEAD~10
```

### Step R2: Four Retrospective Items

| Item | Analysis Method |
|------|----------------|
| **Estimation accuracy** | Infer expected file count from plan task descriptions -> Compare with actual changed file count from `git diff --stat` |
| **Block causes** | Aggregate reason patterns for tasks with `blocked` marker (technical / external dependency / unclear spec) |
| **Quality marker accuracy** | Check whether tasks tagged `[feature:security]` etc. actually had related issues |
| **Scope changes** | Task count at the start of the period vs current task count (additions/deletions) |

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
