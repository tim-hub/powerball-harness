---
name: team-composition
description: "Reference guide for the Harness 3-agent team — Lead/Worker/Reviewer structure, breezing execution flow, permission settings, and Codex CLI mapping. Loaded by harness-work --breezing mode."
---

# Team Composition

3-agent composition for Harness.
Consolidated from 11 agents to 3 agents.

## Team Structure Diagram

```
Lead (Execute skill's --breezing mode) - orchestration only
  |
  +-- Worker (powerball-harness:worker)
  |     Implementation + preflight self-check + build verification + commit preparation
  |     * In --codex mode, delegates to Codex via official plugin
  |
  +-- [Worker #2] (powerball-harness:worker)
  |     Parallel execution of independent tasks
  |
  +-- Reviewer (powerball-harness:reviewer)
        Independent verdict across static / runtime / browser
        REQUEST_CHANGES -> Lead creates fix tasks
```

## Legacy Agent Mapping

| Legacy Agent | Current Agent |
|--------------|--------------|
| task-worker | worker |
| codex-implementer | worker (--codex included) |
| error-recovery | worker (error recovery included) |
| code-reviewer | reviewer |
| plan-critic | reviewer (plan type) |
| plan-analyst | reviewer (scope type) |
| project-analyzer | scaffolder |
| project-scaffolder | scaffolder |
| project-state-updater | scaffolder |
| ci-cd-fixer | worker (CI recovery included) |
| video-scene-generator | extensions/generate-video (separate) |

## Role Definitions

### Lead (Internal to Execute Skill)

| Item | Setting |
|------|------|
| **Phase A** | Preparation and task decomposition |
| **Phase B** | Delegate + review — Worker spawn / review execution / SendMessage / cherry-pick |
| **Phase C** | Rich completion report, Plans.md final update |
| **Prohibited** | Direct Write/Edit during Phase B (implementation is delegated to Worker). However, Bash may be used for review (companion review) and cherry-pick |

### Worker

| Item | Setting |
|------|------|
| **subagent_type** | `powerball-harness:worker` |
| **Model** | sonnet |
| **Count** | 1-3 (based on number of independent tasks) |
| **Tools** | Read, Write, Edit, Bash, Grep, Glob |
| **Prohibited** | Task (recursion prevention) |
| **Responsibilities** | Implementation -> preflight self-check -> CI verification -> worktree commit (does not reflect to main in Breezing mode) |
| **Error recovery** | Up to 3 times. Escalation after 3 failures |

### Reviewer

| Item | Setting |
|------|------|
| **subagent_type** | `powerball-harness:reviewer` |
| **Model** | sonnet |
| **Count** | 1 |
| **Tools** | Read, Grep, Glob (default for static profile) |
| **Prohibited** | Write, Edit, Task |
| **Responsibilities** | Returns verdict across static/runtime/browser based on `sprint-contract` |
| **Verdict** | APPROVE / REQUEST_CHANGES |

### Quality Language

The following short criteria phrases are used in common at the start of Worker / Reviewer / Lead:

- Worker: "Make changes that are testable and easy to fix later"
- Reviewer: "Don't raise concerns without evidence; only block on critical issues"
- Lead: "Stay within scope while leaving things in a form that enables learning from reviews"

When drift or oversights are found in reviews, record them in `review-calibration.jsonl`
as one of `false_positive`, `false_negative`, `missed_bug`, `overstrict_rule`,
and regenerate the few-shot bank.

### Scaffolder (Setup Only)

| Item | Setting |
|------|------|
| **subagent_type** | `powerball-harness:scaffolder` |
| **Model** | sonnet |
| **Count** | 1 |
| **Tools** | Read, Write, Edit, Bash, Grep, Glob |
| **Responsibilities** | Project analysis, scaffolding, state updates |

## Execution Flow (v3.12+ Review Loop Integration)

```
Phase A: Lead decomposes tasks, analyzes dependency graph, scores effort, generates sprint-contract
    |
Phase B: Sequential execution per task (in dependency order)
    |
    B-1. Worker spawn (mode: breezing, isolation: worktree)
         Worker: Implementation -> preflight self-check -> worktree commit -> return results to Lead
    |
    B-2. Lead executes review
         Selects sprint-contract's `reviewer_profile`
         static: Codex exec / Reviewer agent
         runtime: Executes contract's `runtime_validation`
         browser: Delegates to browser-capable evaluator
         Threshold criteria: critical/major -> REQUEST_CHANGES, minor only -> APPROVE
    |
    B-3. On REQUEST_CHANGES: Fix loop (up to 3 times)
         Lead -> SendMessage(to: worker_id) sends feedback
         Worker: Fix -> git commit --amend -> return updated hash
         Lead: Re-review
    |
    B-4. APPROVE -> Lead cherry-picks to main
         git cherry-pick --no-commit {worktree_commit}
         git commit -m "{task description}"
         Plans.md: cc:done [{hash}]
    |
Phase C: Lead outputs rich completion report, final Plans.md confirmation
```

### SendMessage Pattern (Fix Loop)

Syntax when Lead instructs Worker to make fixes:

```
SendMessage(
    to: "{worker_agent_id}",
    message: "Please fix the following critical/major issues:\n\n{issues}\n\nAfter fixing, please git commit --amend and return completion."
)
```

Worker-side receive processing:
1. Receive SendMessage -> parse feedback content
2. Fix the relevant code
3. Update the worktree commit with `git commit --amend`
4. Return the updated commit hash to Lead

### cherry-pick Pattern (After APPROVE)

Lead incorporates the Worker's worktree commit into main:

```bash
# Cherry-pick worktree commit to main
git cherry-pick --no-commit {worktree_commit_hash}
# Lead controls the commit message
git commit -m "feat: {task_description}"
```

> **Note**: By using `--no-commit` to stage changes before committing,
> Lead can control the commit message in a unified format.

### Nested Teammate Policy (v2.1.69)

In CC 2.1.69, nested teammate spawning (nested teammates) is blocked at the platform level.
Harness minimizes redundant prevention wording and standardizes on the following operations:

1. Only Lead spawns teammates
2. Worker/Reviewer prompts focus on "implementation/review responsibilities"
3. Nested prevention relies on the official guard rather than adding hooks (simplifies operations)

## Permission Settings (bypassPermissions / permissionMode)

Teammates run in the background without UI, so explicit permission mode configuration is required.

### v2.1.72+ Recommended: `permissionMode` in frontmatter

Official documentation now documents `permissionMode` as a formal field in agent frontmatter.
**Declaration at the definition level is recommended** over specifying `mode` at spawn time:

```yaml
# agents/worker.md frontmatter
permissionMode: bypassPermissions
```

**Benefit**: Embeds the permission mode in the agent definition itself, independent of the spawn prompt.
Safe even if Lead's spawn code forgets to pass `mode`.

### Safety Layers (Defense in Depth)

1. `permissionMode: bypassPermissions` — declared in frontmatter
2. `disallowedTools` restricts tools
3. PreToolUse hooks maintain guardrails
4. Lead constantly monitors
5. `Agent(worker, reviewer)` limits the types of agents that can be spawned

### Auto Mode (Rollout Target)

A new permission model offered by Anthropic as a safe alternative to `bypassPermissions`.
Claude automatically makes permission decisions, with built-in prompt injection countermeasures.

| Aspect | bypassPermissions | Auto Mode |
|------|-------------------|-----------|
| Permission decisions | All tools unconditionally allowed | Claude decides automatically |
| Safety layers | hooks + disallowedTools | Built-in countermeasures + hooks + disallowedTools |
| Token cost | No additional cost | Slight increase |
| Latency | No additional cost | Slight increase |
| Teammate compatibility | Current shipped default | Validation target when parent session's permission mode is compatible |

#### Current Handling

The shipped default for `/breezing` and `/harness-work --breezing` remains `bypassPermissions` for now.
The `--auto-mode` flag is treated as an opt-in for trying out the Auto Mode rollout, limited to cases where the parent session's permission mode is compatible:

```bash
/breezing all                 # Current default (bypassPermissions) runs all tasks to completion
/breezing --auto-mode all     # Try Auto Mode rollout on compatible parent sessions
/execute --breezing all
```

**Constraint**: If the parent session or subagent frontmatter remains `bypassPermissions`, that permission mode takes priority.
Therefore, to truly make Auto Mode the default, the permission design of the parent session execution path needs to be revisited, not just the teammate execution path. hooks and disallowedTools are maintained as-is.

#### Configuration Policy

| Layer | Adopted Value | Reason |
|---------|--------|------|
| project template (`permissions.defaultMode`) | `bypassPermissions` | `autoMode` is not included in the official docs' documented permission modes |
| agent frontmatter (`permissionMode`) | `bypassPermissions` | Frontmatter also declares only documented permission modes |
| teammate execution path | `bypassPermissions` (current) | To align shipped default with actual permission inheritance |
| `--auto-mode` flag | opt-in rollout | To enable safely after revisiting parent session design |

This separation allows distribution templates to avoid undocumented configuration values while correctly describing the current shipped behavior. Auto Mode will be reconsidered during the next phase of parent session design changes.

### Official Agent Teams Documentation

Agent Teams has been officially documented as an experimental feature.
Activation requires the `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS` environment variable:

```json
// settings.json
{
  "env": {
    "CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS": "1"
  }
}
```

**Impact on Harness**:
- The `breezing` skill depends on Agent Teams -> Add environment variable check during setup
- Official documentation now explicitly documents the `teammateMode` setting (`"in-process"` | `"tmux"` | `"auto"`)
- `TeammateIdle` / `TaskCompleted` with `{"continue": false}` has stabilized as official specification

## Official Agent Teams Best Practices Alignment (2026-03)

Alignment status with best practices from the official Claude Code documentation `agent-teams.md`.

### Task Granularity Guidelines

Official recommendation: **5-6 tasks per teammate**. Harness Lead uses this granularity as a guideline during task decomposition.

| Granularity | Assessment | Example |
|------|------|-----|
| Too small | Coordination > implementation cost | Single-line fix, comment addition |
| Appropriate | Self-contained unit with clear deliverable | Function implementation, test file creation, review |
| Too large | Runs for extended time without check-in | Full module redesign |

### `teammateMode` Setting

Officially supported display modes:

| Mode | Behavior | Recommended Environment |
|--------|------|----------|
| `"auto"` | Split if inside tmux session, otherwise in-process | Default |
| `"in-process"` | Manage all teammates in the same terminal | VS Code integrated terminal |
| `"tmux"` | Individual pane for each teammate | iTerm2 / tmux users |

```json
// settings.json
{ "teammateMode": "in-process" }
```

### Plan Approval Pattern

Official "Require plan approval for teammates" pattern:

```
Lead: "Spawn an architect teammate. Require plan approval before changes."
  -> Teammate investigates and drafts plan in plan mode
  -> Sends plan_approval_request to Lead
  -> Lead APPROVEs -> Teammate begins implementation
  -> Lead REJECTs + feedback -> Teammate revises plan
```

In Harness, this can be used complementarily with Reviewer's `REQUEST_CHANGES` -> Worker fix loop.
Recommended to require plan approval for Worker spawns on complex architecture changes.

### Quality Gate Hooks

Alignment with official hook events:

| Hook | Harness Implementation | Official Documentation |
|------|-------------|--------------|
| `TeammateIdle` | `teammate-idle.sh` (implemented) | exit 2 for feedback + continue instruction |
| `TaskCompleted` | `task-completed.sh` (implemented) | exit 2 for completion rejection + feedback |
| `SubagentStart` | Implemented (subagent-tracker + matcher: worker/reviewer/scaffolder/video-scene-generator) | Filter by agent type in settings.json |
| `SubagentStop` | Implemented (subagent-tracker + matcher + agent frontmatter Stop hook) | Two-layer monitoring via settings.json + frontmatter |

### Team Size Guidelines

Official recommendation: **3-5 teammates**. This aligns with Harness's current composition (Worker 1-3 + Reviewer 1).

> "Three focused teammates often outperform five scattered ones." — from official documentation

## Codex CLI Environment

In Codex CLI environments, Claude Code's Agent/SendMessage API is unavailable.
Use native subagent API (`spawn_agent`, `send_input`, etc.) as alternatives.
See full API mapping, companion patterns, and memory config in
[references/team-composition-codex.md](references/team-composition-codex.md).


## Sandboxing Integration (Phased Rollout)

Claude Code's `/sandbox` feature provides OS-level filesystem/network isolation.
Introduced as an **additional safety layer** on top of the current `bypassPermissions` + hooks defense-in-depth.

### Current vs Sandboxing

| Aspect | bypassPermissions + hooks | Sandbox auto-allow |
|------|--------------------------|-------------------|
| Granularity | Tool-level (determined by hooks) | File path/domain level (OS-enforced) |
| Implementation layer | Claude Code permission system | macOS Seatbelt / Linux bubblewrap |
| Prompt injection | Partially defended by hooks | Fully defended at OS level |
| Worker freedom | All Bash allowed (guarded by hooks) | Only defined paths/domains |
| Token cost | None | None |

### Application Strategy for Worker

```json
// settings.json — Example sandbox configuration for Worker sessions
{
  "sandbox": {
    "enabled": true,
    "filesystem": {
      "allowWrite": [
        "/",
        "~/.claude",
        "//tmp"
      ]
    }
  }
}
```

- `allowWrite: ["/"]` is relative to the settings.json directory (project root)
- `~/.claude` is needed for Agent Memory writes
- `//tmp` is for build output and temporary files


### Lead Model Optimization with `opusplan`

The `opusplan` alias is ideal for Lead sessions:
- **Plan phase**: Opus for task decomposition and architecture decisions (high-quality reasoning)
- **Execute phase**: Sonnet for Worker coordination (cost-efficient)

```bash
# Use opusplan in breezing sessions
claude --model opusplan
/breezing all
```

### Worker Model Control with `CLAUDE_CODE_SUBAGENT_MODEL`

The environment variable `CLAUDE_CODE_SUBAGENT_MODEL` specifies the model for all subagents at once:

```bash
# Reduce costs in CI environment (run Worker/Reviewer with haiku)
export CLAUDE_CODE_SUBAGENT_MODEL=claude-haiku-4-5-20251001
```

> Priority between the agent definition's `model` field and this environment variable is unverified. Scheduled for Phase 2 validation.

## v2.1.68/v2.1.72 Effort Level Changes Impact

### Changes
- Opus 4.6 changed to **medium effort** default (v2.1.68)
- `ultrathink` keyword enables high effort (limited to 1 turn)
- Opus 4 / 4.1 removed from first-party API (auto-migrated to Opus 4.6)
- **v2.1.72**: `max` level removed. Simplified to 3 levels: `low(○)/medium(◐)/high(●)`. `/effort auto` to reset

### Impact on Team
- Worker (`model: sonnet`): Sonnet is not affected by effort levels. No change
- Reviewer (`model: sonnet`): Same. No change
- Lead (when using Opus): medium effort is default. Use ultrathink for complex task coordination
- Codex Worker: Effort control is Claude Code specific. Not applicable in Codex CLI

### Effort Injection Pattern
When Lead spawns Worker/Reviewer, `ultrathink` is added at the beginning of the spawn prompt based on the task's complexity score. See the "Effort Level Control" section in `skills/harness-work/SKILL.md` for details.

### v2.1.72 Agent Tool `model` Parameter Restored
The Agent tool's per-invocation `model` parameter has been restored. Separate from the agent definition's `model`, a temporary model specification is possible at spawn time.
- **Current**: Both Worker/Reviewer operate with fixed `model: sonnet`
- **Phase 2 consideration**: Dynamic model selection based on task characteristics (lightweight -> haiku, high quality -> opus)

### v2.1.72 `/clear` Preserves Background Agents
`/clear` now only stops foreground tasks. Background Workers survive even when Lead uses `/clear` during breezing team execution.

### v2.1.72 Parallel Tool Call Fix
Read/WebFetch/Glob failures no longer cancel sibling calls. Reliability of Worker's parallel file reads improved.
