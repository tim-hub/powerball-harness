# Prompt Writing Best Practices for Ralph Loop

This guide helps you write prompts that work well with the harness-ralph-loop iterative execution model.
A well-structured prompt dramatically improves convergence speed and reduces wasted iterations.

---

## Section A: Prompt Writing Best Practices

### A1. Clear Completion Criteria

The most common cause of stalled Ralph loops is a vague or missing completion criterion.
Ralph's orchestrator matches an **exact** `<promise>...</promise>` tag emitted by the worker,
and the Verify command exit code must be 0. Both must agree before the loop terminates.

**BAD — no measurable criteria:**

```
Build a todo API and make it good.
```

The worker has no idea what "good" means. It may produce code, emit `<promise>COMPLETE</promise>`,
yet fail the Verify command — triggering FT-RALPH-02 VERIFY-MISMATCH on every iteration.

**GOOD — explicit success criteria with `<promise>` requirement:**

```
Build a REST API for todos.

When complete:
- All CRUD endpoints working (GET /todos, POST /todos, PUT /todos/:id, DELETE /todos/:id)
- Input validation in place (400 on missing fields)
- Tests passing (coverage > 80%)
- README with API docs

Output exactly: <promise>COMPLETE</promise>
```

The worker now has an unambiguous checklist. Each iteration narrows the gap. The orchestrator
can verify the promise tag and run the Verify command independently.

---

### A2. Incremental Goals

Ralph is not a "build everything at once" tool. Break large goals into phases with a single
promise emitted at the end. Smaller, well-scoped phases converge faster and produce better results.

**BAD — too broad:**

```
Create a complete e-commerce platform.
```

This will drift: the worker may implement different subsets on each iteration, never fully converging.

**GOOD — phase decomposition with single terminal promise:**

```
Build Phase 1 of the e-commerce platform: authentication only.

Phase 1 scope:
- User registration (POST /auth/register)
- User login (POST /auth/login) returning a JWT
- Protected route middleware
- Tests for all three

Phase 2 (catalog) and Phase 3 (cart) are out of scope for this task.

When all Phase 1 tests pass: <promise>COMPLETE</promise>
```

Map each phase to a separate Plans.md `[ralph]` task. Each task gets its own `Verify:` command.
This gives the loop a tight, measurable target and lets subsequent phases build on confirmed foundations.

---

### A3. Self-Correction

Tell the worker to run its own verify step and fix failures before emitting the promise.
A TDD-style red-green-refactor loop with explicit retry instructions converts "maybe it works"
into "I confirmed it works."

**BAD — no self-check:**

```
Write code for feature X.
```

The worker writes code and emits `<promise>COMPLETE</promise>` without running tests. The
orchestrator runs the Verify command, it fails, and the loop restarts — wasting an iteration.

**GOOD — TDD-style loop with explicit retry steps:**

```
Implement feature X following TDD:

1. Write failing tests
2. Implement the feature
3. Run tests: <verify-command>
4. If any tests fail, debug and fix the implementation
5. Refactor if needed
6. Repeat steps 3-5 until all tests pass
7. Only when all tests are green: <promise>COMPLETE</promise>
```

The explicit self-correction loop means the worker catches failures before the orchestrator
does — converging faster with fewer wasted iterations.

---

### A4. Escape Hatches

Even well-written prompts can hit edge cases where progress stalls. Always provide
two layers of escape-hatch protection.

**Layer 1: `MaxIter` safety net in the task block**

Set `MaxIter:` explicitly when the default is wrong for the task — adjust upward for complex
tasks, downward for simple ones to catch runaway loops early.

For the full task-block format including `Verify:` and `MaxIter:` syntax and the default value,
see [`when-to-ralph.md` — Plans.md Syntax](${CLAUDE_SKILL_DIR}/references/when-to-ralph.md#plansmd-syntax).

**Layer 2: Escape-hatch instructions in the prompt**

Include a structured fallback clause:

```
If still failing after 3 consecutive iterations without improvement:
- Document what is blocking in RALPH_BLOCKERS.md
- List each approach already tried and why it failed
- Suggest 2-3 alternative approaches with trade-offs
- Output: <promise>COMPLETE</promise>

Note: the orchestrator uses ralph-worker-report.v1.summary for blocked diagnostics,
so write a clear one-line summary in that field as well.
```

**Important constraint**: harness-ralph-loop uses an exact-string promise match — you cannot
multiplex "SUCCESS" vs "BLOCKED" through the promise tag itself (both must emit the same
configured promise string). Use the `MaxIter` cap and the `ralph-worker-report.v1.summary`
field to distinguish success from blocked states in post-loop diagnostics.

---

## Section B: Philosophy

The Ralph loop is built on four foundational principles. Understanding them helps you write
better prompts and debug convergence problems faster.

1. **Iteration > Perfection** — Don't aim for a perfect prompt on the first try. Start with
   a clear verify command and success criteria, then refine the prompt based on where iterations
   stall. The loop itself is the refinement mechanism; let it work.

2. **Failures Are Data** — "Deterministically bad" means failures are predictable and tunable.
   Each failed iteration leaves context in the worktree (git log, test output, modified files)
   for the next worker to read. A clear failure with diagnostic output is more valuable than
   a vague partial success.

3. **Operator Skill Matters** — Success depends on prompt quality, not just model strength.
   The same underlying model can converge in 3 iterations or spin for 10 depending on how
   clearly the goal, verify command, and self-correction loop are specified. Prompt engineering
   is the primary lever you control.

4. **Persistence Wins** — Keep trying; the loop handles retry logic automatically. Transient
   failures, minor implementation missteps, and intermediate broken states are expected. The
   orchestrator exists precisely to absorb these and continue. Trust the loop, set a reasonable
   `MaxIter`, and let it run.

---

## Section C: When to Use Ralph

Ralph works best for tasks with an **automatic, unambiguous success signal** — a Verify command
that exits 0 when done and non-0 when not.

For a full mapping of task patterns (what is Ralph-suitable vs. not), DoD quality checks,
and Plans.md syntax examples, see:

[`when-to-ralph.md`](${CLAUDE_SKILL_DIR}/references/when-to-ralph.md)

---

## Section D: Learn More

- **Original technique**: https://ghuntley.com/ralph/
- **Ralph Orchestrator**: https://github.com/mikeyobrien/ralph-orchestrator
