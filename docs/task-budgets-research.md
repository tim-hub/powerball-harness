# Task Budgets (Public Beta) 調査メモ

Phase 44.10.1 作成日: 2026-04-18

本ドキュメントは Anthropic API に追加された **Task Budgets (public beta)** の仕様を整理し、
Harness 既存の制御機構との競合関係を分析したうえで、
**本 Phase での採用を見送る判断の根拠と、将来の採用判定タイミング**を記録する。

---

## 1. Task Budgets の API 仕様要約

> **注意**: 以下は 2026-04-18 時点での public beta 仕様に基づく要約。
> 正確なフィールド名・スキーマ・エラーコードは公式 Anthropic ドキュメントを参照すること。
> フィールド名等で不確実な箇所は `(推定)` と明記する。

### 1-1. 概要

Task Budgets は Anthropic API の Messages API または Agents API に追加されたパラメータ群で、
1 回のエージェント呼び出し（または task 単位）が消費できるリソースの上限を宣言的に指定する機能。

トークン消費・コストを事前制約として渡すことで、
エージェントの暴走・予期しない高額請求を API レイヤで防ぐことを目的とする。

### 1-2. 入力パラメータ（推定）

公式ドキュメント未確定の beta 段階のため、以下は Anthropic の公開情報および既知の類似機能から推定したもの。

| パラメータ名 | 型 | 説明 |
|------------|-----|------|
| `max_input_tokens` (推定) | integer | 1 タスクが消費できる入力トークン上限 |
| `max_output_tokens` (推定) | integer | 1 タスクが生成できる出力トークン上限 |
| `max_cost_usd` (推定) | number | USD 換算の最大コスト上限 |

これらのパラメータは API リクエストの `task_budget` オブジェクト (推定) 内に指定すると見られる。

```json
// 構造例（推定 — 実際の field 名は Anthropic docs を参照）
{
  "model": "claude-opus-4-7",
  "messages": [...],
  "task_budget": {
    "max_input_tokens": 100000,
    "max_output_tokens": 8000,
    "max_cost_usd": 2.00
  }
}
```

### 1-3. 出力・エラー形式（推定）

上限を超えた場合は、通常のストリーミング完了とは異なる `budget_exhausted` エラーが返ると見られる。

```json
// エラーレスポンス例（推定）
{
  "type": "error",
  "error": {
    "type": "budget_exhausted",
    "message": "Task budget exceeded: max_cost_usd limit reached",
    "exhausted_budget": "max_cost_usd"
  }
}
```

エージェントループを組んでいる場合、このエラーを受け取ったループ側で
「予算切れによる早期終了」として扱う必要がある。

### 1-4. 現行ステータス

- **Public beta** — 一部のユーザー/API tier に限定公開
- GA（一般提供）の時期は未公表（2026-04-18 時点）
- スキーマの後方互換性は保証されていない（beta の常）

---

## 2. Harness 既存機構との競合関係

Task Budgets が提供する「リソース上限」機能は、Harness が既に複数の独自機構で実装している概念と重なる。
以下の表で競合ポイントを整理する。

| Harness 既存機構 | 場所 | 制御対象 | Task Budgets との重複度 |
|----------------|------|---------|----------------------|
| Advisor 相談回数上限（最大 3 回） | `agents/worker.md` の Advisor 相談判定 | Worker が Advisor を呼べる回数 | 低（用途が異なる） |
| `maxTurns` | `agents/worker.md` frontmatter（Worker: 100, Reviewer: 50） | エージェントのターン数上限 | 中（間接的にトークン消費を制限） |
| `effort` frontmatter | `agents/worker.md`、各 skill frontmatter | thinking 強度（low/medium/high） | 中（出力トークン消費に影響） |
| `/cost` per-model breakdown | CC 組み込み（v2.1.92） | コスト可視化（事後確認） | 低（事後確認であり、事前制約ではない） |
| `scripts/detect-review-plateau.sh` | `skills/harness-loop` Step 6 | レビューループの行き詰まり検知 | 低（品質ゲート。コスト制御ではない） |
| harness-loop `--max-cycles N` | `skills/harness-loop` | ループサイクル数上限（デフォルト 8） | 中（間接的にセッション全体の消費を制限） |
| Advisor `STOP` 判定 | `agents/worker.md`、`skills/harness-loop` Advisor Strategy | 危険タスクの手動エスカレーション | 低（品質ゲート。コスト制御ではない） |

### 2-1. 最も競合度が高いポイント

**`maxTurns` vs `max_input_tokens` / `max_output_tokens`**:

- `maxTurns: 100` は Worker のターン数を制限し、間接的にトークン消費を抑制する
- `max_input_tokens` はより直接的にトークン数を制限する
- ただし `maxTurns` は「ループ制御」、`max_input_tokens` は「コスト制御」で主目的が異なる

**`harness-loop --max-cycles` vs `max_cost_usd`**:

- `--max-cycles N` はサイクル数で長時間セッションのコストを間接制御する
- `max_cost_usd` はドル換算で直接制御する
- `--max-cycles 8`（デフォルト）は実際のコストが不明なまま制限する粗い方法であり、
  Task Budgets の `max_cost_usd` はより精密

### 2-2. 非競合な領域

以下は Harness 固有の概念であり、Task Budgets では代替できない:

- `plateau 検知` — 品質ループの行き詰まり（コスト制御ではなく品質ゲート）
- Advisor 相談回数上限 — 方針相談の回数制御（コスト制御ではなくガバナンス）
- `effort` レベル — thinking 品質の調整（コスト削減ではなく品質/コストのトレードオフ調整）

---

## 3. 採用するなら「どの skill で」「どの粒度で」

将来 Task Budgets を採用する場合の粒度と適用箇所の候補を整理する。

### 3-1. Per-task budget（1 タスクの Worker spawn に上限）

| 項目 | 内容 |
|------|------|
| 適用箇所 | `agents/worker.md` の Agent 呼び出し時 / `scripts/codex-companion.sh` 経由の Codex spawn 時 |
| 粒度 | sprint-contract に `task_budget` セクションを追加し、タスクごとに上限を指定 |
| メリット | 単一の重いタスクが予算を食い尽くすのを防止できる |
| 課題 | タスクの適切な budget 見積もりが必要。過小見積もりは途中打ち切りを招く |
| 実装ポイント | `scripts/generate-sprint-contract.js` に `task_budget` セクション生成ロジックを追加 |

### 3-2. Per-session budget（1 breezing session 全体に上限）

| 項目 | 内容 |
|------|------|
| 適用箇所 | `skills/harness-loop/SKILL.md` のループ開始時、または `skills/harness-work` の breezing モード |
| 粒度 | セッション全体に `max_cost_usd` を設定し、超過時はループを停止して報告 |
| メリット | 長時間セッションのコスト上限を明確に定義できる。`--max-cycles` よりも精密 |
| 課題 | セッション途中での budget 枯渇を graceful に処理するロジックが必要 |
| 実装ポイント | `harness-loop` の `[Step 9] 次 wake-up 予約` 前に残余 budget をチェックするステップを追加 |

### 3-3. Per-day budget（ユーザー単位の日次上限）

| 項目 | 内容 |
|------|------|
| 適用箇所 | Harness レイヤよりも上位（Anthropic ダッシュボードの API usage limits またはユーザー側の外部制御） |
| 粒度 | 日次の累積コストが閾値を超えたらアラート / 自動停止 |
| メリット | 予期しない高額請求を確実に防止できる |
| 課題 | Harness 側での実装が困難（セッションをまたいだ状態管理が必要）。harness-mem との連携が必要 |
| 実装ポイント | `harness_mem_record_checkpoint` を利用した日次コスト累積記録と、`harness-loop` 起動時のチェックが考えられる |

---

## 4. 本 Phase では採用しない判断 + rationale

**判断: Phase 44 では Task Budgets の実装を見送る。**

### 理由 1: Public beta のため API 安定性が不確実

Task Budgets は 2026-04-18 時点で public beta。
スキーマの後方互換性が保証されず、フィールド名や挙動が GA 前に変更される可能性が高い。
beta API を Harness の中核制御（sprint-contract 生成・harness-loop）に組み込むと、
Anthropic 側の変更が Harness の破壊的変更に直結するリスクがある。

### 理由 2: 既存の `maxTurns` + `--max-cycles` + Advisor STOP で 80% はカバー済み

| 保護したいリスク | 既存の対応 |
|----------------|-----------|
| 単一 Worker の暴走 | `maxTurns: 100` で打ち切り |
| 長時間セッションの過剰消費 | `--max-cycles 8`（デフォルト）で停止 |
| 品質ループの行き詰まり | `detect-review-plateau.sh` + `PIVOT_REQUIRED` |
| 高リスクタスクの過剰相談 | Advisor 相談上限 最大 3 回 |

Task Budgets が提供する「ドル換算の直接制御」は便利だが、
現状の制御機構と組み合わせることで実質的なコスト暴走は防げている。

### 理由 3: Phase 44 の優先度配分

Phase 44 のスコープは以下を中心とする:

- Plugin agent exposure の解消
- Phase 45 sync.go 修正
- PreCompact hook 実装
- Task Budgets は **調査・ドキュメント整備のみ**（本タスク）が Phase 44 の目標

Phase 44 の残りリソースを Task Budgets 実装に投入するより、
上記の確定スコープを完遂する優先度のほうが高い。

### 理由 4: harness-mem との統合設計が未定

Per-day budget のような日次累積管理は `harness-mem` の checkpoint 機能との統合が必要。
統合設計を詰めずに実装すると、後の `harness-mem` 設計変更時に技術的負債になる。
設計を正しく行うには別の調査タスクが必要であり、Phase 44 のスコープには含まない。

---

## 5. 次サイクルでの採用判定タイミング

以下のトリガー条件のいずれかを満たした時点で再評価を行う。

| トリガー | 詳細 | 評価フェーズ |
|---------|------|------------|
| Task Budgets が GA に昇格 | Anthropic が GA を公式アナウンスした時点でスキーマを再確認し、採用可否を判断する | Phase 45 以降、GA 確認後 |
| `maxTurns` だけでは防げない実際のコスト超過が発生した | Harness 運用中に `maxTurns` の上限に達する前に高額課金が発生した場合 | 発生次第、緊急対応タスクを立てる |
| `harness-mem` の累積記録設計が固まった | harness-mem の日次集計機能が実装・安定化した場合、per-day budget の実装を着手する | Phase 45 以降 |
| Anthropic が Harness 向けの推奨実装パターンを公開した | 公式ドキュメントやブログで Task Budgets + agent framework の統合例が示された場合 | 確認後 1 Phase 以内 |

### 再評価時の確認事項

再評価時には以下を確認すること:

1. フィールド名が本ドキュメントの推定と一致しているか（`max_input_tokens` 等）
2. `budget_exhausted` エラーの graceful handling が sprint-contract ループに組み込めるか
3. per-task / per-session / per-day の 3 粒度のうち、どれから着手するかを決定する
4. `scripts/generate-sprint-contract.js` への `task_budget` セクション追加が技術的に可能か

---

## 参照ファイル

- `agents/worker.md` — Advisor 相談上限・`maxTurns`・`effort` frontmatter
- `docs/CLAUDE-feature-table.md` — `task budgets` エントリ（行 210 付近、`A: 明示追従対象`）
- `skills/harness-loop/SKILL.md` — `--max-cycles`・plateau 検知・Advisor Strategy
- `.claude/rules/opus-4-7-prompt-audit.md` — 2.1.111 運用ノブの定義
