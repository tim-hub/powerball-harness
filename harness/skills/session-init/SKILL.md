---
name: session-init
description: "Pre-work status check that verifies environment readiness and gives a Plans.md overview. Use when starting a new session or checking status before work."
when_to_use: "start session, start work, check status, pre-work check, what needs to be done, environment ready"
allowed-tools: ["Read", "Write", "Bash"]
user-invocable: false
---

# Session Init Skill

A skill for verifying the environment and understanding the current task status at session start.

## Overview

The Session Init skill automatically checks the following at Claude Code session start:

1. **Git status**: Current branch, uncommitted changes
2. **Plans.md**: In-progress tasks, requested tasks
3. **AGENTS.md**: Role assignments, prohibited actions
4. **Previous session**: Handoff items to review
5. **Latest snapshot**: Progress snapshot summary and diff from last time

---

## Execution Steps

Full bash commands and templates for all steps: [references/steps.md](${CLAUDE_SKILL_DIR}/references/steps.md)

### Step 0: File Status Check (Auto-cleanup)

Check Plans.md (>200 lines) and session-log.md (>500 lines); suggest cleanup when thresholds are exceeded.

### Step 0.5: Legacy Local Memory Compatibility (Optional)

The current standard is the Unified Harness Memory in Step 0.7.
Checking legacy local memory compatibility is generally unnecessary; refer to it only when special migration verification is needed.

> **Note**: In normal operation, skip this step and treat the shared DB Resume Pack as the sole resumption path.

### Step 0.7: Unified Harness Memory Resume Pack (Required)

Call `harness_mem_resume_pack(project, session_id?, limit=5, include_private=false)` to retrieve resume context from the shared DB (`~/.harness-mem/harness-mem.db`). On failure: run `harness_mem_health()` then follow recovery steps in [references/steps.md](${CLAUDE_SKILL_DIR}/references/steps.md).

### Step 1: Environment Check

Execute git status, Plans.md read, and AGENTS.md head in parallel. See [references/steps.md](${CLAUDE_SKILL_DIR}/references/steps.md) for exact commands.

### Step 2: Understand Task Status

Extract the following from Plans.md:

- `cc:WIP` - Tasks continuing from the previous session
- `pm:requesting` - Newly requested tasks from the PM
- `cc:TODO` - Unstarted but assigned tasks

### Step 3: Output Status Report

Output a structured session-start report (date/time, branch, priority tasks, AGENTS.md notes). Full template in [references/steps.md](${CLAUDE_SKILL_DIR}/references/steps.md).

---

## Related Skills

- `harness-work` - Execute tasks (supports parallel execution)
- `harness-sync` - Progress summary for Plans.md
- `harness-setup` - Initialize a new project (when Plans.md doesn't exist)

## Notes

- **Always check AGENTS.md**: Understand role assignments before starting work
- **If Plans.md doesn't exist**: Suggest running `harness-setup`
- **If previous work was interrupted**: Confirm whether to continue
- **"What should I do?" overlap**: This skill handles session-start context loading. For open-ended guidance on how to work, `vibecoder-guide` may be more appropriate.
