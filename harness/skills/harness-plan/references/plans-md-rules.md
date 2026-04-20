# Plans.md Rules

Ordering conventions, field definitions, and behavioral requirements for Plans.md.
For the canonical template structure, see [plans-md-template.md](${CLAUDE_SKILL_DIR}/references/plans-md-template.md).

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

**Never append a new phase at the bottom.** The non-ascending order (highest phase number nearest the top) is enforced by `harness/scripts/plans-format-check.sh`.

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

Use the complete template from [plans-md-template.md](${CLAUDE_SKILL_DIR}/references/plans-md-template.md). Start with Phase 1 at the top. Include the `## Archive` footer with an empty table (or omit the table body if no archives yet).
