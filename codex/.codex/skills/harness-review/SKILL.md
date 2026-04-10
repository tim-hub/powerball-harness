---
name: HAR:review
description: "HAR:コード・プラン・スコープを多角的にレビュー。セキュリティ・品質チェック。レビュー、コードレビュー、プランレビュー、スコープ分析で起動。実装・新機能・バグ修正・セットアップ・リリースには使わない。"
description-en: "HAR: Multi-angle code, plan, scope review. Security/quality check. Trigger: review, code review, plan review, scope analysis. Do NOT load for: implementation, new features, bugfix, setup, release."
description-ja: "HAR:コード・プラン・スコープを多角的にレビュー。セキュリティ・品質チェック。レビュー、コードレビュー、プランレビュー、スコープ分析で起動。実装・新機能・バグ修正・セットアップ・リリースには使わない。"
allowed-tools: ["Read", "Grep", "Glob", "Bash", "Task", "Monitor"]
argument-hint: "[code|plan|scope] [--dual] [--security]"
context: fork
effort: high
disable-model-invocation: true
---

# Harness Review

Harness の統合レビュースキル。
以下の旧スキルを統合:

- `harness-review` — コード・プラン・スコープ多角的レビュー
- `codex-review` — Codex CLI によるセカンドオピニオン
- `verify` — ビルド検証・エラー復旧・レビュー修正適用
- `troubleshoot` — エラー・障害の診断と修復

---

## 🚀 Step 0: 動作モード決定 (必ず最初に読む)

> **あなたは今この瞬間からレビュアーとして動作します。** 以下の決定木に従って**自動的にレビューを開始**してください。「タスクが不明確」「追加の指示を待つ」で停止してはいけません。bare 呼び出し (`/HAR:review` 引数なし) の場合でも git 状態から対象を自動検出して Code Review を開始します。

### 決定木

```
引数を解析
├── --security が含まれる    → Security Review モード → Step 3.6 へ
├── --dual が含まれる        → Dual Review モード → Step 3.5 へ
├── plan が含まれる          → Plan Review モード → 「## Plan Review フロー」へ
├── scope が含まれる         → Scope Review モード → 「## Scope Review フロー」へ
├── code が含まれる          → Code Review モード → Step 1 へ
└── 引数なし (bare 呼び出し) → 下記「Bare 呼び出し時の default フロー」を実行
```

### Bare 呼び出し時の default フロー

引数無しで `/HAR:review` が呼ばれた場合、以下を順番に実行して**必ず Code Review を自動開始**してください:

#### Step 0.1: git 状態から base ref を自動決定

```bash
# 直近の commit 状況を確認
git log --oneline -15
git status --short

# Base ref を以下の優先順位で自動決定:
# 1. 最後の release tag (例: v4.0.0)
# 2. main/master の HEAD
# 3. HEAD~10 (上記どちらも取れない時)

BASE_REF=""
if LAST_TAG="$(git describe --tags --abbrev=0 2>/dev/null)"; then
  BASE_REF="$LAST_TAG"
elif git rev-parse --verify main >/dev/null 2>&1; then
  BASE_REF="main"
elif git rev-parse --verify master >/dev/null 2>&1; then
  BASE_REF="master"
else
  BASE_REF="HEAD~10"
fi

echo "Auto-detected BASE_REF: ${BASE_REF}"

# 差分が存在することを確認 & スコープ上限チェック
CHANGED_COUNT="$(git log --oneline "${BASE_REF}..HEAD" 2>/dev/null | wc -l | tr -d ' ')"

# 下限フォールバック: 差分ゼロの時は HEAD~5 で再試行
if [ "$CHANGED_COUNT" -eq 0 ]; then
  echo "⚠️ ${BASE_REF}..HEAD に差分がありません。HEAD~5..HEAD にフォールバックします。"
  BASE_REF="HEAD~5"
  CHANGED_COUNT="$(git log --oneline "${BASE_REF}..HEAD" 2>/dev/null | wc -l | tr -d ' ')"
fi

# 上限フォールバック: commits が 10 を超える時は HEAD~10 に絞る
# (最後のリリースタグから多数のコミットが積まれた状態で bare 呼び出し
#  されると、レビュースコープが過大になりレビュー品質が落ちるため)
if [ "$CHANGED_COUNT" -gt 10 ]; then
  echo "⚠️ ${BASE_REF}..HEAD に ${CHANGED_COUNT} commits あります。スコープを HEAD~10 に絞ります。"
  echo "   (フル範囲をレビューしたい場合は明示的に 'code' を指定するか、より古い ref を argument で渡してください)"
  BASE_REF="HEAD~10"
fi
```

#### Step 0.2: レビュータイプを自動判定

base ref から HEAD までのコミットメッセージを調べて、最適なレビュータイプを選ぶ:

```bash
RECENT_TYPES="$(git log --oneline "${BASE_REF}..HEAD" --pretty='%s' | head -20)"

# 判定ロジック:
# - "plan:" で始まる commit が多い → Plan Review
# - "feat|fix|refactor|test|chore|docs|perf|style" 系 → Code Review (default)
# - よくわからない → Code Review (default)

if echo "$RECENT_TYPES" | grep -c '^plan:' | awk '$1 > 2 {exit 0} {exit 1}'; then
  REVIEW_TYPE="plan"
else
  REVIEW_TYPE="code"  # Default
fi

echo "Auto-detected review type: ${REVIEW_TYPE}"
```

#### Step 0.3: 該当のレビューフローへ遷移

- `REVIEW_TYPE=code` → **Step 1 (変更差分を収集) へ進む**。`BASE_REF` 環境変数は Step 0.1 で決定したものを使用
- `REVIEW_TYPE=plan` → **「## Plan Review フロー」セクションへ進む**

**⚠️ 重要**: Step 0 を実行したら、**必ず Step 1 以降に処理を進める**こと。「モードを決定した」だけで停止せず、決定したモードの全フローを最後まで実行してください。

---

## Quick Reference

| ユーザー入力 | サブコマンド | 動作 |
|------------|------------|------|
| "レビューして" / "review" | `code`（自動） | コードレビュー（直近の変更） |
| "`harness-plan` 実行後" | `plan`（自動） | 計画レビュー |
| "スコープ確認" | `scope`（自動） | スコープ分析 |
| `harness-review code` | `code` | コードレビュー強制 |
| `harness-review plan` | `plan` | 計画レビュー強制 |
| `harness-review scope` | `scope` | スコープ分析強制 |
| `harness-review --dual` | `code`（自動） + Codex 並行 | Claude + Codex dual review |
| `harness-review --security` | Security Review | OWASP Top 10 専用セキュリティレビュー（read-only） |

## オプション

| オプション | デフォルト | 説明 |
|-----------|-----------|------|
| `--dual` | なし | Claude Reviewer と Codex Reviewer を並行実行し verdict をマージ。Codex 不可時は自動フォールバック。詳細: [`${CLAUDE_SKILL_DIR}/references/dual-review.md`](${CLAUDE_SKILL_DIR}/references/dual-review.md) |
| `--security` | なし | OWASP Top 10 ベースのセキュリティ専用レビューを実行。read-only（Write/Edit/Bash 書き込み不可）。詳細: [`${CLAUDE_SKILL_DIR}/references/security-profile.md`](${CLAUDE_SKILL_DIR}/references/security-profile.md) |
| `--no-commit` | なし | APPROVE 時の自動コミットを無効化 |

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
  "schema_version": "review-result.v1",
  "verdict": "APPROVE | REQUEST_CHANGES",
  "reviewer_profile": "static | runtime | browser",
  "calibration": {
    "label": "false_positive | false_negative | missed_bug | overstrict_rule",
    "source": "manual | post-review | retrospective",
    "notes": "観察メモ",
    "prompt_hint": "few-shot に使う要点",
    "few_shot_ready": true
  },
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

browser review の場合は `scripts/generate-browser-review-artifact.sh` が `browser_mode` と route / required artifacts を決め、その後に `scripts/write-review-result.sh` で `.claude/state/review-result.json` に正規化して保存する。
このファイルは commit guard と後続フローの共通入力になる。
`calibration` が付くレビュー結果は `scripts/record-review-calibration.sh` で
`.claude/state/review-calibration.jsonl` に追記し、`scripts/build-review-few-shot-bank.sh`
で few-shot bank を更新する。

### Step 3.5: --dual フラグ時の Codex 並行レビュー

`--dual` フラグが指定されている場合、Step 3 の Claude レビューと並行して Codex レビューを実行し、結果をマージする。

1. Codex の利用可否を確認する（`scripts/codex-companion.sh setup --json`）
2. 利用可能であれば `scripts/codex-companion.sh review --base "${BASE_REF:-HEAD~1}"` を起動
3. 両方の verdict を Verdict マージルールで統合する
4. 最終レビュー結果に `dual_review` フィールドを付加する

詳細な手順・出力スキーマ・フォールバック仕様は [`${CLAUDE_SKILL_DIR}/references/dual-review.md`](${CLAUDE_SKILL_DIR}/references/dual-review.md) を参照。

### Step 3.6: --security フラグ時のセキュリティ専用レビュー

`--security` フラグが指定された場合、通常の 5 観点レビューを **スキップ**し、セキュリティ専用フローを実行する。

**Read-only 制約**: このフロー中は Write / Edit / 書き込み系 Bash を一切実行しない。

1. セキュリティプロファイルを読み込む:
   ```
   Read: ${CLAUDE_SKILL_DIR}/references/security-profile.md
   ```
2. OWASP Top 10 全カテゴリを変更差分・関連ファイルに対して確認する
3. 認証・認可フロー、秘密情報取り扱い、依存パッケージ脆弱性をチェックする
4. `reviewer_profile: "security"` を設定して結果を出力する（Step 3 の JSON スキーマに準拠）
5. Security モードの verdict 判定基準（security-profile.md 末尾参照）を適用する

通常の Code Review と `--security` の使い分け:

| | 通常の Code Review | `--security` |
|---|---|---|
| 観点 | Security, Performance, Quality, Accessibility, AI Residuals | Security のみ（OWASP Top 10 全項目） |
| 深度 | セキュリティは概要チェック | 認証・認可・暗号化・依存関係まで網羅 |
| ツール制限 | なし | Read / Grep / Glob / 読み取り Bash のみ |
| 用途 | PR マージ前の総合確認 | セキュリティ集中監査・リリース前の追加確認 |

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
