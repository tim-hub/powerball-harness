# Distribution Scope

Last updated: 2026-03-06

This document is a scope table that explicitly defines items that "exist in the repo but are not always distributed in the same form" for `claude-code-harness`.
When uncertain in `Plans.md`, README, distribution scripts, or validation scripts, treat this table as the authoritative source.

## Scope Table

| Path | Status | Why it exists | Enforcement signal |
|------|--------|---------------|--------------------|
| `core/` | Distribution-included | TypeScript guardrail engine core | `core npm test`, README architecture |
| `skills/` | Distribution-included | Current 5 verb skills | README, mirror sync checks |
| `agents/` | Distribution-included | Current worker / reviewer / scaffolder | README, validate-plugin-v3 |
| `hooks/` | Distribution-included | Runtime guardrails and lifecycle hooks | `hooks/hooks.json`, validate-plugin |
| `scripts/hook-handlers/memory-bridge.sh`, `scripts/hook-handlers/memory-*.sh` | Distribution-included | Bridge and wrapper for harness-mem integration. Hooks reference the stable bridge, wrappers are for compatibility and testing | `validate-plugin`, `test-memory-hook-wiring.sh` |
| `templates/` | Distribution-included | Authoritative source for project init and rules distribution | `check-consistency.sh` |
| `codex/` | Distribution-included | Codex CLI distribution (symlinked skills) | `tests/test-codex-package.sh` |
| `commands/` | Compatibility-retained | Legacy slash commands retained for backwards compatibility; skills/ is the SSOT going forward | README, validate-plugin |
| `mcp-server/` | Development-only and distribution-excluded | Optional feature. Retained in repo for development/investigation but not included in distribution payload | `.gitignore`, CHANGELOG history |
| `harness-ui/`, `harness-ui-archive/` | Development-only and distribution-excluded | Optional UI experiments and legacy implementation archive | `.gitignore`, CHANGELOG history |
| `docs/research/`, `docs/private/` | Private reference | Comparison notes, investigation records, pre-publication drafts | repo reference only |

## Current Decisions

- `commands/` is not treated as deleted. Currently **Compatibility-retained**.
- `mcp-server/` is not treated as deleted. Currently **Development-only and distribution-excluded**.
- `scripts/hook-handlers/memory-bridge.sh` and `memory-*.sh` are **Distribution-included** even as local bridges. They need to be tracked in the repo because hooks reference them.
- When writing "deleted" in README or `Plans.md`, only use it when items have actually been removed from the tree.
- Use "distribution-excluded," "compatibility-retained," and "development-only" according to the labels in this document.

## Update Rule

When any of the following occur, update this table in the same PR / commit.

1. When changing README architecture / install / compatibility descriptions
2. When changing `.gitignore` or build script exclusion rules
3. When changing how directories prone to misunderstanding (like `commands/` or `mcp-server/`) are handled
