# Harness Setup — Codex CLI

Reference for `harness-setup codex`: install Codex CLI and set up Harness skills in your project.

## Prerequisites

- Node.js 20+ (for `npm install`)
- A project already initialized with `harness-setup init`

## Installation

```bash
npm install -g @openai/codex
```

Verify:

```bash
codex --version
```

## Setup

Run the setup subcommand from Claude Code:

```
harness-setup codex
```

This copies Harness config and skills to your project's `.codex/` directory:

```
project/
├── .codexignore          — Codex ignore patterns
├── AGENTS.md             — Agent role reference (points to CLAUDE.md)
└── .codex/
    ├── config.toml       — Multi-agent config (8 named agent profiles)
    ├── rules/
    │   └── harness.rules — Guardrail rules (until Codex supports hooks)
    └── skills/           — All Harness skills
        ├── harness-work/ — Codex-native implementation skill (override)
        ├── breezing/     — Codex-native team-execution skill (override)
        └── ...           — All other skills from skills/ (patched with disable-model-invocation)
```

## Running Codex with Harness Skills

```bash
# Interactive session
codex

# Run a specific task non-interactively
codex exec "implement task 3 from Plans.md"

# Delegate from a Claude Code session (recommended)
bash "${CLAUDE_PLUGIN_ROOT}/scripts/codex-companion.sh" task --write "your prompt"
```

See `.claude/rules/codex-cli-only.md` for the full policy on Codex invocation.

## Re-running Setup

Setup is idempotent — re-running `harness-setup codex` overwrites `.codex/skills/` with fresh copies from the plugin. Run this after:

- Updating the Harness plugin
- Adding new skills to the plugin

## Troubleshooting

| Symptom | Fix |
|---------|-----|
| `codex: command not found` | Re-run `npm install -g @openai/codex` |
| Skills not loading | Re-run `harness-setup codex` to refresh `.codex/skills/` |
| Config not found | Re-run `harness-setup codex` to recopy `config.toml` |
| Multi-agent disabled | Ensure `[features] multi_agent = true` is in `.codex/config.toml` |
