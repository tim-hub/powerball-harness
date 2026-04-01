# `/harness-work all` Evidence Pack

最終更新: 2026-03-06

この evidence pack は、`/harness-work all` の主張を「実行したら何が残るか」で確認するための最小セットです。
現在の前提は、Worker の自己点検だけでは完了にせず、`sprint-contract` と独立 review artifact を通してから完了する、という新しい契約です。

## What is included

| Scenario | Goal | Expected result |
|----------|------|-----------------|
| success | 小さな TODO repo を `work all` で完了させる | テストが green になり、追加コミットが残る |
| failure | 不可能なタスクを投げて quality gate を確認する | テストは fail のまま、追加コミットは作られない |

## Fixtures

- `tests/fixtures/work-all-success/`
- `tests/fixtures/work-all-failure/`

どちらも baseline では `npm test` が失敗するように作ってあります。

## Smoke vs Full

| Mode | Command | What it does |
|------|---------|--------------|
| CI smoke | `./scripts/evidence/run-work-all-smoke.sh` | fixture の整合と baseline failure を確認し、Claude 実行コマンド preview を残す |
| Local full | `./scripts/evidence/run-work-all-success.sh --full` | Claude CLI で success scenario を実行し、rate limit 時は replay overlay で artifact を完成させる |
| Local full (strict) | `./scripts/evidence/run-work-all-success.sh --full --strict-live` | replay を使わず、live Claude 実行だけで success を証明する |
| Local full | `./scripts/evidence/run-work-all-failure.sh --full` | Claude CLI で failure scenario を実行し、commit が増えないことを確認する |

artifact は既定で `out/evidence/work-all/` に保存されます。

## Prerequisites for full runs

- `claude --version` が通ること（strict live を使う場合は必須）
- Claude Code で認証済みであること
- この repo の root から実行すること

full mode は次のコマンドを内部で使います。

```bash
claude --plugin-dir /path/to/claude-code-harness \
  --dangerously-skip-permissions \
  --output-format json \
  --no-session-persistence \
  -p "$(cat PROMPT.md)"
```

## Saved artifacts

- `baseline-test.log`
- `claude-stdout.json`
- `claude-stderr.log`
- `elapsed-seconds.txt`
- `git-status.txt`
- `git-diff-stat.txt`
- `git-diff.patch`
- `git-log.txt`
- `commit-count.txt`
- `result.txt`
- `execution-mode.txt`
- `sprint-contract.json` または contract 生成ログ
- `review-result.json`
- `fallback-reason.txt`
- `rate-limit-detected.txt`
- `replay.log`（rate limit fallback が発生したとき）

## Interpretation

- success で `post_test_status=0` かつ `final_commits > baseline_commits` なら、最小シナリオでは「完走して commit まで到達した」証拠になる
- さらに `review-result.json` が `APPROVE` なら、「独立 review を通して完了した」証拠になる
- failure で `post_test_status!=0` かつ `final_commits == baseline_commits` なら、少なくとも「失敗を隠して commit はしなかった」証拠になる
- 失敗 fixture でテスト改ざんが起きた場合も diff artifact に残るので、quality gate の振る舞いをレビューしやすい

## Live vs Replay

- `execution_mode=live` なら、Claude CLI がそのまま success scenario を完走した artifact
- `execution_mode=replay-after-rate-limit` なら、Claude 実行は rate limit で止まり、fixture に同梱した replay overlay を適用して happy path artifact を作ったことを示す
- 公開文面で「live Claude run で証明済み」と言いたい場合は `--strict-live` の成功 artifact を別途取る
