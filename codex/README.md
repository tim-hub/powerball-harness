# Harness for Codex CLI

Codex CLI compatible distribution of Claude Code Harness.

## Setup

### Option 1: Script (recommended, user-based)

```bash
# Default: install to CODEX_HOME (user-based)
/path/to/claude-code-harness/scripts/setup-codex.sh --user --skip-mcp
```

Project-local install is still available:

```bash
/path/to/claude-code-harness/scripts/setup-codex.sh --project --skip-mcp
```

### Option 1.5: Claude Code (in-session)

If you use Claude Code Harness, you can run:

```
/setup codex
```

### Option 2: Manual (user-based)

```bash
# Clone Harness
git clone https://github.com/Chachamaru127/claude-code-harness.git

# User-level Codex home
CODEX_HOME="${CODEX_HOME:-$HOME/.codex}"
mkdir -p "$CODEX_HOME/skills" "$CODEX_HOME/rules"

# Copy Codex Harness skills/rules
cp -R claude-code-harness/codex/.codex/skills/* "$CODEX_HOME/skills/"
cp -R claude-code-harness/codex/.codex/rules/* "$CODEX_HOME/rules/"

# Optional: project-level AGENTS.md (only if needed)
cp claude-code-harness/codex/AGENTS.md /path/to/project/AGENTS.md
```

## How to Use

- Use `$skill-name` to explicitly invoke skills.
- Codex reads skills from `$CODEX_HOME/skills/<skill-name>/SKILL.md` (user) and `.codex/skills/...` (project override).
- Custom prompts like `/work` are not the primary path for Codex.

## Rules

- `$CODEX_HOME/rules/harness.rules` provides guardrails.

## MCP (optional)

If you want user-level MCP integration, copy the template:

```bash
CODEX_HOME="${CODEX_HOME:-$HOME/.codex}"
cp codex/.codex/config.toml "$CODEX_HOME/config.toml"
```

Then edit the `mcp_servers.harness` path to your local build.

## Notes

- Codex supports hook events internally.
