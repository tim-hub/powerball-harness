---
name: workflow-guide
description: "Use this skill when the user asks how the 2-agent workflow operates, wants to understand the collaboration process, or needs guidance on Cursor/CC roles and responsibilities. Do NOT load for: actual implementation work, executing handoffs (use cc-cursor-cc instead), or workflow configuration setup. Reference guide for Cursor ↔ Claude Code 2-agent collaboration workflow — explains roles, handoff patterns, and process flow."
allowed-tools: ["Read"]
user-invocable: false
---

# Workflow Guide Skill

A skill that provides guidance on the Cursor ↔ Claude Code 2-agent workflow.

---

## Trigger Phrases

This skill is triggered by the following phrases:

- "Tell me about the workflow"
- "How do I collaborate with Cursor?"
- "Explain the work process"
- "How should I proceed?"
- "how does the workflow work?"
- "explain 2-agent workflow"

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
│  - Request work from Claude Code (/handoff-to-claude)  │
│  - Review completion reports                           │
│  - Decide on production deploys                        │
└─────────────────────┬───────────────────────────────────┘
                      │ Task request
                      ▼
┌─────────────────────────────────────────────────────────┐
│                  Claude Code (Worker)                   │
│  - Execute tasks with /work (supports parallel)        │
│  - Implement -> Test -> Commit                         │
│  - Auto-fix on CI failure (up to 3 times)              │
│  - Report completion with /handoff-to-cursor           │
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

## Key Commands

### Claude Code Side

| Command | Purpose |
|---------|---------|
| `/harness-init` | Project setup |
| `/plan-with-agent` | Planning and task breakdown |
| `/work` | Task execution (supports parallel) |
| `/handoff-to-cursor` | Completion report (to Cursor PM) |
| `/sync-status` | Status check |

### Skills (Auto-triggered in Conversation)

| Skill | Trigger Example |
|-------|----------------|
| `handoff-to-pm` | "Report completion to PM" |
| `handoff-to-impl` | "Hand off to the implementer" |

### Cursor Side (Reference)

| Command | Purpose |
|---------|---------|
| `/handoff-to-claude` | Request task from Claude Code |
| `/review-cc-work` | Review completion reports |

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

A: Ask Cursor for clarification, or use `/sync-status` to organize the current status.

### Q: What if CI keeps failing?

A: After 3+ failures, stop auto-fixing and escalate to Cursor.

---

## Related Documents

- AGENTS.md - Detailed role assignments
- CLAUDE.md - Claude Code specific settings
- Plans.md - Task management file
