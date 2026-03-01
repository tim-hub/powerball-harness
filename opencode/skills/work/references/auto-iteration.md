# Auto-Iteration

4+ タスクまたは `all` 指定時に有効化される自動反復ロジック。
旧 `/ultrawork` の自律実行ループを `/work` に統合。

## Overview

```text
/work all (or /work 1-10)
    ↓
┌─────────────────────────────────────────────────────────────┐
│ Phase 0: 範囲確認（ユーザー承認必須）                        │
│  → スコープ対話 or 引数パース → 対象タスク特定              │
│  → 戦略表示: 「並列 3 ワーカー + 自動反復 (最大10回)」      │
└─────────────────────────────────────────────────────────────┘
    ↓ ユーザー承認後
┌─────────────────────────────────────────────────────────────┐
│ Phase 1: 初期化                                              │
│  1. 依存関係グラフ構築                                      │
│  2. 完了条件の設定                                          │
│  3. ワークログ初期化 → .claude/state/work.log.jsonl          │
│  4. ガードバイパス有効化 → work-active.json                  │
│  5. session.json に active_skill: "work" を設定             │
└─────────────────────────────────────────────────────────────┘
    ↓
┌─────────────────────────────────────────────────────────────┐
│ Iteration 1〜N: 自律実行ループ                              │
│  Step 1: 現状評価                                           │
│    - 未完了タスク特定                                       │
│    - 失敗履歴から学習                                       │
│    - 優先順位再計算                                         │
│                                                             │
│  Step 2: 並列実装（task-worker × N）                        │
│    - 独立タスクを並列実行                                   │
│    - 各ワーカーが自己完結（実装→ビルド→テスト）            │
│                                                             │
│  Step 3: 統合検証                                           │
│    - 全体ビルド実行                                         │
│    - テストスイート実行                                     │
│                                                             │
│  Step 3.5: /harness-review + 自己修正ループ                 │
│    - 全タスク完了時のみ                                     │
│    - APPROVE まで自動修正を繰り返す                         │
│    - REJECT/STOP は即停止                                   │
│                                                             │
│  Step 3.7: Auto-Refinement                                  │
│    - harness-review APPROVE 後のみ                          │
│    - /simplify で自動コード洗練                             │
│    - --deep-simplify 時は code-simplifier も実行            │
│    - --no-simplify 時はスキップ                             │
│                                                             │
│  Step 4: 判定                                               │
│    - APPROVE → 完了処理へ                                   │
│    - REQUEST CHANGES → 自己修正ループ                       │
│    - REJECT/STOP → 即停止 + 手動介入                        │
│    - 未完了あり → 次 iteration へ                           │
└─────────────────────────────────────────────────────────────┘
    ↓
┌─────────────────────────────────────────────────────────────┐
│ 完了処理                                                     │
│  0. review_status 確認（必須: "passed" のみ完了可）          │
│  1. work-active.json 削除                                   │
│  2. session.json から active_skill を削除                   │
│  3. 最終コミット                                            │
│  4. 完了レポート + Tip 表示                                 │
└─────────────────────────────────────────────────────────────┘
```

## Self-Learning Mechanism

各イテレーションで前回の失敗から学習し、同じ失敗を繰り返さない。

### 学習フロー

```text
Iteration 1:
  タスク A: 型エラー "User型が見つからない"
  → 失敗をワークログに記録
    ↓
Iteration 2:
  ワークログを読み込み:
  「前回 User 型が見つからなかった」
  → 戦略: "User 型の定義を先に確認してから実装"
  → タスク A: 成功
```

### 学習戦略パターン

| 失敗パターン | 次イテレーションの戦略 |
|-------------|----------------------|
| 型エラー | 関連する型定義を先に確認 |
| import エラー | パス構造を再確認 |
| テスト失敗 | テストケースを読んで期待値を理解 |
| ビルドエラー | 依存関係を確認、順序変更 |
| 3回連続同じエラー | 別アプローチを試行 |

## Worklog Format

`.claude/state/work.log.jsonl`:

```jsonl
{"ts":"2026-02-08T10:00:00Z","event":"start","range":"1-5","max_iterations":10}
{"ts":"2026-02-08T10:00:05Z","event":"iteration_start","iteration":1}
{"ts":"2026-02-08T10:00:30Z","event":"task_complete","task":"Create Header","status":"success"}
{"ts":"2026-02-08T10:00:55Z","event":"task_failed","task":"Create Footer","error":"Import not found"}
{"ts":"2026-02-08T10:01:25Z","event":"iteration_end","iteration":1,"completed":1,"failed":1}
{"ts":"2026-02-08T10:02:00Z","event":"task_complete","task":"Create Footer","learned_from":"iter 1"}
{"ts":"2026-02-08T10:05:00Z","event":"complete","iterations":3,"tasks_completed":5}
```

## Completion Conditions

以下の**全て**を満たしたとき完了:

1. 指定範囲の全タスクが `cc:done`
2. 全体ビルド成功
3. 全テスト通過
4. harness-review で APPROVE
5. `review_status === "passed"`

## 検証実行規則（全て実行、失敗で即停止）

| 順位 | 対象 | コマンド |
|------|------|---------|
| 1 | `./tests/validate-plugin.sh` | `bash ./tests/validate-plugin.sh` |
| 2 | `./scripts/ci/check-consistency.sh` | `bash ./scripts/ci/check-consistency.sh` |
| 3 | `package.json` の `test` script | `{pkg_mgr} test` |
| 4 | `package.json` の `lint` script | `{pkg_mgr} run lint` |
| 5 | `pytest.ini` / `pyproject.toml` | `pytest` |
| 6 | `Cargo.toml` | `cargo test` |
| 7 | `go.mod` | `go test ./...` |

## Resume from Worklog

```bash
# 前回の中断から再開
/work --resume latest

# 内部動作:
# 1. .claude/state/work.log.jsonl を読み込み
# 2. 最後の iteration_end を特定
# 3. 完了タスクをスキップして未完了から再開
# 4. 失敗履歴を学習データとして引き継ぎ
```

## Progress Display

```text
📊 /work Progress: Iteration 2/10

Range: Tasks 1-5
Completed: 2/5 tasks
Time elapsed: 2m 15s

├── Task 1: Create Header ✅ (iter 1, 25s)
├── Task 2: Create Footer ✅ (iter 2, 30s) [learned]
├── Task 3: Create Sidebar ⏳ In progress...
├── Task 4: Create Layout 🔜 Waiting
└── Task 5: Create Page 🔜 Waiting

Last iteration result:
├── Build: ✅ Pass
├── Tests: ⚠️ 14/15 pass
└── Review: ✅ No Critical/High
```
