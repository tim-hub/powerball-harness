# Codex Engine (Breezing)

`/breezing` は Codex ネイティブのマルチエージェントで完走する。
`--claude` 指定時のみ実装とレビューを Claude に切り替える。

## Mode Matrix

| 項目 | デフォルト (`/breezing`) | `--claude` |
|------|--------------------------|------------|
| `impl_engine` | `codex` | `claude` |
| `review_engine` | `codex` | `claude` |
| 実装担当 | Codex implementer role | Claude implementer role |
| レビュー担当 | Codex reviewer role | Claude reviewer role |

## Native Tools

- `spawn_agent`: 実装/レビュー担当を起動
- `wait`: 並列エージェント完了待ち
- `send_input`: リテイク指示
- `resume_agent`: 再開
- `close_agent`: 終了クリーンアップ

## Review Policy

- 既定: Codexレビュー
- `--claude`: Claudeレビュー固定
- `--claude + --codex-review`: 禁止（開始前エラー）

## State

セッション状態は `${CODEX_HOME:-~/.codex}/state/harness/breezing-active.json` を使用する。
