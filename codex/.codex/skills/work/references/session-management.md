# Session Management

## Paths

- `${CODEX_HOME:-~/.codex}/state/harness/work-active.json`
- `${CODEX_HOME:-~/.codex}/state/harness/work.log.jsonl`

## Commands

```bash
cat "${CODEX_HOME:-$HOME/.codex}/state/harness/work-active.json"
tail -20 "${CODEX_HOME:-$HOME/.codex}/state/harness/work.log.jsonl"
```

## Recovery

`resume` 指定時は state を読み、未完了単位から再開する。
