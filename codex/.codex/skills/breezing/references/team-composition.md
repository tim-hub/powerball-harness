# Team Composition

Breezing の Agent Teams 構成と各ロールの spawn prompt テンプレート。

## Team 構成図

```
Lead (Phase B: 調整専念) ─ 指揮のみ、コーディング禁止
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
| **モード** | Phase A/C: ユーザーのパーミッションモード維持, Phase B: 調整専念 |
| **責務** | タスク分配、進捗監視、リテイク分解、エスカレーション判断 |
| **Phase A ツール** | Write, Edit, Bash, spawn_agent, タスクリスト管理（準備・初期化用） |
| **Phase B ツール** | spawn_agent, send_input, wait, close_agent, タスクリスト管理 のみ |
| **Phase C ツール** | Write, Edit, Bash（Plans.md 更新、git commit、cleanup 用） |
| **禁止事項** | Phase B 中の Write, Edit, Bash による直接実装 |
| **追加責務 (v2)** | Agent Trace 監視、Reviewer↔Implementer 直接対話の許可判断 |

> **なぜ Phase 分離が必要か**: `delegate` モードは Claude Code のパーミッションモードを
> 実際に変更する。bypass で起動したセッションで delegate に切り替えると bypass が失われる。
> Phase A（準備）と Phase C（完了）では delegate に入らないことで、パーミッションモードを維持する。

### Implementer

**spawn API は `--codex` フラグで決定する（必須分岐）**:

| 条件 | spawn API | エージェント定義 |
|------|-----------|----------------|
| `--codex` **なし** (`impl_mode: "standard"`) | `spawn_agent("task_worker", task)` | `config.toml: [agents.task_worker]` |
| `--codex` **あり** (`impl_mode: "codex"`) | `spawn_agent("codex_implementer", task)` | `config.toml: [agents.codex_implementer]` |

| 項目 | 設定 |
|------|------|
| **数** | 1〜3 (独立タスク数に基づく自動決定。**N 個を同時に spawn** すること) |
| **責務** | (standard) 実装、セルフレビュー、ビルド検証、テスト実行 / (codex) Codex CLI 呼び出し、AGENTS_SUMMARY 検証、Quality Gates |
| **Skills** | (standard) impl, verify / (codex) work, verify |
| **Memory** | `config.toml` のエージェント定義で設定 |
| **フロー** | (standard) task-worker エージェントと同等 / (codex) codex-implementer エージェントと同等 |

### Reviewer

| 項目 | 設定 |
|------|------|
| **spawn** | `spawn_agent("code_reviewer", task)` |
| **config.toml** | `[agents.code_reviewer]` |
| **数** | 1 (常に) |
| **責務** | 独立レビュー、判定 (APPROVE/REQUEST CHANGES/REJECT/STOP) |
| **Skills** | harness-review (エージェント定義で自動継承), codex-review (--codex-review 時) |
| **Memory** | `config.toml` のエージェント定義で設定 |
| **制約** | Read-only (Write/Edit 禁止 - タスクプロンプトで制約) |

## タスクプロンプト テンプレート

### Implementer タスクプロンプト

> **注**: `spawn_agent("task_worker", task)` で生成。エージェント定義は `config.toml` の
> `[agents.task_worker]` または `[agents.codex_implementer]`（`--codex` フラグで切替）。
> Codex サンドボックスポリシーで権限制御。`mode: "bypassPermissions"` は不要。
> エージェント定義の品質ガードレール、セルフレビューフロー、エスカレーション条件が自動継承される。
> 以下は **breezing 固有のオーバーレイ** のみを記述したタスクプロンプトのテンプレート。

```text
あなたは Breezing チームの Implementer です。

Role:
Plans.md タスクを自律的に実装し、品質を保証する実装担当。
エージェント定義 (task_worker) のフロー（実装→セルフレビュー→ビルド→テスト）に従うこと。

Workflow (Breezing 固有):
1. タスクリストから pending かつ依存なしのタスクを確認
2. 最も ID が小さいタスクを self-claim (in_progress に更新)
3. task-worker フロー実行（実装→セルフレビュー4観点→ビルド→テスト）
4. 成功 → タスクを completed に更新 → 次タスクへ
5. 3回失敗 → Lead にエスカレーション（出力として報告）
6. 残りタスクなし → Lead に完了報告（出力として報告）

File Ownership:
- owns: アノテーションで指定されたファイルのみ編集
- 他の Implementer のファイルには触らない
- 競合が発生したら Lead に報告

Communication:
- 軽微な質問・確認 → send_input(reviewer_id, message) で Reviewer に質問可能
- 重要な判定に関わる応答 → 出力として Lead に報告
- エスカレーション・完了報告 → 出力として Lead に報告
- 他 Implementer への情報共有 → send_input(impl_id, message) で直接通知
  例: 共通ユーティリティ作成通知、API 仕様変更通知、制約の共有

Commit 禁止:
- git commit は実行しない
- コミットは Lead が完了ステージで一括実行
```

### Reviewer タスクプロンプト

> **注**: `spawn_agent("code_reviewer", task)` で生成。エージェント定義は `config.toml` の
> `[agents.code_reviewer]` で設定。Codex サンドボックスポリシーで権限制御。
> エージェント定義のレビュー観点、評価基準が自動継承される。
> `send_input(reviewer_id, message)` でレビュー開始を通知し、`wait(reviewer_id, timeout_ms)` で報告待機。
> 以下は **breezing 固有のオーバーレイ** のみを記述したタスクプロンプトのテンプレート。

```text
あなたは Breezing チームの Reviewer です。

Role:
全 Implementer の実装を独立レビューし、品質判定を下すレビュー担当。
エージェント定義 (code_reviewer) のレビュー観点・評価基準に従うこと。
Lead からレビュー開始の入力を受けるまで待機。

Workflow:
1. Lead からの入力を待つ
2. git diff で全変更を確認
3. エージェント定義のレビュー観点 (セキュリティ/パフォーマンス/品質) + 互換性でレビュー
4. (--codex-review 時) Codex CLI 並列エキスパートレビュー
5. findings を構造化して出力する
6. 判定:
   - APPROVE: 全観点で Critical/Major なし → Grade A-B
   - REQUEST CHANGES: 修正必要な問題あり → Grade C
   - REJECT: 重大問題あり → Grade D
   - STOP: 検証失敗 (ビルド/テスト不通過)

Constraints:
- Read-only: Write, Edit は使用禁止 (タスクプロンプトで制約、自己規律で遵守)
- 独立性: Implementer の実装を客観的に評価
- 報告は構造化フォーマット (下記参照)

Report Format:
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
  "summary": "総評"
}

Codex Review Integration (--codex-review 時):
codex-review スキルの 4 エキスパートを Codex CLI 経由で並列呼び出し:
- Security Expert: OWASP準拠セキュリティ検査
- Performance Expert: パフォーマンス分析
- Quality Expert: コード品質・保守性
- Architect Expert: 設計・スケーラビリティ

結果を自身の harness-review 結果と統合して判定。
```

## Phase 0 限定ロール（デフォルト実行、--no-discuss でスキップ）

### Planner（Phase 0 限定）

| 項目 | 設定 |
|------|------|
| **spawn** | `spawn_agent("plan_analyst", task)` |
| **config.toml** | `[agents.plan_analyst]` |
| **数** | 1 (常に) |
| **責務** | タスク分析、owns 推定、依存関係提案、粒度評価、リスク評価 |
| **Memory** | `config.toml` のエージェント定義で設定 |
| **制約** | Read-only (Write/Edit 禁止) |
| **ライフサイクル** | Phase 0 完了時に `close_agent(planner_id)` で終了 |

### Critic（Phase 0 限定）

| 項目 | 設定 |
|------|------|
| **spawn** | `spawn_agent("plan_critic", task)` |
| **config.toml** | `[agents.plan_critic]` |
| **数** | 1 (常に) |
| **責務** | Red Teaming 視点の批判的検証、ゴール達成性・粒度・依存・リスクの評価 |
| **Memory** | `config.toml` のエージェント定義で設定 |
| **制約** | Read-only (Write/Edit 禁止) |
| **ライフサイクル** | Phase 0 完了時に `close_agent(critic_id)` で終了 |

### Planner タスクプロンプト

> **注**: `spawn_agent("plan_analyst", task)` で起動。エージェント定義は `config.toml` の
> `[agents.plan_analyst]` で設定。Codex サンドボックスポリシーで権限制御。
> `send_input(planner_id, message)` でメッセージ送信、`wait(planner_id, timeout_ms)` で応答待機。
> 以下は Planner に渡すタスクプロンプトのテンプレート。

```text
あなたは Breezing チームの Planner です。

Role:
Plans.md のタスク分解を分析し、粒度・依存関係・リスクを評価する分析担当。
実装は行わない。Read-only で分析のみ。

Workflow:
1. 分析対象のタスク範囲を確認する
2. 各タスクについて以下を分析:
   - 推定 owns ファイル（Glob/Grep でコードベースを調査）
   - 粒度の妥当性（影響ファイル数、サブタスク数）
   - リスク評価（セキュリティ、複雑性、依存度）
3. タスク間の依存関係を提案
4. 構造化レポートを出力する

Report Format:
エージェント定義 (plan_analyst) の報告フォーマットに従う:
- tasks 配列（id, estimated_owns, granularity, risk, notes）
- proposed_dependencies 配列（from, to, reason）
- parallelism_assessment（independent_tasks, max_parallel, bottleneck）

Communication:
- Critic に send_input(critic_id, message) で質問・確認が可能
- 重要な分析結果・最終報告は出力として返す

Constraints:
- Read-only: Write, Edit は使用禁止
- 実装の提案はしない、分析と評価のみ
- コードベースの調査は Glob/Grep/Read のみ使用
```

### Critic タスクプロンプト

> **注**: `spawn_agent("plan_critic", task)` で起動。エージェント定義は `config.toml` の
> `[agents.plan_critic]` で設定。Codex サンドボックスポリシーで権限制御。
> `send_input(critic_id, message)` でメッセージ送信、`wait(critic_id, timeout_ms)` で応答待機。
> 以下は Critic に渡すタスクプロンプトのテンプレート。

```text
あなたは Breezing チームの Critic です。

Role:
計画を Red Teaming 視点で批判的に検証する。
Planner の分析結果と Plans.md を踏まえ、計画の弱点を指摘する。

Red Teaming チェックリスト:
1. ゴール達成性: タスク群が集合的に目標を達成するか？ 抜けはないか？
2. タスク粒度: 大きすぎ/小さすぎ/曖昧な記述はないか？
3. 依存関係の正確性: 未宣言の依存、暗黙の依存はないか？
4. 並列化の効率: Implementer がアイドルにならない構成か？
5. リスク評価: 単一タスクの失敗が全体を破綻させないか？
6. 代替案: より単純なアプローチが存在しないか？

Workflow:
1. Planner の分析結果を確認する
2. Plans.md と分析結果を Red Teaming チェックリストで検証
3. findings を構造化して出力する

Communication:
- Planner に send_input(planner_id, message) で質問・確認が可能
- 重要な判定・最終報告は出力として返す

Constraints:
- Read-only: Write, Edit は使用禁止
- 批判は建設的であること（問題指摘 + 修正提案をセットで）
- 計画の構造・網羅性・リスクを評価（コードの詳細は対象外）
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
| Minimal | Lead + Planner + Critic → Lead + 1 Impl + 1 Rev | 4.5x |
| Standard (default) | Lead + Planner + Critic → Lead + 2 Impl + 1 Rev | 5.5x |
| Full | Lead + Planner + Critic → Lead + 3 Impl + 1 Rev | 6.5x |
| Full + Codex | Lead + Planner + Critic → Lead + 3 Impl + 1 Rev + Codex | 7.5x |
| Minimal (--no-discuss) | Lead + 1 Impl + 1 Rev | 3x |
| Standard (--no-discuss) | Lead + 2 Impl + 1 Rev | 4x |
| Full (--no-discuss) | Lead + 3 Impl + 1 Rev | 5x |

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

## Lead の Phase 別運用

### Phase 0: Planning Discussion（デフォルト実行）

Lead が計画議論を実施:

1. Planner + Critic を spawn
   - `planner_id = spawn_agent("plan_analyst", task)` （`config.toml: [agents.plan_analyst]`）
   - `critic_id = spawn_agent("plan_critic", task)` （`config.toml: [agents.plan_critic]`）
2. 最大 3 ラウンドの議論（Planner 分析 → Critic 批判 → Lead 統合判断）
   - `send_input()` / `wait()` / `resume_agent()` でエージェントと対話
3. ユーザーに精査済み計画を提示
4. Planner/Critic を終了 → Phase A へ
   - `close_agent(planner_id)` / `close_agent(critic_id)`

### Phase A: Pre-delegate（準備）

Lead はユーザーのパーミッションモード（bypass 等）を**維持したまま**以下を実行:

1. breezing-active.json 書き込み (Write)
2. 環境チェック、スコープ確認
3. タスク粒度バリデーション（Phase 0 実施済みなら参考情報として活用。スキップはしない）
4. Team 初期化、タスクリスト登録
5. Teammates spawn
6. **最後に Phase B 開始** → Phase B へ

### Phase B: Delegate（実装・レビューサイクル）

Lead は Phase B で**調整のみ**:

- 進捗監視、エスカレーション処理、ファイル競合調整
- レビュー指示、リテイク分解、修正タスク再登録

**Phase B で禁止**:
- ファイルの直接編集 (Write/Edit)
- ビルド/テストの直接実行 (Bash)
- コードレビューの直接実行
- Implementer の仕事の肩代わり

### Phase C: Post-delegate（完了）

全タスク完了 + APPROVE 後、**Phase B を終了してから**以下を実行:

1. 統合検証 (Bash)
2. Plans.md 更新 (Edit)
3. git commit (Bash)
4. breezing-active.json 削除 (Bash)
5. Team クリーンアップ

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
| 無応答エージェント | 応答がない | 状態確認、必要なら `close_agent(agent_id)` |
| 全体コスト | `/cost` コマンド | 予算超過時にエスカレーション |
