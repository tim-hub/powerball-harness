# Positioning Notes

Last updated: 2026-03-06

In short, the value of `claude-code-harness` for public-facing purposes is not about "adding more skill packs," but about **being able to run Plan -> Work -> Review with runtime enforcement and verification**.

## Core Message

- Harness treats `5 verb skills + TypeScript guardrail engine` as the core product
- The value is not in the number of commands, but in `guardrail`, `review`, `consistency`, and `evidence` working together as a unified system
- Legacy / optional buckets like `commands/` and `mcp-server/` are not weaknesses; they can be explained as operational assets as long as their boundaries are clearly documented

## Public Comparison Language

- Avoid: "overwhelmingly superior to competitors," "complete victory"
- Use: "strong runtime enforcement," "clear verification path," "claims are backed by reproducible evidence"
- In competitor comparisons, do not dismiss their philosophy or adoption track record; instead focus Harness strengths on guardrail / evidence / operator clarity

## Recommended One-liner

> A harness that goes beyond extending Claude Code with skill packs to enable Plan -> Work -> Review with guardrails and verification.

## Proof Points

- TypeScript guardrail engine (`core/`)
- 5 verb skills (`skills-v3/`)
- consistency check and plugin validation
- `/harness-work all` evidence pack
