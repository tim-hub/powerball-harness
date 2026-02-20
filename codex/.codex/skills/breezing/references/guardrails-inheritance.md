# Guardrails

Codex 実行時の制約は次の優先順で適用する。

1. `harness.rules`
2. skill 内の hard rule
3. runtime flag rule（`--claude` routing）

## Required Guarantees

- `--claude` 時は `impl_engine=claude` かつ `review_engine=claude`
- `--claude + --codex-review` は拒否
- 危険コマンドは `harness.rules` に従って確認を挟む
