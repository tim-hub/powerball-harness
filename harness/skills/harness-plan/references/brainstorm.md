# brainstorm Subcommand — Idea → Spec → Plan

Two-stage orchestrator: shapes a rough idea into a design spec via dialogue, then converts the agreed design into Plans.md tasks.

## Quick Reference

| Stage | Skill invoked | Output |
|-------|---------------|--------|
| 1. Shape | `superpowers-extended-cc:brainstorming` | Design spec at `docs/superpowers/specs/YYYY-MM-DD-<topic>-design.md` |
| 2. Plan  | `harness-plan create`                   | Tasks appended to `Plans.md` |

## When to Use This Instead of `create` Alone

- **Use `harness-plan brainstorm`** when the idea is still rough — purpose, constraints, or success criteria are unclear, AND you want the result tracked in Plans.md.
- Use `harness-plan create` alone when requirements are already clear and you just need tasks.

## Skip Conditions

Skip Stage 1 when:
- The user passes `--from-spec <path>`. Read the spec, summarize it back for confirmation, then go straight to Stage 2.
- An approved spec from a prior session already exists and the user references it.

## Stage 1 — Brainstorming

Invoke the `superpowers-extended-cc:brainstorming` skill and follow its checklist verbatim:

1. Explore project context (files, docs, recent commits)
2. Offer the visual companion if upcoming questions are visual (as its own message)
3. Ask clarifying questions — one at a time
4. Propose 2–3 approaches with trade-offs
5. Present the design section by section, getting approval per section
6. Write the spec to `docs/superpowers/specs/YYYY-MM-DD-<topic>-design.md` and commit
7. Spec self-review (placeholders, contradictions, ambiguity, scope)
8. User reviews the written spec

**Terminal-state override:** `brainstorming`'s default terminal step is "Invoke writing-plans skill". Here we replace that step with `harness-plan create` (Stage 2). Do **not** invoke `writing-plans`, `frontend-design`, or any other implementation skill.

## Stage 2 — Plan Generation

Once the user approves the spec:

1. State the spec path back to the user (e.g., `docs/superpowers/specs/2026-05-15-foo-design.md`).
2. Invoke `harness-plan create`.
3. At `create` Step 0 ("Check Conversation Context"), choose option **1 — "From the preceding conversation"**. The approved spec plus the brainstorming dialogue serve as the requirements input — do not run a second hearing.
4. `harness-plan create` produces Plans.md with phase headers, DoD inference, Depends inference, and TDD markers per its own rules.

## Handoff Contract

| What flows Stage 1 → Stage 2 | Where it lives |
|-------------------------------|----------------|
| Approved design spec | `docs/superpowers/specs/<date>-<topic>-design.md` (committed) |
| Requirements, constraints, decisions | Current conversation — consumed by `harness-plan create` Step 0 |
| Feature list and priorities | Re-extracted by `harness-plan create` Steps 4–5 |

## Anti-Patterns

- **Don't** invoke `writing-plans` after Stage 1 — replacing it with `harness-plan create` is the whole point.
- **Don't** ask Plans.md-shaped questions ("how many phases?", "what's the DoD?") during Stage 1. Stage 1 is about *what to build*; Stage 2 is about *how to track it*.
- **Don't** write to Plans.md directly. Always delegate to `harness-plan create` so its markers, DoD inference, and TDD logic apply.
- **Don't** combine the visual-companion offer with clarifying questions — it must be its own message.
