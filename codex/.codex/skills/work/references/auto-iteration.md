# Auto Iteration

大量タスク実行時（4+ または `all`）に反復実行する。

## State

- `${CODEX_HOME:-~/.codex}/state/harness/work-active.json`
- `${CODEX_HOME:-~/.codex}/state/harness/work.log.jsonl`

## Loop

1. 未完了タスク抽出
2. 並列実装
3. レビュー
4. 未完了が残れば次反復
