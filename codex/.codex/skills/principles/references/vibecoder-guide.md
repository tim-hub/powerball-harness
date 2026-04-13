---
name: vibecoder-guide
description: "A skill that guides VibeCoders (non-technical users) to develop using natural language. Use when providing guidance for non-technical users."
allowed-tools: ["Read"]
---

# VibeCoder Guide

A skill that guides VibeCoders (non-technical users) to develop using only natural language.
Automatically responds to questions like "What should I do?", "What's next?", etc.

---

## Trigger Phrases

This skill auto-triggers on the following phrases:

- "What should I do?", "What do I do?"
- "What should I do next?", "What's next?"
- "What can I do?", "What should I work on?"
- "I'm stuck", "I don't understand", "Help"
- "Teach me how to use this"
- "what should I do?", "what's next?", "help"

---

## Overview

VibeCoders can find out their next action simply by asking questions in natural language,
without knowing technical commands or workflows.

---

## Response Patterns

### Pattern 1: No Project Exists

> **Let's start a project first!**
>
> **Example phrases:**
> - "I want to create a blog"
> - "I want to build a task management app"
> - "I want to make a portfolio site"
>
> A rough idea is fine. Just tell me what you want to do.

### Pattern 2: Plans.md Exists But No Tasks in Progress

> **You have a plan. Let's start working!**
>
> **Current plan:**
> - Phase 1: Foundation setup
> - Phase 2: Core features
> - ...
>
> **Example phrases:**
> - "Start Phase 1"
> - "Do the first task"
> - "Do everything"

### Pattern 3: Task in Progress

> **Work is in progress**
>
> **Current task:** {{task name}}
> **Progress:** {{completed}}/{{total}}
>
> **Example phrases:**
> - "Continue"
> - "Next task"
> - "How far along are we?"

### Pattern 4: After Phase Completion

> **Phase completed!**
>
> **What you can do next:**
> - "Check if it works" -> Start the dev server
> - "Review it" -> Code quality check
> - "Move to the next phase" -> Start next work
> - "Commit it" -> Save changes

### Pattern 5: When an Error Occurs

> **A problem occurred**
>
> **Situation:** {{error summary}}
>
> **Example phrases:**
> - "Fix it" -> Attempt automatic fix
> - "Explain it" -> Explain the problem in detail
> - "Skip it" -> Move to the next task

---

## Common Phrase Reference Table

| What You Want to Do | How to Say It |
|---------------------|---------------|
| Start a project | "I want to build [something]" |
| See the plan | "Show me the plan", "What's the current status?" |
| Start working | "Start", "Build it", "Do Phase 1" |
| Continue work | "Continue", "Next" |
| Check behavior | "Run it", "Show me" |
| Review code | "Review it", "Check it" |
| Save | "Commit it", "Save it" |
| When stuck | "What should I do?", "Help" |
| Leave it all to Claude | "Do everything", "Handle it" |

---

## Context Detection

This skill checks the following to select the appropriate response:

1. **Existence of AGENTS.md** -> Whether the project is initialized
2. **Contents of Plans.md** -> Whether a plan exists, progress status
3. **Current task state** -> Presence of `cc:WIP` markers
4. **Recent errors** -> Whether a problem has occurred

---

## Implementation Notes

When this skill is triggered:

1. Analyze the current state
2. Select the appropriate pattern
3. Present specific "example phrases"
4. Wait for the user's next action

**Important**: Avoid technical jargon; explain in plain, easy-to-understand language
