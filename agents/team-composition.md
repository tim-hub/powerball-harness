# Team Composition

Harness の標準チーム構成は 5 ロール。
実装系の teammate を増やす時も、この 5 ロールの責務境界は変えない。

## 構成図

```text
Lead
├── Worker x 1..3
├── Advisor x 0..1
├── Reviewer x 1
└── Scaffolder x 0..1
```

## spawn 権限

- Lead だけが teammate を spawn する
- Worker は teammate を spawn しない
- Reviewer は teammate を spawn しない
- Scaffolder は teammate を spawn しない
- Worker が相談したい時は subagent を増やさず `advisor-request.v1` を返す

## role contract

| Role | subagent_type | 数 | 使うツール | 返すもの |
|------|---------------|----|------------|----------|
| Lead | Execute skill 内部 | 1 | Agent, SendMessage, Bash | task 分解、review 判定、main 反映 |
| Worker | `claude-code-harness:worker` | 1..3 | Read, Write, Edit, Bash, Grep, Glob | 実装結果または `advisor-request.v1` |
| Advisor | `claude-code-harness:advisor` | 0..1 | Read, Grep, Glob | `advisor-response.v1` |
| Reviewer | `claude-code-harness:reviewer` | 1 | Read, Grep, Glob | `review-result.v1` |
| Scaffolder | `claude-code-harness:scaffolder` | 0..1 | Read, Write, Edit, Bash, Grep, Glob | analyze/scaffold/update-state の結果 JSON |

## worker 数の決め方

| 条件 | worker 数 |
|------|-----------|
| 書き込み対象ファイルが 1 グループ、またはファイルが重なる | 1 |
| 書き込み対象ファイルが 2 グループで、互いに重ならない | 2 |
| 書き込み対象ファイルが 3 グループ以上で、互いに重ならない | 3 |

ここでいう「グループ」は、同じ commit にまとめても競合しない書き込み集合を指す。
同じファイルを 2 worker に書かせる分割は禁止。

## 実行フロー

1. Lead が task を分解し、`sprint-contract` を作る
2. Lead が worker を spawn する
3. Worker が実装、preflight、検証、commit 準備を行う
4. Worker が相談条件に当たった時だけ `advisor-request.v1` を返す
5. Lead が Advisor を呼び、`advisor-response.v1` を同じ Worker に返す
6. Worker が結果を返したら Lead が review を実行する
7. `APPROVE` の時だけ Lead が main へ反映する

## review loop

| 条件 | Lead の動き |
|------|-------------|
| `review-result.v1.verdict == APPROVE` | cherry-pick して main に commit |
| `review-result.v1.verdict == REQUEST_CHANGES` | 同じ Worker に修正依頼を返す |

修正ループは最大 3 回。
4 回目には入らず、Lead が task をエスカレーションする。

## SendMessage の固定パターン

Lead が Worker に修正を返す時は、次の構文を使う。

```text
SendMessage(
  to: "{worker_agent_id}",
  message: "以下の critical/major 指摘を修正してください:\n\n{issues}\n\n修正後 git commit --amend して完了を返してください。"
)
```

## breezing 時の main 反映

Worker は worktree または feature branch で commit する。
Lead は `APPROVE` 後に次の 2 コマンドで main へ取り込む。

```bash
git cherry-pick --no-commit {worktree_commit_hash}
git commit -m "feat: {task_description}"
```

Lead が main に反映するまでは、Worker は Plans.md を `cc:完了` に更新しない。

## Advisor の境界

- Advisor は `PLAN | CORRECTION | STOP` だけ返す
- Advisor は `APPROVE | REQUEST_CHANGES` を返さない
- Advisor はコードを編集しない
- Reviewer は advisor の提案文ではなく、最終成果物だけを見る

## Codex bridge

Claude Code から Codex へ委譲する時の標準コマンドは次の 2 つだけ。

```bash
bash scripts/codex-companion.sh task --write "タスク内容"
bash scripts/codex-companion.sh review --base "${TASK_BASE_REF}"
```

raw `codex exec` をチーム標準手順として書かない。

## 2.1.111 優先ルール

- `xhigh` は caller 側の推論強度指定。worker prompt が文字列から自動判定しない
- `/ultrareview` は caller 側の review entrypoint。review artifact の契約は `review-result.v1` のまま
- `--auto-mode` は opt-in rollout。shipped default にしない

## permission mode

現行 shipped default は `bypassPermissions`。
理由は、teammate 実行時の権限継承を agent frontmatter と一致させるため。

| レイヤー | 現行値 |
|---------|--------|
| project template | `bypassPermissions` |
| worker frontmatter | `bypassPermissions` |
| reviewer frontmatter | `bypassPermissions` |
| advisor frontmatter | `bypassPermissions` |
| scaffolder frontmatter | `bypassPermissions` |

`--auto-mode` は rollout 用の opt-in。
既定値にはしない。

## チームサイズ

- 標準は 3 から 5 teammate
- Harness の通常構成は `Worker 1..3 + Reviewer 1`
- Advisor と Scaffolder は必要時のみ追加する
