---
name: workflow-guide
description: "Explains the Harness workflow — roles, execution modes, and process flow. Use when asking how the harness workflow operates."
when_to_use: "how does the workflow work, how does harness work, execution modes, workflow roles, solo vs parallel vs breezing"
allowed-tools: ["Read"]
user-invocable: false
---

# Workflow Guide Skill

A skill that provides guidance on the Claude Code + Harness workflow.

---

## Overview

This skill explains the execution modes and process flow for implementing tasks using Harness.

---

## Execution Modes

| Mode | When | How |
|------|------|-----|
| **Solo** | 1 task | Worker → Reviewer |
| **Parallel** | 2–3 independent tasks | Worker A + Worker B → Reviewer |
| **Breezing** | 4+ tasks or `--breezing` flag | Lead → Workers (worktree isolation) → Reviewer |
| **Codex** | `--codex` flag | Delegate to Codex CLI via codex-plugin-cc |

---

## Task Management with Plans.md

### Marker List

| Marker | Meaning | Set By |
|--------|---------|--------|
| `pm:requesting` | Requested by PM | PM agent or user |
| `cc:TODO` | Not yet started | Either |
| `cc:WIP` | Claude Code working on it | Claude Code |
| `cc:done` | Claude Code completed | Claude Code |
| `pm:confirmed` | PM confirmed complete | PM agent or user |
| `blocked` | Blocked | Either |

### Task State Transitions

```
pm:requesting -> cc:WIP -> cc:done -> pm:confirmed
```

---

## Standard Work Cycle

```
harness-plan → harness-work → harness-review → harness-release
```

1. **Plan**: Add tasks to Plans.md with `/harness-plan`
2. **Work**: Execute tasks with `/harness-work`
3. **Review**: Multi-angle quality check with `/harness-review`
4. **Release**: Version bump + tag + GitHub Release with `/release-this`

---

## Key Skills

| Skill | Purpose |
|-------|---------|
| `harness-setup init` | Project setup |
| `harness-plan` | Planning and task breakdown |
| `harness-work` | Task execution (solo / parallel / breezing) |
| `harness-review` | Code and plan review |
| `harness-sync` | Status check |
| `harness-release` | Release engine |

---

## CI/CD Rules

### Claude Code's Scope of Responsibility

- ✅ Up to staging deploy
- ✅ Auto-fix on CI failure (up to 3 times)
- ❌ Production deploy requires human approval

### 3-Strike Rule

When CI fails 3 consecutive times:
1. Stop auto-fix attempts
2. Generate an escalation report
3. Defer the decision to the user

---

## Related Documents

- AGENTS.md - Detailed role assignments
- CLAUDE.md - Claude Code specific settings
- Plans.md - Task management file
- [Typical workflow examples](${CLAUDE_SKILL_DIR}/examples/typical-workflow.md)
