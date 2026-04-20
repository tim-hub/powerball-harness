<p align="center">
  <em>Turn Claude Code into a disciplined development partner.</em>
</p>

<p align="center">
  <a href="https://github.com/tim-hub/powerball-harness/releases/latest"><img src="https://img.shields.io/github/v/release/tim-hub/powerball-harness?display_name=tag&sort=semver" alt="Latest Release"></a>
  <a href="LICENSE.md"><img src="https://img.shields.io/badge/License-MIT-green.svg" alt="License"></a>
  <a href="docs/CLAUDE_CODE_COMPATIBILITY.md"><img src="https://img.shields.io/badge/Claude_Code-v2.1+-purple.svg" alt="Claude Code"></a>
</p>

A Claude Code plugin for autonomous **Plan → Work → Review** workflows, with a Go-native guardrail engine that protects your repo at runtime.

<p align="center">
  <img src="docs/assets/readme-visuals-en/generated/why-harness-pillars.svg" alt="What changes with Claude Harness: shared plan, runtime guardrails, and rerunnable validation" width="860">
</p>

---

## Requirements

- **Claude Code v2.1+**
- **Go 1.22+** runtime

---

## Install

```bash
# In Claude Code (user scope recommended — applies across all your projects)
/plugin marketplace add tim-hub/powerball-harness
/plugin install harness@powerball-harness-marketplace --scope user
```

---

## How It Works

- Use claude memory to support decisions and patterns, and to maintain a session log of what was done and why
- Use Plans.md a both human readable and LLM friendly document spec to drive implementation.
- Use go based engine for hooks to enforce runtime guardrails and autmate tasks.
- Use `deleted-concepts.yaml` to track and prevent reintroduction of rejected ideas.



## The 5 Verb Workflow

| Command | What it does |
|---------|-------------|
| `/harness-setup` | Project initialization (eg, creates `CLAUDE.md` and `Plans.md`) |
| `/harness-plan` | Ideas → `Plans.md` with acceptance criteria |
| `/harness-work` | Parallel implementation (auto-detects task count) |
| `/harness-review` | 4-perspective code review (security, perf, quality, a11y) |
| `/harness-release` | CHANGELOG, tag, and GitHub Release |

Run everything after plan approval:

```bash
/harness-work all
```

> **Note**: `/harness-setup` is only needed when onboarding a brand new project that doesn't yet have `CLAUDE.md` or `Plans.md`. If your project already has those, or if you're just using the skills directly, you can skip it entirely.

---

## Security Guardrails

<p align="center">
  <img src="docs/assets/readme-visuals-en/generated/safety-guardrails.svg" alt="Safety Protection System" width="640">
</p>

Harness protects your codebase with a **Go-native guardrail engine** (`go/internal/guardrail/`) — 13 declarative rules (R01–R13) evaluated in priority order:

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

Runtime hook behavior is documented in [docs/hardening-parity.md](docs/hardening-parity.md).

---

## Architecture

```
powerball-harness/
├── go/         # Go guardrail engine (bin/harness binary)
├── skills/     # 27 skills (5 core verbs + specialized)
├── agents/     # 7 agents (worker, reviewer, scaffolder + 4 specialized)
├── hooks/      # ~35 hook entries across 27 event types → Go binary
└── scripts/    # Helper scripts
```

> Deep dive into the guardrail engine — rules R01–R13, pre/post-tool pipelines, fail-safe design: **[go/README.md](go/README.md)**

---

## Advisor Strategy

The **Advisor** is a read-only consultation agent (model: Opus) that Workers can consult during task execution. It never writes code or executes tools — it only returns structured guidance.

### When It Triggers

| Trigger | Condition |
|---------|-----------|
| High-risk preflight | Task marked `<!-- advisor:required -->` |
| Repeated failure | Same error signature after >= `retry_threshold` retries |
| Plateau | Task restarted without new commits |

### Decisions

- **`PLAN`** — Executor replans using the suggested approach
- **`CORRECTION`** — Executor applies the provided local fix
- **`STOP`** — Escalate to Reviewer; human decision required

> See [docs/advisor-strategy.md](docs/advisor-strategy.md) for full configuration reference and the 4-agent team model.

---

## Troubleshooting

| Issue | Fix |
|-------|-----|
| Hook errors on every prompt | Run `/harness-setup binary` to manually re-download the platform binary |
| Commands not found | Run `/harness-setup` first |
| Plugin not loading | `rm -rf ~/.claude/plugins/cache/powerball-harness-marketplace/` and restart |

---

## Uninstall

```bash
/plugin uninstall powerball-harness
```

Project files (`Plans.md`, `CLAUDE.md`, SSOT files) remain unchanged.

---

## Documentation

- [Changelog](CHANGELOG.md)
- [Claude Code Compatibility](docs/CLAUDE_CODE_COMPATIBILITY.md)
- [Guardrail Rules](docs/hardening-parity.md)
- [Work All Evidence](docs/evidence/work-all.md)
- [Advisor Strategy](docs/advisor-strategy.md)

> Want to know more about how powerball-harness works under the hood? See **[harness/README.md](harness/README.md)** — full workflow diagrams, skill catalog, agent roles, and the hook event map.

---

## Contributing

Issues and PRs welcome. See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT — [Full License](LICENSE.md)

## Origin

Forked from [claude-code-harness](https://github.com/Chachamaru127/claude-code-harness) and significantly modified.

### What stayed the same?
- Harness flow
- Multi-agent architecture
- Guardrail go engine and rules


### What changed?
- Repository structure reorganized for clarity and maintainability
- Optimise skills descriptions
- Use English for wider audience and consistency with code/comments
- No binrary files committed to repo — use setup script to either build from go source or download from GitHub Releases instead
- Fix [a couple of issues and anti-patterns](https://github.com/tim-hub/powerball-harness/blob/master/CHANGELOG.md#403---2026-04-13).
- Inspired by **[Meta-Harness: End-to-End Optimization of Model Harnesses](https://arxiv.org/abs/2603.28052)** (arXiv:2603.28052): compressed feedback loses causal signal — agents need raw execution history, not summaries. This drives the per-task `.claude/state/traces/` JSONL system and the `harness-review` eval loop.
- Inspired by **[Natural-Language Agent Harnesses](https://arxiv.org/abs/2603.25723)** (arXiv:2603.25723): named failure modes drive recovery strategies. The Failure Taxonomy (`FT-*` IDs) in `.claude/rules/failure-taxonomy.md` is a direct implementation of this element.