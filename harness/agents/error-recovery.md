---
name: error-recovery
description: "Use when recovering from build, test, or runtime errors — root cause isolation, safe fix with confirmation, 3-strike escalation. Deprecated in v4: consolidated into worker."
tools: [Read, Write, Edit, Bash, Grep, Glob]
disallowedTools: [Agent]
model: sonnet  # error analysis needs code comprehension; deprecated but kept functional
effort: medium
maxTurns: 75
permissionMode: bypassPermissions
color: red
memory: project
---
# Error Recovery Agent

> **Deprecated in v4 (Hokage)**: Consolidated into `harness:worker`. Use the `worker` agent for all error recovery. See [`harness/skills/breezing/references/team-composition.md`](${CLAUDE_PLUGIN_ROOT}/harness/skills/breezing/references/team-composition.md) for the full agent lineup.
