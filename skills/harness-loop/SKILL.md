---
name: harness-loop
description: "長時間タスクを /loop （CC dynamic mode）と ScheduleWakeup で wake-up 毎に fresh context で再入実行。harness-work を内部で Agent 呼び出し。長時間、ループ、loop、wake-up、autonomous に対応。Do NOT load for: 単発タスク実行・レビュー・リリース・プランニング。"
description-ja: "長時間タスクを /loop と ScheduleWakeup で wake-up 毎に fresh context で再入実行。harness-work を内部で Agent 呼び出し。長時間、ループ、loop、wake-up、autonomous に対応。"
allowed-tools: ["Read", "Edit", "Bash", "Task", "ScheduleWakeup", "mcp__harness__harness_mem_resume_pack", "mcp__harness__harness_mem_record_checkpoint"]
argument-hint: "[all|N-M] [--max-cycles N] [--pacing worker|ci|plateau|night]"
---

# harness-loop

`/loop`（CC dynamic mode）と `ScheduleWakeup` を組み合わせ、
長時間タスクを **wake-up 毎に fresh context で再入実行** するメタスキル。

各 wake-up で `harness-work --breezing` を Agent 経由で呼び出し、
1 サイクル = 1 タスク完結の再入可能ループを構成する。

## Quick Reference

| 入力 | 動作 |
|------|------|
| `/harness-loop all` | 全未完了タスクをループ実行（default: max 8 サイクル） |
| `/harness-loop all --max-cycles 3` | 3 サイクルで停止 |
| `/harness-loop 41.1-41.3 --pacing ci` | タスク範囲を CI pacing で実行 |
| `/harness-loop all --pacing night` | 深夜バッチ（3600s 間隔） |

## オプション

| オプション | 説明 | デフォルト |
|----------|------|----------|
| `all` | 全未完了タスクを対象 | - |
| `N-M` | タスク番号範囲指定 | - |
| `--max-cycles N` | 最大サイクル数 | `8` |
| `--pacing <mode>` | wake-up 間隔モード | `worker`（270s） |

### pacing 値マッピング

| pacing | delaySeconds | 用途 |
|--------|-------------|------|
| `worker` | 270 | Worker 完了直後（5 min 以内で cache warm） |
| `ci` | 270 | CI 短時間ジョブ待ち |
| `plateau` | 1200 | 20 min（plateau 検知後の再試行間隔） |
| `night` | 3600 | 深夜の長時間放置 |

> **制約**: `ScheduleWakeup` の `delaySeconds` はランタイムで **[60, 3600]** に clamp される。
> `worker` / `ci` の 270s および `night` の 3600s はこの範囲内。
> `plateau` の 1200s も範囲内。値を直接指定する場合は必ず 60 以上 3600 以下にすること。

## 起動フロー（wake-up 毎のエントリ）

詳細版: [`${CLAUDE_SKILL_DIR}/references/flow.md`](${CLAUDE_SKILL_DIR}/references/flow.md)

```
wake-up
  │
  ▼
[Step 1] Plans.md を先に読む
  cc:WIP / cc:TODO の先頭タスクを特定（task_id を得る）
  ※ 未完了タスクなし → ループ終了（正常完了）
  │
  ▼
[Step 2] sprint-contract 存在確認 & 生成
  .claude/state/contracts/${task_id}.sprint-contract.json の有無を確認
  無ければ node scripts/generate-sprint-contract.sh ${task_id} で生成
  生成直後（初回のみ）: bash scripts/enrich-sprint-contract.sh <contract-path> \
    --check "wake-up 自動承認（harness-loop のため DoD を reviewer 観点で確認）" \
    --approve  ← draft → approved に昇格
  （既存 contract は approved 済みのためスキップ）
  │
  ▼
[Step 3] contract readiness チェック
  bash scripts/ensure-sprint-contract-ready.sh <contract-path>
  │
  ▼
[Step 4] Resume pack 再読込
  harness-mem resume-pack（コンテキスト再注入）
  │
  ▼
[Step 5] 1 タスクサイクル実行
  worker_result = Agent(
      subagent_type="claude-code-harness:worker",  # worker エージェント（harness-work ではない）
      prompt="タスク: ${task_id}\nDoD: <Plans.md から抽出>\ncontract_path: ${CONTRACT_PATH}\nmode: breezing",
      isolation="worktree",
      run_in_background=false
  )
  # worker_result: { commit, branch, worktreePath, files_changed, summary }
  │
  ▼
[Step 5.5] Lead レビュー実行
  diff_text = git show worker_result.commit
  verdict = codex_exec_review(diff_text) or reviewer_agent_review(diff_text)
  ※ 詳細は flow.md 参照
  │
  ▼
[Step 5.6] APPROVE → main に cherry-pick / REQUEST_CHANGES → 修正ループ（contract の max_iterations 回、デフォルト 3）
  APPROVE: git cherry-pick → Plans.md を cc:完了 [{hash}] に更新 → feature branch 削除
  REQUEST_CHANGES x MAX_REVIEWS 後も否決: エスカレーション
  ※ 詳細は flow.md 参照
  │
  ▼
[Step 6] plateau 判定
  scripts/detect-review-plateau.sh ${current_task_id}
  │
  ├── PIVOT_REQUIRED（exit 2）  → ループ停止 + ユーザーエスカレーション
  ├── INSUFFICIENT_DATA（exit 1）→ 続行
  └── PIVOT_NOT_REQUIRED（exit 0）→ 続行
  │
  ▼
[Step 7] サイクル数チェック
  │
  ├── cycles >= max_cycles → ループ停止（上限到達）
  │
  ▼
[Step 8] checkpoint 記録
  harness_mem_record_checkpoint(
      session_id, title, content=サイクル結果サマリ
  )
  │
  ▼
[Step 9] 次 wake-up 予約
  ScheduleWakeup(
      delaySeconds=<pacing値>,
      prompt="/harness-loop <同じ引数>",
      reason="サイクル {N}/{max} 完了 — 次タスクへ"
  )
```

## サイクル停止条件

| 条件 | 停止種別 | 対応 |
|------|---------|------|
| `cycles >= max_cycles` | 正常停止（上限到達） | ユーザーに報告 |
| `PIVOT_REQUIRED`（exit 2） | 異常停止（エスカレーション） | ユーザーに判断を仰ぐ |
| 未完了タスクなし | 正常停止（全完了） | 完了報告を出力 |

`--max-cycles 3` 指定時は 3 サイクル完了後に停止する。
default（`--max-cycles 8`）時は 8 サイクルで停止する。

## /loop との連携

このスキルは CC の `/loop`（dynamic mode）と組み合わせて使用する。

`/loop` を有効にすると CC は自律的な再入実行を継続し、
各サイクルの末尾で `ScheduleWakeup` を呼ぶことで次回 wake-up を予約する。

`/loop` のセンチネル: `<<autonomous-loop-dynamic>>`

各 wake-up は **fresh context** で開始されるため、前サイクルのコンテキスト汚染を防ぐ。
`harness-mem resume-pack` による resume pack 再読込が必須（Step 2）。

## checkpoint 記録

`harness_mem_record_checkpoint` スキーマ:

```json
{
  "session_id": "<セッション ID>",
  "title": "harness-loop cycle {N}/{max}: {タスク名}",
  "content": "cycle_result の 1 行サマリ + commit hash"
}
```

## 関連スキル

- `harness-work` — 各サイクルで実行されるタスク実装スキル
- `harness-plan` — ループ対象タスクの計画
- `harness-review` — 個別タスクのレビュー
- `session-control` — セッション状態管理
