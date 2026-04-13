---
name: session-control
description: "Auto-triggered by harness-work when session --resume/--fork flags are present. Do NOT load for: user-facing session management, login state, app state handling, or direct user requests. Internal workflow skill — controls session resume/fork(branch) for /work, updates session.json and session.events.jsonl."
allowed-tools: ["Read", "Bash", "Write", "Edit"]
user-invocable: false
---

# Session Control Skill

Switches session state based on the `--resume` / `--fork` flags of /work.

## Feature Details

| Feature | Details |
|---------|---------|
| **Session resume/fork** | See [references/session-control.md](${CLAUDE_SKILL_DIR}/references/session-control.md) |

## Execution Steps

1. Verify variables passed from the workflow
2. Execute `scripts/session-control.sh` with the appropriate arguments
3. Confirm updates to `session.json` and `session.events.jsonl`
