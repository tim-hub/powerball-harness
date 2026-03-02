---
name: review-harness
description: "Unified review skill for Harness v3. Multi-angle code, plan, and scope review. Use when user mentions: review, code review, plan review, scope analysis, security, performance, quality checks, PRs, diffs, /review-harness. Do NOT load for: implementation, new features, bug fixes, setup, or release."
description-ja: "Harness v3 統合レビュースキル。コード・プラン・スコープを多角的にレビュー。以下で起動: レビュー、コードレビュー、プランレビュー、スコープ分析、セキュリティ、品質チェック、/review-harness。実装・新機能・バグ修正・セットアップ・リリースには使わない。"
allowed-tools: ["Read", "Grep", "Glob", "Bash", "Task"]
argument-hint: "[code|plan|scope]"
context: fork
---

# Review Skill (v3)

Harness v3 の統合レビュースキル。
以下の旧スキルを統合:

- `harness-review` — コード・プラン・スコープ多角的レビュー
- `codex-review` — Codex CLI によるセカンドオピニオン
- `verify` — ビルド検証・エラー復旧・レビュー修正適用
- `troubleshoot` — エラー・障害の診断と修復

## Quick Reference

| ユーザー入力 | サブコマンド | 動作 |
|------------|------------|------|
| "レビューして" / "review" | `code`（自動） | コードレビュー（直近の変更） |
| "/plan-with-agent 後" | `plan`（自動） | 計画レビュー |
| "スコープ確認" | `scope`（自動） | スコープ分析 |
| `/review code` | `code` | コードレビュー強制 |
| `/review plan` | `plan` | 計画レビュー強制 |
| `/review scope` | `scope` | スコープ分析強制 |

## レビュータイプ自動判定

| 直前のアクティビティ | レビュータイプ | 観点 |
|--------------------|--------------|------|
| `/execute` 後 | **Code Review** | Security, Performance, Quality, Accessibility |
| `/plan` 後 | **Plan Review** | Clarity, Feasibility, Dependencies, Acceptance |
| タスク追加後 | **Scope Review** | Scope-creep, Priority, Feasibility, Impact |

## Code Review フロー

### Step 1: 変更差分を収集

```bash
git diff HEAD~1 --stat
git diff HEAD~1 -- <changed_files>
```

### Step 2: 4観点でレビュー

| 観点 | チェック内容 |
|------|------------|
| **Security** | SQLインジェクション, XSS, 機密情報露出, 入力バリデーション |
| **Performance** | N+1クエリ, 不要な再レンダリング, メモリリーク |
| **Quality** | 命名, 単一責任, テストカバレッジ, エラーハンドリング |
| **Accessibility** | ARIA属性, キーボードナビ, カラーコントラスト |

### Step 3: レビュー結果出力

```markdown
## レビュー結果

### APPROVE / REQUEST_CHANGES

**重大な問題**: なし / {{詳細}}

| 観点 | 評価 | 詳細 |
|------|------|------|
| Security | OK / NG | {{詳細}} |
| Performance | OK / NG | {{詳細}} |
| Quality | OK / NG | {{詳細}} |
| Accessibility | OK / NG | {{詳細}} |

### 推奨改善点（必須ではない）
- {{改善提案}}
```

### Step 4: コミット判定

- **APPROVE**: 自動コミット実行（`--no-commit` でなければ）
- **REQUEST_CHANGES**: 問題箇所と修正方針を提示。`/execute` で修正後に再レビュー

## Plan Review フロー

1. Plans.md を読み込む
2. 以下の観点でレビュー:
   - **Clarity**: タスク説明が明確か
   - **Feasibility**: 技術的に実現可能か
   - **Dependencies**: タスク間の依存関係が正しいか
   - **Acceptance**: 完了条件が定義されているか
3. 改善提案を提示

## Scope Review フロー

1. 追加されたタスク/機能をリスト化
2. 以下の観点で分析:
   - **Scope-creep**: 当初スコープからの逸脱
   - **Priority**: 優先度は適切か
   - **Feasibility**: 現在のリソースで実現可能か
   - **Impact**: 既存機能への影響
3. リスクと推奨アクションを提示

## 異常検知

| 状況 | アクション |
|------|----------|
| セキュリティ脆弱性 | 即座に REQUEST_CHANGES |
| テスト改ざん疑い | 警告 + 修正要求 |
| force push 試み | 拒否 + 代替案提示 |

## 関連スキル

- `execute` — レビュー後に修正を実装
- `plan` — 計画を作成・修正
- `release` — レビュー通過後にリリース
