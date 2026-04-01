---
name: harness-work
description: "Harness v3 統合実行スキル（Codex ネイティブ版）。Plans.md タスクを1件から全並列チーム実行まで担当。以下で起動: 実装して、実行して、harness-work、全部やって、breezing、チーム実行、parallel。プランニング・レビュー・リリース・セットアップには使わない。"
description-en: "Unified execution skill for Harness v3 (Codex native). Implements Plans.md tasks from single task to full parallel team runs."
description-ja: "Harness v3 統合実行スキル（Codex ネイティブ版）。Plans.md タスクを1件から全並列チーム実行まで担当。"
argument-hint: "[all] [task-number|range] [--parallel N] [--no-commit] [--breezing]"
effort: high
---

# Harness Work (v3) — Codex Native

> **この SKILL.md は Codex CLI ネイティブ版です。**
> Claude Code 版は `skills-v3/harness-work/SKILL.md` を参照してください。
> サブエージェント API は Codex の `spawn_agent` / `send_input` / `wait_agent` / `close_agent` を使用します。

Harness v3 の統合実行スキル。

## Quick Reference

| ユーザー入力 | モード | 動作 |
|------------|--------|------|
| `harness-work` | **solo** | 次の未完了タスクを1件実行 |
| `harness-work all` | **sequential** | 全未完了タスクを直列実行 |
| `harness-work 3` | solo | タスク3だけ即実行 |
| `harness-work --parallel 3` | parallel | companion `task` で3並列実行（Bash `&` + `wait`） |
| `harness-work --breezing` | breezing | `spawn_agent` によるチーム実行（明示時のみ） |

## Execution Mode Selection

> **重要**: Codex では `spawn_agent` はユーザーが明示的にチーム実行・並列作業を求めた場合にのみ使用する。
> タスク件数だけを根拠に自動昇格しない。

| 条件 | モード | 理由 |
|------|--------|------|
| 引数なし / 1件指定 | **Solo** | 直接実装が最速 |
| `all` / 範囲指定（フラグなし） | **Sequential** | 直列で安全に逐次処理 |
| `--parallel N` | **Parallel** | companion `task` の Bash 並列（明示時のみ） |
| `--breezing` | **Breezing** | `spawn_agent` チーム実行（明示時のみ） |

### ルール

1. **明示フラグは常にデフォルトを上書き**する
2. **`--breezing` と `--parallel` は明示時のみ発動**。件数による自動昇格はしない
3. `--parallel` と `--breezing` は排他（同時指定不可）

## オプション

| オプション | 説明 | デフォルト |
|----------|------|----------|
| `all` | 全未完了タスクを対象 | - |
| `N` or `N-M` | タスク番号/範囲指定 | - |
| `--parallel N` | companion `task` Bash 並列数 | - |
| `--sequential` | 直列実行強制 | - |
| `--no-commit` | main への最終コミット抑制（Solo/Sequential のみ。Breezing/Parallel では非対応） | false |
| `--breezing` | Lead/Worker/Reviewer のチーム実行 | false |
| `--no-tdd` | TDD フェーズスキップ | false |

## スコープダイアログ（引数なし時）

```
harness-work
どこまでやりますか?
1) 次のタスク: Plans.md の次の未完了タスク → Solo で実行
2) 全部: 残りのタスクをすべて直列実行
3) 番号指定: タスク番号を入力（例: 3, 5-7）
```

引数ありなら即実行（対話スキップ）。

## 実行モード詳細

### Solo モード

1. Plans.md を読み込み、対象タスクを特定
   - **Plans.md が存在しない場合**: `harness-plan create --ci` を自動呼び出し → Plans.md を生成して続行
   - ヘッダーに DoD / Depends カラムがない場合: 停止
   - **会話に未記載タスクがある場合**: 直前の会話コンテキストから要件を抽出し、Plans.md に `cc:TODO` で自動追記
1.5. **タスク背景確認**（30 秒）:
   - タスクの「内容」と「DoD」から目的を 1 行で推論表示
   - 推論に自信がある場合: そのまま実装に進む
   - 推論に自信がない場合: ユーザーに 1 問だけ確認
2. タスクを `cc:WIP` に更新。`TASK_BASE_REF=$(git rev-parse HEAD)` を記録
3. **TDD フェーズ**（`[skip:tdd]` なし & テストFW存在時）:
   a. テストファイルを先に作成（Red）
   b. 失敗を確認
4. コードを実装（Green）
5. `git commit` で自動コミット（`--no-commit` で省略可）
6. **自動レビューステージ**（「レビューループ」参照）— TASK_BASE_REF..HEAD の差分をレビュー
7. タスクを `cc:完了 [hash]` に更新
8. **リッチ完了報告**（「完了報告フォーマット」参照）
9. **失敗時の自動再計画**（テスト/CI 失敗時のみ）

### Sequential モード（`all` 指定時のデフォルト）

Plans.md のタスクを依存順に1件ずつ Solo モードで逐次処理する。
各タスク完了後に Plans.md を更新し、次タスクに進む。

### Parallel モード（`--parallel N` 明示時のみ）

独立タスクを Bash の `&` + `wait` で並列実行する。
公式プラグインの companion `task` を使用し、worktree で Worker ごとに分離する。

> **制約**: 同一ファイルを変更する可能性があるタスクは並列化しないこと。
> `git worktree add` で Worker ごとに作業ディレクトリを分離し、Lead がレビュー後に cherry-pick する。

```bash
# Worker ごとに worktree を分離（-b <branch> <path> の順序に注意）
git worktree add -b worker-a-$$ /tmp/worker-a-$$
git worktree add -b worker-b-$$ /tmp/worker-b-$$

# タスク A（cd で worktree に cwd を切り替えてから companion を呼ぶ）
PROMPT_A=$(mktemp /tmp/codex-prompt-XXXXXX.md)
cat > "$PROMPT_A" << EOF
タスク A の内容...

完了後、以下の JSON を stdout に出力してください:
{"commit": "<hash>", "files_changed": ["path1", "path2"], "summary": "..."}
EOF
(cd /tmp/worker-a-$$ && cat "$PROMPT_A" | bash "${PROJECT_ROOT}/scripts/codex-companion.sh" task --write) > /tmp/out-a-$$.json 2>>/tmp/harness-codex-$$.log &

# タスク B（cd で worktree に cwd を切り替えてから companion を呼ぶ）
PROMPT_B=$(mktemp /tmp/codex-prompt-XXXXXX.md)
cat > "$PROMPT_B" << EOF
タスク B の内容...

完了後、以下の JSON を stdout に出力してください:
{"commit": "<hash>", "files_changed": ["path1", "path2"], "summary": "..."}
EOF
(cd /tmp/worker-b-$$ && cat "$PROMPT_B" | bash "${PROJECT_ROOT}/scripts/codex-companion.sh" task --write) > /tmp/out-b-$$.json 2>>/tmp/harness-codex-$$.log &

wait
rm -f "$PROMPT_A" "$PROMPT_B"

# Lead が各 Worker の出力 JSON から commit hash を取得し、個別にレビュー → cherry-pick
# ... レビュー・cherry-pick 処理 ...

# worktree 削除
git worktree remove /tmp/worker-a-$$
git worktree remove /tmp/worker-b-$$
```

### Breezing モード（`--breezing` 明示時のみ）

Lead / Worker / Reviewer の役割分離でチーム実行する。
Codex の native subagent API を使用する。

> **`--breezing` は明示時のみ**。ユーザーが「チーム実行で」「breezing で」と指示した場合に限り使用する。

```
Lead (this agent)
├── Worker (spawn_agent) — 実装担当
│   各 Worker は git worktree で分離された作業ディレクトリで動作
└── Reviewer (companion review --base) — レビュー担当
```

**Phase A: Pre-delegate（準備）**:
1. Plans.md を読み込み、対象タスクを特定
2. 依存グラフを解析し、実行順序を決定（Depends カラム）
3. 各タスクに対応する git worktree を作成
4. `scripts/generate-sprint-contract.sh` で `sprint-contract.json` を生成
5. `scripts/enrich-sprint-contract.sh` で Reviewer 観点を加え、`scripts/ensure-sprint-contract-ready.sh` で未承認なら停止

**Phase B: Delegate（Worker spawn → レビュー → cherry-pick）**:

各タスクについて以下を**逐次**実行する（依存順）:

```
for task in execution_order:
    # B-0. 作業ディレクトリ分離
    worktree_path = "/tmp/worker-{task.number}-$$"
    branch_name = "worker-{task.number}-$$"
    git worktree add -b {branch_name} {worktree_path}
    TASK_BASE_REF = git rev-parse HEAD  # このタスク固有の base ref

    # B-1. sprint-contract を生成
    contract_path = bash("scripts/generate-sprint-contract.sh {task.number}")
    contract_path = bash("scripts/enrich-sprint-contract.sh {contract_path} --check \"DoD を reviewer 観点で確認\" --approve")
    bash("scripts/ensure-sprint-contract-ready.sh {contract_path}")

    # B-2. Worker spawn（Codex native subagent）
    Plans.md: task.status = "cc:WIP"

    worker_id = spawn_agent({
        message: "作業ディレクトリ: {worktree_path} で作業してください。\n\nタスク: {task.内容}\nDoD: {task.DoD}\ncontract_path: {contract_path}\n\n実装してください。完了後 git commit してください。\n\n完了時、以下の JSON を返してください:\n{\"commit\": \"<hash>\", \"files_changed\": [\"path1\"], \"summary\": \"...\"}",
        fork_context: true
    })
    wait_agent({ ids: [worker_id] })
    # Worker の出力から commit hash, files_changed, summary を取得

    # B-3. Lead がレビュー実行（companion review --base TASK_BASE_REF）
    # このタスク固有の diff のみをレビュー（TASK_BASE_REF 起点）
    # 公式プラグインの構造化レビューを使用
    verdict = companion_review(TASK_BASE_REF)  # static review（詳細は「レビューループ」参照）
    profile = jq(contract_path, ".review.reviewer_profile")
    browser_mode = jq(contract_path, ".review.browser_mode // \"scripted\"")
    review_input = "review-output.json"
    if profile == "runtime":
        # worktree 内で runtime checks を実行（Lead の cwd ではなく Worker の成果物に対して）
        review_input = bash("cd {worktree_path} && scripts/run-contract-review-checks.sh {contract_path}")
        runtime_verdict = jq(review_input, ".verdict")
        if runtime_verdict == "REQUEST_CHANGES":
            verdict = "REQUEST_CHANGES"
        elif runtime_verdict == "DOWNGRADE_TO_STATIC":
            # runtime 検証コマンドなし → static review 結果にフォールバック
            review_input = "review-output.json"
    if profile == "browser":
        # browser artifact は PENDING_BROWSER scaffold。実際の browser 実行は reviewer agent が後続で担当。
        # review-result には static review の verdict を書く（PENDING_BROWSER ではなく）。
        browser_artifact = bash("scripts/generate-browser-review-artifact.sh {contract_path}")
        # review_input は static review のまま維持
    # DOWNGRADE_TO_STATIC が review_input に残っていないことを確認
    if review_input != "review-output.json" and jq(review_input, ".verdict") == "DOWNGRADE_TO_STATIC":
        review_input = "review-output.json"
    bash("scripts/write-review-result.sh {review_input} {commit_hash}")

    # B-4. 修正ループ（REQUEST_CHANGES 時、最大 3 回）
    review_count = 0
    while verdict == "REQUEST_CHANGES" and review_count < 3:
        # Worker は完了済みだが close していないので send_input で直接指示可能
        send_input({
            id: worker_id,
            message: "指摘内容: {issues}\n修正して git commit --amend してください。修正後 JSON を再出力してください。"
        })
        wait_agent({ ids: [worker_id] })
        # 再レビュー（TASK_BASE_REF 起点の差分）
        diff_text = git("-C", worktree_path, "diff", TASK_BASE_REF, "HEAD")
        verdict = codex_exec_review(diff_text)
        review_count++

    close_agent({ id: worker_id })

    # B-5. 結果処理
    if verdict == "APPROVE":
        # worktree の commit を main に cherry-pick
        commit_hash = git("-C", worktree_path, "rev-parse", "HEAD")
        git cherry-pick --no-commit {commit_hash}
        git commit -m "{task.内容}"
        Plans.md: task.status = "cc:完了 [{short_hash}]"
    else:
        → ユーザーにエスカレーション（Plans.md は cc:WIP のまま）
        # B-5 以降はスキップ、次タスクも停止

    # B-6. Worktree クリーンアップ
    git worktree remove {worktree_path}
    git branch -D {branch_name}

    # B-7. Progress feed
    print("📊 Progress: Task {completed}/{total} 完了 — {task.内容}")
```

**Phase C: Post-delegate（統合・報告）**:
1. 全タスクの commit log を集計
2. **リッチ完了報告** を出力
3. Plans.md の最終確認（全タスク cc:完了 になっているか）

## CI 失敗時の対応

1. ログを確認してエラーを特定
2. 修正を実施
3. 同一原因で 3 回失敗したら自動修正ループを停止
4. 失敗ログ・試みた修正・残る論点をまとめてエスカレーション

## 失敗タスクの自動再チケット化

タスク完了後にテスト/CI が失敗した場合、修正タスク案を自動生成し、承認後に Plans.md へ反映する。

| 条件 | アクション |
|------|----------|
| `cc:完了` 後にテスト失敗 | 修正タスク案を提示し、承認を待つ |
| CI 失敗（3回未満） | 修正を実施 |
| CI 失敗（3回目） | 修正タスク案を提示 + エスカレーション |

## レビューループ

実装完了後に自動実行される品質検証ステージ。
**全モード共通**（Solo / Sequential / Parallel / Breezing）で統一的に適用される。

### レビュー実行（公式プラグイン companion 経由）

公式プラグイン `codex-plugin-cc` の companion review を使用する。
構造化出力（`review-output.schema.json` 準拠）で verdict を取得する。

> **差分の起点**: 各タスク固有の `TASK_BASE_REF`（タスク着手時の HEAD）を使う。
> 累積差分ではなく、そのタスクの変更のみをレビュー対象にする。

```bash
# タスク開始時に base ref を記録（cc:WIP 更新前に実行）
TASK_BASE_REF=$(git rev-parse HEAD)

# ... 実装完了後 ...

# 公式プラグインの構造化レビューを実行
bash scripts/codex-companion.sh review --base "${TASK_BASE_REF}"
REVIEW_EXIT=$?
```

**verdict マッピング**（公式プラグイン → Harness 形式）:

| 公式 plugin | Harness | verdict 影響 |
|---|---|---|
| `approve` | `APPROVE` | - |
| `needs-attention` | `REQUEST_CHANGES` | - |
| `findings[].severity: critical` | `critical_issues[]` | 1件でも → REQUEST_CHANGES |
| `findings[].severity: high` | `major_issues[]` | 1件でも → REQUEST_CHANGES |
| `findings[].severity: medium/low` | `recommendations[]` | verdict に影響しない |

companion review の出力から verdict を判定する:
- `verdict` が `approve` → `APPROVE`
- `verdict` が `needs-attention` → `REQUEST_CHANGES`
- `findings` に `critical` / `high` severity がある → `REQUEST_CHANGES`

### APPROVE / REQUEST_CHANGES の判定基準

| 重要度 | 定義 | verdict への影響 |
|--------|------|-----------------|
| **critical** | セキュリティ脆弱性、データ損失リスク、本番障害の可能性 | 1 件でも → REQUEST_CHANGES |
| **major** | 既存機能の破壊、仕様との明確な矛盾、テスト不通過 | 1 件でも → REQUEST_CHANGES |
| **minor** | 命名改善、コメント不足、スタイル不統一 | verdict に影響しない |
| **recommendation** | ベストプラクティス提案、将来の改善案 | verdict に影響しない |

> **重要**: minor / recommendation のみの場合は **必ず APPROVE** を返すこと。

### 修正ループ（REQUEST_CHANGES 時）

```
review_count = 0
MAX_REVIEWS = 3

while verdict == "REQUEST_CHANGES" and review_count < MAX_REVIEWS:
    1. レビュー指摘を解析（critical / major のみ対象）
    2. 各指摘に対して修正を実装
    3. git commit --amend
    4. 再度 companion review でレビューを実行（TASK_BASE_REF 起点）
    review_count++

if review_count >= MAX_REVIEWS and verdict != "APPROVE":
    → ユーザーにエスカレーション
    → 「3 回修正しましたが以下の critical/major 指摘が残っています」+ 指摘一覧を表示
    → ユーザー判断を待つ（続行 / 中断）
```

### Breezing モードでの適用

1. Worker が worktree 内で実装・commit → `wait_agent` で完了待ち
2. Lead が companion review でレビュー（TASK_BASE_REF 起点）
3. REQUEST_CHANGES → `send_input` で Worker に修正指示 → Worker が amend
4. 修正後、再レビュー（最大 3 回）
5. `close_agent` で Worker を終了
6. APPROVE → Lead が main に cherry-pick → Plans.md を `cc:完了 [{hash}]` に更新

### Worker の出力契約

Worker プロンプトには、完了時に以下の JSON を返すことを明示する:

```json
{
  "commit": "a1b2c3d",
  "files_changed": ["src/foo.ts", "tests/foo.test.ts"],
  "summary": "foo モジュールに bar 機能を追加"
}
```

Lead はこの JSON を解析して commit hash とファイル一覧を取得する。
Worker が JSON を返さなかった場合は `git log --oneline -1` で直近 commit を取得する。

## 完了報告フォーマット

タスク完了時に自動出力される視覚的サマリ。

### Solo テンプレート

```
┌─────────────────────────────────────────────┐
│  ✓ Task {N} 完了: {タスク名}                    │
├─────────────────────────────────────────────┤
│  ■ 何をしたか                                 │
│    • {変更内容 1}                              │
│    • {変更内容 2}                              │
│  ■ 何が変わるか                                │
│    Before: {旧動作}                            │
│    After:  {新動作}                            │
│  ■ 変更ファイル ({N} files)                    │
│    {ファイルパス 1}                             │
│  ■ 残りの課題                                  │
│    Plans.md に {M} 件の未完了タスクあり          │
│  commit: {hash} | review: {APPROVE}           │
└─────────────────────────────────────────────┘
```

### Breezing テンプレート

```
┌─────────────────────────────────────────────┐
│  ✓ Breezing 完了: {N}/{M} タスク             │
├─────────────────────────────────────────────┤
│  1. ✓ {タスク名 1}            [{hash1}]      │
│  2. ✓ {タスク名 2}            [{hash2}]      │
│  ■ 全体の変更                                 │
│    {N} files changed, {A} insertions(+),     │
│    {D} deletions(-)                          │
│  ■ 残りの課題                                  │
│    Plans.md に {K} 件の未完了タスクあり         │
└─────────────────────────────────────────────┘
```

## Claude Code 版との差分

| 項目 | Claude Code 版 | Codex ネイティブ版（本ファイル） |
|------|---------------|-------------------------------|
| Worker spawn | `Agent(subagent_type="worker")` | `spawn_agent({message, fork_context})` |
| 完了待ち | `Agent` の戻り値 | `wait_agent({ids: [id]})` |
| 修正指示 | `SendMessage(to: agentId)` | `send_input({id, message})` |
| Worker 終了 | 自動（Agent tool 戻り値） | `close_agent({id})` で明示終了 |
| Worktree 分離 | `isolation="worktree"` 自動管理 | `git worktree add` で手動分離 |
| 権限 | `bypassPermissions` | companion `task --write` / `spawn_agent`: セッション権限継承 |
| レビュー | Codex exec → Reviewer agent fallback | companion `review --base` （構造化出力） |
| verdict 取得 | Agent 応答を解析 | companion review の verdict フィールド（approve/needs-attention） |
| モード自動昇格 | タスク数で自動判定 | 明示フラグのみ（自動昇格しない） |
| Effort 制御 | `ultrathink` + `/effort` | `model_reasoning_effort` in config.toml |
| Auto-Refinement | `/simplify` | なし |

## 関連スキル

- `harness-plan` — 実行するタスクを計画する
- `harness-sync` — 実装と Plans.md を同期する
- `harness-review` — 実装のレビュー
- `harness-release` — バージョンバンプ・リリース
