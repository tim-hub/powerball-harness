---
name: harness-plan
description: "Plans and tracks tasks in Plans.md. Use when creating plans, adding tasks, updating markers, or checking progress."
when_to_use: "create a plan, add a task, mark task done, where am I, check progress, sync plans, archive phases"
allowed-tools: ["Read", "Write", "Edit", "Bash", "Grep", "Glob", "WebSearch", "Task"]
argument-hint: "[create|add|update|sync|archive|session-log|sync --no-retro|--ci]"
effort: xhigh
model: opus
---

# Harness Plan

## Quick Reference

| User Input | Subcommand | Behavior |
|------------|------------|----------|
| "create a plan" | `create` | Default behavior, interactive interview → Plans.md generation |
| "add a task" | `add` | Add new task to Plans.md |
| "mark complete" | `update` | Change task marker to cc:done |
| "where am I?" / "check progress" / `harness-plan sync` / "sync status"| `sync` | Compare implementation with Plans.md and sync |
| "archive old phases" / `harness-plan archive` | `archive` | Archive phases in Plans.md to `.claude/memory/archive/`; update `Last archive:` line in the `## Archive` footer |
| "session log too big" / `harness-plan session-log` | `session-log` | Split session-log.md by month; move older months to `.claude/memory/session-log-YYYY-MM.md` |

## Subcommand Details

Each subcommand has its own reference file. Open the matching file when invoking that subcommand.

| Subcommand | Reference |
|------------|-----------|
| `create` | [references/create.md](${CLAUDE_SKILL_DIR}/references/create.md) — interview flow, TDD decision rules, Plans.md generation |
| `add` | [references/add.md](${CLAUDE_SKILL_DIR}/references/add.md) — insertion rules, DoD/Depends auto-inference |
| `update` | [references/update.md](${CLAUDE_SKILL_DIR}/references/update.md) — marker mapping, single-task status changes |
| `sync` | [references/sync.md](${CLAUDE_SKILL_DIR}/references/sync.md) — discrepancy detection, retrospective |
| `archive` | [references/archive.md](${CLAUDE_SKILL_DIR}/references/archive.md) — phase archival, retention, naming |
| `session-log` | [references/session-log.md](${CLAUDE_SKILL_DIR}/references/session-log.md) — monthly split of session-log.md |

**CI mode** (`--ci`) — applies to `create` only: no interview; uses existing Plans.md and only performs task decomposition. See [references/create.md](${CLAUDE_SKILL_DIR}/references/create.md) "CI Mode" section.

**Retrospective skip** (`--no-retro`) — applies to `sync` only: skips the automatic retrospective pass. See [references/sync.md](${CLAUDE_SKILL_DIR}/references/sync.md) "Step 6: Retrospective".

## Plans.md Format Conventions

The canonical template lives in [templates/plans-md-template.md](${CLAUDE_SKILL_DIR}/templates/plans-md-template.md). Ordering rules, field definitions, and behavioral requirements are in [references/plans-md-rules.md](${CLAUDE_SKILL_DIR}/references/plans-md-rules.md).

**Key rules (summary)**:
- **Newest phase on top** — insert above all existing `## Phase` blocks, never at the bottom
- **Non-ascending order** — phase numbers decrease top-to-bottom; gaps are allowed (archived phases removed)
- **Archive footer last** — `## Archive` section stays at the very bottom with a link table to `.claude/memory/archive/`
- **DoD must be verifiable** — yes/no answerable; banned: "looks good", "works properly"

## Marker List

Compact summary. For full semantics (including the `cc:done [hash]` artifact format and `blocked` reason annotation), see [references/plans-md-rules.md](${CLAUDE_SKILL_DIR}/references/plans-md-rules.md) "Status markers".

| Marker | Meaning |
|--------|---------|
| `pm:requested` | Requested by PM |
| `cc:TODO` | Not started |
| `cc:WIP` | In progress |
| `cc:done` | Worker completed |
| `pm:confirmed` | PM review completed |
| `blocked` | Blocked (reason must always be noted) |

## Optional Briefs / Manifest

`harness-plan create` attaches briefs only when needed.

- Tasks involving UI get a `design brief`
- Tasks involving API get a `contract brief`
- Briefs are supplementary materials that briefly define what to build; they do not replace Plans.md
- A machine-readable JSON list of skill frontmatter can be generated with `scripts/generate-skill-manifest.sh`

Reference:

- `docs/plans/briefs-manifest.md`

## Team Mode / Issue Bridge

Plans.md is maintained as the source of truth, and GitHub Issue integration is only used in opt-in team mode.

- Do not use the bridge for solo development
- Team mode creates one tracking issue and generates dry-run sub-issue payloads for each task underneath it
- `scripts/plans-issue-bridge.sh` does not actually update GitHub; it always returns dry-run payloads
- This bridge does not modify Plans.md

Reference:

- `docs/plans/team-mode.md`

## Related Skills

- `harness-sync` — Sync implementation with Plans.md
- `harness-work` — Implement planned tasks
- `harness-review` — Review implementation
- `harness-setup` — Project initialization
