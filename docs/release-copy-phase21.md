# Phase 21 Release Copy Drafts

Last updated: 2026-03-06

These are drafts for external announcements of Phase 21.
Use one topic at a time without mixing `trust repair`, `evidence pack`, and `positioning refresh`.

## Draft 1: Trust Repair

We cleaned up the public-facing surfaces of `claude-code-harness`. README badges, missing docs, and distribution boundary descriptions have been aligned, reducing self-contradictions across README / Plans / docs.

## Draft 2: Evidence Pack

We added success / failure fixtures and a smoke runner for `/harness-work all`. Instead of just claims, we now provide a path for re-verification while examining artifacts.

## Draft 3: Positioning Refresh

We refocused the Harness core message on `5 verb skills + TypeScript guardrail engine`. The emphasis is now on being able to run runtime enforcement and verification as a unified system, not just adding skill packs.

## Current Recommendation

- Evidence artifacts can continue to be collected via replay fallback even when hitting quota limits
- Until full success artifacts are ready, Draft 2 should express that "the rerunnable framework has been established"
- Avoid strong assertions like `production-ready`
- When including competitor comparisons, align vocabulary with `docs/positioning-notes.md`
