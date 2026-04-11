---
name: init-memory-ssot
description: "Initializes the project's SSOT memory (decisions/patterns) and optional session-log. Use during initial setup or when .claude/memory is not yet configured for a project."
allowed-tools: ["Read", "Write"]
---

# Init Memory SSOT

Initializes the **SSOT** files under `.claude/memory/`.

- `decisions.md` (SSOT for important decisions)
- `patterns.md` (SSOT for reusable solutions)
- `session-log.md` (Session log. Recommended for local use)

Detailed policy: `docs/MEMORY_POLICY.md`

---

## Execution Steps

### Step 1: Check for Existing Files

- `.claude/memory/decisions.md`
- `.claude/memory/patterns.md`
- `.claude/memory/session-log.md`

**Do not overwrite** files that already exist.

### Step 2: Initialize from Templates (Only for Missing Files)

Templates:

- `templates/memory/decisions.md.template`
- `templates/memory/patterns.md.template`
- `templates/memory/session-log.md.template`

Replace `{{DATE}}` with the current date (e.g., `2025-12-13`) when generating.

### Step 3: Completion Report

- List of created files
- Git policy (`decisions/patterns` are recommended for sharing, `session-log/.claude/state` are recommended for local use)
