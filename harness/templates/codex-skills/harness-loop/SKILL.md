---
name: harness-loop
description: "Use when running long-running Plans.md tasks in a Codex-native background loop — one cycle at a time, with status/stop controls. Do NOT load for: one-shot implementation (harness-work), review (harness-review), release (harness-release)."
description-en: "Use when running long-running Plans.md tasks in a Codex-native background loop — one cycle at a time, with status/stop controls. Do NOT load for: one-shot implementation (harness-work), review (harness-review), release (harness-release)."
allowed-tools: ["Read", "Bash"]
argument-hint: "[all|N-M] [--max-cycles N] [--pacing worker|ci|plateau|night]"
disable-model-invocation: true
---

# Harness Loop — Codex Native

The Codex version of `harness-loop` starts a **real background runner** rather than a simulated loop.

## In One Line

`$harness-loop` is the entry point for starting a "task supervisor" that continuously finds the next incomplete task, delegates it to Codex, checks the result, and advances to the next one — running in the background.

## Analogy

Instead of a person watching over every step, imagine a supervisor running in the background that repeats: "find the next task → hand it to Codex → check the result → move on."

## Quick Reference

| Input | Behavior |
|-------|----------|
| `$harness-loop all` | Start long-running loop for all incomplete tasks |
| `$harness-loop 41.1-41.4` | Start with a narrowed task range |
| `$harness-loop all --max-cycles 3` | Stop after maximum 3 cycles |
| `$harness-loop all --pacing night` | Use longer wait between cycles |
| `$harness-loop status` | Check current execution status |
| `$harness-loop stop` | Stop the running job and request loop termination |

## Execution Commands

### Start

```bash
harness codex-loop start all
```

With range:

```bash
harness codex-loop start 41.1-41.4 --max-cycles 5 --pacing worker
```

### Status

```bash
harness codex-loop status
harness codex-loop status --json
```

### Stop

```bash
harness codex-loop stop
```

## How It Works

1. Write execution state to `.claude/state/codex-loop/`
2. Find the next `cc:TODO` / `cc:WIP` task from Plans.md
3. Prepare using existing Harness assets such as `generate-sprint-contract.js`
4. Start Codex's actual work via `harness/scripts/codex-companion.sh task --background --write ...`
5. After job completion, perform review / checkpoint / plateau detection
6. If target tasks remain, wait and proceed to the next cycle

## Advisor Trigger Points

When advisor consultation is enabled (default), the loop calls `run-advisor-consultation.sh` at three trigger points:

1. **Pre-task**: When a task has `<!-- advisor:required -->` marker — calls with `reason_code=high_risk_preflight`
2. **Post-plateau**: When plateau is detected (`PIVOT_REQUIRED`) — calls with `reason_code=plateau_before_escalation`
3. **Pre-escalation**: Before surfacing any failure to the user — calls with `reason_code=pre_user_escalation`

Use `--no-advisor` to disable all advisor consultations.

## Pacing

| Value | Use case | Wait (seconds) |
|-------|----------|---------------|
| `worker` | Normal development loop | 270 |
| `ci` | When you want shorter intervals | 270 |
| `plateau` | Retry when hitting a plateau | 1200 |
| `night` | Long overnight run | 3600 |

## State Files

- `.claude/state/codex-loop/run.json`
- `.claude/state/codex-loop/cycles.jsonl`
- `.claude/state/codex-loop/runner.log`
- `.claude/state/codex-loop/current-job.json`
- `.claude/state/locks/codex-loop.lock.d`

## Important Notes

- This **actually runs in the background**. It does not just return an explanation.
- Two instances cannot run simultaneously. Returns `already running` if one is active.
- Rather than skipping failed tasks, the loop stops at the failure point and records the reason.
- `status` and `runner.log` make it easy to see where the loop is currently stuck.

## Example

"I want to automatically run the remaining tasks in Phase 41 throughout today":

```bash
harness codex-loop start 41.1-41.4 --max-cycles 8 --pacing worker
```

Check progress midway:

```bash
harness codex-loop status
```

Stop when done for the night:

```bash
harness codex-loop stop
```

## Why This Form

Codex cannot use the same wake-up mechanism as Claude's `/loop` directly.
Instead, by using **Codex companion's background jobs** as the foundation and having
Harness manage state and re-entry control, long-running tasks support "stop," "resume,"
and "check current state" naturally.

## Related Skills

- `$harness-work` — Single-task implementation used by each cycle
- `$harness-plan` — Plan tasks targeted by the loop
- `$harness-review` — Review individual tasks
