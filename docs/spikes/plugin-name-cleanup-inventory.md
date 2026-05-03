# Plugin Name Reference Cleanup — Inventory (Phase 90.1)

**Date**: 2026-05-03
**Canonical**: plugin name `harness` from `.claude-plugin/marketplace.json`. All subagent refs should use `harness:<agent>`. All slash commands should use `/harness:<skill>`.

**Excluded**: Go source (backward-compat parser), `.claude/memory/archive/`, `.claude/plans/`, `harness/skills/harness-release/scripts/check-residue.py`, CHANGELOG ≤ v5.0.3.

---

## Group (a): Subagent prefixes in `harness/agents/` + `harness/skills/`

```
harness/agents/team-composition.md:16:  +-- Worker (harness:worker)  ──consults──>  Advisor (harness:advisor)
harness/agents/team-composition.md:23:  +-- [Worker #2] (harness:worker)
harness/agents/team-composition.md:26:  +-- Reviewer (harness:reviewer)
harness/agents/team-composition.md:63:| **subagent_type** | `harness:worker` |
harness/agents/team-composition.md:75:| **subagent_type** | `harness:ralph-worker` |
harness/agents/team-composition.md:83:| **subagent_type** | `harness:advisor` |
harness/agents/team-composition.md:95:| **subagent_type** | `harness:reviewer` |
harness/agents/team-composition.md:119:| **subagent_type** | `harness:scaffolder` |
harness/agents/worker.md:189:2. If enabled, invoke `harness:advisor` subagent with: `task_id`, `reason_code`, normalized `error_signature`, `retry_count`, and `taxonomy_id` if available from hook output
harness/skills/harness-work/references/solo-mode.md:36:   - If task has `<!-- advisor:required -->` marker: consult `harness:advisor` with `reason_code: high_risk_preflight`
harness/skills/harness-work/references/breezing-mode.md:88:        subagent_type="harness:worker",
harness/skills/breezing/SKILL.md:64:| Worker xN | `harness:worker` | `bypassPermissions` (current) / Auto Mode (follow-up)* | Implementation |
harness/skills/breezing/SKILL.md:65:| Reviewer | `harness:reviewer` | `bypassPermissions` (current) / Auto Mode (follow-up)* | Independent review |
harness/skills/harness-ralph-loop/SKILL.md:14:Spawns a fresh `harness:ralph-worker` subagent per attempt inside a single persistent worktree,
harness/skills/harness-ralph-loop/SKILL.md:68:- `harness:ralph-worker` — The per-iteration worker agent this skill dispatches
harness/skills/harness-ralph-loop/references/loop-flow.md:45:            subagent_type="harness:ralph-worker",
harness/skills/harness-ralph-loop/references/loop-flow.md:58:            subagent_type="harness:ralph-worker",
harness/skills/harness-schedule-run/SKILL.md:92:      subagent_type="harness:worker",
harness/skills/harness-schedule-run/references/flow.md:151:Spawn `harness:worker` via the Agent tool:
harness/skills/harness-schedule-run/references/flow.md:153:> **Important**: Specify `"harness:worker"` for `subagent_type`, NOT `"harness-work"`.
harness/skills/harness-schedule-run/references/flow.md:159:    subagent_type="harness:worker",  # worker agent (not a skill)
```

---

## Group (b): Subagent prefixes in `harness/templates/`

```
harness/templates/codex-skills/harness-work/SKILL.md:146:   - If task has `<!-- advisor:required -->` marker: consult `harness:advisor` with `reason_code: high_risk_preflight`
harness/templates/codex-skills/harness-work/SKILL.md:239:        subagent_type="harness:worker",
harness/templates/codex-skills/breezing/SKILL.md:63:| Worker xN | `harness:worker` | `bypassPermissions` (current) / Auto Mode (follow-up)* | Implementation |
harness/templates/codex-skills/breezing/SKILL.md:64:| Reviewer | `harness:reviewer` | `bypassPermissions` (current) / Auto Mode (follow-up)* | Independent review |
harness/templates/codex-skills/harness-schedule-run/references/flow.md:151:Spawn `harness:worker` via the Agent tool:
harness/templates/codex-skills/harness-schedule-run/references/flow.md:153:> **Important**: Specify `"harness:worker"` for `subagent_type`, NOT `"harness-work"`.
harness/templates/codex-skills/harness-schedule-run/references/flow.md:159:    subagent_type="harness:worker",  # worker agent (not a skill)
harness/templates/codex-skills/harness-schedule-run/SKILL.md:87:      subagent_type="harness:worker",
harness/templates/opencode/skills/harness-schedule-run/SKILL.md:87:      subagent_type="harness:worker",
harness/templates/opencode/skills/harness-schedule-run/references/flow.md:151:Spawn `harness:worker` via the Agent tool:
harness/templates/opencode/skills/harness-schedule-run/references/flow.md:153:> **Important**: Specify `"harness:worker"` for `subagent_type`, NOT `"harness-work"`.
harness/templates/opencode/skills/harness-schedule-run/references/flow.md:159:    subagent_type="harness:worker",  # worker agent (not a skill)
```

---

## Group (c): Slash-command invocations in Cursor/Opencode templates

```
harness/templates/cursor/commands/review-cc-work.md:61:/claude-code-harness:core:work
harness/templates/cursor/commands/review-cc-work.md:86:/claude-code-harness:core:work
harness/templates/cursor/commands/review-cc-work.md:122:/claude-code-harness:core:work
harness/templates/cursor/commands/handoff-to-claude.md:19:/claude-code-harness:core:work
harness/templates/opencode/commands/review-cc-work.md:61:/claude-code-harness:core:work
harness/templates/opencode/commands/review-cc-work.md:86:/claude-code-harness:core:work
harness/templates/opencode/commands/review-cc-work.md:122:/claude-code-harness:core:work
harness/templates/opencode/commands/handoff-to-claude.md:19:/claude-code-harness:core:work
```

---

## Group (d): Plans.md / CHANGELOG / docs self-references

```
Plans.md:23:| 90.5 | Fix Group (d): Plans.md (Phase 89 task descriptions reference `harness:ralph-worker`), CHANGELOG.md `[Unreleased]` (mentions `harness:ralph-worker` from my v5.0.3-era edit that was wrong), `docs/advisor-strategy.md`, `docs/spikes/ralph-{worktree-persistence,smoke-test-happy-path}.md`. | All 5 files updated; grep clean across the repo (excluding intentional Go + frozen artifacts) | 90.2, 90.3, 90.4 | cc:TODO |
Plans.md:24:| 90.6 | Add regression-prevention validation rule. Extend `tests/validate-plugin.sh` (or add a check to `.claude/skills/release-this/scripts/check-consistency.sh`) that flags any new occurrence of `claude-code-harness:<agent>` or `powerball-harness:<agent>` outside the allowlist (Go source, `.claude/memory/archive/`, `.claude/plans/`, `harness/skills/harness-release/scripts/check-residue.py`, CHANGELOG entries for v5.0.3 and earlier). Synthetic-insertion test: temporarily insert `harness:worker` into a test file → check fails; remove → check passes. | New check is wired into validate-plugin.sh or check-consistency.sh; on a clean tree it passes; with a synthetic violation it fails non-zero with a clear message | 90.2, 90.3, 90.4 | cc:TODO |
Plans.md:33:**Goal**: Add a Ralph-Wiggum-style iterative-loop execution mode to harness. `harness-plan` learns to detect Ralph-suitable tasks (well-defined success criteria, "iterate-until-pass" patterns) and emits them with a new `[ralph]` marker plus per-task `Verify:` and `MaxIter:` lines. A new `harness-ralph-loop` skill orchestrates the loop: it owns one persistent worktree, dispatches a fresh `harness:ralph-worker` subagent per iteration, runs the verify command authoritatively (anti-tamper hard stop on worker self-report mismatch), scans for both a `<promise>{DoD}</promise>` text tag and a structured `ralph-worker-report.v1` field, diffs the worktree to detect stuck iterations, and stops on success / `FT-RALPH-01` / `FT-RALPH-02` / `FT-RALPH-03`. `harness-work` delegates `[ralph]` tasks to the new skill.
Plans.md:37:**Source**: User request (this session) + opus-agent architectural review at `/Users/hbai/.claude/plans/help-me-investigate-serialized-curry.md` (full design + decisions table). Opus review corrected three premises: (1) `isolation="worktree"` is fresh-per-call so the orchestrator must drive worktree reuse via `EnterWorktree(path=...)`; (2) the existing `harness:worker` agent cannot be reused (hard-coded `isolation: worktree` in frontmatter, TDD-shaped initialPrompt, SR-5 conflicts with Ralph's scratchpad-file pattern) — a sibling `ralph-worker` agent is required; (3) verifier-mismatch (worker self-reports verify pass when actual verify fails) is a load-bearing hard stop, not just a warning.
Plans.md:384:**Agent names**: NOT changing. `harness:worker`, `harness:reviewer`, `harness:advisor` are part of the distributed plugin and work correctly as-is.
CHANGELOG.md:57:**After**: A new `harness-ralph-loop` skill orchestrates the loop in subagent space. Iteration 0 spawns a fresh `harness:ralph-worker` with `isolation="worktree"` to create a persistent worktree; iterations 1..N use `EnterWorktree(path=...)` to re-enter the same worktree so each spawn sees prior attempts on disk (the "self-referential through files" Ralph pattern). After each iteration, the orchestrator runs the verify command authoritatively (anti-tampering against worker self-reports), scans the final assistant message for `<promise>{DoD}</promise>`, and consults the structured `ralph-worker-report.v1` JSON before deciding whether to continue, succeed, or hard-stop.
CHANGELOG.md:74:#### 3. `harness:ralph-worker` agent + `ralph-worker-report.v1` schema
CHANGELOG.md:78:**After**: A sibling `harness:ralph-worker` agent is added with no `isolation: worktree` in frontmatter (orchestrator owns the worktree), a Ralph-specific initialPrompt (read prior attempts → implement → run verify → emit `<promise>` only if exit 0), and a new `ralph-worker-report.v1` schema with `iteration`, `verify {command, exit_code, stderr_tail}`, `promise {asserted, dod}`, `files_changed`, `summary`, plus three SR-RALPH-* rules.
CHANGELOG.md:2850:**Before**: Worker and Reviewer agent type names varied across files. `breezing/SKILL.md` used `general-purpose` while `team-composition.md` used `harness:worker`, preventing per-agent hooks (agent type-specific guardrails) from firing correctly.
CHANGELOG.md:2852:**After**: Unified to `harness:worker` / `harness:reviewer` across all files. Worker-specific PreToolUse guards (Write/Edit checks) and Reviewer-specific Stop logs (completion records) now apply reliably.
docs/advisor-strategy.md:56:  +-- Worker (harness:worker)  ──consults──>  Advisor (harness:advisor)
docs/advisor-strategy.md:61:  +-- Reviewer (harness:reviewer)
docs/spikes/ralph-smoke-test-happy-path.md:125:result = Agent(subagent_type="harness:ralph-worker", isolation="worktree")
docs/spikes/ralph-worktree-persistence.md:66:            subagent_type="harness:ralph-worker",
docs/spikes/ralph-worktree-persistence.md:76:            subagent_type="harness:ralph-worker",
```

---

## Mapping: wrong → canonical

| Wrong | Canonical |
|-------|-----------|
| `harness:worker` | `harness:worker` |
| `harness:worker` | `harness:worker` |
| `harness:reviewer` | `harness:reviewer` |
| `harness:reviewer` | `harness:reviewer` |
| `harness:advisor` | `harness:advisor` |
| `harness:advisor` | `harness:advisor` |
| `harness:scaffolder` | `harness:scaffolder` |
| `harness:scaffolder` | `harness:scaffolder` |
| `harness:ralph-worker` | `harness:ralph-worker` |
| `harness:ralph-worker` | `harness:ralph-worker` |
| `/claude-code-harness:core:work` | `/harness:harness-work` |
```
