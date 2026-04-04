# Harness Go Rewrite — Architecture Design

> Zero-base rewrite. No backward compatibility. All accumulated knowledge applied.

## Why Go

| 要件 | Go | Rust | 現行 (bash+TS) |
|------|-----|------|---------------|
| Cold start | 1-2ms | 0.5-1ms | 40-60ms |
| Cross-compile | `GOOS=x go build` | toolchain 管理 | N/A (interpreter) |
| JSON handling | stdlib `encoding/json` | serde (verbose) | jq + node |
| HTTP client | stdlib `net/http` | reqwest (dep) | curl spawn |
| Development speed | seconds | minutes | immediate but fragile |
| Binary size | 5-10MB | 2-5MB | ~200KB scripts |
| Dependencies | stdlib + 1-2 (uuid) | many crates | node + jq + bash |
| Self-referential dev | compile in 2s | compile in 30s+ | edit & run |

**結論**: Harness は「毎秒数回起動され、JSON を読んで判断を返す」ツール。Go の stdlib 中心 + 高速コンパイルが最適。

### 許容する外部依存

- `github.com/google/uuid` — OTel trace ID 生成。stdlib に UUID v4 がない
- `github.com/mattn/go-isatty` — TTY 検出（省略可: `os.Stdin.Stat()` で代替）
- **それ以外は不可。** MCP プロトコルも OTel フォーマットも stdlib の `encoding/json` + `net/http` で自前実装

## Directory Structure

```
harness-go/
├── cmd/
│   └── harness/
│       └── main.go              # Single entry: stdin → route → stdout
│
├── internal/
│   ├── guardrail/               # Guardrail engine (hot path, PreToolUse/PostToolUse/Permission)
│   │   ├── engine.go            # Rule evaluation loop
│   │   ├── rules.go             # Declarative rule table (R01-R13+)
│   │   ├── pretool.go           # PreToolUse: deny/allow/defer
│   │   ├── posttool.go          # PostToolUse: advisory checks
│   │   ├── permission.go        # PermissionRequest: conditional eval
│   │   └── tampering.go         # Test/config tampering detection
│   │
│   ├── session/                 # Session lifecycle + agent tracking
│   │   ├── start.go             # SessionStart: env + memory bridge + init (parallel)
│   │   ├── stop.go              # Stop: summary, memory save, WIP check
│   │   ├── compact.go           # PreCompact save + PostCompact WIP restore
│   │   ├── agent.go             # SubagentStart/Stop tracking + trace
│   │   └── state.go             # State file + JSONL append with rotation
│   │
│   ├── event/                   # Remaining hook events + dispatcher
│   │   ├── dispatcher.go        # stdin → parse → route → execute → stdout
│   │   ├── prompt.go            # UserPromptSubmit: policy injection, tracking
│   │   ├── task.go              # TaskCreated/Completed + webhook trigger
│   │   ├── denied.go            # PermissionDenied: telemetry + retry
│   │   └── misc.go              # Notification, ConfigChange, Elicitation, StopFailure
│   │
│   ├── plans/                   # Plans.md operations + effort scoring
│   │   ├── parser.go            # Parse Plans.md tables (5-column format)
│   │   ├── marker.go            # Status marker read/update
│   │   └── effort.go            # Task complexity scoring
│   │
│   ├── notify/                  # External notifications (webhook + OTel + broadcast)
│   │   ├── webhook.go           # HARNESS_WEBHOOK_URL → HTTP POST (sync with timeout, secret mask)
│   │   ├── otel.go              # OTLP HTTP span export (sync with timeout, fallback JSONL)
│   │   └── broadcast.go         # Inter-session file-based broadcast
│   │
│   └── config/                  # Configuration + security utilities
│       ├── settings.go          # Plugin settings.json
│       ├── env.go               # Environment detection (CLAUDE_*, PROJECT_ROOT)
│       ├── paths.go             # State dir, log file path resolution
│       ├── safefile.go          # Symlink-safe file operations (append, read, mkdir)
│       └── mask.go              # Secret masking for logs/webhook URLs
│
├── pkg/
│   └── hookproto/               # Hook protocol types (public API)
│       ├── input.go             # HookInput struct (all fields)
│       └── output.go            # HookOutput, SystemMessage, Decision, etc.
│
├── skills-v3/                   # Skills (Markdown, unchanged)
├── agents-v3/                   # Agents (Markdown, unchanged)
├── .claude-plugin/
│   ├── plugin.json
│   ├── hooks.json               # Simplified: all → bin/harness hook <event>
│   └── settings.json
│
├── bin/                         # Build output (gitignored)
│   ├── harness-darwin-arm64
│   ├── harness-darwin-amd64
│   ├── harness-linux-amd64
│   └── harness-windows-amd64.exe
│
├── Makefile
├── go.mod
└── go.sum
```

**パッケージ数: 11 → 6** (guardrail, session, event, plans, notify, config)。`pkg/` は hookproto のみ。
JSONL 操作は `config/safefile.go` に統合。agent tracking は `session/agent.go` に統合。
review (security/dual) はスキル側のプロンプト指示で完結するため Go binary から除外。

## Core Design: Single Binary, Subcommand Routing

```
# === Hook events (全 20 event を網羅。現行 hooks.json の全 command hook に対応) ===
bin/harness hook pretool            # PreToolUse
bin/harness hook pretool --browser  # PreToolUse (browser MCP tools)
bin/harness hook posttool           # PostToolUse
bin/harness hook permission         # PermissionRequest
bin/harness hook session-start      # SessionStart (startup + resume)
bin/harness hook session-end        # SessionEnd
bin/harness hook stop               # Stop
bin/harness hook pre-compact        # PreCompact
bin/harness hook post-compact       # PostCompact
bin/harness hook task-completed     # TaskCompleted
bin/harness hook task-created       # TaskCreated (runtime-reactive)
bin/harness hook permission-denied  # PermissionDenied
bin/harness hook teammate-idle      # TeammateIdle
bin/harness hook notification       # Notification
bin/harness hook config-change      # ConfigChange
bin/harness hook elicitation        # Elicitation
bin/harness hook elicitation-result # ElicitationResult
bin/harness hook stop-failure       # StopFailure
bin/harness hook user-prompt        # UserPromptSubmit
bin/harness hook subagent-start     # SubagentStart
bin/harness hook subagent-stop      # SubagentStop
bin/harness hook setup              # Setup (init / init-only / maintenance)
bin/harness hook instructions-loaded # InstructionsLoaded
bin/harness hook worktree-create    # WorktreeCreate
bin/harness hook worktree-remove    # WorktreeRemove
bin/harness hook cwd-changed        # CwdChanged (runtime-reactive)
bin/harness hook file-changed       # FileChanged (runtime-reactive)
bin/harness hook post-tool-failure  # PostToolUseFailure

# === Utilities ===
bin/harness effort <task-desc>      # Effort scoring
bin/harness plans sync              # Plans.md sync
bin/harness plans update <id> <status>  # Marker update
bin/harness version                 # Version info
```

**40+ shell scripts → 1 binary, ~28 subcommands。現行 hooks.json の全 command hook を漏れなくカバー。**
**MCP subcommand は削除（D3: 分離プロセス維持）。**

## hooks.json (Simplified)

```json
{
  "hooks": {
    "PreToolUse": [{
      "matcher": "Write|Edit|MultiEdit|Bash|Read",
      "hooks": [{
        "type": "command",
        "command": "${CLAUDE_PLUGIN_ROOT}/bin/harness hook pretool",
        "timeout": 10
      }]
    }, {
      "matcher": "mcp__chrome-devtools__.*|mcp__playwright__.*",
      "hooks": [{
        "type": "command",
        "command": "${CLAUDE_PLUGIN_ROOT}/bin/harness hook pretool --browser",
        "timeout": 5
      }]
    }],
    "PermissionRequest": [{
      "matcher": "Edit|Write|MultiEdit",
      "hooks": [{
        "type": "command",
        "command": "${CLAUDE_PLUGIN_ROOT}/bin/harness hook permission",
        "timeout": 10
      }]
    }, {
      "matcher": "Bash",
      "if": "Bash(git status*)|Bash(git diff*)|Bash(git log*)|Bash(git branch*)|Bash(git rev-parse*)|Bash(git show*)|Bash(git ls-files*)|Bash(npm test*)|Bash(npm run test*)|Bash(npm run lint*)|Bash(npm run typecheck*)|Bash(npm run build*)|Bash(npm run validate*)|Bash(npm lint*)|Bash(npm typecheck*)|Bash(npm build*)|Bash(pnpm test*)|Bash(pnpm run test*)|Bash(pnpm run lint*)|Bash(pnpm run typecheck*)|Bash(pnpm run build*)|Bash(pnpm run validate*)|Bash(pnpm lint*)|Bash(pnpm typecheck*)|Bash(pnpm build*)|Bash(yarn test*)|Bash(yarn run test*)|Bash(yarn run lint*)|Bash(yarn run typecheck*)|Bash(yarn run build*)|Bash(yarn run validate*)|Bash(yarn lint*)|Bash(yarn typecheck*)|Bash(yarn build*)|Bash(pytest*)|Bash(python -m pytest*)|Bash(go test*)|Bash(cargo test*)",
      "hooks": [{
        "type": "command",
        "command": "${CLAUDE_PLUGIN_ROOT}/bin/harness hook permission",
        "timeout": 10
      }]
    }],
    "PostToolUse": [{
      "matcher": "Write|Edit|MultiEdit|Bash",
      "hooks": [{
        "type": "command",
        "command": "${CLAUDE_PLUGIN_ROOT}/bin/harness hook posttool",
        "timeout": 10
      }]
    }, {
      "matcher": "Write|Edit|Task",
      "hooks": [{
        "type": "command",
        "command": "${CLAUDE_PLUGIN_ROOT}/bin/harness worker auto-test",
        "timeout": 120,
        "async": true
      }]
    }, {
      "matcher": "Bash",
      "hooks": [{
        "type": "command",
        "command": "${CLAUDE_PLUGIN_ROOT}/bin/harness worker ci-check",
        "timeout": 30,
        "async": true
      }]
    }],
    "SessionStart": [{
      "matcher": "startup|resume",
      "hooks": [{
        "type": "command",
        "command": "${CLAUDE_PLUGIN_ROOT}/bin/harness hook session-start",
        "timeout": 15,
        "once": true
      }]
    }],
    "Stop": [{
      "hooks": [{
        "type": "command",
        "command": "${CLAUDE_PLUGIN_ROOT}/bin/harness hook stop",
        "timeout": 20
      }, {
        "type": "agent",
        "prompt": "Check if there are incomplete tasks before allowing session to stop. Read the Plans.md file and look for tasks with status 'cc:WIP'. If any WIP tasks exist, return JSON: {\"decision\": \"block\", \"reason\": \"WIP tasks remain: [list task numbers]. Consider completing them or marking as blocked before stopping.\"}. If no WIP tasks, return nothing (allow stop). Input: $ARGUMENTS",
        "model": "haiku",
        "timeout": 30
      }]
    }],
    "PreCompact": [{
      "hooks": [{
        "type": "command",
        "command": "${CLAUDE_PLUGIN_ROOT}/bin/harness hook pre-compact",
        "timeout": 15
      }, {
        "type": "agent",
        "prompt": "Check Plans.md for tasks with status 'cc:WIP' before the context window is compacted. If any WIP tasks exist, include a warning in systemMessage: 'Warning: Compacting context with WIP tasks in progress: [list task IDs and titles]. Key context about these tasks may be lost after compaction. Consider completing or checkpointing them first.' If no WIP tasks, return nothing. Input: $ARGUMENTS",
        "model": "haiku",
        "timeout": 30
      }]
    }],
    "PostCompact": [{
      "matcher": "*",
      "hooks": [{
        "type": "command",
        "command": "${CLAUDE_PLUGIN_ROOT}/bin/harness hook post-compact",
        "timeout": 10
      }]
    }],
    "TaskCompleted": [{
      "hooks": [{
        "type": "command",
        "command": "${CLAUDE_PLUGIN_ROOT}/bin/harness hook task-completed",
        "timeout": 10
      }]
    }],
    "TaskCreated": [{
      "hooks": [{
        "type": "command",
        "command": "${CLAUDE_PLUGIN_ROOT}/bin/harness hook task-created",
        "timeout": 5
      }]
    }],
    "PermissionDenied": [{
      "hooks": [{
        "type": "command",
        "command": "${CLAUDE_PLUGIN_ROOT}/bin/harness hook permission-denied",
        "timeout": 5
      }]
    }],
    "SubagentStart": [{
      "matcher": "worker|reviewer|scaffolder|video-scene-generator",
      "hooks": [{
        "type": "command",
        "command": "${CLAUDE_PLUGIN_ROOT}/bin/harness hook subagent-start",
        "timeout": 5
      }]
    }],
    "SubagentStop": [{
      "matcher": "worker|reviewer|scaffolder|video-scene-generator",
      "hooks": [{
        "type": "command",
        "command": "${CLAUDE_PLUGIN_ROOT}/bin/harness hook subagent-stop",
        "timeout": 5
      }]
    }],
    "TeammateIdle": [{
      "hooks": [{
        "type": "command",
        "command": "${CLAUDE_PLUGIN_ROOT}/bin/harness hook teammate-idle",
        "timeout": 10
      }]
    }],
    "ConfigChange": [{
      "hooks": [{
        "type": "command",
        "command": "${CLAUDE_PLUGIN_ROOT}/bin/harness hook config-change",
        "timeout": 10
      }]
    }],
    "UserPromptSubmit": [{
      "matcher": "*",
      "hooks": [{
        "type": "command",
        "command": "${CLAUDE_PLUGIN_ROOT}/bin/harness hook user-prompt",
        "timeout": 10
      }]
    }],
    "Notification": [{
      "hooks": [{
        "type": "command",
        "command": "${CLAUDE_PLUGIN_ROOT}/bin/harness hook notification",
        "timeout": 5
      }]
    }],
    "StopFailure": [{
      "hooks": [{
        "type": "command",
        "command": "${CLAUDE_PLUGIN_ROOT}/bin/harness hook stop-failure",
        "timeout": 5
      }]
    }],
    "SessionEnd": [{
      "hooks": [{
        "type": "command",
        "command": "${CLAUDE_PLUGIN_ROOT}/bin/harness hook session-end",
        "timeout": 15
      }]
    }],
    "Setup": [{
      "matcher": "init|init-only",
      "hooks": [{
        "type": "command",
        "command": "${CLAUDE_PLUGIN_ROOT}/bin/harness hook setup --mode init",
        "timeout": 60,
        "once": true
      }]
    }, {
      "matcher": "maintenance",
      "hooks": [{
        "type": "command",
        "command": "${CLAUDE_PLUGIN_ROOT}/bin/harness hook setup --mode maintenance",
        "timeout": 60,
        "once": true
      }]
    }],
    "InstructionsLoaded": [{
      "hooks": [{
        "type": "command",
        "command": "${CLAUDE_PLUGIN_ROOT}/bin/harness hook instructions-loaded",
        "timeout": 10
      }]
    }],
    "WorktreeCreate": [{
      "hooks": [{
        "type": "command",
        "command": "${CLAUDE_PLUGIN_ROOT}/bin/harness hook worktree-create",
        "timeout": 10
      }]
    }],
    "WorktreeRemove": [{
      "hooks": [{
        "type": "command",
        "command": "${CLAUDE_PLUGIN_ROOT}/bin/harness hook worktree-remove",
        "timeout": 10
      }]
    }],
    "CwdChanged": [{
      "hooks": [{
        "type": "command",
        "command": "${CLAUDE_PLUGIN_ROOT}/bin/harness hook cwd-changed",
        "timeout": 10
      }]
    }],
    "FileChanged": [{
      "hooks": [{
        "type": "command",
        "command": "${CLAUDE_PLUGIN_ROOT}/bin/harness hook file-changed",
        "timeout": 10
      }]
    }],
    "Elicitation": [{
      "matcher": "*",
      "hooks": [{
        "type": "command",
        "command": "${CLAUDE_PLUGIN_ROOT}/bin/harness hook elicitation",
        "timeout": 10
      }]
    }],
    "ElicitationResult": [{
      "matcher": "*",
      "hooks": [{
        "type": "command",
        "command": "${CLAUDE_PLUGIN_ROOT}/bin/harness hook elicitation-result",
        "timeout": 5
      }]
    }],
    "PostToolUseFailure": [{
      "hooks": [{
        "type": "command",
        "command": "${CLAUDE_PLUGIN_ROOT}/bin/harness hook post-tool-failure",
        "timeout": 10
      }]
    }]
  }
}
```

**現行 hooks.json の全 command hook event を網羅。** agent hooks (type: "agent") は PreToolUse/PostToolUse 内にそのまま残る。

## Guardrail Engine Design

```go
// internal/guardrail/rules.go

type Rule struct {
    ID          string                         // R01, R02, ...
    Name        string                         // Human-readable name
    Events      []string                       // Which hook events this applies to
    Match       func(input *hookproto.Input) bool
    Evaluate    func(input *hookproto.Input) *Decision
    Severity    Severity                       // block, warn, info
}

type Decision struct {
    Action  Action  // allow, deny, defer, warn
    Reason  string
    Details map[string]any
}

var Rules = []Rule{
    {
        ID:       "R01",
        Name:     "no-hardcoded-secrets",
        Events:   []string{"pretool"},
        Severity: Block,
        Match: func(in *hookproto.Input) bool {
            return in.Tool == "Write" || in.Tool == "Edit"
        },
        Evaluate: func(in *hookproto.Input) *Decision {
            if containsSecret(in.Content) {
                return &Decision{Action: Deny, Reason: "Hardcoded secret detected"}
            }
            return nil // pass
        },
    },
    {
        ID:       "R02",
        Name:     "no-test-tampering",
        Events:   []string{"pretool"},
        Severity: Block,
        Match: func(in *hookproto.Input) bool {
            return (in.Tool == "Write" || in.Tool == "Edit") && isTestFile(in.FilePath)
        },
        Evaluate: func(in *hookproto.Input) *Decision {
            if detectsTampering(in.Content) {
                return &Decision{Action: Deny, Reason: "Test tampering detected: skip/only/assertion removal"}
            }
            return nil
        },
    },
    // ... R03-R13+
}
```

**宣言的ルールテーブル。** ルール追加は struct を 1 つ足すだけ。shell の if/else 連鎖と比較して可読性が桁違い。

## State Management

```go
// internal/config/paths.go

func StateDir() string {
    if d := os.Getenv("CLAUDE_PLUGIN_DATA"); d != "" {
        hash := projectHash(ProjectRoot())
        return filepath.Join(d, "projects", hash)
    }
    return filepath.Join(ProjectRoot(), ".claude", "state")
}

// Symlink safety is built into every file operation
func SafeAppend(path string, data []byte) error {
    if isSymlink(path) || isSymlink(filepath.Dir(path)) {
        return ErrSymlinkRefused
    }
    // ...
}
```

**セキュリティチェック（symlink 拒否、ディレクトリ検証）がライブラリレベルで強制。** 個別スクリプトで忘れる問題が消える。

## Security Design

| 脅威 | 対策 | 実装場所 |
|------|------|---------|
| **Symlink traversal** | `safefile.go` で全 I/O 前に `os.Lstat` チェック。symlink は即拒否 | config/safefile.go |
| **Path traversal (../)** | `filepath.Clean` + `filepath.Rel` で state dir 外への書込を拒否 | config/paths.go |
| **Secret leak (logs)** | `mask.go` で URL、トークン、API キーを `***` にマスク。webhook URL も対象 | config/mask.go |
| **Command injection** | guardrail ホットパス（pretool/posttool/permission）は shell・exec.Command を使わず全て内部処理。worker サブコマンド（auto-test, ci-check）は exec.Command でプロジェクトの test/CI コマンドを実行する（現行 shell 版と同等） | guardrail/*, session/* は内部処理。event/worker.go は exec.Command 許可 |
| **TOCTOU** | ファイル存在チェックせず直接操作 → エラーハンドリング | config/safefile.go |
| **Unbounded growth** | JSONL rotation (500行超 → 400行に切詰) を safefile.Append に内蔵 | config/safefile.go |
| **Secret in hook output** | PreToolUse deny 理由にユーザー入力を含める際はサニタイズ | guardrail/pretool.go |

## Delivery Model: Short-Lived Process での通知保証

Go binary は都度起動→即終了の短命プロセス。async goroutine はプロセス終了で消える。

**設計方針**: webhook と OTel は **sync with timeout** で送信。

```
bin/harness hook task-completed
  → guardrail check (1ms)
  → JSONL append (0.5ms)
  → webhook POST (timeout 3s, sync)  ← プロセス内で完了を待つ
  → OTel POST (timeout 2s, sync)     ← 同上
  → stdout JSON response
  → exit
```

- 送信成功/タイムアウト/接続拒否のいずれかが確定してから exit
- 送信失敗は stderr ログのみ。リトライしない（次回の hook 起動時に新イベントを送る）
- webhook + OTel で最大 5 秒追加。ただしタイムアウトは稀なので通常は数十 ms

**長時間系 hook (PostToolUse) の扱い**:

現行の `auto-test-runner` (120s, async) や `ci-status-checker` (30s, async) は
短命プロセスに収まらない。これらは **分離ワーカー** として扱う:

```json
// hooks.json で Go binary と分離ワーカーを並列配置
"PostToolUse": [{
  "matcher": "Write|Edit|MultiEdit|Bash",
  "hooks": [{
    "type": "command",
    "command": "${CLAUDE_PLUGIN_ROOT}/bin/harness hook posttool",
    "timeout": 10
  }]
}, {
  "matcher": "Write|Edit|Task",
  "hooks": [{
    "type": "command",
    "command": "${CLAUDE_PLUGIN_ROOT}/bin/harness worker auto-test",
    "timeout": 120,
    "async": true
  }]
}, {
  "matcher": "Bash",
  "hooks": [{
    "type": "command",
    "command": "${CLAUDE_PLUGIN_ROOT}/bin/harness worker ci-check",
    "timeout": 30,
    "async": true
  }]
}]
```

`bin/harness worker <name>` は長時間実行用のサブコマンド。
`bin/harness hook posttool` は高速判定 (10s) のみ担当し、重い処理は worker に分離。

## Agent Hooks (type: "agent") の扱い

**決定: hooks.json に残す。Go binary は関与しない。**

agent hooks は CC が LLM を起動して判断を委任する仕組み。Go binary が置換するのは `type: "command"` のフックのみ。

```json
// hooks.json で agent hooks はそのまま残る
{
  "matcher": "Write|Edit",
  "hooks": [
    {
      "type": "command",
      "command": "${CLAUDE_PLUGIN_ROOT}/bin/harness hook pretool",
      "timeout": 10
    },
    {
      "type": "agent",
      "prompt": "Review the code change for quality issues...",
      "model": "haiku",
      "timeout": 30
    }
  ]
}
```

Go binary (command) が高速にガードレール判定を返し、agent hook (LLM) がその後に非同期で品質チェックを行う。責務分離。

## MCP Server の扱い

**決定: 分離プロセス。Go binary に内蔵しない。**

理由:
- harness-mem は SQLite ベースの永続ストア。Go で CGO なしの SQLite は制約が多い
- MCP server は常駐プロセス。hook handler は短命プロセス。ライフサイクルが異なる
- 現行の harness-mem (Node.js) をそのまま使い、Go binary は `memory/bridge.go` でファイルベースの連携のみ行う

```
bin/harness hook session-start → bridge.go → harness-mem にファイル経由で通知
bin/harness hook stop → bridge.go → harness-mem にセッション終了を通知
```

将来 Go 純粋の MCP server が必要になった場合は、`modernc.org/sqlite` (CGO-free) で別バイナリとして実装する。

## Performance Comparison (Estimated)

| Operation | Current (bash+node) | Go rewrite | Speedup |
|-----------|-------------------|------------|---------|
| PreToolUse guardrail | 40-60ms | 2-3ms | **20x** |
| PostToolUse logging | 20-30ms | 1-2ms | **15x** |
| SessionStart init | 500-800ms (4 hooks) | 10-15ms (1 call) | **50x** |
| TaskCompleted + webhook | 100-150ms | 3-5ms | **30x** |
| OTel span export | 100-200ms | 2-3ms (async) | **50x** |
| Plans.md parse | 50-100ms (bash+jq) | 1-2ms | **50x** |
| Total per tool call | 60-90ms overhead | 3-5ms overhead | **20x** |

**Breezing 1000 回のツール呼び出し: 60-90 秒 → 3-5 秒**

### Worst Case Analysis (PreToolUse)

```
stdin JSON read:     0.1ms  (io.ReadAll, unbounded — large Write/Edit payloads safe)
json.Unmarshal:      0.3ms  (encoding/json, typed struct)
Rule matching loop:  0.5ms  (13 rules × Match func, short-circuit on first deny)
State dir access:    0.5ms  (single os.Stat for CLAUDE_PLUGIN_DATA resolution)
JSONL append:        0.3ms  (safefile.Append, includes Lstat check)
json.Marshal output: 0.1ms  (<200B response)
─────────────────────────────
Worst case total:    1.8ms
```

SessionStart (parallel goroutines):
```
env setup:           2ms   ┐
rules load:          1ms   ├── parallel → max(2, 1) = 2ms
memory bridge file:  3ms   │  (sequential after env)
─────────────────────────────
Worst case total:    5ms
```

## What Stays Markdown

| Component | Format | Reason |
|-----------|--------|--------|
| skills-v3/*.md | Markdown | CC がプロンプトとして読む。コンパイル不要 |
| agents-v3/*.md | Markdown | CC がプロンプトとして読む |
| .claude/rules/*.md | Markdown | CC がルールとして読む |
| CLAUDE.md | Markdown | CC が instructions として読む |
| Plans.md | Markdown | Go がパースするが、人間も読む |
| CHANGELOG.md | Markdown | リリースノート |

## What Disappears

| Current | Go Rewrite | Reason |
|---------|------------|--------|
| 40+ shell scripts in scripts/ | 0 | 全て Go binary に統合 |
| core/src/ TypeScript | 0 | Go internal/ に置換 |
| node_modules/ | 0 | Go は stdlib 完結 |
| scripts/run-hook.sh routing | 0 | Go の subcommand routing |
| jq dependency | 0 | Go encoding/json |
| scripts/path-utils.sh | 0 | Go internal/config/paths.go |
| scripts/sync-plugin-cache.sh | 0 | バイナリ 1 つで同期不要 |

## Build & Distribution

```makefile
# Makefile

PLATFORMS := darwin/arm64 darwin/amd64 linux/amd64 linux/arm64 windows/amd64

.PHONY: build
build:
	go build -o bin/harness ./cmd/harness

.PHONY: release
release:
	$(foreach platform,$(PLATFORMS),\
		GOOS=$(word 1,$(subst /, ,$(platform))) \
		GOARCH=$(word 2,$(subst /, ,$(platform))) \
		go build -ldflags="-s -w" \
		-o bin/harness-$(subst /,-,$(platform))$(if $(findstring windows,$(platform)),.exe) ./cmd/harness ;)

.PHONY: test
test:
	go test ./...

.PHONY: lint
lint:
	golangci-lint run
```

## Migration Path

This is a **zero-base rewrite**, not a migration. Both versions coexist:

- `main` branch: Current bash+TS version (production)
- `feat/harness-go-rewrite` branch: Go version (development)

Switch is atomic: replace `.claude-plugin/` and `bin/` in one commit.

## Design Decisions (formerly Open Questions)

| # | 質問 | 決定 | 理由 |
|---|------|------|------|
| D1 | agent hooks (type: "agent") | **hooks.json に残す。Go は command hooks のみ** | LLM judgment は CC の責務。Go は高速なルール評価に集中 |
| D2 | Codex companion | **現行 shell wrapper を維持** | companion は codex-plugin-cc の proxy。Go 化の ROI が低い |
| D3 | Memory MCP | **分離プロセス。Go に内蔵しない** | SQLite CGO 問題、ライフサイクル不一致。Node 版を継続利用 |
| D4 | Plugin bin/ auto-selection | **CC の bin/ feature を使う** | CC v2.1.91+ がプラットフォーム別にバイナリを選択。Makefile で命名規則を合わせるだけ |
| D5 | パッケージ構造 | **6 パッケージに統合** (guardrail, session, event, plans, notify, config) | 11 は過多。Go 慣習に従い機能密度を上げる |
| D6 | review (security/dual) | **Go binary から除外。スキルのプロンプト指示で完結** | review の判断は LLM が行う。Go が持つ必要がない |
| D7 | 外部依存 | **uuid のみ許容。他は stdlib** | MCP/OTel も encoding/json + net/http で自前実装 |

## Binary Size Estimate (Revised)

```
Go stdlib minimal:     1.5MB
+ net/http:           +0.5MB
+ encoding/json:       (included)
+ google/uuid:        +0.1MB
+ ldflags -s -w:      -30%
─────────────────────────────
Expected:              ~2.5MB (darwin/arm64)
```

5 プラットフォーム合計: ~12MB（bin/ ディレクトリ全体）