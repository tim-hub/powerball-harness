# Versioning Rules

Version management standards for Harness. Follows SemVer (Semantic Versioning).

## Version Classification Criteria

| Type of Change | Version | Example |
|-----------|----------|-----|
| Wording fixes or additions to skill definitions (SKILL.md) | **patch** (x.y.Z) | Minor template fixes, description improvements |
| Documentation or rule file updates | **patch** (x.y.Z) | CHANGELOG rewrites, rules/ additions |
| Bug fixes in hooks/scripts | **patch** (x.y.Z) | Escape fix in task-completed.sh |
| Adding new flags/subcommands to existing skills | **minor** (x.Y.0) | `--snapshot`, `--auto-mode` |
| Adding new skills/agents/hooks | **minor** (x.Y.0) | New skill `harness-foo` |
| Changes to the TypeScript guardrail engine | **minor** (x.Y.0) | New rule additions, existing rule changes |
| Claude Code new version compatibility | **minor** (x.Y.0) | CC v2.1.72 support |
| Breaking changes (skill deprecation, format incompatibility) | **major** (X.0.0) | Plans.md v1 support removal |

## Decision Flowchart

```
Does it break existing behavior?
├─ Yes → major
└─ No → Does the user gain new capabilities?
    ├─ Yes → minor
    └─ No → patch
```

## Batch Release Recommendations

- **When multiple Phases are completed on the same day**: Combine into a single minor release
- **Phase completion + documentation fixes**: Use minor for the Phase; bundle documentation fixes (don't create a separate release)
- **CC compatibility + feature additions**: May be combined into a single minor

### Bad Example

```
v3.6.0 (03/08 AM) — Phase 25
v3.7.0 (03/08 PM) — Phase 26    ← Avoid 2 minor bumps on the same day
v3.7.1 (03/09)    — Auto Mode
```

### Good Example

```
v3.6.0 (03/08) — Phase 25 + Phase 26    ← Combined into 1 minor
v3.6.1 (03/09) — Auto Mode prep         ← prep is patch
```

## Pre-Release Checklist

1. **List all changes since the last release**
2. **Determine version type based on the classification criteria**
3. **Consider batching multiple same-day changes**
4. **Verify sync across VERSION / plugin.json / CHANGELOG**
5. **Verify git tags are sequential with no gaps**

## Prohibited

- Deleting or rolling back tags (published versions are immutable)
- More than one minor bump on the same day
- Using a minor bump for patch-level changes
