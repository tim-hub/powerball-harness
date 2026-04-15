---
name: session-control
description: "Apply /work --resume/--fork flags by updating session state files."
allowed-tools: ["Read", "Bash", "Write", "Edit"]
---

# Session Control

## Input

Workflow variables:
- `resume_session_id` (string)
- `resume_latest` (boolean)
- `fork_session_id` (string)
- `fork_reason` (string)

## Execution

### 1) Determine Arguments
- resume:
  - `resume_latest == true` -> `--resume latest`
  - Otherwise, if `resume_session_id` is present -> `--resume <id>`
- fork:
  - If `fork_session_id` is present -> `--fork <id>`, otherwise `--fork current`
  - If `fork_reason` is present -> `--reason "<text>"`

### 2) Run Script
```bash
bash "${CLAUDE_SKILL_DIR}/../../scripts/session-control.sh" --resume <id|latest>
bash "${CLAUDE_SKILL_DIR}/../../scripts/session-control.sh" --fork <id|current> --reason "<text>"
```

## Expected Results
- `.claude/state/session.json` is updated
- A `session.resume` or `session.fork` event is appended to `.claude/state/session.events.jsonl`
- On error, the reason is output to stderr
