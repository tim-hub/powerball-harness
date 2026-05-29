---
name: harness-schedule-run
description: "Runs plans.json tasks in a long-running autonomous loop with ScheduleWakeup, sprint contracts, and plateau detection. Use when running tasks overnight or on a scheduled cadence."
when_to_use: "autonomous loop, overnight run, scheduled loop, continuous execution, long-running loop"
allowed-tools: ["Read", "Edit", "Bash", "Task", "ScheduleWakeup", "mcp__harness__harness_mem_resume_pack", "mcp__harness__harness_mem_record_checkpoint"]
argument-hint: "[all|N-M|--max-cycles N|--pacing worker|ci|plateau|night|--advisor|--no-advisor]"
model: sonnet
effort: medium
---

# Harness Schedule Run

Meta-skill that combines `/loop` (CC dynamic mode) with `ScheduleWakeup` to re-enter long-running tasks with a **fresh context on every wake-up**.

Each wake-up calls `harness-work --breezing` via the Agent tool — 1 cycle = 1 task completion.

> **Renamed in v4.x**: previously `harness-schedule-run`. The slash command and skill identifier are now `/harness-schedule-run`.

## Quick Reference

| Input | Behavior |
|-------|----------|
| `/harness-schedule-run all` | Loop all incomplete tasks (default: max 8 cycles) |
| `/harness-schedule-run all --max-cycles 3` | Stop after 3 cycles |
| `/harness-schedule-run 41.1-41.3 --pacing ci` | Execute task range with CI pacing |
| `/harness-schedule-run all --pacing night` | Overnight batch (3600s interval) |
| `/harness-schedule-run --no-advisor` | Disable advisor consultation at all trigger points |

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

## Launch Flow

Full per-wake-up step walkthrough (Steps 0–9: concurrency guard → state check → plans.json read → sprint-contract → resume-pack → worker execution → review → plateau detection → cycle count → checkpoint → next wake-up):

[`${CLAUDE_SKILL_DIR}/references/flow.md`](${CLAUDE_SKILL_DIR}/references/flow.md)

## Cycle Stop Conditions

| Condition | Stop Type | Response |
|-----------|-----------|----------|
| `cycles >= max_cycles` | Normal stop (limit reached) | Report to user |
| `PIVOT_REQUIRED` (exit 2) | Abnormal stop (escalation) | Ask user for decision |
| No incomplete tasks | Normal stop (all complete) | Output completion report |

## /loop Integration

When `/loop` is enabled, CC continues autonomous re-entry via `ScheduleWakeup` at the end of each cycle. Each wake-up starts with a **fresh context**; `harness-mem resume-pack` (Step 4 in flow.md) reloads context.

`/loop` sentinel: `<<autonomous-loop-dynamic>>`

## Checkpoint Schema

```json
{
  "session_id": "<session ID>",
  "title": "harness-schedule-run cycle {N}/{max}: {task name}",
  "content": "1-line summary of cycle_result + commit hash"
}
```

## Advisor Integration

When enabled (default: on, disable with `--no-advisor`), the advisor is called at three trigger points — pre-task risk check (`high_risk_preflight`), post-plateau (`plateau_before_escalation`), and pre-escalation (`pre_user_escalation`). On a `STOP` response the loop exits.

Full bash call patterns and trigger conditions: [`${CLAUDE_SKILL_DIR}/references/flow.md`](${CLAUDE_SKILL_DIR}/references/flow.md)

## Related Skills

- `harness-work` — Task implementation skill executed each cycle
- `harness-plan` — Plan tasks targeted by the loop
- `harness-review` — Review individual tasks
- `session-control` — Session state management
