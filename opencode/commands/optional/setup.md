---
description: "[Optional] Unified setup command for CI, LSP, MCP, dev-tools, etc."
description-en: "[Optional] Unified setup command for CI, LSP, MCP, dev-tools, etc."
---

# /setup - Unified Setup Command

Consolidates all setup commands into one unified interface.

## Quick Reference

- "**Setup CI**" → `/setup ci`
- "**Setup LSP**" → `/setup lsp`
- "**Setup dev tools**" → `/setup dev-tools` (recommended)
- "**Setup MCP**" → `/setup mcp`
- "**Setup OpenCode**" → `/setup opencode`
- "**Setup webhook**" → `/setup webhook`
- "**Setup Harness UI**" → `/setup ui`

## Usage

```bash
/setup              # Show available options
/setup ci           # Setup CI/CD (GitHub Actions)
/setup dev-tools    # Setup AST-Grep + LSP (recommended)
/setup lsp          # Setup LSP only
/setup mcp          # Setup MCP server
/setup opencode     # Setup OpenCode.ai compatibility
/setup webhook      # Setup GitHub Actions webhook
/setup ui           # Setup Harness UI
```

---

## Subcommands

### `/setup ci` - CI/CD Setup

Sets up GitHub Actions for automated testing and deployment.

**Deliverables**:
- `.github/workflows/*.yml` for lint/typecheck/test/build
- Failure analysis and fix suggestions

**Features**:
- ✅ Lint (ESLint, Prettier)
- ✅ Type Check (TypeScript)
- ✅ Unit Test (Jest, Vitest)
- ✅ E2E Test (Playwright)
- ✅ Build Check

---

### `/setup dev-tools` - Development Tools (Recommended)

Sets up AST-Grep and LSP for enhanced code intelligence.

**Deliverables**:
- AST-Grep MCP server configuration
- LSP server configuration
- `.claude/settings.json` updates

**Why recommended**:
- Structural code search (more accurate than grep)
- Type-aware navigation and diagnostics
- Better code review accuracy

---

### `/setup lsp` - LSP Only

Sets up Language Server Protocol support without AST-Grep.

**Deliverables**:
- LSP server configuration for your language
- `.claude/settings.json` updates

---

### `/setup mcp` - MCP Server

Sets up Model Context Protocol servers.

**Deliverables**:
- MCP server configuration
- Tool availability verification

---

### `/setup opencode` - OpenCode.ai Compatibility

Sets up compatibility with OpenCode.ai.

**Deliverables**:
- OpenCode-compatible file structure
- Symlinks or copies as needed

---

### `/setup webhook` - GitHub Actions Webhook

Sets up webhook for GitHub Actions integration.

**Deliverables**:
- `.github/workflows/review.yml` for PR auto-review
- Webhook endpoint configuration

---

### `/setup ui` - Harness UI

Sets up the Harness UI dashboard.

**Deliverables**:
- UI server configuration
- Dashboard access instructions

---

## Migration Note

This command consolidates the following individual commands:
- `/ci-setup` → `/setup ci`
- `/dev-tools-setup` → `/setup dev-tools`
- `/lsp-setup` → `/setup lsp`
- `/mcp-setup` → `/setup mcp`
- `/opencode-setup` → `/setup opencode`
- `/webhook-setup` → `/setup webhook`
- `/harness-ui-setup` → `/setup ui`

The individual commands are deprecated but still work for backward compatibility.
