# Plans.md → TaskList 変換

Plans.md のタスクを Agent Teams の共有タスクリスト (TaskCreate/TaskUpdate/TaskList/TaskGet) に変換するロジック。

## 概要

```
Plans.md (cc:TODO マーカー付き)
    ↓ 準備ステージで変換
Agent Teams TaskList (TaskCreate) ← SSOT
    ↓ Implementer が自律消化
TaskUpdate (completed)
    ↓ 完了ステージで一括反映
Plans.md (cc:done) に更新
```

**重要**: Agent Teams TaskList が実行中のタスク状態の SSOT。
Plans.md は完了時のみ更新する（途中の cc:WIP 更新は不要）。

## タスク粒度バリデーション（TaskCreate 前の必須チェック）

TaskCreate 登録の**前に**、Plans.md の各タスクを以下の観点でバリデーションする。
問題が検出された場合はユーザーに修正提案を表示し、承認を得てから TaskCreate に進む。

### バリデーション観点

| # | チェック項目 | 検出条件 | 重要度 | アクション |
|---|---|---|---|---|
| V1 | **スコープ過大** | owns 推定ファイル数 > 10、またはサブタスクが 5+ 項目 | warning | 分割を提案 |
| V2 | **記述曖昧** | 具体的ファイルパス/コンポーネント名/API 名がゼロ | warning | 具体化を要求 |
| V3 | **owns 重複過多** | 2タスクが owns の 50% 以上を共有 | warning | マージまたは順次化を提案 |
| V4 | **抽象キーワード** | 「改善」「リファクタリング」「最適化」等のみで受入条件なし | warning | 受入条件の明示を要求 |
| V5 | **依存関係未宣言** | 同一ファイルを触るタスク間に blockedBy がない | error | addBlockedBy を自動付与 |

### バリデーション実行フロー

```text
Plans.md から対象タスクを抽出
    ↓
各タスクに V1〜V5 を適用
    ↓
┌── 問題なし → TaskCreate 登録へ
└── 問題あり → バリデーションレポートをユーザーに表示
                ↓
              ユーザー判断:
                ├── 修正する → Plans.md 修正後に再バリデーション
                ├── そのまま続行 → warning は無視して TaskCreate 登録
                └── 中止 → breezing-active.json 削除して停止
```

### バリデーションレポート例

```text
🏇 Breezing - タスク品質チェック

⚠️ 2 件の警告が見つかりました:

1. [V1] タスク 4.1「UI全体のリファクタリング」
   → スコープ過大: 推定 15 ファイルに影響
   → 提案: 「ヘッダーのリファクタリング」「サイドバーのリファクタリング」等に分割

2. [V2] タスク 4.3「パフォーマンス改善」
   → 記述曖昧: 具体的な対象ファイル/メトリクスが不明
   → 提案: 「src/db/users.ts の N+1 クエリ解消」等に具体化

✅ タスク 4.2「認証ミドルウェアの作成」— 問題なし

修正しますか？ (修正 / そのまま続行 / 中止)
```

### V5 の自動修復

V5（依存関係未宣言）は Lead が**自動的に addBlockedBy を付与**する:

```
検出:
  タスク A: owns: src/auth/login.ts
  タスク B: owns: src/auth/login.ts (blockedBy 未設定)

自動修復:
  → B に addBlockedBy: [A] を設定
  → バリデーションレポートに「自動付与」として記載
```

V5 は error レベルだが自動修復可能なため、ユーザー承認なしで適用する。

## 変換ルール

### 1. タスク粒度

Plans.md の **1 タスク = 1 TaskCreate エントリ**。

```markdown
<!-- Plans.md -->
## Phase 4: 認証機能

### 4.1 ログイン機能の実装 <!-- cc:TODO -->
- ログインフォーム作成
- バリデーション追加
- API エンドポイント接続

### 4.2 認証ミドルウェアの作成 <!-- cc:TODO -->
- JWT 検証
- ルートガード実装
```

↓ 変換

```
TaskCreate:
  subject: "4.1 ログイン機能の実装"
  description: |
    - ログインフォーム作成
    - バリデーション追加
    - API エンドポイント接続
    owns: src/components/LoginForm.tsx, src/app/api/auth/login/route.ts
  activeForm: "ログイン機能を実装中"

TaskCreate:
  subject: "4.2 認証ミドルウェアの作成"
  description: |
    - JWT 検証
    - ルートガード実装
    owns: src/middleware.ts, src/lib/auth.ts
  activeForm: "認証ミドルウェアを作成中"
```

### 2. owns: アノテーション

各タスクが**編集するファイル**を `owns:` で明示。

```
owns: の決定ロジック:

1. Plans.md にファイルパスが記載されていれば使用
2. タスク説明からキーワード抽出 → 既存ファイル検索
   例: "ログインフォーム" → Glob("**/Login*.tsx")
3. 関連ディレクトリの推定
   例: "認証" → src/auth/, src/lib/auth/
4. 特定できない場合 → owns: auto (Implementer が自動判定)
```

### 3. 依存関係 (addBlockedBy)

**同一ファイルを触るタスクは順次化**する。

```
タスク A: owns: src/auth/login.ts
タスク B: owns: src/auth/login.ts, src/auth/session.ts
タスク C: owns: src/components/Header.tsx

→ B に addBlockedBy: [A] を設定 (login.ts が競合)
→ C は独立 (競合なし)
```

#### 競合検出アルゴリズム

```
1. 全タスクの owns: を抽出
2. ファイル → タスク のマッピングを構築
3. 同一ファイルに複数タスクがある場合:
   - Plans.md の順序で addBlockedBy を設定
   - 先のタスクが完了するまで後のタスクは pending
```

### 4. activeForm 生成

TaskCreate の `activeForm` は**タスク名を present continuous 形式に変換**:

| subject | activeForm |
|---------|-----------|
| ログイン機能の実装 | ログイン機能を実装中 |
| 認証ミドルウェアの作成 | 認証ミドルウェアを作成中 |
| テストの追加 | テストを追加中 |
| バグ修正: ログアウト | ログアウトバグを修正中 |

### 5. Spec Driven Development 連携（[feature:tdd] マーカー）

Plans.md に `/plan-with-agent` が付与した `[feature:tdd]` マーカーがある場合、
テスト仕様を TaskCreate に統合する。

#### 検出と変換

```markdown
<!-- Plans.md -->
### 4.1 ログイン機能の実装 <!-- cc:TODO --> <!-- feature:tdd -->
- ログインフォーム作成
- バリデーション追加
- API エンドポイント接続

**テストケース設計:**
- 正常ログイン → トークン返却
- 無効な認証情報 → 401 エラー
- レート制限 → 429 エラー
```

↓ 変換

```
TaskCreate:
  subject: "4.1a ログイン機能のテスト作成"
  description: |
    以下のテストケースを先行実装:
    - 正常ログイン → トークン返却
    - 無効な認証情報 → 401 エラー
    - レート制限 → 429 エラー
    owns: src/__tests__/auth/login.test.ts
  activeForm: "ログイン機能のテストを作成中"

TaskCreate:
  subject: "4.1b ログイン機能の実装"
  description: |
    テストが先行作成済み。テストを通す実装を行う。
    - ログインフォーム作成
    - バリデーション追加
    - API エンドポイント接続
    owns: src/components/LoginForm.tsx, src/app/api/auth/login/route.ts
  addBlockedBy: ["4.1a"]
  activeForm: "ログイン機能を実装中"
```

#### 変換ルール

| 条件 | 変換 |
|---|---|
| `[feature:tdd]` + テストケース設計あり | テスト作成タスク(a) + 実装タスク(b)に分割。b は a に blockedBy |
| `[feature:tdd]` + テストケース設計なし | Implementer spawn prompt に「テスト先行」指示を追加（分割はしない） |
| `[feature:tdd]` なし | 通常変換（分割なし） |

#### Reviewer への影響

`[feature:tdd]` タスクのレビュー時、Reviewer はスペック充足を追加チェック:

```
通常レビュー観点 (セキュリティ/パフォーマンス/品質/互換性)
  +
スペック充足チェック:
  - テストケース設計の全シナリオがテストコードに反映されているか
  - テストが meaningful か（アサーションが適切か）
  - 実装がテストを通しているか
```

### 6. plans_md_mapping 生成

TaskCreate と同時に breezing-active.json の `plans_md_mapping` を生成:

```json
{
  "plans_md_mapping": {
    "task-1": "4.1",
    "task-2": "4.2",
    "task-3": "4.3"
  }
}
```

これにより完了ステージで TaskList ID → Plans.md セクション番号を逆引き可能。

#### `[feature:tdd]` 分割時の 1:N マッピング

`[feature:tdd]` でタスクが分割された場合、1 つの Plans.md セクションに複数のタスク ID がマッピングされる:

```json
{
  "plans_md_mapping": {
    "task-1a": "4.1",
    "task-1b": "4.1",
    "task-2": "4.2"
  }
}
```

**完了ステージの注意**: 同一セクション番号にマッピングされた全タスクが `completed` の場合にのみ `cc:done` に更新する。
一部のみ完了した状態で `cc:done` にしないこと。

## Plans.md マーカーとの対応

| Plans.md マーカー | TaskList status | 同期方向 |
|-------------------|----------------|---------|
| `cc:TODO` | `pending` | Plans.md → TaskList (準備ステージ) |
| (更新しない) | `in_progress` | 同期なし |
| `cc:done` | `completed` | TaskList → Plans.md (完了ステージのみ) |

### 同期フロー（簡素化 v2）

```
準備ステージ:
  Plans.md cc:TODO → TaskCreate(pending)  ← 一方向変換

実装中:
  Implementer が TaskUpdate(in_progress) → Plans.md 更新しない
  Implementer が TaskUpdate(completed) → Plans.md 更新しない

完了ステージ (全タスク完了 + APPROVE):
  Lead が plans_md_mapping を参照
  → 全 completed タスクの Plans.md セクションを cc:done に一括更新
```

**v1 からの変更**: 途中の cc:WIP 同期を廃止。TaskList が SSOT なので Plans.md の途中状態管理は不要。

## グループ化戦略

大量のタスクがある場合、Lead は以下の戦略でグループ化:

### 機能グループ

```
グループ A: 認証機能 (タスク 4.1, 4.2, 4.3)
  → Implementer #1 が担当
  → owns: src/auth/*, src/middleware.ts

グループ B: ユーザー管理 (タスク 5.1, 5.2)
  → Implementer #2 が担当
  → owns: src/users/*, src/app/api/users/*
```

### 依存グラフに基づく並列化

```
独立タスク群 (並列実行可能):
  ├── タスク A (owns: src/auth/*)
  ├── タスク C (owns: src/components/*)
  └── タスク E (owns: src/utils/*)

順次タスク群 (blockedBy 設定):
  └── タスク B (owns: src/auth/*) → blockedBy: [A]
  └── タスク D (owns: src/components/*) → blockedBy: [C]
```

## 「続きやって」の処理

`/breezing 続きやって` 実行時:

```
1. breezing-active.json の plans_md_mapping を読み込み
2. ~/.claude/tasks/{team_name}/ からタスク状態を確認
   → 存在しない場合: Plans.md から未完了タスクを特定
3. 未完了タスクのみ新 Team の TaskCreate で再登録
4. plans_md_mapping のステール ID 調整（下記参照）
5. 依存関係を再構築
6. 実装サイクルから再開
```

### ステール ID 調整ルール

再開時に新しいタスク ID が発行されるため、plans_md_mapping の整合性を保つ:

```text
調整フロー:
1. 旧 mapping から各セクション番号の旧タスク ID を取得
2. completed 状態の旧 ID → そのまま保持（完了済みなので再登録不要）
3. 未完了の旧 ID → 新 Team の TaskCreate で再登録、新 ID を取得
4. mapping を更新: 旧 ID (未完了) → 新 ID に置換
5. 完了判定は active な ID セット（completed + 新 ID）で評価

例:
  旧: {"task-1a": "4.1", "task-1b": "4.1", "task-2": "4.2"}
  task-1a: completed, task-1b: 未完了, task-2: 未完了
  ↓
  新: {"task-1a": "4.1", "task-5": "4.1", "task-6": "4.2"}
  (task-1a は完了済みなので保持、task-1b は task-5 に再登録)
```

**完了判定**: セクション 4.1 の cc:done 判定は `task-1a` (completed) + `task-5` (active) の両方が completed であること。
ステール ID (`task-1b`) は mapping から削除し、判定対象に含めない。

## エッジケース

### Plans.md にファイルパスがない場合

```
タスク: "UIの改善"
→ owns: auto
→ Implementer が Glob/Grep で対象ファイルを自動検出
→ 検出したファイルを Lead に SendMessage で報告
→ Lead が競合をチェック
```

### タスクが曖昧な場合

```
タスク: "パフォーマンス改善"
→ Lead がユーザーに確認を求める（準備ステージで）
→ 具体的なファイル/方法が特定されるまで承認しない
```

### 大量タスク: Progressive Batch 戦略

タスク総数 > 8 の場合、Progressive Batching を有効化する。
TaskList が肥大化するのを防ぎ、Lead のコンテキスト窓を保全する。

#### 有効化条件

| タスク数 | 戦略 |
|---|---|
| ≤ 8 | 全タスクを一括 TaskCreate |
| 9〜20 | Progressive Batch (2-3 バッチ) |
| 21+ | Progressive Batch (4+ バッチ) + ユーザーに分割実行を推奨 |

#### バッチ構成ルール

```text
Batch 1（初期バッチ）:
  → blockedBy が空のタスク群（独立フロンティア）を全て登録
  → 最大 8 タスクまで

Batch 2（次バッチ）: Batch 1 の 60% 完了時に登録
  → Batch 1 のタスクに blockedBy で依存していたタスク群
  → + 追加の独立タスク（Batch 1 に含まれなかった分）
  → 最大 8 タスクまで

Batch N: 前バッチの 60% 完了時に登録（以降同様）
```

#### 60% 完了のトラッキング

```text
TaskCompleted Hook (Lead 側で発火)
  → breezing-timeline.jsonl に記録
  → Lead が TaskList で completed 数をカウント
  → completed / batch_total ≥ 0.6 → 次バッチ登録
```

#### breezing-active.json への記録

```json
{
  "batching": {
    "enabled": true,
    "total_tasks": 15,
    "current_batch": 2,
    "batches": [
      {"batch": 1, "task_ids": ["task-1", "task-2", "task-3", "task-4"], "status": "completed"},
      {"batch": 2, "task_ids": ["task-5", "task-6", "task-7"], "status": "in_progress"},
      {"batch": 3, "task_ids": [], "status": "pending"}
    ]
  }
}
```

#### 注意事項

- Implementer はバッチの存在を意識しない（TaskList の pending タスクを順次消化するだけ）
- バッチ境界でのレビュー判断は Lead の裁量（バッチ完了ごとに部分レビューを推奨）
- 「続きやって」での再開時は、batching メタデータから未登録バッチを復元
