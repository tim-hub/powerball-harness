# Planning Discussion (Phase 0)

`--discuss` フラグ使用時の計画議論フェーズ仕様。
Agent Teams を活用して、実行前にタスク分解の妥当性を議論・精査する。

## 概要

```text
/breezing --discuss all
    ↓
Phase 0: Planning Discussion（実行前の計画議論）
  ├── Planner: タスク分析・粒度精査・owns 推定・リスク評価
  ├── Critic: Red Teaming 視点で計画を批判
  └── Lead: 議論を調整、精査済み計画をユーザーに提示
    ↓
Phase A: Pre-delegate（通常フローに合流）
```

## 有効化条件

| 条件 | Phase 0 起動 |
|---|---|
| `--discuss` フラグ指定 | 常に起動 |
| タスク粒度バリデーション(V1〜V4)で warning 3+ | 自動起動を推奨（Lead 判断） |
| タスク数 10+ | 自動起動を推奨（Lead 判断） |
| フラグなし + 問題なし | スキップ（直接 Phase A へ） |

## Team 構成（Phase 0 限定）

```text
Lead ─── 議論の調整、最終判断
  │
  ├── Planner (sonnet) ─── タスク分析・依存推定・粒度精査
  │     subagent_type: claude-code-harness:project-analyzer
  │     mode: "bypassPermissions"
  │
  └── Critic (sonnet) ─── Red Teaming・批判的検証
        subagent_type: claude-code-harness:plan-critic
        mode: "bypassPermissions"
```

> **注**: Phase 0 の Planner/Critic は Phase B の Implementer/Reviewer とは**別の Teammate**。
> Phase 0 完了時に shutdown し、Phase A で新しい Team を構築する。

## 議論フロー

### Round 1: Planner の初期分析

```text
Lead → SendMessage → Planner:
  「Plans.md のタスク 4.1〜4.5 を分析してください。
   以下を報告:
   ・各タスクの推定 owns ファイル
   ・依存関係の提案
   ・粒度の妥当性評価
   ・リスクの高いタスク」

Planner → SendMessage → Lead:
  {
    "tasks": [
      {
        "id": "4.1",
        "title": "ログイン機能の実装",
        "estimated_owns": ["src/components/LoginForm.tsx", "src/app/api/auth/login/route.ts"],
        "granularity": "appropriate",
        "risk": "low",
        "notes": "独立タスク、並列実行可能"
      },
      {
        "id": "4.3",
        "title": "パフォーマンス改善",
        "estimated_owns": ["unknown"],
        "granularity": "too_vague",
        "risk": "medium",
        "notes": "対象ファイルと改善メトリクスが不明"
      }
    ],
    "proposed_dependencies": [
      {"from": "4.1", "to": "4.2", "reason": "4.2 が 4.1 の認証 API に依存"}
    ]
  }
```

### Round 2: Critic の批判的レビュー

```text
Lead → SendMessage → Critic:
  「Planner の分析結果を踏まえて、計画を批判的に検証してください。
   Red Teaming チェックリスト（ゴール達成性・粒度・依存・並列化・リスク・代替案）
   の各観点で問題を指摘してください。」

Critic → SendMessage → Lead:
  {
    "assessment": "revise_recommended",
    "findings": [
      {
        "severity": "warning",
        "category": "granularity",
        "task": "4.3",
        "issue": "「パフォーマンス改善」は受入条件が不明",
        "suggestion": "具体的なメトリクス（レスポンスタイム < 200ms 等）と対象ファイルを明示"
      },
      {
        "severity": "warning",
        "category": "goal_coverage",
        "issue": "テストタスクが計画に含まれていない",
        "suggestion": "各機能に対するテスト作成タスクを追加"
      },
      {
        "severity": "info",
        "category": "parallelism",
        "issue": "5タスク中3タスクが依存チェーンを形成、Implementer #2 のアイドル時間が長い",
        "suggestion": "タスク 4.4 を独立化できないか検討"
      }
    ],
    "parallelism_score": "medium",
    "summary": "概ね妥当だが、タスク 4.3 の具体化とテストタスクの追加が推奨"
  }
```

### Round 3: Lead の統合判断

Lead は Planner/Critic の分析を統合し、以下のいずれかを判断:

| Critic の assessment | Lead のアクション |
|---|---|
| `approve` | Phase 0 完了 → Phase A へ |
| `revise_recommended` | ユーザーに修正提案を表示、判断を委ねる |
| `revise_required` | ユーザーに修正を強く推奨（critical findings あり） |

### Round 3+: 追加ラウンド（必要に応じて）

ユーザーが Plans.md を修正した場合:
1. Lead が修正内容を Planner/Critic に共有
2. Planner が再分析
3. Critic が再レビュー
4. Lead が最終判断

**最大ラウンド数**: 3（超過時は現状の計画で Phase A に進む）

## ユーザー提示フォーマット

```text
🏇 Breezing - 計画議論の結果

## Planner の分析
- タスク 4.1, 4.2, 4.4: 粒度・明確性 OK
- タスク 4.3: ⚠️ 受入条件不明 → 具体化推奨
- タスク 4.5: ✅ 独立タスク、並列実行可能

## Critic の指摘
- ⚠️ テストタスクが計画に含まれていない
- ⚠️ タスク 4.3 の対象ファイル/メトリクスが不明
- ℹ️ 並列度が中程度（改善の余地あり）

## 提案
1. タスク 4.3 を「src/db/users.ts の N+1 クエリ解消」に具体化
2. テストタスクを各機能に追加

修正しますか？ (修正 / そのまま続行 / 中止)
```

## Phase 0 完了後の引き継ぎ

Phase 0 で得られた情報は Phase A に引き継ぐ:

```text
Phase 0 → Phase A への引き継ぎ:
  1. Planner の estimated_owns → Phase A Step 3 の owns 推定に活用
  2. Planner の proposed_dependencies → Phase A Step 3 の addBlockedBy に反映
  3. Critic の findings → バリデーション済みとして V1〜V4 チェックをスキップ可能
  4. 修正された Plans.md → Phase A の入力として使用
```

## breezing-active.json への記録

Phase 0 を実行した場合、メタデータに記録:

```json
{
  "planning_discussion": {
    "enabled": true,
    "rounds": 2,
    "critic_assessment": "approve",
    "findings_resolved": 2,
    "findings_accepted": 1
  }
}
```

## コスト考慮

Phase 0 は追加の Teammate spawn を伴うため、トークンコストが増加する:

| Phase | 追加コスト（vs Phase 0 なし） |
|---|---|
| Phase 0 (2 rounds) | +1.5x〜2x（Planner + Critic の 2 Teammate） |
| Phase 0 (3 rounds) | +2x〜2.5x |

**トレードオフ**: Phase 0 のコスト増 vs Phase B でのリテイク削減。
タスク数が多いほど、Phase 0 の投資対効果が高くなる。
