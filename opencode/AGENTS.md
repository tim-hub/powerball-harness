# AGENTS.md

This project uses [claude-code-harness](https://github.com/Chachamaru127/claude-code-harness) workflow.

## Available Commands

### Core Commands

| Command | Description |
|---------|-------------|
| `/harness-init` | Project setup |
| `/plan-with-agent` | Create development plan |
| `/work` | Execute tasks |
| `/harness-review` | Code review |
| `/sync-status` | Check project status |

### Optional Commands

| Command | Description |
|---------|-------------|
| `/harness-update` | Update Harness |
| `/mcp-setup` | Setup MCP server |
| `/lsp-setup` | Setup LSP |

## Workflow

```
/plan-with-agent → /work → /harness-review → commit
```

## MCP Integration

This project includes an MCP server for cross-client communication.
See `opencode.json` for configuration.

## More Information

- [Harness Documentation](https://github.com/Chachamaru127/claude-code-harness)
- [OpenCode Documentation](https://opencode.ai/docs/)
