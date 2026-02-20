# Review / Retake Loop

## Loop

1. 実装完了後に reviewer が判定
2. findings があれば修正単位へ分解
3. `send_input` で implementer へ修正依頼
4. 再レビュー
5. APPROVE まで繰り返し

## Engine Routing

- デフォルト: Codex reviewer
- `--claude`: Claude reviewer

## Hard Rule

`--claude` 指定時に Codex reviewer へ切り替えない。
