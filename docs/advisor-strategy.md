# Advisor Strategy

## ひとことで

Advisor Strategy は、
**ふだんは実行役が自走し、難しい場面だけ相談役を呼ぶ**進め方です。

Harness では v1 として、
まず `harness-loop` からこの考え方を入れています。

## たとえると

ずっと横で細かく指示する監督ではなく、
普段は現場の担当者が動き、
「ここは判断が重い」となった時だけ先輩に相談する形です。

この形にすると、
毎回大きな判断役を前に出さずに済むので、
速さと安全のバランスを取りやすくなります。

## 中身

Harness の役割分担は次の 4 つです。

| 役割 | 何をするか |
|------|------------|
| Lead | 全体の流れを整える |
| Worker / executor | 実装や修正を進める |
| Advisor | 方針だけ助言する |
| Reviewer | 最終的な品質判定をする |

大事なのは、
**Advisor は Reviewer の代わりではない**ことです。

Advisor は「次にどう進むか」を返します。
最終的に `APPROVE` するか `REQUEST_CHANGES` にするかは、
これまで通り Reviewer が持ちます。

## いつ advisor が呼ばれるか

v1 では、相談する場面を 3 つに固定しています。

1. 高リスク task の初回実行前
2. 同じ原因の失敗が 2 回続いた後
3. plateau 検知で `PIVOT_REQUIRED` を返す直前

高リスク task とは、今の contract では次のいずれかです。

- `needs-spike`
- `security-sensitive`
- `state-migration`

同じ相談を何度も繰り返さないために、
`trigger_hash` という識別子を使います。

これは、
**「どの task で」「どんな理由で」「どんな失敗だったか」**
をまとめた印です。

同じ `trigger_hash` では 1 回しか相談しません。
さらに task ごとの相談回数は最大 3 回です。

## advisor が返す 3 つの decision

Advisor の返答は `advisor-response.v1` という JSON で固定しています。
decision は次の 3 種類だけです。

| decision | 意味 | Harness の動き |
|----------|------|----------------|
| `PLAN` | 進め方を組み直す | 次の実行 prompt の先頭に助言を入れて再実行 |
| `CORRECTION` | 方針は合っていて局所修正だけ必要 | 修正指示として再実行 |
| `STOP` | これ以上は自走しない方がよい | loop を止め、理由を state に残してエスカレーション |

## 具体例

たとえば、
`state-migration` を含む task を `harness-loop` で回している場面を考えます。

1. loop が sprint contract を読む
2. 高リスク task だと分かる
3. 実装を始める前に advisor に 1 回だけ相談する
4. advisor が `PLAN` を返す
5. loop はその助言を次の prompt の先頭に入れて実行役を走らせる
6. 実装後の最終判定は Reviewer が行う

つまり、
相談役は「実装そのもの」を引き取らず、
実装役が迷わず進めるための方向だけ整えます。

## なぜ `harness-loop` から先に入れるのか

理由は 3 つあります。

1. 長時間実行では、迷った時だけ重い判断を呼ぶ形が特に効くから
2. `run.json` や `cycles.jsonl` があるので、相談の履歴を残しやすいから
3. 既存の Reviewer や checkpoint の流れを崩さずに導入しやすいから

言い換えると、
いきなり全部の実行経路を変えるのではなく、
**いちばん効果が大きく、観察しやすい場所から入れている**
ということです。

## 既知の制約

v1 には、あえて入れていないものがあります。

- Worker が自由に新しい subagent を増やすことはしない
- 自信推定のような自然言語ベース判定はまだ使わない
- advisor の永続化は SQLite ではなく file-based state に留める
- `breezing` と `harness-work` は、まず protocol と文書の統一から進める

## 関連ファイル

- `agents/advisor.md`
- `scripts/run-advisor-consultation.sh`
- `scripts/codex-loop.sh`
- `skills/harness-loop/SKILL.md`
- `skills/harness-loop/references/flow.md`

## なぜこのやり方を取るか

Harness はもともと、
「計画」「実装」「レビュー」を分けて壊れにくくする設計です。

Advisor Strategy を入れる時も、
その土台は残したまま、
**実行役の自走力だけを上げる**方が安全です。

そのため v1 では、
「相談役を追加する」が主で、
「品質判定の責任を移す」はしていません。
