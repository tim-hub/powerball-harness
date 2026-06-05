<p align="center">
  <em>Turn Claude Code into a disciplined development partner.</em>
</p>

<p align="center">
  <a href="https://github.com/tim-hub/powerball-harness/releases/latest"><img src="https://img.shields.io/github/v/release/tim-hub/powerball-harness?display_name=tag&sort=semver" alt="Latest Release"></a>
  <a href="LICENSE.md"><img src="https://img.shields.io/badge/License-MIT-green.svg" alt="License"></a>
  <a href="docs/CLAUDE_CODE_COMPATIBILITY.md"><img src="https://img.shields.io/badge/Claude_Code-v2.1+-purple.svg" alt="Claude Code"></a>
</p>

A Claude Code plugin for autonomous **Plan → Work → Review** workflows, backed by two Go-native guardrail engines that block dangerous operations and credential leaks at runtime.

<p align="center">
  <img src="docs/assets/readme-visuals-en/generated/why-harness-pillars.svg" alt="What changes with Claude Harness: shared plan, runtime guardrails, and rerunnable validation" width="860">
</p>

- **A shared plan drives the work** — every task in `.claude/harness/plans.json` has acceptance criteria; agents follow the plan instead of improvising
- **Two Go-native guardrails** — operation guard (`R01–R13`) blocks dangerous tool calls; content guard (PII Guard) blocks credential leaks
- **Rerunnable validation** — consistency checks, plugin validation, and migration-residue scans gate every change

---

## Installation

**Requirements**: Claude Code v2.1+ · Go 1.22+ runtime

```bash
# Run inside Claude Code (user scope recommended — applies across all your projects)
/plugin marketplace add tim-hub/powerball-harness
/plugin install harness@powerball-harness-marketplace
```

> First-time setup only: run `/harness-setup` to create `CLAUDE.md`, `.claude/harness/plans.json`, and `settings.json`. Existing projects with these files can skip it.

---

## Core Skills

| Command | What it does |
|---------|-------------|
| `/harness-setup` | Project initialization (creates `CLAUDE.md` and `.claude/harness/plans.json`) |
| `/harness-plan` | Ideas → `.claude/harness/plans.json` with acceptance criteria |
| `/harness-work` | Implementation — auto-selects solo (1 task) / parallel (2–3) / breezing (4+) |
| `/harness-review` | 4-perspective code review (security, performance, quality, a11y) |
| `/maintenance` | Housekeeping — log pruning, stale-state cleanup, orphaned worktrees |

Run everything after plan approval:

```bash
/harness-work all
```


Full skill catalog, lifecycle diagrams, and agent roles: [harness/README.md](harness/README.md).

---

## Go Guardrails

<p align="center">
  <img src="docs/assets/readme-visuals-en/generated/safety-guardrails.svg" alt="Safety Protection System" width="640">
</p>

13 declarative rules in [`go/internal/guardrail/`](go/internal/guardrail/), evaluated in priority order on every tool call:

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

Runtime hook behavior: [docs/hardening-parity.md](docs/hardening-parity.md) · Engine internals: [go/README.md](go/README.md).


### PII & Secret Guard

A second guardrail engine in [`go/internal/piiguard/`](go/internal/piiguard/) — 45 rules (15 built-in + 30 from an embedded coding-only catalog) that block AWS / OpenAI / Anthropic / Google / GitHub / Stripe / HuggingFace API keys, JWT and Bearer tokens, PEM private keys, generic `api_key = "..."` assignments, and email addresses. Wired into three hooks:

| Hook event | Action on detection |
|---|---|
| `UserPromptSubmit` | Hard block (`{decision: "block"}` + exit 1) |
| `PreToolUse` (`Write\|Edit\|MultiEdit\|Bash\|Read`) | `permissionDecision: deny` |
| `PostToolUse` (`Bash\|Read`) | Inject redacted view via `additionalContext` |

Disable globally with `HARNESS_PIIGUARD_DISABLED=1` or per-rule with `HARNESS_PIIGUARD_DISABLED_RULES=id1,id2`.


---

## Codex CLI Integration

Harness can delegate implementation and review tasks to [OpenAI Codex CLI](https://github.com/openai/codex) as a parallel execution engine, running the same Harness skills inside Codex. **Prerequisites**: Node.js 20+, `npm install -g @openai/codex`, and the [`openai/codex-plugin-cc`](https://github.com/openai/codex-plugin-cc) Claude Code plugin installed (`/plugin install openai/codex-plugin-cc`). Run `/harness-setup codex` once to copy skills and config into `.codex/` — after that. Full invocation policy: [harness/rules/codex-cli-only.md](harness/rules/codex-cli-only.md).

---

## Other Core Features


### Academic Foundations

Two papers directly shaped this project's design — see [docs/credits-and-references.md](docs/credits-and-references.md) for summaries and links.


---

## `.claude/` Folder

After `/harness-setup`, your project will have a `.claude/` folder. Some paths are tracked in git; others are generated at runtime and git-ignored. Full layout: [docs/credits-and-references.md#claude-folder-layout](docs/credits-and-references.md#-claude-folder-layout).

**Committed** (you should check these in): `agents/`, `harness/plans.json`, `memory/`, `rules/`, `skills/`, `settings.json`

**Git-ignored** (generated): `sessions/`, `logs/`, `state/`, `worktrees/`

---

## Documentation

- [Changelog](CHANGELOG.md)
- [Credits & References](docs/credits-and-references.md)
- [Claude Code Compatibility](docs/CLAUDE_CODE_COMPATIBILITY.md)
- [Guardrail Rules](docs/hardening-parity.md)
- [harness-work all examples](docs/evidence/work-all.md)
- [Harness Plugin Workflow Diagrams · Skill Catalog · Hooks](harness/README.md)

### Troubleshooting

| Issue | Fix |
|---|---|
| Hook errors on every prompt | Run `/harness-setup binary` to re-download the platform binary |
| Commands not found | Run `/harness-setup` first |
| Plugin not loading | `rm -rf ~/.claude/plugins/cache/powerball-harness-marketplace/` and restart |

### Uninstall

```bash
/plugin uninstall powerball-harness
```

Project files (`.claude/harness/plans.json`, `CLAUDE.md`) remain unchanged.

---

## Contributing

Issues and PRs welcome. See [CONTRIBUTING.md](CONTRIBUTING.md).

Credits and academic references: [docs/credits-and-references.md](docs/credits-and-references.md).

## License

MIT — [Full License](LICENSE.md)
