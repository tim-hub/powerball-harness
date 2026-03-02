---
name: work-harness
description: "Unified execution skill for Harness v3. Implements Plans.md tasks from single task to full parallel team runs. Use when user mentions: implement, execute, /work-harness, /work, do everything, build features, run tasks, breezing, team run, --codex, --parallel. Do NOT load for: planning, code review, release, or setup."
description-ja: "Harness v3 統合実行スキル。Plans.md タスクを1件から全並列チーム実行まで担当。以下で起動: 実装して、実行して、/work-harness、/work、全部やって、breezing、チーム実行、--codex、--parallel。プランニング・レビュー・リリース・セットアップには使わない。"
allowed-tools: ["Read", "Write", "Edit", "Grep", "Glob", "Bash", "Task"]
argument-hint: "[all] [task-number|range] [--codex] [--parallel N] [--no-commit] [--resume id] [--breezing]"
---

# Execute Skill (v3)

Harness v3 の統合実行スキル。
以下の旧スキルを統合:

- `work` — Plans.md タスクの実装（スコープ自動判断）
- `impl` — 機能実装（タスクベース）
- `breezing` — チームフル自動実行（Agent Teams）
- `parallel-workflows` — 並列ワークフロー最適化
- `ci` — CI 失敗時の復旧

## Quick Reference

| ユーザー入力 | モード | 動作 |
|------------|--------|------|
| `/execute` | solo | スコープ確認 → 実装 |
| `/execute all` | solo | 全未完了タスクを即実行 |
| `/execute 3` | solo | タスク3だけ即実行 |
| `/execute --parallel 5` | solo-parallel | 5ワーカーで並列実行 |
| `/execute --codex` | codex | Codex CLI に委託 |
| `/execute --breezing` | team | Agent Teams でチーム実行 |

## オプション

| オプション | 説明 | デフォルト |
|----------|------|----------|
| `all` | 全未完了タスクを対象 | - |
| `N` or `N-M` | タスク番号/範囲指定 | - |
| `--parallel N` | 並列ワーカー数 | auto |
| `--sequential` | 直列実行強制 | - |
| `--codex` | Codex CLI で実装委託 | false |
| `--no-commit` | 自動コミット抑制 | false |
| `--resume <id\|latest>` | 前回セッション再開 | - |
| `--breezing` | Agent Teams でチーム実行 | false |
| `--no-simplify` | Auto-Refinement スキップ | false |

## スコープダイアログ（引数なし時）

```
/execute
どこまでやりますか?
1) 次のタスク（推奨）: Plans.md の次の未完了タスク
2) 全部: 残りのタスクをすべて完了
3) 番号指定: タスク番号を入力（例: 3, 5-7）
```

## 実行モード詳細

### Solo モード（デフォルト）

1. Plans.md を読み込み、対象タスクを特定
2. タスクを `cc:WIP` に更新
3. コードを実装（Read/Write/Edit/Bash）
4. `/simplify` で Auto-Refinement（`--no-simplify` で省略可）
5. `git commit` で自動コミット（`--no-commit` で省略可）
6. タスクを `cc:完了` に更新

### Parallel モード（`--parallel N`）

`[P]` マーク付きタスクを N ワーカーで並列実行。
同一ファイルへの書き込みが競合する場合は git worktree で分離。

### Codex モード（`--codex`）

```bash
TIMEOUT=$(command -v timeout || command -v gtimeout || echo "")
$TIMEOUT 120 codex exec "$(cat /tmp/codex-prompt.md)" 2>/dev/null
```

タスク内容を `/tmp/codex-prompt.md` に書き出してから Codex CLI に委託。
結果を検証し、品質基準を満たさない場合は自力で修正。

### Breezing モード（`--breezing`）

Agent Teams（Worker + Reviewer）でチーム実行。

```
Lead (this agent)
├── Worker (task-worker agent) — 実装担当
└── Reviewer (code-reviewer agent) — レビュー担当
```

**フロー**:
1. Lead: タスク割り当て（TaskCreate/TaskUpdate）
2. Worker: 実装 → `cc:完了`
3. Reviewer: コードレビュー → APPROVE / REQUEST_CHANGES
4. REQUEST_CHANGES の場合: 修正タスクを作成 → 再実装

## CI 失敗時の対応

CI が失敗した場合:

1. ログを確認してエラーを特定
2. 修正を実施
3. 同一原因で 3 回失敗したら自動修正ループを停止
4. 失敗ログ・試みた修正・残る論点をまとめてエスカレーション

## 関連スキル

- `plan` — 実行するタスクを計画する
- `review` — 実装のレビュー
- `release` — バージョンバンプ・リリース
