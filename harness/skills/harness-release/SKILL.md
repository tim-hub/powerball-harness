---
name: harness-release
description: "Manages version bumps, CHANGELOG, git tags, and GitHub Releases. Use when cutting a release or updating version metadata."
when_to_use: "release, bump version, update CHANGELOG, create tag, GitHub release, patch release, minor release"
allowed-tools: ["Read", "Write", "Edit", "Bash"]
argument-hint: "[patch|minor|major|--dry-run]"
effort: medium
model: sonnet
---

# Harness Release

## Quick Reference

| User Input | Subcommand | Behavior |
|------------|------------|----------|
| `harness-release` |  | Patch as default version to bump |
| `harness-release patch` | `patch` | Patch version bump (bug fixes, x.y.Z+1) |
| `harness-release minor` | `minor` | Minor version bump (new features, x.Y+1.0) |
| `harness-release major` | `major` | Major version bump (breaking changes, X+1.0.0) |
| `harness-release --dry-run` | `--dry-run` | Preview all phases without writing or publishing |

Key paths:
- Preflight script: `${CLAUDE_SKILL_DIR}/scripts/release-preflight.sh`
- Override plugin root: `HARNESS_RELEASE_PLUGIN_ROOT=/path/to/repo`

**Version classification and batch-release rules**: [`${CLAUDE_SKILL_DIR}/references/versioning-rules.md`](${CLAUDE_SKILL_DIR}/references/versioning-rules.md)

## Project-specific pre-release steps

Before running harness-release, run your project's own pre-release checks:

```bash
# Example: if you have a release-this skill
/release-this patch    # runs build, lint, project checks → then delegates to harness-release

# Or manually:
make build-all          # build binaries / assets
make test               # run full test suite
make validate           # project-specific validation
# then:
/harness-release patch  # generic release flow
```

harness-release is the generic release engine. Projects should create their own pre-release
orchestrator (e.g. `.claude/skills/release-this/`) that runs project-specific checks first.

## Release-only policy

- Normal PRs: Do not touch `VERSION` or versioned `CHANGELOG.md` entries
- Change history for normal PRs: Append to the `[Unreleased]` section of `CHANGELOG.md`
- Only when running `/release`: update version bump, versioned CHANGELOG entry, and tag / GitHub Release together
- `/release --dry-run` runs the same preflight as production execution, catching red flags before publishing

## Branch Policy

- **Solo development**: Direct push to main or master is allowed (CI serves as the quality gate)
- **Collaborative development**: Merge via PR is required

## Execution Flow

### Phase 0: Pre-flight Checks (Required)

```bash
command -v gh &>/dev/null || echo "gh missing: GitHub Release will be skipped"
command -v jq &>/dev/null || echo "jq missing: required for manifest updates"
bash "${CLAUDE_SKILL_DIR}/scripts/release-preflight.sh"
```

Validates: clean working tree, `[Unreleased]` section exists, CI status, no debug/mock remnants.
Adjustable via `HARNESS_RELEASE_PLUGIN_ROOT`, `HARNESS_RELEASE_HEALTHCHECK_CMD`, `HARNESS_RELEASE_CI_STATUS_CMD`.

### Phase 1: Get Current Version

```bash
CURRENT=$(cat VERSION 2>/dev/null)
```

### Phase 2: Calculate New Version

```bash
# patch bump
"${CLAUDE_SKILL_DIR}/scripts/sync-version.sh" bump

# minor bump (manual)
MAJOR=$(cat VERSION | cut -d. -f1); MINOR=$(cat VERSION | cut -d. -f2)
echo "$MAJOR.$((MINOR + 1)).0" > VERSION

# major bump (manual)
MAJOR=$(cat VERSION | cut -d. -f1)
echo "$((MAJOR + 1)).0.0" > VERSION
```

### Phase 3: CHANGELOG Update

Move `[Unreleased]` content into a new `## [X.Y.Z] - YYYY-MM-DD` section.
Keep an empty `## [Unreleased]` placeholder above it.
Use the `writing-changelog` skill for format rules and the Before/After template.

### Phase 4: Commit & Tag

```bash
NEW_VERSION=$(cat VERSION)
git add VERSION CHANGELOG.md
git commit -m "chore: release v$NEW_VERSION"
git tag -a "v$NEW_VERSION" -m "Release v$NEW_VERSION"
```

### Phase 5: Push

```bash
git push origin {main or master} --tags
```

`.github/workflows/release.yml` auto-generates a GitHub Release from CHANGELOG on tag push if one hasn't been created yet.

### Phase 6: Create GitHub Release

```bash
NEW_VERSION=$(cat VERSION)
gh release create "v$NEW_VERSION" \
  --title "v$NEW_VERSION - Title" \
  --notes "$(cat <<'EOF'
## What's Changed

**[Summary]**

### Before / After

| Before | After |
|--------|-------|
| Previous state | New state |

EOF
)"
```

Release notes must be in English. Run `"${CLAUDE_SKILL_DIR}/scripts/validate-release-notes.sh" "v$NEW_VERSION"` to validate.

## `--dry-run` Mode

Runs all pre-flight checks (Phase 0), displays calculated version, CHANGELOG draft, and GitHub Release Notes draft — no writes.

## Regression Checklist

- [ ] `bash "${CLAUDE_SKILL_DIR}/scripts/release-preflight.sh"` — clean
- [ ] `bash "${CLAUDE_SKILL_DIR}/scripts/validate-release-notes.sh" vX.Y.Z` — format OK
- [ ] `bash "${CLAUDE_SKILL_DIR}/scripts/sync-version.sh" check` — VERSION in sync
- [ ] `git tag --sort=-version:refname | head -5` — no gaps
- [ ] CI passes on release commit
- [ ] `git status` clean after push

## Related

- `writing-changelog` skill — CHANGELOG format and Before/After template
- [`${CLAUDE_SKILL_DIR}/references/versioning-rules.md`](${CLAUDE_SKILL_DIR}/references/versioning-rules.md) — SemVer classification and batch-release rules
