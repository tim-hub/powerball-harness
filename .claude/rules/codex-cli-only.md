# Codex Plugin Policy

Use the **official plugin `openai/codex-plugin-cc`** for all Codex invocations.

## Core Policy

Direct invocation of raw `codex exec` is prohibited. Use one of the following two methods to invoke Codex:

1. **`scripts/codex-companion.sh`** — Invocation from within Harness skills and agents
2. **`/codex:*` commands** — Ad-hoc usage in user interactions

## Prohibited

- Direct invocation of `codex exec` (except within `skills-codex/`; see exception below)
- Use of `mcp__codex__codex` (MCP server has been deprecated)
- Searching for Codex MCP via ToolSearch
- Re-registering the MCP server via `claude mcp add codex`

## MCP Block (v2.1.78+)

Legacy MCP tools are blocked via `deny` rules in settings.json (already configured):

```json
{
  "permissions": {
    "deny": ["mcp__codex__*"]
  }
}
```

## Correct Invocation Methods

### Task Delegation (Implementation, Debugging, Investigation)

```bash
# Write-enabled task delegation
bash scripts/codex-companion.sh task --write "Fix the bug"

# Via stdin (for large prompts)
cat "$PROMPT_FILE" | bash scripts/codex-companion.sh task --write

# Resume previous thread
bash scripts/codex-companion.sh task --resume-last --write "Continue where you left off"
```

### Review

```bash
# Review the working tree
bash scripts/codex-companion.sh review

# Review from a specific base ref
bash scripts/codex-companion.sh review --base "${TASK_BASE_REF}"

# Adversarial review (challenge design decisions)
bash scripts/codex-companion.sh adversarial-review
```

### Setup and Job Management

```bash
# Check Codex availability
bash scripts/codex-companion.sh setup --json

# Check running jobs
bash scripts/codex-companion.sh status

# Retrieve job results
bash scripts/codex-companion.sh result <job-id>

# Cancel a job
bash scripts/codex-companion.sh cancel <job-id>
```

### /codex:* Commands (User Interaction)

```
/codex:setup              — Check Codex CLI setup
/codex:rescue             — Task delegation (investigation, implementation, debugging)
/codex:review             — Code review
/codex:adversarial-review — Adversarial review
/codex:status             — Check job status
/codex:result             — Retrieve job results
/codex:cancel             — Cancel a job
```

## Verdict Mapping (Official Plugin <-> Harness)

The official plugin's review output uses a different schema from Harness. Conversion rules:

| Official Plugin | Harness | Notes |
|---|---|---|
| `approve` | `APPROVE` | |
| `needs-attention` | `REQUEST_CHANGES` | |
| `findings[].severity: critical` | `critical_issues[]` | Affects verdict |
| `findings[].severity: high` | `major_issues[]` | Affects verdict |
| `findings[].severity: medium/low` | `recommendations[]` | Does not affect verdict |

## Exception: Codex Native Skills

Skills within `skills-codex/` **run inside the Codex CLI**, so
Codex native APIs such as `spawn_agent` / `wait_agent` / `send_input` / `close_agent`
may continue to be used. However, invoking reviews via the companion script is recommended.

## Official Plugin Features

| Feature | Description |
|------|------|
| Job Management | Thread start, resume, cancel, and result retrieval |
| App Server Protocol | High-reliability Codex communication via JSON-RPC over TCP |
| Structured Output | Structured reviews conforming to `review-output.schema.json` |
| Stop Review Gate | Automatic review gate at session end |
| GPT-5.4 Prompting | Optimized prompt guidance for Codex |
