# Claude Code Harness — Plans.md

最終アーカイブ: 2026-03-08（Phase 17〜24 → `.claude/memory/archive/Plans-2026-03-08-phase17-24.md`）

---

## Phase 35: Harness v4 — Go ゼロベース再構築

作成日: 2026-04-05
目的: 127 本のシェルスクリプト + TypeScript コアを単一 Go バイナリに統合し、フック応答を 5ms 以下に短縮、設定ファイルの二重管理を解消する

設計詳細: [go/DESIGN.md](go/DESIGN.md)

### 設計方針

- CC プラグインプロトコル準拠を最優先。`plugin.json`, `hooks.json`, `settings.json`, `agents/*.md`, `skills/*/SKILL.md` の公式形式を維持
- `harness.toml` はユーザーが編集する唯一のファイル。CC 必須ファイルは `harness sync` で自動生成
- Phase ごとに E2E 検証を通し、未移行スクリプトは既存シムで動作継続
- `bin/harness` を CC プラグインの PATH 経由で解決

### 優先度マトリクス

| 優先度 | Phase | 内容 | タスク数 | 依存 |
|--------|-------|------|---------|------|
| **Required** | 35.0 | プロトコル + ガードレール（最小 MVP） | 4 | なし |
| **Required** | 35.1 | SQLite 状態層 | 3 | 35.0 |
| **Required** | 35.2 | 統合設定 (harness.toml → CC ファイル生成) | 4 | 35.1 |
| **Recommended** | 35.3 | ハンドラ統合 (127 スクリプト段階的吸収) | 5 | 35.2 |
| **Recommended** | 35.4 | エージェントライフサイクル状態マシン | 3 | 35.3 |
| **Recommended** | 35.5 | スキル検証 + SKILL.md バリデータ | 2 | 35.2 |
| **Optional** | 35.6 | Breezing 並行処理 (goroutine/worktree) | 3 | 35.4 |
| **Optional** | 35.7 | npm 配布 + クロスコンパイル | 3 | 35.6 |

合計: **27 タスク**

### 完成基準 (Definition of Done — Phase 35 全体)

Phase 35 は以下の **全項目** を満たした時点で完成とする。

| # | 基準 | 検証方法 | 必須/推奨 |
|---|------|---------|----------|
| 1 | **Node.js ランタイム依存ゼロ**: Go 正本 hook から `node`/`core/dist` 参照なし。allowlist: codex-companion.sh, 未移行 scripts/*.js | `grep -rE "node\|core/dist" hooks/` + allowlist 照合 | 必須 |
| 2 | **フック応答 p99 < 10ms**: PreToolUse の最も重い経路 (SQLite 参照あり) | `hyperfine` 100 回 (空DB/肥大化DB/競合DB) | 必須 |
| 3 | **ガードレールパリティ**: R01-R13 の全テストケースが Go テストで PASS | `go test ./internal/guard/...` | 必須 |
| 4 | **公式プロトコル準拠**: Protocol Truth Table (SPEC.md §2) の documented フィールドが動作。experimental は未実装、unknown は無視 | E2E テストで各フィールド検証 | 必須 |
| 5 | **設定一元化**: `harness.toml` → `harness sync` → CC 正常動作 | `harness sync && validate-plugin.sh` PASS | 必須 |
| 6 | **dual hooks.json 解消**: `harness sync` で hooks.json + .claude-plugin/hooks.json 自動同期 | `check-consistency.sh` PASS | 必須 |
| 7 | **既存スキル・エージェント動作**: 全 30+ スキル + 3 エージェントが Go 環境で正常動作 | hook event matrix による E2E (全 hook event カバー) | 必須 |
| 8 | **State 移行整合性**: 旧 state.db → 新パス移行が可逆。export/import + rollback toggle | `harness doctor --migration` PASS | 必須 |
| 9 | **スクリプト移行率 80%+**: 127 本中 100 本以上が Go サブコマンドに吸収 | `harness doctor` 移行レポート | 推奨 |
| 10 | **クロスプラットフォームビルド**: darwin-arm64, darwin-amd64, linux-amd64 | CI クロスコンパイル成功 | 推奨 |
| 11 | **バイナリサイズ < 10MB**: strip + 最適化済み | `ls -lh bin/harness` | 推奨 |

**最小完成条件**: 基準 1-8 (必須 8 項目) を全て満たせば v4.0.0 リリース可能。基準 9-11 は v4.1.0 以降で達成可。

詳細仕様: [go/SPEC.md](go/SPEC.md)

---

### Phase 35.0: プロトコル + ガードレール [P0]

Purpose: `pre-tool.sh` → `node core/dist/index.js` の呼び出しチェーンを `harness hook pre-tool` 単一バイナリに置換し、p99 < 10ms を達成する

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 35.0.1 | Go モジュール初期化 + `pkg/protocol/types.go` で公式 stdin/stdout JSON スキーマに完全準拠した型定義 | `transcript_path`, `permission_mode`, `hook_event_name`, `defer`, `updatedInput`, `additionalContext` が型に含まれる | - | cc:完了 |
| 35.0.2 | `internal/guard/rules.go` に R01-R13 全ルールを 1:1 移植 + `internal/hook/codec.go` で stdin パーサー実装 | 58 テスト全 PASS | 35.0.1 | cc:完了 |
| 35.0.3 | `cmd/harness/main.go` CLI + PreToolUse/PostToolUse/PermissionRequest ハンドラ + `bin/harness` ビルド | E2E 8 シナリオ PASS、p99 5ms | 35.0.2 | cc:完了 |
| 35.0.4 | hook shim 3 本 (`pre-tool.sh`, `post-tool.sh`, `permission.sh`) を Go バイナリ直呼びに書き換え | Node.js フォールバック削除、バイナリ未発見時は明確なエラー | 35.0.3 | cc:完了 |

---

### Phase 35.1: SQLite 状態層 [P0]

Purpose: `core/src/state/` の Go 移植 + `${CLAUDE_PLUGIN_DATA}` 永続ストレージ活用

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 35.1.1 | `internal/state/schema.go` — 既存 DDL (sessions, signals, task_failures, work_states, schema_meta) + assumptions テーブル追加 | マイグレーション付きスキーマ初期化が動作 | 35.0.3 | cc:完了 |
| 35.1.2 | `internal/state/store.go` — HarnessStore の Go 移植 (WAL mode, busy timeout 5s) | 3 goroutine 並列 INSERT/SELECT でデッドロックなし | 35.1.1 | cc:完了 |
| 35.1.3 | `pre_tool.go` の BuildContext に SQLite work_states 参照を統合 | session_id 付き入力で DB から codexMode/workMode を取得 | 35.1.2 | cc:完了 |

---

### Phase 35.2: 統合設定 [P0]

Purpose: `harness.toml` → CC 必須ファイル自動生成で dual hooks.json sync 問題を根本解決

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 35.2.1 | `pkg/config/toml.go` — harness.toml パーサー ([project], [safety], [agent], [env], [hooks], [telemetry] セクション)。Mapping Table は SPEC.md §5 に準拠 | TOML パース + バリデーション + unsupported key rejection | 35.0.3 | cc:完了 |
| 35.2.2 | `harness sync` — harness.toml → hooks.json + settings.json (permissions/sandbox/env/agent) + plugin.json 自動生成 | 生成ファイルと現行ファイルが機能等価 | 35.2.1 | cc:完了 |
| 35.2.3 | `harness init` サブコマンド — プロジェクト初期化 (harness.toml テンプレート生成) | 新規プロジェクトで `harness init && harness sync` が動作 | 35.2.2 | cc:完了 |
| 35.2.4 | dual hooks.json 同期スクリプト (`sync-plugin-cache.sh`) を `harness sync` に統合 | `sync-plugin-cache.sh` が `harness sync` のラッパーになる | 35.2.2 | cc:完了 |

---

### Phase 35.3: ハンドラ統合 [P1]

Purpose: 127 スクリプトをカテゴリ別に Go サブコマンドへ段階的に吸収

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 35.3.1 | hook-handlers 5本を Go に移植 (session-env, post-tool-failure, post-compact, notification, permission-denied) | 25テスト PASS、symlink チェック付き | 35.2.2 | cc:完了 |
| 35.3.2 | session-* 4本を Go に移植 (init, cleanup, monitor, summary) | 30テスト PASS | 35.3.1 | cc:完了 |
| 35.3.3 | codex-companion.sh は Go 統合 **対象外** (SPEC.md 決定事項)。shell wrapper を維持 | 変更なし (SPEC.md に方針文書化済み) | - | cc:完了 |
| 35.3.4 | ci-status-checker + evidence collector を Go に移植 | 15テスト PASS | 35.3.2 | cc:完了 |
| 35.3.5 | `harness doctor` + `--migration` で hook 移行状況を一覧表示。mixed-mode 警告、hooks.json divergence 検出 | `harness doctor` 11テスト PASS | 35.3.1 | cc:完了 |

---

### Phase 35.4: エージェントライフサイクル [P1]

Purpose: SPAWNING→RUNNING→REVIEWING→APPROVED→COMMITTED 状態マシン + 4段階リカバリ

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 35.4.1 | `internal/lifecycle/state.go` — 状態マシン定義 + 遷移ルール。SPEC.md §8 のフル状態 (FAILED/CANCELLED/STALE/RECOVERING/ABORTED 含む) | 不正遷移を型レベルで防止、異常系状態が全て定義済み | 35.3.2 | cc:完了 |
| 35.4.2 | `internal/lifecycle/recovery.go` — 4段階リカバリ (自己修復→仲間修復→指揮官介入→停止) | 各段階のトリガー条件と動作が定義済み | 35.4.1 | cc:完了 |
| 35.4.3 | SubagentStart/Stop フックとの統合 | 状態遷移が SQLite に永続化され、`harness status` で表示 | 35.4.2, 35.1.2 | cc:完了 |

---

### Phase 35.5: スキル検証 [P1]

Purpose: SKILL.md frontmatter の型安全バリデーション

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 35.5.1 | `harness validate skills` — 全 SKILL.md の frontmatter を公式スキーマに照合。正規表現ベース (外部 YAML 依存なし) | name, description 必須フィールド + オプション型検証 | 35.2.1 | cc:完了 |
| 35.5.2 | `harness validate agents` — agents/*.md の frontmatter 検証 (tools, disallowedTools, isolation, background, maxTurns) | 22テスト PASS | 35.5.1 | cc:完了 |

---

### Phase 35.6: Breezing 並行処理 [P2]

Purpose: goroutine + worktree による安全な並列タスク実行

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 35.6.1 | `internal/breezing/orchestrator.go` — Worker/Reviewer の goroutine 管理 | 最大並列数制御、graceful shutdown | 35.4.3 | cc:完了 |
| 35.6.2 | worktree 自動作成/クリーンアップの Go 実装 | CC の WorktreeCreate/Remove フックと連携 | 35.6.1 | cc:完了 |
| 35.6.3 | タスク依存関係の自動解決 + file-lock claiming | 依存タスクの自動 unblock が動作 | 35.6.2 | cc:完了 |

---

### Phase 35.7: npm 配布 + クロスコンパイル [P2]

Purpose: `bin/` ディレクトリ活用による Go バイナリ配布

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 35.7.1 | クロスコンパイル (darwin-arm64/amd64, linux-amd64)。CGO_ENABLED=0 + modernc.org/sqlite | 3バイナリ全て6.6-6.8MB | 35.3.1 | cc:完了 |
| 35.7.2 | npm パッケージ設定 + postinstall でプラットフォーム別バイナリ配置 | `npm install` で `bin/harness` が PATH に配置 | 35.7.1 | cc:未着手 |
| 35.7.3 | 旧パッケージへの移行通知 + GitHub Release 自動化 | リリースワークフローで Go バイナリが含まれる | 35.7.2 | cc:未着手 |

---

## Phase 34: Feature Table 整合性回復 + 未活用 upstream 機能の実装

作成日: 2026-04-02
目的: Feature Table の「書いてあるが活かしていない」ギャップを全て解消し、Claude / Codex 両文脈で Harness の信頼性と活用度を引き上げる

### 設計方針

- P0 は「嘘を直す」。Feature Table の記載を実態に合わせる
- P1 は最小工数で最大効果。PostCompact WIP 復元と HTTP hook 実用例
- P2 は Codex parity とセキュリティ。Codex Worker の effort 伝播と security review profile
- P3 は中期的な基盤強化。OTel 形式変換と dual review

### Phase 34.0: Feature Table 誇張修正 [P0]

Purpose: Feature Table の記載と実態の乖離を解消し、Harness の信頼性を回復する

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 34.0.1 | Feature Table（CLAUDE.md + docs/CLAUDE-feature-table.md）で実態と乖離する 7 件の記載を修正: HTTP hooks→テンプレートのみ、OTel→独自 JSONL、Analytics Dashboard→計画中、LSP→CC native、Auto Mode→RP Phase 1、Slack→将来対応、Desktop Scheduled Tasks→CC native | 7 件の記載が実態と一致し、「実装済み」と誤読される表現がない | - | cc:完了 |

### Phase 34.1: 即効性の高い実装 [P1]

Purpose: 最小工数で Feature Table の整合性を取り、長時間セッションの品質を向上させる

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 34.1.1 | PostCompact hook で PreCompact が保存した WIP タスク情報を systemMessage として復元する | post-compact.sh が WIP 情報を読み取り再注入し、圧縮後もタスク状態が保持される | 34.0.1 | cc:完了 |
| 34.1.2 | hooks.json に TaskCompleted 用の HTTP hook を追加し、`HARNESS_WEBHOOK_URL` 設定時のみ通知が飛ぶ opt-in 外部通知を実装 | hooks.json に `type: "http"` が 1 件以上存在し、URL 未設定時はノンブロッキングスキップ | 34.0.1 | cc:完了 |

### Phase 34.2: Codex parity + セキュリティ [P2]

Purpose: Codex Worker の実装品質とセキュリティレビューの独立性を確保する

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 34.2.1 | harness-review に `--security` フラグを追加し、OWASP Top 10 + 認証/認可 + データ露出に特化した reviewer profile を実装 | `/harness-review --security` で security profile が起動し、security-specific な checks が review-result に含まれる | 34.1.1 | cc:完了 |
| 34.2.2 | codex-companion.sh の task 呼び出し時に Plans.md のタスク情報から effort level を計算して渡す仕組みを追加 | Codex Worker に effort が伝播し、複雑なタスクで高 effort が適用される | - | cc:完了 |

### Phase 34.3: 監視基盤 + Codex Review 統合 [P3]

Purpose: 運用監視と複数モデル視点によるレビュー品質向上

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 34.3.1 | emit-agent-trace.js の出力を OTel Span JSON 形式に寄せ、`OTEL_EXPORTER_OTLP_ENDPOINT` 設定時のみ OTLP HTTP 送信 | OTel endpoint 設定時にスパンが送信され、未設定時は既存 JSONL にフォールバック | 34.0.1 | cc:完了 |
| 34.3.2 | harness-review に `--dual` フラグを追加し、Claude Reviewer と Codex Reviewer を並行実行して verdict をマージ | `/harness-review --dual` で両方の verdict が出て、最終判定が統合される | 34.2.1 | cc:完了 |

---

## Phase 32: Long-running harness hardening from Anthropic article

作成日: 2026-03-30
目的: Anthropic の long-running apps 設計知見を Claude Harness に取り込み、自己評価バイアス、context anxiety、Sprint Contract 不在、静的レビュー偏重、Codex 側 continuity 未完了をまとめて解消する

### 設計方針

- 「レビューする」ではなく「独立 Evaluator が実行可能な基準で判定する」方向へ寄せる
- Compaction 依存を減らし、構造化 handoff artifact + 戦略的 reset を正規経路にする
- Plans.md の DoD を出発点にしつつ、実装前に Sprint Contract へ昇格させる
- Claude / Codex で意味が分岐しないよう、contract / handoff / telemetry は共通 artifact で持つ
- 追加コンポーネントには assumption を明示し、将来の削除判断まで含めて設計する

### 優先度マトリクス

| 優先度 | Phase | 内容 | タスク数 | 依存 |
|--------|-------|------|---------|------|
| **Required** | 32.0 | 独立 Reviewer の全モード適用 + Self-review の役割縮小 | 3 | なし |
| **Required** | 32.1 | Sprint Contract + Context Reset/Handoff + Codex continuity | 4 | 32.0 |
| **Recommended** | 32.2 | Runtime/Browser evaluator と calibration loop | 3 | 32.1 |
| **Recommended** | 32.3 | Assumption Registry + prompt language + per-agent telemetry | 3 | 32.1 |

合計: **13 タスク**

---

### Phase 32.0: 独立 Reviewer の全モード適用 [P0]

Purpose: 記事の核心である「自己評価バイアス回避」を、Breezing だけでなく Solo / Sequential / Codex 経路まで標準化する

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 32.0.1 | `harness-work` の Solo / Sequential / Codex 実行フローを見直し、Worker の self-review を「実装前後の preflight」に格下げし、最終 verdict は常に独立 Reviewer または read-only review runner が返すようにする | 全モードで `cc:完了` 前に独立 verdict が必須になり、Worker 単独で完了確定しない | - | cc:完了 |
| 32.0.2 | Reviewer の出力契約を `review-result.json` 相当の共通 artifact に統一し、`verdict`, `checks`, `gaps`, `followups` を機械可読で残す | Claude / Codex / Breezing の review artifact 形式が統一され、差分比較や再評価に使える | 32.0.1 | cc:完了 |
| 32.0.3 | `README`, `team-composition`, `harness-work`, `harness-review`, evidence 文書を更新し、「self-review は補助」「独立 review が完了条件」という新しい契約にそろえる | 説明・skill・docs・evidence で review 責務の表現が一致し、旧説明が残っていない | 32.0.1, 32.0.2 | cc:完了 |

### Phase 32.1: Sprint Contract と Context Reset/Handoff [P0]

Purpose: 実装前の成功基準合意と、長時間実行での state 継承を first-class な artifact にする

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 32.1.1 | `Plans.md` の DoD/Depends から `sprint-contract.json` を生成し、Worker 着手前に Reviewer が `checks`, `non_goals`, `runtime_validation`, `risk_flags` を追加して合意するフローを設計・実装する | タスク開始前に contract artifact が生成され、未合意のまま Worker が実装開始しない | 32.0.1 | cc:完了 |
| 32.1.2 | `pre-compact` / `post-compact` / `session-init` / `session-resume` を拡張し、`handoff artifact` に `previous_state`, `next_action`, `open_risks`, `failed_checks`, `decision_log` を保存・再読込する | handoff artifact が安定形式で保存され、resume 時に要約ではなく構造化状態として再利用される | 32.0.2 | cc:完了 |
| 32.1.3 | Claude 側に「戦略的 context reset」ポリシーを追加し、turn 数・compaction 直前・Phase 切替などの条件で reset 候補と handoff 生成を行う | reset 条件、生成物、再開手順が定義され、少なくとも dry-run/fixture で再現できる | 32.1.2 | cc:完了 |
| 32.1.4 | 既存の `31.1.2` を吸収する形で、Codex の `plugin-first workflow` と `resume-aware effort continuity` を `sprint-contract` / `handoff artifact` / `session state` に接続する | Codex 経路でも effort と未完了 contract が resume/fork 後に保持され、`31.1.2` を完了扱いにできる根拠がそろう | 32.1.1, 32.1.2 | cc:完了 |

### Phase 32.2: Runtime/Browser evaluator と calibration loop [P1]

Purpose: 静的コードレビューだけでは見えない UX・実動作・レビュー精度のズレを補足する

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 32.2.1 | Reviewer を profile 化し、`static`, `runtime`, `browser` の 3 種類を導入する。`runtime` は Bash で test/lint/typecheck/API probe を行い、contract にある検証を実行する | タスク種別ごとに reviewer profile が選択され、runtime profile で contract 上の検証コマンドを実行できる | 32.1.1 | cc:完了 |
| 32.2.2 | Web アプリ向けに `browser` evaluator を追加し、既存の browser/Chrome 導線を reviewer フローへ統合する。レイアウト崩れ、主要 UI フロー、スクリーンショット差分を contract ベースで検証する | browser profile が利用可能で、少なくとも 1 つの fixture で UI フロー検証 artifact が残る | 32.2.1 | cc:完了 |
| 32.2.3 | review artifact を蓄積し、`false_positive`, `false_negative`, `missed_bug`, `overstrict_rule` を記録する calibration ループと few-shot 更新フローを作る | Reviewer の判断ログから drift を見つけて基準更新できる手順と保存先が存在する | 32.0.2, 32.2.1 | cc:完了 |
| 32.2.4 | browser reviewer の route policy を再設計し、既定値を `Playwright` 中心に切り替える。`playwright | agent-browser | chrome-devtools` の 3 route を正式サポートし、判定順を `contract 明示指定 > repo に Playwright 基盤あり > AgentBrowser 利用可 > Chrome fallback` に統一する。browser_mode: scripted | `sprint-contract`, browser artifact, docs, skill, fixture test で 3 route を表現でき、環境依存ではなく repo/contract ベースで route が決まる | 32.2.2 | cc:完了 |
| 32.2.5 | browser reviewer に `browser_mode`（`scripted` / `exploratory`）を追加し、`scripted` は Playwright、`exploratory` は AgentBrowser を優先する運用へ整理する。artifact も trace/screenshot 系と snapshot/ui-flow-log 系で役割を分ける。browser_mode: exploratory | contract で `browser_mode` を指定でき、mode ごとに既定 route・必要 artifact・review 手順が切り替わる | 32.2.4 | cc:完了 |

### Phase 32.3: Assumption Registry・Prompt Language・Telemetry [P2]

Purpose: 追加した harness 要素を「なぜ必要か」「いつ外せるか」「いくら重いか」まで追跡できる状態にする

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 32.3.1 | guardrail / skill / agent ごとに `assumption` と `retirement signal` を持つ `Assumption Registry` を追加し、「どのモデル限界に対処しているか」を台帳化する | 少なくとも主要 rule/agent/skill に assumption と削除条件が記録され、モデル更新時の棚卸しに使える | 32.1.1 | cc:完了 |
| 32.3.2 | Worker / Reviewer / Lead の prompt language を再設計し、手順だけでなく品質姿勢を明文化する。あわせて文言差分の A/B 比較手順を作る | initialPrompt / review prompt に品質言語が導入され、少なくとも比較観察ログが残せる | 32.0.1, 32.2.3 | cc:完了 |
| 32.3.3 | per-agent の duration / token / cost / retry count / artifact count を集計し、Solo / Breezing / Codex 各モードの ROI を比較できる telemetry surface を追加する | Worker / Reviewer / Lead 単位の集計が見え、コストと成功率をモード別に比較できる | 32.0.2, 32.1.4 | cc:完了 |
## Phase 33: Claude 2.1.87-2.1.90 / Codex 0.118 upstream update integration

作成日: 2026-04-02
目的: CC 2.1.89 の PermissionDenied hook・defer decision と 2.1.90 の guardrail 修正を Harness に取り込み、auto mode 拒否追跡と Breezing 安全弁を実装。自動継承分は Feature Table に明示分類

### 設計方針

- Claude 側は「PermissionDenied handler による拒否追跡」を即時実装対象とする
- defer decision はドキュメント整備のみ。運用パターン蓄積後に具体的な defer ルールを設計
- CC 2.1.90 の PreToolUse exit 2 修正は既存 guardrail の信頼性向上（CC 自動継承）
- Codex 0.118 は prompt-plus-stdin を比較軸として残す

### Phase 33.0: 即時実装 [P0]

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 33.0.1 | `PermissionDenied` hook handler を作成し、auto mode 拒否を telemetry 記録 + Breezing Lead 通知 | `permission-denied-handler.sh` 存在・実行可能・hooks.json 両方に wiring | - | cc:完了 |
| 33.0.2 | hooks-editing.md に `PermissionDenied`・`defer`・`updatedInput+AskUserQuestion`・hook output >50K・exit 2 修正を追記 | hooks-editing.md にイベント一覧と設計指針が更新されている | 33.0.1 | cc:完了 |
| 33.0.3 | Feature Table（CLAUDE.md + docs/CLAUDE-feature-table.md）に v2.1.84-2.1.90 を A/C 分類付きで追加 | Feature Table に新エントリがあり、B（書いただけ）が 0 件 | 33.0.1 | cc:完了 |
| 33.0.4 | upstream integration test に PermissionDenied wiring 検証を追加 | テストが green | 33.0.1, 33.0.2 | cc:完了 |

### Phase 33.1: 将来拡張 [P1]

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 33.1.1 | `defer` permission decision の具体的なルール設計（本番 DB 書込・外部 API・destructive git 操作等） | defer 条件と resume フローが設計文書として残っている | 33.0.2 | cc:完了 |
| 33.1.2 | ~~Codex 0.118 prompt-plus-stdin~~ | companion script は `exec node` で stdin を透過済み。Harness 側の変更不要 | - | cc:完了（対応不要） |
| 33.1.3 | `PermissionDenied` の蓄積データを分析し、auto mode の permission 設定最適化を提案する導線を追加 | 拒否ログの集計と改善提案のスクリプトまたは skill が存在 | 33.0.1 | cc:完了 |

---

## Phase 31: Claude 2.1.80-2.1.86 / Codex 0.117 upstream update integration

作成日: 2026-03-28
目的: Claude Code と Codex の最新アップデートを「書いただけ」で終わらせず、Claude 側は Harness の既存フック・設定・エージェント・ルール生成に実効改善として取り込み、Codex 側は次に伸ばす価値軸を整理する

### 設計方針

- 公式 changelog / releases を先に確認し、一次情報ベースで候補を絞る
- Claude 側は「Harness で 2 倍以上の価値に変換できるもの」だけを実装対象にする
- Codex 側は比較軸と将来タスクを残し、今回の主実装は Claude 優先で進める
- 非配布の内部スキルとして、次回以降も同じ流れで再実行できる形に保存する

### Phase 31.0: 調査と即時実装 [P0]

Purpose: 最新更新の中から今すぐ効くものを選び、Harness の体験改善までつなげる

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 31.0.1 | Claude changelog (`2.1.80`〜`2.1.86`) と Codex releases (`0.117.0`) を調査し、Harness で意味がある候補を整理する | Claude / Codex の候補が「実装対象」「比較軸」「将来対応」に分かれている | - | cc:完了 |
| 31.0.2 | Claude Code `hooks conditional if field` を `PermissionRequest` に取り込み、Bash の安全コマンドだけに permission hook を起動するようにする | `.claude-plugin/hooks.json` と `hooks/hooks.json` で Bash `PermissionRequest` に `if` が入り、`claude plugin validate` を通る | 31.0.1 | cc:完了 |
| 31.0.3 | `PermissionRequest` の編集系 matcher を `Edit|Write|MultiEdit` にそろえ、hooks と core の自動承認面を一致させる | `MultiEdit` が hooks 側でも取りこぼしなく permission flow に乗る | 31.0.2 | cc:完了 |
| 31.0.4 | `sandbox.failIfUnavailable` と `CLAUDE_CODE_SUBPROCESS_ENV_SCRUB=1` を `.claude-plugin/settings.json` に取り込み、sandbox 失敗時の unsandboxed 継続と subprocess への認証情報引き継ぎを抑える | settings に両項目があり、validate / integration test で確認できる | 31.0.1 | cc:完了 |
| 31.0.5 | `TaskCreated` / `CwdChanged` / `FileChanged` hook と `runtime-reactive.sh` を追加し、バックグラウンド task・Plans 更新・worktree 切替を記録できるようにする | hooks wiring と handler が存在し、runtime reactive test が通る | 31.0.1 | cc:完了 |
| 31.0.6 | rules template と `scripts/localize-rules.sh` の `paths:` を YAML list 形式へ移行し、複数 glob を壊れにくくする | template / generated rule の paths が YAML list になり、validate で壊れていない | 31.0.1 | cc:完了 |
| 31.0.7 | skill `effort` frontmatter と agent `initialPrompt` を `skills-v3` / `agents-v3` / mirror に追加し、重いフローの初動品質を安定化する | skill / agent frontmatter が更新され、integration test と validate が通る | 31.0.1 | cc:完了 |
| 31.0.8 | Feature Table / CHANGELOG / upstream integration test を更新し、「追従した」ではなく「Harness でどう強くなったか」を記録する | docs / changelog / tests で 2.1.80〜2.1.86 反映が確認できる | 31.0.2, 31.0.3, 31.0.4, 31.0.5, 31.0.6, 31.0.7 | cc:完了 |

### Phase 31.1: 将来拡張の保存 [P1]

Purpose: 今回実装しないが価値の高い更新を、次回すぐ取り込める形で残す

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 31.1.1 | Claude Code `PreToolUse updatedInput` を使った AskUserQuestion 自動補完・入力正規化の実装方針を内部スキルへ残す | 次回着手時に対象 surface と実装案が追える | 31.0.1 | cc:完了 |
| 31.1.2 | Codex の `plugin-first workflow` と `resume-aware effort continuity` を比較軸として Plans に残す | Claude / Codex の差分を埋める次フェーズ候補が明文化されている | 31.0.1 | cc:完了 |
| 31.1.3 | この一連の調査・実装フローを非配布の内部スキルとして保存する | `skills/claude-codex-upstream-update/SKILL.md` が存在し、ローカル専用運用が書かれている | 31.0.1 | cc:完了 |

## Phase 30: Claude Code / Codex 両対応 hardening parity

作成日: 2026-03-25
目的: `claude-code-hardened` から得た示唆を Harness に取り込み、Claude Code は hook による runtime enforcement、Codex は wrapper / quality gate / merge gate による近似 enforcement として両経路に反映する

### 設計方針

- shell hook をそのまま移植せず、Harness の既存 surface（`core/guardrails`, `hooks`, `scripts/codex*`, quality gate）へ吸収する
- 共通化するのは「ポリシー」であり、実装は Claude Code と Codex で分ける
- Claude Code は deny / warn / ask を PreToolUse / PostToolUse で実行し、Codex は実行前注入・実行後検査・マージ前検証で同等の事故防止を目指す
- プラットフォーム制約により完全一致しない差分は docs に明示し、`validate` / `doctor` 系の出力で利用者に見える化する

### Phase 30.0: 共通 hardening policy 定義 [P0]

Purpose: 先に「何を守るか」を定義し、Claude Code / Codex で別実装しても意味がズレないようにする

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 30.0.1 | `docs/` または `skills-v3/harness-work` 参照文書に hardening policy matrix を追加し、対象ルール（`--no-verify` / `--no-gpg-sign` 禁止、protected branch への `reset --hard` 禁止、protected files 警告、pre-push secrets チェック）の意図と severity を定義 | 対象ルール、適用 surface、deny/warn/ask の基準が表で確認できる | - | cc:完了 |
| 30.0.2 | Claude Code / Codex の適用方式マッピングを定義し、「共通ポリシー・実装差分・既知の非対称性」を明文化する | 各ルールごとに CC hook / Codex wrapper / quality gate / docs-only のどれで実現するかが決まっている | 30.0.1 | cc:完了 |

### Phase 30.1: Claude Code 経路の runtime hardening 追加 [P0]

Purpose: Claude Code では hook で直接止められるものを増やし、Git 事故と重要ファイル誤編集を実行前に減らす

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 30.1.1 | `core/src/guardrails/rules.ts` に `--no-verify` / `--no-gpg-sign` 禁止ルールと protected branch への `git reset --hard` deny / direct push warn ルールを追加 | 危険 Git コマンドが期待どおり deny / warn され、既存 force-push ルールと競合しない | 30.0.2 | cc:完了 |
| 30.1.2 | protected files プロファイル（例: `package.json`, `Dockerfile`, `.github/workflows`, `schema.prisma`）を warn / deny できる設定面を追加する | 既定または opt-in の protected files 一覧があり、Write/Edit 時に警告またはブロックできる | 30.0.2 | cc:完了 |
| 30.1.3 | `core/src/guardrails/__tests__/rules.test.ts` と統合テストを更新し、上記 hardening の回帰テストを追加する | deny / warn の主要ケースにテストがあり、既存 guardrail テストが全通過する | 30.1.1, 30.1.2 | cc:完了 |

### Phase 30.2: Codex 経路の parity hardening 追加 [P0]

Purpose: Codex Hooks 未搭載の制約下で、wrapper / prompt 注入 / quality gate / merge gate を組み合わせて近い安全性を出す

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 30.2.1 | `scripts/codex/codex-exec-wrapper.sh` と関連 rules 注入フローを拡張し、Codex 実行前に hardening contract を base instructions へ注入する | Codex 実行前に hardening policy が必ず prompt context に含まれ、スキップ手段が明示的に制限される | 30.0.2 | cc:完了 |
| 30.2.2 | `scripts/codex-worker-quality-gate.sh` などの post-exec 検証に protected files / no-verify 相当 / secrets 検査を追加し、危険変更を merge 前に落とす | Codex の出力結果に対して hardening policy 違反が検出され、失敗理由が安定出力される | 30.2.1 | cc:完了 |
| 30.2.3 | Codex 経路の既知限界（hook 不在による runtime 非対称）を docs に記載し、Claude Code との差を利用者に見える化する | docs に「どこまで同等化できたか / できないか」が明記されている | 30.2.2 | cc:完了 |

### Phase 30.3: validate / doctor / docs の見える化 [P1]

Purpose: hardening の実装有無と適用範囲を、人が読んで判断できる状態にする

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 30.3.1 | `tests/validate-plugin.sh` または新規スクリプトに hardening surface の検査を追加し、Claude Code / Codex 両経路の有効化状態を確認できるようにする | validate 実行で主要 hardening policy の有効/未設定が判別できる | 30.1.3, 30.2.2 | cc:完了 |
| 30.3.2 | README または docs に「Claude Code では直接 enforcement、Codex では近似 enforcement」という説明と運用例を追加する | 利用者が両経路の差と推奨運用を 1 ページで理解できる | 30.3.1 | cc:完了 |
| 30.3.3 | CHANGELOG の `[Unreleased]` に cross-runtime hardening 追加を記録する | 利用者価値と制約が CHANGELOG に簡潔に記載されている | 30.3.2 | cc:完了 |

---

## Phase 29: CCAGI レポート由来の高価値要素を Harness に取り込む

作成日: 2026-03-24
目的: CCAGI 調査で見えた「運用上は効くがベンダー依存が強すぎる要素」を分解し、Harness の公開・軽量・汎用方針を崩さずに再実装する

### 設計方針

- 取り込むのは「失敗しやすい場所の型化」であり、Issue-first や AWS/認証基盤固定ではない
- デフォルトは今の軽さを維持し、チーム向け機能は opt-in にする
- LLM 判断だけに寄せず、再実行可能な検証は scripts へ降ろす

### 推奨実行順（2026-04-01 更新）

- まず `29.2.x` を進め、release 前の現実チェックを先に標準化する
- 次に `29.1.x` で Plans.md と GitHub Issue の opt-in bridge を dry-run 前提で追加する
- 続いて `29.3.x` で optional brief と machine-readable manifest を整え、比較・監査・自動 docs の足場を作る
- 最後に `29.4.x` で validate / consistency / CHANGELOG を通し、Phase 29 全体を「計画だけでなく再実行できる状態」で閉じる
- `25.5.3` の GitHub Release 作成は実際に公開を切るタイミングでのみ実施し、通常の実装残タスクとは分けて扱う

### Phase 29.0: AI 残骸レビューゲート [P0]

Purpose: AI 実装で混入しやすい mock, dummy, localhost, TODO 残骸を Harness Review の既定観点に追加し、「動くが出荷できない」状態を減らす

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 29.0.1 | `skills-v3/harness-review/SKILL.md` に 5つ目の観点 `AI Residuals` を追加し、検出対象（`mockData`, `dummy`, `localhost`, `TODO`, テスト無効化, ハードコード設定）の verdict ルールを定義 | SKILL.md に `AI Residuals` 観点と severity 判定表があり、minor/major の境界が明記されている | - | cc:完了 |
| 29.0.2 | 差分または対象ファイルを静的に走査する `scripts/review-ai-residuals.sh` を新設し、安定フォーマットで検出結果を出力 | スクリプトが対象ファイル入力を受け、検出 0 件でもエラーなく安定出力する | 29.0.1 | cc:完了 |
| 29.0.3 | `harness-review code` から上記スクリプトを呼ぶ手順と、最小回帰テスト/fixture を追加 | review フローとテスト/fixture が追加され、残骸検出がレビュー結果へ反映される | 29.0.2 | cc:完了 |

### Phase 29.1: Plans.md ⇄ GitHub Issue ブリッジ（opt-in） [P1]

Purpose: Plans.md を SSOT のまま維持しつつ、チーム利用時だけ Issue と橋渡しできるようにする。CCAGI の Issue-first を強制せず、必要時だけ使える形にする

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 29.1.1 | `harness-plan` に opt-in の team mode 仕様を追加し、`Plans.md -> tracking issue / sub-issue` の変換ルールを定義 | SKILL.md または reference に opt-in 条件、tracking issue 形式、非採用時の既定動作が明記されている | - | cc:完了 |
| 29.1.2 | `scripts/plans-issue-bridge.sh` を新設し、Plans.md から issue payload の dry-run 出力（JSON または Markdown）を生成 | スクリプトが Plans.md から task, DoD, Depends, Status を抽出し、stable な dry-run 出力を返す | 29.1.1 | cc:完了 |
| 29.1.3 | `harness-plan` / docs に「ソロ開発では不要、チーム開発では推奨」の使い分けと運用例を追加 | README か docs に使い分け例があり、デフォルトフローが重くなっていない | 29.1.2 | cc:完了 |

### Phase 29.2: ベンダー非依存の pre-release verification [P1]

Purpose: CCAGI の deploy 前チェック思想だけを汎用化し、AWS 固定なしで「出荷前に見るべき現実チェック」を Harness Release に追加する

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 29.2.1 | `skills-v3/harness-release/SKILL.md` に vendor-neutral な pre-release チェック項目（未コミット差分, env/healthcheck, debug/mock 残骸, CHANGELOG, CI 状態）を追加 | release フローに pre-release verification セクションがあり、チェック失敗時の中断条件が明記されている | - | cc:完了 |
| 29.2.2 | `scripts/release-preflight.sh` を新設し、上記チェックを安定出力で実行する | スクリプトが主要チェックの pass/fail を一覧出力し、失敗時に非0終了する | 29.2.1 | cc:完了 |
| 29.2.3 | release skill / docs / tests に preflight 実行導線を追加し、dry-run でも確認できるようにする | `/harness-release --dry-run` で preflight が案内され、関連テストまたは fixture が追加されている | 29.2.2 | cc:完了 |

### Phase 29.3: 軽量 brief と machine-readable manifest [P2]

Purpose: CCAGI の document-first の重さは持ち込まず、必要な場面だけ設計の足場を出す。また Harness の skill/command を比較・監査しやすくする

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 29.3.1 | `harness-plan create` に optional brief 生成ルールを追加し、UI タスクでは `design brief`、API タスクでは `contract brief` を出せるようにする | create フローに optional brief 条件があり、UI/API の brief テンプレが存在する | - | cc:完了 |
| 29.3.2 | skill frontmatter / routing rules / mirror 情報から `machine-readable manifest` を生成するスクリプトまたは doc 生成フローを追加 | skill 名、用途、禁止用途、関連 surface を含む manifest が生成できる | - | cc:完了 |
| 29.3.3 | 上記 2 つの成果物を README または docs で説明し、比較・監査・自動 docs への用途を明記 | docs に brief/manifest の用途と生成方法が記載されている | 29.3.1, 29.3.2 | cc:完了 |

### Phase 29.4: 統合検証・CHANGELOG [P2]

Purpose: Phase 29 の追加を既存の plugin/skill/docs 検証ループに乗せ、単なる計画で終わらせない

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 29.4.1 | `validate-plugin.sh`, `check-consistency.sh`, 必要な skill mirror/check を通し、Phase 29 追加後の回帰がないことを確認 | 主要検証が全通過し、追加スクリプトが CI 想定で再実行可能 | 29.0〜29.3 | cc:完了 |
| 29.4.2 | CHANGELOG の `[Unreleased]` に「CCAGI 調査から取り込んだ価値」を Before/After 形式で記録 | CHANGELOG に Phase 29 の要点と利用者価値が記載されている | 29.4.1 | cc:完了 |

---

## Maintenance: Claude Code v2.1.77〜v2.1.79 統合

作成日: 2026-03-20
目的: CC v2.1.77〜v2.1.79 の新機能・修正を Harness に統合し、Feature Table・hooks・guardrail docs を最新化する

### Phase M-CC79.0: ドキュメント・フック基盤統合 [P0]

Purpose: Feature Table 更新と StopFailure フック新設で基盤を整える

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| CC79.0.1 | CLAUDE.md Feature Table に v2.1.77〜v2.1.79 の全新機能（21項目）を追加、バージョン表記を 2.1.79+ に更新 | Feature Table に 21 行追加、表記が 2.1.79+ | - | cc:完了 |
| CC79.0.2 | docs/CLAUDE-feature-table.md に詳細セクション追加（各機能の動作概要・Harness 活用方法） | 詳細セクションが存在し、既存フォーマットと一致 | CC79.0.1 | cc:完了 |
| CC79.0.3 | hooks.json (×2) に `StopFailure` イベント定義を追加 + `stop-failure.sh` ハンドラーを新設 | `StopFailure` が hooks.json に存在、ハンドラーが実行可能 | - | cc:完了 |
| CC79.0.4 | hooks-editing.md のイベント型一覧・タイムアウト表・バージョン注記を更新 | `StopFailure`, `ConfigChange` が一覧に存在、v2.1.77/78 注記あり | CC79.0.3 | cc:完了 |
| CC79.0.5 | core/src/types.ts の `SignalType` に `stop_failure` を追加 | 型定義が存在 | CC79.0.3 | cc:完了 |
| CC79.0.6 | session-control スキルの description を `/fork` → `/branch` に更新 | description に `/branch` が記載 | - | cc:完了 |
| CC79.0.7 | CHANGELOG.md [Unreleased] に統合変更を記録 | 変更点が Before/After 形式で記載 | CC79.0.1〜CC79.0.6 | cc:完了 |

### Phase M-CC79.1: settings.json deny パターン移行 [P1]

Purpose: フックベースの MCP ブロックから settings.json deny に移行し、v2.1.77 の allow/deny 優先順位を活用

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| CC79.1.1 | `.claude/settings.json` に `deny: ["mcp__codex__*"]` のコメント付きテンプレートを追加 | settings.json に deny テンプレートが存在 | - | cc:完了 |
| CC79.1.2 | `codex-cli-only.md` に v2.1.78 の settings.json deny パターンを推奨として追記 | ルールファイルに deny パターンの説明がある | CC79.1.1 | cc:完了 |

### Phase M-CC79.2: プラグイン永続ステート移行準備 [P1]

Purpose: `${CLAUDE_PLUGIN_DATA}` 変数を活用し、プラグイン更新でのステート消失を防止

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| CC79.2.1 | hook handler のステート保存パスを `${CLAUDE_PLUGIN_DATA}` 対応に段階移行（フォールバック付き） | `CLAUDE_PLUGIN_DATA` 設定時はそちらに保存、未設定時は旧パス | - | cc:完了 |
| CC79.2.2 | harness-setup スキルに `${CLAUDE_PLUGIN_DATA}` と `ANTHROPIC_CUSTOM_MODEL_OPTION` の説明を追記 | SKILL.md に両変数の説明がある | - | cc:完了 |

### Phase M-CC79.3: CI 検証強化 + Agent effort 宣言 [P2]

Purpose: `claude plugin validate` の CI 統合と Agent frontmatter の effort フィールド活用

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| CC79.3.1 | `tests/validate-plugin.sh` に `claude plugin validate` を追加（v2.1.77+ 必要） | CI で frontmatter + hooks.json の構文検証が実行される | - | cc:完了 |
| CC79.3.2 | Worker/Reviewer エージェント定義に `effort` フィールドを検討・追加（Worker: medium, Reviewer: medium, セキュリティレビュー時は high） | agent frontmatter に effort が記載 | - | cc:完了 |

---

## Phase 28: CC アプデ追従の品質革命 — 「書いただけ禁止」+ 付加価値実装

作成日: 2026-03-20
起点: CC v2.1.77〜v2.1.79 統合のセルフレビューで「21項目中、Harness ならではの付加価値は3件のみ」と判明
目的: (1) 今後の CC アプデ追従で「書いただけ」を構造的に防止するガードレールスキルを作る (2) 既存の「書いただけ」項目に本当の付加価値を実装する

### 背景

- 3エージェント並列レビュー（悪魔の代弁者 / プロダクト価値アーキテクト / UX アナリスト）の結論が一致
- Feature Table 21項目のうち14項目が「CC の恩恵を記載しただけ」
- ペルソナ別の改善実感: ソロ開発者 4/10、Breezing ユーザー 7/10、VibeCoder 1/10
- Harness の本当の価値は「セッション間・プロジェクト間の統治」にある
- CC が 1 セッション内の自動化を極めるほど、Harness は「メタレイヤー」に徹すべき

### 設計原則

1. **「CC 機能の転記」は付加価値ではない** — Feature Table に載せるなら「Harness がどう活用するか」の実装が必須
2. **自動で体験が変わること** — ユーザーが Feature Table を読まなくても恩恵を受ける設計
3. **CC にできないことだけ実装する** — 1 セッション完結の機能は CC に委譲。Harness は複数セッション・複数タスクの統治

### 優先度マトリクス

| 優先度 | Phase | 内容 | タスク数 | 依存 |
|--------|-------|------|---------|------|
| **Required** | 28.0 | 「書いただけ禁止」ガードレールスキル | 3 | なし |
| **Required** | 28.1 | StopFailure → 自動復旧（Breezing 信頼性の根本改善） | 3 | なし |
| **Required** | 28.2 | Effort 動的注入（既存スコアリングとの接続） | 2 | なし |
| **Recommended** | 28.3 | StopFailure ログ可視化 + allowRead sandbox 自動設定 | 3 | 28.1 |
| **Required** | 28.4 | 統合検証・CHANGELOG | 2 | 28.0〜28.3 |

合計: **13 タスク**

---

### Phase 28.0: 「書いただけ禁止」ガードレールスキル [P0]

Purpose: 今後の CC アプデ追従で「Feature Table に書いただけ」を構造的に防止する。非配布の内部専用スキル。

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 28.0.1 | `skills/cc-update-review/` を新設（`user-invocable: false`）。CC アプデ統合 PR を対象に「各 Feature Table 項目に対応する実装変更があるか」を検証するチェックリストスキルを作成 | スキルが存在し、frontmatter に `user-invocable: false` が設定されている | - | cc:完了 |
| 28.0.2 | スキル内に 3 分類の判定基準を定義: (A) 実装あり = hooks/scripts/agents/skills に変更がある (B) 書いただけ = Feature Table のみ変更 (C) CC 自動継承 = Harness 側の変更不要（パフォーマンス改善・バグ修正等）。B は「付加価値の実装案」を必須出力にする | 判定基準が SKILL.md に記載され、B 判定時に実装案が出力される | 28.0.1 | cc:完了 |
| 28.0.3 | `.claude/rules/cc-update-policy.md` を新設。「Feature Table への追加は、対応する実装変更またはカテゴリ C（CC 自動継承）の明示的な分類を伴うこと」をルール化 | ルールファイルが存在し、CLAUDE.md からリンクされている | 28.0.2 | cc:完了 |

### Phase 28.1: StopFailure → 自動復旧 [P0]

Purpose: Breezing で Worker がレート制限で死んだ時、Lead が自動検出・バックオフ・再開する。CC 単体にはない「チーム統治」の付加価値。

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 28.1.1 | `breezing/SKILL.md` の Lead Phase B に StopFailure 検出ロジックを追加。`.claude/state/stop-failures.jsonl` を定期 scan し、429 エラーの Worker を特定 | Lead が StopFailure ログから失敗 Worker を特定する手順が SKILL.md に記載 | - | cc:完了 |
| 28.1.2 | `breezing/SKILL.md` にエラーコード別の自動アクションを定義: 429 → 指数バックオフ（30s/60s/120s）後に `SendMessage` で Worker に再開指示、401 → Lead が systemMessage でユーザーに通知、500 → Plans.md にブロッカー記録 | エラーコード別のアクション表が SKILL.md に存在 | 28.1.1 | cc:完了 |
| 28.1.3 | `scripts/hook-handlers/stop-failure.sh` に `systemMessage` 出力を追加。429 検出時に Lead へ「Worker X がレート制限で停止。30 秒後に自動再開します」を通知 | stop-failure.sh が 429 時に systemMessage JSON を出力する | 28.1.1 | cc:完了 |

### Phase 28.2: Effort 動的注入 [P0]

Purpose: harness-work の既存スコアリング（閾値 ≥ 3 で ultrathink）と Agent frontmatter の `effort` フィールドを接続する。

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 28.2.1 | `harness-work/SKILL.md` のスコアリングセクションを拡張。スコア ≥ 3 の場合、spawn prompt への `ultrathink` 注入に加えて Agent tool の `model` パラメータ経由での effort 指定を明記（注: Agent frontmatter の `effort: medium` はデフォルト値、spawn 時の指定が上書き） | スコアリング → effort 注入のフローが SKILL.md に明記 | - | cc:完了 |
| 28.2.2 | `agents-v3/worker.md` の Effort 制御セクションに「Lead からの動的 effort 上書き」の説明を追加。完了後に「effort: high で足りたか」を agent memory に記録する指示を追記 | worker.md に動的 effort の受け取りと事後記録の手順が記載 | 28.2.1 | cc:完了 |

### Phase 28.3: ログ可視化 + Sandbox 自動設定 [P1]

Purpose: 「記録するだけ」から「見える・使える」へ。CC 単体にはないプロジェクト横串の可視化。

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 28.3.1 | `harness-sync/SKILL.md` に `--show-failures` サブコマンドを追加。`.claude/state/stop-failures.jsonl` を集計し、エラーコード別・時間帯別のサマリーを表示 | `/harness-sync --show-failures` で直近のエラーサマリーが表示される | 28.1 | cc:完了 |
| 28.3.2 | `.claude-plugin/settings.json` に `allowRead` sandbox テンプレートを追加。Reviewer が `.env.example`、`config/public-*`、`docs/` を読めるが `.env`、秘密鍵は読めない設定 | settings.json に sandbox.allowRead が存在し、Reviewer のセキュリティレビュー精度が向上する設計 | - | cc:完了 |
| 28.3.3 | `harness-setup/SKILL.md` の `init` サブコマンドに sandbox 自動設定ステップを追加。プロジェクト種別に応じて allowRead/denyRead を自動生成 | `harness-setup init` で sandbox 設定が自動生成される手順が記載 | 28.3.2 | cc:完了 |

### Phase 28.4: 統合検証・CHANGELOG [P2]

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 28.4.1 | `validate-plugin.sh` + `check-consistency.sh` 全体検証 | 全検証パス | 28.0〜28.3 | cc:完了 |
| 28.4.2 | CHANGELOG.md [Unreleased] に Phase 28 の変更を記録 | 変更点が Before/After 形式で記載 | 28.4.1 | cc:完了 |

### Phase 28.5: ランタイム確実性の強化 [P0]

Purpose: SKILL.md の指示（LLM 判断依存）ではなく、hooks/scripts で確定的に動く仕組みに昇格すべきものだけを実装する

**スクリプト化する基準**:
- hooks で自動発火し、LLM の判断なしに確定出力するもの → スクリプト化
- Lead の文脈判断が必要なもの（バックオフ待機時間、effort の妥当性判断）→ SKILL.md のまま

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 28.5.1 | `scripts/show-failures.sh` を新設。`stop-failures.jsonl` を読み込み、エラーコード別集計・直近5件・推奨アクションを stdout に出力するスタンドアロンスクリプト | `bash scripts/show-failures.sh` で集計サマリーが表示される。JSONL が空でもエラーなし | - | cc:完了 |
| 28.5.2 | `harness-sync/SKILL.md` の `--show-failures` セクションを更新。LLM が手動集計するのではなく `scripts/show-failures.sh` を Bash 実行する手順に変更 | SKILL.md が `Bash("scripts/show-failures.sh")` を指示している | 28.5.1 | cc:完了 |
| 28.5.3 | `validate-plugin.sh` + `check-consistency.sh` で回帰確認 | 全検証パス | 28.5.1〜28.5.2 | cc:完了 |

**スクリプト化しないもの（理由付き）**:
- Lead のバックオフ+再開 → 待機時間は Worker の状況次第で変わる。固定スクリプトより Lead の判断が適切
- Effort 動的注入 → スコアリングは spawn prompt のコンテキスト（タスク内容、影響範囲）に依存。hooks では入力情報が足りない
- Sandbox 自動設定 → `settings.json` に既にテンプレート適用済み。init 時の自動生成は `harness-setup` スキルの責務

---

## Fix: プラグイン利用者向け品質改善（Issue #64, #65）

作成日: 2026-03-19
目的: プラグインインストール後に利用者が遭遇する致命的エラー・UX 問題を修正する（Issue #64: MODULE_NOT_FOUND, Issue #65: HTTP hook エラー）

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| F1 | `.gitignore` から `/core/dist/` を除外解除し、ビルド済み JS をリポジトリに含める | `core/dist/index.js` が git tracked になり、`claude plugin install` 後にフックが動作する | - | cc:完了 |
| F2 | `hooks.json` (×2) から `localhost:9090` HTTP hook エントリを削除し、`docs/examples/` にテンプレートとして移動 | デフォルト状態で HTTP hook エラーが出ない。テンプレートがドキュメントで参照可能 | - | cc:完了 |
| F3 | 壊れたシンボリックリンク `skills-v3/extensions/codex-review` を削除 | `find -type l -xtype l` で broken symlink が 0 件 | - | cc:完了 |
| F4 | `marketplace.json` のライセンスを `plugin.json` と統一（MIT） | 両ファイルの license フィールドが一致 | - | cc:完了 |
| F5 | CHANGELOG.md の `[Unreleased]` にプラグイン品質改善の変更点を記録 | CHANGELOG に全変更が Before/After 形式で記載 | F1-F4 | cc:完了 |

---

## Maintenance: v3.10.3 release closeout

作成日: 2026-03-14
目的: 未公開の M10〜M18 をまとめて patch release として確定し、version / tag / GitHub Release / main push まで完了する

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| M19 | `3.10.3` として release metadata を更新し、検証・tag・push・GitHub Release を完了する | `VERSION` / `plugin.json` / `CHANGELOG` / tag / GitHub Release / `origin/main` が `3.10.3` で一致し、主要検証が通る | M10-M18 | cc:完了 |

---

## Maintenance: Claude Code 2.1.76 統合

作成日: 2026-03-14
目的: CC 2.1.76 の新機能（MCP Elicitation, PostCompact hook, -n/--name, worktree.sparsePaths 等）を Harness に組み込み、Feature Table・hooks・skills を最新化する

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| M13 | hooks.json (×2) に Elicitation/ElicitationResult/PostCompact フックエントリ追加 + handler scripts 新規作成 | 新フック 3 種が hooks.json に追加、handler が `${CLAUDE_PLUGIN_ROOT}` 経由で実行可能 | - | cc:完了 |
| M14 | CLAUDE.md Feature Table に CC 2.1.76 の全新機能行を追加（~10 行）、バージョン表記を 2.1.76+ に更新 | Feature Table に新機能行があり、表記が 2.1.76+ | - | cc:完了 |
| M15 | docs/CLAUDE-feature-table.md に CC 2.1.76 の詳細セクション追加 | 各新機能の動作概要・Harness 活用方法・制約事項が記載 | M14 | cc:完了 |
| M16 | breezing/SKILL.md + harness-work/SKILL.md に `-n`/`--name`、`worktree.sparsePaths`、部分結果保持、`/effort` コマンド参照を追記 | 4 機能が skills に反映 | - | cc:完了 |
| M17 | hooks-editing.md に Elicitation/ElicitationResult/PostCompact イベント追記 + `--plugin-dir` 破壊的変更をドキュメント反映 | hooks-editing.md に 3 イベント、docs に破壊的変更の注記あり | - | cc:完了 |
| M18 | CHANGELOG.md [Unreleased] に CC 2.1.76 統合の変更点を記録 | CHANGELOG に全変更点が記録 | M13-M17 | cc:完了 |

---

## Maintenance: Codex command surface + stale skill cleanup

作成日: 2026-03-13
目的: Codex の native multi-agent / subagent 導線に合わせて Harness の Codex 側コマンドを更新し、昔の skill/command が `~/.codex/skills` に残る問題を解消する

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| M12 | Codex 配布 docs / AGENTS / setup scripts / tests を更新し、`harness-*` コマンド面と legacy skill cleanup を現行 Codex に合わせる | `test-codex-package.sh` と関連検証が通り、Codex で推奨コマンド面と stale skill cleanup が説明できる | M11 | cc:完了 |

---

## Maintenance: PR61 selective merge rescue

作成日: 2026-03-13
目的: PR #61 を release metadata ごと取り込まず、現行 `main` に不足している実質差分だけを救済して merge-ready にする

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| M11 | PR61 の docs 差分を現行 release-only policy に沿って取り込み、不要な version bump / release entry を排除したうえで回帰確認を通す | `check-version-bump.sh` / `check-consistency.sh` / `validate-plugin.sh` / `validate-plugin-v3.sh` / `test-codex-package.sh` が通り、PR61 の rescue 方針を説明できる | M10 | cc:完了 |

---

## Maintenance: release-only versioning workflow

作成日: 2026-03-13
目的: feature PR で version / version badge / versioned CHANGELOG が先行して競合・赤CIを生まないよう、release 時だけ metadata を更新する運用へ切り替える

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| M10 | pre-commit / CI / ドキュメント / release skill を「通常PRでは VERSION を触らず、release 時だけ bump する」方針に統一し、PR61 のような drift を再発防止する | `validate-plugin.sh` / `check-consistency.sh` / `test-codex-package.sh` / 必要な追加回帰テストが通り、運用手順と merge 方針を説明できる | - | cc:完了 |

---

## Maintenance: v3.10.2 release closeout

作成日: 2026-03-12
目的: TaskCompleted finalize hardening と Claude Code 2.1.74 docs 追従を README / CHANGELOG / version metadata まで揃えて正式 release する

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| M9 | `VERSION` / `plugin.json` / README 英日 / CHANGELOG / 互換性 docs を 3.10.2 と最新検証結果に同期し、commit・push・tag・GitHub Release まで完了する | `check-consistency.sh` と関連テストが通り、`v3.10.2` の tag / GitHub Release / main push が確認できる | M8 | cc:完了 |

---

## Maintenance: TaskCompleted finalize hardening

作成日: 2026-03-12
目的: 全タスク完了時に harness-mem finalize を安全に前倒しし、Stop 前クラッシュ時の記録取りこぼしを減らす

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| M8 | `task-completed.sh` に idempotent な finalize 呼び出しを追加し、専用回帰テストで「最後のタスクだけ finalize」「重複 finalize しない」「session_id 未解決時は skip」を検証する | `tests/test-task-completed-finalize.sh` と既存関連テストが通り、TaskCompleted ベース finalize の挙動と安全条件を説明できる | - | cc:完了 |

---

## Maintenance: Auto Mode review follow-up

作成日: 2026-03-12
目的: Auto Mode 既定化まわりの表現と実装実態のズレを是正し、agent skill preload 名と breezing mirror チェックを整えてレビューを通す

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| M6 | Auto Mode を rollout/opt-in 表現へ戻し、agents-v3 の skills 名を実在 `harness-*` に統一し、breezing mirror drift を CI で検知できるようにする | `./scripts/sync-v3-skill-mirrors.sh --check` / `./scripts/ci/check-consistency.sh` / `./tests/validate-plugin.sh` / `./tests/test-codex-package.sh` が通り、follow-up review で重大指摘がなくなる | - | cc:完了 |

---

## Maintenance: PR59/60 Auto Mode default merge prep

作成日: 2026-03-12
目的: PR #59 / #60 を Auto Mode 既定方針で merge できるよう、skill 正本・docs・README 版表記・mirror を同期し、validate の残ブロッカーを解消する

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| M5 | Breezing の Auto Mode 既定化を teammate 実行層に統一し、README / feature docs / CHANGELOG / skill mirror を同期して merge-ready に戻す | `./scripts/sync-v3-skill-mirrors.sh --check` / `./scripts/ci/check-consistency.sh` / `./tests/validate-plugin.sh` が通り、README 英日・CHANGELOG・skills-v3/mirror が一致している | - | cc:完了 [6983808] |

---

## Maintenance: PR58 pre-merge stabilization

作成日: 2026-03-11
目的: PR #58 の docs / CI / mirror 整合を修正し、merge 可否を再判定できる状態に戻す

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| M3 | Auto Mode ドキュメント誤記、README/CHANGELOG の版ズレ、validate-plugin の baseline 破綻、opencode mirror ドリフトを修正する | `validate-plugin.sh` / `check-consistency.sh` / `node scripts/build-opencode.js` / `core` テストが通り、PR #58 の残ブロッカーが整理されている | - | cc:完了 [cb625b12] |

---

## Maintenance: v3.9.0 release redo

作成日: 2026-03-11
目的: 新バージョンを切らずに v3.9.0 を正式 release としてやり直し、README / CHANGELOG / tag / GitHub Release を一致させる

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| M4 | CHANGELOG の未公開版番表記を整理し、v3.9.0 の tag と GitHub Release を作成して release 整合を回復する | README 英日・VERSION・plugin.json・CHANGELOG・tag・GitHub Release が v3.9.0 で一致している | - | cc:完了 [7618428c] |

---

## Maintenance: Claude-mem MCP 削除

作成日: 2026-03-08
目的: Claude-mem を MCP として接続する経路と、その前提ドキュメント/検証導線を repo から外す

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| M1 | Claude-mem MCP ラッパー・セットアップ/検証スクリプト・Cursor向け参照を削除し、残る文言を整合させる | `rg` で対象参照が実運用ファイルから消えている | - | cc:完了 |
| M2 | `harness-mem` を維持したまま、旧メモリ名称の user-facing 文言を live な setup/hook/skill から除去する | `rg` で対象参照が live 設定・主要スキルから消えている。内部互換パスは除く | M1 | cc:完了 |

---

## Phase 25: ソロモード PM フレームワーク強化

作成日: 2026-03-08
起点: pm-skills (phuryn/pm-skills) との比較分析 — ソロモードでの PM 思考フレームワーク欠如を特定
目的: ソロモード（Claude Code 単独運用）で PM 不在を補う「構造化された自問機構」を既存スキルに埋め込む

### 背景

- ハーネスは 2-Agent（Cursor PM + Claude Code Worker）前提で設計されたため、ソロモードでは PM 側の思考フレームワークが薄い
- pm-skills は 65 スキル / 36 チェーンワークフローで PM の思考構造化（Discovery, Strategy, Execution）をカバー
- ハーネスの強み（Evals 必須化、Plans.md マーカー、ガードレール）と pm-skills の強み（フレームワーク適用、段階的チェックポイント）は補完関係
- 新規スキル/コマンドは作らず、全て既存スキルの拡張として実装する

### 完了条件

1. harness-plan create の優先度判定が Impact × Risk の 2 軸マトリクスになっている
2. Plans.md テーブルに DoD カラムが追加され、create 時に自動生成される
3. harness-review の Plan Review に Value 軸が追加されている
4. harness-plan sync にレトロスペクティブ機能が統合されている
5. breezing の Phase 0 に構造化 3 問チェックが定義されている
6. harness-work Solo フローにタスク背景確認ステップが追加されている
7. Plans.md テーブルに Depends カラムが追加され、breezing が依存グラフを活用できる

### 優先度マトリクス

| 優先度 | Phase | 内容 | タスク数 | 依存 |
|--------|-------|------|---------|------|
| **Required** | 25.0 | Plans.md フォーマット拡張（DoD + Depends カラム） | 3 | なし |
| **Required** | 25.1 | harness-plan create 強化（2 軸マトリクス + DoD 自動生成） | 3 | 25.0 |
| **Required** | 25.2 | harness-review Plan Review 拡張（Value 軸） | 2 | なし |
| **Recommended** | 25.3 | harness-plan sync レトロ機能 | 2 | なし |
| **Recommended** | 25.4 | breezing Phase 0 構造化 + harness-work Solo 背景確認 | 3 | 25.0 |
| **Required** | 25.5 | 統合検証・バージョン・リリース | 3 | 25.0〜25.4 |

合計: **16 タスク**

---

### Phase 25.0: Plans.md フォーマット拡張 [P0]

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 25.0.1 | `harness-plan/references/create.md` の Plans.md 生成テンプレート（Step 6）を `| Task | 内容 | DoD | Depends | Status |` の 5 カラムに拡張 | テンプレートが 5 カラム形式になっている | - | cc:完了 |
| 25.0.2 | `harness-plan/references/sync.md` の差分検出ロジックを 5 カラム形式に対応させる（3 カラム Plans.md との後方互換を維持） | 旧 3 カラム Plans.md でもエラーなく動作する | 25.0.1 | cc:完了 |
| 25.0.3 | `harness-plan/SKILL.md` の Plans.md フォーマット規約セクションを 5 カラムに更新し、DoD / Depends の記法ガイドを追記 | SKILL.md 内のフォーマット規約が新テンプレートと一致 | 25.0.1 | cc:完了 |

### Phase 25.1: harness-plan create 強化 [P1]

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 25.1.1 | `harness-plan/references/create.md` の Step 5 を 2 軸マトリクス（Impact × Risk）に拡張。高 Impact × 高 Risk のタスクに `[needs-spike]` マーカーを自動付与し、spike タスクを自動生成 | Step 5 が 2 軸で評価され、高リスクタスクに spike が付く | 25.0.1 | cc:完了 |
| 25.1.2 | `harness-plan/references/create.md` の Step 6 で DoD カラムをタスク内容から自動推論して生成するロジックを追加 | 生成された Plans.md の全タスクに DoD が埋まっている | 25.0.1 | cc:完了 |
| 25.1.3 | `harness-plan/references/create.md` の Step 6 で Depends カラムをフェーズ内の依存関係から自動推論して生成するロジックを追加 | 依存のないタスクは `-`、依存ありは タスク番号が入る | 25.0.1 | cc:完了 |

### Phase 25.2: harness-review Plan Review 拡張 [P2] [P]

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 25.2.1 | `harness-review/SKILL.md` の Plan Review フローに Value 軸を追加（5 軸目: ユーザー課題との紐付き、代替手段の検討、Elephant 検出） | Plan Review が 5 軸（Clarity / Feasibility / Dependencies / Acceptance / Value）で評価される | - | cc:完了 |
| 25.2.2 | `harness-review/SKILL.md` の Plan Review で DoD カラム・Depends カラムの品質チェックを追加（空欄検出、検証不能な DoD の警告） | DoD 未記入タスクが警告される | - | cc:完了 |

### Phase 25.3: harness-plan sync レトロ機能 [P3] [P]

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 25.3.1 | `harness-plan/references/sync.md` に `--retro` フラグ対応を追加。完了タスクの振り返り（見積もり精度、ブロック原因パターン、スコープ変動）を出力 | `sync --retro` で振り返りサマリーが表示される | - | cc:完了 |
| 25.3.2 | `harness-plan/SKILL.md` の argument-hint と sync サブコマンド説明に `--retro` を追記 | SKILL.md に --retro の説明がある | 25.3.1 | cc:完了 |

### Phase 25.4: breezing Phase 0 構造化 + harness-work Solo 背景確認 [P4]

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 25.4.1 | `breezing/SKILL.md` の Phase 0: Planning Discussion に構造化 3 問チェック（スコープ確認、依存関係確認、リスクフラグ）を定義 | Phase 0 に 3 つの具体的チェック項目がある | 25.0.1 | cc:完了 |
| 25.4.2 | `harness-work/SKILL.md` の Solo フロー Step 1 と Step 2 の間に Step 1.5（タスク背景 30 秒確認）を追加。目的と影響範囲を推論表示し、自信がない場合のみ 1 問確認 | Solo フローに背景確認ステップが存在する | - | cc:完了 |
| 25.4.3 | `breezing/SKILL.md` の Phase 0 で Depends カラムを読み取り、依存グラフに基づくタスク割り当て順序を自動決定するロジックを追加 | Depends が空のタスクから先に Worker に割り当てられる | 25.0.1 | cc:完了 |

### Phase 25.5: 統合検証・バージョン・リリース [P5]

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 25.5.1 | `./tests/validate-plugin.sh` + `./scripts/ci/check-consistency.sh` 全体検証 | 全検証パス | 25.0〜25.4 | cc:完了 |
| 25.5.2 | VERSION バンプ + plugin.json 同期 + CHANGELOG 追記 | バージョンが同期されている | 25.5.1 | cc:完了 |
| 25.5.3 | GitHub Release 作成 | リリースが公開されている | 25.5.2 | cc:完了 |

---

## Phase 26: まさお理論適用 — 状態中心アーキテクチャへの転換

作成日: 2026-03-08
起点: まさお氏「マクロハーネス・ミクロハーネス・Project OS」3要素理論の分析
目的: 会話中心の運用から状態中心の運用へ転換し、自律実行の信頼性とセッション継続性を向上

### 背景

- まさお理論の3要素（マクロ/ミクロ/Project OS）と Harness を対照分析
- ミクロハーネス（breezing, guardrails, Agent Teams）は成熟済み — アップデート不要
- マクロハーネス（計画・監視・再計画）と Project OS（状態基盤）にギャップあり
- 3エージェント（Red Team / Architect / PM-UX）による多角的レビューで以下を確定:
  - KPI/Story 層は P0 から降格（ソロ開発では「管理」より「自動化」が優先）
  - Plans.md フォーマット変更は統一設計を先行（競合変更の防止）
  - プログレスフィード（breezing 中の進捗可視化）を新規追加

### 設計原則（3エージェント議論から導出）

1. **「管理」ではなく「自動化」を増やす** — 管理層を厚くするとユーザーが管理層を管理する逆説に陥る
2. **半自動→全自動の段階的移行** — 精度が安定するまでは提案→承認のフロー
3. **Plans.md 変更は一括設計してから実装** — 同じファイル群への競合変更を防ぐ
4. **任意フィールドをデフォルトにする** — 運用されない必須項目は害悪
5. **既存インフラを活用する** — 新しい仕組みより既存 hooks/skills の拡張を優先

### 優先度マトリクス

| 優先度 | Phase | 内容 | タスク数 | 依存 |
|--------|-------|------|---------|------|
| **Required** | 26.0 | 失敗→再チケット化フロー（半自動MVP） | 3 | なし |
| **Required** | 26.1 | harness-sync --snapshot | 3 | なし |
| **Recommended** | 26.2 | Artifact 軽量紐付け + プログレスフィード | 4 | なし |
| **Optional** | 26.3 | Plans.md v3 フォーマット統一設計 | 3 | 26.2 |
| **Required** | 26.4 | 統合検証・バージョン・リリース | 3 | 26.0〜26.3 |

合計: **16 タスク**

---

### Phase 26.0: 失敗→再チケット化フロー（半自動MVP） [P0] [P]

Purpose: 自己修正ループ失敗時に「止まるだけ」から「次の一手を提案してくれる」へ転換

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 26.0.1 | `harness-work/SKILL.md` の自己修正ループ終了処理に失敗原因分析ステップを追加。3回 STOP 時に失敗ログの要約 + 推奨アクション + 修正タスク案を生成 | 3回STOPで原因分析と修正タスク案が出力される | - | cc:完了 |
| 26.0.2 | 修正タスク案のユーザー承認フローを追加。承認時に Plans.md へ `cc:TODO` で自動追加、却下時はスキップ | 承認→Plans.md 追加、却下→スキップが動作する | 26.0.1 | cc:完了 |
| 26.0.3 | 全自動昇格条件を `decisions.md` に D30 として記録（提案採用率 80%+ で全自動化を検討） | D30 が記録されている | 26.0.1 | cc:完了 |

### Phase 26.1: harness-sync --snapshot [P0] [P]

Purpose: セッション再開時の「どこまでやったっけ」問題の根本解決

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 26.1.1 | `harness-sync/SKILL.md` に `--snapshot` サブコマンドを追加。Plans.md の WIP/TODO カウント + 最新 3 コミット + 未解決ブロッカーを 1 出力に集約 | `/harness-sync --snapshot` で状態サマリーが得られる | - | cc:完了 |
| 26.1.2 | `harness-sync/references/sync.md` に snapshot 生成ロジックを追加。Plans.md + 直近の decisions.md エントリ + git log を読み取り | snapshot が Plans.md 以外の状態も含む | 26.1.1 | cc:完了 |
| 26.1.3 | `harness-sync/SKILL.md` の argument-hint と sync サブコマンド説明に `--snapshot` を追記 | SKILL.md に --snapshot の説明がある | 26.1.1 | cc:完了 |

### Phase 26.2: Artifact 軽量紐付け + プログレスフィード [P1] [P]

Purpose: タスク完了の追跡性向上 + breezing 中のユーザー体験改善

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 26.2.1 | `harness-work/SKILL.md` のタスク完了処理で、`cc:完了` マーカー更新時に直近の commit hash を Status 内に付与（例: `cc:完了 [a1b2c3d]`） | タスク完了時に commit hash が自動付与される | - | cc:完了 |
| 26.2.2 | `harness-plan/references/sync.md` の差分検出ロジックを `cc:完了 [hash]` 形式に対応させる（後方互換: hash なしでもエラーなし） | 旧形式 Plans.md でもエラーなく動作する | 26.2.1 | cc:完了 |
| 26.2.3 | `breezing/SKILL.md` の Lead フローに、Worker タスク完了時の 1 行プログレスサマリー出力を追加（「Task 3/7 完了: ユーザー認証 API 実装」形式） | breezing 実行中にタスク完了ごとに進捗が表示される | - | cc:完了 |
| 26.2.4 | `scripts/hook-handlers/task-completed.sh` に進捗サマリー出力を追加（既存 TaskCompleted hook 基盤を活用） | TaskCompleted hook で進捗情報が出力される | 26.2.3 | cc:完了 |

### Phase 26.3: Plans.md v3 フォーマット統一設計 [P2]

Purpose: 将来の KPI/Story/Artifact カラム追加を一括設計し、競合変更を防止

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 26.3.1 | Plans.md v3 フォーマット仕様を設計。任意 Purpose 行（Phase ヘッダー）+ Artifact 表記の標準化 + 影響ファイル一覧を文書化 | 仕様書が作成され、影響ファイル一覧がある | - | cc:完了 |
| 26.3.2 | `harness-plan/references/create.md` の Plans.md 生成テンプレートに任意 Purpose 行を追加。デフォルトでは入力を求めない | Purpose 行が生成可能（省略可）。既存 Plans.md との後方互換維持 | 26.3.1 | cc:完了 |
| 26.3.3 | `decisions.md` に D31 として Plans.md v3 フォーマット設計判断を記録 | D31 が記録されている | 26.3.1 | cc:完了 |

### Phase 26.4: 統合検証・バージョン・リリース [P3]

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 26.4.1 | `./tests/validate-plugin.sh` + `./scripts/ci/check-consistency.sh` 全体検証 | 全検証パス | 26.0〜26.3 | cc:完了 |
| 26.4.2 | VERSION バンプ + plugin.json 同期 + CHANGELOG 追記 | バージョンが同期されている | 26.4.1 | cc:完了 |
| 26.4.3 | GitHub Release 作成 | リリースが公開されている | 26.4.2 | cc:完了 [56cdd77] |

---

## Phase 27: まさお理論適用の実装整合ハードニング

作成日: 2026-03-10
起点: `56cdd777 feat: state-centric architecture with masao theory` のレビュー
目的: Phase 26 で導入した「状態中心」機能のうち、説明先行になっている部分を実装・再開導線・追跡性まで含めて本当に閉じる

### 背景

- まさお理論との方向性自体は正しいが、「説明ではできること」と「実ランタイムで起きること」に一部ズレがある
- 特に「失敗→再チケット化」は TaskCompleted hook 側で修正タスク追加まで到達しておらず、実質は原因分析 + エスカレーション止まり
- `--snapshot` は保存設計までは入ったが、セッション再開時の自動読込・比較は未接続
- Project OS の最小要件（目的 / 受け入れ条件 / 上流参照 / 成果物リンク）のうち、上流参照がまだ薄い

### 優先度マトリクス

| 優先度 | Phase | 内容 | タスク数 | 依存 |
|--------|-------|------|---------|------|
| **Required** | 27.0 | 失敗→再チケット化の実ランタイム実装 | 3 | なし |
| **Required** | 27.1 | snapshot の再開導線接続 | 2 | なし |
| **Recommended** | 27.2 | Project OS 最小トレーサビリティ補強 | 3 | 27.1 |

合計: **8 タスク**

---

### Phase 27.0: 失敗→再チケット化の実ランタイム実装 [P0]

Purpose: Phase 26 の「次の一手をチケットとして残す」を説明ではなく実動作にする

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 27.0.1 | `scripts/hook-handlers/task-completed.sh` に `.fix` タスク案の構造化出力を追加し、失敗カテゴリ・元タスク番号・DoD・Depends を machine-readable に返す | 3回失敗時に修正タスク案が JSON もしくは安定フォーマットで取得できる | - | cc:完了 |
| 27.0.2 | 修正タスク案を `Plans.md` へ安全に追記する承認フローを実装する（承認時のみ追加、重複追加防止） | 承認で `.fix` タスクが1回だけ追加され、却下時は Plans.md が変化しない | 27.0.1 | cc:完了 |
| 27.0.3 | `skills/harness-work/SKILL.md` / `CHANGELOG.md` / `.claude/memory/decisions.md` の再チケット化説明を実装と一致させ、回帰検証を追加する | 「提案まで」「承認後追加」「全自動」の境界が全ファイルで一致し、再現手順がある | 27.0.2 | cc:完了 |

### Phase 27.1: snapshot の再開導線接続 [P0]

Purpose: 保存した状態を次セッションで本当に使えるようにして、状態中心アーキテクチャを閉じる

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 27.1.1 | `session-init` / `session-resume.sh` 系に最新 snapshot 読込を追加し、再開時に前回との差分サマリーを表示する | snapshot がある状態で再開すると差分サマリーが出る。ない場合は静かにスキップする | - | cc:完了 |
| 27.1.2 | `harness-sync --snapshot` の初回保存・2回目比較・再開時読込の検証手順を `tests/` またはドキュメント化された再現スクリプトとして固定する | 手順どおりに snapshot 保存→比較→再開確認を再現できる | 27.1.1 | cc:完了 |
| 27.1.3 | `session-init` と usage 記録フックの stdout ノイズを分離し、hook 出力が JSON 本体だけになることを回帰検証する | `session-init` / usage tracking が telemetry を出しても hook stdout が壊れず、直接実行の検証がある | 27.1.2 | cc:完了 |
| 27.1.4 | Claude SessionStart/UserPromptSubmit/PostToolUse/Stop を harness-mem runtime に接続し、`session-init` / `session-resume` で continuity briefing を初手に表示する | hooks.json が memory hook を呼び、Claude の SessionStart additionalContext に `Continuity Briefing` が出て、pending artifact が二重注入されない | 27.1.3 | cc:完了 |
| 27.1.5 | Claude memory lifecycle 回帰検証を追加し、開始→発話→停止まで同一 continuity chain が流れることを固定する | lifecycle 統合テストで `record-event` / `resume-pack` / `finalize-session` が同じ `correlation_id` を使い、SessionStart briefing が表示される | 27.1.4 | cc:完了 |

### Phase 27.2: Project OS 最小トレーサビリティ補強 [P1]

Purpose: 「なぜこのチケットがあるか」を上流へ辿れる最小フォーマットを、管理過多にならない範囲で足す
