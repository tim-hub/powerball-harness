# `/harness-work all` Evidence Pack

Last updated: 2026-03-06

This evidence pack is the minimum set for verifying `/harness-work all` claims by examining "what remains after execution."
The current assumption is a new contract where Worker self-inspection alone is not sufficient for completion; tasks must pass through `sprint-contract` and an independent review artifact before being marked complete.

## What is included

| Scenario | Goal | Expected result |
|----------|------|-----------------|
| success | Complete a small TODO repo with `work all` | Tests go green, additional commits are created |
| failure | Submit an impossible task to verify the quality gate | Tests remain failed, no additional commits are created |

## Fixtures

- `tests/fixtures/work-all-success/`
- `tests/fixtures/work-all-failure/`

Both are designed so that `npm test` fails at baseline.

## Smoke vs Full

| Mode | Command | What it does |
|------|---------|--------------|
| CI smoke | `./scripts/evidence/run-work-all-smoke.sh` | Verifies fixture integrity and baseline failure, leaves a Claude execution command preview |
| Local full | `./scripts/evidence/run-work-all-success.sh --full` | Executes the success scenario with Claude CLI; falls back to replay overlay for artifact completion on rate limit |
| Local full (strict) | `./scripts/evidence/run-work-all-success.sh --full --strict-live` | Proves success using only live Claude execution without replay |
| Local full | `./scripts/evidence/run-work-all-failure.sh --full` | Executes the failure scenario with Claude CLI; verifies that no new commits are created |

Artifacts are saved to `out/evidence/work-all/` by default.

## Prerequisites for full runs

- `claude --version` must work (required for strict-live mode)
- Must be authenticated with Claude Code
- Must run from the root of this repo

Full mode internally uses the following command:

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
- `sprint-contract.json` or contract generation log
- `review-result.json`
- `fallback-reason.txt`
- `rate-limit-detected.txt`
- `replay.log` (when rate limit fallback occurs)

## Interpretation

- If success shows `post_test_status=0` and `final_commits > baseline_commits`, it serves as evidence that the minimum scenario "ran to completion and reached a commit"
- If `review-result.json` also shows `APPROVE`, it serves as evidence that completion "passed an independent review"
- If failure shows `post_test_status!=0` and `final_commits == baseline_commits`, it serves as at minimum evidence that "failures were not hidden and no commits were made"
- If test tampering occurs in the failure fixture, it will remain in the diff artifact, making quality gate behavior easy to review

## Live vs Replay

- `execution_mode=live` means the Claude CLI completed the success scenario as-is
- `execution_mode=replay-after-rate-limit` means the Claude execution was stopped by rate limits, and the replay overlay bundled with the fixture was applied to produce the happy path artifact
- To claim "proven with a live Claude run" in public copy, a separate `--strict-live` success artifact is needed
