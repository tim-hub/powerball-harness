# Claude Harness

[English](README.md) | 日本語

![Claude Harness](docs/images/claude-harness-logo-with-text.png)

**Claude Code を「自己修正する開発チーム」に変える**

Claude Harness は Claude Code を **Plan → Work → Review** の自律サイクルで運用し、
ミスを出荷前にキャッチします。

[![Version](https://img.shields.io/badge/version-2.9.20-blue.svg)](CHANGELOG_ja.md)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE.ja.md)
[![Claude Code](https://img.shields.io/badge/Claude_Code-v2.1.6+-purple.svg)](docs/CLAUDE_CODE_COMPATIBILITY.md)

---

## 動作を見る

```bash
/plan-with-agent   # ブレスト → 計画を作成
/work              # 計画を実行（並列ワーカー対応）
/harness-review    # 多角的コードレビュー
```

**以上です。** 3 つのコマンドで、ラフなアイデアがレビュー済みの本番コードに変わります。

---

## なぜ Claude Harness？

ソロ開発者が直面する 4 つの問題を、すべて解決します：

| 問題 | 症状 | Harness の解決策 |
|------|------|-----------------|
| **混乱** | 「どこから始めれば？」 | `/plan-with-agent` でアイデアをタスクに分解 |
| **雑さ** | プレッシャーで品質が落ちる | `/harness-review` で 8 人のエキスパートが並列レビュー |
| **事故** | 危険なコマンドが通ってしまう | Hooks が `rm -rf` をブロック、`.env` を保護、秘密鍵をガード |
| **忘却** | 過去の決定がセッション間で失われる | SSOT ファイル + Claude-mem でコンテキストを永続化 |

---

## クイックスタート

**要件**: Claude Code v2.1.6+（[互換性ガイド](docs/CLAUDE_CODE_COMPATIBILITY.md)）

```bash
# 1. Claude Code でプロジェクトを開く
cd /path/to/your-project && claude

# 2. プラグインをインストール
/plugin marketplace add Chachamaru127/claude-code-harness
/plugin install claude-code-harness@claude-code-harness-marketplace

# 3. 初期化
/harness-init
```

**完了。** `/plan-with-agent` で最初の計画を作成しましょう。

<details>
<summary>代替手段: ローカルクローン</summary>

```bash
git clone https://github.com/Chachamaru127/claude-code-harness.git ~/claude-plugins/claude-code-harness
cd /path/to/your-project
claude --plugin-dir ~/claude-plugins/claude-code-harness
```

</details>

---

## 主な機能

### 並列フルサイクル自動化

```bash
/work --full --parallel 3
```

各タスクで **実装 → セルフレビュー → 修正 → コミット** を並列実行。
各ワーカーは完了前に自分のコードをレビューします。

### 8 人のエキスパートレビュー

```bash
/harness-review
```

セキュリティ、パフォーマンス、アクセシビリティ、保守性—8 人のスペシャリストが同時にコードをレビュー。オプションで [Codex](https://github.com/openai/codex) のセカンドオピニオンも。

### 安全フック

| 保護対象 | アクション |
|----------|-----------|
| `.git/`, `.env`, 秘密鍵 | 書き込みブロック |
| `rm -rf`, `sudo`, `git push --force` | 確認が必要 |
| `git status`, `npm test` | 自動許可 |

### セッション継続性

- **SSOT ファイル**: `decisions.md`（なぜ） + `patterns.md`（どうやって）
- **Claude-mem 統合**: 過去の学びがセッションを超えて持続
- **セッション再開**: `/resume` で作業状態をそのまま復元

---

## 誰のためのツール？

| あなた | メリット |
|--------|---------|
| **ソロ開発者** | 品質を犠牲にせず、より速く出荷 |
| **フリーランサー** | レビューレポートを品質の証明として納品 |
| **VibeCoder** | 自然言語でアプリを構築 |
| **Cursor ユーザー** | 2-Agent ワークフローで計画と実装を分離 |

---

## コマンド

### コアワークフロー

| コマンド | 用途 |
|---------|------|
| `/plan-with-agent` | アイデアを計画に変換 |
| `/work` | Plans.md からタスクを実行 |
| `/harness-review` | マルチエキスパートレビュー |
| `/sync-status` | 進捗確認、次のアクションを提案 |

### 運用

| コマンド | 用途 |
|---------|------|
| `/harness-init` | プロジェクト初期化 |
| `/harness-update` | プラグインファイルを更新 |
| `/codex-review` | Codex のみのセカンドオピニオン |
| `/skill-list` | 全 67 スキルを表示 |

### 2-Agent ワークフロー（Cursor）

| コマンド | 用途 |
|---------|------|
| `/handoff-to-cursor` | PM に完了報告を送信 |
| `/plan-with-cc` | (Cursor) 計画し、Claude Code に渡す |
| `/review-cc-work` | (Cursor) 実装をレビュー |

---

## スキル

リクエストに応じて自動トリガー：

| スキル | トリガー |
|--------|---------|
| `impl` | 「実装して」「機能追加」「作って」 |
| `review` | 「レビュー」「セキュリティチェック」「監査」 |
| `verify` | 「ビルド」「エラー修正」「復旧」 |
| `auth` | 「ログイン」「Stripe」「決済」 |
| `deploy` | 「デプロイ」「Vercel」「本番」 |
| `ui` | 「ヒーローセクション」「コンポーネント」「フォーム」 |

**22 カテゴリ、67 スキル。** `/skill-list` で全て表示。

---

## アーキテクチャ

```
claude-code-harness/
├── commands/     # 21 スラッシュコマンド
├── skills/       # 67 スキル（22 カテゴリ）
├── agents/       # 6 サブエージェント（並列ワーカー）
├── hooks/        # 安全性 & 自動化フック
├── scripts/      # ガードスクリプト
└── templates/    # 生成テンプレート
```

---

## ドキュメント

| ガイド | 説明 |
|--------|------|
| [変更履歴](CHANGELOG_ja.md) | 各バージョンの更新内容 |
| [実装ガイド](IMPLEMENTATION_GUIDE.md) | 内部構造の詳細 |
| [開発フロー](DEVELOPMENT_FLOW_GUIDE.md) | ハーネスの拡張方法 |
| [Cursor 連携](docs/CURSOR_INTEGRATION.md) | 2-Agent ワークフローの設定 |
| [Claude Code 互換性](docs/CLAUDE_CODE_COMPATIBILITY.md) | バージョン要件 |

---

## 謝辞

- **階層的スキル構造**: [AI Masao](https://note.com/masa_wunder)
- **テスト改ざん防止**: [Beagle](https://github.com/beagleworks)（Claude Code Meetup Tokyo 2025.12.22）

---

## ライセンス

**MIT License** — 自由に使用、改変、商用利用可能。

[English](LICENSE.md) | [日本語](LICENSE.ja.md)
