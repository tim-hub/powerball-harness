# harness-loop: wake-up フロー詳細

`harness-loop` の各 wake-up エントリ手順の詳細版。
SKILL.md のサマリを補完する実装リファレンス。

---

## wake-up 毎のエントリ手順（詳細）

### Step 1: Plans.md を先に読む

```bash
# cc:WIP / cc:TODO タスクを抽出し、先頭タスクの task_id を特定
grep -E "cc:(WIP|TODO)" Plans.md | head -1
```

- `cc:WIP` タスクが残っている場合: 前サイクルで中断された可能性あり → task_id を取得して継続
- `cc:TODO` タスクがある場合: 次のターゲットタスクとして task_id を取得
- どちらもない場合: **全タスク完了** → ループ正常終了

> **41.1.2 前提**: `plans-watcher.sh` が flock で Plans.md を保護している場合、
> Plans.md 読み取りはその flock スコープ内で実行すること。
> 41.1.2 リリース前は flock なしで直接読み取り可。

### Step 2: sprint-contract 存在確認 & 生成

```bash
CONTRACT_PATH=".claude/state/contracts/${task_id}.sprint-contract.json"

if [ ! -f "${CONTRACT_PATH}" ]; then
    # contract 未生成 → 生成する
    node scripts/generate-sprint-contract.sh "${task_id}"

    # Step 2.5: draft → approved に昇格（初回生成時のみ）
    # generate-sprint-contract.sh は review.status == "draft" で初期化するため、
    # ensure-sprint-contract-ready.sh（approved 要求）の前に必ず昇格させる
    bash scripts/enrich-sprint-contract.sh "${CONTRACT_PATH}" \
      --check "wake-up 自動承認（harness-loop のため DoD を reviewer 観点で確認）" \
      --approve
fi
```

- `.claude/state/contracts/${task_id}.sprint-contract.json` の有無を確認
- 存在しない場合は `node scripts/generate-sprint-contract.sh ${task_id}` で生成
  （※ 41.5.1 で .sh→.js リネーム予定だが、現時点は既存名を node 経由で呼ぶ）
- **生成直後（初回のみ）**: `enrich-sprint-contract.sh --approve` で `draft` → `approved` に昇格
  - `generate-sprint-contract.sh` は `review.status == "draft"` で初期化する
  - `ensure-sprint-contract-ready.sh`（次の Step 3）は `approved` しか受け付けない
  - `if [ ! -f ... ]` ブロック内に入れることで、既存 contract（前サイクルで approved 済み）には適用しない
- 生成後は `${CONTRACT_PATH}` を以降のステップで使い回す

### Step 3: contract readiness チェック

```bash
bash scripts/ensure-sprint-contract-ready.sh "${CONTRACT_PATH}"
```

- sprint-contract の `review.status == "approved"` を確認
- 未承認 contract が残っている場合はエラーで停止

### Step 4: Resume pack 再読込

```
Step 4. harness-mem resume-pack 再読込:
  mcp__harness__harness_mem_resume_pack ツールを呼ぶ。
  必須引数:
    - project: 現在のプロジェクト名（既存 session-init スキルの実装例に倣う。
              例: リポジトリ root を `basename $(git rev-parse --show-toplevel)` で取得して渡す）
  optional: session_id（前セッションから再開する場合）

  例（擬似コード）:
    resume_pack = mcp__harness__harness_mem_resume_pack(
      project="claude-code-harness",
      session_id=<前回 checkpoint の session_id>
    )
```

fresh context での wake-up 直後は前サイクルのメモリが失われている。
`harness-mem resume-pack` 相当の操作で以下を再注入する:

- `decisions.md` — アーキテクチャ決定事項
- `patterns.md` — 再利用パターン
- `session-state` — 前回の作業状態
- 直前サイクルの `checkpoint` — 何を完了したか

> **注意**: resume pack 再読込は Step 3（contract readiness チェック）の後に実行すること。
> スキップすると前サイクルの成果物を重複実装するリスクがある。

### Step 5: 1 タスクサイクル実行

Agent tool 経由で `claude-code-harness:worker` を spawn する:

> **重要**: `subagent_type` には `"harness-work"` ではなく `"claude-code-harness:worker"` を指定すること。
> `harness-work` はスキルであり agent ではない。実在する agent は `worker` / `reviewer` / `scaffolder`。
> `"harness-work"` を指定すると Agent spawn が失敗し、ループが初回 Worker 起動で停止する。

```python
worker_result = Agent(
    subagent_type="claude-code-harness:worker",  # ← worker エージェント（スキルではない）
    prompt="""
    タスク: ${task_id}
    DoD: <Plans.md から抽出>
    contract_path: ${CONTRACT_PATH}
    mode: breezing
    完了後: commit hash・branch・変更サマリを返却してください。
    """,
    isolation="worktree",
    run_in_background=false  # フォアグラウンド実行（完了まで待機）
)
# worker_result: { commit, branch, worktreePath, files_changed, summary }
```

Worker は `mode: breezing` で動作するため:
- feature branch 上に commit するだけで main には触らない
- `worktreePath` に変更内容が格納される
- Lead（harness-loop）が Step 5.5/5.6 でレビュー → cherry-pick を担当する

> **実装上の注意**: `Bash("harness-work --breezing")` でも代替可能だが、
> Agent tool 経由の方がコンテキスト分離が明確でデバッグしやすい。

### Step 5.5: Lead レビュー実行

Worker が返却した commit に対して Lead がレビューを実行する:

```bash
# diff 取得（worktree 内の commit を対象）
diff_text=$(git -C "${worker_result.worktreePath}" show "${worker_result.commit}")

# レビュー実行（Codex exec 優先、フォールバックで内部 Reviewer agent）
bash scripts/codex-companion.sh review --diff "${diff_text}"
# → review-output.json に verdict が書き込まれる
```

**verdict 判定**:

| verdict | アクション |
|---------|----------|
| `APPROVE` | Step 5.6 へ（cherry-pick） |
| `REQUEST_CHANGES` | 修正ループへ（最大 3 回） |

**修正ループ（REQUEST_CHANGES 時）**:

```python
review_count = 0
latest_commit = worker_result.commit
worker_id = worker_result.agentId

while verdict == "REQUEST_CHANGES" and review_count < 3:
    # Worker に修正を指示（SendMessage で再開）
    SendMessage(to=worker_id, message=f"指摘内容: {issues}\n修正して amend してください")
    updated_result = wait_for_response(worker_id)
    latest_commit = updated_result.commit
    diff_text = git("-C", worker_result.worktreePath, "show", latest_commit)
    verdict = codex_exec_review(diff_text) or reviewer_agent_review(diff_text)
    review_count += 1

if review_count >= 3 and verdict != "APPROVE":
    # エスカレーション
    raise PivotRequired(f"3 回修正後も REQUEST_CHANGES: {issues}")
```

### Step 5.6: APPROVE → main に cherry-pick

```bash
# main ブランチに戻る（Worker は feature branch で作業）
git checkout main

# feature branch の commit が main に未マージかを確認（再入防止）
if ! git merge-base --is-ancestor "${latest_commit}" HEAD; then
    git cherry-pick --no-commit "${latest_commit}"
    git commit -m "${task_title}"
fi

# Worker が作成した feature branch を削除
if [ -n "${worker_result.branch}" ] && \
   [ "${worker_result.branch}" != "main" ] && \
   [ "${worker_result.branch}" != "master" ]; then
    git branch -D "${worker_result.branch}"
fi
```

Plans.md を更新:

```bash
# cc:WIP → cc:完了 [{hash}] に更新
HASH=$(git rev-parse --short HEAD)
# Plans.md の該当タスク行を更新
```

### Step 6: plateau 判定

```bash
bash scripts/detect-review-plateau.sh ${current_task_id}
PLATEAU_EXIT=$?
# ※ current_task_id は Step 1 で特定した task_id
```

| exit code | 意味 | アクション |
|-----------|------|----------|
| `0` | `PIVOT_NOT_REQUIRED` | 続行 |
| `1` | `INSUFFICIENT_DATA` | 続行（データ不足） |
| `2` | `PIVOT_REQUIRED` | **ループ停止** + エスカレーション |

**PIVOT_REQUIRED 時のエスカレーションメッセージ**:

```
harness-loop: plateau 検知により停止（サイクル {N}/{max}）

検知された問題:
  {plateau の詳細: detect-review-plateau.sh の出力}

対応案:
  1. 手動でタスク内容を見直す
  2. `--pacing plateau` で間隔を延ばして再実行
  3. 問題タスクをスキップして `/harness-loop` を再起動

現在の Plans.md 状態を確認してください。
```

### Step 7: サイクル数チェック

```
cycles_completed += 1
if cycles_completed >= max_cycles:
    ループ停止
    print(f"harness-loop: {max_cycles} サイクル完了で停止")
    return
```

- default `max_cycles = 8`
- `--max-cycles N` 指定時は N サイクルで停止

**サイクルカウントの永続化**:
- `ScheduleWakeup` の `prompt` 引数にカウントを埋め込む:
  ```
  /harness-loop all --max-cycles 8 --cycles-done {N} --pacing worker
  ```
- wake-up 時に `--cycles-done N` を読み取り、カウントを復元する

### Step 8: checkpoint 記録

```json
{
  "session_id": "<現在のセッション ID>",
  "title": "harness-loop cycle {N}/{max}: {task_completed}",
  "content": "cycle {N} 完了。commit: {commit}。変更: {files_changed}。次: {next_task}"
}
```

`harness_mem_record_checkpoint` ツールでメモリに記録する。
次の wake-up の resume pack に自動的に含まれる。

### Step 9: 次 wake-up 予約

```
ScheduleWakeup(
    delaySeconds=<pacing に対応する値>,
    prompt="/harness-loop <同じ引数> --cycles-done {N}",
    reason="サイクル {N}/{max} 完了: {task_completed}"
)
```

**pacing に対応する delaySeconds**:

| pacing | delaySeconds | 選定理由 |
|--------|-------------|---------|
| `worker` | 270 | Worker 完了直後の再入（5 min cache warm 以内） |
| `ci` | 270 | CI ジョブの最短完了を想定した待機 |
| `plateau` | 1200 | 20 min 冷却期間（plateau 回避） |
| `night` | 3600 | 深夜バッチ（最大 clamp 値） |

> **clamp 制約**: `ScheduleWakeup` は `delaySeconds` を `[60, 3600]` にランタイムで clamp する。
> 60 未満を指定すると 60 に切り上げ、3600 超を指定すると 3600 に切り下げられる。
> 設計値は全て範囲内だが、将来的な変更時は要注意。

---

## サイクル停止条件マトリクス

| 条件 | サイクル数 | exit | 停止理由 | ユーザー通知 |
|------|-----------|------|---------|------------|
| `cycles >= max_cycles` | N (上限) | 0 | 正常上限 | 「{N} サイクル完了で停止」 |
| `PIVOT_REQUIRED` | 任意 | 2 | plateau 検知 | エスカレーション詳細 |
| 未完了タスクなし | 任意 | 0 | 全タスク完了 | 完了報告 |
| ユーザーキャンセル | 任意 | - | 手動中断 | - |

---

## pacing 選択ガイド

### どの pacing を使うべきか

```
タスクの性質は？
│
├── Worker 完了直後に再入したい
│     → worker（270s）
│
├── CI / テストの完了を待つ必要がある
│     → ci（270s）
│     ※ CI が 270s 以上かかる場合は手動で --pacing を調整
│
├── plateau を検知して間隔を空けたい
│     → plateau（1200s）
│
└── 深夜に放置して翌朝確認したい
      → night（3600s）
```

### pacing 変更のタイミング

- **初回起動時**: 通常は `worker`（デフォルト）で良い
- **CI 待ちが多い場合**: `--pacing ci` に切り替え
- **plateau 検知後**: `--pacing plateau` で自動切り替えを検討（Step 5 参照）
- **夜間放置**: `--pacing night` で起動してそのまま就寝

---

## ScheduleWakeup の制約詳細

### delaySeconds のランタイム制約

```
ScheduleWakeup(delaySeconds=X)
  → X < 60  → clamp to 60
  → X > 3600 → clamp to 3600
  → 60 <= X <= 3600 → そのまま使用
```

### cache TTL との関係

ScheduleWakeup の cache TTL は **5 min（300s）**。

- `worker` / `ci` の 270s は 5 min 以内 → cache warm な状態で wake-up
- `plateau` の 1200s、`night` の 3600s は cache 失効後に wake-up
  → Step 2（resume pack 再読込）が特に重要

### prompt の引数引き継ぎ

サイクルカウントを次の wake-up に引き継ぐ方法:

```bash
# 現在の cycle count を prompt に埋め込む
NEXT_PROMPT="/harness-loop ${SCOPE} --max-cycles ${MAX_CYCLES} --cycles-done ${CYCLES_DONE} --pacing ${PACING}"

ScheduleWakeup(
    delaySeconds=${DELAY},
    prompt="${NEXT_PROMPT}",
    reason="cycle ${CYCLES_DONE}/${MAX_CYCLES} 完了"
)
```

---

## 参考: spike 41.0.0 の検証結果

この設計は spike 41.0.0 の実証結果に基づく:

- `ScheduleWakeup`: 内部ツールとして存在確認済み。delay [60, 3600] clamp、cache 5min TTL
- `/loop`: CC dynamic mode として存在確認済み。sentinel `<<autonomous-loop-dynamic>>`
- `harness_mem_record_checkpoint`: 存在確認済み（schema: session_id / title / content 必須）

これらの前提が変わった場合は本ファイルを更新すること。
