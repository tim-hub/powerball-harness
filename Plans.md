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
| 17.0.2 | `core/` ディレクトリ作成。`package.json`（`better-sqlite3`, `tsx`, `vitest` を devDependencies）、`tsconfig.json`（strict, ESM, NodeNext）を配置 | cc:TODO |
| 17.0.3 | `core/index.ts` エントリポイント作成。stdin JSON → パース → ルーティング → stdout JSON の基本パイプライン | cc:TODO |
| 17.0.4 | `core/types.ts` 作成。`HookInput`, `HookResult`, `GuardRule`, `Signal`, `TaskFailure` の型定義 | cc:TODO |
| 17.0.5 | CI（`.github/workflows/`）に `npm test`（vitest）ステップを追加 | cc:TODO |

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
| 17.2.6 | `hooks/session.sh` + `core/engine/lifecycle.ts` 作成 | cc:TODO |

### Phase 17.3: スキル統合 42→5 + 拡張パック分離 [P1]

| Task | 内容 | Status |
|------|------|--------|
| 17.3.1 | `skills-v3/plan/SKILL.md` 作成（planning + plans-management + sync-status 統合） | cc:完了 |
| 17.3.2 | `skills-v3/execute/SKILL.md` 作成（work + impl + breezing + parallel + ci 統合） | cc:完了 |
| 17.3.3 | `skills-v3/review/SKILL.md` 作成（harness-review + codex-review + verify + troubleshoot 統合） | cc:完了 |
| 17.3.4 | `skills-v3/release/SKILL.md` 作成（release-har + x-release-harness + handoff 統合） | cc:完了 |
| 17.3.5 | `skills-v3/setup/SKILL.md` 作成（setup + harness-init + harness-update + maintenance 統合） | cc:完了 |
| 17.3.6 | `skills-v3/extensions/` に拡張パック移動（auth, crud, ui 等 11スキル） | cc:完了 |
| 17.3.7 | `core/engine/lifecycle.ts` 作成（session系5スキル吸収） | cc:TODO |
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
| 17.7.4 | breezing-bench v2 vs v3 比較ベンチマーク | cc:TODO |
| 17.7.5 | VERSION 3.0.0 バンプ + CHANGELOG + plugin.json | cc:完了 |
| 17.7.6 | main マージ + GitHub Release | cc:TODO |
