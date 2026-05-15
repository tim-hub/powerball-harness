# Plans.md Rules

Ordering conventions, field definitions, and behavioral requirements for Plans.md.
For the canonical template structure, see [plans-md-template.md](${CLAUDE_SKILL_DIR}/templates/plans-md-template.md).

---

## Ordering Rules

### 1. Newest phase on top — always

New phases are inserted **immediately after the `---` header separator**, above all existing `## Phase` blocks.

```
---            ← header separator

## Phase N+1   ← NEW phase goes here (top)

---

## Phase N    ← previous newest

---
```

**Never append a new phase at the bottom.** The non-ascending order (highest phase number nearest the top) is enforced by `${CLAUDE_PLUGIN_ROOT}/scripts/plans-format-check.sh`.

### 2. Non-ascending phase numbers (gaps allowed)

Phase numbers must decrease top-to-bottom. Gaps are allowed because archiving removes completed phases.

| Example | Valid? | Reason |
|---------|--------|--------|
| 79, 78, 77, 76 | ✅ | Strictly descending, no gaps |
| 79, 77, 74 | ✅ | Descending with gaps (archived phases removed) |
| 74, 75, 76 | ❌ | Ascending — violation |
| 79, 78, 80 | ❌ | 80 > 78 — violation |

### 3. Archive footer stays at the bottom

The `## Archive` section is the **last section** in Plans.md, below `## Future Considerations`.
After every `harness-plan archive` run, update the `Last archive:` bullet in the `## Archive` footer with the new date and archive filename.

### 4. Future Considerations section

Always present, even when empty. Use `(none currently)` as placeholder text — never omit the section or leave it blank.

---

## Field Definitions

### Header block (required)

```
# [Project Name] — Plans.md

Last release: vA.B.C on YYYY-MM-DD (description)

---
```

- `Last release`: updated by `harness-release` after each release
- `Last archive`: lives in the `## Archive` footer (bottom of file), updated by `harness-plan archive` after each run

### Phase block (one per phase)

```markdown
## Phase N: Short Title

Created: YYYY-MM-DD

**Goal**: ...

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| N.1  | ...         | ... | -       | cc:TODO |
```

**DoD (Definition of Done)**: Must be a verifiable yes/no condition. Banned phrases: "looks good", "works properly", "is done", "is complete".

**Depends**: Use `-` for no dependency, `N.1` for a single task, `N.1, N.2` for multiple, `Phase N` for a full phase dependency.

**Status markers**:

| Marker | Meaning |
|--------|---------|
| `cc:TODO` | Not started |
| `cc:WIP` | In progress |
| `cc:done [hash]` | Worker completed (include short git hash) |
| `pm:confirmed` | PM review confirmed |
| `blocked` | Blocked — **always add reason in parentheses** |

**Quality markers** (appear inline in the Description column):

| Marker | Meaning |
|--------|---------|
| `[needs-spike]` | High Impact × High Risk — spike task required before implementation |
| `[skip:tdd]` | Skip TDD phase (docs, config, style, trivial changes) |
| `[feature:security]` | Auth, login, payment tasks — security review required |
| `[feature:a11y]` | UI/screen/component tasks — accessibility review required |
| `[bugfix:reproduce-first]` | Bug tasks — must reproduce before fixing |
| `[ralph]` | Iterative loop task — executed by `harness-ralph-loop` (see below) |

**`[ralph]` tasks — extra per-task lines**:

`[ralph]` tasks require two lines below the task row (not additional columns):

```markdown
| N.2  | Fix flaky tests until passing [ralph] | All tests pass | N.1 | cc:TODO |
Verify: npm test
MaxIter: 15
```

- `Verify:` (required) — bash command that must exit 0 for the loop to succeed. Auto-inferred from project type if omitted at creation.
- `MaxIter:` (optional) — max worker iterations. Default: 10.

See [references/ralph-tasks.md](${CLAUDE_SKILL_DIR}/references/ralph-tasks.md) for full format and serialization rules.

### Archive footer (required, always last)

```markdown
## Archive

- Last archive: YYYY-MM-DD (Phase X–Y → `.claude/memory/archive/Plans-YYYY-MM-DD-phaseX-Y.md`)
- Other older phases have been moved to `.claude/memory/archive/` to keep this file lean.
```

---

## What `harness-plan` Must Do

### `add` — insert new phase

1. Determine next phase number (highest existing + 1)
2. Insert the new phase block **after `---` header separator and before the first existing `## Phase` block**
3. Never append to the bottom

### `archive` — remove completed phases

1. Identify phases where all tasks are `cc:done` or `pm:confirmed`
2. Retain the 10 most recently completed phases in Plans.md
3. Write archived phases to `.claude/memory/archive/Plans-YYYY-MM-DD-phaseX-Y.md`
4. Update `Last archive:` in the footer
5. Verify remaining phases are still non-ascending after removal

### `create` — generate new Plans.md

Use the complete template from [plans-md-template.md](${CLAUDE_SKILL_DIR}/templates/plans-md-template.md). Start with Phase 1 at the top. Include the `## Archive` footer with an empty table (or omit the table body if no archives yet).

---

## Native task mirror rule

Plans.md is the SSOT for task status. The native Claude Code task list (managed via `TaskCreate` / `TaskUpdate`) is a **mirror only** — never authoritative.

**Contract**: any skill that writes `cc:WIP` or `cc:done` to a Plans.md row MUST also call `TaskUpdate` on the matching native task. The match is by **title prefix**: a native task whose title starts with the Plans.md task ID (e.g., `99.1`) is the mirror.

| Plans.md write | Required mirror call |
|----------------|----------------------|
| `cc:WIP` | `TaskUpdate(status="in_progress")` |
| `cc:done` / `cc:done [hash]` | `TaskUpdate(status="completed")` |
| `blocked` | no mirror — leave native task alone (visible signal that work paused) |
| `cc:TODO` | no mirror — let user manage TODO restarts themselves |

**Silent on no match**: if no native task matches the prefix, skip — TaskUpdate is mirror-only, never authoritative.

**Call sites covered by this contract**:

1. `harness-work` solo mode — `references/solo-mode.md` Step 2.5 (cc:WIP) and Step 11 (cc:done)
2. `harness-work` breezing mode — `references/breezing-mode.md` B-2 (cc:WIP) and B-5 (cc:done)
3. `harness-ralph-loop` — `references/loop-flow.md` SUCCESS terminal state (cc:done)
4. `harness-plan update` — `references/update.md` Step 6, when new marker is `done` or `WIP`

**End-of-scope sweep** (in `harness-work` SKILL.md, `## Native Task Reconciliation`): after the per-task loop completes, reconcile any remaining drift by running `TaskList` and comparing every native task's title prefix to its matching Plans.md row.
