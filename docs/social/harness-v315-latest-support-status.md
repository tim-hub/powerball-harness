# How Far Can Harness v3.15.0 Go with Claude and Codex Latest Support

## Article Title Candidates

1. Harness v3.15.0: Claude is latest-ready. How far can we go with Codex?
2. Claude is latest-ready, Codex is at stable latest with next steps clearly defined
3. We won't say "perfect." An honest look at Harness v3.15.0's latest support status
4. Claude and Codex latest updates: How much has Harness incorporated?

## Chosen Title

How Far Can Harness v3.15.0 Go with Claude and Codex Latest Support

## Introduction

When communicating about AI development tools, there's always a temptation to make strong claims like "now supporting the latest version."

But what really matters is:

- What is actually supported
- What can we confidently claim
- What is still an implementation task for the future

Without being ambiguous about these distinctions.

With Harness v3.15.0, the Claude Code side has reached a state where we can make strong claims. Meanwhile, the Codex side has made significant progress in organizing the latest stable version's integration strategy, but we won't yet say "fully tracking the latest."

This article organizes those differences based on official information and implementation/verification results as of 2026-03-28.

## 1. The Bottom Line

Stating the conclusion upfront, the current status is:

### Using Harness from Claude

- Safe to assume the latest Claude Code `2.1.86`
- Important updates from `2.1.80` through `2.1.86` have been implemented on the Harness side
- Tangible improvements in "responsiveness," "safety," and "reduced assumption drift"

### Using Harness from Codex

- Comparison and organization against the latest stable `0.117.0` is complete
- However, `plugin-first workflow` and `resume-aware effort continuity` remain as next implementation tasks
- In other words, "organization based on latest stable is done; full tracking is still to come"

### Release Status

- The release is in quite good shape
- But we won't say "perfect" or "absolutely no regressions"
- The accurate statement is that no major issues are visible in current automated verification

## 2. Why Claude Side Can Be Called "Latest-Ready"

Looking at Claude Code's official changelog, there are features that clearly synergize well with Harness among the recent updates.

Key items actually incorporated this time include:

- `TaskCreated` / `FileChanged` / `CwdChanged` hooks
- `sandbox.failIfUnavailable`
- `CLAUDE_CODE_SUBPROCESS_ENV_SCRUB=1`
- skill `effort`
- agent `initialPrompt`
- hooks conditional `if` field
- rules `paths:` YAML list support

On the Harness side, these weren't just "introduced" -- they were transformed into forms that make a difference in daily development.

For example, for those using Claude, the following changes apply:

### 2-1. Fewer unnecessary permission checks, slightly lighter experience

Using the `hooks conditional if field`, permission hooks are now scoped to "safe-leaning Bash commands only."

For example, operations like:

- `git status`
- `git diff`
- `pytest`
- `npm run lint`

now incur fewer unnecessary hook evaluations for relatively safe operations.

This means:

- Less likely to slow down
- Less noise
- Safe auto-approvals are preserved

A good balance.

### 2-2. Easier to notice when assumptions change

With reactive hooks, changes like:

- `Plans.md` changed
- Rules changed
- Settings changed
- Worktree or workspace changed

are now caught, prompting re-confirmation before the next action.

Subtle, but highly effective in long work sessions.

### 2-3. Safety settings move from "available if you know about them" to "enabled by default"

`sandbox.failIfUnavailable` prevents continuing in an unsafe state when sandbox is unavailable.

`CLAUDE_CODE_SUBPROCESS_ENV_SCRUB=1` prevents passing too many credentials to background processes like Bash and hooks.

These are now built into Harness settings, so they take effect from the start without users having to remember them each time.

This is why the Claude side can say "latest-ready" -- **the value from the latest changelog directly translates into runtime improvements**.

## 3. Why Codex Side Won't Say "Fully Tracked" Yet

Honesty matters here.

Codex's latest stable is `0.117.0`. We've examined it and organized where the value lies.

However, the main focus of immediate implementation this time was the Claude side.

Progress on the Codex side this time includes:

- Mirror alignment
- `effort` frontmatter integration for heavy workflows
- Stabilizing initial quality for heavy skills and flows

These are meaningful improvements.

But the items identified as "truly impactful next" after reviewing the latest stable:

- `plugin-first workflow`
- `resume-aware effort continuity`

are still next-phase tasks.

So for Codex:

- Investigation complete
- Comparison axes are clear
- Partially integrated
- But we won't say all value from the latest stable has been incorporated

This wording may appear weak.

But in reality, **not being ambiguous about what's incomplete makes the next improvement stronger**.

## 4. How to Answer "Is the Release Perfect?"

Here too, it's better not to overstate.

What can be said about the current release:

- No uncommitted changes
- `v3.15.0` is at `main` HEAD
- Latest `validate-plugin` is success
- `harness-review` shows 0 critical / major
- `validate-plugin` shows `43 pass / 2 warning / 0 fail`

This is quite good.

But we won't say "perfect" or "absolutely no regressions."

The reason is simple: in software, what truly matters is "whether there are problems within what we can currently see," not making guarantees about the future.

So the accurate statement is:

> The release is in quite good shape.
> No major issues are visible based on automated verification.
> But we don't promise perfection or zero risk.

## 5. Translated to User Perspective

Setting aside technical details, this time's message is quite simple.

### For Claude users

- Slightly lighter than before
- Slightly safer
- Harder to lose track of assumptions in long sessions
- Easier to benefit from latest updates

### For Codex users

- Initial quality for heavy flows has improved
- But not yet at the stage of fully leveraging the latest stable's strengths
- What to strengthen next is already clear

In other words, Claude has "immediately effective improvements" at the center, while Codex is at the "next strengthening targets are now clearly visible" stage.

## 6. Summary

Harness v3.15.0 examined the latest updates for both Claude and Codex, then:

- Strongly incorporates Claude
- Doesn't force-fit Codex, but clearly defines next tasks

This was a deliberate release decision.

This approach is understated but quite important.

Rather than claiming "everything is supported":

- Strongly assert what's truly supported
- Say "not yet" for what's not
- But be clear about what comes next

This results in a product that builds trust more effectively.

This update embodies that Harness philosophy.

## Questions / CTA

If you're interested in this kind of approach -- "not just echoing upstream updates as promotions, but transforming them into real operational strength" -- we'd love to hear whether you primarily use Claude or Codex in your daily work.

## Cover Image Prompt

High-resolution, white-background tech infographic. Center heading: "Claude Latest, Codex In Progress, Honest Release Status." Left: "Claude 2.1.86 Ready," right: "Codex 0.117 Stable Tracked." Center bottom banner: "Release: Strong, Not Overclaimed." Three cards each showing "Latest Support," "No Known Regression in Automated Checks," and "Honest Boundaries." Clean, futuristic, organized like a SaaS announcement. Colors: white, deep blue, teal, with a touch of orange. Text readable even at thumbnail size. 1:1, generous whitespace, low noise, not overly flashy.

## Alt Text

Infographic organizing three points: Claude is latest-ready, Codex is tracking the latest stable, and the release is in good shape without overclaiming.
