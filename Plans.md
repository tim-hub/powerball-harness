# Claude Code Harness — Plans.md

作成日: 2026-03-02
前回アーカイブ: Phase 13〜15 → `.claude/memory/archive/Plans-2026-03-02-pre-phase16.md`

---

## Phase 16: Claude Code v2.1.63 対応 — /simplify・/batch 統合 + インフラ更新

作成日: 2026-03-02
起点: Claude Code CHANGELOG v2.1.52〜v2.1.63 の全変更分析
目的: ハーネスの対応バージョンを v2.1.51 → v2.1.63 に引き上げ、新機能を活用

### 背景

Claude Code v2.1.63 で追加された `/simplify`（3並列レビュー→自動修正）と `/batch`（大規模並列マイグレーション）の2つの bundled コマンドがハーネスの既存ワークフローと補完関係にある。加えて HTTP hooks、auto-memory の worktree 共有等のインフラ改善をハーネスに反映する。

### 3つの「simplify」の関係

| 名前 | 提供元 | アーキテクチャ | CLAUDE.md 参照 | 呼び出し方 |
|------|--------|-------------|---------------|----------|
| `/simplify` (bundled) | CC v2.1.63 組み込み | 3並列エージェント（Reuse/Quality/Efficiency） | する | Skill tool |
| `code-simplifier` (プラグイン) | Anthropic 公式マーケットプレイス | 単一 Opus エージェント（Clarity/Consistency/Maintainability） | する | Task tool (`subagent_type: "code-simplifier:code-simplifier"`) |
| `harness-review` | ハーネス | 4観点レビュー（指摘のみ、修正しない） | する | Skill tool |

`/simplify` は `code-simplifier` にインスパイアされた進化版（Boris Cherny 時系列より）。両者は別実装で共存可能。

### `/batch` と `/breezing` の棲み分け

| 観点 | `/batch` | `/breezing` |
|------|---------|------------|
| 用途 | 横展開（同じ変更の大量適用） | 縦展開（異なるタスクの Plan→Work→Review） |
| 入力 | 自然言語の指示 1 行 | Plans.md のタスクリスト |
| レビュー | `/simplify` 自動適用 | 独立 Reviewer Teammate（三者分離） |
| 出力 | 複数の PR（ユニットごと） | 1つの git commit |
| 耐障害性 | なし | breezing-active.json + TaskList 二層永続化 |

→ 競合ではなく補完。横展開タスクでは breezing Lead が `/batch` に委任する設計。

### 優先度マトリクス

| 優先度 | Phase | 内容 | タスク数 |
|--------|-------|------|---------|
| **Required** | 16.1 | /work に Phase 3.5 Auto-Refinement 追加 | 5 |
| **Required** | 16.2 | breezing に /batch 委任判断ロジック追加 | 3 |
| **Required** | 16.3 | feature-table v2.1.63 更新 | 3 |
| **Recommended** | 16.4 | hooks-editing.md に HTTP hooks 仕様追記 | 2 |
| **Recommended** | 16.5 | CLAUDE.md + ドキュメント バージョン表記更新 | 3 |
| Required | 16.6 | 検証 + 同期 | 3 |

合計: **19 タスク**

---

### Phase 16.1: /work に Phase 3.5 Auto-Refinement 追加 [P1] [feature:quality]

`/work` フローの Phase 3（Review APPROVE 後）と Phase 4（Auto-commit 前）の間に、自動コード洗練ステップを追加。

**改訂後フロー**:
```
Phase 2: 実装 → Phase 3: harness-review APPROVE → Phase 3.5: Auto-Refinement → Phase 4: Auto-commit
```

| Task | 内容 | Status |
|------|------|--------|
| 16.1.1 | `skills/work/SKILL.md` の Default Flow に Phase 3.5 Auto-Refinement を追加。デフォルト: `/simplify` 実行。`--deep-simplify`: `/simplify` 後に `code-simplifier` も実行。`--no-simplify`: スキップ | cc:完了 |
| 16.1.2 | `skills/work/references/execution-flow.md` に Phase 3.5 の詳細手順を追記: Review APPROVE 後に `/simplify` を Skill tool で呼び出し → 変更があれば差分確認 → Phase 4 へ | cc:完了 |
| 16.1.3 | `skills/work/references/auto-iteration.md` の Step 3.5 に `/simplify` 統合を反映: 全タスク完了時の harness-review 後に Auto-Refinement を実行 | cc:完了 |
| 16.1.4 | Options テーブルに `--deep-simplify`（`/simplify` + `code-simplifier` 両方実行）と `--no-simplify`（スキップ）を追加 | cc:完了 |
| 16.1.5 | `work-active.json` スキーマに `simplify_mode: "default" | "deep" | "skip"` フィールドを追加（Compaction 復元用） | cc:完了 |

### Phase 16.2: breezing に /batch 委任判断ロジック追加 [P1]

横展開パターン（「全ファイルの○○を変更」系タスク）を検出し、breezing Lead が `/batch` に委任する戦略を追加。

| Task | 内容 | Status |
|------|------|--------|
| 16.2.1 | `skills/breezing/references/execution-flow.md` の Phase A に横展開パターン検出ロジックを追加: Plans.md タスクが「migrate」「replace all」「add ... to all」等のパターンを含み、かつ単一の均質な変更である場合に `/batch` 委任を提案 | cc:完了 |
| 16.2.2 | `skills/breezing/references/execution-flow.md` に `/batch` 委任時の Phase B 代替フローを追記: Lead が `/batch <instruction>` を Skill tool で呼び出し → `/batch` が worktree + PR を自動処理 → Lead は PR リストを breezing-active.json に記録 → Phase C は PR マージ確認に変更 | cc:完了 |
| 16.2.3 | `skills/breezing/SKILL.md` の Quick Reference と Feature Details に `/batch` 委任の説明を追加。`/batch` との関係性（横展開 vs 縦展開）を明記 | cc:完了 |

### Phase 16.3: feature-table v2.1.63 更新 [P1]

| Task | 内容 | Status |
|------|------|--------|
| 16.3.1 | `docs/CLAUDE-feature-table.md` に v2.1.52〜v2.1.63 の新機能を追加: `/simplify`（Phase 3.5 統合）、`/batch`（breezing 委任）、`code-simplifier` プラグイン（`--deep-simplify`）、HTTP hooks、auto-memory worktree 共有、`/clear` スキルキャッシュリセット、`ENABLE_CLAUDEAI_MCP_SERVERS=false` | cc:完了 |
| 16.3.2 | `docs/CLAUDE-feature-table.md` の既存行「メモリリーク修正 (v2.1.50)」を v2.1.63 まで拡大。15+ の修正を反映 | cc:完了 |
| 16.3.3 | `CLAUDE.md` の Feature Table 要約（上位5機能）を更新: `/simplify` + `/batch` を追加 | cc:完了 |

### Phase 16.4: hooks-editing.md に HTTP hooks 仕様追記 [P2]

| Task | 内容 | Status |
|------|------|--------|
| 16.4.1 | `.claude/rules/hooks-editing.md` に HTTP hooks セクションを追加: `type: "http"` のフォーマット（`url`, `headers`, `allowedEnvVars`）、常に POST、2xx レスポンス仕様、command hook との差異表（ブロッキングは 2xx + JSON が必要、`async: true` 非対応、`/hooks` メニューから追加不可） | cc:完了 |
| 16.4.2 | `.claude/rules/hooks-editing.md` に HTTP hooks サンプルテンプレートを追加: Slack 通知、メトリクス収集、外部ダッシュボード更新の3例 | cc:完了 |

### Phase 16.5: CLAUDE.md + ドキュメント バージョン表記更新 [P2]

| Task | 内容 | Status |
|------|------|--------|
| 16.5.1 | `CLAUDE.md` のバージョン表記を `2.1.51+` → `2.1.63+` に更新 | cc:完了 |
| 16.5.2 | `skills/breezing/references/guardrails-inheritance.md` に auto-memory worktree 共有の注記を追加（v2.1.63 で worktree エージェントがプロジェクト auto-memory にアクセス可能に） | cc:完了 |
| 16.5.3 | `skills/troubleshoot/SKILL.md` の診断手順に「`/clear` でスキルキャッシュをリセット」を追加（v2.1.63 で `/clear` がキャッシュされたスキルもリセットする動作に変更） | cc:完了 |

### Phase 16.6: 検証 + 同期

| Task | 内容 | Status |
|------|------|--------|
| 16.6.1 | `./tests/validate-plugin.sh && ./scripts/ci/check-consistency.sh` で構造検証 | cc:完了 |
| 16.6.2 | ミラー同期: `rsync -av --delete skills/ codex/.codex/skills/` + `opencode/skills/` | cc:完了 |
| 16.6.3 | CHANGELOG.md + CHANGELOG_ja.md エントリ追加 + バージョンバンプ | cc:完了 |

検証: /work SKILL.md の Phase 3.5 記述整合 / breezing の /batch 委任フロー整合 / feature-table 全行の正確性 / hooks-editing.md の HTTP hooks サンプル構文 / バージョン表記の一貫性 / validate-plugin + check-consistency 全パス

対象外: session-memory と auto-memory の関係（D22 で解決済み）/ command hooks から HTTP hooks への移行（既存フックは command 型を維持）/ `/simplify` のカスタマイズ（bundled のためバイナリ内蔵、変更不可）

リスク評価:

| リスク | 深刻度 | 軽減策 |
|--------|--------|--------|
| `/simplify` がサブエージェント内で呼べない可能性 | 中 | Phase 3.5 は Lead が直接呼ぶ設計（task-worker 内ではなく） |
| `code-simplifier` プラグイン未インストール環境 | 低 | `--deep-simplify` はオプション。未インストール時はスキップ + 案内表示 |
| `/batch` が Plans.md を認識しない | 低 | breezing Lead がタスク内容を自然言語に変換して `/batch` に渡す |
