# CLAUDE_CODE_SIMPLE Mode Compatibility

This document describes how `CLAUDE_CODE_SIMPLE` mode (CC v2.1.50+) affects claude-code-harness functionality.

## Overview

When `CLAUDE_CODE_SIMPLE=1` is set, Claude Code strips **skills**, **memory**, and **agents** from the session. Only **hooks** continue to operate. This means the majority of Harness automation features become unavailable.

## Impact Summary

| Category | Status in SIMPLE Mode | Count | Severity |
|----------|----------------------|-------|----------|
| **Skills** | Stripped (unavailable) | 37 SKILL.md | CRITICAL |
| **Agents** | Stripped (unavailable) | 11 agent files | CRITICAL |
| **Memory** | Unavailable | project memory system | HIGH |
| **Hooks** | Still operational | hooks.json + scripts | OK |
| **Direct Bash** | Still operational | all Bash capabilities | OK |

## What Still Works

### Hooks (Fully Operational)

All hooks defined in `hooks/hooks.json` continue to function:

| Hook Event | Handlers | Purpose |
|-----------|----------|---------|
| `PreToolUse` | pretooluse-guard, inbox-check, browser-guide | Safety guards, file protection |
| `PostToolUse` | quality-pack, tampering-detector, ci-status-checker, agent-trace | Quality gates, CI monitoring |
| `UserPromptSubmit` | inject-policy, track-command, breezing-signal-injector | Policy injection |
| `SessionStart` | session-init, session-monitor | Session initialization |
| `Stop` | session-summary, stop-session-evaluator | Session evaluation |
| `ConfigChange` | config-change | Configuration auditing |
| `Setup` | setup-hook | Init/maintenance |
| `PreCompact` | pre-compact-save | State preservation |
| `SessionEnd` | session-cleanup | Cleanup |

### Basic Claude Code Features

- File reading/writing/editing (with hook guards)
- Git operations (with safety checks from pretooluse-guard)
- Web search and fetching
- TodoWrite for task management

## What Breaks

### Skills (37 total — ALL unavailable)

| Skill | Impact | Description |
|-------|--------|-------------|
| `/work` | **CRITICAL** | Core implementation workflow — cannot spawn task-workers |
| `/breezing` | **CRITICAL** | Agent Teams orchestration — multi-agent workflows fail |
| `/planning` | **CRITICAL** | Plan creation with Task tool |
| `/harness-review` | **CRITICAL** | Code review orchestration with parallel agents |
| `/impl` | HIGH | Implementation skill (used by task-worker) |
| `/verify` | HIGH | Verification skill (used by task-worker) |
| `/setup` | MEDIUM | Setup automation |
| `/ci` | MEDIUM | CI/CD automation |
| `/codex-review` | MEDIUM | Codex integration |
| `/memory` | HIGH | Project memory integration |
| `/session-*` | MEDIUM | Session management (4 skills) |
| `/maintenance` | LOW | Plugin maintenance |
| `/troubleshoot` | LOW | Diagnostic skill |
| Others (22+) | LOW-MEDIUM | Specialized features |

### Agents (11 total — ALL unavailable)

| Agent | Impact | Description |
|-------|--------|-------------|
| `task-worker` | **CRITICAL** | Primary implementation agent — spawned by `/work` and `/breezing` |
| `code-reviewer` | **CRITICAL** | Parallel 4-point code review |
| `codex-implementer` | HIGH | Codex-delegated implementation |
| `ci-cd-fixer` | HIGH | Auto-recovery from CI failures |
| `project-analyzer` | MEDIUM | Project analysis |
| `project-scaffolder` | MEDIUM | Project scaffolding |
| `project-state-updater` | MEDIUM | State management |
| `plan-analyst` | MEDIUM | Planning analysis (Phase 0) |
| `plan-critic` | MEDIUM | Red Team review (Phase 0) |
| `video-scene-generator` | LOW | Async video generation |
| Others | LOW | Specialty agents |

### Memory System

- **Project memory** (`memory: project` in agents) — unavailable
- **Unified Memory Gate** (harness_mem_*) — cannot execute
- **Cross-session learning** — context lost between sessions

### Core Workflow Failures

| Feature | Why It Fails | Workaround |
|---------|-------------|------------|
| `/work` command | Skill + Task tool + agent invocation required | Manual implementation only |
| `/breezing` (Agent Teams) | Skill + Task tool + multiple agents | Not possible in SIMPLE mode |
| `/harness-review` | Review skill + parallel review agents | Manual code review |
| Parallel task execution | Task tool spawns task-worker agents | Single-threaded manual work |
| Auto-commit workflow | Handoff skill invocation | Manual `git add && git commit` |
| CI auto-recovery | ci-cd-fixer agent spawning | Manual CI fix |

## Detection

### How Harness Detects SIMPLE Mode

Harness detects SIMPLE mode at two points:

1. **SessionStart hook** (`scripts/session-init.sh`): Checks `CLAUDE_CODE_SIMPLE` env var and displays a warning banner in both stderr and `additionalContext`
2. **Setup hook** (`scripts/setup-hook.sh`): Checks on init/maintenance and includes warning in output

### Detection Utility

```bash
# Source the utility
source scripts/check-simple-mode.sh

# Check mode
if is_simple_mode; then
  echo "SIMPLE mode active"
fi

# Get localized warning message
simple_mode_warning "en"  # English
simple_mode_warning "ja"  # Japanese
```

### Environment Variable

```bash
# SIMPLE mode is activated by:
CLAUDE_CODE_SIMPLE=1

# Normal mode (default):
# CLAUDE_CODE_SIMPLE is unset or "0"
```

## Recommendations

### For Users in SIMPLE Mode

1. **Use basic Claude Code** without Harness automation features
2. **Hooks still protect you** — file guards, safety checks, and quality gates remain active
3. **Manual workflows** — use direct Claude Code commands instead of Harness skills
4. **Consider disabling SIMPLE mode** if you need Harness features

### For Developers Extending Harness

1. **Always check for SIMPLE mode** before referencing skills or agents
2. **Use `check-simple-mode.sh`** utility for consistent detection
3. **Hooks are the only reliable extension point** in SIMPLE mode
4. **Do not assume skills/agents exist** — graceful degradation is required

## Version Requirements

| Feature | Required CC Version |
|---------|-------------------|
| `CLAUDE_CODE_SIMPLE` mode | v2.1.50+ |
| SIMPLE mode detection in Harness | v2.25.0+ |
| Hook-only operation | All CC versions |

## Related Documentation

- [CLAUDE_CODE_COMPATIBILITY.md](./CLAUDE_CODE_COMPATIBILITY.md) — Full compatibility matrix
- [check-simple-mode.sh](../scripts/check-simple-mode.sh) — Detection utility
