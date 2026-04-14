---
name: harness-setup
description: "Use when initializing a project, setting up CI/Codex/memory config, configuring 2-agent workflow, or running /harness-setup. Do NOT load for: implementation, review, release, or planning."
allowed-tools: ["Read", "Write", "Edit", "Grep", "Glob", "Bash"]
argument-hint: "[init|binary|codex|opencode|duo|cleanup|gitignore]"
effort: medium
---

# Harness Setup

## Quick Reference

| User Input | Subcommand | Behavior |
|------------|------------|----------|
| `harness-setup` (no args) | `init` | Runs `binary` → `gitignore` → project initialization (CLAUDE.md + Plans.md + settings.json) |
| `harness-setup binary` | `binary` | Check or build/install the platform binary from Go source |
| `harness-setup init` | `init` | New project initialization: binary build → gitignore → CLAUDE.md + Plans.md + settings.json |
| `harness-setup gitignore` | `gitignore` | Merge harness-managed block into .gitignore (runs `scripts/merge-gitignore.sh`) |
| `harness-setup cleanup` | `cleanup` | Periodic maintenance: delete old logs, compress Plans.md, trim traces |
| `harness-setup codex` | `codex` | Set up Codex CLI: copy config, rules, and skills to project `.codex/` |
| `harness-setup opencode` | `opencode` | Set up OpenCode: copy config, commands, and skills to project `.opencode/` |
| `harness-setup duo` | `duo` | Set up both Codex and OpenCode in one step |

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
    ├── memory/          # Harness SSOT (decisions.md + patterns.md)
    ├── output-styles/   # Custom output styles (if any)
    ├── rules/           # Custom rules (if any)
    ├── scripts/         # Custom scripts (if any)
    └── skills/          # Custom skills (if any)
    └── settings.local.json # Local custom settings (gitignored) for user overrides
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

### codex — Codex CLI Setup

Sets up the project for Codex CLI by copying Harness config templates and skills.

**Prerequisites**: Codex CLI installed (`npm install -g @openai/codex`)

**What it does**:
1. Checks that `codex` is installed; prints install instructions if not
2. Copies `templates/codex/*` → `.codex/` (config.toml, rules/harness.rules, .codexignore)
3. Copies `AGENTS.md` to the project root
4. Copies all skills from plugin `skills/` → `.codex/skills/` with `disable-model-invocation: true` patched into each SKILL.md frontmatter
5. Overlays `templates/codex-skills/` → `.codex/skills/` (codex-native variants of breezing and harness-work override the generic copies)

```bash
bash "${CLAUDE_PLUGIN_ROOT}/skills/harness-setup/scripts/setup-codex.sh"
```

> **Note**: Excluded from the no-args `init` run — requires an external `@openai/codex` install that should be explicit opt-in.

### opencode — OpenCode Setup

Sets up the project for OpenCode by copying Harness config templates, commands, and skills.

**Prerequisites**: OpenCode installed (see https://opencode.ai/)

**What it does**:
1. Checks that `opencode` is installed; prints install instructions if not
2. Copies `templates/opencode/opencode.json` → `.opencode/opencode.json`
3. Copies `templates/opencode/commands/` → `.opencode/commands/`
4. Copies `AGENTS.md` to the project root
5. Copies all skills from plugin `skills/` → `.opencode/skills/` as-is (opencode ignores unknown frontmatter fields)

```bash
bash "${CLAUDE_PLUGIN_ROOT}/skills/harness-setup/scripts/setup-opencode.sh"
```

> **Note**: Excluded from the no-args `init` run — requires OpenCode to be installed explicitly.

### duo — Codex + OpenCode Setup

Runs both `codex` and `opencode` setup subcommands in sequence.

```bash
bash "${CLAUDE_PLUGIN_ROOT}/skills/harness-setup/scripts/setup-codex.sh"
bash "${CLAUDE_PLUGIN_ROOT}/skills/harness-setup/scripts/setup-opencode.sh"
```

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
