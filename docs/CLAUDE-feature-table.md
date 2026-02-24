# Claude Code 2.1.51+ 新機能活用ガイド（完全版）

> **概要**: Harness が活用する Claude Code 2.1.51+ の全機能一覧。
> CLAUDE.md の Feature Table の完全版（詳細説明付き）。

## 機能一覧

| 機能 | 活用スキル | 用途 |
|------|-----------|------|
| **Task tool メトリクス** | parallel-workflows | サブエージェントのトークン/ツール/時間を集計 |
| **`/debug` コマンド** | troubleshoot | 複雑なセッション問題の診断 |
| **PDF ページ範囲** | notebookLM, harness-review | 大型ドキュメントの効率的な処理 |
| **Git log フラグ** | harness-review, CI, release-harness | 構造化されたコミット分析 |
| **OAuth 認証** | codex-review | DCR 非対応 MCP サーバーの設定 |
| **68% メモリ最適化** | session-memory, session | `--resume` の積極的活用 |
| **サブエージェント MCP** | task-worker | 並列実行時の MCP ツール共有 |
| **Reduced Motion** | harness-ui | アクセシビリティ設定 |
| **TeammateIdle/TaskCompleted Hook** | breezing | チーム監視の自動化 |
| **Agent Memory (memory frontmatter)** | task-worker, code-reviewer | 永続的学習 |
| **Fast mode (Opus 4.6)** | 全スキル | 高速出力モード |
| **自動メモリ記録** | session-memory | セッション間知識の自動永続化 |
| **スキルバジェットスケーリング** | 全スキル | コンテキスト窓の 2% に自動調整 |
| **Task(agent_type) 制限** | agents/ | サブエージェント種類制限 |
| **Plugin settings.json** | setup | init トークン削減・即時セキュリティ保護 |
| **Worktree isolation** | breezing, parallel-workflows | 同一ファイル並列書き込み安全化 |
| **Background agents** | generate-video | 非同期シーン生成 |
| **ConfigChange hook** | hooks | 設定変更監査 |
| **last_assistant_message** | session-memory | セッション品質評価 |
| **Sonnet 4.6 (1M context)** | 全スキル | 大規模コンテキスト処理 |
| **メモリリーク修正 (v2.1.50)** | breezing, work | 長時間チームセッションの安定性向上 |
| **`claude agents` CLI (v2.1.50)** | troubleshoot | エージェント定義の診断・確認 |
| **WorktreeCreate/Remove hook (v2.1.50)** | breezing | Worktree ライフサイクル管理（将来対応） |
| **`claude remote-control` (v2.1.51)** | 将来対応 | 外部ビルドとローカル環境サービング |

## 機能詳細

### Task tool メトリクス

サブエージェントが消費したトークン数・ツール呼び出し数・実行時間を集計できる。
`parallel-workflows` スキルでは複数サブエージェントのメトリクスを集約し、コスト分析に使用。

```
metrics: {tokens: 40000, tools: 7, duration: 67s}
```

### `/debug` コマンド

セッション診断用コマンド。複雑なエラーや予期しない挙動の原因調査に使用。
`troubleshoot` スキルが自動的に起動し、問題を体系的に診断。

### PDF ページ範囲指定

大型 PDF を読み込む際にページ範囲を指定可能（例: `pages: "1-5"`）。
`notebookLM` スキルでのドキュメント処理、`harness-review` での大型仕様書参照に活用。

### Git log フラグ

`git log` の構造化オプション（`--format`, `--stat`, `--since` 等）を活用。
リリースノート生成、コミット分析、変更追跡を効率化。

### OAuth 認証

DCR（Dynamic Client Registration）非対応 MCP サーバーへの OAuth 認証設定。
`codex-review` スキルでの Codex CLI 接続に使用。

### 68% メモリ最適化

`--resume` フラグによるセッション再開時のメモリ使用量削減。
長時間作業セッションでのコンテキスト継続に有効。

### サブエージェント MCP

Task tool で起動したサブエージェントが親セッションの MCP ツールを共有できる。
`task-worker` での並列実装時に、各エージェントが同じ MCP ツールセットを使用可能。

### Reduced Motion

アクセシビリティ設定。モーション/アニメーションを削減するオプション。
`harness-ui` スキルで UI 生成時に考慮。

### TeammateIdle/TaskCompleted Hook

Breezing チームのメンバーがアイドル状態になった時、またはタスク完了時に発火するフック。
`scripts/hook-handlers/teammate-idle.sh` と `task-completed.sh` で処理。

```json
"TeammateIdle": [{"hooks": [{"type": "command", "command": "...teammate-idle", "timeout": 10}]}],
"TaskCompleted": [{"hooks": [{"type": "command", "command": "...task-completed", "timeout": 10}]}]
```

### Agent Memory (memory frontmatter)

エージェント定義 YAML の `memory: project` フィールドで永続メモリを有効化。
`task-worker`, `code-reviewer` が過去の実装パターン・失敗と解決策を跨ぎセッションで学習。

### Fast mode (Opus 4.6)

`/fast` コマンドで切り替える高速出力モード。同じ Opus 4.6 モデルを使用。
全スキルで利用可能。長い実装タスクでの待ち時間短縮に有効。

### 自動メモリ記録

セッション終了時に学習内容を自動的にメモリファイルへ永続化。
`session-memory` スキルが管理。次のセッションで前回の文脈を自動復元。

### スキルバジェットスケーリング

SKILL.md の文字数予算がコンテキスト窓の 2% に自動調整される。
推奨 500 行は目安値。実効上限はモデルのコンテキスト窓サイズに依存。

### Task(agent_type) 制限

Task tool 呼び出し時に `subagent_type` を指定し、サブエージェントの種類を制限。
`agents/` 定義と組み合わせて、意図したエージェントのみを起動することを保証。

### Plugin settings.json

プラグインの `settings.json` で初期化時の設定を事前定義。
init トークン消費を削減し、セキュリティポリシーをセッション開始直後から適用。

### Worktree isolation

`git worktree` を使って同一ファイルへの並列書き込みを安全化。
`breezing` と `parallel-workflows` での複数エージェント並列実装時のコンフリクト防止。

### Background agents

非同期でバックグラウンドエージェントを起動。完了を待たずに他の処理を継続可能。
`generate-video` スキルでの複数シーン並列生成に使用。

### ConfigChange hook

設定ファイル（`settings.json` 等）が変更された時に発火するフック。
`scripts/hook-handlers/config-change.sh` で変更を記録・監査。

### last_assistant_message

セッション終了時の最後のアシスタントメッセージを参照できる機能。
`session-memory` スキルがセッション品質の自己評価に使用。

### Sonnet 4.6 (1M context)

最大 1M トークンのコンテキスト窓を持つ Sonnet 4.6 モデル。
大規模コードベースの分析、長大なドキュメント処理に対応。全スキルで利用可能。

### メモリリーク修正 (v2.1.50)

CC 2.1.50 で LSP 診断データ、大型ツール出力、ファイル履歴、シェル実行に関するメモリリークが修正された。
完了タスクのガベージコレクションも実装され、`/breezing` 等の長時間チームセッションの安定性が大幅に改善。
Harness 側は JSONL ローテーション（500→400 行）やアトミック更新で既に独自対策を実施済み。

### `claude agents` CLI (v2.1.50)

`claude agents list` で登録済みエージェントの一覧を表示。
`troubleshoot` スキルでエージェント spawn 失敗時の診断に活用。

```bash
claude agents list   # 登録済みエージェントの一覧
```

### WorktreeCreate/WorktreeRemove hook (v2.1.50)

Worktree の作成・削除時に発火する新しいライフサイクルフック。
`/breezing` 並列ワークフローでの自動セットアップ・クリーンアップに将来活用可能。
現状は Harness 未実装。`skills/breezing/references/guardrails-inheritance.md` に記載。

### `claude remote-control` (v2.1.51)

外部ビルドシステムとローカル環境のサービングを可能にするサブコマンド。
将来的に Breezing のクロスセッション制御や CI 連携に活用の余地あり。

## 関連ドキュメント

- [CLAUDE.md](../CLAUDE.md) - 開発ガイド（Feature Table の要約版）
- [CLAUDE-skill-catalog.md](./CLAUDE-skill-catalog.md) - スキルカタログ
- [CLAUDE-commands.md](./CLAUDE-commands.md) - コマンドリファレンス
- [ARCHITECTURE.md](./ARCHITECTURE.md) - アーキテクチャ概要
