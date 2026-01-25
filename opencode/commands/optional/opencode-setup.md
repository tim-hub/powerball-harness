---
description: opencode.ai 用にプロジェクトをセットアップ
---

# /opencode-setup - OpenCode セットアップ

現在のプロジェクトに opencode.ai 互換のコマンドと設定ファイルを生成します。

## VibeCoder Quick Reference

- "**opencode でも使いたい**" → このコマンド
- "**GPT でも Harness 使いたい**" → opencode セットアップ
- "**マルチ LLM 開発したい**" → opencode 互換設定

## Deliverables

- `.opencode/commands/` - opencode 用コマンド
- `AGENTS.md` - opencode 用ルールファイル
- `opencode.json` - MCP 設定（オプション）

---

## Usage

```bash
/opencode-setup
```

---

## Execution Flow

### Step 1: 確認

> 🔧 **opencode.ai 互換ファイルを生成します**
>
> 以下のファイルが作成されます：
> - `.opencode/commands/` (Harness コマンド)
> - `AGENTS.md` (ルールファイル)
>
> 続行しますか？ (y/n)

**ユーザーの回答を待つ**

### Step 2: ディレクトリ作成

```bash
mkdir -p .opencode/commands/core
mkdir -p .opencode/commands/optional
```

### Step 3: コマンドファイル生成

Harness プラグインの `opencode/commands/` からコピー:

```bash
# プラグインディレクトリを特定
PLUGIN_DIR="${CLAUDE_PLUGIN_ROOT:-$(dirname $(dirname $0))}"

# コマンドをコピー
cp -r "$PLUGIN_DIR/opencode/commands/"* .opencode/commands/
```

### Step 4: AGENTS.md 生成

プロジェクトルートに `AGENTS.md` を作成:

```markdown
# AGENTS.md

This project uses [claude-code-harness](https://github.com/Chachamaru127/claude-code-harness) workflow.

## Available Commands

| Command | Description |
|---------|-------------|
| `/plan-with-agent` | Create development plan |
| `/work` | Execute tasks |
| `/harness-review` | Code review |
| `/sync-status` | Check progress |

## Workflow

/plan-with-agent → /work → /harness-review → commit
```

### Step 5: MCP 設定（オプション）

> 🔧 **MCP サーバーも設定しますか？**
>
> MCP を設定すると、opencode から Harness のワークフローツールが使えます。
>
> 設定しますか？ (y/n)

**「y」の場合:**

`opencode.json` を生成:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "harness": {
      "type": "local",
      "enabled": true,
      "command": ["node", "<PLUGIN_DIR>/mcp-server/dist/index.js"]
    }
  }
}
```

### Step 6: 完了メッセージ

> ✅ **OpenCode セットアップ完了**
>
> 📁 **生成されたファイル:**
> - `.opencode/commands/` - Harness コマンド
> - `AGENTS.md` - ルールファイル
> - `opencode.json` - MCP 設定（選択時）
>
> **使い方:**
> ```bash
> # opencode を起動
> opencode
>
> # コマンドを実行
> /plan-with-agent
> /work
> /harness-review
> ```
>
> **ドキュメント:** https://github.com/Chachamaru127/claude-code-harness

---

## Notes

- 既存の `.opencode/` ディレクトリがある場合は上書き確認
- `AGENTS.md` が既存の場合はバックアップを作成
- MCP サーバーを使う場合は事前にビルドが必要

---

## Related Commands

- `/mcp-setup` - MCP サーバーセットアップ
- `/harness-init` - Harness プロジェクト初期化
- `/harness-update` - Harness 更新
