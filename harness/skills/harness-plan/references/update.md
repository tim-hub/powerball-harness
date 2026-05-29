# update Subcommand — Update Marker

Changes a task's status via `harness plan-cli update`. The SSOT is `.claude/harness/plans.json`; never edit a status cell in `Plans.md` markdown.

## Syntax

```
harness-plan update [task-name|task-number] [WIP|done|blocked|TODO]
```

## Marker Mapping

| Command | `--status` value |
|---------|------------------|
| `WIP` | `cc:WIP` |
| `done` | `cc:done` |
| `blocked` | `blocked` |
| `TODO` | `cc:TODO` |

For full status marker semantics — including the `cc:done [hash]` artifact notation and the required reason annotation for `blocked` — see [task-fields.md](${CLAUDE_SKILL_DIR}/references/task-fields.md) ("Status Markers" section).

## Flow

1. Locate the target task by name or task ID (e.g. `2.3`) using `harness plan-cli list`/`get`.
   - Ambiguity → list candidates and ask the user.
   - Not found → abort with a clear "task not found" message.
2. Apply the status change: `harness plan-cli update <task-id> --status <mapped-status>`.
3. For `blocked`, prompt the user for a one-line reason and pass it via `--reason` (required for `blocked`):
   `harness plan-cli update <task-id> --status blocked --reason "waiting on external API spec"`
4. For `done`, pass the short git hash via `--hash` (required for `cc:done`) when a recent commit can be unambiguously attributed to this task:
   `harness plan-cli update <task-id> --status cc:done --hash a1b2c3d`
5. The CLI performs an atomic write to `plans.json` without reordering.

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
