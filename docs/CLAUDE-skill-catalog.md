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

## Skill Directory (27 skills)

Skills are flat — each skill lives in its own directory with a `SKILL.md`. Routing is description-based (auto-loaded by trigger matching).

```
harness/skills/
├── harness-work/          # Task implementation (Plans.md tasks, parallel workers, breezing)
├── harness-plan/          # Plans.md authoring — create tasks, add acceptance criteria
├── harness-review/        # Code/plan/scope review — pre-merge quality gate, security, performance
├── harness-release/       # Generic release engine: version bumps, CHANGELOG, git tags, GitHub Releases (any project)
├── harness-setup/         # Project init, CI/Codex/memory config, binary download
├── harness-sync/          # Plans.md ↔ implementation drift detection and marker updates
├── harness-loop/          # Autonomous ScheduleWakeup-based loop runtime with sprint-contracts
├── breezing/              # Full team end-to-end run with parallel Workers (auto-detects task count)
├── memory/                # SSOT management — decisions.md, patterns.md, memory search
├── maintenance/           # Periodic cleanup — session log pruning, stale state, cache purge
├── session/               # Session lifecycle: list, inbox checks, broadcast
├── session-init/          # Session start — status check, Plans.md overview, env readiness
├── session-state/         # Internal — auto-triggered at harness-work phase boundaries
├── session-control/       # Internal — auto-triggered for --resume/--fork flags
├── session-memory/        # Cross-session context recall and persistence
├── ci/                    # CI/CD failure diagnosis — GitHub Actions, pipeline errors
├── agent-browser/         # Browser automation — URLs, forms, scraping, UI testing
├── cc-cursor-cc/          # Cursor ↔ Claude Code handoff — PM plan validation, Plans.md sync
├── crud/                  # Scaffold CRUD — API endpoints, database models, REST resources
├── deploy/                # Deploy to Vercel or Netlify — health checks, monitoring
├── auth/                  # Authentication, OAuth, sessions, payments (Clerk, Supabase, Stripe)
├── ui/                    # UI components — hero sections, landing pages, feedback forms
├── gogcli-ops/            # Google Drive/Docs/Sheets/Slides via gogcli
├── notebook-lm/           # NotebookLM YAML, structured slide content
├── principles/            # Coding principles, development guidelines, safe-editing practices
├── vibecoder-guide/       # Plain-language workflow guidance for newcomers
└── workflow-guide/        # 2-agent workflow (Cursor ↔ Claude Code roles)
```

## Full Skill Category Listing

| Skill | Purpose | Trigger Examples |
|-------|---------|-----------|
| `harness-work` | Task implementation (auto-scope detection, parallel workers) | "implement", "do it all", "/harness-work" |
| `harness-plan` | Create/update Plans.md with tasks and acceptance criteria | "plan", "add task", "/harness-plan" |
| `harness-review` | Code review, quality checks, security audit | "review this", "security", "performance" |
| `harness-release` | Generic release engine: version bump, CHANGELOG, tag, GitHub Release (usable by any project) | "release", "tag", "publish" |
| `release-this` | Plugin-specific release: build-all → checks → harness-release (use this to release THIS plugin) | "release this", "release plugin", "publish harness" |
| `harness-setup` | Project init, binary download, CI config | "setup", "initialize", "install binary" |
| `harness-sync` | Drift detection between Plans.md and implementation | "sync", "check drift", "update markers" |
| `harness-loop` | Autonomous loop runtime with ScheduleWakeup and sprint-contracts | "loop", "autonomous run", "harness-loop" |
| `breezing` | Full auto-run with parallel Agent Teams | "run with team", "breezing", "all tasks" |
| `memory` | SSOT management, decisions.md, patterns.md | "SSOT", "decisions", "memory search" |
| `maintenance` | Periodic housekeeping — log pruning, stale state, worktrees | "prune logs", "clean state", "/maintenance" |
| `session` | Session lifecycle: list, inbox, broadcast | "/session", "session status" |
| `session-init` | Session start check — Plans.md overview, env readiness | (auto-triggered on session start) |
| `ci` | Diagnose CI/CD failures — GitHub Actions errors | "CI failed", "pipeline broken" |
| `agent-browser` | Browser automation — scraping, screenshots, UI testing | "open URL", "click", "screenshot" |
| `cc-cursor-cc` | Cursor ↔ Claude Code handoff | "hand off to Cursor", "sync plans" |
| `crud` | Scaffold CRUD API, database models | "create endpoint", "scaffold CRUD" |
| `deploy` | Deploy to Vercel/Netlify, health checks | "deploy", "Vercel", "Netlify" |
| `auth` | Auth, OAuth, sessions, payments | "login", "Clerk", "Stripe", "payments" |
| `ui` | UI component generation | "component", "hero", "landing page form" |
| `gogcli-ops` | Google Drive/Docs/Sheets via gogcli | "Google Drive", "Sheets", "gogcli" |
| `notebook-lm` | NotebookLM YAML, slide content | "document", "NotebookLM", "slides" |
| `principles` | Development principles, guidelines | "principles", "VibeCoder", "safety" |
| `vibecoder-guide` | Newcomer-friendly workflow guide | "how does this work", "explain workflow" |
| `workflow-guide` | 2-agent workflow reference | "cursor role", "claude code role", "2-agent" |

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
