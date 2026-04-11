---
name: breezing
description: "Use when user says 'breezing', 'do everything', 'run the whole plan', 'team run', 'full auto', or wants all tasks executed end-to-end with parallel workers. Do NOT load for: single-task implementation, planning, code review, release, or setup. Team execution mode — runs Plans.md tasks with full parallel team orchestration. High-level alias for harness-work in team mode."
allowed-tools: ["Read", "Write", "Edit", "Bash", "Grep", "Glob", "Task", "WebSearch"]
argument-hint: "[all|N-M|--codex|--parallel N|--no-commit|--no-discuss|--auto-mode]"
user-invocable: true
---

# Breezing — Team Execution Mode

> **Backward-compatible alias**: Runs `harness-work` in team execution mode.

## Quick Reference

```bash
breezing                        # Ask for scope, then execute
breezing all                    # Run all Plans.md tasks to completion
breezing 3-6                    # Run tasks 3 through 6 to completion
breezing --codex all            # Run all tasks to completion via Codex CLI
breezing --parallel 2 all       # Run all tasks to completion with 2 parallel workers
breezing --no-discuss all       # Run all tasks to completion, skipping planning discussion
breezing --auto-mode all        # Try Auto Mode rollout with a compatible parent session
```

## Options

| Option | Description | Default |
|--------|-------------|---------|
| `all` | Target all incomplete tasks | - |
| `N` or `N-M` | Specify task number/range | - |
| `--codex` | Delegate implementation to Codex CLI | false |
| `--parallel N` | Number of parallel Implementers | auto |
| `--no-commit` | Suppress automatic commits | false |
| `--no-discuss` | Skip planning discussion | false |
| `--auto-mode` | Explicitly enable Auto Mode rollout. Only considered when the parent session's permission mode is compatible | false |

## Execution

**This skill delegates to `harness-work`.** Execute `harness-work` with the following settings:

1. **Pass arguments directly to `harness-work`**
2. **Force team execution mode** — Three-way separation: Lead -> Worker spawn -> Reviewer spawn
3. **Lead focuses solely on delegation** — Does not write code directly
4. **Auto Mode is opt-in** — `--auto-mode` is accepted as a rollout flag for compatible parent sessions

### Differences from `harness-work`

| Characteristic | `harness-work` | `breezing` (this skill) |
|------|-----------------|------------------------|
| Parallelization approach | Automatic splitting based on need | **Lead/Worker/Reviewer role separation** |
| Lead's role | Coordination + implementation | **Delegate (coordination only)** |
| Review | Lead self-review | **Independent Reviewer** |
| Default scope | Next task | **All tasks** |

### Team Composition

| Role | Agent Type | Mode | Responsibility |
|------|-----------|------|------|
| Lead | (self) | - | Coordination, direction, task distribution |
| Worker xN | `claude-code-harness:worker` | `bypassPermissions` (current) / Auto Mode (follow-up)* | Implementation |
| Reviewer | `claude-code-harness:reviewer` | `bypassPermissions` (current) / Auto Mode (follow-up)* | Independent review |

> *If the parent session or frontmatter uses `bypassPermissions`, that takes priority. The distributed template currently uses `bypassPermissions`, so Auto Mode is a follow-up rollout target, not the default behavior.

### Codex Mode (`--codex`)

Mode that delegates all implementation to Codex CLI via the official plugin `codex-plugin-cc`:

```bash
# Task delegation (write-enabled)
bash scripts/codex-companion.sh task --write "task content"

# Via stdin (for large prompts)
CODEX_PROMPT=$(mktemp /tmp/codex-prompt-XXXXXX.md)
# Write task content
cat "$CODEX_PROMPT" | bash scripts/codex-companion.sh task --write
rm -f "$CODEX_PROMPT"
```

## Flow Summary

```
breezing [scope] [--codex] [--parallel N] [--no-discuss] [--auto-mode]
    |
    v Load harness-work with team mode
    |
Phase 0: Planning Discussion (skipped with --no-discuss)
Phase A: Pre-delegate (team initialization)
Phase B: Delegate (Worker implementation + Reviewer review)
Phase C: Post-delegate (integration verification + Plans.md update + commit)
```

### Progress Feed (progress notifications during Phase B)

The Lead outputs progress in the following format after each Worker task completion:

```
📊 Progress: Task {completed}/{total} done — "{task_subject}"
```

**Output example**:
```
📊 Progress: Task 1/5 done — "Add failure re-ticketing to harness-work"
📊 Progress: Task 2/5 done — "Add --snapshot to harness-sync"
📊 Progress: Task 3/5 done — "Add progress feed to breezing"
```

> **Design intent**: Breezing often runs for extended periods.
> This ensures users can glance at the terminal and immediately see how far along things are.
> The task-completed.sh hook outputs equivalent information via systemMessage, complementing the Lead's output.

### Review Policy (unified across all modes)

Even in Breezing mode, reviews follow the unified policy of **Codex exec preferred -> internal Reviewer fallback**.
See the "Review Loop" section in `harness-work` for details.

- Worker implements and commits in worktree -> returns result to Lead
- Lead reviews via Codex exec (120s timeout, fallback: Reviewer agent)
- REQUEST_CHANGES -> Lead sends fix instructions to Worker via SendMessage, Worker amends (up to 3 times)
- APPROVE -> **Lead** cherry-picks to main -> Updates Plans.md to `cc:done [{hash}]`

### Completion Report (Phase C — generated by Lead)

After all tasks are complete, the **Lead** generates a rich completion report with the following steps:

1. Collect all cherry-pick commits via `git log --oneline {base_ref}..HEAD`
2. Get overall change scale via `git diff --stat {base_ref}..HEAD`
3. Extract remaining `cc:TODO` / `cc:WIP` tasks from Plans.md
4. Output using the Breezing template from `harness-work`'s "Completion Report Format"

> **The generator is the Lead**. Not Workers or hooks. The Lead reads git + Plans.md during Phase C to generate this.

### Phase 0: Planning Discussion (structured 3-question check)

Before executing all tasks, verify plan soundness with the following 3 questions.
All skipped when `--no-discuss` is specified.

**Q1. Scope confirmation**:
> "We will execute {{N}} tasks. Is the scope appropriate?"

If too many, suggest narrowing by priority (Required > Recommended > Optional).

**Q2. Dependency confirmation** (only when Plans.md has a Depends column):
> "Task {{X}} depends on {{Y}}. Is the execution order correct?"

Read the Depends column and display the dependency chain. Error if circular dependencies exist.

**Q3. Risk flags** (only when `[needs-spike]` tasks exist):
> "Task {{Z}} is marked [needs-spike]. Should we spike first?"

If there are incomplete `[needs-spike]` tasks, confirm whether to run the spike first.

If all 3 questions pass without issues, proceed to Phase A (designed to complete in 30 seconds total).

### Dependency Graph-Based Task Assignment

When Plans.md has a Depends column (v2 format), execute tasks according to the dependency graph:

1. **Execute tasks with Depends = `-`** first. If multiple independent tasks exist, they can be spawned in parallel
2. After each Worker completes, Lead reviews -> cherry-picks (see harness-work Phase B)
3. Once a dependency source task is cherry-picked to main, execute tasks that depended on it next
4. Repeat until all tasks are complete

> **Note**: The "Worker completion -> review -> cherry-pick" sequence for each task is sequential.
> Only the Worker spawn portion of independent tasks (Depends = `-`) can be parallelized.

## Codex Native Orchestration

Codex uses native subagents.
The primary control surfaces are `spawn_agent`, `wait`, `send_input`, `resume_agent`, `close_agent`.

> **Claude Code vs Codex communication API** (SSOT: API mapping table in `team-composition.md`):
> - Claude Code: `SendMessage(to: agentId, message: "...")` to send fix instructions to Worker
> - Codex: `resume_agent(agent_id)` to resume Worker -> `send_input(agent_id, "...")` to send instructions
>
> Pseudocode in harness-work is written in Claude Code syntax. In Codex environments, translate to the above.

## Related Skills

- `harness-work` — Single task to team execution (main implementation)
- `harness-sync` — Progress synchronization
- `harness-review` — Code review (auto-launched within breezing)
