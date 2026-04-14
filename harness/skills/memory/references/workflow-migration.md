---
name: migrate-workflow-files
description: "Migrates existing project AGENTS.md/CLAUDE.md/Plans.md to the new format by reviewing existing content, interactively confirming carry-over items, with backups and task-preserving Plans merge."
allowed-tools: ["Read", "Write", "Edit", "Bash"]
---

# Migrate Workflow Files (Interactive Merge)

## Purpose

**Updates the following files to the new format while respecting existing content** in projects already in operation.

- `AGENTS.md`
- `CLAUDE.md`
- `Plans.md`

Key points:

- **Interactively confirm carry-over information** (never silently discard / never silently overwrite)
- **Always create a backup** before making changes
- For `Plans.md`, follow the `merge-plans` approach to **update structure while preserving tasks**

---

## Prerequisites (Important)

To balance "safety on first application" with "intended behavior (new format),"
this skill proceeds in the order: **user consent -> backup -> generation -> diff review**.

---

## Input (Auto-detected within this skill)

- `project_name`: Inferred from `basename $(pwd)`
- `date`: `YYYY-MM-DD`
- Existence of existing files:
  - `AGENTS.md`
  - `CLAUDE.md`
  - `Plans.md`
- Reference templates for the new format:
  - `templates/AGENTS.md.template`
  - `templates/CLAUDE.md.template`
  - `templates/Plans.md.template`

---

## Execution Flow

### Step 0: Detection and Consent (Required)

1. Use `Read` to check for existing `AGENTS.md` / `CLAUDE.md` / `Plans.md`.
2. If they exist, confirm with the user:
   - **Is it okay to migrate (update to the new format)?**
   - Important: Migration **includes content reorganization** (= some rearrangement and wording changes may occur)

If user says NO:

- Abort this skill (do not rewrite anything)
- Instead, suggest safe operations such as "merge `.claude/settings.json` only"

### Step 1: Review Existing Content (Summary)

`Read` each file and extract the following, presenting a brief summary:

- **AGENTS.md**: Role assignments, handoff procedures, prohibited actions, environment/prerequisites
- **CLAUDE.md**: Important constraints (prohibited actions/permissions/branch policies), test procedures, commit conventions, operational rules
- **Plans.md**: Task structure, marker conventions, current WIP/requested tasks

### Step 2: Confirm Carry-over Items (Interactive)

Based on the summary, ask the user about items to **retain/adjust** (5-10 questions is sufficient):

- Constraints that must absolutely be kept (e.g., production deploy prohibition, specific directory restrictions, security requirements)
- Role assignment assumptions (Solo/2-agent)
- Branch workflow (main/staging, etc.)
- Representative test/build commands
- Plans marker conventions (reconcile if existing rules exist)

### Step 3: Create Backup (Required)

Backups are collected in `.claude-code-harness/backups/` within the project (often not wanted in git).

Example:

- `.claude-code-harness/backups/2025-12-13/AGENTS.md`
- `.claude-code-harness/backups/2025-12-13/CLAUDE.md`
- `.claude-code-harness/backups/2025-12-13/Plans.md`

Using `Bash` with `mkdir -p` and `cp` is fine.

### Step 4: Generate New Format (Merge)

#### 4-1. Plans.md (Task-preserving Merge)

Execute following the `merge-plans` approach:

- Preserve existing 🔴🟡🟢📦 tasks
- Update marker legend and last-updated info from the template
- If unparseable, keep the backup and adopt the template

#### 4-2. AGENTS.md / CLAUDE.md (Template + Carry-over Blocks)

Build the skeleton from the template and **place items confirmed in Step 2 into the appropriate locations in the new format**.

Minimum approach:

- Do not remove existing "important rules"; keep them as a **"Project-specific Rules (Migrated)"** section
- Rewrite role assignments/flows to match the template format (preserve the meaning)

### Step 5: Diff Review and Completion

- Briefly summarize changes via `git diff` (or file diff)
- Final confirmation that key points (permissions/prohibited actions/task state) are as intended
- Fix immediately if issues are found

---

## Deliverables (Completion Criteria)

- **New-format versions** of `AGENTS.md` / `CLAUDE.md` / `Plans.md` based on existing content
- Backups exist in `.claude-code-harness/backups/`
- Plans tasks have not been lost (preserved)
