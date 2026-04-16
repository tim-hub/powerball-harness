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

### Phase 37.5: hooks.json 最終書き換え + レガシー削除 [cc:完了]

| Task | 内容 | DoD | Status |
|------|------|-----|--------|
| 37.5.1 | hooks.json の残り 37 エントリを全て `bin/harness hook <name>` に書き換え | `grep -c 'run-hook.sh' hooks/hooks.json` が 0 | cc:完了 |
| 37.5.2 | `scripts/run-hook.sh` + `scripts/run-script.js` 削除 | ファイルが存在しない | cc:完了 |
| 37.5.3 | `package.json` 削除 (npm 依存の完全排除) | ファイルが存在しない | cc:完了 |
| 37.5.4 | `core/` (TypeScript エンジン) 削除 | ディレクトリが存在しない | cc:完了 |
| 37.5.5 | E2E テスト: 全フックイベントが Go binary 経由で動作 | `go/test-e2e.sh` 全パス | cc:完了 |
| 37.5.6 | `harness doctor` が Node.js 依存ゼロを確認 | `grep -rE "node\|run-script" hooks/` が 0 件 | cc:完了 |

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

## 📦 アーカイブ

完了済み Phase は以下のファイルへ切り出し済み（git history にも残存）:

- [Phase 39 + 40 + 41.0](.claude/memory/archive/Plans-2026-04-15-phase39-40-41.0.md) — レビュー体験改善 / Migration Residue Scanner / Long-Running Harness Spike

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

### Phase 41.1: /loop 統合（独立 harness-loop スキル） [P0]

Purpose: `skills/harness-loop/` を新設し、起動責務のみを担う薄いスキルとして実装する。内部で `harness-work` を Agent 呼び出しして既存ロジックを再利用

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 41.1.1 | `skills/harness-loop/SKILL.md` + `skills/harness-loop/references/flow.md` を新設。frontmatter: `name: harness-loop`, `description: "長時間タスクを /loop （CC dynamic mode）と ScheduleWakeup で wake-up 毎に fresh context で再入実行。harness-work を内部で Agent 呼び出し。"`, `allowed-tools: [Read, Edit, Bash, Task, ScheduleWakeup]`, `argument-hint: "[all|N-M] [--max-cycles N] [--pacing worker|ci|plateau|night]"`。wake-up 毎のエントリ手順: (1) `scripts/ensure-sprint-contract-ready.sh` で state 健全性確認、(2) harness-mem resume-pack 再読込、(3) Plans.md の `cc:WIP` / `cc:TODO` 確認（41.1.2 で `plans-watcher.sh` に追加される flock 下で実行）、(4) 1 タスクサイクル実行（`harness-work --breezing` を Agent で spawn）、(5) `scripts/detect-review-plateau.sh` で plateau 判定、(6) PIVOT_REQUIRED なら停止 + エスカレーション、そうでなければ `ScheduleWakeup(delaySeconds, prompt="/loop ...")` で次 wake-up を予約。spike 41.0.0 の結果次第で `/loop` user-facing コマンド + `ScheduleWakeup` internal tool の組み合わせを正式採択する | (a) `/harness-loop all` で起動し wake-up が反復発火、(b) 8 サイクル（default）で自動停止、(c) `--max-cycles 3` で 3 サイクル後に停止、(d) pacing 引数で delaySeconds が 270/270/1200/3600 から選択される、(e) SKILL.md 500 行以下、(f) frontmatter の description に「長時間、ループ、loop、wake-up、autonomous」など検索キーワード含む | 41.0.2, 41.0.3 | cc:完了 [f6c719e] |
| 41.1.2 | harness-loop の冪等性ガードを実装: (a) `.claude/state/locks/loop-session.lock` で同一セッション内の多重起動を防止、(b) wake-up 冒頭で `bash tests/validate-plugin.sh --quick`（既存に `--quick` がなければ新設、最小限の state 整合性のみチェック）を実行、失敗なら loop 停止、(c) `harness-work` Phase B-5 の `git commit` 直後かつ Plans.md 書き換え**直後**に `scripts/auto-checkpoint.sh` を呼ぶ行を追加（既存の Plans.md 更新行の 1 行後に挿入、破壊的改変を避ける）、(d) **PreCompact 抑制は hooks.json の matcher 変更ではなく**、既存 agent hook の prompt 冒頭に `.claude/state/locks/loop-session.lock` の存在チェックを追加し、lock がある場合は WIP 警告を出力せず即 return する方式（現行 matcher は文字列パターンのみで環境変数否定条件を書けないため）、(e) `scripts/plans-watcher.sh` に flock ガードを新規追加（現状 flock なし、wake-up と Worker 並行書き換えのロストアップデート防止。既存挙動は変えず排他のみ強化） | (a) 2 回目の `/harness-loop` 呼び出しで "already running" エラー、(b) Plans.md を意図的に破損 → wake-up で即停止、(c) 3 サイクル実行後 `.claude/state/checkpoint-events.jsonl` に 3 件、(d) `/harness-loop` セッション中に context compaction を誘発しても PreCompact agent hook が WIP 警告を出さない（lock 検出で suppress）、(e) plans-watcher.sh の flock 下で 2 プロセス同時書き込みテストがロストアップデートなし | 41.1.1 | cc:完了 [79a6248] |
| 41.1.3 | harness-loop の Lead レビュー実装を実動作可能にする（41.1.1 再レビューで発見された Step 5.5/5.6 の 3 実装ギャップ）: (a) **Codex companion review で Worker feature branch 差分を正しくレビュー**: `scripts/codex-companion.sh review` は `--diff` を受け付けないため、Lead が worker_result.worktreePath に `cd` してから review を呼ぶか、`git fetch <worker_path> <branch>` で Lead 側に取り込んでからレビューする方式に変更（skills/harness-loop/references/flow.md Step 5.5 を修正）、(b) **reviewer_profile 分岐を harness-loop にも追加**: sprint-contract の `reviewer_profile == "runtime"` なら `scripts/run-contract-review-checks.sh` を追加実行、`== "browser"` なら `scripts/generate-browser-review-artifact.sh` を実行（既存 breezing Phase B-3 の分岐ロジックを harness-loop Step 5.5 にも移植）、(c) **worktree remove → branch -D の順序修正**: `isolation="worktree"` で checkout された feature branch は `git branch -D` で削除できない（`branch is checked out at ...` エラー）。APPROVE 後の cleanup 手順を「cherry-pick → `git worktree remove <path>` → `git branch -D <branch>`」の順に変更（flow.md Step 5.6） | (a) Lead レビューが Worker feature branch の**実際の差分**を見て verdict を返す（main の空 diff ではない）、(b) `reviewer_profile: "runtime"` の contract で `run-contract-review-checks.sh` が実行され verdict に反映、`reviewer_profile: "browser"` で browser artifact が生成、(c) `/harness-loop` の 2 タスク連続実行で branch 削除エラーなく完走（現行の `git branch -D` 失敗問題が再発しない）、(d) **validate-plugin.sh `--quick` の jq フォールバック**: `command -v jq` ガードなしで `jq empty` 実行しているため、macOS 素の環境で contract が 1 件でもあると誤判定 → Python フォールバック or `command -v jq` ガード追加（41.1.2 の 3 回目 Codex レビューで指摘）、(e) **plans-watcher.sh の fail-closed 化**: lock 取得タイムアウト時に warning だけ出して処理続行（fail-open）のため 5 秒超の競合で plans-state.json の read-modify-write が無保護。abort or retry に変更（同レビューで指摘） | (a)(c)(d)(e) 達成、(b) は 41.1.4 へ繰越（CONTRACT_PATH 上書きバグ発見） | 41.1.1 | cc:完了 [4fcb9df] |
| 41.1.4 | 41.1.3 の 3 回目 Codex レビューで発見された残フィンディング 3 件を解消: (a) **Go `writePMNotification` の CWD 整合**（P1）: `go/internal/hookhandler/plans_watcher.go` の `writePMNotification` 関数が `os.Getwd()` を使って pm-notifications パスを決定している。`HandlePlansWatcher` が採用した `input.CWD` ベースのパス決定と食い違うため、plans-state.json と pm-notifications が異なる worktree に分裂する可能性あり。`writePMNotification(cwd string, ...)` に signature 変更 + 呼び出し側で `input.CWD` を渡す、(b) **Go plans-watcher の lock fallback 追加**（P2）: 現行の `syscall.Flock` のみの実装では NFS/SMB マウント上で silent no-op になる。bash 側の 3 段 fallback（flock → lockf → mkdir）と対応させ、Go 側にも `mkdir` atomic fallback を追加（Flock 失敗時のみ）、(c) **flow.md Step 2 の CONTRACT_PATH 上書きバグ**（P1, DoD (b) 無効化）: skills/harness-loop/references/flow.md Step 2 で `CONTRACT_PATH=$(...enrich...)` の後に `CONTRACT_PATH="${task_contract_path:-}"` と書いており、`task_contract_path` 変数は定義されていないため空文字で上書きされる。結果として Step 5.5 の `reviewer_profile` 分岐が発火しない。該当行を削除または正しい変数参照に修正 | (a) `writePMNotification` が `input.CWD` ベースのパスを使い、`HandlePlansWatcher` の状態ファイルと同じ worktree に書かれる（Go unit test で検証）、(b) Flock が失敗する環境（モック）で mkdir fallback が発火して fail-closed、(c) flow.md Step 2 で CONTRACT_PATH が enrich の戻り値のまま保持され、browser profile タスクで `generate-browser-review-artifact.sh` が実行される（integration test で検証） | 41.1.3 | cc:完了 [6902de2] |

---

### Phase 41.2: 反復制御（MAX_REVIEWS + Browser verdict） [P1]

Purpose: 主観/長サイクル評価に備えて反復上限を可変化し、browser reviewer の verdict を最終判定に組み込む

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 41.2.1 | `MAX_REVIEWS = 3` 直書き（`skills/harness-work/SKILL.md` L263 相当）を可変化。sprint-contract の `review.max_iterations`（新フィールド）で上書き可能に。profile ごとの default: `static:3`, `runtime:3`, `browser:5`, `ui-rubric:10`。`scripts/generate-sprint-contract.js` の `detectProfile()` 関数（L96-104）でデフォルト値を埋め込む。**未定義時の fallback は必ず 3**（旧 contract 互換） | (a) 旧 contract（`max_iterations` なし）で MAX_REVIEWS=3 が適用、(b) `reviewer_profile: browser` で 5、(c) Plans.md のタスク行 DoD にインラインコード ``<!-- max_iterations: N -->`` を追記すると N が優先される（N は 1-30 の整数で、説明文中の `N` は例示なので browser default を誤上書きしない）、(d) 指定回数到達で自動停止 + ユーザーエスカレーション、(e) `harness-work` SKILL.md の該当擬似コードが更新されている | 41.0.0 | cc:完了 [a95132ab] |
| 41.2.2 | `scripts/run-contract-review-checks.sh` L56-73 の `browser` 分岐を拡張: (a) `browser_runner.sh`（Playwright / agent-browser / chrome-devtools のいずれか既存 detector）を実行して APPROVE/REQUEST_CHANGES を返す、(b) タイムアウト 120s で static verdict にデグレ、(c) `review-result.v1` schema に `browser_verdict` フィールドを追加、(d) `scripts/write-review-result.sh` が `--browser-result` オプションを受け取り、最終 verdict は **static AND browser の AND 結合**（両方 APPROVE なら APPROVE、どちらか REQUEST_CHANGES なら REQUEST_CHANGES、browser が PENDING_BROWSER なら static のみを採用＝既存動作を維持）、(e) `harness-work` SKILL.md L248-256 の擬似コード注記を更新、(f) **回帰テスト**: `tests/unit/browser-verdict-fallback.sh` を新設し、browser=PENDING_BROWSER かつ static=APPROVE の場合に最終 verdict が APPROVE になる（既存の commit guard が `.verdict == "APPROVE"` を参照する前提を崩さない）ことを明示検証 | (a) Playwright が動く環境で browser profile タスクを実行 → browser_verdict が review-result.json に記録、(b) browser runner が 120s で timeout → static にデグレ + log 記録、(c) Playwright 未インストール環境 → PENDING_BROWSER 維持（既存動作）、(d) 既存の static/runtime/security profile タスクに regressions なし、(e) review-result schema v1 のバージョンは据え置き（後方互換フィールド追加のため）、(f) 回帰テスト PASS（commit guard が既存通り動作） | 41.0.2, 41.1.1 | cc:完了 [67499d43] |

---

### Phase 41.3: UI rubric profile [P2]

Purpose: 記事の Frontend Design Loop に倣い、主観品質評価の 4 軸ルーブリックを追加する

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 41.3.1 | `skills/harness-review/references/ui-rubric.md` を新設。構成は `security-profile.md` に倣う。内容: (a) 4 軸 Design Quality / Originality / Craft / Functionality を 0-10 で採点、(b) 各軸に**アンカー例**（0/5/10 の具体的判定基準）を記載、(c) sprint-contract の `review.rubric_target`（新フィールド、例: `{design:7, originality:6, craft:8, functionality:9}`）と照合し、1 軸でも target 未達なら REQUEST_CHANGES、(d) `skills/harness-review/SKILL.md` の決定木（L33-38）に `--ui-rubric` 分岐を追加、(e) `scripts/generate-sprint-contract.js` の `detectProfile()` に `ui-rubric` 検出パターン追加（タスク内容に「design」「UI」「styling」「aesthetic」「layout」が含まれる場合）、(f) mirror 同期（`codex/.codex/skills/harness-review/references/ui-rubric.md` + `opencode/...`） | (a) `/harness-review --ui-rubric` で 4 軸採点が実行される、(b) sprint-contract に `rubric_target` があれば閾値判定、なければ default threshold=6、(c) 各軸のアンカー例が非専門家にも判定可能な日本語で書かれている、(d) mirror 3 箇所完全一致（`check-consistency.sh` PASS） | 41.2.1 | cc:完了 [6902de2] |

---

### Phase 41.4: ドキュメント + mirror + preflight [P0]

Purpose: Phase 41 の新機能を運用可能にし、リリース時の schema drift を preflight で catch する

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 41.4.1 | `docs/long-running-harness.md` を新規作成。内容: (1) 記事の要約と Harness との対応表（B1-B12 の 12 軸）、(2) /loop + ScheduleWakeup の使い方（cheatsheet）、(3) pacing プリセットの選び方（cache 境界の説明込み）、(4) wake-up 回数上限・lock・冪等性ガードの仕組み、(5) plateau 検知の閾値と golden fixture の配置、(6) **Phase 41 のスコープ明示**（同一セッション内限定、ホスト跨ぎは Phase 42 以降）、(7) 既知の制約（`bypassPermissions` との併用ガイド、Plans.md flock の限界）、(8) `CLAUDE.md` に docs/long-running-harness.md への reference 1 行追加 | ファイル存在、markdown valid、非専門家が「/loop で長時間タスクを実行する方法」を 10 分で理解できる。CLAUDE.md に reference 追加 | 41.1.2 | cc:完了 [6902de2] |
| 41.4.2 | `tests/integration/` 配下に 4 本のテストを追加: (a) `loop-3cycle.sh` — DoD #1 検証、(b) `loop-compaction-resume.sh` — DoD #2、(c) `loop-max-cycles.sh` — DoD #5、(d) `loop-plans-concurrent.sh` — DoD #7。`tests/validate-plugin.sh` に integration セクションを追加し、これら 4 本を optional category として集計（fail しても既存の required test 集計には影響させない） | (a) 4 本とも単体で PASS、(b) validate-plugin.sh から呼び出しても PASS、(c) 意図的に /loop ロジックを壊すと少なくとも 1 本が FAIL、(d) 実行時間合計 10 分以内 | 41.1.2 | cc:完了 [6902de2] |
| 41.4.3 | `scripts/release-preflight.sh` に sprint-contract schema validator を追加。`.claude/state/contracts/*.sprint-contract.json` をスキャンし、(a) 新フィールド `max_iterations` / `rubric_target` / `loop_pacing` / `browser_verdict` の型が正しい、(b) `reviewer_profile` が `static|runtime|browser|security|ui-rubric` のいずれか（**security は既存出力のため必ず許容**、ui-rubric は 41.3.1 未実装時も将来互換のため許容）、(c) `max_iterations` が 1-30 の範囲内、を検証。違反があれば preflight 失敗。mirror 同期: `skills/harness-loop/` を `codex/.codex/skills/` と `opencode/skills/` に mirror、`check-consistency.sh` が mirror 完全一致を確認 | (a) 意図的に `max_iterations: 100` を contract に埋め込む → preflight 失敗、(b) 正常な contract で PASS、(c) `reviewer_profile: "security"` の既存 contract で regressions なし（preflight PASS）、(d) `reviewer_profile: "ui-rubric"` の contract も PASS（41.3.1 未実装でも許容）、(e) `codex/.codex/skills/harness-loop/SKILL.md` が `skills/harness-loop/SKILL.md` と完全一致、(f) `check-consistency.sh` 全 PASS | 41.1.1 | cc:完了 [6902de2] |

---

### Phase 41.5: ランタイム整合性 cleanup [P1]

Purpose: Hokage 移行以降も残った node 系周辺スクリプトについて、(a) 嘘拡張子の修正（bug）、(b) Go 化対象の評価（policy）を行う

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 41.5.1 | `scripts/generate-sprint-contract.js` へリネームして嘘拡張子を解消。旧ファイルは `#!/usr/bin/env node` なのに shell 風の名前だったため、`bash` 実行で `syntax error near unexpected token (`` と失敗した。(a) ファイルを `scripts/generate-sprint-contract.js` にリネーム（`git mv`）、(b) 呼び出し元 11 ファイルの参照を `.sh` → `.js` に一括更新（`skills/harness-work/SKILL.md`、`skills/harness-loop/references/flow.md`、`opencode/skills/harness-work/SKILL.md`、`skills-codex/harness-work/SKILL.md`、`skills-codex/breezing/SKILL.md`、`codex/.codex/skills/harness-work/SKILL.md`、`tests/test-sprint-contract-approval.sh`、`tests/test-generate-sprint-contract.sh`、`tests/test-generate-browser-review-artifact.sh`、`Plans.md`、`scripts/release-preflight.sh` 等）、(c) 呼び出し側が旧 shell 風コマンドなら `node scripts/generate-sprint-contract.js` に変更、(d) `check-consistency.sh` で mirror の一括同期を確認 | (a) 旧 shell 風ファイル名が存在しない（not found）、(b) `node scripts/generate-sprint-contract.js <task-id>` で contract が従来通り生成される、(c) `tests/test-generate-sprint-contract.sh` が PASS、(d) 旧 shell 風ファイル名の参照が source 上 0 件、(e) `check-consistency.sh` が PASS | - | cc:完了 [6902de2] |
| 41.5.2 | scripts/*.js（7 件: `record-usage.js`, `emit-agent-trace.js`, `generate-agent-telemetry.js`, `migrate-usage-history.js`, `build-opencode.js`, `validate-opencode.js`, `hook-handlers/pre-compact-save.js` + 41.5.1 で改名された `generate-sprint-contract.js`）の Go 移植評価を実施。(a) 各スクリプトの役割・呼び出し頻度・Go 化の ROI を 1 行で表にまとめる、(b) 「Hokage 原則との整合性」軸で A（即 Go 化）/ B（保留）/ C（Node のまま維持）に分類、(c) A 判定のものだけ後続 Phase のタスクとして `Plans.md` に追加提案（実装はこのタスクではやらない。評価のみ）、(d) 結果を `.claude/memory/decisions.md` の新規 decision として記録（D33 想定） | (a) 8 件すべてに A/B/C の分類が付く、(b) 評価表を含む decision が `.claude/memory/decisions.md` に追記される、(c) A 判定が 1 件以上あれば後続タスク案を Plans.md 末尾にコメント付きで提示（実装追加はユーザー承認後） | 41.5.1 | cc:完了 [6902de2] |

---

### ロードマップ（着手順）

推奨実行順（依存グラフに基づく）:

1. **Week 1**: 41.0.0（Spike、2-3 日）→ 判明事項を decisions.md に記録
2. **Week 1-2**: 41.0.1 → 41.0.2 / 41.0.3（並列可）
3. **Week 2**: 41.1.1 → 41.1.2
4. **Week 3**: 41.2.1 / 41.2.2（並列可、41.2.1 を先）
5. **Week 3-4**: 41.3.1（Optional、時間があれば）
6. **Week 4**: 41.4.1 / 41.4.2 / 41.4.3（並列可）

<!-- 41.5.2 follow-up proposals (evaluation only; do not implement without approval)
- 41.5.3: `scripts/generate-sprint-contract.js` を Go 側 helper / hookhandler に寄せ、profile 判定と contract schema を Hokage 側で一元管理する
- 41.5.4: `scripts/emit-agent-trace.js` と `scripts/hook-handlers/pre-compact-save.js` を Go hookhandler に統合し、long-running session の hot path を Node 依存から外す
-->
7. **Week 4-5**: DoD #12（試金石タスク）を手動セッションで実行、学びを `.claude/memory/patterns.md` に記録

---

### Phase 42: Go hot-path migration [P1]

Purpose: D33 で A 判定になった「長時間実行の本線」を、優先順を固定したうえで Go 側に寄せる。まず sprint-contract の SSOT を Go に集約し、その後に trace と pre-compact を hookhandler に統合する。

推奨実行順（この順に着手）:

1. `42.1.1` `generate-sprint-contract.js`
2. `42.1.2` `emit-agent-trace.js`
3. `42.1.3` `pre-compact-save.js`

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 42.1.1 | `scripts/generate-sprint-contract.js` を Go helper / hookhandler 側へ移植し、reviewer_profile / `max_iterations` / `rubric_target` / browser route 判定の SSOT を Go に一本化する。JS 側は削除または薄いラッパーにし、`harness-work` / `harness-loop` / preflight / test 群の参照を Go 実装へ切り替える | (a) `node scripts/generate-sprint-contract.js <task-id>` と同等の contract を Go 実装が生成し、既存テスト `tests/test-generate-sprint-contract.sh` / `tests/unit/max-iterations-default.sh` / `tests/test-generate-browser-review-artifact.sh` / `tests/test-sprint-contract-approval.sh` が回帰なしで PASS、(b) Go unit test で profile 判定・HTML コメント override・ui-rubric default target・browser route 判定を直接検証、(c) 旧 JS 実装が SSOT ではなくなったことが source 上で明確（薄い wrapper なら Go 呼び出しだけに縮退） | - | cc:完了 |
| 42.1.2 | `scripts/emit-agent-trace.js` を Go hookhandler に統合し、PostToolUse の trace 記録と OTel 送信を Go 側で処理する。repo root 推定、JSONL rotation、VCS 情報取得、OTLP 送信の責務を Go 実装へ寄せる | (a) PostToolUse 相当の入力 fixture から `.claude/state/agent-trace.jsonl` が従来互換で出力される、(b) OTel endpoint 指定時の non-blocking 挙動と timeout を Go unit / integration test で検証、(c) `tests/validate-plugin.sh` と trace 周辺の既存回帰が PASS、(d) JS 版は削除または Go 実装を呼ぶ薄い wrapper のみになり、hot path の本体ロジックが Go 側へ移っている | 42.1.1 | cc:完了 |
| 42.1.3 | `scripts/hook-handlers/pre-compact-save.js` を Go hookhandler に統合し、PreCompact 時の handoff-artifact / snapshot 保存を Go 側で一元化する。Plans.md の WIP 抽出、state 読み込み、artifact schema 管理を Go に寄せる | (a) PreCompact fixture から `handoff-artifact.json` と `precompact-snapshot.json` が従来互換で生成される、(b) compaction 前後の long-running session 回帰を integration test で確認し、`tests/integration/loop-compaction-resume.sh` が PASS、(c) Go unit test で Plans.md の WIP 抽出・欠損 state file・壊れた JSON の耐障害性を検証、(d) JS 版は削除または薄い wrapper のみになり、compaction hot path の本体ロジックが Go 側へ移っている | 42.1.2 | cc:完了 |

### Non-Goals（Phase 41 でやらないこと）

- ホスト CC プロセス終了をまたぐ /loop 継続（tmux 常駐、systemd daemon 化）→ Phase 42 で検討
- harness-mem 側への新 MCP エンドポイント追加（41.0.0 spike で既存 API に相当が見つからなかった場合、Phase 42 で別途切り出し）
- UI rubric の自動採点の LLM 較正ループ（ui-rubric profile の採点ばらつきの calibration 自動化）→ Phase 43 以降
- `/loop` + `--codex`（Codex CLI モード）の組み合わせ検証 → Phase 42 で別途

---
