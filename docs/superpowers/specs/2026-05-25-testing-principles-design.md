# Testing Principles — Design Spec

**Date**: 2026-05-25  
**Scope**: Downstream projects only (propagated via `harness-setup`)  
**Status**: Approved — ready for implementation

---

## Problem

The harness TDD infrastructure (`tdd-guidelines.md.template`, `tdd-order-check.sh`) is
advisory-only: it suggests test-first but never enforces it through the planning or work
loop. Two gaps exist:

1. Planner creates a single task for a feature; the Worker can implement before writing
   a test with no structural barrier.
2. Phases have no required verification step — a phase can be marked complete with no
   proof that the feature works end-to-end.

---

## Design

Minimal-change approach: edit three existing files. No new files, no script lib
extraction, no movement of `testing-anti-patterns.md`.

### Files changed

| File | Change summary |
|------|---------------|
| `harness/templates/rules/tdd-guidelines.md.template` | Flip suggestion → enforcement; add three-lane phase-bottom verification section |
| `harness/skills/harness-plan/references/create.md` | Step 5.5 task-split rule; Step 6 marker + template; Step 6.5 self-review checks |
| `harness/skills/harness-work/references/solo-mode.md` | Step 1 dependency guard; Step 3 `.a`/`.b` behaviour table; Step 11 phase-close gate |

---

## Section 1 — Template: `tdd-guidelines.md.template`

### Edit 1 — Flip suggestion to enforcement

Replace the `## Notes` section:

```markdown
## Notes
- **Enforcement**: Test-first is the default. Tasks tagged `[feature]` or `[bugfix]`
  are automatically split by the planner (see below). Add `[skip:tdd]` to a task to
  opt out.
- **Respect existing tests**: Don't create duplicates.
- **When no test framework is detected**: Auto-apply `[skip:tdd]`.
```

### Edit 2 — Update Decision Logic

```
Task received
    |
Check tags and [skip:tdd] absence
    |
    |-- [feature] or [bugfix] (no [skip:tdd])
    |       → Planner SPLITS into:
    |           N.a  [tdd:test-first]  Write failing tests
    |           N.b                    Implement to pass (blocked-by N.a)
    |
    |-- [skip:tdd] present → single task, no split
    |
    |-- [config] / [docs] / [style] / [refactor] → auto-apply [skip:tdd]
```

### Edit 3 — Add Phase-bottom Verification section

```markdown
## Phase-bottom Verification

At the end of every phase, the planner appends one verification task.
The Worker cannot close a phase until this task is `cc:done`.

### Lane Selection

All lanes are **agent-executable** — the Worker runs them via Bash, no human action required.

| Project type | Lane | Agent execution method | Verification criterion |
|---|---|---|---|
| Backend / API | **curl** | `bash`: `curl -f <endpoint>` | All critical routes return HTTP 2xx |
| Frontend / UI | **chrome-devtools** | `mcp__chrome_devtools__*` tools (Claude Code native Chrome) | Golden path completes, 0 console errors |
| CLI tool | **CLI** | `bash`: run the command end-to-end | Exit 0 + expected stdout matches |
| All tasks are `[skip:tdd]` | **— skip —** | N/A | No phase-bottom task appended |

Lane is inferred from the phase's task content. If mixed, use the dominant type.

### Verification Task Format

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| `N.e2e` | `[verify:e2e]` Phase N E2E — `<lane>` | `<lane-specific criterion>` | `<last-impl-task>` | `cc:TODO` |

**Skip rule**: If every task in the phase carries `[skip:tdd]` (docs/config-only phase),
no `N.e2e` task is appended.
```

---

## Section 2 — Planner: `create.md`

### Edit 1 — Step 5.5: Task-split rule

Append after the skip-condition table:

```markdown
### Task Split — Test-First Pair

For every task that does NOT receive `[skip:tdd]`, split into a blocking pair:

| Suffix | Tag | Description | DoD | Depends |
|--------|-----|-------------|-----|---------|
| `N.a` | `[tdd:test-first]` | Write failing tests: {{task name}} | Failing test file exists and runs red | prior dep or `-` |
| `N.b` | _(original tags)_ | {{original description}} | {{original DoD}} | `N.a` |

Rules:
- `N.b` always lists `N.a` in `Depends`.
- Original DoD moves to `N.b`; `N.a` DoD is always "Failing test file exists and runs red".
- If the original task had a prior dependency (e.g., `1.1`), `N.a` inherits it; `N.b` depends on `N.a`.
- `[bugfix]` tasks: `N.a` DoD is "Reproduction test exists and fails on current code".
```

### Edit 2 — Step 6: Quality Marker Assignment

Extend marker block:

```
+-- [feature] / [bugfix] (no [skip:tdd]) → split into N.a [tdd:test-first] + N.b
```

### Edit 3 — Step 6: Generation template

Show the split and phase-bottom row in the example Plans.md block:

```markdown
| Task  | Description | DoD | Depends | Status |
|-------|-------------|-----|---------|--------|
| 1.1.a | [tdd:test-first] Write failing tests: User login | Failing test runs red | - | cc:TODO |
| 1.1.b | Implement user login [feature:security] | Tests pass, curl returns 200 | 1.1.a | cc:TODO |
| 1.2.a | [tdd:test-first] Write failing tests: Password reset | Failing test runs red | - | cc:TODO |
| 1.2.b | Implement password reset | Tests pass | 1.2.a | cc:TODO |
| 1.e2e | [verify:e2e] Phase 1 E2E — curl: login + reset return 2xx | `curl -f` exits 0 for both routes | 1.2.b | cc:TODO |
```

### Edit 4 — Step 6.5: Self-Review

Add two new checks:

```markdown
4. **TDD pair completeness** — Every task without `[skip:tdd]` must have a corresponding
   `.a` row. Fix any missing splits before presenting to the user.
5. **Phase-bottom verification** — Every phase with at least one non-`[skip:tdd]` task
   must end with an `N.e2e` row. Fix any missing.
```

---

## Section 3 — Worker: `solo-mode.md`

### Edit 1 — Step 1: Dependency guard

After "identify the target task":

```markdown
**Dependency check**: Before claiming a task, verify its `Depends` column.
If any listed dependency is not yet `cc:done`, skip this task and select the next
eligible one. For `.b` tasks whose `.a` is still open: redirect the Worker to `.a` first.
```

### Edit 2 — Step 3: `.a` vs `.b` TDD behaviour

Replace the current TDD phase block:

```markdown
**TDD Phase** — behaviour depends on task tag:

| Task type | Step 3 action |
|---|---|
| `[tdd:test-first]` (an `.a` task) | **This task IS the TDD phase.** Write the failing test file; confirm it runs red. Commit as `test: failing tests for {{feature}}`. |
| `.b` task (follows a `.a`) | Confirm tests from `.a` still run red, then proceed to Step 6. No new test file needed. |
| No split (legacy task, no `[skip:tdd]`) | Existing behaviour — create test file first, confirm failure. |
| `[skip:tdd]` present | Skip Step 3 entirely. |
```

### Edit 3 — Step 11: Phase-close gate

After marking `cc:done`:

```markdown
**Phase-close check**: Scan the current phase in Plans.md.
- All tasks except `[verify:e2e]` are `cc:done` AND `N.e2e` is `cc:TODO`:
  → "Phase N implementation complete. Next: run the E2E verification task (N.e2e)."
  Auto-select `N.e2e` as the next task (or surface it in manual mode).
- `[verify:e2e]` task is `cc:done`: phase is fully closed.
- No `[verify:e2e]` task (docs/config-only phase): phase closes normally.
```

---

## End-to-end flow

```
harness-plan creates:
  1.1.a  [tdd:test-first]  Write failing tests           → Depends: -
  1.1.b  Implement feature                               → Depends: 1.1.a
  1.e2e  [verify:e2e]      Phase 1 E2E — curl: login+reset 2xx  → Depends: 1.1.b

harness-work picks up:
  Step 1  → 1.1.b blocked; selects 1.1.a
  Step 3  → writes failing tests, commits "test: failing tests for feature"
  Step 11 → marks 1.1.a cc:done
  ─────────────────────────────────────────────────
  Step 1  → 1.1.a done; 1.1.b now eligible
  Step 3  → confirms tests still red, implements
  Step 11 → marks 1.1.b cc:done → phase-close check fires
            "Phase 1 implementation complete — run N.e2e next"
  ─────────────────────────────────────────────────
  Runs 1.e2e → curl smoke → marks cc:done → phase fully closed
```

---

## Out of scope

These were considered and explicitly deferred:

| Item | Decision |
|------|---------|
| Extract shell helpers to `harness/scripts/lib/tdd-checks.sh` | Deferred — `tdd-order-check.sh` stays as-is |
| Move `testing-anti-patterns.md` to `templates/rules/` | Deferred — stays in `harness-work/references/` |
| `.claude/rules/` changes | Out of scope — downstream-only |

---

## Acceptance criteria

- [ ] `tdd-guidelines.md.template` no longer says "Suggestion, not enforcement"; three-lane phase-bottom section present
- [ ] `create.md` Step 5.5 has the task-split rule and format example
- [ ] `create.md` Step 6 marker logic includes `[tdd:test-first]` split; template shows `.a`/`.b`/`.e2e` rows
- [ ] `create.md` Step 6.5 has TDD pair completeness and phase-bottom checks
- [ ] `solo-mode.md` Step 1 has dependency guard
- [ ] `solo-mode.md` Step 3 has the `.a`/`.b` behaviour table
- [ ] `solo-mode.md` Step 11 has phase-close gate
- [ ] Docs/config-only phases (all `[skip:tdd]`) produce no `N.e2e` task
