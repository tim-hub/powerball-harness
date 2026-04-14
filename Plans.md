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
| 35.7.2 | npm パッケージ設定 + postinstall でプラットフォーム別バイナリ配置 | `npm install` で `bin/harness` が PATH に配置 | 35.7.1 | cc:完了 |
| 35.7.3 | 旧パッケージへの移行通知 + GitHub Release 自動化 | リリースワークフローで Go バイナリが含まれる | 35.7.2 | cc:完了 |

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

---

## Phase 36: skills-v3/ 統合 — SSOT を skills/ に一本化

作成日: 2026-04-07
目的: `skills-v3/` ディレクトリを廃止し、`skills/` を唯一の SSOT にする。ミラー同期は `skills/ → codex/.codex/skills/`, `opencode/skills/` の 2 方向のみに簡素化。

### 背景

Phase A（前セッション完了済み）で `agents-v3/ → agents/`、`skills-v3-codex/ → skills-codex/` のリネームは完了。
Phase B は `skills-v3/` の統合で、ミラースクリプト群の再設計が必要なため別 PR に分離していた。

### 設計方針

- `skills-v3/` の 6 コアスキル + breezing + routing-rules.md は既に `skills/` にミラーコピーが存在
- 統合後は `skills/` が正本。`skills-v3/` は削除
- `skills-v3/extensions/` の symlink 10 本は `skills/` に実体があるため不要
- ミラー同期スクリプトのソースパスを `skills-v3/` → `skills/` に変更
- "v3" という名前を全てのパス・ドキュメントから除去（ただし CHANGELOG の過去エントリは除く）

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 36.1 | `skills-v3/` の内容を `skills/` に最終同期し、差分がないことを確認 | `diff -r skills-v3/{skill} skills/{skill}` で差分 0 | - | cc:完了 |
| 36.2 | `sync-v3-skill-mirrors.sh` を書き換え: ソースを `skills/` に変更、スクリプト名を `sync-skill-mirrors.sh` にリネーム | `./scripts/sync-skill-mirrors.sh` と `--check` が正常動作 | 36.1 | cc:完了 |
| 36.3 | `check-consistency.sh` のセクション [10/12] を更新: `skills-v3/` 参照を `skills/` に変更 | `./scripts/ci/check-consistency.sh` が全パス | 36.2 | cc:完了 |
| 36.4 | `validate-plugin-v3.sh` を更新: `skills-v3/` → `skills/` に変更 | スクリプトが正常動作 | 36.2 | cc:完了 |
| 36.5 | `generate-skill-manifest.sh` の roots から `skills-v3` を削除 | マニフェスト生成が `skills/` のみをスキャン | 36.2 | cc:完了 |
| 36.6 | `fix-symlinks.sh` のソースパスを `skills/` に変更 | Windows 互換修復ロジックが `skills/` を正本として動作 | 36.2 | cc:完了 |
| 36.7 | `set-locale.sh` の `skills-v3` 参照を除去 | ロケール切替が `skills/` のみを処理 | 36.2 | cc:完了 |
| 36.8 | ドキュメント一括更新: README.md, README_ja.md, v3-architecture.md, CLAUDE.md 等の `skills-v3` 参照を更新 | `grep -r 'skills-v3' --include='*.md'` が CHANGELOG 以外で 0 件 | 36.1 | cc:完了 |
| 36.9 | SKILL.md 内の `skills-v3` 参照を更新（全ミラー含む） | 全 SKILL.md で `skills-v3` 参照なし | 36.8 | cc:完了 |
| 36.10 | `skills-v3/` ディレクトリ + 旧 `sync-v3-skill-mirrors.sh` を削除 | `ls skills-v3/` が存在しない | 36.1-36.9 | cc:完了 |
| 36.11 | テスト群の `skills-v3` 参照を更新 | `grep -r 'skills-v3' tests/` が 0 件 | 36.10 | cc:完了 |
| 36.12 | 統合検証: `check-consistency.sh` + `validate-plugin.sh` + `go build` + `go test` 全パス | CI 相当の全検証パス（既知の pre-existing issue 除く） | 36.10, 36.11 | cc:完了 |

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

## Phase 38: CC 2.1.89-2.1.100 追従 + Go v4 リリース完璧化

作成日: 2026-04-10
目的: Claude Code 2.1.89-2.1.100 の未取込 hook/permission/plugin 変更を Harness v4 Go ガードレールに取り込み、Go v4.0.0 リリース前にセキュリティ退行ゼロの完璧な状態を達成する

背景: CC 2.1.98 で本体が塞いだ 2 つの脆弱性 (backslash-escaped flag bypass, env-var prefix bypass) が Harness 二層目ガードレールでまだ開いたまま。さらに CC 2.1.89 の `DecisionDefer` は `go/pkg/hookproto/types.go` に型定義済みだが `go/internal/guardrail/pre_tool.go` の `PreToolToOutput()` switch で拾えていない既知ギャップあり。.husky 保護 / symlink 解決 / wildcard 正規化 / plugin.json skills 明示化 / Monitor ツール取込も v4 リリース前に纏めて取り込む。

### 優先度マトリクス

| 優先度 | Phase | 内容 | タスク数 | 依存 |
|--------|-------|------|---------|------|
| **Required** | 38.0 | セキュリティ緊急修正 (permission.go ハードニング + DecisionDefer ワイヤリング) | 2 | なし |
| **Required** | 38.1 | セキュリティ補強 (.husky + symlink + wildcard 正規化) | 2 | なし |
| **Required** | 38.2 | プラグイン/スキル整合 + Monitor ツール取込 | 2 | なし |
| **Required** | 38.3 | 統合検証・バイナリ再ビルド・CHANGELOG | 3 | 38.0-38.2 全て |

合計: **9 タスク**

### 完成基準 (Definition of Done — Phase 38 全体)

| # | 基準 | 検証方法 | 必須/推奨 |
|---|------|---------|----------|
| 1 | ガードレール新規セキュリティテスト 16+ 本追加、全 PASS | `go test -v ./internal/guardrail/...` | 必須 |
| 2 | 全 Go テスト PASS | `go test ./...` | 必須 |
| 3 | プラグイン検証 PASS | `./tests/validate-plugin.sh` | 必須 |
| 4 | 一貫性チェック PASS | `./scripts/ci/check-consistency.sh` | 必須 |
| 5 | Feature Table に CC 2.1.98 Monitor 追記 + 付加価値列 "A: 実装あり" | `docs/CLAUDE-feature-table.md` 目視 | 必須 |
| 6 | CHANGELOG.md Unreleased に Phase 38 の 7 項目 Before/After 追記 | `CHANGELOG.md` 目視 | 必須 |
| 7 | 3 プラットフォームバイナリ再ビルド済み | `ls -la bin/harness-*` | 必須 |
| 8 | CC 2.1.89-2.1.100 の security hardening が Go v4 に完全反映 | 基準 1-7 全達成 | 必須 |

---

### Phase 38.0: セキュリティ緊急修正 (permission.go + DecisionDefer) [P0]

Purpose: CC 2.1.98 で本体が塞いだ 2 つの脆弱性 (backslash-escape / env-var prefix) を Harness 二層目ガードでも塞ぐ。加えて 2.1.89 で追加された `DecisionDefer` が switch case で拾われていない既知ギャップを解消する

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 38.0.1 | `go/internal/guardrail/permission.go` ハードニング: (1) `hasBackslashEscape(cmd string) bool` 関数を追加し、`\-`, `\\ ` (空白エスケープ), `\--` 等のパターンを正規表現 `\\[\-\s]` で検出、(2) `stripSafeEnvPrefix(cmd string) (string, bool)` 関数と `knownSafeEnvVars` map (`LANG`, `LANGUAGE`, `TZ`, `NO_COLOR`, `FORCE_COLOR`) を追加。`LC_*` は prefix match で許可、(3) `isSafeCommand()` の先頭で両方を呼び、バックスラッシュエスケープ検出時 or 未知 env-var 検出時は即 `false` を返す [feature:security] | `go/internal/guardrail/permission_test.go` に最低 8 テスト追加し全 PASS: (a) `git\ status` reject, (b) `git\ push\ --force` reject, (c) `rm\ -rf\ /` reject, (d) `LANG=C git status` pass, (e) `TZ=UTC git log` pass, (f) `EVIL=x git status` reject, (g) `LANG=C NO_COLOR=1 git status` pass, (h) `LANG=C EVIL=x git status` reject。`go test ./internal/guardrail/...` 全 PASS | - | cc:完了 [aa9f4bb] |
| 38.0.2 | `go/internal/guardrail/pre_tool.go` の `PreToolToOutput()` 関数の switch 文に `case hookproto.DecisionDefer:` を追加し、`inner.PermissionDecision = "defer"` + `inner.PermissionDecisionReason = result.Reason` を設定する。既に `go/pkg/hookproto/types.go:39` に `DecisionDefer` 定数定義があるが switch で拾われていない既知問題を解消 [feature:security] | `go/internal/guardrail/pre_tool_test.go` に DecisionDefer 返却時のテスト追加、出力 JSON に `"permissionDecision": "defer"` と `"permissionDecisionReason": "<reason>"` が含まれる。`go test ./internal/guardrail/...` 全 PASS | - | cc:完了 [aa9f4bb] |

---

### Phase 38.1: セキュリティ補強 (helpers.go + rules.go) [P0]

Purpose: CC 2.1.89 の symlink target 解決、CC 2.1.90 の `.husky` 保護パス追加、CC 2.1.98 の wildcard whitespace 正規化の 3 つを Harness Go ガードレールに追従させる

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 38.1.1 | `go/internal/guardrail/helpers.go` 拡張: (1) `protectedPathPatterns` スライスに `.husky/` パターンを追加 (git hooks ディレクトリ保護、CC 2.1.90 追従), (2) `isProtectedPath(path string) bool` 内部で `filepath.EvalSymlinks()` を呼び、symlink の解決後実パスが protected patterns にマッチする場合も deny する (CC 2.1.89 追従)。`EvalSymlinks` エラー時は fail-safe として `true` (deny) を返す [feature:security] | `go/internal/guardrail/helpers_test.go` に最低 5 テスト追加し全 PASS: (a) `.husky/pre-commit` write deny, (b) `.husky/hooks/commit-msg` deny, (c) 一時ディレクトリで symlink `link-env → .env` を作成しアクセスが deny, (d) 2 段 symlink chain (link1 → link2 → .env) も deny, (e) symlink loop で `EvalSymlinks` エラーが返る場合も fail-safe で deny。`go test ./internal/guardrail/...` 全 PASS | - | cc:完了 [aa9f4bb] |
| 38.1.2 | `go/internal/guardrail/rules.go` の wildcard pattern 評価で、ユーザーコマンド側の連続 whitespace (スペース/タブ) を `regexp.MustCompile(\`\s+\`).ReplaceAllString(cmd, " ")` で単一スペースに正規化してからパターンマッチを行う。CC 2.1.98 で本体が修正した `Bash(git push -f:*)` が複数スペース/タブコマンドにマッチする挙動を二層目でも再現 [feature:security] | `go/internal/guardrail/rules_test.go` に最低 3 テスト追加し全 PASS: (a) `git  push  --force` (連続スペース) が force-push rule で deny, (b) `git\tpush\t-f` (タブ区切り) deny, (c) `git push   --force-with-lease` deny。既存テスト全 PASS を維持。`go test ./internal/guardrail/...` 全 PASS | - | cc:完了 [aa9f4bb] |

---

### Phase 38.2: プラグイン/スキル整合 + Monitor ツール取込 [P0]

Purpose: CC 2.1.94 の plugin skill invocation name 仕様に明示対応、CC 2.1.98 で追加された Monitor ツールを Breezing 等の長時間実行スキルで活用できるよう宣言・文書化する

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 38.2.1 | `.claude-plugin/plugin.json` に `"skills": ["./"]` フィールドを追加する。CC 2.1.94 以降の plugin skill invocation name が frontmatter `name` フィールド基準になる仕様に明示対応。既存 auto-discover との互換性を保つ (既存スキル全 32 個が引き続き invocation 可能であること) | `.claude-plugin/plugin.json` に `"skills": ["./"]` が存在。`./tests/validate-plugin.sh` PASS。`./scripts/ci/check-consistency.sh` PASS。jq で `.skills` が `["./"]` であることを確認 | - | cc:完了 [ebdf47b] |
| 38.2.2 | CC 2.1.98 Monitor ツール対応: (1) `skills/breezing/SKILL.md`, `skills/harness-work/SKILL.md`, `skills/ci/SKILL.md`, `skills/deploy/SKILL.md`, `skills/harness-review/SKILL.md` の frontmatter `allowed-tools` 配列に `"Monitor"` を追加、(2) `skills/breezing/SKILL.md` に「### Monitor ツール活用ガイド (CC 2.1.98+)」節を追加 (Worker 観察は Agent 層が完了通知するため Monitor 不要、シェルの長時間コマンド監視は Monitor を優先、という使い分け基準を明記。`gh run watch`, `go test -v`, `codex-companion.sh watch <job-id>` 等を具体例として列挙)、(3) `docs/CLAUDE-feature-table.md` に "Monitor ツール" 行を追加し付加価値列に "A: 実装あり (allowed-tools + 運用ガイド + Feature Table)" と記載 | 5 SKILL.md の frontmatter に `"Monitor"` が含まれる (`grep -l '"Monitor"' skills/*/SKILL.md` で 5 件)。breezing SKILL.md に `### Monitor ツール活用ガイド (CC 2.1.98+)` 節が存在。`docs/CLAUDE-feature-table.md` に Monitor 行が存在し付加価値列に "A: 実装あり" 記載。`.claude/rules/cc-update-policy.md` の「書いただけ検出」に該当しない (実装を伴うため) | - | cc:完了 [ebdf47b] |

---

### Phase 38.3: 統合検証・バイナリ再ビルド・CHANGELOG [P0]

Purpose: Phase 38.0-38.2 の全変更を統合し、バイナリを 3 プラットフォームで再ビルド、CHANGELOG に Before/After 形式で記録して Go v4.0.0 リリース可能状態に到達する

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 38.3.1 | `bin/harness` を darwin-arm64, darwin-amd64, linux-amd64 の 3 プラットフォームで再ビルド。既存の cross-platform ビルドスクリプト (`scripts/build-cross-platform.sh` 等があればそれを使用) がなければ `GOOS=darwin GOARCH=arm64 go build -o bin/harness-darwin-arm64 ./go/cmd/harness` を 3 パターン実行 | `bin/harness-darwin-arm64`, `bin/harness-darwin-amd64`, `bin/harness-linux-amd64` が全て更新されている (`ls -la` で最新タイムスタンプ)。各バイナリが `--version` で v4.0.0 を返す。バイナリサイズは Phase 37.5 追加分で ~10-11MB に増加 (Phase 38 の影響は無視できる程度) | 38.0.1, 38.0.2, 38.1.1, 38.1.2 | cc:完了 [fbed2f9] |
| 38.3.2 | `CHANGELOG.md` の `[Unreleased]` セクションに Phase 38 の全項目を日本語 Before/After 形式で追記。CC 2.1.89-2.1.100 の 7 項目 (backslash / env-var / defer / .husky+symlink / wildcard / plugin.json / Monitor) を網羅。`.claude/rules/github-release.md` の「CC バージョン統合時の CHANGELOG パターン」 (CC のアプデ → Harness での活用) 形式を使用 | `CHANGELOG.md` Unreleased に `#### N. Claude Code 2.1.98 統合` + `##### N-X.` 形式で 7 エントリ追加。各エントリに「CC のアプデ」「Harness での活用」の 2 段書き。VERSION / plugin.json version / harness.toml version は変更しない (リリース作業ではないため) | 38.0.1, 38.0.2, 38.1.1, 38.1.2, 38.2.1, 38.2.2 | cc:完了 [fbed2f9] |
| 38.3.3 | 統合テスト: (1) `go test ./...` 全 PASS, (2) `./tests/validate-plugin.sh` PASS, (3) `./scripts/ci/check-consistency.sh` PASS の 3 点確認。失敗した場合は原因特定して修正するまでタスク完了としない | 3 コマンド全 PASS。新規追加した 24 本のセキュリティテストが全て含まれていること、既存テストで退行がないことを確認。validate-plugin.sh は baseline 7 件失敗から 6 件失敗に 1 件改善 (plugin.json skills field 追加効果)、check-consistency.sh は baseline 2 件維持 (全て v4 cleanup 残骸で Phase 38 とは無関係) | 38.3.1, 38.3.2 | cc:完了 [fbed2f9] |

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
| 41.0.0 | **[Spike]** 2 つの API を実環境で確認する。(a) `scripts/harness-mem-client.sh` の実エンドポイント一覧を取得し、checkpoint 相当の操作（`ingest` / `record_event` / `finalize_session` / 新 API のいずれか）を特定、(b) CC が提供する `/loop` コマンドと内部 `ScheduleWakeup` ツールの実制約値（最小/最大 delaySeconds、cron syntax、wake-up ごとの状態継承）を検証、(c) `PreCompact` / `PostCompact` hooks が /loop の wake-up 間で発火するかを実測、(d) 結果を `.claude/memory/decisions.md` に D25「Phase 41 前提の実測結果」として記録 | (a) harness-mem の checkpoint 相当 API が特定されている（またはカスタム実装の必要性が結論付けられている）、(b) `ScheduleWakeup` の [60,3600]s 制約が実測で確認、(c) `/loop <interval>` の最小値が数値で判明、(d) decisions.md に 4 項目全て記録 + 41.0.2 以降のタスクで使う API 名が確定 | - | cc:TODO |
| 41.0.1 | `scripts/record-review-calibration.sh` L41-67 に新フィールド `critical_count` / `major_count` / `score_delta` を追加。既存レコード（2 件）には影響せず、新規記録からのみ書き込み。`scripts/build-review-few-shot-bank.sh` が新フィールドを読めるよう対応 | 新規 calibration 記録後、jsonl に 3 フィールドが含まれる。旧レコードへの読み出しは `// 0` default で動作。既存 few-shot bank 再生成テストが PASS | 41.0.0 | cc:TODO |
| 41.0.2 | `scripts/auto-checkpoint.sh` を新設。引数: task_id + commit_hash + sprint_contract_path + review_result_path。内部で 41.0.0 で特定した harness-mem API を呼び出して checkpoint を記録。成功・失敗いずれの経路でも `.claude/state/checkpoint-events.jsonl` に 1 行の audit レコード（`{"type":"checkpoint","status":"ok|failed","task":...,"commit":...,"timestamp":...}`）を必ず追記する（ローカル監査ログ）。harness-mem 失敗時は追加で `.claude/state/session-events.jsonl` にもデグレ出力（失敗を静かに吸収しない）。`.claude/state/locks/phase-b.lock` を flock で取得し同期保護 | (a) 正常系: harness-mem に 1 レコード追加 + checkpoint-events.jsonl に 1 行、(b) 異常系: harness-mem API 不達時に checkpoint-events.jsonl に `status:"failed"` 1 行 + session-events.jsonl に `checkpoint_failed` 1 行、(c) phase-b.lock が既に取得されている場合は timeout 10s 待機後に abort、(d) 単体実行 `bash scripts/auto-checkpoint.sh <task> <hash> <contract> <result>` で全挙動 OK、(e) 10 回連続実行でも lock デッドロックなし | 41.0.1 | cc:TODO |
| 41.0.3 | `scripts/detect-review-plateau.sh` を新設。入力: `.claude/state/review-calibration.jsonl` + 現在の task_id。ロジック: 同一 task_id の直近 N=3 エントリを抽出し、(a) iteration 数 ≥ 3、かつ (b) 修正対象ファイル集合の Jaccard 類似度 > 0.7 の両方を満たすなら `PIVOT_REQUIRED` を返す。N<3 なら `INSUFFICIENT_DATA`、条件不成立なら `PIVOT_NOT_REQUIRED`。`tests/fixtures/review-calibration/` に 3 種類の golden fixture を配置してテスト可能にする | (a) golden fixture `plateau.jsonl`（3 行全て Jaccard>0.7） → `PIVOT_REQUIRED`、(b) `improved.jsonl`（最終行で major=0） → `PIVOT_NOT_REQUIRED`、(c) `insufficient.jsonl`（2 行）→ `INSUFFICIENT_DATA`。exit code それぞれ 2 / 0 / 1。`bash scripts/detect-review-plateau.sh <task_id>` 単体で動作 | 41.0.1 | cc:TODO |

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
