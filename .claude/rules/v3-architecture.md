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
├── skills/         # Named skills (harness-work, harness-plan, breezing, etc.)
├── codex/          # Codex CLI distribution (symlinked skills)
│   └── .codex/
│       ├── config.toml       # Multi-agent config
│       ├── rules/            # Codex guardrail rules
│       └── skills/           # Symlinks → ../../../skills/
├── agents/         # 6 agents (3 core + 3 specialized)
│   ├── worker.md             # Implementation agent
│   ├── reviewer.md           # Review agent (Read-only)
│   ├── scaffolder.md         # Scaffolding and state update agent
│   ├── team-composition.md   # Team composition guide
│   ├── ci-cd-fixer.md        # CI/CD failure recovery
│   └── video-scene-generator.md # Remotion video generation
├── hooks/          # Thin shims (→ delegates to core/src/index.ts)
├── scripts/        # Hook handlers, session management, Codex companion
└── .claude/
    └── agent-memory/
        ├── claude-code-harness-worker/
        ├── claude-code-harness-reviewer/
        └── claude-code-harness-scaffolder/
```

## Agent Consolidation History

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

## Codex Symlink Structure

Skills in `codex/.codex/skills/` are symlinks to `../../../skills/`:

```bash
codex/.codex/skills/harness-work -> ../../../skills/harness-work
codex/.codex/skills/harness-plan -> ../../../skills/harness-plan
codex/.codex/skills/breezing     -> ../../../skills/breezing
# ...etc
```

`check-consistency.sh` validates symlink integrity.
