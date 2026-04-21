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

## Version Determination Criteria (SemVer)

Decision flowchart:

```
Does existing behavior break?
├─ Yes → major
└─ No → Can the user do something new?
    ├─ Yes → minor
    └─ No → patch
```

| Type of Change | Version | Example |
|----------------|---------|---------|
| Wording fixes or documentation updates | **patch** | Minor template adjustments |
| hooks/scripts bug fixes | **patch** | Escape fixes |
| New features/flags/commands added | **minor** | `--dual`, new skill |
| Breaking changes (legacy removal, format incompatibility) | **major** | Plans.md v1 removal |

**Batch releases recommended**: When multiple changes occur on the same day, combine them into one minor release. Two or more minor bumps on the same day is prohibited.

## Version Distribution

The canonical version lives in `VERSION`. Your project may sync it to additional manifests (package.json, Cargo.toml, pyproject.toml, etc.) — see Phase 4.

## Execution Flow

### Phase 0: Pre-flight Checks (Required)

```bash
# 1. Verify required tools
command -v gh &>/dev/null || echo "gh missing: GitHub Release will be skipped"
command -v jq &>/dev/null || echo "jq missing: required for manifest updates"

# 2. vendor-neutral preflight (common to production and dry-run)
bash "${CLAUDE_SKILL_DIR}/scripts/release-preflight.sh"
```

`${CLAUDE_SKILL_DIR}/scripts/release-preflight.sh` validates the following:

- Whether the working tree is clean
- Whether `CHANGELOG.md` has an `[Unreleased]` section
- Diff between `.env.example` and `.env` (warning only under managed secrets)
- `healthcheck` / `preflight` commands (run if available)
- Whether `agents/` / `hooks/` / `scripts/` shipped surfaces contain debug / mock / placeholder remnants
- CI status (when available)

Adjustable per repository via environment variables:

- `HARNESS_RELEASE_PLUGIN_ROOT`
- `HARNESS_RELEASE_HEALTHCHECK_CMD`
- `HARNESS_RELEASE_CI_STATUS_CMD`

Details: `docs/release-preflight.md` (project-root)

### Phase 1: Get Current Version

```bash
CURRENT=$(cat VERSION 2>/dev/null)
echo "Current version: $CURRENT"
```

### Phase 2: Calculate New Version

`${CLAUDE_SKILL_DIR}/scripts/sync-version.sh` only supports patch bumps. For minor / major, manually edit VERSION:

```bash
# patch bump (x.y.Z → x.y.(Z+1))
"${CLAUDE_SKILL_DIR}/scripts/sync-version.sh" bump

# minor bump (manual: x.Y.z → x.(Y+1).0)
CURRENT=$(cat VERSION)
MAJOR=$(echo "$CURRENT" | cut -d. -f1)
MINOR=$(echo "$CURRENT" | cut -d. -f2)
NEW_VERSION="$MAJOR.$((MINOR + 1)).0"
echo "$NEW_VERSION" > VERSION

# major bump (manual: X.y.z → (X+1).0.0)
CURRENT=$(cat VERSION)
MAJOR=$(echo "$CURRENT" | cut -d. -f1)
NEW_VERSION="$((MAJOR + 1)).0.0"
echo "$NEW_VERSION" > VERSION
```

Sync VERSION to your project's manifest files (if any) — see `sync-version.sh sync` documentation.

### Phase 3: CHANGELOG Update

Finalize the `[Unreleased]` section into a versioned entry. Use the `writing-changelog` skill for format rules, the Before/After template, and the CC version integration pattern.

Key requirements:
- Move `[Unreleased]` content into a new `## [X.Y.Z] - YYYY-MM-DD` section
- Keep an empty `## [Unreleased]` placeholder above it for the next release
- Each feature gets its own `#### N. Feature Name` section with **Before** / **After**

### Phase 4: Commit & Tag

```bash
NEW_VERSION=$(cat VERSION)

# Stage release files — add your manifest file here if you sync VERSION to one
git add VERSION CHANGELOG.md
# If you sync VERSION to a manifest file, add it here too:
# git add VERSION CHANGELOG.md your-manifest-file

git commit -m "chore: release v$NEW_VERSION"
git tag -a "v$NEW_VERSION" -m "Release v$NEW_VERSION"
```

### Phase 5: Push

```bash
git push origin {main or master} --tags
```

**Note**: `.github/workflows/release.yml` detects tag pushes and runs a safety net that auto-generates a GitHub Release from CHANGELOG if one hasn't been created yet. If you create the GitHub Release manually first, the workflow will automatically skip.

### Phase 6: Create GitHub Release

```bash
NEW_VERSION=$(cat VERSION)

gh release create "v$NEW_VERSION" \
  --title "v$NEW_VERSION - Title" \
  --notes "$(cat <<'EOF'
## What's Changed

**[Summary of changes (English)]**

### Before / After

| Before | After |
|--------|-------|
| Previous state | New state |

---

## Added

- **Feature**: Description

## Changed

- **Change**: Description

## Fixed

- **Fix**: Description

EOF
)"
```

GitHub Release Notes rules:
- Language: **English** (for public repository)
- Required: `## What's Changed`, bold summary, Before / After table, footer

Release notes validation:

```bash
"${CLAUDE_SKILL_DIR}/scripts/validate-release-notes.sh" "v$NEW_VERSION"
```

## `--dry-run` Mode

`--dry-run` executes the following without making actual changes:

1. Run **all** pre-flight checks (Phase 0)
2. Display calculated version (without writing)
3. Display CHANGELOG draft (without writing)
4. Display GitHub Release Notes draft (without creating)

Skipped items: VERSION writes, git commit/tag/push, GitHub Release creation

---

# Additional Guidelines

## Regression Checklist

Verify the following before release:

- [ ] **Preflight** — `bash "${CLAUDE_SKILL_DIR}/scripts/release-preflight.sh"` (working tree, CHANGELOG, CI, remnants)
- [ ] **Release notes** — `bash "${CLAUDE_SKILL_DIR}/scripts/validate-release-notes.sh" vX.Y.Z` (GitHub Release format)
- [ ] **VERSION sync** — `bash "${CLAUDE_SKILL_DIR}/scripts/sync-version.sh" check` (VERSION matches any manifest files)
- [ ] **Tag continuity** — `git tag --sort=-version:refname | head -5` (no missing tags)
- [ ] **CI status** — verify CI passes on the release commit
- [ ] **GitHub Release** — confirm `gh release view vX.Y.Z` shows the expected notes
- [ ] **Working tree** — `git status` is clean after push

## CI Safety Net

1. Detects `v*` tag push
2. Checks if a GitHub Release with the same name already exists
3. If not, auto-generates from CHANGELOG (safety net)
4. If it exists, does nothing

The recommended flow is to create the GitHub Release manually before pushing.
The safety net only rescues "forgotten Release creation."

## PM Handoff

Completion report to PM after release:

```markdown
## Release Completion Report

**Version**: v{{NEW_VERSION}}
**Release Date**: {{DATE}}

### Completed Work
{{CHANGELOG contents}}

### GitHub Release
{{URL}}

### Next Actions
- PM review of release notes
- Production deployment (if applicable)
```

## Prohibited Actions

- Deleting or rolling back tags (published versions are immutable)
- Two or more minor bumps on the same day
- Minor bump for patch-level changes
- Force push via `--force` / `--force-with-lease`
- Mixing implementation changes other than VERSION / CHANGELOG (and any synced manifests) into release commits

## Related

- Your project's build/test/review skills (project-specific pre-release orchestration)
- `writing-changelog` skill — CHANGELOG format and Before/After template
