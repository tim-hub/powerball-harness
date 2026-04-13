---
name: principles
description: "Use this skill when the user asks about coding principles, development guidelines, safe editing practices, or VibeCoder guidance. Also use when another skill needs to reference general development standards. Do NOT load for: actual code implementation (use harness-work instead), code reviews, or project setup. Reference for development principles, guidelines, and VibeCoder best practices — including diff-aware editing, repo context reading, and safety guardrails."
allowed-tools: ["Read"]
user-invocable: false
---

# Principles Skills

A collection of skills that provide development principles and guidelines.

## Feature Details

| Feature | Details |
|---------|--------|
| **General Principles** | See [references/general-principles.md](${CLAUDE_SKILL_DIR}/references/general-principles.md) |
| **Diff-Aware Editing** | See [references/diff-aware-editing.md](${CLAUDE_SKILL_DIR}/references/diff-aware-editing.md) |
| **Context Reading** | See [references/repo-context-reading.md](${CLAUDE_SKILL_DIR}/references/repo-context-reading.md) |
| **VibeCoder** | See [references/vibecoder-guide.md](${CLAUDE_SKILL_DIR}/references/vibecoder-guide.md) |

## Execution Steps

1. Classify the user's request
2. Read the appropriate reference file from "Feature Details" above
3. Reference and apply its contents
