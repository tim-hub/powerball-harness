---
name: breezing
description: "Full team execution with Lead/Worker/Reviewer agents running all Plans.md tasks end-to-end. Use when running the complete breezing flow with parallel workers."
when_to_use: "full team run, run all tasks with team, parallel team execution, end-to-end execution"
allowed-tools: ["Read", "Write", "Edit", "Bash", "Grep", "Glob", "Task", "WebSearch"]
argument-hint: "[all|N-M|--codex|--parallel N|--no-commit|--no-discuss|--auto-mode|--advisor|--no-advisor]"
user-invocable: true
model: sonnet
---

# Breezing — Team Execution Mode

> **Backward-compatible alias**: Runs `harness-work` in team execution mode.

## Quick Reference

| User Input | Subcommand | Behavior |
|------------|------------|----------|
| `breezing` | _(none)_ | Ask for scope before executing |
| `breezing all` | `all` | Complete all tasks in Plans.md |
| `breezing 3-6` | `N-M` | Complete tasks 3 through 6 |
| `breezing --codex all` | `--codex` | Complete all tasks via Codex CLI |
| `breezing --parallel 2 all` | `--parallel N` | Complete all tasks with 2 parallel workers |
| `breezing --no-commit all` | `--no-commit` | Complete all tasks, suppress automatic commits |
| `breezing --no-discuss all` | `--no-discuss` | Complete all tasks, skipping planning discussion |
| `breezing --auto-mode all` | `--auto-mode` | Try Auto Mode rollout on a compatible parent session |

## Options

| Option | Description | Default |
|--------|-------------|---------|
| `all` | Target all incomplete tasks | - |
| `N` or `N-M` | Task number/range specification | - |
| `--codex` | Delegate implementation to Codex CLI | false |
| `--parallel N` | Number of parallel Implementers | auto |
| `--no-commit` | Suppress automatic commits | false |
| `--no-discuss` | Skip planning discussion | false |
| `--auto-mode` | Explicitly opt in to Auto Mode rollout. Only considered when the parent session's permission mode is compatible | false |
| `--advisor` | Enable advisor consultation at risk/failure trigger points | from config |
| `--no-advisor` | Disable advisor; escalate directly to user | false |

## Execution

**This skill delegates to `harness-work`.** Run `harness-work` with the following settings:

1. **Pass arguments directly to `harness-work`**
2. **Force team execution mode** — Three-way separation: Lead → Worker spawn → Reviewer spawn
3. **Lead focuses on delegation only** — Does not write code directly
4. **Auto Mode is opt-in** — `--auto-mode` is accepted as a rollout flag for compatible parent sessions

### Differences from `harness-work`

| Aspect | `harness-work` | `breezing` (this skill) |
|--------|-----------------|------------------------|
| Parallelization | Automatic splitting based on need | **Lead/Worker/Reviewer role separation** |
| Lead's role | Coordination + implementation | **Delegation only (coordination focused)** |
| Review | Lead self-review | **Independent Reviewer** |
| Default scope | Next task | **All tasks** |

### Team Composition

| Role | Agent Type | Mode | Responsibility |
|------|-----------|------|----------------|
| Lead | (self) | - | Coordination, command, task distribution |
| Worker xN | `harness:worker` | `bypassPermissions` | Implementation |
| Reviewer | `harness:reviewer` | `bypassPermissions` | Independent review |

## Advisor Integration

When `advisor.enabled: true` in config (or `--advisor` flag), the Lead checks for advisor consultation at two points:
- **Pre-spawn**: tasks marked `<!-- advisor:required -->` trigger a preflight consultation before Worker spawn
- **Post-STOP**: if a Worker signals STOP, Lead consults the Advisor before escalating to the user

### Codex Mode (`--codex`)

Delegates implementation to Codex CLI via `codex-plugin-cc`.
Bash invocation snippets and Codex native orchestration API: [references/codex-orchestration.md](${CLAUDE_SKILL_DIR}/references/codex-orchestration.md)

## Flow Summary

```
breezing [scope] [--codex] [--parallel N] [--no-discuss] [--auto-mode]
    │
    ↓ Load harness-work with team mode
    │
Phase 0: Planning Discussion (skipped with --no-discuss) — [details](${CLAUDE_SKILL_DIR}/references/planning-discussion.md)
Phase A: Pre-delegate (team initialization)
Phase B: Delegate (Worker implementation + Reviewer review)
Phase C: Post-delegate (integration verification + Plans.md update + commit)
```

### Progress Feed (Phase B notifications)

Output per completed task:
```
📊 Progress: Task {completed}/{total} done — "{task_subject}"
```
The `task-completed.sh` hook also outputs equivalent information via systemMessage.

### Review Policy

Codex exec first → internal Reviewer fallback (unified across all modes).
See [`harness-work/references/review-loop.md`](${CLAUDE_SKILL_DIR}/../../harness-work/references/review-loop.md).
APPROVE → Lead cherry-picks to main → `cc:done [{hash}]` in Plans.md.

### Completion Report (Phase C)

Lead collects cherry-pick commits, diff stat, and remaining tasks, then outputs per
[`harness-work/templates/completion-report.md`](${CLAUDE_SKILL_DIR}/../../harness-work/templates/completion-report.md).

## Related Skills

- `harness-work` — From single tasks to team execution (core)
- `harness-sync` — Progress synchronization
- `harness-review` — Code review (auto-triggered within breezing)
