---
name: principles
description: "Use when consulting coding principles, development guidelines or safe-editing practices. Do NOT load for: direct implementation (harness-work), reviews, or setup."
allowed-tools: ["Read"]
user-invocable: false
---

# Principles Skills

Development principles and guidelines for coding, editing, and VibeCoder workflows.

## Feature Details

| Feature | Details |
|---------|--------|
| **General Principles** | See [references/general-principles.md](${CLAUDE_SKILL_DIR}/references/general-principles.md) |
| **Diff-Aware Editing** | See [references/diff-aware-editing.md](${CLAUDE_SKILL_DIR}/references/diff-aware-editing.md) |
| **Context Reading** | See [references/repo-context-reading.md](${CLAUDE_SKILL_DIR}/references/repo-context-reading.md) |

<!-- OPEN: vibecoder-guide.md may be redundant with the standalone vibecoder-guide skill (harness:vibecoder-guide). Recommend verifying content overlap and removing this reference file if the standalone skill supersedes it. -->

## Execution Steps

1. Classify the user's request
2. Read the appropriate reference file from "Feature Details" above
3. Reference and apply its contents
