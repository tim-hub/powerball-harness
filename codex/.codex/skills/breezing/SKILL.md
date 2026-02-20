---
name: breezing
description: "CodexのマルチエージェントでPlans.mdを完走。Use when user mentions '/breezing', team execution, full auto completion, multi-agent workflow, 'チームで完走', 'チームで全部'. Do NOT load for: single tasks, reviews, setup, or /work (direct implementation)."
description-en: "Complete Plans.md with Codex native multi-agent orchestration. Use when user mentions '/breezing', team execution, full auto completion, multi-agent workflow. Do NOT load for: single tasks, reviews, setup, or /work (direct implementation)."
description-ja: "CodexのマルチエージェントでPlans.mdを完走。Use when user mentions '/breezing', team execution, full auto completion, multi-agent workflow, 'チームで完走', 'チームで全部'. Do NOT load for: single tasks, reviews, setup, or /work (direct implementation)."
allowed-tools: ["Read", "Write", "Edit", "Grep", "Glob", "Bash", "Task"]
argument-hint: "[all] [task-number|range] [--claude] [--codex] [--parallel N] [--no-commit]"
disable-model-invocation: true
---

# Breezing Skill

Codex ネイティブのマルチエージェントで Plans.md の未完了タスクを**全自動完走**する。

## Philosophy

> **「Lead は統括、実装とレビューは専任エージェント」**
>
> 既定では Codex 実装 + Codex レビュー。
> `--claude` 指定時のみ、実装もレビューも Claude に委譲。

## Quick Reference

```bash
/breezing                                     # スコープを聞いてから実行
/breezing all                                 # Plans.md 全タスクを完走
/breezing 3-6                                 # タスク3〜6を完走
/breezing 認証機能からユーザー管理まで          # 自然言語で範囲指定
/breezing --claude all                        # Claude 実装 + Claude レビュー
/breezing --claude --parallel 2 ログイン機能   # Claude 2並列
/breezing --codex all                         # 互換エイリアス（既定と同じ）
/breezing --parallel 2 ログイン機能            # Codex 2並列
/breezing 続きやって                           # 前回中断から再開
```

## Options

| Option | Description | Default |
|--------|-------------|---------|
| `all` | 全未完了タスクを対象 | - |
| `N` or `N-M` | タスク番号/範囲指定 | - |
| `--claude` | 実装・レビューを Claude CLI に委譲 | false |
| `--codex` | 互換エイリアス（既定と同じ Codex 実行） | true (既定) |
| `--parallel N` | Implementer 並列数 | auto |
| `--no-commit` | 自動コミット抑制 | false |

## Scope Dialog (引数なし時)

```text
/breezing
どこまでやりますか?
1) 全部 (推奨): 残りタスクを完走
2) 指定: タスク番号や範囲を指定 (例: 3, 3-6)
```

## `/work` との違い

| 特徴 | `/work` | `/breezing` |
|------|---------|-------------|
| スコープ | 単発〜中規模 | 大規模完走 |
| 並列 | 必要時のみ | 強めに並列 |
| レビュー | 1サイクル中心 | リテイク込み完走 |
| デフォルト対象 | 次タスク | 全タスク |

## Codex-First Engine

| 項目 | デフォルト | `--claude` |
|------|-----------|------------|
| `impl_engine` | `codex` | `claude` |
| `review_engine` | `codex` | `claude` |
| 実装経路 | Codex マルチエージェント | Claude CLI ワーカー |
| レビュー経路 | Codex レビューエージェント | Claude CLI レビューエージェント |

互換性メモ: `--codex` は legacy alias。挙動は既定と同じ。

## フラグ整合ルール

- `--claude` + `--codex-review` は同時指定不可。
- 同時指定時は開始前にエラー終了する。

## Native Multi-Agent Flow

`spawn_agent` / `wait` / `send_input` / `resume_agent` / `close_agent` を使って、
Lead が実装担当とレビュー担当を統括する。

## State Path

- セッション状態は `${CODEX_HOME:-~/.codex}/state/harness/` に保存する。
- 旧来の Claude用状態パスは使用しない。

## Feature Details

| Feature | Reference |
|---------|-----------|
| **Execution Flow** | See [references/execution-flow.md](references/execution-flow.md) |
| **Team Composition** | See [references/team-composition.md](references/team-composition.md) |
| **Review/Retake Loop** | See [references/review-retake-loop.md](references/review-retake-loop.md) |
| **Plans.md Mapping** | See [references/plans-to-tasklist.md](references/plans-to-tasklist.md) |
| **Codex Engine** | See [references/codex-engine.md](references/codex-engine.md) |
| **Session Resilience** | See [references/session-resilience.md](references/session-resilience.md) |
| **Guardrails** | See [references/guardrails-inheritance.md](references/guardrails-inheritance.md) |
| **Codex Review Integration** | See [references/codex-review-integration.md](references/codex-review-integration.md) |

## Prerequisites

1. Codex CLI `>= 0.102.0`
2. `features.multi_agent = true`
3. Plans.md に未完了タスクがあること
4. `--claude` 使用時のみ `which claude` が成功すること

## Completion Conditions

1. 指定範囲の全タスクが `cc:done`
2. 統合ビルド成功
3. 全テスト通過
4. 最終レビューが APPROVE

## Completion Tip

```text
Done! 5 tasks completed by team.
Tip: /work で1タスクだけ進めることもできます
Tip: --claude を付けると実装・レビューを Claude に委譲できます
```

## VibeCoder Hints

| What You Want | How to Say |
|---------------|------------|
| 全部終わらせて | `/breezing all` |
| Codex で全部 | `/breezing all` |
| Claude で全部 | `/breezing --claude all` |
| この機能だけ | `/breezing ログイン機能を完了して` |
| ここからここまで | `/breezing 認証からユーザー管理まで` |
| 前回の続きから | `/breezing 続きやって` |
| 1タスクだけ | → `/work` |

## Related Skills

- `work` - Codex が直接実装
- `harness-review` - コードレビュー
- `codex-review` - Codex セカンドオピニオンレビュー
