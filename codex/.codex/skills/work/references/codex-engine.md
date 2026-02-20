# Codex Engine (Work)

`/work` は Codex ネイティブマルチエージェントを既定で使う。
`--claude` 指定時のみ実装とレビューを Claude に固定する。

## Mode Matrix

| 項目 | デフォルト (`/work`) | `--claude` |
|------|----------------------|------------|
| `impl_engine` | `codex` | `claude` |
| `review_engine` | `codex` | `claude` |
| 実装 | Codex implementer role | Claude implementer role |
| レビュー | Codex reviewer role | Claude reviewer role |

## Required Tools

- `spawn_agent`
- `wait`
- `send_input`
- `resume_agent`
- `close_agent`

## Conflict

`--claude + --codex-review` は開始前にエラー。
