---
name: session
description: "Manages Claude Code session lifecycle — listing, inbox, and broadcasting. Use when running session commands or managing active sessions."
when_to_use: "list sessions, inbox check, broadcast message, session lifecycle, manage sessions"
allowed-tools: ["Read", "Bash", "Write", "Edit", "Glob"]
argument-hint: "[list|inbox|broadcast \"message\"]"
---

# Session Skill (Unified)

Consolidates all session-related functionality into one skill: listing sessions, inbox checks, and cross-session broadcasting.

## Quick Reference

| User Input | Subcommand | Behavior |
|------------|------------|----------|
| `/session list` | `list` | Show all active Claude Code sessions in the current project |
| `/session inbox` | `inbox` | Check for incoming messages from other sessions |
| `/session broadcast "message"` | `broadcast` | Send a message to all active sessions |
| `/session` (no args) | _(none)_ | Show available subcommands and usage |

## Subcommands

### `/session list` - List Active Sessions

Shows all active Claude Code sessions in the current project.

```
📋 Active Sessions

| Session ID | Status | Last Activity |
|------------|--------|---------------|
| abc123     | active | 2 min ago     |
| def456     | idle   | 15 min ago    |
```

### `/session inbox` - Check Inbox

Checks for incoming messages from other sessions.

```
📬 Session Inbox

| From | Time | Message |
|------|------|---------|
| abc123 | 5m ago | "Ready for review" |
| def456 | 10m ago | "API implementation done" |
```

### `/session broadcast "message"` - Broadcast Message

Sends a message to all active sessions.

```bash
/session broadcast "Review complete, ready for merge"
```

---

## Capabilities

| Feature | Description | Reference |
|---------|-------------|-----------|
| **Initialization** | Start new session, load context | See `session-init` skill |
| **Memory** | Persist learnings across sessions | See [references/memory.md](${CLAUDE_SKILL_DIR}/references/memory.md) |
| **State Control** | Resume/fork session based on flags | See `references/session-control.md` |
| **State Transitions** | Session state machine details | See `references/state-transition.md` |
| **Communication** | Cross-session messaging | See `session-state` skill |

---

## Memory Optimization (CC 2.1.49+)

Session resume memory reduced 68% since CC 2.1.49. Best practices, workflow examples, and recommendations:
[references/memory-optimization.md](${CLAUDE_SKILL_DIR}/references/memory-optimization.md)

---

## When to Use

- Session initialization (use the `session-init` skill)
- Session resume/fork (`harness-work --resume`, `harness-work --fork`)
- Memory persistence (automatic)
- Cross-session communication (`/session broadcast`)

## Execution Flow & Files

Session initialization → control (resume/fork) → memory persistence → broadcast flow, plus managed file paths:
[references/execution-flow.md](${CLAUDE_SKILL_DIR}/references/execution-flow.md)

## Migration Note

This skill consolidates:
- `session-init` → Session initialization
- `session-memory` → Memory persistence (content now in [references/memory.md](${CLAUDE_SKILL_DIR}/references/memory.md))
- `session-control` → Resume/fork control
- `session-state` → State management & communication

The individual skills are deprecated but still work for backward compatibility.
