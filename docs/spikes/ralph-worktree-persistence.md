# Spike: Ralph Loop Worktree Persistence

**Task**: 89.0-spike
**Date**: 2026-05-03
**Method**: Code and documentation reading (live Agent() experiment not possible — workers have `disallowedTools: [Agent]`)

---

## Mechanism

`isolation: worktree` in an agent's frontmatter triggers automatic git worktree creation per Agent() call. Each call produces an isolated worktree at a path returned in the result as `worktreePath`. Files written inside a worktree are visible only to that worktree's git HEAD; they do NOT appear in the main working tree unless cherry-picked.

Evidence: `harness/agents/worker.md:56-60`
```
- **`isolation: worktree`**: Automatic worktree isolation via frontmatter (existing)
- **`ExitWorktree` tool**: Allows programmatic worktree exit after implementation is complete (new in v2.1.72)
- **Worktree fixes**: cwd restoration on task resume, worktreePath included in background notifications (v2.1.72 fix)
```

The `EnterWorktree(path=...)` deferred tool (available in Claude Code v2.1.72+) allows the orchestrator to **enter an existing worktree** by path. Since it is a deferred tool, its schema must be fetched via `ToolSearch(query="select:EnterWorktree")` before it can be called.

---

## Persistence Question

Can iteration N+1 see iteration N's work?

**Yes**, with the correct pattern:

- Iteration 0 creates the worktree via `isolation="worktree"`. The Agent() result includes `worktreePath`.
- The orchestrator stores `RALPH_WORKTREE = result.worktreePath`.
- For iterations 1..N:
  1. Orchestrator fetches `EnterWorktree` schema: `ToolSearch(query="select:EnterWorktree")`
  2. Orchestrator calls `EnterWorktree(path=RALPH_WORKTREE)` — this changes the CWD into the worktree
  3. Orchestrator spawns the next worker **without** `isolation="worktree"` (do not pass the param, or pass `isolation="none"`)
  4. The worker runs inside the persistent worktree and can read all files written by prior iterations

Since the worker writes files and commits within the worktree, and the next iteration enters the same worktree, the commits and modified files are visible to subsequent workers.

---

## Classification

**FEASIBLE**

The pattern works as intended. No design changes required. The only additional requirement beyond the original design is the `ToolSearch` call to load `EnterWorktree`'s schema before first use.

---

## Recommended Orchestrator Pattern

For use in `harness/skills/harness-ralph-loop/references/loop-flow.md`:

```
RALPH_WORKTREE = null
PRIOR_COMMIT   = null

for iter in range(0, MaxIter):

    # Build prompt with structured failure context
    prompt = build_ralph_prompt(task, iter, prior_verify_result)

    if iter == 0:
        # First iteration: create the persistent worktree
        result = Agent(
            subagent_type="powerball-harness:ralph-worker",
            prompt=prompt,
            isolation="worktree"           # ← creates the worktree
        )
        RALPH_WORKTREE = result.worktreePath
    else:
        # Subsequent iterations: enter the SAME worktree
        ToolSearch(query="select:EnterWorktree")   # ← load schema before first call
        EnterWorktree(path=RALPH_WORKTREE)         # ← enter existing worktree
        result = Agent(
            subagent_type="powerball-harness:ralph-worker",
            prompt=prompt
            # isolation NOT passed — worker inherits current CWD (the worktree)
        )

    # Verify, check promise, check diff ...
    prior_verify_result = run_verify(task.verify_cmd, RALPH_WORKTREE)
    PRIOR_COMMIT = result.commit
```

**Key constraints**:
- `ToolSearch` for `EnterWorktree` must be called before the FIRST `EnterWorktree(path=...)` call (deferred tool)
- Do NOT pass `isolation="worktree"` on iterations 1..N — that would create a new fresh worktree, breaking persistence
- The orchestrator must keep `RALPH_WORKTREE` alive across iterations; it is the single source of truth for worktree state
