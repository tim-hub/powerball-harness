# Harness Sync — Detailed Steps

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

Save to `.claude/state/snapshots/` as a timestamped JSON:

```bash
SNAPSHOT_DIR="${PROJECT_ROOT}/.claude/state/snapshots"
mkdir -p "${SNAPSHOT_DIR}"
SNAPSHOT_FILE="${SNAPSHOT_DIR}/progress-$(date -u +%Y%m%dT%H%M%SZ).json"
```

Snapshot schema:

```json
{
  "timestamp": "2026-03-08T10:30:00Z",
  "phase": "Phase 26",
  "progress": { "total": 16, "todo": 5, "wip": 3, "done": 6, "confirmed": 2 },
  "progress_rate": 50,
  "recent_commits": ["abc1234 feat: ..."],
  "recent_files": ["skills/harness-work/SKILL.md"],
  "notes": ""
}
```

When a previous snapshot exists, display a diff table (Previous vs Current: progress rate, done count, WIP count).

## Step 5: Next Action Proposal

```
Next steps

**Priority 1**: {{task}}
- Reason: {{in progress / waiting to unblock}}

**Recommended**: harness-work, harness-review
```

## Step 6: Retrospective (ON by default, skip with `--no-retro`)

Runs automatically when ≥ 1 `cc:Done` task exists.

### R1: Collect Completed Tasks

```bash
grep -E 'cc:Done|pm:confirmed' Plans.md
git log --oneline --since="7 days ago"
git diff --stat HEAD~10
```

### R2: Four Retrospective Items

| Item | Analysis Method |
|------|-----------------|
| **Estimation Accuracy** | Expected file count (from task descriptions) vs actual from `git diff --stat` |
| **Block Causes** | Aggregate reason patterns for `blocked` markers |
| **Quality Marker Accuracy** | Check whether `[feature:security]` tasks actually hit related issues |
| **Scope Changes** | Compare initial Plans.md task count vs current |

### R3: Retrospective Summary Output

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

### R4: Record to harness-mem

Record retrospective results to harness-mem for reference in future `create` invocations.
Storage: `.claude/agent-memory/` under the corresponding agent memory key.
