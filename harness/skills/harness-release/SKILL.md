---
name: harness-release
description: "Manages version bumps, CHANGELOG, git tags, and GitHub Releases. Use when cutting a release or updating version metadata."
when_to_use: "release, bump version, update CHANGELOG, create tag, GitHub release, patch release, minor release"
allowed-tools: ["Read", "Write", "Edit", "Bash", "AskUserQuestion", "Skill"]
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
| `harness-release` (bare) | _(auto)_ | Review gate → commit reviewed work → release |

Key paths:
- Preflight script: `${CLAUDE_SKILL_DIR}/scripts/release-preflight.sh`
- Override plugin root: `HARNESS_RELEASE_PLUGIN_ROOT=/path/to/repo`

**Version classification and batch-release rules**: [`${CLAUDE_SKILL_DIR}/references/versioning-rules.md`](${CLAUDE_SKILL_DIR}/references/versioning-rules.md)

## Bare invocation contract

if $ARGUMENTS == "":
  → Interpret as "commit and release the work done so far"; start Review Gate detection
  → Auto-advance to Review Gate only when exactly one work target can be identified
  → When target is ambiguous or no review state exists: show AskUserQuestion with options, then proceed

Emit this literal marker in the first response on bare invocation:
`RELEASE_AUTOSTART: target=<work-summary>, base_ref=<ref>, mode=<patch|minor|major|auto>`

**Forbidden actions on bare invocation**: "task unclear", "awaiting instructions", "no tasks found", "awaiting further instructions".

<!-- AUTO-START CONTRACT: P27 solution 3-piece set (machine-readable condition + forbidden-action literals + AUTOSTART marker) -->

When `harness-release` is invoked without a bump level (bare), treat it as:
**"commit and release the work done so far."**

Do not stop with "no tasks found" or "awaiting instruction." Instead, before the regular release preflight, run two pre-stages:

**Review Gate:**

1. Check `git status --porcelain` and `git log @{upstream}..HEAD` to identify the work in scope
2. Read `.claude/state/review-result.json` and `.claude/state/review-approved.json` — verify APPROVE exists for the work
3. If no APPROVE review is found, use `AskUserQuestion`:

```text
harness-release will commit and release work done so far, but no APPROVE review was found.
How would you like to proceed?
  - Start with review (Recommended): run harness-review; only advance to commit/release on APPROVE
  - release dry-run: preview the release plan without writing files
  - Abort: stop without review or release
```

4. If the user selects "Start with review": invoke `harness-review`; do not advance until APPROVE
5. If `harness-review` returns REQUEST_CHANGES: hold the release; fix with `harness-work`, then re-run `harness-review`; repeat until APPROVE
6. After APPROVE: create the work commit (see Work Commit Gate below)
7. After clean working tree: proceed to the normal release preflight → confirmation gate → tag → GitHub Release

Return control to the user only when:
- A Plans.md / spec / API / permission / migration decision is needed via `AskUserQuestion`
- Multiple fix strategies exist and the choice changes user value or compatibility
- The user chose "dry-run" or "Abort" from the Ask

Do not treat REQUEST_CHANGES alone as a terminal stop.

**Work Commit Gate:**

When the working tree has uncommitted changes under a bare release, create a work commit separately from the version bump commit:

```bash
git status --short
git diff --stat
git add <reviewed files>
git commit -m "<type>: <summary>"
```

Generate the commit message from the review summary, Plans.md task, or branch name.
If judgment is unclear, use `AskUserQuestion` to present 2–3 candidate messages.
After creating the work commit, verify or update `commit_hash` in `.claude/state/review-result.json`,
then proceed to the regular release preflight.

Once inside the regular release preflight, the working tree dirty check applies as normal.
Do not advance to version bump / tag / GitHub Release with a dirty tree.

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

**Delegation**: delegate to `harness-releaser` agent via `setup` invocation (runs Phase 0 + 1 + 2 together).

```bash
command -v gh &>/dev/null || echo "gh missing: GitHub Release will be skipped"
command -v jq &>/dev/null || echo "jq missing: required for manifest updates"
bash "${CLAUDE_SKILL_DIR}/scripts/release-preflight.sh"
```

Validates: clean working tree, `[Unreleased]` section exists, CI status, no debug/mock remnants.
Adjustable via `HARNESS_RELEASE_PLUGIN_ROOT`, `HARNESS_RELEASE_HEALTHCHECK_CMD`, `HARNESS_RELEASE_CI_STATUS_CMD`.

> **Bare release note**: This working tree clean check is the standard release preflight gate.
> For bare releases where uncommitted work exists, complete the Review Gate and Work Commit Gate
> (see above) before this check. Do not abort on a dirty tree alone without first running those gates.

### Phase 1: Get Current Version

**Delegation**: included in the `harness-releaser` agent `setup` invocation.

```bash
CURRENT=$(cat VERSION 2>/dev/null)
```

### Phase 2: Calculate New Version

**Delegation**: included in the `harness-releaser` agent `setup` invocation.

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

### Phase 2.5 → Agent `setup` invocation

Delegates Phases 0–2 to the `harness-releaser` agent:

```
Agent(
  subagent_type: "harness-releaser",
  description: "run preflight and bump version",
  prompt: "{\"schema_version\":\"releaser-request.v1\",\"invocation\":\"setup\",\"bump_type\":\"<patch|minor|major>\"}"
)
```

Parse `new_version` from the `releaser-response.v1` response. If `status` is `"error"`, surface the `error` field to the user and abort.

> **Smoke test required**: The releaser agent path was introduced in Phase 106. Run `/harness-release --dry-run` to validate the setup invocation before using in a live release.

### Phase 3: CHANGELOG Update

**skill-owned — drafting required, no agent delegation**

Move `[Unreleased]` content into a new `## [X.Y.Z] - YYYY-MM-DD` section.
Keep an empty `## [Unreleased]` placeholder above it.
Use [`${CLAUDE_SKILL_DIR}/references/writing-changelog.md`](${CLAUDE_SKILL_DIR}/references/writing-changelog.md) for format rules and the Before/After template.

### Phase 4: Commit & Tag

**Delegation**: delegate to `harness-releaser` agent via `finalize` invocation (runs Phase 4 + 5 together).

```bash
NEW_VERSION=$(cat VERSION)
git add VERSION CHANGELOG.md
git commit -m "chore: release v$NEW_VERSION"
git tag -a "v$NEW_VERSION" -m "Release v$NEW_VERSION"
```

### Phase 5: Push

**Delegation**: included in the `harness-releaser` agent `finalize` invocation.

```bash
git push origin {main or master} --tags
```

### Phase 4.5 + 5 → Agent `finalize` invocation

After Phase 3 CHANGELOG is written, delegates Phases 4–5 to the `harness-releaser` agent:

```
Agent(
  subagent_type: "harness-releaser",
  description: "commit, tag, push release v<new_version>",
  prompt: "{\"schema_version\":\"releaser-request.v1\",\"invocation\":\"finalize\",\"version\":\"<new_version>\"}"
)
```

Parse `git_hash` from the response. If `status` is `"error"`, surface the `error` and `changes` fields (idempotent — re-invocation after fixing the issue is safe).

`.github/workflows/release.yml` auto-generates a GitHub Release from CHANGELOG on tag push if one hasn't been created yet.

### Phase 6: Create GitHub Release

**skill-owned — drafting required, no agent delegation**

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

↑ Claude will summarize this result. Type a new prompt to redirect or press Enter to continue.

## Related

- [`${CLAUDE_SKILL_DIR}/references/writing-changelog.md`](${CLAUDE_SKILL_DIR}/references/writing-changelog.md) — CHANGELOG format and Before/After template
- [`${CLAUDE_SKILL_DIR}/references/versioning-rules.md`](${CLAUDE_SKILL_DIR}/references/versioning-rules.md) — SemVer classification and batch-release rules

## Related Agents

- `harness-releaser` — Haiku executor for Phases 0–2 (`setup`) and Phases 4–5 (`finalize`); never drafts content
