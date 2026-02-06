---
name: breezing
description: "楽勝で流す。Agent Teamsで完全自走、寝てる間にゴール。Use when user mentions '/breezing', agent teams, team execution, full auto completion, or multi-agent workflow. Do NOT load for: single tasks, reviews, setup, ultrawork, or codex worker."
description-en: "Auto-complete Plans.md with Agent Teams, fully autonomous. Use when user mentions '/breezing', agent teams, team execution, full auto completion, or multi-agent workflow. Do NOT load for: single tasks, reviews, setup, ultrawork, or codex worker."
description-ja: "楽勝で流す。Agent Teamsで完全自走、寝てる間にゴール。Use when user mentions '/breezing', agent teams, team execution, full auto completion, or multi-agent workflow. Do NOT load for: single tasks, reviews, setup, ultrawork, or codex worker."
allowed-tools: ["Read", "Write", "Edit", "Grep", "Glob", "Bash", "Task"]
argument-hint: "[natural language range] [--codex-review] [--parallel N]"
disable-model-invocation: true
---

# Breezing Skill

Agent Teams を活用して Plans.md の未完了タスクを**チーム協調で完全自動完走**する。

## Philosophy

> **「Lead は指揮するだけ、手を動かすのは Teammate」**
>
> delegate mode で Lead は調整に専念。
> 実装は Implementer、レビューは Reviewer。三者分離の完全自律。

## Quick Reference

```bash
/breezing 全部やって                    # Plans.md 全タスクを完走
/breezing 認証機能からユーザー管理まで    # 範囲指定
/breezing --codex-review 全部やって      # Codex MCP レビュー付き
/breezing --parallel 2 ログイン機能      # 並列数指定
/breezing 続きやって                     # 前回中断から再開
```

## `/work` / `/ultrawork` との違い

| 特徴 | `/work` | `/ultrawork` | `/breezing` |
|------|---------|-------------|-------------|
| 規模 | 1-3 タスク | 3-10 タスク | 制限なし |
| 並列手段 | Task tool | Task tool | **Agent Teams** |
| Lead の役割 | 調整+実装 | 調整+実装 | **delegate (調整専念)** |
| レビュー | Lead 自己レビュー | Lead 自己レビュー | **独立 Reviewer Teammate** |
| リテイク | 手動 | 自動 (Lead 自己修正) | **自動 (Lead 分解 → Impl 修正)** |
| コスト | 低 | 中 | 高 |

## Feature Details

| Feature | Reference |
|---------|-----------|
| **Execution Flow** | See [references/execution-flow.md](references/execution-flow.md) |
| **Team Composition** | See [references/team-composition.md](references/team-composition.md) |
| **Review/Retake Loop** | See [references/review-retake-loop.md](references/review-retake-loop.md) |
| **Plans.md → TaskList** | See [references/plans-to-tasklist.md](references/plans-to-tasklist.md) |
| **Codex Review Integration** | See [references/codex-review-integration.md](references/codex-review-integration.md) |
| **Session Resilience** | See [references/session-resilience.md](references/session-resilience.md) |
| **Guardrails Inheritance** | See [references/guardrails-inheritance.md](references/guardrails-inheritance.md) |

## Prerequisites

1. **Agent Teams 有効化**: `settings.json` に `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1`
2. **Plans.md** が存在し、未完了タスクがあること
3. (--codex-review) Codex MCP サーバーが登録済み

## Execution Flow Summary

```
/breezing [range] [--codex-review] [--parallel N]
    │
準備: 範囲確認 → ユーザー承認 → Team 初期化 → delegate mode → Teammates spawn
    │
実装・レビューサイクル (Lead の判断で柔軟に運用):
  ├── 実装: Implementer 並列実装 → self-claim → build/test
  ├── レビュー: Reviewer 独立レビュー (部分/全体、任意タイミング)
  └── リテイク: findings → 修正タスク → 再レビュー (直接対話可)
    │
完了: 全タスク完了 + APPROVE → 統合検証 → commit → cleanup
```

## Completion Conditions

以下を**全て**満たしたとき完了:

1. 指定範囲の全タスクが `cc:done`
2. 統合ビルド成功
3. 全テスト通過
4. Reviewer が最終 APPROVE (Critical/Major findings = 0)

## VibeCoder Hints

| What You Want | How to Say |
|---------------|------------|
| 全部終わらせて | `/breezing 全部やって` |
| レビュー付きで | `/breezing --codex-review 全部やって` |
| この機能だけ | `/breezing ログイン機能を完了して` |
| ここからここまで | `/breezing 認証からユーザー管理まで` |
| 前回の続きから | `/breezing 続きやって` |
