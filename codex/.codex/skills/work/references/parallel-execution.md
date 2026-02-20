# Parallel Execution

## Policy

- 独立タスクは並列
- 依存タスクは順次

## Runtime Pattern

1. 並列グループ生成
2. 各グループで `spawn_agent` を複数起動
3. `wait` で全完了待ち
4. 失敗時は `send_input` で再試行指示

## Notes

- 並列上限は `--parallel N` を優先
- 未指定時は安全側で自動決定
