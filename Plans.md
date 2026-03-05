# Claude Code Harness — Plans.md (v3 Rewrite Branch)

作成日: 2026-03-02
ブランチ: worktree-v3-full-rewrite

---

## Phase 17: Harness v3 — フルリライト（アーキテクチャ再設計）

作成日: 2026-03-02
起点: 現行アーキテクチャの構造的限界に対する再設計議論
目的: テスト可能・保守可能・拡張可能なアーキテクチャへの全面移行

### 設計原則

1. **プラグインは薄い接着剤** — ロジックをBashに書かない。TypeScriptで型安全に
2. **宣言的ルール** — ガードレールは条件→アクションのルールテーブル
3. **状態は1箇所** — SQLite 1ファイルに統合。ファイル散在を排除
4. **5動詞スキル** — plan / execute / review / release / setup
5. **シンボリックリンク** — ミラーはrsyncではなくリンク

### 優先度マトリクス

| 優先度 | Phase | 内容 | タスク数 | 依存 |
|--------|-------|------|---------|------|
| **Required** | 17.0 | v3ブランチ + TS基盤構築 | 5 | なし |
| **Required** | 17.1 | ガードレールエンジン（Bash→TS） | 8 | 17.0 |
| **Required** | 17.2 | SQLite状態管理 | 6 | 17.0 |
| **Required** | 17.3 | スキル統合 42→5 + 拡張パック | 9 | 17.0 |
| **Required** | 17.4 | ミラー廃止（rsync→symlink） | 4 | 17.3 |
| **Recommended** | 17.5 | エージェント統合 11→3 | 5 | 17.3 |
| **Recommended** | 17.6 | リポジトリ整理（80%ドキュメント削減） | 5 | なし |
| **Required** | 17.7 | テスト + 検証 + カットオーバー | 6 | 17.1, 17.2, 17.3, 17.4 |

合計: **48 タスク**

---

### Phase 17.0: v3ブランチ + TypeScript基盤構築 [P1]

| Task | 内容 | Status |
|------|------|--------|
| 17.0.1 | v3ブランチ確認（worktree-v3-full-rewrite で作業中） | cc:完了 |
| 17.0.2 | `core/` ディレクトリ作成。`package.json`（`better-sqlite3`, `tsx`, `vitest` を devDependencies）、`tsconfig.json`（strict, ESM, NodeNext）を配置 | cc:完了 |
| 17.0.3 | `core/index.ts` エントリポイント作成。stdin JSON → パース → ルーティング → stdout JSON の基本パイプライン | cc:完了 |
| 17.0.4 | `core/types.ts` 作成。`HookInput`, `HookResult`, `GuardRule`, `Signal`, `TaskFailure` の型定義 | cc:完了 |
| 17.0.5 | CI（`.github/workflows/`）に `npm test`（vitest）ステップを追加 | cc:完了 |

### Phase 17.1: ガードレールエンジン — Bash→TypeScript [P1] [P]

| Task | 内容 | Status |
|------|------|--------|
| 17.1.1 | `core/guardrails/rules.ts` 作成。宣言的ルールテーブル。pretooluse-guard.sh の全ルール移植 | cc:完了 |
| 17.1.2 | `core/guardrails/pre-tool.ts` 作成。`evaluate(input): HookResult` 関数 | cc:完了 |
| 17.1.3 | `core/guardrails/tampering.ts` 作成。tampering-detector の全検出パターン移植 | cc:完了 |
| 17.1.4 | `core/guardrails/post-tool.ts` 作成。9スクリプト → Promise.allSettled 統合 | cc:完了 |
| 17.1.5 | `core/guardrails/permission.ts` 作成。permission-request.sh 移植 | cc:完了 |
| 17.1.6 | `hooks/pre-tool.sh` 薄いシム作成（5行以内） | cc:完了 |
| 17.1.7 | `hooks/post-tool.sh` 薄いシム作成 + hooks.json 差し替え | cc:完了 |
| 17.1.8 | `core/guardrails/__tests__/rules.test.ts` 単体テスト（カバレッジ90%+） | cc:完了 |

### Phase 17.2: SQLite状態管理 [P1] [P]

| Task | 内容 | Status |
|------|------|--------|
| 17.2.1 | `core/state/schema.ts` 作成。テーブル定義 | cc:完了 |
| 17.2.2 | `core/state/store.ts` 作成。better-sqlite3 ラッパー | cc:完了 |
| 17.2.3 | `core/state/migration.ts` 作成。JSON/JSONL→SQLite移行 | cc:完了 |
| 17.2.4 | `core/state/__tests__/store.test.ts` 単体テスト | cc:完了 |
| 17.2.5 | guardrails のJSONスタブをSQLiteストアに差し替え | cc:完了 |
| 17.2.6 | `hooks/session.sh` + `core/engine/lifecycle.ts` 作成 | cc:完了 |

### Phase 17.3: スキル統合 42→5 + 拡張パック分離 [P1]

| Task | 内容 | Status |
|------|------|--------|
| 17.3.1 | `skills-v3/plan/SKILL.md` 作成（planning + plans-management + sync-status 統合） | cc:完了 |
| 17.3.2 | `skills-v3/execute/SKILL.md` 作成（work + impl + breezing + parallel + ci 統合） | cc:完了 |
| 17.3.3 | `skills-v3/review/SKILL.md` 作成（harness-review + codex-review + verify + troubleshoot 統合） | cc:完了 |
| 17.3.4 | `skills-v3/release/SKILL.md` 作成（release-har + x-release-harness + handoff 統合） | cc:完了 |
| 17.3.5 | `skills-v3/setup/SKILL.md` 作成（setup + harness-init + harness-update + maintenance 統合） | cc:完了 |
| 17.3.6 | `skills-v3/extensions/` に拡張パック移動（auth, crud, ui 等 11スキル） | cc:完了 |
| 17.3.7 | `core/engine/lifecycle.ts` 作成（session系5スキル吸収） | cc:完了 |
| 17.3.8 | `skills-v3/routing-rules.md` 作成（5エントリ） | cc:完了 |
| 17.3.9 | CLAUDE.md にガイダンス統合（vibecoder-guide, workflow-guide, principles） | cc:完了 |

### Phase 17.4: ミラー廃止 — rsync→シンボリックリンク [P1]

| Task | 内容 | Status |
|------|------|--------|
| 17.4.1 | `codex/.codex/skills/` → シンボリックリンクに置換 | cc:完了 |
| 17.4.2 | `opencode/skills/`, `.opencode/skills/` → シンボリックリンクに置換 | cc:完了 |
| 17.4.3 | `check-consistency.sh` のミラーチェック → symlink チェックに更新 | cc:完了 |
| 17.4.4 | rsync 参照をすべて削除・更新 | cc:完了 |

### Phase 17.5: エージェント統合 11→3 [P2]

| Task | 内容 | Status |
|------|------|--------|
| 17.5.1 | `agents-v3/worker.md` 作成（task-worker + codex-implementer + error-recovery 統合） | cc:完了 |
| 17.5.2 | `agents-v3/reviewer.md` 作成（code-reviewer + plan-critic + plan-analyst 統合） | cc:完了 |
| 17.5.3 | `agents-v3/scaffolder.md` 作成（project-analyzer + project-scaffolder + project-state-updater 統合） | cc:完了 |
| 17.5.4 | team-composition.md を3エージェント構成に更新 | cc:完了 |
| 17.5.5 | `.claude/agent-memory/` を3エージェントに再編 | cc:完了 |

### Phase 17.6: リポジトリ整理 [P2] [P]

| Task | 内容 | Status |
|------|------|--------|
| 17.6.1 | `commands/` ディレクトリ全体を削除 | cc:完了 |
| 17.6.2 | `docs/` を精選（残す4件、アーカイブ、削除） | cc:完了 |
| 17.6.3 | `CHANGELOG_ja.md` を削除（英語版に一本化） | cc:完了 |
| 17.6.4 | `benchmarks/evals-v2/`, `evals-v3/` を削除 | cc:完了 |
| 17.6.5 | プラグイン外コード分離（mcp-server/, profiles/ を削除。workflows/, templates/ は残す） | cc:完了 |

### Phase 17.7: テスト・検証・カットオーバー [P1]

| Task | 内容 | Status |
|------|------|--------|
| 17.7.1 | `core/guardrails/__tests__/integration.test.ts` E2Eテスト | cc:完了 |
| 17.7.2 | `core/state/__tests__/migration.test.ts` 移行テスト | cc:完了 |
| 17.7.3 | `tests/validate-plugin-v3.sh` v3バリデータ | cc:完了 |
| 17.7.4 | breezing-bench v2 vs v3 比較ベンチマーク | cc:完了 |
| 17.7.5 | VERSION 3.0.0 バンプ + CHANGELOG + plugin.json | cc:完了 |
| 17.7.6 | main マージ + GitHub Release | cc:完了 |

---

## Phase 18: Codex CLI 0.107.0 対応 + README ビジュアル改善

作成日: 2026-03-03
起点: Codex CLI 0.107.0 リリース（2026-03-02）+ README 訴求力向上要件
目的: Codex CLI 上での Harness 使用時の互換性・安全性を確保し、README の視覚的訴求力を向上

### 背景

- Codex CLI が 0.104.0 → 0.107.0 に更新（thread forking, 設定可能メモリ, sandbox 強化）
- Harness の Codex 統合コード（`codex/.codex/`, `scripts/codex/`, `setup-codex.sh`）に廃止済み MCP 残骸・並列競合リスクあり
- README に Nano Banana Pro 生成画像を追加し、機能差分の直感的理解を向上

### Phase 18.0: README ビジュアル改善 [P1]

| Task | 内容 | Status |
|------|------|--------|
| 18.0.1 | Nano Banana Pro で3枚の画像生成（hero-comparison, core-loop, safety-guardrails） | cc:完了 |
| 18.0.2 | README.md にブランド T&M に沿った画像配置 + セクション構造改善 | cc:完了 |
| 18.0.3 | ロゴファイル `docs/images/claude-harness-logo-with-text.png` 修復 | cc:完了 |

### Phase 18.1: MCP 残骸除去（High） [P1] [P]

| Task | 内容 | Status |
|------|------|--------|
| 18.1.1 | `codex/.codex/config.toml` から `[mcp_servers.harness]` セクション削除 | cc:完了 |
| 18.1.2 | `scripts/setup-codex.sh` から `--with-mcp` フラグ + `setup_mcp_template()` 関数を削除 | cc:完了 |
| 18.1.3 | `scripts/codex-worker-engine.sh` の `mcp-params.json` → `codex-exec-params.json` にリネーム | cc:完了 |
| 18.1.R1 | `scripts/codex-setup-local.sh` の MCP 残骸除去（Reviewer 指摘） | cc:完了 |
| 18.1.R2 | `--skip-mcp` 残存参照の一掃（README, codex/README, tests） | cc:完了 |

### Phase 18.2: 並列実行安全性（High） [P1]

| Task | 内容 | Status |
|------|------|--------|
| 18.2.1 | `skills-v3/harness-work/SKILL.md` の `/tmp/codex-prompt.md` 固定パス → `mktemp` 一意パスに変更 | cc:完了 |
| 18.2.2 | `codex exec` 呼び出しに `-a never -s workspace-write` フラグ明示（正式フラグ名に修正） | cc:完了 |
| 18.2.3 | `2>/dev/null` のエラー握りつぶし → ログファイルへのリダイレクト（`2>>/tmp/harness-codex-$$.log`） | cc:完了 |
| 18.2.R1 | `codex-cli-only.md` と README の固定パス・旧フラグ名修正（Reviewer 指摘） | cc:完了 |

### Phase 18.3: Codex 環境での Harness スキル互換性（Medium） [P2]

| Task | 内容 | Status |
|------|------|--------|
| 18.3.1 | `skills-v3/harness-review/SKILL.md` に Codex 環境での代替フロー記載（Task ツール非対応時のフォールバック） | cc:完了 |
| 18.3.2 | `agents-v3/team-composition.md` に Codex 環境の注記追加（`bypassPermissions` → `-a never`） | cc:完了 |
| 18.3.3 | `codex/.codex/config.toml` に `[notify]` セクション追加（after_agent → メモリブリッジ接続） | cc:完了 |
| 18.3.4 | `codex/.codex/config.toml` の reviewer エージェントに Read-only sandbox 制限追加 | cc:完了 |

### Phase 18.4: Codex 0.107.0 新機能活用（Medium） [P2]

| Task | 内容 | Status |
|------|------|--------|
| 18.4.1 | Thread forking 活用検討: 時期尚早（`codex exec fork` は未実装、Issue #11750 提案段階） | cc:完了 |
| 18.4.2 | 設定可能メモリ: `memory: project` の Codex 側マッピング定義を team-composition.md に記載 | cc:完了 |
| 18.4.3 | stdin パイプ方式に改善（`cat file \| codex exec -`）、`--input-file` は存在せず | cc:完了 |

### Phase 18.5: 品質改善（Low） [P3]

| Task | 内容 | Status |
|------|------|--------|
| 18.5.1 | `codex-exec-wrapper.sh` の構造化出力調査: `--output-schema` 将来移行可能、現状マーカー方式維持 | cc:完了 |
| 18.5.2 | `worker.md` に Codex 環境での `memory`/`skills` フィールドの非互換に関する注記追加 | cc:完了 |
| 18.5.3 | `codex/.codex/skills/` の CLAUDE.md ノイズ化対策（.codexignore 追加 + ルート CLAUDE.md 削除） | cc:完了 |
| 18.5.4 | README_ja.md にも同等のビジュアル改善を反映 | cc:完了 |
| 18.5.5 | CHANGELOG.md に Phase 18 の変更を追記（[3.1.0] - 2026-03-03） | cc:完了 |

---

## Phase 19: Claude Code v2.1.68 対応 + Feature Table 活用機能の実装

作成日: 2026-03-05
起点: Claude Code v2.1.63→v2.1.68 の新機能（effort levels, agent hooks, voice mode 等）+ Feature Table「将来対応」の実装格上げ
目的: Harness を最新 Claude Code に最適化し、未活用の公式機能を実装に移行

### 背景

- Claude Code v2.1.68 で Opus 4.6 の **medium effort デフォルト化** + **ultrathink キーワード再導入**
- Opus 4/4.1 が first-party API から削除（自動的に Opus 4.6 に移行）
- 公式 Hooks ドキュメントに `type: "agent"` フック（LLM エージェントベースのフック）が登場
- `type: "prompt"` フックが全イベントで利用可能に（Harness ルールでは Stop/SubagentStop 限定と誤記載）
- Voice mode (`/voice`) がローリングアウト開始
- Feature Table の「将来対応」3件（WorktreeRemove, remote-control, HTTP hooks 実用化）が実装可能な段階に

### 設計判断（3エージェントレビューにより確定）

1. **effort 判定は多要素スコアリング** — ファイル数だけでなく、対象ディレクトリ（core/, guardrails/）、タスクキーワード（security, architecture, design）、agent memory の失敗記録を組み合わせる
2. **breezing の effort 制御は harness-work に一本化** — breezing は harness-work の委譲エイリアスなので独自追加せず継承
3. **agent hooks はコスト上限を事前定義** — matcher で対象を絞り、1フック当たりの上限トークン・月間上限を定義。超過時は自動 rollback（command 型に戻す）
4. **hooks-editing.md の 3 タスク（19.1.1, 19.1.2, 旧 19.2.4）は 1 タスクに統合** — 同一ファイルの連続編集による中途半端状態を防止
5. **調査ファースト** — remote-control 調査（19.3.4）を Feature Table 更新（19.2）より先に実施し、結果を反映
6. **hooks.json 編集は直列化** — agent hook（19.1）、Worktree hooks（19.3）、HTTP hooks（19.4）は並列不可

### 優先度マトリクス

| 優先度 | Phase | 内容 | タスク数 | 依存 |
|--------|-------|------|---------|------|
| **Required** | 19.0 | Effort レベル制御（多要素スコアリング） | 5 | なし |
| **Required** | 19.1 | Agent hooks 対応（ルール整備 + プロトタイプ + 検証） | 6 | なし |
| **Recommended** | 19.2 | Feature Table「将来対応」実装格上げ + 調査 | 5 | なし |
| **Required** | 19.3 | ドキュメント・Feature Table 統合更新 | 4 | 19.0, 19.1, 19.2 |
| **Recommended** | 19.4 | 既存機能の活用強化 | 5 | 19.2 |
| **Required** | 19.5 | バージョン・品質・リリース | 4 | 19.0〜19.4 |

合計: **29 タスク**

---

### Phase 19.0: Opus 4.6 Effort レベル制御 [P0] [P]

| Task | 内容 | Status |
|------|------|--------|
| 19.0.1 | `skills-v3/harness-work/SKILL.md` に多要素 effort 判定ロジック追加 | cc:完了 |
| 19.0.2 | `opencode/commands/pm/` 配下の既存 `ultrathink` 使用を体系化 | cc:完了 |
| 19.0.3 | `agents-v3/worker.md` に effort 制御セクション追加 | cc:完了 |
| 19.0.4 | `agents-v3/reviewer.md` に effort 制御セクション追加 | cc:完了 |
| 19.0.5 | `agents-v3/team-composition.md` に v2.1.68 effort 変更の影響を追記 | cc:完了 |

### Phase 19.1: Agent hooks 対応 + Prompt/Agent type ルール整備 [P1]

| Task | 内容 | Status |
|------|------|--------|
| 19.1.1 | `.claude/rules/hooks-editing.md` 統合更新（4タイプ体系、prompt全イベント対応修正） | cc:完了 |
| 19.1.2 | agent hook 移行候補の特定と設計（rules.ts 分析） | cc:完了 |
| 19.1.3 | hooks.json に agent hook プロトタイプ追加（PreToolUse + Stop） | cc:完了 |
| 19.1.4 | PostToolUse agent hook 追加（軽量自動コードレビュー） | cc:完了 |
| 19.1.5 | agent hook 動作検証 + コスト実測 | cc:完了 |
| 19.1.6 | 検証結果に基づく agent hook 最終判断 → D28 記録 | cc:完了 |

### Phase 19.2: Feature Table「将来対応」実装格上げ + 調査 [P2] [P]

| Task | 内容 | Status |
|------|------|--------|
| 19.2.1 | `worktree-create.sh` 新規作成（worktree 環境初期化） | cc:完了 |
| 19.2.2 | `worktree-remove.sh` 新規作成（worktree クリーンアップ） | cc:完了 |
| 19.2.3 | hooks.json に WorktreeCreate/Remove 両イベント登録 | cc:完了 |
| 19.2.4 | `claude remote-control` 調査 → Research Preview、Breezing 不適合と判定 | cc:完了 |
| 19.2.5 | remote-control 実装スキップ（19.2.4 結果: 将来対応維持） | cc:完了 |

### Phase 19.3: ドキュメント・Feature Table 統合更新 [P3]

| Task | 内容 | Status |
|------|------|--------|
| 19.3.1 | CLAUDE.md Feature Table を 2.1.68+ に更新、新機能行追加 | cc:完了 |
| 19.3.2 | docs/CLAUDE-feature-table.md 統合更新（新機能・将来対応・参照修正） | cc:完了 |
| 19.3.3 | decisions.md 更新（D15 修正 + D27 新規追加） | cc:完了 |
| 19.3.4 | README.md + README_ja.md に Feature Table セクション新設 | cc:完了 |

### Phase 19.4: 既存機能の活用強化 [P4] [P]

| Task | 内容 | Status |
|------|------|--------|
| 19.4.1 | PostToolUse HTTP hook（metrics 収集テンプレート）追加 | cc:完了 |
| 19.4.2 | PreCompact agent hook（WIP タスク警告）追加 | cc:完了 |
| 19.4.3 | session-env-setup.sh 新規作成 + SessionStart 登録 | cc:完了 |
| 19.4.4 | Auto-memory worktree 共有テスト（手動検証が必要） | cc:完了 |
| 19.4.5 | hooks-editing.md タイムアウトガイドライン更新 | cc:完了 |

### Phase 19.5: バージョン・品質・リリース [P5]

| Task | 内容 | Status |
|------|------|--------|
| 19.5.1 | validate-plugin.sh + check-consistency.sh 全体検証 | cc:完了 |
| 19.5.2 | VERSION バンプ 3.2.0 → 3.3.0 + plugin.json 同期 | cc:完了 |
| 19.5.3 | CHANGELOG.md に [3.3.0] - 2026-03-05 追記 | cc:完了 |
| 19.5.4 | GitHub Release 作成 | cc:完了 |

---

## Phase 20: Claude Code v2.1.69 対応 + Codex 0.110.0 統合

作成日: 2026-03-06
起点: Claude Code v2.1.68→v2.1.69 リリース（多数の新機能・破壊的変更）+ Codex CLI 0.110.0 アップデート
目的: Harness の全コンポーネントを最新 Claude Code / Codex CLI に最適化し、新機能を即座に活用可能にする

### 背景

**Claude Code v2.1.69 の主要変更**:
- `${CLAUDE_SKILL_DIR}` 変数導入（スキル内の相対パス参照を公式変数に移行）
- `InstructionsLoaded` フックイベント新設
- フックイベントに `agent_id` / `agent_type` フィールド追加
- `TeammateIdle` / `TaskCompleted` で `{"continue": false, "stopReason": "..."}` をサポート
- `WorktreeCreate` / `WorktreeRemove` フックの動作修正（以前は silently ignored）
- `/reload-plugins` コマンド追加
- `includeGitInstructions: false` 設定でトークン節約
- `git-subdir` プラグインソース対応
- Sonnet 4.5 → 4.6 自動マイグレーション
- nested teammates 防止（teammates が更に teammate を spawn するのをブロック）
- 複数セキュリティ修正（symlink bypass, nested skill discovery, interactive tools auto-allow）

**Codex CLI 0.110.0 の主要変更**（v3.2.0 で一部対応済み）:
- `[[skills.config]]` path-based skill loading
- memory config renames (`no_memories_if_mcp_or_web_search`)
- workspace-scoped memory writes
- polluted memories flag

### 3エージェント監査結果（計画前調査）

| Track | 対象 | 発見 |
|-------|------|------|
| Hooks | hooks.json, pretooluse-guard.sh | InstructionsLoaded: 参照0件（完全新規）、agent_id/agent_type: 未使用、TeammateIdle/TaskCompleted: teammate_name ベースで agent_id 未活用 |
| Skills | 17 SKILL.md files | `references/` パスを `${CLAUDE_SKILL_DIR}/references/` に変更必要。description 欠損: 0件 |
| Docs | 8 files, 17 箇所 | "2.1.68" → "2.1.69" 更新必要。Sonnet 4.5 参照: 0件（対応済み）。includeGitInstructions: 参照0件 |

### 優先度マトリクス

| 優先度 | Phase | 内容 | タスク数 | 依存 |
|--------|-------|------|---------|------|
| **Required** | 20.0 | Hooks & ガードレール更新 | 6 | なし |
| **Required** | 20.1 | Skills `${CLAUDE_SKILL_DIR}` 対応 | 4 | なし |
| **Required** | 20.2 | ドキュメント・Feature Table 更新 | 5 | なし |
| **Recommended** | 20.3 | Breezing・Plugin・チーム構成更新 | 5 | 20.0 |
| **Required** | 20.4 | 統合検証・バージョン・リリース | 4 | 20.0〜20.3 |

合計: **24 タスク**

---

### Phase 20.0: Hooks & ガードレール更新 [P0]

| Task | 内容 | Status |
|------|------|--------|
| 20.0.1 | `hooks.json` + `.claude-plugin/hooks.json`: TeammateIdle / TaskCompleted ハンドラに `{"continue": false, "stopReason": "..."}` レスポンス形式サポートを追加。現在 teammate_name のみ使用 → agent_id / agent_type も活用 | cc:完了 |
| 20.0.2 | `hooks.json` + `.claude-plugin/hooks.json`: `InstructionsLoaded` フックイベント登録。用途: Harness ルール注入・セッション環境の事前検証 | cc:完了 |
| 20.0.3 | `scripts/pretooluse-guard.sh`: `agent_id` / `agent_type` フィールド活用。現在 session_id のみ → エージェント種別に応じたガードレール分岐を追加 | cc:完了 |
| 20.0.4 | `scripts/hook-handlers/teammate-idle.sh` + `task-completed.sh`: `{"continue": false}` 判定ロジック追加（全タスク完了時やエラー時に自動停止） | cc:完了 |
| 20.0.5 | WorktreeCreate / WorktreeRemove ハンドラの動作確認。2.1.69 で silently ignored → 正常発火に修正済みのため、既存スクリプトが期待通り動作するか検証 | cc:完了 |
| 20.0.6 | `.claude/rules/hooks-editing.md` 更新: InstructionsLoaded イベント追加、agent_id/agent_type フィールド仕様追記、`{"continue": false}` パターン追記 | cc:完了 |

### Phase 20.1: Skills `${CLAUDE_SKILL_DIR}` 対応 [P1] [P]

| Task | 内容 | Status |
|------|------|--------|
| 20.1.1 | 全 17 SKILL.md の `references/` パスを `${CLAUDE_SKILL_DIR}/references/` に変更（対象: skills/ 配下の本体） | cc:完了 |
| 20.1.2 | symlink 先の skills-v3/ 配下も同様に更新（codex/.codex/skills/, opencode/skills/ は symlink なので自動反映を確認） | cc:完了 |
| 20.1.3 | 全スキルの `description:` frontmatter に colon 含みの値が正しくクォートされているか確認（2.1.69 のパーサー変更対応） | cc:完了 |
| 20.1.4 | `.claude/rules/skill-editing.md` 更新: `${CLAUDE_SKILL_DIR}` 変数の使用ガイドライン追記、references パスの標準テンプレート更新 | cc:完了 |

### Phase 20.2: ドキュメント・Feature Table・モデル参照更新 [P2] [P]

| Task | 内容 | Status |
|------|------|--------|
| 20.2.1 | `CLAUDE.md` の Feature Table を 2.1.69+ に更新。新機能行追加: `${CLAUDE_SKILL_DIR}`, InstructionsLoaded, agent_id/agent_type, continue:false, /reload-plugins, includeGitInstructions, git-subdir source | cc:完了 |
| 20.2.2 | `docs/CLAUDE-feature-table.md` 統合更新: 8ファイル17箇所の "2.1.68" → "2.1.69" 一括更新 + 新機能行追加 | cc:完了 |
| 20.2.3 | モデル参照の確認: Sonnet 4.5 → 4.6 自動マイグレーションに関する注記追加（既に参照は更新済みだが、ドキュメント上で明記） | cc:完了 |
| 20.2.4 | `includeGitInstructions: false` 設定の活用検討・ドキュメント化。Breezing Worker で git instructions 不要なケースを特定 | cc:完了 |
| 20.2.5 | `.claude/memory/decisions.md` 更新: D29 として 2.1.69 対応の設計判断を記録 | cc:完了 |

### Phase 20.3: Breezing・Plugin・チーム構成更新 [P3]

| Task | 内容 | Status |
|------|------|--------|
| 20.3.1 | `agents-v3/team-composition.md`: nested teammates 防止の spawn プロンプト更新。2.1.69 で公式にブロックされたため、Harness 側の防止策を公式仕様に合わせて簡素化 | cc:完了 |
| 20.3.2 | `skills-v3/execute/SKILL.md` (harness-work): Breezing モードの spawn プロンプトから冗長な nested teammate 防止指示を削除（2.1.69 で公式対応済み） | cc:完了 |
| 20.3.3 | `git-subdir` プラグインソース対応: `.claude-plugin/plugin.json` に source 設定追加の検討・ドキュメント化 | cc:完了 |
| 20.3.4 | `/reload-plugins` コマンドの Harness 開発ワークフローへの統合: スキル編集後の即時反映手順をドキュメント化 | cc:完了 |
| 20.3.5 | セキュリティ修正の影響確認: symlink bypass 修正が skills/ → skills-v3/ の symlink 構成に影響ないか検証 | cc:完了 |
| 20.3.6 | `scripts/ci/check-consistency.sh`: Auto Mode（Research Preview）向けに `defaultMode=autoMode` を許容する互換チェックを追加 | cc:完了 |

### Phase 20.4: 統合検証・バージョン・リリース [P4]

| Task | 内容 | Status |
|------|------|--------|
| 20.4.1 | `tests/validate-plugin.sh` + `scripts/ci/check-consistency.sh` 全体検証 | cc:完了 |
| 20.4.2 | VERSION バンプ + plugin.json 同期（3.3.1 → 3.4.0） | cc:完了 |
| 20.4.3 | CHANGELOG.md に [3.4.0] - 2026-03-06 追記 | cc:完了 |
| 20.4.4 | GitHub Release 作成 + X 告知文生成（前セッションで作成済みの画像を活用） | cc:TODO |
