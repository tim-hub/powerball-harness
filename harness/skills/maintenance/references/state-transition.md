---
name: state-transition
description: "Execute session state transitions using session-state.sh"
allowed-tools: [Read, Bash]
---

# State Transition

Execute session state transitions.

## Input

Workflow variables:
- `target_state` (string): Target state to transition to
- `event_name` (string): Trigger event
- `event_data` (string, optional): Additional event data (JSON)

## Valid States

| State | Description |
|-------|-------------|
| `idle` | Session not started |
| `initialized` | SessionStart completed |
| `planning` | Preparing for Plan/Work |
| `executing` | /work in progress |
| `reviewing` | Review in progress |
| `verifying` | Build/test in progress |
| `escalated` | Awaiting human confirmation |
| `completed` | Deliverables finalized |
| `failed` | Unrecoverable |
| `stopped` | Stop hook reached |

## Typical Transitions

| From | Event | To |
|------|-------|----|
| idle | session.start | initialized |
| initialized | plan.ready | planning |
| planning | work.start | executing |
| executing | work.task_complete | reviewing |
| reviewing | verify.start | verifying |
| verifying | verify.passed | completed |
| verifying | verify.failed | escalated |
| * | session.stop | stopped |
| stopped | session.resume | initialized |

## Execution

```bash
./scripts/session-state.sh --state <state> --event <event> [--data <json>]
```

### Example: Transition to executing state

```bash
./scripts/session-state.sh --state executing --event work.start
```

### Example: Escalation (with data)

```bash
./scripts/session-state.sh --state escalated --event escalation.requested \
  --data '{"reason":"Build failed 3 times","retry_count":3}'
```

## Expected Results

- The `state`, `updated_at`, `last_event_id`, and `event_seq` fields in `.claude/state/session.json` are updated
- An event is appended to `.claude/state/session.events.jsonl`
- Invalid transitions produce an error on stderr + non-zero exit

## Error Handling

When a transition fails (e.g., invalid transition):
1. Output the current state and allowed transitions to stderr
2. Return a non-zero exit code
3. The caller (workflow) handles the escalation
