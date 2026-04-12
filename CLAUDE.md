# CLAUDE.md - Claude Harness Development Guide

This file provides guidance for Claude Code when working in this repository.

## Project Overview

**Claude harness** is a plugin for autonomous operation of Claude Code in a "Plan → Work → Review" workflow.

**Special note**: This project is self-referential — it uses the harness itself to improve the harness.

## Claude Code 2.1.79+ Feature Utilization Guide

Harness makes full use of new features introduced in Claude Code 2.1.79.

| Feature | Skill | Purpose |
|---------|-------|---------|
| **Agent Memory (memory frontmatter)** | task-worker, code-reviewer | Persistent learning |
| **TeammateIdle/TaskCompleted Hook** | breezing | Automated team monitoring |
| **Skill budget scaling** | All skills | Auto-adjusts to 2% of context window |
| **Fast mode (Opus 4.6)** | All skills | High-speed output mode |
| **Worktree isolation** | breezing, parallel-workflows | Safe parallel writes to the same file |
| **`/simplify` Auto-Refinement** | work | Automatic code simplification after implementation |
| **HTTP hooks (v2.1.63)** | hooks | JSON POST template provided. Practical use enabled when `HARNESS_WEBHOOK_URL` is set |
| **Effort levels + ultrathink (v2.1.68)** | harness-work | Multi-factor scoring injects ultrathink for complex tasks |
| **Agent hooks (v2.1.68)** | hooks | LLM-based code quality guard (type: "agent") |
| **`${CLAUDE_SKILL_DIR}` variable (v2.1.69)** | all skills | Stable skill-local reference path resolution |
| **InstructionsLoaded hook (v2.1.69)** | hooks | Pre-session instruction load tracking and environment checks |
| **`agent_id` / `agent_type` fields (v2.1.69)** | hooks, breezing | Robust teammate identity and role-aware guarding |
| **`{"continue": false}` teammate response (v2.1.69)** | breezing | Stop team loop when all tasks are completed or stop is requested |
| **`/reload-plugins` (v2.1.69)** | all skills | Immediate reflection after skill/hook edits without restarting |
| **`includeGitInstructions: false` (v2.1.69)** | breezing, work | Reduce prompt token overhead for git-instruction-light tasks |
| **`git-subdir` plugin source (v2.1.69)** | setup, release | Support plugin source managed from repository subdirectories |
| **Sonnet 4.5 → 4.6 auto-migration** | all skills | Legacy Sonnet references migrate to 4.6 behavior automatically |
| **WorktreeCreate/Remove hook (v2.1.50)** | breezing | Worktree lifecycle auto-setup and cleanup |
| **Auto Mode (RP Phase 1)** | breezing, work | CC native feature. Harness only tracks PermissionDenied. Decision logic not implemented |
| **Per-agent hooks (v2.1.69+)** | agents-v3/ | Worker PreToolUse guard + Reviewer Stop log in agent frontmatter |
| **Agent `isolation: worktree` (v2.1.50+)** | agents-v3/worker | Auto worktree isolation for parallel writes with shared Agent Memory |
| **`/loop` + Cron scheduling (v2.1.71)** | breezing, harness-work | Periodic task monitoring with `/loop 5m /sync-status` |
| **PostToolUseFailure hook (v2.1.70)** | hooks | Auto-escalation after 3 consecutive failures |
| **Background Agent output fix (v2.1.71)** | breezing | Safe background agent usage with output path in completion notification |
| **Compaction image retention (v2.1.70)** | all skills | Images preserved during context compaction |
| **Subagent `background` field (v2.1.71+)** | breezing | Always-background agent execution via frontmatter |
| **Subagent `local` memory scope (v2.1.71+)** | agents-v3/ | Non-VCS agent memory in `.claude/agent-memory-local/` |
| **Agent Teams experimental flag (v2.1.71+)** | breezing | `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS` env var for official Agent Teams |
| **`/agents` command (v2.1.71+)** | setup, troubleshoot | Interactive agent management UI (create/edit/delete) |
| **Desktop Scheduled Tasks (v2.1.71+)** | harness-work | CC native feature. No Harness default config (CronCreate tool available) |
| **`--agents` CLI flag (v2.1.71+)** | breezing, CI | Session-level JSON agent definitions without disk persistence |
| **`ExitWorktree` tool (v2.1.72)** | breezing, work | Programmatic worktree exit for agent workflows |
| **Effort levels simplified (v2.1.72)** | harness-work | Persistent levels are `low/medium/high` (`○ ◐ ●`). `max` remains as Opus 4.6 session-only option |
| **Agent tool `model` param restored (v2.1.72)** | breezing | Per-invocation model overrides re-enabled |
| **`/plan` description argument (v2.1.72)** | harness-plan | `/plan fix the auth bug` enters plan mode with context |
| **Parallel tool call fix (v2.1.72)** | breezing, work | Failed Read/WebFetch/Glob no longer cancel sibling calls |
| **Worktree isolation fixes (v2.1.72)** | breezing | Task resume cwd restore + background notification worktreePath |
| **`/clear` preserves background agents (v2.1.72)** | breezing | `/clear` only kills foreground tasks; background agents survive |
| **Hooks fixes (v2.1.72)** | hooks | transcript_path fix, skill hooks double-fire fix, async stdin fix |
| **HTML comments hidden in CLAUDE.md (v2.1.72)** | all | `<!-- -->` hidden from auto-injection; visible via Read tool |
| **Sandboxing (`/sandbox`)** | breezing, work | OS-level filesystem/network isolation complementing bypassPermissions |
| **`opusplan` model alias** | breezing | Auto-switches Opus (plan) ↔ Sonnet (execute) for Lead sessions |
| **`CLAUDE_CODE_SUBAGENT_MODEL` env var** | breezing, work | Centralized subagent model control for Worker/Reviewer |
| **Checkpointing (`/rewind`)** | work | Session state tracking, rewind, and selective summarization |
| **Code Review (managed, RP)** | harness-review | Multi-agent PR review with `REVIEW.md` guidance. Teams/Enterprise |
| **Status Line (`/statusline`)** | all skills | Custom shell-script status bar for context/cost/git monitoring |
| **1M Context (`sonnet[1m]`)** | harness-review, breezing | 1M token context window for large codebase analysis |
| **Chrome Integration (`--chrome`, beta)** | harness-work, harness-review | Browser automation for UI testing, console debugging, data extraction |
| **`modelOverrides` setting (v2.1.73)** | setup, breezing | Map model picker entries to custom provider model IDs (Bedrock ARNs, etc.) |
| **`/output-style` deprecated (v2.1.73)** | all skills | Use `/config` instead; output style selection moved to config menu |
| **Bedrock/Vertex Opus 4.6 default (v2.1.73)** | breezing | Default Opus on cloud providers updated from 4.1 to 4.6 |
| **`autoMemoryDirectory` setting (v2.1.74)** | session-memory, setup | Custom auto-memory storage path for project-specific memory isolation |
| **`CLAUDE_CODE_SESSIONEND_HOOKS_TIMEOUT_MS` (v2.1.74)** | hooks | Configurable SessionEnd hooks timeout (was fixed 1.5s kill) |
| **Full model ID fix (v2.1.74)** | agents-v3/, breezing | `claude-opus-4-6` etc. now recognized in agent frontmatter and JSON config |
| **Streaming API memory leak fix (v2.1.74)** | breezing, work | Unbounded RSS growth in streaming response buffers fixed |
| **LSP server integration (`.lsp.json`)** | setup | CC native feature. No Harness default `.lsp.json` config |
| **`SubagentStart`/`SubagentStop` matcher** | breezing, hooks | Agent type-specific lifecycle monitoring with matcher filtering |
| **Agent Teams: Task Dependencies** | breezing | Auto-unblocking dependent tasks with file-lock claiming |
| **`--teammate-mode` CLI flag** | breezing | Per-session display mode override (`in-process`/`tmux`) |
| **`skills` field in agent frontmatter** | agents-v3/ | Preload skill content into subagent context at startup |
| **`--remote` / Cloud Sessions** | breezing, harness-work | Terminal-to-cloud async task execution with `/teleport` retrieval |
| **`CLAUDE_ENV_FILE` SessionStart persistence** | hooks | Persist env vars from SessionStart hooks to subsequent Bash commands |
| **`PreCompact` hook** | hooks | Pre-compaction state save + WIP task warning (implemented) |
| **Slack Integration (`@Claude`)** | — | Future support (requires Teams/Enterprise). No Harness implementation |
| **Analytics Dashboard** | — | Planned (not implemented). PR attribution, metrics, leaderboard for future |
| **OpenTelemetry Monitoring** | hooks, breezing | Custom JSONL trace output. OTel format conversion only when `OTEL_EXPORTER_OTLP_ENDPOINT` is set (planned) |
| **`/security-review` command** | harness-review | Analyze pending changes for security vulnerabilities (injection, auth, data exposure) |
| **`/insights` command** | session-memory | Session analysis report: project areas, interaction patterns, friction points |
| **`/stats` command** | session | Daily usage visualization, session history, streaks, model preferences |
| **Prompt Suggestions** | all skills | Git-history-based context-aware autocomplete; Tab to accept, Enter to submit |
| **PR Review Status footer** | breezing, harness-review | Clickable PR link with color-coded review status (green/yellow/red/gray/purple) |
| **`CLAUDE_CODE_TASK_LIST_ID` env var** | breezing | Named task list sharing across sessions: `CLAUDE_CODE_TASK_LIST_ID=my-project claude` |
| **`fastModePerSessionOptIn` setting** | setup, breezing | Admin control: fast mode resets each session, users must `/fast` to re-enable |
| **1M Context Window (`opus[1m]`) (v2.1.75)** | breezing, harness-review | Opus 4.6 1M context window. Auto-upgrade on Max/Team/Enterprise |
| **Memory file timestamps (v2.1.75)** | session-memory, memory | Last-updated timestamps on memory files. Supports freshness-based memory decisions |
| **Async hook suppression (v2.1.75)** | breezing, hooks | Async hook completion messages hidden by default. Show with `--verbose` |
| **`/effort max` session-only (v2.1.75+)** | harness-work, harness-plan | Opus 4.6 only deepest reasoning mode. Per-session activation, not persisted |
| **MCP Elicitation support (v2.1.76)** | hooks, breezing | Structured input requests from MCP servers. Auto-skipped in Breezing |
| **`Elicitation`/`ElicitationResult` hooks (v2.1.76)** | hooks | Intercept and log before/after MCP elicitation |
| **`PostCompact` hook (v2.1.76)** | hooks, breezing | Post-compaction context re-injection (counterpart to PreCompact) |
| **`-n`/`--name` CLI flag (v2.1.76)** | breezing | Set session display name. Used for identification in session list |
| **`worktree.sparsePaths` setting (v2.1.76)** | breezing, setup | Worktree sparse-checkout for monorepos. Faster parallel worker startup |
| **`/effort` slash command (v2.1.76)** | harness-work | Switch effort level during session (low/medium/high) |
| **`--worktree` faster startup (v2.1.76)** | breezing | Direct git refs read + skip redundant fetch |
| **Background agent partial result retention (v2.1.76)** | breezing | Partial results preserved in context even on kill |
| **Stale worktree auto-cleanup (v2.1.76)** | breezing | Auto-delete worktrees from interrupted parallel executions |
| **Auto-compaction circuit breaker (v2.1.76)** | all skills | Auto-stop after 3 consecutive failures (prevents infinite retry) |
| **`--plugin-dir` spec change (v2.1.76, breaking)** | setup | Multiple directories specified by repeating `--plugin-dir` |
| **Deferred Tools schema fix (v2.1.76)** | all skills | ToolSearch tool schema retained after compaction |
| **`/context` command (v2.1.74)** | all skills | Visualize context consumption and suggest optimizations. Prevent bloat in long sessions |
| **`maxTurns` agent safety limit** | agents-v3/ | Worker: 100, Reviewer: 50, Scaffolder: 75. Safety valve to prevent runaway |
| **`Notification` hook implementation** | hooks | Log notification events (permission_prompt, idle_prompt, etc.). Improves Breezing observability |
| **Output token limits 64k/128k (v2.1.77)** | all skills | Opus 4.6 / Sonnet 4.6 default output 64k, max 128k tokens |
| **`allowRead` sandbox setting (v2.1.77)** | harness-review | Re-allow read access to specific paths within `denyRead` regions |
| **PreToolUse `allow` respects `deny` (v2.1.77)** | guardrails | Hook `allow` no longer overrides settings.json `deny` rules (security hardening) |
| **Agent `resume` → `SendMessage` (v2.1.77)** | breezing | Agent tool `resume` parameter deprecated. Migrated to `SendMessage({to: agentId})` |
| **`/branch` (was `/fork`) (v2.1.77)** | session | `/fork` renamed to `/branch` (`/fork` remains as alias) |
| **`claude plugin validate` enhanced (v2.1.77)** | setup | Added frontmatter + hooks.json syntax validation |
| **`--resume` 45% faster (v2.1.77)** | session | Fork-heavy session resume up to 45% faster, 100-150MB memory reduction |
| **Stale worktree race fix (v2.1.77)** | breezing | Fixed race condition where active agent worktrees were incorrectly deleted |
| **`StopFailure` hook event (v2.1.78)** | hooks | Capture session stop failures on API errors (rate limits, auth failures) |
| **`${CLAUDE_PLUGIN_DATA}` variable (v2.1.78)** | hooks, setup | Persistent state directory variable that survives plugin updates |
| **Agent `effort`/`maxTurns`/`disallowedTools` frontmatter (v2.1.78)** | agents-v3/ | Declaratively set effort, turn limits, and tool restrictions in plugin agent definitions |
| **`deny: ["mcp__*"]` permission fix (v2.1.78)** | setup | settings.json deny rules now correctly block MCP tools |
| **`ANTHROPIC_CUSTOM_MODEL_OPTION` env var (v2.1.78)** | setup | Add custom model entries to the `/model` picker |
| **`--worktree` skills/hooks loading fix (v2.1.78)** | breezing | Skills and hooks now load correctly when using the worktree flag |
| **Large session truncation fix (v2.1.78)** | session | Fixed `cc log` / `--resume` truncating sessions over 5MB |
| **`--console` auth flag (v2.1.79)** | setup | `claude auth login --console` for Anthropic Console API billing auth |
| **Turn duration toggle (v2.1.79)** | all skills | Toggle turn execution time display in `/config` |
| **`CLAUDE_CODE_PLUGIN_SEED_DIR` multiple dirs (v2.1.79)** | setup | Specify multiple seed directories with platform path separator |
| **SessionEnd hooks fix in `/resume` (v2.1.79)** | hooks | SessionEnd hooks now fire correctly during interactive `/resume` session switching |
| **18MB startup memory reduction (v2.1.79)** | all skills | Reduced startup memory usage by ~18MB |
| **Hooks conditional `if` field (v2.1.85)** | hooks, guardrails | Scope `PermissionRequest` to safe Bash and edit tools only, reducing unnecessary hook invocations |
| **`TaskCreated` hook blocking (v2.1.84)** | hooks | Hooks fire synchronously on `TaskCreate`. Used for runtime-reactive workflows |
| **Rules `paths:` YAML list (v2.1.84)** | setup | Rule definition `paths:` now supports YAML list format. Structured multi-glob support |
| **MCP tool description cap 2KB (v2.1.84)** | all skills | Prevent context bloat from large OpenAPI-derived MCP schemas |
| **`PermissionDenied` hook event (v2.1.89)** | hooks, breezing | Track auto mode denials and notify Breezing Lead. `{retry:true}` for retry guidance |
| **`"defer"` permission decision (v2.1.89)** | hooks, breezing | Pause headless sessions for Lead judgment, then re-evaluate on resume |
| **Hook output >50K disk save (v2.1.89)** | hooks | Large hook output saved to disk with preview instead of direct context injection |
| **Hooks `if` compound command fix (v2.1.89)** | hooks | Compound commands like `ls && git push` now match `if` conditions correctly |
| **Autocompact thrash loop fix (v2.1.89)** | all skills | Stops after 3 consecutive compact-then-refill cycles. CC auto-inherited |
| **Nested CLAUDE.md re-injection fix (v2.1.89)** | all skills | Fixed CLAUDE.md duplicate injection in long sessions. CC auto-inherited |
| **PreToolUse exit 2 JSON fix (v2.1.90)** | hooks, guardrails | Fixed JSON stdout + exit 2 blocking behavior. pre-tool.sh deny now more reliable |
| **PostToolUse format-on-save fix (v2.1.90)** | hooks | Fixed Edit/Write failures after hook-triggered file rewrites. CC auto-inherited |
| **`--resume` prompt-cache miss fix (v2.1.90)** | session | Regression since v2.1.69. Fixed cache miss on resume. CC auto-inherited |
| **SSE/transcript performance (v2.1.90)** | all skills | SSE frame processing O(n²)→O(n), faster transcript writes. CC auto-inherited |

Full details: [docs/CLAUDE-feature-table.md](docs/CLAUDE-feature-table.md)

## Development Rules

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/): `feat:` / `fix:` / `docs:` / `refactor:` / `test:` / `chore:`

### Version Management

Keep `VERSION` and `.claude-plugin/plugin.json` in sync.
Normal feature/docs PRs must leave both files unchanged and record changes under `CHANGELOG.md`'s `[Unreleased]` section.
Use `./scripts/sync-version.sh bump` only when cutting a release.

### CHANGELOG

Details: [.claude/rules/changelog.md](.claude/rules/changelog.md) (Keep a Changelog format; include Before/After tables for major changes)

### Language

All responses must be in **English** (including `context: fork` skills).

### Code Style

- Use clear and descriptive names
- Add comments for complex logic
- Keep commands/agents/skills single-responsibility

## Repository Structure

`.claude-plugin/` Plugin manifest / `agents/` Sub-agents / `skills/` Skills / `hooks/` Hooks / `scripts/` Shell scripts / `docs/` Documentation / `tests/` Validation

## Using Skills (Important)

**Before starting work:** If a relevant skill exists, launch it with the Skill tool first.

> For heavy tasks, skills spawn sub-agents from `agents/` in parallel via the Task tool.

### Top Skill Categories (Top 5)

| Category | Purpose | Trigger Examples |
|---------|---------|-----------------|
| work | Task implementation (auto-scope detection, --codex support) | "implement", "do it all", "/work" |
| breezing | Full auto-run with Agent Teams (--codex support) | "run with team", "breezing" |
| harness-review | Code review, quality checks | "review", "security", "performance" |
| setup | Setup integration hub (init, harness-mem, Codex CLI, etc.) | "setup", "initialize", "harness-mem", "codex-setup" |
| memory | SSOT management, memory search, SSOT promotion | "SSOT", "decisions.md", "memory search", "claude-mem" |

Full category list and hierarchy: [docs/CLAUDE-skill-catalog.md](docs/CLAUDE-skill-catalog.md)

## Development Flow

0. **When editing skills/hooks**: run `/reload-plugins` to refresh runtime cache immediately
1. **Plan**: Use `/plan-with-agent` to add tasks to Plans.md
2. **Implement**: `/work` (Claude implements) or `/breezing` (team full-run). Both support `--codex`
3. **Review**: Runs automatically (manual: `/harness-review`)
4. **Validate**: Run `./tests/validate-plugin.sh` for structural validation

## Testing

```bash
./tests/validate-plugin.sh          # Validate plugin structure
./scripts/ci/check-consistency.sh   # Consistency check
```

Details: [docs/CLAUDE-commands.md](docs/CLAUDE-commands.md)

## Notes

- **Watch for self-reference**: Running `/work` on this plugin means editing its own code
- **Hooks run automatically**: PreToolUse/PostToolUse guards are active
- **VERSION sync**: Leave version files untouched in normal PRs; update them only for releases

## Key Commands (for development)

| Command | Purpose |
|---------|---------|
| `/plan-with-agent` | Add improvement tasks to Plans.md |
| `/work` | Implement tasks (auto-scope detection, --codex support) |
| `/breezing` | Full team parallel run with Agent Teams (--codex support) |
| `/harness-review` | Review changes |
| `/validate` | Validate plugin |
| `/remember` | Record learnings |

Details & handoff: [docs/CLAUDE-commands.md](docs/CLAUDE-commands.md)

## SSOT (Single Source of Truth)

- `.claude/memory/decisions.md` - Decisions (Why)
- `.claude/memory/patterns.md` - Reusable patterns (How)

## Test Tampering Prevention

> **Absolutely prohibited**: Tampering with tests to fake "success"

Details: [.claude/rules/test-quality.md](.claude/rules/test-quality.md) / [.claude/rules/implementation-quality.md](.claude/rules/implementation-quality.md)
