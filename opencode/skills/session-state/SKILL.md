---
name: session-state
description: "Auto-triggered by the orchestration system at /work phase boundaries. Do NOT load for: user-facing session management, login state, app state handling, or direct user requests. Internal workflow skill — manages session state transitions per SESSION_ORCHESTRATION.md, escalated transitions on error, and initialized restoration on session resume."
allowed-tools: ["Read", "Bash"]
user-invocable: false
---

# Session State Skill

An internal skill that manages session state transitions.
Validates and executes transitions according to the state machine defined in `docs/SESSION_ORCHESTRATION.md`.

## Feature Details

| Feature | Details |
|---------|---------|
| **State transitions** | See [references/state-transition.md](${CLAUDE_SKILL_DIR}/references/state-transition.md) |

## When to Use

- State updates at `/work` phase boundaries
- `escalated` transition on error
- `stopped` transition at session end
- `initialized` restoration on session resume

## Notes

- This skill is for internal use only
- Not intended to be invoked directly by users
- State transition rules are defined in `docs/SESSION_ORCHESTRATION.md`
