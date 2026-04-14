# Team Mode and Issue Bridge

`Plans.md` remains the authoritative source, and GitHub Issue integration is only used via opt-in team mode.

## When to Use

- Solo development does not use the issue bridge
- Team mode creates a single tracking issue, with sub-issue payloads generated in dry-run mode for each task underneath
- The issue bridge does not update Plans.md
- Completes in dry-run only; no actual updates are made to GitHub

## Conversion Rules

`harness/scripts/plans-issue-bridge.sh` expands each task in Plans.md into the following:

- tracking issue
  - Parent issue for aggregation
  - Contains a list of phases and tasks in the body
- sub-issue
  - Individual payload for each task
  - Retains `task id`, `DoD`, `Depends`, `Status` in the body

## Example

```bash
harness/scripts/plans-issue-bridge.sh --team-mode --plans Plans.md
```

Specifying `--format markdown` switches to a human-readable dry-run.

## Benefits

- Plans.md can remain as the authoritative source
- Only team work gets issue-based visibility
- Solo development does not incur extra overhead
