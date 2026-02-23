# スキルカタログ

スキル階層構造・全カテゴリ一覧・開発用スキルの参照ドキュメント。

## スキル評価フロー

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

## スキル階層構造

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

## 全スキルカテゴリ一覧

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

## 開発用スキル（非公開）

以下のスキルは開発・実験用であり、リポジトリには含まれません（.gitignore で除外）：

```
skills/
├── test-*/      # テスト用スキル
└── x-promo/     # X投稿作成スキル（開発用）
```

これらのスキルは個別の開発環境でのみ使用し、プラグイン配布には含めないこと。

## 関連ドキュメント

- [CLAUDE.md](../CLAUDE.md) - プロジェクト開発ガイド（概要）
- [docs/CLAUDE-feature-table.md](./CLAUDE-feature-table.md) - Claude Code 新機能活用テーブル
- [docs/CLAUDE-commands.md](./CLAUDE-commands.md) - 主要コマンド一覧
- [.claude/rules/skill-editing.md](../.claude/rules/skill-editing.md) - スキルファイル編集ルール
