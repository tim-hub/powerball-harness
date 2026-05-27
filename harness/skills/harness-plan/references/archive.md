# archive Subcommand — Plans.md Archiving

Moves fully-completed phases out of Plans.md into `.claude/memory/archive/` to keep the active file lean. A phase is eligible when every task in it has a `cc:done` or `pm:confirmed` marker.

## Flow

1. Read Plans.md and identify phases where **all** tasks are `cc:done` / `pm:confirmed`.
2. Apply the retention rule below to decide which eligible phases actually move.
3. Write the archived phases to `.claude/memory/archive/Plans-YYYY-MM-DD-phaseX-Y.md` (using today's date and the range of archived phase numbers).
4. Remove those phases from Plans.md.
5. Update the `Last archive:` bullet in the `## Archive` footer at the bottom of Plans.md to record the date and archive filename.
6. Verify remaining phases are still non-ascending after removal.

## What Stays in Plans.md

- Any phase with at least one task that is `cc:TODO`, `cc:WIP`, or `blocked`.
- The **10 most recent completed phases**, even if they are fully `cc:done` / `pm:confirmed`, to maintain recent history and context.

## Naming Convention

`Plans-YYYY-MM-DD-phaseX-Y.md`

- `YYYY-MM-DD` — today's date
- `X` — lowest archived phase number
- `Y` — highest archived phase number

Example: `Plans-2026-04-15-phase35-48.md` (archived Phases 35 through 48 on 2026-04-15).

## Archive Footer Format

The `## Archive` section stays at the very bottom of Plans.md and is updated after each archive run. See [plans-md-rules.md](${CLAUDE_SKILL_DIR}/references/plans-md-rules.md) ("Archive footer" subsection) for the exact format and ordering verification rules.

## References

- [plans-md-rules.md](${CLAUDE_SKILL_DIR}/references/plans-md-rules.md) — ordering rules and footer format
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
