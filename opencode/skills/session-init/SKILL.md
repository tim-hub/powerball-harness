---
name: session-init
description: "Use when starting a new session — pre-work status check, environment readiness verification, or Plans.md overview. Do NOT load for: mid-session implementation, reviews, or ongoing tasks."
allowed-tools: ["Read", "Write", "Bash"]
user-invocable: false
---

# Session Init Skill

A skill for verifying the environment and understanding the current task status at session start.

---

## Trigger Phrases

This skill is triggered by the following phrases:

- "Start session"
- "Start work"
- "Start today's work"
- "Check the status"
- "What should I do?"
- "start session"
- "what should I work on?"

---

## Overview

The Session Init skill automatically checks the following at Claude Code session start:

1. **Git status**: Current branch, uncommitted changes
2. **Plans.md**: In-progress tasks, requested tasks
3. **AGENTS.md**: Role assignments, prohibited actions
4. **Previous session**: Handoff items to review
5. **Latest snapshot**: Progress snapshot summary and diff from last time

---

## Execution Steps

### Step 0: File Status Check (Auto-cleanup)

Check file sizes before starting the session:

```bash
# Check Plans.md line count
if [ -f "Plans.md" ]; then
  lines=$(wc -l < Plans.md)
  if [ "$lines" -gt 200 ]; then
    echo "⚠️ Plans.md has ${lines} lines. Recommend cleanup with 'clean up'"
  fi
fi

# Check session-log.md line count
if [ -f ".claude/memory/session-log.md" ]; then
  lines=$(wc -l < .claude/memory/session-log.md)
  if [ "$lines" -gt 500 ]; then
    echo "⚠️ session-log.md has ${lines} lines. Recommend cleanup with 'clean up session log'"
  fi
fi
```

If cleanup is needed, a suggestion is displayed (does not affect work).

### Step 0.5: Legacy Local Memory Compatibility (Optional)

The current standard is the Unified Harness Memory in Step 0.7.
Checking legacy local memory compatibility is generally unnecessary; refer to it only when special migration verification is needed.

> **Note**: In normal operation, skip this step and treat the shared DB Resume Pack as the sole resumption path.

### Step 0.7: Unified Harness Memory Resume Pack (Required)

Retrieve resume context from the Codex / Claude / OpenCode shared DB (`~/.harness-mem/harness-mem.db`).

Required call:

```text
harness_mem_resume_pack(project, session_id?, limit=5, include_private=false)
```

Operational rules:
- `project` must always specify the current project name
- `session_id` is obtained from `$CLAUDE_SESSION_ID`, falling back to `.session_id` in `.claude/state/session.json`
- Using the first result of `harness_mem_sessions_list(project, limit=1)` is limited to read-only (resume confirmation); do not use it for writes via `record_checkpoint` / `finalize_session`
- Inject retrieved results into the session start context
- On retrieval failure, check daemon status with `harness_mem_health()`, report the failure explicitly, and continue
- Recovery order: `scripts/harness-memd doctor` -> `scripts/harness-memd cleanup-stale` -> `scripts/harness-memd start`

### Step 1: Environment Check

Execute the following in parallel:

```bash
# Git status
git status -sb
git log --oneline -3
```

```bash
# Plans.md
cat Plans.md 2>/dev/null || echo "Plans.md not found"
```

```bash
# Key points from AGENTS.md
head -50 AGENTS.md 2>/dev/null || echo "AGENTS.md not found"
```

### Step 2: Understand Task Status

Extract the following from Plans.md:

- `cc:WIP` - Tasks continuing from the previous session
- `pm:requesting` - Newly requested tasks from the PM (compat: cursor:requesting)
- `cc:TODO` - Unstarted but assigned tasks

### Step 3: Output Status Report

```markdown
## 🚀 Session Start

**Date/Time**: {{YYYY-MM-DD HH:MM}}
**Branch**: {{branch}}
**Session ID**: ${CLAUDE_SESSION_ID}

---

### 📋 Today's Tasks

**Priority Tasks**:
- {{pm:requesting (compat: cursor:requesting) or cc:WIP tasks}}

**Other Tasks**:
- {{List of cc:TODO tasks}}

---

### ⚠️ Notes

{{Important constraints and prohibitions from AGENTS.md}}

---

**Ready to start work?**
```

---

## Output Format

At session start, present the following information concisely:

| Item | Content |
|------|---------|
| Current branch | e.g., `staging` |
| Priority tasks | Top 1-2 most important |
| Notes | Summary of prohibitions |
| Next action | Specific suggestions |

---

## Related Commands

- `/work` - Execute tasks (supports parallel execution)
- `/sync-status` - Progress summary for Plans.md
- `/maintenance` - Auto-cleanup of files

---

## Notes

- **Always check AGENTS.md**: Understand role assignments before starting work
- **If Plans.md doesn't exist**: Suggest `/harness-init`
- **If previous work was interrupted**: Confirm whether to continue
