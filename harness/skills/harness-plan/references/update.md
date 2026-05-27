# update Subcommand — Update Marker

Changes a task's status marker in Plans.md.

## Syntax

```
harness-plan update [task-name|task-number] [WIP|done|blocked|TODO]
```

## Marker Mapping

| Command | Marker |
|---------|--------|
| `WIP` | `cc:WIP` |
| `done` | `cc:done` |
| `blocked` | `blocked` |
| `TODO` | `cc:TODO` |

For full status marker semantics — including the `cc:done [hash]` artifact notation and the required reason annotation for `blocked` — see [plans-md-rules.md](${CLAUDE_SKILL_DIR}/references/plans-md-rules.md) ("Status markers" subsection).

## Flow

1. Locate the target row in Plans.md by task name or task number (e.g. `2.3`).
   - Ambiguity → list candidates and ask the user.
   - Not found → abort with a clear "task not found" message.
2. Replace the **Status** cell with the mapped marker.
3. For `blocked`, prompt the user for a one-line reason and append it in parentheses:
   `blocked (waiting on external API spec)`
4. For `done`, optionally append the short git hash if a recent commit can be unambiguously attributed to this task: `cc:done [a1b2c3d]`.
5. Save Plans.md without reordering rows.

## Notes

- Marker updates **never** change phase ordering. Use `harness-plan archive` to remove fully completed phases.
- Use `harness-plan sync` (not `update`) when discrepancy detection across many tasks is needed — `update` is for a single deliberate marker change.

## Agent Delegation

This subcommand can be delegated to the `harness-planner` agent (Haiku, low effort) when a caller (skill or agent) wants to mark a task without running the skill flow inline.

Request shape (`planner-request.v1`):

```json
{
  "schema_version": "planner-request.v1",
  "operation": "update",
  "task_id": "98.3",
  "marker": "done",
  "commit_hash": "a1b2c3d"
}
```

For `marker: "blocked"` the caller must also supply a `reason` field. See [`harness/agents/harness-planner.md`](${CLAUDE_SKILL_DIR}/../../agents/harness-planner.md).
