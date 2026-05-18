# Quick / Codex Closeout

## Summary

For small changes: fix the scope, verify Codex advice against the actual code, and stop when the result is clean.

## Target selection decision tree

1. Working tree is dirty
   - Recommended: uncommitted changes only
   - base: `HEAD`
   - Include untracked files
2. PR branch / feature branch has commits
   - Recommended: `upstream..HEAD` or `origin/main..HEAD`
   - If working tree is also dirty, use `AskUserQuestion` to choose: uncommitted only / all / commits only
3. Clean tree with no branch diff
   - Recommended: most recent 1 commit
   - Most recent 5 commits if needed
4. User specified `--base` / `--commit`
   - Honor the explicit specification

## Advisory rule

Codex findings are advisory — they are reference opinions, not facts.

Always do the following:

- Read the cited location in the actual code
- Confirm reproducibility with the diff and tests
- Separate into accepted findings / rejected findings
- For each rejected finding, document why it was not adopted

## Stop-on-clean

Stop-on-clean: do not run additional reviews just for appearance after a clean result.

Example:

- Codex review: no major issues
- Focused tests: pass
- Manual spot check: pass

Stop in this state. Run additional heavy reviews only before release, for security-sensitive changes, spec source-of-truth changes, or when the user explicitly requests one.

## Helper contract

`harness/scripts/harness-review-closeout.sh` is a helper that locks down the execution plan for lightweight closeout.

Supported inputs:

- `--dry-run`
- `--parallel-tests`
- `--base REF`
- `--commit REF`
- `--uncommitted`
- `--test CMD`
- `--json`

Examples:

```bash
bash harness/scripts/harness-review-closeout.sh --dry-run --uncommitted
bash harness/scripts/harness-review-closeout.sh --base origin/main --parallel-tests --test "bash harness/tests/test-harness-review-governance.sh"
bash harness/scripts/harness-review-closeout.sh --commit HEAD --json
```

If Codex is unavailable:

- Fall back to full manual pass
- Do not treat failure as success
- Record `codex_available: false` in the final report

## Final report

Required fields:

- review command
- tests
- accepted findings
- rejected findings
- clean result
- fallback reason

Minimum JSON structure:

```json
{
  "schema_version": "harness-review-closeout.v1",
  "target": "working_tree | branch_range | commit",
  "base_ref": "HEAD",
  "review_command": "bash harness/scripts/codex-companion.sh review --base HEAD --json",
  "tests": [],
  "accepted_findings": [],
  "rejected_findings": [],
  "clean_result": true,
  "codex_available": true,
  "fallback": ""
}
```
