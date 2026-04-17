# Harness Architecture

Last updated: 2026-04-17 (v4.6.1)

## 1. Overview

Harness is a Claude Code plugin for autonomous **Plan → Work → Review** workflows. It extends Claude Code with: a **Go-native guardrail engine** for runtime protection, **27 skills** for structured workflows, **7 agents** for multi-role task execution, and a **concurrent hook system** covering 27 event types.

## 2. Three-Layer Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│  Skills Layer  (harness/skills/ — 27 SKILL.md files)                │
│  Description-based auto-loading; each skill is a standalone unit     │
├─────────────────────────────────────────────────────────────────────┤
│  Agent Layer   (harness/agents/ — 7 agent definitions)              │
│  worker · reviewer · scaffolder · advisor · ci-cd-fixer · ...       │
├─────────────────────────────────────────────────────────────────────┤
│  Hook/Guardrail Layer  (harness/hooks/ + go/internal/)              │
│  27 event types → Go binary (CGO_ENABLED=0 static binary)           │
└─────────────────────────────────────────────────────────────────────┘
```

- **Skills Layer**: 27 standalone `SKILL.md` files, each with `description` (trigger matching), `allowed-tools`, and optional `references/` subdirectory. Auto-loaded by Claude Code based on task shape.
- **Agent Layer**: Markdown agent definitions spawned by skills via the Task tool. Roles: Worker (implement), Reviewer (approve/deny), Scaffolder (state), Advisor (read-only guidance, Opus model), CI-CD-Fixer, Error-Recovery.
- **Hook/Guardrail Layer**: `hooks.json` maps 27 CC event types to thin bash shims. Shims call `bin/harness hook <name>` which dispatches to the Go binary. Concurrent fan-out (`post-tool-batch`, `pre-tool-batch`) parallelizes independent hooks.

## 3. Directory Structure

```
powerball-harness/
├── .claude-plugin/         # Plugin metadata (marketplace.json)
├── harness/                # Plugin payload (distributed to users)
│   ├── skills/             # 27 skills (SKILL.md + references/)
│   ├── agents/             # 7 agent definitions
│   ├── hooks/              # hooks.json + thin bash shims
│   ├── scripts/            # 90+ shell scripts
│   ├── bin/                # Pre-compiled harness binaries (gitignored)
│   ├── templates/          # Managed-block templates for setup
│   └── VERSION             # Current version (e.g. 4.6.1)
├── go/                     # Go guardrail engine source
│   ├── cmd/harness/        # CLI entry point
│   └── internal/
│       ├── guardrail/      # R01–R13 rule evaluation
│       ├── hookhandler/    # Concurrent fan-out orchestration
│       ├── session/        # Session state management
│       ├── breezing/       # Breezing runtime primitives
│       └── sprint/         # Sprint Contract Go package
├── local-scripts/          # Dev-only tooling (not distributed)
└── tests/                  # validate-plugin.sh and CI gates
```

## 4. Key Components

### 4.1. Skills

27 skills covering the 5-verb workflow (`plan/work/review/release/setup`) plus specialized skills (session management, CI, browser automation, CRUD scaffolding, etc.). Each skill is independently auto-loaded based on its `description` field — no parent/child routing.

### 4.2. Go Guardrail Engine

13 rules (R01–R13) in `go/internal/guardrail/rules.go`, evaluated in priority order. First matching rule wins. Compiled as a static binary (CGO_ENABLED=0) for darwin-arm64, darwin-amd64, linux-amd64. Fail-open: if binary is missing, hooks exit 0 with a one-time warning.

Rule actions: **Deny** (block), **Ask** (require confirmation), **Approve + Warning** (allow with system message).

### 4.3. Hook System

27 event types covered via `harness/hooks/hooks.json`. PostToolUse uses goroutine fan-out (`post-tool-batch`) for concurrent side-effect execution. PreToolUse uses deny-wins merge (`pre-tool-batch`) so any guardrail denial blocks the tool call. Session lifecycle hooks (SessionStart, Stop, PreCompact, PostCompact) manage state and memory-bridge sync.

### 4.4. Agents and Breezing

`harness-work --parallel` spawns Worker agents via Task tool to implement Plans.md tasks concurrently. `breezing` runs the full team (Worker + Reviewer + Advisor) end-to-end. `harness-loop` uses ScheduleWakeup for autonomous multi-session runs with sprint-contract gating and plateau detection.
