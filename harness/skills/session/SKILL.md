---
name: session
description: "Use when running /session — listing sessions, inbox checks, broadcasting, or managing Claude Code session lifecycle. Do NOT load for: application user sessions, login/auth state, or auth features."
allowed-tools: ["Read", "Bash", "Write", "Edit", "Glob"]
argument-hint: "[list|inbox|broadcast \"message\"]"
---

# Session Skill (Unified)

Consolidates all session-related functionality into one skill.

## Usage

```bash
/session              # Show available options
/session list         # Show active sessions
/session inbox        # Check incoming messages
/session broadcast "message"  # Send message to all sessions
```

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
| **Initialization** | Start new session, load context | See [../session-init/SKILL.md](../session-init/SKILL.md) |
| **Memory** | Persist learnings across sessions | See [../session-memory/SKILL.md](../session-memory/SKILL.md) |
| **State Control** | Resume/fork session based on flags | See [references/session-control.md](${CLAUDE_SKILL_DIR}/references/session-control.md) |
| **Communication** | Cross-session messaging | See [../session-state/SKILL.md](../session-state/SKILL.md) |

---

## Memory Optimization (CC 2.1.49+)

Since Claude Code 2.1.49, memory usage on session resume has been **reduced by 68%**.

### Best Practices for Long Session Management

| Workload | Recommended Strategy |
|----------|---------------------|
| **Normal implementation** | Resume with `--resume` every 1-2 hours |
| **Large-scale refactoring** | Split sessions by feature unit, use `--resume` for each |
| **Parallel tasks** | Run in parallel with `/work all`, use `--resume` midway for long sessions |
| **Memory warning** | Resume immediately with `--resume` (faster than before) |

### Auto-generated Session Names (CC 2.1.41+)

Running `/rename` without arguments auto-generates a session name from the conversation context.
This makes it easier to identify sessions in long-running or `--resume`-heavy workflows.

### Efficient Workflow Example

```bash
# Implementation phase 1
claude "Implement authentication feature"
# -> 1 hour later

# Resume session (memory-efficient)
claude --resume "Add password reset feature"
# -> 1 hour later

# Resume again
claude --resume "Add tests"
```

### Memory Management Recommendations

| Recommendation | Reason |
|---------------|--------|
| **Actively resume sessions** | Low resume cost with 68% memory reduction |
| **Resume periodically** | Keeps context organized and maintains focus |
| **Split by feature unit** | Break large tasks into smaller chunks for resuming |
| **Use Plans.md** | Smooth handoff when resuming |

> 💡 Memory efficiency has been significantly improved, so actively take advantage of session resumption.

---

## When to Use

- Session initialization (`/harness-init`)
- Session resume/fork (`/work --resume`, `/work --fork`)
- Memory persistence (automatic)
- Cross-session communication (`/session broadcast`)

## Execution Flow

### 1. Session Initialization

```
/harness-init
    ↓
├── Load project context
├── Initialize session.json
├── Load previous session memory (if exists)
└── Display session status
```

### 2. Session Control (from /work)

```
/work --resume
    ↓
├── Check session.json exists
├── Load session state
└── Continue from last checkpoint

/work --fork
    ↓
├── Create new session branch
├── Copy relevant context
└── Start fresh with context
```

### 3. Memory Persistence

```
Session end
    ↓
├── Extract learnings (gotchas, patterns)
├── Update .claude/memory/*.md
└── Prepare handoff summary
```

### 4. Cross-Session Communication

```
/session broadcast "message"
    ↓
├── Find active sessions
├── Write to session.events.jsonl
└── Notify all sessions
```

## Files Managed

| File | Purpose |
|------|---------|
| `.claude/state/session.json` | Current session state |
| `.claude/state/session.events.jsonl` | Event log for cross-session communication |
| `.claude/memory/*.md` | Persistent memory files |

## Migration Note

This skill consolidates:
- `session-init` → Session initialization
- `session-memory` → Memory persistence
- `session-control` → Resume/fork control
- `session-state` → State management & communication

The individual skills are deprecated but still work for backward compatibility.
