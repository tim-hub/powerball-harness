---
name: harness-plan
description: "Use this skill whenever the user asks to create a plan, add tasks, update Plans.md, mark tasks complete, check progress, sync status, or says 'where am I' or 'what's next'. Also use when the user runs /harness-plan, /harness-sync, or needs to organize work into actionable tasks. Do NOT load for: code implementation (use harness-work), code review (use harness-review), or release tasks (use harness-release). Unified planning skill for Harness v3 — task planning, Plans.md management, and progress sync."
allowed-tools: ["Read", "Write", "Edit", "Bash", "Grep", "Glob", "WebSearch", "Task"]
argument-hint: "[create|add|update|sync|sync --no-retro|--ci]"
effort: medium
---

# Harness Plan (v3)

Unified planning skill for Harness v3.
Consolidates the following 3 legacy skills:

- `planning` (plan-with-agent) — Turning ideas into Plans.md
- `plans-management` — Task state management and marker updates
- `sync-status` — Sync verification between Plans.md and implementation

## Quick Reference

| User Input | Subcommand | Behavior |
|------------|------------|------|
| "create a plan" | `create` | Interactive hearing -> Plans.md generation |
| "add a task" | `add` | Add new task to Plans.md |
| "mark complete" / "mark as done" | `update` | Change task marker to cc:done |
| "where am I?" / "check progress" | `sync` | Cross-reference implementation with Plans.md and sync |
| `harness-sync` | `sync` | Progress check (equivalent to standalone sync surface) |
| `harness-plan create` | `create` | Create plan |

## Subcommand Details

### create — Plan Creation

See [references/create.md](${CLAUDE_SKILL_DIR}/references/create.md)

Gather ideas and requirements through a hearing process, then generate an actionable Plans.md.

**Flow**:
1. Check conversation context (extract from recent discussion or start new hearing)
2. Ask what to build (max 3 questions)
3. Technical research (WebSearch)
4. Feature list extraction
5. Priority matrix (Required / Recommended / Optional)
6. TDD adoption decision (test design)
7. Plans.md generation (with `cc:TODO` markers)
8. Next action guidance

**CI Mode** (`--ci`):
No hearing. Uses existing Plans.md as-is and only performs task decomposition.

### add — Add Task

Add a new task to Plans.md.

```
harness-plan add task name: detailed description [--phase phase-number]
```

Tasks are added with the `cc:TODO` marker.

### update — Update Marker

Change a task's status marker.

```
harness-plan update [task-name|task-number] [WIP|done|blocked]
```

Marker mapping:

| Command | Marker |
|---------|---------|
| `WIP` | `cc:WIP` |
| `done` | `cc:done` |
| `blocked` | `blocked` |
| `TODO` | `cc:TODO` |

### sync — Progress Sync

Cross-reference implementation status with Plans.md, detecting and updating discrepancies.

See [references/sync.md](${CLAUDE_SKILL_DIR}/references/sync.md)

**Flow**:
1. Get current Plans.md state
2. Detect Plans.md format (v1: 3 columns / v2: 5 columns)
3. Get implementation status from git status / git log
4. Check agent trace (`.claude/state/agent-trace.jsonl`)
5. Detect drift between Plans.md and implementation
6. Propose automatic fixes for outdated markers
7. Present next actions

**Retrospective** (default ON):
Automatically runs a retrospective when 1 or more `cc:done` tasks exist.
Analyzes estimation accuracy, block cause patterns, and scope variation, then records learnings.
Can be explicitly skipped with `sync --no-retro`.

### team mode / issue bridge

Plans.md remains the source of truth; GitHub Issue integration is used only in opt-in team mode.

- Do not use the bridge in solo development
- Team mode creates a single tracking issue and generates sub-issue payloads per task as a dry-run
- `scripts/plans-issue-bridge.sh` never actually updates GitHub; it always returns dry-run payloads
- The bridge does not modify Plans.md

Reference:

- `docs/plans/team-mode.md`

## Plans.md Format Convention

### Format

```markdown
# [Project Name] Plans.md

Created: YYYY-MM-DD

---

## Phase N: Phase Name

| Task | Description | DoD | Depends | Status |
|------|------|-----|---------|--------|
| N.1  | Description | Tests pass | - | cc:TODO |
| N.2  | Description | lint errors 0 | N.1 | cc:WIP |
| N.3  | Description | Migration runnable | N.1, N.2 | cc:done |
```

**DoD (Definition of Done)**: Write verifiable completion criteria in one line. "Looks good" or "works properly" is prohibited. Must be Yes/No decidable.

**Depends**: Inter-task dependencies. `-` (no dependency), task number (`N.1`), comma-separated (`N.1, N.2`), phase dependency (`Phase N`).

### optional briefs / manifest

`harness-plan create` attaches briefs only when needed.

- Tasks involving UI get a `design brief`
- Tasks involving API get a `contract brief`
- Briefs are supplementary materials that briefly fix "what to build" and do not replace Plans.md
- Skill frontmatter listings can be exported as machine-readable JSON via `scripts/generate-skill-manifest.sh`

Reference:

- `docs/plans/briefs-manifest.md`

### Marker List

| Marker | Meaning |
|---------|------|
| `pm:requested` | Requested by PM |
| `cc:TODO` | Not started |
| `cc:WIP` | In progress |
| `cc:done` | Worker completed |
| `pm:confirmed` | PM review completed |
| `blocked` | Blocked (reason must be stated) |

## Related Skills

- `harness-sync` — Sync implementation with Plans.md
- `harness-work` — Implement planned tasks
- `harness-review` — Review implementation
- `harness-setup` — Project initialization
