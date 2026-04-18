# Opus 4.7 Vision 使用ガイド

Opus 4.7 で強化された vision 機能（解像度上限 ~2576px）の運用ガイド。
harness-review での PDF・設計図・UI スクリーンショットレビューに適用する。

> **出典**: Claude / Opus 4.7 リリースノート および Claude Code ドキュメント記載の vision 仕様。
> 「短辺 2576px まで安全」はこれらのドキュメントに基づく値であり、それ以外の数値は使用しない。

---

## 基本ガイドライン

### 解像度上限

**短辺 2576px** が Opus 4.7 の vision における運用上の安全上限。

| 画像サイズ | 対応 |
|-----------|------|
| 短辺 2576px 以下 | Read tool でそのまま渡せる |
| 短辺 2576px 超 | **事前リサイズ必須**（下記参照） |

- 「短辺」は縦横のうち小さい方。例: 3840×2160 の画像は短辺 = 2160px（上限内）
- 例: 5000×3000 の画像は短辺 = 3000px（上限超 → リサイズ必要）
- 長辺は 2576px を超えていても短辺が 2576px 以下なら問題ない

---

## 2576px を超える場合の事前リサイズ手順

### macOS (sips コマンド)

```bash
# 解像度確認
sips -g pixelWidth -g pixelHeight input.png

# リサイズ: 長辺・短辺のうち大きい方を 2576px に収める
sips -Z 2576 input.png --out output.png
```

`-Z 2576` はアスペクト比を保ったまま、長辺を 2576px に収める。
短辺が 2576px を超えている場合（縦長画像など）も同様に機能する。

### ImageMagick (クロスプラットフォーム)

```bash
# リサイズ: 縦横どちらも 2576px を超えないよう縮小（アスペクト比保持）
convert input.png -resize 2576x2576\> output.png
```

`\>` は「元のサイズが指定値より大きい場合のみ縮小」する修飾子。
2576px 以下の画像は変更されない。

### 複数ファイルを一括リサイズ (macOS sips)

```bash
# カレントディレクトリの PNG を全てリサイズして resized/ に出力
mkdir -p resized
for f in *.png; do
  sips -Z 2576 "$f" --out "resized/$f"
done
```

---

## PDF の場合の注意点

PDF は **ページ単位**で vision モデルに渡される。
各ページのレンダリング解像度（DPI）が高いと、1 ページが 2576px を超える場合がある。

### DPI と実効解像度の関係

| DPI | A4 ページの実効解像度（縦×横） | 短辺 |
|-----|-------------------------------|------|
| 72 dpi  | 595 × 842 px | 595px（上限内） |
| 150 dpi | 1240 × 1754 px | 1240px（上限内） |
| 200 dpi | 1654 × 2340 px | 1654px（上限内） |
| 250 dpi | 2067 × 2926 px | 2067px（上限内） |
| 300 dpi | 2480 × 3508 px | 2480px（上限内） |
| 360 dpi | 2976 × 4210 px | **2976px（上限超）** |

A4 サイズの場合、300 dpi まではほぼ安全。360 dpi 以上は要注意。

### PDF の DPI を調整してエクスポートする（Ghostscript）

```bash
# 150 dpi で再エクスポート（ファイルサイズも削減）
gs -dNOPAUSE -dBATCH -sDEVICE=pdfwrite \
   -dPDFSETTINGS=/screen \
   -sOutputFile=output_150dpi.pdf input.pdf

# 特定の解像度を明示指定
gs -dNOPAUSE -dBATCH -sDEVICE=pdfwrite \
   -dCompatibilityLevel=1.4 \
   -dDownsampleColorImages=true \
   -dColorImageResolution=200 \
   -sOutputFile=output_200dpi.pdf input.pdf
```

### Read tool での PDF 読み込み

```
Read tool: file_path="spec.pdf", pages="1-5"
```

- `pages` パラメータで読み込むページ範囲を指定する（例: `"1-5"`, `"3"`, `"10-20"`）
- 1 回のリクエストで最大 20 ページまで指定可能
- 20 ページを超える PDF は 20 ページ単位で分割して読み込む

---

## メモリ消費の目安

高解像度画像を複数渡す場合、token 消費が増加する。以下を参考に枚数を調整する。

| 画像 1 枚あたりの解像度 | 概算 token 消費（vision 入力分） |
|------------------------|-------------------------------|
| 512 × 512 px | ~85 トークン |
| 1024 × 1024 px | ~340 トークン |
| 2048 × 2048 px | ~1360 トークン |
| 2576 × 2576 px | ~2100 トークン（上限付近） |

> 上記は概算値。実際の消費量は画像の内容・圧縮率・モデルの内部処理によって変動する。

### N 枚渡す場合の換算例

| 枚数 × 解像度 | 概算 token 消費 |
|--------------|----------------|
| 5 枚 × 2576px | ~10,500 トークン |
| 10 枚 × 2576px | ~21,000 トークン |
| 20 枚 × 2048px | ~27,200 トークン |

1M コンテキスト窓を持つ Opus 4.7 では、これらは全体の 2〜3% 程度に収まる。
ただし、大量の高解像度画像を同一セッションで処理する場合はバッチ分割を推奨する。

---

## よくあるエラーと対処

| 症状 | 原因 | 対処 |
|------|------|------|
| Read tool が画像を返さない | ファイルパスが正しくない、または対応外の形式 | パスを確認。PNG / JPG / GIF / WebP / PDF に限定 |
| レビュー結果が「画像が不明瞭」 | 解像度が低すぎる（100px 以下等） | 高解像度版を用意するか、テキスト補足を添える |
| PDF の一部ページが欠落する | pages 指定が PDF の総ページ数を超えている | `pages` を有効範囲に収める |
| 処理が遅い / タイムアウト | 高解像度画像を大量に渡している | 5 枚単位にバッチ分割して処理する |

---

## 関連ドキュメント

- [`skills/harness-review/references/vision-high-res-flow.md`](../skills/harness-review/references/vision-high-res-flow.md) — 典型シナリオ別フロー（PDF / 設計図 / UI スクリーンショット）
- [`skills/harness-review/SKILL.md`](../skills/harness-review/SKILL.md) — harness-review メインスキル定義
- [`docs/CLAUDE-feature-table.md`](CLAUDE-feature-table.md) — Opus 4.7 機能一覧（vision 2576px エントリ）
