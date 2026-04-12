---
name: harness-plan
description: "Use this skill whenever the user asks to create a plan, add tasks, update Plans.md, mark tasks complete, check progress, sync status, or says 'where am I' or 'what's next'. Also use when the user runs /harness-plan, /harness-sync, or needs to organize work into actionable tasks. Do NOT load for: code implementation (use harness-work), code review (use harness-review), or release tasks (use harness-release). Unified planning skill for Harness v3 — task planning, Plans.md management, and progress sync."
allowed-tools: ["Read", "Write", "Edit", "Bash", "Grep", "Glob", "WebSearch", "Task"]
argument-hint: "[create|add|update|sync|sync --no-retro|--ci]"
effort: medium
model: opus
---

# Harness Plan (v3)

Unified planning skill for Harness v3.
Consolidates the following 3 legacy skills:

- `planning` (plan-with-agent) — Convert ideas into Plans.md
- `plans-management` — Task state management and marker updates
- `sync-status` — Sync verification between Plans.md and implementation

## Quick Reference

| User Input | Subcommand | Action |
|------------|------------|--------|
| "create a plan" | `create` | Interactive interview → Plans.md generation |
| "add a task" | `add` | Add new task to Plans.md |
| "mark complete" | `update` | Change task marker to cc:done |
| "where am I?" / "check progress" | `sync` | Compare implementation with Plans.md and sync |
| `harness-sync` | `sync` | Progress check (equivalent to standalone sync surface) |
| `harness-plan create` | `create` | Create plan |

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
