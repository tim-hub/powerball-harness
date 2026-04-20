# Plans.md Format Reference

Canonical structure and ordering conventions for Plans.md.
All `harness-plan` subcommands (`create`, `add`, `archive`) must produce output that conforms to this template.

---

## Complete Template

```markdown
# [Project Name] — Plans.md

Last archive: YYYY-MM-DD (Phase X–Y → `.claude/memory/archive/Plans-YYYY-MM-DD-phaseX-Y.md`)
Last release: vA.B.C on YYYY-MM-DD (short description)

---

## Phase N: [Phase Name — newest phase always here]

Created: YYYY-MM-DD

**Goal**: One paragraph explaining what this phase achieves and why.

**Depends on**: Phase M (reason). Omit section if no dependency.

**Non-goals**: What this phase explicitly does NOT do. Omit section if not needed.

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| N.1  | What to implement | Verifiable yes/no condition | - | cc:TODO |
| N.2  | What to implement | Verifiable yes/no condition | N.1 | cc:TODO |
| N.3  | What to implement | Verifiable yes/no condition | N.1, N.2 | cc:TODO |

---

## Phase N-1: [Older Phase]

...

---

## Phase N-2: [Even Older Phase]

...

---

## Future Considerations

(none currently)

---

## Archive

Older phases have been moved to `.claude/memory/archive/` to keep this file lean.

| Archive file | Phases | Date |
|---|---|---|
| [Plans-YYYY-MM-DD-phaseX-Y.md](.claude/memory/archive/Plans-YYYY-MM-DD-phaseX-Y.md) | Phase X–Y | YYYY-MM-DD |
```

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
After every `harness-plan archive` run, update the archive table with a new row for the archived file.

---

## Field Definitions

### Header block (required)

```
# [Project Name] — Plans.md

Last archive: YYYY-MM-DD (Phase X–Y → path)
Last release: vA.B.C on YYYY-MM-DD (description)

---
```

- `Last archive`: updated by `harness-plan archive` after each run
- `Last release`: updated by `harness-release` after each release

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

Older phases have been moved to `.claude/memory/archive/` to keep this file lean.

| Archive file | Phases | Date |
|---|---|---|
| [Plans-2026-04-18-phase62-73.md](.claude/memory/archive/Plans-2026-04-18-phase62-73.md) | Phase 62–73 | 2026-04-18 |
```

Add one row per archive file, newest archive at the top of the table.

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
4. Update `Last archive:` in the header
5. Add a new row to the `## Archive` footer table (newest at top)
6. Verify remaining phases are still non-ascending after removal

### `create` — generate new Plans.md

Use the complete template above. Start with Phase 1 at the top. Include the `## Archive` footer with an empty table (or omit the table body if no archives yet).
