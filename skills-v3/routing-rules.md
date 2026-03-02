# Skill Routing Rules — v3 (Reference)

Harness v3 の5動詞スキル間ルーティングルールのリファレンス。

> **SSOT の場所**: 各スキルの `description` フィールドがルーティングの SSOT です。
> このファイルは詳細な説明と例を提供するリファレンスであり、実際のルーティングは各スキルの description に依存します。

## 5動詞スキル ルーティングテーブル

| スキル | トリガーキーワード | 除外 |
|--------|-----------------|------|
| `plan-harness` | create a plan, add tasks, update Plans.md, mark complete, check progress, sync status, where am I, /plan-harness, /sync-status | implementation, code review, release |
| `work-harness` | implement, execute, /work-harness, /work, do everything, build features, run tasks, breezing, team run, --codex, --parallel | planning, code review, release, setup |
| `review-harness` | review, code review, plan review, scope analysis, security, performance, quality checks, PRs, diffs, /review-harness | implementation, new features, bug fixes, setup, release |
| `release-harness` | release, version bump, create tag, publish, /release-harness | implementation, code review, planning, setup |
| `setup-harness` | setup, initialization, new project, CI setup, codex CLI setup, harness-mem, agent setup, symlinks, /setup-harness | implementation, code review, release, planning |

## 詳細ルーティング

### plan-harness スキル

**トリガー**（いずれかにマッチ）:
- "計画を作って" / "create a plan"
- "タスクを追加して" / "add a task"
- "Plans.md を更新して"
- "完了にして" / "mark complete" / "mark as done"
- "今どこ？" / "where am I" / "check progress"
- "進捗確認"
- "/plan-harness" / "/sync-status"
- "sync status" / "sync Plans.md"

**除外**（いずれかにマッチしたら除外）:
- "実装して" / "implement"
- "コードレビュー"
- "リリース"

### work-harness スキル

**トリガー**（いずれかにマッチ）:
- "実装して" / "implement"
- "実行して" / "execute"
- "/work-harness" / "/work"
- "全部やって" / "do everything"
- "ここだけ" / "just this"
- "breezing" / "チーム実行"
- "--codex" / "--parallel"
- "ビルドして" / "build"

**除外**（いずれかにマッチしたら除外）:
- "計画" / "plan"（実装なし）
- "レビュー"（実装なし）
- "リリース"
- "セットアップ"

### review-harness スキル

**トリガー**（いずれかにマッチ）:
- "レビューして" / "review"
- "コードレビュー" / "code review"
- "プランレビュー" / "plan review"
- "スコープ確認"
- "セキュリティチェック"
- "品質チェック"
- "PR レビュー"
- "/review"
- "diff を見て" / "変更を確認"

**除外**（いずれかにマッチしたら除外）:
- "実装して"（実装依頼）
- "新機能を追加"
- "バグを修正"
- "セットアップ"
- "リリース"

### release-harness スキル

**トリガー**（いずれかにマッチ）:
- "リリース" / "release"
- "バージョンバンプ" / "version bump"
- "タグを作成" / "create tag"
- "公開" / "publish"
- "CHANGELOG を更新"
- "/release-harness"

**除外**（いずれかにマッチしたら除外）:
- "実装して"
- "コードレビュー"
- "計画"
- "セットアップ"

### setup-harness スキル

**トリガー**（いずれかにマッチ）:
- "セットアップ" / "setup"
- "初期化" / "initialization" / "init"
- "新規プロジェクト" / "new project"
- "CI セットアップ"
- "Codex CLI セットアップ"
- "harness-mem"
- "エージェント設定"
- "symlink 更新"
- "/setup-harness"

**除外**（いずれかにマッチしたら除外）:
- "実装して"
- "コードレビュー"
- "リリース"
- "計画を作って"

## 優先順位ルール

1. **除外が最優先**: 除外キーワードにマッチしたスキルは絶対にロードしない
2. **具体的なキーワードが優先**: 完全一致 > 部分一致
3. **あいまいな場合**: `plan` > `execute` > `review` の順で優先（保守的な方を選択）

## 拡張パック（extensions/）

コアスキル以外の機能は `skills-v3/extensions/` に格納:

| スキル | 用途 |
|--------|------|
| `auth` | 認証・決済機能（Clerk, Stripe） |
| `crud` | CRUD 自動生成 |
| `ui` | UIコンポーネント生成 |
| `agent-browser` | ブラウザ自動化 |
| `gogcli-ops` | Google Workspace 操作 |
| `codex-review` | Codex セカンドオピニオン |
| `notebookLM` | NotebookLM 連携 |
| `generate-slide` | スライド生成 |
| `deploy` | デプロイ自動化 |
| `memory` | SSOT・メモリ管理 |
| `cc-cursor-cc` | Cursor ↔ Claude Code 連携 |

## 更新ルール

1. **description = SSOT**: 各スキルの `description` フィールドがルーティングの正式な定義
2. **このファイルの役割**: 詳細な説明と判定フローのリファレンス（SSOT ではない）
3. **完全リスト維持**: 汎用表現を使わず、具体的なキーワードを列挙する
