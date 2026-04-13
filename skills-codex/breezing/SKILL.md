---
name: breezing
description: "Use when running breezing or full team-mode execution in Codex — alias for harness-work with team orchestration."
description-en: "Team execution mode (Codex native) — backward-compatible alias for harness-work with team orchestration using Codex native subagent API."
argument-hint: "[all|N-M|--max-workers N|--no-discuss]"
user-invocable: true
effort: high
---

# Breezing — Team Execution Mode (Codex Native)

> **This SKILL.md is the Codex CLI native version.**
> For the Claude Code version, see `skills/breezing/SKILL.md`.
> Subagent APIs use Codex's `spawn_agent` / `send_input` / `wait_agent` / `close_agent`.

**Backward-compatible alias**: Runs `harness-work --breezing` in team execution mode.

## Quick Reference

```bash
breezing                        # Ask for scope then execute
breezing all                    # Run all Plans.md tasks to completion
breezing 3-6                    # Run tasks 3-6 to completion
breezing --max-workers 2 all     # Limit simultaneous spawns of independent tasks to 2
breezing --no-discuss all       # Skip planning discussion and run all tasks to completion
```

## Options

| Option | Description | Default |
|--------|-------------|---------|
| `all` | Target all incomplete tasks | - |
| `N` or `N-M` | Task number / range specification | - |
| `--max-workers N` | Max simultaneous spawn count for independent tasks (breezing-specific option) | 1 (sequential) |
| `--no-commit` | Not supported (Breezing requires Worker temporary commits and Lead cherry-picks) | - |
| `--no-discuss` | Skip planning discussion | false |

## Execution

**This skill delegates to `harness-work --breezing`.** Execute with the following settings:

1. **Pass arguments to `harness-work --breezing`** (`--max-workers N` is interpreted as a breezing-specific option, distinct from `harness-work`'s `--parallel`)
2. **Force team execution mode** — Three-party separation: Lead → Worker spawn → companion review Reviewer
3. **Lead focuses on delegation** — Does not write code directly

### Differences from `harness-work`

| Feature | `harness-work` | `breezing` (this skill) |
|------|-----------------|------------------------|
| Default mode | Solo / Sequential | **Breezing (team execution)** |
| Parallelism | companion `task` Bash parallel | **Subagent delegation via `spawn_agent`** |
| Lead role | Coordination + implementation | **Delegate (coordination only)** |
| Review | Lead self-review | **Independent review via companion review** |
| Default scope | Next task | **All tasks** |

### Team Composition (Codex Native)

| Role | Execution method | Permissions | Responsibility |
|------|---------|------|------|
| Lead | (self) | Current session inherited | Coordination, direction, task distribution, cherry-pick |
| Worker ×N | `spawn_agent({message, fork_context})` | Session permissions inherited | Implementation (git worktree isolated) |
| Reviewer | companion `review --base` | read-only | Independent review |

## Flow Summary

```
breezing [scope] [--max-workers N] [--no-discuss]
    │
    ↓ Load harness-work --breezing
    │
Phase 0: Planning Discussion (skipped with --no-discuss)
Phase A: Pre-delegate (team initialization + worktree preparation)
Phase B: Delegate (Worker implementation + companion review)
Phase C: Post-delegate (integration verification + Plans.md update + commit)
```

### Phase 0: Planning Discussion (structured 3-question check)

Before executing all tasks, verify plan health with the following 3 questions.
All skipped when `--no-discuss` is specified.

**Q1. Scope confirmation**:
> "Will execute {{N}} tasks. Is the scope appropriate?"

**Q2. Dependency confirmation** (only when Plans.md has a Depends column):
> "Task {{X}} depends on {{Y}}. Is the execution order correct?"

**Q3. Risk flags** (only when there are `[needs-spike]` tasks):
> "Task {{Z}} is [needs-spike]. Do you want to spike first?"

### Phase A: Pre-delegate

1. Load Plans.md and identify target tasks
2. Analyze dependency graph and determine execution order
3. Create git worktree for each task

### Phase B: Delegate (Codex Native Subagent Orchestration)

```
for task in execution_order:
    # B-0. Isolate work directory
    worktree_path = "/tmp/worker-{task.number}-$$"
    branch_name = "worker-{task.number}-$$"
    git worktree add -b {branch_name} {worktree_path}
    TASK_BASE_REF = git rev-parse HEAD

    # B-1. Generate sprint-contract
    contract_path = bash("scripts/generate-sprint-contract.sh {task.number}")
    contract_path = bash("scripts/enrich-sprint-contract.sh {contract_path} --check \"Verify DoD from reviewer perspective\" --approve")
    bash("scripts/ensure-sprint-contract-ready.sh {contract_path}")

    # B-2. Worker spawn
    Plans.md: task.status = "cc:WIP"

    worker_id = spawn_agent({
        message: "Work in directory: {worktree_path}.\n\nTask: {task.content}\nDoD: {task.DoD}\ncontract_path: {contract_path}\n\nPlease implement. When done, git commit.\n\nWhen complete, return the following JSON:\n{\"commit\": \"<hash>\", \"files_changed\": [...], \"summary\": \"...\"}",
        fork_context: true
    })
    wait_agent({ ids: [worker_id] })

    # B-3. Lead runs review (from TASK_BASE_REF)
    # Use official plugin companion review (see harness-work "Review Loop"):
    #   bash scripts/codex-companion.sh review --base {TASK_BASE_REF}
    #   → verdict mapping: approve→APPROVE, needs-attention→REQUEST_CHANGES
    VERDICT = review_task(worktree_path, TASK_BASE_REF)  # static review (see harness-work)
    PROFILE = jq(contract_path, ".review.reviewer_profile")
    BROWSER_MODE = jq(contract_path, ".review.browser_mode // \"scripted\"")
    REVIEW_INPUT = "review-output.json"
    if PROFILE == "runtime":
        # Run runtime checks inside worktree
        REVIEW_INPUT = bash("cd {worktree_path} && scripts/run-contract-review-checks.sh {contract_path}")
        RUNTIME_VERDICT = jq(REVIEW_INPUT, ".verdict")
        if RUNTIME_VERDICT == "REQUEST_CHANGES":
            VERDICT = "REQUEST_CHANGES"
        elif RUNTIME_VERDICT == "DOWNGRADE_TO_STATIC":
            REVIEW_INPUT = "review-output.json"  # fall back to static review
    if PROFILE == "browser":
        # browser artifact is PENDING_BROWSER scaffold. reviewer agent runs it later.
        BROWSER_ARTIFACT = bash("scripts/generate-browser-review-artifact.sh {contract_path}")
        # REVIEW_INPUT stays as static review
    if REVIEW_INPUT != "review-output.json" and jq(REVIEW_INPUT, ".verdict") == "DOWNGRADE_TO_STATIC":
        REVIEW_INPUT = "review-output.json"
    bash("scripts/write-review-result.sh {REVIEW_INPUT} {commit_hash}")

    # B-4. Fix loop (on REQUEST_CHANGES, max 3 times)
    review_count = 0
    while VERDICT == "REQUEST_CHANGES" and review_count < 3:
        send_input({
            id: worker_id,
            message: "Issues found: {issues}\nFix them and run git commit --amend. Output JSON again after fixing."
        })
        wait_agent({ ids: [worker_id] })
        VERDICT = review_task(worktree_path, TASK_BASE_REF)
        review_count++

    # B-5. Terminate Worker
    close_agent({ id: worker_id })

    # B-6. Result handling
    if VERDICT == "APPROVE":
        commit_hash = git("-C", worktree_path, "rev-parse", "HEAD")
        git cherry-pick --no-commit {commit_hash}
        git commit -m "{task.content}"
        Plans.md: task.status = "cc:Done [{short_hash}]"
    else:
        → Escalate to user (Plans.md stays as cc:WIP)
        → Stop subsequent tasks as well

    # B-7. Worktree cleanup
    git worktree remove {worktree_path}
    git branch -D {branch_name}

    # B-8. Progress feed
    print("📊 Progress: Task {completed}/{total} done — {task.content}")
```

### Parallel spawn for independent tasks (when `--max-workers N` is specified)

When there are multiple tasks with no dependencies, control simultaneous spawn count with `--max-workers N`:

> **`wait_agent` semantics**: `wait_agent({ids: [a, b]})` returns the first one that completes (not all completions).
> Therefore, to wait for all Workers to complete, call `wait_agent` individually in a loop.

```
# Spawn independent tasks A, B in parallel (each with isolated worktree)
worker_a = spawn_agent({ message: "Work in directory: /tmp/worker-a-$$ ...", fork_context: true })
worker_b = spawn_agent({ message: "Work in directory: /tmp/worker-b-$$ ...", fork_context: true })

# Wait for each Worker individually → review → cherry-pick (sequential)
# wait_agent returns the first one, so remaining Workers are still running
for worker_id in [worker_a, worker_b]:
    wait_agent({ ids: [worker_id] })    # Wait for this Worker to complete
    VERDICT = review_task(worktree_path, TASK_BASE_REF)  # see harness-work
    # Fix loop (if needed)...
    close_agent({ id: worker_id })
    if VERDICT == "APPROVE":
        cherry-pick → update Plans.md
```

> **Constraint**: Only tasks with Depends set to `-` (independent tasks) can be parallelized.
> Review → cherry-pick runs sequentially (to avoid write conflicts on main).

### Worker Output Contract

The Worker prompt must explicitly specify that the following JSON is returned on completion:

```json
{
  "commit": "a1b2c3d",
  "files_changed": ["src/foo.ts", "tests/foo.test.ts"],
  "summary": "Add bar feature to foo module"
}
```

Lead parses this JSON to get the commit hash and file list.

### Progress Feed (progress notifications during Phase B)

```
📊 Progress: Task 1/5 done — "Add auto re-ticketing to harness-work"
📊 Progress: Task 2/5 done — "Add --snapshot to harness-sync"
```

### Completion Report (Phase C)

After all tasks complete, Lead generates a rich completion report with the following steps:

1. Collect all cherry-pick commits with `git log --oneline {session_base_ref}..HEAD`
2. Get overall change volume with `git diff --stat {session_base_ref}..HEAD`
3. Extract remaining tasks from Plans.md
4. Output following the Breezing template

## Differences from Claude Code Version

| Item | Claude Code version | Codex native version (this file) |
|------|---------------|-------------------------------|
| Worker spawn | `Agent(subagent_type="worker", isolation="worktree")` | `spawn_agent({message, fork_context})` + `git worktree add` |
| Wait for completion | `Agent` return value | `wait_agent({ids: [id]})` |
| Fix instruction | `SendMessage(to: agentId, message: "...")` | `send_input({id, message})` |
| Worker termination | Automatic | `close_agent({id})` |
| Review | Codex exec → Reviewer agent fallback | companion `review --base` (structured output) |
| Permissions | `bypassPermissions` + hooks | companion `task --write` / `spawn_agent`: inherits session permissions |
| Agent Teams | `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS` env var | Codex native (standard feature) |
| Worktree | `isolation="worktree"` auto-managed | `git worktree add/remove` manual management |
| Mode promotion | Auto for 4+ tasks | Only with explicit `--breezing` |

## Related Skills

- `harness-work` — From single task to team execution (main)
- `harness-sync` — Progress sync
- `harness-review` — Code review
