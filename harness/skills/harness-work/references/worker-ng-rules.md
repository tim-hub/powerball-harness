# Worker NG Rules

Three hard constraints applied to every Worker — NG-1 (Plans.md markers are Lead-owned), NG-2 (no embedded git repos), NG-3 (no nested teammate spawn). Violations are grounds for immediate escalation.

---

Three hard constraints applied to every Worker before and during implementation. Violations are grounds for immediate escalation — the Lead must not cherry-pick a commit that broke any NG rule.

## NG-1: Plans.md Markers Are Lead-Owned

`cc:TODO`, `cc:WIP`, `cc:Done`, and all other `cc:*` / `pm:*` status markers in `Plans.md` are **exclusively managed by the Lead**. Workers must not read, write, or touch `Plans.md` during implementation.

**Why**: status markers are the coordination layer between Lead, Worker, and Reviewer. A Worker modifying its own `cc:WIP` entry, or pre-emptively writing `cc:Done`, breaks the Lead's ability to track actual completion state and can cause double-commits or skipped reviews.

**Check**: Lead verifies `git diff HEAD Plans.md` is empty before cherry-picking a Worker commit. Any `Plans.md` modification in the diff → reject + send amendment request.

## NG-2: No Embedded Git Repositories

Workers must not create, clone, or leave an embedded `.git` directory inside the worktree. An embedded repo inside the worktree causes `git` operations in the parent (cherry-pick, diff, status) to behave unpredictably and can silently omit files from the diff.

**Why**: git treats subdirectories containing a `.git` folder as submodules or ignores them entirely. This causes the Lead's diff to be incomplete and review to miss changed files.

**Check**: Before committing, run `find . -mindepth 2 -name ".git" -not -path "./.git"` in the worktree root. Any hit → abort and clean up before committing.

## NG-3: No Nested Teammate Spawning

Workers must not spawn Reviewer, Advisor, or other Worker agents. Team member spawning is exclusively the Lead's responsibility.

**Why**: Claude Code v2.1.69+ prohibits nested teammate spawning at the platform level. Attempting it causes unpredictable behavior (silent drop or error). Even where technically possible, nested spawning creates untracked review loops that bypass the Lead's quality gate.

**Check**: Worker prompts must not contain `Agent(subagent_type=...)` calls targeting harness agent types. The Lead validates this by reviewing the Worker's implementation plan before spawning.
