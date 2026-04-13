# An Honest Look at Where Harness Stands Today

## Article Title Candidates

1. Claude is latest-ready, Codex has come far. An honest look at where Harness stands
2. AI coding operations are stronger without exaggeration. Harness release status and latest support
3. Claude is latest-ready. Harness honestly explains how far it's tracked Codex
4. We won't say "perfect." Organizing Harness release status and latest version support

## Chosen Title

Claude is latest-ready, Codex has come far. An honest look at where Harness stands

## Introduction

When communicating about AI coding tools, there's a tendency to make strong claims like "now supporting the latest," "fully compatible," or "zero regressions."

But what truly earns trust is **separating what has been verified from what cannot yet be definitively claimed**.

This time with Harness, we re-verified:

- How far can the release be considered safe
- Are there visible regressions
- Is Claude fully latest-ready
- How far has Codex integration progressed

As a result, we can say with considerable confidence that the Claude side is "latest-ready." Meanwhile, the Codex side has completed important organization against the latest stable, but is not yet at the "fully compatible" stage.

This article summarizes those differences in a user-friendly way.

## 1. Conclusion First

The current conclusion fits in four lines:

- The release is in quite good shape
- No known regressions are visible
- Claude can be called latest-ready
- Codex reflects the latest stable, but won't claim full tracking yet

This wording may seem modest.

But for this kind of tool, **being honest about what has been verified earns more trust over time** than embellishing claims.

## 2. Release Is "Quite Good" But Won't Say "Perfect"

First, the release status.

Harness has now reached `v3.15.0`, with a clean working tree. The latest `validate-plugin` was a success.

Looking at that alone, the temptation is to say "it's perfect."

However, there was actually a failure just before the latest run.

What matters here is stating both:

- The latest verification result is good
- But the recent history includes failures

So the honest view of the release is:

- Not in a dangerous state
- Not in a state where it shouldn't be released
- But won't claim "flawlessly green all along"

Harness doesn't round these edges for appearances.

## 3. Are There Regressions?

Regressions are "things that used to work but broke with this change."

Based on current evidence, **no known regressions are visible**.

Verifications performed include:

- Full plugin validation
- Upstream integration point validation
- Reactive hook runtime validation
- sync-plugin-cache validation
- AI residuals check -- confirming no AI-generated remnants or dangerous leftovers

These all passed.

However, even here, we avoid overclaiming.

Even if automated tests pass, if we haven't traced end-to-end through:

- The actual latest CLI
- Actual long work sessions
- Both Claude and Codex pathways

then it's safer not to say "absolutely zero regressions."

So the most accurate statement is:

> Based on automated verification, no regressions are visible

## 4. For Those Using Harness from Claude, "Latest-Ready" Is a Fair Claim

This can be stated quite positively.

After reviewing Claude Code's latest changelog, Harness has incorporated important updates from `2.1.80` through `2.1.86`.

Key improvements include:

### 4-1. Daily operations are slightly lighter

Using the `hooks conditional if` from `2.1.85`, permission hooks are now scoped to "safe-leaning Bash" only.

This means fewer unnecessary hook evaluations for operations like:

- `git status`
- `git diff`
- `pytest`
- `npm run lint`

Not flashy, but effective since it's used every day.

### 4-2. Harder to lose track of assumption changes

`TaskCreated`, `FileChanged`, and `CwdChanged` hooks were incorporated, recording changes to Plans, rules, settings, and worktree switches, and prompting re-confirmation.

This reduces drift like:

- Implementing against outdated Plans
- Continuing with old assumptions after switching worktrees
- Working without noticing rule changes

### 4-3. Safety has been strengthened

`sandbox.failIfUnavailable` and `CLAUDE_CODE_SUBPROCESS_ENV_SCRUB=1` are now built into settings.

In plain terms:

- If sandbox isn't available, don't continue in an unsafe state
- Don't over-share credentials with background processes

So for Claude users, it's fair to say **not just tracking the latest features, but transforming them into forms that make a difference in daily development**.

## 5. For Those Using Harness from Codex, "Significant Progress, But Won't Claim Full Tracking" Is Accurate

This requires a more careful explanation.

Looking at Codex's latest releases, there are newer pre-release versions, but the latest stable is `0.117.0`.

Harness reviewed `0.117.0` and organized:

- Which updates are high-value
- What should be incorporated immediately
- What should be handled in the next phase

Key improvements on the Codex side this time:

### 5-1. Initial quality for heavy flows has stabilized

Skill `effort` and agent `initialPrompt` were organized to encourage deeper thinking from the start in heavy work.

This is effective in scenarios with high oversight costs:

- Implementation
- Review
- Release

### 5-2. Mirror and operational records are aligned

The Codex and OpenCode mirror sides were also updated, along with documentation and test descriptions.

This kind of alignment isn't flashy, but it prevents issues like:

- One side being outdated
- Documentation and implementation drifting
- Different assumptions at each entry point

However, the important point here is:

The key Codex items identified as high-value:

- plugin-first workflow
- resume-aware effort continuity

are still listed as "items with high value for next integration."

So for Codex:

> Important organization against latest stable is done.
> But claiming full compatibility is still premature.

is the most honest statement.

## 6. Why Use Seemingly Weaker Wording

The reason is simple.

In AI coding platforms:

- What has truly been verified
- What hasn't been verified yet
- What's already usable
- What needs to be extended

Mixing these causes inevitable problems later.

What users want to know isn't polished marketing copy, but:

- Can I safely use this right now
- Where should I still be careful
- What's getting stronger next

Harness will continue explaining things without being ambiguous here.

## Summary

In one sentence, Harness today is:

**Claude is latest-ready with high confidence, Codex is moving forward against latest stable but still has a next phase.**

And for the release:

**In quite good shape, but won't claim perfect. No regressions visible in current automated verification, but won't claim absolute zero.**

This isn't flashy wording.

But this kind of honesty makes an operational tool stronger, we believe.

## Short Announcement Text

We organized an honest assessment of where Harness stands today.

- Claude is latest-ready with high confidence
- Codex is progressing against latest stable
- Release is in quite good shape
- No regressions visible in automated verification

But we won't casually say "perfect" or "fully compatible."
The article includes those boundaries.

## Thread Draft

### 1/6

AI coding communications tend toward strong claims like "now supporting the latest" or "fully compatible."

But what really matters is
separating what has been verified
from what cannot yet be definitively claimed.

We organized an honest look at where Harness stands.

### 2/6

Bottom line:

- Release is in quite good shape
- No known regressions visible
- Claude can be called latest-ready
- Codex reflects latest stable but won't claim full tracking

### 3/6

The Claude side can be stated quite positively.

Incorporated important updates from 2.1.80-2.1.86:
- Reduced unnecessary permission hooks
- Reactive hooks make assumption changes harder to miss
- sandbox / env scrub improve safety

### 4/6

We won't say "everything's done" for the Codex side.

Against latest stable, we've organized:
- Skill / agent initial quality
- Mirror alignment
- High-value axes for next integration

But plugin-first and resume continuity are next phase.

### 5/6

Regarding regressions:

Not visible based on automated verification.

But
we won't say "absolutely zero."

This kind of tool earns more trust by not embellishing here.

### 6/6

Rather than flashy claims,

what matters is:
- Can I safely use this right now
- Where should I still be careful
- What's getting stronger next

We organized Harness in that format.

## Cover Image Prompt

High-resolution, white-background tech infographic. Center heading: "Claude Latest, Codex In Progress, Honest Release Status." Left: "Claude 2.1.86 Ready," right: "Codex 0.117 Stable Tracked." Center bottom banner: "Release: Strong, Not Overclaimed." Three cards each showing "Latest Support," "No Known Regression in Automated Checks," and "Honest Boundaries." Clean, futuristic, organized SaaS announcement feel. Colors: white, deep blue, teal, with a touch of orange. Text readable even at thumbnail size. 1:1, generous whitespace, low noise, not overly flashy.

## Alt Text

Infographic organizing three points: Claude is latest-ready, Codex is tracking the latest stable, and the release is in good shape without overclaiming.
