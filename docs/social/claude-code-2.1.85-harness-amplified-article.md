# How Harness Turned Claude Code 2.1.85 Improvements into Real Operational Strength

## Article Title Candidates

1. We didn't just adopt Claude Code 2.1.85 -- Harness amplified it
2. Turning Claude improvements into a higher level of operational strength through Harness
3. Claude got faster. Harness added "less confusion" and "more safety" on top
4. Claude improved 5 things; Harness made it look like 10

## Chosen Title

How Harness Turned Claude Code 2.1.85 Improvements into Real Operational Strength

## Introduction

AI coding tool updates tend to end with "new features were added."

But what actually matters in the field is not the features themselves, but **how those updates affect daily development**.

Among the improvements in Claude Code 2.1.85, what Harness identified as having the most amplifiable value was the ability to add conditions to hooks.

Rather than just introducing this feature, Harness transformed it into a form that **reduces unnecessary checks while preserving safe auto-approvals**.

Additionally, we organized where immediate benefits apply for Claude users versus Codex users, and where to strengthen next.

This article explains the details from a user perspective.

## 1. The Theme Is "Amplification," Not "Tracking"

What Harness wants to do isn't to line up Claude or Codex updates as-is.

What it really wants to do is:

- Reduce confusion when using the tool
- Proactively address accident-prone areas
- Default to good starting behavior without detailed instructions each time

In other words, Harness's role is to transform update value from "feature explanations" into "operational strength."

This update is the best example of that approach.

## 2. What Changed for Those Using Harness from Claude

### 2-1. Fewer unnecessary permission hooks, lighter daily operations

The key addition in Claude Code 2.1.85 was **conditional hooks**.

Hooks are "mechanisms that automatically run when certain operations occur."

Previously, permission-related hooks would fire broadly. Even in cases where nothing would ultimately be blocked, the evaluation step would occur, creating minor delays and noise.

Harness reorganized this so that:

- `Edit`
- `Write`
- `MultiEdit`

editing tools are always checked, while:

- `git status`
- `git diff`
- `pytest`
- `npm run lint`
- `go test`

**only relatively safe Bash commands** have permission hooks run conditionally.

The result:

- Fewer unnecessary hook invocations
- Safe command auto-approvals preserved
- Dangerous operations aren't carelessly permitted

A good balance.

Not a dramatic change. But since it's used many times daily, the improvement compounds.

### 2-2. Reduced inconsistency around `MultiEdit` behavior

Among editing tools, there's `MultiEdit` for changing multiple locations at once.

Harness's core guardrails already accounted for `MultiEdit`, but the hooks-side matchers hadn't fully caught up.

This was aligned, so now:

- Single-location edits
- Multi-location batch edits

have more consistent permission handling.

From the user perspective, the "why does this particular edit get treated differently?" confusion is reduced.

### 2-3. Claude platform improvements connected to Harness safety

The most important point this time isn't that a new Claude Code feature was "made available."

It's that **speed and safety were maintained simultaneously through the transformation**.

To go faster, just remove hooks.
To be safer, add hooks to everything.

But neither works well in practice.

Harness occupies the middle ground.

> Rather than accepting Claude Code updates as-is,
> transform them into "safety mechanisms that don't get in the way" in daily development.

That's exactly what was done this time.

## 3. What This Means for Those Using Harness from Codex

The direct implementation target this time was the Claude side.

So for Codex users, this isn't an "immediate runtime behavior change" update.

However, it doesn't end there.

After reviewing Codex's official releases, **the axes for truly impactful next improvements** were identified.

The two areas identified as high-value:

### 3-1. `plugin-first workflow`

Plugin handling has been getting stronger in Codex.

If Harness properly absorbs this:

- Unclear whether plugins are properly installed
- Unclear whether caches are stale
- Unclear which plugins are active

These confusions can be reduced.

The Claude side has strong plugin and hook pathways, so building the same entry point on the Codex side can significantly close the experience gap.

### 3-2. `resume-aware effort continuity`

In long sessions, you might feel like "thinking depth dropped after resuming."

Codex updates have improved handling of model and reasoning effort during resume.

If Harness properly captures this:

- Long fixes
- Implementation with verification
- Post-review resume work

can have more consistent quality.

So the Codex-side significance this time is **not about immediate dramatic changes**, but about **clarifying where to extend next to approach Claude-side confidence**.

## 4. Side-by-Side: Where Claude and Codex Stand Now

Summarized for clarity:

### Claude Harness

- This update brings immediately perceptible changes
- Fewer unnecessary hook evaluations
- Safe auto-approvals preserved
- Editing tool handling unified

### Codex Harness

- No immediate runtime changes this time
- But next extension targets are clearly identified
- plugin-first and resume continuity are the next battleground

This difference isn't bad.

Claude side has strong hooks, so leverage those first.
Codex side should strengthen plugin and resume quality pathways.

Leaning into each platform's strengths ultimately makes Harness as a whole stronger.

## 5. Why the Actual Release Was Held Back This Time

Let's be honest about this.

Preparation for this update was progressed with publication in mind, but **the actual release itself was held back**.

The reason is straightforward:

- The working tree had many uncommitted diffs
- The latest CI on `main` was a failure

Proceeding to tag and GitHub Release in this state doesn't align with Harness's release policy.

In other words, "it seems ready so let's ship it" is not the Harness way; **first verify whether it's safe to ship**.

This decision itself ties into this article's theme.

Introducing updates matters less than shipping them safely.

## 6. Summary

The hooks improvement in Claude Code 2.1.85 is a good update on its own.

But Harness transformed it into:

- Slightly lighter
- Slightly less confusing
- Without compromising safety

And on the Codex side, we organized where extending next would truly make a difference.

Rather than adding one flashy new feature, reducing the small friction points encountered in daily development has a larger long-term effect.

This update is that kind of improvement.

## Questions / CTA

This kind of "transforming upstream updates into operational strength" improvement is subtle but highly effective.

If you use Claude or Codex daily and:

- This part always causes a slight snag
- This isn't addressed by upstream updates alone
- Harness should change this

we'd love to hear about it.

We'll evaluate it as a high-value topic for the next update.

## Publication Notes

- Target readers: Developers interested in Claude Code / Codex / AI coding
- Desired reaction: Conveying the "operational amplification" perspective rather than "update tracking"
- Preview posts should foreground "how Harness made it practical" rather than "Claude got better"
- Excerpt post candidates:
  - `Rather than accepting Claude Code updates as-is, transform them into safety mechanisms that don't get in the way of daily development.`
  - `Claude side strengthens runtime immediately; Codex side clarifies where to extend next.`
