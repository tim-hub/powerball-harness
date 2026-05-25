# Scripts — Orphaned (zero references)

Scripts listed here have no references in skills, agents, templates, hooks, or other scripts.
They are candidates for wiring up, consolidation, or eventual deletion.

## Advisor

- [ ] `advisor-check-cache.sh` — cache check helper, not wired to advisor agent
- [ ] `advisor-load-context.sh` — context loader, not wired to advisor agent

## Codex (legacy)

- [ ] `codex-loop.sh` — old Codex loop runner (replaced by codex-companion.sh)
- [ ] `codex-setup-local.sh` — old local Codex setup
- [ ] `codex-worker-engine.sh` — old worker engine
- [ ] `codex-worker-lock.sh` — old worker lock
- [ ] `codex-worker-merge.sh` — old worker merge
- [ ] `codex-worker-setup.sh` — old worker setup
- [ ] `check-codex.sh` — Codex availability check, superseded by codex-companion.sh setup

## Session

- [ ] `session-list.sh` — lists sessions, not wired to session skill
- [ ] `session-monitor.sh` — session monitoring, not wired
- [ ] `session-cleanup.sh` — session cleanup, not wired
- [ ] `session-summary.sh` — session summary, not wired
- [ ] `session-auto-broadcast.sh` — auto-broadcast, not wired

## Hook handlers (written but not in hooks.json)

- [ ] `stop-plans-reminder.sh` — stop hook reminder, not wired
- [ ] `stop-cleanup-check.sh` — stop hook cleanup check, not wired
- [ ] `stop-check-pending.sh` — stop hook pending check, not wired
- [ ] `posttooluse-clear-pending.sh` — PostToolUse handler, not wired
- [ ] `posttooluse-commit-cleanup.sh` — PostToolUse handler, not wired
- [ ] `posttooluse-log-toolname.sh` — PostToolUse handler, not wired
- [ ] `posttooluse-quality-pack.sh` — PostToolUse handler, not wired
- [ ] `pretooluse-browser-guide.sh` — PreToolUse handler, not wired
- [ ] `pretooluse-inbox-check.sh` — PreToolUse handler, not wired
- [ ] `userprompt-track-command.sh` — UserPromptSubmit handler, not wired

## Analytics / telemetry

- [ ] `usage-tracker.sh` — usage tracking, not wired
- [ ] `skill-trigger-telemetry.sh` — skill telemetry, not wired
- [ ] `skill-description-drift-report.sh` — skill description drift detector
- [ ] `subagent-tracker.sh` — subagent tracking, not wired
- [ ] `build-weak-supervision-cues.sh` — weak supervision data builder, not wired

## Setup utilities

- [ ] `quick-install.sh` — quick install helper, not wired to harness-setup
- [ ] `setup-existing-project.sh` — setup for existing projects, not wired
- [ ] `setup-hook.sh` — hook setup helper, not wired
- [ ] `opencode-setup-local.sh` — OpenCode local setup, not wired

## Misc

- [ ] `tdd-order-check.sh` — TDD order validation, not wired to tdd-guidelines template
- [ ] `todo-sync.sh` — TodoWrite ↔ Plans.md sync, not wired
- [ ] `track-changes.sh` — change tracking, not wired
- [ ] `plans-format-migrate.sh` — Plans.md format migration, not wired
- [ ] `localize-rules.sh` — rules localization, not wired
- [ ] `collect-cleanup-context.sh` — cleanup context collector, not wired
- [ ] `show-failures.sh` — failure display, not wired
- [ ] `sync-skill-mirrors.sh` — skill mirror sync, not wired
- [ ] `statusline-harness.sh` — status line helper, not wired
- [ ] `auto-test-runner.sh` — auto test runner, not wired
- [ ] `generate-x-article-image.sh` — X/Twitter article image generator, not wired
- [ ] `calculate-effort.sh` — effort calculation, not wired
- [ ] `check-simple-mode.sh` — simple mode check, not wired
