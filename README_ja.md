<p align="center">
  <img src="docs/images/claude-harness-logo-with-text.png" alt="Claude Harness" width="400">
</p>

<p align="center">
  <strong>Plan. Work. Review. Ship.</strong><br>
  <em>Claude Code を規律ある開発パートナーに変える</em>
</p>

<p align="center">
  <a href="VERSION"><img src="https://img.shields.io/badge/version-2.16.12-blue.svg" alt="Version"></a>
  <a href="LICENSE.md"><img src="https://img.shields.io/badge/License-MIT-green.svg" alt="License"></a>
  <a href="docs/CLAUDE_CODE_COMPATIBILITY.md"><img src="https://img.shields.io/badge/Claude_Code-v2.1+-purple.svg" alt="Claude Code"></a>
  <img src="https://img.shields.io/badge/Skills-42-orange.svg" alt="Skills">
</p>

<p align="center">
  <a href="README.md">English</a> | 日本語
</p>

---

## なぜ Harness？

Claude Code は強力だが、時に構造が必要になる。

```mermaid
graph LR
    A[アイデア] --> B["/plan-with-agent"]
    B --> C["Plans.md"]
    C --> D["/work"]
    D --> E["コード + セルフレビュー"]
    E --> F["/harness-review"]
    F --> G["リリース"]
```

**3つのコマンド。1つのワークフロー。本番品質のコード。**

> **VibeCoder向け**: 「メールバリデーション付きのログインフォームが欲しい」と言うだけで、Harness が計画・実装・レビューを自動で行います。

---

## 動作要件

インストール前に以下を確認:

- **Claude Code v2.1+** ([インストールガイド](https://docs.anthropic.com/claude-code))
- **Node.js 18+** (セーフティフック用)

---

## 30秒でインストール

```bash
# プロジェクトで Claude Code を起動
claude

# マーケットプレイスを追加してインストール
/plugin marketplace add Chachamaru127/claude-code-harness
/plugin install claude-code-harness@claude-code-harness-marketplace

# プロジェクトを初期化
/harness-init
```

これだけ。`/plan-with-agent` から始めよう。

---

## コアループ

### 1. Plan（計画）

アイデアを構造化されたタスクに変換。

```bash
/plan-with-agent
```

> 「メールバリデーション付きのログインフォームが欲しい」

Harness が明確な受入条件付きの `Plans.md` を作成。

### 2. Work（実装）

並列ワーカーでタスクを実行。

```bash
/work              # 並列数を自動検出
/work --parallel 5 # 5ワーカーで同時実行
```

各ワーカーが実装、セルフレビュー、報告を行う。

### 3. Review（レビュー）

4つの視点で並列コードレビュー。

```bash
/harness-review
```

| 視点 | 焦点 |
|------|------|
| Security | 脆弱性、インジェクション、認証 |
| Performance | ボトルネック、メモリ、スケーリング |
| Quality | パターン、命名、保守性 |
| Accessibility | WCAG準拠、スクリーンリーダー、UX |

---

## 何が変わるか

| Harness なし | Harness あり |
|--------------|--------------|
| すぐにコードを書き始める | まず計画、それから実行 |
| 頼まれたときだけレビュー | 全ての変更を自動レビュー |
| 過去の決定を忘れる | SSOT ファイルでコンテキストを保持 |
| `rm -rf` が警告なく実行される | 危険なコマンドをブロック |
| 手動で git 操作 | 承認時に自動コミット |
| 一度に1タスク | 並列ワーカー |

> **SSOT** (Single Source of Truth): 決定事項やパターンをセッション横断で保存するファイル。

---

## セーフティファースト

Harness はフック（自動安全チェック）でコードベースを保護:

| 保護対象 | アクション |
|----------|------------|
| `.git/`, `.env`, シークレット | 書き込みブロック |
| `rm -rf`, `sudo`, `--force` | 確認が必要 |
| `git status`, `npm test` | 自動許可 |
| テスト改ざん | 警告をトリガー |

---

## 42スキル、設定不要

スキルはコンテキストに応じて自動ロード。スラッシュコマンドでも自然言語でも起動可能。

| こう言うと | このスキルが起動 |
|------------|------------------|
| 「ログインを実装して」 | `impl` |
| 「このコードをレビューして」 | `harness-review` |
| 「ビルドエラーを直して」 | `verify` |
| 「Stripe決済を追加して」 | `auth` |
| 「Vercelにデプロイして」 | `deploy` |
| 「ヒーローセクションを作って」 | `ui` |

> **Note**: すべてのスキルは `/スキル名` コマンドまたは自然言語で起動できます。

---

## 誰のためのツール？

| あなたが | Harness でできること |
|----------|---------------------|
| **開発者** | 組み込み QA で高速に出荷 |
| **フリーランサー** | クライアントにレビューレポートを納品 |
| **インディーハッカー** | 壊さずに素早く動く |
| **VibeCoder** | 自然言語でアプリを構築 |
| **チームリード** | プロジェクト横断で標準を強制 |

---

## コマンド一覧

### コアワークフロー

| コマンド | 機能 |
|----------|------|
| `/plan-with-agent` | アイデア → `Plans.md` |
| `/work` | タスクを並列実行 |
| `/harness-review` | 4視点レビュー |

### オペレーション

| コマンド | 機能 |
|----------|------|
| `/harness-init` | プロジェクト初期化 |
| `/harness-update` | プラグイン更新 |
| `/sync-status` | 進捗確認 |
| `/maintenance` | 古いタスクを整理 |

### メモリ

| コマンド | 機能 |
|----------|------|
| `/sync-ssot-from-memory` | 決定事項を SSOT に昇格 |
| `/memory` | SSOT ファイルを管理 |

> **仕組み**: v2.16 でコマンドからスキルに移行。`/コマンド` でも自然言語でも起動可能—同じ機能、賢いローディング。

---

## アーキテクチャ

```
claude-code-harness/
├── skills/       # 42のスキル定義
├── agents/       # 8つのサブエージェント（並列ワーカー）
├── hooks/        # セーフティ & オートメーション
├── scripts/      # ガードスクリプト
└── templates/    # 生成テンプレート
```

---

## 高度な機能

<details>
<summary><strong>並列実行</strong></summary>

```bash
/work --parallel 5
```

各ワーカーは独立して動作：
1. 割り当てられたタスクを実装
2. セルフレビューを実行
3. 完了を報告

全ワーカー完了後にグローバルレビューを実行。

</details>

<details>
<summary><strong>2-Agent モード（Cursor連携）</strong></summary>

Cursor を PM として、Claude Code を実装者として使用。

```bash
/handoff       # Cursor PM に報告
```

Plans.md が両者間で同期。

</details>

<details>
<summary><strong>Codex 連携</strong></summary>

OpenAI Codex でセカンドオピニオンを追加：

```bash
/harness-review  # 4視点 + Codex
```

Codex が16種のスペシャリストから4人の関連エキスパートを選出。

> **セットアップが必要**: [Codex CLI](https://github.com/openai/codex) をインストールし、APIキーを設定してください。これはオプション機能です。

</details>

<details>
<summary><strong>動画生成</strong></summary>

Remotion でプロダクト動画を生成：

```bash
/generate-video
```

AI生成のシーン、ナレーション、エフェクト。

> **依存関係**: [Remotion](https://www.remotion.dev/) プロジェクトのセットアップと ffmpeg が必要です。上級者向けのオプション機能です。

</details>

---

## トラブルシューティング

| 問題 | 解決策 |
|------|--------|
| コマンドが見つからない | まず `/harness-init` を実行 |
| プラグインが読み込まれない | キャッシュをクリア: `rm -rf ~/.claude/plugins/cache/claude-code-harness-marketplace/` して再起動 |
| フックが動作しない | Node.js 18+ がインストールされているか確認 |

詳しいヘルプは [Issue を作成](https://github.com/Chachamaru127/claude-code-harness/issues)してください。

---

## アンインストール

```bash
/plugin uninstall claude-code-harness
```

プラグインを削除します。プロジェクトファイル（Plans.md、SSOT ファイル）はそのまま残ります。

---

## ドキュメント

| リソース | 説明 |
|----------|------|
| [Changelog](CHANGELOG.md) | バージョン履歴 |
| [Claude Code 互換性](docs/CLAUDE_CODE_COMPATIBILITY.md) | 動作要件 |
| [Cursor 連携](docs/CURSOR_INTEGRATION.md) | 2-Agent セットアップ |

---

## コントリビュート

Issue と PR を歓迎します。[CONTRIBUTING.md](CONTRIBUTING.md) を参照。

---

## 謝辞

- [AI Masao](https://note.com/masa_wunder) — 階層的スキル設計
- [Beagle](https://github.com/beagleworks) — テスト改ざん防止パターン

---

## ライセンス

**MIT License** — 自由に使用、改変、商用利用可能。

[English](LICENSE.md) | [日本語](LICENSE.ja.md)
