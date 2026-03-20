# Claude Code Harness — Plans.md

最終アーカイブ: 2026-03-08（Phase 17〜24 → `.claude/memory/archive/Plans-2026-03-08-phase17-24.md`）

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
| 25.5.3 | GitHub Release 作成 | リリースが公開されている | 25.5.2 | cc:TODO |

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

### Phase 27.2: Project OS 最小トレーサビリティ補強 [P1]

Purpose: 「なぜこのチケットがあるか」を上流へ辿れる最小フォーマットを、管理過多にならない範囲で足す
