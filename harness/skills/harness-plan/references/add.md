# add Subcommand — Add Task

Adds a new task (or new phase) to the plan (`.claude/harness/plans.json`) via `harness plan-cli`. The SSOT is `.claude/harness/plans.json`; never edit `Plans.md` markdown.

## Syntax

```
harness-plan add task-name: detailed description [--phase phase-number]
```

## Behavior

- New tasks are created with the `cc:TODO` status by default.
- **New phases**: `harness plan-cli add-phase` prepends, so the new phase lands newest-on-top automatically. **Never append at the bottom.**
- When `--phase N` is supplied and Phase `N` exists, the new task is added to Phase `N` via `harness plan-cli add-task N …` instead of creating a new phase.

## Flow

1. Parse `task-name`, `description`, and optional `--phase`.
2. Determine target phase (query existing state with `harness plan-cli list`):
   - If `--phase N` given and Phase N exists → `harness plan-cli add-task N --name "…" --dod "…" [--depends "…"] [--description "…"]`.
   - Otherwise → create a new phase first with `harness plan-cli add-phase --title "…" --goal "…"`, then add the task to the returned phase ID.
3. Auto-fill DoD using the inference logic in [create.md](${CLAUDE_SKILL_DIR}/references/create.md) (Step 6 — "DoD Auto-Inference Logic"); pass it via `--dod`.
4. Auto-fill Depends using the inference logic in [create.md](${CLAUDE_SKILL_DIR}/references/create.md) (Step 6 — "Depends Auto-Inference Logic"); pass it via `--depends`.
5. Issue the `harness plan-cli add-task` (and, if needed, `add-phase`) calls. Ordering is handled by the CLI.
6. Verify the result with `harness plan-cli list`.

## References

- [task-fields.md](${CLAUDE_SKILL_DIR}/references/task-fields.md) — field definitions (DoD, Depends, Status markers, quality markers)

## Agent Delegation

This subcommand can be delegated to the `harness-planner` agent (Haiku, low effort) when the caller already has all content fields (DoD, Depends) materialized and just wants the row inserted.

Request shape (`planner-request.v1`):

```json
{
  "schema_version": "planner-request.v1",
  "operation": "add",
  "task_name": "Wire planner into worker sweep",
  "description": "Worker delegates plan status updates to harness-planner",
  "dod": "Worker SR-1 step emits planner-request.v1; plan task updated by planner via harness plan-cli",
  "depends": "-",
  "phase": 110
}
```

Omit `phase` to create a new top phase. The planner will **not** invent missing fields — if `dod` or `depends` is absent the request is rejected with `status: "error"`. Auto-inference of DoD/Depends (Steps 3–4 above) remains a responsibility of the `harness-plan` skill running in the main session, not the agent. See [`harness/agents/harness-planner.md`](${CLAUDE_SKILL_DIR}/../../agents/harness-planner.md).
