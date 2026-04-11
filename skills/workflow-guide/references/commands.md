# Command Reference

Details of commands used in the 2-agent workflow.

---

## Claude Code Side Commands

### /setup

Initial project setup (formerly `/harness-init`).

```
/setup
```

**Generated files**:
- Plans.md - Task management
- AGENTS.md - Role assignment definitions
- CLAUDE.md - Claude Code configuration
- .claude/rules/ - Project rules

---

### /setup codex

Introduces/updates Harness configuration for Codex CLI at the **user level** (`${CODEX_HOME:-~/.codex}`).

```
/setup codex
```

**Generated files (default)**:
- ${CODEX_HOME:-~/.codex}/skills/
- ${CODEX_HOME:-~/.codex}/rules/
- (optional) ${CODEX_HOME:-~/.codex}/config.toml

**Project mode only**:
- .codex/skills/
- .codex/rules/
- AGENTS.md

---

### /plan-with-agent

Plan and decompose tasks.

```
/plan-with-agent [task description]
```

**Example**:
```
/plan-with-agent I want to implement user authentication
```

**Output**: Tasks are added to Plans.md

---

### /work

Execute tasks from Plans.md.

```
/work
```

**Features**:
- Auto-detects tasks marked `cc:TODO` or `pm:requested`
- Supports parallel execution of multiple tasks
- Automatically updates status to `cc:done` upon completion

---

### /sync-status

Output a summary of the current state.

```
/sync-status
```

**Output example**:
```
Current State
- In progress: 2
- Not started: 5
- Completed (awaiting confirmation): 1
```

---

### /handoff-to-cursor

Completion report to the Cursor PM.

```
/handoff-to-cursor
```

**Included information**:
- List of completed tasks
- Changed files
- Test results
- Suggested next actions

---

## Cursor Side Commands (Reference)

### /handoff-to-claude

Task request to Claude Code.

### /review-cc-work

Review the completion report from Claude Code.
If not approved (request_changes), update Plans.md and **generate a correction request with `/claude-code-harness/handoff-to-claude` and pass it directly**.

---

## Skills (Auto-triggered in conversation)

### handoff-to-pm

**Trigger**: "report completion to PM", "report work completed"

Generates a completion report from Worker to PM.

### handoff-to-impl

**Trigger**: "hand off to the implementer", "request Claude Code"

Formats a task request from PM to Worker.

---

## Command Usage Flow

```
[Session Start]
    |
    v
/sync-status  <-- Check current status
    |
    v
/work  <-- Execute tasks
    |
    v
/handoff-to-cursor  <-- Completion report
    |
    v
[Session End]
```
