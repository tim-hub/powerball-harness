<!-- Generated from CLAUDE.md by build-opencode.js -->
<!-- opencode.ai compatible version of Claude Code Harness -->

# AGENTS.md - Claude harness 開発ガイド

このファイルは Claude Code がこのリポジトリで作業する際の指針です。

## プロジェクト概要

**Claude harness** は、Claude Code を「Plan → Work → Review」の型で自律運用するためのプラグインです。

**特殊な点**: このプロジェクトは「ハーネス自身を使ってハーネスを改善する」自己参照的な構成です。

## Claude Code 2.1.49+ 新機能活用ガイド

Harness は Claude Code 2.1.49 の新機能をフル活用しています。

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

詳細は各スキルの SKILL.md を参照してください。

## 開発ルール

### コミットメッセージ

[Conventional Commits](https://www.conventionalcommits.org/) に従う:

- `feat:` - 新機能
- `fix:` - バグ修正
- `docs:` - ドキュメント変更
- `refactor:` - リファクタリング
- `test:` - テスト追加/更新
- `chore:` - メンテナンス

### バージョン管理

バージョンは 2 箇所で定義（同期必須）:
- `VERSION` - ソース・オブ・トゥルース
- `.claude-plugin/plugin.json` - プラグインシステム用

変更時は `./scripts/sync-version.sh bump` を使用。

### CHANGELOG 記載ルール

詳細: [.claude/rules/changelog.md](.claude/rules/changelog.md)

- [Keep a Changelog](https://keepachangelog.com/ja/1.0.0/) フォーマットに準拠
- セクション: Added / Changed / Deprecated / Removed / Fixed / Security
- 大きな変更時は Before/After テーブルを追加

### 言語設定

すべての応答は **日本語** で行うこと（`context: fork` スキル含む）。

### コードスタイル

- 明確で説明的な名前を使う
- 複雑なロジックにはコメントを追加
- コマンド/エージェント/スキルは単一責任に保つ

## リポジトリ構成

```
claude-code-harness/
├── .claude-plugin/     # プラグインマニフェスト
├── commands/           # スラッシュコマンド（ユーザー向け）
├── agents/             # サブエージェント定義（Task tool で並列起動可能）
├── skills/             # エージェントスキル
├── hooks/              # ライフサイクルフック
├── scripts/            # シェルスクリプト（ガード、自動化）
├── templates/          # テンプレートファイル
├── docs/               # ドキュメント
└── tests/              # 検証スクリプト
```

## スキルの活用（重要）

### スキル評価フロー

> 💡 重いタスク（並列レビュー、CI修正ループ）では、スキルが `agents/` のサブエージェントを Task tool で並列起動します。

**作業を開始する前に、必ず以下のフローを実行すること:**

1. **評価**: 利用可能なスキルを確認し、今回の依頼に該当するものがあるか評価
2. **起動**: 該当するスキルがあれば、Skill ツールで起動してから作業開始
3. **実行**: スキルの手順に従って作業を進める

```
ユーザーの依頼
    ↓
スキルを評価（該当するものがあるか？）
    ↓
YES → Skill ツールで起動 → スキルの手順に従う
NO  → 通常の推論で対応
```

### スキルの階層構造

スキルは **親スキル（カテゴリ）** と **子スキル（具体的な機能）** の階層構造になっています。

```
skills/
├── impl/                  # 実装（機能追加、テスト作成）
├── harness-review/        # レビュー（品質、セキュリティ、パフォーマンス）
├── verify/                # 検証（ビルド、エラー復旧、修正適用）
├── setup/                 # 統合セットアップ（プロジェクト初期化、ツール設定、2-Agent、harness-mem、Codex CLI、ルールローカライズ）
├── memory/                # メモリ管理（SSOT、decisions.md、patterns.md、SSOT昇格、記憶検索）
├── troubleshoot/          # 診断・修復（エラー、CI障害含む）
├── principles/            # 原則・ガイドライン（VibeCoder、差分編集）
├── auth/                  # 認証・決済（Clerk、Supabase、Stripe）
├── deploy/                # デプロイ（Vercel、Netlify、アナリティクス）
├── ui/                    # UI（コンポーネント、フィードバック）
├── handoff/               # ワークフロー（ハンドオフ、自動修正）
├── notebookLM/            # ドキュメント（NotebookLM、YAML）
└── maintenance/           # メンテナンス（クリーンアップ）
```

**使い方:**
1. 親スキルを Skill ツールで起動
2. 親スキルがユーザーの意図に応じて適切な子スキル（doc.md）にルーティング
3. 子スキルの手順に従って作業実行

### 開発用スキル（非公開）

以下のスキルは開発・実験用であり、リポジトリには含まれません（.gitignore で除外）：

```
skills/
├── test-*/      # テスト用スキル
└── x-promo/     # X投稿作成スキル（開発用）
```

これらのスキルは個別の開発環境でのみ使用し、プラグイン配布には含めないこと。

### 主要スキルカテゴリ

| カテゴリ | 用途 | トリガー例 |
|---------|------|-----------|
| work | タスク実装（スコープ自動判断、--codex 対応） | 「実装して」「全部やって」「/work」 |
| breezing | Agent Teams で完全自動完走（--codex 対応） | 「チームで完走」「breezing」 |
| impl | 実装、機能追加、テスト作成 | 「実装して」「機能追加」「コードを書いて」 |
| harness-review | コードレビュー、品質チェック | 「レビューして」「セキュリティ」「パフォーマンス」 |
| verify | ビルド検証、エラー復旧 | 「ビルド」「エラー復旧」「検証して」 |
| setup | セットアップ統合ハブ（プロジェクト初期化、ツール設定、2-Agent、harness-mem、Codex CLI、ルールローカライズ） | 「セットアップ」「CLAUDE.md」「初期化」「CI setup」「2-Agent」「Cursor設定」「harness-mem」「codex-setup」 |
| memory | SSOT管理、記憶検索、SSOT昇格、Cursor連携メモリ | 「SSOT」「decisions.md」「マージ」「SSOT昇格」「記憶検索」「claude-mem」 |
| principles | 開発原則、ガイドライン | 「原則」「VibeCoder」「安全性」 |
| auth | 認証、決済機能 | 「ログイン」「Clerk」「Stripe」「決済」 |
| deploy | デプロイ、アナリティクス | 「デプロイ」「Vercel」「GA」 |
| ui | UIコンポーネント生成 | 「コンポーネント」「ヒーロー」「フォーム」 |
| handoff | ハンドオフ、自動修正 | 「ハンドオフ」「PMに報告」「自動修正」 |
| notebookLM | ドキュメント生成 | 「ドキュメント」「NotebookLM」「スライド」 |
| troubleshoot | 診断と修復（CI障害含む） | 「動かない」「エラー」「CIが落ちた」 |
| maintenance | ファイル整理 | 「整理して」「クリーンアップ」 |

## 開発フロー

1. **計画**: `/plan-with-agent` でタスクを Plans.md に落とす
2. **実装**: `/work` (Claude が実装) or `/breezing` (チームで完走)。両方 `--codex` 対応
3. **レビュー**: 自動実行（手動は `/harness-review`）
4. **検証**: `./tests/validate-plugin.sh` で構造検証

## テスト方法

```bash
# プラグイン構造の検証
./tests/validate-plugin.sh
./scripts/ci/check-consistency.sh

# 別プロジェクトでローカルテスト
cd /path/to/test-project
claude --plugin-dir /path/to/claude-code-harness
```

## 注意事項

- **自己参照に注意**: このプラグインの `/work` を実行すると、自分自身のコードを編集することになる
- **Hooks は自動実行**: PreToolUse/PostToolUse で自動ガードが働く
- **VERSION 同期**: コード変更時は必ずバージョンを更新

## 主要コマンド（開発時に使用）

| コマンド | 用途 |
|---------|------|
| `/plan-with-agent` | 改善タスクを Plans.md に追加 |
| `/work` | タスクを実装（スコープ自動判断、--codex 対応） |
| `/breezing` | Agent Teams でチーム並列完走（--codex 対応） |
| `/harness-review` | 変更内容をレビュー |
| `/validate` | プラグイン検証 |
| `/remember` | 学習事項を記録 |

### ハンドオフ

| コマンド | 用途 |
|---------|------|
| `/handoff-to-cursor` | Cursor 運用時の完了報告 |

**スキル（会話で自動起動）**:
- `handoff-to-impl` - 「実装役に渡して」→ PM → Impl への依頼
- `handoff-to-pm` - 「PMに完了報告」→ Impl → PM への完了報告

## SSOT（Single Source of Truth）

- `.claude/memory/decisions.md` - 決定事項（Why）
- `.claude/memory/patterns.md` - 再利用パターン（How）

## テスト改ざん防止（品質保証）

詳細: [D9: テスト改ざん防止の3層防御戦略](.claude/memory/decisions.md#d9-テスト改ざん防止の3層防御戦略)

| ルールファイル | 内容 |
|---------------|------|
| [test-quality.md](.claude/rules/test-quality.md) | テスト改ざん禁止パターン |
| [implementation-quality.md](.claude/rules/implementation-quality.md) | 形骸化実装禁止パターン |

> ⚠️ **絶対禁止**: テストを改ざんして「成功」を偽装すること

