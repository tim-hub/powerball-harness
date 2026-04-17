---
name: vibecoder-guide
description: "Plain-language workflow guidance for users unfamiliar with the Harness development cycle. Use when explaining what to do next or getting someone started."
when_to_use: "what do I do next, explain workflow, getting started, harness guide, how does this work"
allowed-tools: ["Read"]
user-invocable: false
---

# VibeCoder Guide Skill

Guides non-technical users (VibeCoders) through development using only natural language. Responds to general orientation questions ("What should I do?", "What's next?", "Help") with context-aware suggestions about what to work on next.

> **Distinction from `session-init`**: This skill provides general orientation and onboarding guidance for users unfamiliar with the harness workflow. Use `session-init` when explicitly starting a new work session (e.g., beginning a day's work with a known plan).

---

## Overview

VibeCoders can find out their next action just by asking in plain language,
without needing to know technical commands or workflows.

---

## Response Patterns

### Pattern 1: No Project Exists

> 🎯 **Let's start a project first!**
>
> **Example phrases:**
> - "I want to build a blog"
> - "I want to create a task management app"
> - "I want to make a portfolio site"
>
> A rough idea is fine. Just tell me what you want to do.

### Pattern 2: Plans.md Exists but No In-Progress Tasks

> 📋 **There's a plan. Let's start working!**
>
> **Current plan:**
> - Phase 1: Foundation setup
> - Phase 2: Core features
> - ...
>
> **Example phrases:**
> - "Start phase 1"
> - "Do the first task"
> - "Do everything"

### Pattern 3: Task In Progress

> 🔧 **Work in progress**
>
> **Current task:** {{task name}}
> **Progress:** {{completed}}/{{total}}
>
> **Example phrases:**
> - "Continue"
> - "Next task"
> - "How far along are we?"

### Pattern 4: After Phase Completion

> ✅ **Phase complete!**
>
> **What you can do next:**
> - "Check it works" -> Start the dev server
> - "Review it" -> Code quality check
> - "Next phase" -> Start the next phase of work
> - "Commit it" -> Save the changes

### Pattern 5: When an Error Occurs

> ⚠️ **A problem occurred**
>
> **Situation:** {{error summary}}
>
> **Example phrases:**
> - "Fix it" -> Attempt auto-fix
> - "Explain it" -> Explain the problem in detail
> - "Skip it" -> Move to the next task

---

## Common Phrase Reference Table

| What You Want to Do | How to Say It |
|--------------------|---------------|
| Start a project | "I want to build XX" |
| View the plan | "Show me the plan", "What's the status?" |
| Start working | "Start", "Build it", "Do phase 1" |
| Continue | "Continue", "Next" |
| Test it | "Run it", "Show me" |
| Review code | "Review it", "Check it" |
| Save | "Commit it", "Save it" |
| When stuck | "What should I do?", "Help" |
| Leave it all to you | "Do everything", "You handle it" |

---

## Context Determination

This skill checks the following to select the appropriate response:

1. **Existence of AGENTS.md** -> Whether the project has been initialized
2. **Contents of Plans.md** -> Whether a plan exists, progress status
3. **Current task state** -> Presence of `cc:WIP` marker
4. **Recent errors** -> Whether a problem has occurred

---

## Implementation Notes

When this skill is triggered:

1. Analyze the current state
2. Select the appropriate pattern
3. Present specific "example phrases"
4. Wait for the user's next action

**Important**: Avoid technical jargon and explain in plain, simple language
