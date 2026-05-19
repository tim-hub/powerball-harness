# Hook Features: CC 2.1.139+

Documents Claude Code hook features introduced in or after v2.1.139.
Harness uses these features in `harness/scripts/` hook handlers.

---

## 1. `continueOnBlock` Field

When a `stop` hook returns `{"decision":"block"}`, Claude Code normally halts
the session. Setting `continueOnBlock: true` on the hook causes CC to ignore
the block and continue — useful for advisory-only hooks that should never halt.

### Schema

```json
{
  "hooks": {
    "Stop": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "bash harness/scripts/stop-plans-reminder.sh",
            "continueOnBlock": true
          }
        ]
      }
    ]
  }
}
```

### Harness usage

`harness/scripts/stop-plans-reminder.sh` — the reminder is advisory; blocking
would frustrate users when Plans.md still has WIP tasks.

---

## 2. Exec Form (`args` Array)

Instead of a shell `command` string, you can specify a command as an `args`
array. CC executes it directly via `execvp` — no shell quoting issues, no
PATH surprises, and faster startup.

### Schema

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "args": ["bash", "harness/scripts/hook-handlers/post-tool-use.sh"]
          }
        ]
      }
    ]
  }
}
```

### Shell form (compatible with all versions)

```json
{
  "type": "command",
  "command": "bash harness/scripts/hook-handlers/post-tool-use.sh"
}
```

Use the `args` form when you need guaranteed argument boundaries (e.g., a
script path with spaces, or passing JSON arguments without extra quoting).

---

## 3. `$CLAUDE_EFFORT` Environment Variable

CC 2.1.139+ exports `$CLAUDE_EFFORT` to every hook subprocess. The value
reflects the current effort level set by the user (`low`, `medium`, or `high`).

### Values

| Value    | Corresponds to |
|----------|---------------|
| `low`    | `○` (low effort) |
| `medium` | `◐` (medium effort, default for Opus 4.6) |
| `high`   | `●` (high effort) |

### Harness usage

Hook handlers can inspect `$CLAUDE_EFFORT` to gate expensive operations:

```bash
# Skip heavy analysis during low-effort runs
if [ "${CLAUDE_EFFORT:-medium}" = "low" ]; then
  echo '{"decision":"approve","reason":"effort=low, skipping heavy check"}'
  exit 0
fi
```

None of the current Harness hook handlers gate on effort level. This is
available for project-level customization.

---

## 4. `terminalSequence` Hook Output Field

CC 2.1.141+ (semantic superset of 2.1.139+) processes a `terminalSequence`
field in hook JSON responses. The field value is written verbatim to the
terminal — enabling bell beeps, window-title updates, and desktop notifications
without a separate process.

### Schema

```json
{
  "decision": "approve",
  "reason": "task completed",
  "terminalSequence": "]9;Claude Code: task completed"
}
```

### Opt-in via `HARNESS_TERMINAL_NOTIFY`

Harness gates `terminalSequence` output on the `HARNESS_TERMINAL_NOTIFY`
environment variable (unset by default — no breaking change for existing users).

| Value    | Sequence emitted |
|----------|-----------------|
| unset/`0` | none (off) |
| `1`/`bell` | BEL (`\x07`) |
| `title` | OSC 0 — window title update |
| `osc9` | OSC 9 — macOS/iTerm notification |
| `notify` | OSC 777 — KDE/GNOME desktop notification |

### Handlers with `terminalSequence` support

| Handler | Go source | Shell source |
|---------|-----------|--------------|
| Notification | `go/internal/hookhandler/notification_handler.go` | `harness/scripts/hook-handlers/notification-handler.sh` |
| TaskCompleted | `go/internal/hookhandler/task_completed.go` | — |
| Webhook | — | `harness/scripts/hook-handlers/webhook-notify.sh` |

The shared Go helper is `go/internal/hookhandler/terminal_notify.go`;
the shared shell helper is `harness/scripts/lib/terminal-notify.sh`.

### Security

Control characters (0x00–0x1F, 0x7F) are stripped from `title` and `body`
inputs before sequence construction to prevent terminal corruption via
injected OSC/ESC sequences. Printable non-ASCII characters pass through.

---

## Compatibility Matrix

| Feature | Minimum CC version | Harness implementation |
|---------|-------------------|----------------------|
| `continueOnBlock` | 2.1.139 | `stop-plans-reminder.sh` |
| `args` exec form | 2.1.139 | optional — `command` form still used |
| `$CLAUDE_EFFORT` | 2.1.139 | available; no handlers gate on it |
| `terminalSequence` | 2.1.141 | `terminal_notify.go`, `terminal-notify.sh` |

Harness hooks are backward-compatible with CC versions before 2.1.139 —
the new fields are silently ignored by older CC versions.
