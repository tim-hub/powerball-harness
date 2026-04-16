---
name: x-article
description: "X の記事機能向けに、タイトル設計、ヘッダー画像、差し込み画像、品質チェック、公開後の拡散導線まで含めた長文記事パッケージを作成。Use when user mentions: X article, Xの記事, 記事として投稿, 長文記事, article header image, inline image, 記事カバー. Do NOT load for: short X post, tweet, thread-only announcement, GitHub release notes."
description-en: "Create a full X Articles package with title options, header image, inline images, article draft, quality checks, and promotion plan. Use when user mentions: X article, longform article, article header image, inline image. Do NOT load for: short tweet, thread-only announcement, GitHub release notes."
description-ja: "X の記事機能向けに、タイトル設計、ヘッダー画像、差し込み画像、品質チェック、公開後の拡散導線まで含めた長文記事パッケージを作成。短いポストではなく、X の記事機能で出す前提のときに使う。"
allowed-tools: ["Read", "Write", "Edit", "Bash", "Glob", "Grep", "Agent", "AskUserQuestion"]
argument-hint: "[theme|release|latest|url]"
user-invocable: false
---

# X Article - X記事パッケージ生成

X の記事機能向けに、本文だけでなく「開かれるタイトル」「ヘッダー画像」「差し込み画像」「品質チェック」「公開後に伸ばすための導線」までまとめて作るスキル。

このスキルでは、画像を単なるプロンプト案で終わらせず、**Nano Banana Pro（Gemini の画像生成 API）で実画像まで作る**ことを基本とする。

## Quick Reference

- "**Xの記事にして**" → this skill
- "**ポストじゃなくて記事として出したい**" → this skill
- "**長文で公開したい。タイトルと画像も作って**" → this skill
- "**X Articles 用に整えて**" → this skill

## まず理解しておくこと

- **Xの記事機能**: X 上で長文を公開する機能。X 公式ヘルプでは、プレミアム系会員向け機能として案内されている。
- **ヘッダー画像**: 記事の入口になる大きな画像。X 公式も「読者に記事を開いてもらうための重要な要素」と案内している。
- **差し込み画像**: 記事の途中に入れる画像。読み疲れを減らし、理解と共有を助ける。

X 公式仕様の要点は [x-articles-spec.md](${CLAUDE_SKILL_DIR}/references/x-articles-spec.md) を読むこと。

## このスキルが作るもの

1. 記事タイトル候補 3 本以上
2. 記事本文の完成原稿
3. ヘッダー画像プロンプト
4. 差し込み画像プロンプト 2〜4 本
5. alt テキスト
6. 品質チェック結果
7. 公開後の拡散メモ
8. 画像生成リクエスト / レスポンスの保存物

## 出力先

```
out/social/
└── <slug>/
    ├── article.md             # 記事本文の生成原稿
    ├── image-prompts.md       # 画像生成プロンプトと alt テキスト
    ├── quality-report.md      # 記事・画像の品質チェック
    ├── header/                # ヘッダー画像候補
    ├── inline/                # 差し込み画像候補
    └── responses/             # Nano Banana Pro の request / response
```

生成中のパッケージは、**記事本文も画像も同じ `out/social/<slug>/` にまとめる**。

`docs/social/` は、公開用に整えた正本をあとで昇格させるときだけ使う。
つまり `x-article` のデフォルト出力は `out/social/` で完結する。

## 実行フロー

```
/x-article [theme|release|latest|url]
    │
    ├─[Step 1] テーマを固める
    │   ├── 誰向けの記事か
    │   ├── 読後に何を思ってほしいか
    │   ├── 何をしてほしいか
    │   └── 根拠になる事実・実例を集める
    │
    ├─[Step 2] タイトルと構成を作る
    │   ├── タイトル候補を 3 本以上
    │   ├── 冒頭フックを作る
    │   └── 小見出し付きの骨子を作る
    │
    ├─[Step 3] 本文を完成させる
    │   ├── 短い段落で書く
    │   ├── 3〜5 段落ごとに小見出し
    │   ├── 箇条書きと太字で読みやすくする
    │   └── 強い締めと CTA を置く
    │
    ├─[Step 4] 画像を作る
    │   ├── ヘッダー画像を 1 枚以上
    │   ├── 差し込み画像を 2〜4 枚
    │   ├── Nano Banana Pro で生成
    │   ├── claude-code-harness ロゴを参照画像として入力
    │   ├── 日本語テキストで生成
    │   └── alt テキストを付ける
    │
    ├─[Step 5] 品質チェック
    │   ├── 記事の可読性
    │   ├── タイトルの強さ
    │   ├── 根拠と具体例
    │   ├── 画像品質
    │   └── X 公式の推奨との整合
    │
    └─[Step 6] 公開後の伸ばし方を添える
        ├── 予告ポスト
        ├── 抜粋ポスト
        ├── ピン留め
        └── 初動返信方針
```

## 実行ルール

### 1. テーマが曖昧なら先に整理する

最低でも次を確認する。

- 読者: 誰に向けた記事か
- 約束: 読み終わった読者に何を持ち帰ってほしいか
- 根拠: その主張を支える事実、実例、数字、比較
- 行動: 読後に何をしてほしいか

これが曖昧なまま書き始めない。

### 2. 事実と提案を分ける

- **事実**: X 公式ヘルプや一次情報で確認できること
- **提案**: Harness としての伸ばし方、書き方、画像の作り方

本文やメモでは、この 2 つを混ぜない。

### 3. 画像は「飾り」ではなく「理解を助ける部品」にする

- ヘッダー画像: クリックを取る入口
- 差し込み画像: 説明の補助、比較、要点整理

単にきれいなだけの画像で終わらせない。

### 4. ブランドトーンを固定する

画像生成では、毎回次の 2 点を守る。

1. **白背景メインのミニマルなデザイン**
2. **`claude-code-harness` ロゴを適切に使う**

ロゴは後から雑に合成するのではなく、**Gemini へ参照画像として渡して、最初から構図に統合させる**のを基本とする。

### 5. 画像内テキストは日本語を基本にする

見出しやラベルを画像に入れる場合は、日本語を基本とする。

英語の短い補助語を使うのは、次の条件を両方満たすときだけに限る。

- 日本語主体の主張を邪魔しない
- 補助語としての役割が明確

## 読むべき reference

- **X 公式仕様と推奨**: [x-articles-spec.md](${CLAUDE_SKILL_DIR}/references/x-articles-spec.md)
- **本文の構成と伸ばし方**: [article-playbook.md](${CLAUDE_SKILL_DIR}/references/article-playbook.md)
- **画像生成の進め方**: [image-generation.md](${CLAUDE_SKILL_DIR}/references/image-generation.md)
- **品質チェック表**: [quality-check.md](${CLAUDE_SKILL_DIR}/references/quality-check.md)

## 完了条件

次がそろって初めて「完了」とする。

- タイトル候補があり、最終採用タイトルの理由が説明できる
- ヘッダー画像と差し込み画像がある
- 画像が Nano Banana Pro で生成されている
- ロゴ参照画像を使った request / response が保存されている
- 品質チェックで必須項目を満たしている
- 記事本文、画像生成プロンプト、品質レポートが同じ `out/social/<slug>/` に保存されている

## 関連スキル

- `x-announce` - 短いポストやスレッドの告知向け。記事そのものではなく、公開告知が必要なときに使う
- `harness-release` - リリース情報の整理が先に必要なときに使う
