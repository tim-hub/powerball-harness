# Phase 21 Release Checklist

Last updated: 2026-03-06

This checklist is a verification table for release decisions involving `trust repair`, `evidence pack`, and `positioning refresh` changes.

## Surfaces

- [ ] `VERSION` and `.claude-plugin/plugin.json` are in sync
- [ ] README / README_ja use the latest release badge
- [ ] README / README_ja have no broken links
- [ ] Descriptions in `docs/distribution-scope.md` and `Plans.md` are consistent
- [ ] Classification in `docs/claims-audit.md` does not contradict current wording

## Evidence

- [ ] `./tests/validate-plugin.sh`
- [ ] `./tests/validate-plugin-v3.sh`
- [ ] `./.claude/scripts/check-consistency.sh`
- [ ] `cd core && npm test`
- [ ] `./scripts/evidence/run-work-all-smoke.sh`
- [ ] `./scripts/evidence/run-work-all-success.sh --full` if needed
- [ ] `./scripts/evidence/run-work-all-success.sh --full --strict-live` if you want to demonstrate live Claude completion
- [ ] `./scripts/evidence/run-work-all-failure.sh --full` if needed

## Artifact Review

- [ ] The description in `docs/evidence/work-all.md` matches the generated artifacts
- [ ] The most recent artifacts in `out/evidence/work-all/` have been reviewed
- [ ] The release note specifies which of success / failure is unverified

## Release Decision

- [ ] Determined whether this change requires release metadata updates
- [ ] Obtained explicit approval for GitHub Release / tag creation
- [ ] Organized the announcement copy without mixing `trust repair`, `evidence pack`, and `positioning refresh`

## Current Recommendation (2026-03-06)

- If only shipping evidence tooling with replay fallback, release is possible.
- However, if you want to strongly announce "live Claude completed the happy path as-is," wait until the `--strict-live` artifact is obtained.
