---
name: 2agent
description: "PMと実装役の二人三脚を設定。息ぴったりの開発体制を構築。Use when user mentions 2-Agent setup, PM coordination, Cursor setup, or 2-agent operations. Do NOT load for: solo operation, workflow execution, or handoff processing."
allowed-tools: ["Read", "Write", "Edit", "Bash"]
argument-hint: "[setup|cursor-rules|auto]"
---

# 2-Agent Skills

2-Agent ワークフローの設定を担当するスキル群です。

## 機能詳細

| 機能 | 詳細 |
|------|------|
| **初期設定** | See [references/2agent-setup.md](references/2agent-setup.md) |
| **ファイル更新** | See [references/2agent-updating.md](references/2agent-updating.md) |
| **Cursor連携** | See [references/cursor-setup.md](references/cursor-setup.md) |

## 実行手順

1. ユーザーのリクエストを分類
2. 上記の「機能詳細」から適切な参照ファイルを読む
3. その内容に従って設定
