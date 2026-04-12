# AGENTS.md - Codex Harness Development Guide

This file guides Codex CLI when working in this repository.

## Project Overview

**Harness** is a guide for running Codex CLI in a "Plan → Work → Review" loop.

**Special note**: This project is self-referential — it uses the harness itself to improve the harness.

## Codex CLI Assumptions

- Codex reads `${CODEX_HOME:-~/.codex}/skills/<skill-name>/SKILL.md` (user-based) and `.codex/skills/...` (project override), invoked via `$skill-name`
- Codex prioritizes `AGENTS.override.md`, then `AGENTS.md`, then configured fallback names
- Hooks are not supported; temporary guards use `prefix_rule()` in `.codex/rules/*.rules`

## Development Rules

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/): `feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`

### Versioning

`VERSION` is the source of truth. Do not change `VERSION` or `.claude-plugin/marketplace.json` during regular work. Use `./scripts/sync-version.sh bump` only when cutting a release.

### Code Style

- Clear, descriptive names
- Comments for complex logic
- Single-responsibility skills/agents

## Repository Structure

```
claude-code-harness/
├── codex/              # Codex CLI distribution (symlinked skills)
├── agents/             # Sub-agent definitions
├── skills/             # Agent skills (SSOT, symlinked into codex/)
├── scripts/            # Shell scripts (guards, automation)
├── docs/               # Documentation
└── tests/              # Validation scripts
```

## Skill Architecture

Skills in `codex/.codex/skills/` are **symlinks** to `../../../skills/`. Single source of truth, no duplication.

## Primary Skills

| Skill | Purpose | Trigger |
|-------|---------|---------|
| `$harness-plan` | Planning, task decomposition | "plan this", "add a task" |
| `$harness-work` | Implementation, parallel execution | "implement", "do everything" |
| `$breezing` | Full team execution (Lead/Worker/Reviewer) | "breezing", "team run" |
| `$harness-review` | Code review, quality checks | "review this" |
| `$harness-setup` | Project initialization | "setup", "initialize" |
| `$harness-sync` | Sync implementation with Plans.md | "check progress" |

## Development Flow

1. **Plan**: `$harness-plan` to add tasks to Plans.md
2. **Implement**: `$harness-work` or `$breezing` to execute tasks
3. **Review**: `$harness-review` for quality checks
4. **Validate**: `./tests/validate-plugin.sh`

## Runtime Behavior

- `$harness-work` and `$breezing` use Codex native multi-agent orchestration
- Native flow uses `spawn_agent`, `wait`, `send_input`, `resume_agent`, `close_agent`

## SSOT (Single Source of Truth)

- `.claude/memory/decisions.md` - Decisions (Why)
- `.claude/memory/patterns.md` - Reusable patterns (How)

## Test Tampering Prevention

Absolutely prohibited. Fix the implementation, never modify tests to make them pass.

## Rules

`$CODEX_HOME/rules/harness.rules` provides command guardrails.
