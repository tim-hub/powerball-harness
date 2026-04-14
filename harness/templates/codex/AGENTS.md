# AGENTS.md — Codex Agent Configuration

> For full project rules and development workflow, see [CLAUDE.md](CLAUDE.md).

## Agent Roles

| Agent | Role | Access |
|-------|------|--------|
| `implementer` | Implements tasks from Plans.md | Read/Write |
| `reviewer` | Reviews code and plans | Read-only |
| `task_worker` | Breezing standard implementation worker | Read/Write |
| `code_reviewer` | Breezing independent code review | Read-only |
| `plan_analyst` | Phase 0 planning analyst | Read-only |
| `plan_critic` | Phase 0 plan red-team reviewer | Read-only |

## Skill Loading

Harness skills are available in `.codex/skills/`. Codex CLI loads them from this directory.

## Key Skills

| Skill | Purpose |
|-------|---------|
| `harness-plan` | Ideas → Plans.md |
| `harness-work` | Task implementation (Codex native) |
| `harness-review` | Multi-angle code review |
| `harness-setup` | Project initialization |
| `breezing` | Full team-mode execution (Codex native) |

## Rules

Guardrail rules are in `.codex/rules/harness.rules`.
When Codex hooks become available, these rules will migrate to hook-based enforcement.
