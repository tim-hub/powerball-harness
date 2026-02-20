# Execution Flow

## Phase 1: Plan Parsing

1. Plans.md を読み込み
2. 対象範囲を選定
3. 依存解析で並列グループを作成

## Phase 2: Multi-Agent Run

1. グループごとに `spawn_agent` で implementer を起動
2. `wait` でグループ完了を収束
3. 次グループへ進む

## Phase 3: Review

1. reviewer を `spawn_agent` で起動
2. findings があれば `send_input` で修正依頼
3. APPROVE で次へ

## Phase 4: Finalize

1. build/test
2. Plans.md 更新
3. commit（`--no-commit` なし時）
4. `close_agent`

## Engine Rule

- 既定: Codex実装 + Codexレビュー
- `--claude`: Claude実装 + Claudeレビュー
