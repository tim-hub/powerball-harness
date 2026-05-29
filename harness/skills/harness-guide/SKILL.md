---
name: harness-guide
description: "Orientation and workflow guidance for the Harness development cycle. Use when explaining how Harness works, what to do next, or getting started."
when_to_use: "how does this work, explain workflow, what do I do next, getting started, harness guide, execution modes, harness roles, vibecoder"
allowed-tools: ["Read"]
user-invocable: false
model: opus
effort: high
---

# Harness Guide

Provides orientation and workflow guidance for users at any experience level — from first-timers asking "what do I do?" to developers wanting to understand execution modes and agent roles.

## Quick Reference

| User Need | Guidance |
|-----------|---------|
| "What do I do next?" | See [Orientation Patterns](#orientation-patterns) below |
| "How does the workflow work?" | See [Workflow Overview](#workflow-overview) below |
| "What are the execution modes?" | See [references/workflow.md](${CLAUDE_SKILL_DIR}/references/workflow.md) |
| "Non-technical / plain language" | See [references/vibecoder.md](${CLAUDE_SKILL_DIR}/references/vibecoder.md) |

---

## Orientation Patterns

Check current state and respond with context-aware guidance:

### No Project Exists

> Let's start a project first! Tell me what you want to build — a rough idea is fine.
> Example: "I want to build a blog" → run `/harness-setup` then `/harness-plan create`

### Plans.md Exists, No In-Progress Tasks

> There's a plan. Let's start working!
> Say: "Start phase 1", "Do the first task", or "Do everything" → `/harness-work`

### Task In Progress

> Work in progress — current task: `{{task name}}` ({{completed}}/{{total}})
> Say: "Continue", "Next task", or "How far along are we?"

### After Phase Completion

> Phase complete! Options: "Check it works" (dev server), "Review it" (`/harness-review`), "Next phase" (`/harness-work`), "Commit it"

### When an Error Occurs

> A problem occurred: `{{error summary}}`
> Say: "Fix it" (auto-fix), "Explain it" (detail), or "Skip it" (next task)

**Context check**: Existence of AGENTS.md (project init), Plans.md (plan + progress), `cc:WIP` marker (in-progress task), recent errors.

---

## Workflow Overview

The Harness lifecycle: **Setup → Plan → Work → Review → Release**

```
/harness-setup    → creates CLAUDE.md, Plans.md
/harness-plan     → adds tasks with acceptance criteria
/harness-work     → implements (auto-selects solo/parallel/breezing)
/harness-review   → code review gate
/harness-release  → version bump, CHANGELOG, tag, GitHub Release
```

**Execution modes** (auto-selected by `/harness-work` based on task count):

| Mode | Tasks | Behavior |
|------|-------|---------|
| Solo | 1 | Worker → Reviewer |
| Parallel | 2–3 | Workers A+B → Reviewer |
| Breezing | 4+ | Lead → Workers → Reviewer (worktree isolation) |
| Codex | any | `--codex` flag — delegates to Codex CLI |

Full mode details and agent roles: [references/workflow.md](${CLAUDE_SKILL_DIR}/references/workflow.md)

Plain-language phrase guide: [references/vibecoder.md](${CLAUDE_SKILL_DIR}/references/vibecoder.md)
