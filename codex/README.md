# Harness for Codex CLI

Codex CLI compatible distribution of Claude Code Harness.

## Setup

### Option 1: Script (recommended)

```bash
# From your target project
/path/to/claude-code-harness/scripts/setup-codex.sh
```

### Option 1.5: Claude Code (in-session)

If you use Claude Code Harness, you can run:

```
/setup codex
```

### Option 2: Manual

```bash
# Clone Harness
git clone https://github.com/Chachamaru127/claude-code-harness.git

# Copy Codex Team Config
cp -r claude-code-harness/codex/.codex your-project/.codex

# Copy AGENTS.md (project instructions)
cp claude-code-harness/codex/AGENTS.md your-project/AGENTS.md
```

## How to Use

- Use `$skill-name` to explicitly invoke skills.
- Codex reads skills from `.codex/skills/<skill-name>/SKILL.md`.
- Custom prompts like `/work` are not the primary path for Codex.

## Rules (temporary)

- `.codex/rules/harness.rules` provides temporary guardrails.
- Each rule includes `HOOK:<id>` so it can be migrated to Hooks later.

## MCP (optional)

If you want MCP integration, copy the template:

```bash
cp codex/.codex/config.toml your-project/.codex/config.toml
```

Then edit the `mcp_servers.harness` path to your local build.

## Notes

- Codex Hooks are not supported yet. Use Rules until Hooks land.
