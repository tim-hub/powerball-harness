---
name: build-dependency-graph
description: "Build dependency graph from Plans.md tasks to determine parallel execution groups. Analyzes task descriptions to extract target files and infer dependencies."
allowed-tools: ["Read", "Grep", "Glob"]
---

# Build Dependency Graph

Plans.mdのタスクから依存関係を解析し、並列実行可能なグループを決定するスキル。

---

## 入力

- **Plans.md**: タスク一覧を含むファイル
- **target_tasks**: 実行対象のタスクリスト（plans-managementスキルから取得）

---

## 出力

```json
{
  "parallel_groups": [
    {
      "group_id": 1,
      "tasks": ["task1", "task2", "task3"],
      "can_parallelize": true,
      "reason": "独立したファイルを編集"
    },
    {
      "group_id": 2,
      "tasks": ["task4"],
      "can_parallelize": false,
      "reason": "group1の出力に依存",
      "depends_on": [1]
    }
  ],
  "task_files": {
    "task1": ["src/components/Header.tsx"],
    "task2": ["src/components/Footer.tsx"],
    "task3": ["src/components/Sidebar.tsx"],
    "task4": ["src/components/Layout.tsx"]
  }
}
```

---

## 依存関係判定ルール

### 並列化可能な条件

1. **別ファイルを編集**: タスクが異なるファイルを編集する
2. **データ依存なし**: タスクAの出力がタスクBの入力にならない
3. **順序依存なし**: 実行順序が結果に影響しない

### 並列化不可の条件

1. **同一ファイル編集**: 複数タスクが同じファイルを編集する
2. **import依存**: タスクAが作成するファイルをタスクBがimportする
3. **明示的な依存**: Plans.mdに`depends:T001`などの依存記法がある

---

## 実行手順

### Step 1: Plans.mdからタスクと対象ファイルを抽出

```bash
# Plans.mdを読み込む
cat Plans.md

# タスク行を抽出（マーカー付き）
grep -E "^\s*- \[ \].*`cc:(TODO|WIP)`" Plans.md

# タスク説明からファイルパスを抽出
# パターン例:
# - "src/components/Header.tsx を作成"
# - "Header.tsx を実装"
# - "関連ファイル: src/components/Header.tsx"
```

### Step 2: ファイル依存関係の推定

#### 2.1: ファイル名から依存を推定

```bash
# 各タスクの対象ファイルをリスト化
# 例:
# task1: ["src/components/Header.tsx"]
# task2: ["src/components/Footer.tsx"]
# task3: ["src/components/Layout.tsx"]  # Header.tsx, Footer.tsx を import する可能性
```

#### 2.2: import文の確認（既存ファイルがある場合）

```bash
# Layout.tsxが既に存在する場合、import文を確認
grep -E "import.*from.*['\"](\.\.?/.*Header|\.\.?/.*Footer)" src/components/Layout.tsx

# 存在する場合、Layout作成タスクはHeader/Footer作成タスクに依存
```

#### 2.3: Plans.mdの依存記法を確認

```bash
# depends:記法を抽出
grep -E "depends:[T0-9,]+" Plans.md

# 例: "depends:T001,T002" → タスクIDで依存関係を特定
```

### Step 3: 依存グラフの構築

```
判定ロジック:
├── 同一ファイル編集 → 競合 → 直列実行
├── import依存あり → 依存 → A→B の順序
├── depends:記法あり → 依存 → 指定順序
└── 互いに独立 → 並列可能 → 同一グループ
```

### Step 4: 並列グループの決定

```javascript
// 疑似コード例
const groups = [];
let currentGroup = { group_id: 1, tasks: [], depends_on: [] };

for (const task of target_tasks) {
  if (hasDependency(task, currentGroup.tasks)) {
    // 依存がある場合は新しいグループ
    groups.push(currentGroup);
    currentGroup = { group_id: groups.length + 1, tasks: [task], depends_on: [groups.length] };
  } else {
    // 依存がない場合は現在のグループに追加
    currentGroup.tasks.push(task);
  }
}
groups.push(currentGroup);
```

---

## 出力フォーマット

### parallel_groups

各グループの構造：

```json
{
  "group_id": 1,
  "tasks": ["task1", "task2", "task3"],
  "can_parallelize": true,
  "reason": "独立したファイルを編集",
  "depends_on": []
}
```

| フィールド | 説明 |
|-----------|------|
| `group_id` | グループ番号（実行順序） |
| `tasks` | このグループに含まれるタスクIDのリスト |
| `can_parallelize` | グループ内で並列実行可能か |
| `reason` | 判定理由（デバッグ用） |
| `depends_on` | 依存しているグループIDのリスト |

### task_files

各タスクが編集するファイルのマッピング：

```json
{
  "task1": ["src/components/Header.tsx"],
  "task2": ["src/components/Footer.tsx"]
}
```

---

## 使用例

### 例1: 独立したコンポーネント作成

**Plans.md**:
```markdown
- [ ] Headerコンポーネント作成 `cc:TODO`
  - 関連ファイル: src/components/Header.tsx
- [ ] Footerコンポーネント作成 `cc:TODO`
  - 関連ファイル: src/components/Footer.tsx
- [ ] Sidebarコンポーネント作成 `cc:TODO`
  - 関連ファイル: src/components/Sidebar.tsx
```

**出力**:
```json
{
  "parallel_groups": [
    {
      "group_id": 1,
      "tasks": ["Header作成", "Footer作成", "Sidebar作成"],
      "can_parallelize": true,
      "reason": "独立したファイルを編集",
      "depends_on": []
    }
  ],
  "task_files": {
    "Header作成": ["src/components/Header.tsx"],
    "Footer作成": ["src/components/Footer.tsx"],
    "Sidebar作成": ["src/components/Sidebar.tsx"]
  }
}
```

### 例2: 依存関係あり

**Plans.md**:
```markdown
- [ ] Headerコンポーネント作成 `cc:TODO`
  - 関連ファイル: src/components/Header.tsx
- [ ] Footerコンポーネント作成 `cc:TODO`
  - 関連ファイル: src/components/Footer.tsx
- [ ] Layoutコンポーネント作成 `cc:TODO` depends:Header,Footer
  - 関連ファイル: src/components/Layout.tsx
```

**出力**:
```json
{
  "parallel_groups": [
    {
      "group_id": 1,
      "tasks": ["Header作成", "Footer作成"],
      "can_parallelize": true,
      "reason": "独立したファイルを編集",
      "depends_on": []
    },
    {
      "group_id": 2,
      "tasks": ["Layout作成"],
      "can_parallelize": false,
      "reason": "group1の出力に依存",
      "depends_on": [1]
    }
  ],
  "task_files": {
    "Header作成": ["src/components/Header.tsx"],
    "Footer作成": ["src/components/Footer.tsx"],
    "Layout作成": ["src/components/Layout.tsx"]
  }
}
```

---

## 注意事項

- **ファイルパス抽出の精度**: タスク説明から正確にファイルパスを抽出できない場合、`task_files`が空になる可能性がある
- **既存ファイルの確認**: import依存を正確に判定するには、既存ファイルのimport文を確認する必要がある
- **依存記法の優先**: Plans.mdに`depends:`記法がある場合、それを最優先で使用する
- **安全側に倒す**: 依存関係が不明確な場合は並列化しない（安全側に倒す）
