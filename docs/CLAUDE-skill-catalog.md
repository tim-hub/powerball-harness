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

## Skill Hierarchy

Skills are organized in a hierarchy of **parent skills (categories)** and **child skills (specific features)**.

```
harness/skills/
├── harness-review/        # Review (quality, security, performance)
├── setup/                 # Integrated setup (project init, tool config, 2-Agent, harness-mem, Codex CLI, rule localization)
├── memory/                # Memory management (SSOT, decisions.md, patterns.md, SSOT promotion, memory search)
├── principles/            # Principles and guidelines (VibeCoder, diff editing)
├── auth/                  # Authentication and payments (Clerk, Supabase, Stripe)
├── deploy/                # Deployment (Vercel, Netlify, analytics)
├── ui/                    # UI (components, feedback)
└── notebook-lm/           # Documentation (NotebookLM, YAML)
```

**Usage:**
1. Launch the parent skill with the Skill tool
2. The parent skill routes to the appropriate child skill (doc.md) based on user intent
3. Execute work following the child skill's procedures

## Full Skill Category Listing

| Category | Purpose | Trigger Examples |
|---------|------|-----------|
| work | Task implementation (auto-scope detection, --codex support) | "implement", "do it all", "/work" |
| breezing | Full auto-run with Agent Teams (--codex support) | "run with team", "breezing" |
| harness-review | Code review, quality checks | "review this", "security", "performance" |
| setup | Integrated setup hub (project init, tool config, 2-Agent, harness-mem, Codex CLI, rule localization) | "setup", "CLAUDE.md", "initialize", "CI setup", "2-Agent", "Cursor config", "harness-mem", "codex-setup" |
| memory | SSOT management, memory search, SSOT promotion, Cursor-linked memory | "SSOT", "decisions.md", "merge", "SSOT promotion", "memory search", "harness-mem" |
| principles | Development principles, guidelines | "principles", "VibeCoder", "safety" |
| auth | Authentication, payment features | "login", "Clerk", "Stripe", "payments" |
| deploy | Deployment, analytics | "deploy", "Vercel", "GA" |
| ui | UI component generation | "component", "hero", "form" |
| notebook-lm | Document generation | "document", "NotebookLM", "slides" |

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
