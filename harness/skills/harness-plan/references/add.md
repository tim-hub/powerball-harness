# add Subcommand — Add Task

Adds a new task (or new phase) to Plans.md.

## Syntax

```
harness-plan add task-name: detailed description [--phase phase-number]
```

## Behavior

- New tasks are added with the `cc:TODO` marker.
- **Insertion point**: a new phase block goes immediately after the `---` header separator, above all existing `## Phase` blocks (newest phase on top). **Never append at the bottom.**
- When `--phase N` is supplied and Phase `N` exists, the new task is appended to Phase `N`'s task table instead of creating a new phase.

## Flow

1. Parse `task-name`, `description`, and optional `--phase`.
2. Determine target phase:
   - If `--phase N` given and Phase N exists → append row to that phase's task table.
   - Otherwise → create a new phase block with the next phase number (highest existing + 1).
3. Auto-fill DoD using the inference logic in [create.md](${CLAUDE_SKILL_DIR}/references/create.md) (Step 6 — "DoD Auto-Inference Logic").
4. Auto-fill Depends using the inference logic in [create.md](${CLAUDE_SKILL_DIR}/references/create.md) (Step 6 — "Depends Auto-Inference Logic").
5. Insert the new phase block (or row) per the ordering rules in [plans-md-rules.md](${CLAUDE_SKILL_DIR}/references/plans-md-rules.md).
6. Verify the file remains non-ascending after insertion.

## References

- [plans-md-template.md](${CLAUDE_SKILL_DIR}/templates/plans-md-template.md) — canonical phase-block structure
- [plans-md-rules.md](${CLAUDE_SKILL_DIR}/references/plans-md-rules.md) — ordering rules and field definitions (DoD, Depends, Status markers)
