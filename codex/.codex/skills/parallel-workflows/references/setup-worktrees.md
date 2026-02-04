---
name: setup-worktrees
description: "Create git worktrees for parallel task execution. Each task gets its own worktree branch for complete isolation."
allowed-tools: ["Bash", "Read"]
---

# Setup Worktrees

`--isolation=worktree`指定時に、各タスク用のgit worktreeを作成するスキル。

---

## 入力

- **parallel_groups**: 依存グラフから取得した並列実行グループ
- **task_files**: 各タスクが編集するファイルのマッピング
- **base_branch**: ベースブランチ（デフォルト: `main`または`master`）

---

## 出力

```json
{
  "worktree_paths": {
    "task1": "/path/to/project/.worktrees/task1-worktree",
    "task2": "/path/to/project/.worktrees/task2-worktree",
    "task3": "/path/to/project/.worktrees/task3-worktree"
  },
  "worktree_task_slugs": {
    "task1": "task1",
    "task2": "task2",
    "task3": "task3"
  },
  "worktree_branches": [
    "worktree/task1",
    "worktree/task2",
    "worktree/task3"
  ],
  "worktree_branch_map": {
    "task1": "worktree/task1",
    "task2": "worktree/task2",
    "task3": "worktree/task3"
  }
}
```

---

## 実行手順

### Step 1: ベースブランチの確認

```bash
# 現在のブランチを確認
CURRENT_BRANCH=$(git branch --show-current)

# デフォルトブランチを確認（main または master）
if git show-ref --verify --quiet refs/heads/main; then
  BASE_BRANCH="main"
elif git show-ref --verify --quiet refs/heads/master; then
  BASE_BRANCH="master"
else
  BASE_BRANCH="$CURRENT_BRANCH"
fi
```

### Step 2: worktreeディレクトリの準備

```bash
# .worktreesディレクトリを作成（存在しない場合）
mkdir -p .worktrees

# .gitignoreに追加（既に追加されていない場合）
if ! grep -q "^\.worktrees$" .gitignore 2>/dev/null; then
  echo ".worktrees" >> .gitignore
fi
```

### Step 3: 各タスク用のworktreeを作成

**安全なブランチ名の生成**:
タスクIDにスペース/記号/日本語が含まれる場合は、ブランチ名として使えるスラッグに変換する。

```bash
# 例: "Header 作成" → "header-sakusei"
# 例: "T001:認証" → "t001-auth"
slugify() {
  echo "$1" | tr '[:upper:]' '[:lower:]' | sed -E 's/[^a-z0-9]+/-/g' | sed -E 's/^-+|-+$//g'
}
```

```bash
# 各タスクに対して
for task_id in "${task_ids[@]}"; do
  TASK_SLUG=$(slugify "$task_id")
  WORKTREE_DIR=".worktrees/${TASK_SLUG}-worktree"
  WORKTREE_BRANCH="worktree/${TASK_SLUG}"
  
  # ブランチを作成（ベースブランチから）
  git branch "$WORKTREE_BRANCH" "$BASE_BRANCH" 2>/dev/null || true
  
  # worktreeを作成
  git worktree add "$WORKTREE_DIR" "$WORKTREE_BRANCH"
  
  echo "Created worktree: $WORKTREE_DIR (branch: $WORKTREE_BRANCH)"
done
```

### Step 4: pnpm使用時の容量節約（オプション）

pnpmを使用している場合、shared storeを活用：

```bash
# pnpmの場合、node_modulesはシンボリックリンクで共有される
# 各worktreeでpnpm installを実行しても容量は+54MB程度
cd "$WORKTREE_DIR"
pnpm install  # shared storeを使用
```

---

## 出力フォーマット

### worktree_paths

各タスクIDに対応するworktreeディレクトリのパス：

```json
{
  "task1": "/absolute/path/to/project/.worktrees/task1-worktree",
  "task2": "/absolute/path/to/project/.worktrees/task2-worktree"
}
```

### worktree_task_slugs

タスクIDとスラッグの対応表：

```json
{
  "Header 作成": "header-sakusei",
  "T001:認証": "t001-auth"
}
```

### worktree_branches

各タスクIDに対応するworktreeブランチ名：

```json
[
  "worktree/task1",
  "worktree/task2"
]
```

### worktree_branch_map（オプション）

タスクIDとブランチ名の対応表：

```json
{
  "task1": "worktree/task1",
  "task2": "worktree/task2"
}
```

---

## 使用例

### 例1: 3タスクのworktree作成

**入力**:
```json
{
  "parallel_groups": [
    {
      "group_id": 1,
      "tasks": ["Header作成", "Footer作成", "Sidebar作成"]
    }
  ],
  "task_files": {
    "Header作成": ["src/components/Header.tsx"],
    "Footer作成": ["src/components/Footer.tsx"],
    "Sidebar作成": ["src/components/Sidebar.tsx"]
  }
}
```

**実行**:
```bash
# Header作成用
git worktree add .worktrees/Header作成-worktree worktree/Header作成

# Footer作成用
git worktree add .worktrees/Footer作成-worktree worktree/Footer作成

# Sidebar作成用
git worktree add .worktrees/Sidebar作成-worktree worktree/Sidebar作成
```

**出力**:
```json
{
  "worktree_paths": {
    "Header作成": "/path/to/project/.worktrees/Header作成-worktree",
    "Footer作成": "/path/to/project/.worktrees/Footer作成-worktree",
    "Sidebar作成": "/path/to/project/.worktrees/Sidebar作成-worktree"
  },
  "worktree_branches": {
    "Header作成": "worktree/Header作成",
    "Footer作成": "worktree/Footer作成",
    "Sidebar作成": "worktree/Sidebar作成"
  }
}
```

---

## 注意事項

- **既存worktreeの確認**: 同じタスクIDのworktreeが既に存在する場合は削除してから作成
- **ブランチ名の衝突**: `worktree/{task_id}`形式のブランチ名が既に存在する場合は番号を付与
- **.gitignoreへの追加**: `.worktrees`ディレクトリは`.gitignore`に追加する（既に追加されていない場合）
- **pnpm使用時**: pnpmのshared storeにより、各worktreeの`node_modules`は約54MBで済む
- **クリーンアップ**: Phase3完了後、`merge-worktrees`スキルでマージし、その後worktreeを削除する

---

## トラブルシューティング

### Q: worktree作成時にエラーが発生

**A**: 既存のworktreeやブランチを確認：

```bash
# 既存worktree一覧
git worktree list

# 既存ブランチ一覧
git branch -a | grep worktree

# 必要に応じて削除
git worktree remove .worktrees/task1-worktree
git branch -D worktree/task1
```

### Q: pnpm installが遅い

**A**: shared storeが正しく設定されているか確認：

```bash
# pnpm store pathを確認
pnpm store path

# 各worktreeで同じstoreを使用していることを確認
cd .worktrees/task1-worktree
pnpm store path  # 同じパスが表示されるはず
```
