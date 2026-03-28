---
name: harness-review
description: "Harness v3 統合レビュースキル。コード・プラン・スコープを多角的にレビュー。以下で起動: レビュー、コードレビュー、プランレビュー、スコープ分析、セキュリティ、品質チェック、harness-review。実装・新機能・バグ修正・セットアップ・リリースには使わない。"
description-en: "Unified review skill for Harness v3. Multi-angle code, plan, and scope review. Use when user mentions: review, code review, plan review, scope analysis, security, performance, quality checks, PRs, diffs, harness-review. Do NOT load for: implementation, new features, bug fixes, setup, or release."
description-ja: "Harness v3 統合レビュースキル。コード・プラン・スコープを多角的にレビュー。以下で起動: レビュー、コードレビュー、プランレビュー、スコープ分析、セキュリティ、品質チェック、harness-review。実装・新機能・バグ修正・セットアップ・リリースには使わない。"
allowed-tools: ["Read", "Grep", "Glob", "Bash", "Task"]
argument-hint: "[code|plan|scope]"
context: fork
effort: high
---

# Harness Review (v3)

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
| "`harness-plan` 実行後" | `plan`（自動） | 計画レビュー |
| "スコープ確認" | `scope`（自動） | スコープ分析 |
| `harness-review code` | `code` | コードレビュー強制 |
| `harness-review plan` | `plan` | 計画レビュー強制 |
| `harness-review scope` | `scope` | スコープ分析強制 |

## レビュータイプ自動判定

| 直前のアクティビティ | レビュータイプ | 観点 |
|--------------------|--------------|------|
| `harness-work` 後 | **Code Review** | Security, Performance, Quality, Accessibility, AI Residuals |
| `harness-plan` 後 | **Plan Review** | Clarity, Feasibility, Dependencies, Acceptance |
| タスク追加後 | **Scope Review** | Scope-creep, Priority, Feasibility, Impact |

## Code Review フロー

### Step 1: 変更差分を収集

```bash
# BASE_REF が harness-work から渡された場合はそれを使用、なければ HEAD~1 にフォールバック
CHANGED_FILES="$(git diff --name-only --diff-filter=ACMR "${BASE_REF:-HEAD~1}")"
git diff ${BASE_REF:-HEAD~1} --stat
git diff ${BASE_REF:-HEAD~1} -- ${CHANGED_FILES}
```

### Step 1.5: AI Residuals を静的走査

LLM の印象だけで判定せず、再実行できる形で残骸候補を拾う。`scripts/review-ai-residuals.sh` は stable な JSON を返すので、その結果をレビュー根拠として使う。

```bash
# 差分ベース
AI_RESIDUALS_JSON="$(bash scripts/review-ai-residuals.sh --base-ref "${BASE_REF:-HEAD~1}")"

# 対象ファイルを明示したい場合
bash scripts/review-ai-residuals.sh path/to/file.ts path/to/config.sh
```

### Step 2: 5観点でレビュー

| 観点 | チェック内容 |
|------|------------|
| **Security** | SQLインジェクション, XSS, 機密情報露出, 入力バリデーション |
| **Performance** | N+1クエリ, 不要な再レンダリング, メモリリーク |
| **Quality** | 命名, 単一責任, テストカバレッジ, エラーハンドリング |
| **Accessibility** | ARIA属性, キーボードナビ, カラーコントラスト |
| **AI Residuals** | `mockData`, `dummy`, `fake`, `localhost`, `TODO`, `FIXME`, `it.skip`, `describe.skip`, `test.skip`, ハードコードされた秘密情報/環境依存 URL, 明らかな仮実装コメント |

### Step 2.2: AI Residuals の severity 判定表

`AI Residuals` は、まず `scripts/review-ai-residuals.sh` の JSON を確認し、その後に diff 文脈で「本当に出荷リスクか」を最終判断する。

| 重要度 | 代表例 | 判定の考え方 |
|--------|--------|-------------|
| **major** | `localhost` / `127.0.0.1` / `0.0.0.0` の接続先、`it.skip` / `describe.skip` / `test.skip`、ハードコードされた秘密情報っぽい値、dev/staging 固定 URL | 本番事故、誤設定、検証抜けに直結しやすい。1 件でも `REQUEST_CHANGES` |
| **minor** | `mockData`, `dummy`, `fakeData`, `TODO`, `FIXME` | 残骸の可能性は高いが、即事故とは限らない。修正推奨だが verdict は変えない |
| **recommendation** | `temporary implementation`, `replace later`, `placeholder implementation` のような仮実装コメント | コメント単体では即バグ断定できないが、追跡・明確化を促したい |

### Step 2.5: 閾値基準による verdict 判定

各指摘を以下の重要度に分類し、**この基準のみ**で verdict を決定する。

| 重要度 | 定義 | verdict への影響 |
|--------|------|-----------------|
| **critical** | セキュリティ脆弱性、データ損失リスク、本番障害の可能性 | 1 件でも → REQUEST_CHANGES |
| **major** | 既存機能の破壊、仕様との明確な矛盾、テスト不通過 | 1 件でも → REQUEST_CHANGES |
| **minor** | 命名改善、コメント不足、スタイル不統一 | verdict に影響しない |
| **recommendation** | ベストプラクティス提案、将来の改善案 | verdict に影響しない |

> **重要**: minor / recommendation のみの場合は **必ず APPROVE** を返すこと。
> 「あったほうが良い改善」は REQUEST_CHANGES の理由にならない。
> `AI Residuals` でも同じ。`major` に入るのは「出荷事故や誤設定に直結しやすいもの」だけで、単なる残骸候補は `minor` または `recommendation` に留める。

### Step 3: レビュー結果出力

```json
{
  "verdict": "APPROVE | REQUEST_CHANGES",
  "critical_issues": [],
  "major_issues": [],
  "observations": [
    {
      "severity": "critical | major | minor | recommendation",
      "category": "Security | Performance | Quality | Accessibility | AI Residuals",
      "location": "ファイル名:行番号",
      "issue": "問題の説明",
      "suggestion": "修正案"
    }
  ],
  "recommendations": ["必須ではない改善提案"]
}
```

### Step 4: コミット判定

- **APPROVE**: 自動コミット実行（`--no-commit` でなければ）
- **REQUEST_CHANGES**: critical/major の指摘箇所と修正方針を提示。`harness-work` の修正ループで自動修正後に再レビュー（最大 3 回）

## Plan Review フロー

1. Plans.md を読み込む
2. 以下の **5 観点** でレビュー:
   - **Clarity**: タスク説明が明確か
   - **Feasibility**: 技術的に実現可能か
   - **Dependencies**: タスク間の依存関係が正しいか（Depends カラムと実際の依存が一致しているか）
   - **Acceptance**: 完了条件（DoD カラム）が定義され、検証可能か
   - **Value**: このタスクはユーザー課題を解くか？
     - 「誰の、どんな問題」が明示されているか
     - 代替手段（作らない選択肢）は検討されたか
     - Elephant（全員気づいているが放置されている問題）はないか
3. DoD / Depends カラムの品質チェック:
   - DoD が空欄のタスク → 警告（「完了条件が未定義です」）
   - DoD が検証不能（「いい感じ」「ちゃんと動く」等） → 警告 + 具体化提案
   - Depends に存在しないタスク番号 → エラー
   - 循環依存 → エラー
4. 改善提案を提示

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

## Codex Environment

Codex CLI 環境（`CODEX_CLI=1`）では一部ツールが利用不可のため、以下のフォールバックを使用する。

| 通常環境 | Codex フォールバック |
|---------|-------------------|
| `TaskList` でタスク一覧取得 | Plans.md を `Read` して WIP/TODO タスクを確認 |
| `TaskUpdate` でステータス更新 | Plans.md のマーカーを `Edit` で直接更新（例: `cc:WIP` → `cc:完了`） |
| レビュー結果を Task に書き込み | レビュー結果を stdout に出力 |

### 検出方法

```bash
if [ "${CODEX_CLI:-}" = "1" ]; then
  # Codex 環境: Plans.md ベースのフォールバック
fi
```

### Codex 環境でのレビュー出力

Task ツール非対応のため、レビュー結果は標準出力にマークダウン形式で出力する。
Lead エージェントまたはユーザーが結果を読み取り、次のアクションを判断する。

## 関連スキル

- `harness-work` — レビュー後に修正を実装
- `harness-plan` — 計画を作成・修正
- `harness-release` — レビュー通過後にリリース
