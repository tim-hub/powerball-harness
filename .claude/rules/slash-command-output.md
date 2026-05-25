# Slash Command Output Summarization Rule (P35)

## Rule

When a skill invocation returns a `<local-command-stdout>` tool result that is **10 or more lines**, the host Claude **must**:

1. Produce a **1–3 line assistant message** that summarizes the result
2. Explicitly state the **next action**: wait for user input, continue autonomously, or request a decision

Do not pass the raw long output to the user without a summary. The summary is the primary signal; the raw output is context.

## Trigger

- Tool name contains `Skill` or `SlashCommand`
- The `<local-command-stdout>` block in the result is ≥ 10 lines

## Required Summary Structure

```
[1-3 line summary of what the skill output means]
Next: [wait / continuing with X / need decision on Y]
```

## Skill-side Instruction Literal

Skills that produce long output should include this instruction line at their conclusion block:

```
↑ Claude will summarize this result. Type a new prompt to redirect or press Enter to continue.
```

This tells the host Claude that summarization is expected.

## Why

Long skill outputs (validation reports, governance checks, release notes) often exceed the cognitive bandwidth of a single message. The summarization rule ensures:
- Key signals (pass/fail, blockers, next task) are surfaced immediately
- The user can redirect without reading hundreds of lines
- Automated sessions (breezing, workers) don't stall on long intermediate output

## Reference

- Upstream: patterns.md P35, codified 2026-05-19
- Applies to: harness-release, harness-review, any skill outputting structured reports
