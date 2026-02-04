---
name: merge-worktrees
description: "Merge worktree branches back to main branch before Phase3 commit. Detects conflicts and reports to user."
allowed-tools: ["Bash", "Read", "Grep"]
---

# Merge Worktrees

Phase3（コミット実行）の前に、各worktreeブランチをメインブランチにマージするスキル。

---

## 入力

- **worktree_branches**: 各タスクIDに対応するworktreeブランチ名
- **base_branch**: ベースブランチ（デフォルト: `main`または`master`）
- **worker_results**: Phase1のtask-worker実行結果
- **commit_strategy**: `task` / `phase` / `all`

---

## 出力

```json
{
  "merged": true,
  "merged_branches": ["worktree/task1", "worktree/task2"],
  "conflicts": [],
  "conflict_files": []
}
```

または、競合がある場合：

```json
{
  "merged": false,
  "merged_branches": [],
  "conflicts": [
    {
      "branch": "worktree/task1",
      "files": ["src/components/Layout.tsx"],
      "reason": "同一行への変更"
    }
  ],
  "conflict_files": ["src/components/Layout.tsx"]
}
```

---

## 実行手順

### Step 1: ベースブランチの確認とチェックアウト

```bash
# 現在のブランチを確認
CURRENT_BRANCH=$(git branch --show-current)

# ベースブランチを確認
if git show-ref --verify --quiet refs/heads/main; then
  BASE_BRANCH="main"
elif git show-ref --verify --quiet refs/heads/master; then
  BASE_BRANCH="master"
else
  BASE_BRANCH="$CURRENT_BRANCH"
fi

# ベースブランチにチェックアウト
git checkout "$BASE_BRANCH"
```

### Step 2: 各worktreeブランチを順次マージ

```bash
CONFLICTS=()

for branch in "${worktree_branches[@]}"; do
  echo "Merging $branch into $BASE_BRANCH..."
  
  # マージを試行
  if [ "$commit_strategy" = "task" ]; then
    # task: 既存のコミットを保持するため通常マージ
    if git merge --no-ff --no-edit "$branch"; then
      echo "✅ Merged $branch (preserve commits)"
    else
      CONFLICTS+=("$branch")
      echo "⚠️ Conflict detected in $branch"
      CONFLICT_FILES=$(git diff --name-only --diff-filter=U)
      echo "Conflict files: $CONFLICT_FILES"
      git merge --abort
    fi
  else
    # phase/all: 単一コミットにまとめるためsquashマージ
    if git merge --squash "$branch"; then
      echo "✅ Squashed $branch (no commit)"
    else
      CONFLICTS+=("$branch")
      echo "⚠️ Conflict detected in $branch"
      CONFLICT_FILES=$(git diff --name-only --diff-filter=U)
      echo "Conflict files: $CONFLICT_FILES"
      git merge --abort
    fi
  fi
done
```

### Step 3: 3-way mergeの自動解消試行

競合が発生した場合、自動解消を試行：

```bash
# 競合ファイルごとに処理
for file in $CONFLICT_FILES; do
  # 変更内容を確認
  git diff "$BASE_BRANCH" "$branch" -- "$file"
  
  # 自動マージ可能か判定
  # - 異なる行への変更 → 自動マージ可能
  # - 同一行への変更 → 手動解決必要
done
```

### Step 4: 競合レポートの生成

競合がある場合、ユーザーに報告：

```markdown
⚠️ マージ競合が検出されました

以下のブランチで競合が発生しています：

| ブランチ | 競合ファイル | 理由 |
|---------|-------------|------|
| worktree/task1 | src/components/Layout.tsx | 同一行への変更 |
| worktree/task2 | src/utils/helpers.ts | 同一行への変更 |

**対応方法**:
1. 手動で競合を解決してから続行
2. 該当タスクをスキップして続行
3. マージを中止して最初からやり直す

どれを選択しますか？
```

---

## 競合検出ロジック

### 自動マージ可能なケース

1. **異なるファイルへの変更**: タスクが異なるファイルを編集
2. **異なる行への変更**: 同じファイルでも異なる行を編集
3. **片方が削除、片方が追加**: 一方が削除、他方が追加

### 手動解決が必要なケース

1. **同一行への変更**: 複数のタスクが同じ行を変更
2. **依存関係の競合**: タスクAが作成したファイルをタスクBが削除
3. **構造的な競合**: 同じ関数/クラスを複数タスクが変更

---

## 使用例

### 例1: 競合なしのマージ

**入力**:
```json
{
  "worktree_branches": [
    "worktree/Header作成",
    "worktree/Footer作成",
    "worktree/Sidebar作成"
  ],
  "base_branch": "main"
}
```

**実行**:
```bash
git checkout main
git merge --no-commit --no-ff worktree/Header作成
git commit -m "Merge worktree/Header作成 into main"

git merge --no-commit --no-ff worktree/Footer作成
git commit -m "Merge worktree/Footer作成 into main"

git merge --no-commit --no-ff worktree/Sidebar作成
git commit -m "Merge worktree/Sidebar作成 into main"
```

**出力**:
```json
{
  "merged": true,
  "merged_branches": [
    "worktree/Header作成",
    "worktree/Footer作成",
    "worktree/Sidebar作成"
  ],
  "conflicts": [],
  "conflict_files": []
}
```

### 例2: 競合ありのマージ

**入力**:
```json
{
  "worktree_branches": [
    "worktree/Layout作成",
    "worktree/Page作成"
  ],
  "base_branch": "main"
}
```

**状況**: `Layout作成`と`Page作成`が両方`src/app/layout.tsx`の同じ行を変更

**実行**:
```bash
git checkout main
git merge --no-commit --no-ff worktree/Layout作成
# ✅ 成功

git merge --no-commit --no-ff worktree/Page作成
# ⚠️ 競合発生: src/app/layout.tsx
git merge --abort
```

**出力**:
```json
{
  "merged": false,
  "merged_branches": ["worktree/Layout作成"],
  "conflicts": [
    {
      "branch": "worktree/Page作成",
      "files": ["src/app/layout.tsx"],
      "reason": "同一行への変更"
    }
  ],
  "conflict_files": ["src/app/layout.tsx"]
}
```

---

## マージ後のクリーンアップ

マージ完了後、worktreeを削除：

```bash
# 各worktreeを削除
for task_id in "${task_ids[@]}"; do
  WORKTREE_DIR=".worktrees/${task_id}-worktree"
  WORKTREE_BRANCH="worktree/${task_id}"
  
  # worktreeを削除
  git worktree remove "$WORKTREE_DIR"
  
  # ブランチを削除（オプション）
  git branch -D "$WORKTREE_BRANCH"
done

# 不要なworktreeをクリーンアップ
git worktree prune
```

---

## 注意事項

- **マージ順序**: 依存関係を考慮してマージ順序を決定（依存グラフの順序に従う）
- **競合時の処理**: 競合が発生した場合、ユーザーに手動解決を依頼する
- **worktree削除**: マージ完了後にworktreeを削除してクリーンアップする
- **3-way merge**: 可能な限り3-way mergeで自動解消を試みる
- **コミットメッセージ**: マージコミットには適切なメッセージを付与する
