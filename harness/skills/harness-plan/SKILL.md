---
name: harness-plan
description: "Use when creating plans, adding tasks, updating Plans.md, marking tasks done, or checking progress/sync. Do NOT load for: implementation (harness-work), review (harness-review), or release (harness-release)."
allowed-tools: ["Read", "Write", "Edit", "Bash", "Grep", "Glob", "WebSearch", "Task"]
argument-hint: "[create|add|update|sync|archive|session-log|sync --no-retro|--ci]"
effort: medium
model: opus
---

# Harness Plan

Unified planning skill for Harness.
Consolidates the following 3 legacy skills:

- `planning` (plan-with-agent) — Convert ideas into Plans.md
- `plans-management` — Task state management and marker updates
- `sync-status` — Sync verification between Plans.md and implementation

## Quick Reference

| User Input | Subcommand | Behavior |
|------------|------------|----------|
| "create a plan" | `create` | Interactive interview → Plans.md generation |
| "add a task" | `add` | Add new task to Plans.md |
| "mark complete" | `update` | Change task marker to cc:done |
| "where am I?" / "check progress" | `sync` | Compare implementation with Plans.md and sync |
| `harness-plan sync` / "sync status" | `sync` | Progress check via harness-plan's embedded sync subcommand |
| `harness-plan create` | `create` | Create plan |
| "archive old phases" / `harness-plan archive` | `archive` | Archive phases in Plans.md to `.claude/memory/archive/`; update `Last archive:` header |
| "session log too big" / `harness-plan session-log` | `session-log` | Split session-log.md by month; move older months to `.claude/memory/session-log-YYYY-MM.md` |

## Subcommand Details

### create — Plan Creation

See [references/create.md](${CLAUDE_SKILL_DIR}/references/create.md)

Interviews for ideas and requirements, then generates an actionable Plans.md.

**Flow**:
1. Check conversation context (extract from preceding discussion or start new interview)
2. Ask what to build (max 3 questions)
3. Technical research (WebSearch)
4. Feature list extraction
5. Priority matrix (Required / Recommended / Optional)
6. TDD adoption decision (test design)
7. Plans.md generation (with `cc:TODO` markers)
8. Suggest next actions

**CI mode** (`--ci`):
No interview. Uses existing Plans.md as-is and only performs task decomposition.

### add — Add Task

Adds a new task to Plans.md.

```
harness-plan add task-name: detailed description [--phase phase-number]
```

Tasks are added with the `cc:TODO` marker.

### update — Update Marker

Changes a task's status marker.

```
harness-plan update [task-name|task-number] [WIP|done|blocked]
```

Marker mapping:

| Command | Marker |
|---------|--------|
| `WIP` | `cc:WIP` |
| `done` | `cc:done` |
| `blocked` | `blocked` |
| `TODO` | `cc:TODO` |

### sync — Progress Sync

Compares implementation status with Plans.md and detects/updates differences.

See [references/sync.md](${CLAUDE_SKILL_DIR}/references/sync.md)

**Flow**:
1. Retrieve current state of Plans.md
2. Detect Plans.md format (v1: 3 columns / v2: 5 columns)
3. Retrieve implementation status from git status / git log
4. Check agent traces (`.claude/state/agent-trace.jsonl`)
5. Detect differences between Plans.md and implementation
6. Propose automatic corrections for outdated markers
7. Present next actions

**Retrospective** (ON by default):
Automatically runs a retrospective if there is at least one `cc:done` task.
Analyzes estimation accuracy, blocker cause patterns, and scope changes, then records learnings.
Can be explicitly skipped with `sync --no-retro`.

### archive — Plans.md Archiving

Moves fully-completed phases out of Plans.md into `.claude/memory/archive/` to keep the active file lean. A phase is eligible when every task in it has a `cc:done` or `pm:confirmed` marker.

**Flow**:
1. Read Plans.md and identify phases where all tasks are `cc:done` / `pm:confirmed`
2. Write the completed phases to `.claude/memory/archive/Plans-YYYY-MM-DD-phaseX-Y.md` (using today's date and the range of archived phase numbers)
3. Remove those phases from Plans.md
4. Update the `Last archive:` line at the top of Plans.md to record the date and archive filename

**What stays in Plans.md**: 
- Any phase with at least one task that is `cc:TODO`, `cc:WIP`, or `blocked`.
- The 10 most recent completed phases, even if they are fully `cc:done` / `pm:confirmed`, to maintain recent history and context.

**Naming convention**: `Plans-YYYY-MM-DD-phaseX-Y.md` where X is the lowest and Y the highest archived phase number. Example: `Plans-2026-04-15-phase35-48.md`.

### session-log — Split session-log.md by Month

Moves sessions from past months out of `.claude/memory/session-log.md` into per-month archive files, keeping the active file lean.

**Flow**:
1. Read `.claude/memory/session-log.md`
2. Parse each `## Session: YYYY-MM-DDTHH:MM:SSZ` header; group all content blocks by `YYYY-MM`
3. Identify months older than the current month — these are candidates for archiving
4. For each older month: write its session blocks (including their `---` separators) to `.claude/memory/session-log-YYYY-MM.md`
5. Rewrite session-log.md keeping only: the file header (first 10 lines up to and including `---`), the current month's sessions, and an updated `## Index` section with links to each archived file
6. Commit the changes

**What stays in session-log.md**:
- The file header and Index section
- All sessions from the current calendar month

**Archive naming**: `session-log-YYYY-MM.md` — e.g. `session-log-2026-03.md` for March 2026.

**When nothing to archive** (all sessions are current month): report "session-log.md is already current — nothing to archive" and exit without modifying any files.

### team mode / issue bridge

Plans.md is maintained as the source of truth, and GitHub Issue integration is only used in opt-in team mode.

- Do not use the bridge for solo development
- Team mode creates one tracking issue and generates dry-run sub-issue payloads for each task underneath it
- `scripts/plans-issue-bridge.sh` does not actually update GitHub; it always returns dry-run payloads
- This bridge does not modify Plans.md

Reference:

- `docs/plans/team-mode.md`

## Plans.md Format Conventions

### Format

```markdown
# [Project Name] Plans.md

Created: YYYY-MM-DD

---

## Phase N: Phase Name

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| N.1  | Description | Tests pass | - | cc:TODO |
| N.2  | Description | 0 lint errors | N.1 | cc:WIP |
| N.3  | Description | Migration executable | N.1, N.2 | cc:done |
```

**DoD (Definition of Done)**: Write a verifiable completion condition in one line. Vague criteria like "looks good" or "works properly" are prohibited. Must be answerable with Yes/No.

**Depends**: Dependencies between tasks. `-` (no dependencies), task number (`N.1`), comma-separated (`N.1, N.2`), or phase dependency (`Phase N`).

### optional briefs / manifest

`harness-plan create` attaches briefs only when needed.

- Tasks involving UI get a `design brief`
- Tasks involving API get a `contract brief`
- Briefs are supplementary materials that briefly define what to build; they do not replace Plans.md
- A machine-readable JSON list of skill frontmatter can be generated with `scripts/generate-skill-manifest.sh`

Reference:

- `docs/plans/briefs-manifest.md`

### Marker List

| Marker | Meaning |
|--------|---------|
| `pm:requested` | Requested by PM |
| `cc:TODO` | Not started |
| `cc:WIP` | In progress |
| `cc:done` | Worker completed |
| `pm:confirmed` | PM review completed |
| `blocked` | Blocked (reason must always be noted) |

## Related Skills

- `harness-sync` — Sync implementation with Plans.md
- `harness-work` — Implement planned tasks
- `harness-review` — Review implementation
- `harness-setup` — Project initialization
