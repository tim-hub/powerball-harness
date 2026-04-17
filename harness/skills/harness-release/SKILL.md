---
name: harness-release
description: "Manages version bumps, CHANGELOG, git tags, and GitHub Releases. Use when cutting a release or updating version metadata."
when_to_use: "release, bump version, update CHANGELOG, create tag, GitHub release, patch release, minor release"
allowed-tools: ["Read", "Write", "Edit", "Bash"]
argument-hint: "[patch|minor|major|--dry-run|--complete]"
context: fork
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
| `harness-release --complete` | `--complete` | Release completion marking only (Phase 9) |

Key paths:
- Preflight script: `${CLAUDE_SKILL_DIR}/scripts/release-preflight.sh`
- Override plugin root: `HARNESS_RELEASE_PLUGIN_ROOT=/path/to/repo`

## Release-only policy

- Normal PRs: Do not touch `VERSION` / `.claude-plugin/marketplace.json` / versioned `CHANGELOG.md` entries
- Change history for normal PRs: Append to the `[Unreleased]` section of `CHANGELOG.md`
- Only when running `/release`: update version bump, versioned CHANGELOG entry, and tag / GitHub Release together
- `/release --dry-run` runs the same preflight as production execution, catching red flags before publishing

## Branch Policy

- **Solo development**: Direct push to main or master is allowed (CI serves as the quality gate)
- **Collaborative development**: Merge via PR is required

## Version Determination Criteria (SemVer)

Decision flowchart based on `.claude/rules/versioning.md`:

```
Does existing behavior break?
├─ Yes → major
└─ No → Can the user do something new?
    ├─ Yes → minor
    └─ No → patch
```

| Type of Change | Version | Example |
|----------------|---------|---------|
| Skill definition wording fixes/additions | **patch** | Minor template adjustments |
| hooks/scripts bug fixes | **patch** | Escape fixes |
| New skills/flags/agents added | **minor** | `--dual`, new skill |
| CC new version compatibility | **minor** | CC v2.1.90 support |
| Breaking changes (legacy skill removal, format incompatibility) | **major** | Plans.md v1 removal |

**Batch releases recommended**: When multiple changes occur on the same day, combine them into one minor release. Two or more minor bumps on the same day is prohibited.

## Version Distribution

Only the following 2 files are subject to version management:

- `VERSION` -- Source of truth
- `.claude-plugin/marketplace.json` -- Plugin manifest

## Distribution Surfaces and Mirror Sync

`skills/` is the SSOT (Single Source of Truth). Codex CLI uses symlinks from `codex/.codex/skills/` → `../../../skills/`, so no manual sync is needed.

| Surface | Path | Target Users |
|---------|------|-------------|
| Claude | `skills/harness-release/` | Claude Code users |
| Codex | `codex/.codex/skills/harness-release/` (symlink) | Codex CLI users |

## Execution Flow

### Phase 0: Pre-flight Checks (Required)

```bash
# 1. Verify required tools
command -v gh &>/dev/null || echo "gh missing: GitHub Release will be skipped"
command -v jq &>/dev/null || echo "jq missing: required for marketplace.json update"

# 2. vendor-neutral preflight (common to production and dry-run)
bash "${CLAUDE_SKILL_DIR}/scripts/release-preflight.sh"

# 3. Plugin structure validation
bash tests/validate-plugin.sh

# 4. Consistency check
bash "${CLAUDE_SKILL_DIR}/scripts/check-consistency.sh"

# 5. Verify codex symlinks
ls -la codex/.codex/skills/
```

`${CLAUDE_SKILL_DIR}/scripts/release-preflight.sh` validates the following:

- Whether the working tree is clean
- Whether `CHANGELOG.md` has an `[Unreleased]` section
- Diff between `.env.example` and `.env` (warning only under managed secrets)
- `healthcheck` / `preflight` commands (run if available)
- Whether `agents/` / `core/` / `hooks/` / `scripts/` shipped surfaces contain debug / mock / placeholder remnants
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
"${CLAUDE_SKILL_DIR}/scripts/sync-version.sh" sync

# major bump (manual: X.y.z → (X+1).0.0)
CURRENT=$(cat VERSION)
MAJOR=$(echo "$CURRENT" | cut -d. -f1)
NEW_VERSION="$((MAJOR + 1)).0.0"
echo "$NEW_VERSION" > VERSION
"${CLAUDE_SKILL_DIR}/scripts/sync-version.sh" sync
```

`${CLAUDE_SKILL_DIR}/scripts/sync-version.sh sync` applies the `VERSION` value to `.claude-plugin/marketplace.json`.

### Phase 3: CHANGELOG Update

Finalize the `[Unreleased]` section into a versioned entry. Use the `writing-changelog` skill for format rules, the Before/After template, and the CC version integration pattern.

Key requirements:
- Move `[Unreleased]` content into a new `## [X.Y.Z] - YYYY-MM-DD` section
- Keep an empty `## [Unreleased]` placeholder above it for the next release
- Each feature gets its own `#### N. Feature Name` section with **Before** / **After**

### Phase 4: Update Version Files

```bash
# VERSION was already updated in Phase 2
# Sync marketplace.json
"${CLAUDE_SKILL_DIR}/scripts/sync-version.sh" sync

# Verify sync
"${CLAUDE_SKILL_DIR}/scripts/sync-version.sh" check
```

### Phase 5: Verify Codex Symlinks

```bash
# Only run when Codex CLI is installed
if command -v codex &>/dev/null; then
  ls -la codex/.codex/skills/
else
  echo "  (Codex CLI not installed — symlink check skipped)"
fi
```

### Phase 6: Commit & Tag

```bash
NEW_VERSION=$(cat VERSION)

# Stage only the files a release commit should touch
git add harness/VERSION harness/harness.toml CHANGELOG.md

git commit -m "chore: release v$NEW_VERSION"
git tag -a "v$NEW_VERSION" -m "Release v$NEW_VERSION"
```

### Phase 7: Push

```bash
git push origin {main or master} --tags
```

**Note**: `.github/workflows/release.yml` detects tag pushes and runs a safety net that auto-generates a GitHub Release from CHANGELOG if one hasn't been created yet. If you create the GitHub Release manually first, the workflow will automatically skip.

### Phase 8: Create GitHub Release

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
- Detailed format: See `.claude/rules/github-release.md`

Release notes validation (plugin-level script):

```bash
"${CLAUDE_SKILL_DIR}/scripts/validate-release-notes.sh" "v$NEW_VERSION"
```

### Phase 9: Release Completion Marking

```bash
git commit --allow-empty -m "chore: mark v$NEW_VERSION release complete"
git push origin {main or master}
```

This empty commit serves as an explicit marker that "all release work is complete."

## `--dry-run` Mode

`--dry-run` executes the following without making actual changes:

1. Run **all** pre-flight checks (Phase 0)
2. Display calculated version (without writing)
3. Display CHANGELOG draft (without writing)
4. Display GitHub Release Notes draft (without creating)
5. Display mirror sync diff (without writing)

Skipped items: VERSION/marketplace.json writes, git commit/tag/push, GitHub Release creation

## `--complete` Mode

Performs only the "release completion" marking after tag creation:

```bash
/release --complete
```

Executes only Phase 9. Verifies that the GitHub Release was not missed, then creates the completion commit.

---

# Additional Guidelines

## Regression Checklist

Verify the following regressions before release:

- [ ] **Plugin structure** — `bash tests/validate-plugin.sh` (marketplace.json, skills, hooks, scripts)
- [ ] **Consistency** — `bash "${CLAUDE_SKILL_DIR}/scripts/check-consistency.sh"` (templates, versions, CHANGELOG)
- [ ] **Templates** — `test -f ${CLAUDE_SKILL_DIR}/../../templates/codex/config.toml && test -f ${CLAUDE_SKILL_DIR}/../../templates/opencode/opencode.json` (setup templates present)
- [ ] **Preflight** — `bash "${CLAUDE_SKILL_DIR}/scripts/release-preflight.sh"` (working tree, CHANGELOG, CI, remnants)
- [ ] **Release notes** — `bash "${CLAUDE_SKILL_DIR}/scripts/validate-release-notes.sh" vX.Y.Z` (GitHub Release format)
- [ ] **VERSION sync** — `bash "${CLAUDE_SKILL_DIR}/scripts/sync-version.sh" check` (VERSION matches marketplace.json)
- [ ] **Guardrails** — R01-R13 in `go/internal/guardrail/rules.go` (Go rule health)
- [ ] **Tag continuity** — `git tag --sort=-version:refname | head -5` (no missing tags)
- [ ] **Migration residue** — `bash "${CLAUDE_SKILL_DIR}/scripts/check-residue.sh"` (no deleted-concept references)

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
- Mixing implementation changes other than VERSION / marketplace.json / CHANGELOG into release commits

## Related Skills

- `harness-review` -- Perform code review before release
- `harness-work` -- Implement next tasks after release
- `harness-plan` -- Create plans for the next version
- `harness-setup` -- Mirror sync and plugin configuration setup

## Related Rules

- `.claude/rules/versioning.md` -- SemVer determination criteria and batch release recommendations
- `.claude/rules/github-release.md` -- GitHub Release Notes format (English)
- `.claude/rules/cc-update-policy.md` -- Feature Table quality criteria for CC update integration
