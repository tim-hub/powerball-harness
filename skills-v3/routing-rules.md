# Skill Routing Rules — v3 (Reference)

Reference for routing rules between Harness v3's 5-verb skills.

> **SSOT location**: Each skill's `description` field is the SSOT for routing.
> This file provides detailed explanations and examples as a reference; actual routing depends on each skill's description.

## 5-Verb Skill Routing Table

| Skill | Trigger Keywords | Exclusions |
|--------|-----------------|------|
| `harness-plan` | create a plan, add tasks, update Plans.md, mark complete, check progress, sync status, where am I, harness-plan, harness-sync | implementation, code review, release |
| `harness-work` | implement, execute, harness-work, do everything, build features, run tasks, breezing, team run, --codex, --parallel | planning, code review, release, setup |
| `harness-review` | review, code review, plan review, scope analysis, security, performance, quality checks, PRs, diffs, harness-review | implementation, new features, bug fixes, setup, release |
| `harness-release` | release, version bump, create tag, publish, /harness-release | implementation, code review, planning, setup |
| `harness-setup` | setup, initialization, new project, CI setup, codex CLI setup, harness-mem, agent setup, symlinks, /harness-setup | implementation, code review, release, planning |

## Detailed Routing

### harness-plan Skill

**Triggers** (match any):
- "create a plan"
- "add a task"
- "update Plans.md"
- "mark complete" / "mark as done"
- "where am I" / "check progress"
- "harness-plan" / "harness-sync"
- "sync status" / "sync Plans.md"

**Exclusions** (exclude if any match):
- "implement"
- "code review"
- "release"

### harness-work Skill

**Triggers** (match any):
- "implement"
- "execute"
- "harness-work"
- "do everything"
- "just this"
- "breezing" / "team execution"
- "--codex" / "--parallel"
- "build"

**Exclusions** (exclude if any match):
- "plan" (without implementation)
- "review" (without implementation)
- "release"
- "setup"

### harness-review Skill

**Triggers** (match any):
- "review"
- "code review"
- "plan review"
- "check scope"
- "security check"
- "quality check"
- "PR review"
- "harness-review"
- "check diff" / "review changes"

**Exclusions** (exclude if any match):
- "implement" (implementation request)
- "add new feature"
- "fix bug"
- "setup"
- "release"

### harness-release Skill

**Triggers** (match any):
- "release"
- "version bump"
- "create tag"
- "publish"
- "update CHANGELOG"
- "harness-release"

**Exclusions** (exclude if any match):
- "implement"
- "code review"
- "plan"
- "setup"

### harness-setup Skill

**Triggers** (match any):
- "setup"
- "initialization" / "init"
- "new project"
- "CI setup"
- "Codex CLI setup"
- "harness-mem"
- "agent configuration"
- "symlink update"
- "harness-setup"

**Exclusions** (exclude if any match):
- "implement"
- "code review"
- "release"
- "create a plan"

## Priority Rules

1. **Exclusions take highest priority**: Skills matching an exclusion keyword are never loaded
2. **Specific keywords take priority**: Exact match > partial match
3. **When ambiguous**: `plan` > `execute` > `review` (choose the more conservative option)

## Extension Pack (extensions/)

Non-core features are stored in `skills-v3/extensions/`:

| Skill | Purpose |
|--------|------|
| `auth` | Authentication and payment features (Clerk, Stripe) |
| `crud` | CRUD auto-generation |
| `ui` | UI component generation |
| `agent-browser` | Browser automation |
| `gogcli-ops` | Google Workspace operations |
| `codex-review` | Codex second opinion |
| `notebookLM` | NotebookLM integration |
| `generate-slide` | Slide generation |
| `deploy` | Deployment automation |
| `memory` | SSOT and memory management |
| `cc-cursor-cc` | Cursor <-> Claude Code integration |

## Update Rules

1. **description = SSOT**: Each skill's `description` field is the official routing definition
2. **This file's role**: Detailed explanations and decision flow reference (not the SSOT)
3. **Maintain complete list**: Use specific keywords, not generic expressions
