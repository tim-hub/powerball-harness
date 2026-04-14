# Harness for OpenCode

OpenCode configuration template for Claude Code Harness.

## Setup

Run `harness-setup opencode` from Claude Code to install Harness skills and configuration into your project's `.opencode/` directory.

## What Gets Installed

- `.opencode/opencode.json` — OpenCode project configuration
- `.opencode/commands/` — OpenCode slash commands (handoff-to-claude, plan-with-cc, etc.)
- `.opencode/skills/` — All Harness skills (same as Claude Code skills)

## Usage

Launch OpenCode in your project directory:

```bash
opencode
```

Harness skills will be available as slash commands.

## Related

- [Claude Code Harness](https://github.com/tim-hub/powerball-harness)
- [OpenCode Documentation](https://opencode.ai/docs/)
