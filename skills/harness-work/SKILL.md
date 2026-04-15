---
name: harness-work
description: "HAR:Plans.md タスクを1件から全並列チーム実行まで担当。実装して、実行して、全部やって、breezing、チーム実行、parallel で起動。プランニング・レビュー・リリース・セットアップには使わない。"
description-en: "HAR: Execute Plans.md tasks from single task to full parallel team run. Trigger: implement, execute, do everything, breezing, team run, parallel. Do NOT load for: planning, review, release, setup."
description-ja: "HAR:Plans.md タスクを1件から全並列チーム実行まで担当。実装して、実行して、全部やって、breezing、チーム実行、parallel で起動。プランニング・レビュー・リリース・セットアップには使わない。"
allowed-tools: ["Read", "Write", "Edit", "Grep", "Glob", "Bash", "Task", "Monitor"]
argument-hint: "[all] [task-number|range] [--codex] [--parallel N] [--no-commit] [--resume id] [--breezing] [--auto-mode]"
effort: high
---

# Harness Work

Harness の統合実行スキル。
以下の旧スキルを統合:

- `work` — Plans.md タスクの実装（スコープ自動判断）
- `impl` — 機能実装（タスクベース）
- `breezing` — チームフル自動実行
- `parallel-workflows` — 並列ワークフロー最適化
- `ci` — CI 失敗時の復旧

## Quick Reference

| ユーザー入力 | モード | 動作 |
|------------|--------|------|
| `harness-work` | **auto** | タスク数で自動判定（下記参照） |
| `harness-work all` | **auto** | 全未完了タスクを自動モードで実行 |
| `harness-work 3` | solo | タスク3だけ即実行 |
| `harness-work --parallel 5` | parallel | 5ワーカーで並列実行（強制） |
| `harness-work --codex` | codex | Codex CLI に委託（明示時のみ） |
| `harness-work --breezing` | breezing | チーム実行を強制 |

## Execution Mode Auto Selection（フラグなし時の自動判定）

明示的なモードフラグ（`--parallel`, `--breezing`, `--codex`）がない場合、
対象タスク数に応じて最適なモードを自動選択する:

| 対象タスク数 | 自動選択モード | 理由 |
|-------------|---------------|------|
| **1 件** | Solo | オーバーヘッド最小。直接実装が最速 |
| **2〜3 件** | Parallel（Task tool） | Worker 分離のメリットが出始める閾値 |
| **4 件以上** | Breezing | Lead 調整 + Worker 並列 + Reviewer 独立の三者分離が効果的 |

### ルール

1. **明示フラグは常にオートモードを上書き**する
   - `--parallel N` → Parallel モード（タスク数に関係なく）
   - `--breezing` → Breezing モード（タスク数に関係なく）
   - `--codex` → Codex モード（タスク数に関係なく）
2. **`--codex` は明示時のみ発動**。Codex CLI が未インストールの環境があるため、自動選択しない
3. `--codex` は他モードと組み合わせ可能: `--codex --breezing` → Codex + Breezing

## オプション

| オプション | 説明 | デフォルト |
|----------|------|----------|
| `all` | 全未完了タスクを対象 | - |
| `N` or `N-M` | タスク番号/範囲指定 | - |
| `--parallel N` | 並列ワーカー数 | auto |
| `--sequential` | 直列実行強制 | - |
| `--codex` | Codex CLI で実装委託（明示時のみ、自動選択しない） | false |
| `--no-commit` | 自動コミット抑制 | false |
| `--resume <id\|latest>` | 前回セッション再開 | - |
| `--breezing` | Lead/Worker/Reviewer のチーム実行 | false |
| `--no-tdd` | TDD フェーズスキップ | false |
| `--no-simplify` | Auto-Refinement スキップ | false |
| `--auto-mode` | Auto Mode rollout を明示。親セッションの permission mode が互換な場合のみ採用を検討 | false |

> **Token Optimization (v2.1.69+)**: git 操作を伴わない軽量タスクでは
> plugin settings の `includeGitInstructions: false` を有効にして
> プロンプトトークンを削減できる。

## スコープダイアログ（引数なし時）

```
harness-work
どこまでやりますか?
1) 次のタスク: Plans.md の次の未完了タスク → Solo で実行
2) 全部（推奨）: 残りのタスクをすべて完了 → タスク数で自動モード選択
3) 番号指定: タスク番号を入力（例: 3, 5-7）→ 件数で自動モード選択
```

引数ありなら即実行（対話スキップ）:
- `harness-work all` → 全タスク、自動モード選択
- `harness-work 3-6` → 4件なので Breezing 自動選択

## Effort レベル制御（v2.1.68+, v2.1.72 簡素化）

Claude Code v2.1.68 で Opus 4.6 は **medium effort** (`◐`) がデフォルト。
v2.1.72 で `max` レベルが廃止され、3段階 `low(○)/medium(◐)/high(●)` に簡素化。
`/effort auto` でデフォルトにリセット可能。
複雑なタスクには `ultrathink` キーワードで high effort (`●`) を有効化する。

### 多要素スコアリング

タスク着手時に以下のスコアを合算し、**閾値 3 以上**で ultrathink を注入:

| 要素 | 条件 | スコア |
|------|------|--------|
| ファイル数 | 変更対象 4 ファイル以上 | +1 |
| ディレクトリ | core/, guardrails/, security/ を含む | +1 |
| キーワード | architecture, security, design, migration を含む | +1 |
| 失敗履歴 | agent memory に同タスクの失敗記録あり | +2 |
| 明示指定 | PM テンプレートに ultrathink 記載あり | +3（自動採用） |

### 注入方法

スコア ≥ 3 の場合、Worker spawn prompt の冒頭に `ultrathink` を追加。
breezing モードでも同じロジックが適用される（harness-work が一本化して管理）。

## 実行モード詳細

### Solo モード（1 件時の自動選択）

1. Plans.md を読み込み、対象タスクを特定
   - **Plans.md が存在しない場合**: `harness-plan create --ci` を自動呼び出し → Plans.md を生成して続行
   - ヘッダーに DoD / Depends カラムがない場合: `Plans.md が旧フォーマットです。harness-plan create で再生成してください。` → **停止**
   - **会話に未記載タスクがある場合**: 直前の会話コンテキストから要件を抽出し、Plans.md に `cc:TODO` で自動追記
     - 抽出ロジック: ユーザー発言からアクション動詞（「〜を追加」「〜を修正」「〜を実装」）を検出
     - 追記時は v2 フォーマット（Task / 内容 / DoD / Depends / Status）に準拠
     - 追記後、ユーザーに「Plans.md に以下を追記しました」と表示（5 秒タイムアウト付きプロンプト、デフォルト: 続行）
1.5. **タスク背景確認**（30 秒）:
   - タスクの「内容」と「DoD」から **目的**（このタスクが解く課題）を 1 行で推論表示
   - `git grep` / `Glob` で **影響範囲**（変更が及ぶファイル/モジュール）を推論表示
   - 推論に自信がある場合: そのまま実装に進む（フロー遅延なし）
   - 推論に自信がない場合: ユーザーに 1 問だけ確認（「この理解で合っていますか？」）
2. タスクを `cc:WIP` に更新
3. **TDD フェーズ**（`[skip:tdd]` なし & テストFW存在時）:
   a. テストファイルを先に作成（Red）
   b. 失敗を確認
4. `scripts/generate-sprint-contract.sh <task-id>` で `sprint-contract.json` を生成
5. Reviewer 観点の追記を `scripts/enrich-sprint-contract.sh` で加え、`scripts/ensure-sprint-contract-ready.sh` で approved を確認
6. コードを実装（Green）（Read/Write/Edit/Bash）
7. `/simplify` で Auto-Refinement（`--no-simplify` で省略可）
8. **自動レビューステージ**（「レビューループ」参照）:
   - Codex exec 優先でレビュー実行 → フォールバックで内部 Reviewer agent
   - `sprint-contract.json` の `reviewer_profile` が `runtime` の場合は `scripts/run-contract-review-checks.sh` を実行
   - REQUEST_CHANGES の場合: 指摘を元に修正→再レビュー（最大 3 回）
   - APPROVE で次ステップへ。self-check だけでは完了を確定しない
9. `scripts/write-review-result.sh` で review artifact を正規化して保存
10. `git commit` で自動コミット（`--no-commit` で省略可）
11. タスクを `cc:完了` に更新（commit hash 付与）
   - `git log --oneline -1` で直近の commit hash（短縮形 7 文字）を取得
   - Plans.md の Status を `cc:完了 [a1b2c3d]` 形式で更新
   - commit がない場合（`--no-commit` 時）は hash なしで `cc:完了` のみ
12. **リッチ完了報告**（「完了報告フォーマット」参照）
13. **失敗時の自動再計画**（テスト/CI 失敗時のみ）:
    - テスト実行結果を確認
    - 失敗した場合: 修正タスク案を state に保存し、承認コマンド経由で Plans.md に追加（「失敗タスクの自動再チケット化」参照）
    - 成功した場合: 次タスクへ進む

### Parallel モード（2〜3 件時の自動選択 / `--parallel N` で強制）

`[P]` マーク付きタスクを N ワーカーで並列実行。
`--parallel N` で明示指定した場合は、タスク数に関係なくこのモードを使用。
同一ファイルへの書き込みが競合する場合は git worktree で分離。

### Codex モード（`--codex` 明示時のみ）

公式プラグイン `codex-plugin-cc` の companion 経由で Codex CLI にタスクを委託する。

```bash
# タスク委託（書き込み可能）
bash scripts/codex-companion.sh task --write "タスク内容"

# stdin 経由（大きなプロンプト向け）
CODEX_PROMPT=$(mktemp /tmp/codex-prompt-XXXXXX.md)
# タスク内容を書き出し
cat "$CODEX_PROMPT" | bash scripts/codex-companion.sh task --write
rm -f "$CODEX_PROMPT"

# 前回スレッドの続行
bash scripts/codex-companion.sh task --resume-last --write "続きをやって"
```

companion は App Server Protocol 経由で Codex と通信し、
Job 管理・thread resume・構造化出力を提供する。
結果を検証し、品質基準を満たさない場合は自力で修正。

### Breezing モード（4 件以上で自動選択 / `--breezing` で強制）

Lead / Worker / Reviewer の役割分離でチーム実行する。
Codex では `spawn_agent`, `wait`, `send_input`, `resume_agent`, `close_agent`
を使った native subagent orchestration を前提にし、
古い TeamCreate / TaskCreate ベースの説明を採らない。

**権限ポリシー**:
- 現行の shipped default は `bypassPermissions`
- `--auto-mode` は互換な親セッション向けの opt-in rollout フラグとして扱う
- `permissions.defaultMode` や agent frontmatter の `permissionMode` には未文書化の `autoMode` 値を書かない

> **CC v2.1.69+**: nested teammates はプラットフォーム側で禁止されるため、
> Worker/Reviewer プロンプトには冗長な nested 防止文言を追加しない。

```
Lead (this agent)
├── Worker (task-worker agent) — 実装担当
└── Reviewer (code-reviewer agent) — レビュー担当
```

**Phase A: Pre-delegate（準備）**:
1. Plans.md を読み込み、対象タスクを特定
2. 依存グラフを解析し、実行順序を決定（Depends カラム）
3. 各タスクの effort スコアリング（ultrathink 注入判定）
4. `scripts/generate-sprint-contract.sh` で `sprint-contract.json` を生成
5. `scripts/enrich-sprint-contract.sh` で Reviewer 観点を加え、`scripts/ensure-sprint-contract-ready.sh` で未承認なら停止

**Phase B: Delegate（Worker spawn → レビュー → cherry-pick）**:

各タスクについて以下を**逐次**実行する（依存順）:

> **API 注記**: 以下は Claude Code の API 構文で記述。
> Codex 環境では `Agent(...)` → `spawn_agent(...)`, `SendMessage(...)` → `send_input(...)` に読み替え。
> 詳細は `team-composition.md` の API マッピング表を参照。

```
for task in execution_order:
    # B-1. sprint-contract を生成
    contract_path = bash("scripts/generate-sprint-contract.sh {task.number}")
    contract_path = bash("scripts/enrich-sprint-contract.sh {contract_path} --check \"DoD を reviewer 観点で確認\" --approve")
    bash("scripts/ensure-sprint-contract-ready.sh {contract_path}")

    # B-2. Worker spawn（フォアグラウンド、worktree 分離）
    # Agent tool の戻り値に agentId が含まれる — 修正ループで SendMessage に使用
    Plans.md: task.status = "cc:WIP"  # 着手時に更新（未着手タスクは cc:TODO のまま）

    worker_result = Agent(
        subagent_type="claude-code-harness:worker",
        prompt="タスク: {task.内容}\nDoD: {task.DoD}\ncontract_path: {contract_path}\nmode: breezing",
        isolation="worktree",
        run_in_background=false  # フォアグラウンドで実行 → Worker 完了まで待機
    )
    worker_id = worker_result.agentId  # SendMessage 用に保持
    # worker_result には {commit, worktreePath, files_changed, summary} が含まれる

    # B-3. Lead がレビュー実行（Codex exec 優先）
    diff_text = git("-C", worker_result.worktreePath, "show", worker_result.commit)
    verdict = codex_exec_review(diff_text) or reviewer_agent_review(diff_text)
    profile = jq(contract_path, ".review.reviewer_profile")
    review_input = "review-output.json"
    if profile == "runtime":
        review_input = bash("cd {worker_result.worktreePath} && scripts/run-contract-review-checks.sh {contract_path}")
        runtime_verdict = jq(review_input, ".verdict")
        if runtime_verdict == "REQUEST_CHANGES":
            verdict = "REQUEST_CHANGES"
        elif runtime_verdict == "DOWNGRADE_TO_STATIC":
            pass  # runtime 検証コマンドなし → static verdict をそのまま使う
    if profile == "browser":
        # browser artifact は PENDING_BROWSER scaffold を生成。
        # 実際の browser 実行は reviewer agent が後続で担当する。
        # review-result には static review の verdict を書く（PENDING_BROWSER ではなく）。
        browser_artifact = bash("scripts/generate-browser-review-artifact.sh {contract_path}")
        # browser artifact は参照用に保存するが、review-result の verdict は static のまま
    # review_input が DOWNGRADE_TO_STATIC の場合は static review 結果を使う
    if review_input != "review-output.json" and jq(review_input, ".verdict") == "DOWNGRADE_TO_STATIC":
        review_input = "review-output.json"  # static review の結果にフォールバック
    bash("scripts/write-review-result.sh {review_input} {latest_commit}")

    # B-4. 修正ループ（REQUEST_CHANGES 時、最大 3 回）
    # Worker はフォアグラウンドで完了済みだが、SendMessage で再開可能
    # （CC: SendMessage(to: agentId) / Codex: resume_agent(agent_id) + send_input）
    review_count = 0
    latest_commit = worker_result.commit
    while verdict == "REQUEST_CHANGES" and review_count < 3:
        SendMessage(to=worker_id, message="指摘内容: {issues}\n修正して amend してください")
        # Worker が修正 → amend → 更新された commit hash を返す
        updated_result = wait_for_response(worker_id)
        latest_commit = updated_result.commit
        diff_text = git("-C", worker_result.worktreePath, "show", latest_commit)
        verdict = codex_exec_review(diff_text) or reviewer_agent_review(diff_text)
        review_count++

    # B-5. APPROVE → main に cherry-pick（feature ブランチ経由）
    # Worker の Branch Guard により main HEAD は動かず、commit は feature ブランチ上にある想定
    if verdict == "APPROVE":
        git checkout main  # safety: 既に main なら no-op
        # feature ブランチの commit が既に main にある（Branch Guard 失敗時のフォールバック）か確認
        if git("merge-base", "--is-ancestor", latest_commit, "HEAD"):
            pass  # 既に main 上 — cherry-pick 不要（再入防止）
        else:
            git cherry-pick --no-commit {latest_commit}  # feature branch → main
            git commit -m "{task.内容}"
        # Worker が作成した feature ブランチを削除
        if worker_result.branch and worker_result.branch not in ["main", "master"]:
            git branch -D {worker_result.branch}
        Plans.md: task.status = "cc:完了 [{hash}]"
        # auto-checkpoint 記録（冪等性ガード (c)）
        # Plans.md 書き換え直後に呼ぶ。失敗しても fail-open（|| true）でループを止めない
        HASH=$(git rev-parse --short HEAD)
        REVIEW_RESULT_PATH=".claude/state/review-results/${task.number}.review-result.json"
        bash scripts/auto-checkpoint.sh \
            "${task.number}" "${HASH}" "${contract_path}" "${REVIEW_RESULT_PATH}" \
            || true  # fail-open: harness-mem 未起動環境でも継続
    else:
        → ユーザーにエスカレーション

    # B-6. Progress feed
    print("📊 Progress: Task {completed}/{total} 完了 — {task.内容}")
```

### Sprint Contract

`sprint-contract` は「このタスクを何で合格にするか」を機械でも人でも同じ意味で読める形にする小さな契約ファイルです。
既定の保存先は `.claude/state/contracts/<task-id>.sprint-contract.json` です。

```bash
scripts/generate-sprint-contract.sh 32.1.1
```

生成物には次を含めます。

- `checks`: DoD を分解した確認項目
- `non_goals`: 今回やらないこと
- `runtime_validation`: test, lint, typecheck などの検証コマンド
- `browser_validation`: browser reviewer が残すべき UI フロー検証項目
- `browser_mode`: `scripted` または `exploratory`
- `route`: browser reviewer が `playwright` / `agent-browser` / `chrome-devtools` のどれを使うか
- `risk_flags`: `needs-spike`, `security-sensitive`, `ux-regression` など
- `reviewer_profile`: `static`, `runtime`, `browser`

**Phase C: Post-delegate（統合・報告）**:
1. 全タスクの commit log を集計
2. **リッチ完了報告**（「完了報告フォーマット」の Breezing テンプレート）を出力
3. Plans.md の最終確認（全タスク cc:完了 になっているか）

## CI 失敗時の対応

CI が失敗した場合:

1. ログを確認してエラーを特定
2. 修正を実施
3. 同一原因で 3 回失敗したら自動修正ループを停止
4. 失敗ログ・試みた修正・残る論点をまとめてエスカレーション

## 失敗タスクの自動再チケット化

タスク完了後にテスト/CI が失敗した場合、修正タスク案を自動生成し、承認後に Plans.md へ反映する:

### トリガー条件

| 条件 | アクション |
|------|----------|
| `cc:完了` 後にテスト失敗 | 修正タスク案を state に保存し、承認を待つ |
| CI 失敗（3回未満） | 修正を実施し、失敗カウントをインクリメント |
| CI 失敗（3回目） | 修正タスク案を提示 + エスカレーション |

### 修正タスクの自動生成

1. 失敗原因を分類（syntax_error / import_error / type_error / assertion_error / timeout / runtime_error）
2. `.claude/state/pending-fix-proposals.jsonl` に修正タスク案を保存:
   - 番号: 元タスク番号 + `.fix` サフィックス（例: `26.1.fix`）
   - 内容: `fix: [元タスク名] - [失敗原因カテゴリ]`
   - DoD: テスト/CI が通ること
   - Depends: 元タスク番号
3. ユーザーが `approve fix <task_id>` を送ると Plans.md に `cc:TODO` で追加
4. `reject fix <task_id>` で提案を破棄。pending が1件だけのときは `yes` / `no` でも応答可能

## レビューループ

実装完了後（ステップ 5 の後）に自動実行される品質検証ステージ。
**全モード共通**（Solo / Parallel / Breezing）で統一的に適用される。
Parallel モードでは各 Worker が step 10（外部レビュー受付）として同じループを実行する。

### レビュー実行の優先順位

```
1. Codex exec（優先）
   ↓ codex コマンドが存在しない or タイムアウト（120s）
2. 内部 Reviewer agent（フォールバック）
```

### APPROVE / REQUEST_CHANGES の判定基準

レビュアーには以下の閾値基準を渡し、**この基準のみ**で verdict を判定させる。
基準外の改善提案は `recommendations` として返すが、verdict には影響しない。

| 重要度 | 定義 | verdict への影響 |
|--------|------|-----------------|
| **critical** | セキュリティ脆弱性、データ損失リスク、本番障害の可能性 | 1 件でも → REQUEST_CHANGES |
| **major** | 既存機能の破壊、仕様との明確な矛盾、テスト不通過 | 1 件でも → REQUEST_CHANGES |
| **minor** | 命名改善、コメント不足、スタイル不統一 | verdict に影響しない |
| **recommendation** | ベストプラクティス提案、将来の改善案 | verdict に影響しない |

> **重要**: minor / recommendation のみの場合は **必ず APPROVE** を返すこと。
> 「あったほうが良い改善」は REQUEST_CHANGES の理由にならない。

### Codex exec レビュー（公式プラグイン経由）

タスク開始時の HEAD を `BASE_REF` として保持し、その ref との差分をレビュー対象にする。
公式プラグイン `codex-plugin-cc` の companion review を使用する。

```bash
# タスク開始時に base ref を記録（Step 2 の cc:WIP 更新前に実行）
BASE_REF=$(git rev-parse HEAD)

# ... 実装完了後 ...

# 公式プラグインの構造化レビューを実行
bash scripts/codex-companion.sh review --base "${BASE_REF}"
REVIEW_EXIT=$?
```

**verdict マッピング**（公式プラグイン → Harness 形式）:

公式プラグインは `review-output.schema.json` 準拠の構造化出力を返す。
Harness の verdict 形式への変換ルール:

| 公式 plugin | Harness | verdict 影響 |
|---|---|---|
| `approve` | `APPROVE` | - |
| `needs-attention` | `REQUEST_CHANGES` | - |
| `findings[].severity: critical` | `critical_issues[]` | 1件でも → REQUEST_CHANGES |
| `findings[].severity: high` | `major_issues[]` | 1件でも → REQUEST_CHANGES |
| `findings[].severity: medium/low` | `recommendations[]` | verdict に影響しない |

AI Residuals スキャンは引き続き `scripts/review-ai-residuals.sh` で実行し、
companion review の結果と合わせて最終 verdict を判定する。

```bash
# AI Residuals スキャン（companion review と並行実行可能）
AI_RESIDUALS_JSON="$(bash scripts/review-ai-residuals.sh --base-ref "${BASE_REF}" 2>/dev/null || echo '{"tool":"review-ai-residuals","scan_mode":"diff","base_ref":null,"files_scanned":[],"summary":{"verdict":"APPROVE","major":0,"minor":0,"recommendation":0,"total":0},"observations":[]}')"
```

### 内部 Reviewer agent フォールバック

Codex exec が使えない場合（`command -v codex` が失敗、または exit code ≠ 0）:

```
Agent tool: subagent_type="reviewer"
prompt: "以下の変更をレビューしてください。判定基準: critical/major → REQUEST_CHANGES、minor/recommendation のみ → APPROVE。diff: {git diff ${BASE_REF}}"
```

Reviewer agent は Read-only（Write/Edit/Bash 無効）で安全にレビューを実行する。

### 修正ループ（REQUEST_CHANGES 時）

```
review_count = 0
MAX_REVIEWS = 3

while verdict == "REQUEST_CHANGES" and review_count < MAX_REVIEWS:
    1. レビュー指摘を解析（critical / major のみ対象）
    2. 各指摘に対して修正を実装
    3. 再度レビューを実行（同じ判定基準・同じ優先順位）
    review_count++

if review_count >= MAX_REVIEWS and verdict != "APPROVE":
    → ユーザーにエスカレーション
    → 「3 回修正しましたが以下の critical/major 指摘が残っています」+ 指摘一覧を表示
    → ユーザー判断を待つ（続行 / 中断）
```

### Breezing モードでの適用

Breezing モードでは **Lead** がレビューループを実行する（上記 Phase B 参照）:

1. Worker が worktree 内で実装・commit → Lead に結果返却
2. Lead が Codex exec でレビュー（優先）/ Reviewer agent（フォールバック）
3. REQUEST_CHANGES → Lead が SendMessage で Worker に修正指示 → Worker が amend
4. 修正後、再レビュー（最大 3 回）
5. APPROVE → Lead が main に cherry-pick → Plans.md を `cc:完了 [{hash}]` に更新

## 完了報告フォーマット

タスク完了時（`cc:完了` + commit 後）に自動出力される視覚的サマリ。
非専門家にも変更内容と影響が伝わることを目的とする。

### テンプレート

```
┌─────────────────────────────────────────────┐
│  ✓ Task {N} 完了: {タスク名}                    │
├─────────────────────────────────────────────┤
│                                              │
│  ■ 何をしたか                                 │
│    • {変更内容 1}                              │
│    • {変更内容 2}                              │
│                                              │
│  ■ 何が変わるか                                │
│    Before: {旧動作}                            │
│    After:  {新動作}                            │
│                                              │
│  ■ 変更ファイル ({N} files)                    │
│    {ファイルパス 1}                             │
│    {ファイルパス 2}                             │
│                                              │
│  ■ 残りの課題                                  │
│    • Task {X} ({status}): {内容}  ← Plans.md  │
│    • Task {Y} ({status}): {内容}  ← Plans.md  │
│    （Plans.md に {M} 件の未完了タスクあり）       │
│                                              │
│  commit: {hash} | review: {APPROVE}           │
└─────────────────────────────────────────────┘
```

### 生成ルール

1. **何をしたか**: `git diff --stat HEAD~1` と commit message から自動抽出。技術用語は最小限にし、動詞で始める
2. **何が変わるか**: タスクの「内容」と「DoD」から Before/After を推論。ユーザー体験の変化を重視
3. **変更ファイル**: `git diff --name-only HEAD~1` から取得。5 ファイル超は省略して件数表示
4. **残りの課題**: Plans.md の `cc:TODO` / `cc:WIP` タスクを一覧表示。Plans.md に記載済みかどうかを明示
5. **review**: レビュー結果（APPROVE / REQUEST_CHANGES → APPROVE）を表示

### Parallel モードでの報告

- **1 タスク**（`--parallel` 強制時）: Solo テンプレートを使用
- **複数タスク**: Breezing 集約テンプレートを使用（下記参照）

### Breezing モードでの報告

全タスク完了後にまとめて出力。各タスクは簡略版（何をしたか + commit hash のみ）で一覧し、
最後に全体サマリ（合計変更ファイル数 + 残り課題）を出力する:

```
┌─────────────────────────────────────────────┐
│  ✓ Breezing 完了: {N}/{M} タスク             │
├─────────────────────────────────────────────┤
│                                              │
│  1. ✓ {タスク名 1}            [{hash1}]      │
│  2. ✓ {タスク名 2}            [{hash2}]      │
│  3. ✓ {タスク名 3}            [{hash3}]      │
│                                              │
│  ■ 全体の変更                                 │
│    {N} files changed, {A} insertions(+),     │
│    {D} deletions(-)                          │
│                                              │
│  ■ 残りの課題                                  │
│    Plans.md に {K} 件の未完了タスクあり         │
│    • Task {X}: {内容}                         │
│                                              │
└─────────────────────────────────────────────┘
```

## 関連スキル

- `harness-plan` — 実行するタスクを計画する
- `harness-sync` — 実装と Plans.md を同期する
- `harness-review` — 実装のレビュー
- `harness-release` — バージョンバンプ・リリース
