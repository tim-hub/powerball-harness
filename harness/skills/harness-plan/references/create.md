# create Subcommand -- Plan Creation Flow

Gathers ideas and requirements through a hearing, then creates an actionable plan by issuing `harness plan-cli` calls. The SSOT is `.claude/harness/plans.json`; never write `Plans.md` markdown directly.

## Step 0: Check Conversation Context

If requirements can be extracted from the preceding conversation, confirm:

> Choose how to create the plan:
> 1. From the preceding conversation -- Create a plan based on the brainstorm content
> 2. From scratch -- Start with a hearing

If "From the preceding conversation": Extract requirements, ideas, and decisions and confirm with the user.
After confirmation, skip to Step 3 (technical research).

## Step 0.5: Scope Check

Before asking questions, assess whether the request spans multiple independent subsystems (e.g., "build a platform with auth, payments, and an admin dashboard").

If it does, suggest breaking it into **phases** rather than one monolithic plan:

> This covers [N] independent areas. I recommend splitting into phases:
> - Phase 1: [Subsystem A] — can be built and verified independently
> - Phase 2: [Subsystem B] — depends on Phase 1's foundation
> - ...
>
> Each phase produces working, testable software on its own. Want to start with Phase 1?

If the user agrees: proceed with Phase 1 scope only. Capture remaining subsystems for later (e.g., as a comment on the phase via `harness plan-cli comment`, or note them back to the user) rather than creating their tasks now.

If the scope is appropriately contained for a single plan, continue to Step 1.

## Step 1: Ask What to Build

If there is no user input, ask:

> What do you want to build?
>
> Examples: Reservation management system / Blog site / Task management app / API server
>
> A rough idea is fine!

## Step 1.5: Memory Conflict Check

Before creating the plan, read `.claude/memory/decisions.md` and `.claude/memory/patterns.md`.

Scan the proposed approach against recorded decisions and patterns:

- **If a conflict is found** (e.g., the plan uses a pattern previously rejected, or contradicts a decision): present it clearly and ask the user to choose:
  1. **Update memory** — the new approach supersedes the old decision; run `harness-remember` to record it first, then continue planning
  2. **Adjust the plan** — keep the existing decision in place and revise the proposed approach to align with it

- **If no conflict is found**: proceed silently. Do not mention this step to the user.

> **Silent-pass rule**: this step must not slow down planning when there is no conflict. Surface only genuine conflicts.

## Step 1.6: New Decision Capture

If planning surfaces a new architectural decision not yet recorded in `.claude/memory/decisions.md`, prompt the user once before creating the plan:

> "I noticed a new decision: **[decision summary]**. Would you like to record it in memory before we write the plan? (yes / skip)"

- If yes → invoke `harness-remember` to capture it, then continue.
- If skip → continue without recording.

> This step only fires when a genuinely new decision is surfaced. Do not prompt for known patterns.

## Step 1.7: Planning Quality Gate

For `create` (new plan) and `add` (high-impact task addition), run the planning quality contract before creating the plan.

See [`${CLAUDE_SKILL_DIR}/references/planning-quality.md`](${CLAUDE_SKILL_DIR}/references/planning-quality.md) for the full 8-step protocol:

| Step | Name | When to run |
|------|------|-------------|
| 0 | Applicability | Always — decide if the contract applies |
| 1 | Input Decomposition | Break user input into subject, intent, facts, evidence |
| 2 | Latest-information Fetch | WebSearch for external facts |
| 3 | Local-source-of-truth Check | Reconcile against the plan (`harness plan-cli list`/`get`), specs, docs |
| 4 | Memory Check | Check harness-mem / local memory |
| 5 | Subagent Debate | 3–4 independent perspectives (Product, Arch, QA, Skeptic) |
| 6 | Neutral Scoring Review | 5-point rubric across 6 axes |
| 7 | Quality Contract Output | Decision-ready summary |
| 8 | Plan / Spec Output | Convert adopted proposals to task contracts via `harness plan-cli` |

**Skip criteria** (Step 0): marker-only `update`, status-only `sync`, typo/README/CHANGELOG changes, or narrow changes with a fixed spec answer.

## Step 2: Increase Resolution (Up to 3 Questions)

> Tell me a bit more:
>
> 1. Who will use it? (Just you? A team? Public?)
> 2. Any reference services you'd like to emulate?
> 3. How far do you want to go? (MVP? Full features?)

## Step 3: Technical Research (WebSearch)

Do not ask the user -- Claude Code researches and proposes.

```
WebSearch:
- "{{project type}} tech stack 2025"
- "{{similar service}} architecture"
```

## Step 4: Extract Feature List

Extract a concrete feature list from the requirements.

Example: For a reservation management system
- User registration/login
- Reservation calendar display
- Reservation creation/editing/cancellation
- Admin dashboard
- Email notifications
- Payment functionality

## Step 4.5: Optional Brief Generation

Attach a brief only when needed. The brief does not replace the plan -- it is a supplementary document that briefly locks down implementation prerequisites.

- For tasks involving UI, include a `design brief`
- For tasks involving API, include a `contract brief`
- When UI and API coexist, separate the briefs

### design brief

A brief for UI tasks should include at minimum:

- What you want to achieve
- Who will use it
- Important screen states
- Appearance and interaction constraints
- Completion criteria

### contract brief

A brief for API tasks should include at minimum:

- What to receive / what to return
- Input validation rules
- Failure behavior
- External dependencies
- Completion criteria

## Step 5: Create Priority Matrix (2-Axis Evaluation)

Evaluate each feature on **Impact x Risk (uncertainty)** across 2 axes:

- **Impact**: User value x Number of affected users (high/low)
- **Risk**: Technical unknowns x External dependencies (high/low)

| Impact \ Risk | Low Risk | High Risk |
|--------------|---------|---------|
| **High Impact** | ★ **Required** -- Top priority (value is certain) | ▲ **Required + [needs-spike]** -- Needs early validation |
| **Low Impact** | ○ **Recommended** -- Address if capacity allows | ✕ **Optional** -- Defer or reduce scope |

### `[needs-spike]` Marker

Tasks with High Impact x High Risk are automatically tagged `[needs-spike]`.
Tasks tagged `[needs-spike]` automatically get a **spike (technical validation) task** added ahead of them via `harness plan-cli add-task`:

```bash
harness plan-cli add-task <phase-id> --name "[spike] Technical validation for {{task name}}" \
  --dod "Create validation result report" --marker "spike"
harness plan-cli add-task <phase-id> --name "{{task name}} [needs-spike]" \
  --dod "{{DoD}}" --depends "<spike-task-id>"
```

The spike task's completion criterion is "leave a validation result report (feasible / infeasible / needs design change)."

## Step 5.5: TDD (Test-Driven Development) — Enabled by Default

TDD is encouraged and enabled by default. Only tasks matching one of the following conditions receive a `[skip:tdd]` marker to skip:

| Skip Condition | Reason |
|---------------|--------|
| Documentation/comments only | Does not affect executable code |
| Configuration files only (JSON, YAML, .env) | No testable logic |
| Single-line or trivial fix (typo) | Test cost exceeds benefit |
| Style/formatting changes only | Does not affect behavior |
| Dependency updates only | No implementation logic change |
| README/CHANGELOG updates | Documentation only |
| Refactoring (no behavior change) | Covered by existing tests |

Tasks not matching the above have TDD automatically applied (test-first).

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

## Step 5.7: Plan Field Conventions

The plan (`.claude/harness/plans.json`) supports the following conceptual fields:

### Phase Goal / Purpose (Optional)

Each phase can carry a one-line goal describing what problem it solves. Set it via `--goal` when creating the phase:

```bash
harness plan-cli add-phase --title "[Phase Name]" --goal "[What problem this phase solves, in one line]"
```

- **Default**: Do not prompt for input (omit `--goal` if blank)
- **When included**: Displayed during breezing Phase 0 scope confirmation
- **Generation rule**: Auto-include only when the user explicitly states the phase's purpose

### Artifact Notation (commit hash on completion)

Record the commit hash when a task is marked done, via `harness plan-cli update`:

```bash
harness plan-cli update 1.1 --status cc:done --hash a1b2c3d
```

- **Format**: 7-char short hash, passed with `--hash`
- **When applied**: Automatically applied at `harness-work` Solo Step 7
- **Backward compatibility**: `cc:done` without a hash remains valid

### Affected Files

Files related to these field conventions:

| File | Impact |
|------|--------|
| `skills/harness-plan/references/create.md` | Phase goal set via `add-phase --goal` |
| `skills/harness-plan/references/sync.md` | Discrepancy detection recognizes `cc:done` with/without hash |
| `skills/harness-work/SKILL.md` | Hash applied at Solo Step 7, re-ticketing on failure |
| `skills/harness-sync/SKILL.md` | Snapshot saved with --snapshot |
| `skills/breezing/SKILL.md` | Progress displayed in Progress Feed |

## Step 6: Create the Plan via `harness plan-cli`

Auto-generate quality markers + DoD + Depends, then create the phase and its tasks by issuing `harness plan-cli` calls (one `add-phase`, then one `add-task` per task). Do **not** write `Plans.md` markdown — the SSOT is `.claude/harness/plans.json`.

### Quality Marker Assignment Logic
```
Analyze task content
    |
    +-- "auth" "login" "API" -> [feature:security]
    +-- "component" "UI" "screen" -> [feature:a11y]
    +-- "fix" "bug" -> [bugfix:reproduce-first]
    +-- "docs" "comment" "README" "CHANGELOG" -> [skip:tdd]
    +-- "config" "json" "yaml" "env" -> [skip:tdd]
    +-- "style" "format" "lint" -> [skip:tdd]
    +-- "refactor" (no behavior change) -> [skip:tdd]
    +-- "payment" "billing" -> [feature:security]
    +-- "until tests pass" / "iterate until" / "fix until" / "loop until" / "until clean" -> [ralph]
    +-- other -> no marker (TDD enabled by default)
    +-- [feature] / [bugfix] (no [skip:tdd]) → split into N.a [tdd:test-first] + N.b (see Step 5.5)
```

When `[ralph]` is applied, also auto-fill the `Verify:` line below the task row using project-type inference:

| Project-type signal | Auto-inferred `Verify:` command |
|--------------------|---------------------------------|
| `package.json` present | `npm test` |
| `pyproject.toml` or `setup.py` present | `pytest` |
| `Cargo.toml` present | `cargo test` |
| `go.mod` present | `go test ./...` |
| None found | `# TODO: set Verify command` |

See [references/ralph-tasks.md](${CLAUDE_SKILL_DIR}/references/ralph-tasks.md) for full `[ralph]` task format including `MaxIter:` defaults and worked example.

### DoD Auto-Inference Logic

Infer DoD from task "Description" keywords and auto-fill:

| Task Description Keywords | DoD Inference |
|--------------------------|---------------|
| "create" "new" "add" | File exists with expected structure |
| "test" | Tests pass (`npm test` / `pytest`, etc.) |
| "fix" "bug" | Issue no longer reproduces |
| "UI" "screen" "component" | Visual confirmation (screenshot or browser) |
| "API" "endpoint" | Response confirmed via curl/httpie |
| "config" "settings" | Configuration values take effect |
| "documentation" "docs" | File exists with no broken links |
| "migration" "DB" | Migration can be executed |
| "refactoring" | All existing tests pass + 0 lint errors |

Inference results are default values only. If the user specifies concrete acceptance criteria, those take priority.

### Depends Auto-Inference Logic

Infer dependencies between tasks within a phase using the following rules:

1. **DB/schema tasks** -> Depended on by other implementation tasks (predecessor)
2. **UI tasks** -> Depend on API/logic tasks (successor)
3. **Test/verification tasks** -> Depend on implementation tasks (last in sequence)
4. **Config/environment tasks** -> Depended on by other tasks (predecessor)
5. **Tasks with no clear dependency** -> `-` (can run in parallel)

When inference confidence is low, set to `-` and request user confirmation.

**Generation sequence** (create the phase first, then add each task, wiring `--depends` to the task IDs returned by prior `add-task` calls):

```bash
# 1. Create the phase (capture the returned phase-id, e.g. 1)
harness plan-cli add-phase --title "[Phase Name]" --goal "[Phase purpose (optional)]"

# 2. Add each task to that phase (new tasks default to cc:TODO)
harness plan-cli add-task 1 --name "Write failing tests: User login" \
  --dod "Failing test runs red" --marker "tdd:test-first"
harness plan-cli add-task 1 --name "Implement user login" \
  --dod "Tests pass, curl returns 200" --depends "1.1.a" --marker "feature:security"
harness plan-cli add-task 1 --name "Write failing tests: Password reset" \
  --dod "Failing test runs red" --marker "tdd:test-first"
harness plan-cli add-task 1 --name "Implement password reset" \
  --dod "Tests pass" --depends "1.2.a"
harness plan-cli add-task 1 --name "[verify:e2e] Phase 1 E2E — curl: login + reset return 2xx" \
  --dod "\`curl -f\` exits 0 for both routes" --depends "1.2.b"
```

- New tasks are created with the `cc:TODO` status by default.
- Pass quality markers (Step 6 logic), DoD, and Depends as flags. Use `--depends` to reference task IDs created earlier in the sequence.
- Archived phases stay in `.claude/harness/plans.json` with `status: archived` (no separate archive markdown file). Defer-only subsystems can be noted via `harness plan-cli comment`.

**Purpose line**:
- Auto-include only when the user states the phase purpose
- Omit the entire Purpose line if no input (do not leave a blank line)
- Must be a single line (no multi-line)

**DoD (Definition of Done) notation**:
- Write as a single verifiable line (e.g., "Tests pass", "Migration can be executed", "0 lint errors")
- Must be answerable with Yes/No

**Banned DoD phrases** — never write these:
- "looks good", "works properly", "works correctly", "seems fine"
- "TBD", "TODO", "implement later"
- "add appropriate error handling" (without specifying what)
- "add validation" (without specifying what to validate)
- "handle edge cases" (without naming the cases)
- "similar to task N" or "as above"

**Depends notation**:
- No dependency: `-`
- Single dependency: Task number (e.g., `1.1`)
- Multiple dependencies: Comma-separated (e.g., `1.1, 1.2`)
- Phase dependency: Phase number (e.g., `Phase 1`)

### Team mode output

Only when the user explicitly requests team mode, provide an issue bridge dry-run alongside the created plan.

- Only one tracking issue
- List sub-issue payloads for each task
- `.claude/harness/plans.json` remains the source of truth
- Provide in a form directly usable from `scripts/plans-issue-bridge.sh --team-mode` dry-run

## Step 6.5: Self-Review

After creating the plan, review it (query state via `harness plan-cli list`/`get`) before presenting to the user:

1. **Feature coverage** — Does every feature from Step 4 map to at least one task? Add any missing tasks via `harness plan-cli add-task`.
2. **DoD quality** — Scan all task DoDs for banned phrases (see Step 6). Fix any found by re-adding or updating the task.
3. **Dependency consistency** — Do all task IDs referenced in `--depends` exist in the phase?
4. **TDD pair completeness** — Every task without `[skip:tdd]` must have a corresponding
   `.a` row. Fix any missing splits before presenting to the user.
5. **Phase-bottom verification** — Every phase with at least one non-`[skip:tdd]` task
   must end with an `N.e2e` row. Fix any missing.

Fix issues inline. No need to re-review — just fix and move on.

## Step 7: Next Action Guidance

> Plan complete!
>
> Next steps:
> - Start implementation with `harness-work`
> - Or say "start from Phase 1"
> - Add features with `harness-plan add [feature name]`
> - Defer features with `harness-plan update [task] blocked`

## CI Mode (--ci)

No hearing. Uses the existing plan as-is and only performs task decomposition.

1. Load the plan via `harness plan-cli list` (the SSOT is `.claude/harness/plans.json`)
2. List cc:TODO tasks in priority order (`harness plan-cli list --status cc:TODO`)
3. Mark parallelizable tasks with `[P]`
4. Suggest the next task to execute
