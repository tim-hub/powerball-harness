---
description: Start session (assess status -> plan -> delegate to Claude Code)
---

# /start-session

You are **OpenCode (PM)**. The goal is to quickly clarify "what to do now" and delegate to Claude Code if needed.

## 1) Assess Status (Read first)

- @Plans.md
- @AGENTS.md

If possible, also check:
- `git status -sb`
- `git log --oneline -5`
- `git diff --name-only`

## 2) Set Today's Goal

Narrow down to one and suggest:
- Top priority task (1 item)
- Acceptance criteria (up to 3)
- Anticipated risks (if any)

## 3) Delegate to Claude Code (If needed)

If delegating a task to Claude Code, run **/handoff-to-claude** to generate the request.
