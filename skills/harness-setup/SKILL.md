---
name: harness-setup
description: "Use when initializing a project, setting up CI/Codex/memory config, configuring 2-agent workflow, or running /harness-setup. Do NOT load for: implementation, review, release, or planning."
allowed-tools: ["Read", "Write", "Edit", "Grep", "Glob", "Bash"]
argument-hint: "[init|ci|codex|harness-mem|mirrors|agents|localize]"
effort: medium
---

# Harness Setup

## Quick Reference

| Subcommand | Behavior |
|------------|----------|
| `harness-setup` (no args) | Run binary install check, then init |
| `harness-setup binary` | Download/install the platform binary from GitHub releases |
| `harness-setup init` | New project initialization (CLAUDE.md + Plans.md + hooks) |
| `harness-setup ci` | CI/CD pipeline configuration |
| `harness-setup codex` | Codex CLI installation and configuration |
| `harness-setup harness-mem` | harness-mem integration and memory configuration |
| `harness-setup mirrors` | skills/ → public mirror bundle update |
| `harness-setup agents` | agents/ agent configuration |
| `harness-setup localize` | CLAUDE.md rule localization |

## Subcommand Details

### binary — Platform Binary Install

Downloads and installs the `harness-<os>-<arch>` binary from the GitHub release into `$CLAUDE_PLUGIN_ROOT/bin/`.
Run this first if hooks are silently passing through (binary not yet installed).

Implementation: [`scripts/download-binary.sh`](${CLAUDE_SKILL_DIR}/scripts/download-binary.sh)

```bash
bash "${CLAUDE_PLUGIN_ROOT}/skills/harness-setup/scripts/download-binary.sh"
```

**When to run**: After fresh plugin install if you see `UserPromptSubmit hook error` messages.

### init — Project Initialization

Introduce Harness to a new project.

**Generated files** (user's project):
```
project/
├── CLAUDE.md            # Project configuration
├── Plans.md             # Task management (empty template)
├── .gitignore           # Standard ignore rules (harness-managed block appended)
└── .claude/
    └── settings.json    # Claude Code permissions/sandbox/env
```

> **Note**: Neither `hooks/` nor `harness.toml` is generated into a user's project.
> - Hooks ship inside the installed plugin (`.claude-plugin/hooks.json`) — Claude Code
>   loads them from there automatically.
> - `harness.toml` + `harness sync` is a *plugin-author* workflow for regenerating
>   `.claude-plugin/*` files from a single TOML SSOT. User projects have no
>   `.claude-plugin/` to regenerate, so the TOML would be an orphaned file. Users who
>   want unified TOML authoring for their own `.claude/settings.json` can opt in later
>   via a dedicated subcommand (future work).

**Flow**:
1. Detect project type (Node.js/Python/Go/Rust/Other)
2. Generate minimal CLAUDE.md
3. Generate Plans.md template
4. Generate `.claude/settings.json` (permissions/sandbox/env — safe defaults)
5. Merge `templates/gitignore-harness` into `.gitignore` (idempotent — skips if `# >>> harness-managed >>>` marker already present)

**gitignore merge logic**:
```bash
MARKER="# >>> harness-managed >>>"
if grep -qF "$MARKER" .gitignore 2>/dev/null; then
  echo ".gitignore already contains harness-managed block — skipping"
else
  echo "" >> .gitignore
  cat "${CLAUDE_PLUGIN_ROOT}/templates/gitignore-harness" >> .gitignore
  echo "Appended harness-managed gitignore block"
fi
```

The block ignores `.claude/sessions/`, `logs/`, `settings.local.json`, and `states/`,
while force-tracking `.claude/memory/decisions.md`, `.claude/memory/patterns.md`,
`.claude/output-styles/`, `.claude/rules/`, `.claude/scripts/`, `.claude/skills/`,
and `.claude/settings.json`.

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
mkdir -p .claude/agent-memory/powerball-harness-worker
mkdir -p .claude/agent-memory/powerball-harness-reviewer

# Deploy MEMORY.md template
cat > .claude/agent-memory/powerball-harness-worker/MEMORY.md << 'EOF'
# Worker Agent Memory

## Project Context
[Project overview]

## Patterns
[Learned patterns]
EOF
```

### mirrors — Public Skill Bundle Sync

On Windows with `core.symlinks=false`, repository symlinks become regular files, and `harness-*` skills may not appear in the command list. Public bundles are synced as real directory mirrors.

Codex skills are symlinks to `skills/` — no manual sync needed:

```bash
ls -la codex/.codex/skills/
```

Update targets:

- `skills/` (SSOT)
- `codex/.codex/skills/` (symlinks → `../../../skills/`)

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
