# Memory Architecture & Usage Examples

## 3-Layer Memory Architecture (CC 2.1.32+)

| Layer | System | Content | Management |
|-------|--------|---------|------------|
| **Layer 1** | Claude Code Auto-Memory | General learnings (mistake avoidance, tool usage) | Implicit, automatic |
| **Layer 2** | Harness SSOT | Project-specific decisions and patterns | Explicit, manual |
| **Layer 3** | Agent Memory | Per-agent task learnings | Agent-defined |

**Usage guidelines**:
- If Layer 1 insights are important project-wide, promote to Layer 2 with `/memory ssot`
- Leave everyday learning to Layer 1 (do not disable)
- Be cautious of concurrent writes when using Agent Teams

Details: `.claude/memory/decisions.md` — entry D22: 3-Layer Memory Architecture (project-root)

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

## Memory Optimization (CC 2.1.49+)

Memory usage on session resume reduced by 68% since CC 2.1.49.

| Scenario | Recommendation |
|----------|---------------|
| Long implementation | Resume session every 1-2 hours |
| Large-scale refactoring | Split sessions by feature unit |
| Memory shortage warning | Resume immediately with `--resume` |

Use `claude --resume` to continue long work sessions efficiently.
