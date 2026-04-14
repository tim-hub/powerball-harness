# Harness for OpenCode

Claude Code Harness の opencode.ai 互換版です。

## セットアップ方法

### 方法 1: Claude Code からセットアップ（推奨）

```bash
# Claude Code 内で実行
/harness-setup --platform opencode
```

### 方法 2: 手動セットアップ

```bash
# Harness をクローン
git clone https://github.com/tim-hub/powerball-harness.git

# opencode 用コマンドをコピー
cp -r claude-code-harness/opencode/commands/ your-project/.opencode/commands/
cp claude-code-harness/opencode/AGENTS.md your-project/AGENTS.md
```

---

## MCP サーバーセットアップ（オプション）

MCP サーバーを使うと、opencode から Harness のワークフローツールを直接呼び出せます。

```bash
# MCP サーバーをビルド
cd claude-code-harness/mcp-server
npm install
npm run build

# opencode.json をプロジェクトにコピーしてパスを調整
cp claude-code-harness/opencode/opencode.json your-project/
# opencode.json 内のパスを実際のパスに変更
```

Unified memory daemon（共通DB）も併用する場合:

```bash
# memory daemon 起動
./scripts/harness-memd start

# health 確認
./scripts/harness-mem-client.sh health
```

または `harness-mem` で診断まで実行:

```bash
/path/to/claude-code-harness/scripts/harness-mem doctor --platform opencode --fix
```

---

## 利用可能なコマンド

| コマンド | 説明 |
|----------|------|
| `/harness-init` | プロジェクトセットアップ |
| `/plan-with-agent` | 開発プラン作成 |
| `/work` | タスク実行 |
| `/harness-review` | コードレビュー |
| `/sync-status` | 進捗確認 |
| `/handoff-to-opencode` | OpenCode PM への完了報告生成 |

---

## PM モード (OpenCode で計画管理)

OpenCode を PM (Project Manager) として使用する場合のコマンド:

| コマンド | 説明 |
|----------|------|
| `/start-session` | セッション開始（状況把握→計画） |
| `/plan-with-cc` | 計画作成（Evals含む） |
| `/project-overview` | プロジェクト概要把握 |
| `/handoff-to-claude` | Claude Code への依頼生成 |
| `/review-cc-work` | 作業レビュー・承認 |

### ワークフロー（PM モード）

```
OpenCode (PM)                    Claude Code (Impl)
    |                                   |
    | /start-session                    |
    | /plan-with-cc                     |
    | /handoff-to-claude ─────────────> |
    |                                   | /work
    |                                   | /handoff-to-opencode
    | <─────────────────────────────────|
    | /review-cc-work                   |
    |    ├── approve → 次タスク ────────>|
    |    └── request_changes ──────────>|
```

---

## MCP ツール

MCP サーバー経由で以下のツールが利用可能です：

| ツール | 説明 |
|--------|------|
| `harness_workflow_plan` | プラン作成 |
| `harness_workflow_work` | タスク実行 |
| `harness_workflow_review` | コードレビュー |
| `harness_session_broadcast` | セッション間通知 |
| `harness_status` | 状態確認 |
| `harness_mem_resume_pack` | 再開コンテキスト取得 |
| `harness_mem_search` | 共通メモリ検索 |
| `harness_mem_record_checkpoint` | チェックポイント記録 |
| `harness_mem_finalize_session` | セッション確定 |

---

## 使い方

```bash
# opencode を起動
cd your-project
opencode

# コマンドを実行
/plan-with-agent  # プラン作成
/work             # タスク実行
/harness-review   # コードレビュー
```

---

## 制限事項

- Harness プラグインシステム（`.claude-plugin/`）は opencode では使用できません
- memory hooks は `opencode/plugins/harness-memory/index.ts` で提供します（`chat.message` / `session.idle` / `session.compacted`）
- `description-en` フィールドは自動削除されます

---

## 関連リンク

- [Claude Code Harness](https://github.com/tim-hub/powerball-harness)
- [OpenCode Documentation](https://opencode.ai/docs/)
- [OpenCode Commands](https://opencode.ai/docs/commands/)
