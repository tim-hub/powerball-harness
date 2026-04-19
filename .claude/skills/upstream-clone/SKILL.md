---
name: upstream-clone
description: "Analyzes upstream fork PRs against the local codebase and produces a structured triage (can port / cannot port / already done better), then invokes harness-plan to produce an actionable integration phase. Use when reviewing upstream PRs for selective porting into the local downstream branch, understanding what an upstream fork changed, or planning a cherry-pick integration without rebasing."
when_to_use: "upstream PRs, check upstream changes, port upstream, upstream integration, cherry-pick upstream, upstream fork analysis, what changed upstream, bring upstream changes"
user-invocable: true
allowed-tools: ["Bash", "Read", "Glob", "Grep", "WebFetch", "Agent"]
---

# Upstream Clone — Selective Upstream Porting

A local skill for understanding what changed in an upstream fork's PRs, deciding what to bring into the local branch, and producing a harness-plan phase to implement the portable items.

**Core principle**: Never rebase the upstream. Cherry-pick by concern, remapped to the local file layout.

## When to Use

- You have one or more upstream PR URLs and want to know what to borrow
- You want to audit how far the upstream fork has diverged from the local branch
- You want to generate a Plans.md phase (via `harness-plan`) for the portable items

## When Not to Use

- Full upstream rebase or sync (use `git merge` or `git rebase` directly)
- Reviewing internal PRs or local branch history (use `harness-review`)

---

## Prerequisites

1. **Upstream remote configured**: `git remote -v` must show an `upstream` entry pointing at the fork.
   If missing, add it:
   ```bash
   git remote add upstream https://github.com/<org>/<repo>.git
   ```

2. **PR numbers or URLs**: Collect the upstream PR numbers before starting. URLs in the format
   `https://github.com/<org>/<repo>/pull/<N>` — extract the number `N`.

3. **Local branch is clean**: `git status` should show no uncommitted changes before fetching.

---

## Execution Flow

### Step 0: Fetch upstream PRs

Fetch each PR as a local branch. Do **not** use `gh pr view` — it has known TLS issues in this environment. Use direct git fetch instead:

```bash
git fetch upstream pull/91/head:pr-upstream-91
git fetch upstream pull/92/head:pr-upstream-92
# repeat for each PR number
```

This creates local branches `pr-upstream-N` that you can diff against master.

### Step 1: Read each PR's diff and purpose

For each PR, run:

```bash
git log --oneline master..pr-upstream-N        # commit summary
git diff --stat master...pr-upstream-N         # files changed
git diff master...pr-upstream-N                # full diff (pipe to head -200 for large PRs)
```

Read enough of the diff to answer:
- **What problem does this PR solve?** (the user-facing or developer-facing improvement)
- **What files does it touch?** (understand coupling — does it touch core/, skills/, hooks/, go/, tests/?)
- **Is it self-contained or does it depend on other upstream-only concepts?**

### Step 2: Compare to local codebase

For each meaningful change in the PR, grep and read the local equivalent:

```bash
git grep -l "<function or concept name>" --   # does local have an analogue?
```

Read the local implementation to determine whether:
- We have the same feature but wired differently
- We have a stronger or more architecturally appropriate version
- The concept is completely absent locally
- The upstream version assumes a different architecture (TypeScript vs Go, different schema, etc.)

### Step 3: Triage into three buckets

See `${CLAUDE_SKILL_DIR}/references/triage-framework.md` for the full decision matrix and examples.

**Summary**:

| Bucket | Label | Meaning |
|--------|-------|---------|
| ✅ Can Port | `PORTABLE` | Clean improvement applicable to our architecture |
| ❌ Cannot Port | `SKIP` | Architecture mismatch, incompatible schema, or platform-specific |
| 🔄 Already Done | `HAVE_IT` | We have the same or a better version already |

For each item, write a one-line rationale. Be specific:
- `SKIP — upstream uses TypeScript guardrail engine; we use Go-native bin/harness`
- `HAVE_IT — our CheckAdvisorDrift already scans last 200 lines; upstream deferred this optimization`
- `PORTABLE — filepath.Clean on config paths is a pure defensive improvement, zero coupling`

### Step 4: Produce the triage report

Output a structured report in this format before proceeding to harness-plan:

```
## Upstream Triage: PRs #N, #M, ...
Source: <upstream repo URL>
Date: YYYY-MM-DD

### PR #N — <PR title or one-line summary>

| Change | Bucket | Rationale |
|--------|--------|-----------|
| <feature/fix description> | PORTABLE | <why it applies> |
| <another change> | SKIP | <why it doesn't apply> |
| <another change> | HAVE_IT | <what we already have> |

### PR #M — ...
...

### Summary
- PORTABLE: N items (these become tasks)
- SKIP: M items
- HAVE_IT: K items
```

Present this report to the user and wait for acknowledgement before proceeding to Step 5. The user may want to move items between buckets.

### Step 5: Invoke harness-plan to create the integration phase

After the user confirms the triage, invoke the `harness-plan` skill to create a new phase in Plans.md.

Provide harness-plan with:
- The phase goal: "Port [N] items from upstream PRs #X, #Y, #Z — selective integration, no rebase"
- One task per PORTABLE item, with:
  - A precise description of what to implement
  - A clear DoD (Definition of Done) that is verifiable (test name, config key, command output)
  - Dependencies between tasks (e.g., if task B requires a Go type added in task A)
  - Whether it's mandatory or optional (optional = lower priority, skip without breaking the phase)
- A phase-level note listing SKIP and HAVE_IT items with rationale (so future contributors know why they were excluded)

The harness-plan skill will write the phase to Plans.md in the v2 table format.

---

## Language Policy

All output — triage reports, Plans.md tasks, rationale text — must be in **English**. This includes task descriptions, DoD, and SKIP/HAVE_IT rationale. Do not translate from upstream Japanese or other languages.

---

## Key Lessons (from Phase 75 and Phase 77)

- **`gh pr view` has TLS issues in this environment** — always use `git fetch upstream pull/N/head:pr-upstream-N`
- **Upstream architecture diverges** — their repo may still reference TypeScript core/, old agent structures, or a different version numbering. Map to local equivalents before deciding PORTABLE vs SKIP
- **Upstream deferrals are opportunities** — if the upstream PR says "deferred optimization: read last N lines instead of whole file", that's a chance to adopt the stronger version from day one
- **CodeRabbit security flags travel** — if upstream has a CodeRabbit comment about a security issue (e.g., `os.Executable()` vs hardcoded binary path), apply the fix in our port even if their original PR didn't include it
- **Preflight before wiring** — for hook wiring tasks, always check whether our existing handler already covers the same ground before adding new wiring (risk: double-injection)
- **Keep SKIP rationale narrow** — write `SKIP — this specific reason` not `SKIP — incompatible`, so future reviewers understand the exact decision

---

## Quick Reference

| Phrase | What to do |
|--------|-----------|
| "understand upstream changes" | Run this skill |
| "what can we borrow from upstream PR #N" | Run this skill |
| "port upstream changes" | Run this skill, then harness-plan |
| "upstream PR triage" | Run this skill |
| "bring changes from fork" | Run this skill |
