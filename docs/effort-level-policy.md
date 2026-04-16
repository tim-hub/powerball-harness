# Effort Level Policy

ひとことで:
Phase 44 の方針では、`xhigh` は見送りではなく正式対象です。ただし「どこで無条件に上げるか」ではなく、「どの役割に使うと得か」を分けて運用します。

たとえると:
料理で火力を上げる話に近いです。全部を強火にすると速い反面、焦げやすくもなります。`xhigh` は「ここだけ火を強くする」と決めて使うのが向いています。

## 前提

- Claude Code `2.1.111` で `xhigh` が追加された
- 公開 changelog では、`xhigh` は Opus 4.7 専用で、他モデルでは `high` にフォールバックするとされている
- Harness は「モデル側にあるから全面採用」ではなく、Plan / Work / Review の役割差で使い分ける

## 対応マトリクス

| レイヤー | 指定値 | 実効値の考え方 | 方針 |
|---------|--------|----------------|------|
| Claude Code `/effort` | `low`, `medium`, `high`, `xhigh`, `max` | UI と CLI から指定可能 | `xhigh` を正式対象に含める |
| `--effort` | 同上 | 実行時指定 | 長い review / advisory で有効 |
| model picker | 同上 | Opus 4.7 利用時に意味がある | reviewer / advisor に相性が良い |
| 非 Opus 4.7 モデル | `xhigh` 指定時でも `high` 相当へフォールバック | changelog 明記あり | docs で明示する |

## Harness の採用方針

| フロー | 推奨 effort | 理由 |
|--------|-------------|------|
| Plan | `high` | 速さと整理力のバランスが良い |
| Work | `high` | 実装は長考より反復確認が重要 |
| Review | `xhigh` | 比較・反証・抜け漏れ検知に向く |
| Advisor | `xhigh` | correction / stop 判断の精度を優先したい |
| Release / Setup | `high` | 手順遵守が中心で、常時 `xhigh` は過剰になりやすい |

## 正式対象としての結論

### 採用するもの

- `xhigh` は Harness docs 上の正式対象に含める
- `/ultrareview` と組み合わせる review 文脈では、`xhigh` を前提に説明してよい
- reviewer / advisor 系の説明では、`high` より一段重い選択肢として扱う

### 採用しないもの

- 全 skill / 全 agent を一律 `xhigh` にすること
- Opus 4.7 以外でも常に `xhigh` と同じ結果が得られる前提で書くこと

## 運用ルール

1. review と advisory を優先して `xhigh` の対象にする  
理由: バグ検知や反証は、実装そのものより thinking 増分の効果が出やすいからです。

2. work は既定 `high` を維持する  
理由: 実装はトークン消費より、短いサイクルでの検証の方が効くことが多いからです。

3. docs では「Opus 4.7 以外は `high` へフォールバック」を毎回明記する  
理由: 利用者が「`xhigh` と書いたのに効いていない」と誤解しやすいためです。

## 具体例

具体例:
大きい PR のレビューでは、`/harness-review` や `/ultrareview` を `xhigh` で回す価値があります。  
一方で、単純な typo 修正や docs 整理まで `xhigh` にすると、待ち時間とコストだけ増えやすいです。

## 注意点

- `xhigh` は「賢くなる魔法」ではなく、より深く考えるための余白です
- 曖昧な指示のままだと、深く考えてもズレた方向に精密化されることがあります
- だからこそ Opus 4.7 では effort だけでなく、指示文の具体性も同時に必要です

## なぜこのやり方か

Harness は Plan → Work → Review の分業があるため、全部を同じ火力で動かすより、役割に合わせて火力を変える方が合理的だからです。
