# OpenCode Plugin Policy

Use the **official plugin `tasict/opencode-plugin-cc`** for all OpenCode invocations.

## Core Policy

Direct invocation of raw `opencode run` is prohibited. Use one of the following two methods to invoke OpenCode:

1. **`scripts/opencode-companion.sh`** — Invocation from within Harness skills and agents
2. **`/opencode:*` commands** — Ad-hoc usage in user interactions

## Prohibited

- Direct invocation of `opencode run` (except within `templates/opencode-skills/` if applicable; see exception below)
- Searching for OpenCode MCP via ToolSearch (no MCP server exists for OpenCode)
- Attempting to register an OpenCode MCP server

## No MCP Block

Unlike Codex, OpenCode does **not** use an MCP server. It communicates via an HTTP REST API
(`opencode serve` headless server). No `deny` rules are needed in settings.json for OpenCode.

## Correct Invocation Methods

### Task Delegation (Implementation, Debugging, Investigation)

```bash
# Write-enabled task delegation
bash scripts/opencode-companion.sh task "Fix the bug"

# Via stdin (for large prompts)
cat "$PROMPT_FILE" | bash scripts/opencode-companion.sh task

# Resume previous thread
bash scripts/opencode-companion.sh task --resume-last "Continue where you left off"
```

### Review

```bash
# Review the working tree
bash scripts/opencode-companion.sh review

# Review from a specific base ref
bash scripts/opencode-companion.sh review --base "${TASK_BASE_REF}"

# Adversarial review (challenge design decisions)
bash scripts/opencode-companion.sh adversarial-review
```

### Setup and Job Management

```bash
# Check OpenCode availability
bash scripts/opencode-companion.sh setup --json

# Check running jobs
bash scripts/opencode-companion.sh status

# Retrieve job results
bash scripts/opencode-companion.sh result <job-id>

# Cancel a job
bash scripts/opencode-companion.sh cancel <job-id>
```

### /opencode:* Commands (User Interaction)

```
/opencode:setup              — Check OpenCode CLI setup
/opencode:rescue             — Task delegation (investigation, implementation, debugging)
/opencode:review             — Code review
/opencode:adversarial-review — Adversarial review
/opencode:status             — Check job status
/opencode:result             — Retrieve job results
/opencode:cancel             — Cancel a job
```

## Verdict Mapping (opencode-plugin-cc <-> Harness)

The official plugin's review output uses a different schema from Harness. Conversion rules:

| opencode-plugin-cc | Harness | Notes |
|---|---|---|
| `approve` | `APPROVE` | |
| `needs-attention` | `REQUEST_CHANGES` | |
| `findings[].severity: critical` | `critical_issues[]` | Affects verdict |
| `findings[].severity: high` | `major_issues[]` | Affects verdict |
| `findings[].severity: medium/low` | `recommendations[]` | Does not affect verdict |

## Exception: OpenCode Native Skills

Skills within `templates/opencode-skills/` **run inside the OpenCode CLI**, so
OpenCode native APIs may continue to be used. However, invoking reviews via the companion script is recommended.

## Official Plugin Features

| Feature | Description |
|------|------|
| Job Management | Thread start, resume, cancel, and result retrieval |
| HTTP REST API | High-reliability OpenCode communication via `opencode serve` headless server |
| Structured Output | Structured reviews conforming to `review-output.schema.json` |
| Stop Review Gate | Automatic review gate at session end |
| ACP Server | Agent Client Protocol server for multi-agent coordination |
