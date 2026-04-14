---
name: session-state
description: "Use when orchestration system hits /work phase boundaries (auto-triggered). Internal workflow skill. Do NOT load for: user-facing session management, login state, or direct user requests."
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
