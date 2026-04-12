# v3 Architecture Details

## Directory Structure

```
claude-code-harness/
├── core/           # TypeScript core engine
│   ├── src/
│   │   ├── index.ts          # stdin → route → stdout pipeline
│   │   ├── types.ts          # Type definitions (HookInput, HookResult, etc.)
│   │   └── guardrails/       # Guardrail engine
│   │       ├── rules.ts      # Declarative rule table (R01-R09)
│   │       ├── pre-tool.ts   # PreToolUse hook
│   │       ├── post-tool.ts  # PostToolUse hook
│   │       ├── permission.ts # PermissionRequest hook
│   │       └── tampering.ts  # Tampering detection
│   ├── package.json          # standalone TypeScript package
│   └── tsconfig.json         # strict, NodeNext ESM
├── skills-v3/      # 5-verb skills
│   ├── plan/       # planning + plans-management + sync-status consolidated
│   ├── execute/    # work + breezing + codex consolidated
│   ├── review/     # harness-review + codex-review consolidated
│   ├── release/    # release-har + handoff consolidated
│   ├── setup/      # harness-init + harness-mem consolidated
│   └── extensions/ # Extension packs (symlink → skills/)
├── agents-v3/      # 3 agents (11→3 consolidated)
│   ├── worker.md        # Implementation agent
│   ├── reviewer.md      # Review agent (Read-only)
│   ├── scaffolder.md    # Scaffolding and state update agent
│   └── team-composition.md  # Team composition guide
├── skills/         # Legacy skills (retained for backward compatibility)
├── hooks/          # Thin shims (→ delegates to core/src/index.ts)
└── .claude/
    └── agent-memory/
        ├── claude-code-harness-worker/
        ├── claude-code-harness-reviewer/
        └── claude-code-harness-scaffolder/
```

## 5-Verb Skill Mapping

| v3 Skill | Consolidated From (Legacy Skills) |
|----------|----------------|
| `plan` | planning, plans-management, sync-status |
| `execute` | work, impl, breezing, parallel-workflows, ci |
| `review` | harness-review, codex-review, verify, troubleshoot |
| `release` | release-har, x-release-harness, handoff |
| `setup` | setup, harness-init, harness-update, maintenance |

## 3-Agent Mapping

| v3 Agent | Consolidated From (Legacy Agents) |
|--------------|------------------|
| `worker` | task-worker, codex-implementer, error-recovery |
| `reviewer` | code-reviewer, plan-critic, plan-analyst |
| `scaffolder` | project-analyzer, project-scaffolder, project-state-updater |

## TypeScript Configuration

- `exactOptionalPropertyTypes: true` — Use conditional assignment for optional fields
- `noUncheckedIndexedAccess: true` — Array access requires undefined checks
- `NodeNext` module resolution — ESM
- `better-sqlite3` is in `optionalDependencies` (Node 24 compat)

## Symlink Structure (v3)

The 5-verb skills in `codex/.codex/skills/` and `opencode/skills/` are symlinks to `skills-v3/`:

```bash
codex/.codex/skills/plan -> ../../../../skills-v3/plan
opencode/skills/execute   -> ../../../skills-v3/execute
# ...etc
```

`check-consistency.sh` validates symlink integrity.
