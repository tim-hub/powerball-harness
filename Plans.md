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
| 5.1.1 | `pretooluse-guard.sh` symlink bypass 対策（realpath 検証追加） | |
| 5.1.2 | `permission-request.sh:58` npm/pnpm/yarn 自動承認をリポジトリ別 allowlist 方式に変更 | |
| 5.1.3 | `userprompt-track-command.sh:77` prompt_preview のパーミッション hardening (umask 077) | |
| 5.1.4 | `session-monitor.sh:275` resume_token の chmod 600 + umask 077 | |
| 5.1.5 | `pretooluse-guard.sh:354` eval を直接パース（jq/python）に置換 | |

### 5.2 ドキュメント・リンク修正

| Task | 内容 | Status |
|------|------|--------|
| 5.2.1 | `docs/QUALITY_GUARD_DESIGN.md` broken SSOT link 修正 | |
| 5.2.2 | `docs/PLAN_RULES_IMPROVEMENT.md` stale command refs 修正 | |
| 5.2.3 | `docs/plans/claude-mem-integration.md` stale paths 修正 | |
| 5.2.4 | `skills/workflow-guide/references/commands.md` path mismatch 修正 | |
| 5.2.5 | templates 内の `/skills-update` 参照を削除または更新 | |

### 5.3 generate-video メンテナンス

| Task | 内容 | Status |
|------|------|--------|
| 5.3.1 | `agents/video-scene-generator.md` Remotion paths 更新 | |
| 5.3.2 | `skills/generate-video/` references 内の Remotion paths 更新 | |
| 5.3.3 | `generate-video/src/schemas/*.ts` z.any() → z.unknown() / proper unions に修正 | |

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
| 5.6.1 | `check-checklist-sync.sh` empty gate logic 修正 | |
| 5.6.2 | `workflows/default/init.yaml` project-analyzer 参照修正 | |
| 5.6.3 | `build-opencode.js` commands/ 空ディレクトリ対応 | |
| 5.6.4 | `harness-ui` command catalog 空対応（統合後） | |
| 5.6.5 | `parse-work-flags.md` internal inconsistency 修正 | |

### 5.7 命名・ルーティング整理

| Task | 内容 | Status |
|------|------|--------|
| 5.7.1 | `/planning` → `/plan-with-agent` 完全統一（dual naming 解消） | |
| 5.7.2 | `verify` skill の `user-invocable` 整合性確認 | |
| 5.7.3 | setup と codex-review の Codex セットアップ重複整理 | |
| 5.7.4 | `_archived/` 配下からの dangling references 削除 | |

---

## 対象外（今回は見送り）

- `/gogcli-ops` — 独立した外部ツール連携。統合先がない。使用頻度に応じて別途判断
- `/deploy` — 高インパクト操作。明示的なコマンドとして維持（Codex も非表示に反対）
