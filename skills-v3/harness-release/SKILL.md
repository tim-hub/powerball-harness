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
/release          # Interactive (confirm version type)
/release patch    # Patch version bump (bug fixes)
/release minor    # Minor version bump (new features)
/release major    # Major version bump (breaking changes)
/release --dry-run   # Preview only (no actual execution)
/release --announce  # Also execute X (Twitter) announcement
/release --complete  # Release completion marking (post-tag finishing)
```

## Release-only policy

- Normal PRs: Do not touch `VERSION` / `.claude-plugin/plugin.json` / versioned `CHANGELOG.md` entries
- Change history for normal PRs: Append to `[Unreleased]` in `CHANGELOG.md`
- Only during `/release` execution: Update version bump, versioned CHANGELOG entry, tag / GitHub Release together
- `/release --dry-run` runs the same preflight as production execution, catching danger signals before publishing

## Branch Policy

- **Solo development**: Direct push to main is allowed (CI serves as quality gate)
- **Collaborative development**: Merge via PR is required
- Force push (`--force` / `--force-with-lease`) is always prohibited

## Version Determination Criteria (SemVer)

Decision flowchart based on `.claude/rules/versioning.md`:

```
Does existing behavior break?
├─ Yes -> major
└─ No -> Can the user do something new?
    ├─ Yes -> minor
    └─ No -> patch
```

| Change Type | Version | Example |
|-----------|----------|-----|
| Skill definition wording fixes/additions | **patch** | Template minor fix |
| hooks/scripts bug fixes | **patch** | Escape fix |
| New skill/flag/agent additions | **minor** | `--dual`, new skill |
| CC new version compatibility | **minor** | CC v2.1.90 support |
| Breaking changes (legacy skill removal, format incompatibility) | **major** | Plans.md v1 removal |

**Batch release recommended**: When multiple changes occur on the same day, consolidate into a single minor. More than 2 minor bumps on the same day is prohibited.

## About NPM Distribution

This project is a Claude Code plugin and is not distributed as an npm package.
There is no `package.json` at the root (`core/package.json` is for internal TypeScript builds).
Version management targets only these 2 files:

- `VERSION` -- Source of truth
- `.claude-plugin/plugin.json` -- Plugin manifest

## Distribution Surfaces and Mirror Sync

`skills-v3/` is the SSOT (Single Source of Truth). The following 3 distribution surfaces are synced as mirrors:

| Surface | Path | Target Users |
|--------|------|------------|
| Claude | `skills/harness-release/` | Claude Code users |
| Codex | `codex/.codex/skills/harness-release/` | Codex CLI users |
| OpenCode | `opencode/skills/harness-release/` | OpenCode users |

**Important**: After editing `skills-v3/`, always sync mirrors before release:

```bash
./scripts/sync-v3-skill-mirrors.sh
```

Verification only (no writes):

```bash
./scripts/sync-v3-skill-mirrors.sh --check
```

## Internationalization (i18n)

Skill description fields can be switched between Japanese and English. Verify locale settings are as intended before release:

```bash
# Set to Japanese (description-ja -> description)
./scripts/i18n/set-locale.sh ja

# Set to English (description-en -> description)
./scripts/i18n/set-locale.sh en
```

Current default: description is in Japanese (identical to `description-ja`). `description-en` is always kept as an English backup.

## Execution Flow

### Phase 0: Pre-flight Check (required)

```bash
# 1. Required tools check
command -v gh &>/dev/null || echo "gh missing: GitHub Release will be skipped"
command -v jq &>/dev/null || echo "jq missing: required for plugin.json update"

# 2. Vendor-neutral preflight (shared for actual run / dry-run)
bash scripts/release-preflight.sh

# 3. Plugin structure validation
bash tests/validate-plugin.sh

# 4. Consistency check
bash scripts/ci/check-consistency.sh

# 5. Mirror sync state check
bash scripts/sync-v3-skill-mirrors.sh --check
```

`scripts/release-preflight.sh` validates:

- Is the working tree clean?
- Does `CHANGELOG.md` contain `[Unreleased]`?
- `.env.example` vs `.env` diff (warning-level for managed secrets)
- `healthcheck` / `preflight` commands (run if present)
- No debug / mock / placeholder residuals in `agents/` / `core/` / `hooks/` / `scripts/` shipped surface
- CI status (when available)

Environment variables for per-repository adjustment:

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
# Patch bump (x.y.Z -> x.y.(Z+1))
./scripts/sync-version.sh bump

# Minor bump (manual: x.Y.z -> x.(Y+1).0)
CURRENT=$(cat VERSION)
MAJOR=$(echo "$CURRENT" | cut -d. -f1)
MINOR=$(echo "$CURRENT" | cut -d. -f2)
NEW_VERSION="$MAJOR.$((MINOR + 1)).0"
echo "$NEW_VERSION" > VERSION
./scripts/sync-version.sh sync

# Major bump (manual: X.y.z -> (X+1).0.0)
CURRENT=$(cat VERSION)
MAJOR=$(echo "$CURRENT" | cut -d. -f1)
NEW_VERSION="$((MAJOR + 1)).0.0"
echo "$NEW_VERSION" > VERSION
./scripts/sync-version.sh sync
```

`sync-version.sh sync` propagates the `VERSION` value to `.claude-plugin/plugin.json`.

### Phase 3: CHANGELOG Update

Release entries finalize changes accumulated under `[Unreleased]` from normal PRs into a versioned section.

Write in **detailed Before/After format** (Japanese) for the CHANGELOG.
Separate each feature into numbered sections and explain "Before" and "After" with concrete examples.

```markdown
## [X.Y.Z] - YYYY-MM-DD

### Theme: [one-line summary of all changes]

**[1-2 sentences about user value]**

---

#### 1. [Feature name]

**Before**: [Old behavior described concretely. Describe the problem users experienced]

**After**: [New behavior described concretely. What gets resolved]

```
[actual output or command examples]
```

#### 2. [Next feature name]

**Before**: ...

**After**: ...
```

**CC version integration pattern**: When integrating CC updates, use the "CC update -> Harness utilization" format instead of the usual "Before / After".
See "CC Version Integration CHANGELOG Pattern" in `.claude/rules/github-release.md` for details.

**Writing rules**:

| Rule | Description |
|--------|------|
| Language | **Japanese** |
| Each feature as independent section | Numbered as `#### N. Feature name` |
| "Before" describes the problem | Concretely describe the inconvenience users experienced |
| "After" shows the resolution | What changes and how + concrete examples (code/output) |
| Always include concrete examples | Command examples, output examples, Plans.md snippets, etc. |
| Keep technical details minimal | File names and step numbers as minimal supplements in "After" |
| Length is OK | 3-10 lines per feature. Readability is top priority |

Keep the `[Unreleased]` section for the next release instead of removing it:

```markdown
## [Unreleased]

## [X.Y.Z] - YYYY-MM-DD
...
```

### Phase 4: Version File Update

```bash
# VERSION already updated in Phase 2
# Sync plugin.json
./scripts/sync-version.sh sync

# Verify sync
./scripts/sync-version.sh check
```

### Phase 5: Mirror Sync

```bash
# Mirror sync from skills-v3 -> skills, codex, opencode
./scripts/sync-v3-skill-mirrors.sh

# Verify sync
./scripts/sync-v3-skill-mirrors.sh --check
```

### Phase 6: Commit & Tag

```bash
NEW_VERSION=$(cat VERSION)

# Staging (explicitly specify target files)
git add VERSION .claude-plugin/plugin.json CHANGELOG.md
git add skills/ codex/.codex/skills/ opencode/skills/

git commit -m "chore: release v$NEW_VERSION"
git tag -a "v$NEW_VERSION" -m "Release v$NEW_VERSION"
```

### Phase 7: Push

```bash
git push origin main --tags
```

**Note**: `.github/workflows/release.yml` detects tag pushes and auto-generates a GitHub Release from CHANGELOG if one doesn't exist yet (safety net). If you create the GitHub Release manually first, the workflow auto-skips.

### Phase 8: Create GitHub Release

```bash
NEW_VERSION=$(cat VERSION)

gh release create "v$NEW_VERSION" \
  --title "v$NEW_VERSION - Title" \
  --notes "$(cat <<'EOF'
## What's Changed

**[Change summary (English)]**

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
- Language: **English** (public repository)
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

Invoke the `/x-announce` skill to generate X (Twitter) announcement threads:

```
Skill: x-announce
Args: v$NEW_VERSION
```

Generates 5 post texts + 5 Gemini images in one shot.

## `--dry-run` Mode

`--dry-run` executes the following without making actual changes:

1. Run **all** Pre-flight checks (Phase 0)
2. Display version calculation (no write)
3. Display CHANGELOG draft (no write)
4. Display GitHub Release Notes draft (no creation)
5. Display mirror sync diff (no write)

Skipped: VERSION/plugin.json writes, git commit/tag/push, GitHub Release creation, announcement

## `--complete` Mode

Only performs "release complete" marking after tag creation:

```bash
/release --complete
```

Executes Phase 9 only. Verifies no GitHub Release creation was missed before committing the completion marker.

## Regression Checklist

Verify the following regressions before release:

| Check Item | Verification Method | Notes |
|------------|---------|------|
| Plugin structure | `tests/validate-plugin.sh` | Validates plugin.json, skills, hooks, scripts |
| Consistency | `scripts/ci/check-consistency.sh` | Templates, versions, mirrors, CHANGELOG |
| Mirror sync | `scripts/sync-v3-skill-mirrors.sh --check` | Match between skills-v3 and 3 distribution surfaces |
| Preflight | `scripts/release-preflight.sh` | Working tree, CHANGELOG, CI, residuals |
| Release notes | `scripts/validate-release-notes.sh vX.Y.Z` | GitHub Release format validation |
| VERSION sync | `scripts/sync-version.sh check` | Match between VERSION and plugin.json |
| Guardrails | `core/src/guardrails/rules.ts` R01-R13 | TypeScript rule health |
| Tag continuity | `git tag --sort=-version:refname \| head -5` | No gaps in sequence |
| Locale | Match between description and description-ja | Switchable via `set-locale.sh` |

## CI Safety Net

`.github/workflows/release.yml` runs automatically on tag push:

1. Detects `v*` tag push
2. Checks if a GitHub Release with the same name already exists
3. If not, auto-generates from CHANGELOG (safety net)
4. If exists, does nothing

Creating the GitHub Release manually before pushing is the recommended flow.
The safety net only rescues "forgotten Release creation."

## PM Handoff

Completion report to PM after release:

```markdown
## Release Completion Report

**Version**: v{{NEW_VERSION}}
**Release Date**: {{DATE}}

### Changes Implemented
{{CHANGELOG contents}}

### GitHub Release
{{URL}}

### Next Actions
- PM review of release notes
- Production deployment (if applicable)
```

## Prohibited Actions

- Deleting or rewinding tags (published versions are immutable)
- More than 2 minor bumps on the same day
- Minor bump for patch-level changes
- Force push via `--force` / `--force-with-lease`
- Mixing implementation changes with release commits (only VERSION / plugin.json / CHANGELOG allowed)

## Related Skills

- `harness-review` -- Conduct code review before release
- `harness-work` -- Implement next tasks after release
- `harness-plan` -- Create plan for next version
- `x-announce` -- Generate X (Twitter) release announcement threads
- `harness-setup` -- Mirror sync and plugin configuration setup

## Related Rules

- `.claude/rules/versioning.md` -- SemVer determination criteria and batch release recommendations
- `.claude/rules/github-release.md` -- GitHub Release Notes format (English)
- `.claude/rules/cc-update-policy.md` -- Feature Table quality criteria for CC update tracking
