# Skill Catalog

Skill hierarchy, full category listing, and development skill reference documentation.

## Skill Evaluation Flow

> For heavy tasks (parallel reviews, CI fix loops), skills spawn sub-agents from `agents/` in parallel via the Task tool.

**Before starting work, always follow this flow:**

1. **Evaluate**: Check available skills and assess whether any match the current request
2. **Launch**: If a matching skill exists, launch it with the Skill tool before starting work
3. **Execute**: Follow the skill's procedures to carry out the work

```
User request
    |
Evaluate skills (is there a match?)
    |
YES -> Launch with Skill tool -> Follow skill procedures
NO  -> Handle with standard reasoning
```

## Skill Directory (21 skills)

Skills are flat — each skill lives in its own directory with a `SKILL.md`. Routing is description-based (auto-loaded by trigger matching).

```
harness/skills/
├── harness-work/          # Task implementation (Plans.md tasks, parallel workers, breezing)
├── harness-plan/          # Plans.md authoring — create/add/sync/brainstorm subcommands
├── harness-review/        # Code/plan/scope review — pre-merge quality gate, security, performance
├── harness-setup/         # Project init, CI/Codex/memory config, binary download
├── harness-remember/      # SSOT management — decisions.md, patterns.md, memory search
├── maintenance/           # Periodic cleanup and session lifecycle — pruning, state, cache, list/inbox/broadcast
└── harness-ralph-loop/    # Iterative loop orchestrator for [ralph]-marked tasks
```

## Full Skill Category Listing

| Skill | Purpose | Trigger Examples |
|-------|---------|-----------|
| `harness-work` | Task implementation (auto-scope detection, parallel workers) | "implement", "do it all", "/harness-work" |
| `harness-plan` | Create/update Plans.md — create, add, sync, brainstorm subcommands | "plan", "add task", "brainstorm", "/harness-plan" |
| `harness-review` | Code review, quality checks, security audit | "review this", "security", "performance" |
| `release-this` | Plugin release: build-all → checks → version bump → CHANGELOG → tag → GitHub Release | "release this", "release plugin", "publish harness" |
| `harness-setup` | Project init, binary download, CI config | "setup", "initialize", "install binary" |
| `harness-remember` | SSOT management, decisions.md, patterns.md | "SSOT", "decisions", "memory search" |
| `maintenance` | Periodic housekeeping and session lifecycle — log pruning, stale state, worktrees, list/inbox/broadcast | "prune logs", "clean state", "/maintenance", "list sessions", "session inbox" |

## Development Skills (Private)

The following skills are for development and experimentation, and are not included in the repository (excluded via .gitignore):

```
harness/skills/
├── test-*/      # Test skills
└── x-promo/     # X post creation skills (development use)
```

These skills are used only in individual development environments and should not be included in plugin distribution.

## Related Documentation

- [CLAUDE.md](../CLAUDE.md) - Project development guide (overview)
- [docs/CLAUDE-feature-table.md](./CLAUDE-feature-table.md) - Claude Code feature utilization table
- [docs/CLAUDE-commands.md](./CLAUDE-commands.md) - Key command reference
- [.claude/rules/skill-editing.md](../.claude/rules/skill-editing.md) - Skill file editing rules
