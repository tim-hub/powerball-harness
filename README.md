<p align="center">
  <img src="docs/images/claude-harness-logo-with-text.png" alt="Claude Harness" width="400">
</p>

<p align="center">
  <strong>Plan. Work. Review. Ship.</strong><br>
  <em>Turn Claude Code into a disciplined development partner.</em>
</p>

<p align="center">
  <a href="https://github.com/Chachamaru127/claude-code-harness/releases/latest"><img src="https://img.shields.io/github/v/release/Chachamaru127/claude-code-harness?display_name=tag&sort=semver" alt="Latest Release"></a>
  <a href="LICENSE.md"><img src="https://img.shields.io/badge/License-MIT-green.svg" alt="License"></a>
  <a href="docs/CLAUDE_CODE_COMPATIBILITY.md"><img src="https://img.shields.io/badge/Claude_Code-v2.1+-purple.svg" alt="Claude Code"></a>
  <img src="https://img.shields.io/badge/Skills-5_Verbs-orange.svg" alt="Skills">
  <img src="https://img.shields.io/badge/Core-Go_Native-00ADD8.svg" alt="Go Core">
  <img src="https://img.shields.io/badge/v4.0-Hokage-FF4500.svg" alt="Hokage">
</p>

<p align="center">
  English | <a href="README_ja.md">日本語</a>
</p>

<p align="center">
  <img src="docs/images/hokage/hokage-hero.jpg" alt="Hokage v4.0 — The Silent Blade" width="860">
</p>

---

## v4.0 "Hokage" — What's New

> **Go-native engine. 25x faster hooks. Zero Node.js dependency.**

Every tool call Claude makes passes through Harness hooks. In v3, each pass cost 40-60ms of bash + Node.js overhead — a subtle drag you feel across hundreds of calls per session. v4 replaces the entire stack with a single Go binary:

| | v3 (bash + Node.js) | v4 "Hokage" (Go) |
|---|---|---|
| **PreToolUse** | 40-60ms | **10ms** |
| **SessionStart** | 500-800ms | **10-30ms** |
| **PostToolUse** | 20-30ms | **10ms** |
| **Node.js** | Required (18+) | **Not needed** |

**What you'll notice:**
- The micro-pauses between tool calls disappear — Claude feels more responsive
- No more `npm install` or Node.js version issues on setup
- Session startup is instant (10-30ms vs almost a second)
- Optional [harness-mem](https://github.com/Chachamaru127/harness-mem) integration: sessions remember what you worked on last time

Just update the plugin — no configuration changes needed:
```
/plugin update claude-code-harness
```

---

## Why Harness?

Claude Code is powerful. Harness turns that raw capability into a delivery loop that is easier to trust and harder to derail.

<p align="center">
  <img src="assets/readme-visuals-en/generated/why-harness-pillars.svg" alt="What changes with Claude Harness: shared plan, runtime guardrails, and rerunnable validation" width="860">
</p>

The 5 verb skills keep setup, plan, work, review, and release on one path. The Go-native guardrail engine protects execution with sub-10ms response, and validation can be rerun when you need proof.

## Compared With Popular Claude Code Harnesses

What matters here is not the theoretical ceiling of Claude Code. It is what becomes the **default operating model** once you install a harness.

This is a **user-facing workflow** snapshot as of **2026-03-06**, not a popularity contest.
Full notes and source links: [docs/github-harness-plugin-benchmark.md](docs/github-harness-plugin-benchmark.md)

The card below focuses on what becomes the default operating path after install.

<p align="center">
  <img src="assets/readme-visuals-en/generated/harness-feature-matrix.svg" alt="How the default workflow changes after installing Claude Harness, Superpowers, or cc-sdd" width="860">
</p>

Claude Harness is the clearest fit if you want the default path itself to stay planned, guarded, reviewed, and rerunnable.

Supported baseline and latest verified snapshot: see [Claude Code Compatibility](docs/CLAUDE_CODE_COMPATIBILITY.md).

---

## Requirements

- **Claude Code v2.1+** ([Install Guide](https://docs.anthropic.com/claude-code))
- **No Node.js required** (v4.0 Hokage uses Go-native engine)

---

## Who Is This For?

| You Are | Harness Helps You |
|---------|-------------------|
| **Developer** | Ship faster with built-in QA |
| **Freelancer** | Deliver review reports to clients |
| **Indie Hacker** | Move fast without breaking things |
| **VibeCoder** | Build apps with natural language |
| **Team Lead** | Enforce standards across projects |

---

## Install in 30 Seconds

```bash
# Start Claude Code in your project
claude

# Add the marketplace & install
/plugin marketplace add Chachamaru127/claude-code-harness
/plugin install claude-code-harness@claude-code-harness-marketplace

# Initialize your project
/harness-setup
```

That's it. Start with `/harness-plan`.

---

## 🪄 TL;DR: Verified Work All

**Don't want to read all this?** Just type:

```
/harness-work all
```

**One command runs the full loop after plan approval.** Plan → Parallel Implementation → Review → Commit.

<p align="center">
  <img src="assets/readme-visuals-en/work-all-flow.svg" alt="/work all pipeline" width="700">
</p>

> ⚠️ **Experimental workflow**: Once you approve the plan, Claude runs to completion. Validate the success/failure contract in [docs/evidence/work-all.md](docs/evidence/work-all.md) before depending on it in production.

---

## The 5 Verb Workflow

<p align="center">
  <img src="assets/readme-visuals-en/generated/core-loop.svg" alt="Plan → Work → Review cycle" width="560">
</p>

### 0. Setup

```bash
/harness-setup
```

Bootstraps project files, rules, and command surfaces so the rest of the loop runs against the same conventions.

### 1. Plan

```bash
/harness-plan
```

> "I want a login form with email validation"

Harness creates `Plans.md` with clear acceptance criteria.

### 2. Work

```bash
/harness-work              # Auto-detect parallelism
/harness-work --parallel 5 # 5 workers simultaneously
```

Each worker implements, runs a preflight self-check, and waits for an independent review verdict before completion.

<p align="center">
  <img src="assets/readme-visuals-en/parallel-workers.svg" alt="Parallel workers" width="640">
</p>

### 3. Review

```bash
/harness-review
```

<p align="center">
  <img src="assets/readme-visuals-en/review-perspectives.svg" alt="4-perspective review" width="640">
</p>

| Perspective | Focus |
|-------------|-------|
| Security | Vulnerabilities, injection, auth |
| Performance | Bottlenecks, memory, scaling |
| Quality | Patterns, naming, maintainability |
| Accessibility | WCAG compliance, screen readers |

### 4. Release

```bash
/harness-release
```

Packages the verified result into CHANGELOG, tag, and release handoff steps after implementation and review are complete.

---

## Safety First

<p align="center">
  <img src="assets/readme-visuals-en/generated/safety-guardrails.svg" alt="Safety Protection System" width="640">
</p>

Harness v4 protects your codebase with a **Go-native guardrail engine** (`go/internal/guardrail/`) — 13 declarative rules (R01–R13), single binary with sub-10ms response:

| Rule | Protected | Action |
|------|-----------|--------|
| R01 | `sudo` commands | **Deny** |
| R02 | `.git/`, `.env`, secrets | **Deny** write |
| R03 | Shell writes to protected files | **Deny** |
| R04 | Writes outside project | **Ask** |
| R05 | `rm -rf` | **Ask** |
| R06 | `git push --force` | **Deny** |
| R07–R09 | Mode-specific and secret-read guards | Context-aware |
| R10 | `--no-verify`, `--no-gpg-sign` | **Deny** |
| R11 | `git reset --hard main/master` | **Deny** |
| R12 | Direct push to `main` / `master` | **Warn** |
| R13 | Protected file edits | **Warn** |
| Post | `it.skip`, assertion tampering | **Warning** |
| Perm | `git status`, `npm test` | **Auto-allow** |

Runtime differences between Claude Code hooks and Codex CLI gates are documented in [docs/hardening-parity.md](docs/hardening-parity.md).

---

## 5 Verb Skills, Zero Config

v4 unifies 42 skills into **5 verb skills**. Start with the verbs first, then add Breezing, Codex, or 2-agent flows only when you need them.

<table>
<tr>
<td align="center" width="20%"><h3>/plan</h3>Ideas → Plans.md</td>
<td align="center" width="20%"><h3>/work</h3>Parallel implementation</td>
<td align="center" width="20%"><h3>/review</h3>4-angle code review</td>
<td align="center" width="20%"><h3>/release</h3>Tag + GitHub Release</td>
<td align="center" width="20%"><h3>/setup</h3>Project init & config</td>
</tr>
</table>

<p align="center">
  <img src="assets/readme-visuals-en/skills-ecosystem.svg" alt="Skills ecosystem" width="640">
</p>

### Key Commands

| Command | What It Does | Legacy Redirect |
|---------|--------------|-----------------|
| `/harness-plan` | Ideas → `Plans.md` | `/plan-with-agent`, `/planning` |
| `/harness-work` | Parallel implementation | `/work`, `/breezing`, `/impl` |
| `/harness-work all` | Approved plan → implement → review → commit | `/work all` |
| `/harness-review` | 4-perspective code review | `/harness-review`, `/verify` |
| `/harness-release` | CHANGELOG, tag, GitHub Release | `/release-har`, `/handoff` |
| `/harness-setup` | Initialize project | `/harness-init`, `/setup` |
| `/memory` | Manage SSOT files | — |
| `harness doctor --residue` | Detect stale references to deleted code | — |

---

## Architecture

```
claude-code-harness/
├── go/                # Go-native guardrail + hookhandler engine
│   ├── cmd/harness/   #   CLI entry point (sync, doctor, validate)
│   ├── internal/      #   guardrail / hookhandler / state / lifecycle / breezing
│   └── pkg/           #   config / hookproto (public API)
├── bin/               # Compiled harness binaries (darwin-arm64/amd64, linux-amd64)
├── skills/            # 5 verb skills + extension skills (plan/execute/review/release/setup, etc.)
├── agents/            # 3 agents (worker / reviewer / scaffolder)
├── hooks/             # CC hook configuration (hooks.json)
├── scripts/           # Auxiliary shell scripts
└── templates/         # Generation templates
```

---

## Advanced Features

### Breezing (Agent Teams)

Run entire task lists with autonomous agent teams:

```bash
/harness-work breezing all                    # Plan review + parallel implementation
/harness-work breezing --no-discuss all       # Skip plan review, go straight to coding
```

<p align="center">
  <img src="assets/readme-visuals-en/breezing-agents.svg" alt="Breezing agent teams" width="640">
</p>

**Phase 0 (Planning Discussion)** runs by default — Planner analyzes task quality, Critic challenges the plan, then you approve before coding starts. 8+ tasks auto-split into manageable batches.

### Session Memory (harness-mem)

When [harness-mem](https://github.com/Chachamaru127/harness-mem) is running, Harness automatically records session events — what you worked on, what tools were used, and how the session ended. Next time you start a session, that context is available for retrieval.

- **Without harness-mem**: events are logged locally to `.claude/state/memory-bridge-events.jsonl` (no external dependency)
- **With harness-mem**: events are also sent to the memory server for cross-session search and retrieval

No configuration needed — Harness detects harness-mem automatically.

<details>
<summary><strong>Codex Engine</strong></summary>

Delegate implementation tasks to OpenAI Codex in parallel. Codex implements, self-reviews, and reports back.

```bash
/harness-work --codex implement these 5 API endpoints
/harness-review --codex  # 4 perspectives + Codex second opinion
```

> **Setup**: Install [Codex CLI](https://github.com/openai/codex) and configure API key. Or run `./scripts/setup-codex.sh --user` to use Harness skills inside Codex CLI directly.

</details>

<details>
<summary><strong>2-Agent Mode (with Cursor)</strong></summary>

Use Cursor as PM, Claude Code as implementer. Plans.md syncs between both.

```bash
/harness-release handoff  # Report to Cursor PM
```

</details>

<details>
<summary><strong>Content Generation (Slides & Video)</strong></summary>

```bash
/generate-slide   # 3 visual patterns, quality scoring, auto-export
/generate-video   # JSON Schema-driven pipeline with Remotion
```

> **Dependencies**: Slides need `GOOGLE_AI_API_KEY`. Video needs [Remotion](https://www.remotion.dev/) + ffmpeg.

</details>

---

## Why Harness vs Skill-Pack Only?

Skill packs can teach a prompt. Harness also enforces behavior at runtime.

- **Guardrail engine** blocks destructive writes, secret exposure, and force-push patterns on the actual execution path.
- **Hooks + review flow** keep quality checks close to the tools that edit your repo.
- **Validation scripts + evidence pack** give you a rerunnable way to confirm docs, packaging, and `/harness-work all` behavior.

---

## Troubleshooting

| Issue | Solution |
|-------|----------|
| Command not found | Run `/harness-setup` first |
| `harness-*` commands missing on Windows | Update or reinstall the plugin. Public command skills now ship as real directories, so `core.symlinks=false` no longer hides them. |
| Plugin not loading | Clear cache: `rm -rf ~/.claude/plugins/cache/claude-code-harness-marketplace/` and restart |
| Hooks not working | Run `bin/harness doctor` to diagnose (Go binary, no Node.js needed) |
| Stale v3 references after migration | Run `bin/harness doctor --residue` — auto-detects leftover references to deleted code |

For more help, [open an issue](https://github.com/Chachamaru127/claude-code-harness/issues).

---

## Uninstall

```bash
/plugin uninstall claude-code-harness
```

Project files (Plans.md, SSOT files) remain unchanged.

---

## Claude Code Feature Highlights

Harness leverages the latest Claude Code features automatically. Here are the ones you'll notice:

| What You Get | How It Works |
|-------------|-------------|
| **Parallel safe writes** | Worktree isolation lets multiple workers edit the same file |
| **Smart effort scaling** | Complex tasks auto-trigger ultrathink mode |
| **Auto-escalation** | 3 consecutive tool failures trigger recovery |
| **LLM quality guards** | Agent hooks run security and quality checks on every edit |
| **Team monitoring** | Breezing auto-detects when workers finish or idle |
| **Model flexibility** | Use any provider (Bedrock, Vertex, etc.) via `modelOverrides` |

Full technical list (19 features): [docs/CLAUDE-feature-table.md](docs/CLAUDE-feature-table.md)

---

## Documentation

| Resource | Description |
|----------|-------------|
| [Changelog](CHANGELOG.md) | Version history |
| [Claude Code Compatibility](docs/CLAUDE_CODE_COMPATIBILITY.md) | Requirements |
| [Distribution Scope](docs/distribution-scope.md) | Included vs compatibility vs development-only paths |
| [Work All Evidence Pack](docs/evidence/work-all.md) | Success/failure verification contract |
| [Cursor Integration](docs/CURSOR_INTEGRATION.md) | 2-Agent setup |
| [Benchmark Rubric](docs/benchmark-rubric.md) | Static vs executed evidence scoring |
| [Positioning Notes](docs/positioning-notes.md) | Public-facing differentiation language |
| [Content Layout](docs/content-layout.md) | Source docs vs generated outputs convention |

---

## Contributing

Issues and PRs welcome. See [CONTRIBUTING.md](CONTRIBUTING.md).

---

## Acknowledgments

- [AI Masao](https://note.com/masa_wunder) — Hierarchical skill design
- [Beagle](https://github.com/beagleworks) — Test tampering prevention patterns

---

## License

**MIT License** — Free to use, modify, commercialize.

[English](LICENSE.md) | [日本語](LICENSE.ja.md)
