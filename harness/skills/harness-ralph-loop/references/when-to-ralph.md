# When to Use Ralph — Task Pattern Guide

This guide maps the Ralph-loop philosophy to harness Plans.md task patterns.
Use it to decide whether a task should receive the `[ralph]` marker when planning with `harness-plan`.

---

## Core Philosophy

Ralph's value comes from **iteration with memory**: each worker attempt reads everything the previous
attempt left behind (git log, modified files, test output in the worktree). This is more powerful than
retrying from scratch, and more tractable than a single long agent run — the orchestrator can inject
structured failure context and enforce a hard stop when the loop stalls.

The flip side: Ralph only works when there is an **automatic, unambiguous success signal** — a command
that exits 0 when done, non-0 when not. Without that signal, the orchestrator cannot determine when to stop.

---

## Good for Ralph

These task patterns benefit from iterative loop execution. Use `[ralph]` freely here.

| Pattern | Example Plans.md Description | Why Ralph Fits |
|---------|------------------------------|----------------|
| Fix failing tests until passing | "Fix auth module until tests pass [ralph]" | `npm test` or `pytest` provides a clear exit-0 signal |
| Refactor until lint clean | "Refactor legacy API until lint clean [ralph]" | `eslint src/` exits 0 when all violations are resolved |
| Greenfield until tests green | "Build payment service until `pytest -x` exits 0 [ralph]" | TDD loop: red → green is a natural Ralph use case |
| Fix flaky integration tests | "Fix flaky integration tests [ralph]" | Iteration naturally converges; each attempt improves coverage |
| Migrate until type-check passes | "Migrate user module to TypeScript until `tsc --noEmit` clean [ralph]" | Compiler errors are precise; each iteration reduces the count |
| Cleanup until no warnings | "Remove deprecated API calls until `go build ./...` warning-free [ralph]" | Compiler warnings are enumerable and diminish per iteration |
| Build a feature until acceptance test passes | "Add JWT refresh until `test/auth/refresh.test.ts` exits 0 [ralph]" | Narrow acceptance test is the verify command |

**Common trigger phrases** (harness-plan detects these automatically):

- "until tests pass" / "until test passes"
- "iterate until X" / "until X is clean"
- "fix until passing" / "fix until clean"
- "loop until clean"
- "greenfield until tests green"
- "until `<command>` exits 0"

---

## Not Good for Ralph

These patterns should NOT use `[ralph]`. Write them as standard tasks without the marker.

| Anti-Pattern | Why Ralph Doesn't Help |
|--------------|------------------------|
| **Design decisions requiring human judgment** | No automatic verifier. "Design the new API shape" has no command that exits 0 when the design is good. |
| **One-shot operations** | Migrations, schema changes, one-time scripts — retry loops add risk without benefit. Use a standard task with a manual DoD. |
| **Tasks with unclear or unmeasurable DoD** | If you can't write a Verify command, you can't use Ralph. "Improve code quality" is not Ralph-suitable; "fix all eslint violations" is. |
| **Production debugging** | Use `/diagnose` instead. Live system state changes between iterations; Ralph assumes a stable verify baseline. |
| **Exploratory spikes** | Spikes are open-ended. There's no pre-defined "done" to verify against. |
| **Tasks that must interact with a human mid-flow** | Ralph workers run headless. Any task requiring confirmation, approval, or input during execution will stall. |
| **Concurrent-unsafe operations** | Database migrations, file system destructive ops — iterating these risks leaving the system in an undefined state. |

---

## DoD Quality Check

A `[ralph]` task's Verify command must be a command where **exit 0 = unambiguously done**.

### Good Verify Commands

```bash
# Test runner: exits 0 when all tests pass
npm test
pytest
go test ./...
cargo test

# Specific test file
npx jest src/auth/auth.test.ts
pytest tests/test_auth.py

# Type checking
npx tsc --noEmit
mypy src/

# Lint
eslint src/ --max-warnings 0
golangci-lint run

# Custom acceptance test
bash scripts/acceptance-test.sh
test -f /tmp/marker-file.txt && echo OK
```

### Bad Verify Commands (Do Not Use)

```bash
# Human evaluation required — not automatable
echo "Check if this looks good"

# Always succeeds regardless of state
true
exit 0

# Vague, not tied to the specific DoD
ls -la

# Depends on external state that may not exist
curl https://production.example.com/health
```

**If you can't write a Verify command, the task isn't Ralph-suitable.**

Rewrite the task as a standard `cc:TODO` task with a manual DoD, or decompose it until you can
identify a subset that IS automatically verifiable.

---

## Plans.md Syntax

When adding a `[ralph]` task manually (rather than via `harness-plan`), use this format:

```markdown
| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 3.2  | Fix flaky integration tests [ralph] | All tests pass (`npm test` exits 0) | 3.1 | cc:TODO |
Verify: npm test
MaxIter: 15
```

- `[ralph]` appears **inline** in the Description column
- `Verify:` and `MaxIter:` appear as **plain lines below** the task row, not as columns
- `MaxIter:` is optional; omit to use the default of 10
- Only one `[ralph]` task may be active (cc:WIP) at a time in a session — they serialize

---

## Serialization Warning

**Only one `[ralph]` task may be active at a time within a session.**

Ralph tasks serialize because each worker attempt may mutate shared state (files, test fixtures, databases).
Running two Ralph loops concurrently risks conflicting writes and non-deterministic outcomes.

If you need parallel Ralph execution, run them in separate sessions.
