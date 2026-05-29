# archive Subcommand — Phase Archiving

Marks fully-completed phases as archived in the plan (`harness plan-cli archive <phase-id>`) to keep the active view lean. Archived phases stay in `.claude/harness/plans.json` with `status: archived` — there is **no separate archive markdown file**. A phase is eligible when every task in it has a `cc:done` or `pm:confirmed` marker.

## Flow

1. Query phases via `harness plan-cli list` and identify those where **all** tasks are `cc:done` / `pm:confirmed`.
2. Apply the retention rule below to decide which eligible phases actually archive.
3. For each phase to archive, run `harness plan-cli archive <phase-id>` (sets `status: archived` in plans.json).
4. Verify the result with `harness plan-cli list --status archived`.

## What Stays Active

- Any phase with at least one task that is `cc:TODO`, `cc:WIP`, or `blocked`.
- The **10 most recent completed phases**, even if they are fully `cc:done` / `pm:confirmed`, to maintain recent history and context.

## Archived Phases

Archived phases are soft-deleted: they remain in `.claude/harness/plans.json` with `status: archived` and can be listed with `harness plan-cli list --status archived` (or `--status all`). No external archive markdown file is written.

## References

- [task-fields.md](${CLAUDE_SKILL_DIR}/references/task-fields.md) — field definitions (DoD, Depends, Status markers)
- [session-log.md](${CLAUDE_SKILL_DIR}/references/session-log.md) — sibling subcommand for archiving session-log.md by month

## Agent Delegation

This subcommand can be delegated to the `harness-planner` agent (Haiku, low effort). The agent applies the same retention rule and footer update.

Request shape (`planner-request.v1`):

```json
{
  "schema_version": "planner-request.v1",
  "operation": "archive"
}
```

When no phases are eligible, the planner returns `status: "skipped"`. See [`harness/agents/harness-planner.md`](${CLAUDE_SKILL_DIR}/../../agents/harness-planner.md).
