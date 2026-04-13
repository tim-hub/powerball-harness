---
name: harness-release
description: "Use this skill whenever the user mentions releasing, version bumps, creating tags, publishing, cutting a release, or runs /harness-release. Also use when the user asks about the release process or wants to finalize and ship changes. Do NOT load for: code implementation (use harness-work), code review (use harness-review), planning (use harness-plan), or project setup. Unified release skill for Harness v3 — automates CHANGELOG updates, version bumps, git tags, GitHub Releases, mirror sync, and release validation."
allowed-tools: ["Read", "Write", "Edit", "Bash"]
argument-hint: "[patch|minor|major|--dry-run|--announce|--complete]"
context: fork
effort: high
---

# Harness Release (v3)

Unified release skill for Harness v3.
Consolidates the following legacy skills:

- `release-har` -- General-purpose release automation
- `x-release-harness` -- Harness-specific release automation
- `handoff` -- PM handoff and completion reporting

## Quick Reference

```bash
/release          # Interactive (confirms version type)
/release patch    # Patch version bump (bug fixes)
/release minor    # Minor version bump (new features)
/release major    # Major version bump (breaking changes)
/release --dry-run   # Preview only (no execution)
/release --announce  # Also post X (Twitter) announcement
/release --complete  # Release completion marking (finishing after tagging)
```

## Release-only policy

- Normal PRs: Do not touch `VERSION` / `.claude-plugin/plugin.json` / versioned `CHANGELOG.md` entries
- Change history for normal PRs: Append to the `[Unreleased]` section of `CHANGELOG.md`
- Only when running `/release`: update version bump, versioned CHANGELOG entry, and tag / GitHub Release together
- `/release --dry-run` runs the same preflight as production execution, catching red flags before publishing

## Branch Policy

- **Solo development**: Direct push to main is allowed (CI serves as the quality gate)
- **Collaborative development**: Merge via PR is required
- Force push (`--force` / `--force-with-lease`) is always prohibited

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

## NPM Distribution

This project is a Claude Code plugin and is not distributed as an npm package.
There is no `package.json` at the root (`core/package.json` is for the internal TypeScript build).
Only the following 2 files are subject to version management:

- `VERSION` -- Source of truth
- `.claude-plugin/plugin.json` -- Plugin manifest

## Distribution Surfaces and Mirror Sync

`skills/` is the SSOT (Single Source of Truth). Codex CLI uses symlinks from `codex/.codex/skills/` → `../../../skills/`, so no manual sync is needed.

| Surface | Path | Target Users |
|---------|------|-------------|
| Claude | `skills/harness-release/` | Claude Code users |
| Codex | `codex/.codex/skills/harness-release/` (symlink) | Codex CLI users |

## Internationalization (i18n)

The skill's description field can be switched between Japanese and English. Verify the locale setting is as intended before releasing:

```bash
# Set to Japanese (description-ja → description)
./scripts/i18n/set-locale.sh ja

# Set to English (description-en → description)
./scripts/i18n/set-locale.sh en
```

Current default: description is in Japanese (identical to `description-ja`). `description-en` is always maintained as an English backup.

## Execution Flow

### Phase 0: Pre-flight Checks (Required)

```bash
# 1. Verify required tools
command -v gh &>/dev/null || echo "gh missing: GitHub Release will be skipped"
command -v jq &>/dev/null || echo "jq missing: required for plugin.json update"

# 2. vendor-neutral preflight (common to production and dry-run)
bash scripts/release-preflight.sh

# 3. Plugin structure validation
bash tests/validate-plugin.sh

# 4. Consistency check
bash scripts/ci/check-consistency.sh

# 5. Verify codex symlinks
ls -la codex/.codex/skills/
```

`scripts/release-preflight.sh` validates the following:

- Whether the working tree is clean
- Whether `CHANGELOG.md` has an `[Unreleased]` section
- Diff between `.env.example` and `.env` (warning only under managed secrets)
- `healthcheck` / `preflight` commands (run if available)
- Whether `agents/` / `core/` / `hooks/` / `scripts/` shipped surfaces contain debug / mock / placeholder remnants
- CI status (when available)

Adjustable per repository via environment variables:

- `HARNESS_RELEASE_PROJECT_ROOT`
- `HARNESS_RELEASE_HEALTHCHECK_CMD`
- `HARNESS_RELEASE_CI_STATUS_CMD`

Details: [docs/release-preflight.md](${CLAUDE_SKILL_DIR}/../../docs/release-preflight.md)

### Phase 1: Get Current Version

```bash
CURRENT=$(cat VERSION 2>/dev/null)
echo "Current version: $CURRENT"
```

### Phase 2: Calculate New Version

`scripts/sync-version.sh` only supports patch bumps. For minor / major, manually edit VERSION:

```bash
# patch bump (x.y.Z → x.y.(Z+1))
./scripts/sync-version.sh bump

# minor bump (manual: x.Y.z → x.(Y+1).0)
CURRENT=$(cat VERSION)
MAJOR=$(echo "$CURRENT" | cut -d. -f1)
MINOR=$(echo "$CURRENT" | cut -d. -f2)
NEW_VERSION="$MAJOR.$((MINOR + 1)).0"
echo "$NEW_VERSION" > VERSION
./scripts/sync-version.sh sync

# major bump (manual: X.y.z → (X+1).0.0)
CURRENT=$(cat VERSION)
MAJOR=$(echo "$CURRENT" | cut -d. -f1)
NEW_VERSION="$((MAJOR + 1)).0.0"
echo "$NEW_VERSION" > VERSION
./scripts/sync-version.sh sync
```

`sync-version.sh sync` applies the `VERSION` value to `.claude-plugin/plugin.json`.

### Phase 3: CHANGELOG Update

The release entry finalizes changes accumulated in `[Unreleased]` from normal PRs into a versioned section.

Write using the **detailed Before/After format** (in Japanese).
Split each feature into numbered sections, explaining "Before" and "After" with concrete examples.

```markdown
## [X.Y.Z] - YYYY-MM-DD

### Theme: [One-line summary of all changes]

**[Value to users in 1-2 sentences]**

---

#### 1. [Feature Name]

**Before**: [Describe the old behavior concretely. Paint the pain point users experienced]

**After**: [Describe the new behavior concretely. What gets resolved]

```
[Actual output or command examples]
```

#### 2. [Next Feature Name]

**Before**: ...

**After**: ...
```

**CC version integration pattern**: Instead of the usual "Before / After" format, use the "CC update → Harness utilization" format.
See the "CC Version Integration CHANGELOG Pattern" section in `.claude/rules/github-release.md` for details.

**Writing rules**:

| Rule | Description |
|------|-------------|
| Language | **Japanese** |
| Each feature as a separate section | Numbered with `#### N. Feature Name` |
| "Before" describes the pain point | Concretely describe the inconvenience users experienced |
| "After" shows the resolution | What changes and how + concrete examples (code/output) |
| Always include concrete examples | Command examples, output examples, Plans.md snippets, etc. |
| Minimize technical details | File names and step numbers as supplements in "After" |
| Longer is OK | 3-10 lines per feature. Readability is the top priority |

Do not empty the `[Unreleased]` section; keep it for the next release:

```markdown
## [Unreleased]

## [X.Y.Z] - YYYY-MM-DD
...
```

### Phase 4: Update Version Files

```bash
# VERSION was already updated in Phase 2
# Sync plugin.json
./scripts/sync-version.sh sync

# Verify sync
./scripts/sync-version.sh check
```

### Phase 5: Verify Codex Symlinks

```bash
# Verify codex symlinks resolve correctly
ls -la codex/.codex/skills/
```

### Phase 6: Commit & Tag

```bash
NEW_VERSION=$(cat VERSION)

# Staging (explicitly specify target files)
git add VERSION .claude-plugin/plugin.json CHANGELOG.md
git add skills/ codex/.codex/skills/

git commit -m "chore: release v$NEW_VERSION"
git tag -a "v$NEW_VERSION" -m "Release v$NEW_VERSION"
```

### Phase 7: Push

```bash
git push origin main --tags
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

---

Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

GitHub Release Notes rules:
- Language: **English** (for public repository)
- Required: `## What's Changed`, bold summary, Before / After table, footer
- Detailed format: See `.claude/rules/github-release.md`

Release notes validation:

```bash
./scripts/validate-release-notes.sh "v$NEW_VERSION"
```

### Phase 9: Release Completion Marking

```bash
git commit --allow-empty -m "chore: mark v$NEW_VERSION release complete"
git push origin main
```

This empty commit serves as an explicit marker that "all release work is complete."

### Phase 10: Announcement (`--announce` only)

Invokes the `/x-announce` skill to generate an announcement thread for X (Twitter):

```
Skill: x-announce
Args: v$NEW_VERSION
```

Outputs 5 post texts + 5 Gemini images in a single pass.

## `--dry-run` Mode

`--dry-run` executes the following without making actual changes:

1. Run **all** pre-flight checks (Phase 0)
2. Display calculated version (without writing)
3. Display CHANGELOG draft (without writing)
4. Display GitHub Release Notes draft (without creating)
5. Display mirror sync diff (without writing)

Skipped items: VERSION/plugin.json writes, git commit/tag/push, GitHub Release creation, announcements

## `--complete` Mode

Performs only the "release completion" marking after tag creation:

```bash
/release --complete
```

Executes only Phase 9. Verifies that the GitHub Release was not missed, then creates the completion commit.

## Regression Checklist

Verify the following regressions before release:

| Check Item | Verification Method | Notes |
|------------|-------------------|-------|
| Plugin structure | `tests/validate-plugin.sh` | Validates plugin.json, skills, hooks, scripts |
| Consistency | `scripts/ci/check-consistency.sh` | Templates, versions, mirrors, CHANGELOG |
| Codex symlinks | `ls -la codex/.codex/skills/` | All symlinks resolve to skills/ |
| Preflight | `scripts/release-preflight.sh` | Working tree, CHANGELOG, CI, remnants |
| Release notes | `scripts/validate-release-notes.sh vX.Y.Z` | GitHub Release format validation |
| VERSION sync | `scripts/sync-version.sh check` | Match between VERSION and plugin.json |
| Guardrails | R01-R13 in `core/src/guardrails/rules.ts` | TypeScript rule health |
| Tag continuity | `git tag --sort=-version:refname \| head -5` | No missing tags |
| Locale | Match between description and description-ja | Switchable via `set-locale.sh` |

## CI Safety Net

`.github/workflows/release.yml` runs automatically on tag push:

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
- Mixing implementation changes other than VERSION / plugin.json / CHANGELOG into release commits

## Related Skills

- `harness-review` -- Perform code review before release
- `harness-work` -- Implement next tasks after release
- `harness-plan` -- Create plans for the next version
- `x-announce` -- Generate X (Twitter) release announcement threads
- `harness-setup` -- Mirror sync and plugin configuration setup

## Related Rules

- `.claude/rules/versioning.md` -- SemVer determination criteria and batch release recommendations
- `.claude/rules/github-release.md` -- GitHub Release Notes format (English)
- `.claude/rules/cc-update-policy.md` -- Feature Table quality criteria for CC update integration
