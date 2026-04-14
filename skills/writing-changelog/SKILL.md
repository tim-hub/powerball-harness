---
name: writing-changelog
description: "Use when writing or updating CHANGELOG.md — adding entries to [Unreleased], finalizing a versioned release section, or writing GitHub Release notes. Do NOT load for: running the full release workflow (harness-release), code review, or implementation."
allowed-tools: ["Read", "Write", "Edit"]
argument-hint: "[unreleased|release|github-release]"
---

# Writing Changelog

CHANGELOG entries in this project use a **detailed Before/After format** — each change explains what the user experienced before and what they get after, with concrete examples.

## When to Use Each Mode

| User Says | Mode | Action |
|-----------|------|--------|
| Finished a feature / fixed a bug | `unreleased` | Append to `## [Unreleased]` section |
| Cutting a release | `release` | Finalize `[Unreleased]` into a versioned section |
| Writing GitHub Release notes | `github-release` | Write English release notes for `gh release create` |

## CHANGELOG Entry Format

### Structure

```markdown
## [X.Y.Z] - YYYY-MM-DD

### Theme: [One-line summary of all changes]

**[Value to users in 1-2 sentences]**

---

#### 1. [Feature Name]

**Before**: [Describe the old behavior concretely — the inconvenience users experienced]

**After**: [Describe the new behavior — what changed and how, with concrete examples]

```
[Actual output or command example]
```

#### 2. [Next Feature Name]

**Before**: ...

**After**: ...
```

### Writing Rules

| Rule | Guidance |
|------|----------|
| Language | **English** |
| Structure | Each feature in its own `#### N. Feature Name` section |
| "Before" | Describes the **pain point** — what users had to do or put up with |
| "After" | Shows the **resolution** — what changes and how, with concrete command/output examples |
| Examples | Always include real command output, Plans.md snippets, or before/after code |
| Detail level | 3–10 lines per feature; readability beats brevity |
| Technical detail | File names and step numbers are supplements in "After", not the headline |

### Preserving `[Unreleased]`

Never empty the `[Unreleased]` section — keep it as a placeholder for the next release:

```markdown
## [Unreleased]

## [X.Y.Z] - YYYY-MM-DD
...
```

## CC Version Integration Pattern

For releases that bundle a new Claude Code version, use the **"CC Update → Harness Usage"** format instead of Before/After. This makes it clear which changes come from upstream CC and which are Harness-specific responses.

See `.claude/rules/github-release.md` — "CHANGELOG Pattern for CC Version Integration" for the full template and examples.

## GitHub Release Notes Format

GitHub Releases use **English** and a condensed format with a Before/After table:

```markdown
## What's Changed

**[One-line value description]**

### Before / After

| Before | After |
|--------|-------|
| Previous state | New state |

---

## Added
- **Feature**: Description

## Changed / Fixed
- **Change**: Description


```

Full format rules: `.claude/rules/github-release.md`

## Related

- `harness-release` — Full release workflow (calls this skill at Phase 3)
- `.claude/rules/github-release.md` — Detailed GitHub Release format + CC integration pattern
- `.claude/rules/versioning.md` — Deciding patch / minor / major
