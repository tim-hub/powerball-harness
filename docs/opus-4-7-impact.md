# Opus 4.7 影響整理

ひとことで:
Opus 4.7 では `literal instruction following` と `xhigh`、`/ultrareview`、Auto Mode 拡張が Harness にとって実務上の本命で、docs・agents・skills の前提を書き換える必要があります。

たとえると:
同じ運転手でも車のハンドル反応が鋭くなった状態です。前より正確に曲がる分、曖昧な指示のままだと意図しない方向にもそのまま進みやすくなります。

## この文書の見方

- 一次情報は `Claude Code 2.1.111` changelog と Phase 44 の計画
- Opus 4.7 関連で Harness が見るべき 8 項目を整理
- `影響あり` は Harness 側で明示追従が必要
- `影響なし` は自動継承または今回の scope 外

## 8 項目マッピング

| 項目 | 影響 | 影響箇所 | 対応方針 |
|------|------|----------|----------|
| literal instruction following | 影響あり | `agents/`, `skills/`, `CLAUDE.md`, docs | 曖昧語を減らし、判断条件を具体化する |
| `xhigh` effort | 影響あり | `skills/`, `agents/`, `docs/effort-level-policy.md` | 正式対象として採用判断を明文化する |
| task budgets | 影響あり | docs, 将来の `skills/` | 今 phase は調査メモ化。即実装はしない |
| tokenizer 改善 | 影響なし | 全体 | 本体改善として自動継承 |
| vision 2576px | 影響あり | review 系 docs / references | 高解像度レビューの運用上限を明記する |
| memory 改善 | 影響あり | `session-memory`, docs | 長時間セッションと再開品質の説明を更新する |
| `/ultrareview` | 影響あり | `skills/harness-review/`, docs | `harness-review` と競合ではなく関係整理を明記する |
| Auto Mode 拡大 | 影響あり | docs, guardrails 方針 | `--enable-auto-mode` 前提を捨てる |

## 項目別メモ

### 1. literal instruction following

- Opus 4.7 は曖昧な指示を「良い感じに補完する」より、「書いてある通りに忠実に従う」寄りです
- そのため Harness の agent prompt では、次をより具体化する必要があります
  - いつ止まるか
  - 何を報告するか
  - どのファイルを触るか
  - 何を禁止するか

具体例:
「必要に応じて確認する」より、「破壊的操作の前にだけ確認する」の方が Opus 4.7 と相性が良いです。

### 2. `xhigh` effort

- `xhigh` は `high` と `max` の間です
- docs 側では「見送り候補」ではなく正式対象として扱います
- ただし Claude Code frontmatter と API effort の対応が常に一致するとは限らないため、運用ポリシー文書が必要です

### 3. task budgets

- task budgets は魅力的ですが、Harness にはすでに `max_consults` や cost 管理があります
- いきなり導入すると二重管理になりやすいので、Phase 44 では研究メモ止まりにするのが妥当です

### 4. tokenizer 改善

- 同じ内容でも token 消費が少し良くなる方向の改善です
- Harness 独自のコード変更は不要なので、docs では「自動継承」として扱います

### 5. vision 2576px

- 画像レビューで安全に扱える上限が広がると、設計図や PDF の読み取り精度が上がります
- ただし「大きいほど良い」ではなく、上限超過時の事前リサイズ運用はまだ必要です

### 6. memory 改善

- 長い会話の再開品質や長時間実行の安定性に効きます
- Harness では `session-memory` と long-running docs の説明を現実に合わせる必要があります

### 7. `/ultrareview`

- `/ultrareview` は cloud 上で多エージェント並列レビューを回す専用 review です
- Harness の `/harness-review` とは役割が少し違います
- Harness 側の価値は次です
  - Plans.md 連動
  - sprint contract / scope review
  - Codex adversarial review

つまり「どちらかを消す」ではなく、「いつどちらを使うか」を決める問題です。

### 8. Auto Mode 拡大

- `2.1.111` で Auto Mode は `--enable-auto-mode` 前提ではなくなりました
- そのため docs に古い有効化手順が残っていると誤案内になります
- 今回の docs 更新ではこの前提を落とします

## Harness への結論

| 領域 | 結論 |
|------|------|
| docs | 最優先で更新する |
| agents | 曖昧表現を減らして literal 対応する |
| skills | `xhigh` と `/ultrareview` の扱いを明文化する |
| hooks / guardrails | Auto Mode と permission 周辺の説明を更新する |
| code | tokenizer のような本体恩恵は無理に触らない |

## なぜこの方針か

Opus 4.7 の変化は「できることが少し増えた」より、「指示との付き合い方が変わった」が本質だからです。  
そのため、まず docs と prompt 設計を正すのが最も費用対効果が高いです。
