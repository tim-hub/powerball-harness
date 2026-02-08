# Execution Flow

Agent Teams を活用した `/breezing` の実行フロー。Lead は状況に応じてステージ間を柔軟に判断する。

## フロー全体図

```text
/breezing 認証機能からユーザー管理まで完了して
    ↓
┌─────────────────────────────────────────────────────────────┐
│ Step 0: breezing-active.json 即時書き込み (impl_mode: standard)│
│  → Compaction 耐性の確保（環境チェックより前）               │
└─────────────────────────────────────────────────────────────┘
    ↓
┌─────────────────────────────────────────────────────────────┐
│ 準備（必須、最初に 1 回）                                    │
│  環境チェック → ユーザー承認 → Team 初期化 → ロール登録     │
│  ※環境チェック失敗時は breezing-active.json を削除           │
└─────────────────────────────────────────────────────────────┘
    ↓
┌─────────────────────────────────────────────────────────────┐
│ 実装・レビューサイクル（Lead の判断で柔軟に運用）            │
│                                                              │
│  ┌─────────────┐   ┌─────────────┐   ┌─────────────┐       │
│  │ 実装         │ ↔ │ レビュー     │ ↔ │ リテイク     │       │
│  │ Implementer  │   │ Reviewer     │   │ Impl↔Rev    │       │
│  │ 自律タスク消化│   │ 部分/全体    │   │ 直接対話可   │       │
│  └─────────────┘   └─────────────┘   └─────────────┘       │
│                                                              │
│  Lead はこのサイクルを監視し、状況に応じて:                   │
│  ・半分完了 → 部分レビューを指示                             │
│  ・軽微な問題 → Reviewer↔Implementer 直接対話で解決          │
│  ・重大な問題 → タスク分解して修正タスク登録                 │
│  ・3回リテイク超過 → ユーザーにエスカレーション              │
└─────────────────────────────────────────────────────────────┘
    ↓ 全タスク完了 + APPROVE
┌─────────────────────────────────────────────────────────────┐
│ 完了                                                         │
│  統合検証 → Plans.md 更新 → git commit → メトリクスレポート  │
└─────────────────────────────────────────────────────────────┘
```

## 準備ステージ

### 0. breezing-active.json 即時書き込み（最優先）

**環境チェックよりも前に実行する。** Compaction 対策として、モード情報を永続化する。

```jsonc
// .claude/state/breezing-active.json に即時書き込み
{
  "session_id": "breezing-{timestamp}",
  "started_at": "{ISO8601}",
  "impl_mode": "standard",  // ← この値が compaction 後の復元キー
  "task_range": "{ユーザー指定の範囲}"
  // 残りのフィールド (team_name, plans_md_mapping 等) は Step 3 で追記
}
```

**`impl_mode` フィールド**:

| 値 | 設定者 | 意味 | ランタイムガード |
|---|--------|------|-----------------|
| `"standard"` | `/breezing` | task-worker が直接実装 | なし（通常の breezing role guard のみ） |
| `"codex"` | `/breezing --codex` | Codex MCP 経由で実装 | pretooluse-guard が Write/Edit/Bash を制限 |

**なぜ最初に書くか**: Compaction が準備ステージ中に発生しても、`impl_mode` が永続化されていれば復元できる。

**早期中断時のクリーンアップ**: Step 0 で書き込んだ breezing-active.json を削除してから停止する。部分的なファイルが残ると「続きやって」で誤復元されるリスクがあるため。

```text
以下のいずれかで中断 → breezing-active.json 削除 → 停止:
  ・環境チェック失敗（Agent Teams 未有効）
  ・ユーザーが範囲確認で拒否/キャンセル
  ・Team 初期化の失敗
```

### 1. 環境チェック

```bash
# Agent Teams 有効化チェック
# CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1 が必要
```

未設定時のメッセージ:

```text
⚠️ Agent Teams が有効化されていません。

以下を settings.json に追加してください:
{
  "env": {
    "CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS": "1"
  }
}

Agent Teams なしで実行する場合は `/work all` を使用してください。
```

### 2. 範囲確認（ユーザー承認必須）

| 指定パターン | 解釈 |
|-------------|------|
| `認証機能からユーザー管理まで` | 「認証」〜「ユーザー管理」を含むタスク |
| `ログイン機能を終わらせて` | 「ログイン」を含む全タスク |
| `Header, Footer, Sidebar` | 列挙されたキーワードを含むタスク |
| `全部やって` | Plans.md の全未完了タスク |
| `続きやって` | breezing-active.json から未完了タスクを復元 |

```text
🏇 Breezing - 範囲を確認させてください

指定: 「認証機能からユーザー管理まで」

対象タスク:
├── 3. ログイン機能の実装 (cc:TODO)
├── 4. 認証ミドルウェアの作成 (cc:TODO)
└── 5. セッション管理 (cc:TODO)

Team 構成:
├── Lead: delegate mode (調整専念)
├── Implementer: 2 個 (独立タスク数に基づく)
└── Reviewer: 1 個

計 3 タスクを Implementer 2 並列で完走します。

これで合っていますか？
```

### 3. Team 初期化

1. breezing-active.json 作成（メタデータのみ — タスク状態は Agent Teams TaskList に委譲）
2. delegate mode ON → Lead は指揮専念
3. Plans.md タスクを TaskCreate で共有タスクリストに登録
   - owns: アノテーション付与
   - addBlockedBy で依存関係設定
4. Implementer Teammates spawn (N 個)
   - `subagent_type: "claude-code-harness:task-worker"` で spawn
   - エージェント定義の `memory: project` により永続メモリが自動注入
   - spawn prompt でロールマーカーファイル Write を指示
5. Reviewer Teammate spawn (1 個)
   - `subagent_type: "claude-code-harness:code-reviewer"` で spawn
   - エージェント定義の `memory: project` により永続メモリが自動注入
   - spawn prompt でロールマーカーファイル Write を指示
6. (--codex-review) Codex MCP レビュー設定

詳細: team-composition.md 参照

### breezing-active.json スキーマ (v2)

**ファイル**: `.claude/state/breezing-active.json`

タスク状態は Agent Teams TaskList (`~/.claude/tasks/`) に一元化。
breezing-active.json はメタデータのみを保持。

```json
{
  "session_id": "breezing-20260206-0300",
  "started_at": "2026-02-06T03:00:00Z",
  "impl_mode": "standard",
  "team_name": "breezing-auth-feature",
  "task_range": "認証機能からユーザー管理まで",
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

### Implementer 数の自動決定

```
独立タスク数 = 依存関係なしで並列実行可能なタスク数

Implementer 数 = min(独立タスク数, --parallel N, 3)

デフォルト上限: 3 (トークンコスト抑制)
```

### TaskCreate 登録ルール

Plans.md のタスクを Agent Teams の共有タスクリスト (TaskCreate) に変換する際:

1. **タスク粒度**: Plans.md の 1 タスク = 1 TaskCreate エントリ
2. **owns: アノテーション**: 各タスクが触るファイルを description に記載
3. **依存関係**: 同一ファイルを触るタスクは `addBlockedBy` で順次化
4. **activeForm**: 進捗表示用の present continuous 形式

詳細: plans-to-tasklist.md 参照

## 実装・レビューサイクル

### Lead の運用ガイドライン

Lead はサイクル内で以下を**自律的に判断**する:

| 状況 | Lead の判断 |
|------|------------|
| 全タスク未着手 | 実装を開始させる |
| 半数のタスクが完了 | 部分レビューを Reviewer に指示可能 |
| 全タスク完了 | 全体レビューを Reviewer に指示 |
| Reviewer から軽微な質問 | Reviewer↔Implementer 直接対話を許可 |
| REQUEST CHANGES | findings を修正タスクに分解、Implementer に指示 |
| 3回リテイク超過 | ユーザーにエスカレーション |

**重要**: Lead は「この順番で進めなければならない」という制約はない。
状況を見て最適な判断をすること。

### Implementer の自律ループ

各 Implementer は以下を独立して繰り返す:

```
1. TaskList で pending かつ blockedBy が空のタスクを検索
2. 最も ID が小さいタスクを self-claim (TaskUpdate → in_progress)
3. task-worker フロー実行:
   - 実装 → セルフレビュー4観点 → ビルド → テスト
4. 成功 → TaskUpdate(completed)
5. 失敗 (3回) → Lead にエスカレーション (SendMessage)
6. 残りタスクあり → Step 1 へ
7. 残りタスクなし → Lead に完了報告 (SendMessage)
```

### ファイル競合回避

```
Lead が準備ステージで検出:
  タスク A: src/auth/login.ts を編集
  タスク B: src/auth/login.ts を編集
  → B に addBlockedBy: [A] を設定

結果: A が完了するまで B は pending のまま
```

### レビューのタイミング

Lead は以下のタイミングで Reviewer にレビューを指示できる:

```
パターン A: 全完了後レビュー（デフォルト）
  全 Implementer 完了 → Reviewer に全体レビュー指示

パターン B: 部分レビュー
  独立タスクグループ A が完了 → Reviewer にグループ A のレビュー指示
  並行してグループ B の実装を継続

パターン C: 即時レビュー
  重要度の高いタスクが完了 → すぐにレビュー指示
```

### エスカレーション処理

```
Implementer → SendMessage → Lead:
  "タスク X が 3回失敗。原因: 型エラー解消不能"

Lead の判断:
  1. 別 Implementer に再割当て
  2. タスク分割して再登録
  3. ユーザーにエスカレーション (重大問題時)
```

### リテイクループ

詳細: review-retake-loop.md 参照

## 完了ステージ

### 前提

- 全タスクが completed
- Reviewer の最終判定が APPROVE

### 処理

1. 統合ビルド・テスト最終確認
2. Plans.md 更新 (cc:TODO → cc:done)
3. git commit (Conventional Commits 形式)
4. breezing-active.json 削除
5. Team クリーンアップ
6. メトリクスレポート生成

### 検証実行規則

Phase 4 の統合検証は ultrawork と同一:

| 順位 | 対象 | コマンド |
|------|------|---------|
| 1 | `./tests/validate-plugin.sh` | `bash ./tests/validate-plugin.sh` |
| 2 | `./scripts/ci/check-consistency.sh` | `bash ./scripts/ci/check-consistency.sh` |
| 3 | `package.json` の `test` script | `{pkg_mgr} test` |
| 4 | `package.json` の `lint` script | `{pkg_mgr} run lint` |
| 5 | `pytest.ini` / `pyproject.toml` | `pytest` |
| 6 | `Cargo.toml` | `cargo test` |
| 7 | `go.mod` | `go test ./...` |

> **注**: 該当ファイルが存在する場合は必ず実行し、**失敗した時点で即停止**する。

### 完了レポート

```markdown
🏇 Breezing Complete!

## Summary
- 対象: 認証機能からユーザー管理まで (3 タスク)
- 所要時間: 15 分
- Implementer: 2 並列
- リテイク: 1 回

## Tasks
✅ 3. ログイン機能の実装
✅ 4. 認証ミドルウェアの作成
✅ 5. セッション管理

## Team Activity (TaskCompleted Hook)
| Teammate | 完了タスク | 完了時刻 |
|----------|-----------|---------|
| Implementer #1 | #1, #3 | 14:32, 14:45 |
| Implementer #2 | #2 | 14:38 |
| Reviewer | (レビュー完了) | 14:50 |

## Review
- 判定: APPROVE (Grade: A)
- Codex Review: N/A

## Build & Test
- ビルド: ✅ 成功
- テスト: ✅ 12/12 通過

## Commit
- abc1234: feat: implement auth flow (login, middleware, session)

楽勝でした 🐎💨
```

> **メトリクスの制限**: PostToolUse Hook は Teammate に継承されないため、Teammate 別のトークン数・ツール使用数は取得不可。
> TaskCompleted Hook（Lead 側で発火）により「誰がどのタスクをいつ完了したか」のタイムラインは記録される。
> Lead 自身の agent-trace.jsonl は正常に記録される。全体コストは `/cost` コマンドで確認可能。
