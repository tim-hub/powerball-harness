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

## Future Considerations

(none currently)

---
