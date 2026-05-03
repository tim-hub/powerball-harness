---
name: harness-ralph-loop
description: "Iterates a fresh ralph-worker subagent per attempt in a persistent worktree until <promise> tag and verify command both confirm success. Use when running [ralph]-marked tasks, iterate-until-pass loops, or stub-fail mode tests."
when_to_use: "ralph task, loop until pass, iterate until tests pass, ralph mode, run ralph loop, [ralph] task"
allowed-tools: ["Read", "Write", "Edit", "Bash", "Task", "Skill"]
argument-hint: "[task-id|help|--max-iter N|--verify CMD|--stub-fail|--stub-stuck|--stub-mismatch]"
model: sonnet
effort: high
---

# Harness Ralph Loop

Iterative-loop orchestrator for `[ralph]`-marked Plans.md tasks.
Spawns a fresh `claude-code-harness:ralph-worker` subagent per attempt inside a single persistent worktree,
runs the Verify command authoritatively, and loops until both the `<promise>` text tag and verify exit 0 agree.

## Quick Reference

| Subcommand / Argument | Behavior |
|---|---|
| `help` | Print usage, philosophy, and link to prompt-best-practices.md |
| `<task-id>` | Run the loop for a Plans.md task that has `[ralph]` marker |
| `--max-iter N` | Override MaxIter (default: 10) |
| `--verify CMD` | Override the Verify command from the task |
| `--stub-fail` | Test mode: stub a failing verify (smoke test for FT-RALPH-03) |
| `--stub-stuck` | Test mode: stub a zero-change worker (smoke test for FT-RALPH-01) |
| `--stub-mismatch` | Test mode: stub a verify-mismatch worker (smoke test for FT-RALPH-02) |

## Execution Flow

See [`references/loop-flow.md`](${CLAUDE_SKILL_DIR}/references/loop-flow.md) for the complete orchestrator pseudocode.

Summary: On iteration 0 the orchestrator spawns a `ralph-worker` with `isolation="worktree"` to create the persistent worktree, capturing `RALPH_WORKTREE` from `result.worktreePath`. For iterations 1..N it fetches the `EnterWorktree` deferred-tool schema via `ToolSearch`, calls `EnterWorktree(path=RALPH_WORKTREE)`, then spawns the next worker without the isolation param. After each iteration the orchestrator runs the Verify command itself (anti-tamper), compares the exit code against `ralph-worker-report.v1.verify.exit_code`, scans the final assistant message for `<promise>…</promise>`, and diffs the worktree to detect stuck iterations. The loop terminates on success or on any of the three hard-stop failure modes (FT-RALPH-01 STUCK, FT-RALPH-02 VERIFY-MISMATCH, FT-RALPH-03 MAX-ITER).

## When to Use Ralph

See [`references/when-to-ralph.md`](${CLAUDE_SKILL_DIR}/references/when-to-ralph.md) for guidance on which tasks are Ralph-suitable and which are not.

## Reference Files

| File | Purpose |
|------|---------|
| [`references/loop-flow.md`](${CLAUDE_SKILL_DIR}/references/loop-flow.md) | Full orchestrator pseudocode, prompt template, decision matrix, terminal-state handling |
| [`references/when-to-ralph.md`](${CLAUDE_SKILL_DIR}/references/when-to-ralph.md) | When to use Ralph (task patterns, DoD quality checks) |

## Related Skills and Agents

- `harness-plan` — Emits `[ralph]` marker + `Verify:` line for suitable tasks
- `harness-work` — Detects `[ralph]` marker and routes to this skill
- `claude-code-harness:ralph-worker` — The per-iteration worker agent this skill dispatches
