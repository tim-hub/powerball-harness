# Claude Code Harness — Plans.md

最終アーカイブ: 2026-04-15（Phase 25-36 + 38 → `.claude/memory/archive/Plans-2026-04-15-phase25-36-38.md`）
前回アーカイブ: 2026-03-08（Phase 17〜24 → `.claude/memory/archive/Plans-2026-03-08-phase17-24.md`）

---

## Phase 37: 全フックハンドラの Go 移植 — "Hokage 完全体"

作成日: 2026-04-09
目的: hooks.json に残る 37 個の bash/Node.js ハンドラを Go サブコマンドに移植し、run-hook.sh + run-script.js + Node.js ランタイム依存を完全排除する
前提: Phase 35.0-35.2 完了済み、v4.0 "Hokage" コミット済み (14 フック Go 化済み)

### 設計方針

- 各ハンドラを `go/cmd/harness/main.go` の `runHook()` サブコマンドとして実装
- 既存 bash の動作を 1:1 で移植（機能追加はしない）
- `internal/hookhandler/` パッケージに実装を集約
- ハンドラごとに `_test.go` を作成（最低限の入出力テスト）
- hooks.json の該当エントリを `bin/harness hook <name>` に書き換え

### Node.js 依存状況

| 依存 | 件数 | 対象 |
|------|------|------|
| Node.js 必須 | 2 | pre-compact-save.js (783行), emit-agent-trace.js (808行) |
| 純 bash | 35 | 残り全て |

**Node.js ゼロ達成に必須なのは 2 ファイルの移植のみ。**

---

### Phase 37.1: Trivial ハンドラ (10個) [cc:TODO]

難易度: 低 / 各 ~50-100行 / ファイルI/O + JSON出力のみ

| Task | ハンドラ | 元ファイル | 行数 | 内容 | Status |
|------|---------|-----------|------|------|--------|
| 37.1.1 | pretooluse-inbox-check | scripts/pretooluse-inbox-check.sh | 82 | 他セッションからの未読メッセージチェック (5分スロットル) | cc:TODO |
| 37.1.2 | pretooluse-browser-guide | scripts/pretooluse-browser-guide.sh | 84 | agent-browser CLI 検出 + MCP ブラウザツール推奨 | cc:TODO |
| 37.1.3 | memory-bridge | scripts/hook-handlers/memory-bridge.sh + 4サブハンドラ | 55 | harness-mem MCP ブリッジディスパッチャ (session-start/user-prompt/post-tool/stop) | cc:TODO |
| 37.1.4 | worktree-create | scripts/hook-handlers/worktree-create.sh | 93 | .claude/state/ 作成 + worktree-info.json 記録 | cc:TODO |
| 37.1.5 | worktree-remove | scripts/hook-handlers/worktree-remove.sh | 73 | tmp ファイル削除 + worktree-info.json 削除 | cc:TODO |
| 37.1.6 | posttooluse-commit-cleanup | scripts/posttooluse-commit-cleanup.sh | 50 | git commit 検出 → review-approved.json 削除 | cc:TODO |
| 37.1.7 | posttooluse-clear-pending | scripts/posttooluse-clear-pending.sh | 28 | pending-skills/*.pending 削除 (スキル完了シグナル) | cc:TODO |
| 37.1.8 | session-auto-broadcast | scripts/session-auto-broadcast.sh | 103 | src/api/, types/, schema 変更時のチームメイト通知 | cc:TODO |
| 37.1.9 | config-change | scripts/hook-handlers/config-change.sh | 92 | ConfigChange → breezing-timeline.jsonl 記録 | cc:TODO |
| 37.1.10 | instructions-loaded | scripts/hook-handlers/instructions-loaded.sh | 86 | InstructionsLoaded → jsonl ログ + hooks.json 存在検証 | cc:TODO |

---

### Phase 37.2: Medium ハンドラ (12個) [cc:TODO]

難易度: 中 / 各 ~100-350行 / JSONL管理・状態追跡・条件分岐

| Task | ハンドラ | 元ファイル | 行数 | 内容 | Status |
|------|---------|-----------|------|------|--------|
| 37.2.1 | setup-hook | scripts/setup-hook.sh | 188 | プラグインキャッシュ同期 + .claude/state 初期化 + テンプレート検証 | cc:TODO |
| 37.2.2 | runtime-reactive | scripts/hook-handlers/runtime-reactive.sh | 168 | FileChanged/CwdChanged/TaskCreated → コンテキスト注入 | cc:TODO |
| 37.2.3 | teammate-idle | scripts/hook-handlers/teammate-idle.sh | 186 | チームメンバー idle 記録 + continue:false 停止シグナル | cc:TODO |
| 37.2.4 | userprompt-track-command | scripts/userprompt-track-command.sh | 107 | /slash コマンド検出 + usage 記録 + pending-skills マーカー | cc:TODO |
| 37.2.5 | breezing-signal-injector | scripts/hook-handlers/breezing-signal-injector.sh | 183 | breezing-signals.jsonl → systemMessage 注入 + consumed マーク | cc:TODO |
| 37.2.6 | ci-status-checker | scripts/hook-handlers/ci-status-checker.sh | 192 | git push/gh pr 検出 → CI ステータス非同期チェック | cc:TODO |
| 37.2.7 | usage-tracker | scripts/usage-tracker.sh | 108 | Skill/Task ツール使用追跡 | cc:TODO |
| 37.2.8 | todo-sync | scripts/todo-sync.sh | 118 | TodoWrite → Plans.md マーカー同期 (pending→cc:TODO等) | cc:TODO |
| 37.2.9 | auto-cleanup-hook | scripts/auto-cleanup-hook.sh | 118 | Write/Edit 後のファイルサイズ警告 (>10KB) | cc:TODO |
| 37.2.10 | track-changes | scripts/track-changes.sh | 185 | ファイル変更記録 + 2時間 dedup + パス正規化 | cc:TODO |
| 37.2.11 | plans-watcher | scripts/plans-watcher.sh | 201 | Plans.md 変更検出 + WIP/TODO/done マーカーサマリ注入 | cc:TODO |
| 37.2.12 | tdd-order-check | scripts/tdd-order-check.sh | 115 | 実装ファイル先行編集の警告 (TDD 順序強制) | cc:TODO |

---

### Phase 37.3: Medium ハンドラ — 既存 Go 補完 (7個) [cc:TODO]

Go binary に既にルーティングがあるが、hooks.json がまだ bash を呼んでいるもの

| Task | ハンドラ | 元ファイル | 行数 | 内容 | Status |
|------|---------|-----------|------|------|--------|
| 37.3.1 | elicitation-handler | scripts/hook-handlers/elicitation-handler.sh | 139 | MCP Elicitation → ログ + Breezing 時自動スキップ | cc:TODO |
| 37.3.2 | elicitation-result | scripts/hook-handlers/elicitation-result.sh | 123 | ElicitationResult → jsonl ログ | cc:TODO |
| 37.3.3 | stop-session-evaluator | scripts/hook-handlers/stop-session-evaluator.sh | 106 | Stop → セッション状態評価 + session.json 更新 | cc:TODO |
| 37.3.4 | stop-failure | scripts/hook-handlers/stop-failure.sh | 178 | StopFailure → API エラーログ (rate limit, auth) | cc:TODO |
| 37.3.5 | notification-handler | scripts/hook-handlers/notification-handler.sh | 166 | Notification → notification-events.jsonl 記録 | cc:TODO |
| 37.3.6 | permission-denied-handler | scripts/hook-handlers/permission-denied-handler.sh | 197 | PermissionDenied → denial ログ + Breezing Lead 通知 | cc:TODO |
| 37.3.7 | posttooluse-quality-pack | scripts/posttooluse-quality-pack.sh | 190 | Write/Edit 後の品質チェック (Prettier, tsc, console.log 検出) | cc:TODO |

---

### Phase 37.4: Hard ハンドラ (8個) [cc:TODO]

難易度: 高 / 各 ~300-900行 / 状態機械・プロセス制御・Node.js 移植

| Task | ハンドラ | 元ファイル | 行数 | 内容 | Status |
|------|---------|-----------|------|------|--------|
| 37.4.1 | userprompt-inject-policy | scripts/userprompt-inject-policy.sh | 351 | メモリ resume コンテキスト注入 + セマフォロック + RESUME_MAX_BYTES 制限 | cc:TODO |
| 37.4.2 | fix-proposal-injector | scripts/hook-handlers/fix-proposal-injector.sh | 338 | pending-fix-proposals.jsonl → 提案表示 + 承認/却下 → Plans.md 同期 | cc:TODO |
| 37.4.3 | posttooluse-log-toolname | scripts/posttooluse-log-toolname.sh | 333 | ツール使用ログ + LSP 追跡 + セッションイベントログ (500行ローテーション) + flock | cc:TODO |
| 37.4.4 | auto-test-runner | scripts/auto-test-runner.sh | 326 | ソースファイル変更検出 → テスト自動実行 (async) + Vitest/Jest/pytest 自動判定 | cc:TODO |
| 37.4.5 | task-completed | scripts/hook-handlers/task-completed.sh | 911 | タスク完了記録 + fix proposal 生成 + Breezing タイムライン + Plans.md 同期 (最大) | cc:TODO |
| 37.4.6 | **pre-compact-save.js** ⚡ | scripts/hook-handlers/pre-compact-save.js | 783 | **Node.js** — handoff-artifact.json + precompact-snapshot.json 生成 + Git 情報収集 | cc:TODO |
| 37.4.7 | **emit-agent-trace.js** ⚡ | scripts/emit-agent-trace.js | 808 | **Node.js** — agent-trace.jsonl 記録 + OpenTelemetry span + 10MB/3世代ローテーション | cc:TODO |
| 37.4.8 | post-compact (拡張) | scripts/hook-handlers/post-compact.sh | 380 | PostCompact 拡張 — WIP コンテキスト + handoff artifact 再注入 (現 Go 版の補完) | cc:TODO |

⚡ = Node.js 依存。これら 2 ファイルの移植で Node.js ゼロ達成。

---

### Phase 37.5: hooks.json 最終書き換え + レガシー削除 [cc:TODO]

| Task | 内容 | DoD | Status |
|------|------|-----|--------|
| 37.5.1 | hooks.json の残り 37 エントリを全て `bin/harness hook <name>` に書き換え | `grep -c 'run-hook.sh' hooks/hooks.json` が 0 | cc:TODO |
| 37.5.2 | `scripts/run-hook.sh` + `scripts/run-script.js` 削除 | ファイルが存在しない | cc:TODO |
| 37.5.3 | `package.json` 削除 (npm 依存の完全排除) | ファイルが存在しない | cc:TODO |
| 37.5.4 | `core/` (TypeScript エンジン) 削除 | ディレクトリが存在しない | cc:TODO |
| 37.5.5 | E2E テスト: 全フックイベントが Go binary 経由で動作 | `go/scripts/test-e2e.sh` 全パス | cc:TODO |
| 37.5.6 | `harness doctor` が Node.js 依存ゼロを確認 | `grep -rE "node\|run-script" hooks/` が 0 件 | cc:TODO |

---

### Phase 37 完成基準

| # | 基準 | 検証方法 |
|---|------|---------|
| 1 | hooks.json に `run-hook.sh` 参照が 0 件 | `grep 'run-hook' hooks/hooks.json` |
| 2 | `scripts/run-hook.sh`, `run-script.js`, `package.json`, `core/` が削除済み | `ls` で確認 |
| 3 | `node` コマンドへの参照がハーネス内に 0 件 (codex-companion 除く) | `grep -r 'node ' scripts/ hooks/` |
| 4 | 全 37 ハンドラに Go テスト (`_test.go`) が存在 | `go test ./internal/hookhandler/...` |
| 5 | `harness doctor` が全チェック PASS | `bin/harness doctor` |
| 6 | `go/scripts/test-e2e.sh` が全フックイベントをカバー | E2E 全パス |

合計: **37 ハンドラ移植 + 6 クリーンアップ = 43 タスク**

---

## Phase 39: レビュー体験改善 + v4.0.1 リリース前 polish

作成日: 2026-04-11
目的: Phase 38 完了後に発見した改善機会 (/HAR:review 出力の非専門家化、sync.go の plugin.json auto-revert 根本修正、test assertion 厳密化、v3 cleanup 残骸、bare review scope cap) を事後整理し、v4.0.1 リリース前に CHANGELOG を更新して出荷可能状態にする

背景: Phase 38 後の独立レビューで 3 つの follow-up が発見された (HAR:* 検証・jq assertion 緩和・scope 過大)。これらを修正する過程で名前空間の不整合、レビュー出力の UX 問題、harness sync の plugin.json 再生成時に skills field が消える根本バグなど、より深い問題が次々に発覚。全て Phase 38 スコープ外だが v4.0.1 リリース前に解消すべき品質改善なのでここに集約する。

### 優先度マトリクス

| 優先度 | Phase | 内容 | タスク数 | 依存 |
|--------|-------|------|---------|------|
| **完了** | 39.0 | /HAR:review 出力改善 (bare flow + Japanese + 非専門家向け) | 3 | - |
| **完了** | 39.1 | インフラ修正 (sync.go Skills + test assertion + scope cap) | 3 | - |
| **完了** | 39.2 | 名前整合性 (HAR:* → harness-* revert + SSOT 回復) | 1 | - |
| **完了** | 39.3 | v3 cleanup 残骸除去 + テストスクリプトの v4 migration | 5 | - |
| **Required** | 39.4 | CHANGELOG [Unreleased] 更新 | 1 | 39.0-39.3 |
| **Recommended** | 39.5 | レビュアー指摘のフォローアップ (shell pipeline + memory_bridge) | 2 | - |

合計: **15 タスク** (12 完了済み + 3 未着手、うち 1 件がリリースブロッカー)

### 完成基準 (Definition of Done — Phase 39 全体)

| # | 基準 | 検証方法 | 必須/推奨 |
|---|------|---------|----------|
| 1 | /HAR:review が bare 呼び出しで動作し、日本語結果サマリー冒頭出力 + 非専門家向け 4 段構造 | `/harness-review` 実行して目視確認 | 必須 |
| 2 | `harness sync` 実行後も `plugin.json.skills = ["./"]` が保持される | `jq '.skills' .claude-plugin/plugin.json` | 必須 |
| 3 | `go test ./...` 全 12 パッケージ PASS | コマンド実行 | 必須 |
| 4 | `./tests/validate-plugin.sh` 42 合格 / 0 失敗 | コマンド実行 | 必須 |
| 5 | `./scripts/ci/check-consistency.sh` 全合格 | コマンド実行 | 必須 |
| 6 | CHANGELOG [Unreleased] に Phase 38 + Phase 39 の改善が日本語 Before/After で記載 | 目視確認 | 必須 |
| 7 | Phase 39.5 は v4.0.2 以降で対応 (リリース前はブロッカーにしない) | Plans.md で cc:TODO のまま | 推奨 |

---

### Phase 39.0: /HAR:review 出力改善 [P0 完了]

Purpose: レビュー体験を「技術者のための英語 JSON 出力」から「非専門家にも読める日本語サマリー + 技術詳細は折り畳み」に刷新

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 39.0.1 | `skills/harness-review/SKILL.md` に bare 呼び出しデフォルトフロー (Step 0) を追加。`git describe --tags` → main → HEAD~10 の自動 BASE_REF 決定ロジック + Step 1 への自動遷移。3 mirror 同期 | bare 呼び出しで Step 1 に自動到達することを目視確認。validate-plugin / check-consistency 維持 | - | cc:完了 [8d2b89cc] |
| 39.0.2 | Step 0 に「出力言語・フォーマット (絶対遵守)」ブロック追加、Step 3 に結果サマリー冒頭出力ルール追加。日本語出力必須化 + JSON は後ろに補足 | CLAUDE.md の "context: fork skills も日本語" ルールを明示引用。`/harness-review` 実行で日本語出力を確認 | 39.0.1 | cc:完了 [af915fb4] |
| 39.0.3 | 非専門家向けテンプレート再設計 (情報粒度 MID / 認知負荷 MIN)。判定 → ✨ 良かったところ → ⚠️ 気になったところ (4 段: 日本語タイトル → 問題 → 対応 → 重要度 → 技術詳細) → 🎬 次のアクション → 📊 自動検証 → 📦 詳細データ (JSON 降格) の 6 セクション構造。3 mirror 同期 | `/harness-review` 出力が新テンプレートに沿う。重要度は「🔴 致命的 / 🟠 重要 / 🟡 軽微 / 🟢 推奨」の日本語+絵文字表記。英語重要度語と技術用語を本文から隔離 | 39.0.2 | cc:完了 [7481f98f] |

---

### Phase 39.1: インフラ修正 [P0 完了]

Purpose: レビューで発見された 3 つの根本バグを修正 (plugin.json auto-revert、jq assertion 緩和、bare review scope 過大)

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 39.1.1 | `go/cmd/harness/sync.go` の `pluginJSON` struct に `Skills []string` フィールド追加、`generatePluginJSON()` で `[]string{"./"}` をハードコード設定。これで `harness sync` が毎回 plugin.json から `"skills": ["./"]` を削除する auto-revert ループを根本解消 | `TestSync_GeneratesPluginJSON` に `skills == ["./"]` アサーション追加して PASS。`harness sync .` 実行後 `jq '.skills' plugin.json` が `["./"]` を返す | - | cc:完了 [009faf74] |
| 39.1.2 | `tests/test-memory-hook-wiring.sh` の SessionStart matcher チェックを `contains("startup")` から pipe-token 正規表現 `test("(^|\\|)startup($|\\|)")` に厳密化。`startup-only` 等のタイポを silently pass する false positive を防止 | 6 エッジケースで検証 (`startup`, `startup|resume`, `resume|startup` は match、`startup-only`, `startup_special`, `resume|startup-only` は reject)。`bash tests/test-memory-hook-wiring.sh` OK | - | cc:完了 [f7146d3e] |
| 39.1.3 | `skills/harness-review/SKILL.md` の Step 0.1 に上限フォールバック追加: 最後のタグから HEAD までの commit 数が 10 超なら HEAD~10 に自動 clamp。bare 呼び出し時のスコープ過大を防止 | 10 commits 超で自動 clamp 動作。レビュー実行時にサマリーで元候補と絞込結果の両方を表示。3 mirror 同期 | - | cc:完了 [9103377f] |

---

### Phase 39.2: 名前整合性 [P0 完了]

Purpose: ff4ee422 で frontmatter `name` を `HAR:*` に変更したが、directory 名と不一致で skill-editing.md SSOT ルール違反 + 内部テキストが `harness-*` のまま残る 3-way split が発生。整合性を戻す

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 39.2.1 | 18 ファイル (6 skills × 3 locations: main/codex/opencode) の frontmatter `name:` を `HAR:*` から `harness-*` に revert。description 先頭の "HAR:" ブランドは維持 (視覚的識別性のため) | 18 ファイル全ての `name:` が `harness-*` (grep 検証)。description の `"HAR:` は 54 箇所 (18 ファイル × 3 description fields) で維持。validate-plugin / check-consistency 維持 | - | cc:完了 [af915fb4] |

---

### Phase 39.3: v3 cleanup 残骸除去 + テストスクリプト v4 migration [P0 完了]

Purpose: v4.0.0 リリース時に掃除漏れだった v3 時代の参照 (deleted TypeScript rules.ts への参照、README の "TypeScript engine" 記述、v3 hook 呼び出しパターンのテスト、v3 命名の帖) を除去

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 39.3.1 | `tests/validate-plugin.sh` の RULES_FILE パスを deleted `core/src/guardrails/rules.ts` から `go/internal/guardrail/rules.go` に変更、R12 expected pattern を `warn-direct-push-protected-branch` から `deny-direct-push-protected-branch` に同期 | validate-plugin.sh から 4 件の R10-R13 失敗が消失、合格数が 35 → 40 に増加 | - | cc:完了 [cbea4620] |
| 39.3.2 | `scripts/ci/check-consistency.sh` の README 期待文字列を `"TypeScript guardrail engine"` / `"TypeScript ガードレールエンジン"` から `"Go-native guardrail engine"` / `"Go ネイティブガードレールエンジン"` に同期 | check-consistency.sh から 2 件の TypeScript 参照失敗が消失、全合格 | - | cc:完了 [cbea4620] |
| 39.3.3 | `tests/test-memory-hook-wiring.sh` と `tests/test-claude-upstream-integration.sh` の jq クエリを v3 shell パス (`hook-handlers/memory-bridge`) から v4 Go binary 形式 (`bin/harness hook memory-bridge` 相当) に migrate。agent-type hook の `command` null 対応も追加 | 両 test script が直接実行 OK、validate-plugin.sh から 2 件の "missing wiring" 失敗が消失 | - | cc:完了 [c91b21c1] |
| 39.3.4 | `tests/test-claude-upstream-integration.sh` の PermissionDenied wiring check を `contains("permission-denied-handler")` から `contains("permission-denied")` に同期 (v4 の `bin/harness hook permission-denied` 形式に対応) | test-claude-upstream-integration.sh 直接実行 OK、validate-plugin.sh の最後の 1 件失敗が解消して 42/0 全合格に到達 | 39.3.3 | cc:完了 [04026f3a] |
| 39.3.5 | v4 cleanup 残骸の削除: 2 つの JSON 名 ghost directory (Agent tool isolation エラーの副産物)、`core/` 残骸 (node_modules + package-lock.json)、`infographic-check.png` (debug screenshot)、`.orphaned_at` (旧 session marker) を削除 | ルート直下に上記 5 件が存在しない。git status に影響なし (全 untracked) | - | cc:完了 [Lead 直接実行] |

---

### Phase 39.4: CHANGELOG 更新 [P0 Required — v4.0.1 リリースブロッカー]

Purpose: Phase 38 と Phase 39 の全改善を CHANGELOG.md の [Unreleased] セクションに日本語 Before/After 形式で記載し、v4.0.1 リリース時に変更内容がユーザーに正しく伝わる状態にする

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 39.4.1 | `CHANGELOG.md` の `[Unreleased]` セクションに Phase 39 の全改善を追記。既存の Phase 38 エントリ (CC 2.1.89-2.1.100 追従) はそのまま維持し、その下に「#### 8. レビュー体験の改善 (Phase 39)」等の新セクションを追加。各変更は日本語 Before/After で記述 (`.claude/rules/github-release.md` 準拠)。対象は: (a) /HAR:review bare 呼び出し + 非専門家テンプレート、(b) sync.go Skills field 根本修正、(c) SessionStart matcher 厳密化、(d) bare review scope cap、(e) name revert、(f) v3 cleanup 残骸除去、(g) test scripts v4 migration | CHANGELOG.md Unreleased に Phase 39 の 5 サブエントリ (項目 8-12) が追加されている。VERSION / plugin.json version / harness.toml version は変更しない (リリース作業ではないため)。`./scripts/ci/check-consistency.sh` PASS、`./tests/validate-plugin.sh` 42 合格 / 0 失敗 | 39.0.1-39.3.5 | cc:完了 [c96ca7d1] |

---

### Phase 39.5: レビュアー指摘のフォローアップ [P1 Recommended — v4.0.2 以降]

Purpose: 前回および今回の /HAR:review で recommendation として発見された 2 件を追跡起票。いずれも現状動作に問題なく、v4.0.1 リリースのブロッカーにはしない

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 39.5.1 | `skills/harness-review/SKILL.md` の Step 0.2 にある `grep -c '^plan:' \| awk '$1 > 2 {exit 0} {exit 1}'` パイプラインを strict-mode 耐性に強化。`awk '/^plan:/ {n++} END {exit (n<=2)}'` のような単一 awk 統合、または `\|\| true` による保険を追加。3 mirror 同期 | `set -euo pipefail` 下で bare 呼び出しが plan: 0 件のケースでも誤停止しない。既存挙動 (plan: > 2 件で plan review モード判定) は維持 | - | cc:TODO |
| 39.5.2 | `go/internal/hookhandler/memory_bridge.go` の `validTargets` map が kebab-case (`session-start`, `user-prompt` 等) を期待しているが、CC は `hook_event_name` に PascalCase (`SessionStart`, `UserPromptSubmit` 等) を送っているため常に "unknown target" fail-open branch に落ちて memory bridge が実質動作していない問題を修正。HookEventName を正規化するか validTargets キーを PascalCase に揃える | memory bridge が実際に session-start / user-prompt / post-tool-use / stop の各 event を dispatch する。hook handler テストに PascalCase 入力のケース追加 | - | cc:TODO |

---

## Phase 40: Migration Residue Scanner — inclusion → exclusion verification

作成日: 2026-04-11
目的: Harness の検証層に **exclusion-based verification (削除された概念への参照が残っていないかを systematic に検出する層)** を追加する。現状は「X が含まれるか」の inclusion-based チェックのみで、major migration (v3→v4 等) 後の残骸を偶然発見に頼る状態。今セッションで検出した 13 件の v3 残骸バグが将来自動的に捕捉される状態にする

### 背景 (Why this phase exists)

v4.0.0 "Hokage" リリース (2026-04-09) から v4.0.1 までの 2 日間で、13 件の v3 残骸バグが**偶然発見**された:

| # | 残骸バグ | 発見経緯 | 影響 |
|---|---|---|---|
| 1 | `validate-plugin.sh` が削除済み `core/src/guardrails/rules.ts` を grep | 初回 validate 実行で 4 件失敗 | 検証スクリプトが false negative |
| 2 | `check-consistency.sh` が README に `"TypeScript guardrail engine"` を期待 | 初回 consistency 実行で 2 件失敗 | 同上 |
| 3 | `tests/test-memory-hook-wiring.sh` が v3 shell path を厳密一致期待 | validate の下流失敗 | 同上 |
| 4 | `tests/test-claude-upstream-integration.sh` の `permission-denied-handler` | Worker C の部分修正後に再発見 | 同上 |
| 5 | 18 SKILL.md frontmatter の `"Harness v3"` 文字列 | ユーザーがスラッシュパレットで気づく | ユーザー混乱 |
| 6 | `agents/*.md` の `v3` narrative | grep 再走査で副次発見 | 同上 |
| 7 | SKILL.md H1 タイトルの `(v3)` サフィックス | 同上 | 同上 |
| 8 | `harness.toml` → `plugin.json` sync が `skills: ["./"]` を削除 | auto-revert 現象で気づく | 機能退行 |
| 9 | `/HAR:review` SKILL.md 本体が英語中心 | fork subagent が英語で返答 | UX 問題 |
| 10 | `README.md` ファイルツリーの `core/ engine` 言及 | ユーザー最終指摘 | 誤誘導 |
| 11 | `README.md` ファイルツリーの `skills/`/`agents/` 重複バグ | 同上 | 同上 |
| 12 | `README_ja.md` 同じ問題 | 同上 | 同上 |
| 13 | `README.md` troubleshooting の `Node.js 18+` 要求 | 同上 | 誤誘導 |

全て **偶然発見**: テスト失敗・ユーザー指摘・レビュー指摘のいずれか。systematic scanner があれば **v4.0.0 リリース前に全て検出できた class の bug**。これは Harness の verification 戦略に関する**根本的な欠陥**を示している: inclusion-based (「X が含まれるか」) のみで exclusion-based (「削除された X が残っていないか」) の視点が欠けている。

### 優先度マトリクス

| 優先度 | Phase | 内容 | タスク数 | 依存 |
|--------|-------|------|---------|------|
| **Required** | 40.0 | 基盤 (deleted-concepts.yaml + check-residue.sh) | 2 | なし |
| **Required** | 40.1 | 統合 (doctor --residue + validate-plugin + release preflight) | 3 | 40.0 |
| **Required** | 40.2 | ドキュメント (migration-policy.md) | 1 | 40.0, 40.1 |

合計: **6 タスク**

### 完成基準 (Definition of Done — Phase 40 全体)

| # | 基準 | 検証方法 | 必須/推奨 |
|---|------|---------|----------|
| 1 | scanner が今セッションの 13 件の v3 残骸を retroactive に全て検出 | v4.0.1 時点の 1 つ前の commit に戻して `bash scripts/check-residue.sh` → 13 件検出 / リリース後の commit → 0 件 | 必須 |
| 2 | scanner の false positive 率ゼロ (CHANGELOG.md 等の歴史記述を allowlist で正しく除外) | v4.0.1 HEAD で実行して 0 件 | 必須 |
| 3 | `bin/harness doctor --residue` が scanner を呼び出して結果表示 | コマンド実行で期待出力 (件数 + ファイル + 行番号) | 必須 |
| 4 | `validate-plugin.sh` に residue 統合、失敗時に合計 fail カウントに加算 | 意図的に v3 残骸を混入 → validate-plugin が失敗 | 必須 |
| 5 | `harness-release` skill の preflight に `harness doctor --residue` を組み込み、失敗時にリリース中止 | SKILL.md に明記 + dry-run で動作確認 | 必須 |
| 6 | `.claude/rules/migration-policy.md` が存在、deleted-concepts.yaml の更新ルールを明文化 | ファイル存在 + 内容確認 | 必須 |
| 7 | `.claude/rules/deleted-concepts.yaml` にこのセッションで検出した 13 件を全てエントリ化 | yaml を parse、最低 8-10 のエントリ (13 件はパターンに集約可能) | 必須 |
| 8 | Go test 全パス、validate-plugin 43+/0 (residue check 分増加)、check-consistency 全合格 | 既存のテスト走査 | 必須 |

---

### Phase 40.0: 基盤 [P0]

Purpose: `.claude/rules/deleted-concepts.yaml` を SSOT として定義し、それを読んで repo を scan する `check-residue.sh` を実装する

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 40.0.1 | `.claude/rules/deleted-concepts.yaml` を新規作成。スキーマは `deleted_paths[]` と `deleted_concepts[]` の 2 セクション。各エントリは `path`/`term`, `term_ja` (任意), `replacement`/`replacement_ja` (任意), `deleted_in` (version), `deleted_by` (commit hash 任意), `reason`, `allowlist[]` を含む。今セッションで検出した 13 件の v3 残骸をパターンに集約してエントリ化: (a) `core/src/guardrails`, (b) `core/dist`, (c) `core/package.json`, (d) `scripts/run-hook.sh`, (e) 用語 `"TypeScript guardrail engine"` / `"TypeScript ガードレールエンジン"`, (f) 用語 `"Harness v3"` / `"(v3)"`, (g) troubleshooting 文言 `"Node.js 18+ is installed"` / `"Node.js 18+ が必要"` / `"Ensure Node.js"`, (h) v3 shell invocation pattern `"hook-handlers/memory-bridge"` / `"hook-handlers/runtime-reactive"` / `"hook-handlers/permission-denied-handler"`。allowlist には `CHANGELOG.md`, `.claude/memory/archive/**`, `benchmarks/**`, `README.md` の "Before / After" table 領域を含める | YAML が valid (`yq` で parse OK)。8-10 個のエントリ + 各エントリに allowlist 配列。`reason` フィールドで「なぜ削除されたか」が各エントリで説明されている (v4.0.0 Hokage migration, CC 2.1.94 対応等) | - | cc:完了 [20654143, 191cdde4] |
| 40.0.2 | `scripts/check-residue.sh` を実装。`.claude/rules/deleted-concepts.yaml` を `yq` で読み込み、`deleted_paths[]` と `deleted_concepts[]` を順次 `grep -rln -F` でスキャン。allowlist 適用 (`.gitignore` 風のマッチングか prefix match)。hit があれば exit code 1、詳細レポートを stdout に出力 (件数、ファイル、行番号、該当文字列、どのエントリの violation か)。`set -euo pipefail` 下でも動作すること。エラー時の fallback 仕様も実装 (yq 未インストール → python3 + yaml module でパースする fallback) | (a) v4.0.0 release commit (`8d8ce3c8`) + v4.0.1 前の時点で scanner 実行 → 今セッションの 13 件の v3 残骸を全て検出 (retroactive validation)、(b) v4.0.1 以降の HEAD で実行 → 残骸 0 件 (false positive ゼロ)、(c) 意図的に `core/src/guardrails/rules.ts` への reference を README に追加 → 即検出。`bash scripts/check-residue.sh` 単体実行で全挙動 OK | 40.0.1 | cc:完了 [20654143, 191cdde4] |

---

### Phase 40.1: 統合 [P0]

Purpose: 3 つの検証ポイント (developer ad-hoc / PR ごと / リリース前) に scanner を組み込む

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 40.1.1 | `go/cmd/harness/doctor.go` に `--residue` フラグを追加。内部で `scripts/check-residue.sh` を subprocess 呼び出しして結果をフォーマット表示。または yaml を直接 parse する Go 実装でも可 (その場合 `gopkg.in/yaml.v3` 等を go.mod に追加)。既存の `--migration` 等のフラグと独立して動作。`bin/harness doctor` 単体実行時は residue check をスキップ (高速化のため opt-in) | `bin/harness doctor --residue` を実行すると: (a) scan 中のメッセージ表示、(b) 検出された場合はファイル + 行番号 + 該当エントリ、(c) 0 件なら "✓ No migration residue detected" を表示、(d) exit code 0 (clean) or 1 (residue)。`go test ./cmd/harness/ -run TestDoctor_Residue` で意図的 residue 混入テストが PASS | 40.0.2 | cc:完了 [470a05bd] |
| 40.1.2 | `tests/validate-plugin.sh` に residue scan を組み込む。新しい test カテゴリ "migration residue check" として追加、既存の合格/失敗/警告カウントに統合 (residue 0 件 → +1 合格、residue 1 件以上 → +1 失敗) | `./tests/validate-plugin.sh` 実行時に residue check が最後のセクションとして走る。clean state で合格数 +1 (42 → 43)。意図的な residue 混入時は失敗数 +1。既存テストは全て維持 | 40.0.2 | cc:完了 [1e886ad9] |
| 40.1.3 | `skills/harness-release/SKILL.md` の preflight セクション (Step 1 相当) に `bin/harness doctor --residue` を追加。residue 検出時はリリース中止 + ユーザーへ修正指示を日本語で表示。3 mirror (`skills/`, `codex/.codex/skills/`, `opencode/skills/`) 同期 | `harness-release` の preflight テーブルに residue check 行追加。residue 検出時の error message が「Phase 40 の scanner が N 件の削除済み概念への参照を検出しました。修正してから再実行してください。」のように日本語で明確。`check-consistency.sh` PASS (mirror 完全一致) | 40.0.2 | cc:完了 [60199c01] |

---

### Phase 40.2: ドキュメント [P0]

Purpose: 今後の major migration で scanner を正しく運用するためのルール化

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 40.2.1 | `.claude/rules/migration-policy.md` を新規作成。内容: (1) major version migration 時に `.claude/rules/deleted-concepts.yaml` を更新する義務、(2) 更新タイミングは削除 PR と同時 (遅延禁止)、(3) allowlist の運用基準 (歴史記述 CHANGELOG は常に allowlist、Before/After table は文脈で allowlist、docs/archive は常に allowlist)、(4) retroactive validation の実施方法、(5) 今セッション (v4.0.0 → v4.0.1) で検出された 13 件の v3 残骸事例を付録として記録 (なぜこの機能が生まれたかのストーリー付き)、(6) `CLAUDE.md` に migration-policy.md への参照を追加 | ファイルが存在、markdown valid。非専門家にも「なぜ deleted-concepts.yaml が必要か」が 5 分で理解できる。CLAUDE.md に 1 行 reference 追加 | 40.0.1, 40.1.3 | cc:完了 [719c08bd] |

---

## Phase 41: Long-Running Harness — Anthropic harness-design 記事統合

作成日: 2026-04-14
目的: Anthropic の記事 "Harness Design for Long-Running AI Agent Applications" (https://www.anthropic.com/engineering/harness-design-long-running-apps) から導出した 6 つの改善軸（B6/B7/B9/B10/B11 + /loop 統合）を実装し、記事が示す 6h+ クラスの自律実行を **同一セッション内での context reset パターン**で実現する

### 背景 (Why this phase exists)

現行 Harness は記事のベンチマーク 12 軸のうち 9 軸をクリアしているが、以下 4 点で弱点がある:

1. **B3 context 管理の checkpoint 頻度不足**: `harness_mem_record_checkpoint` 相当の呼び出しが `scripts/` / `skills/` / `agents/` / `hooks/` で **grep 0 件**。手動・セッション終了時のみで、Phase B 各サイクル後の永続化が未実装
2. **B7 browser verdict が最終判定に乗らない**: `skills/harness-work/SKILL.md` L250-252 で "browser artifact は参照用に保存するが、review-result の verdict は static のまま" と明記
3. **B10 反復上限 3 回一律**: `MAX_REVIEWS = 3` が `harness-work` SKILL.md L263 に直書き。UI/design タスクの 5-15 反復に対応できない
4. **B11 plateau / pivot 機構なし**: 3 回失敗でエスカレーション停止のみ

また、記事の Game Maker (6h) / DAW (3h50m) 相当を実現するには、単一会話の context 膨張を避ける仕組みが必要。**`/loop` コマンド（CC 提供の dynamic mode / 内部 CronCreate backend）と、Claude 側から呼ぶ `ScheduleWakeup` ツール（同じ dynamic mode の pacing 制御）** を活用し、wake-up ごとに fresh context で再入することで、Opus 4.6 / Sonnet 両系統で context anxiety を構造的に回避する。両 API の実制約は 41.0.0 spike で確定させ、計画内の記述を spike 結果に一本化する。

> **記事との対応注記**: Phase 41 は**同一セッション内**での context reset により、記事の 6h 級を **"部分再現"** するもの（3h 程度の連続反復が現実的目標）。ホスト CC プロセスをまたぐ継続は Non-Goals とし Phase 42 以降で別途検討。

### 設計方針（TeamAgent 3 視点合意事項）

以下は plan-critic / scaffolder / architect の 3 エージェント並列議論で合意した非交渉事項:

- **A-β**: `/loop` 統合は `skills/harness-work/SKILL.md` への `--loop` フラグ追加ではなく、**独立スキル `skills/harness-loop/` を新設**。SKILL.md 責務膨張と mirror 同期コストを回避
- **B-α 改**: auto-checkpoint は Phase B-5 の `git commit` 直後ではなく、**Plans.md 書き換え完了直後**に移動。臨界路を短く保ち、タスク紐付けを確実にする
- **C-β**: plateau 検知は critical/major 件数のみの単信号ではなく、**iteration 数 ≥ 3 AND 修正対象ファイル集合の Jaccard 類似度 > 0.7** の AND 条件。false positive 耐性を確保
- **スコープ制限**: Phase 41 は**同一 CC セッション内**での context reset + resume に限定。ホスト CC プロセス終了をまたぐ継続（tmux 常駐前提）は Phase 42 以降に defer
- **wake-up 回数上限**: default 8 サイクル。`bypassPermissions` 下で destructive action が累積するリスクを制御
- **permission mode**: /loop 中も現行の `bypassPermissions` を維持。`--auto-mode` との組み合わせは Phase 41 では opt-in のまま

### 優先度マトリクス

| 優先度 | Phase | 内容 | タスク数 | 依存 |
|--------|-------|------|---------|------|
| **Required** | 41.0 | Spike + 基盤（harness-mem/CronCreate API 調査 + calibration 拡張 + 新規 scripts 2 本） | 4 | なし |
| **Required** | 41.1 | /loop 統合（独立 harness-loop スキル + 冪等性ガード） | 2 | 41.0 |
| **Recommended** | 41.2 | 反復制御（MAX_REVIEWS 可変化 + browser verdict AND 結合） | 2 | 41.0, 41.1 |
| **Optional** | 41.3 | UI rubric profile 新設 | 1 | 41.2 |
| **Required** | 41.4 | ドキュメント + mirror 同期 + preflight 拡張 | 3 | 41.1, 41.2 |

合計: **12 タスク**（Required 9 / Recommended 2 / Optional 1）

### 完成基準 (Definition of Done — Phase 41 全体)

| # | 基準 | 検証方法 | 必須/推奨 |
|---|------|---------|----------|
| 1 | **同一セッション内で 3 サイクル連続 /loop 実行**: Plans.md の `cc:WIP` → `cc:完了` 状態遷移が grep で整合確認できる | `tests/integration/loop-3cycle.sh` 新設 → 3 タスク連続実行後に Plans.md を grep で検証 | 必須 |
| 2 | **context compaction 発生後の resume 整合性**: `CLAUDE_AUTOCOMPACT_PCT_OVERRIDE=50` で閾値を下げて compaction を強制誘発、その後 resume-pack 再読込で直前タスク ID が復元される | `tests/integration/loop-compaction-resume.sh` 新設 | 必須 |
| 3 | **plateau 検知の golden fixture テスト**: 3 行の calibration レコード（全て非減少 + Jaccard > 0.7）に対して `PIVOT_REQUIRED`、改善した fixture に対して `PIVOT_NOT_REQUIRED`、N<3 に対して `INSUFFICIENT_DATA` | `tests/fixtures/review-calibration/*.jsonl` + unit test | 必須 |
| 4 | **auto-checkpoint が Plans.md 更新直後に発火**: Phase B-5 末尾で `harness-mem` への checkpoint record が 1 件追加される（spike 41.0.0 で確定した API 経由）。併せて `.claude/state/checkpoint-events.jsonl` にローカル audit ログが 1 行追加される（harness-mem 成功/失敗に関わらず必ず書き込み）| `.claude/state/checkpoint-events.jsonl` の件数が cycle 数と一致（harness-mem 側の記録は 41.0.0 で API 確定後に別途検証） | 必須 |
| 5 | **wake-up 回数上限の強制**: 9 サイクル目の wake-up で自動停止 + ユーザーエスカレーション | `tests/integration/loop-max-cycles.sh` | 必須 |
| 6 | **phase-b.lock による race 防止**: 並行 Worker が B-5 実行中は wake-up の checkpoint 呼び出しがブロックされる | `.claude/state/locks/phase-b.lock` の flock テスト | 必須 |
| 7 | **Plans.md flock**: Lead の wake-up と Worker の Plans.md 書き換えが同時発火してもロストアップデートが発生しない | `tests/integration/loop-plans-concurrent.sh` | 必須 |
| 8 | **sprint-contract 後方互換**: `max_iterations` / `rubric_target` / `loop_pacing` フィールドがない旧 contract で default fallback（MAX_REVIEWS=3）が動作 | 旧 contract の re-run テスト | 必須 |
| 9 | **harness-release preflight で Phase 41 スキーマ検証**: `sprint-contract` の新フィールドが schema 違反なら preflight 失敗 | `scripts/release-preflight.sh` に schema validator 追加 | 必須 |
| 10 | **Go test / validate-plugin / check-consistency / check-residue 全パス**: 既存テスト regressions ゼロ | 通常の CI 一式 | 必須 |
| 11 | **mirror 同期**: `codex/.codex/skills/harness-loop/` と `opencode/skills/harness-loop/` が `skills/harness-loop/` と完全一致 | `check-consistency.sh` PASS | 必須 |
| 12 | **試金石: Harness 内部の design-heavy タスクを 3h+ /loop 実行**: `generate-video` スキル UI レビュー相当タスクを 10 反復、browser profile + ui-rubric 組み合わせで完走 | 数値基準: (a) `.claude/state/session-events.jsonl` に完了イベント ≥ 10 件、(b) unexpected abort（exit code ≠ 0）ゼロ、(c) 10 反復全てで Plans.md の `cc:WIP` → `cc:完了` 遷移が grep で整合、(d) 通算実行時間 3h+ を `session-events.jsonl` の timestamp 差分で確認 | 推奨 |

---

### Phase 41.0: Spike + 基盤 [P0]

Purpose: 実装前に 2 つの API（harness-mem checkpoint / `/loop`＋`ScheduleWakeup`）の実在と制約を確認し、plateau 検知の入力源を準備する

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 41.0.0 | **[Spike]** 2 つの API を実環境で確認する。(a) `scripts/harness-mem-client.sh` の実エンドポイント一覧を取得し、checkpoint 相当の操作（`ingest` / `record_event` / `finalize_session` / 新 API のいずれか）を特定、(b) CC が提供する `/loop` コマンドと内部 `ScheduleWakeup` ツールの実制約値（最小/最大 delaySeconds、cron syntax、wake-up ごとの状態継承）を検証、(c) `PreCompact` / `PostCompact` hooks が /loop の wake-up 間で発火するかを実測、(d) 結果を `.claude/memory/decisions.md` に D32「Phase 41 前提の実測結果」として記録 | (a) harness-mem の checkpoint 相当 API が特定されている（またはカスタム実装の必要性が結論付けられている）、(b) `ScheduleWakeup` の [60,3600]s 制約が実測で確認、(c) `/loop <interval>` の最小値が数値で判明（partial: dynamic mode は確定、interval 指定版は未調査）、(d) decisions.md に 4 項目全て記録 + 41.0.2 以降のタスクで使う API 名が確定 | - | cc:完了 [a23c222] |
| 41.0.1 | `scripts/record-review-calibration.sh` L41-67 に新フィールド `critical_count` / `major_count` / `score_delta` を追加。既存レコード（2 件）には影響せず、新規記録からのみ書き込み。`scripts/build-review-few-shot-bank.sh` が新フィールドを読めるよう対応 | 新規 calibration 記録後、jsonl に 3 フィールドが含まれる。旧レコードへの読み出しは `// 0` default で動作。既存 few-shot bank 再生成テストが PASS | 41.0.0 | cc:完了 [f85207a] |
| 41.0.2 | `scripts/auto-checkpoint.sh` を新設。引数: task_id + commit_hash + sprint_contract_path + review_result_path。内部で 41.0.0 で特定した harness-mem API を呼び出して checkpoint を記録。成功・失敗いずれの経路でも `.claude/state/checkpoint-events.jsonl` に 1 行の audit レコード（`{"type":"checkpoint","status":"ok|failed","task":...,"commit":...,"timestamp":...}`）を必ず追記する（ローカル監査ログ）。harness-mem 失敗時は追加で `.claude/state/session-events.jsonl` にもデグレ出力（失敗を静かに吸収しない）。`.claude/state/locks/phase-b.lock` を flock で取得し同期保護 | (a) 正常系: harness-mem に 1 レコード追加 + checkpoint-events.jsonl に 1 行、(b) 異常系: harness-mem API 不達時に checkpoint-events.jsonl に `status:"failed"` 1 行 + session-events.jsonl に `checkpoint_failed` 1 行、(c) phase-b.lock が既に取得されている場合は timeout 10s 待機後に abort、(d) 単体実行 `bash scripts/auto-checkpoint.sh <task> <hash> <contract> <result>` で全挙動 OK、(e) 10 回連続実行でも lock デッドロックなし | 41.0.1 | cc:完了 [eb0ea7b] |
| 41.0.3 | `scripts/detect-review-plateau.sh` を新設。入力: `.claude/state/review-calibration.jsonl` + 現在の task_id。ロジック: 同一 task_id の直近 N=3 エントリを抽出し、(a) iteration 数 ≥ 3、かつ (b) 修正対象ファイル集合の Jaccard 類似度 > 0.7 の両方を満たすなら `PIVOT_REQUIRED` を返す。N<3 なら `INSUFFICIENT_DATA`、条件不成立なら `PIVOT_NOT_REQUIRED`。`tests/fixtures/review-calibration/` に 3 種類の golden fixture を配置してテスト可能にする | (a) golden fixture `plateau.jsonl`（3 行全て Jaccard>0.7） → `PIVOT_REQUIRED`、(b) `improved.jsonl`（最終行で major=0） → `PIVOT_NOT_REQUIRED`、(c) `insufficient.jsonl`（2 行）→ `INSUFFICIENT_DATA`。exit code それぞれ 2 / 0 / 1。`bash scripts/detect-review-plateau.sh <task_id>` 単体で動作 | 41.0.1 | cc:完了 [8f6f787] |

---

### Phase 41.1: /loop 統合（独立 harness-loop スキル） [P0]

Purpose: `skills/harness-loop/` を新設し、起動責務のみを担う薄いスキルとして実装する。内部で `harness-work` を Agent 呼び出しして既存ロジックを再利用

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 41.1.1 | `skills/harness-loop/SKILL.md` + `skills/harness-loop/references/flow.md` を新設。frontmatter: `name: harness-loop`, `description: "長時間タスクを /loop （CC dynamic mode）と ScheduleWakeup で wake-up 毎に fresh context で再入実行。harness-work を内部で Agent 呼び出し。"`, `allowed-tools: [Read, Edit, Bash, Task, ScheduleWakeup]`, `argument-hint: "[all|N-M] [--max-cycles N] [--pacing worker|ci|plateau|night]"`。wake-up 毎のエントリ手順: (1) `scripts/ensure-sprint-contract-ready.sh` で state 健全性確認、(2) harness-mem resume-pack 再読込、(3) Plans.md の `cc:WIP` / `cc:TODO` 確認（41.1.2 で `plans-watcher.sh` に追加される flock 下で実行）、(4) 1 タスクサイクル実行（`harness-work --breezing` を Agent で spawn）、(5) `scripts/detect-review-plateau.sh` で plateau 判定、(6) PIVOT_REQUIRED なら停止 + エスカレーション、そうでなければ `ScheduleWakeup(delaySeconds, prompt="/loop ...")` で次 wake-up を予約。spike 41.0.0 の結果次第で `/loop` user-facing コマンド + `ScheduleWakeup` internal tool の組み合わせを正式採択する | (a) `/harness-loop all` で起動し wake-up が反復発火、(b) 8 サイクル（default）で自動停止、(c) `--max-cycles 3` で 3 サイクル後に停止、(d) pacing 引数で delaySeconds が 270/270/1200/3600 から選択される、(e) SKILL.md 500 行以下、(f) frontmatter の description に「長時間、ループ、loop、wake-up、autonomous」など検索キーワード含む | 41.0.2, 41.0.3 | cc:TODO |
| 41.1.2 | harness-loop の冪等性ガードを実装: (a) `.claude/state/locks/loop-session.lock` で同一セッション内の多重起動を防止、(b) wake-up 冒頭で `bash tests/validate-plugin.sh --quick`（既存に `--quick` がなければ新設、最小限の state 整合性のみチェック）を実行、失敗なら loop 停止、(c) `harness-work` Phase B-5 の `git commit` 直後かつ Plans.md 書き換え**直後**に `scripts/auto-checkpoint.sh` を呼ぶ行を追加（既存の Plans.md 更新行の 1 行後に挿入、破壊的改変を避ける）、(d) **PreCompact 抑制は hooks.json の matcher 変更ではなく**、既存 agent hook の prompt 冒頭に `.claude/state/locks/loop-session.lock` の存在チェックを追加し、lock がある場合は WIP 警告を出力せず即 return する方式（現行 matcher は文字列パターンのみで環境変数否定条件を書けないため）、(e) `scripts/plans-watcher.sh` に flock ガードを新規追加（現状 flock なし、wake-up と Worker 並行書き換えのロストアップデート防止。既存挙動は変えず排他のみ強化） | (a) 2 回目の `/harness-loop` 呼び出しで "already running" エラー、(b) Plans.md を意図的に破損 → wake-up で即停止、(c) 3 サイクル実行後 `.claude/state/checkpoint-events.jsonl` に 3 件、(d) `/harness-loop` セッション中に context compaction を誘発しても PreCompact agent hook が WIP 警告を出さない（lock 検出で suppress）、(e) plans-watcher.sh の flock 下で 2 プロセス同時書き込みテストがロストアップデートなし | 41.1.1 | cc:TODO |

---

### Phase 41.2: 反復制御（MAX_REVIEWS + Browser verdict） [P1]

Purpose: 主観/長サイクル評価に備えて反復上限を可変化し、browser reviewer の verdict を最終判定に組み込む

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 41.2.1 | `MAX_REVIEWS = 3` 直書き（`skills/harness-work/SKILL.md` L263 相当）を可変化。sprint-contract の `review.max_iterations`（新フィールド）で上書き可能に。profile ごとの default: `static:3`, `runtime:3`, `browser:5`, `ui-rubric:10`。`scripts/generate-sprint-contract.sh` の `detectProfile()` 関数（L96-104）でデフォルト値を埋め込む。**未定義時の fallback は必ず 3**（旧 contract 互換） | (a) 旧 contract（`max_iterations` なし）で MAX_REVIEWS=3 が適用、(b) `reviewer_profile: browser` で 5、(c) sprint-contract に `max_iterations: 15` を明示指定すると 15 が優先、(d) 15 回到達で自動停止 + ユーザーエスカレーション、(e) `harness-work` SKILL.md の該当擬似コードが更新されている | 41.0.0 | cc:TODO |
| 41.2.2 | `scripts/run-contract-review-checks.sh` L56-73 の `browser` 分岐を拡張: (a) `browser_runner.sh`（Playwright / agent-browser / chrome-devtools のいずれか既存 detector）を実行して APPROVE/REQUEST_CHANGES を返す、(b) タイムアウト 120s で static verdict にデグレ、(c) `review-result.v1` schema に `browser_verdict` フィールドを追加、(d) `scripts/write-review-result.sh` が `--browser-result` オプションを受け取り、最終 verdict は **static AND browser の AND 結合**（両方 APPROVE なら APPROVE、どちらか REQUEST_CHANGES なら REQUEST_CHANGES、browser が PENDING_BROWSER なら static のみを採用＝既存動作を維持）、(e) `harness-work` SKILL.md L248-256 の擬似コード注記を更新、(f) **回帰テスト**: `tests/unit/browser-verdict-fallback.sh` を新設し、browser=PENDING_BROWSER かつ static=APPROVE の場合に最終 verdict が APPROVE になる（既存の commit guard が `.verdict == "APPROVE"` を参照する前提を崩さない）ことを明示検証 | (a) Playwright が動く環境で browser profile タスクを実行 → browser_verdict が review-result.json に記録、(b) browser runner が 120s で timeout → static にデグレ + log 記録、(c) Playwright 未インストール環境 → PENDING_BROWSER 維持（既存動作）、(d) 既存の static/runtime/security profile タスクに regressions なし、(e) review-result schema v1 のバージョンは据え置き（後方互換フィールド追加のため）、(f) 回帰テスト PASS（commit guard が既存通り動作） | 41.0.2, 41.1.1 | cc:TODO |

---

### Phase 41.3: UI rubric profile [P2]

Purpose: 記事の Frontend Design Loop に倣い、主観品質評価の 4 軸ルーブリックを追加する

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 41.3.1 | `skills/harness-review/references/ui-rubric.md` を新設。構成は `security-profile.md` に倣う。内容: (a) 4 軸 Design Quality / Originality / Craft / Functionality を 0-10 で採点、(b) 各軸に**アンカー例**（0/5/10 の具体的判定基準）を記載、(c) sprint-contract の `review.rubric_target`（新フィールド、例: `{design:7, originality:6, craft:8, functionality:9}`）と照合し、1 軸でも target 未達なら REQUEST_CHANGES、(d) `skills/harness-review/SKILL.md` の決定木（L33-38）に `--ui-rubric` 分岐を追加、(e) `scripts/generate-sprint-contract.sh` の `detectProfile()` に `ui-rubric` 検出パターン追加（タスク内容に「design」「UI」「styling」「aesthetic」「layout」が含まれる場合）、(f) mirror 同期（`codex/.codex/skills/harness-review/references/ui-rubric.md` + `opencode/...`） | (a) `/harness-review --ui-rubric` で 4 軸採点が実行される、(b) sprint-contract に `rubric_target` があれば閾値判定、なければ default threshold=6、(c) 各軸のアンカー例が非専門家にも判定可能な日本語で書かれている、(d) mirror 3 箇所完全一致（`check-consistency.sh` PASS） | 41.2.1 | cc:TODO |

---

### Phase 41.4: ドキュメント + mirror + preflight [P0]

Purpose: Phase 41 の新機能を運用可能にし、リリース時の schema drift を preflight で catch する

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 41.4.1 | `docs/long-running-harness.md` を新規作成。内容: (1) 記事の要約と Harness との対応表（B1-B12 の 12 軸）、(2) /loop + ScheduleWakeup の使い方（cheatsheet）、(3) pacing プリセットの選び方（cache 境界の説明込み）、(4) wake-up 回数上限・lock・冪等性ガードの仕組み、(5) plateau 検知の閾値と golden fixture の配置、(6) **Phase 41 のスコープ明示**（同一セッション内限定、ホスト跨ぎは Phase 42 以降）、(7) 既知の制約（`bypassPermissions` との併用ガイド、Plans.md flock の限界）、(8) `CLAUDE.md` に docs/long-running-harness.md への reference 1 行追加 | ファイル存在、markdown valid、非専門家が「/loop で長時間タスクを実行する方法」を 10 分で理解できる。CLAUDE.md に reference 追加 | 41.1.2 | cc:TODO |
| 41.4.2 | `tests/integration/` 配下に 4 本のテストを追加: (a) `loop-3cycle.sh` — DoD #1 検証、(b) `loop-compaction-resume.sh` — DoD #2、(c) `loop-max-cycles.sh` — DoD #5、(d) `loop-plans-concurrent.sh` — DoD #7。`tests/validate-plugin.sh` に integration セクションを追加し、これら 4 本を optional category として集計（fail しても既存の required test 集計には影響させない） | (a) 4 本とも単体で PASS、(b) validate-plugin.sh から呼び出しても PASS、(c) 意図的に /loop ロジックを壊すと少なくとも 1 本が FAIL、(d) 実行時間合計 10 分以内 | 41.1.2 | cc:TODO |
| 41.4.3 | `scripts/release-preflight.sh` に sprint-contract schema validator を追加。`.claude/state/contracts/*.sprint-contract.json` をスキャンし、(a) 新フィールド `max_iterations` / `rubric_target` / `loop_pacing` / `browser_verdict` の型が正しい、(b) `reviewer_profile` が `static|runtime|browser|security|ui-rubric` のいずれか（**security は既存出力のため必ず許容**、ui-rubric は 41.3.1 未実装時も将来互換のため許容）、(c) `max_iterations` が 1-30 の範囲内、を検証。違反があれば preflight 失敗。mirror 同期: `skills/harness-loop/` を `codex/.codex/skills/` と `opencode/skills/` に mirror、`check-consistency.sh` が mirror 完全一致を確認 | (a) 意図的に `max_iterations: 100` を contract に埋め込む → preflight 失敗、(b) 正常な contract で PASS、(c) `reviewer_profile: "security"` の既存 contract で regressions なし（preflight PASS）、(d) `reviewer_profile: "ui-rubric"` の contract も PASS（41.3.1 未実装でも許容）、(e) `codex/.codex/skills/harness-loop/SKILL.md` が `skills/harness-loop/SKILL.md` と完全一致、(f) `check-consistency.sh` 全 PASS | 41.1.1 | cc:TODO |

---

### ロードマップ（着手順）

推奨実行順（依存グラフに基づく）:

1. **Week 1**: 41.0.0（Spike、2-3 日）→ 判明事項を decisions.md に記録
2. **Week 1-2**: 41.0.1 → 41.0.2 / 41.0.3（並列可）
3. **Week 2**: 41.1.1 → 41.1.2
4. **Week 3**: 41.2.1 / 41.2.2（並列可、41.2.1 を先）
5. **Week 3-4**: 41.3.1（Optional、時間があれば）
6. **Week 4**: 41.4.1 / 41.4.2 / 41.4.3（並列可）
7. **Week 4-5**: DoD #12（試金石タスク）を手動セッションで実行、学びを `.claude/memory/patterns.md` に記録

### Non-Goals（Phase 41 でやらないこと）

- ホスト CC プロセス終了をまたぐ /loop 継続（tmux 常駐、systemd daemon 化）→ Phase 42 で検討
- harness-mem 側への新 MCP エンドポイント追加（41.0.0 spike で既存 API に相当が見つからなかった場合、Phase 42 で別途切り出し）
- UI rubric の自動採点の LLM 較正ループ（ui-rubric profile の採点ばらつきの calibration 自動化）→ Phase 43 以降
- `/loop` + `--codex`（Codex CLI モード）の組み合わせ検証 → Phase 42 で別途

---
