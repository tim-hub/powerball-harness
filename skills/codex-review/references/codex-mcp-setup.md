---
name: codex-cli-setup
description: "Codex CLI の検出・認証・`codex exec` 動作確認を行うセットアップ手順（legacy filename: codex-mcp-setup.md）"
allowed-tools: ["Bash", "Read"]
---

# Codex CLI セットアップ（legacy filename: `codex-mcp-setup.md`）

OpenAI Codex CLI をレビュー実行で使うためのセットアップ手順。

## 前提

- `codex exec` を実行できること
- ネットワーク接続があること

## セットアップフロー

1. Codex CLI のインストール
2. Codex 認証
3. `codex exec` の動作確認
4. harness-review 連携設定（任意）

---

## Step 1: Codex CLI インストール

```bash
# macOS (Homebrew)
brew install openai-codex

# cross-platform (npm)
npm install -g @openai/codex
```

確認:

```bash
which codex
codex --version
```

---

## Step 2: Codex 認証

```bash
codex login
codex login status
```

`codex login status` が成功すれば認証完了。

---

## Step 3: `codex exec` 動作確認

```bash
TIMEOUT=$(command -v timeout || command -v gtimeout || echo "")

if [ -n "$TIMEOUT" ]; then
  $TIMEOUT 15 codex exec "echo test" >/tmp/codex-cli-smoke.txt 2>/dev/null
else
  codex exec "echo test" >/tmp/codex-cli-smoke.txt 2>/dev/null
fi
```

確認:

```bash
test -s /tmp/codex-cli-smoke.txt && echo "ok" || echo "failed"
```

---

## Step 4: harness-review で Codex を有効化（任意）

`.claude-code-harness.config.yaml`:

```yaml
review:
  mode: codex
  codex:
    enabled: true
    timeout_ms: 60000
```

この設定で `/harness-review` 時に Codex レビューを使える。

---

## トラブルシューティング

### 問題: `codex` コマンドが見つからない

対処:

```bash
which codex
echo "$PATH"
```

- Homebrew なら `brew list openai-codex`
- npm なら `npm list -g @openai/codex --depth=0`

### 問題: 認証エラー

対処:

```bash
codex logout
codex login
codex login status
```

### 問題: `codex exec` がタイムアウトする

対処:

```bash
# macOS の場合
brew install coreutils

# 再確認
TIMEOUT=$(command -v timeout || command -v gtimeout || echo "")
$TIMEOUT 30 codex exec "echo test"
```

### 問題: レビュー結果が返らない

- ネットワーク接続
- アカウント状態
- `timeout_ms` を増やして再試行

---

## 完了メッセージ（例）

```markdown
✅ Codex CLI セットアップ完了

- codex: 利用可能
- login: 認証済み
- codex exec: 動作確認済み

次の手順:
- `/harness-review` で Codex 併用レビューを実行
```
