# Benchmark Rubric

Last updated: 2026-03-06

This document is a rerunnable rubric for comparing `claude-code-harness` with other tools.
Rather than relying on README impressions, it separates scoring into static evidence and executed evidence.

## Evidence Classes

| Class | Examples | When to Use |
|------|----|-----------|
| Static evidence | README, repo tree, hooks definitions, tests, docs, package metadata | Comparing mechanism existence, design clarity, distribution paths |
| Executed evidence | test run, smoke run, benchmark logs, evidence pack, CI artifact | Comparing whether claims are reproducible, whether guardrails actually work |

## Scoring Axes

| Axis | Weight | What to inspect |
|------|--------|-----------------|
| Runtime enforcement | 25 | Hooks, guardrails, deny/warn behavior, lifecycle automation |
| Verification and test credibility | 25 | Unit/integration tests, consistency checks, evidence pack, CI coverage |
| Onboarding and operator clarity | 20 | install flow, docs completeness, claim consistency, quickstart quality |
| Scope discipline and maintainability | 15 | distribution boundary, compatibility story, residue management |
| Positioning and adoption proof | 15 | public narrative, stars/users, reproducible showcase, differentiation |

Total: 100 points

## Review Flow

1. Gather static evidence
2. List claims that require executed evidence
3. Separate verified claims from unverified/pending claims
4. Score each axis, noting the evidence type
5. Write strengths and weaknesses separately, e.g., "design is strong but unproven," "market is strong but runtime enforcement is thin"

## Required Output Format

Comparison reports must include at minimum:

- Comparison date
- Target repositories / versions / commit or default branch snapshot
- List of commands executed
- Distinction between static evidence and executed evidence
- Score per axis
- Items that could not be fully reproduced

## Reusable Template

```md
# Benchmark Report

- Compared at:
- Repositories / versions:
- Commands executed:

## Static evidence

- Repo structure:
- Docs and claims:
- Guardrails / hooks / tests:

## Executed evidence

- Validation commands:
- Benchmark or smoke runs:
- Evidence artifacts:

## Scores

| Axis | Score | Evidence type | Notes |
|------|-------|---------------|-------|
| Runtime enforcement |  | Static / Executed |  |
| Verification and test credibility |  | Static / Executed |  |
| Onboarding and operator clarity |  | Static / Executed |  |
| Scope discipline and maintainability |  | Static / Executed |  |
| Positioning and adoption proof |  | Static / Executed |  |

## Unverified or blocked items

- None

## Harness-specific Notes

- Strong claims like `/harness-work all` should only receive high scores after execution evidence from `docs/evidence/work-all.md` is available
- Residual artifacts like `commands/` or `mcp-server/` are not penalized for existence; **only penalize when their explanation is ambiguous**
- If README claims do not align with tests/CI/distribution boundaries, lower the `Onboarding and operator clarity` score
