# Execution Flow

## Phase A: Prepare

1. scope を確定
2. `${CODEX_HOME:-~/.codex}/state/harness/breezing-active.json` を初期化
3. Plans.md を実行単位へ変換
4. `spawn_agent` で implementer / reviewer を起動

## Phase B: Run

1. implementer 群が並列実装
2. `wait` で実装完了を収束
3. reviewer が判定
4. NG時は `send_input` でリテイク指示
5. APPROVE まで反復

## Phase C: Close

1. 統合検証（build/test）
2. Plans.md 更新
3. commit（`--no-commit` なし時）
4. `close_agent` 実行
5. 状態ファイル削除

## Guard Conditions

- `--claude` 時はレビューも Claude に固定
- `--claude + --codex-review` はエラー終了
