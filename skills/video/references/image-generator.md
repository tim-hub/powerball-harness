# Image Generator - Nano Banana Pro 画像自動生成

Nano Banana Pro（Google DeepMind）を使用して、動画シーン用の高品質画像を自動生成します。

---

## 概要

`/generate-video` のシーン生成フェーズで、素材画像が必要と判定された場合に自動実行されます。
2枚生成 → Claude が品質判定 → NG なら再生成、という品質保証ループを実装しています。

## 前提条件

- `GOOGLE_AI_API_KEY` 環境変数が設定済み
- Google AI Studio で Nano Banana Pro（Gemini 3 Pro Image）が有効化済み

---

## API 仕様

### エンドポイント

```
POST https://generativelanguage.googleapis.com/v1beta/models/gemini-3-pro-image:generateImage
```

### 認証

```bash
# Query parameter 方式（推奨）
?key=${GOOGLE_AI_API_KEY}

# または Header 方式
Authorization: Bearer ${GOOGLE_AI_API_KEY}
```

### リクエスト形式

```json
{
  "prompt": "A modern SaaS dashboard interface with clean design, showing analytics charts and user metrics, professional UI mockup, light theme",
  "config": {
    "numberOfImages": 2,
    "aspectRatio": "16:9",
    "outputMimeType": "image/png"
  }
}
```

### レスポンス形式

```json
{
  "images": [
    {
      "mimeType": "image/png",
      "bytesBase64Encoded": "iVBORw0KGgoAAAANS..."
    },
    {
      "mimeType": "image/png",
      "bytesBase64Encoded": "iVBORw0KGgoAAAANS..."
    }
  ]
}
```

---

## 解像度オプション

| 設定 | 解像度 | 用途 | コスト目安 |
|------|--------|------|-----------|
| `1K` | 1024×1024 | プレビュー、テスト | ~$0.02/枚 |
| `2K` | 2048×2048 | 標準品質 | ~$0.06/枚 |
| `4K` | 4096×4096 | 高品質、プロフェッショナル | ~$0.12/枚 |

### アスペクト比

| 比率 | 用途 |
|------|------|
| `16:9` | 動画シーン（推奨） |
| `1:1` | アイコン、ロゴ |
| `9:16` | 縦型動画 |
| `4:3` | プレゼン資料 |

---

## プロンプト設計ガイドライン

### 基本構造

```
[主題] + [スタイル] + [品質指定] + [制約]
```

### シーンタイプ別プロンプトテンプレート

#### イントロ/タイトルシーン

```
Professional product logo and title card for "{product_name}",
modern minimalist design, clean typography,
{brand_color} accent color, dark background,
cinematic quality, 4K render
```

#### UI デモシーン（補助画像）

```
Modern web application interface showing {feature_description},
clean UI design, light theme, subtle shadows,
professional SaaS aesthetic, mockup style,
no text labels, focus on visual hierarchy
```

#### CTA シーン

```
Call-to-action banner for {product_name},
action-oriented design, prominent button,
{brand_color} gradient, professional marketing style,
clear visual hierarchy, engaging composition
```

#### アーキテクチャ/概念図

```
Technical architecture diagram showing {concept},
isometric illustration style, modern tech aesthetic,
clear visual flow, connected components,
professional documentation quality, clean lines
```

### プロンプト品質向上のコツ

| 追加要素 | 効果 |
|---------|------|
| `professional quality` | 全体の品質向上 |
| `clean design` | 不要な要素の削減 |
| `modern aesthetic` | 現代的なデザイン |
| `cinematic lighting` | ドラマチックな照明 |
| `4K render` | 高解像度 |
| `no text` | テキストなし（後で追加する場合） |

### 避けるべきプロンプト

| NG パターン | 理由 |
|------------|------|
| 曖昧な指示 | 「いい感じの画像」→ 結果が不安定 |
| 過度に複雑 | 要素が多すぎると品質低下 |
| テキスト指定 | AI 生成テキストは品質不安定 |
| 著作権物 | ブランドロゴ等は生成不可 |

---

## 実行フロー

```
シーン生成フェーズ
    │
    ├── [Step 1] 素材必要判定
    │   └─ シーンタイプ、既存素材の有無を確認
    │       ├── 素材あり → スキップ
    │       └── 素材なし → Step 2 へ
    │
    ├── [Step 2] プロンプト生成
    │   ├─ シーン情報からプロンプト構築
    │   ├─ ブランド情報（色、スタイル）を反映
    │   └─ テンプレートを適用
    │
    ├── [Step 3] 画像生成（2枚同時）
    │   └─ Nano Banana Pro API 呼び出し
    │       numberOfImages: 2
    │
    ├── [Step 4] 品質判定
    │   └─ → image-quality-check.md 参照
    │
    ├── [Step 5] 結果処理
    │   ├── 成功 → 画像保存、シーンに組み込み
    │   └── 失敗 → Step 6 へ
    │
    └── [Step 6] 再生成ループ（最大3回）
        ├─ プロンプト改善（Claude が提案）
        └─ Step 3 に戻る
```

---

## Bash 実行例

### curl での API 呼び出し

```bash
# 環境変数確認
echo $GOOGLE_AI_API_KEY

# 画像生成リクエスト
curl -X POST \
  "https://generativelanguage.googleapis.com/v1beta/models/gemini-3-pro-image:generateImage?key=${GOOGLE_AI_API_KEY}" \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "Modern SaaS dashboard interface, clean design, light theme, professional UI",
    "config": {
      "numberOfImages": 2,
      "aspectRatio": "16:9",
      "outputMimeType": "image/png"
    }
  }' \
  -o response.json

# Base64 デコードして保存
cat response.json | jq -r '.images[0].bytesBase64Encoded' | base64 -d > out/assets/generated/image_1.png
cat response.json | jq -r '.images[1].bytesBase64Encoded' | base64 -d > out/assets/generated/image_2.png
```

### 画像保存先

```
out/
└── assets/
    └── generated/
        ├── intro_1.png
        ├── intro_2.png
        ├── cta_1.png
        └── cta_2.png
```

---

## 再生成ループ制御

### 最大試行回数

```
max_attempts = 3
```

### 再生成時のプロンプト改善

各試行で Claude がプロンプトを改善:

| 試行 | 改善戦略 |
|------|---------|
| 1回目 | 初期プロンプトで生成 |
| 2回目 | 品質指摘を反映してプロンプト調整 |
| 3回目 | より具体的な指示を追加、スタイル変更 |

### 改善プロンプト生成

```
前回の画像が以下の理由で不採用でした:
- {rejection_reason}

改善案:
1. {improvement_1}
2. {improvement_2}

新しいプロンプト:
{improved_prompt}
```

### 3回失敗時のフォールバック

```
⚠️ 画像生成が3回失敗しました

シーン: {scene_name}
最後のエラー: {last_error}

選択肢:
1. 「続行」→ プレースホルダー画像で進める
2. 「スキップ」→ このシーンを画像なしで生成
3. 「手動」→ ユーザーが画像を提供
```

---

## エラーハンドリング

### API エラー

| エラーコード | 原因 | 対処 |
|-------------|------|------|
| `400` | 不正なプロンプト | プロンプト内容を確認 |
| `401` | 認証失敗 | API キーを確認 |
| `429` | レート制限 | 60秒待機して再試行 |
| `500` | サーバーエラー | 30秒待機して再試行 |

### コンテンツポリシー違反

```
⚠️ コンテンツポリシー違反

プロンプトが Google のポリシーに違反しています。
以下を削除/変更してください:
- {violation_reason}

自動修正を試みますか？ (y/n)
```

### 環境変数未設定

```
⚠️ GOOGLE_AI_API_KEY が設定されていません

設定方法:
1. Google AI Studio でAPIキーを取得
   https://ai.google.dev/aistudio

2. 環境変数に設定
   export GOOGLE_AI_API_KEY="your-api-key"

3. または .env.local に追加
   GOOGLE_AI_API_KEY=your-api-key
```

---

## コスト見積もり

### シーンあたりのコスト

```
基本: 2枚 × $0.12 = $0.24
最大（3回再生成）: 6枚 × $0.12 = $0.72
```

### 動画あたりのコスト目安

| 動画タイプ | シーン数 | 画像生成シーン | コスト目安 |
|-----------|---------|---------------|-----------|
| 90秒ティザー | 5 | 2-3 | $0.48-$0.72 |
| 3分デモ | 8 | 3-4 | $0.72-$0.96 |
| 5分アーキテクチャ | 12 | 4-6 | $0.96-$1.44 |

---

## 関連ドキュメント

- [image-quality-check.md](./image-quality-check.md) - 品質判定ロジック
- [generator.md](./generator.md) - 並列シーン生成エンジン
- [planner.md](./planner.md) - シナリオプランナー
