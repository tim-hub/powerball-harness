# Error Handling

## Common Cases

- 実装失敗: 対象agentへ再指示
- レビューNG: 修正ループへ戻す
- 設定不整合: 開始前エラーで停止

## Flag Error

`--claude + --codex-review` は不正指定として即停止する。
