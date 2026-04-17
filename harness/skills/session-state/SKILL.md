---
name: session-state
description: "Internal skill for harness-work phase boundary management. Auto-triggered by orchestration."
when_to_use: "phase boundary, save state, restore state"
allowed-tools: ["Read", "Bash"]
user-invocable: false
---

# Session State Skill

An internal skill that manages session state transitions.
Validates and executes transitions according to the state machine (idle → initialized → planning → executing → reviewing → verifying → completed/escalated/stopped).

## Feature Details

| Feature | Details |
|---------|---------|
| **State transitions** | See [state-transition.md](${CLAUDE_SKILL_DIR}/references/state-transition.md) |

## When to Use

- State updates at `harness-work` phase boundaries
- `escalated` transition on error
- `stopped` transition at session end
- `initialized` restoration on session resume

## Notes

- This skill is for internal use only
- Not intended to be invoked directly by users
- State transition rules and valid states are documented in [state-transition.md](${CLAUDE_SKILL_DIR}/references/state-transition.md)
