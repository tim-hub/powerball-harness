# create Subcommand — Plan Creation Flow

Gather ideas and requirements through a hearing process, then generate an actionable Plans.md.

## Step 0: Check Conversation Context

If requirements can be extracted from the recent conversation, confirm:

> Choose how to create the plan:
> 1. From recent conversation — Create plan based on brainstorming content
> 2. From scratch — Start with a hearing

For "From recent conversation": Extract requirements, ideas, and decisions, then confirm with user.
After confirmation, skip to Step 3 (technical research).

## Step 1: Ask What to Build

If no user input, ask:

> What do you want to build?
>
> Examples: Booking system / Blog site / Task management app / API server
>
> A rough idea is fine!

## Step 2: Increase Resolution (max 3 questions)

> Tell me a bit more:
>
> 1. Who will use it? (Just yourself? Team? Public?)
> 2. Any services you'd like to reference?
> 3. How far do you want to go? (MVP? Full feature set?)

## Step 3: Technical Research (WebSearch)

Claude Code researches and proposes without asking the user.

```
WebSearch:
- "{{project type}} tech stack 2025"
- "{{similar service}} architecture"
```

## Step 4: Feature List Extraction

Extract a concrete feature list from the requirements.

Example: For a booking management system
- User registration/login
- Booking calendar display
- Create/edit/cancel bookings
- Admin dashboard
- Email notifications
- Payment functionality

## Step 4.5: Optional Brief Generation

Attach briefs only when needed. Briefs are supplementary materials that briefly fix implementation assumptions and do not replace Plans.md.

- Tasks involving UI get a `design brief`
- Tasks involving API get a `contract brief`
- When both UI and API are involved, separate the briefs

### design brief

For UI task briefs, include at minimum:

- What to achieve
- Who uses it
- Important screen states
- Visual and interaction constraints
- Completion criteria

### contract brief

For API task briefs, include at minimum:

- What to receive / return
- Input validation conditions
- Failure behavior
- External dependencies
- Completion criteria

## Step 5: Priority Matrix (2-axis evaluation)

Evaluate each feature on **Impact x Risk (uncertainty)** axes:

- **Impact**: User value x number of target users (high/low)
- **Risk**: Technical unknowns x external dependencies (high/low)

| Impact \ Risk | Low Risk | High Risk |
|-------------|---------|---------|
| **High Impact** | ★ **Required** — Top priority (reliably delivers value) | ▲ **Required + [needs-spike]** — Needs early validation |
| **Low Impact** | ○ **Recommended** — Address if capacity allows | ✕ **Optional** — Defer or reduce scope |

### `[needs-spike]` Marker

Tasks with High Impact x High Risk automatically get the `[needs-spike]` marker.
Tasks with `[needs-spike]` get an auto-generated **spike (technical validation) task** that runs first:

```markdown
| N.X-spike | [spike] {{task name}} technical validation | Create validation result report | - | cc:TODO |
| N.X       | {{task name}} [needs-spike] | {{DoD}} | N.X-spike | cc:TODO |
```

The spike task's completion criteria is "produce a validation result report (feasible/infeasible/requires design change)."

## Step 5.5: TDD Skip Decision (enabled by default)

TDD is enabled by default. Only tasks matching the following criteria get the `[skip:tdd]` marker:

| Skip Condition | Reason |
|-------------|------|
| Documentation/comments only | Does not affect executable code |
| Configuration files only (JSON, YAML, .env) | No logic to test |
| Single-line or trivial fix (typo) | Test cost exceeds benefit |
| Style/formatting changes only | Does not affect behavior |
| Dependency updates only | No implementation logic changes |
| README/CHANGELOG updates | Documentation only |
| Refactoring (no behavior change) | Covered by existing tests |

Tasks not matching the above have TDD automatically applied (test-first recommended).

## Step 5.7: Plans.md v3 Format Specification

Plans.md v3 includes the following format extensions:

### Phase Header Purpose Line (optional)

Each Phase header can include a one-line Purpose. Omitted when no input is provided:

```markdown
### Phase N.X: [Phase Name] [Px]

Purpose: [The problem this phase solves in one line]
```

- **Default**: Input is not requested (omitted when empty)
- **When specified**: Displayed during breezing Phase 0 scope verification
- **Generation rule**: Auto-populated only when the user explicitly states the phase's purpose

### Artifact Notation (Status column)

Commit hash is appended to Status upon task completion:

```markdown
| Task | Description | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 1.1  | ... | ... | - | cc:done [a1b2c3d] |
| 1.2  | ... | ... | 1.1 | cc:TODO |
```

- **Format**: `cc:done [7-char hash]`
- **Timing**: Auto-appended during `harness-work` Solo Step 7
- **Backward compatibility**: Hashless `cc:done` remains valid

### Related Files

Files related to the v3 format:

| File | Impact |
|---------|------|
| `skills/harness-plan/references/create.md` | Purpose line added to Step 6 template |
| `skills/harness-plan/references/sync.md` | Drift detection recognizes `cc:done [hash]` format |
| `skills/harness-work/SKILL.md` | Hash appended in Solo Step 7, failure re-ticketing |
| `skills/harness-sync/SKILL.md` | Snapshot saving with --snapshot |
| `skills/breezing/SKILL.md` | Progress display in Progress Feed |

## Step 6: Plans.md Generation

Auto-generate quality markers + DoD + Depends to produce Plans.md.

### Quality Marker Assignment Logic
```
Analyze task content
    ↓
├── "auth" "login" "API" -> [feature:security]
├── "component" "UI" "screen" -> [feature:a11y]
├── "fix" "bug" -> [bugfix:reproduce-first]
├── "docs" "comment" "README" "CHANGELOG" -> [skip:tdd]
├── "config" "json" "yaml" "env" -> [skip:tdd]
├── "style" "format" "lint" -> [skip:tdd]
├── "refactor" (no behavior change) -> [skip:tdd]
├── "payment" "billing" -> [feature:security]
└── Other -> No marker (TDD enabled by default)
```

### DoD Auto-Inference Logic

Infer DoD from task "Description" keywords and auto-populate:

| Description Keywords | Inferred DoD |
|---------------------|---------|
| "create" "new" "add" | File exists and has expected structure |
| "test" | Tests pass (`npm test` / `pytest` etc.) |
| "fix" "bug" | Problem no longer reproduces |
| "UI" "screen" "component" | Display verified (screenshot or browser) |
| "API" "endpoint" | Response verified via curl/httpie |
| "config" "configuration" | Configuration values are applied |
| "documentation" "docs" | File exists and no broken links |
| "migration" "DB" | Migration is runnable |
| "refactoring" | All existing tests pass + lint errors 0 |

Inferred results are defaults. When the user specifies explicit acceptance criteria, those take priority.

### Depends Auto-Inference Logic

Infer inter-task dependencies within a phase using the following rules:

1. **DB/schema tasks** -> Depended on by other implementation tasks (predecessor)
2. **UI tasks** -> Depend on API/logic tasks (successor)
3. **Test/verification tasks** -> Depend on implementation tasks (last)
4. **Config/environment tasks** -> Depended on by other tasks (predecessor)
5. **Tasks with no clear dependency** -> `-` (can run in parallel)

When confidence is low, set to `-` and ask the user for confirmation.

**Generation template**:

```markdown
# [Project Name] Plans.md

Created: YYYY-MM-DD

---

## Phase 1: [Phase Name]

Purpose: [Phase purpose (optional)]

| Task | Description | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 1.1  | [Task description] [feature:security] | [Verifiable completion criteria] | - | cc:TODO |
| 1.2  | [Task description] | [Verifiable completion criteria] | 1.1 | cc:TODO |
```

**Purpose line**:
- Auto-populated only when the user explicitly states the phase's purpose
- Omit the entire Purpose line when no input is provided (do not leave an empty line)
- Must be a single line (multi-line prohibited)

**DoD (Definition of Done) notation**:
- Write in one verifiable line (e.g., "Tests pass", "Migration is runnable", "lint errors 0")
- "Looks good" or "works properly" is prohibited. Must be decidable as Yes/No

**Depends notation**:
- No dependency: `-`
- Single dependency: Task number (e.g., `1.1`)
- Multiple dependencies: Comma-separated (e.g., `1.1, 1.2`)
- Phase dependency: Phase number (e.g., `Phase 1`)

### Team mode output

Only when the user explicitly requests team mode, provide issue bridge dry-run guidance alongside Plans.md.

- Only a single tracking issue
- List sub-issue payloads per task
- Plans.md remains the source of truth
- Provide guidance in a form that can directly use `scripts/plans-issue-bridge.sh --team-mode` dry-run output

## Step 7: Next Action Guidance

> Plans.md complete!
>
> Next steps:
> - Start implementation with `harness-work`
> - Or say "start from Phase 1"
> - Add features with `harness-plan add [feature name]`
> - Defer features with `harness-plan update [task] blocked`

## CI Mode (--ci)

No hearing. Uses existing Plans.md as-is and only performs task decomposition.

1. Read Plans.md
2. List cc:TODO tasks in priority order
3. Mark parallelizable tasks with `[P]`
4. Propose next execution tasks
