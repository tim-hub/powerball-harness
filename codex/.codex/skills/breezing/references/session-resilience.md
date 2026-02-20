# Session Resilience

## State Files

- `${CODEX_HOME:-~/.codex}/state/harness/breezing-active.json`
- `${CODEX_HOME:-~/.codex}/state/harness/breezing-timeline.jsonl`

## Resume Flow

1. active state を読む
2. 未完了実行単位を復元
3. 必要な agent を `resume_agent` / `spawn_agent` で再接続
4. `wait` 待機を再開

## Corruption Handling

状態ファイルが壊れている場合はバックアップを作成し、新規セッションで再開する。
