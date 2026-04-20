# Plans.md Template

Canonical template for Plans.md. Copy this structure when creating or regenerating Plans.md.
For ordering rules, field definitions, and what `harness-plan` must do, see [plans-md-rules.md](${CLAUDE_SKILL_DIR}/references/plans-md-rules.md).

---

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

(none currently — add forward-looking items, deferred scope, or open questions here)

---

## Archive

Older phases have been moved to `.claude/memory/archive/` to keep this file lean.

| Archive file | Phases | Date |
|---|---|---|
| [Plans-YYYY-MM-DD-phaseX-Y.md](.claude/memory/archive/Plans-YYYY-MM-DD-phaseX-Y.md) | Phase X–Y | YYYY-MM-DD |
```
