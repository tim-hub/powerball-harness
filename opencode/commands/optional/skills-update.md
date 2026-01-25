---
description: "スキル設定を更新（add/remove/list/enable/disable）。Skills Gate の対象スキルを管理"
---

# /skills-update

プロジェクトの Skills Gate 設定を更新します。

## 使い方

```
/skills-update [action] [skill_name]
```

### アクション

| アクション | 説明 | 例 |
|-----------|------|-----|
| `list` | 現在の設定を表示 | `/skills-update list` |
| `add <skill>` | スキルを追加 | `/skills-update add auth` |
| `remove <skill>` | スキルを削除 | `/skills-update remove auth` |
| `enable` | Skills Gate を有効化 | `/skills-update enable` |
| `disable` | Skills Gate を無効化 | `/skills-update disable` |

## 実行手順

### 1. 引数を解析

ユーザーの入力から action と skill_name を抽出します。

### 2. skills-config.json を更新

```bash
STATE_DIR=".claude/state"
SKILLS_CONFIG="${STATE_DIR}/skills-config.json"

mkdir -p "$STATE_DIR"

# 存在しない場合は初期化
if [ ! -f "$SKILLS_CONFIG" ]; then
  echo '{"enabled": true, "skills": ["impl", "review"]}' > "$SKILLS_CONFIG"
fi
```

### 3. アクション別処理

#### list
```bash
cat "$SKILLS_CONFIG" | jq .
```

#### add <skill>
```bash
jq --arg s "$SKILL_NAME" '.skills += [$s] | .skills |= unique' "$SKILLS_CONFIG" > tmp && mv tmp "$SKILLS_CONFIG"
```

#### remove <skill>
```bash
jq --arg s "$SKILL_NAME" '.skills -= [$s]' "$SKILLS_CONFIG" > tmp && mv tmp "$SKILLS_CONFIG"
```

#### enable
```bash
jq '.enabled = true' "$SKILLS_CONFIG" > tmp && mv tmp "$SKILLS_CONFIG"
```

#### disable
```bash
jq '.enabled = false' "$SKILLS_CONFIG" > tmp && mv tmp "$SKILLS_CONFIG"
```

### 4. 結果を報告

更新後の設定を表示し、変更内容をユーザーに伝えます。

## 利用可能なスキル

| スキル | 用途 |
|--------|------|
| impl | 実装、機能追加 |
| review | コードレビュー |
| ui | UI コンポーネント |
| auth | 認証・決済 |
| deploy | デプロイ |
| ci | CI/CD |
| verify | ビルド検証 |
| docs | ドキュメント生成 |

## 注意事項

- Skills Gate が有効な場合、コード編集前に Skill ツールの使用が必要
- 設定は `.claude/state/skills-config.json` に保存
- セッション再起動後に反映
