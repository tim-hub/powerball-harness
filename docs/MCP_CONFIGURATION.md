# MCP Configuration Guide

Claude Code の MCP (Model Context Protocol) サーバー設定ガイド。

## Auto-Enable Threshold (v2.1.9+)

MCP ツールの自動有効化を閾値で制御できます。

### 構文

```json
{
  "mcpServers": {
    "my-server": {
      "command": "npx",
      "args": ["-y", "@my/mcp-server"],
      "autoApprove": "auto:5"
    }
  }
}
```

### 設定値

| 値 | 動作 |
|-------|----------|
| `true` | 常に自動承認 |
| `false` | 手動承認のみ（デフォルト） |
| `"auto:N"` | N 回の手動承認後に自動承認に切り替え |

### auto:N の動作

1. 最初の N 回は手動で承認を求められる
2. N 回承認すると、以降は自動承認される
3. セッションをまたいで承認カウントが保持される

### 使用例

```json
{
  "mcpServers": {
    // 信頼済みサーバー: 常に自動承認
    "harness-ui": {
      "command": "node",
      "args": ["./harness-ui/dist/mcp/server.js"],
      "autoApprove": true
    },

    // 新しいサーバー: 5回の手動承認後に自動化
    "external-api": {
      "command": "npx",
      "args": ["-y", "@external/mcp-server"],
      "autoApprove": "auto:5"
    },

    // 慎重なサーバー: 10回の手動承認後に自動化
    "database-access": {
      "command": "npx",
      "args": ["-y", "@db/mcp-server"],
      "autoApprove": "auto:10"
    },

    // 高リスクサーバー: 常に手動承認
    "production-deploy": {
      "command": "./scripts/deploy-mcp.sh",
      "autoApprove": false
    }
  }
}
```

## 推奨設定

| サーバー種別 | 推奨設定 | 理由 |
|-------------|---------|------|
| 内部/自作サーバー | `true` | 信頼性が確認済み |
| 公式/検証済みサーバー | `"auto:3"` | 数回の確認で十分 |
| 新規/未検証サーバー | `"auto:10"` | 慎重に確認 |
| 本番環境操作 | `false` | 常に明示的な承認が必要 |

## MCP Auto Mode (v2.1.7+)

Claude Code v2.1.7 以降では、コンテキストウィンドウの 10% を超える MCP ツールは自動的に `MCPSearch` ツール経由でアクセスされます。

### 無効化

```json
{
  "disallowedTools": ["MCPSearch"]
}
```

### カスタム閾値

コンテキストウィンドウの N% を閾値として設定:

```bash
# 20% を閾値に設定
claude --mcp-auto-threshold 20
```

## Harness での MCP 活用

### harness-ui MCP サーバー

ハーネスには組み込みの MCP サーバーがあり、以下の機能を提供します:

- プロジェクト状態の取得
- Plans.md の読み取り/更新
- セッション情報の取得

設定例:

```json
{
  "mcpServers": {
    "harness-ui": {
      "command": "node",
      "args": ["${HARNESS_PATH}/harness-ui/dist/mcp/server.js"],
      "autoApprove": true
    }
  }
}
```

### Codex MCP サーバー

外部レビュー用の Codex MCP サーバー:

```json
{
  "mcpServers": {
    "codex": {
      "command": "npx",
      "args": ["-y", "@openai/codex-mcp"],
      "autoApprove": "auto:3",
      "env": {
        "OPENAI_API_KEY": "${OPENAI_API_KEY}"
      }
    }
  }
}
```

## 関連ドキュメント

- [CLAUDE_CODE_COMPATIBILITY.md](./CLAUDE_CODE_COMPATIBILITY.md) - 互換性マトリクス
- [Claude Code MCP ドキュメント](https://docs.anthropic.com/claude-code/mcp)
