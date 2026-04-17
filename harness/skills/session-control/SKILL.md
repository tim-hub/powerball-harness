---
name: session-control
description: "Internal skill for --resume and --fork workflow boundaries in harness-work. Auto-triggered by orchestration."
when_to_use: "resume session, fork session"
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

1. Verify variables passed from the workflow (`resume_session_id`, `resume_latest`, `fork_session_id`, `fork_reason`)
2. Execute `${CLAUDE_SKILL_DIR}/../../scripts/session-control.sh` with the appropriate arguments (see [references/session-control.md](${CLAUDE_SKILL_DIR}/references/session-control.md) for argument logic)
3. Confirm updates to `.claude/state/session.json` and `.claude/state/session.events.jsonl`
