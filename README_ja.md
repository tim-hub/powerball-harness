<p align="center">
  <img src="docs/images/claude-harness-logo-with-text.png" alt="Claude Harness" width="400">
</p>

<p align="center">
  <strong>Claude Code を、自己修正する開発チームに変える。</strong>
</p>

<p align="center">
  <a href="VERSION"><img src="https://img.shields.io/badge/version-2.14.10-blue.svg" alt="Version"></a>
  <a href="LICENSE.md"><img src="https://img.shields.io/badge/License-MIT-green.svg" alt="License"></a>
  <a href="docs/CLAUDE_CODE_COMPATIBILITY.md"><img src="https://img.shields.io/badge/Claude_Code-v2.1.21+-purple.svg" alt="Claude Code"></a>
</p>

<p align="center">
  <a href="README.md">English</a> | 日本語
</p>

---

## 問題

**Claude は優秀。でも、忘れる。迷走する。壊す。**

| 問題 | 症状 |
|------|------|
| **忘れる** | 前回の決定がセッション間で消える |
| **迷走する** | 計画なしでコーディング開始、方向性を見失う |
| **壊す** | テストをスキップ、lint を無視、バグを出荷 |
| **ごまかす** | 詰まると近道を取る—テスト改ざん、空の catch ブロック |

心当たりありませんか？

---

## 解決策

**Claude Harness は Claude Code をガードレールで包み、規律を強制します。**

```
Plan  →  Work  →  Review  →  Commit
```

覚えるべき3つのコアコマンド。

```bash
/plan-with-agent   # 壁打ち → 構造化された計画
/work              # 並列ワーカーで実行 + セルフレビュー
/harness-review    # 4観点の並列コードレビュー
```

**結果:** プロトタイプではなく、本番品質のコード。

---

## クイックスタート

### 10秒インストール

```bash
# 任意のプロジェクトディレクトリで
claude

# 以下を実行:
/plugin marketplace add Chachamaru127/claude-code-harness
/plugin install claude-code-harness@claude-code-harness-marketplace
/harness-init
```

**完了。** `/plan-with-agent` で始めましょう。

<details>
<summary>別の方法: ワンライナースクリプト</summary>

```bash
curl -fsSL https://raw.githubusercontent.com/Chachamaru127/claude-code-harness/main/scripts/quick-install.sh | bash
```

開発ツール付き (AST-Grep + LSP):
```bash
curl -fsSL https://raw.githubusercontent.com/Chachamaru127/claude-code-harness/main/scripts/quick-install.sh | bash -s -- --with-dev-tools
```

</details>

<details>
<summary>別の方法: ローカルクローン</summary>

```bash
git clone https://github.com/Chachamaru127/claude-code-harness.git ~/claude-plugins/claude-code-harness
cd /path/to/your-project
claude --plugin-dir ~/claude-plugins/claude-code-harness
```

</details>

---

## Before → After

| Before (素の Claude Code) | After (Harness 導入後) |
|---------------------------|------------------------|
| すぐにコーディング開始 | まず計画、それから実行 |
| 頼まないとレビューしない | 全ての変更を自動レビュー |
| 過去の決定を忘れる | SSOT ファイルがコンテキストを保持 |
| `rm -rf` が警告なしで実行 | 危険なコマンドは確認を要求 |
| 作業後に手動で `git commit` | レビュー通過で自動コミット |
| 一度に1タスク | 並列ワーカーで高速化 |

---

## 主な機能

### 🎯 Plan → Work → Review サイクル

すべてのアイデアが同じループを通過:

1. **Plan** — `/plan-with-agent` が曖昧なアイデアを `Plans.md` に変換
2. **Work** — `/work` が並列ワーカーでタスクを実行
3. **Review** — `/harness-review` が4観点で並列レビュー

### 🛡️ 安全フック

| 保護対象 | アクション |
|----------|------------|
| `.git/`, `.env`, 秘密鍵 | 書き込みブロック |
| `rm -rf`, `sudo`, `git push --force` | 確認を要求 |
| `git status`, `npm test` | 自動許可 |

### 🧠 永続的メモリ

- **SSOT ファイル**: `decisions.md`（なぜ）+ `patterns.md`（どうやって）
- **Claude-mem 統合**: 過去の学びがセッションを跨いで生き残る
- **セッション再開**: 中断したところから正確に再開

### ⚡ 並列実行

```bash
/work --parallel 5   # 5ワーカーを並列実行
```

各ワーカーが実装とセルフレビュー。全体レビュー通過後に自動コミット。

### 🔍 4観点の並列コードレビュー

```bash
/harness-review
```

セキュリティ、パフォーマンス、アクセシビリティ、品質—4観点が同時並列でレビュー。[Codex](https://github.com/openai/codex) を追加すれば16種のエキスパートから4つを選択してセカンドオピニオンを取得可能。

### 🔧 コードインテリジェンス

```bash
/dev-tools-setup   # 初回のみ
```

AST-Grep + LSP による構造的検索とセマンティック分析。

---

## 誰のためのツール？

| あなた | メリット |
|--------|----------|
| **個人開発者** | 品質を犠牲にせず速く出荷 |
| **フリーランス** | レビュー結果を品質証明として納品 |
| **VibeCoder** | 自然言語でアプリを構築 |
| **Cursor ユーザー** | 2-Agent ワークフロー: Cursor が計画、Claude Code が実装 |

---

## コマンド

### コアワークフロー

| コマンド | 用途 |
|----------|------|
| `/plan-with-agent` | アイデアを実行可能な計画に変換 |
| `/work` | Plans.md のタスクを実行 |
| `/harness-review` | 4観点の並列コードレビュー |
| `/sync-status` | 進捗確認、次のアクションを提案 |

### セットアップ & 運用

| コマンド | 用途 |
|----------|------|
| `/harness-init` | プロジェクト初期化 |
| `/harness-update` | プラグイン更新 |
| `/dev-tools-setup` | AST-Grep + LSP セットアップ |
| `/skill-list` | 全28スキルカテゴリを表示 |

### 2-Agent (Cursor)

| コマンド | 用途 |
|----------|------|
| `/handoff-to-cursor` | PM への完了報告 |

---

## スキル

リクエストに応じて自動起動:

| スキル | トリガー |
|--------|----------|
| `impl` | 「実装して」「機能追加」「作って」 |
| `review` | 「レビューして」「セキュリティチェック」 |
| `verify` | 「ビルドして」「エラー修正」 |
| `auth` | 「ログイン機能」「Stripe決済」 |
| `deploy` | 「デプロイ」「Vercel」 |
| `ui` | 「ヒーローセクション」「コンポーネント」 |

**28スキルカテゴリ。** `/skill-list` で全て確認。

---

## アーキテクチャ

```
claude-code-harness/
├── commands/     # 31 スラッシュコマンド
├── skills/       # 28 スキルカテゴリ
├── agents/       # 8 サブエージェント（並列ワーカー）
├── hooks/        # 安全性 & 自動化
├── scripts/      # ガードスクリプト
└── templates/    # 生成テンプレート
```

---

## ドキュメント

| ガイド | 説明 |
|--------|------|
| [変更履歴](CHANGELOG_ja.md) | 各バージョンの変更点 |
| [Claude Code 互換性](docs/CLAUDE_CODE_COMPATIBILITY.md) | バージョン要件 |
| [Cursor 統合](docs/CURSOR_INTEGRATION.md) | 2-Agent ワークフローのセットアップ |
| [OpenCode 互換性](docs/OPENCODE_COMPATIBILITY.md) | 他の LLM での使用 |

---

## 動作要件

- **Claude Code v2.1.21+** (推奨)
- 詳細は[互換性ガイド](docs/CLAUDE_CODE_COMPATIBILITY.md)を参照

---

## 謝辞

- **階層型スキル構造**: [AIまさお氏](https://note.com/masa_wunder)
- **テスト改ざん防止**: [びーぐる氏](https://github.com/beagleworks)（Claude Code Meetup Tokyo 2025.12.22）

---

## ライセンス

**MIT License** — 使用・改変・配布・商用利用が自由です。

[English](LICENSE.md) | [日本語](LICENSE.ja.md)
