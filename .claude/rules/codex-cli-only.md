# Codex CLI Only Rule

Codex の呼び出しには **必ず `codex exec` (Bash)** を使用すること。

## 禁止事項

- `mcp__codex__codex` の使用（MCP サーバーは廃止済み）
- ToolSearch で Codex MCP を検索する行為
- `claude mcp add codex` による MCP サーバー再登録

## 推奨ブロック方法（v2.1.78+）

settings.json の `deny` ルールで MCP ツールを直接ブロックするのが最もクリーンな方法:

```json
{
  "permissions": {
    "deny": ["mcp__codex__*"]
  }
}
```

v2.1.77 以降、PreToolUse フックの `allow` 応答は settings.json の `deny` を上書きできないため、
`deny` ルールは最も確実なブロック手段となる。Harness の `.claude-plugin/settings.json` には設定済み。

## 正しい呼び出し方

```bash
# 基本（stdin 方式 + 一意テンポラリファイル）
CODEX_PROMPT=$(mktemp /tmp/codex-prompt-XXXXXX.md)
# タスク内容を書き出し
cat "$CODEX_PROMPT" | $TIMEOUT 120 codex exec - -a never -s workspace-write 2>>/tmp/harness-codex-$$.log
rm -f "$CODEX_PROMPT"

# 並列実行
PROMPT1=$(mktemp /tmp/codex-prompt-XXXXXX.md)
PROMPT2=$(mktemp /tmp/codex-prompt-XXXXXX.md)
cat "$PROMPT1" | $TIMEOUT 120 codex exec - -a never -s workspace-write > /tmp/out1-$$.txt 2>>/tmp/harness-codex-$$.log &
cat "$PROMPT2" | $TIMEOUT 120 codex exec - -a never -s workspace-write > /tmp/out2-$$.txt 2>>/tmp/harness-codex-$$.log &
wait
rm -f "$PROMPT1" "$PROMPT2"
```

## timeout の取得

```bash
TIMEOUT=$(command -v timeout || command -v gtimeout || echo "")
```

macOS では `brew install coreutils` で `gtimeout` をインストール。
