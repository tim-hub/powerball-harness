---
name: remember-this
description: "Syncs this repo's project specs with powerball-harness workflow conventions. Use when aligning Plans.md markers, PM/Impl handoff routes, or project spec consistency for this plugin."
when_to_use: "sync project specs, remember-this, sync-across, align markers, PM handoff routes, AGENTS.md PM/Impl, powerball-harness workflow, Plans.md marker normalization"
allowed-tools: ["Read", "Write", "Edit", "Bash", "Grep"]
argument-hint: "[sync-across]"
context: fork
---

# Remember This

Project-specific memory operations for the `claude-code-harness` plugin repo.

Handles spec alignment tasks that reference `powerball-harness` operations, Plans.md marker conventions, and PM/Impl handoff routes — things that don't belong in the generic `harness-remember` skill.

## Quick Reference

| User Input | Subcommand | Behavior |
|------------|------------|----------|
| "Sync project specs" / "align Plans.md markers" / `remember-this sync-across` | `sync-across` | Align Plans.md, AGENTS.md, rules/ with current powerball-harness conventions (see `references/sync-project-specs.md`) |

## Execution Steps

1. Classify the user's request
2. Read the appropriate reference file from "Quick Reference" above
3. Execute according to its contents

## Related Skills

- `harness-remember` — Generic SSOT operations (decisions.md, patterns.md, search, record)
- `harness-plan sync` — Progress sync and marker drift detection
