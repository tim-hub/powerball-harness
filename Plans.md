# スキル統合プラン — 28 → 19 スキルへ

## 背景

v2.19.0 で実装コマンドを 5→2 に統合した。次はスキル全体の整理。
ユーザーに見えるスキル数を 28 → 19 に削減し、認知負荷を下げる。

---

## Phase 1: memory 統合 (3→1)

`/memory` に `/sync-ssot-from-memory` と `/cursor-mem` を吸収。

### 1.1 memory SKILL.md に統合機能を追加

| Task | 内容 | Status |
|------|------|--------|
| 1.1.1 | `/memory` SKILL.md の description に sync-ssot, cursor-mem のトリガーフレーズを追加 | ✅ |
| 1.1.2 | SKILL.md 本文に「SSOT 昇格」と「記憶検索」セクションを追加 | ✅ |
| 1.1.3 | sync-ssot-from-memory の処理ロジックを `references/sync-ssot.md` として移設 | ✅ |
| 1.1.4 | cursor-mem の処理ロジックを `references/cursor-mem-search.md` として移設 | ✅ |

### 1.2 旧スキルのアーカイブ

| Task | 内容 | Status |
|------|------|--------|
| 1.2.1 | `skills/sync-ssot-from-memory/` → `skills/_archived/sync-ssot-from-memory/` に移動 | ✅ |
| 1.2.2 | `skills/cursor-mem/` → `skills/_archived/cursor-mem/` に移動 | ✅ |

---

## Phase 2: setup 統合 (5→1)

`/setup` に `/harness-mem`, `/codex-setup`, `/2agent`, `/localize-rules` を吸収。
`/setup-tools` をベースに、サブコマンド的に分岐する構成。

### 2.1 setup SKILL.md の拡張

| Task | 内容 | Status |
|------|------|--------|
| 2.1.1 | `/setup-tools` SKILL.md の description に統合対象のトリガーフレーズを追加 | ✅ |
| 2.1.2 | SKILL.md 本文にルーティングテーブル追加（ユーザー意図 → 適切な reference へ分岐） | ✅ |
| 2.1.3 | harness-mem の処理ロジックを `references/harness-mem.md` として移設 | ✅ |
| 2.1.4 | codex-setup の処理ロジックを `references/codex-setup.md` として移設 | ✅ |
| 2.1.5 | 2agent の SKILL.md + references/ を `references/2agent-setup.md` + `references/2agent/` として移設 | ✅ |
| 2.1.6 | localize-rules の処理ロジックを `references/localize-rules.md` として移設 | ✅ |

### 2.2 スキル名変更

| Task | 内容 | Status |
|------|------|--------|
| 2.2.1 | `skills/setup-tools/` → `skills/setup/` にリネーム（name: setup に変更） | ✅ |

### 2.3 旧スキルのアーカイブ

| Task | 内容 | Status |
|------|------|--------|
| 2.3.1 | `skills/harness-mem/` → `skills/_archived/harness-mem/` に移動 | ✅ |
| 2.3.2 | `skills/codex-setup/` → `skills/_archived/codex-setup/` に移動 | ✅ |
| 2.3.3 | `skills/2agent/` → `skills/_archived/2agent/` に移動 | ✅ |
| 2.3.4 | `skills/localize-rules/` → `skills/_archived/localize-rules/` に移動 | ✅ |

---

## Phase 3: 非表示化 (3スキル)

`user-invocable: false` を設定。description のトリガーフレーズは維持し、
他スキルからの内部呼び出しは引き続き可能にする。

| Task | 内容 | Status |
|------|------|--------|
| 3.1 | `skills/x-release-harness/SKILL.md` に `user-invocable: false` 追加 | ✅ |
| 3.2 | `skills/ci/SKILL.md` に `user-invocable: false` 追加。`/troubleshoot` の description に「CIが落ちた」トリガーを追加し、内部で ci を呼ぶ導線を確保 | ✅ |
| 3.3 | `skills/agent-browser/SKILL.md` に `user-invocable: false` 追加。description のトリガーフレーズ（「ブラウザで操作」等）は維持し自動ロード経由のアクセスを確保 | ✅ |

---

## Phase 4: CLAUDE.md 更新 + ミラー同期

| Task | 内容 | Status |
|------|------|--------|
| 4.1 | CLAUDE.md のスキルカテゴリテーブルを更新（統合後の 19 スキル反映） | ✅ |
| 4.2 | CLAUDE.md のスキル階層構造ツリーを更新 | ✅ |
| 4.3 | ミラー同期 (`rsync skills/ → codex/.codex/skills/, opencode/skills/, .opencode/skills/`) | ✅ |
| 4.4 | バージョンバンプ (v2.20.0) + CHANGELOG エントリ追加 | ✅ |
| 4.5 | `./tests/validate-plugin.sh && ./scripts/ci/check-consistency.sh` で検証 | ✅ |

---

## 検証方法

1. **構造検証**: `./tests/validate-plugin.sh && ./scripts/ci/check-consistency.sh`
2. **統合後の動作**: `/memory sync`, `/memory search` で旧機能がルーティングされること
3. **setup ルーティング**: `/setup codex`, `/setup 2agent` 等で正しい reference にルーティング
4. **非表示確認**: スキルリストに ci, agent-browser, release-harness が出ないこと
5. **自動ロード確認**: 「CIが落ちた」→ troubleshoot 経由で ci にルーティングされること
6. **ミラー一致**: `diff -rq skills/ codex/.codex/skills/`

## Phase C: Codex レビュー修正ループ (R1-R10)

3エキスパート（Security, Quality, Architect）による Codex 並列レビュー → 修正 → 再レビューを10ラウンド実施。

| Task | 内容 | Status |
|------|------|--------|
| C.1 | Security エキスパート: Score A 達成（R5で達成、R10まで維持） | ✅ |
| C.2 | Quality エキスパート: consolidation スコープ High ゼロ達成（R10） | ✅ |
| C.3 | Architect エキスパート: consolidation スコープ High ゼロ達成（R10） | ✅ |
| C.4 | 累計修正: ~34ファイル、壊れたリンク・旧スキル名参照・コマンド名不一致を修正 | ✅ |
| C.5 | ミラー同期 + validate-plugin.sh + check-consistency.sh 全パス | ✅ |

---

## Phase 5: DEFER 項目（Codex レビューで検出された pre-existing 問題）

R1-R10 で検出されたが、consolidation スコープ外の pre-existing 問題。

### 5.1 Security 強化

| Task | 内容 | Status |
|------|------|--------|
| 5.1.1 | `pretooluse-guard.sh` symlink bypass 対策（realpath 検証追加） | ✅ |
| 5.1.2 | `permission-request.sh:58` npm/pnpm/yarn 自動承認をリポジトリ別 allowlist 方式に変更 | ✅ |
| 5.1.3 | `userprompt-track-command.sh:77` prompt_preview のパーミッション hardening (umask 077) | ✅ |
| 5.1.4 | `session-monitor.sh:275` resume_token の chmod 600 + umask 077 | ✅ |
| 5.1.5 | `pretooluse-guard.sh:354` eval を直接パース（jq/python）に置換 | ✅ |

### 5.2 ドキュメント・リンク修正

| Task | 内容 | Status |
|------|------|--------|
| 5.2.1 | `docs/QUALITY_GUARD_DESIGN.md` broken SSOT link 修正 | ✅ |
| 5.2.2 | `docs/PLAN_RULES_IMPROVEMENT.md` stale command refs 修正 | ✅ |
| 5.2.3 | `docs/plans/claude-mem-integration.md` stale paths 修正 | ✅ |
| 5.2.4 | `skills/workflow-guide/references/commands.md` path mismatch 修正 | ✅ |
| 5.2.5 | templates 内の `/skills-update` 参照を削除または更新 | ✅ |

### 5.3 generate-video メンテナンス

| Task | 内容 | Status |
|------|------|--------|
| 5.3.1 | `agents/video-scene-generator.md` Remotion paths 更新 | ✅ |
| 5.3.2 | `skills/generate-video/` references 内の Remotion paths 更新 | ✅ |
| 5.3.3 | `generate-video/src/schemas/*.ts` z.any() → z.unknown() / proper unions に修正 | ✅ |

### 5.4 Architecture: Hook オーケストレーター

| Task | 内容 | Status |
|------|------|--------|
| 5.4.1 | PostToolUse fan-out (9スクリプト) を単一 Node オーケストレーターに統合 | |
| 5.4.2 | stdin JSON パース共通化（scripts/lib/hook-input.js） | |

### 5.5 Architecture: State 管理

| Task | 内容 | Status |
|------|------|--------|
| 5.5.1 | `.claude/state/*.json` のスキーマ定義 + atomic write helper 導入 | |
| 5.5.2 | ロック戦略の統一（flock or advisory lock） | |

### 5.6 ビルド・ツーリング整理

| Task | 内容 | Status |
|------|------|--------|
| 5.6.1 | `check-checklist-sync.sh` empty gate logic 修正 | ✅ |
| 5.6.2 | `workflows/default/init.yaml` project-analyzer 参照修正 | ✅ |
| 5.6.3 | `build-opencode.js` commands/ 空ディレクトリ対応 | ✅ |
| 5.6.4 | `harness-ui` command catalog 空対応（統合後） | ✅ |
| 5.6.5 | `parse-work-flags.md` internal inconsistency 修正 | ✅ |

### 5.7 命名・ルーティング整理

| Task | 内容 | Status |
|------|------|--------|
| 5.7.1 | `/planning` → `/plan-with-agent` 完全統一（dual naming 解消） | ✅ |
| 5.7.2 | `verify` skill の `user-invocable` 整合性確認 | ✅ |
| 5.7.3 | setup と codex-review の Codex セットアップ重複整理 | ✅ |
| 5.7.4 | `_archived/` 配下からの dangling references 削除 | ✅ |

---

## Phase 6: リリースクリーンアップ (v2.20.0 再統合)

v2.20.1 の内容を v2.20.0 に統合し、リポジトリ品質を向上。

### 6.1 Breezing Teammate 権限修正

| Task | 内容 | Status |
|------|------|--------|
| 6.1.1 | Teammate の "prompts unavailable" 根本原因調査（公式仕様検証） | ✅ |
| 6.1.2 | `mode: "bypassPermissions"` + PreToolUse hooks 多層防御の採用決定 | ✅ |
| 6.1.3 | execution-flow.md, team-composition.md, codex-engine.md, guardrails-inheritance.md に反映 | ✅ |
| 6.1.4 | session-resilience.md のコンパクション回復処理にも反映 | ✅ |

### 6.2 Breezing Phase A/B/C 分離

| Task | 内容 | Status |
|------|------|--------|
| 6.2.1 | Phase A (Pre-delegate): ユーザー権限維持でTeam初期化・spawn | ✅ |
| 6.2.2 | Phase B (Delegate): Lead は TaskCreate/TaskUpdate/SendMessage のみ | ✅ |
| 6.2.3 | Phase C (Post-delegate): delegate 解除→統合検証・コミット・クリーンアップ | ✅ |

### 6.3 英語リリース + gitignore クリーンアップ

| Task | 内容 | Status |
|------|------|--------|
| 6.3.1 | GitHub リリースノートを英語に統一（ルール・スキル更新） | ✅ |
| 6.3.2 | 不要ファイルの精査: ビルド成果物, 開発ドキュメント, ロックファイル | ✅ |
| 6.3.3 | .gitignore 更新 + 33ファイル untrack | ✅ |
| 6.3.4 | v2.20.1 を v2.20.0 に統合 (amend + force push) | ✅ |
| 6.3.5 | CHANGELOG.md / CHANGELOG_ja.md を v2.20.0 に統合更新 | ✅ |

---

## 対象外（今回は見送り）

- `/gogcli-ops` — 独立した外部ツール連携。統合先がない。使用頻度に応じて別途判断
- `/deploy` — 高インパクト操作。明示的なコマンドとして維持（Codex も非表示に反対）

---

## Phase 8: Issue #40 — PostToolUse Hook 修正

GitHub Issue #40: `posttooluse-tampering-detector.sh` の bash パーサーエラー修正 + 全スクリプト python3 フォールバック改善。

### 8.1 tampering-detector 本体修正 (Phase 1+3+4)

| Task | 内容 | Status |
|------|------|--------|
| 8.1.1 | `set -euo pipefail` → `set +e` (他 PostToolUse スクリプトと統一) | cc:done |
| 8.1.2 | 行39 `\|\| true` 削除 (syntax error 根本原因) | cc:done |
| 8.1.3 | `echo \| grep -qE` → `[[ =~ ]]` (6箇所、パフォーマンス改善) | cc:done |
| 8.1.4 | eval + shlex.quote 維持、python3 -c 方式に変更 (Reviewer 判断) | cc:done |
| 8.1.5 | `echo -e` → `printf '%b'` (POSIX 準拠) | cc:done |
| 8.1.6 | 警告メッセージのバイリンガル化 (日本語 + English) | cc:done |
| 8.1.7 | jq パス: `echo "$INPUT"` → `printf '%s' "$INPUT"` | cc:done |

### 8.2 python3 フォールバック修正 (Phase 2, 9スクリプト)

| Task | 内容 | Status |
|------|------|--------|
| 8.2.1 | `posttooluse-log-toolname.sh` python3 -c 方式に変更 | cc:done |
| 8.2.2 | `auto-test-runner.sh` python3 -c 方式に変更 | cc:done |
| 8.2.3 | `permission-request.sh` python3 -c 方式に変更 | cc:done |
| 8.2.4 | `skill-child-reminder.sh` python3 -c 方式に変更 | cc:done |
| 8.2.5 | `plans-watcher.sh` python3 -c 方式に変更 | cc:done |
| 8.2.6 | `auto-cleanup-hook.sh` python3 -c 方式に変更 | cc:done |
| 8.2.7 | `track-changes.sh` python3 -c 方式に変更 | cc:done |
| 8.2.8 | `posttooluse-security-review.sh` python3 -c 方式に変更 | cc:done |
| 8.2.9 | `posttooluse-quality-pack.sh` python3 -c 方式に変更 | cc:done |

### 8.3 検証 + ミラー同期 (Phase 5)

| Task | 内容 | Status |
|------|------|--------|
| 8.3.1 | `bash -n` 全修正スクリプト構文チェック | cc:done |
| 8.3.2 | `./tests/validate-plugin.sh` 実行 | cc:done |
| 8.3.3 | `./scripts/ci/check-consistency.sh` 実行 | cc:done |
| 8.3.4 | ミラー同期確認 (.claude-plugin/hooks.json 等) | cc:done |

---

## Phase 9: Claude Code 2.1.30 → 2.1.38 対応

Claude Code が 2.1.30 から 2.1.38 まで 8 バージョン進行。Harness の CLAUDE.md は「2.1.30+」で止まっている。
Breezing の一部ドキュメント（guardrails-inheritance.md, team-composition.md）は既に 2.1.33 の情報を反映済みだが、
hooks.json の実装・CLAUDE.md の Feature Table・各スキルのバージョン参照は未更新。

### 対応バージョン範囲

| Ver | リリース内容（Harness に関連するもの） | 影響度 |
|-----|---------------------------------------|--------|
| 2.1.31 | セッション再開ヒント、全角スペース対応、PDF ロック修正 | 低 |
| 2.1.32 | Opus 4.6、Agent Teams preview、自動メモリ記録、--add-dir スキル自動ロード、スキル文字バジェット 2% スケーリング | **高** |
| 2.1.33 | TeammateIdle/TaskCompleted hook、Task(agent_type) 制限、memory frontmatter 公式化、プラグイン名スキル表示 | **高** |
| 2.1.34 | Bash ask permission bypass セキュリティ修正 | 中 |
| 2.1.36 | Fast mode for Opus 4.6 | 中 |
| 2.1.38 | .claude/skills 書き込みブロック（sandbox）、heredoc 解析強化 | 中 |

### 優先度マトリクス

| 優先度 | 項目 | 理由 |
|--------|------|------|
| **Required** | CLAUDE.md Feature Table 更新 | ユーザーへの正確な情報提供 |
| **Required** | TeammateIdle/TaskCompleted hook 実装 | Breezing docs に記述済みだが hooks.json に未実装 |
| **Required** | 自動メモリ記録との共存設計 | 競合の可能性があり調査必須 |
| **Recommended** | Task(agent_type) 制限の検討 | エージェントセキュリティ強化 |
| **Recommended** | スキル文字バジェットガイドライン更新 | skill-editing.md の 500 行ルール見直し |
| **Recommended** | Fast mode ドキュメント | ユーザー向け情報 |
| **Optional** | .claude/skills sandbox 書き込みブロック検証 | pretooluse-guard との相互作用確認 |
| **Optional** | heredoc 解析互換性検証 | hooks スクリプトへの影響確認 |
| **Optional** | マイナーバージョン参照更新 | 各スキルの「CC 2.1.30+」表記を「CC 2.1.38+」に統一 |

---

### 9.1 CLAUDE.md Feature Table 更新 [feature:docs]

`CLAUDE.md` の「Claude Code 2.1.30+ 新機能活用ガイド」テーブルを 2.1.38+ に更新。

| Task | 内容 | Status |
|------|------|--------|
| 9.1.1 | ヘッダーを「2.1.30+」→「2.1.38+」に変更 | cc:done |
| 9.1.2 | 既存 8 行の機能・活用スキル・用途を最新に更新（古い機能は維持、表現を調整） | cc:done |
| 9.1.3 | 新規行追加: `TeammateIdle/TaskCompleted Hook` → breezing → チーム監視の自動化 | cc:done |
| 9.1.4 | 新規行追加: `Agent Memory (memory frontmatter)` → task-worker, code-reviewer → 永続的学習 | cc:done |
| 9.1.5 | 新規行追加: `Fast mode (Opus 4.6)` → 全スキル → 高速出力モード | cc:done |
| 9.1.6 | 新規行追加: `自動メモリ記録` → session-memory → セッション間知識の自動永続化 | cc:done |
| 9.1.7 | 新規行追加: `スキルバジェットスケーリング` → 全スキル → コンテキスト窓の 2% に自動調整 | cc:done |
| 9.1.8 | 新規行追加: `Task(agent_type) 制限` → agents/ → サブエージェント制限構文 | cc:done |

---

### 9.2 TeammateIdle / TaskCompleted Hook 実装

Breezing docs（guardrails-inheritance.md, team-composition.md）では既にこれらの Hook イベントを記述しているが、
`hooks.json` に実際のハンドラが存在しない。Lead 側で発火するイベントであり、チーム監視に活用する。

**設計判断**: TeammateIdle と TaskCompleted は Lead のコンテキストで発火する。
ペイロード: `teammate_name`, `team_name` (Idle) / `teammate_name`, `task_id`, `task_subject`, `task_description` (TaskCompleted)。
トークン数・ツール使用数は含まれない（2.1.33 で検証済み、guardrails-inheritance.md に記載）。

| Task | 内容 | Status |
|------|------|--------|
| 9.2.1 | `scripts/hook-handlers/teammate-idle.sh` 新規作成: イベントを `.claude/state/breezing-timeline.jsonl` に追記 | cc:done |
| 9.2.2 | `scripts/hook-handlers/task-completed.sh` 新規作成: タスク完了をタイムラインに追記 | cc:done |
| 9.2.3 | `hooks/hooks.json` に `TeammateIdle` ハンドラ追加 (timeout: 10) | cc:done |
| 9.2.4 | `hooks/hooks.json` に `TaskCompleted` ハンドラ追加 (timeout: 10) | cc:done |
| 9.2.5 | `.claude-plugin/hooks.json` に同期 | cc:done |
| 9.2.6 | `breezing/references/execution-flow.md` の「Team Activity」セクションでフックの実装状態を「実装済み」に更新 | cc:done |
| 9.2.7 | `bash -n` 構文チェック + 動作検証 | cc:done |

---

### 9.3 自動メモリ記録との共存設計

Claude Code 2.1.32 で「自動メモリ記録」機能が追加された。これは Harness の session-memory スキル・
`.claude/memory/` SSOT・agent memory (`memory: project`) とは別系統の Claude Code 組み込み機能。
共存設計を明確にし、ドキュメントに反映する。

**調査ポイント**:
- Claude Code 自動メモリは `~/.claude/` 配下に保存される（プロジェクト固有ではない）
- Harness の `.claude/memory/decisions.md`, `patterns.md` は SSOT として手動管理
- Agent memory (`memory: project`) は `.claude/agent-memory/` 配下
- 3 系統のメモリが独立して動作するか、干渉するかを確認

| Task | 内容 | Status |
|------|------|--------|
| 9.3.1 | Claude Code 自動メモリの保存先・フォーマット・トリガー条件を調査 (WebSearch + 公式ドキュメント) | cc:done |
| 9.3.2 | Harness memory (SSOT) との責務分界を定義: 自動メモリ = 暗黙的/汎用、SSOT = 明示的/プロジェクト固有 | cc:done |
| 9.3.3 | `skills/session-memory/SKILL.md` に「自動メモリとの関係」セクション追加 | cc:done |
| 9.3.4 | `skills/memory/SKILL.md` に自動メモリとの関係に関する注記追加 | cc:done |
| 9.3.5 | `.claude/memory/decisions.md` に D-next: 「3 層メモリアーキテクチャ」決定を記録 | cc:done |

---

### 9.4 Agent Type Restriction (Task(agent_type)) の検討

Claude Code 2.1.33 で `tools` フロントマターに `Task(agent_type)` 構文が追加された。
これにより、エージェントがスポーン可能なサブエージェントの種類を制限できる。

**現状分析**:
- `task-worker`: `disallowedTools: [Task]` → Task 完全禁止（適切: ワーカーはサブエージェント不要）
- `code-reviewer`: `disallowedTools: [Write, Edit, Bash, Task]` → Task 完全禁止（適切: レビュワーは read-only）
- `codex-implementer`: `tools: [Read, Write, Edit, Bash, Grep, Glob]` → Task 未リスト（暗黙的に使用不可）
- 他エージェント: 同様に Task 未リスト

**結論**: 現在の設計では Task をスポーンする必要のあるエージェントがないため、構文変更の実益は小さい。
ただし、将来的に Lead エージェントや orchestrator を定義する際に有用なため、ドキュメントに記録する。

| Task | 内容 | Status |
|------|------|--------|
| 9.4.1 | `agents/CLAUDE.md` または `.claude/memory/patterns.md` に `Task(agent_type)` パターンを記録 | cc:done |
| 9.4.2 | `skills/breezing/references/guardrails-inheritance.md` に利用可能な制限手段として追記 | cc:done |

---

### 9.5 スキルバジェットスケーリング対応

Claude Code 2.1.32 で「スキル文字バジェットがコンテキスト窓の 2% にスケール」に変更。
現在の `.claude/rules/skill-editing.md` は「500 行以下」のハードルールを記載。

| Task | 内容 | Status |
|------|------|--------|
| 9.5.1 | `skill-editing.md` の「File Size Guidelines」に 2.1.32 のスケーリング動作を注記 | cc:done |
| 9.5.2 | 500 行ルールを「推奨」に緩和し、「2% スケーリングにより実効上限はモデルのコンテキスト窓に依存」と説明追加 | cc:done |

---

### 9.6 バージョン参照の一括更新

各スキルの「CC 2.1.30+」表記を「CC 2.1.38+」に更新。CLAUDE.md の更新に合わせて実施。

**対象ファイル** (skills/ 内、ミラー除外):
- `skills/session/SKILL.md` (72行目)
- `skills/session-memory/SKILL.md` (176行目)
- `skills/harness-review/SKILL.md` (216行目, 482行目)
- `skills/harness-ui/SKILL.md` (29行目)
- `skills/parallel-workflows/references/run-task-workers.md` (307, 311, 409行目)
- `skills/codex-review/references/codex-mcp-setup.md` (102行目)
- `skills/troubleshoot/SKILL.md` (CC 2.1.30 参照があれば)
- `agents/task-worker.md` (355行目: MCP ツールアクセス)

| Task | 内容 | Status |
|------|------|--------|
| 9.6.1 | 上記ファイルの「CC 2.1.30+」→「CC 2.1.38+」一括更新 | cc:done |
| 9.6.2 | 新規追加の機能参照（Fast mode, 自動メモリ等）を適切なスキルに追記 | cc:done |
| 9.6.3 | `CHANGELOG_ja.md` の 2.1.30 対応行も更新 | cc:done |

---

### 9.7 セキュリティ・互換性検証

| Task | 内容 | Status |
|------|------|--------|
| 9.7.1 | `.claude/skills` sandbox 書き込みブロック: pretooluse-guard.sh との相互作用テスト | cc:done |
| 9.7.2 | heredoc 解析強化: hooks スクリプト内の heredoc パターンの互換性チェック (`bash -n` + 実テスト) | cc:done |
| 9.7.3 | Bash ask permission bypass 修正 (2.1.34): `autoAllowBashIfSandboxed` 設定が有効な場合の hooks 動作確認 | cc:done |

---

### 9.8 ミラー同期 + 検証 + リリース

| Task | 内容 | Status |
|------|------|--------|
| 9.8.1 | ミラー同期: `rsync -av --delete skills/ codex/.codex/skills/` + `opencode/skills/` | cc:done |
| 9.8.2 | `.claude-plugin/hooks.json` 同期 (9.2.5 と重複、最終確認) | cc:done |
| 9.8.3 | `./tests/validate-plugin.sh` 実行 | cc:done |
| 9.8.4 | `./scripts/ci/check-consistency.sh` 実行 | cc:done |
| 9.8.5 | バージョンバンプ: `./scripts/sync-version.sh bump` (patch or minor) | cc:done |
| 9.8.6 | CHANGELOG.md エントリ追加 | cc:done |
| 9.8.7 | コミット: `feat: support Claude Code 2.1.38 features` | cc:done |

---

### Phase 9 検証チェックリスト

| # | 検証項目 | 方法 |
|---|---------|------|
| 1 | CLAUDE.md Feature Table が 2.1.38+ を反映 | 目視確認 |
| 2 | hooks.json に TeammateIdle/TaskCompleted が存在 | `jq '.hooks.TeammateIdle' hooks/hooks.json` |
| 3 | 新規スクリプトの構文チェック通過 | `bash -n scripts/hook-handlers/teammate-idle.sh` |
| 4 | 自動メモリとの共存設計が decisions.md に記録 | 目視確認 |
| 5 | skill-editing.md がバジェットスケーリングを反映 | 目視確認 |
| 6 | 全スキルのバージョン参照が 2.1.38+ に更新 | `grep -r "2\.1\.30" skills/` でゼロ件 |
| 7 | ミラー同期完了 | `diff -rq skills/ codex/.codex/skills/` |
| 8 | validate-plugin.sh + check-consistency.sh 全パス | CI 実行 |
