---
name: breezing
description: "チーム実行モード（Codex ネイティブ版）— harness-work のチーム協調エイリアス。breezing, チーム実行, 全部やって でトリガー。"
description-en: "Team execution mode (Codex native) — backward-compatible alias for harness-work with team orchestration using Codex native subagent API."
description-ja: "チーム実行モード（Codex ネイティブ版）— harness-work のチーム協調エイリアス。breezing, チーム実行, 全部やって でトリガー。"
argument-hint: "[all|N-M|--max-workers N|--no-discuss]"
user-invocable: true
effort: high
---

# Breezing — Team Execution Mode (Codex Native)

> **この SKILL.md は Codex CLI ネイティブ版です。**
> Claude Code 版は `skills-v3/breezing/SKILL.md` を参照してください。
> サブエージェント API は Codex の `spawn_agent` / `send_input` / `wait_agent` / `close_agent` を使用します。

**後方互換エイリアス**: `harness-work --breezing` をチーム実行モードで動かします。

## Quick Reference

```bash
breezing                        # スコープを聞いてから実行
breezing all                    # Plans.md 全タスクを完走
breezing 3-6                    # タスク3〜6を完走
breezing --max-workers 2 all     # 独立タスクの同時 spawn 上限を2に
breezing --no-discuss all       # 計画議論スキップで全タスク完走
```

## Options

| Option | Description | Default |
|--------|-------------|---------|
| `all` | 全未完了タスクを対象 | - |
| `N` or `N-M` | タスク番号/範囲指定 | - |
| `--max-workers N` | 独立タスクの同時 spawn 数上限（breezing 固有オプション） | 1（直列） |
| `--no-commit` | 非対応（Breezing では Worker の一時 commit と Lead の cherry-pick が必須） | - |
| `--no-discuss` | 計画議論スキップ | false |

## Execution

**このスキルは `harness-work --breezing` に委譲します。** 以下の設定で実行してください:

1. **引数を `harness-work --breezing` に渡す**（`--max-workers N` は breezing 固有オプションとして解釈し、`harness-work` の `--parallel` とは別概念）
2. **チーム実行モードを強制** — Lead → Worker spawn → codex exec Reviewer の三者分離
3. **Lead は delegate 専念** — コードを直接書かない

### `harness-work` との違い

| 特徴 | `harness-work` | `breezing` (このスキル) |
|------|-----------------|------------------------|
| デフォルトモード | Solo / Sequential | **Breezing（チーム実行）** |
| 並列手段 | `codex exec` Bash 並列 | **`spawn_agent` によるサブエージェント委譲** |
| Lead の役割 | 調整+実装 | **delegate (調整専念)** |
| レビュー | Lead 自己レビュー | **codex exec 独立レビュー** |
| デフォルトスコープ | 次のタスク | **全部** |

### Team Composition（Codex Native）

| Role | 実行方式 | 権限 | 責務 |
|------|---------|------|------|
| Lead | (self) | 現セッション継承 | 調整・指揮・タスク分配・cherry-pick |
| Worker ×N | `spawn_agent({message, fork_context})` | セッション権限継承 | 実装（git worktree 分離） |
| Reviewer | `codex exec --sandbox read-only` | read-only | 独立レビュー |

## Flow Summary

```
breezing [scope] [--max-workers N] [--no-discuss]
    │
    ↓ Load harness-work --breezing
    │
Phase 0: Planning Discussion (--no-discuss でスキップ)
Phase A: Pre-delegate（チーム初期化 + worktree 準備）
Phase B: Delegate（Worker 実装 + codex exec レビュー）
Phase C: Post-delegate（統合検証 + Plans.md 更新 + commit）
```

### Phase 0: Planning Discussion（構造化 3 問チェック）

全タスク実行前に、以下の 3 問で計画の健全性を確認する。
`--no-discuss` 指定時は全スキップ。

**Q1. スコープ確認**:
> 「{{N}} 件のタスクを実行します。スコープは適切ですか？」

**Q2. 依存関係確認**（Plans.md に Depends カラムがある場合のみ）:
> 「タスク {{X}} は {{Y}} に依存しています。実行順序は合っていますか？」

**Q3. リスクフラグ**（`[needs-spike]` タスクがある場合のみ）:
> 「タスク {{Z}} は [needs-spike] です。先に spike しますか？」

### Phase A: Pre-delegate

1. Plans.md を読み込み、対象タスクを特定
2. 依存グラフを解析し、実行順序を決定
3. 各タスク用に git worktree を作成

### Phase B: Delegate（Codex Native Subagent Orchestration）

```
for task in execution_order:
    # B-0. 作業ディレクトリ分離
    worktree_path = "/tmp/worker-{task.number}-$$"
    branch_name = "worker-{task.number}-$$"
    git worktree add -b {branch_name} {worktree_path}
    TASK_BASE_REF = git rev-parse HEAD

    # B-1. Worker spawn
    Plans.md: task.status = "cc:WIP"

    worker_id = spawn_agent({
        message: "作業ディレクトリ: {worktree_path} で作業してください。\n\nタスク: {task.内容}\nDoD: {task.DoD}\n\n実装してください。完了後 git commit してください。\n\n完了時、以下の JSON を返してください:\n{\"commit\": \"<hash>\", \"files_changed\": [...], \"summary\": \"...\"}",
        fork_context: true
    })
    wait_agent({ ids: [worker_id] })

    # B-2. Lead がレビュー実行（TASK_BASE_REF 起点）
    # レビュー手順は harness-work の「レビューループ」セクションと完全に同一:
    #   codex exec -C {worktree_path} - --sandbox read-only -o {REVIEW_OUT}
    #   → grep '"verdict"' で APPROVE/REQUEST_CHANGES を抽出
    VERDICT = review_task(worktree_path, TASK_BASE_REF)  # harness-work 参照

    # B-3. 修正ループ（REQUEST_CHANGES 時、最大 3 回）
    review_count = 0
    while VERDICT == "REQUEST_CHANGES" and review_count < 3:
        send_input({
            id: worker_id,
            message: "指摘内容: {issues}\n修正して git commit --amend してください。修正後 JSON を再出力してください。"
        })
        wait_agent({ ids: [worker_id] })
        VERDICT = review_task(worktree_path, TASK_BASE_REF)
        review_count++

    # B-4. Worker 終了
    close_agent({ id: worker_id })

    # B-5. 結果処理
    if VERDICT == "APPROVE":
        commit_hash = git("-C", worktree_path, "rev-parse", "HEAD")
        git cherry-pick --no-commit {commit_hash}
        git commit -m "{task.内容}"
        Plans.md: task.status = "cc:完了 [{short_hash}]"
    else:
        → ユーザーにエスカレーション（Plans.md は cc:WIP のまま）
        → 後続タスクも停止

    # B-6. Worktree クリーンアップ
    git worktree remove {worktree_path}
    git branch -D {branch_name}

    # B-7. Progress feed
    print("📊 Progress: Task {completed}/{total} 完了 — {task.内容}")
```

### 独立タスクの並列 spawn（`--max-workers N` 指定時）

依存のないタスクが複数ある場合、`--max-workers N` で同時 spawn 数を制御:

> **`wait_agent` のセマンティクス**: `wait_agent({ids: [a, b]})` は最初に完了した1つを返す（全完了待ちではない）。
> したがって、全 Worker の完了を待つにはループで個別に `wait_agent` を呼ぶ。

```
# 独立タスク A, B を並列 spawn（各自 worktree 分離済み）
worker_a = spawn_agent({ message: "作業ディレクトリ: /tmp/worker-a-$$ ...", fork_context: true })
worker_b = spawn_agent({ message: "作業ディレクトリ: /tmp/worker-b-$$ ...", fork_context: true })

# 各 Worker の完了を個別に待ち → レビュー → cherry-pick（直列）
# wait_agent は最初の1つを返すので、残りの Worker はまだ動作中
for worker_id in [worker_a, worker_b]:
    wait_agent({ ids: [worker_id] })    # この Worker の完了を待つ
    VERDICT = review_task(worktree_path, TASK_BASE_REF)  # harness-work 参照
    # 修正ループ（必要なら）...
    close_agent({ id: worker_id })
    if VERDICT == "APPROVE":
        cherry-pick → Plans.md 更新
```

> **制約**: 並列化できるのは Depends が `-` の独立タスクのみ。
> レビュー → cherry-pick は直列実行（main への書き込みが競合するため）。

### Worker の出力契約

Worker プロンプトには、完了時に以下の JSON を返すことを明示する:

```json
{
  "commit": "a1b2c3d",
  "files_changed": ["src/foo.ts", "tests/foo.test.ts"],
  "summary": "foo モジュールに bar 機能を追加"
}
```

Lead はこの JSON を解析して commit hash とファイル一覧を取得する。

### Progress Feed（Phase B 中の進捗通知）

```
📊 Progress: Task 1/5 完了 — "harness-work に失敗再チケット化を追加"
📊 Progress: Task 2/5 完了 — "harness-sync に --snapshot を追加"
```

### 完了報告（Phase C）

全タスク完了後、Lead が以下の手順でリッチ完了報告を生成:

1. `git log --oneline {session_base_ref}..HEAD` で全 cherry-pick コミットを収集
2. `git diff --stat {session_base_ref}..HEAD` で全体の変更規模を取得
3. Plans.md の残タスクを抽出
4. Breezing テンプレートに従い出力

## Claude Code 版との差分

| 項目 | Claude Code 版 | Codex ネイティブ版（本ファイル） |
|------|---------------|-------------------------------|
| Worker spawn | `Agent(subagent_type="worker", isolation="worktree")` | `spawn_agent({message, fork_context})` + `git worktree add` |
| 完了待ち | `Agent` の戻り値 | `wait_agent({ids: [id]})` |
| 修正指示 | `SendMessage(to: agentId, message: "...")` | `send_input({id, message})` |
| Worker 終了 | 自動 | `close_agent({id})` |
| レビュー | Codex exec → Reviewer agent fallback | `codex exec --sandbox read-only` のみ |
| 権限 | `bypassPermissions` + hooks | `codex exec`: `--full-auto` / `spawn_agent`: セッション権限継承 |
| Agent Teams | `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS` 環境変数 | Codex native（標準機能） |
| Worktree | `isolation="worktree"` 自動管理 | `git worktree add/remove` 手動管理 |
| モード昇格 | タスク4件以上で自動 | `--breezing` 明示時のみ |

## Related Skills

- `harness-work` — 単一タスクからチーム実行まで（本体）
- `harness-sync` — 進捗同期
- `harness-review` — コードレビュー
