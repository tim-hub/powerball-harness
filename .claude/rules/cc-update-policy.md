---
description: Quality policy for tracking CC updates
globs: ["CLAUDE.md", "docs/CLAUDE-feature-table.md"]
---

# CC Update Tracking Policy

Quality standards for updating the Feature Table when adding support for new Claude Code versions.

## Core Principle

Additions to the Feature Table must be accompanied by **corresponding implementation changes** or **explicit classification as Category C (CC auto-inherited)**.

A PR that only adds rows to the Feature Table without implementation must not be merged.

## 3-Category Classification

| Category | Definition | PR Merge |
|---------|------|----------|
| **(A) Has Implementation** | Corresponding implementation changes exist in hooks / scripts / agents / skills / core | Allowed |
| **(B) Documentation Only** | Only the Feature Table was changed. No implementation | **Not allowed** -- An implementation proposal is required |
| **(C) CC Auto-Inherited** | CC core fix requires no Harness-side changes (performance improvements, bug fixes, etc.) | Allowed (must note "CC auto-inherited" in the Feature Table) |

## Rules

### 1. Feature Table Additions Must Include Implementation or Classification

When adding new rows to the Feature Table, one of the following must be satisfied:

- **(A)** The same PR contains corresponding implementation file changes
- **(C)** The Feature Table entry explicitly states it is "CC auto-inherited"

If neither applies, the item is classified as Category B (documentation only).

### 2. Block PR and Request Implementation Proposal When Category B Is Detected

If even one Category B item exists:

- **Block** the PR from merging
- Request an **implementation proposal** for each Category B item, including:
  - Explanation of the unique value Harness provides
  - Target files and specific changes
  - User experience improvement (before / after)

After the implementation proposal is approved, create additional commits or a follow-up PR that includes the implementation.

### 3. Adding a "Value-Add" Column Is Recommended

Adding a "Value-Add" column to the Feature Table to visualize A / B / C classification is recommended.

```markdown
| Feature | Skill | Purpose | Value-Add |
|---------|-------|---------|---------|
| PostCompact hook | hooks | Context re-injection | A: Has implementation |
| Streaming leak fix | all | Memory leak fix | C: CC auto-inherited |
```

This column enables:
- Immediate detection of remaining Category B items during review
- Self-documenting why each Feature Table entry exists
- Reference to past decisions during future CC update integrations

## Scope

This policy applies when modifying the following files:

- The Feature Table section of `CLAUDE.md`
- `docs/CLAUDE-feature-table.md`

It does not apply to regular implementation PRs, documentation fixes, or release operations.
