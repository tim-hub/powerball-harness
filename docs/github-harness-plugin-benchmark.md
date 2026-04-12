# GitHub Harness Plugin Benchmark

Last updated: 2026-03-06

This document is a dated snapshot comparing `claude-code-harness` against popular **harness / workflow plugins for Claude Code** on GitHub, from the perspective of **how standard operations change after adoption**.

- This is a **harness comparison**, not a **popularity contest**
- GitHub stars are used only as "selection criteria for comparison targets"
- First we list "what becomes standard after adoption," then explain what the differences mean
- General AI coding agents (Aider, OpenHands, etc.) and curated lists are excluded from this comparison as they are **not standalone harnesses**

## Compared Repositories

As of 2026-03-06, the targets are publicly available repos on GitHub that claim to be "multi-step workflow / plugin / harness for Claude Code" and have sufficient public information for comparison.

| Repo | GitHub stars | Included because |
|------|--------------|------------------|
| [obra/superpowers](https://github.com/obra/superpowers) | 71,993 | The most popular workflow / skills plugin. Essential comparison target |
| [gotalab/cc-sdd](https://github.com/gotalab/cc-sdd) | 2,770 | A popular Claude Code harness emphasizing requirements-driven development flow |
| [Chachamaru127/claude-code-harness](https://github.com/Chachamaru127/claude-code-harness) | 232 | This repo |

## User-Visible Comparison Table

Legend:

- `✅` Available as a standard flow immediately after adoption
- `△` Possible with effort, but not the primary path
- `—` Not a primary value proposition

| What users care about | Claude Harness | Superpowers | cc-sdd |
|------------------------|----------------|-------------|--------|
| Plans survive in the repo instead of disappearing in conversation | ✅ | ✅ | ✅ |
| Implementation flows naturally after approval | ✅ | ✅ | △ |
| Review is part of the standard process before completion | ✅ | ✅ | △ |
| Dangerous operations are stopped by runtime guards | ✅ | △ | — |
| Verification can be re-run later using the same procedures | ✅ | △ | ✅ |
| After approval, the entire flow can run end-to-end | ✅ | △ | — |

## What These Differences Mean

### Claude Harness

- The strongest points are **standardized flow**, **runtime guards**, and **re-runnable verification**
- Plan -> Work -> Review are set up as independent paths, with `/harness-work all` as a shortcut for end-to-end execution
- Suited for those who want "every run to follow the same pattern without falling apart," rather than "just make it work each time"

### Superpowers

- The strongest points are **workflow breadth** and **clear adoption story**
- The flow from planning through implementation, review, and debugging is easy to see, with strong auto-triggers
- However, mechanisms to stop dangerous operations via runtime rules and re-runnable evidence trails are not as prominently positioned in the standard flow compared to Harness

### cc-sdd

- The strongest point is **spec-driven discipline**
- The `Requirements -> Design -> Tasks -> Implementation` flow is clear, with dry-run, validate-gap, and validate-design
- However, from the public perspective, independent review stages and end-to-end execution paths do not appear as strongly in the standard flow compared to Harness

## README Presentation

For README and landing pages, the following wording is natural:

> If you want to expand your workflow toolkit, try Superpowers.
> If you want to strengthen requirements -> design -> task discipline, try cc-sdd.
> If you want to turn planning, implementation, review, and verification into a robust standard flow, try Claude Harness.

## Scoring Notes

- `Plans survive in the repo instead of disappearing in conversation`
  - Harness: `Plans.md` / `/harness-plan`
  - Superpowers: brainstorming / writing-plans workflow
  - cc-sdd: requirements / design / tasks workflow
- `Implementation flows naturally after approval`
  - Harness: `/harness-work --parallel`, Breezing, worker/reviewer flows are part of the standard flow
  - Superpowers: parallel agent execution / subagent workflows are clearly visible publicly
  - cc-sdd: Claude agent variant shows multiple subagents, but this is not positioned as a central feature across all usage patterns
- `Review is part of the standard process before completion`
  - Harness: `/harness-review` and `/harness-work all`
  - Superpowers: code review workflow is explicit
  - cc-sdd: validate commands are explicit, but the degree to which code review is positioned as an independent stage is somewhat weaker
- `Dangerous operations are stopped by runtime guards`
  - Harness: TypeScript guardrail engine + deny / warn rules
  - Superpowers: workflow discipline and hooks are visible, but compiled deny / warn runtime engine is not front-and-center
  - cc-sdd: An explicit runtime safety engine is not clearly visible from the public README
- `Verification can be re-run later using the same procedures`
  - Harness: validate scripts + consistency checks + evidence pack
  - Superpowers: verify-oriented workflows exist, but artifact pack is not prominently featured
  - cc-sdd: dry-run / validate-gap / validate-design are available
- `After approval, the entire flow can run end-to-end`
  - Harness: `/harness-work all`
  - Superpowers: auto-triggered workflows exist, but a published single command for the same purpose is not prominently featured
  - cc-sdd: spec-based command set exists, but a single path to aggregate a full loop after approval is not prominently featured

## Notes

- Stars change daily, so this table is a **dated snapshot**
- This comparison focuses on "user-visible harness feature differences," not "market popularity"
- There are axes where `Superpowers > Claude Harness`. Particularly, ecosystem / adoption / workflow story strength stands out
- There are axes where `cc-sdd > Claude Harness`. Particularly, the clarity of requirements-driven discipline is a strength
- When including in README, it is more natural to write **who each is suited for based on priorities** rather than making definitive win/lose claims

## Evidence Used

### Local evidence

- [README.md](../README.md)
- [docs/claims-audit.md](claims-audit.md)
- [docs/distribution-scope.md](distribution-scope.md)
- [docs/evidence/work-all.md](evidence/work-all.md)

### Public GitHub sources

- [obra/superpowers](https://github.com/obra/superpowers)
- [gotalab/cc-sdd](https://github.com/gotalab/cc-sdd)
- [Chachamaru127/claude-code-harness](https://github.com/Chachamaru127/claude-code-harness)
