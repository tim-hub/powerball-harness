---
name: harness-loop
description: "Runs Plans.md tasks in a long-running autonomous loop with ScheduleWakeup, sprint contracts, and plateau detection. Use when running tasks overnight or in a continuous loop."
when_to_use: "autonomous loop, overnight run, scheduled loop, continuous execution, long-running loop"
allowed-tools: ["Read", "Edit", "Bash", "Task", "ScheduleWakeup", "mcp__harness__harness_mem_resume_pack", "mcp__harness__harness_mem_record_checkpoint"]
argument-hint: "[all|N-M|--max-cycles N|--pacing worker|ci|plateau|night|--advisor|--no-advisor]"
model: sonnet
effort: medium
---

# Harness Loop

Meta-skill that combines `/loop` (CC dynamic mode) with `ScheduleWakeup` to re-enter long-running tasks with a **fresh context on every wake-up**.

Each wake-up calls `harness-work --breezing` via the Agent tool, forming a re-entrant loop of 1 cycle = 1 task completion.

## Quick Reference

| Input | Behavior |
|-------|----------|
| `/harness-loop all` | Loop all incomplete tasks (default: max 8 cycles) |
| `/harness-loop all --max-cycles 3` | Stop after 3 cycles |
| `/harness-loop 41.1-41.3 --pacing ci` | Execute task range with CI pacing |
| `/harness-loop all --pacing night` | Overnight batch (3600s interval) |
| `/harness-loop --no-advisor` | Disable advisor consultation at all trigger points |

## Options

| Option | Description | Default |
|--------|-------------|---------|
| `all` | Target all incomplete tasks | - |
| `N-M` | Task number range | - |
| `--max-cycles N` | Maximum cycle count | `8` |
| `--pacing <mode>` | Wake-up interval mode | `worker` (270s) |
| `--advisor` | Enable advisor consultation (default) | enabled |
| `--no-advisor` | Disable advisor consultation at all trigger points | - |

### Pacing Values

| pacing | delaySeconds | Use case |
|--------|-------------|----------|
| `worker` | 270 | Immediately after Worker completion (within 5 min cache warm) |
| `ci` | 270 | Waiting for short CI jobs |
| `plateau` | 1200 | 20 min (retry interval after plateau detection) |
| `night` | 3600 | Long overnight batch |

> **Constraint**: `ScheduleWakeup`'s `delaySeconds` is clamped to **[60, 3600]** at runtime.
> All pacing values are within this range. When specifying values directly, always use 60–3600.

## Launch Flow (per wake-up entry)

Full details: [`${CLAUDE_SKILL_DIR}/references/flow.md`](${CLAUDE_SKILL_DIR}/references/flow.md)

```
wake-up
  │
  ▼
[Step 0] Flock-based concurrency guard
  Prevent concurrent loop instances via lock directory
  │
  ▼
[Step 0.5] State consistency check
  bash tests/validate-plugin.sh --quick
  │
  ▼
[Step 1] Read Plans.md first
  Identify the leading cc:WIP / cc:TODO task (get task_id)
  No incomplete tasks → loop ends (normal completion)
  │
  ▼
[Step 2] Check sprint-contract existence & generate
  Check .claude/state/contracts/${task_id}.sprint-contract.json
  If absent: node harness/scripts/generate-sprint-contract.js ${task_id}
  On first generation: bash harness/scripts/enrich-sprint-contract.sh <contract-path> \
    --check "auto-approve (harness-loop — confirm DoD from reviewer perspective)" \
    --approve  ← draft → approved
  (Existing contracts already approved — skip)
  │
  ▼
[Step 3] Contract readiness check
  bash harness/scripts/ensure-sprint-contract-ready.sh <contract-path>
  │
  ▼
[Step 4] Resume pack reload
  harness-mem resume-pack (context re-injection)
  │
  ▼
[Step 5] Execute 1 task cycle
  worker_result = Agent(
      subagent_type="claude-code-harness:worker",
      prompt="Task: ${task_id}\nDoD: <extracted from Plans.md>\ncontract_path: ${CONTRACT_PATH}\nmode: breezing",
      isolation="worktree",
      run_in_background=false
  )
  # worker_result: { commit, branch, worktreePath, files_changed, summary }
  │
  ▼
[Step 5.5] Lead review execution
  diff_text = git show worker_result.commit
  verdict = codex_exec_review(diff_text) or reviewer_agent_review(diff_text)
  See flow.md for details
  │
  ▼
[Step 5.6] APPROVE → cherry-pick to main / REQUEST_CHANGES → fix loop (max_iterations from contract, default 3)
  APPROVE: git cherry-pick → update Plans.md to cc:Done [{hash}] → delete feature branch
  REQUEST_CHANGES x MAX_REVIEWS still rejected: escalation
  See flow.md for details
  │
  ▼
[Step 6] Plateau detection
  bash harness/scripts/detect-review-plateau.sh ${current_task_id}
  │
  ├── PIVOT_REQUIRED (exit 2)   → loop stop + user escalation (advisor called if enabled)
  ├── INSUFFICIENT_DATA (exit 1) → continue
  └── PIVOT_NOT_REQUIRED (exit 0) → continue
  │
  ▼
[Step 7] Cycle count check
  │
  ├── cycles >= max_cycles → loop stop (limit reached)
  │
  ▼
[Step 8] Record checkpoint
  harness_mem_record_checkpoint(
      session_id, title, content=cycle result summary
  )
  │
  ▼
[Step 9] Schedule next wake-up
  ScheduleWakeup(
      delaySeconds=<pacing value>,
      prompt="/harness-loop <same args>",
      reason="Cycle {N}/{max} complete — proceeding to next task"
  )
```

## Cycle Stop Conditions

| Condition | Stop Type | Response |
|-----------|-----------|----------|
| `cycles >= max_cycles` | Normal stop (limit reached) | Report to user |
| `PIVOT_REQUIRED` (exit 2) | Abnormal stop (escalation) | Ask user for decision |
| No incomplete tasks | Normal stop (all complete) | Output completion report |

With `--max-cycles 3`, stops after 3 completed cycles.
Default (`--max-cycles 8`) stops after 8 cycles.

## /loop Integration

This skill is used in combination with CC's `/loop` (dynamic mode).

When `/loop` is enabled, CC continues autonomous re-entry, scheduling the next wake-up via `ScheduleWakeup` at the end of each cycle.

`/loop` sentinel: `<<autonomous-loop-dynamic>>`

Each wake-up starts with a **fresh context**, preventing context contamination from the previous cycle.
Reload via `harness-mem resume-pack` is required (Step 4).

## Checkpoint Schema

```json
{
  "session_id": "<session ID>",
  "title": "harness-loop cycle {N}/{max}: {task name}",
  "content": "1-line summary of cycle_result + commit hash"
}
```

## Advisor Integration

When advisor consultation is enabled (default: on, disable with `--no-advisor`), the loop pauses and calls `run-advisor-consultation.sh` at three trigger points. On a `STOP` response the loop exits immediately with a summary.

### Trigger Point 1: Pre-task Risk Check

**Trigger**: Before starting any task annotated with `<!-- advisor:required -->` in Plans.md.

**Reason code**: `high_risk_preflight`

**Script call**:
```bash
bash harness/scripts/run-advisor-consultation.sh \
  --reason-code high_risk_preflight \
  --task-id "${task_id}"
```

**Behavior**: The Advisor reviews the task description, DoD, and current repo state. Returns `PLAN` to proceed or `STOP` to exit with explanation.

### Trigger Point 2: Post-plateau (PIVOT_REQUIRED)

**Trigger**: When `detect-review-plateau.sh` returns exit 2 (`PIVOT_REQUIRED`).

**Reason code**: `plateau_before_escalation`

**Script call**:
```bash
bash harness/scripts/run-advisor-consultation.sh \
  --reason-code plateau_before_escalation \
  --task-id "${task_id}"
```

**Behavior**: The Advisor receives plateau details. Returns `PLAN` (retry with different approach) or `STOP` (escalate to user).

### Trigger Point 3: Pre-escalation

**Trigger**: Before surfacing any STOP/failure condition to the user.

**Reason code**: `pre_user_escalation`

**Script call**:
```bash
bash harness/scripts/run-advisor-consultation.sh \
  --reason-code pre_user_escalation \
  --task-id "${task_id}"
```

**Behavior**: Final check before user involvement. Advisor may provide a resolution path or confirm escalation is necessary.

## Related Skills

- `harness-work` — Task implementation skill executed each cycle
- `harness-plan` — Plan tasks targeted by the loop
- `harness-review` — Review individual tasks
- `session-control` — Session state management
