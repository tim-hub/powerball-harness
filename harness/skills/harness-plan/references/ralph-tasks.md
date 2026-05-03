# Ralph Tasks — Iterative Loop Task Reference

`[ralph]` marks a task that should be executed by `harness-ralph-loop`.
The loop spawns a fresh Worker subagent per attempt, iterating until both a `<promise>` tag
AND a verify command succeed.

---

## When harness-plan Should Emit `[ralph]`

Apply `[ralph]` when the task description contains any of the following trigger phrases:

| Trigger Phrase | Example Task Description |
|----------------|--------------------------|
| "until tests pass" | "Fix auth module until tests pass" |
| "iterate until X" | "Iterate until lint is clean" |
| "fix until passing" | "Fix CI failures until passing" |
| "loop until clean" | "Loop until clean" |
| "until clean" | "Refactor database layer until clean" |
| "greenfield until tests green" | "Greenfield payment service until tests green" |

Detection is case-insensitive and matches substrings.

---

## Required Per-Task Fields When `[ralph]` Is Applied

When a task receives `[ralph]`, two extra lines are appended **below** the task row
(similar to comment lines — they are not columns):

### `Verify:` line (required)

A bash command that must exit 0 for the loop to terminate.
Auto-inferred from project type when not explicitly specified by the user:

| Project-type signal | Auto-inferred `Verify:` command |
|--------------------|---------------------------------|
| `package.json` present | `npm test` |
| `pyproject.toml` or `setup.py` present | `pytest` |
| `go.mod` present | `go test ./...` |
| None of the above found | `# TODO: set Verify command` |

Inference priority: check in the order listed above; use the first match.

### `MaxIter:` line (optional)

Maximum number of Worker attempts before the loop gives up.
Default: `10` (applied automatically when the line is omitted).

---

## Worked Example

The following is a concrete Plans.md excerpt showing a `[ralph]` task block:

```markdown
| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 3.2  | Fix flaky integration tests [ralph] | All tests pass (`npm test` exits 0) | 3.1 | cc:TODO |
Verify: npm test
MaxIter: 15
```

Key points:
- `[ralph]` appears inline in the Description column
- `Verify:` and `MaxIter:` appear as plain lines **below** the task row, not as columns
- `MaxIter:` is optional; omit to use the default of 10

---

## Serialization Warning

**Only one `[ralph]` task may be active at a time within a session.**

Ralph tasks serialize within a session because each attempt spawns a fresh Worker subagent
that may mutate shared state (files, test fixtures, databases).
Running two `[ralph]` loops concurrently risks conflicting writes and non-deterministic outcomes.

`harness-ralph-loop` enforces this constraint: if another `[ralph]` task is already in a
`cc:WIP` state, the new loop will wait rather than start a second concurrent loop.
