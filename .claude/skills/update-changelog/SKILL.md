---
name: update-changelog
description: "Generates CHANGELOG.md entries after a version bump. Use this skill whenever the version in marketplace.json has been changed, after a release, when the user says 'update changelog', 'generate changelog', 'write release notes', or asks to document what changed between versions. Also trigger when you detect that marketplace.json version differs from the latest CHANGELOG entry. Do NOT load for: reading the changelog, planning work, or code implementation."
allowed-tools: ["Read", "Write", "Edit", "Bash", "Grep", "Glob"]
argument-hint: "[version]"
---

# Update Changelog

Generate a CHANGELOG.md entry by comparing the current state against the last released version. This skill follows the project's Before/After changelog format documented in `.claude/rules/github-release.md`.

## When to Run

- After bumping the version in `.claude-plugin/marketplace.json`
- When the user asks to update or generate changelog entries
- When `[Unreleased]` section in CHANGELOG.md is empty but there are unreleased commits

## Execution Flow

### Step 1: Determine version boundaries

Read the current version from `.claude-plugin/marketplace.json` (field: `plugins[0].version`).

Find the previous version by checking git tags:

```bash
# Get the two most recent version tags
git tag --sort=-v:refname | head -5
```

If no tags exist, fall back to reading the CHANGELOG.md to find the last `## [X.Y.Z]` entry and using `git log` to find approximate boundaries.

### Step 2: Gather changes

Collect all changes between the two versions:

```bash
# Commits between versions
git log v<old>..HEAD --oneline

# Files changed
git diff v<old>..HEAD --stat

# Detailed diff for understanding what changed
git diff v<old>..HEAD -- '*.md' '*.sh' '*.ts' '*.json'
```

Focus on user-facing changes. Internal refactors matter only if they change behavior.

### Step 3: Categorize changes

Group changes into logical features/fixes. Each group becomes a numbered section. Think about what the user experiences differently — not what files changed.

Categories to consider:
- New features or capabilities
- Changed behavior
- Bug fixes
- Removed functionality
- Claude Code version integration (use the CC Update/Harness Integration pattern)

### Step 4: Write the entry

Follow this exact format (matching existing CHANGELOG entries):

```markdown
## [X.Y.Z] - YYYY-MM-DD

### Theme: <one-line summary of the release>

**<1-2 sentence value description of what this release does for the user.>**

---

#### 1. <Feature or change name>

**Before**: <What the user experienced before. Describe the inconvenience or limitation concretely.>

**After**: <What's different now. Include concrete examples, commands, or output when helpful.>

#### 2. <Next feature or change>

**Before**: ...

**After**: ...
```

### Format Rules

These rules come from the project's `.claude/rules/github-release.md`:

- Every feature gets its own `#### N. Name` section
- **Before** describes the problem using "users had to..." pattern
- **After** provides a concrete picture of the solution with examples
- Keep technical details (file names, step numbers) minimal — supplement in "After"
- Readability is the top priority; longer entries are fine if they're clear
- For CC version integration changes, use the **CC Update** / **Harness Integration** pattern instead of Before/After

### Step 5: Insert into CHANGELOG.md

If there's an `## [Unreleased]` section with content, convert it to the versioned entry.

If `## [Unreleased]` is empty, insert the new entry between `## [Unreleased]` and the previous version entry.

Preserve the `## [Unreleased]` heading (leave it empty for future changes).

Also add the version comparison link at the bottom of CHANGELOG.md:

```markdown
[X.Y.Z]: https://github.com/tim-hub/powerball-harness/compare/vOLD...vNEW
```

### Step 6: Verify

After writing, verify:
- [ ] Version in CHANGELOG matches marketplace.json
- [ ] Date is today's date
- [ ] Theme line exists and is concise
- [ ] Bold summary line exists
- [ ] Every feature has Before/After (or CC Update/Harness Integration)
- [ ] `[Unreleased]` heading preserved at top
- [ ] Comparison link added at bottom

## Example: Small Patch Release

```markdown
## [3.17.1] - 2026-04-06

### Theme: harness-mem integration fix (emergency patch)

**Fixed an issue where harness-mem integration was broken for marketplace users.**

---

#### 1. Fix search paths in harness-mem-bridge.sh

**Before**: `harness-mem-bridge.sh` only searched hardcoded development paths.
For marketplace users, the search failed silently and no resume pack was generated.

**After**: Added the standard installation path as highest priority in search order.
Session resume now properly restores previous work context.
```

## Example: CC Integration Release

```markdown
#### 1. Auto mode denial tracking via PermissionDenied hook

**CC Update**: A `PermissionDenied` hook now fires when auto mode denies a command (v2.1.89).

**Harness Integration**: Implemented `permission-denied-handler.sh` to record denial events.
When a Breezing Worker is denied, Lead is notified to consider alternatives.
```
