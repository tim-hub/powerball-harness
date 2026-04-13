---
description: Quick Reference table format for SKILL.md files with multiple subcommands
paths: "skills/**/SKILL.md"
---

# Skill Quick Reference Format

Rules for the **Quick Reference** section in `skills/<skill>/SKILL.md` files.

## When This Rule Applies

Both conditions must be true:

1. The skill declares **multiple subcommands** in its frontmatter `argument-hint` (e.g. `"[create|update|sync|add]"`)
2. The skill includes a `## Quick Reference` section

If the skill has only a single subcommand, or has no Quick Reference section, this rule does not apply.

## Required Table Format

Use a **3-column table** with these exact headers, in this exact order:

```markdown
## Quick Reference

| User Input | Subcommand | Behavior |
|------------|------------|----------|
| <what the user types or says> | <subcommand token from argument-hint> | <what happens, including pointer to reference file if any> |
```

### Column semantics

| Column | Contents |
|--------|----------|
| **User Input** | Literal command (`harness-review code`), trigger phrase (`"Review this"`), or flag (`harness-review --dual`). Use backticks for commands/flags, quotes for phrases. |
| **Subcommand** | The matching token from `argument-hint`. Mark auto-detected dispatch as `` `code` (auto) ``. Combinations use `+` (e.g. `` `code` (auto) + Codex parallel ``). |
| **Behavior** | One-line description of what happens. If the behavior delegates to a reference file, name the file in this column. |

### Linking to references in the Behavior column

When a row's behavior is documented in a `references/*.md` file, mention the file inline. Examples:

```markdown
| `harness-plan sync`      | `sync`    | Progress check (see `references/sync.md`)          |
| `harness-plan create`    | `create`  | Interactive plan creation (see `references/create.md`) |
```

The reference filename is enough — no need to wrap in a full markdown link inside the table cell (tables stay readable that way).

## Placement

The Quick Reference section should appear **near the top of SKILL.md**, immediately after the skill's one-paragraph overview and before deeper sections like "Subcommand Details" or "Feature Details". This matches the convention used in `harness-plan`, `harness-work`, `harness-review`, `harness-setup`, and `harness-sync`.

## Prohibited

- 2-column tables (`User Input | Behavior`) when the skill has multiple subcommands — the Subcommand column must be visible so the token binding is explicit
- Column header renames (`Trigger` instead of `User Input`, `Action` instead of `Behavior`, etc.) — keep wording uniform across skills
- Putting MCP tools, external CLIs, or non-dispatch capabilities in the Subcommand column — that column is strictly for tokens listed in `argument-hint`. Non-dispatch capabilities belong in a separate section below Quick Reference.

## Rationale

The Quick Reference table doubles as a routing hint when a skill is auto-loaded from its `description`. Keeping `argument-hint` tokens, user-facing input, and behavior visible in one place lets the model dispatch without having to open reference files. Uniform column naming across skills makes this instant-readable.

## Related

- [skill-editing.md](./skill-editing.md) — general SKILL.md editing rules (frontmatter, file size, references/)
