---
name: harness-setup
description: "Use this skill whenever the user mentions setup, initialization, starting a new project, CI setup, Codex CLI setup, harness-mem, agent configuration, symlinks, mirrors, or runs /harness-setup. Also use when the user needs to configure the harness environment or onboard a new repository. Do NOT load for: code implementation (use harness-work), code review (use harness-review), release (use harness-release), or planning (use harness-plan). Unified setup skill for Harness v3 — project initialization, tool configuration, 2-agent setup, memory config, symlinks, and mirror sync."
allowed-tools: ["Read", "Write", "Edit", "Grep", "Glob", "Bash"]
argument-hint: "[init|ci|codex|harness-mem|mirrors|agents|localize]"
effort: medium
---

# Harness Setup (v3)

Unified setup skill for Harness v3.
Consolidates the following legacy skills:

- `setup` — Unified setup hub
- `harness-init` — Project initialization
- `harness-update` — Harness updates
- `maintenance` — File cleanup and organization

## Quick Reference

| Subcommand | Behavior |
|------------|----------|
| `harness-setup init` | New project initialization (CLAUDE.md + Plans.md + hooks) |
| `harness-setup ci` | CI/CD pipeline configuration |
| `harness-setup codex` | Codex CLI installation and configuration |
| `harness-setup harness-mem` | harness-mem integration and memory configuration |
| `harness-setup mirrors` | skills/ → public mirror bundle update |
| `harness-setup agents` | agents/ agent configuration |
| `harness-setup localize` | CLAUDE.md rule localization |

## Subcommand Details

### init — Project Initialization

Introduce Harness v3 to a new project.

**Generated files**:
```
project/
├── CLAUDE.md            # Project configuration
├── Plans.md             # Task management (empty template)
├── .claude/
│   ├── settings.json    # Claude Code configuration
│   └── hooks.json       # Hook configuration (v3 shim)
└── hooks/
    ├── pre-tool.sh      # Thin shim (→ core/src/index.ts)
    └── post-tool.sh     # Thin shim (→ core/src/index.ts)
```

**Flow**:
1. Detect project type (Node.js/Python/Go/Rust/Other)
2. Generate minimal CLAUDE.md
3. Generate Plans.md template
4. Deploy hooks.json

### ci — CI/CD Configuration

Configure GitHub Actions workflows.

```yaml
# .github/workflows/ci.yml generation example
name: CI
on:
  push:
    branches: [main]
  pull_request:
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: npm ci && npm test
```

### codex — Codex CLI Configuration

```bash
# Installation check
which codex || npm install -g @openai/codex

# Timeout command check (macOS)
TIMEOUT=$(command -v timeout || command -v gtimeout || echo "")
# For macOS: brew install coreutils
```

**Usage patterns** (via official plugin):
```bash
bash scripts/codex-companion.sh task --write "task content"
# Or via stdin
cat /tmp/prompt.md | bash scripts/codex-companion.sh task --write
```

### harness-mem — Memory Configuration

Configure Unified Harness Memory.

```bash
# Create memory directories
mkdir -p .claude/agent-memory/claude-code-harness-worker
mkdir -p .claude/agent-memory/claude-code-harness-reviewer

# Deploy MEMORY.md template
cat > .claude/agent-memory/claude-code-harness-worker/MEMORY.md << 'EOF'
# Worker Agent Memory

## Project Context
[Project overview]

## Patterns
[Learned patterns]
EOF
```

### mirrors — Public Skill Bundle Sync

On Windows with `core.symlinks=false`, repository symlinks become regular files, and `harness-*` skills may not appear in the command list. Public bundles are synced as real directory mirrors.

```bash
./scripts/sync-v3-skill-mirrors.sh
./scripts/sync-v3-skill-mirrors.sh --check
```

Update targets:

- `skills/`
- `codex/.codex/skills/`
- `opencode/skills/`

### agents — Agent Configuration

Configure the 3-agent structure in agents/.

```
agents/
├── worker.md      # Implementation (task-worker + codex-implementer + error-recovery)
├── reviewer.md    # Review (code-reviewer + plan-critic)
└── scaffolder.md  # Scaffolding (project-analyzer + scaffolder)
```

### localize — Rule Localization

Adapt `.claude/rules/` rules to the current project.

```bash
# Check rule list
ls .claude/rules/

# Add project-specific rules
cat >> .claude/rules/project-rules.md << 'EOF'
# Project-Specific Rules
[Project-specific rules here]
EOF
```

## Plugin Installation (v2.1.71+ Marketplace)

Marketplace stability was significantly improved in v2.1.71.

### Recommended Installation Method

```bash
# Pin version with @ref format (recommended)
claude plugin install owner/repo@v3.5.0

# Latest version
claude plugin install owner/repo
```

The `owner/repo@vX.X.X` format is recommended. With the `@ref` parser fix, tags, branches, and commit hashes are all resolved accurately.

### Updates

```bash
claude plugin update owner/repo
```

Merge conflicts during updates were fixed in v2.1.71, enabling stable updates.

### Other Improvements

- MCP server deduplication: Automatically prevents duplicate registration of the same MCP server
- `/plugin uninstall` uses `settings.local.json`: Accurately reflected in user-local settings

## Maintenance — File Cleanup

Periodic maintenance tasks:

| Task | Command |
|------|---------|
| Delete old logs | `find .claude/logs -mtime +30 -delete` |
| Compress Plans.md | Move completed tasks to an archive section |
| Delete old traces | `tail -1000 .claude/state/agent-trace.jsonl > /tmp/trace && mv /tmp/trace .claude/state/agent-trace.jsonl` |

## Related Skills

- `harness-plan` — Create project plans after setup
- `harness-work` — Execute tasks after setup
- `harness-review` — Review setup configuration
