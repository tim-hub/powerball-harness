# Plans Maintenance

Last updated: 2026-03-06

`Plans.md` is the authoritative source, but letting it grow indefinitely makes it prone to drift between "past completion statements" and "current repo state."
This document defines minimal operational rules to reduce drift.

## Lightweight Rule

1. Before starting a new major improvement phase, treat only the most recent 1-2 phases as the active zone
2. Older completed phases should be archived to a history location such as `docs/plans-history/` if needed
3. Wording prone to conflicting with the current tree (such as "deleted" or "migration complete") should have correction notes added when state changes in subsequent phases
4. When changing the handling of README / docs / `.gitignore` / build scripts, also fix the corresponding language in `Plans.md` in the same commit

## When to Archive

Consider archiving old completed phases when any of the following are met:

- The primary work target in `Plans.md` requires looking back 3+ phases
- Terms like "deleted" or "consolidated" create misunderstandings with the current repo
- The cost of reading past history on each sync-status run becomes noticeable

## Recommended Shape

- `Plans.md`: Only the current active phase and most recently completed phases
- `docs/plans-history/`: Fixed snapshots of past phases
- `docs/distribution-scope.md`: Current truth about residual artifacts and distribution boundaries

## Phase 21 Decision

- This time, archiving was not performed; instead, the priority was **correcting misleading completion statements**
- Before the next major phase addition, it is recommended to archive Phase 17 and earlier completion history to `docs/plans-history/`
