# Codex Mode

> **Status: Experimental**
>
> この機能は実装済みですが、実験段階です。本番環境での使用前にテストを推奨します。

`/ultrawork --codex` で Codex Worker を使った並列実行モード。

## Overview

```
Claude Code (Orchestrator)
    │
    ├─ タスク分析 & グループ化
    │     - Plans.md からタスク取得
    │     - owns: アノテーションで担当ファイル特定
    │     - 重複なし → 並列グループ
    │     - 重複あり → 順次グループ
    │
    ├─ 並列実行
    │     ├─ Worker A ─→ ../worktrees/worker-a ─→ タスク A
    │     ├─ Worker B ─→ ../worktrees/worker-b ─→ タスク B
    │     └─ Worker C ─→ ../worktrees/worker-c ─→ タスク C
    │
    ├─ 品質ゲート
    │     - AGENTS_SUMMARY 証跡検証
    │     - lint / test 実行
    │     - 改ざん検出
    │
    └─ マージ & 完了
          - cherry-pick / merge
          - Plans.md 更新
          - Worktree クリーンアップ
```

## Claude の役割（PM モード）

`--codex` モード時、Claude は **PM（Project Manager）** として機能します。

### 許可される操作

| 操作 | 許可 | 説明 |
|------|------|------|
| ファイル読み込み | ✅ | Read, Glob, Grep |
| Codex Worker 呼び出し | ✅ | `mcp__codex__codex` |
| レビューと判定 | ✅ | 品質ゲート、証跡検証 |
| Plans.md 更新 | ✅ | 状態マーカーの更新のみ |
| Edit/Write | ❌ | **禁止**（pretooluse-guard でブロック） |
| Bash（読み取り系） | ✅ | cat, grep, ls, git status 等 |
| Bash（書き込み系） | ⚠️ | リダイレクト/tee/sed -i はブロック、mv/cp/rm は確認 |

### 保証範囲

| 対象 | 保証レベル | 説明 |
|------|-----------|------|
| Edit/Write | **厳格** | Plans.md 以外は完全にブロック |
| Bash | **ヒューリスティック** | 主要な書き込みパターンを検出してブロック/確認 |

> **Note**: Bash の制限はブラックリスト方式のため、`python -c`, `git apply` 等一部のコマンドはすり抜ける可能性があります。厳密な書き込み禁止が必要な場合は、Codex Worker を使用してください。

### 初期化時の設定

`--codex` フラグ指定時、`ultrawork-active.json` に `codex_mode: true` を設定：

```json
{
  "active": true,
  "started_at": "2025-02-06T10:00:00Z",
  "codex_mode": true,
  "bypass_guards": ["rm_rf", "git_push"],
  "allowed_rm_paths": ["node_modules", "dist", ".cache"]
}
```

**重要**: `codex_mode: true` が設定されている間、Claude の Edit/Write は `pretooluse-guard.sh` によってブロックされます。すべての実装は `mcp__codex__codex` 経由で Codex Worker に委譲してください。

## Usage

```bash
# 基本（並列数はデフォルト 3）
/ultrawork --codex 全部やって

# 並列数を指定
/ultrawork --codex --parallel 5 認証機能から完了して

# シーケンシャル実行（並列なし）
/ultrawork --codex --parallel 1 ログイン機能を実装
```

## Options

| オプション | 説明 | デフォルト |
|-----------|------|-----------|
| `--codex` | Codex Worker モードを有効化 | - |
| `--parallel N` | 並列 Worker 数 | 3 |
| `--worktree-base PATH` | Worktree 作成先 | `../worktrees` |

## Prerequisites

1. **Codex CLI**: `codex --version` >= 0.92.0
2. **MCP 登録**: `claude mcp list` に codex が含まれる
3. **Git worktree**: `git --version` >= 2.5.0

セットアップ確認:
```bash
./scripts/codex-worker-setup.sh --check-only
```

## Execution Flow

### Step 0: Worktree 必要性判定（必須）

`--codex` 指定時、まず **Worktree が本当に必要か** を判定する。

```
タスク分析
    ↓
┌─────────────────────────────────────────────┐
│ Worktree 必要性判定                          │
├─────────────────────────────────────────────┤
│  1. タスク数を確認                           │
│  2. owns: アノテーションから所有ファイル取得  │
│  3. ファイル重複チェック                     │
│  4. 並列実行可能性を判定                     │
└─────────────────────────────────────────────┘
    ↓
├── Worktree 使用 → Step 1〜8 実行
└── 直接実行 → 通常の /ultrawork フロー
```

#### 判定基準

| 条件 | Worktree | 理由 |
|------|----------|------|
| タスク 1 つのみ | ❌ 不要 | 並列の意味がない |
| 全タスクが順次依存 | ❌ 不要 | 結局直列実行になる |
| owns: が全て重複 | ❌ 不要 | 同じファイルを触るため並列不可 |
| 変更予定ファイル < 5 | ❌ 不要 | Worktree オーバーヘッド > 効果 |
| 並列可能タスク 2+ & ファイル分離 | ✅ 使用 | Worktree の価値あり |

#### 判定アルゴリズム

```javascript
function shouldUseWorktree(tasks, parallelCount) {
  // 1. タスク数チェック
  if (tasks.length < 2) return { use: false, reason: "single_task" };

  // 2. 並列数チェック
  if (parallelCount < 2) return { use: false, reason: "sequential_mode" };

  // 3. owns: からファイルパターンを抽出
  const ownershipMap = tasks.map(t => ({
    id: t.id,
    owns: extractOwnsPatterns(t)
  }));

  // 4. ファイル重複チェック
  const parallelGroups = groupByNoOverlap(ownershipMap);
  if (parallelGroups.length < 2) {
    return { use: false, reason: "all_files_overlap" };
  }

  // 5. 変更ファイル数チェック
  const totalFiles = countUniqueFiles(ownershipMap);
  if (totalFiles < 5) {
    return { use: false, reason: "small_changeset" };
  }

  return { use: true, parallelGroups };
}
```

#### 出力例

**Worktree 使用時:**
```
📊 Worktree 判定結果

タスク数: 5
並列可能グループ: 3
ファイル重複: なし
変更予定ファイル: 12

→ ✅ Worktree モードで実行（3 Worker 並列）
```

**直接実行時:**
```
📊 Worktree 判定結果

タスク数: 2
並列可能グループ: 0（全て順次依存）
理由: all_files_overlap

→ ❌ 直接実行モード（Worktree スキップ）
   通常の /ultrawork フローで実行します
```

---

### Step 1: タスク分析

```json
{
  "tasks": [
    {"id": "task-1", "owns": ["src/auth/*"], "group": 1},
    {"id": "task-2", "owns": ["src/api/*"], "group": 1},
    {"id": "task-3", "owns": ["src/auth/*"], "group": 2}
  ],
  "groups": [
    {"id": 1, "parallel": true, "tasks": ["task-1", "task-2"]},
    {"id": 2, "parallel": false, "tasks": ["task-3"]}
  ]
}
```

### Step 2: Worktree 作成

```bash
git worktree add ../worktrees/worker-1 HEAD
git worktree add ../worktrees/worker-2 HEAD
```

### Step 3: ロック取得

```bash
./scripts/codex-worker-lock.sh acquire --path "src/auth/*" --worker worker-1
./scripts/codex-worker-lock.sh acquire --path "src/api/*" --worker worker-2
```

### Step 4: 並列 Worker 実行

各 Worker は独立した worktree で動作:

```bash
# Worker 1
./scripts/codex-worker-engine.sh --task "タスク1" --worktree ../worktrees/worker-1

# Worker 2 (並列)
./scripts/codex-worker-engine.sh --task "タスク2" --worktree ../worktrees/worker-2
```

MCP 呼び出し:
```json
{
  "prompt": "タスク内容 + AGENTS_SUMMARY 証跡出力指示",
  "base-instructions": "Rules 連結 + AGENTS.md 強制読み込み指示",
  "cwd": "/path/to/worktree",
  "approval-policy": "never",
  "sandbox": "workspace-write"
}
```

### Step 5: 証跡検証

各 Worker の出力に AGENTS_SUMMARY があることを確認:

```
AGENTS_SUMMARY: タスク1の実装を完了 | HASH:a1b2c3d4
```

- 正規表現: `/AGENTS_SUMMARY:\s*(.+?)\s*\|\s*HASH:([A-Fa-f0-9]{8})/`
- 証跡欠落 → 即失敗（再試行なし、手動対応）
- ハッシュ不一致 → 再試行（最大 3 回）

### Step 6: 品質ゲート

```bash
npm run lint
npm test
```

失敗時:
- lint エラー → 自動修正指示 → 再実行
- テスト失敗 → 修正指示 → 再実行（最大 3 回）
- 改ざん検出 → 即座に中断 → 手動対応

### Step 7: マージ

```bash
# Worktree でコミット
cd ../worktrees/worker-1
git add .
git commit -m "feat: タスク1の実装"

# メインブランチに cherry-pick
cd /original/repo
git cherry-pick <commit-hash>
```

### Step 8: クリーンアップ

```bash
./scripts/codex-worker-lock.sh release --path "src/auth/*" --worker worker-1
git worktree remove ../worktrees/worker-1
git worktree prune
```

## Error Handling

### Worker 失敗時

- エラーログ記録
- Worktree を保持（デバッグ用）
- 他の Worker は継続
- 失敗タスクは後で再実行

### 競合発生時

1. 競合を検出
2. ユーザーに通知
3. 手動解決 or 自動リトライ選択
4. 解決後、残りの Worker を再開

## Monitoring

### Progress Display

```
🚀 Parallel Execution Started (--codex mode)
├─ Worker 1: ログイン機能 [████████░░] 80%
├─ Worker 2: API実装     [██████████] 100% ✅
└─ Worker 3: UI作成      [██░░░░░░░░] 20%

Completed: 1/3 | In Progress: 2 | Failed: 0
```

### Log Files

- `.claude/state/parallel-execution.log` - 並列実行ログ
- `.claude/state/locks.log` - ロック操作ログ
- `.claude/state/codex-worker/` - 各 Worker のパラメータ

## Related

- [codex-worker skill](../../codex-worker/SKILL.md) - 単体 Worker 実行
- [parallel-strategy.md](../../codex-worker/references/parallel-strategy.md) - 並列戦略詳細
- [task-ownership.md](../../codex-worker/references/task-ownership.md) - ロック機構詳細
