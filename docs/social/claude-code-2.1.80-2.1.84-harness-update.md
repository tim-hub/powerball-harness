# How Harness Incorporated Claude Code 2.1.80-2.1.84

## Article Title Draft

How Harness "Amplified" Claude Code Updates

## Introduction

Claude Code 2.1.80 through 2.1.84 included important updates around hooks, skill frontmatter, agent frontmatter, safety settings, and rules scope.

What Harness did this time wasn't simply "adding support for new features." It was about connecting improvements that Claude Code makes on its own to the actual development flow, making them even more effective.

Put another way, where Claude Code improves by 10 on its own, adding Harness extends that to 15 or 20.

This article summarizes -- in accessible terms -- what changed on the Claude Code side, how Harness incorporated it, and what improved as a result.

## 1. The Theme Is "Amplification," Not "Tracking"

AI development tool updates tend to end with "new features were added."

But in reality, what matters more than the features themselves is:

- How to connect them to daily implementation
- How to reduce incidents in team operations
- How to stabilize without per-prompt adjustments each time

Harness is the layer for that. Rather than just accepting Claude Code's new features, it connects them to planning, implementation, review, safety mechanisms, and rule operations, carrying them all the way to where they make a difference in the daily development experience.

## 2. Claude Code Updates and Harness Integration

### 2-1. New hooks were turned into development-time awareness

First, hooks. Hooks are "mechanisms that automatically run something the moment a certain event occurs."

Claude Code 2.1.83 and 2.1.84 made the following events available:

- `TaskCreated`
- `CwdChanged`
- `FileChanged`

On their own, this is just "more events are available."

Harness connected new reactive hooks to these:

- When a task is created, record it
- When `Plans.md` changes, prompt a re-read before the next implementation
- When rules or settings change, alert that assumptions have changed
- When moving to a different worktree or workspace, prompt context re-confirmation

This transforms "being able to capture events" into "making it harder to lose track of assumption changes."

This reduces:

- Proceeding with implementation against outdated plans
- Operating under old assumptions after moving to a different worktree
- Working without noticing rule updates

### 2-2. Safety settings were made to take effect automatically

Next, safety.

Claude Code 2.1.83 introduced particularly valuable settings:

- `sandbox.failIfUnavailable`
- `CLAUDE_CODE_SUBPROCESS_ENV_SCRUB=1`

In plain terms:

- If sandbox is unavailable, don't continue in an unsafe state
- Minimize passing credentials to background Bash, hooks, and integration processes

Harness didn't leave these as "configurable" -- they were built in as plugin defaults. They were also reflected in distribution cache sync, ensuring they take effect in real operation after installation.

This difference is significant.

Features that merely exist may not take effect unless you configure them yourself. But building them into Harness reduces cases of "unknowingly forgetting to enable safety features."

### 2-3. Heavy work now thinks deeply from the start

Claude Code 2.1.80 added the ability to set `effort` in skill frontmatter.

Frontmatter is the configuration section at the top of Markdown files. `effort` specifies "how deeply to think."

Harness assigned this to key skills:

- `harness-work`: high
- `harness-review`: high
- `harness-release`: high
- `harness-plan`: medium
- `harness-sync`: medium
- `harness-setup`: medium

The benefit is that thinking depth automatically starts higher for the task at hand, without saying "think deeply this time" each session.

This is especially effective in review and release scenarios where oversight costs are high.

### 2-4. Sub-agent initial behavior stabilized

Claude Code 2.1.83 added the ability to set `initialPrompt` for agents.

This means providing sub-agents with "what to prioritize thinking about" at the very moment of startup.

Harness added initial behavior instructions for:

- Worker
- Reviewer
- Scaffolder

For example:

- Worker first organizes the task, DoD, target files, and verification approach
- Reviewer first confirms verdict criteria and avoids confusing minor with major
- Scaffolder first organizes the existing structure and current setup goals

This reduces the risk of sub-agents immediately running in the wrong direction.

### 2-5. Rules scope was made more robust

Claude Code 2.1.84 enabled YAML arrays for `paths:` in rules and skills.

Previously, multiple target patterns were often held in a single string, which was:

- Hard to read
- Hard to extend
- Prone to breaking on commas or whitespace

Harness updated templates and `localize-rules.sh` to use YAML arrays for `paths:`.

This looks minor but:

- Scope becomes more readable
- Future additions are easier
- Fewer accidents in auto-generation

### 3. What Improved as a Result

Summarizing for non-specialists, this update improved five things:

- Easier to notice when assumptions change
- Safety features now take effect in real operation, not just configuration
- Heavy work thinks deeply from the start
- Sub-agents are less likely to deviate from their roles
- Rule operations are more readable and less fragile

None of these are dramatic visual changes, but they make a difference when used daily.

That's why this update is best described not as "Claude Code got new features" but as "Harness turned those improvements into real development strength."

## 4. What Was Deliberately Not Expanded This Time

Not everything was incorporated at equal weight this time.

The reason is simple: even when Claude Code improves, some items see their value significantly extended through Harness while others don't.

This time, priority went to items that "directly connect to Harness's strengths":

- Reactive hooks
- Safety settings
- Skill / agent initial quality
- Rules maintainability

## Summary

The Claude Code 2.1.80-2.1.84 updates are valuable on their own.

But adding Harness allowed transforming that value into "forms that affect the development flow."

This update targeted exactly that.

Rather than accepting the improved parts of Claude Code as-is, they were connected to the implementation, review, safety, and operational layers, making them stronger to use.
That's what this Harness-side update is about.

## Article Cover Image Prompt

High-resolution, white-background tech infographic. Center heading: "Harness Amplifies Claude Code Updates." Left: "Claude Code 2.1.80-2.1.84," right: "Harness." Five strong arrows pointing from center to right. Five cards labeled "Reactive Hooks," "Safer Sandbox," "Skill Effort," "Agent Initial Prompt," and "YAML Paths." Clean, futuristic editorial design with SaaS announcement feel. Colors: white, deep blue, teal, minimal orange. Text readable and clearly organized, works well as X article cover at thumbnail size. 1:1, generous whitespace, low noise, minimal icons, tastefully understated design.

## Cover Image Text Draft

Claude Code Updates
Amplified by Harness

- Reactive Hooks
- Safer Sandbox
- Better Agent Starts
- Deeper Skill Execution
- More Reliable Rules

## Alt Text

Infographic showing how Harness amplifies Claude Code updates. Five items are organized: Reactive Hooks, Safer Sandbox, Skill Effort, Agent Initial Prompt, and YAML Paths.

## Short Announcement for Article Sharing

We incorporated Claude Code 2.1.80-2.1.84 into Harness.
This time's theme is "amplification," not "tracking."

New hooks, safety settings, skill / agent initial behavior, and rules scope -- all connected to forms that affect real operations.
Summarized in an article.
