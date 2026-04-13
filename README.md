<p align="center">
  <em>Turn Claude Code into a disciplined development partner.</em>
</p>

<p align="center">
  <a href="https://github.com/tim-hub/powerball-harness/releases/latest"><img src="https://img.shields.io/github/v/release/tim-hub/powerball-harness?display_name=tag&sort=semver" alt="Latest Release"></a>
  <a href="LICENSE.md"><img src="https://img.shields.io/badge/License-MIT-green.svg" alt="License"></a>
  <a href="docs/CLAUDE_CODE_COMPATIBILITY.md"><img src="https://img.shields.io/badge/Claude_Code-v2.1+-purple.svg" alt="Claude Code"></a>
</p>

A Claude Code plugin for autonomous **Plan → Work → Review** workflows, with a Go-native guardrail engine that protects your repo at runtime.

---

## Requirements

- **Claude Code v2.1+**
- **Go 1.22+** runtime (pre-built binary included in releases — no separate install needed)

---

## Install

```bash
# In Claude Code (user scope recommended — applies across all your projects)
/plugin marketplace add tim-hub/powerball-harness
/plugin install harness@powerball-harness-marketplace --scope user
```


---

## The 4 Verb Workflow

| Command | What it does |
|---------|-------------|
| `/harness-plan` | Ideas → `Plans.md` with acceptance criteria |
| `/harness-work` | Parallel implementation (auto-detects task count) |
| `/harness-review` | 4-perspective code review (security, perf, quality, a11y) |
| `/harness-release` | CHANGELOG, tag, and GitHub Release |

Run everything after plan approval:

```bash
/harness-work all
```

---

## Guardrails

A Go-native engine (`go/internal/guardrail/`) enforces 13 declarative rules at runtime — blocking `sudo`, force-push, secret writes, `rm -rf`, and more. See [docs/hardening-parity.md](docs/hardening-parity.md) for the full rule table.

---

## Architecture

```
claude-code-harness/
├── go/         # Go guardrail engine (bin/harness binary)
├── skills/     # 5 verb skills
├── agents/     # worker / reviewer / scaffolder
├── hooks/      # Thin shims → Go binary
└── scripts/    # Helper scripts
```

---

## Troubleshooting

| Issue | Fix |
|-------|-----|
| Hook errors on every prompt | Run `/harness-setup binary` to manually re-download the platform binary |
| Commands not found | Run `/harness-setup` first |
| Plugin not loading | `rm -rf ~/.claude/plugins/cache/claude-code-harness-marketplace/` and restart |

---

## Documentation

- [Changelog](CHANGELOG.md)
- [Claude Code Compatibility](docs/CLAUDE_CODE_COMPATIBILITY.md)
- [Guardrail Rules](docs/hardening-parity.md)
- [Work All Evidence](docs/evidence/work-all.md)

---

## Contributing

Issues and PRs welcome. See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT — [Full License](LICENSE.md)

## Origin

Forked from [claude-code-harness](https://github.com/Chachamaru127/claude-code-harness) and significantly modified.
