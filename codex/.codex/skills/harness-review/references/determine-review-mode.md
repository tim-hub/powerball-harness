---
name: determine-review-mode
description: "Determine review mode (default/codex) from .claude-code-harness.config.yaml. Falls back to 'default' if Codex CLI is not available or not enabled."
allowed-tools: ["Read"]
---

# Determine Review Mode

`.claude-code-harness.config.yaml` からレビューモードを決定し、**Codex CLI (`codex exec`) の利用可否**を判断する。利用不可の場合は `default` にフォールバックする。

---

## 入力

- **config_file**: 設定ファイルのコンテキスト（任意）

---

## 出力

```json
{
  "review_mode": "default"
}
```

または、Codex が有効な場合：

```json
{
  "review_mode": "codex"
}
```

---

## 実行手順

### Step 1: 設定ファイルの読み込み

```bash
# config_file が存在する場合は読み込む
# 存在しない場合は default にフォールバック
```

### Step 2: レビューモードの確認

設定ファイルの構造：

```yaml
review:
  mode: codex  # または "default"
  codex:
    enabled: true      # または false
    timeout_ms: 60000  # Codex CLI 実行タイムアウト（ミリ秒）
```

### Step 3: Codex CLI 可用性チェック（必須）

Codex CLI の実行可否を確認する。確認に失敗したら `default` にフォールバック。

```bash
# 1) CLI が存在するか
command -v codex

# 2) 認証済みか
codex login status

# 3) 実行確認（timeout は timeout_ms を秒に変換して使用）
TIMEOUT=$(command -v timeout || command -v gtimeout || echo "")
$TIMEOUT 10 codex exec "echo test" >/dev/null 2>&1
```

> `timeout` / `gtimeout` がない環境では、タイムアウト制御なしで短い `codex exec` を行う。

### Step 4: 判定ロジック

```javascript
// 疑似コード例
let review_mode = "default";
const timeoutMs = config.review?.codex?.timeout_ms ?? 60000;
const codexAvailable = detectCodexCli(timeoutMs);

if (config.review?.mode === "codex") {
  if (config.review?.codex?.enabled === true && codexAvailable) {
    review_mode = "codex";
  }
}

return { review_mode };
```

### Step 5: フォールバック条件

以下の場合は `default` にフォールバック：

1. `config_file` が存在しない
2. `review.mode` が未設定または `default`
3. `review.mode` が `codex` だが `review.codex.enabled` が `false`
4. `review.codex.enabled` が未設定
5. `codex` コマンドが見つからない
6. `codex login status` が失敗する
7. `codex exec` の実行確認が失敗する（`timeout_ms` 超過含む）

---

## 使用例

### 例1: Codex 有効

**設定ファイル**:
```yaml
review:
  mode: codex
  codex:
    enabled: true
```

**出力**:
```json
{
  "review_mode": "codex"
}
```

### 例2: Codex 無効

**設定ファイル**:
```yaml
review:
  mode: codex
  codex:
    enabled: false
```

**出力**:
```json
{
  "review_mode": "default"
}
```

### 例3: 設定ファイル未設定

**設定ファイル**: 存在しない、または `review` セクションなし

**出力**:
```json
{
  "review_mode": "default"
}
```

---

## 注意事項

- **安全側に倒す**: Codex 設定が不明確な場合は `default` にフォールバック
- **設定ファイルの優先**: 環境変数などではなく、設定ファイルの値を優先
- **CLI-only 方針**: Codex 呼び出しは `codex exec`（Bash）を使用し、旧サーバー登録方式には依存しない
