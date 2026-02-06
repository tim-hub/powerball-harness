# Review / Retake Loop

Breezing における三者分離リテイクフローと Reviewer↔Implementer 直接対話パターン。

## 概要

```
ultrawork: Lead が自己レビュー → 自己修正 (一人三役)
breezing: Reviewer が独立レビュー → Lead が分解 → Implementer が修正 (三者分離)
```

三者分離により:
- レビューの客観性が向上 (自分のコードを自分でレビューしない)
- Lead は調整に専念できる
- 修正と再レビューが並行可能

## 通信パターン (v2)

### パターン A: 軽微な質問・確認（直接対話）

```text
Reviewer → SendMessage → Implementer:
  「この関数、なぜ async にしていない？」

Implementer → SendMessage → Reviewer:
  「DB アクセスがないので同期で十分です」

Reviewer: 了解、問題なし → レビュー継続
```

**使用条件**: 実装意図の確認、命名の理由、設計判断の質問など
**Lead の関与**: 不要（監視のみ）

### パターン B: 修正指示（Lead 経由）

```text
Reviewer → SendMessage → Lead:
  findings 報告 (構造化フォーマット)

Lead:
  1. findings を修正タスクに分解
  2. TaskCreate で修正タスク登録
  3. Implementer に SendMessage で修正指示

Implementer:
  修正タスクを claim → 実装 → 完了報告
```

**使用条件**: コードの変更が必要な場合
**Lead の関与**: 必須（タスク分解と管理）

### パターン C: エスカレーション（Lead 経由）

```text
Implementer → SendMessage → Lead:
  「タスク X が 3回失敗。原因: 型エラー解消不能」

Lead の判断:
  1. 別 Implementer に再割当て
  2. タスク分割して再登録
  3. ユーザーにエスカレーション (重大問題時)
```

**使用条件**: 自力解決不能な問題
**Lead の関与**: 必須（判断と対応）

### パターン使い分け基準

| 状況 | パターン | 理由 |
|------|---------|------|
| 「なぜこう書いた？」という質問 | A (直接) | 確認のみ、コード変更不要 |
| 「ここのロジックは正しい？」 | A (直接) | 確認のみ |
| 「入力バリデーションが不足」 | B (Lead経由) | コード変更が必要 |
| 「N+1クエリを修正して」 | B (Lead経由) | コード変更が必要 |
| 「セキュリティ Critical 検出」 | B (Lead経由) | 重大問題、Lead 判断必要 |
| 「ビルドが通らない」 | C (エスカレーション) | 自力解決不能 |

## レビューフロー図

```text
Lead がレビュー指示 (任意のタイミング)
    ↓
┌──────────────────────────────────────────────────────┐
│ Step 1: Reviewer がレビュー実行                       │
│  a. git diff --merge-base で全変更を確認              │
│  b. harness-review 4観点:                             │
│     ├── セキュリティ                                  │
│     ├── パフォーマンス                                │
│     ├── 品質                                          │
│     └── 互換性                                        │
│  c. (--codex-review) Codex MCP 4エキスパート          │
│  d. 不明点は Implementer に直接質問 (パターン A)      │
│  e. findings 集約 → 判定                              │
└──────────────────────────────────────────────────────┘
    ↓
┌──────────────────────────────────────────────────────┐
│ Step 2: Reviewer → Lead に SendMessage (判定報告)     │
│                                                       │
│  ├── APPROVE (Grade A-B)                              │
│  │   → 完了処理へ                                    │
│  │                                                    │
│  ├── REQUEST CHANGES (Grade C)                        │
│  │   → Step 3 へ (リテイクループ)                     │
│  │                                                    │
│  ├── REJECT (Grade D)                                 │
│  │   → 即停止 + ユーザーに手動修正を要請              │
│  │                                                    │
│  └── STOP (検証失敗)                                  │
│      → 即停止 + ビルド/テスト失敗報告                 │
└──────────────────────────────────────────────────────┘
    ↓ REQUEST CHANGES の場合
┌──────────────────────────────────────────────────────┐
│ Step 3: Lead がリテイク処理                           │
│  a. findings を修正タスクに分解                       │
│  b. 修正タスクを TaskCreate で登録                    │
│  c. 担当 Implementer に SendMessage で修正指示        │
│  d. retake_count++ (breezing-active.json 更新)       │
│  e. retake_count > 3 → ユーザーにエスカレーション    │
└──────────────────────────────────────────────────────┘
    ↓
┌──────────────────────────────────────────────────────┐
│ Step 4: Implementer が修正実行                        │
│  - 修正タスクを self-claim → 実装 → ビルド → テスト  │
│  - 完了 → Lead に SendMessage で修正完了報告          │
└──────────────────────────────────────────────────────┘
    ↓
┌──────────────────────────────────────────────────────┐
│ Step 5: Lead → Reviewer に再レビュー指示              │
│  「修正が完了しました。再レビューをお願いします」      │
│  → Step 1 へ戻る                                      │
└──────────────────────────────────────────────────────┘
```

## 判定基準

### APPROVE

```json
{
  "decision": "APPROVE",
  "grade": "A",
  "conditions": [
    "Critical/Major findings が 0 件",
    "全観点で Grade B 以上",
    "ビルド成功",
    "テスト全通過"
  ]
}
```

### REQUEST CHANGES

```json
{
  "decision": "REQUEST_CHANGES",
  "grade": "C",
  "conditions": [
    "Critical findings が 0 件",
    "Major findings が 1 件以上",
    "自動修正可能な問題が主"
  ]
}
```

### REJECT

```json
{
  "decision": "REJECT",
  "grade": "D",
  "conditions": [
    "Critical findings が 1 件以上",
    "セキュリティ脆弱性",
    "設計レベルの根本的問題"
  ]
}
```

### STOP

```json
{
  "decision": "STOP",
  "grade": "N/A",
  "conditions": [
    "ビルドが失敗",
    "テストが失敗",
    "環境エラー"
  ]
}
```

## findings → 修正タスク分解

Lead が REQUEST CHANGES を受けた際の分解ロジック:

### 分解ルール

```
1. findings をファイル別にグループ化
2. 同一ファイルの findings → 1 修正タスクにまとめる
3. auto_fixable: true の findings を優先
4. Critical → Major → Warning の優先順位で対応
```

### 修正タスク生成例

Reviewer からの findings:

```json
{
  "findings": [
    {
      "severity": "warning",
      "category": "security",
      "file": "src/auth/login.ts",
      "line": 15,
      "issue": "入力バリデーション不足",
      "suggestion": "zod スキーマでバリデーション追加",
      "auto_fixable": true
    },
    {
      "severity": "warning",
      "category": "quality",
      "file": "src/auth/login.ts",
      "line": 42,
      "issue": "エラーハンドリング不足",
      "suggestion": "try-catch で適切にハンドリング",
      "auto_fixable": true
    },
    {
      "severity": "warning",
      "category": "performance",
      "file": "src/db/users.ts",
      "line": 8,
      "issue": "N+1 クエリ",
      "suggestion": "JOIN または一括取得に変更",
      "auto_fixable": false
    }
  ]
}
```

Lead が生成する修正タスク:

```
TaskCreate:
  subject: "src/auth/login.ts のセキュリティ・品質修正"
  description: |
    以下の 2 件を修正:
    1. L15: 入力バリデーション不足 → zod スキーマ追加
    2. L42: エラーハンドリング不足 → try-catch 追加
    owns: src/auth/login.ts

TaskCreate:
  subject: "src/db/users.ts の N+1 クエリ修正"
  description: |
    L8: N+1 クエリ → JOIN または一括取得に変更
    owns: src/db/users.ts
```

## リテイク回数管理

### retake_count の管理

breezing-active.json の `review.retake_count` でリテイク回数を追跡:

```
REQUEST CHANGES 受信 → retake_count++
APPROVE 受信 → リテイクループ終了
retake_count > max_retakes → エスカレーション
```

### エスカレーション条件

| 条件 | アクション |
|------|-----------|
| retake_count > 3 | ユーザーにエスカレーション |
| REJECT | 即停止 + ユーザー報告 |
| STOP | 即停止 + 検証失敗報告 |

### エスカレーション時のメッセージ

Lead が会話コンテキスト内の Reviewer 報告履歴からメッセージを構成:

```text
⚠️ Breezing: リテイク上限 (3回) に達しました

## 未解決の問題

| # | ファイル | 問題 | 重要度 |
|---|---------|------|--------|
| 1 | src/auth/login.ts:15 | 入力バリデーション | warning |
| 2 | src/db/users.ts:8 | N+1 クエリ | warning |

## リテイク経過 (Lead の Reviewer 報告受信履歴から構成)

| 回 | 判定 | Grade | 指摘数 |
|----|------|-------|--------|
| 1 | REQUEST_CHANGES | C | 5 |
| 2 | REQUEST_CHANGES | C | 3 |
| 3 | REQUEST_CHANGES | C | 2 |

手動で修正するか、`/breezing 続きやって` で再開してください。
```

> **データ源**: リテイク履歴は breezing-active.json には保存しない。
> Lead の会話コンテキスト内に蓄積された Reviewer の SendMessage 報告から動的に構成する。
> `review.retake_count` のみ breezing-active.json で永続化。

## Codex Review 統合 (--codex-review)

### フロー

```
Reviewer の通常レビュー (harness-review 4観点)
    +
Codex MCP 並列エキスパートレビュー (4エキスパート)
    ↓
結果統合 → 総合判定
```

### 統合ルール

```
1. harness-review の findings を基盤とする
2. Codex エキスパートの findings を追加
3. 重複する findings はマージ (harness-review 側を優先)
4. 最終判定は harness-review の基準に従う
5. Codex が Critical を検出した場合は Grade を1段階下げる
```

詳細: codex-review-integration.md 参照

## REJECT / STOP テンプレート

### REJECT

```markdown
### Manual Intervention Required

**Decision**: REJECT
**Grade**: D
**Reason**: 重大な問題があり、自動修正では対応できません

**Critical Issues**:
1. [ファイル:行] 問題の説明
2. ...

Breezing を停止しました。手動で修正してください。
```

### STOP

```markdown
### Verification Failed

**Decision**: STOP
**Grade**: N/A (blocked)
**Failure Type**: [lint_failure | test_failure | build_failure | environment_error]
**Failed Command**: ...

**Required Fixes**:
1. ...

Breezing を停止しました。問題を解決して `/breezing 続きやって` で再開してください。
```
