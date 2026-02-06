---
name: run-task-workers
description: "Execute task-worker agents in parallel for each group. Collects commit_ready/needs_escalation status from each worker."
allowed-tools: ["Read", "Task", "Bash"]
---

# Run Task Workers

依存グラフで決定された並列グループごとに`task-worker`エージェントを起動し、結果を集約するスキル。

---

## 入力

- **parallel_groups**: 依存グラフから取得した並列実行グループ
- **task_files**: 各タスクが編集するファイルのマッピング
- **parallel_count**: 同時実行数の上限（デフォルト: 1）
- **max_iterations**: 改善ループの上限（デフォルト: 3）
- **isolation_mode**: `lock`または`worktree`
- **commit_strategy**: `task` / `phase` / `all`
- **worktree_paths**: `isolation_mode=worktree`時の作業ディレクトリパス（オプション）

---

## 出力

```json
{
  "worker_results": [
    {
      "task_id": "task1",
      "status": "commit_ready",
      "iterations": 2,
      "changes": [
        { "file": "src/components/Header.tsx", "action": "created" }
      ],
      "self_review": {
        "quality": { "grade": "A", "issues": [] },
        "security": { "grade": "A", "issues": [] },
        "performance": { "grade": "A", "issues": [] },
        "compatibility": { "grade": "A", "issues": [] }
      },
      "build_result": "pass",
      "test_result": "pass",
      "metrics": {
        "tokenCount": 2340,
        "toolUses": 12,
        "duration": 45000
      }
    },
    {
      "task_id": "task2",
      "status": "needs_escalation",
      "escalation_reason": "max_iterations_exceeded",
      "context": {
        "attempted_fixes": ["型エラー修正", "import パス修正"],
        "remaining_issues": [
          {
            "file": "src/utils/helpers.ts",
            "line": 42,
            "issue": "型 'unknown' を 'User' に変換できません"
          }
        ]
      },
      "metrics": {
        "tokenCount": 1890,
        "toolUses": 8,
        "duration": 32000
      }
    }
  ],
  "all_ready": false,
  "escalation_count": 1,
  "changed_files": [
    "src/components/Header.tsx",
    "src/components/Footer.tsx"
  ],
  "total_metrics": {
    "tokenCount": 4230,
    "toolUses": 20,
    "duration": 77000
  }
}
```

---

## 実行手順

### Step 1: 並列グループを順次処理

各グループは前のグループ完了後に実行（依存関係を考慮）。

```javascript
// 疑似コード例
for (const group of parallel_groups) {
  if (group.depends_on.length > 0) {
    // 依存グループの完了を待つ
    await waitForGroups(group.depends_on);
  }
  
  // グループ内のタスクを並列実行（parallel_countで上限を制御）
  await executeGroupInParallel(group, parallel_count);
}
```

### Step 2: グループ内タスクの並列起動

各グループ内で`Task`ツールを使用して`task-worker`を並列起動：

```yaml
# Task tool呼び出し例（1つのレスポンス内で複数並列）
Task:
  subagent_type: "task-worker"
  run_in_background: true
  prompt: |
    タスク: {{task_description}}
    対象ファイル: {{task_files[task_id]}}
    max_iterations: {{max_iterations}}
    review_depth: "standard"
```

**重要**: `run_in_background: true`で起動し、`TaskOutput`で結果を収集。

### Step 3: 結果の収集と待機

```javascript
// 疑似コード例
const results = [];
for (const taskId of group.tasks) {
  const taskOutput = await getTaskOutput(taskId);
  results.push({
    task_id: taskId,
    status: taskOutput.status, // "commit_ready" | "needs_escalation" | "failed"
    ...taskOutput
  });
}
```

### Step 3.5: commit_strategy に応じた処理

#### commit_strategy = task
各task-worker完了時にコミットを実行（task単位でcommit）:

```bash
# task-worker内で実行
git add {{task_files[task_id]}}
git commit -m "feat: {{task_id}}"
```

#### commit_strategy = phase / all
このスキルではコミットしない（Phase3のexecute-commitで実行）。

### Step 4: エスカレーションの集約

`needs_escalation`が返されたタスクを集約：

```markdown
⚠️ エスカレーション（{{escalation_count}}件）

{{#each escalation_tasks}}
タスク{{task_id}}: {{escalation_reason}}
  → 提案: {{suggestion}}

{{/each}}

どう対応しますか？
1. 提案を適用して続行
2. スキップして次へ
3. 手動で修正する
```

### Step 5: changed_files の集約

```javascript
// 疑似コード例
const changedFiles = new Set();
for (const result of results) {
  if (result.changes) {
    for (const change of result.changes) {
      changedFiles.add(change.file);
    }
  }
}
```

---

## isolation_mode別の処理

### `isolation_mode=lock`（デフォルト）

- 同一worktreeでファイルロックを使用
- 各task-workerは同じディレクトリで動作
- ファイル競合はロックで防止

### `isolation_mode=worktree`

- 各タスクに`git worktree add`でブランチ作成
- `worktree_paths`が提供されている場合、それを使用
- 各task-workerは独立したworktreeで動作
- 完全な並列ビルド/テストが可能

**worktree使用時のTask tool呼び出し**:

```yaml
Task:
  subagent_type: "task-worker"
  run_in_background: true
  cwd: "{{worktree_paths[task_id]}}"  # worktreeパスを指定
  prompt: |
    タスク: {{task_description}}
    対象ファイル: {{task_files[task_id]}}
    max_iterations: {{max_iterations}}
```

---

## 進捗表示

実行中は以下の形式で進捗を表示：

```markdown
📊 Phase 1 進捗: 2/5 完了

├── [worker-1] Header.tsx ✅ commit_ready (35秒)
├── [worker-2] Footer.tsx ✅ commit_ready (28秒)
├── [worker-3] Sidebar.tsx ⏳ セルフレビュー中...
├── [worker-4] Utils.ts ⏳ 実装中...
└── [worker-5] Types.ts 🔜 待機中（依存: Utils.ts）
```

---

## エラーハンドリング

### 一部タスク失敗時

- 成功したタスクの結果は保持
- 失敗タスクのみ再実行オプションを提示
- `needs_escalation`はユーザー確認待ち

### 全タスク失敗時

- 共通の原因がある可能性を分析
- 依存ファイルの不足などを確認
- 実行順序の見直しを提案

---

## 使用例

### 例1: 3並列実行

**入力**:
```json
{
  "parallel_groups": [
    {
      "group_id": 1,
      "tasks": ["Header作成", "Footer作成", "Sidebar作成"],
      "can_parallelize": true
    }
  ],
  "task_files": {
    "Header作成": ["src/components/Header.tsx"],
    "Footer作成": ["src/components/Footer.tsx"],
    "Sidebar作成": ["src/components/Sidebar.tsx"]
  },
  "parallel_count": 3,
  "max_iterations": 3,
  "isolation_mode": "lock",
  "commit_strategy": "task"
}
```

**実行**:
```
Task tool で3つのtask-workerを並列起動:
- worker-1: Header作成
- worker-2: Footer作成
- worker-3: Sidebar作成

全員の完了を待機 → 結果を集約
```

**出力**:
```json
{
  "worker_results": [
    { "task_id": "Header作成", "status": "commit_ready", ... },
    { "task_id": "Footer作成", "status": "commit_ready", ... },
    { "task_id": "Sidebar作成", "status": "commit_ready", ... }
  ],
  "all_ready": true,
  "escalation_count": 0,
  "changed_files": [
    "src/components/Header.tsx",
    "src/components/Footer.tsx",
    "src/components/Sidebar.tsx"
  ]
}
```

---

## メトリクス集計表示（CC 2.1.30+）

### Task tool メトリクスの取得

Claude Code 2.1.30 以降、Task tool の結果に以下のメトリクスが含まれます：

- `tokenCount` - 消費トークン数
- `toolUses` - ツール使用回数
- `duration` - 実行時間（ミリ秒）

### 集計表示例

並列実行完了時に全ワーカーのメトリクスを集計して表示：

```markdown
🚀 Ultrawork 完了

| Worker | Status | Tokens | Tools | Duration |
|--------|--------|--------|-------|----------|
| task-1 | ✅ | 2,340 | 12 | 45s |
| task-2 | ✅ | 1,890 | 8 | 32s |
| task-3 | ✅ | 3,120 | 15 | 58s |
| **合計** | | **7,350** | **35** | **135s (2m 15s)** |

📊 平均効率:
- トークン/タスク: 2,450
- ツール/タスク: 11.7
- 時間/タスク: 45s
```

### 実装例

```javascript
// メトリクス集計関数
function aggregateMetrics(workerResults) {
  const totals = { tokenCount: 0, toolUses: 0, duration: 0 };
  let count = 0;

  for (const result of workerResults) {
    if (result.metrics) {
      totals.tokenCount += result.metrics.tokenCount || 0;
      totals.toolUses += result.metrics.toolUses || 0;
      totals.duration += result.metrics.duration || 0;
      count++;
    }
  }

  return {
    totals,
    averages: {
      tokenCount: Math.round(totals.tokenCount / count),
      toolUses: (totals.toolUses / count).toFixed(1),
      duration: Math.round(totals.duration / count)
    }
  };
}

// 表示用フォーマット
function formatDuration(ms) {
  if (ms < 60000) return `${Math.round(ms / 1000)}s`;
  const minutes = Math.floor(ms / 60000);
  const seconds = Math.round((ms % 60000) / 1000);
  return `${minutes}m ${seconds}s`;
}

function formatNumber(num) {
  return num.toLocaleString();
}
```

### AgentTrace への記録

メトリクスは `.claude/state/agent-trace.jsonl` に自動記録されます（emit-agent-trace.js による）：

```json
{
  "version": "0.3.0",
  "id": "uuid",
  "timestamp": "2026-02-04T...",
  "tool": "Task",
  "metrics": {
    "tokenCount": 2340,
    "toolUses": 12,
    "duration": 45000
  },
  "metadata": {
    "project": "my-project",
    "taskId": "task-1"
  }
}
```

詳細: [docs/AGENT_TRACE_SCHEMA.md](../../../docs/AGENT_TRACE_SCHEMA.md)

---

## 注意事項

- **並列数の上限**: `parallel_count`で指定された数を超えない（デフォルト: 1、上限: 10）
- **依存グループの待機**: 依存グループが完了するまで次のグループは実行しない
- **エスカレーション処理**: `needs_escalation`が返された場合、ユーザー確認を待つ
- **worktreeクリーンアップ**: `isolation_mode=worktree`時は後でworktreeを削除する必要がある（Phase3でマージ後）
- **メトリクス収集**: CC 2.1.30+ では Task tool の結果に自動的にメトリクスが含まれる（古いバージョンでは null）
