# Codex Review Integration

`--codex-review` フラグ使用時の Codex CLI レビュー統合仕様。

## 概要

```
Reviewer の通常レビュー (harness-review 4観点)
    +
Codex CLI 並列エキスパートレビュー (4エキスパート)
    ↓
結果統合 → 総合判定
```

## 有効化条件

| 条件 | 必須 |
|------|------|
| `--codex-review` フラグ指定 | Yes |
| Codex CLI がインストール済み (`which codex`) | Yes |

未設定時のメッセージ:

```text
⚠️ Codex CLI がインストールされていません。

`npm install -g @openai/codex` でインストールするか、--codex-review なしで実行してください。
```

## Reviewer のレビューフロー (--codex-review 時)

```
Step 1: harness-review 4観点レビュー (通常通り)
    ↓
Step 2: Codex CLI 4エキスパート並列レビュー
    ├── Security Expert (OWASP準拠)
    ├── Performance Expert (パフォーマンス分析)
    ├── Quality Expert (コード品質・保守性)
    └── Architect Expert (設計・スケーラビリティ)
    ↓
Step 3: 結果統合
    ↓
Step 4: 総合判定 → Lead に報告
```

## Codex エキスパート呼び出し

### 4エキスパートの並列呼び出し

Reviewer は Bash (`codex exec`) を使い、4つのレビュー観点を並列で実行:

```
エキスパート呼び出し (並列):
  1. Security Expert → セキュリティ脆弱性検出
  2. Performance Expert → パフォーマンス問題検出
  3. Quality Expert → コード品質評価
  4. Architect Expert → 設計レビュー
```

### レート制限への配慮

```
Codex CLI の推奨並列数: 最大 4 (エキスパート数)
各エキスパートのタイムアウト: timeout 120 秒
タイムアウト時 (exit 124): harness-review の結果のみで判定
```

## 結果統合ルール

### 統合アルゴリズム

```
1. harness-review の findings を基盤リストとする
2. 各 Codex エキスパートの findings を追加:
   a. 既存 finding と重複チェック (同一ファイル + 同一行 + 類似 issue)
   b. 重複あり → harness-review 側の finding を優先、Codex の詳細を追記
   c. 重複なし → 新規 finding として追加 (source: "codex" タグ付き)
3. severity の調整:
   - Codex が Critical を検出 → harness-review に Critical がなくても追加
   - harness-review と Codex で severity が異なる → 高い方を採用
```

### 判定への影響

```
harness-review の判定を基本とし、Codex の結果で調整:

1. harness-review: APPROVE + Codex: 問題なし → APPROVE
2. harness-review: APPROVE + Codex: Critical 検出 → REQUEST CHANGES に格下げ
3. harness-review: REQUEST CHANGES + Codex: 結果あり → findings を追加
4. harness-review: REJECT + Codex: 任意 → REJECT (変更なし)
```

## 統合レポートフォーマット

```json
{
  "decision": "REQUEST_CHANGES",
  "grade": "C",
  "review_sources": {
    "harness_review": {
      "findings_count": 3,
      "grade": "B"
    },
    "codex_review": {
      "findings_count": 2,
      "experts_responded": 4,
      "grade_adjustment": "-1 (Critical found by Security Expert)"
    }
  },
  "findings": [
    {
      "severity": "critical",
      "category": "security",
      "file": "src/auth/login.ts",
      "line": 15,
      "issue": "SQL インジェクション脆弱性",
      "suggestion": "パラメータ化クエリを使用",
      "source": "codex:security-expert",
      "auto_fixable": true
    },
    {
      "severity": "warning",
      "category": "quality",
      "file": "src/auth/login.ts",
      "line": 42,
      "issue": "エラーハンドリング不足",
      "suggestion": "try-catch で適切にハンドリング",
      "source": "harness-review",
      "auto_fixable": true
    }
  ],
  "summary": "Codex Security Expert が SQL インジェクション脆弱性を検出。harness-review の判定を B → C に格下げ。"
}
```

## Codex なしへのフォールバック

以下の場合、Codex なしで継続:

| 状況 | 対応 |
|------|------|
| Codex CLI 未インストール | Phase 0 で警告、harness-review のみで実行 |
| Codex CLI タイムアウト (exit 124) | harness-review の結果のみで判定 |
| Codex CLI エラー | エラーログ記録、harness-review のみで判定 |
| 4エキスパート中 1-3 が応答 | 応答したエキスパートの結果のみ統合 |

## codex-review スキルとの関係

| 項目 | codex-review スキル (単体) | breezing --codex-review |
|------|---------------------------|------------------------|
| 呼び出し元 | ユーザー直接 or Lead | Reviewer Teammate |
| レビュー対象 | git diff | git diff (Phase 3 時点) |
| harness-review 統合 | なし | あり (Reviewer が統合) |
| リテイク | 手動 | 自動 (三者分離ループ) |

> breezing の --codex-review は、codex-review スキルの機能を
> Reviewer Teammate 内で再利用する形で統合しています。
