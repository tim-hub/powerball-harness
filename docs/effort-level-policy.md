# Effort Level Policy

## 概要

CC frontmatter の `effort` フィールドと Anthropic API の effort パラメータの対応関係、および Harness における採用方針を定義する。

## CC Frontmatter と API Effort の対応マトリクス

CC v2.1.72 で `max` が廃止され、v2.1.111 で `xhigh` が追加された。

| CC frontmatter `effort` 値 | API effort 実効値 | Opus 4.7 での動作 | 非 Opus 4.7 での動作 |
|----------------------------|------------------|-------------------|---------------------|
| `low` | low | low | low |
| `medium` | medium | medium | medium |
| `high` | high | high | high |
| `xhigh` | xhigh (extended thinking) | xhigh（最大 thinking budget） | `high` にフォールバック（changelog 明記） |

**注記**:
- `xhigh` は CC v2.1.111 で frontmatter に追加された（`CLAUDE-feature-table.md` / `cc-2.1.99-2.1.111-impact.md` 参照）
- `max` は CC v2.1.72 で廃止済み。frontmatter に書いても無効
- `xhigh` を Opus 4.7 以外のモデル（Sonnet 系など）で指定した場合、CC が `high` に自動ダウングレードする

### xhigh が CC 経由で API に渡せるかの判定

**判定: 採用（xhigh を frontmatter で受け付ける証拠あり）**

根拠:
1. `docs/CLAUDE-feature-table.md` の v2.1.111 セクションに `xhigh effort` が `A: 明示追従対象` として記録されている
2. 同ファイルの Opus 4.7 セクションにも `xhigh effort` が `A: 明示追従対象` として記録されている
3. `docs/cc-2.1.99-2.1.111-impact.md` に v2.1.111 での `xhigh` 追加が文書化されている
4. Harness の `opus-4-7-prompt-audit.md` にて「`xhigh`: 呼び出し側が選ぶ推論強度」と定義されている

`xhigh` を frontmatter に書いた場合、CC は Anthropic API に extended thinking を有効にしたリクエストを送る。非 Opus 4.7 モデルではサイレントに `high` 相当へダウングレードされる。reject や error にはならない。

## Harness の採用方針

| フロー | 採用 effort | 理由 |
|--------|------------|------|
| Plan | `high` | 速さと整理力のバランスが良い |
| Work (Worker agent) | `high` | 実装は長考より反復確認が重要 |
| Review (Reviewer agent, harness-review) | `xhigh` | 比較・反証・抜け漏れ検知に thinking 増分の効果が出る |
| Advisor | `xhigh` | PLAN / CORRECTION / STOP の判断精度を優先 |
| Release / Setup | `high` | 手順遵守が中心で、常時 `xhigh` は過剰 |

### frontmatter 更新対象

| ファイル | 変更前 | 変更後 | 理由 |
|--------|--------|--------|------|
| `agents/reviewer.md` | `effort: medium` | `effort: xhigh` | Review に xhigh を採用 |
| `agents/advisor.md` | `effort: high` | `effort: xhigh` | Advisor に xhigh を採用 |
| `skills/harness-review/SKILL.md` | `effort: high` | 変更なし | スキルの effort は呼び出し側が上書きするため high を維持 |

## 運用ルール

1. **review と advisory を優先して `xhigh` の対象にする**
   理由: バグ検知や反証は、実装そのものより thinking 増分の効果が出やすい。

2. **work は既定 `high` を維持する**
   理由: 実装はトークン消費より、短いサイクルでの検証の方が効くことが多い。

3. **docs では「Opus 4.7 以外は `high` へフォールバック」を明記する**
   理由: 利用者が「`xhigh` と書いたのに効いていない」と誤解しやすい。

4. **全 skill / 全 agent を一律 `xhigh` にしない**
   理由: コストとレイテンシが無駄に増加する。役割差で使い分けること。

## 見送り rationale（採用しないもの）

以下は採用しない。見送り理由を明記する。

| 項目 | 見送り理由 |
|------|-----------|
| Worker agent を `xhigh` にすること | 実装ループは長考より速い反復が重要。xhigh のコスト増分に見合う品質向上が得られない |
| Setup / Release スキルを `xhigh` にすること | 手順遵守が中心で、judgment より recall が重要な場面が多い |
| `max` の復活 | CC v2.1.72 で廃止済み。`xhigh` がその後継 |

## 注意点

- `xhigh` は「賢くなる魔法」ではなく、より深く考えるための余白
- 曖昧な指示のままだと、深く考えてもズレた方向に精密化される
- Opus 4.7 以外のモデルでは `xhigh` を指定しても `high` 相当にフォールバックするため、期待した効果が出ない場合がある
- `opus-4-7-prompt-audit.md` の合格条件 5: `xhigh` は「呼び出し側が選ぶ推論強度」であり、agent prompt が free-text marker から推測するものではない

## 関連ファイル

- `docs/CLAUDE-feature-table.md` — v2.1.111 / Opus 4.7 の機能一覧
- `docs/cc-2.1.99-2.1.111-impact.md` — xhigh 追加の詳細
- `.claude/rules/opus-4-7-prompt-audit.md` — xhigh の運用ノブ定義
- `agents/reviewer.md` — Reviewer effort 設定
- `agents/advisor.md` — Advisor effort 設定
