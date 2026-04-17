# Decisions (SSOT)

This file is the Single Source of Truth (SSOT) for important decisions.
We avoid excessive discussion logs and keep **conclusions, rationale, and trade-offs** short and reliable.

## Index

- D1: 2026-04-12: Adopt 5-verb skills + 7-agent structure #architecture
- D2: 2026-04-12: Implement declarative guardrails with TypeScript core engine #guardrails #typescript _(superseded by D9)_
- D3: 2026-04-12: Translate all comments, tests, and output in core/ to English #i18n #core
- D4: 2026-04-12: Manage better-sqlite3 as optionalDependencies #dependencies #node24
- D5: 2026-04-12: Unify Codex integration through official plugin #codex #policy
- D6: 2026-04-12: Require implementation or category classification for Feature Table additions #quality #policy
- D7: 2026-04-12: Keep hooks/ as thin shims, delegate logic to Go binary #architecture #hooks
- D8: 2026-04-12: Retire OpenCode platform, keep Codex via symlinks #architecture #cleanup
- D9: 2026-04-13: Migrate guardrail engine from TypeScript to Go native binary #guardrails #go
- D10: 2026-04-13: Hooks fail-open (exit 0) when platform binary is missing #hooks #ux
- D11: 2026-04-13: Auto-download platform binary on plugin install via Setup hook #distribution #binary
- D12: 2026-04-15: Makefile as stable CI interface layer between workflows and script paths #ci #architecture
- D13: 2026-04-15: MARKETPLACE_NAME and PLUGIN_NAME in cache scripts must match marketplace.json #distribution #cache
- D14: 2026-04-15: Consistency check sections must be explicitly skipped, never silently no-op #quality #ci
- D15: 2026-04-17: Concurrent hook fan-out and ScheduleWakeup-based harness-loop runtime #architecture #hooks #breezing

---

## D1: 2026-04-12: Adopt 5-verb skills + 3-agent structure in v3 architecture #architecture #v3

### Conclusion

- Consolidate skills into 5 verbs (plan / execute / review / release / setup)
- Consolidate agents into 7 (worker / reviewer / scaffolder / team-composition / ci-cd-fixer / error-recovery / advisor), reduced from 11
- Use skills/ as the source of truth; codex/ references via symlinks (opencode retired in Phase 36)

### Background

- Under the old structure, 20+ skills and 11 agents led to growing duplication and ambiguous responsibilities

### Options

- A: Continue maintaining existing skills individually
- B: Consolidate around verbs with clear single responsibility

### Rationale

- Verb-based organization maps directly to user intent ("I want to plan" -> plan)
- Reducing agent count significantly lowers context consumption and maintenance cost
- Symlinks maintain compatibility with Codex CLI (OpenCode retired)

### Impact / Trade-offs

- Legacy skill names (work, breezing, etc.) must be retained in skills/ for backward compatibility
- Symlink health checks need to be integrated into CI

### Review Conditions

- If Claude Code natively provides multi-agent orchestration

### Related

- rules: `.claude/rules/v3-architecture.md`

---

## D2: 2026-04-12: Implement declarative guardrails with TypeScript core engine #guardrails #typescript

### Conclusion

- Define a declarative rule table (R01-R13) in core/src/guardrails/rules.ts
- Each rule has the structure `{id, toolPattern, evaluate()}` and is evaluated sequentially with priority ordering
- Migration from hooks/ bash scripts to core/ TypeScript engine is complete

### Background

- The bash guardrail (pretooluse-guard.sh, removed in Phase 36) had increasingly complex conditional branching and was difficult to test

### Options

- A: Continue improving the bash script
- B: Migrate to a type-safe declarative rule engine in TypeScript

### Rationale

- Enables individual rule unit testing with vitest (target coverage 90%+)
- Adding a new rule requires only a single table entry
- The stdin -> route -> stdout pipeline integrates simply with hooks.json

### Impact / Trade-offs

- Increases Node.js runtime dependency (no longer runs on bash alone)
- Native builds of better-sqlite3 may be required depending on the environment

### Review Conditions

- If Claude Code natively provides guardrail functionality

### Related

- files: `core/src/guardrails/rules.ts`, `core/src/guardrails/pre-tool.ts`

---

## D3: 2026-04-12: Translate all comments, tests, and output in core/ to English #i18n #core

### Conclusion

- Translated comments, test descriptions, and user-facing output from Japanese to English across all files (17 files) under core/src/
- Synchronized core/dist/ via tsc rebuild
- Guard rule reason/systemMessage values also translated to English (e.g., "Warning" instead of Japanese equivalent)

### Background

- As a public OSS repository, internationalization of the core engine was needed
- Test assertions depended on Japanese strings, requiring synchronization when output language changed

### Options

- A: Keep core/ in Japanese and only translate documentation to English
- B: Translate all of core/ to English (code, tests, output)

### Rationale

- The core engine should be language-agnostic
- Test assertions and actual output must use the same language or tests break
- Makes it easier for international contributors to participate

### Impact / Trade-offs

- CLAUDE.md now mandates English for all code, comments, documentation, and communication
- Potential inconsistency if skill-side code expects Japanese guardrail output

### Review Conditions

- If an i18n framework is introduced for runtime language switching

### Related

- session: 2026-04-12 translation session

---

## D4: 2026-04-12: Manage better-sqlite3 as optionalDependencies #dependencies #node24

### Conclusion

- Place better-sqlite3 in `optionalDependencies` in core/package.json
- Design tolerates native module build failures for Node 24 compatibility

### Background

- Node 24 changed some native module ABIs, causing build failures in certain cases

### Rationale

- Prevents guardrail evaluation (which does not need SQLite) from being blocked by native build failures
- SQLite-related features use graceful degradation

### Review Conditions

- If better-sqlite3 officially supports Node 24

---

## D5: 2026-04-12: Unify Codex integration through official plugin #codex #policy

### Conclusion

- Direct calls to raw `codex exec` are prohibited
- Calls must go through `scripts/codex-companion.sh` or `/codex:*` commands
- MCP server (`mcp__codex__*`) is deprecated and blocked via deny settings

### Background

- Direct calls bypass permission controls and review gates

### Rationale

- The official plugin provides job management, structured output, and Stop Review Gate
- The companion script makes Harness-specific workflow integration straightforward

### Related

- rules: `.claude/rules/codex-cli-only.md`

---

## D6: 2026-04-12: Require implementation or category classification for Feature Table additions #quality #policy

### Conclusion

- New rows added to the Feature Table must be classified into one of 3 categories: (A) Has implementation / (B) Written only / (C) CC auto-inherited
- Category B blocks PR merge and requires an implementation proposal

### Background

- "Written only" entries accumulated in the Feature Table, and the gap from reality undermined its trustworthiness

### Rationale

- An institutional guardrail to maintain an "honest Feature Table"

### Related

- rules: `.claude/rules/cc-update-policy.md`

---

## D7: 2026-04-12: Keep hooks/ as thin shims, delegate logic to Go binary #architecture #hooks

> _(Updated 2026-04-13: core/ TypeScript → go/ Go binary. See D9.)_

### Conclusion

- Bash scripts in hooks/ are thin shims that call `"${CLAUDE_PLUGIN_ROOT}/bin/harness" hook <name>`
- All logic (guardrail evaluation, permission handling, session management) is centralized in the Go binary at `go/internal/`

### Background

- Writing logic in bash makes testing difficult and leads to complex conditional branching
- TypeScript core/ was the original centralization layer (D2), superseded by Go binary (D9)

### Rationale

- A single entry point makes testing and debugging straightforward
- Logic can be updated without changing hooks.json configuration
- Go binary is statically compiled — no runtime dependencies

### Review Conditions

- If Claude Code makes hooks execution environment natively Go or provides a native plugin SDK

---

## D8: 2026-04-12: Retire OpenCode platform, keep Codex via symlinks #architecture #cleanup

### Conclusion

- OpenCode platform fully retired (all scripts, workflows, and references removed)
- Codex CLI integration preserved using symlinks: `codex/.codex/skills/` -> `../../../skills/`
- Pre-consolidation agents removed (8 agents superseded by v3 consolidation)
- Unwired shell scripts removed (pretooluse-guard.sh, stop-* scripts, security/tampering detectors, mirror sync)
- Total: ~5,364 lines removed across 43 files

### Background

- Exploration revealed ~8,800+ lines of dead code accumulated from OpenCode retirement, v3 agent consolidation, and TypeScript migration
- OpenCode was never widely adopted; Codex CLI remained the active secondary platform

### Options

- A: Remove everything including Codex (minimal maintenance)
- B: Keep Codex via symlinks, remove everything else (zero duplication, Codex stays usable)
- C: Keep both Codex and OpenCode (high maintenance cost for unused platform)

### Rationale

- Option B chosen: symlinks create zero duplication (single SSOT in skills/), zero maintenance overhead
- Git history preserves all removed code for future recovery if needed
- Pre-consolidation agents were already superseded by worker.md, reviewer.md, scaffolder.md

### Impact / Trade-offs

- Codex CLI can still be used as a secondary worker
- OpenCode users (if any) would need to re-setup after pulling this change
- error-recovery.md was initially removed but restored as it is still actively referenced by worker.md and team-composition.md

### Review Conditions

- If OpenCode is revived, restore from git history
- If Codex CLI is deprecated upstream, remove codex/ directory

---

## D9: 2026-04-13: Migrate guardrail engine from TypeScript to Go native binary #guardrails #go

### Conclusion

- Guardrail engine moved from `core/src/` (TypeScript) to `go/internal/guardrail/` (Go)
- Compiled binary at `bin/harness-<os>-<arch>`, dispatched by `bin/harness` wrapper script
- hooks/ shims now call `"${CLAUDE_PLUGIN_ROOT}/bin/harness" hook <name>` instead of Node.js

### Background

- v4.0.0 "Hokage" release completed the TypeScript → Go migration
- Go binary is CGO_ENABLED=0 (static), no runtime dependencies

### Rationale

- No Node.js runtime required at install time
- Faster hook execution (no interpreter startup)
- Cross-compiled for darwin-arm64, darwin-amd64, linux-amd64 via CI

### Related

- files: `go/cmd/harness/`, `go/internal/guardrail/rules.go`, `bin/harness`
- rules: `.claude/rules/v3-architecture.md` (historical TypeScript reference)

---

## D10: 2026-04-13: Hooks fail-open (exit 0) when platform binary is missing #hooks #ux

### Conclusion

- `bin/harness` exits 0 (approve/no-op) when the platform binary is not installed
- Warns **once** to stderr: `[harness] binary not installed — guardrails inactive. Run /harness-setup binary to install.`
- Warning suppressed after first occurrence via flag file `~/.claude/harness-binary-missing.warned`
- Flag is cleared when binary is successfully installed (by download-binary.sh)
- Previously exited 1, causing `UserPromptSubmit hook error` on every prompt for new users

### Background

- Binary is gitignored (built artifact); new installs had no binary until manually downloaded
- Exit 1 from any hook surfaces as a visible error in Claude Code UI
- Silent exit 0 (original fix) was invisible — user had no indication guardrails were inactive

### Rationale

- Fail-open is safer UX: hooks that can't run should not block the user
- One-time warning gives visibility without becoming noise on every prompt
- The guardrails are a hardening layer, not a blocker — absence degrades gracefully

### Review Conditions

- If Claude Code adds a hook health-check mechanism that can surface "binary missing" as a warning instead of error

---

## D11: 2026-04-13: Auto-download platform binary on plugin install via Setup hook #distribution #binary

### Conclusion

- `skills/harness-setup/scripts/download-binary.sh` runs as first step of `Setup: init` hook
- Pure POSIX shell — no Go binary needed to run the downloader
- Fetches latest release tag from GitHub API, downloads correct `harness-<os>-<arch>` asset
- Exits 0 on any failure so install never blocks

### Background

- Binaries are release assets (not committed to git), so fresh installs had no binary
- Previously required manual `/harness-setup binary` after install

### Rationale

- Zero-friction install: binary is ready before the first prompt
- Shell-only downloader avoids chicken-and-egg (can't use the binary to download itself)

### Review Conditions

- If Claude Code marketplace supports bundling release assets directly

---

## D12: 2026-04-15: Makefile as stable CI interface layer between workflows and script paths #ci #architecture

### Conclusion

- All CI workflow steps call `make <target>` instead of direct script paths (e.g. `make validate` not `bash ./tests/validate-plugin.sh`)
- The Makefile is the single point of update when script paths change; CI workflows stay untouched
- Dev-only scripts live in `local-scripts/` (not plugin-distributed); plugin-distributed scripts stay in `harness/scripts/`

### Background

- Phase 52 restructure moved 780+ files into `harness/`. CI steps referencing raw script paths broke silently (wrong paths not caught until CI ran).
- `local-scripts/` (repo dev tooling) vs `harness/scripts/` (plugin scripts shipped to users) needed a clean separation.

### Rationale

- Indirection via `make` means path changes require one edit (Makefile) not N edits across workflow files
- The `local-scripts/` vs `harness/scripts/` split makes the distribution boundary visible in directory structure

### Impact / Trade-offs

- Contributors must have `make` installed (standard on macOS/Linux; available via Git for Windows)
- Adding a new dev script requires a Makefile target, not just a script file

### Review Conditions

- If CI platform moves away from make

---

## D13: 2026-04-15: MARKETPLACE_NAME and PLUGIN_NAME in cache scripts must match marketplace.json #distribution #cache

### Conclusion

- `harness/scripts/sync-plugin-cache.sh` uses `MARKETPLACE_NAME="powerball-harness-marketplace"` and `PLUGIN_NAME="harness"` — matching `.claude-plugin/marketplace.json` top-level `name` and the plugin entry's `name` respectively
- Cache directory path: `~/.claude/plugins/cache/<MARKETPLACE_NAME>/<PLUGIN_NAME>/<VERSION>/`

### Background

- Script shipped with `MARKETPLACE_NAME="claude-code-harness-marketplace"` and `PLUGIN_NAME="claude-code-harness"` — the old pre-marketplace monolithic names. Step 2 wrote files to a non-existent cache path; script exited 0, making the failure invisible.

### Rationale

- CC derives the cache path from the identifiers in `marketplace.json`. Any mismatch means synced files land where CC never reads.
- Verified from `~/.claude/plugins/cache/` directory structure at review time.

### Review Conditions

- If CC changes cache directory structure or naming conventions

---

## D14: 2026-04-15: Consistency check sections must be explicitly skipped, never silently no-op #quality #ci

### Conclusion

- Any section of `check-consistency.sh` (or similar gate scripts) that iterates over a path must guard with a directory existence check and print an explicit "skipped" message if the path does not exist
- A section that prints ✅ regardless of actual state is indistinguishable from a passing check and destroys trust in CI output

### Background

- `check-consistency.sh` [2/13] iterated `for cmd in "$PLUGIN_ROOT/commands"/*.md` — `commands/` was removed in v2.17.0 (migrated to skills). With nullglob off, bash loops once with the unexpanded glob string, `grep` silently finds nothing, section prints ✅. This was undetected until the v4.2.0 → HEAD review.

### Rationale

- Silent no-ops accumulate over time as directories are renamed or deleted. Explicit "skipped" messages surface these drift events and keep the check list honest.
- Applied fix: `if [ -d "$PLUGIN_ROOT/commands" ]; then ... else echo "skipped (commands/ removed)"; fi`

### Review Conditions

- Pattern should be applied proactively whenever a check section's target path changes

---

## D15: 2026-04-17: Concurrent hook fan-out and ScheduleWakeup-based harness-loop runtime #architecture #hooks #breezing

### Conclusion

- `PostToolUse` and `PreToolUse` hooks use goroutine fan-out (`post-tool-batch`, `pre-tool-batch`) to parallelize subprocess invocations — 9 sequential forks reduced to 2 concurrent via `post-tool-batch`
- `harness-loop` graduated from a basic loop to a full ScheduleWakeup-based autonomous runtime with `--max-cycles`, `--pacing`, flock guard, sprint-contracts, and plateau detection
- Sprint Contract defined as a Go package (`go/internal/sprint`) — structured commitment between Planner and Worker with acceptance criteria

### Background

- Phase 63: Hook chain was blocking on sequential subprocess calls; fan-out eliminates the bottleneck
- Phase 69: harness-loop needed to be autonomous enough to run unattended across multiple Claude Code sessions using ScheduleWakeup events

### Rationale

- Fan-out is safe for PostToolUse (side effects, not blocking decisions); deny-wins merge semantics for PreToolUse ensures safety is preserved
- ScheduleWakeup gives the loop persistence across session boundaries without requiring the user to stay present

### Impact / Trade-offs

- `flock` guard prevents concurrent harness-loop instances from stepping on each other (Plans.md exclusive access)
- `--pacing` allows rate-limiting to avoid overwhelming external resources during automated runs

### Review Conditions

- If Claude Code provides a native autonomous loop primitive (ScheduleWakeup replacement)
- If fan-out introduces race conditions that surface in CI

### Related

- files: `go/internal/hookhandler/`, `harness/skills/harness-loop/`, `harness/scripts/codex-loop.sh`
