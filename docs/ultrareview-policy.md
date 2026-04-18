# `/ultrareview` と `/harness-review` の連携方針

Phase 44.8.1 で確定した方針ドキュメント。

---

## 1. `/ultrareview` の挙動

`/ultrareview` は Claude Code 2.1.111 で追加された **built-in slash command**。

| 属性 | 内容 |
|------|------|
| セッション種別 | single-turn dedicated review session |
| 実行主体 | CC ネイティブ（Harness agent 外） |
| 入力 | 現在の作業ツリー差分（自動収集） |
| 出力 | インラインの自然言語レビュー結果 |
| 出力スキーマ | 未定義（CC 内部形式） |
| Plans.md 連動 | なし |
| sprint-contract 検証 | なし |
| Codex adversarial review | なし |
| Reviewer agent 呼び出し | なし |

`/ultrareview` は「ユーザーが直接 CC に対してアドホックなレビューを求める」entrypoint であり、
Harness の自動化フロー（Plan → Work → Review）の外側で動作する。

---

## 2. `/harness-review` との差分

| 観点 | `/ultrareview` | `/harness-review` |
|------|----------------|-------------------|
| 実行主体 | CC ネイティブ | Harness skill (context: fork) |
| セッション | single-turn | multi-step（Step 0〜4） |
| Plans.md 連動 | なし | あり（cc:WIP 確認・cc:完了 更新） |
| sprint-contract 検証 | なし | あり（`.claude/state/contracts/<task>.sprint-contract.json`） |
| Codex adversarial review | なし | あり（`--dual` フラグ時） |
| Reviewer agent | なし | あり（`reviewer` agent、`review-result.v1` 出力） |
| 出力スキーマ | 非定義 | `review-result.v1`（機械可読 JSON） |
| AI Residuals スキャン | なし | あり（`scripts/review-ai-residuals.sh`） |
| 修正ループ | なし | あり（REQUEST_CHANGES 時、最大 3 回） |
| Security 専用モード | なし | あり（`--security`、OWASP Top 10） |
| UI Rubric モード | なし | あり（`--ui-rubric`、4 軸採点） |
| 対象ユーザー | ユーザーが直接 | Lead / breezing フローの自動呼び出し |

---

## 3. 確定方針: **(B) `/harness-review` 優先 — Harness flow 内で `/ultrareview` を呼ばない**

### 3.1 rationale

**ルール 5 との整合**: `.claude/rules/opus-4-7-prompt-audit.md` は
「`/ultrareview` は呼び出し側の review entrypoint。agent 定義側では `review-result.v1` を契約にする」
と定めている。Harness の Reviewer agent・harness-review skill は `review-result.v1` を出力契約とする。
`/ultrareview` をその内部で呼ぶことは、`review-result.v1` の機械可読保証を失わせる。

**スキーマ不一致**: `/ultrareview` の出力は CC 内部形式であり、
`review-result.v1` の `verdict`, `critical_issues`, `major_issues` フィールドを含まない。
Harness の修正ループ・commit guard・sprint-contract 検証はすべて `review-result.v1` に依存しており、
スキーマ変換のオーバーヘッドを正当化できるメリットがない。

**責務の分離**: `/ultrareview` はユーザーがアドホックに CC へ要求する entrypoint。
Harness flow 内の自動レビューは `reviewer` agent（`review-result.v1`）と
`codex-companion.sh review` がカバーする。両者は用途が異なり並立で問題ない。

**フォールバックの安全性**: `codex-companion.sh review` が利用不可の場合は
`reviewer` agent（static / runtime / browser profile）にフォールバックする。
`/ultrareview` を追加するとフォールバックパスが増えてデバッグが困難になる。

### 3.2 使い分けガイド

| シーン | 推奨コマンド |
|--------|------------|
| PR マージ前の総合確認（Harness 外） | `/ultrareview` |
| Harness Plan→Work 後の自動レビュー | `/harness-review`（自動呼び出し） |
| Codex セカンドオピニオン付きレビュー | `/harness-review --dual` |
| セキュリティ集中監査 | `/harness-review --security` |
| UI 品質採点 | `/harness-review --ui-rubric` |

---

## 4. 今後の対応

- `/ultrareview` は CC built-in として成熟した段階で再評価する（次回評価 Phase: 45 以降）
- Harness 内で `/ultrareview` を呼ぶ場合は、`review-result.v1` へのスキーマ変換レイヤーが
  `scripts/codex-companion.sh` に実装されてからとする（現時点では未実装）
- 方針変更は `.claude/rules/opus-4-7-prompt-audit.md` ルール 5 の改訂と同時に行う

---

*決定: Phase 44.8.1 / 2026-04-18*
