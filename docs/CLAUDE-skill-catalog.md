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
├── harness-release/       # Generic release engine: version bumps, CHANGELOG, git tags, GitHub Releases (any project)
├── harness-setup/         # Project init, CI/Codex/memory config, binary download
├── harness-schedule-run/  # Autonomous ScheduleWakeup-based scheduled run runtime with sprint-contracts
├── breezing/              # Full team end-to-end run with parallel Workers (auto-detects task count)
├── harness-remember/      # SSOT management — decisions.md, patterns.md, memory search
├── maintenance/           # Periodic cleanup — session log pruning, stale state, cache purge
├── session/               # Session lifecycle: list, inbox checks, broadcast, init, resume/fork
├── harness-guide/         # Workflow orientation and plain-language guidance
└── harness-ralph-loop/    # Iterative loop orchestrator for [ralph]-marked tasks
```

## Full Skill Category Listing

| Skill | Purpose | Trigger Examples |
|-------|---------|-----------|
| `harness-work` | Task implementation (auto-scope detection, parallel workers) | "implement", "do it all", "/harness-work" |
| `harness-plan` | Create/update Plans.md — create, add, sync, brainstorm subcommands | "plan", "add task", "brainstorm", "/harness-plan" |
| `harness-review` | Code review, quality checks, security audit | "review this", "security", "performance" |
| `harness-release` | Generic release engine: version bump, CHANGELOG, tag, GitHub Release (usable by any project) | "release", "tag", "publish" |
| `release-this` | Plugin-specific release: build-all → checks → harness-release (use this to release THIS plugin) | "release this", "release plugin", "publish harness" |
| `harness-setup` | Project init, binary download, CI config | "setup", "initialize", "install binary" |
| `harness-schedule-run` | Autonomous scheduled run runtime with ScheduleWakeup and sprint-contracts (renamed from `harness-loop`) | "scheduled run", "autonomous run", "harness-schedule-run" |
| `breezing` | Full auto-run with parallel Agent Teams | "run with team", "breezing", "all tasks" |
| `harness-remember` | SSOT management, decisions.md, patterns.md | "SSOT", "decisions", "memory search" |
| `maintenance` | Periodic housekeeping — log pruning, stale state, worktrees | "prune logs", "clean state", "/maintenance" |
| `session` | Session lifecycle: list, inbox, broadcast, init, resume/fork | "/session", "session status" |
| `harness-guide` | Workflow orientation and plain-language guidance | "how does this work", "explain workflow", "what do I do next", "execution modes" |

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
