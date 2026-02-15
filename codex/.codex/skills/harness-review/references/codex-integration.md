---
name: codex-integration
description: "Codex 実行時に harness-review を Claude CLI（claude -p）へ委譲する手順"
allowed-tools: ["Read", "Bash"]
---

# Codex 実行時の Claude 委譲レビュー

Codex 上で既存のレビューフローを実行する際に、レビュー依頼先を
Claude CLI（`claude -p`）へ委譲する手順。

---

## 🎯 概要

`review.mode: codex` が有効な場合、Codex 側でオーケストレーションしつつ、
レビュー本体は Claude CLI（`claude -p`）に依頼して結果を統合します。

```
/harness-review 実行（Codex）
    ↓
設定確認: review.codex.enabled
    │
    ├── false → Claude 単独レビュー（従来通り）
    │
    └── true → Claude 委譲レビュー
            │
            ├── Codex 側レビュー集約
            └── Claude CLI 呼び出し（並列）
                    ↓
            結果統合 → 出力
```

---

## 実行フロー

### Step 1: 設定確認

`.claude-code-harness.config.yaml` を確認:

```yaml
review:
  codex:
    enabled: true   # これが true の場合のみ Codex 統合
    auto: false     # false の場合は確認を求める
    timeout_ms: 60000 # Claude CLI 呼び出しのタイムアウト（ミリ秒）
```

### Step 2: 確認プロンプト（auto: false の場合）

> **自動実行ループ**（`/work all` など）では `auto: true` を前提にし、確認プロンプトは出さない。

```markdown
🤖 Claude 委譲レビュー

Claude にレビューを依頼しますか？
- Codex から `claude -p` で依頼し、結果を統合します
- 追加で数秒〜数十秒かかります

[Y] はい / [N] いいえ
```

### Step 3: 並列レビュー実行

**Claude レビュー（従来通り）**:
- 品質判定ゲート
- 変更ファイル分析
- 重点領域のレビュー

**Claude CLI 呼び出し（並列）**:
```bash
# 各レビュー観点を並列実行
# macOS: brew install coreutils
TIMEOUT=$(command -v timeout || command -v gtimeout || echo "")
$TIMEOUT 120 claude -p "$(cat /tmp/claude-review-prompt.md)" \
  > /tmp/claude-review-result.txt 2>/dev/null &
```

### Step 4: 結果統合

両方のレビュー結果を統合して表示:

```markdown
## 📊 統合レビュー結果

### Claude のレビュー

| 観点 | 指摘数 | 重要度 |
|------|--------|--------|
| セキュリティ | 1 | 高 |
| 品質 | 3 | 中 |
| パフォーマンス | 0 | - |

**主な指摘**:
1. [高] SQL インジェクションの可能性 (src/api/users.ts:45)
2. [中] 未使用の import (src/components/Form.tsx:3)
3. ...

---

### Claude 委譲レビュー（`claude -p`）

| 観点 | 指摘数 |
|------|--------|
| コードスタイル | 2 |
| 設計パターン | 1 |

**主な指摘**:
1. 関数が長すぎる（50行超）→ 分割を推奨
2. 命名規則の不統一（camelCase と snake_case の混在）
3. Strategy パターンの適用を検討

---

### 統合サマリ

| 観点 | Codex 側 | Claude 委譲 | アクション |
|------|--------|-------|-----------|
| セキュリティ | ⚠️ 1件 | - | **対応必須** |
| コードスタイル | - | 2件 | 推奨 |
| 設計 | - | 1件 | 検討 |
| 品質 | 3件 | - | 推奨 |
```

---

## タイムアウト設定（必須）

Claude CLI 呼び出しは **`timeout` コマンド** で制御し、超過した場合はフォールバックする。

```yaml
review:
  codex:
    timeout_ms: 60000 # 60秒
```

**動作**:
- タイムアウト発生 → Claude 委譲結果はスキップ
- Codex 側のレビュー結果のみで判定を継続

## フォールバック処理

### Claude CLI エラー時

```markdown
⚠️ Claude 委譲レビューがスキップされました

理由: CLI タイムアウト（exit code 124）
詳細: timeout 120s exceeded

Codex 側のレビュー結果のみを表示します。
```

**対応**:
1. Codex 側のレビュー結果はそのまま表示
2. エラーをログに記録
3. 次回実行時の参考情報を表示

### Claude CLI 未認証・未ログイン

```markdown
⚠️ Claude CLI の認証が必要です

以下のコマンドで再認証してください:
\`\`\`bash
claude login
\`\`\`
```

---

## パフォーマンス考慮

### 大規模変更時

変更ファイルが多い場合（10ファイル超）:

1. **チャンク分割**: 5ファイルずつに分割して Claude に送信
2. **優先度付け**: 変更行数の多いファイルを優先
3. **タイムアウト**: 60秒でタイムアウト、部分結果を表示

### 並列実行の最適化

```
Codex 側レビュー ────┐
                     ├──→ 結果統合
Claude CLI 呼び出し ─┘

最大待機時間: max(Codex 側, Claude CLI) ≈ Claude CLI 待機時間
```

---

## VibeCoder 向け

```markdown
💡 セカンドオピニオンの使い方

**言い方**:
- 「他の AI にも見てもらって」
- 「Claude CLI でダブルチェックして」
- 「ダブルチェックして」

**結果の見方**:
- Codex 側と Claude 委譲で同じ指摘 → 確実に対応が必要
- 片方だけの指摘 → 内容を確認して判断
- 矛盾する指摘 → 両方の理由を確認
```

---

## 関連ドキュメント

- [codex-review/SKILL.md](../../codex-review/SKILL.md) - Codex 統合スキル
- [codex-mcp-setup.md](../../codex-review/references/codex-mcp-setup.md) - セットアップ手順
