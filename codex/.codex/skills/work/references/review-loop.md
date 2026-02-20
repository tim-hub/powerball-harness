# Review Loop

## Flow

1. reviewer が変更を判定
2. 重大指摘あり: implementer に修正依頼
3. 再レビュー
4. APPROVE で完了

## Routing

- 既定: Codex reviewer
- `--claude`: Claude reviewer（固定）

## Constraint

`--claude` 時は Codex reviewer にフォールバックしない。
