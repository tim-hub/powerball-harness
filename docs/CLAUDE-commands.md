# 主要コマンド一覧

Claude harness 開発時に使用するコマンドとハンドオフの一覧です。

## 主要コマンド（開発時に使用）

| コマンド | 用途 |
|---------|------|
| `/plan-with-agent` | 改善タスクを Plans.md に追加 |
| `/work` | タスクを実装（スコープ自動判断、--codex 対応） |
| `/breezing` | Agent Teams でチーム並列完走（--codex 対応） |
| `/reload-plugins` | スキル/フック編集後の即時反映（再起動不要） |
| `/harness-review` | 変更内容をレビュー |
| `/validate` | プラグイン検証 |
| `/remember` | 学習事項を記録 |

## ハンドオフ

| コマンド | 用途 |
|---------|------|
| `/handoff-to-cursor` | Cursor 運用時の完了報告 |

**スキル（会話で自動起動）**:
- `handoff-to-impl` - 「実装役に渡して」→ PM → Impl への依頼
- `handoff-to-pm` - 「PMに完了報告」→ Impl → PM への完了報告

## 関連ドキュメント

- [CLAUDE.md](../CLAUDE.md) - プロジェクト開発ガイド
- [docs/CLAUDE-skill-catalog.md](./CLAUDE-skill-catalog.md) - スキルカタログ
- [docs/CLAUDE-feature-table.md](./CLAUDE-feature-table.md) - 新機能活用テーブル
