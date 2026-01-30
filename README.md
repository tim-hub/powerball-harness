<p align="center">
  <img src="docs/images/claude-harness-logo-with-text.png" alt="Claude Harness" width="400">
</p>

<p align="center">
  <strong>Turn Claude Code into a self-correcting development team.</strong>
</p>

<p align="center">
  <a href="VERSION"><img src="https://img.shields.io/badge/version-2.14.10-blue.svg" alt="Version"></a>
  <a href="LICENSE.md"><img src="https://img.shields.io/badge/License-MIT-green.svg" alt="License"></a>
  <a href="docs/CLAUDE_CODE_COMPATIBILITY.md"><img src="https://img.shields.io/badge/Claude_Code-v2.1.21+-purple.svg" alt="Claude Code"></a>
</p>

<p align="center">
  English | <a href="README_ja.md">日本語</a>
</p>

---

## The Problem

**Claude is brilliant. But it forgets. It wanders. It breaks things.**

| Issue | What Happens |
|-------|--------------|
| **Forgets** | Past decisions vanish between sessions |
| **Wanders** | Starts coding without a plan, loses direction |
| **Breaks** | Skips tests, bypasses lint, ships bugs |
| **Rushes** | Takes shortcuts when stuck—test tampering, empty catch blocks |

Sound familiar?

---

## The Solution

**Claude Harness wraps Claude Code in guardrails that enforce discipline.**

```
Plan  →  Work  →  Review  →  Commit
```

Three core commands to remember.

```bash
/plan-with-agent   # Brainstorm → Structured plan
/work              # Execute with parallel workers + self-review
/harness-review    # 4-perspective parallel code review
```

**Result:** Production-ready code, not prototypes.

---

## Quick Start

### 10-Second Install

```bash
# In any project directory
claude

# Then run:
/plugin marketplace add Chachamaru127/claude-code-harness
/plugin install claude-code-harness@claude-code-harness-marketplace
/harness-init
```

**Done.** Start with `/plan-with-agent`.

<details>
<summary>Alternative: One-liner script</summary>

```bash
curl -fsSL https://raw.githubusercontent.com/Chachamaru127/claude-code-harness/main/scripts/quick-install.sh | bash
```

With dev tools (AST-Grep + LSP):
```bash
curl -fsSL https://raw.githubusercontent.com/Chachamaru127/claude-code-harness/main/scripts/quick-install.sh | bash -s -- --with-dev-tools
```

</details>

<details>
<summary>Alternative: Local clone</summary>

```bash
git clone https://github.com/Chachamaru127/claude-code-harness.git ~/claude-plugins/claude-code-harness
cd /path/to/your-project
claude --plugin-dir ~/claude-plugins/claude-code-harness
```

</details>

---

## Before → After

| Before (Raw Claude Code) | After (With Harness) |
|--------------------------|----------------------|
| Starts coding immediately | Plans first, then executes |
| Reviews only if you ask | Auto-reviews every change |
| Forgets past decisions | SSOT files preserve context |
| `rm -rf` runs without warning | Dangerous commands require confirmation |
| Manual `git commit` after work | Auto-commits when review passes |
| One task at a time | Parallel workers for speed |

---

## Key Features

### 🎯 Plan → Work → Review Cycle

Every idea goes through the same loop:

1. **Plan** — `/plan-with-agent` turns vague ideas into `Plans.md`
2. **Work** — `/work` executes tasks with parallel workers
3. **Review** — `/harness-review` runs 4 perspectives in parallel

### 🛡️ Safety Hooks

| Protected | Action |
|-----------|--------|
| `.git/`, `.env`, secrets | Write blocked |
| `rm -rf`, `sudo`, `git push --force` | Confirmation required |
| `git status`, `npm test` | Auto-allowed |

### 🧠 Persistent Memory

- **SSOT files**: `decisions.md` (why) + `patterns.md` (how)
- **Claude-mem integration**: Past learnings survive across sessions
- **Session resume**: Pick up exactly where you left off

### ⚡ Parallel Execution

```bash
/work --parallel 5   # 5 workers in parallel
```

Each worker implements and self-reviews. Auto-commit after global review passes.

### 🔍 4-Perspective Parallel Code Review

```bash
/harness-review
```

Security, performance, accessibility, quality—4 perspectives review simultaneously in parallel. Add [Codex](https://github.com/openai/codex) for second opinions (selects 4 relevant experts from 16 specialist types).

### 🔧 Code Intelligence

```bash
/dev-tools-setup   # One-time setup
```

AST-Grep + LSP for structural search and semantic analysis.

---

## Who Is This For?

| You | Benefit |
|-----|---------|
| **Solo developers** | Ship faster without sacrificing quality |
| **Freelancers** | Deliver review reports as proof of quality |
| **VibeCoder** | Build apps with natural language |
| **Cursor users** | 2-Agent workflow: Cursor plans, Claude Code implements |

---

## Commands

### Core Workflow

| Command | Purpose |
|---------|---------|
| `/plan-with-agent` | Turn ideas into actionable plans |
| `/work` | Execute tasks from Plans.md |
| `/harness-review` | 4-perspective parallel code review |
| `/sync-status` | Check progress, suggest next action |

### Setup & Ops

| Command | Purpose |
|---------|---------|
| `/harness-init` | Initialize project |
| `/harness-update` | Update plugin |
| `/dev-tools-setup` | Setup AST-Grep + LSP |
| `/skill-list` | Show all 28 skill categories |

### 2-Agent (Cursor)

| Command | Purpose |
|---------|---------|
| `/handoff-to-cursor` | Report completion to PM |

---

## Skills

Skills auto-trigger based on your request:

| Skill | Triggers |
|-------|----------|
| `impl` | "implement", "add feature", "build" |
| `review` | "review", "check security", "audit" |
| `verify` | "build", "fix errors", "recover" |
| `auth` | "login", "Stripe", "payment" |
| `deploy` | "deploy", "Vercel", "production" |
| `ui` | "hero section", "component", "form" |

**28 skill categories.** Run `/skill-list` to see all.

---

## Architecture

```
claude-code-harness/
├── commands/     # 31 slash commands
├── skills/       # 28 skill categories
├── agents/       # 8 sub-agents (parallel workers)
├── hooks/        # Safety & automation
├── scripts/      # Guard scripts
└── templates/    # Generation templates
```

---

## Documentation

| Guide | Description |
|-------|-------------|
| [Changelog](CHANGELOG.md) | What's new in each version |
| [Claude Code Compatibility](docs/CLAUDE_CODE_COMPATIBILITY.md) | Version requirements |
| [Cursor Integration](docs/CURSOR_INTEGRATION.md) | 2-Agent workflow setup |
| [OpenCode Compatibility](docs/OPENCODE_COMPATIBILITY.md) | Use with other LLMs |

---

## Requirements

- **Claude Code v2.1.21+** (recommended)
- See [compatibility guide](docs/CLAUDE_CODE_COMPATIBILITY.md) for details

---

## Acknowledgments

- **Hierarchical Skill Structure**: [AI Masao](https://note.com/masa_wunder)
- **Test Tampering Prevention**: [Beagle](https://github.com/beagleworks) (Claude Code Meetup Tokyo 2025.12.22)

---

## License

**MIT License** — Free to use, modify, and commercialize.

[English](LICENSE.md) | [日本語](LICENSE.ja.md)
