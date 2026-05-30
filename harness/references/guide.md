# Harness Guide

Orientation and workflow guidance for the Harness development cycle.

## Orientation Patterns

Check current state and respond with context-aware guidance:

### No Project Exists

> Let's start a project first! Tell me what you want to build — a rough idea is fine.
> Example: "I want to build a blog" → run `/harness-setup` then `/harness-plan create`

### Plans.md Exists, No In-Progress Tasks

> There's a plan. Let's start working!
> Say: "Start phase 1", "Do the first task", or "Do everything" → `/harness-work`

### Task In Progress

> Work in progress — current task: `{{task name}}` ({{completed}}/{{total}})
> Say: "Continue", "Next task", or "How far along are we?"

### After Phase Completion

> Phase complete! Options: "Check it works" (dev server), "Review it" (`/harness-review`), "Next phase" (`/harness-work`), "Commit it"

### When an Error Occurs

> A problem occurred: `{{error summary}}`
> Say: "Fix it" (auto-fix), "Explain it" (detail), or "Skip it" (next task)

**Context check**: Existence of AGENTS.md (project init), Plans.md (plan + progress), `cc:WIP` marker (in-progress task), recent errors.

---

## Workflow Overview

The Harness lifecycle: **Setup → Plan → Work → Review → Release**

```
/harness-setup    → creates CLAUDE.md, Plans.md
/harness-plan     → adds tasks with acceptance criteria
/harness-work     → implements (auto-selects solo/parallel/breezing)
/harness-review   → code review gate
/harness-release  → version bump, CHANGELOG, tag, GitHub Release
```

### Primary Lifecycle

```
User / New Project
    ↓
① Setup       /harness-setup init
    ↓
② Planning    /harness-plan create|add
    ↓
③ Execution   /harness-work all  (auto-selects mode)
    ↓
④ Review      /harness-review
    ↓
⑤ Release     /harness-release patch|minor|major
```

### Execution Mode Decision

`/harness-work` auto-selects based on task count (override with flags):

| Flag / Count | Mode | Description |
|---|---|---|
| 1 task or `--solo` | **Solo** | Single Worker → Reviewer cycle |
| 2–3 tasks or `--parallel N` | **Parallel** | Workers A+B in parallel → Reviewer |
| 4+ tasks or `--breezing` | **Breezing** | Lead → Workers (worktrees) → Reviewer |
| `--codex` | **Codex** | Delegates to Codex CLI via `scripts/codex-companion.sh` |

### Agent Roles

| Agent | Role | Tools |
|-------|------|-------|
| **Worker** | TDD, implementation, self-check, git commit | Read / Write / Edit / Bash |
| **Reviewer** | Independent verdict against sprint-contract | Read / Grep / Glob (read-only) |
| **Scaffolder** | Project init, tech-stack detection, CLAUDE.md/Plans.md | Read / Write / Edit |
| **Advisor** | Consults on high-risk preflight or repeated failures | Read-only, returns PLAN/CORRECTION/STOP |
| **CI-CD-Fixer** | CI failure diagnosis and fix (3-strike escalation) | Read / Write / Edit / Bash |

### Breezing Fix Loop

```
Lead spawns Worker (worktree, task + sprint-contract)
Worker: TDD → implement → self-check → commit
Lead reviews with Reviewer (diff + sprint-contract)
  APPROVE  → cherry-pick to main → plans.json cc:done
  REQUEST_CHANGES (≤3 retries) → SendMessage fixes → re-review
```

### Key Skills

| Skill | Purpose |
|-------|---------|
| `harness-setup` | Project initialization |
| `harness-plan` | Create/update Plans.md tasks |
| `harness-work` | Implement tasks (all modes) |
| `harness-review` | Multi-angle code review |
| `harness-release` | Version bump + CHANGELOG + GitHub Release |
| `harness-remember` | SSOT — decisions.md, patterns.md, session log |

Full catalog: `docs/CLAUDE-skill-catalog.md`

---

## Plain-Language Phrase Guide

For users who prefer natural language over slash commands.

### Common Phrase Reference

| What You Want to Do | How to Say It |
|--------------------|---------------|
| Start a project | "I want to build XX" |
| View the plan | "Show me the plan", "What's the status?" |
| Start working | "Start", "Build it", "Do phase 1" |
| Continue | "Continue", "Next" |
| Test it | "Run it", "Show me" |
| Review code | "Review it", "Check it" |
| Save | "Commit it", "Save it" |
| When stuck | "What should I do?", "Help" |
| Leave it all to me | "Do everything", "You handle it" |

### Principles

- Avoid technical jargon — explain in plain, simple language
- Analyze current state before responding (Plans.md, AGENTS.md, cc:WIP, recent errors)
- Present specific "example phrases" the user can say next
- Wait for the user's next action rather than proceeding automatically
