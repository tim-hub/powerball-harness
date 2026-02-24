# Claude Code 互換性マトリクス

このドキュメントは Claude Code Harness と Claude Code CLI の互換性を定義します。

## 現在の対応状況

| Harness バージョン | Claude Code 最小バージョン | 推奨バージョン | 備考 |
|-------------------|-------------------------|--------------|------|
| v2.9.0 | v2.1.1+ | v2.1.6+ | hooks, skills 基本機能 |
| v2.9.24 | v2.1.6+ | v2.1.21+ | Setup hook, plansDirectory, context_window, セッション間通信 |
| v2.14.9 | v2.1.6+ | v2.1.21+ | 4観点並列レビュー、auto-commit、OpenCode対応、MCP code intelligence |
| **v2.20.6** | v2.1.1+ | **v2.1.41+** | Agent Teams Bedrock/Vertex/Foundry 修正、Hook stderr 表示修正、起動性能改善 |
| **v2.21.0** | v2.1.1+ | **v2.1.49+** | Plugin settings.json、Worktree isolation、Background agents、ConfigChange hook、last_assistant_message |
| **v2.24.0** | v2.1.1+ | **v2.1.51+** | メモリリーク修正、WorktreeCreate/Remove hook、`claude agents` CLI、remote-control、ツール出力 50K 閾値変更 |

## バージョン別機能対応

### v2.1.51 (2026-02-24)

| 機能 | Harness 対応 | 備考 |
|------|-------------|------|
| `claude remote-control` サブコマンド（外部ビルドとローカル環境サービング） | 将来対応 | 外部ビルドシステムとの連携に将来活用可能。Breezing のクロスセッション制御に応用の余地あり |
| カスタム npm レジストリ・バージョンピニング（プラグインインストール） | 有利 | エアギャップ環境やバージョン固定が必要な環境でのプラグイン管理が改善 |
| ツール出力 50K 閾値変更（100K → 50K でディスク永続化） | **互換** | Harness フックの出力は最大 500B に切り詰め済み（`ci-status-checker.sh`）。影響なし |
| `statusLine`/`fileSuggestion` フックのワークスペース信頼検証修正 | **影響なし** | Harness はこれらのフックタイプを使用していない |
| SKILL.md YAML 配列 description のオートコンプリートクラッシュ修正 | **影響なし** | 全スキルの description は文字列形式を使用。脆弱性なし |
| モデルピッカーの人間可読ラベル表示（"Sonnet 4.5" 等） | 有利 | `/model` コマンドでの UX 向上 |
| `CLAUDE_CODE_ACCOUNT_UUID` 等の新環境変数 | 有利 | SDK テレメトリ用メタデータ。Harness は現時点で未使用 |

### v2.1.50 (2026-02-23)

| 機能 | Harness 対応 | 備考 |
|------|-------------|------|
| メモリリーク修正（LSP 診断・大型ツール出力・ファイル履歴・シェル実行） | **重要** | `/breezing` 等の長時間チームセッションの安定性が大幅改善。Harness 側は JSONL ローテーション（500→400 行）で既に対策済み |
| 完了タスクの GC（ガベージコレクション）修正 | **重要** | 多数のタスクを spawn する `/breezing` セッションでメモリ圧迫を防止 |
| `WorktreeCreate`/`WorktreeRemove` フックイベント | 将来対応 | Worktree 作成・削除時のカスタム VCS セットアップに活用可能。Breezing 並列ワークフローの自動化候補 |
| `claude agents` CLI コマンド（エージェント一覧・詳細表示） | **対応済み** | `troubleshoot` スキルの診断テーブルに追加。エージェント spawn 失敗時の診断に有用 |
| `isolation: worktree` のエージェント定義サポート（宣言的 worktree） | 有利 | v2.1.49 の Task tool パラメータに加え、エージェント定義で宣言的に worktree 分離を指定可能に |
| LSP `startupTimeout` 設定 | **対応済み** | `skills/setup/references/lsp-setup.md` に既に文書化済み |
| シンボリックリンクディレクトリでのセッション永続性修正 | 有利 | シンボリックリンクを使用するプロジェクトでのセッション可視性が改善 |
| `CLAUDE_CODE_SIMPLE` モードで skills/memory/agents を除外 | **対応済み** | SIMPLE モード使用時は Harness スキル（37）・エージェント（11）・メモリが無効化。フックのみ動作。SessionStart/Setup フックで自動検出・警告表示（v2.25.0+）。詳細: [SIMPLE_MODE_COMPATIBILITY.md](./SIMPLE_MODE_COMPATIBILITY.md) |
| WASM メモリの無制限成長修正（tree-sitter パーサーの定期リセット） | 有利 | 長時間セッションでの WASM メモリリーク解消 |
| Headless モード起動高速化（WASM/UI インポート遅延） | 有利 | `-p` フラグ使用時の起動性能向上 |

### v2.1.49 (2026-02-21)

| 機能 | Harness 対応 | 備考 |
|------|-------------|------|
| Plugin settings.json（プラグイン同梱の settings.json） | **対応済み** | `.claude-plugin/settings.json` でセキュリティルール・MCP 権限を即時適用。init トークン削減・インストール直後から保護が有効 |
| Worktree isolation（Task tool `isolation: "worktree"` パラメータ） | **対応済み** | `/breezing` の並列 Implementer に指定すると同一ファイル並列書き込みが安全化。`skills/breezing/references/guardrails-inheritance.md` 参照 |
| Background agents（エージェント定義の `background: true`） | **対応済み** | `agents/video-scene-generator.md` に `background: true` 追加。非同期シーン生成が可能に |
| ConfigChange hook（設定変更時に発火するライフサイクルフック） | **対応済み** | `hooks/hooks.json` に ConfigChange ハンドラ追加。設定変更を監査ログに記録 |
| WASM memory fix | 有利 | WebAssembly ベースのツール（Rust/Go MCP 等）のメモリ問題修正。安定性向上 |

### v2.1.47 (2026-02-19)

| 機能 | Harness 対応 | 備考 |
|------|-------------|------|
| `last_assistant_message` in Stop hook | **対応済み** | `scripts/stop-session-evaluator.sh` でセッション終了時の最終メッセージ品質評価に活用 |
| Agent model field fix（カスタムエージェントの `model` フィールド継承修正） | **対応済み** | `video-scene-generator.md` 等の `model: sonnet` 指定が Teammate spawn 時に確実に反映されるように |
| メモリトリミング（長期セッションのメモリ使用量最適化） | 有利 | `/breezing` 等の長時間チームセッションでのメモリ効率が改善 |
| plan mode compaction fix | 有利 | `/planning` スキル使用中のセッション圧縮の安定性向上 |
| 並列操作耐障害性向上 | 有利 | `/breezing` 並列 Implementer のネットワーク断等への耐性強化 |

### v2.1.46 (2026-02-18)

| 機能 | Harness 対応 | 備考 |
|------|-------------|------|
| claude.ai MCP connectors | 将来対応 | Web 版 Claude から MCP サーバーへの接続。Harness の MCP 統合に将来活用可能 |

### v2.1.45 (2026-02-17)

| 機能 | Harness 対応 | 備考 |
|------|-------------|------|
| **Sonnet 4.6**（最新モデル追加） | **重要** | `/breezing` の Implementer/Reviewer に Sonnet 4.6 (1M context) を利用可能。大規模コンテキスト処理に有利 |
| enabledPlugins from `--add-dir`（追加ディレクトリのプラグイン有効化） | 有利 | モノレポ構成でのプラグイン読み込みが改善 |
| Agent Teams on Bedrock/Vertex fix（追加修正） | **重要** | v2.1.41 に続く追加修正。Bedrock/Vertex/Foundry での `/breezing` 動作がさらに安定 |

### v2.1.44 (2026-02-16)

| 機能 | Harness 対応 | 備考 |
|------|-------------|------|
| ENAMETOOLONG fix（長いファイルパスのエラー修正） | 有利 | 深いネスト構造の Harness スキルディレクトリでのエラーを防止 |

### v2.1.43 (2026-02-15)

| 機能 | Harness 対応 | 備考 |
|------|-------------|------|
| AWS auth timeout fix | 有利 | Bedrock 環境での認証タイムアウト修正。troubleshoot スキルの診断が改善 |

### v2.1.42 (2026-02-14)

| 機能 | Harness 対応 | 備考 |
|------|-------------|------|
| Zod スキーマ遅延構築による起動性能改善 | 有利 | CC の起動が高速化。Harness 側の対応不要 |
| プロンプトキャッシュヒット率改善（日付をシステムプロンプト外に移動） | 有利 | API コスト削減・レイテンシ改善。対応不要 |
| Opus 4.6 effort callout（対象ユーザーへの1回限り通知） | - | 情報通知のみ。Harness 影響なし |
| `/resume` 中断メッセージのセッションタイトル表示修正 | 有利 | session-control スキルの UX 向上。セッション一覧がよりわかりやすく |
| 画像サイズ制限エラーで `/compact` を提案 | 有利 | harness-review 等で大きな画像を扱う際のエラーメッセージが改善 |

### v2.1.41 (2026-02-13)

| 機能 | Harness 対応 | 備考 |
|------|-------------|------|
| CC 内での CC 二重起動ガード | 有利 | `codex exec` 等での事故的な CC 再起動を防止 |
| Agent Teams: Bedrock/Vertex/Foundry モデル ID 修正 | **重要** | `/breezing` が Bedrock/Vertex/Foundry 環境で正常動作するように。v2.1.41+ 推奨の主因 |
| MCP ツールのストリーミング中画像コンテンツでクラッシュ修正 | 有利 | agent-browser (Chrome DevTools MCP) 等の安定性向上 |
| `/resume` プレビューの raw XML タグ表示修正 | 有利 | セッション再開時にスキル名が正しく表示される |
| Bedrock/Vertex/Foundry のエラーメッセージ改善 | 有利 | troubleshoot スキルでの診断が容易に（フォールバック提案付き） |
| プラグイン browse の "Space to Toggle" ヒント修正 | - | UI レベルの修正。Harness 影響なし |
| Hook blocking エラー (exit code 2) の stderr 表示修正 | **重要** | pretooluse-guard.sh 等のブロック理由がユーザーに正しく表示されるように |
| Hook blocking stderr の UI 表示修正（重複修正） | **重要** | 上記と合わせて Hook エラーの可視性が大幅改善 |
| OTel イベント/スパンに `speed` 属性追加 | 将来対応 | AgentTrace と連携して fast mode の可視化に活用可能 |
| `claude auth login/status/logout` サブコマンド追加 | 対応済み | troubleshoot スキルの診断テーブルに `claude auth status` を追加 |
| Windows ARM64 (win32-arm64) ネイティブバイナリ対応 | - | プラットフォームサポート拡大。Harness 影響なし |
| `/rename` が引数なしでセッション名自動生成 | 有利 | session スキルで活用可能。コンテキストから自動命名 |
| 狭いターミナルのプロンプトフッター改善 | - | UI レイアウト修正。Harness 影響なし |
| @-mention のアンカーフラグメント修正（`@README.md#installation`） | 有利 | スキルドキュメントでファイル内特定セクション参照が可能に |
| FileReadTool の FIFO/dev/stdin/大ファイルブロック修正 | 有利 | ファイル読み取りのハング防止。安定性向上 |
| ストリーミング Agent SDK のバックグラウンドタスク通知修正 | 有利 | `/breezing` の Teammate 完了通知が確実に届くように |
| classifier ルール入力のカーソルジャンプ修正 | - | UI レベルの修正。Harness 影響なし |
| markdown リンク表示テキストが raw URL で消える修正 | 有利 | スキル出力のリンク可読性向上 |
| auto-compact 失敗エラー通知の非表示化 | 有利 | 不要なエラー通知が抑制され UX 改善 |
| 権限待ち時間がサブエージェント経過時間に含まれる問題修正 | 有利 | AgentTrace メトリクスの精度向上（Task tool のduration が正確に） |
| plan mode 中の proactive ticks 発火修正 | 有利 | `/planning` スキル使用中の安定性向上 |
| 設定変更時の古い権限ルールクリア | 有利 | ディスク上の settings.json 変更がリアルタイム反映 |

### v2.1.39 (2026-02-11)

| 機能 | Harness 対応 | 備考 |
|------|-------------|------|
| ターミナルレンダリング性能改善 | 有利 | 大量出力時の描画が高速化 |
| 致命的エラーが飲み込まれる問題の修正 | 有利 | エラー診断がより確実に |
| セッションクローズ後のプロセスハング修正 | 有利 | セッション終了時の安定性向上 |
| ターミナル画面境界での文字ロス修正 | 有利 | 日本語表示の安定性向上 |
| verbose transcript view の空行修正 | - | デバッグ表示の改善。Harness 影響なし |

### v2.1.22 (2026-01-28)

| 機能 | Harness 対応 | 備考 |
|------|-------------|------|
| 非対話モード (`-p`) の structured outputs 修正 | - | Harness 影響なし |

### v2.1.21 (2026-01-28)

| 機能 | Harness 対応 | 備考 |
|------|-------------|------|
| ファイル操作ツール優先（Read/Edit/Write > cat/sed/awk） | 有利 | PostToolUse `Write\|Edit` フックの発火頻度が増加 = 品質ガード範囲拡大。`Bash(cat:*)` 権限の発火頻度は低下するが維持 |
| 全角数字入力対応（日本語 IME） | 有利 | 日本語ユーザーの選択肢入力が改善 |
| セッション中断後の再開時 API エラー修正 | 有利 | session-resume.sh の安定性向上 |
| auto-compact の早期発火修正 | 有利 | 大出力トークンモデルでのコンテキスト保持が改善 |
| Task ID の再利用問題修正 | - | Harness は TodoWrite を使用、影響なし |
| シェル補完キャッシュ修正 | - | Harness 影響なし |
| 読み取り/検索プログレスインジケーター改善 | - | UX 改善（Harness 影響なし） |
| [VSCode] Python venv 自動アクティベーション | - | VSCode 拡張機能のみ |
| [VSCode] ボタン背景色修正 | - | VSCode 拡張機能のみ |
| [VSCode] Windows ファイル検索修正 | - | VSCode 拡張機能のみ |

### v2.1.20 (2026-01-27)

| 機能 | Harness 対応 | 備考 |
|------|-------------|------|
| Background agent 起動前の権限プロンプト | 要注意 | `/work` 並列実行時に権限承認が必要。`permissions.allow` で事前承認推奨 |
| Setup hook `--init-only` フラグ | 対応済み | hooks.json に `init-only` マッチャー追加 |
| `Bash(*)` ワイルドカードが `Bash` と同等 | 互換 | harness-update の破壊的変更検知で除外対応 |
| PR レビューステータスインジケーター | - | プロンプトフッターに PR 状態表示（Harness 影響なし） |
| `CLAUDE.md` 追加ディレクトリ読み込み | 互換 | `--add-dir` + `CLAUDE_CODE_ADDITIONAL_DIRECTORIES_CLAUDE_MD=1` でモノレポ対応 |
| Task 削除（TaskUpdate ツール） | - | Harness は TodoWrite を使用、影響なし |
| Session compaction 修正 | 有利 | セッション resume の安定性向上 |
| Agent がユーザーメッセージを無視する問題修正 | 有利 | 並列 task-worker 実行中のユーザー介入が可能に |
| CJK/emoji レンダリング修正 | 有利 | 日本語表示の改善 |
| MCP Unicode JSON パース修正 | 有利 | Codex MCP 呼び出しの安定性向上 |
| Config バックアップのタイムスタンプ付きローテーション | 互換 | Claude Code 側で設定バックアップを5世代管理 |
| `/commit-push-pr` Slack 自動投稿 | - | MCP 経由で PR URL を Slack 投稿（Harness の auto-commit と補完関係） |

### v2.1.19 (2026-01-24)

| 機能 | Harness 対応 | 備考 |
|------|-------------|------|
| `CLAUDE_CODE_ENABLE_TASKS` env var | - | Harness 影響なし |
| `$ARGUMENTS[0]` 構文 | 互換 | Harness では未使用 |
| 権限/フックなしスキルは承認不要 | 有利 | Harness スキルの UX 向上 |
| バックグラウンドフック修正 | 有利 | Harness フックの安定性向上 |

### v2.1.18 (2026-01-23)

| 機能 | Harness 対応 | 備考 |
|------|-------------|------|
| `/keybindings` コマンド | - | Harness 影響なし（ターミナル機能） |

### v2.1.17 (2026-01-22)

| 機能 | Harness 対応 | 備考 |
|------|-------------|------|
| Task management system | 対応済み | TodoWrite ↔ Plans.md 同期 |

### v2.1.10 (2026-01-17)

| 機能 | Harness 対応 | 備考 |
|------|-------------|------|
| Setup hook event | 対応済み | `--init` / `--maintenance` フック |

### v2.1.9 (2026-01-16)

| 機能 | Harness 対応 | 備考 |
|------|-------------|------|
| PreToolUse additionalContext | 対応済み | 品質ガイドライン自動注入 |
| plansDirectory 設定 | 対応済み | Plans.md 配置カスタマイズ |
| ${CLAUDE_SESSION_ID} | 部分対応 | session-init.sh でマッピング |
| MCP auto:N syntax | 対応済み | [MCP_CONFIGURATION.md](./MCP_CONFIGURATION.md) 参照 |

### v2.1.7 (2026-01-14)

| 機能 | Harness 対応 | 備考 |
|------|-------------|------|
| MCP auto mode | 対応済み | ドキュメント簡略化 |
| showTurnDuration | - | Harness 影響なし |

### v2.1.6 (2026-01-13)

| 機能 | Harness 対応 | 備考 |
|------|-------------|------|
| Nested skills directory | 互換 | 将来的な構造変更で活用予定 |
| context_window percentage | 対応済み | harness-ui ダッシュボードで表示 |

### v2.1.3 (2026-01-09)

| 機能 | Harness 対応 | 備考 |
|------|-------------|------|
| Hook timeout 10分 | 対応済み | 重い処理のタイムアウト延長 |
| Commands/Skills 統合 | 互換 | 既存構造で対応 |

### v2.1.2 (2026-01-09)

| 機能 | Harness 対応 | 備考 |
|------|-------------|------|
| SessionStart agent_type | 対応済み | サブエージェント軽量初期化 |
| OSC 8 hyperlinks | - | ターミナル機能、Harness 影響なし |

### v2.0.74 (2025-12-19)

| 機能 | Harness 対応 | 備考 |
|------|-------------|------|
| LSP tool | 対応済み | impl/harness-review スキルで活用推奨 |

## 互換性チェック方法

```bash
# Claude Code バージョン確認
claude --version

# Harness バージョン確認
cat /path/to/harness/VERSION
```

## 非互換の可能性

### 破壊的変更はなし

現時点で Claude Code の変更による Harness の破壊的変更はありません。
ただし、以下の機能は新しいバージョンでのみ利用可能です:

- additionalContext（v2.1.9+）
- agent_type（v2.1.2+）
- LSP tool（v2.0.74+）
- Agent Teams Bedrock/Vertex/Foundry 正常動作（v2.1.41+）
- `CLAUDE_CODE_SIMPLE` モードでスキル・メモリ・エージェント無効化（v2.1.50+）
  - **影響**: 37 スキル・11 エージェント・プロジェクトメモリがすべて無効化
  - **動作するもの**: フック（安全ガード・セッション管理）のみ
  - **検出**: SessionStart / Setup フックで自動検出・警告表示（Harness v2.25.0+）
  - **詳細**: [docs/SIMPLE_MODE_COMPATIBILITY.md](./SIMPLE_MODE_COMPATIBILITY.md)

古いバージョンの Claude Code でも Harness は動作しますが、上記機能は無効化または制限されます。

## 更新履歴

- 2026-02-24: v2.1.50〜v2.1.51 対応追加（メモリリーク修正、WorktreeCreate/Remove hook、`claude agents` CLI、remote-control、ツール出力閾値変更、SIMPLE モード注意事項）。推奨バージョンを v2.1.51+ に引き上げ
- 2026-02-21: v2.1.43〜v2.1.49 対応追加（Plugin settings.json、Worktree isolation、Background agents、ConfigChange hook、last_assistant_message、Sonnet 4.6）。推奨バージョンを v2.1.49+ に引き上げ
- 2026-02-14: v2.1.39〜v2.1.42 対応追加（Agent Teams モデルID修正、Hook stderr修正、起動性能改善）。推奨バージョンを v2.1.41+ に引き上げ
- 2026-01-30: Harness v2.14.9 対応追加（4観点並列レビュー、auto-commit、OpenCode対応）
- 2026-01-28: v2.1.21〜v2.1.22 対応追加（ファイル操作ツール優先、全角数字入力、セッション再開修正）
- 2026-01-27: v2.1.20 対応追加（init-only フック、権限プロンプト対応、Bash(*) ワイルドカード）
- 2026-01-24: v2.1.18〜v2.1.19 対応追加
- 2026-01-16: 初版作成（v2.1.2〜v2.1.9 対応）
