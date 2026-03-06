# Benchmark Rubric

最終更新: 2026-03-06

この文書は `claude-code-harness` と他ツールを比較するときの再実行可能な rubric です。
README の印象ではなく、静的証拠と実行証拠を分けて採点します。

## Evidence Classes

| Class | 例 | 使いどころ |
|------|----|-----------|
| Static evidence | README, repo tree, hooks 定義, tests, docs, package metadata | 仕組みの有無、設計の明確さ、配布導線の比較 |
| Executed evidence | test run, smoke run, benchmark logs, evidence pack, CI artifact | 主張が再現できるか、guardrail が実際に効くかの比較 |

## Scoring Axes

| Axis | Weight | What to inspect |
|------|--------|-----------------|
| Runtime enforcement | 25 | Hooks, guardrails, deny/warn behavior, lifecycle automation |
| Verification and test credibility | 25 | Unit/integration tests, consistency checks, evidence pack, CI coverage |
| Onboarding and operator clarity | 20 | install flow, docs completeness, claim consistency, quickstart quality |
| Scope discipline and maintainability | 15 | distribution boundary, compatibility story, residue management |
| Positioning and adoption proof | 15 | public narrative, stars/users, reproducible showcase, differentiation |

合計: 100 点

## Review Flow

1. Static evidence を集める
2. Executed evidence が必要な claim を列挙する
3. 実行できた claim と、未実行で保留の claim を分ける
4. 各 axis を採点し、証拠の種別を明記する
5. 「設計は強いが未実証」「市場は強いが runtime enforcement は薄い」など、強みと弱みを別々に書く

## Required Output Format

比較レポートは最低でも次を含める。

- 比較日時
- 対象リポジトリ / バージョン / commit or default branch snapshot
- 実行したコマンド一覧
- Static evidence と Executed evidence の区分
- 軸ごとの点数
- 再現しきれなかった項目

## Reusable Template

```md
# Benchmark Report

- Compared at:
- Repositories / versions:
- Commands executed:

## Static evidence

- Repo structure:
- Docs and claims:
- Guardrails / hooks / tests:

## Executed evidence

- Validation commands:
- Benchmark or smoke runs:
- Evidence artifacts:

## Scores

| Axis | Score | Evidence type | Notes |
|------|-------|---------------|-------|
| Runtime enforcement |  | Static / Executed |  |
| Verification and test credibility |  | Static / Executed |  |
| Onboarding and operator clarity |  | Static / Executed |  |
| Scope discipline and maintainability |  | Static / Executed |  |
| Positioning and adoption proof |  | Static / Executed |  |

## Unverified or blocked items

- なし

## Harness-specific Notes

- `/harness-work all` のような強い claim は、`docs/evidence/work-all.md` の実行証拠が揃ってから高得点化する
- `commands/` や `mcp-server/` のような残置物は減点対象ではなく、**説明が曖昧なときだけ減点**する
- README の主張とテスト/CI/配布境界が噛み合っていない場合は `Onboarding and operator clarity` を下げる
