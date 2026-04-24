# Harness for OpenCode

This directory contains the opencode.ai-compatible distribution of Claude Code
Harness.

## Language Policy

English is the default for distributed docs and setup output. Japanese remains
available as an explicit opt-in through `i18n.language: ja`,
`CLAUDE_CODE_HARNESS_LANG=ja`, and the Japanese setup templates under
`templates/locales/ja/`.

OpenCode-specific docs should stay in English by default. Do not mix Japanese
instructions into this file unless the section is explicitly about the Japanese
opt-in path.

## Setup

### Option 1: One-Command Setup (Recommended)

You can set up OpenCode support even if Claude Code is not installed:

```bash
cd your-project
curl -fsSL https://raw.githubusercontent.com/Chachamaru127/claude-code-harness/main/scripts/setup-opencode.sh | bash
```

To set up Unified Memory as well:

```bash
cd your-project
/path/to/claude-code-harness/scripts/harness-mem setup --platform opencode
```

### Option 2: Setup From Claude Code

If you already use Claude Code Harness, run:

```bash
# Run inside Claude Code
/opencode-setup
```

### Option 3: Manual Setup

```bash
# Clone Harness
git clone https://github.com/Chachamaru127/claude-code-harness.git

# Copy OpenCode commands
cp -r claude-code-harness/opencode/commands/ your-project/.opencode/commands/
cp claude-code-harness/opencode/AGENTS.md your-project/AGENTS.md
```

## MCP Server Setup (Optional)

MCP servers let OpenCode call Harness workflow tools directly.

```bash
# Build the MCP server
cd claude-code-harness/mcp-server
npm install
npm run build

# Copy opencode.json into your project and adjust paths
cp claude-code-harness/opencode/opencode.json your-project/
```

If you also use the unified memory daemon:

```bash
# Start the memory daemon
./scripts/harness-memd start

# Check health
./scripts/harness-mem-client.sh health
```

You can also run diagnostics through `harness-mem`:

```bash
/path/to/claude-code-harness/scripts/harness-mem doctor --platform opencode --fix
```

## Available Commands

| Command | Description |
|---------|-------------|
| `/harness-init` | Project setup |
| `/plan-with-agent` | Create a development plan |
| `/work` | Execute tasks |
| `/harness-review` | Review code |
| `/sync-status` | Check progress |
| `/handoff-to-opencode` | Generate a completion report for the OpenCode PM |

## PM Mode

When OpenCode acts as the project manager:

| Command | Description |
|---------|-------------|
| `/start-session` | Start a session and inspect context |
| `/plan-with-cc` | Create a plan, including evals when needed |
| `/project-overview` | Understand the project structure |
| `/handoff-to-claude` | Generate a request for Claude Code |
| `/review-cc-work` | Review and approve Claude Code work |

### Workflow

```
OpenCode (PM)                    Claude Code (Impl)
    |                                   |
    | /start-session                    |
    | /plan-with-cc                     |
    | /handoff-to-claude -------------> |
    |                                   | /work
    |                                   | /handoff-to-opencode
    | <-------------------------------- |
    | /review-cc-work                   |
    |    |-- approve -> next task ----> |
    |    `-- request_changes --------> |
```

## MCP Tools

The MCP server exposes these tools:

| Tool | Description |
|------|-------------|
| `harness_workflow_plan` | Create a plan |
| `harness_workflow_work` | Execute tasks |
| `harness_workflow_review` | Review code |
| `harness_session_broadcast` | Send cross-session notifications |
| `harness_status` | Check status |
| `harness_mem_resume_pack` | Fetch resume context |
| `harness_mem_search` | Search shared memory |
| `harness_mem_record_checkpoint` | Record a checkpoint |
| `harness_mem_finalize_session` | Finalize a session |

## Usage

```bash
# Start OpenCode
cd your-project
opencode

# Run commands
/plan-with-agent  # Create a plan
/work             # Execute tasks
/harness-review   # Review code
```

## Limitations

- OpenCode does not use the Claude Code plugin system under `.claude-plugin/`.
- Memory hooks live in `opencode/plugins/harness-memory/index.ts`
  (`chat.message`, `session.idle`, `session.compacted`).
- Skill frontmatter keeps both `description-en` and `description-ja` so OpenCode
  follows the same language contract as Claude Code and Codex.

## Links

- [Claude Code Harness](https://github.com/Chachamaru127/claude-code-harness)
- [OpenCode Documentation](https://opencode.ai/docs/)
- [OpenCode Commands](https://opencode.ai/docs/commands/)
