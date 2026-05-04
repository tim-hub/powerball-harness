# Session Memory File Formats

Templates for each file in the `.claude/memory/` structure.

## session-log.md

Each session record is assigned a session ID via `${CLAUDE_SESSION_ID}`:

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

> `${CLAUDE_SESSION_ID}` is set automatically by Claude Code per session.

## decisions.md

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

## patterns.md

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

## context.json

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
