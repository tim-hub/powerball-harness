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

## Agent Delegation

This subcommand can be delegated to the `harness-planner` agent (Haiku, low effort) when the caller already has all content fields (DoD, Depends) materialized and just wants the row inserted.

Request shape (`planner-request.v1`):

```json
{
  "schema_version": "planner-request.v1",
  "operation": "add",
  "task_name": "Wire planner into worker sweep",
  "description": "Worker delegates Plans.md marker updates to harness-planner",
  "dod": "Worker SR-1 step emits planner-request.v1; Plans.md row updated by planner",
  "depends": "-",
  "phase": 110
}
```

Omit `phase` to create a new top phase. The planner will **not** invent missing fields — if `dod` or `depends` is absent the request is rejected with `status: "error"`. Auto-inference of DoD/Depends (Steps 3–4 above) remains a responsibility of the `harness-plan` skill running in the main session, not the agent. See [`harness/agents/harness-planner.md`](${CLAUDE_SKILL_DIR}/../../agents/harness-planner.md).
