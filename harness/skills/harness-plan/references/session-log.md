# session-log Subcommand — Split session-log.md by Month

Moves sessions from past months out of `.claude/memory/session-log.md` into per-month archive files, keeping the active file lean.

## Flow

1. Read `.claude/memory/session-log.md`.
2. Parse each `## Session: YYYY-MM-DDTHH:MM:SSZ` header; group all content blocks by `YYYY-MM`.
3. Identify months **older than the current month** — these are candidates for archiving.
4. For each older month: write its session blocks (including their `---` separators) to `.claude/memory/session-log-YYYY-MM.md`.
5. Rewrite session-log.md keeping only:
   - The file header (first 10 lines up to and including `---`).
   - The current month's sessions.
   - An updated `## Index` section with links to each archived file.
6. Commit the changes.

## What Stays in session-log.md

- The file header and Index section.
- All sessions from the **current calendar month**.

## Archive Naming

`session-log-YYYY-MM.md`

Example: `session-log-2026-03.md` for March 2026.

## When Nothing to Archive

If all sessions in session-log.md are from the current month, report:

> session-log.md is already current — nothing to archive

…and exit without modifying any files.

## References

- [archive.md](${CLAUDE_SKILL_DIR}/references/archive.md) — sibling subcommand for archiving completed phases out of Plans.md
