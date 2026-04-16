---
name: breezing
description: "Use when running the full team/breezing flow end-to-end — all tasks with parallel workers. Do NOT load for: single-task implementation, planning, review, release, or setup."
allowed-tools: ["Read", "Write", "Edit", "Bash", "Grep", "Glob", "Task", "WebSearch"]
argument-hint: "[all|N-M|--codex|--parallel N|--no-commit|--no-discuss|--auto-mode|--advisor|--no-advisor]"
user-invocable: true
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
| Worker xN | `claude-code-harness:worker` | `bypassPermissions` (current) / Auto Mode (follow-up)* | Implementation |
| Reviewer | `claude-code-harness:reviewer` | `bypassPermissions` (current) / Auto Mode (follow-up)* | Independent review |

> *If the parent session or frontmatter specifies `bypassPermissions`, that takes precedence. The distributed template currently uses `bypassPermissions`, so Auto Mode is a follow-up rollout target and not the default behavior.

## Advisor Integration

When `advisor.enabled: true` in config (or `--advisor` flag), the Lead checks for advisor consultation at two points:
- **Pre-spawn**: tasks marked `<!-- advisor:required -->` trigger a preflight consultation before Worker spawn
- **Post-STOP**: if a Worker signals STOP, Lead consults the Advisor before escalating to the user

Use `--no-advisor` to bypass and escalate directly.

### Codex Mode (`--codex`)

A mode that delegates all implementation to Codex CLI via the official plugin `codex-plugin-cc`:

```bash
# Task delegation (writable)
bash "${CLAUDE_SKILL_DIR}/../../scripts/codex-companion.sh" task --write "task content"

# Via stdin (for large prompts)
CODEX_PROMPT=$(mktemp /tmp/codex-prompt-XXXXXX.md)
# Write task content
cat "$CODEX_PROMPT" | bash "${CLAUDE_SKILL_DIR}/../../scripts/codex-companion.sh" task --write
rm -f "$CODEX_PROMPT"
```

## Flow Summary

```
breezing [scope] [--codex] [--parallel N] [--no-discuss] [--auto-mode]
    │
    ↓ Load harness-work with team mode
    │
Phase 0: Planning Discussion (skipped with --no-discuss)
Phase A: Pre-delegate (team initialization)
Phase B: Delegate (Worker implementation + Reviewer review)
Phase C: Post-delegate (integration verification + Plans.md update + commit)
```

### Progress Feed (progress notifications during Phase B)

The Lead outputs progress in the following format each time a Worker completes a task:

```
📊 Progress: Task {completed}/{total} done — "{task_subject}"
```

**Example output**:
```
📊 Progress: Task 1/5 done — "Add failure re-ticketing to harness-work"
📊 Progress: Task 2/5 done — "Add --snapshot to harness-sync"
📊 Progress: Task 3/5 done — "Add progress feed to breezing"
```

> **Design intent**: Breezing often involves long-running execution.
> This allows users to see "how far along things are" at a glance when checking the terminal.
> The task-completed.sh hook outputs equivalent information via systemMessage, complementing the Lead's output.

### Review Policy (unified across all modes)

Even in Breezing mode, reviews follow the unified policy of **Codex exec first → internal Reviewer fallback**.
See the "Review Loop" section of `harness-work` for details.

- Worker implements and commits within the worktree → returns results to Lead
- Lead reviews via Codex exec (120s timeout, fallback: Reviewer agent)
- REQUEST_CHANGES → Lead sends fix instructions to Worker via SendMessage, Worker amends (up to 3 times)
- APPROVE → **Lead** cherry-picks to main → updates Plans.md to `cc:done [{hash}]`

### Completion Report (Phase C — generated by Lead)

After all tasks are complete, the **Lead** generates a rich completion report with the following steps:

1. Collect all cherry-pick commits with `git log --oneline {base_ref}..HEAD`
2. Get the overall change scope with `git diff --stat {base_ref}..HEAD`
3. Extract remaining `cc:TODO` / `cc:WIP` tasks from Plans.md
4. Output according to the Breezing template in `harness-work`'s "Completion Report Format"

> **The Lead generates this report**, not Workers or hooks. The Lead reads git + Plans.md during Phase C to produce it.

### Phase 0: Planning Discussion (structured 3-question check)

Before executing all tasks, verify plan health with the following 3 questions.
All are skipped when `--no-discuss` is specified.

**Q1. Scope confirmation**:
> "Executing {{N}} tasks. Is the scope appropriate?"

If too many, suggest narrowing by priority (Required > Recommended > Optional).

**Q2. Dependency confirmation** (only when Plans.md has a Depends column):
> "Task {{X}} depends on {{Y}}. Is the execution order correct?"

Read the Depends column and display the dependency chain. Error if circular dependencies exist.

**Q3. Risk flag** (only when `[needs-spike]` tasks exist):
> "Task {{Z}} is [needs-spike]. Should we spike it first?"

If there are incomplete `[needs-spike]` tasks, confirm whether to run the spike first.

If all 3 questions pass, proceed to Phase A (designed to complete in 30 seconds total).

### Task Assignment Based on Dependency Graph

When Plans.md has a Depends column (v2 format), tasks are executed following the dependency graph:

1. Execute **tasks with Depends set to `-`** first. If multiple independent tasks exist, they can be spawned in parallel
2. After each Worker completes, Lead reviews → cherry-picks (see harness-work Phase B)
3. Once a dependency source task is cherry-picked to main, execute tasks that depended on it next
4. Repeat until all tasks are complete

> **Note**: The "Worker complete → review → cherry-pick" cycle for each task is sequential.
> Only the Worker spawn portion of independent tasks (Depends is `-`) can be parallelized.

## Codex Native Orchestration

Codex uses native subagents.
Key control surfaces are `spawn_agent`, `wait`, `send_input`, `resume_agent`, `close_agent`.

> **Claude Code vs Codex communication API** (SSOT: API mapping table in `team-composition.md`):
> - Claude Code: `SendMessage(to: agentId, message: "...")` to send fix instructions to Workers
> - Codex: `resume_agent(agent_id)` to resume Workers → `send_input(agent_id, "...")` to send instructions
>
> Pseudo-code in harness-work is written in Claude Code syntax. Translate to the above when running in a Codex environment.

## Related Skills

- `harness-work` — From single tasks to team execution (core)
- `harness-sync` — Progress synchronization
- `harness-review` — Code review (auto-triggered within breezing)
