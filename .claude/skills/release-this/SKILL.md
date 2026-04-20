---
name: release-this
description: "Orchestrates the full powerball-harness plugin release with build, validation, and version checks. Use when releasing this plugin."
when_to_use: "release this plugin, release harness, cut a release, publish harness, ship harness, release-this"
allowed-tools: ["Read", "Write", "Edit", "Bash"]
argument-hint: "[patch|minor|major|--dry-run|--complete]"
model: sonnet
effort: low
---

# Release This Plugin

Project-specific release orchestrator for the claude-code-harness plugin.
Runs all plugin-specific checks (build, consistency, validation, symlinks, version sync) **before** delegating to the generic `harness-release` skill for the actual release (version bump, CHANGELOG, tag, GitHub Release).

## Quick Reference

| Argument | Behavior |
|----------|----------|
| `patch` | Patch version bump (bug fixes, x.y.Z+1) |
| `minor` | Minor version bump (new features, x.Y+1.0) |
| `major` | Major version bump (breaking changes, X+1.0.0) |
| `--dry-run` | Run steps 1–5 only; print what would happen; skip publishing |
| `--complete` | Mark release complete only (step 7 — empty commit + push) |

## Execution Flow

### Step 1: Cross-platform binary build if there are changes in `go/` since the last release

Run `make build-all` to compile darwin-arm64, darwin-amd64, and linux-amd64 binaries unless there are no changes in `go/` source files since the last release. This ensures that any build issues are caught before proceeding with release validation or publishing steps.

```bash
make build-all
```

Fail immediately if `make build-all` exits non-zero. Do not proceed to step 2.

### Step 2: Plugin consistency checks

Run the 13 plugin-specific consistency checks. These verify symlink integrity, template consistency, version alignment, and structural invariants.

```bash
bash "${CLAUDE_SKILL_DIR}/scripts/check-consistency.sh"
```

### Step 3: Full plugin validation

```bash
bash ./tests/validate-plugin.sh
```

This runs all CI-gated checks: marketplace.json structure, skill frontmatter, hook schemas, section 9 migration residue scan, and section 10 skill-description audit.

### Step 4: Codex symlink verification

Verify that `codex/.codex/skills/` symlinks resolve correctly. Broken symlinks would prevent Codex CLI users from loading skills.

```bash
ls -la codex/.codex/skills/
```

Confirm all listed entries are valid symlinks (no broken/dangling entries).

### Step 5: Version sync check

Verify that `harness/VERSION` and `harness/harness.toml` agree before proceeding.

```bash
bash harness/skills/harness-release/scripts/sync-version.sh check
```

If there is a mismatch, stop and ask the user to run `./harness/skills/harness-release/scripts/sync-version.sh bump` or manually reconcile the files.

---

**Dry-run stops here.** In `--dry-run` mode, steps 1–5 run normally but no release is published, no files are written, and no git commits or tags are created. Report what the release would contain and what the new version would be, then exit.

---

### Step 6: Invoke harness-release

Delegate to the generic release skill using the same argument provided by the user (patch / minor / major).

```
/harness-release <patch|minor|major>
```

The `harness-release` skill handles:
- Phase 0: Pre-flight checks (`release-preflight.sh`)
- Phase 1–2: Version display and bump
- Phase 3: CHANGELOG update (`[Unreleased]` → versioned entry)
- Phase 4: Commit and tag (`chore: release vX.Y.Z`)
- Phase 5: Push branch and tags
- Phase 6: GitHub Release creation

Wait for `harness-release` to complete successfully before proceeding to step 7.

### Step 7: Completion marking commit

After `harness-release` finishes, create an empty commit to mark the release as fully complete, then push it.

```bash
NEW_VERSION=$(cat harness/VERSION)
git commit --allow-empty -m "chore: mark v${NEW_VERSION} release complete"
git push origin "$(git rev-parse --abbrev-ref HEAD)"
```

This empty commit is the explicit record that "all release work is done." It is separate from the `chore: release vX.Y.Z` commit created by `harness-release`.

---

## `--complete` Mode

When called with `--complete`, execute only step 7 (the completion marking commit). Use this if `harness-release` already finished but the completion commit was not created.

```bash
NEW_VERSION=$(cat harness/VERSION)
# Verify a GitHub Release exists for this version before marking complete
gh release view "v${NEW_VERSION}" || { echo "ERROR: GitHub Release v${NEW_VERSION} not found. Create it first."; exit 1; }
git commit --allow-empty -m "chore: mark v${NEW_VERSION} release complete"
git push origin "$(git rev-parse --abbrev-ref HEAD)"
```

## Related Skills

- `harness-release` — Generic release engine (delegated at step 6)
- `harness-review` — Run before release to catch issues early
- `harness-plan` — Plan the next release tasks
- `harness-work` — Work on tasks
- `harness-loop` — Run review-work iterations on release tasks

## Related Rules

- `.claude/rules/versioning.md` — SemVer classification criteria and batch release policy
- `.claude/rules/github-release.md` — GitHub Release Notes format
- `harness/skills/harness-release/SKILL.md` — Generic release skill (delegated in step 6)
