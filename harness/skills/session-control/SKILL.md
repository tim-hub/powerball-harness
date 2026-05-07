---
name: session-control
description: "Internal skill for session state transitions and --resume/--fork workflow boundaries in harness-work. Auto-triggered by orchestration."
when_to_use: "resume session, fork session, phase boundary, save state, restore state, state transition"
allowed-tools: ["Read", "Bash", "Write", "Edit"]
user-invocable: false
---

# Session Control Skill

Internal skill consolidating two session management paths:
- **State transitions**: validates and executes state machine transitions (`idle → initialized → … → completed/stopped`)
- **Resume/fork**: switches session state based on `--resume` / `--fork` flags of `/work`

## Feature Details

| Feature | Details |
|---------|---------|
| **State transitions** | See [references/state-transition.md](${CLAUDE_SKILL_DIR}/references/state-transition.md) |
| **Session resume/fork** | See [references/session-control.md](${CLAUDE_SKILL_DIR}/references/session-control.md) |

## When to Use

- State updates at `harness-work` phase boundaries → use state-transition path
- `escalated` / `stopped` / `initialized` transitions on error or resume → use state-transition path
- `--resume` / `--fork` flags in `/work` → use session-control path

## Notes

- This skill is for internal use only; not intended to be invoked directly by users
- Consolidated from the former `session-state` (state machine) and `session-control` (resume/fork) skills
