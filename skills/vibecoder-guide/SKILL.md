---
name: vibecoder-guide
description: "Use this skill when the user appears non-technical, asks 'what should I do next?', 'how does this work?', seems confused about the development process, or needs step-by-step guidance in plain language. Do NOT load for: experienced developers making direct implementation requests, code reviews, or technical debugging. Guides non-technical users (VibeCoders) through natural language development — explains what to do next, how the system works, and how to express requirements."
allowed-tools: ["Read"]
user-invocable: false
---

# VibeCoder Guide Skill

A skill that guides VibeCoders (non-technical users) through development using only natural language.
Automatically responds to questions like "What should I do?" or "What's next?"

---

## Trigger Phrases

This skill is automatically triggered by the following phrases:

- "What should I do?", "What can I do?"
- "What should I do next?", "What's next?"
- "What's possible?", "What should I work on?"
- "I'm stuck", "I don't understand", "Help"
- "Show me how to use this"
- "what should I do?", "what's next?", "help"

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
