# Feature Candidates

Ideas explored but not yet planned. Kept here for future reference only.

---

## Compound Engineering Feedback Loop

**Source**: [every.to/guides/compound-engineering](https://every.to/guides/compound-engineering)

**Core idea**: After each plan→work→review cycle, capture reusable insights back into the system so future cycles improve automatically. The article frames it as a fourth step: *capture → make findable → update system → verify learning*.

**Where the harness is already strong (~70%):**
- SSOT memory (`decisions.md`, `patterns.md`, `session-log.md`) — knowledge extraction exists
- Worker/Reviewer/Scaffolder agents have full environment access
- Plans-driven flow with hook enforcement

**The gap (~30%):** No structured path from *Reviewer observation* → *system improvement*. Recurring review comments don't automatically harden into guardrails or skill constraints.

**Ideas explored:**

1. Post-review pattern synthesis — classify recurring Reviewer comments, propose them as new deny rules or guardrail entries
2. A `compound-rules.md` file under `.claude/rules/` updated after each cycle

**Why shelved**: The routing problem is non-trivial. Insights belong in different layers depending on their nature:

| Insight type | Right target |
|---|---|
| "Don't do X" — enforceable | `settings.json` deny or Go guardrail rule |
| "Prefer Y pattern" | `patterns.md` |
| "We decided Z because..." | `decisions.md` |
| "Next time, check W first" | relevant `SKILL.md` |

A single file becomes a dumping ground. Automating the routing adds complexity without clear ROI at current scale. Soft markdown rules also don't satisfy the article's test: *"Would the system catch this automatically next time?"* — only enforced rules (Go engine, deny rules) truly do.

**Revisit when**: The Reviewer is catching the same issue 3+ times per week across sessions.
