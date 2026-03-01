---
description: Rules for editing hook configuration (hooks.json)
paths: "**/hooks.json"
---

# Hooks Editing Rules

Rules applied when editing `hooks.json` files.

## Important: Dual hooks.json Sync (Required)

**Two hooks.json files exist and must always be in sync:**

```
hooks/hooks.json           ← Source file (for development)
.claude-plugin/hooks.json  ← For plugin distribution (sync required)
```

### Editing Flow

1. Edit `hooks/hooks.json`
2. Apply the same changes to `.claude-plugin/hooks.json`
3. Sync cache with `./scripts/sync-plugin-cache.sh`

```bash
# Always run after changes
./scripts/sync-plugin-cache.sh
```

## Hook Types

3 つのタイプが利用可能です: `command`（汎用）、`prompt`（Stop/SubagentStop 限定）、`http`（v2.1.63+ 外部連携）。

### command Type (General Purpose)

Available for all events:

```json
{
  "type": "command",
  "command": "node \"${CLAUDE_PLUGIN_ROOT}/scripts/run-script.js\" script-name",
  "timeout": 30
}
```

### prompt Type (Stop/SubagentStop Only)

**Official Support**: Only available for Stop and SubagentStop events

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

### http Type (v2.1.63+)

JSON を URL に POST する新しいフック形式。外部サービスとの連携に使用。

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

#### HTTP hook 専用フィールド

| フィールド | 必須 | 説明 |
|-----------|------|------|
| `url` | Yes | POST 先の URL |
| `headers` | No | 追加 HTTP ヘッダー。`$VAR` / `${VAR}` で環境変数展開可 |
| `allowedEnvVars` | No | `headers` で展開を許可する環境変数名リスト。未指定時は展開されない |

#### レスポンス仕様

| レスポンス | 動作 |
|-----------|------|
| `2xx` + 空ボディ | 成功、続行 |
| `2xx` + JSON ボディ | 成功、JSON は command hook と同じスキーマで解析 |
| `非 2xx` / タイムアウト | ノンブロッキングエラー、実行続行 |

#### command hook との主な違い

| 項目 | command hook | http hook |
|------|-------------|-----------|
| 入力 | stdin (JSON) | POST body (JSON) |
| 成功判定 | exit code 0 | 2xx ステータス |
| ブロッキング | exit 2 | 2xx + `permissionDecision: "deny"` の JSON |
| 非同期実行 | `async: true` 対応 | 非対応 |
| `/hooks` メニュー | 追加可能 | 不可（JSON 直接編集のみ） |
| 環境変数 | シェル環境で自動展開 | `allowedEnvVars` に明示リスト必要 |

#### サンプルテンプレート

**Slack 通知**:
```json
{
  "type": "http",
  "url": "https://hooks.slack.com/services/T00/B00/xxx",
  "timeout": 10
}
```

**メトリクス収集**:
```json
{
  "type": "http",
  "url": "http://localhost:9090/metrics/hook",
  "timeout": 5,
  "headers": { "X-Source": "claude-code-harness" }
}
```

**外部ダッシュボード更新**:
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

**Note**: Set timeouts according to processing nature. Don't make them unnecessarily long.

### Recommended Values by Event Type

| Hook Type | Recommended | Reason |
|-----------|-------------|--------|
| SessionStart | 30s | Initialization may take time |
| SubagentStart/Stop | 10s | Tracking only, lightweight processing |
| PreToolUse | 30s | Guard processing, file validation |
| PostToolUse | 5-30s | Depends on processing content |
| Stop | 20s | Ensure completion of termination processing |
| UserPromptSubmit | 10-30s | Policy injection, tracking |

### Special Considerations for Stop Hooks

Stop hooks execute at session termination, so:
- Too short timeouts may interrupt processing
- 20 seconds or more recommended (D14 decision)

## Hook Structure

### Event Types

```json
{
  "hooks": {
    "PreToolUse": [],      // Before tool execution
    "PostToolUse": [],     // After tool execution
    "SessionStart": [],    // At session start
    "Stop": [],            // At session end
    "SubagentStart": [],   // Subagent start
    "SubagentStop": [],    // Subagent end
    "UserPromptSubmit": [],// On user input
    "PermissionRequest": [] // On permission request
  }
}
```

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

- ❌ Editing only one hooks.json
- ❌ Using `type: "prompt"` for events other than Stop/SubagentStop
- ❌ Not instructing `{ok, reason}` schema for prompt type
- ❌ Hooks without timeout
- ❌ Absolute paths other than `${CLAUDE_PLUGIN_ROOT}`
- ❌ Commits without running sync-plugin-cache.sh

## Related Decisions

- **D14**: Hook timeout optimization
- **D15**: Stop hook prompt type official spec compliance (`{ok, reason}` schema)

Details: [.claude/memory/decisions.md](../memory/decisions.md)
