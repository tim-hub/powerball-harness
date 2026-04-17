---
name: workflow-guide
description: "Explains the 2-agent Cursor and Claude Code workflow — roles, handoffs, and process flow. Use when asking how the dual-agent workflow operates."
when_to_use: "how does the workflow work, Cursor and Claude Code, 2-agent workflow, handoff process, workflow roles"
allowed-tools: ["Read"]
user-invocable: false
---

# Workflow Guide Skill

A skill that provides guidance on the Cursor ↔ Claude Code 2-agent workflow.

---

## Overview

This skill explains the role assignments and collaboration methods between Cursor (PM) and Claude Code (Worker).

---

## 2-Agent Workflow

### Role Assignments

| Agent | Role | Responsibilities |
|-------|------|-----------------|
| **Cursor** | PM (Project Manager) | Task assignment, reviews, production deploy decisions |
| **Claude Code** | Worker | Implementation, testing, CI fixes, staging deploy |

### Workflow Diagram

```
┌─────────────────────────────────────────────────────────┐
│                    Cursor (PM)                          │
│  - Add tasks to Plans.md                               │
│  - Request work from Claude Code (cc-cursor-cc)        │
│  - Review completion reports                           │
│  - Decide on production deploys                        │
└─────────────────────┬───────────────────────────────────┘
                      │ Task request
                      ▼
┌─────────────────────────────────────────────────────────┐
│                  Claude Code (Worker)                   │
│  - Execute tasks with harness-work (supports parallel) │
│  - Implement -> Test -> Commit                         │
│  - Auto-fix on CI failure (up to 3 times)              │
│  - Report completion with cc-cursor-cc                 │
└─────────────────────┬───────────────────────────────────┘
                      │ Completion report
                      ▼
┌─────────────────────────────────────────────────────────┐
│                    Cursor (PM)                          │
│  - Review changes                                      │
│  - Verify staging behavior                             │
│  - Execute production deploy (after approval)          │
└─────────────────────────────────────────────────────────┘
```

---

## Task Management with Plans.md

### Marker List

| Marker | Meaning | Set By |
|--------|---------|--------|
| `pm:requesting` | Requested by PM (compat: cursor:requesting) | PM (Cursor/PM Claude) |
| `cc:TODO` | Not yet started by Claude Code | Either |
| `cc:WIP` | Claude Code working on it | Claude Code |
| `cc:done` | Claude Code completed | Claude Code |
| `pm:confirmed` | PM confirmed complete (compat: cursor:confirmed) | PM (Cursor/PM Claude) |
| `cursor:requesting` | (compat) Synonym for pm:requesting | Cursor |
| `cursor:confirmed` | (compat) Synonym for pm:confirmed | Cursor |
| `blocked` | Blocked | Either |

### Task State Transitions

```
pm:requesting -> cc:WIP -> cc:done -> pm:confirmed
```

---

## Key Skills

### Claude Code Side

| Skill | Purpose |
|-------|---------|
| `harness-setup init` | Project setup |
| `harness-plan` | Planning and task breakdown |
| `harness-work` | Task execution (supports parallel) |
| `cc-cursor-cc` | Completion report (to Cursor PM) or task handoff |
| `harness-sync` | Status check |

### Skills (Auto-triggered in Conversation)

| Skill | Trigger Example |
|-------|----------------|
| `cc-cursor-cc` | "Report completion to PM" |
| `harness-review` | "Review this code" |

### Cursor Side (Reference)

| Skill | Purpose |
|-------|---------|
| `cc-cursor-cc` | Request task from Claude Code |
| `harness-review` | Review completion reports |

---

## CI/CD Rules

### Claude Code's Scope of Responsibility

- ✅ Up to staging deploy
- ✅ Auto-fix on CI failure (up to 3 times)
- ❌ Production deploy is prohibited

### 3-Strike Rule

When CI fails 3 consecutive times:
1. Stop auto-fix attempts
2. Generate an escalation report
3. Defer the decision to Cursor

---

## Frequently Asked Questions

### Q: What if Cursor is not available?

A: Even when working solo, using Plans.md for task management is recommended.
Perform production deploys manually and carefully.

### Q: What if the task is unclear?

A: Ask Cursor for clarification, or use `harness-sync` to organize the current status.

### Q: What if CI keeps failing?

A: After 3+ failures, stop auto-fixing and escalate to Cursor.

---

## Related Documents

- AGENTS.md - Detailed role assignments
- CLAUDE.md - Claude Code specific settings
- Plans.md - Task management file
- [Typical workflow examples](${CLAUDE_SKILL_DIR}/examples/typical-workflow.md)
