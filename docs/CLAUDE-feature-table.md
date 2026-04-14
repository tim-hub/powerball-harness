# Claude Code 2.1.74+ Feature Utilization Guide (Complete Edition)

> **Overview**: Complete listing of all Claude Code 2.1.74+ features utilized by Harness.
> Full version of the CLAUDE.md Feature Table (with detailed descriptions).


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
| **Per-agent hooks (v2.1.69+)** | agents/ | Worker PreToolUse guard + Reviewer Stop log in agent frontmatter |
| **Agent `isolation: worktree` (v2.1.50+)** | agents/worker | Auto worktree isolation for parallel writes with shared Agent Memory |
| **`/loop` + Cron scheduling (v2.1.71)** | breezing, harness-work | Periodic task monitoring with `/loop 5m /sync-status` |
| **PostToolUseFailure hook (v2.1.70)** | hooks | Auto-escalation after 3 consecutive failures |
| **Background Agent output fix (v2.1.71)** | breezing | Safe background agent usage with output path in completion notification |
| **Compaction image retention (v2.1.70)** | all skills | Images preserved during context compaction |
| **Subagent `background` field (v2.1.71+)** | breezing | Always-background agent execution via frontmatter |
| **Subagent `local` memory scope (v2.1.71+)** | agents/ | Non-VCS agent memory in `.claude/agent-memory-local/` |
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
| **Full model ID fix (v2.1.74)** | agents/, breezing | `claude-opus-4-6` etc. now recognized in agent frontmatter and JSON config |
| **Streaming API memory leak fix (v2.1.74)** | breezing, work | Unbounded RSS growth in streaming response buffers fixed |
| **LSP server integration (`.lsp.json`)** | setup | CC native feature. No Harness default `.lsp.json` config |
| **`SubagentStart`/`SubagentStop` matcher** | breezing, hooks | Agent type-specific lifecycle monitoring with matcher filtering |
| **Agent Teams: Task Dependencies** | breezing | Auto-unblocking dependent tasks with file-lock claiming |
| **`--teammate-mode` CLI flag** | breezing | Per-session display mode override (`in-process`/`tmux`) |
| **`skills` field in agent frontmatter** | agents/ | Preload skill content into subagent context at startup |
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
| **`maxTurns` agent safety limit** | agents/ | Worker: 100, Reviewer: 50, Scaffolder: 75. Safety valve to prevent runaway |
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
| **Agent `effort`/`maxTurns`/`disallowedTools` frontmatter (v2.1.78)** | agents/ | Declaratively set effort, turn limits, and tool restrictions in plugin agent definitions |
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



## Details Feature List

| Feature | Utilized Skills | Purpose |
|------|-----------|------|
| **Task tool metrics** | parallel-workflows | Aggregate sub-agent token/tool/time metrics |
| **`/debug` command** | troubleshoot | Diagnose complex session issues |
| **PDF page ranges** | notebook-lm, harness-review | Efficient processing of large documents |
| **Git log flags** | harness-review, CI, harness-release | Structured commit analysis |
| **OAuth authentication** | codex-review | Configure MCP servers without DCR support |
| **68% memory optimization** | session-memory, session | Active use of `--resume` |
| **Sub-agent MCP** | task-worker | Share MCP tools during parallel execution |
| **Reduced Motion** | harness-ui | Accessibility settings |
| **TeammateIdle/TaskCompleted Hook** | breezing | Automated team monitoring |
| **Agent Memory (memory frontmatter)** | task-worker, code-reviewer | Persistent learning |
| **Fast mode (Opus 4.6)** | all skills | High-speed output mode |
| **Auto memory recording** | session-memory | Automatic persistence of cross-session knowledge |
| **Skill budget scaling** | all skills | Auto-adjusts to 2% of context window |
| **Task(agent_type) restrictions** | agents/ | Sub-agent type restrictions |
| **Plugin settings.json** | setup | Reduce init tokens and provide immediate security protection |
| **Worktree isolation** | breezing, parallel-workflows | Safe parallel writes to the same file |
| **Background agents** | generate-video | Async scene generation |
| **ConfigChange hook** | hooks | Configuration change auditing |
| **last_assistant_message** | session-memory | Session quality evaluation |
| **Sonnet 4.6 (1M context)** | all skills | Large-scale context processing |
| **Memory leak fixes (v2.1.50-v2.1.63)** | breezing, work | Improved stability for long team sessions |
| **`claude agents` CLI (v2.1.50)** | troubleshoot | Agent definition diagnosis and verification |
| **WorktreeCreate/Remove hook (v2.1.50)** | breezing | Worktree lifecycle auto-setup and cleanup (implemented) |
| **`claude remote-control` (v2.1.51)** | investigated, future support | External builds and local environment serving |
| **`/simplify` (v2.1.63)** | work | Phase 3.5 Auto-Refinement: automatic code refinement after implementation |
| **`/batch` (v2.1.63)** | breezing | Parallel migration delegation for cross-cutting tasks |
| **`code-simplifier` plugin** | work | Deep refactoring with `--deep-simplify` |
| **HTTP hooks (v2.1.63)** | hooks | JSON POST template provided. TaskCompleted notification enabled when `HARNESS_WEBHOOK_URL` is set |
| **Auto-memory worktree sharing (v2.1.63)** | breezing | Memory sharing between worktree agents |
| **`/clear` skill cache reset (v2.1.63)** | troubleshoot | Diagnose cache issues during skill development |
| **`ENABLE_CLAUDEAI_MCP_SERVERS` (v2.1.63)** | setup | Option to disable claude.ai MCP servers |
| **Effort levels + ultrathink (v2.1.68)** | harness-work | Multi-factor scoring auto-injects ultrathink for complex tasks |
| **Agent hooks (v2.1.68)** | hooks | LLM agent-based code quality guard via type: "agent" |
| **Opus 4/4.1 removal (v2.1.68)** | — | Removed from first-party API. Auto-migration to Opus 4.6 |
| **`${CLAUDE_SKILL_DIR}` variable (v2.1.69)** | all skills | Stable skill-local reference path resolution |
| **InstructionsLoaded hook (v2.1.69)** | hooks | Track pre-session instructions load events |
| **`agent_id` / `agent_type` addition (v2.1.69)** | hooks, breezing | Stable teammate identity and role-aware guarding |
| **`{"continue": false}` teammate response (v2.1.69)** | breezing | Enable automatic stop when all tasks are completed |
| **`/reload-plugins` (v2.1.69)** | all skills | Immediate reflection after skill/hook edits |
| **`includeGitInstructions: false` (v2.1.69)** | work, breezing | Token reduction for tasks that don't need git instructions |
| **`git-subdir` plugin source (v2.1.69)** | setup, release | Support plugin source managed from repository subdirectories |
| **Auto Mode (RP Phase 1)** | breezing, work | CC native feature. Harness only tracks PermissionDenied. Decision logic not implemented. Current default is `bypassPermissions` |
| **Per-agent hooks (v2.1.69+)** | agents/ | Added `hooks` field to agent definition frontmatter. Worker gets PreToolUse guard, Reviewer gets Stop log |
| **Agent `isolation: worktree` (v2.1.50+)** | agents/worker | Added `isolation: worktree` to Worker agent definition. Auto worktree isolation for parallel writes |
| **Compaction image retention (v2.1.70)** | notebook-lm, harness-review | Images retained in summary requests. Improved prompt cache reuse |
| **Sub-agent final report simplification (v2.1.70)** | breezing, harness-work | Reduced token consumption for sub-agent completion reports |
| **`--resume` skill list re-injection removed (v2.1.70)** | session | ~600 tokens saved on session resume |
| **Plugin hooks fixes (v2.1.70)** | hooks | Stop/SessionEnd fire after /plugin, template collision resolved, WorktreeCreate/Remove working correctly |
| **Teammate nesting prevention additional fix (v2.1.70)** | breezing | Additional nesting prevention fix beyond v2.1.69 |
| **PostToolUseFailure hook (v2.1.70)** | hooks | New hook event that fires on tool call failures |
| **`/loop` + Cron scheduling (v2.1.71)** | breezing, harness-work | `/loop 5m <prompt>` for periodic execution. Used for automatic task progress monitoring |
| **Background Agent output path fix (v2.1.71)** | breezing, parallel-workflows | Completion notification includes output file path. Results recoverable even after compaction |
| **`--print` team agent hang fix (v2.1.71)** | CI integration | Fixed team agent hang in `--print` mode |
| **Plugin install parallel execution fix (v2.1.71)** | breezing | Stabilized plugin state during multiple simultaneous instances |
| **Marketplace improvements (v2.1.71)** | setup | @ref parser fix, update merge conflict fix, MCP server deduplication, /plugin uninstall uses settings.local.json |
| **Subagent `background` field (v2.1.71+)** | breezing, parallel-workflows | Added `background: true` to agent definitions. Always runs as background task |
| **Subagent `local` memory scope (v2.1.71+)** | agents/ | `memory: local` saves to `.claude/agent-memory-local/`. Isolates sensitive learning that shouldn't be committed to VCS |
| **Agent Teams experimental flag (v2.1.71+)** | breezing | `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS` env var enables Agent Teams. Officially documented |
| **`/agents` command (v2.1.71+)** | troubleshoot, setup | Interactive agent management UI. Create, edit, delete, and list via GUI |
| **Desktop Scheduled Tasks (v2.1.71+)** | harness-work | CC native feature. No Harness default config (CronCreate tool available) |
| **`CronCreate/CronList/CronDelete` tools (v2.1.71+)** | breezing, harness-work | Internal tools for `/loop`. Create and manage periodic tasks within a session |
| **`CLAUDE_CODE_DISABLE_CRON` env var (v2.1.71+)** | setup | `=1` disables Cron scheduler. For environments where security policy restricts periodic execution |
| **`--agents` CLI flag (v2.1.71+)** | breezing, CI | Pass agent definitions as JSON at session level. Temporary config not persisted to disk |
| **`ExitWorktree` tool (v2.1.72)** | breezing, harness-work | Tool to programmatically exit a worktree session |
| **Effort levels simplification (v2.1.72)** | harness-work | `max` removed, 3 levels `low/medium/high` + `○ ◐ ●` symbols. `/effort auto` resets to default |
| **Agent tool `model` parameter restored (v2.1.72)** | breezing | Per-invocation model override re-enabled |
| **`/plan` description argument (v2.1.72)** | harness-plan | Enter plan mode with description like `/plan fix the auth bug` |
| **Parallel tool call fix (v2.1.72)** | breezing, harness-work | Failed Read/WebFetch/Glob no longer cancel sibling calls (only Bash errors cascade) |
| **Worktree isolation fixes (v2.1.72)** | breezing | Task resume cwd restore, background notification includes worktreePath |
| **`/clear` preserves background agents (v2.1.72)** | breezing | `/clear` only stops foreground tasks. Background agents survive |
| **Hooks fixes (v2.1.72)** | hooks | transcript_path fix, PostToolUse double display fix, async hooks stdin fix, skill hooks double-fire fix |
| **HTML comments hidden (v2.1.72)** | all skills | `<!-- -->` in CLAUDE.md hidden from auto-injection. Still visible via Read tool |
| **Bash auto-approval additions (v2.1.72)** | guardrails | `lsof`, `pgrep`, `tput`, `ss`, `fd`, `fdfind` added to allow list |
| **Prompt cache fix (v2.1.72)** | all skills | Fixed SDK `query()` cache invalidation. Up to 12x input token cost reduction |
| **Output Styles (v2.1.72+)** | all skills | Define custom output styles in `.claude/output-styles/`. `harness-ops` provides structured output for Plan/Work/Review |
| **`permissionMode` in agent frontmatter (v2.1.72+)** | agents/ | Declaratively specify `permissionMode` in agent definition YAML. No need for `mode` specification at spawn time |
| **Agent Teams official best practices (v2.1.72+)** | breezing | 5-6 tasks/teammate guideline, `teammateMode` setting, plan approval pattern reflected in team-composition |
| **Sandboxing (`/sandbox`)** | breezing, harness-work | OS-level filesystem/network isolation. Complementary layer to `bypassPermissions` |
| **`opusplan` model alias** | breezing | Auto-switches Opus (plan) / Sonnet (execute). Optimal for Lead's Plan -> Execute flow |
| **`CLAUDE_CODE_SUBAGENT_MODEL` env var** | breezing, harness-work | Centralized sub-agent model control. Consolidates Worker/Reviewer model management |
| **`availableModels` setting** | setup | Model restriction list. Model governance for enterprise operations |
| **Checkpointing (`/rewind`)** | harness-work | Session state tracking, rewind, and summarization. Supports safe exploration and experimentation |
| **Code Review (managed service)** | harness-review | Multi-agent PR review + `REVIEW.md`. Teams/Enterprise Research Preview |
| **Status Line (`/statusline`)** | all skills | Custom shell script status bar. Constant monitoring of context usage, cost, and git state |
| **1M Context Window (`sonnet[1m]`)** | harness-review, breezing | 1M token context window for large codebase analysis |
| **Per-model Prompt Caching Control** | all skills | `DISABLE_PROMPT_CACHING_*` for per-model cache control. Debugging and cost optimization |
| **`CLAUDE_CODE_DISABLE_ADAPTIVE_THINKING`** | harness-work | Disable Adaptive Reasoning to revert to fixed thinking budget. Predictable cost control |
| **Chrome Integration (`--chrome`, beta)** | harness-work, harness-review | Browser automation for UI testing, form input, console debugging. `/chrome` for in-session toggle |
| **LSP server integration (`.lsp.json`)** | setup | CC native feature. No Harness default `.lsp.json` config (configurable via `/setup lsp`) |
| **`SubagentStart`/`SubagentStop` matcher (v2.1.72+)** | breezing, hooks | Monitor sub-agent lifecycle by agent type at settings.json level. Individual tracking for Worker/Reviewer/Scaffolder/Video Generator |
| **Agent Teams: Task Dependencies** | breezing | Automatic task dependency management. Blocked tasks auto-unblock when dependencies complete. File-lock preventing claim conflicts |
| **`--teammate-mode` CLI flag (v2.1.72+)** | breezing | Per-session display mode toggle: `claude --teammate-mode in-process` |
| **`CLAUDE_CODE_DISABLE_BACKGROUND_TASKS` (v2.1.72+)** | setup | `=1` disables all background task features. For environments where security policy restricts background execution |
| **`CLAUDE_AUTOCOMPACT_PCT_OVERRIDE` (v2.1.72+)** | breezing, harness-work | Adjust sub-agent auto-compaction threshold (default 95%). `50` for early compaction, improving long-running Worker stability |
| **`cleanupPeriodDays` setting (v2.1.72+)** | setup | Auto-cleanup period for sub-agent transcripts (default 30 days) |
| **`/btw` side question (v2.1.72+)** | all skills | Short questions while preserving current context. No tool access, not saved to history. Lightweight alternative to sub-agent |
| **Plugin CLI commands (v2.1.72+)** | setup | `claude plugin install/uninstall/enable/disable/update` + `--scope` flag. Script automation support |
| **Remote Control enhancements (v2.1.72+)** | investigated, future support | `/remote-control` (`/rc`) to enable in-session. `--name`, `--sandbox`, `--verbose` flags. `/mobile` for QR code. Auto-reconnection support |
| **`skills` field in agent frontmatter (v2.1.72+)** | agents/ | Preload skills into sub-agents. Worker gets `harness-work`+`harness-review`, Reviewer gets `harness-review`, Scaffolder gets `harness-setup`+`harness-plan` (implemented) |
| **`modelOverrides` setting (v2.1.73)** | setup, breezing | Map model picker entries to custom provider model IDs (Bedrock ARNs, etc.) |
| **`/output-style` deprecation (v2.1.73)** | all skills | Migrated to `/config`. Output style selection moved to config menu |
| **Bedrock/Vertex Opus 4.6 default (v2.1.73)** | breezing | Cloud provider default Opus updated from 4.1 to 4.6 |
| **`autoMemoryDirectory` setting (v2.1.74)** | session-memory, setup | Customize auto-memory storage path. Project-specific memory isolation |
| **`CLAUDE_CODE_SESSIONEND_HOOKS_TIMEOUT_MS` (v2.1.74)** | hooks | Configurable SessionEnd hook timeout (previously fixed at 1.5s kill) |
| **Full model ID fix (v2.1.74)** | agents/, breezing | `claude-opus-4-6` etc. now recognized in agent frontmatter and JSON config |
| **Streaming API memory leak fix (v2.1.74)** | breezing, harness-work | Fixed unbounded RSS growth in streaming response buffers |
| **`--remote` / Cloud Sessions** | breezing, harness-work | Launch cloud sessions from terminal with `--remote`. Async task execution |
| **`/teleport` (`/tp`)** | session | Import cloud sessions to local terminal |
| **`CLAUDE_CODE_REMOTE` env var** | hooks, session-env-setup | Detect cloud vs local execution. Used for hook conditional branching |
| **`CLAUDE_ENV_FILE` SessionStart persistence** | hooks, session-env-setup | Persist env vars from SessionStart hooks to subsequent Bash commands |
| **Slack Integration (`@Claude`)** | — | Future support (requires Teams/Enterprise). No Harness implementation |
| **Server-managed settings (public beta)** | setup | Server-delivered bulk settings management. For Teams/Enterprise |
| **Microsoft Foundry** | setup, breezing | Added as new cloud provider |
| **`PreCompact` hook** | hooks | Pre-compaction state save and WIP task warning (implemented) |
| **`Notification` hook event** | hooks | Custom handler for notification events (implemented) |
| **`/context` command (v2.1.74)** | all skills | Visualize context consumption and suggest optimizations |
| **`maxTurns` agent safety limit** | agents/ | Runaway prevention via turn limits. Worker: 100, Reviewer: 50, Scaffolder: 75 |
| **Output token limits 64k/128k (v2.1.77)** | all skills | Opus 4.6 / Sonnet 4.6 default 64k, max 128k tokens |
| **`allowRead` sandbox setting (v2.1.77)** | harness-review | Re-allow read access to specific paths within `denyRead` regions |
| **PreToolUse `allow` respects `deny` (v2.1.77)** | guardrails | Hook `allow` no longer overrides settings.json `deny` rules |
| **Agent `resume` -> `SendMessage` (v2.1.77)** | breezing | Agent tool `resume` deprecated. Migrated to `SendMessage({to: agentId})` |
| **`/branch` (formerly `/fork`) (v2.1.77)** | session | `/fork` -> `/branch` rename. Alias remains |
| **`claude plugin validate` enhanced (v2.1.77)** | setup | Added frontmatter + hooks.json syntax validation |
| **`--resume` 45% faster (v2.1.77)** | session | Fork-heavy session resume speedup and memory reduction |
| **Stale worktree race fix (v2.1.77)** | breezing | Prevention of active worktree false deletion |
| **`StopFailure` hook event (v2.1.78)** | hooks | Capture session stop failures on API errors |
| **`${CLAUDE_PLUGIN_DATA}` variable (v2.1.78)** | hooks, setup | Persistent state directory that survives plugin updates |
| **Agent `effort`/`maxTurns`/`disallowedTools` frontmatter (v2.1.78)** | agents/ | Declarative control for plugin agents |
| **`deny: ["mcp__*"]` fix (v2.1.78)** | setup | settings.json deny now correctly blocks MCP tools |
| **`ANTHROPIC_CUSTOM_MODEL_OPTION` (v2.1.78)** | setup | Custom model picker entries |
| **`--worktree` skills/hooks loading fix (v2.1.78)** | breezing | Skills and hooks load correctly with worktree flag |
| **Skill `effort` frontmatter (v2.1.80)** | harness-work, harness-review, harness-plan, harness-release | Set thinking depth on 5-verb skills themselves, improving initial quality for heavy flows |
| **Agent `initialPrompt` frontmatter (v2.1.83)** | agents/ | Stabilize the first turn of Worker / Reviewer / Scaffolder per role |
| **`sandbox.failIfUnavailable` (v2.1.83)** | setup, guardrails | Prevent silent fallback to unsandboxed when sandbox fails to start |
| **`CLAUDE_CODE_SUBPROCESS_ENV_SCRUB=1` (v2.1.83)** | hooks, setup | Reduce credential exposure surface to hook / Bash / MCP stdio subprocesses |
| **`TaskCreated` / `CwdChanged` / `FileChanged` hooks (v2.1.83-2.1.84)** | hooks, session | Add reactive state tracking and Plans / rules re-read reminders |
| **Rules / skills `paths:` YAML list (v2.1.84)** | setup, localize-rules | Hold multiple globs in structured format, making rule scope more readable and less fragile |
| **Hooks conditional `if` field (v2.1.85)** | hooks, guardrails | Scope `PermissionRequest` to safe Bash and edit tools only, reducing unnecessary hook invocations and false warnings |
| **Large session truncation fix (v2.1.78)** | session | Fixed truncation of sessions over 5MB |
| **`--console` auth flag (v2.1.79)** | setup | Anthropic Console API billing authentication |
| **Turn duration display (v2.1.79)** | all skills | Toggle turn execution time display in `/config` |
| **`CLAUDE_CODE_PLUGIN_SEED_DIR` multiple dirs (v2.1.79)** | setup | Multiple seed directory specification |
| **SessionEnd hooks `/resume` fix (v2.1.79)** | hooks | SessionEnd fires correctly during interactive session switching |
| **18MB startup memory reduction (v2.1.79)** | all skills | Reduced startup memory usage |
| **MCP tool description cap 2KB (v2.1.84)** | all skills | Prevent context bloat from large OpenAPI-derived MCP schemas. CC auto-inherited |
| **`TaskCreated` hook blocking (v2.1.84)** | hooks | Hooks fire synchronously on TaskCreate. Used for runtime-reactive state tracking |
| **Idle-return prompt 75min (v2.1.84)** | session | Suggests `/clear` after 75+ minutes idle. Prevents stale session token waste. CC auto-inherited |
| **`X-Claude-Code-Session-Id` header (v2.1.86)** | setup | Session ID header added to API requests. Usable for proxy-side aggregation. CC auto-inherited |
| **Cowork Dispatch fix (v2.1.87)** | breezing | Fixed Cowork Dispatch message delivery. CC auto-inherited |
| **`PermissionDenied` hook event (v2.1.89)** | hooks, breezing | Fires when auto mode classifier denies. `{retry:true}` for retry guidance. Implemented for Breezing Worker denial tracking and Lead notification |
| **`"defer"` permission decision (v2.1.89)** | hooks, breezing | Return `"defer"` from PreToolUse to pause headless sessions, then re-evaluate on resume. Safety valve for Breezing |
| **`updatedInput` + `AskUserQuestion` (v2.1.89)** | hooks | External UI collects questions in headless environments and injects `allow` + answer. For future interactive flow normalization |
| **Hook output >50K disk save (v2.1.89)** | hooks | Large hook output saved to disk with preview. Prevents context bloat |
| **Hooks `if` compound command fix (v2.1.89)** | hooks | Compound commands like `ls && git push` and `FOO=bar git push` now match `if` conditions correctly. CC auto-inherited |
| **Autocompact thrash loop fix (v2.1.89)** | all skills | Stops with actionable error after 3 consecutive compact-then-refill cycles. CC auto-inherited |
| **Nested CLAUDE.md re-injection fix (v2.1.89)** | all skills | Fixed bug where CLAUDE.md was re-injected dozens of times in long sessions. CC auto-inherited |
| **Thinking summaries default off (v2.1.89)** | all skills | Thinking summaries generation stopped by default. Restore with `showThinkingSummaries:true`. CC auto-inherited |
| **PreToolUse exit 2 JSON fix (v2.1.90)** | hooks, guardrails | Fixed blocking behavior with JSON stdout + exit 2. pre-tool.sh deny now more reliable |
| **PostToolUse format-on-save fix (v2.1.90)** | hooks | Fixed Edit/Write failures after hook-triggered file rewrites. CC auto-inherited |
| **`--resume` prompt-cache miss fix (v2.1.90)** | session | Regression fix since v2.1.69. Resume cache miss with deferred tools/MCP/agents. CC auto-inherited |
| **SSE/transcript performance (v2.1.90)** | all skills | SSE frame O(n^2)->O(n), transcript writes quadratic->linear. CC auto-inherited |
| **`/powerup` interactive lessons (v2.1.90)** | — | Animated demos for learning Claude Code features. CC auto-inherited |

## Feature Details

### Task tool metrics

Aggregates token count, tool call count, and execution time consumed by sub-agents.
`parallel-workflows` skill aggregates metrics from multiple sub-agents for cost analysis.

```
metrics: {tokens: 40000, tools: 7, duration: 67s}
```

### `/debug` command

Session diagnostic command. Used to investigate complex errors and unexpected behavior.
`troubleshoot` skill is auto-triggered for systematic diagnosis.

### PDF page range specification

Specify page ranges when loading large PDFs (e.g., `pages: "1-5"`).
Used for document processing in `notebook-lm` skill and large spec reference in `harness-review`.

### Git log flags

Utilize structured options for `git log` (`--format`, `--stat`, `--since`, etc.).
Streamlines release note generation, commit analysis, and change tracking.

### OAuth authentication

OAuth authentication configuration for MCP servers that don't support DCR (Dynamic Client Registration).
Used for Codex CLI connection in `codex-review` skill.

### 68% memory optimization

Reduced memory usage when resuming sessions with the `--resume` flag.
Effective for context continuation in long work sessions.

### Sub-agent MCP

Sub-agents launched via Task tool can share the parent session's MCP tools.
Each agent can use the same MCP toolset during parallel implementation with `task-worker`.

### Reduced Motion

Accessibility setting. Option to reduce motion/animations.
Considered during UI generation by `harness-ui` skill.

### TeammateIdle/TaskCompleted Hook

Hooks that fire when Breezing team members become idle or when tasks complete.
Processed by `scripts/hook-handlers/teammate-idle.sh` and `task-completed.sh`.

```json
"TeammateIdle": [{"hooks": [{"type": "command", "command": "...teammate-idle", "timeout": 10}]}],
"TaskCompleted": [{"hooks": [{"type": "command", "command": "...task-completed", "timeout": 10}]}]
```

### Agent Memory (memory frontmatter)

Enable persistent memory via `memory: project` field in agent definition YAML.
`task-worker` and `code-reviewer` learn implementation patterns and failure resolutions across sessions.

### Fast mode (Opus 4.6)

High-speed output mode toggled via `/fast` command. Uses the same Opus 4.6 model.
Available across all skills. Effective for reducing wait times during long implementation tasks.

### Auto memory recording

Automatically persists learning content to memory files at session end.
Managed by `session-memory` skill. Automatically restores previous session context.

### Skill budget scaling

SKILL.md character budget auto-adjusts to 2% of the context window.
The recommended 500 lines is a guideline; the effective limit depends on the model's context window size.

### Task(agent_type) restrictions

Specify `subagent_type` when calling the Task tool to restrict sub-agent types.
Combined with `agents/` definitions, ensures only intended agents are launched.

### Plugin settings.json

Pre-define settings at initialization via the plugin's `settings.json`.
Reduces init token consumption and applies security policies from session start.

### Worktree isolation

Use `git worktree` to safely enable parallel writes to the same file.
Prevents conflicts during multi-agent parallel implementation in `breezing` and `parallel-workflows`.

### Background agents

Launch background agents asynchronously. Other processing can continue without waiting for completion.
Used for parallel multi-scene generation in `generate-video` skill.

### ConfigChange hook

Hook that fires when configuration files (`settings.json`, etc.) are changed.
Changes are recorded and audited by `scripts/hook-handlers/config-change.sh`.

### last_assistant_message

Access the last assistant message at session end.
Used by `session-memory` skill for session quality self-evaluation.

### Sonnet 4.6 (1M context)

Sonnet 4.6 model with up to 1M token context window.
Supports large codebase analysis and long document processing. Available across all skills.

> Note: In v2.1.69, legacy Sonnet 4.5 references are assumed to auto-migrate to Sonnet 4.6.

### Memory leak fixes (v2.1.50-v2.1.63)

CC 2.1.50 fixed memory leaks related to LSP diagnostic data, large tool output, file history, and shell execution.
Garbage collection for completed tasks was also implemented, significantly improving stability of long team sessions like `/breezing`.
v2.1.63 added fixes for MCP reconnection leaks, git root cache, JSON parse cache, teammate message retention, and shell command prefix cache leaks.
Harness has already implemented its own countermeasures via JSONL rotation (500->400 lines) and atomic updates.

### `claude agents` CLI (v2.1.50)

`claude agents list` displays registered agents.
Used by `troubleshoot` skill for diagnosing agent spawn failures.

```bash
claude agents list   # List registered agents
```

### WorktreeCreate/WorktreeRemove hook (v2.1.50)

Lifecycle hooks that fire on worktree creation and removal.
Used for auto-setup and cleanup in `/breezing` parallel workflows.
Implemented in `scripts/hook-handlers/worktree-create.sh` and `worktree-remove.sh`.

### `claude remote-control` (v2.1.51)

Subcommand enabling external build systems and local environment serving.
Potential future use in Breezing cross-session control and CI integration.

### `/simplify` (v2.1.63)

Auto code refinement command added in CC 2.1.63.
Integrated as Phase 3.5 Auto-Refinement in `/work`, automatically simplifying and organizing code after implementation.
Can be combined with `code-simplifier` plugin for deep refactoring via `--deep-simplify`.

### `/batch` (v2.1.63)

Command for parallel delegation of cross-cutting tasks (applying the same change across multiple files, migrations, etc.).
Used with `/breezing` to delegate bulk migrations to the Breezing team for parallel execution.
Effective for automating repetitive work and reducing human error.

### `code-simplifier` plugin

External plugin providing deep refactoring mode for `/simplify`.
Triggered by `--deep-simplify`, automatically performs logic decomposition, removal of unnecessary abstractions, and naming improvements.
Standard `/simplify` is lightweight; `--deep-simplify` provides more thorough refactoring.

### HTTP hooks (v2.1.63)

New hook format added in CC 2.1.63. `http` type is now available alongside existing `command` / `prompt` types.
POSTs JSON to a specified URL, enabling integration with external services (Slack, dashboards, metrics collection, etc.).
See the "http Type" section in [.claude/rules/hooks-editing.md](../.claude/rules/hooks-editing.md) for details.

### Auto-memory worktree sharing (v2.1.63)

CC 2.1.63 enabled Agent Memory sharing between worktrees when using `isolation: "worktree"`.
Parallel Implementers in `/breezing` can reference and update the same MEMORY.md while working in isolated worktrees.
Enables knowledge sharing between Implementers and prevents duplicate work on the same bug.

### `/clear` skill cache reset (v2.1.63)

Skill cache reset command added in CC 2.1.63.
Resolves the issue of stale cache behavior after editing skill files (common during skill development) via `/clear`.
Integrated into `troubleshoot` skill's cache diagnostic steps.

### `ENABLE_CLAUDEAI_MCP_SERVERS` (v2.1.63)

Environment variable added in CC 2.1.63. Setting to `false` disables MCP servers provided by claude.ai.
Intended for environments where security policy restricts connections to external MCP servers.
Added to `setup` skill's environment initialization checklist.

### Agent hooks (v2.1.68)

`type: "agent"` hooks added in CC 2.1.68. LLM agents make hook decisions, enabling dynamic judgment of code quality issues that are difficult to detect with regex.
Harness uses them in 3 limited locations, with `model: "haiku"` and `matcher` to control costs:

- **PreToolUse Write|Edit**: Guards against secret embedding, TODO stubs, and security vulnerabilities
- **Stop**: WIP task guard (checks for remaining `cc:WIP` tasks in Plans.md)
- **PostToolUse Write|Edit**: Async code review (quality, naming, single responsibility)

Designed to roll back to `command` type if effectiveness is insufficient.

### Effort levels + ultrathink (v2.1.68)

CC 2.1.68 changed Opus 4.6 to **medium effort** by default. The `ultrathink` keyword enables high effort (extended thinking) for one turn.
`harness-work` skill calculates a multi-factor score (changed file count, target directories, keywords, failure history, PM explicit specification) and auto-injects `ultrathink` at Worker spawn prompt start when score >= 3.
See the "Effort Level Control" section in `skills/harness-work/SKILL.md` for details.

### Opus 4/4.1 removal (v2.1.68)

Opus 4 and Opus 4.1 were removed from the first-party API in CC 2.1.68. When Harness specifies `model: opus` equivalent for target agents, they auto-migrate to Opus 4.6.
Worker/Reviewer agents use `model: sonnet` and are unaffected. Only Lead (when using Opus) receives the change to medium effort as default.

### `${CLAUDE_SKILL_DIR}` variable (v2.1.69)

Skill execution base path variable `${CLAUDE_SKILL_DIR}` introduced in CC 2.1.69.
Harness standardized links from `SKILL.md` to `references/*.md` using `${CLAUDE_SKILL_DIR}/references/...`, maintaining consistent references across mirror configurations (codex/opencode).

### InstructionsLoaded hook (v2.1.69)

`InstructionsLoaded` event added in CC 2.1.69. Harness created
`scripts/hook-handlers/instructions-loaded.sh` for lightweight tracking and pre-verification at instructions load completion.

### `agent_id` / `agent_type` addition (v2.1.69)

`agent_id` / `agent_type` added to teammate events.
Harness guardrails extended from `session_id`-based to `agent_id`-preferred (fallback: `session_id`) for stable role guarding.

### `{"continue": false}` teammate response (v2.1.69)

`TeammateIdle` / `TaskCompleted` can now return `{"continue": false, "stopReason": "..."}`.
Harness returns this response on stop request receipt and all-tasks-complete, making breezing stop decisions explicit.

### `/reload-plugins` (v2.1.69)

Added `/reload-plugins` to the development flow for reflecting changes without session restart after skill/hook edits.
The standard procedure is: edit -> `/reload-plugins` -> re-execute.

### `includeGitInstructions: false` (v2.1.69)

For tasks that don't need constant git instruction embedding, `includeGitInstructions: false` suppresses token consumption.
Harness recommends this for lightweight tasks (documentation updates, etc.) in breezing/work.

### `git-subdir` plugin source (v2.1.69)

`git-subdir` method for managing plugin source in monorepo subdirectories is now supported.
Harness currently does not force additional fields in `.claude-plugin/plugin.json`, specifying `plugin source` explicitly at release time (compatibility priority).

### Compaction image retention (v2.1.70)

CC 2.1.70 made summary requests retain images during context compaction.
This maintains image context after compaction in sessions containing screenshots and diagrams.
Prompt cache reuse rate also improved, benefiting all skills that handle images.

### Sub-agent final report simplification (v2.1.70)

Sub-agent completion reports were simplified, reducing token consumption.
For sessions launching many sub-agents in `breezing` and `harness-work`, cumulative token savings are significant.

### `--resume` skill list re-injection removed (v2.1.70)

Skill list re-injection on `--resume` session resume was removed.
This saves ~600 tokens, making the resume flow in `session` skill lighter.

### Plugin hooks fixes (v2.1.70)

Multiple Plugin hooks bugs fixed in v2.1.70:
- `Stop` / `SessionEnd` hooks fire correctly after `/plugin` command execution
- Template collision between hooks sharing the same template resolved
- `WorktreeCreate` / `WorktreeRemove` hooks confirmed working correctly

### Teammate nesting prevention additional fix (v2.1.70)

Additional fix for teammate nesting prevention beyond v2.1.69.
Strengthened prevention of the cascading problem where agents infinitely spawn other agents.

### PostToolUseFailure hook (v2.1.70)

`PostToolUseFailure` event added in CC 2.1.70. A new hook event that fires on tool call failure.
Harness uses it in `hooks` skill and `error-recovery` for auto-escalation on consecutive failures (stop after 3 consecutive failures).

```json
"PostToolUseFailure": [{
  "hooks": [{
    "type": "command",
    "command": "...post-tool-failure.sh",
    "timeout": 10
  }]
}]
```

### `/loop` + Cron scheduling (v2.1.71)

`/loop` command added in CC 2.1.71. Specifying interval and prompt like `/loop 5m <prompt>` enables cron-style periodic command execution.
`breezing` uses `/loop 5m /sync-status` for periodic task progress monitoring.
Unlike existing `TeammateIdle` (passive, event-driven), this enables active periodic monitoring.

### Background Agent output path fix (v2.1.71)

CC 2.1.71 now includes output file path in Background Agent completion notifications.
This enables safe recovery of background agent results even after compaction.
Makes `run_in_background: true` practical in `breezing` and `parallel-workflows`.

### `--print` team agent hang fix (v2.1.71)

Fixed team agent hang issue in `--print` mode.
Improved team agent stability during `claude --print` execution in CI pipelines.

### Plugin install parallel execution fix (v2.1.71)

Fixed state race condition when multiple Claude Code instances install plugins simultaneously.
Improved plugin loading stability when multiple teammates launch simultaneously in `breezing`.

### Marketplace improvements (v2.1.71)

Multiple Marketplace improvements in CC 2.1.71:
- `@ref` parser fix: `owner/repo@vX.X.X` reference resolution now accurate
- Update merge conflict fix: plugin updates more stable
- MCP server deduplication: prevents duplicate registration of the same MCP server
- `/plugin uninstall` uses `settings.local.json`: accurate reflection to user-local settings

### Per-agent hooks (v2.1.69+)

CC 2.1.69 added `hooks` field to agent definition frontmatter.
Agent-specific hooks can be defined separately from global hooks.json.

Harness usage:
- **Worker**: `PreToolUse` applies `pre-tool.sh` guardrail on Write/Edit
- **Reviewer**: `Stop` logs review session completion

Agent definition hooks are only active during that agent's lifecycle and auto-cleanup on exit.

### Agent `isolation: worktree` (v2.1.50+)

Adding `isolation: worktree` to agent definition frontmatter causes the agent to automatically create a git worktree on startup and work in an independent repository copy.
When there are no changes, the worktree is auto-cleaned up.

Harness added `isolation: worktree` to the Worker agent.
Combined with `memory: project`, Agent Memory (MEMORY.md) is shared between worktrees, enabling parallel Workers to reference and update the same learning content.

### Auto Mode rollout policy

Auto Mode is being organized as a migration candidate to move Claude Code team execution toward the safer side.
However, the shipped default is still `bypassPermissions`, and only documented permission modes are kept in project templates and frontmatter.

| Layer | Value Used | Reason |
|---------|--------|------|
| project template (`permissions.defaultMode`) | `bypassPermissions` | `autoMode` is not included in documented permission modes |
| agent frontmatter (`permissionMode`) | `bypassPermissions` | Declarative settings use only documented values |
| teammate execution path | `bypassPermissions` (current) | Matches shipped default with actual permission inheritance |
| `--auto-mode` | opt-in marker | Only attempts rollout when parent session has compatible permission mode |

Default command examples:

```bash
/breezing all
/execute --breezing all
```

### Subagent `background` field

Adding `background: true` to agent definition frontmatter causes the agent to always run as a background task.
Without explicitly specifying `run_in_background: true`, every launch via Agent tool becomes a background execution.

```yaml
---
name: long-running-analyzer
background: true
---
```

Harness could apply this to `breezing` Worker spawns, but currently Lead explicitly controls `run_in_background`, so additional application is considered for Phase 2+.

### Subagent `local` memory scope

`memory: local` saves to `.claude/agent-memory-local/<name>/` and this path should be added to `.gitignore`.
Difference from `project`:

| Scope | Path | VCS Commit | Use Case |
|---------|------|-------------|------------|
| `user` | `~/.claude/agent-memory/<name>/` | Not tracked | Cross-project learning |
| `project` | `.claude/agent-memory/<name>/` | Shareable | Team-shared project knowledge |
| `local` | `.claude/agent-memory-local/<name>/` | Not recommended | Personal or sensitive learning |

Harness uses `memory: project` for both Worker/Reviewer. `local` is suitable for personal debugging patterns, but current settings are maintained to prioritize team sharing.

### Agent Teams experimental flag

Agent Teams are enabled as an experimental feature via the `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS` environment variable.
Also configurable via settings.json:

```json
{
  "env": {
    "CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS": "1"
  }
}
```

Harness's `breezing` skill depends on Agent Teams, so a verification step to confirm this env var is set was added to setup.

### Desktop Scheduled Tasks

Desktop app Scheduled Tasks are saved to `~/.claude/scheduled-tasks/<task-name>/SKILL.md`.
Define `name` and `description` in YAML frontmatter, with the prompt in the body.

Schedule settings (frequency, time, folder) are managed from the Desktop app UI.
Can be used for periodic execution of `/harness-work` or `/harness-review`.

### `/agents` command

Interactive agent management interface. Supports:
- List all available agents (built-in, user, project, plugin)
- Guided or Claude-generated agent creation
- Edit settings and tool access for existing agents
- Delete custom agents

Non-interactive listing from CLI: `claude agents`

### `--agents` CLI flag

Pass agent definitions as JSON at session startup. Temporary configuration not persisted to disk:

```bash
claude --agents '{
  "quick-reviewer": {
    "description": "Quick code review",
    "prompt": "Review for critical issues only",
    "tools": ["Read", "Grep", "Glob"],
    "model": "haiku"
  }
}'
```

Useful for temporary agent injection in CI/CD pipelines.

### `ExitWorktree` tool (v2.1.72)

`ExitWorktree` tool added in CC 2.1.72. Enables programmatic exit from worktree sessions created with `EnterWorktree`.
Previously, manual selection via prompt at worktree session end was the only option; now agents can automatically exit worktrees after implementation completion.

Harness usage:
- `breezing` Worker explicitly closes worktree via `ExitWorktree` after completing work with `isolation: worktree`
- Improves worktree cleanup reliability (can combine with existing auto-delete behavior when no changes exist)

### Effort levels simplification (v2.1.72)

CC 2.1.72 simplified effort levels to 3 tiers: `low/medium/high`. `max` level was removed, display symbols unified to `○ ◐ ●`. `/effort auto` resets to default (medium).

Harness impact:
- `ultrathink` keyword for high effort injection remains effective (no change)
- harness-work scoring logic needs no change (ultrathink -> high effort mapping maintained)
- Document references to `max` unified to `high`

### Agent tool `model` parameter restored (v2.1.72)

CC 2.1.72 restored the Agent tool `model` parameter. Per-invocation model specification for sub-agent launches is possible again.
Separate from the agent definition `model` field, temporary model specification at spawn time is available.

Harness usage potential:
- Lightweight tasks (doc updates, format fixes, etc.) can spawn with `model: "haiku"` for cost reduction
- Security reviews and architecture changes can spawn with `model: "opus"` for maximum quality
- Currently Worker/Reviewer are fixed at `model: sonnet`. Dynamic model switching by Lead based on task characteristics is considered for Phase 2+

### `/plan` description argument (v2.1.72)

CC 2.1.72 enabled `/plan` command to accept an optional description argument.
Enter plan mode instantly with description like `/plan fix the auth bug`.

Harness usage:
- Complementary with `harness-plan` skill's `create` subcommand
- Recommended as a shortcut for users wanting quick plan mode entry

### Parallel tool call fix (v2.1.72)

Important bug fix for parallel tool calls in CC 2.1.72.
Previously, if any of Read, WebFetch, or Glob failed, sibling calls running in parallel were also cancelled.
After fix, only Bash errors cascade; other tool failures are processed independently.

Harness impact:
- Improved stability when running file reads and web searches in parallel in `breezing` and `harness-work`
- Fixed issue where non-existent file Read cancelled other valid Reads
- Improved Worker agent reliability during exploration phase

### Worktree isolation fixes (v2.1.72)

Two worktree isolation bugs fixed in CC 2.1.72:

1. **Task resume cwd restore**: Tasks resumed via `resume` parameter now correctly restore the worktree's working directory
2. **Background notification worktreePath**: Background task completion notifications now include a `worktreePath` field

Harness impact:
- Improved reliability when `breezing` Worker works with `isolation: worktree` and Lead retrieves results
- Can now obtain worktree path from completion notification for Worker spawned with `run_in_background: true`

### `/clear` preserves background agents (v2.1.72)

CC 2.1.72 changed `/clear` behavior. Only foreground tasks are stopped; background agents and Bash tasks are unaffected.

Harness impact:
- Background Workers survive even when user runs `/clear` during `breezing` team execution
- Lead can organize context with `/clear` without interrupting running tasks, improving safety

### Hooks fixes (v2.1.72)

Multiple hook-related bugs fixed in CC 2.1.72:

1. **transcript_path**: Correctly set in `--resume` / `--fork` sessions
2. **PostToolUse block reason double display**: Fixed duplicate display of block reason messages
3. **async hooks stdin**: Async hooks now correctly receive stdin
4. **skill hooks double-fire**: Fixed skill hooks firing twice per event

Harness impact:
- `pre-tool.sh` / `post-tool.sh` guardrail hooks now accurately fire once, improving log reliability
- `session-memory` transcript references work correctly in `--resume` sessions

### HTML comments hidden (v2.1.72)

CC 2.1.72 hides HTML comments (`<!-- ... -->`) in CLAUDE.md files from auto-injection.
Still visible when reading files directly with the Read tool.

Harness impact:
- `<!-- This section is auto-generated by claude-mem. -->` markers used by claude-mem are hidden during auto-injection
- **No practical impact**: Markers are informational comments, and the activity log table itself exists outside comments, so display is unaffected
- Important instructions and settings should not be placed inside HTML comments going forward

### Bash auto-approval additions (v2.1.72)

CC 2.1.72 added the following commands to the Bash auto-approval allow list:
`lsof`, `pgrep`, `tput`, `ss`, `fd`, `fdfind`

Harness impact:
- Worker can now execute process checking (`pgrep`) and file searching (`fd`) without permission prompts
- guardrails `pre-tool.sh` continues to pass these through (not in block targets)

### Prompt cache fix (v2.1.72)

CC 2.1.72 fixed prompt cache invalidation bug in SDK `query()` calls.
Up to 12x reduction in input token costs.

Harness impact:
- Significant cost reduction when spawning many sub-agents in `breezing` and `harness-work`
- Especially effective for repetitive API call patterns within the same session

### Output Styles (v2.1.72+)

CC Output Styles enable system prompt customization.
A different layer from CLAUDE.md (added as user message) and Skills (task-specific).

Harness provides `.claude/output-styles/harness-ops.md`:
- `keep-coding-instructions: true` -- maintain coding instructions while optimizing operational flow
- Structured progress reporting format (done / current / next action)
- Quality Gate tabular output
- Review verdict structured format
- Escalation (3-strike rule) standard output format

```bash
# Enable
/output-style harness-ops
```

### `permissionMode` in agent frontmatter (v2.1.72+)

`permissionMode` officially documented as an agent frontmatter field.

Harness integration:
- Added `permissionMode: bypassPermissions` to all 3 agents (Worker/Reviewer/Scaffolder)
- Declarative permission management without relying on `mode` specification at spawn time
- Auto Mode organized as rollout candidate; current shipped default remains `bypassPermissions`

```yaml
# agents/worker.md frontmatter
permissionMode: bypassPermissions  # added
```

### Agent Teams official best practices (v2.1.72+)

Claude Code official `agent-teams.md` established as independent documentation.
Reflected in Harness's `agents/team-composition.md`:

1. **Task granularity guideline**: 5-6 tasks/teammate recommended
2. **`teammateMode` setting**: Official support for `"auto"` / `"in-process"` / `"tmux"`
3. **Plan Approval pattern**: Official pattern requiring plan mode for Workers
4. **Quality Gate Hooks**: `TeammateIdle`/`TaskCompleted` exit 2 feedback pattern
5. **Team size**: 3-5 teammates recommended (aligns with Harness's Worker 1-3 + Reviewer 1)

### Sandboxing (`/sandbox`)

OS-level sandbox natively integrated into Claude Code. Uses Seatbelt on macOS and bubblewrap on Linux to restrict filesystem/network access for Bash commands.

**Two modes**:
- **Auto-allow mode**: Commands within sandbox are auto-approved. Access outside constraints falls back to normal permission flow
- **Regular permissions mode**: All commands within sandbox require approval

**Harness strategy**:
- Position as a **complementary layer** to `bypassPermissions` (not a replacement)
- Add OS-level safety boundaries to Worker agent Bash commands
- Explicitly restrict Worker write scope with `sandbox.filesystem.allowWrite`
- Restrict external access to trusted domains with `sandbox.network` (exfiltration prevention)

**Phased adoption plan**:

| Phase | Worker Permissions | Sandbox |
|---------|-----------|---------|
| Current | `bypassPermissions` + hooks guard | Not applied |
| Verification phase | `bypassPermissions` + hooks + sandbox auto-allow | Applied to Worker Bash |
| After stabilization | sandbox auto-allow only (`bypassPermissions` retirement consideration) | Applied to all Bash |

```json
// settings.json (verification phase)
{
  "sandbox": {
    "enabled": true,
    "filesystem": {
      "allowWrite": ["~/.claude", "//tmp"]
    }
  }
}
```

> `@anthropic-ai/sandbox-runtime` is published as OSS, also usable for sandboxing MCP servers.

### `opusplan` model alias

Hybrid alias that auto-switches to Opus in plan mode and Sonnet in execution mode.

**Harness usage**:
- Optimal for Breezing Lead sessions: Plan phase (task decomposition, architecture decisions) leverages Opus reasoning, Worker spawn coordination after uses Sonnet for cost efficiency
- Enable via `claude --model opusplan` or `/model opusplan`

**Environment variable control**:
```bash
# Customize opusplan internal mapping
ANTHROPIC_DEFAULT_OPUS_MODEL=claude-opus-4-6    # Plan mode
ANTHROPIC_DEFAULT_SONNET_MODEL=claude-sonnet-4-6  # Execution mode
```

### `CLAUDE_CODE_SUBAGENT_MODEL` env var

Environment variable for centralized sub-agent (Worker/Reviewer) model specification.

**Harness usage**:
- Current: Worker/Reviewer fixed at `model: sonnet` in agent definitions
- This env var enables model switching without changing agent definitions
- Useful for cost control in CI environments (`CLAUDE_CODE_SUBAGENT_MODEL=haiku` for test runs)

```bash
# Run all sub-agents with haiku (CI cost reduction)
export CLAUDE_CODE_SUBAGENT_MODEL=claude-haiku-4-5-20251001
```

### `availableModels` setting

Setting to restrict user-selectable models. When configured via managed/policy settings, restrictions apply to `/model`, `--model`, and `ANTHROPIC_MODEL` alike.

**Harness usage**:
- Enterprise model governance: Prevent Worker/Reviewer from using unintended models
- Combined with `model`, controls the model experience for all users

```json
// managed settings
{
  "model": "sonnet",
  "availableModels": ["sonnet", "haiku", "opusplan"]
}
```

### Checkpointing (`/rewind`)

Auto-tracks file edits during sessions, enabling rollback to any point.
Checkpoints are auto-created at each user prompt.

**Operation**:
- `Esc + Esc` or `/rewind` opens rewind menu
- Options: code restore / conversation restore / both / summarize from here

**Harness usage**:
- Roll back to pre-implementation state when issues found during `harness-work` self-review phase
- "Summarize from here" recovers context window from verbose debugging sessions
- Difference from `/compact`: checkpoints can selectively specify compaction range

**Limitations**:
- File changes via Bash commands are not tracked (`rm`, `mv`, `cp`, etc.)
- External manual changes are not tracked
- Not a git replacement; session-level "local undo"

### Code Review (managed service)

Multi-agent PR review service running on Anthropic infrastructure. Teams/Enterprise Research Preview.

**Operation overview**:
1. Auto-launches on PR creation/update
2. Multiple specialized agents analyze diffs and codebase in parallel
3. Verification step filters false positives
4. Deduplication and importance ranking before posting as inline comments

**Importance levels**:
| Marker | Level | Meaning |
|---------|--------|------|
| Red circle | Normal | Bug to fix before merge |
| Yellow circle | Nit | Minor issue (non-blocking) |
| Purple circle | Pre-existing | Bug that existed before this PR |

**`REVIEW.md`**: Review-specific guidance file placed at repository root. Defines rules applied only during review, separate from `CLAUDE.md`.

**Harness usage**:
- Considering `REVIEW.md` template generation for `harness-review` skill Code Review support
- Harness Worker self-review and managed Code Review are complementary (local + remote double check)
- Average cost $15-25/review. Note: `on-push` triggers incur cost per push

### Status Line (`/statusline`)

Customizable status bar displayed at the bottom of Claude Code terminal. Passes JSON session data to a shell script and displays the output text.

**Available data**:
- `model.id`, `model.display_name` -- current model
- `context_window.used_percentage` -- context usage rate
- `cost.total_cost_usd` -- session cost
- `cost.total_duration_ms` -- elapsed time
- `worktree.*` -- worktree information
- `agent.name` -- agent name
- `output_style.name` -- output style name

**Harness usage**:
- `scripts/statusline-harness.sh` provides Harness-specific status line
- Constant display of model name, context usage, session cost, git branch, Harness version
- ANSI color thresholds for context usage (70% yellow, 90% red)

### 1M Context Window (`sonnet[1m]`)

1 million token context window available for Opus 4.6 and Sonnet 4.6. Long-context pricing applies above 200K tokens.

**Harness usage**:
- Useful for `harness-review` large codebase analysis
- `breezing` sessions handling many files simultaneously
- Enable via `/model sonnet[1m]`. Disable with `CLAUDE_CODE_DISABLE_1M_CONTEXT=1`

### Per-model Prompt Caching Control

Environment variable group for per-model prompt cache control.

| Environment Variable | Purpose |
|---------|------|
| `DISABLE_PROMPT_CACHING` | Disable caching for all models |
| `DISABLE_PROMPT_CACHING_HAIKU` | Haiku only |
| `DISABLE_PROMPT_CACHING_SONNET` | Sonnet only |
| `DISABLE_PROMPT_CACHING_OPUS` | Opus only |

**Harness usage**:
- Disable cache for specific models during debugging to verify behavior
- Selective control when cache implementations differ across cloud providers (Bedrock/Vertex)

### `CLAUDE_CODE_DISABLE_ADAPTIVE_THINKING`

Environment variable to disable Adaptive Reasoning for Opus 4.6 / Sonnet 4.6, reverting to fixed thinking budget controlled by `MAX_THINKING_TOKENS`.

**Harness usage**:
- Useful in CI environments requiring predictable token costs
- Not mutually exclusive with `harness-work` effort scoring (both can be used, but controlling via ultrathink with adaptive thinking enabled is typically more effective)

### Chrome Integration (`--chrome`)

Beta feature connecting Claude Code with Chrome extension for browser automation from terminal.
Start session with `--chrome` flag, or enable in-session with `/chrome`.

**Key features**:
- Live debugging: Read console errors and immediately fix source code
- UI testing: Form validation, visual regression checking, user flow verification
- Data extraction: Extract structured data from web pages and save locally
- GIF recording: Record browser operation sequences as GIF

**Harness usage**:
- Auto-verification after UI component implementation in `harness-work`
- Visual review of web applications in `harness-review`
- Worker can execute browser tests with `/chrome` enabled

**Limitations**: Google Chrome / Microsoft Edge only. Brave, Arc, etc. not supported. WSL not supported.

### LSP server integration (`.lsp.json`)

Integrate Language Server Protocol servers via plugins for real-time code diagnostics.

**Available LSP plugins**:
| Plugin | Language Server | Installation |
|-----------|----------------|------------|
| `pyright-lsp` | Pyright (Python) | `pip install pyright` |
| `typescript-lsp` | TypeScript Language Server | `npm install -g typescript-language-server typescript` |
| `rust-lsp` | rust-analyzer | See rust-analyzer official guide |

**Provided capabilities**:
- Instant diagnostics: Show errors/warnings immediately after edits
- Code navigation: Go to definition, find references, hover info
- Type information: Symbol type and documentation display

**Configuration example** (`.lsp.json`):
```json
{
  "typescript": {
    "command": "typescript-language-server",
    "args": ["--stdio"],
    "extensionToLanguage": {
      ".ts": "typescript",
      ".tsx": "typescriptreact"
    }
  }
}
```

### `SubagentStart`/`SubagentStop` matcher

Hooks for monitoring sub-agent lifecycle by agent type at the settings.json level.
Official documentation now documents the pattern of specifying agent name in matcher.

**Harness implementation**:
- `SubagentStart`: Individual tracking for Worker/Reviewer/Scaffolder/Video Generator startups
- `SubagentStop`: Individual recording of each agent's completion
- Matchers added to existing `subagent-tracker` Node.js script

```json
"SubagentStart": [
  { "matcher": "worker", "hooks": [{ "type": "command", "command": "...subagent-tracker start" }] },
  { "matcher": "reviewer", "hooks": [{ "type": "command", "command": "...subagent-tracker start" }] }
]
```

### Agent Teams: Task Dependencies

Task dependencies can be set for Agent Teams tasks. Blocked tasks auto-unblock when dependencies complete.

**Behavior**:
- Tasks have 3 states: `pending`, `in_progress`, `completed`
- Pending tasks with unresolved dependencies cannot be claimed
- Auto-unblock on dependency completion (no manual intervention)
- File locks prevent simultaneous claiming by multiple teammates

**Harness usage**:
- Breezing Lead explicitly specifies dependencies during task decomposition
- Example: Order guarantee for "API endpoint implementation" -> "test creation" -> "documentation update"

### `--teammate-mode` CLI flag

Flag to specify Agent Teams display mode per session.

```bash
claude --teammate-mode in-process  # All teammates in same terminal
claude --teammate-mode tmux        # Individual pane per teammate
```

Overrides `teammateMode` setting in settings.json. `in-process` recommended for VS Code integrated terminal.

### `CLAUDE_CODE_DISABLE_BACKGROUND_TASKS`

Environment variable that disables all background task features when set to `=1`.

**Harness usage**:
- For environments where security policy restricts background execution
- Note: Breezing background Worker spawns are also disabled

### `CLAUDE_AUTOCOMPACT_PCT_OVERRIDE`

Environment variable to adjust sub-agent auto-compaction threshold (default 95%).

**Harness usage**:
- Set to `50` for early compaction. Improves long-running Worker stability
- Prevents context overflow when Breezing Workers read many files

### `cleanupPeriodDays` setting

Setting to control auto-cleanup period for sub-agent transcripts (default 30 days).
Transcripts are saved to `~/.claude/projects/{project}/{sessionId}/subagents/agent-{agentId}.jsonl`.

### `/btw` side question

Command for short questions while preserving current context.
Answers don't remain in main conversation history, so they don't consume context window.

**Comparison with sub-agents**:
- `/btw`: Questions answerable with current context (no tool access)
- Sub-agents: Independent investigation/implementation tasks (tool access available)

### Plugin CLI commands

Non-interactive plugin management commands. Supports script automation.

```bash
claude plugin install <plugin> [--scope user|project|local]
claude plugin uninstall <plugin> [--scope user|project|local]
claude plugin enable <plugin> [--scope user|project|local]
claude plugin disable <plugin> [--scope user|project|local]
claude plugin update <plugin> [--scope user|project|local|managed]
```

### Remote Control enhancements

`/remote-control` (`/rc`) can now enable Remote Control from within a session.

**New features**:
- `--name "My Project"`: Session name specification
- `--sandbox` / `--no-sandbox`: Enable/disable sandbox
- `--verbose`: Verbose logging
- `/mobile`: QR code display for quick iOS/Android app connection
- Auto-reconnect: Auto-recovery from network drops (within 10 minutes)
- `/config` -> "Enable Remote Control for all sessions" for always-on

### `skills` field in agent frontmatter

Adding `skills` field to sub-agent frontmatter preloads full skill content at startup.
Parent conversation skills are not inherited, so they must be explicitly listed.

**Harness implementation status**:
- Worker: `skills: [harness-work, harness-review]` -- Preloads implementation and self-review skills
- Reviewer: `skills: [harness-review]` -- Preloads review skill
- Scaffolder: `skills: [harness-setup, harness-plan]` -- Preloads setup and planning skills

> Inverse pattern of `skills` in skill (`context: fork`). Instead of skill controlling agent, agent loads skill.

### `modelOverrides` setting (v2.1.73)

Setting added in CC 2.1.73. Maps model picker (`/model` menu) entries to custom provider model IDs.
Provider-specific identifiers like Bedrock ARNs and Vertex AI model IDs can be specified.

**Harness usage**:
- In enterprise environments using Anthropic models via Bedrock/Vertex, `modelOverrides` maps picker display names to actual provider model IDs
- Worker/Reviewer `model: sonnet` auto-resolves to provider-specific ARN
- Combined with `availableModels`, can control team-wide model experience

```json
// settings.json
{
  "modelOverrides": {
    "sonnet": "arn:aws:bedrock:us-east-1::foundation-model/anthropic.claude-sonnet-4-6-20250514-v1:0",
    "opus": "arn:aws:bedrock:us-east-1::foundation-model/anthropic.claude-opus-4-6-20250610-v1:0"
  }
}
```

### `/output-style` deprecation (v2.1.73)

`/output-style` command deprecated in CC 2.1.73; output style selection moved to `/config` menu.
Existing `/output-style harness-ops` etc. continue to work, but `/config` is officially recommended.

**Harness impact**:
- Documentation references to `/output-style harness-ops` recommended to update to `/config` path
- `.claude/output-styles/harness-ops.md` itself remains valid (no change to configuration file location)
- Any skill scripts executing `/output-style` should consider switching to `/config`

### Bedrock/Vertex Opus 4.6 default (v2.1.73)

CC 2.1.73 updated the default Opus model on cloud providers (Amazon Bedrock / Google Vertex AI) from 4.1 to 4.6.
First-party API had Opus 4.6 as default since v2.1.68, now unified across cloud providers.

**Harness impact**:
- Lead (when using Opus) operates with medium effort default on Bedrock/Vertex environments too
- `opusplan` alias references Opus 4.6 on Bedrock/Vertex environments
- `ANTHROPIC_DEFAULT_OPUS_MODEL` env var override remains effective

### `autoMemoryDirectory` setting (v2.1.74)

Setting added in CC 2.1.74. Customizable storage directory for auto-memory.
Can be changed from default `~/.claude/` to a project-specific path.

**Harness usage**:
- Isolate auto-memory per project when using Harness across multiple projects
- Save memory to temp directory in CI, cleaning up on session end
- Different layer from Agent Memory (`memory: project`) -- auto-memory is user-level learning

```json
// settings.json (project level)
{
  "autoMemoryDirectory": ".claude/auto-memory"
}
```

### `CLAUDE_CODE_SESSIONEND_HOOKS_TIMEOUT_MS` (v2.1.74)

Environment variable added in CC 2.1.74. Specify `SessionEnd` hook timeout in milliseconds.
Previously fixed at 1.5s kill, causing heavy cleanup processes to be interrupted before completion.

**Harness usage**:
- Ensure sufficient timeout when `SessionEnd` hook runs `harness-mem` session recording or JSONL rotation
- Recommended: `5000` (5 seconds). For complex cleanup: up to `10000` (10 seconds)

```bash
export CLAUDE_CODE_SESSIONEND_HOOKS_TIMEOUT_MS=5000
```

### Full model ID fix (v2.1.74)

CC 2.1.74 made full model IDs with hyphen format (`claude-opus-4-6`, `claude-sonnet-4-6`, etc.) correctly recognized in agent frontmatter and JSON config.
Previously only aliases (`opus`, `sonnet`) worked reliably.

**Harness impact**:
- Full model IDs can now be specified in agent definition `model` field (e.g., `model: claude-sonnet-4-6`)
- Full model IDs also work in `--agents` CLI flag JSON
- Harness currently uses aliases (`sonnet`, `opus`), so no immediate impact. Useful for Bedrock/Vertex environments requiring full ID specification

```yaml
# agents/worker.md frontmatter (full model ID example)
model: claude-sonnet-4-6
```

### Streaming API memory leak fix (v2.1.74)

CC 2.1.74 fixed unbounded RSS (Resident Set Size) growth in streaming API response buffers.
Resolved issue of indefinite memory growth in Node.js processes during long streaming sessions.

**Harness impact**:
- Improved stability for `breezing` long team sessions
- Stabilized memory consumption for `harness-work` long Worker sessions with heavy file reads/writes
- Follows the v2.1.50-v2.1.63 memory leak fix series (LSP diagnostics, tool output, file history, etc.)
- Combined with Harness's own JSONL rotation measures for double stability assurance

### `--remote` / Cloud Sessions

CC `--remote` flag launches cloud sessions from terminal. Tasks run on Anthropic-managed isolated VMs, with PR creation possible after completion.

**Harness usage**:
- Delegate large `breezing` tasks to cloud, saving local resources
- Launch multiple tasks in parallel with `--remote` (each task as independent cloud session)
- Import cloud results locally with `/teleport`, connecting to subsequent `/harness-review`

```bash
# Execute task in cloud
claude --remote "Fix the authentication bug in src/auth/login.ts"

# Import to local after completion
/teleport
```

### `/teleport` (`/tp`)

Command to import cloud sessions to local terminal. Interactively select sessions with `/teleport` or `/tp`, or directly specify with `claude --teleport <session-id>`.

**Prerequisites**:
- Local git working directory must be clean
- Must run from the same repository
- Must be authenticated with the same Claude.ai account

### `CLAUDE_CODE_REMOTE` env var

`CLAUDE_CODE_REMOTE=true` is set within cloud sessions. Harness's `session-env-setup.sh` persists this as `HARNESS_IS_REMOTE`, enabling other hook handlers to skip local-only processing.

```bash
# Cloud detection in hook scripts
if [ "$HARNESS_IS_REMOTE" = "true" ]; then
  # Skip local-only processing in cloud environment
  exit 0
fi
```

### `CLAUDE_ENV_FILE` SessionStart persistence

CC `SessionStart` hooks can persist environment variables to subsequent Bash commands by writing `KEY=VALUE` to the file pointed to by `CLAUDE_ENV_FILE`.

Harness's `session-env-setup.sh` uses this mechanism to make `HARNESS_VERSION`, `HARNESS_AGENT_TYPE`, `HARNESS_IS_REMOTE`, etc. available session-wide.

### Slack Integration (`@Claude`)

Mentioning `@Claude` with a coding task in a Slack channel automatically creates a cloud session. Requires GitHub repository integration.

**Relationship with Harness**:
- Harness HTTP hooks (`type: "http"`) can be configured with Slack Webhook URL for task completion notifications
- `.claude/settings.json` hooks work within cloud sessions, so Harness guardrails apply to Slack-originated tasks

### Server-managed settings (public beta)

Server-delivered Claude Code settings for entire teams via Claude.ai admin panel. For Teams/Enterprise.

**Harness usage**:
- Centralized management of team-wide `permissions.deny` rules
- Deliver Harness hook settings via server (security confirmation dialog appears for hook settings)
- Combined with `availableModels` + `model` to control team model experience

### Microsoft Foundry

Azure-based new cloud provider. Added as the third third-party provider after Bedrock / Vertex.
Mappable to Foundry model IDs via `modelOverrides` setting.

### `PreCompact` hook

Hook event firing just before context compaction. Harness implements in 2 layers:

1. **`pre-compact-save.js`**: Persist session state (progress, metrics)
2. **agent hook**: Check for remaining `cc:WIP` tasks in Plans.md and inject warning message

```json
"PreCompact": [
  { "hooks": [
    { "type": "command", "command": "...pre-compact-save.js" },
    { "type": "agent", "prompt": "Check Plans.md for WIP tasks...", "model": "haiku" }
  ]}
]
```

### `Notification` hook event

Hook event that fires when Claude Code issues notifications. Listed in plugin reference.
Usable for notification forwarding to external monitoring tools and dashboards.

### `--plugin-dir` spec change (v2.1.76, breaking)

**Change**: `--plugin-dir` now accepts only one path. Multiple directories specified by repeating the flag.

```bash
# Old (no longer supported)
claude --plugin-dir path1,path2

# New
claude --plugin-dir path1 --plugin-dir path2
```

**Harness impact**: No impact for typical configurations using only the Harness plugin.
Syntax change only needed when using multiple plugins simultaneously.

---

## Claude Code 2.1.76 New Features

### MCP Elicitation support

**Operation overview**: Protocol enabling MCP servers to request structured input from users during task execution. Displays interactive dialog through form fields or browser URLs.

**Harness usage**:
- Breezing background Worker/Reviewer cannot interact with UI, so `Elicitation` hook implements auto-skip
- Normal sessions pass through (user responds via dialog)
- `elicitation-handler.sh` logs events

**Limitations**:
- Background agents cannot respond to elicitation (hook-based auto-processing required)
- MCP server must support elicitation

### `Elicitation`/`ElicitationResult` hooks

**Operation overview**: Two new hook events enabling interception before and after MCP Elicitation. `Elicitation` fires before response is returned to MCP server, `ElicitationResult` fires after.

**Harness usage**:
- `Elicitation`: Auto-skip decision during Breezing sessions + logging
- `ElicitationResult`: Result logging (`.claude/state/elicitation-events.jsonl`)
- Both event handlers registered in hooks.json

**Limitations**:
- Blocking (deny) via `Elicitation` hook prevents input from reaching MCP server
- Recommended timeout: Elicitation 10s / ElicitationResult 5s

### `PostCompact` hook

**Operation overview**: New hook event firing after context compaction completes. Counterpart to `PreCompact` hook (existing).

**Harness usage**:
- Post-compaction context re-injection (WIP task state restoration)
- Event recording in `.claude/state/compaction-events.jsonl`
- Improved state continuity in long sessions
- Symmetric structure: PreCompact (state save) -> PostCompact (state restore)

**Limitations**:
- Recommended timeout: 15s
- PostCompact may not fire when compaction fails (circuit breaker activation)

### `-n`/`--name` CLI flag

**Operation overview**: CLI flag to set display name at session startup. Use like `claude -n "auth-refactor"` for identification in session lists.

**Harness usage**:
- Auto-set `breezing-{timestamp}` format names for Breezing sessions
- Used for filtering and tracking in session lists
- Easier session identification in log analysis

**Code example**:
```bash
claude -n "breezing-$(date +%Y%m%d-%H%M%S)"
```

### `worktree.sparsePaths` setting

**Operation overview**: Setting that uses git sparse-checkout to check out only needed directories when using `claude --worktree` in large monorepos. Significantly improves worktree creation performance.

**Harness usage**:
- Reduce Breezing parallel Worker startup time (large repositories)
- Configure in `.claude/settings.json`:
```json
{
  "worktree": {
    "sparsePaths": ["src/", "tests/", "package.json"]
  }
}
```

**Limitations**:
- Files in paths not included in sparse-checkout are inaccessible to Workers
- All dependent directories must be included in sparsePaths

### `/effort` slash command

**Operation overview**: Slash command to switch effort level (low/medium/high) during a session. `/effort auto` resets to default.

**Harness usage**:
- Works with harness-work multi-factor scoring for task complexity-based effort control
- Manually set `/effort high` (enables ultrathink) for complex tasks
- Use `/effort low` to reduce token consumption for simple tasks

### `--worktree` faster startup

**Operation overview**: Reduced `--worktree` startup time through direct git refs reading and skipping redundant `git fetch` when remote branch is available.

**Harness usage**:
- Automatically reduced Breezing Worker startup overhead
- Especially beneficial when launching many Workers simultaneously

### Background agent partial result retention

**Operation overview**: Partial results are preserved in conversation context even when background agents are killed.

**Harness usage**:
- When Breezing Worker is interrupted by timeout or manual stop, partial work is communicated to Lead
- Enables reassignment leveraging partial Worker output
- Reduces wasted "start over" effort

### Stale worktree auto-cleanup

**Operation overview**: Stale worktrees remaining from interrupted parallel executions are automatically cleaned up.

**Harness usage**:
- Complements manual cleanup via `worktree-remove.sh`
- Auto-recovery after Breezing session crashes
- Prevents wasteful disk consumption

### Auto-compaction circuit breaker

**Operation overview**: Circuit breaker stops auto-compaction after 3 consecutive failures. Prevents infinite retry token waste.

**Harness usage**:
- Aligns with Harness "3-strike rule" design philosophy (3-attempt limit for CI failures)
- Prevents unexpected cost increases in long Breezing sessions
- Connects with PostToolUseFailure hook for escalation on circuit breaker activation

### Deferred Tools schema fix

**Operation overview**: Fixed issue where tools loaded via `ToolSearch` lost their input schema after compaction, causing array and numeric parameters to be rejected with type errors.

**Harness usage**:
- Improved stability of ToolSearch-loaded tools in long sessions
- MCP tools work correctly after Breezing compaction

### `/context` command (v2.1.74)

**Operation overview**: Analyzes context window consumption, identifies tools and memory bloating context. Displays actionable optimization suggestions (disconnecting unnecessary MCP servers, cleaning up bloated memory, etc.).

**Harness usage**:
- Root cause identification for "why is compaction happening so frequently" in long Breezing sessions
- Context optimization in environments with many hooks or MCP servers connected
- Instant analysis by running `/context` during a session

**Limitations**:
- Available only during sessions (not in batch mode)
- Not available within sub-agents

### `maxTurns` agent safety limit

**Operation overview**: Frontmatter field limiting sub-agent maximum turns. Agent automatically stops and returns results upon reaching the limit. Safety mechanism recommended in CC official docs.

**Harness usage**:
- Worker: `maxTurns: 100` -- For complex implementation tasks. Generous room while preventing runaway
- Reviewer: `maxTurns: 50` -- Specialized for read-only analysis. Not completing in 50 turns indicates a problem
- Scaffolder: `maxTurns: 75` -- Intermediate complexity for scaffolding and state updates

**Design decisions**:
- Lead can retrieve partial results when limit is reached
- Functions as a safety valve when combined with `bypassPermissions`

### `Notification` hook implementation

**Operation overview**: Hook event that fires when Claude Code issues notifications. Intercepts events like `permission_prompt` (permission checks), `idle_prompt` (idle notification), `auth_success` (authentication success), etc.

**Harness usage**:
- `notification-handler.sh` logs all notification events to `.claude/state/notification-events.jsonl`
- Tracks `permission_prompt` from Breezing background Workers (for post-analysis)
- Documented in hooks-editing.md since v3.10.3, but hooks.json implementation completed this time

**Log format**:
```json
{"event":"notification","notification_type":"permission_prompt","session_id":"...","agent_type":"worker","timestamp":"2026-03-15T..."}
```

### Output token limits 64k/128k (v2.1.77)

CC 2.1.77 increased default max output tokens for Opus 4.6 and Sonnet 4.6 to 64k, with upper limit extended to 128k tokens.

**Harness impact**:
- Long implementation code and large refactoring output less likely to be truncated
- Improved reliability when Worker agents output large volumes of file changes at once
- 128k output can increase costs, so cost management awareness is also needed

### `allowRead` sandbox setting (v2.1.77)

Now possible to re-allow read access to specific paths within `sandbox.filesystem.denyRead` broad blocks via `allowRead`.

**Harness usage**:
- Reviewer agent sandbox can denyRead `/etc/` while allowReading specific config files
- Provides restricted read access to sensitive directories during security review

### PreToolUse `allow` respects `deny` (v2.1.77)

CC 2.1.77 made PreToolUse hooks that return `"allow"` still respect settings.json `deny` permission rules. Previously hook `allow` overrode global `deny`.

**Harness impact**:
- Guardrails security model strengthened
- `deny: ["mcp__codex__*"]` in settings.json now reliably blocks regardless of PreToolUse hook decisions
- settings.json deny becomes recommended pattern alongside hook-based MCP blocking in `.claude/rules/codex-cli-only.md`

### Agent `resume` -> `SendMessage` (v2.1.77)

CC 2.1.77 deprecated Agent tool `resume` parameter. To resume stopped agents, use `SendMessage({to: agentId})`. `SendMessage` automatically resumes stopped agents in background.

**Harness impact**:
- `breezing` skill Lead communicates with Worker/Reviewer using `SendMessage`
- `SendMessage` documented as official communication method in `team-composition.md` Lead Phase B

### `/branch` (formerly `/fork`) (v2.1.77)

CC 2.1.77 renamed `/fork` command to `/branch`. `/fork` continues to function as an alias.

### `claude plugin validate` enhanced (v2.1.77)

CC 2.1.77 added YAML frontmatter and hooks.json syntax validation to `claude plugin validate` for skills, agents, and commands.

**Harness usage**:
- Add `claude plugin validate` to CI pipeline for early frontmatter error detection
- Usable as complement to `tests/validate-plugin.sh`

### `StopFailure` hook event (v2.1.78)

`StopFailure` event added in CC 2.1.78. Fires when session stop fails due to API errors (rate limit 429, auth failure 401, etc.).

**Harness usage**:
- `stop-failure.sh` handler logs error info to `.claude/state/stop-failures.jsonl`
- Used for post-analysis when Breezing Worker fails to stop due to rate limits
- Implemented as lightweight handler with 10-second timeout (no recovery processing needed)

### Hooks conditional `if` field (v2.1.85)

CC 2.1.85 enabled adding `if` conditions to hook definitions, finely controlling "which inputs trigger the hook." Uses permission rule syntax, so patterns like `Bash(git status*)` can specify tool name and input pattern together.

**Harness usage**:
- Split `PermissionRequest` into 2 tracks: `Edit|Write|MultiEdit` always evaluated, `Bash` pre-filtered by `if` to only safe command candidates
- `hooks/permission.sh` safety judgment preserved, while reducing unnecessary Bash permission hook invocations in the first place
- `MultiEdit` also included in matcher, eliminating auto-approval gaps that core guardrail had already addressed

**User experience improvement**:
- Before: Bash permission hooks fired broadly, incurring invocation cost even for ultimately-passed cases
- After: Hooks only fire for safe-read / test Bash, reducing response noise and unnecessary evaluation while maintaining auto-approval precision

### `${CLAUDE_PLUGIN_DATA}` variable (v2.1.78)

`${CLAUDE_PLUGIN_DATA}` directory variable added in CC 2.1.78. Persistent state storage that survives plugin updates.

**Harness usage potential**:
- Currently using `${CLAUDE_PLUGIN_ROOT}/.claude/state/`, which may be lost on plugin update
- Long-term: Consider migrating persistent data (metrics, notification logs, etc.) to `${CLAUDE_PLUGIN_DATA}`
- Migration pattern: `STATE_DIR="${CLAUDE_PLUGIN_DATA:-${CLAUDE_PLUGIN_ROOT}/.claude/state}"`

### Agent frontmatter: `effort`/`maxTurns`/`disallowedTools` (v2.1.78)

CC 2.1.78 officially supported `effort`, `maxTurns`, `disallowedTools` in plugin agent definition frontmatter.

**Harness current status**:
- `maxTurns`: Already implemented since v3.10.4 (Worker: 100, Reviewer: 50, Scaffolder: 75)
- `disallowedTools`: Worker `[Agent]`, Reviewer `[Write, Edit, Bash, Agent]` implemented
- `effort`: Not yet used. Can be added to Worker/Reviewer definitions to declaratively control default thinking level

### `deny: ["mcp__*"]` fix (v2.1.78)

CC 2.1.78 fixed settings.json `deny` permission rules to work correctly against MCP server tools.

**Harness usage**:
- Codex MCP blocking recommended in `.claude/rules/codex-cli-only.md` can migrate from hook-based to settings.json `deny`
- `"permissions": { "deny": ["mcp__codex__*"] }` is the clean pattern

### `--console` auth flag (v2.1.79)

CC 2.1.79 added `claude auth login --console` flag for Anthropic Console API billing authentication.

### SessionEnd hooks `/resume` fix (v2.1.79)

CC 2.1.79 fixed `SessionEnd` hooks to fire correctly during interactive `/resume` session switching. Previously SessionEnd didn't fire on session switch, causing cleanup processing to be skipped.

### `PermissionDenied` hook event (v2.1.89)

CC 2.1.89 added `PermissionDenied` hook firing when auto mode classifier denies a command. Returning `{retry: true}` communicates retry possibility to the model. Denied commands also appear in `/permissions` -> Recent tab.

**Harness usage**:
- `permission-denied-handler.sh` implemented to log denial events to `permission-denied.jsonl` as telemetry
- When Breezing Worker is denied, Lead receives `systemMessage` notification prompting alternative approach consideration
- `agent_id` / `agent_type` fields used to track which agent was denied what

**User experience improvement**:
- Before: Auto mode denials were only notified, not recorded, making the same denials likely to repeat
- After: Denial patterns accumulate, and in Breezing, Lead can immediately recognize and respond

### `"defer"` permission decision (v2.1.89)

CC 2.1.89 enabled returning `"defer"` permission decision from PreToolUse hooks. In headless sessions (`-p` mode), when a hook returns defer, the session pauses and the hook is re-evaluated on `claude -p --resume`.

**Harness usage potential**:
- Safety valve for when Breezing Worker encounters judgment-difficult operations like production environment writes or external service requests
- Add "defer conditions" to `pre-tool.sh` guardrail for specific patterns to pause Worker -> Lead decides
- Currently documentation only. Specific defer rules to be designed after operational pattern accumulation

### Hook output >50K disk save (v2.1.89)

CC 2.1.89 saves hook output exceeding 50K characters to disk instead of direct context injection, referenced as file path + preview.

**Harness impact**:
- Hooks potentially producing large output (quality-pack, ci-status-checker, etc.) should be designed with this behavior in mind
- Current Harness hooks have lightweight output, so direct impact is small; documented as design constraint for future extensions

### PreToolUse exit 2 JSON fix (v2.1.90)

CC 2.1.90 fixed blocking behavior when PreToolUse hooks output JSON to stdout and exit with code 2. Previously this pattern didn't correctly block.

**Harness impact**:
- `pre-tool.sh` uses JSON + exit 2 pattern for denials, working more reliably since v2.1.90
- Cases where guardrails "issued deny but tool was executed" may have been caused by this bug

## Related Documentation

- [CLAUDE.md](../CLAUDE.md) - Development guide (Feature Table summary version)
- [CLAUDE-skill-catalog.md](./CLAUDE-skill-catalog.md) - Skill catalog
- [CLAUDE-commands.md](./CLAUDE-commands.md) - Command reference
- [ARCHITECTURE.md](./ARCHITECTURE.md) - Architecture overview
