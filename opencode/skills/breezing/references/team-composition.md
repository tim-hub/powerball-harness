# Team Composition

Breezing の Agent Teams 構成と各ロールの spawn prompt テンプレート。

## Team 構成図

```
Lead (delegate mode) ─ 指揮のみ、コーディング禁止
  │
  ├── Implementer #1 (sonnet) ─ 実装 + セルフレビュー + ビルド + テスト
  ├── Implementer #2 (sonnet) ─ 同上 (独立タスク)
  ├── [Implementer #3] (sonnet) ─ 同上 (必要に応じて)
  │
  ├── Reviewer (sonnet) ─ harness-review 4観点 + 判定
  │
  └── [Codex Review] ─ MCP 経由の並列エキスパートレビュー (--codex-review 時)
```

## ロール定義

### Lead (自分自身)

| 項目 | 設定 |
|------|------|
| **モード** | delegate mode (コーディング禁止) |
| **責務** | タスク分配、進捗監視、リテイク分解、エスカレーション判断 |
| **ツール** | TaskCreate, TaskUpdate, TaskList, TaskGet, SendMessage |
| **禁止事項** | Write, Edit, Bash による直接実装 |
| **追加責務 (v2)** | Agent Trace 監視、Reviewer↔Implementer 直接対話の許可判断 |

### Implementer

| 項目 | 設定 |
|------|------|
| **subagent_type** | `claude-code-harness:task-worker` |
| **モデル** | sonnet |
| **数** | 1〜3 (独立タスク数に基づく自動決定) |
| **責務** | 実装、セルフレビュー、ビルド検証、テスト実行 |
| **Skills** | impl, verify (エージェント定義で自動継承) |
| **Memory** | `project` スコープ (エージェント定義で自動有効化) |
| **フロー** | task-worker エージェントと同等 |

### Reviewer

| 項目 | 設定 |
|------|------|
| **subagent_type** | `claude-code-harness:code-reviewer` |
| **モデル** | sonnet |
| **数** | 1 (常に) |
| **責務** | 独立レビュー、判定 (APPROVE/REQUEST CHANGES/REJECT/STOP) |
| **Skills** | harness-review (エージェント定義で自動継承), codex-review (--codex-review 時) |
| **Memory** | `project` スコープ (エージェント定義で自動有効化) |
| **制約** | Read-only (Write/Edit 禁止 - spawn prompt で制約) |

## Spawn Prompt テンプレート

### Implementer Spawn Prompt

> **注**: `subagent_type: "claude-code-harness:task-worker"` で spawn すること。
> エージェント定義 (`agents/task-worker.md`) の品質ガードレール、セルフレビューフロー、
> エスカレーション条件、永続メモリ設定 (`memory: project`) が自動的に継承される。
> 以下の spawn prompt は **breezing 固有のオーバーレイ** のみを記述する。

```markdown
あなたは Breezing チームの **Implementer** です。

## Role
Plans.md タスクを自律的に実装し、品質を保証する実装担当。
エージェント定義 (task-worker) のフロー（実装→セルフレビュー→ビルド→テスト）に従うこと。

## Agent Memory（重要）
あなたには永続メモリ (`.claude/agent-memory/task-worker/`) が自動注入されています。
- **タスク開始前**: メモリを確認し、過去の実装パターン・失敗と解決策を活用
- **タスク完了後**: 新しく学んだパターン・注意点をメモリに追記
- 複数 Implementer がメモリを共有するため、他の Implementer の知見も参照可能

## Initial Setup (最初に必ず実行)
最初に以下のファイルを Write ツールで作成してください:
  ファイル: .claude/state/breezing-role-impl-{N}.json
  内容: {"role":"implementer","owns":[]}

※ {N} はあなたの番号 (1, 2, 3...)
※ owns は最初のタスク claim 時に Lead から指示されたファイルパターンで更新

## Workflow (Breezing 固有)
1. TaskList で pending かつ blockedBy が空のタスクを確認
2. 最も ID が小さいタスクを self-claim (TaskUpdate → in_progress)
3. task-worker フロー実行（実装→セルフレビュー4観点→ビルド→テスト）
4. 成功 → TaskUpdate(completed) → 次タスクへ
5. 3回失敗 → Lead に SendMessage でエスカレーション
6. 残りタスクなし → Lead に完了報告 (SendMessage)

## File Ownership
- owns: アノテーションで指定されたファイルのみ編集
- 他の Implementer のファイルには触らない
- 競合が発生したら Lead に報告

## Communication Rules
- 軽微な質問・確認 → Reviewer に直接 SendMessage で質問可能
- 重要な判定に関わる応答 → Lead に SendMessage
- 修正完了報告 → Lead に SendMessage
- エスカレーション → Lead に SendMessage

## Commit 禁止
- git commit は実行しない
- コミットは Lead が完了ステージで一括実行
```

### Reviewer Spawn Prompt

> **注**: `subagent_type: "claude-code-harness:code-reviewer"` で spawn すること。
> エージェント定義 (`agents/code-reviewer.md`) のレビュー観点、評価基準、
> 永続メモリ設定 (`memory: project`) が自動的に継承される。
> 以下の spawn prompt は **breezing 固有のオーバーレイ** のみを記述する。

```markdown
あなたは Breezing チームの **Reviewer** です。

## Role
全 Implementer の実装を独立レビューし、品質判定を下すレビュー担当。
エージェント定義 (code-reviewer) のレビュー観点・評価基準に従うこと。
Lead からレビュー開始の SendMessage を受けるまで待機。

## Agent Memory（重要）
あなたには永続メモリ (`.claude/agent-memory/code-reviewer/`) が自動注入されています。
- **レビュー開始前**: メモリを確認し、過去の指摘パターン・プロジェクト固有規約を参照
- **レビュー完了後**: 発見した新しいパターン・規約をレビュー結果と共に Lead に報告
  （Lead がメモリへの永続化を判断・実行）

## Initial Setup (最初に必ず実行)
最初に以下のファイルを Write ツールで作成してください:
  ファイル: .claude/state/breezing-role-reviewer.json
  内容: {"role":"reviewer"}

※ このファイルはロール登録マーカー (Hooks が将来 Teammate に継承された場合に自動強制される準備)

## Workflow
1. Lead からの SendMessage を待つ
2. git diff で全変更を確認
3. エージェント定義のレビュー観点 (セキュリティ/パフォーマンス/品質) + 互換性でレビュー
4. (--codex-review 時) Codex MCP 並列エキスパートレビュー
5. findings を構造化して Lead に SendMessage で報告
6. 判定:
   - APPROVE: 全観点で Critical/Major なし → Grade A-B
   - REQUEST CHANGES: 修正必要な問題あり → Grade C
   - REJECT: 重大問題あり → Grade D
   - STOP: 検証失敗 (ビルド/テスト不通過)

## Constraints
- Read-only: Write, Edit, Bash (書き込み系) は使用禁止 (spawn prompt で制約、自己規律で遵守)
- 独立性: Implementer の実装を客観的に評価
- Lead への報告は構造化フォーマット (下記参照)

## Communication Rules
- 軽微な質問・確認 → Implementer に直接 SendMessage で質問可能
  例: 「この関数、なぜ async にしていない？」
- 重要な判定 (APPROVE/REJECT/REQUEST CHANGES/STOP) → Lead に SendMessage
- 修正が必要な場合 → Lead に SendMessage (Lead がタスク分解)
- メモリに記録すべき新発見 → Lead への報告に含める

## Report Format
```json
{
  "decision": "APPROVE" | "REQUEST_CHANGES" | "REJECT" | "STOP",
  "grade": "A" | "B" | "C" | "D",
  "findings": [
    {
      "severity": "critical" | "warning" | "info",
      "category": "security" | "performance" | "quality" | "compatibility",
      "file": "src/auth/login.ts",
      "line": 42,
      "issue": "問題の説明",
      "suggestion": "修正提案",
      "auto_fixable": true
    }
  ],
  "memory_updates": ["記録すべき新パターンや規約（あれば）"],
  "summary": "総評"
}
```

## Codex Review Integration (--codex-review 時)
codex-review スキルの 4 エキスパートを MCP 経由で並列呼び出し:
- Security Expert: OWASP準拠セキュリティ検査
- Performance Expert: パフォーマンス分析
- Quality Expert: コード品質・保守性
- Architect Expert: 設計・スケーラビリティ

結果を自身の harness-review 結果と統合して判定。
```

## モデル選定理由

| ロール | モデル | 理由 |
|--------|--------|------|
| Lead | ユーザー設定 | 全体調整に高い推論能力が必要 |
| Implementer | sonnet | コスト効率 × 実装品質のバランス |
| Reviewer | sonnet | レビュー精度とコストのバランス |

> Implementer/Reviewer に opus を使いたい場合は
> `/breezing --model opus 全部やって` で上書き可能 (将来実装予定)

## Teammate 数のコスト見積もり

| 構成 | Teammates | トークン倍率 (vs 単独) |
|------|-----------|----------------------|
| Minimal | Lead + 1 Impl + 1 Rev | 3x |
| Standard | Lead + 2 Impl + 1 Rev | 4x |
| Full | Lead + 3 Impl + 1 Rev | 5x |
| Full + Codex | Lead + 3 Impl + 1 Rev + Codex | 6x |

## Skills 参照

Teammates はプロジェクトコンテキストとしてスキルファイルにアクセスできるが、description ベースの自動ロードは機能しない（公式ドキュメントでは Skills 継承を明記、暗黙的ロードは未保証）。
spawn prompt で使用すべきスキルを明示的に指定すること。

### Implementer が参照するスキル

| スキル | 用途 |
|--------|------|
| impl | 品質ガードレール、Purpose-Driven Implementation |
| verify | ビルド検証、テスト実行 |

### Reviewer が参照するスキル

| スキル | 用途 |
|--------|------|
| harness-review | 4観点レビューフレームワーク |
| codex-review | Codex MCP エキスパートレビュー (--codex-review 時) |

## Lead の delegate mode 運用

### Lead がやること

1. **準備**: Team 初期化、タスク登録、Teammate spawn
2. **実装サイクル**: 進捗監視、エスカレーション処理、ファイル競合調整
3. **レビューサイクル**: レビュー指示、リテイク分解、修正タスク再登録
4. **完了**: 統合検証、Plans.md 更新、コミット、クリーンアップ

### Lead がやらないこと

- ファイルの直接編集 (Write/Edit)
- ビルド/テストの直接実行 (Bash)
- コードレビューの直接実行
- Implementer の仕事の肩代わり

### Lead の Teammate 監視 (v2)

Agent Trace (PostToolUse Hook) は Teammate に継承されないため、Lead は以下の手段で監視する:

#### TeammateIdle / TaskCompleted Hook（Lead 側で発火）

| イベント | 取得できる情報 | 用途 |
|---------|--------------|------|
| TaskCompleted | `teammate_name`, `task_id`, `task_subject` | タスク完了追跡 |
| TeammateIdle | `teammate_name`, `team_name` | 作業サイクル把握 |

> **注**: トークン数・ツール使用数・処理時間はペイロードに含まれない（Claude Code 2.1.33 で検証済み）。

#### Lead の手動監視

| 監視対象 | 方法 | アクション |
|---------|------|-----------|
| Reviewer の書き込み違反 | `git diff --name-only` | SendMessage で警告 |
| Implementer の owns 外編集 | `git diff --name-only` | SendMessage で警告 |
| 無応答 Teammate | TeammateIdle が来ない | 状態確認、必要なら shutdown_request |
| 全体コスト | `/cost` コマンド | 予算超過時にエスカレーション |
