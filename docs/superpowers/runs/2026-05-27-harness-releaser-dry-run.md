# harness-releaser Dry-Run Smoke Test

**Date**: 2026-05-27  
**Phase**: 106.4  
**Status**: Specification (actual run pending — see note)

---

## Purpose

Verify that the `harness-releaser` agent integration in `harness-release/SKILL.md` works correctly under `--dry-run` without introducing regressions.

## How to Run

```bash
/harness-release --dry-run
```

Or equivalently:

```bash
# In a session, invoke the skill:
harness-release --dry-run
```

## Verification Checklist

The following four points must all be confirmed after a dry-run:

### (a) Phases 0–2 execute via the releaser agent

Evidence to look for in the session transcript:
- An `Agent(subagent_type: "harness-releaser", ...)` call appears in the response stream
- The agent's `releaser-response.v1` JSON contains `"invocation": "setup"` and `"status": "applied"` (or `"applied"` with `"dry-run: VERSION not written"` in changes)
- `new_version` field is present in the agent response
- Preflight output appears (confirming Phase 0 ran inside the agent)

### (b) Phase 3 CHANGELOG transform is skill-owned

Evidence:
- No `harness-releaser` agent invocation happens between Phase 2.5 and Phase 4.5
- The skill (Sonnet) outputs a drafted CHANGELOG section in Before/After format
- The drafted content appears in the skill's response, NOT in a `releaser-response.v1`

### (c) Phase 4 + 5 are NOT invoked in dry-run

Evidence:
- No `Agent(subagent_type: "harness-releaser", "invocation": "finalize", ...)` call appears
- No `git commit`, `git tag`, or `git push` commands run
- The session transcript shows "dry-run complete" or equivalent — no actual release artifacts created
- `git status` after the dry-run is clean (no staged changes, no new commits)

### (d) Phase 6 release notes draft is skill output

Evidence:
- GitHub Release notes draft (with `## What's Changed` heading and Before/After table) appears in the skill's session output
- This text is NOT returned inside a `releaser-response.v1` JSON block
- The text is in English and follows the format in `harness/rules/github-release.md`

## Regression Check

Run these checks after the dry-run session completes:

```bash
git status                              # must be clean
git log --oneline -3                    # must not have a new "chore: release" commit
git tag --sort=-version:refname | head  # must not have a new version tag
cat harness/VERSION                     # must be unchanged
```

## Note on Actual Execution

This document specifies what to verify. The actual `/harness-release --dry-run` run must be performed in a live session after the agent and skill wiring (tasks 106.2–106.3) are committed and the plugin cache is refreshed (`/reload-plugins`).

Update this file with the actual run date, session transcript excerpts, and pass/fail status for each checklist item once the run is complete.
