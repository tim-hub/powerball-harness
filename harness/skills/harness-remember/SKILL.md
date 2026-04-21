---
name: harness-remember
description: "Manages SSOT memory — decisions.md, patterns.md, and cross-session learnings. Use when recording decisions, searching memory, or promoting learnings."
when_to_use: "record decision, search memory, update patterns, SSOT, promote learning, decisions.md"
allowed-tools: ["Read", "Write", "Edit", "Bash", "Grep", "mcp__harness__harness_mem_*"]
argument-hint: "[ssot|sync|sync-across|migrate|merge|search|record]"
context: fork
model: opus
effort: high
---

# Memory Skills

SSOT (Single Source of Truth) and cross-tool memory management for Harness.

## Quick Reference

| User Input | Subcommand | Behavior |
|------------|------------|----------|
| "Init SSOT" / `memory ssot` | `ssot` | Bootstrap `.claude/memory/decisions.md` + `patterns.md` (see `references/ssot-initialization.md`) |
| "Save what we learned" / "promote to SSOT" / `memory sync` | `sync` | Pull learnings from Claude Code auto memory (Layer 1) into SSOT (Layer 2) decisions/patterns (see `references/sync-ssot-from-memory.md`) |
| "Sync across agents" / `memory sync-across` | `sync-across` | Reconcile memory artifacts across agent workspaces / project specs (see `references/sync-project-specs.md`) |
| "Migrate from AGENTS.md" / `memory migrate` | `migrate` | Run interactive workflow migration (see `references/workflow-migration.md`) |
| "Merge Plans.md" / `memory merge` | `merge` | Consolidate multiple Plans.md files (see `references/plans-merging.md`) |
| "Search memory for X" / `memory search <term>` | `search` | Keyword/regex search; local first, MCP extends (see `references/search.md`) |
| "Record this decision" / `memory record` | `record` | Validate SSOT-worthiness, write local first, MCP mirrors (see `references/record.md`) |

> For `search` and `record`, local SSOT is authoritative and MCP extends reach — see each reference file for the full contract.

## Execution Steps

1. Classify the user's request
2. Read the appropriate reference file from "Quick Reference" above
3. Execute according to its contents

## Shared References

- [Unified Harness Memory (shared DB through harness_mem_* MCP)](${CLAUDE_SKILL_DIR}/references/harness-mem-mcp.md)

- Relationship with Claude Code Auto Memory (D22): Harness SSOT (Layer 2) coexists with Claude Code's auto memory (Layer 1). Auto memory records general learnings passively; SSOT explicitly curates project-specific decisions. Use `memory sync` when a Layer 1 observation has become important enough to preserve across sessions and contributors.
  - Details: see `decisions.md` in the project's `.claude/memory/` directory (project-root), entry D22: 3-Layer Memory Architecture.