---
name: work
description: "Plans.mdのタスクを実装。Codexネイティブのマルチエージェントで単発から全体まで実行。Use when user mentions '/work', execute plan, implement tasks, build features, '実装して', '全部やって'."
description-en: "Execute Plans.md tasks with Codex native multi-agent from single task to full completion."
description-ja: "Plans.mdのタスクを実装。Codexネイティブのマルチエージェントで単発から全体まで実行。Use when user mentions '/work', execute plan, implement tasks, build features, '実装して', '全部やって'."
allowed-tools: ["Read", "Write", "Edit", "Grep", "Glob", "Bash", "Task"]
argument-hint: "[all] [task-number|range] [--claude] [--codex] [--parallel N] [--no-commit] [--resume id]"
disable-model-invocation: true
---

# Work Skill

Plans.md のタスクを実装する主力スキル。
Codex ネイティブマルチエージェントで、単発から全体完了まで実行する。

## Quick Reference

```bash
/work                    # スコープを聞いて実行
/work 3                  # タスク3を実行
/work all                # 全タスク実行
/work 3-6                # 範囲実行
/work --claude           # 実装・レビューをClaudeへ委譲
/work --claude all       # Claudeで全体実行
/work --codex            # 互換エイリアス（既定と同じ）
/work --parallel 5 all   # 並列5で全体実行
/work --no-commit        # 自動コミット抑制
/work --resume latest    # 前回再開
```

## Options

| Option | Description | Default |
|--------|-------------|---------|
| `all` | 全未完了タスク | - |
| `N` or `N-M` | タスク番号/範囲指定 | - |
| `--claude` | 実装・レビューを Claude CLI に委譲 | false |
| `--codex` | 互換エイリアス（既定と同じ） | true (既定) |
| `--parallel N` | 並列ワーカー数 | auto |
| `--no-commit` | 自動コミット抑制 | false |
| `--max-iterations N` | 反復上限（all時） | 10 |
| `--resume <id|latest>` | セッション再開 | - |

## Scope Dialog

```text
/work
どこまでやりますか?
1) 次のタスク (推奨)
2) 全部
3) 指定 (例: 3, 3-6)
```

## Strategy Selection

| スコープ | 戦略 |
|---------|------|
| 1タスク | 単体実装 |
| 2-3タスク | 並列実行 |
| 4+ / all | 並列 + 自動反復 |

## Codex-First Engine

| 項目 | デフォルト | `--claude` |
|------|-----------|------------|
| `impl_engine` | `codex` | `claude` |
| `review_engine` | `codex` | `claude` |
| 実装経路 | Codex roles | Claude roles |
| レビュー経路 | Codex reviewer | Claude reviewer |

互換性メモ: `--codex` は legacy alias。

## Native Multi-Agent Tools

- `spawn_agent`
- `wait`
- `send_input`
- `resume_agent`
- `close_agent`

## Flag Validation

- `--claude + --codex-review` は同時指定不可（開始前エラー）

## State Path

- `${CODEX_HOME:-~/.codex}/state/harness/work-active.json`
- `${CODEX_HOME:-~/.codex}/state/harness/work.log.jsonl`

## Feature Details

| Feature | Reference |
|---------|-----------|
| Scope Dialog | [references/scope-dialog.md](references/scope-dialog.md) |
| Auto Iteration | [references/auto-iteration.md](references/auto-iteration.md) |
| Codex Engine | [references/codex-engine.md](references/codex-engine.md) |
| Execution Flow | [references/execution-flow.md](references/execution-flow.md) |
| Parallel Execution | [references/parallel-execution.md](references/parallel-execution.md) |
| Session Management | [references/session-management.md](references/session-management.md) |
| Review Loop | [references/review-loop.md](references/review-loop.md) |
| Auto Commit | [references/auto-commit.md](references/auto-commit.md) |
| Error Handling | [references/error-handling.md](references/error-handling.md) |

## Completion Tip

```text
Done! tasks completed.
Tip: /breezing で大規模範囲を完走できます
Tip: --claude で実装・レビューを Claude 委譲に切り替えできます
```
