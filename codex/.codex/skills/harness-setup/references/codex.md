# Harness Setup — Codex CLI Configuration

Reference for `harness-setup codex`: install and configure Codex CLI with Harness project skills.

## Prerequisites

- Node.js 20+ (for `npm install`)
- A project already initialized with `harness-setup init`

## Installation

```bash
npm install -g @openai/codex
```

Verify the installation:

```bash
codex --version
```

## Project Configuration

Copy the Harness team config template to your project root:

```bash
cp "${CLAUDE_PLUGIN_ROOT}/codex/.codex/config.toml" .codex/config.toml
```

Or for user-wide config (applies to all projects):

```bash
mkdir -p ~/.codex
cp "${CLAUDE_PLUGIN_ROOT}/codex/.codex/config.toml" ~/.codex/config.toml
```

The config enables multi-agent mode (`multi_agent = true`) and defines named agent profiles (`implementer`, `reviewer`, `task_worker`, `code_reviewer`) that Harness breezing mode uses.

## Skill Sync

Harness skills are pre-configured as symlinks in `codex/.codex/skills/` → `skills/`. These are already set up in the plugin. No manual action is needed.

Verify the symlinks are healthy:

```bash
ls -la codex/.codex/skills/
```

Each entry should resolve to `../../../skills/<skill-name>`.

## Running Codex with Harness Skills

```bash
# Interactive session with all Harness skills available
codex

# Run a specific task non-interactively
codex exec "implement task 3 from Plans.md"

# Delegate from a Claude Code session (recommended)
bash "${CLAUDE_PLUGIN_ROOT}/scripts/codex-companion.sh" task --write "your prompt"
```

See `.claude/rules/codex-cli-only.md` for the full policy on Codex invocation (always use `codex-companion.sh` from within Harness, not raw `codex exec`).

## Verification

After setup, confirm Codex can see Harness skills:

```bash
codex list-skills 2>/dev/null | grep harness || echo "skills loaded via .codex/skills/ directory"
ls .codex/skills/
```

## Troubleshooting

| Symptom | Fix |
|---------|-----|
| `codex: command not found` | Re-run `npm install -g @openai/codex` |
| Skills not loading | Verify `.codex/skills/` symlinks resolve: `ls -la .codex/skills/` |
| Config not found | Copy `codex/.codex/config.toml` to `.codex/config.toml` in project root |
| Multi-agent disabled | Ensure `[features] multi_agent = true` is in config |
