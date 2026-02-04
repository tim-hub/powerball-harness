---
name: plan-feature
description: "feature_request を Plans.md のタスクに分解し、cc:TODO で追記する。"
allowed-tools: ["Read", "Edit"]
---

# Plan Feature

要望を実行可能なタスクに分解し、Plans.md に追加するスキル。

---

## 入力

- **Plans.md**: 既存タスク一覧
- **feature_request**: 要望の要約（文字列）
- **user_prompt**: ユーザーの指示（任意）
- **tech_stack**: 技術スタック（任意）

---

## 出力

- Plans.md へのタスク追加（`cc:TODO`）
- 変数: `task_count`, `phase_name`

---

## 実行手順

1. `feature_request` を 3〜7 個のタスクに分解  
2. 依存関係がある場合は `depends:` を付与  
3. Plans.md の「未着手」セクションに追記  
4. `task_count` と `phase_name` を出力  

---

## 出力例

```json
{
  "task_count": 4,
  "phase_name": "実装"
}
```

