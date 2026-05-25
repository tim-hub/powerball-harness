# Harness Workflow Reference

Full lifecycle, execution modes, agent roles, and process flow.

## Primary Lifecycle

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

## Execution Mode Decision

`/harness-work` auto-selects based on task count (override with flags):

| Flag / Count | Mode | Description |
|---|---|---|
| 1 task or `--solo` | **Solo** | Single Worker → Reviewer cycle |
| 2–3 tasks or `--parallel N` | **Parallel** | Workers A+B in parallel → Reviewer |
| 4+ tasks or `--breezing` | **Breezing** | Lead → Workers (worktrees) → Reviewer |
| `--codex` | **Codex** | Delegates to Codex CLI via `scripts/codex-companion.sh` |

## Agent Roles

| Agent | Role | Tools |
|-------|------|-------|
| **Worker** | TDD, implementation, self-check, git commit | Read / Write / Edit / Bash |
| **Reviewer** | Independent verdict against sprint-contract | Read / Grep / Glob (read-only) |
| **Scaffolder** | Project init, tech-stack detection, CLAUDE.md/Plans.md | Read / Write / Edit |
| **Advisor** | Consults on high-risk preflight or repeated failures | Read-only, returns PLAN/CORRECTION/STOP |
| **CI-CD-Fixer** | CI failure diagnosis and fix (3-strike escalation) | Read / Write / Edit / Bash |

## Breezing Fix Loop

```
Lead spawns Worker (worktree, task + sprint-contract)
Worker: TDD → implement → self-check → commit
Lead reviews with Reviewer (diff + sprint-contract)
  APPROVE  → cherry-pick to main → Plans.md cc:done
  REQUEST_CHANGES (≤3 retries) → SendMessage fixes → re-review
```

## Key Skills

| Skill | Purpose |
|-------|---------|
| `harness-setup` | Project initialization |
| `harness-plan` | Create/update Plans.md tasks |
| `harness-work` | Implement tasks (all modes) |
| `harness-review` | Multi-angle code review |
| `harness-release` | Version bump + CHANGELOG + GitHub Release |
| `harness-remember` | SSOT — decisions.md, patterns.md, session log |
| `harness-schedule-run` | Autonomous scheduled run (overnight cadence) |

Full catalog: `docs/CLAUDE-skill-catalog.md`
