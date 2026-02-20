# Plans Mapping

Plans.md の各タスクを breezing 実行単位に変換する。

## Mapping Rules

- Plans.md 1タスク = 実行単位1件
- `blocked` は依存関係として扱う
- mapping は `${CODEX_HOME:-~/.codex}/state/harness/breezing-active.json` に保存

## Stored Fields

- `task_range`
- `plans_md_mapping`
- `parallel`
- `impl_engine`
- `review_engine`
