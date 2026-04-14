# AGENTS.md — OpenCode Agent Configuration

> For full project rules and development workflow, see [CLAUDE.md](CLAUDE.md).

## Agent Roles

| Agent | Role | Access |
|-------|------|--------|
| `implementer` | Implements tasks from Plans.md | Read/Write |
| `reviewer` | Reviews code and plans | Read-only |
| `task_worker` | Breezing standard implementation worker | Read/Write |
| `code_reviewer` | Breezing independent code review | Read-only |

## Skill Loading

Harness skills are available in `.opencode/skills/`. OpenCode loads them automatically from this directory.

## Key Skills

| Skill | Purpose |
|-------|---------|
| `harness-plan` | Ideas → Plans.md |
| `harness-work` | Task implementation |
| `harness-review` | Multi-angle code review |
| `harness-setup` | Project initialization |
| `breezing` | Full team-mode execution |
