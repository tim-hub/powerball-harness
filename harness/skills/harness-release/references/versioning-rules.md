# Versioning Rules

Version management standards for Harness. Follows SemVer.

## Version Classification Criteria

| Type of Change | Version | Example |
|----------------|---------|---------|
| Wording fixes or additions to skill definitions (SKILL.md) | **patch** (x.y.Z) | Minor template fixes, description improvements |
| Documentation or rule file updates | **patch** (x.y.Z) | CHANGELOG rewrites, rules/ additions |
| Bug fixes in hooks/scripts | **patch** (x.y.Z) | Escape fix in task-completed.sh |
| Adding new flags/subcommands to existing skills | **minor** (x.Y.0) | `--snapshot`, `--auto-mode` |
| Adding new skills/agents/hooks | **minor** (x.Y.0) | New skill `harness-foo` |
| Changes to the Go guardrail engine | **minor** (x.Y.0) | New rule additions, existing rule changes |
| Claude Code new version compatibility | **minor** (x.Y.0) | CC v2.1.72 support |
| Breaking changes (skill deprecation, format incompatibility) | **major** (X.0.0) | Plans.md v1 support removal |

## Decision Flowchart

```
Does existing behavior break?
├─ Yes → major
└─ No → Can the user do something new?
    ├─ Yes → minor
    └─ No → patch
```

## Batch Release Recommendations

- **Multiple phases completed same day**: Combine into a single minor release
- **Phase completion + documentation fixes**: Use minor for the Phase; bundle docs (don't create separate release)
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

1. List all changes since the last release
2. Determine version type based on the classification criteria above
3. Consider batching multiple same-day changes
4. Verify sync across VERSION / plugin.json / CHANGELOG
5. Verify git tags are sequential with no gaps

## Version Distribution

The canonical version lives in `VERSION`. Sync it to additional manifests (package.json, Cargo.toml, pyproject.toml, etc.) as needed — see `sync-version.sh sync` documentation.

## Prohibited

- Deleting or rolling back tags (published versions are immutable)
- More than one minor bump on the same day
- Using a minor bump for patch-level changes
- Force push via `--force` / `--force-with-lease`
- Mixing implementation changes other than VERSION / CHANGELOG into release commits
