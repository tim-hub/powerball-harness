---
name: harness-brainstorming
description: "Runs interactive brainstorming to capture intent and produce a design spec, then hands off to harness-plan to generate Plans.md. Use when a rough idea needs both shaping and a concrete task list."
when_to_use: "rough idea to plan, interview then plan, design before plan, brainstorm and plan, spec then Plans.md, /harness-brainstorming, fuzzy requirements"
allowed-tools: ["Read", "Write", "Edit", "Bash", "Grep", "Glob", "Task", "WebSearch"]
argument-hint: "[topic|--from-spec <path>]"
effort: high
model: opus
---

# Harness Brainstorming

Two-stage orchestrator: shape the idea through dialogue, then convert the agreed design into Plans.md tasks.

## Quick Reference

| Stage | Skill invoked | Output |
|-------|---------------|--------|
| 1. Shape | `superpowers-extended-cc:brainstorming` | Design spec at `docs/superpowers/specs/YYYY-MM-DD-<topic>-design.md` |
| 2. Plan  | `harness-plan` (subcommand `create`)     | Tasks appended to `Plans.md` |

## When to Use This Skill Instead of Either One Alone

- **Use `harness-brainstorming`** when the idea is still rough — purpose, constraints, or success criteria are unclear, AND you want the result tracked in `Plans.md`.
- Use `superpowers-extended-cc:brainstorming` alone when you only need a spec (no `Plans.md` tracking).
- Use `harness-plan create` alone when requirements are already clear and you just need tasks.

## Execution Flow

### Stage 1 — Brainstorming

Invoke the `superpowers-extended-cc:brainstorming` skill and follow its checklist verbatim:

1. Explore project context (files, docs, recent commits)
2. Offer the visual companion if upcoming questions are visual
3. Ask clarifying questions — one at a time
4. Propose 2-3 approaches with trade-offs
5. Present the design section by section, getting approval per section
6. Write the spec to `docs/superpowers/specs/YYYY-MM-DD-<topic>-design.md` and commit
7. Spec self-review (placeholders, contradictions, ambiguity, scope)
8. User reviews the written spec
9. **Redirect terminal state — see Stage 2 below**

**Terminal-state override:** `brainstorming`'s default terminal step is "Invoke writing-plans skill". In `harness-brainstorming` we replace that step with `harness-plan create` (Stage 2). Do **not** invoke `writing-plans`, `frontend-design`, `mcp-builder`, or any other implementation skill.

### Stage 2 — Plan Generation

Once the user approves the spec:

1. State the spec path back to the user (e.g., `docs/superpowers/specs/2026-05-15-foo-design.md`).
2. Invoke the `harness-plan` skill with subcommand `create`.
3. At `create` Step 0 ("Check Conversation Context"), choose option **1 — "From the preceding conversation"**. The approved spec plus the brainstorming dialogue serve as the requirements input — do not run a second hearing.
4. `harness-plan create` produces `Plans.md` with phase headers, DoD inference, Depends inference, and TDD markers per its own rules.

## Handoff Contract

| What flows from Stage 1 → Stage 2 | Where it lives |
|-----------------------------------|----------------|
| Approved design spec | `docs/superpowers/specs/<date>-<topic>-design.md` (committed) |
| Requirements, constraints, decisions | Current conversation — consumed by `harness-plan create` Step 0 |
| Feature list and priorities | Re-extracted by `harness-plan create` Steps 4-5 |

The spec file is the durable artifact; the conversation is the bridge. Both must exist before Stage 2 starts.

## Skip Conditions

Skip Stage 1 entirely when:

- The user passes `--from-spec <path>`. Read the spec, summarize it back for confirmation, then go straight to Stage 2.
- An approved spec from a prior session already exists and the user references it.

In both cases, Stage 2 still runs.

## Anti-Patterns

- **Don't** invoke `writing-plans` after Stage 1 — replacing it with `harness-plan` is the whole point of this skill.
- **Don't** ask Plans.md-shaped questions ("how many phases?", "what's the DoD?") during Stage 1. Stage 1 is about *what to build*; Stage 2 is about *how to track it*.
- **Don't** write to `Plans.md` directly. Always delegate to `harness-plan create` so its markers, DoD inference, and TDD logic apply.
- **Don't** combine the visual-companion offer with clarifying questions. `brainstorming` requires the offer to be its own message.

## Related Skills

- `superpowers-extended-cc:brainstorming` — Stage 1 engine (interview + spec writing)
- `harness-plan` — Stage 2 engine (subcommand `create`)
- `harness-work` — Run the generated tasks after `Plans.md` exists
- `harness-setup` — Initialize a project that has no `Plans.md` yet
