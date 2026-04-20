# create Subcommand -- Plan Creation Flow

Gathers ideas and requirements through a hearing, then generates an actionable Plans.md.

## Step 0: Check Conversation Context

If requirements can be extracted from the preceding conversation, confirm:

> Choose how to create the plan:
> 1. From the preceding conversation -- Create a plan based on the brainstorm content
> 2. From scratch -- Start with a hearing

If "From the preceding conversation": Extract requirements, ideas, and decisions and confirm with the user.
After confirmation, skip to Step 3 (technical research).

## Step 1: Ask What to Build

If there is no user input, ask:

> What do you want to build?
>
> Examples: Reservation management system / Blog site / Task management app / API server
>
> A rough idea is fine!

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

Attach a brief only when needed. The brief does not replace Plans.md -- it is a supplementary document that briefly locks down implementation prerequisites.

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
Tasks tagged `[needs-spike]` automatically get a **spike (technical validation) task** generated ahead of them:

```markdown
| N.X-spike | [spike] Technical validation for {{task name}} | Create validation result report | - | cc:TODO |
| N.X       | {{task name}} [needs-spike] | {{DoD}} | N.X-spike | cc:TODO |
```

The spike task's completion criterion is "leave a validation result report (feasible / infeasible / needs design change)."

## Step 5.5: TDD Skip Decision (Enabled by Default)

TDD is enabled by default. Only tasks matching one of the following conditions receive a `[skip:tdd]` marker to skip:

| Skip Condition | Reason |
|---------------|--------|
| Documentation/comments only | Does not affect executable code |
| Configuration files only (JSON, YAML, .env) | No testable logic |
| Single-line or trivial fix (typo) | Test cost exceeds benefit |
| Style/formatting changes only | Does not affect behavior |
| Dependency updates only | No implementation logic change |
| README/CHANGELOG updates | Documentation only |
| Refactoring (no behavior change) | Covered by existing tests |

Tasks not matching the above have TDD automatically applied (test-first recommended).

## Step 5.7: Plans.md v3 Format Specification

Plans.md v3 includes the following format extensions:

### Phase Header Purpose Line (Optional)

Each Phase header can include a one-line Purpose. Omit if no input is provided:

```markdown
### Phase N.X: [Phase Name] [Px]

Purpose: [What problem this phase solves, in one line]
```

- **Default**: Do not prompt for input (omit if blank)
- **When included**: Displayed during breezing Phase 0 scope confirmation
- **Generation rule**: Auto-include only when the user explicitly states the phase's purpose

### Artifact Notation (Status Column)

Attach commit hash to Status upon task completion:

```markdown
| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 1.1  | ... | ... | - | cc:done [a1b2c3d] |
| 1.2  | ... | ... | 1.1 | cc:TODO |
```

- **Format**: `cc:done [7-char hash]`
- **When applied**: Automatically applied at `harness-work` Solo Step 7
- **Backward compatibility**: Hashless `cc:done` remains valid

### Affected Files

Files related to the v3 format:

| File | Impact |
|------|--------|
| `skills/harness-plan/references/create.md` | Purpose line added to Step 6 template |
| `skills/harness-plan/references/sync.md` | Discrepancy detection recognizes `cc:done [hash]` format |
| `skills/harness-work/SKILL.md` | Hash applied at Solo Step 7, re-ticketing on failure |
| `skills/harness-sync/SKILL.md` | Snapshot saved with --snapshot |
| `skills/breezing/SKILL.md` | Progress displayed in Progress Feed |

## Step 6: Generate Plans.md

Auto-generate quality markers + DoD + Depends and produce Plans.md.

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
    +-- other -> no marker (TDD enabled by default)
```

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

**Generation template**:

```markdown
# [Project Name] — Plans.md

Last release: (none yet)

---

## Phase 1: [Phase Name]

Created: YYYY-MM-DD

Purpose: [Phase purpose (optional)]

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 1.1  | [Task description] [feature:security] | [Verifiable completion criteria] | - | cc:TODO |
| 1.2  | [Task description] | [Verifiable completion criteria] | 1.1 | cc:TODO |

---

## Future Considerations

(none currently)

---

## Archive

- Last archive: (none yet)
- Other older phases have been moved to `.claude/memory/archive/` to keep this file lean.
```

**Purpose line**:
- Auto-include only when the user states the phase purpose
- Omit the entire Purpose line if no input (do not leave a blank line)
- Must be a single line (no multi-line)

**DoD (Definition of Done) notation**:
- Write as a single verifiable line (e.g., "Tests pass", "Migration can be executed", "0 lint errors")
- Phrases like "looks good" or "works properly" are prohibited. Must be answerable with Yes/No

**Depends notation**:
- No dependency: `-`
- Single dependency: Task number (e.g., `1.1`)
- Multiple dependencies: Comma-separated (e.g., `1.1, 1.2`)
- Phase dependency: Phase number (e.g., `Phase 1`)

### Team mode output

Only when the user explicitly requests team mode, provide an issue bridge dry-run alongside Plans.md.

- Only one tracking issue
- List sub-issue payloads for each task
- Plans.md remains the source of truth
- Provide in a form directly usable from `scripts/plans-issue-bridge.sh --team-mode` dry-run

## Step 7: Next Action Guidance

> Plans.md complete!
>
> Next steps:
> - Start implementation with `harness-work`
> - Or say "start from Phase 1"
> - Add features with `harness-plan add [feature name]`
> - Defer features with `harness-plan update [task] blocked`

## CI Mode (--ci)

No hearing. Uses the existing Plans.md as-is and only performs task decomposition.

1. Load Plans.md
2. List cc:TODO tasks in priority order
3. Mark parallelizable tasks with `[P]`
4. Suggest the next task to execute
