# Codex Review Integration

`--codex-review` は Codex 実行時の拡張レビュー。

## Availability

- 既定モード（Codex実装）で利用可能
- `--claude` 指定時は利用不可（レビューはClaude固定）

## Conflict Rule

`--claude` と `--codex-review` を同時指定した場合は開始前にエラー終了する。
