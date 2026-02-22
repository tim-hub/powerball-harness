# Planning Discussion (Phase 0) - Codex Native

計画議論フェーズ仕様（デフォルト実行、`--no-discuss` でスキップ）。
Codex マルチエージェント API を使用して、実行前にタスク分解の妥当性を議論・精査する。

## 概要

```text
breezing
    ↓
Phase 0: Planning Discussion（デフォルトで実行）
  ├── plan_analyst: タスク分析・粒度精査・owns 推定・リスク評価
  ├── plan_critic:  Red Teaming 視点で計画を批判
  └── Lead: 議論を調整、精査済み計画をユーザーに提示
    ↓
Phase A: Pre-delegate（通常フローに合流）
```

## 有効化条件

| 条件 | Phase 0 起動 |
|---|---|
| フラグなし（デフォルト） | 常に起動 |
| `--no-discuss` フラグ指定 | スキップ（直接 Phase A へ） |

> **注**: `--no-discuss` でスキップした場合でも、Phase A の V1〜V4 バリデーションは実行される。
> Phase 0 は戦略/アーキテクチャ評価、V1〜V4 は技術的詳細チェックで役割が異なる。

## エージェント構成（Phase 0 限定）

```text
Lead ─── 議論の調整、最終判断
  │
  ├── plan_analyst ─── タスク分析・依存推定・粒度精査
  │     config.toml: [agents.plan_analyst]
  │     spawn: spawn_agent("plan_analyst", analyst_task)
  │
  └── plan_critic ─── Red Teaming・批判的検証
        config.toml: [agents.plan_critic]
        spawn: spawn_agent("plan_critic", critic_task)
```

> **注**: Phase 0 の plan_analyst/plan_critic は Phase B の implementer/reviewer とは**別のエージェント**。
> Phase 0 完了時に `close_agent()` で終了し、Phase A で新しいエージェントを spawn する。

## Codex マルチエージェント API

```python
# エージェント生成
analyst_id = spawn_agent("plan_analyst", task)
critic_id  = spawn_agent("plan_critic", task)

# エージェントへの入力送信
send_input(analyst_id, message)
send_input(critic_id, message)

# 応答待機（タイムアウト: ms）
analyst_response = wait(analyst_id, timeout_ms=120000)
critic_response  = wait(critic_id, timeout_ms=120000)

# エージェント再開（中断からの継続）
resume_agent(analyst_id, input)

# エージェント終了
close_agent(analyst_id)
close_agent(critic_id)
```

## 議論フロー

### Phase 0 初期化

```python
# Plans.md を読み込み、タスクリストを抽出
plans_content = read_file("Plans.md")

# Round 1 タスクを準備
analyst_task = f"""
あなたは plan_analyst エージェントです（Read-only）。
以下の Plans.md を分析し、JSON で報告してください。

分析観点:
1. 各タスクの推定 owns ファイル (Glob/Grep/Read で調査)
2. 依存関係の提案
3. 粒度の妥当性評価 (appropriate/too_broad/too_vague/too_small)
4. リスク評価 (high/medium/low)

Plans.md:
{plans_content}

報告フォーマット:
{{
  "tasks": [
    {{
      "id": "タスクID",
      "title": "タスク名",
      "estimated_owns": ["ファイルパス"],
      "granularity": "appropriate",
      "risk": "low",
      "notes": "メモ"
    }}
  ],
  "proposed_dependencies": [
    {{"from": "A", "to": "B", "reason": "理由"}}
  ],
  "parallelism_assessment": {{
    "independent_tasks": N,
    "max_parallel": N,
    "bottleneck": "ボトルネックの説明"
  }}
}}
"""

analyst_id = spawn_agent("plan_analyst", analyst_task)
```

### Round 1: plan_analyst の初期分析

```python
# plan_analyst の分析完了を待機
analyst_result = wait(analyst_id, timeout_ms=120000)

# analyst_result 例:
# {
#   "tasks": [
#     {
#       "id": "4.1",
#       "title": "ログイン機能の実装",
#       "estimated_owns": ["src/components/LoginForm.tsx", "src/app/api/auth/login/route.ts"],
#       "granularity": "appropriate",
#       "risk": "low",
#       "notes": "独立タスク、並列実行可能"
#     },
#     {
#       "id": "4.3",
#       "title": "パフォーマンス改善",
#       "estimated_owns": ["unknown"],
#       "granularity": "too_vague",
#       "risk": "medium",
#       "notes": "対象ファイルと改善メトリクスが不明"
#     }
#   ],
#   "proposed_dependencies": [
#     {"from": "4.1", "to": "4.2", "reason": "4.2 が 4.1 の認証 API に依存"}
#   ]
# }
```

### Round 2: plan_critic の批判的レビュー + エージェント間直接対話

**設計判断**: `send_input` を使ったエージェント間の非同期対話（intra-round discussion）を中核に据える。
Lead を仲介せず、plan_analyst と plan_critic が直接やりとりして疑問点を解消する。

```python
# plan_critic を spawn（analyst 結果を初期コンテキストとして渡す）
critic_task = f"""
あなたは plan_critic エージェントです（Read-only）。
plan_analyst の分析結果を踏まえ、Red Teaming 視点で計画を批判的に検証してください。

plan_analyst の分析結果:
{analyst_result}

Plans.md:
{plans_content}

不明点は plan_analyst に直接質問してください。
質問する場合は "QUESTION_TO_ANALYST: <質問内容>" の形式で出力してください。

報告フォーマット:
{{
  "assessment": "approve | revise_recommended | revise_required",
  "findings": [
    {{
      "severity": "critical | warning | info",
      "category": "goal_coverage | granularity | dependency | parallelism | risk",
      "task": "タスクID（任意）",
      "issue": "問題の説明",
      "suggestion": "改善提案"
    }}
  ],
  "planner_consultations": N,
  "parallelism_score": "high | medium | low",
  "summary": "総評"
}}
"""

critic_id = spawn_agent("plan_critic", critic_task)

# エージェント間対話ループ（最大 3 往復）
max_consultations = 3
consultations = 0

while consultations < max_consultations:
    critic_output = wait(critic_id, timeout_ms=60000)

    # critic が analyst に質問している場合
    if "QUESTION_TO_ANALYST:" in critic_output:
        question = extract_question(critic_output)  # "QUESTION_TO_ANALYST:" 以降を抽出

        # analyst に質問を転送
        send_input(analyst_id, f"plan_critic からの質問: {question}")
        analyst_answer = wait(analyst_id, timeout_ms=60000)

        # analyst の回答を critic に送信
        send_input(critic_id, f"plan_analyst の回答: {analyst_answer}")
        consultations += 1
    else:
        # 最終報告（QUESTION_TO_ANALYST がない = 完了）
        critic_result = critic_output
        break

# critic 対話ループ例:
# critic → "QUESTION_TO_ANALYST: タスク 4.2 が 4.1 の認証 API に依存するとのことですが、
#            4.2 は JWT 検証だけなので独立実装可能では？"
# analyst → "src/middleware.ts を確認したところ、4.1 で作成する loginHandler の
#             レスポンス型を 4.2 の JWT 検証が参照しています。依存は正当です。"
# critic → {最終 JSON 報告}
```

### Round 3: Lead の統合判断

```python
# critic の assessment に基づき判断
assessment = parse_assessment(critic_result)

if assessment == "approve":
    # Phase 0 完了 → Phase A へ
    pass

elif assessment == "revise_recommended":
    # ユーザーに修正提案を表示、判断を委ねる
    print_review_summary(analyst_result, critic_result)
    user_choice = prompt_user("修正しますか？ (修正 / そのまま続行 / 中止)")

    if user_choice == "修正":
        # ユーザーが Plans.md を修正後、追加ラウンドを実行
        pass  # → Round 3+ フロー

    elif user_choice == "そのまま続行":
        pass  # → Phase A へ

    else:  # 中止
        close_agent(analyst_id)
        close_agent(critic_id)
        return

elif assessment == "revise_required":
    # critical findings あり、ユーザーに強く推奨
    print_review_summary(analyst_result, critic_result)
    user_choice = prompt_user("critical な問題が検出されました。修正を強く推奨します。(修正 / 強行 / 中止)")
    # 以降は revise_recommended と同様

# Phase 0 終了処理
close_agent(analyst_id)
close_agent(critic_id)
```

### Round 3+: 追加ラウンド（必要に応じて）

ユーザーが Plans.md を修正した場合:

```python
max_rounds = 3
current_round = 1

while current_round < max_rounds:
    updated_plans = read_file("Plans.md")

    # analyst に再分析を依頼
    send_input(analyst_id, f"Plans.md が更新されました。再分析してください。\n\n{updated_plans}")
    analyst_result = wait(analyst_id, timeout_ms=120000)

    # critic に再レビューを依頼
    send_input(critic_id, f"analyst の再分析結果です。再レビューしてください。\n\n{analyst_result}")
    critic_result = wait(critic_id, timeout_ms=120000)

    assessment = parse_assessment(critic_result)
    if assessment == "approve":
        break

    current_round += 1

# 最大ラウンド超過時は現状の計画で Phase A に進む
```

## ユーザー提示フォーマット

```text
Breezing - 計画議論の結果

## Planner の分析
- タスク 4.1, 4.2, 4.4: 粒度・明確性 OK
- タスク 4.3: [WARNING] 受入条件不明 → 具体化推奨
- タスク 4.5: [OK] 独立タスク、並列実行可能

## Critic の指摘
- [WARNING] テストタスクが計画に含まれていない
- [WARNING] タスク 4.3 の対象ファイル/メトリクスが不明
- [INFO] 並列度が中程度（改善の余地あり）

## 提案
1. タスク 4.3 を「src/db/users.ts の N+1 クエリ解消」に具体化
2. テストタスクを各機能に追加

修正しますか？ (修正 / そのまま続行 / 中止)
```

## Phase 0 完了後の引き継ぎ

Phase 0 で得られた情報は Phase A に引き継ぐ:

```text
Phase 0 → Phase A への引き継ぎ:
  1. analyst の estimated_owns → Phase A Step 3 の owns 推定に活用（Glob 再検索を省略可能）
  2. analyst の proposed_dependencies → Phase A Step 3 の addBlockedBy に反映
  3. critic の findings → V1〜V4 バリデーションの参考情報（スキップはしない）
     ※ Phase 0 は戦略/アーキテクチャ評価、V1〜V4 は技術的詳細チェック。
       役割が異なるため、Phase 0 を経てもバリデーションは必ず実行する。
  4. 修正された Plans.md → Phase A の入力として使用
```

## breezing-active.json への記録

Phase 0 を実行した場合、メタデータに記録:

```json
{
  "planning_discussion": {
    "enabled": true,
    "rounds": 3,
    "critic_assessment": "approve",
    "findings_resolved": 2,
    "findings_accepted": 1,
    "handoff": {
      "estimated_owns": {
        "4.1": ["src/components/LoginForm.tsx", "src/app/api/auth/login/route.ts"],
        "4.2": ["src/middleware.ts", "src/lib/auth.ts"]
      },
      "proposed_dependencies": [
        {"from": "4.1", "to": "4.2", "reason": "loginHandler レスポンス型依存"}
      ],
      "findings_digest": ["4.3 の受入条件を具体化済み", "テストタスク追加済み"]
    }
  }
}
```

> **Compaction 対策**: `handoff` フィールドに Phase 0 の成果物（owns 推定、依存提案、
> findings の要約）を永続化する。コンテキスト圧縮が Phase 0 と Phase A の間に発生しても、
> breezing-active.json から復元可能。

## コスト考慮

Phase 0 は追加のエージェント spawn を伴うため、トークンコストが増加する:

| Phase | 追加コスト（vs Phase 0 なし） |
|---|---|
| Phase 0 (2 rounds) | +1.5x〜2x（plan_analyst + plan_critic の 2 エージェント） |
| Phase 0 (3 rounds) | +2x〜2.5x |

**トレードオフ**: Phase 0 のコスト増 vs Phase B でのリテイク削減。
タスク数が多いほど、Phase 0 の投資対効果が高くなる。

## Claude 版との差分

| 観点 | Claude 版 | Codex 版 |
|---|---|---|
| エージェント生成 | Task tool (role: plan-analyst) | `spawn_agent("plan_analyst", task)` |
| メッセージ送信 | `SendMessage` tool | `send_input(agent_id, message)` |
| 応答待機 | Task tool の完了を待つ | `wait(agent_id, timeout_ms)` |
| エージェント間対話 | Teammate 直接対話（SendMessage） | Lead が中継: `send_input` → `wait` → `send_input` |
| エージェント終了 | shutdown_request | `close_agent(agent_id)` |
| エージェント定義 | `agents/*.md` ファイル | `config.toml` の `[agents.*]` セクション |
| 権限制御 | `mode: "bypassPermissions"` | config.toml のエージェント定義で制御 |
