---
name: harness-loop
description: "Use when running Plans.md tasks in a continuous autonomous loop until all are done or advisor stops execution. Do NOT load for: single-task work (harness-work), planning (harness-plan), review (harness-review)."
argument-hint: "[N-iterations|--until-done|--advisor|--no-advisor|--max-failures N]"
user-invocable: true
effort: high
disable-model-invocation: true
---

# Harness Loop — Codex Native

> **This SKILL.md is the Codex CLI native version.**
> For the Claude Code version, see `skills/harness-loop/SKILL.md`.
> Subagent APIs use Codex's `spawn_agent` / `send_input` / `wait_agent` / `close_agent`.

Continuous autonomous loop that iterates over Plans.md tasks one by one, consulting an Advisor agent at configured trigger points, and runs until all tasks reach `cc:done` status or an explicit stop condition is met.

## Quick Reference

| User Input | Flag | Behavior |
|------------|------|----------|
| `harness-loop --until-done` | `--until-done` | Loop until all tasks are `cc:done` |
| `harness-loop 5` | `N` | Run at most 5 iterations |
| `harness-loop --no-advisor` | `--no-advisor` | Disable advisor consultation at all trigger points |
| `harness-loop --max-failures 3` | `--max-failures N` | Stop after 3 consecutive task failures |

> **Note**: In codex-native mode, an explicit argument is required. The interactive scope dialog
> from the Claude Code version is not available. Use `--until-done`, a numeric iteration count,
> or a task range instead.

## Loop Execution Model

Each iteration of the loop follows this sequence:

1. **Pick next task** — Find the next `cc:TODO` task in Plans.md (respecting `Depends` column ordering)
2. **Invoke harness-work** — Delegate implementation to `harness-work` in solo mode
3. **Update status** — Set task to `cc:done [hash]` on success, or increment failure streak on failure
4. **Check exit conditions** — Evaluate all exit conditions before starting the next iteration

### State File

The loop persists its state to `.claude/state/loop-active.json` so it can survive interruptions and provide accurate progress reporting:

```json
{
  "iteration": 4,
  "failure_streak": 0,
  "last_task": "3.2",
  "started_at": "2026-04-16T10:00:00Z",
  "config": {
    "max_iterations": null,
    "until_done": true,
    "max_failures": 3,
    "advisor_enabled": true
  }
}
```

On normal exit (all tasks done, advisor `STOP`, or user interrupt), the state file is removed. On abnormal exit it is left in place so the next run can report what happened.

## Advisor Trigger Points

When advisor consultation is enabled (default), the loop pauses and queries the Advisor agent at three trigger points. On a `STOP` response, the loop exits immediately and presents a summary to the user.

> **Codex-native mode**: CORRECTION responses are auto-applied without user confirmation.
> STOP responses cause the loop to exit cleanly, writing a final summary to stdout.

### 1. Pre-task Risk Check

**Trigger**: Before starting any task annotated with `<!-- advisor:required -->` in Plans.md.

**Reason code**: `high_risk_preflight`

**Behavior**: The Advisor reviews the task description, DoD, and current repo state. If it returns `PROCEED`, the task runs normally. If it returns `STOP`, the loop exits with a summary and explanation.

### 2. Post-failure Retry Gate

**Trigger**: When the same error signature has been observed on `retry_threshold` or more consecutive iterations (default threshold: 2).

**Reason code**: `repeated_failure`

**Behavior**: The loop presents the repeated failure pattern to the Advisor. The Advisor may return `PROCEED` (attempt once more with a different approach), `SKIP` (mark the task blocked and continue to the next), or `STOP` (exit and surface the failure to the user).

### 3. Plateau Detection

**Trigger**: A task has been restarted (i.e., returned to `cc:WIP`) without producing any new commits since its last attempt.

**Reason code**: `plateau_before_escalation`

**Behavior**: The Advisor is given the task content, the previous attempt's diff (empty), and the failure log. If it cannot resolve the plateau it returns `STOP`, which causes the loop to exit and escalate to the user with a full summary.

## Loop Exit Conditions

The loop terminates when any of the following conditions is met:

| Condition | Exit Type | Message |
|-----------|-----------|---------|
| All tasks are `cc:done` | Normal completion | "All tasks complete." |
| `--until-done` convergence: 0 remaining `cc:TODO` tasks | Normal completion | "Converged — no remaining tasks." |
| Advisor returns `STOP` | Advised stop | "Advisor requested stop: {reason}" |
| `--max-failures N` consecutive failures reached | Failure limit | "Stopped after {N} consecutive failures." |

On any exit, the loop writes a final summary showing: tasks completed this run, tasks remaining, failure count, and the exit reason.

## Differences from Claude Code Version

| Item | Claude Code version | Codex native version (this file) |
|------|---------------|-------------------------------|
| Scope dialog | Interactive prompt asking for scope | Not available — explicit argument required |
| Advisor CORRECTION | Pauses for user confirmation | Auto-applied without confirmation |
| Advisor STOP | Interactive summary presented to user | Loop exits cleanly, summary written to stdout |
| Worker spawn | `Task` tool (Agent) | `spawn_agent` / `wait_agent` |

## Related Skills

- `harness-work` — Single-task implementation engine invoked by each loop iteration
- `harness-plan` — Create or update Plans.md before starting the loop
- `harness-review` — Run a post-loop review after all tasks are complete
