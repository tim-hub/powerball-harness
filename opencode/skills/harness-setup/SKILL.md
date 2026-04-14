---
name: harness-setup
description: "Use when initializing a project, setting up CI/Codex/memory config, configuring 2-agent workflow, or running /harness-setup. Do NOT load for: implementation, review, release, or planning."
allowed-tools: ["Read", "Write", "Edit", "Grep", "Glob", "Bash"]
argument-hint: "[init|binary|codex|cleanup|gitignore]"
effort: medium
---

# Harness Setup

## Quick Reference

| User Input | Subcommand | Behavior |
|------------|------------|----------|
| `harness-setup` (no args) | `init` | Runs `binary` → `gitignore` → project initialization (CLAUDE.md + Plans.md + settings.json) |
| `harness-setup binary` | `binary` | Check or download/install the platform binary from GitHub releases |
| `harness-setup init` | `init` | New project initialization: binary download → gitignore → CLAUDE.md + Plans.md + settings.json |
| `harness-setup gitignore` | `gitignore` | Merge harness-managed block into .gitignore (runs `scripts/merge-gitignore.sh`) |
| `harness-setup cleanup` | `cleanup` | Periodic maintenance: delete old logs, compress Plans.md, trim traces |
| `harness-setup codex` | `codex` | Codex CLI installation and configuration (see `references/codex.md`) |

## Subcommand Details

### binary — Platform Binary Build

Builds the `harness-<os>-<arch>` binary from Go source and installs it into `$CLAUDE_PLUGIN_ROOT/bin/`.
Run this first if hooks are silently passing through (binary not yet installed).
Requires `go` to be installed on the system.

Implementation: [`scripts/build-binary.sh`](${CLAUDE_SKILL_DIR}/scripts/build-binary.sh)

```bash
bash "${CLAUDE_PLUGIN_ROOT}/skills/harness-setup/scripts/build-binary.sh"
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
1. Run the `binary` subcommand (download/install platform binary if not already present)
2. Run the `gitignore` subcommand (idempotent — calls `scripts/merge-gitignore.sh`)
3. Detect project type (Node.js/Python/Go/Rust/Other)
4. Generate minimal CLAUDE.md
5. Generate Plans.md template
6. Generate `.claude/settings.json` (permissions/sandbox/env — safe defaults)

### gitignore — Harness .gitignore Block

Merges the harness-managed block into the project's `.gitignore`. Safe to run multiple times — skips if the marker is already present.

Implementation: [`scripts/merge-gitignore.sh`](${CLAUDE_SKILL_DIR}/scripts/merge-gitignore.sh)

```bash
bash "${CLAUDE_PLUGIN_ROOT}/skills/harness-setup/scripts/merge-gitignore.sh"
# Or with an explicit target path:
bash "${CLAUDE_PLUGIN_ROOT}/skills/harness-setup/scripts/merge-gitignore.sh path/to/.gitignore"
```

The block ignores `.claude/sessions/`, `logs/`, `settings.local.json`, and `states/`,
while force-tracking `.claude/memory/`, `.claude/output-styles/`, `.claude/rules/`,
`.claude/scripts/`, `.claude/skills/`, and `.claude/settings.json`.

### codex — Codex CLI Configuration

See [`references/codex.md`](${CLAUDE_SKILL_DIR}/references/codex.md) for full setup and usage instructions.

> **Note**: This subcommand is intentionally excluded from the no-args run because it
> requires an external npm install (`@openai/codex`) that should be an explicit opt-in.

### Cleanup

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
