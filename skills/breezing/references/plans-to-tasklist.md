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

### 5. plans_md_mapping 生成

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
4. plans_md_mapping を新タスク ID で更新
5. 依存関係を再構築
6. 実装サイクルから再開
```

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

### 大量タスク (10+)

```
1. 最初に独立タスクのみ TaskCreate (batch 1)
2. batch 1 が半分完了 → 次の batch を TaskCreate
3. メモリ効率のため、一度に全タスクを登録しない
```
