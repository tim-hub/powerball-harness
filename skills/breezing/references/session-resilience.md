# Session Resilience

Agent Teams のセッション再開非互換への対策と、`/breezing 続きやって` による再開メカニズム。

## 問題

Agent Teams は experimental 機能であり、以下の制約がある:

```
制約: /resume で Teammates は復元されない
→ セッション再開時に Team が消滅
→ 進捗が失われる
```

## 解決策: 二層永続化

```
Agent Teams TaskList (~/.claude/tasks/{team_name}/)
  → 実行中タスク状態の SSOT (Source of Truth)
  → Agent Teams が管理、Harness は読むだけ
  → ディスク上にファイルとして永続化（セッション切れでも消えない）
  → "Clean up the team" 実行時のみ削除

breezing-active.json (.claude/state/)
  → Team 構成メタデータのみ（TaskList にない情報）
  → session_id, options, team 構成, Plans.md との紐付け
  → タスク状態は持たない（TaskList に委譲）
```

## breezing-active.json スキーマ (v2)

**ファイルパス**: `.claude/state/breezing-active.json`

**設計原則**: タスク状態は Agent Teams TaskList に一元化。breezing-active.json はメタデータのみ保持。

```json
{
  "session_id": "breezing-20260206-0300",
  "started_at": "2026-02-06T03:00:00Z",
  "team_name": "breezing-auth-feature",
  "task_range": "認証機能からユーザー管理まで",
  "impl_mode": "standard",
  "plans_md_mapping": {
    "task-1": "4.1",
    "task-2": "4.2",
    "task-3": "4.3"
  },
  "options": {
    "codex_review": false,
    "parallel": 2
  },
  "team": {
    "implementer_count": 2,
    "reviewer_count": 1,
    "model": "sonnet"
  },
  "review": {
    "retake_count": 0,
    "max_retakes": 3
  }
}
```

### フィールド説明

| フィールド | 用途 | 更新タイミング |
|-----------|------|--------------|
| `session_id` | セッション識別子 | 準備ステージで設定 |
| `team_name` | Agent Teams Team 名 | 準備ステージで設定 |
| `task_range` | ユーザー指定の範囲テキスト | 準備ステージで設定 |
| `impl_mode` | 実装モード (`"standard"` or `"codex"`)。Compaction 復元キー | Step 0 で即時書き込み |
| `plans_md_mapping` | TaskList ID → Plans.md セクション番号 | 準備ステージで設定 |
| `options` | 実行オプション | 準備ステージで設定 |
| `team` | Team 構成情報 | 準備ステージで設定 |
| `review.retake_count` | リテイク回数 | リテイク発生時にインクリメント |
| `review.max_retakes` | リテイク上限 | 固定値 (デフォルト 3) |

### Progressive Batch 状態の永続化

タスク数 > 8 で Progressive Batching が有効な場合、`batching` フィールドを追加:

```json
{
  "batching": {
    "enabled": true,
    "total_tasks": 15,
    "current_batch": 2,
    "batches": [
      {"batch": 1, "task_ids": ["task-1", "task-2", "task-3"], "status": "completed"},
      {"batch": 2, "task_ids": ["task-4", "task-5"], "status": "in_progress"},
      {"batch": 3, "task_ids": [], "status": "pending"}
    ]
  }
}
```

「続きやって」での再開時:
1. `batching.current_batch` で現在のバッチを特定
2. `status: "in_progress"` のバッチの未完了タスクを再登録
3. `status: "pending"` のバッチは元の Plans.md セクションから復元

### v1 → v2 の変更点

| v1 | v2 | 理由 |
|----|-----|------|
| `status` フィールド | 削除 | TaskList の状態で判断可能 |
| `tasks[]` 配列 | 削除 | Agent Teams TaskList に一元化 |
| `review.status` | 削除 | Lead が判断で管理 |
| `review.last_findings` | 削除 | Reviewer が SendMessage で伝達 |
| `review.history` | 削除 | Agent Trace で自動記録 |
| なし | `team_name` 追加 | TaskList 永続化パスに必要 |
| なし | `plans_md_mapping` 追加 | TaskList ↔ Plans.md の紐付け |
| なし | `impl_mode` 追加 | Compaction 復元キー (`"standard"` or `"codex"`) |

## Plans.md の更新タイミング

**重要**: Plans.md は**完了時のみ**更新する。途中の `cc:WIP` 更新は不要。

```
準備ステージ:
  Plans.md cc:TODO → TaskCreate(pending)  ← 一方向

実装中:
  Implementer が TaskUpdate(in_progress) → Plans.md 更新しない
  Implementer が TaskUpdate(completed) → Plans.md 更新しない

完了ステージ (全タスク完了 + APPROVE 後):
  Lead が plans_md_mapping を参照
  → Plans.md の対応セクションを cc:TODO → cc:done に一括更新
```

### 理由

- 途中更新は同期ずれリスクを生む
- TaskList が SSOT なので Plans.md の途中状態は不要
- 完了時の一括更新で Plans.md の整合性を保証

## Compaction 復元フロー（同一セッション内）

Compaction はセッション再開とは異なり、**同一セッション内でコンテキストが圧縮される現象**。
スキルの指示が失われ、Lead が直接実装を始めてしまうリスクがある。

### 問題

```
Compaction 前: Lead はスキルのフロー全体を記憶 → Team spawn → Implementer
Compaction 後: Lead は「タスクを完了すべき」しか残っていない → 直接実装
```

### 対策: breezing-active.json による復元

```text
Compaction 発生
    ↓
Lead: breezing-active.json を Read
    ↓
impl_mode を確認:
  ├── "codex" → breezing-codex モード → Team 再構築 (codex-implementer spawn)
  ├── "standard" → 通常 breezing → Team 再構築 (task-worker spawn)
  └── なし or 未設定 → 通常 breezing として扱う（後方互換）
    ↓
現在のスキルとモード不一致チェック:
  ├── /breezing --codex 実行中に impl_mode="standard" → ユーザーに確認
  └── /breezing 実行中に impl_mode="codex" → ユーザーに確認
    ↓
team_name で既存 Team を確認:
  ├── Team 存在 → TaskList で未完了タスクを確認 → サイクル再開
  └── Team 消失 → 新 Team 作成 → 未完了タスクを再登録 → サイクル再開
```

### 実装側の責務

breezing-active.json への `impl_mode` 書き込みは**スキル実行の最初のステップ**で行う。
環境チェックよりも前に書き込むことで、準備ステージ中の compaction にも対応できる。

```jsonc
// Step 0 で即時書き込み
{
  "impl_mode": "codex",  // breezing-codex の場合
  "session_id": "...",
  "started_at": "..."
}
```

### breezing（通常版）への適用

通常の breezing でも同様の問題が発生しうる。breezing-active.json に以下を追加推奨:

```jsonc
{
  "impl_mode": "standard",  // 通常 breezing
  // ... 既存フィールド
}
```

## 「続きやって」再開フロー

```text
/breezing 続きやって
    ↓
┌──────────────────────────────────────────────────────┐
│ Step 1: breezing-active.json の存在確認               │
│  → なし: 「前回のセッションが見つかりません」で停止    │
│  → あり: Step 2 へ                                    │
└──────────────────────────────────────────────────────┘
    ↓
┌──────────────────────────────────────────────────────┐
│ Step 2: 状態の読み込みと表示                          │
│                                                       │
│  team_name → ~/.claude/tasks/{team_name}/ を確認      │
│  plans_md_mapping → Plans.md の対応セクションを表示    │
│                                                       │
│  前回のセッション: breezing-20260206-0300             │
│  開始: 2026-02-06 03:00                               │
│  Team: breezing-auth-feature                          │
│                                                       │
│  TaskList 状態:                                       │
│  ✅ task-1 (4.1) ログイン機能の実装                   │
│  ⏸ task-2 (4.2) 認証ミドルウェアの作成 (中断)        │
│  ⏳ task-3 (4.3) セッション管理                       │
│                                                       │
│  続行しますか？                                       │
└──────────────────────────────────────────────────────┘
    ↓ ユーザー承認
┌──────────────────────────────────────────────────────┐
│ Step 3: 新 Team 作成 (Phase A として実行)              │
│  1. 新しい Agent Team を作成                          │
│  2. 未完了タスクのみ TaskCreate で再登録               │
│     → in_progress だったタスクは pending にリセット    │
│  3. plans_md_mapping を新タスク ID で更新              │
│  4. Implementer spawn (前回と同数)                    │
│  5. Reviewer spawn                                    │
│  6. delegate mode ON → Phase B へ遷移                │
│  7. 実装サイクルから再開                              │
└──────────────────────────────────────────────────────┘
```

### TaskList 永続化パスについて

```
Agent Teams TaskList のファイル保存先:
  ~/.claude/tasks/{team_name}/

例:
  ~/.claude/tasks/breezing-auth-feature/
  ├── task-1.json
  ├── task-2.json
  └── task-3.json

注意: "Clean up the team" を実行するとこのディレクトリが削除される。
Breezing は完了ステージで自動的にクリーンアップを実行する。
中断時は plans_md_mapping で復元可能。
```

### in_progress タスクの扱い

```
前回 in_progress だったタスク:
  → 部分的な実装が残っている可能性
  → pending にリセットして再実装
  → Implementer は既存コードを確認してから作業開始
  → 前回の変更が git にコミットされていれば、差分から継続
```

## 異常状態への対応

### breezing-active.json が壊れている場合

```
1. JSON パースエラー → ユーザーに報告
2. 必須フィールド欠落 → ユーザーに報告
3. plans_md_mapping に対応する Plans.md セクションが見つからない
   → Plans.md を信頼源としてマッピングを再構築
```

### TaskList ディレクトリが消えている場合

```
1. ~/.claude/tasks/{team_name}/ が存在しない
   → plans_md_mapping + Plans.md から未完了タスクを特定
   → 新 Team 作成時に再登録
2. 一部のタスクファイルが破損
   → 破損タスクを pending として再登録
```

### Plans.md が変更されている場合

```
1. plans_md_mapping のタスクと Plans.md を照合
2. Plans.md で新タスクが追加 → 追加タスクを含めるか確認
3. Plans.md でタスクが削除 → 削除されたタスクをスキップ
4. Plans.md のマーカーが cc:done に変更 → 対応タスクをスキップ
```

## クリーンアップ

### 正常完了時

```
完了ステージ実行:
  1. Plans.md の対応タスクを cc:done に更新
  2. breezing-active.json 削除
  3. breezing-session-roles.json 削除
  4. breezing-role-*.json 削除
  5. Team クリーンアップ (Agent Teams)
```

### 手動クリーンアップ

```
異常終了や中断が繰り返される場合:
  rm .claude/state/breezing-active.json
  rm .claude/state/breezing-session-roles.json
  rm .claude/state/breezing-role-*.json
  → 次回 /breezing 実行時に新規セッションとして開始
```

## ultrawork との比較

| 項目 | ultrawork | breezing |
|------|-----------|---------|
| 状態ファイル | ultrawork-active.json | breezing-active.json (メタデータのみ) |
| タスク状態の SSOT | ultrawork-active.json | Agent Teams TaskList |
| Plans.md 更新 | 随時 (cc:WIP, cc:done) | 完了時のみ (cc:done) |
| 復元対象 | タスク進捗のみ | タスク進捗 + Team 構成 + Plans.md mapping |
| Team 復元 | N/A (Task tool) | 新 Team 再作成 |
| 再開コマンド | 自動検出 | `/breezing 続きやって` |
