# Version Drift Detection

## What to Check

VERSION and the version field in .claude-plugin/plugin.json must always match.
When a mismatch is detected, suggest running `./harness/skills/harness-release/scripts/sync-version.sh` (do not run it automatically).

## Feature Table Freshness

Items marked "planned (not yet implemented)" or "scheduled for implementation" in
docs/CLAUDE-feature-table.md should be proposed for deletion after 6 months have passed.

## Why This Rule Is Needed

D2 (inaccurate information) recurs even after being corrected once.
Version mismatches and Feature Table decay are the most common drift patterns.
