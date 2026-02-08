# Parallel Full-Cycle Automation

> `/work --full` で「実装→セルフレビュー→改善→commit」のフルサイクルを並列自動化する機能のドキュメント。

## 概要

フェーズ 32 で追加された task-worker 統合により、Plans.md のタスクを並列で自動処理できるようになりました。

```
/work --full --parallel 3
```

このコマンド1つで以下が自動実行されます：

1. **Phase 1**: 依存グラフ構築 → task-worker 並列起動 → セルフレビュー
2. **Phase 2**: クロスレビュー（Codex利用可ならCodex 8並列、未設定時は通常レビューにフォールバック）
3. **Phase 3**: コンフリクト検出・解消 → 最終ビルド検証 → Conventional Commit
4. **Phase 4**: Deploy（オプション）

---

## クイックスタート

### 基本的な使い方

```bash
# シンプルなフルサイクル（並列1、デフォルト設定）
/work --full

# 3並列で実行
/work --full --parallel 3

# 完全分離モード（大規模タスク向け）
/work --full --parallel 5 --isolation worktree

# デプロイまで自動化
/work --full --deploy
```

### オプション一覧

| オプション | 説明 | デフォルト |
|-----------|------|-----------|
| `--full` | フルサイクル実行モード | false |
| `--parallel N` | 並列数指定 | 1 |
| `--isolation` | `lock` / `worktree` | lock |
| `--commit-strategy` | `task` / `phase` / `all` | task |
| `--deploy` | commit 後にデプロイ | false |
| `--max-iterations` | 改善ループ上限 | 3 |
| `--skip-cross-review` | Phase 2 スキップ | false |

---

## アーキテクチャ

### task-worker エージェント

各タスクは独立した `task-worker` エージェントで処理されます。

```
task-worker の内部フロー:
┌─────────────────────────────────────────────────────────┐
│  [入力: タスク説明 + 対象ファイル]                       │
│                    ↓                                    │
│  Step 1: 実装（impl スキルの知識を内包）                │
│                    ↓                                    │
│  Step 2: セルフレビュー（4観点）                        │
│  ├── 品質: 命名、構造、可読性                           │
│  ├── セキュリティ: 入力検証、機密情報                   │
│  ├── パフォーマンス: N+1、不要な再計算                  │
│  └── 互換性: 既存コードとの整合性                       │
│                    ↓                                    │
│  [問題あり？] → YES → Step 3（修正）→ ループ           │
│              └ NO → Step 4 へ                           │
│                    ↓                                    │
│  Step 4: ビルド検証                                     │
│                    ↓                                    │
│  Step 5: テスト実行                                     │
│                    ↓                                    │
│  [commit_ready を返す]                                  │
└─────────────────────────────────────────────────────────┘
```

### 入出力スキーマ

**入力**:
```json
{
  "task": "タスク説明（Plans.md から抽出）",
  "files": ["対象ファイルパス"] | "auto",
  "max_iterations": 3,
  "review_depth": "light" | "standard" | "strict"
}
```

**出力**:
```json
{
  "status": "commit_ready" | "needs_escalation" | "failed",
  "iterations": 2,
  "changes": [
    { "file": "src/foo.ts", "action": "created" | "modified" }
  ],
  "self_review": {
    "quality": { "grade": "A", "issues": [] },
    "security": { "grade": "A", "issues": [] },
    "performance": { "grade": "B", "issues": ["N+1クエリの可能性"] },
    "compatibility": { "grade": "A", "issues": [] }
  },
  "build_result": "pass" | "fail",
  "build_log": "エラーメッセージ（失敗時のみ）",
  "test_result": "pass" | "fail" | "skipped",
  "test_log": "失敗したテストの詳細（失敗時のみ）",
  "escalation_reason": null | "max_iterations_exceeded" | "build_failed_3x" | "test_failed_3x" | "review_failed_3x"
}
```

---

## 依存グラフと並列実行

### 依存グラフ構築

Plans.md のタスクを解析し、並列実行可能なグループを決定します。

```
判定ルール:
├── 同一ファイルを複数タスクが編集 → 直列実行
├── タスクAの出力がタスクBの入力 → A→B の順序
└── 互いに独立 → 並列実行可能
```

### 実行例

```
📋 Plans.md から 5 タスクを検出

依存関係分析:
├── [独立] Header 作成
├── [独立] Footer 作成
├── [独立] Sidebar 作成
├── [依存] Layout 作成 ← Header, Footer, Sidebar に依存
└── [依存] Page 作成 ← Layout に依存

実行計画:
🚀 並列グループ1: Header, Footer, Sidebar (同時実行)
   ↓
🔧 直列: Layout 作成
   ↓
🔧 直列: Page 作成
```

---

## ワークスペース分離

### --isolation=lock（デフォルト）

- 同一 worktree でファイルロックを使用
- 小〜中規模タスク向け
- ディスク容量を節約

### --isolation=worktree

- 各タスクに `git worktree add` でブランチ作成
- 完全な並列ビルド/テストが可能
- **pnpm 使用時は容量節約（+54MB/worktree）**

```bash
# pnpm + worktree の組み合わせ
/work --full --parallel 5 --isolation worktree
```

pnpm の shared store により、通常 2GB 以上かかる node_modules が約 54MB で済みます。

#### 前提条件

- Gitリポジトリが初期化されていること
- ベースブランチ（main/master）が存在すること
- `.worktrees`ディレクトリが`.gitignore`に追加されていること（自動追加される）

#### 後始末

Phase3完了後、worktreeブランチは自動的にマージされ、worktreeディレクトリは削除されます。

手動でクリーンアップする場合：

```bash
# 不要なworktreeを削除
git worktree prune

# 特定のworktreeブランチを削除
git branch -D worktree/task1
```

---

## commit_ready 基準

タスクが `commit_ready` を返すには以下を**全て**満たすこと：

1. ✅ セルフレビュー全観点で Critical/Major 指摘なし
2. ✅ ビルドコマンドが成功（exit code 0）
3. ✅ 該当テストが成功（または該当テストなし）
4. ✅ 既存テストの回帰なし
5. ✅ 品質ガードレール違反なし

---

## エスカレーション

task-worker が 3 回修正しても問題が解決しない場合、親に集約してユーザーに一括確認します。

```
⚠️ エスカレーション（2件）

タスクA: 型 'unknown' を 'User' に変換できません
  → 提案: User 型の定義を確認するか、型ガードを追加

タスクB: テスト 'should validate email' が失敗
  → 提案: 正規表現パターンを修正

どう対応しますか？
1. 提案を適用して続行
2. スキップして次へ
3. 手動で修正する
```

---

## Phase 2: クロスレビューのフォールバック動作

Phase2のクロスレビューは、Codex設定に応じて自動的にモードを切り替えます：

### Codexモード（推奨）

**条件**:
- `.claude-code-harness.config.yaml`に`review.mode: codex`が設定されている
- `review.codex.enabled: true`が設定されている

**動作**:
- Codex MCP経由で8つのエキスパートを並列レビュー
- セルフレビューでは見落とす問題を検出

### 通常レビューモード（フォールバック）

**条件**（以下のいずれか）:
- `.claude-code-harness.config.yaml`が存在しない
- `review.mode`が未設定または`default`
- `review.mode=codex`だが`review.codex.enabled=false`
- `review.codex.enabled`が未設定

**動作**:
- 既存の`review`スキルを実行
- セキュリティ/パフォーマンス/品質/アクセシビリティをチェック
- Codex未設定でもフルサイクルは正常に完了

**注意**: Codex未設定時でも`/work --full`は正常に動作します。Phase2が通常レビューにフォールバックするだけで、フローは継続します。

---

## VibeCoder 向けガイド

技術的な詳細がわからなくても使えます。

| やりたいこと | 言い方 |
|-------------|--------|
| 全タスクを一気にやって | `/work --full` |
| 速く終わらせて | `/work --full --parallel 3` |
| デプロイまで全自動で | `/work --full --deploy` |
| レビューはスキップして | `/work --full --skip-cross-review` |

---

## トラブルシューティング

### Q: 並列実行中にエラーが発生した

A: 成功したタスクの結果は保持されます。失敗タスクのみ再実行するか、エラー内容を確認して修正してください。

### Q: worktree がクリーンアップされない

A: `git worktree prune` で不要な worktree を削除できます。

### Q: Codex レビューで Critical が検出された

A: Phase 1 に自動的に戻り、該当タスクが修正されます。3 回修正しても解決しない場合はエスカレーションされます。

---

## 関連ドキュメント

- [task-worker エージェント](../agents/task-worker.md)
- [/work スキル](../skills/work/SKILL.md)
- [Codex 並列レビュー](../skills/codex-review/references/codex-parallel-review.md)
