---
description: Rules for editing hook configuration (hooks.json)
paths: "**/hooks.json"
---

# Hooks Editing Rules

Rules applied when editing `hooks.json` files.

## SSOT: harness/hooks/hooks.json

**`harness/hooks/hooks.json` is the single source of truth for hook configuration.**

There is only one hooks.json file. The old `.claude-plugin/hooks.json` no longer exists.

### Editing Flow

1. Edit `harness/hooks/hooks.json`
2. Run the full validation suite to verify changes: `make validate`

## Hook Types

4 types are available: `command` (general purpose), `http` (external integration), `prompt` (single LLM judgment), and `agent` (LLM agent judgment). The latter two support all events as of v2.1.63+.

> **CC v2.1.69+**: `InstructionsLoaded` event, `agent_id` / `agent_type` fields, and `{"continue": false, "stopReason": "..."}` response were added.
>
> **CC v2.1.76+**: `Elicitation`, `ElicitationResult`, and `PostCompact` events were added.
> MCP Elicitation cannot interact with UI in background agents, so hooks must handle it automatically.
> PostCompact pairs with PreCompact and is used for post-compaction context re-injection.
>
> **CC v2.1.77+**: Even if a PreToolUse hook returns `"allow"`, settings.json `deny` rules now take precedence.
> If a deny rule exists, the action is denied regardless of the hook's allow response. Keep this priority in mind when designing guardrails.
>
> **CC v2.1.78+**: `StopFailure` event was added. It fires when session stop fails due to API errors (rate limits, auth failures, etc.).
> Use for error logging and recovery processing.
>
> **CC v2.1.89+**: `PermissionDenied` event was added. It fires when the auto mode classifier denies a command.
> Returning `{retry: true}` tells the model the action can be retried. Used for Breezing Worker denial tracking.
>
> **CC v2.1.89+**: `"defer"` was added to PreToolUse `permissionDecision`.
> When a hook returns `"defer"` in headless sessions (`-p` mode), the session pauses.
> The hook is re-evaluated when resumed with `claude -p --resume`. Can be used as a safety valve when a Breezing Worker encounters an operation it cannot judge.
>
> **CC v2.1.89+**: Combining `updatedInput` with `AskUserQuestion` in PreToolUse allows
> headless sessions to collect answers via an external UI and inject them along with `permissionDecision: "allow"`.
>
> **CC v2.1.89+**: When hook output exceeds 50K characters, it is saved to disk and injected into context as a file path + preview.
> Design hooks that return large output with this behavior in mind.
>
> **CC v2.1.90+**: Fixed the blocking behavior when a PreToolUse hook outputs JSON to stdout and exits with code 2.
> Previously, blocking did not work correctly with this pattern. Since Harness's pre-tool.sh uses the exit 2 pattern,
> guardrail deny works more reliably on v2.1.90+.

### command Type (General Purpose)

Available for all events:

```json
{
  "type": "command",
  "command": "node \"${CLAUDE_PLUGIN_ROOT}/scripts/run-script.js\" script-name",
  "timeout": 30
}
```

### prompt Type

**Official Support**: Available for all hook events (v2.1.63+)

```json
{
  "type": "prompt",
  "prompt": "Evaluation instructions...\n\n[IMPORTANT] Always respond in this JSON format:\n{\"ok\": true} or {\"ok\": false, \"reason\": \"reason\"}",
  "timeout": 30
}
```

**Response Schema (Required)**:
```json
{"ok": true}                          // Allow action
{"ok": false, "reason": "explanation"}  // Block action
```

⚠️ **Note**: If you don't explicitly instruct JSON format in the prompt, the LLM may return natural language and cause a `JSON validation failed` error

### agent Type (v2.1.63+)

A new hook format that delegates judgment to an LLM agent. It can analyze code using Read, Grep, and Glob tools to make allow/deny decisions.

```json
{
  "type": "agent",
  "prompt": "Check if the code change introduces security vulnerabilities. $ARGUMENTS",
  "model": "haiku",
  "timeout": 60
}
```

#### agent Hook-Specific Fields

| Field | Required | Description |
|-----------|------|------|
| `prompt` | Yes | Prompt sent to the agent. Use `$ARGUMENTS` to reference the hook input JSON |
| `model` | No | Model to use (default: fast model). `haiku` recommended for cost management |

#### Key Differences from command Hooks

| Aspect | command hook | agent hook |
|------|-------------|-----------|
| Decision method | Rule-based (regex, conditionals) | LLM understands context and decides |
| Tools | Shell commands | Read, Grep, Glob (no side effects) |
| Cost | Low (process startup only) | High (LLM inference token consumption) |
| Use case | Deterministic rules | Context-dependent quality judgments |
| Async | `async: true` supported | Not supported |

#### Cost Management Guidelines

- Use matcher to minimize scope (e.g., `Write|Edit` only)
- Use `model: "haiku"` to keep costs low
- Recommended token limit per invocation: 2,000
- Consider rolling back to command type if monthly costs exceed budget

### http Type (v2.1.63+)

A new hook format that POSTs JSON to a URL. Used for integration with external services.

```json
{
  "type": "http",
  "url": "http://localhost:8080/hooks/pre-tool-use",
  "timeout": 30,
  "headers": {
    "Authorization": "Bearer $MY_TOKEN"
  },
  "allowedEnvVars": ["MY_TOKEN"]
}
```

#### HTTP Hook-Specific Fields

| Field | Required | Description |
|-----------|------|------|
| `url` | Yes | URL to POST to |
| `headers` | No | Additional HTTP headers. `$VAR` / `${VAR}` for environment variable expansion |
| `allowedEnvVars` | No | List of environment variable names allowed for expansion in `headers`. Not expanded if unspecified |

#### Response Specification

| Response | Behavior |
|-----------|------|
| `2xx` + empty body | Success, continue |
| `2xx` + JSON body | Success, JSON parsed with same schema as command hook |
| Non-`2xx` / timeout | Non-blocking error, execution continues |

#### Key Differences from command Hooks

| Aspect | command hook | http hook |
|------|-------------|-----------|
| Input | stdin (JSON) | POST body (JSON) |
| Success criteria | exit code 0 | 2xx status |
| Blocking | exit 2 | 2xx + `permissionDecision: "deny"` JSON |
| Async execution | `async: true` supported | Not supported |
| `/hooks` menu | Can be added | Not available (JSON direct edit only) |
| Environment variables | Auto-expanded in shell environment | Requires explicit `allowedEnvVars` list |

#### Sample Templates

**Slack Notification**:
```json
{
  "type": "http",
  "url": "https://hooks.slack.com/services/T00/B00/xxx",
  "timeout": 10
}
```

**Metrics Collection**:
```json
{
  "type": "http",
  "url": "http://localhost:9090/metrics/hook",
  "timeout": 5,
  "headers": { "X-Source": "claude-code-harness" }
}
```

**External Dashboard Update**:
```json
{
  "type": "http",
  "url": "https://dashboard.example.com/api/events",
  "timeout": 15,
  "headers": { "Authorization": "Bearer $DASHBOARD_TOKEN" },
  "allowedEnvVars": ["DASHBOARD_TOKEN"]
}
```

### Recommended Pattern

Execute command type via `run-script.js`:

```json
{
  "type": "command",
  "command": "node \"${CLAUDE_PLUGIN_ROOT}/scripts/run-script.js\" {script-name}",
  "timeout": 30
}
```

## Timeout Setting Guidelines

> **Claude Code v2.1.3+**: Maximum timeout for tool hooks extended from 60 seconds → 10 minutes

### Guidelines by Processing Nature

| Hook Type | Recommended Timeout | Notes |
|-----------|-------------------|-------|
| Lightweight check (guard) | 5-10s | File existence checks, etc. |
| Normal processing (cleanup) | 30-60s | File operations, git operations |
| Heavy processing (test) | 60-120s | Test execution, builds |
| External API integration | 60-180s | Codex reviews, etc. |
| agent hook (LLM judgment) | 30-60s | Depends on model and prompt size. 30s for haiku, 60s for sonnet |
| http hook (external integration) | 5-15s | 5s for local server, 15s for external services. Non-blocking on timeout |

**Note**: Set timeouts according to processing nature. Don't make them unnecessarily long.

#### agent Hook Measured Guidelines (haiku model)

| Prompt Size | Expected Latency | Recommended timeout |
|------------|-------------|------------|
| ~500 tokens | 3-8s | 15s |
| ~1,000 tokens | 5-15s | 30s |
| ~2,000 tokens | 10-25s | 45s |
| Over 2,000 tokens | Not recommended | — |

Cost estimate (haiku): ~$0.01-0.05/day for 100 sessions/day. Under $1-2/month is the normal range.

### Recommended Values by Event Type

| Hook Type | Recommended | Reason |
|-----------|-------------|--------|
| InstructionsLoaded | 5-10s | Lightweight initial context verification only |
| SessionStart | 30s | Initialization may take time |
| SubagentStart/Stop | 10s | Tracking only, lightweight processing |
| TeammateIdle / TaskCompleted | 10-20s | Team progress and stop decision (use `continue:false` if needed) |
| PreToolUse | 30s | Guard processing, file validation |
| PostToolUse | 5-30s | Depends on processing content |
| Stop | 20s | Ensure completion of termination processing |
| SessionEnd | 30s | Session end processing. Controllable via `CLAUDE_CODE_SESSIONEND_HOOKS_TIMEOUT_MS` |
| UserPromptSubmit | 10-30s | Policy injection, tracking |
| Elicitation | 10s | MCP elicitation interception. Auto-skipped in Breezing |
| ElicitationResult | 5s | Result logging only, lightweight processing |
| PostCompact | 15s | Context re-injection. Includes WIP task state restoration |
| PermissionDenied | 10s | Auto mode denial recording/notification. Lightweight processing (v2.1.89+) |
| StopFailure | 10s | API error log recording only. No recovery processing needed (v2.1.78+) |
| ConfigChange | 10s | Configuration change audit recording |

### Special Considerations for Stop Hooks

Stop hooks execute at session termination, so:
- Too short timeouts may interrupt processing
- 20 seconds or more recommended (D14 decision)

### Special Considerations for SessionEnd Hooks

**CC v2.1.74+**: SessionEnd hooks timeout is now controllable via the `CLAUDE_CODE_SESSIONEND_HOOKS_TIMEOUT_MS` environment variable.
Previously, hooks were killed at a fixed 1.5 seconds regardless of the `hook.timeout` setting.

```bash
# Harness recommendation: Set 45 seconds for session-cleanup (timeout: 30s)
export CLAUDE_CODE_SESSIONEND_HOOKS_TIMEOUT_MS=45000
```

- To ensure the Harness `session-cleanup` hook (configured with timeout: 30s in hooks.json) completes reliably, 45 seconds or more is recommended
- If the environment variable is not set, CC's default value applies (v2.1.74+ respects the hook.timeout setting)

## Hook Structure

### Event Types

```json
{
  "hooks": {
    "PreToolUse": [],      // Before tool execution
    "PostToolUse": [],     // After tool execution
    "InstructionsLoaded": [], // Instruction load completed (v2.1.69+)
    "SessionStart": [],    // At session start
    "Stop": [],            // At session end
    "SubagentStart": [],   // Subagent start
    "SubagentStop": [],    // Subagent end
    "TeammateIdle": [],    // Teammate idle event (team mode)
    "TaskCompleted": [],   // Teammate task completion event (team mode)
    "WorktreeCreate": [],  // Worktree lifecycle start
    "WorktreeRemove": [],  // Worktree lifecycle end
    "UserPromptSubmit": [],// On user input
    "PermissionRequest": [], // On permission request
    "PreCompact": [],      // Before context compaction
    "PostCompact": [],     // After context compaction (v2.1.76+)
    "Elicitation": [],     // MCP elicitation request (v2.1.76+)
    "ElicitationResult": [], // MCP elicitation result (v2.1.76+)
    "Notification": [],    // On notification dispatch
    "PermissionDenied": [], // Auto mode permission denial (v2.1.89+)
    "StopFailure": [],     // API error during session stop (v2.1.78+)
    "ConfigChange": []     // Settings change event
  }
}
```

### Teammate Event Fields (v2.1.69+)

For `TeammateIdle` / `TaskCompleted` and related events, prioritize the following fields:

- `agent_id` (recommended key)
- `agent_type` (worker/reviewer, etc.)
- `session_id` (backward-compatible key)

Do not rely solely on `session_id`; reference `agent_id` first with fallback to `session_id`.

### Stop Response Pattern (v2.1.69+)

To stop processing in team events, return the following format:

```json
{"continue": false, "stopReason": "all_tasks_completed"}
```

To continue as before, return `{"decision":"approve"}`.

### matcher Patterns

```json
// Match specific tool
{ "matcher": "Write|Edit|Bash" }

// Match all
{ "matcher": "*" }

// Multiple tools
{ "matcher": "Skill|Task|SlashCommand" }
```

### once Option

Execute only once per session:

```json
{
  "type": "command",
  "command": "...",
  "timeout": 30,
  "once": true  // Recommended for SessionStart
}
```

## Prohibited

- Not instructing `{ok, reason}` schema for prompt type
- Hooks without timeout
- Absolute paths other than `${CLAUDE_PLUGIN_ROOT}`

## Related Decisions

- **D14**: Hook timeout optimization
- **D15**: Stop hook prompt type official spec compliance (`{ok, reason}` schema)

Details: [.claude/memory/decisions.md](../memory/decisions.md)
