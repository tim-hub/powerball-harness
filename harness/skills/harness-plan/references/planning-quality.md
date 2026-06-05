# Planning Quality Contract — harness-plan Standard Flow

`harness-plan` does not blindly convert user input into task entries.
For plan creation and significant task additions, it filters input through latest information,
existing specs, memory, and multi-perspective debate — including only well-supported elements
in the Plans.md task contract.

This is not a standalone subcommand. It is the standard quality gate for `create` and
high-impact `add` operations.

## Step 0: Applicability

Apply this quality contract when any of the following conditions are met:

- Creating a new plan with `create`
- Adding a task with `add` that affects product behavior, API, data model, permissions, billing, external integrations, or distribution surfaces
- The user has provided an external product, competitor feature, draft spec, improvement proposal, or comparison material
- The proposal may conflict with existing specs, Plans.md, memory, or past decisions
- The user has requested maximum rigor, thorough comparison, neutral scoring, or regression prevention

The following can be handled lightly without running this contract:

- `update` for marker changes only
- `sync` for status reconciliation only
- Typo, formatting, README, or CHANGELOG-only changes
- Narrow changes where the correct answer is fixed by existing specs and tests

## Step 1: Input Decomposition

Split the user's input into four categories:

| Category | Examples |
|----------|---------|
| Subject of evaluation | External products, competitor features, draft specs, design decisions, operational proposals |
| User's intent | What to improve, what to avoid |
| Uncertain facts | Recency, pricing, APIs, constraints, competitor state, existing repo state |
| Evidence needed for adoption decisions | Official docs, empirical measurements, existing specs, memory, test results |

Do not stop to ask questions when something is unclear. Evaluate based on the most reasonable inferred intent, and surface "decision branches" only when judgment genuinely splits.

## Step 2: Latest-information Fetch

Use WebSearch when external facts are involved. Priority order:

1. Official documentation, official blog posts, release notes, GitHub repositories
2. Standard specifications, papers, primary technical sources
3. Reliable comparison articles, case studies, issue discussions

For critical facts, verify against at least 2 sources when possible.
When sources conflict, document the contradiction and its effect on the adoption decision.

When WebSearch is unavailable or the network fails:

- Record `latest-info: unverified`
- Use local evidence only for a provisional evaluation
- Explicitly note in the final output which claims still need web verification

## Step 3: Local-source-of-truth Check

Proposals for inclusion in the product must be reconciled against existing sources of truth.

Minimum checks:

```bash
cat Plans.md
rg -n "<related-keyword>" README.md CLAUDE.md docs skills scripts tests
find docs -maxdepth 3 -type f | sort
git status --short --branch
```

Review perspectives:

- Does the proposal contradict existing product promises?
- Does it conflict with existing skill roles, triggers, or allowed-tools?
- Does it compete with incomplete tasks in Plans.md?
- Does it affect distribution mirrors, Codex mirror, i18n surfaces?
- If a spec SSOT exists, should it be updated before Plans.md?

## Step 4: Memory Check

When harness-mem, harness-recall, or local agent memory is available, check past decisions using relevant keywords.
Scope searches to the current project unless the user explicitly requests cross-project search.

Sources to check:

- harness-mem / harness-recall search results
- `.claude/agent-memory/`
- `.claude/state/memory-bridge-events.jsonl`
- Presence of `.harness-mem/`
- Prior decisions recorded in repo docs or Plans.md

Notes:

- Do not assume direct access to the harness-mem database
- If harness-mem is unconfigured, unhealthy, or unsearchable: record `memory: unverified`
- Memory is weaker than current repo state. When old memory conflicts with current git/docs, prefer the current repo state

## Step 5: Subagent Debate

When the Task tool is available, run at least 3 independent perspectives. Each agent should be specified as read-only, evidence-backed, and conclusion-first.

Standard roles:

| Role | Focus |
|------|-------|
| Product / Strategy | Adoption value, differentiation, user value, opportunity cost |
| Architecture / Implementation | Feasibility, consistency with existing design, maintenance burden |
| QA / Regression | Regressions, tests, distribution mirrors, compatibility |
| Skeptic | Reasons not to adopt, overinvestment, ambiguous assumptions |

Required fields in each agent output:

- Adopt / Conditional adopt / Reject
- Evidence
- Biggest risk
- Additional verification needed
- Conflicts with existing specs or memory

Synthesizing the debate:

1. Extract points of agreement
2. Preserve points of disagreement
3. State your own judgment
4. Classify into Required / Recommended / Optional / Reject

When subagents are unavailable: evaluate the same 4 perspectives explicitly in sequence as a single agent, and note `subagent-debate: unavailable`.

## Step 6: Neutral Scoring Review

Score on a 5-point scale. 5 = strong, 1 = weak.

| Axis | 5 | 3 | 1 |
|------|---|---|---|
| Product Fit | Core to this product | Useful but peripheral | Another product or workflow would suffice |
| Evidence Strength | Primary sources + measurements + existing evidence | One side verified only | Primarily inference |
| User Value | Substantially improves decision quality or execution speed | Effective in some workflows | Minimal perceived value |
| Implementation Feasibility | Small and localized | Medium-scale but manageable | Large-scale with high maintenance burden |
| Regression Safety | Low-risk and testable | Some blast radius | Likely to break existing flows |
| Strategic Leverage | Becomes long-term differentiation | Stays a convenience feature | One-off value only |

Correction rules:

- If Evidence Strength ≤ 2: Required status is blocked
- If Regression Safety ≤ 2: Add a spike or spec task first
- If Implementation Feasibility ≤ 2 and User Value ≤ 3: lean toward Reject
- If Product Fit ≤ 2: route to docs or external workflow instead of this product

## Step 7: Quality Contract Output

Convert the evaluation into a decision-ready format. Do not output raw scoring — translate into actionable form.

Required structure:

```markdown
In one sentence:
{{adoption decision in one sentence}}

Scoring review:
| Proposal | Score | Verdict | Evidence | Unverified |
|----------|-------|---------|----------|------------|

Proposals to include:
| Priority | Proposal | Reason | What changes |
|----------|----------|--------|--------------|

Regression check:
- Spec:
- Plans.md:
- Memory:
- Mirrors / distribution:
- Tests:

Next steps:
1. ...
2. ...
3. ...
```

Style rules:

- Lead with the conclusion
- Translate technical terms immediately and briefly
- Do not judge by feel ("amazing", "innovative")
- Limit proposals to 1–3; do not present too many candidates
- Distinguish facts, inferences, and unverified claims

## Step 8: Plans.md / Spec Output

Convert only the adopted proposals into task contracts.

Order:

1. If a spec SSOT is needed, create or update the project spec first
2. Add only Required tasks to Plans.md
3. Tag high-risk proposals with `[needs-spike]`
4. Place a verifiable DoD in each task
5. Tag tasks that require TDD with `[tdd:required]`
6. If mirrors, i18n, or package surfaces are affected, add separate verification tasks

Prohibited:

- Creating implementation tasks when the spec's correct conditions are still undefined
- Substituting a "note" for a regression-check task
