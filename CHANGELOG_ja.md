# Changelog

このプロジェクトのすべての注目すべき変更は、このファイルに記録されます。

フォーマットは [Keep a Changelog](https://keepachangelog.com/ja/1.0.0/) に基づいており、
このプロジェクトは [Semantic Versioning](https://semver.org/spec/v2.0.0.html) に準拠しています。

> **📝 記載ルール**: ユーザー体験に影響する変更を中心に記載。内部修正は簡潔に。

## [2.20.8] - 2026-02-14

### Changed

- **Claude Code 2.1.41/2.1.42 対応**: 互換性マトリクスと推奨バージョンを v2.1.41+ に引き上げ
  - `docs/CLAUDE_CODE_COMPATIBILITY.md` に v2.1.39〜v2.1.42 の4バージョン・30+行の機能対応を追加
  - 推奨バージョンを v2.1.38+ → **v2.1.41+** に引き上げ（Agent Teams の Bedrock/Vertex/Foundry モデルID修正、Hook stderr 表示修正が主因）
- **Breezing Bedrock/Vertex/Foundry 注記**: `guardrails-inheritance.md` に CC 2.1.41+ 必須の注記追加
- **セッション `/rename` 自動命名**: session スキルに CC 2.1.41+ のセッション名自動生成ドキュメント追加
- **トラブルシュート `claude auth` コマンド**: 診断テーブルに CC 2.1.41+ の `claude auth login/status/logout` を追加

---
## [2.20.7] - 2026-02-14

### Fixed

- **Stop フックの「JSON validation failed」エラー (#42)**: 信頼性の低い `type: "prompt"` フックを確定的な `type: "command"` フック（`stop-session-evaluator.sh`）に置換
  - 根本原因: prompt type フックが LLM に JSON 形式での応答を指示していたが、モデルが自然言語を返すことが多く、毎ターン JSON パースエラーが発生
  - 新しいコマンドベースの評価スクリプトは常に有効な JSON を出力し、バリデーションエラーを完全に排除
  - `hooks/hooks.json` と `.claude-plugin/hooks.json` の両方を同期して更新

---
## [2.20.6] - 2026-02-14

### Fixed

- **session-auto-broadcast.sh の hookEventName バリデーションエラー** (#41):
  - `hookEventName` を `"AutoBroadcast"` → `"PostToolUse"` に修正（4箇所）
  - `session-broadcast.sh` の `hookEventName` を `"Broadcast"` → `"PostToolUse"` に修正
  - subprocess の stdout 汚染を防止（`>/dev/null` リダイレクト追加）
  - `test-hook-event-names.sh` テスト追加（hookEventName 一貫性の回帰テスト）

---
## [2.20.5] - 2026-02-12

### Fixed

- **Breezing `--codex` の subagent_type 選択を強制化**: `--codex` フラグが Implementer spawn 時に無視される問題を修正
  - 根本原因: `execution-flow.md` の Step 3 が `task-worker` をハードコードし、`--codex` 分岐が存在しなかった
  - SKILL.md、execution-flow.md、team-composition.md に `impl_mode` による必須分岐を追加
  - 3つの「絶対禁止」ルールを追加: codex モードでは `codex-implementer` 必須、standard モードでは `task-worker` 必須、codex モードの Lead はソースコードの直接 Write/Edit 禁止
  - 並列 spawn の明示指示を追加: N 個の Implementer を同時 spawn（`N = min(独立タスク数, --parallel N, 3)`）
  - Compaction Recovery が `impl_mode` に基づき正しい subagent_type を復元するよう修正

---

## [2.20.4] - 2026-02-11

### Fixed

- **Codex MCP → CLI 移行（Phase 7 完了）**:
  - `pretooluse-guard.sh`（4箇所）と `codex-worker-engine.sh`（1箇所）の `mcp__codex__codex` テキスト参照を `codex exec (CLI)` に置換
  - `codex-review/SKILL.md` から MCP レガシー注記を削除
  - `.claude/rules/codex-cli-only.md` ルールを追加（再発防止）
  - PreToolUse フック failsafe を追加: `mcp__codex__*` ツール呼び出しを `emit_deny` + `msg()` パターンでローカライズメッセージ付き拒否
  - opencode/codex ミラーの開発専用スキル（`test-*`, `x-promo`, `x-release-harness`）を `.gitignore` に追加

### Security

- **Codex MCP 二重防御**: 廃止済み MCP 使用に対する3層防御（テキスト修正 + フックブロック + ルールファイル）。Codex レビュー: Security A, Architect B

---

## [2.20.3] - 2026-02-10

### Fixed

- **フックハンドラのセキュリティ強化** (Codex レビュー Round 1-3):
  - 手動 JSON エスケープを `jq -nc --arg` / `python3 json.dumps` に置換（安全な JSON 構築）
  - Python コードインジェクション脆弱性を修正: データを `sys.argv`/`stdin` 経由で渡すように変更
  - `set -euo pipefail` 下での `grep` 失敗を `|| true` で修正
  - `grep -F` による固定文字列マッチング（正規表現メタ文字問題を回避）
  - `.claude/state` ディレクトリに `chmod 700` を追加
  - description 切り詰め時の `tostring` 型安全ガードを追加
  - TeammateIdle イベントの 5 秒重複抑制を追加
  - JSONL ローテーション（500 → 400 行）で無制限増加を防止

---

## [2.20.2] - 2026-02-10

### Added

- **TeammateIdle/TaskCompleted フックハンドラ**: `scripts/hook-handlers/teammate-idle.sh` と `task-completed.sh` を新規作成。Agent Teams のイベントを `.claude/state/breezing-timeline.jsonl` に記録
- **3層メモリアーキテクチャ (D22)**: Claude Code 自動メモリ、Harness SSOT、Agent Memory の共存設計を `decisions.md` に記録
- **Task(agent_type) パターン (P18)**: サブエージェント種別制限構文を `patterns.md` に記録

### Changed

- **Claude Code 2.1.38+ 対応**: CLAUDE.md の Feature Table に 6 行追加（TeammateIdle/TaskCompleted Hook、Agent Memory、Fast mode、自動メモリ記録、スキルバジェットスケーリング、Task(agent_type) 制限）
- **バージョン参照更新**: 全スキル・エージェントの「CC 2.1.30+」を「CC 2.1.38+」に更新（16+ ファイル）
- **スキルバジェットスケーリング**: `skill-editing.md` の 500 行ハードルールを推奨に緩和、CC 2.1.32+ の 2% スケーリングを注記
- **セッションメモリ**: `session-memory/SKILL.md` と `memory/SKILL.md` に「自動メモリとの関係（D22）」セクション追加
- **Breezing 実行フロー**: `execution-flow.md` のフック実装状態を「実装済み」に更新
- **ガードレール継承**: 安全メカニズムテーブルに Task(agent_type) を追加

---

## [2.20.0] - 2026-02-08

### 🎯 What's Changed for You

**28スキルを19に統合。Breezing の Phase A/B/C 分離、Teammate 権限修正、リポジトリクリーンアップを実施。**

| Before | After |
|--------|-------|
| `memory`, `sync-ssot-from-memory`, `cursor-mem` の3スキル | `memory` 1つに統合（SSOT昇格・記憶検索を references に移設） |
| `setup` 系が6スキルに分散 | `setup` 1つに統合（ルーティングテーブルで分岐） |
| `ci`, `agent-browser`, `x-release-harness` がメニューに露出 | `user-invocable: false` で非表示化（自動ロード経由でアクセス可能） |
| `/breezing` 開始直後に bypass permissions が失われる | Phase A で bypass 維持 → Phase B でのみ delegate |
| 完了ステージでも delegate mode のまま | Phase C で delegate 解除 → Lead が直接コミット可能 |
| Teammate が "prompts unavailable" で Bash 自動拒否 | `mode: "bypassPermissions"` + PreToolUse hooks で安全に解決 |
| ビルド成果物・開発専用ドキュメントが git に追跡されていた | 33ファイルを untrack、.gitignore 更新 |

### Changed

- **スキル統合 (28 → 19)**:
  - `/memory`: `sync-ssot-from-memory` と `cursor-mem` を吸収
  - `/setup`: `setup-tools`, `harness-mem`, `codex-setup`, `2agent`, `localize-rules` を吸収
  - `/troubleshoot`: CI 障害トリガーを description に追加
- **Breezing Phase 分離**: Phase A (Pre-delegate) / Phase B (Delegate) / Phase C (Post-delegate) の3段階に構造化
  - Phase A: ユーザーのパーミッションモードを維持したまま Team 初期化・spawn
  - Phase B: delegate mode で Lead は調整専念（TaskCreate/TaskUpdate/SendMessage のみ）
  - Phase C: delegate 解除後に統合検証・コミット・クリーンアップ
- **Teammate 権限モデル**: 全 Teammate spawn に `mode: "bypassPermissions"` を指定
  - PreToolUse hooks は権限システムと独立して発火（公式仕様）
  - 安全層: disallowedTools + spawn prompt 制約 + .claude/rules/ + Lead 監視
- **英語リリース**: GitHub リリースノートを英語に統一。リリースルール・スキルを更新
- **全関連ドキュメント更新**: execution-flow.md, team-composition.md, codex-engine.md, guardrails-inheritance.md, session-resilience.md

### Added

- `skills/memory/references/cursor-mem-search.md` - Cursor 連携メモリ検索リファレンス
- `skills/setup/references/harness-mem.md` - Harness-Mem セットアップリファレンス
- `skills/setup/references/localize-rules.md` - ルールローカライズリファレンス
- **Codex 初回チェックフック**: `/codex-review` 初回使用時に `check-codex.sh` を自動実行（`once: true`）
- **timeout/gtimeout 検出**: macOS ユーザー向けに `brew install coreutils` を案内

### Fixed

- **Codex レビュー指摘 22 件修正**: pretooluse-guard の JSON パース統合（5→1 jq call）、symlink セキュリティガード追加、session-monitor の `eval` 除去
- **macOS 互換性**: 全ドキュメントの `timeout N codex exec` → `$TIMEOUT N codex exec` に置換（GNU coreutils 非依存化）
- **Teammate Bash 自動拒否**: バックグラウンド Teammate の "prompts unavailable" エラーを解決

### Removed

- **33ファイルを untrack**: `mcp-server/dist/`（ビルド成果物 24件）、`docs/design/`（2件）、`docs/slides/`（1件）、`docs/claude-mem-japanese-setup.md`、開発専用ドキュメント（3件）、ロックファイル（2件）
- **アーカイブ済みスキル**: `sync-ssot-from-memory`, `cursor-mem`, `setup-tools`, `harness-mem`, `codex-setup`, `2agent`, `localize-rules` → `skills/_archived/`

---

## [2.18.11] - 2026-02-06

### 🎯 What's Changed for You

**`--codex` モード時、Claude は PM として振る舞い、Edit/Write が自動ブロックされるようになりました**

| Before | After |
|--------|-------|
| `--codex` 時も Claude が直接編集可能 | Claude の Edit/Write は Plans.md 以外ブロック |
| 役割分担が曖昧 | PM（Claude）と Worker（Codex）の明確な分離 |

### Added

- **breezing スキル**: Agent Teams を活用した完全自動タスク完走
  - Lead は delegate mode で調整専念、実装は Implementer、レビューは独立 Reviewer
  - `--codex-review` でマルチ AI レビュー統合
  - `/ultrawork` より大規模なタスクセット向け
- **Codex モードガード**: `pretooluse-guard.sh` に Codex モード検出を追加
  - Claude が PM 役として機能し、実装は Codex Worker に委譲
  - `ultrawork-active.json` の `codex_mode: true` で有効化
  - Plans.md の状態マーカー更新のみ許可

### Changed

- **Codex レビュー改善**: 並列レビューの品質向上
  - SSOT（decisions.md/patterns.md）を考慮した文脈あるレビュー
  - 出力制限を 1500 → 2500 文字に緩和（十分な分析のため）
  - 終了条件を明確化（Critical/High = 0 で APPROVE）
  - 「重箱の隅つつき」問題を解消（Low/Medium のみは APPROVE）

### Internal

- エキスパートテンプレートの軽微な修正

---

## [2.18.10] - 2026-02-06

### Added

- **エージェント永続メモリ**: 全7エージェントに `memory: project/user` 設定を追加
  - サブエージェントが会話間で制度的知識を蓄積可能に
  - セキュリティ: Read-only エージェント（code-reviewer, project-analyzer）は Bash/Write/Edit 禁止を維持
  - プライバシーガード: 各エージェントにシークレット/PII 保存禁止ルールを明記

---

## [2.18.7] - 2026-02-05

### Changed

- **Claude ガード**: 通常の `git push` では止めず、`-f/--force/--force-with-lease` のみ prompt するように変更しました。

---

## [2.18.6] - 2026-02-05

### Fixed

- **Codex ガード**: `.codex/rules/harness.rules` が安定してパースされ、`git clean -n` / `sudo -n true` のような安全コマンドで止まりにくくなりました（破壊的コマンドは prompt）。
- **Claude ガード**: `templates/claude/settings.security.json.template` のパーミッション構文を修正し、破壊的操作のみ prompt するように見直しました。

### Internal

- **Codex パッケージテスト**: rules の例（match/not_match）検証を追加し、起動時のパース失敗を防止。

---

## [2.18.5] - 2026-02-05

### Added

- **gogcli-ops スキル**: Google Workspace CLI 操作（Drive/Sheets/Docs/Slides）
  - 認証フロー / 複数アカウント選択
  - URL → ID 解決（`gog_parse_url.py`）
  - 原則 read-only、書き込みは明示確認

---

## [2.18.4] - 2026-02-04

### Added

- **Codex セットアップコマンド**: `/codex-setup` と `scripts/codex-setup-local.sh` を追加
- **Setup tools**: `/setup-tools codex` でセッション内セットアップ
- **Harness init/update**: `/harness-init` と `/harness-update` に Codex CLI 同期を追加

---

## [2.18.2] - 2026-02-04

### Added

- **Codex CLI 配布物**: `codex/.codex` に全スキルと暫定 Rules ガードを追加
- **Codex セットアップ**: `scripts/setup-codex.sh` と `codex/README.md` を追加
- **Codex AGENTS**: `$skill` 呼び出し向け `codex/AGENTS.md` を追加
- **Codex パッケージテスト**: `tests/test-codex-package.sh` を追加

### Changed

- **ドキュメント**: README に Codex CLI セットアップ手順を追記

---

## [2.18.1] - 2026-02-04

### Added

- **Aivis/VOICEVOX TTS 対応**: generate-video スキルで日本語 TTS プロバイダーを追加
  - `aivis`: Aivis Cloud API（speaker_id, intonation_scale 等）
  - `voicevox`: VOICEVOX（ずんだもん等のキャラクター音声）
  - サンプルキャラクター設定を追加

### Changed

- **MCP サーバーのオプション化**: `.mcp.json` を削除し、mcp-server を配布から除外
  - 必要なユーザーは別途セットアップ

---

## [2.18.0] - 2026-02-04

### Added

- **Claude Code 2.1.38 対応**: 新機能との完全統合
  - **AgentTrace v0.3.0**: Task tool メトリクス対応（tokenCount, toolUses, duration）
  - **`/debug` コマンド連携**: troubleshoot スキルが複雑なセッション問題に `/debug` を案内
  - **PDF ページ範囲読み込み**: notebookLM, harness-review で `pages` パラメータ対応
  - **Git log 拡張フラグ**: `--format`, `--raw`, `--cherry-pick` を活用
  - **68% メモリ最適化**: `--resume` の恩恵をドキュメント化
  - **サブエージェント MCP アクセス**: MCP ツール共有のバグ修正対応

---

## [2.17.10] - 2026-02-04

### Added

- **PreCompact/SessionEnd フック**: セッション状態の自動保存・クリーンアップに対応
- **AgentTrace v0.2.0**: Attribution フィールド追加（プラグイン帰属情報の記録）
- **Sandbox 設定テンプレート**: `templates/settings/harness-sandbox.json` を追加

### Changed

- **context: fork 追加**: deploy/generate-video/memory/verify スキルで独立コンテキストを使用
- **release → release-harness**: Claude Code 組み込みコマンドとの名前衝突を回避

---

## [2.17.9] - 2026-02-04

### Changed

- **Codex モードをデフォルトに**: 新規プロジェクトの設定テンプレートで `review.mode: codex` がデフォルトに
- **Worktree 必要性判定**: `/ultrawork --codex` 実行時に Worktree が本当に必要か自動判定
  - タスク1つのみ、全タスク順次依存、ファイル重複あり → 直接実行モードにフォールバック
  - 不要な Worktree 作成オーバーヘッドを回避

---

## [2.17.8] - 2026-02-04

### Fixed

- **release スキル**: Skill ツールで `/release` が起動できない問題を修正
  - `disable-model-invocation: true` を削除

---

## [2.17.3] - 2026-02-03

### 🎯 あなたにとって何が変わるか

**Ultrawork がレビュー後に自動で自己修正ループに入るようになりました**

| Before | After |
|--------|-------|
| レビュー後に手動でプロンプト入力が必要 | APPROVE まで自動修正ループ |
| Codex 有無を手動で指定 | Codex MCP 自動検出 + フォールバック |
| 改善方法が不明確 | 「🎯 How to Achieve A」で改善指針を明示 |

### Added

- **自己修正ループ**: `/harness-review` 実行後、APPROVE になるまで自動で修正を繰り返す
  - リトライ状態管理（`ultrawork-retry.json`）で進捗追跡
  - REJECT/STOP は即停止して手動介入を促す
  - 最大3回のリトライ後に STOP

- **検証全実行規則**: 存在する検証スクリプトを優先順で全て実行し、失敗で即停止

- **改善指針テンプレート**: 「🎯 How to Achieve A」セクションで A 評価達成方法を明示
  - Decision 別統一フォーマット（APPROVE/REQUEST CHANGES/REJECT/STOP）

### Changed

- **Codex 自動検出**: Codex MCP が利用可能な場合は自動で Codex モードに切り替え
  - 利用不可の場合はサブエージェント並列にフォールバック
  - `timeout_ms`（ミリ秒単位）でタイムアウト設定可能

- **差分計算改善**: `merge-base` 基準で変更ファイル数を算出
  - staged/unstaged 差分も含む
  - 初回コミット/マージにも対応

- **review_aspects 検出**: パスベースの正規表現で決定的に判定

---

## [2.17.2] - 2026-02-03

### 🎯 あなたにとって何が変わるか

**Codex Worker 完了時に Plans.md が自動更新されるようになりました**

| Before | After |
|--------|-------|
| 作業完了後に手動で Plans.md を更新 | スキルが自動で `cc:done` に更新 |

### Added

- **Plans.md 自動更新**: Codex Worker スキル完了時に必ずタスク完了処理を実行
  - 該当タスクを自動特定
  - `[ ]` → `[x]`, `cc:WIP` → `cc:done` に更新
  - タスクが見つからない場合はユーザーに確認

### Internal

- Codex Worker スクリプト品質改善（共通ライブラリ化、セキュリティ強化）

---

## [2.16.20] - 2026-02-03

### Added

- `ultrawork` スキルに Options テーブルと Quick Reference 例を追加（`--codex`, `--parallel`, `--worktree-base`）

### Internal

- スキルルーティングルールを `skills/routing-rules.md` に一元化（SSOT パターン導入）
- `codex-review` と `codex-worker` のルーティングを決定的に（文脈判定を排除）

---

## [2.16.19] - 2026-02-03

### Fixed

- Stop フックの reason が2回表示される問題を軽減（キーワードのみ出力に変更）

---

## [2.16.17] - 2026-02-03

### 🎯 あなたにとって何が変わるか

**スキルの使い方ヒントがオートコンプリートに表示されるようになりました**

| Before | After |
|--------|-------|
| `/harness-review` | `/harness-review [code|plan|scope]` |
| `/troubleshoot` | `/troubleshoot [build|test|runtime]` |

### Added

- 17スキルに使い方ヒント（`argument-hint`）を追加
- セッション間通知機能（複数セッション連携時に便利）

### Internal

- CI/テスト/ドキュメントを Skills 移行後の構造に更新

---

## [2.16.14] - 2026-02-02

### 🎯 あなたにとって何が変わるか

**実装を依頼すると、自動的に Plans.md に登録されます**

| Before | After |
|--------|-------|
| 口頭依頼が Plans.md に残らない | すべてのタスクが Plans.md に記録 |
| 進捗が追いにくい | `/sync-status` で全体把握可能 |

---

## [2.16.11] - 2026-02-02

### 🎯 あなたにとって何が変わるか

**コマンドがスキルに統合されました（使い方は変わりません）**

| Before | After |
|--------|-------|
| `/work`, `/harness-review` がコマンドとして存在 | 同じ名前でスキルとして動作 |
| 内部スキル (impl, verify) がメニューに表示 | 非表示に（ノイズ軽減） |
| `dev-browser`, `docs`, `video` | `agent-browser`, `notebookLM`, `generate-video` に改名 |

### Internal

- README を VibeCoder 向けにリライト（トラブルシューティング・アンインストール追加）
- CI スクリプトを Skills 構造に対応

---

## [2.16.5] - 2026-01-31

### 🎯 あなたにとって何が変わるか

**`/generate-video` が AI 画像生成・BGM・字幕・視覚効果に対応**

| Before | After |
|--------|-------|
| 画像素材は手動で用意 | AI が自動生成（Nano Banana Pro） |
| BGM・字幕なし | 著作権フリー BGM、日本語字幕対応 |
| 基本トランジションのみ | GlitchText, Particles 等のエフェクト |

---

## [2.16.0] - 2026-01-31

### 🎯 あなたにとって何が変わるか

**`/ultrawork` で rm -rf と git push の確認回数が減りました（実験的機能）**

| Before | After |
|--------|-------|
| rm -rf で毎回確認 | 計画時に許可したパスのみ自動承認 |
| git push で毎回確認 | ultrawork 中は自動承認（force除く） |

---

## [2.15.0] - 2026-01-26

### 🎯 あなたにとって何が変わるか

**OpenCode との完全互換モードを追加**

| Before | After |
|--------|-------|
| OpenCode 向けに別途設定が必要 | `/setup-opencode` で自動セットアップ |
| skills/ 構造が異なる | 同一スキルが両環境で動作 |

---

## [2.14.0] - 2026-01-16

### 🎯 あなたにとって何が変わるか

**`/work --full` で並列タスク実行が可能に**

| Before | After |
|--------|-------|
| タスクを1つずつ実行 | `--parallel 3` で最大3並列実行 |
| 完了報告を手動で確認 | 各 worker が自律的にセルフレビュー |

---

## [2.13.0] - 2026-01-14

### 🎯 あなたにとって何が変わるか

**Codex MCP による並列レビューを追加**

| Before | After |
|--------|-------|
| Claude 単体でレビュー | Codex 4エキスパートが並列でレビュー |
| 一度に1観点 | セキュリティ/品質/パフォーマンス/a11y を同時チェック |

---

## [2.12.0] - 2026-01-10

### Added

- **Harness UI ダッシュボード** (`/harness-ui`) - ブラウザで進捗確認
- **ブラウザ自動化** (`agent-browser`) - ページ操作・スクリーンショット

---

## [2.11.0] - 2026-01-08

### Added

- **セッション間メッセージング** - 複数 Claude Code セッション間でメッセージ送受信
- **CRUD 自動生成** (`crud` スキル) - Zod バリデーション付きエンドポイント生成

---

## [2.10.0] - 2026-01-04

### Added

- **LSP 統合** - Go-to-definition, Find-references で正確なコード理解
- **AST-Grep 統合** - 構造的なコードパターン検索

---

## 過去バージョン

v2.9.x 以前の詳細は [GitHub Releases](https://github.com/Chachamaru127/claude-code-harness/releases) を参照してください。

[2.20.8]: https://github.com/Chachamaru127/claude-code-harness/compare/v2.20.7...v2.20.8
[2.20.7]: https://github.com/Chachamaru127/claude-code-harness/compare/v2.20.6...v2.20.7
[2.20.6]: https://github.com/Chachamaru127/claude-code-harness/compare/v2.20.5...v2.20.6
[2.20.5]: https://github.com/Chachamaru127/claude-code-harness/compare/v2.20.4...v2.20.5
[2.18.7]: https://github.com/Chachamaru127/claude-code-harness/compare/v2.18.6...v2.18.7
