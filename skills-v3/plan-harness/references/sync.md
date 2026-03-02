# sync サブコマンド — 進捗同期フロー

実装状況と Plans.md を照合し、差分を検出・更新する。

## Step 1: 現状収集（並列）

```bash
# Plans.md の状態
cat Plans.md

# Git 変更状態
git status
git diff --stat HEAD~3

# 直近コミット履歴
git log --oneline -10

# エージェントトレース（直近の編集ファイル）
tail -20 .claude/state/agent-trace.jsonl 2>/dev/null | jq -r '.files[].path' | sort -u
```

## Step 1.5: Agent Trace 分析

Agent Trace から直近の編集履歴を取得し、Plans.md のタスクと照合する:

```bash
# 直近の編集ファイル一覧
RECENT_FILES=$(tail -20 .claude/state/agent-trace.jsonl 2>/dev/null | \
  jq -r '.files[].path' | sort -u)

# プロジェクト情報
PROJECT=$(tail -1 .claude/state/agent-trace.jsonl 2>/dev/null | \
  jq -r '.metadata.project')
```

**照合ポイント**:

| チェック項目 | 検出方法 |
|------------|----------|
| Plans.md にないファイル編集 | Agent Trace vs タスク記述 |
| タスク記述と異なるファイル | 想定ファイル vs 実際の編集 |
| 長時間編集がないタスク | Agent Trace 時系列 vs WIP 期間 |

## Step 2: 差分検出

| チェック項目 | 検出方法 |
|------------|----------|
| 完了済みなのに `cc:WIP` | コミット履歴 vs マーカー |
| 着手済みなのに `cc:TODO` | 変更ファイル vs マーカー |
| `cc:完了` なのに未コミット | git status vs マーカー |

## Step 3: Plans.md 更新提案

差分が検出された場合、提案して実行する:

```
Plans.md 更新が必要です

| Task | 現在 | 変更後 | 理由 |
|------|------|--------|------|
| XX   | cc:WIP | cc:完了 | コミット済み |
| YY   | cc:TODO | cc:WIP | ファイル編集済み |

更新しますか？ (yes / no)
```

## Step 4: 進捗サマリー出力

```markdown
## 進捗サマリー

**プロジェクト**: {{project_name}}

| ステータス | 件数 |
|----------|------|
| 未着手 (cc:TODO) | {{count}} |
| 作業中 (cc:WIP) | {{count}} |
| 完了 (cc:完了) | {{count}} |
| PM確認済 (pm:確認済) | {{count}} |

**進捗率**: {{percent}}%

### 直近の編集ファイル (Agent Trace)
- {{file1}}
- {{file2}}
```

## Step 5: 次のアクション提案

```
次にやること

**優先 1**: {{タスク}}
- 理由: {{依頼中 / アンブロック待ち}}

**推奨**: /execute, /review
```

## 異常検知

| 状況 | 警告 |
|------|------|
| 複数の `cc:WIP` | 複数タスクが同時進行中 |
| `pm:依頼中` が未処理 | PM の依頼を先に処理する |
| 大きな乖離 | タスク管理が追いついていない |
| WIP が 3日以上更新なし | ブロックされていないか確認 |
