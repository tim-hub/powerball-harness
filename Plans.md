# Powerball Harness — Plans.md

Last archive: 2026-04-18 (Phase 62–73 → `.claude/memory/archive/Plans-2026-04-18-phase62-73.md`)
Last release: v4.9.1 on 2026-04-18 (fix platform binary version embedding)

---

## Phase 74: Code-space skill search — proof-of-concept on `harness-review`

Created: 2026-04-17

**Goal**: Borrow Meta-Harness's code-space search idea (arxiv 2603.28052) — instead of hand-editing `SKILL.md`, generate variants and score them against an evaluation suite. Scope: one skill (`harness-review`), 3-5 variants, measurable outcome. If the POC improves the skill, formalize the loop; if not, record why and discard.

**Depends on**: Phase 72 (traces provide failure signal for the proposer) + Phase 73 (advisor pattern reusable for proposer scaffolding).

**Score function** (decided 2026-04-18): **golden verdict / pass-fail** — each test case has an expected APPROVE or REQUEST_CHANGES; score = correct / total. Binary, automatable, CI-friendly.

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 74.1 | Define score function + build eval runner `local-scripts/eval-skill.sh <skill-dir> <eval-suite-dir>`. Runs fixed inputs through the skill, captures outputs, emits JSON score report | `eval-skill.sh harness/skills/harness-review tests/skill-eval/harness-review` prints a reproducible score | 72.1 | cc:Done [8ed646c] |
| 74.2 | Build evaluation suite for `harness-review` at `tests/skill-eval/harness-review/` — 5 PR diffs (2 with real bugs, 2 clean, 1 scope-creep) + expected verdicts | `ls tests/skill-eval/harness-review/*.{diff,expected.json}` shows 10 files; running baseline skill against suite produces a score | 74.1 | cc:Done [cbbe5af] |
| 74.3 | Write proposer script `local-scripts/propose-skill-variants.sh <skill-dir>`. Given SKILL.md + eval output + recent traces, generates 3 SKILL.md variants to `/tmp/skill-variants/harness-review-v{1,2,3}/SKILL.md` via Claude subagent | Running against a broken baseline (intentionally degraded `harness-review`) produces 3 syntactically valid SKILL.md files with meaningful diffs | 74.1 | cc:Done [9ef0a1e] |
| 74.4 | Run end-to-end search loop: baseline → generate 3 variants → score each → pick winner. Emit report at `.claude/state/code-search/harness-review-<YYYY-MM-DD>.md` with scores, diffs, chosen winner | Report exists; winner's score ≥ baseline's score; report includes rationale | 74.2, 74.3 | cc:Done [local] |
| 74.5 | Decision gate: if winner beats baseline by ≥10%, promote to main and add pattern to `patterns.md`. Otherwise document the null result. Update CHANGELOG [Unreleased] | Either a commit promoting the variant OR a patterns.md entry "code-space search POC attempted; no gain — reasons X, Y" | 74.4 | cc:Done [747f999] |

---

## Phase 75: Port upstream Worker contract improvements (PR #88, #89)

Created: 2026-04-19

**Source**: Chachamaru127/claude-code-harness PRs #88 (v4.2.0, Hokage) / #89 (v4.3.0, Arcana) / #90 (release marker, no code).

**Goal**: Bring in the portable behavioral improvements from the upstream fork without rebasing — their v4.2/v4.3 diverged significantly from our line (we're at v4.9.3 with a different Go-native architecture). We cherry-pick what's a clean, low-coupling improvement and skip what our infra already covers differently.

**Investigation summary**:
- **PR #88** is a CC 2.1.99-110 + Opus 4.7 integration release. Most of it is CC-version-specific (guardrail rule refresh, agent frontmatter migration, plugin schema migration to `plugins-reference`) — we're already past it structurally. Two portable items: `scripts/enable-1h-cache.sh` (opt-in 1-hour prompt cache) and a role-based PreCompact decision (our `pre-compact-save.js` only warns; upstream's `pre_compact.go` blocks compaction for Worker sessions).
- **PR #89** is the high-value one: four independent Worker/Reviewer contract tightenings (#84–#87). All directly applicable.
- **PR #90** is a release-completion marker with no code changes — skip.

**What we already have and are NOT porting**:
- PreCompact hook: we have `pre-compact-save.js` warning on `cc:WIP` tasks. Upstream's role-based blocking is a nice-to-have but not urgent — deferred to 75.6 as optional.
- Guardrail engine: our Go-native `bin/harness` already covers bypass detection.
- Plugin schema: our v4.9 architecture is already past upstream's `plugins-reference` migration.

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 75.1 | Add Worker NG rules to `harness-work` SKILL.md (upstream issue #85). NG-1: Plans.md `cc:*` markers are Lead-owned — Workers must not modify. NG-2: embedded git repositories inside the worktree are rejected. NG-3: nested teammate spawning is forbidden (reinforce the platform-level block with an explicit contract clause) | `harness-work/SKILL.md` has a "Worker NG Rules" section listing all three; each rule has a one-line rationale and a check point (manual or scripted). Optional: pre-delegate preflight script | - | cc:Done [6030060] |
| 75.2 | Add `REVIEW_AUTOSTART` handshake to `harness-review` SKILL.md Step 0 (upstream issue #84). Purpose: fix the fork-mode stall where `/harness-review` with no args waits forever for input. First 3 lines of Step 0 emit/detect `REVIEW_AUTOSTART` marker; add a prohibition list for 5 fork-context failure modes | `harness-review/SKILL.md` Step 0 top-matter includes the handshake marker and the prohibition list; a smoke test (`bash` or manual) confirms no stall when invoked in fork context with no args | - | cc:Done [390a473] |
| 75.3 | Add Worker self-review gate (`worker-report.v1` JSON schema) to `harness-work` SKILL.md (upstream issue #86). Before the Lead spawns a Reviewer, the Worker must emit a `worker-report.v1` JSON with 5 self-review rule verdicts + evidence fields. Lead rejects incomplete submissions; up to 2 amendment cycles before escalation | `worker-report.v1` schema documented in `harness-work/SKILL.md` or `references/`; Lead flow updated to validate and either accept or reject with amendment loop; schema has `rule_id`, `verified: bool`, `evidence: string` per entry | 75.1 | cc:Done [37b87cf] |
| 75.4 | Add universal violations session injection to `breezing` mode (upstream issue #87). Reviewer memory updates gain a `scope: universal \| task-specific` field; universal-scope items are collected and prepended to the next Worker's briefing within the same breezing session. Prevents repeat violations across tasks | Reviewer memory schema has `scope` field; breezing Phase B prepends universal items to Worker spawn prompt; documented in `harness-work/SKILL.md` breezing section | 75.3 | cc:Done [9d08753] |
| 75.5 | Port `scripts/enable-1h-cache.sh` (upstream PR #88). Opt-in shell script that exports Anthropic's 1-hour prompt cache setting for sessions expected to exceed 30 minutes. Uses `export` format so subprocesses inherit the setting | `harness/scripts/enable-1h-cache.sh` exists and is executable; `source`-able or `eval`-able from a parent shell; a short usage note in `harness-work/SKILL.md` or `docs/` explains when to use it | - | cc:Done [ae36e86] |
| 75.6 | (Optional, lower priority) Upgrade `pre-compact-save.js` from warn-only to role-based blocking (upstream PR #88 `pre_compact.go`). Worker sessions with `cc:WIP` tasks in Plans.md block compaction; Reviewer / Advisor roles pass through. Requires session-role detection in the hook | `pre-compact-save.js` detects session role from session-state and blocks compaction for Worker+WIP combo; test covers all 4 role × state permutations | - | cc:Done [84b0429] |

---

## Future Considerations

(none currently)

---
