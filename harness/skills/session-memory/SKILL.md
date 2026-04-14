---
name: session-memory
description: "Use when recalling prior sessions, continuing past work, referencing earlier decisions, or persisting cross-session context. Do NOT load for: implementation, reviews, ad-hoc notes, or in-session logging."
allowed-tools: ["Read", "Write", "Edit"]
user-invocable: false
---

# Session Memory Skill

A skill for managing cross-session learning and memory.
Records and references past work, decisions, and learned patterns.

---

## Trigger Phrases

This skill is automatically triggered by the following phrases:

- "What did we do last time?", "Continue from last time"
- "Show me the history", "Past work"
- "Tell me about this project"
- "what did we do last time?", "continue from before"

---

## Overview

This skill saves work history to `.claude/memory/` and enables
knowledge continuity across sessions.

It also clarifies where important information should be stored (details: `docs/MEMORY_POLICY.md`).

---

## Memory Structure

```
.claude/
├── memory/
│   ├── session-log.md      # Per-session log
│   ├── decisions.md        # Important decisions
│   ├── patterns.md         # Learned patterns
│   └── context.json        # Project context
└── state/
    └── agent-trace.jsonl   # Agent Trace (tool execution history)
```

### Recommended Operation (SSOT/Local Separation)

- **SSOT (shared, recommended)**: `decisions.md` / `patterns.md`  
  - Aggregate "Decisions (Why)" and "Reusable solutions (How)"
  - Each entry should have a **title + tags** (e.g., `#decision #db`) with an **Index** at the top
- **Local (recommended)**: `session-log.md` / `context.json` / `.claude/state/`  
  - Prone to noise/bloat, so generally not managed in Git (decide on a case-by-case basis if needed)

---

## Automatically Recorded Information

### session-log.md

Each session record is assigned a session ID using the `${CLAUDE_SESSION_ID}` environment variable.
This improves traceability across sessions.

```markdown
## Session: 2024-01-15 14:30 (session: abc123def)

### Completed Tasks
- [x] User authentication implementation
- [x] Login page creation

### Generated Files
- src/lib/auth.ts
- src/app/login/page.tsx

### Important Decisions
- Authentication method: Adopted Supabase Auth

### Handoff for Next Session
- Logout feature not yet implemented
- Password reset also needed
```

> **Note**: `${CLAUDE_SESSION_ID}` is an environment variable automatically set by Claude Code.
> A unique ID is assigned per session, useful for log tracking and issue investigation.

### decisions.md

```markdown
## Technology Choices

| Date | Decision | Reason |
|------|----------|--------|
| 2024-01-15 | Supabase Auth | Free tier available, easy setup |
| 2024-01-14 | Next.js App Router | Latest best practices |

## Architecture

- Components: `src/components/`
- Utilities: `src/lib/`
- Type definitions: `src/types/`
```

### patterns.md

```markdown
## Patterns for This Project

### Component Naming
- PascalCase
- Example: `UserProfile.tsx`, `LoginForm.tsx`

### API Endpoints
- `/api/v1/` prefix
- RESTful design

### Error Handling
- Wrap in try-catch
- Error messages
```

### context.json

```json
{
  "project_name": "my-blog",
  "created_at": "2024-01-14",
  "stack": {
    "frontend": "next.js",
    "backend": "next-api",
    "database": "supabase",
    "styling": "tailwind"
  },
  "current_phase": "Phase 2: Core Features",
  "last_session": "2024-01-15T14:30:00Z"
}
```

---

## Processing Flow

### At Session Start

1. Load `.claude/memory/context.json`
2. Review previous session log
3. **Retrieve recent edit history from Agent Trace**
4. Identify incomplete tasks
5. Generate context summary

**Agent Trace Usage**:
```bash
# Get list of recently edited files
tail -50 .claude/state/agent-trace.jsonl | jq -r '.files[].path' | sort -u

# Get project information
tail -1 .claude/state/agent-trace.jsonl | jq '.metadata'
```

### During Session

1. Record important decisions in `decisions.md`
2. Add new patterns to `patterns.md`
3. Record file generation in `session-log.md`

### At Session End

1. Generate session summary
2. Update `context.json`
3. Record handoff items for next session

---

## Memory Optimization (CC 2.1.49+)

Since Claude Code 2.1.49, memory usage on session resume has been **reduced by 68%**.

### Recommended Workflow

```bash
# Use --resume for long work sessions
claude --resume

# Split large tasks and resume sessions
claude --resume "continue from where we left off"
```

| Scenario | Recommendation |
|----------|---------------|
| Long implementation | Resume session every 1-2 hours |
| Large-scale refactoring | Split sessions by feature unit |
| Memory shortage warning | Resume immediately with `--resume` |

> 💡 Memory efficiency has been significantly improved, so actively take advantage of session resumption.

---

## Usage Examples

### Continue from Previous Session

```
User: "Continue from last time"

Claude Code:
📋 Previous Session (2024-01-15)

Completed tasks:
- User authentication
- Login page

Incomplete:
- Logout feature
- Password reset

Say "build the logout feature" to continue implementation.
```

### Check Project Status

```
User: "Tell me about this project"

Claude Code:
📁 Project: my-blog

Tech stack:
- Next.js + Tailwind CSS + Supabase

Current phase: Core feature development
Progress: 40% complete

Recent decisions:
- Adopted Supabase Auth
- Using App Router
```

---

## Relationship with Claude Code Auto-Memory (D22)

Claude Code 2.1.32+ has an "auto-memory" feature that automatically saves cross-session learnings to `~/.claude/projects/<project>/memory/MEMORY.md`.

It coexists with the Harness memory system as a **3-layer architecture**:

| Layer | System | Content | Management |
|-------|--------|---------|------------|
| **Layer 1** | Claude Code Auto-Memory | General learnings (mistake avoidance, tool usage) | Implicit, automatic |
| **Layer 2** | Harness SSOT | Project-specific decisions and patterns | Explicit, manual |
| **Layer 3** | Agent Memory | Per-agent task learnings | Agent-defined |

**Usage guidelines**:
- If Layer 1 insights are important project-wide, promote to Layer 2 with `/memory ssot`
- Leave everyday learning to Layer 1 (do not disable)
- Be cautious of concurrent writes when using Agent Teams

Details: [D22: 3-Layer Memory Architecture](../../.claude/memory/decisions.md#d22-3-layer-memory-architecture)

---

## Notes

- **Auto-save**: Recommended to use `hooks/Stop` to auto-append summaries to `session-log.md` at session end (manual operation is fine if not set up)
- **Privacy**: Do not record confidential information
- **Git policy**: `decisions.md`/`patterns.md` are recommended for sharing; `session-log.md`/`context.json`/`.claude/state/` are recommended to keep local (details: `docs/MEMORY_POLICY.md`)
- **Capacity management**: When logs grow large, recommend "clean up session log"
