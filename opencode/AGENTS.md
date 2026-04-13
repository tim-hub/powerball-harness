<!-- Generated from CLAUDE.md by build-opencode.js -->
<!-- opencode.ai compatible version of Claude Code Harness -->

## General Rules
- This project use English as the main language for all code, comments, documentation, and communication. 
- When merge or rebase from upstream, for all skills, agents, etc this kind of markdown files, keep using the local version. Do not merge Japanese translation over English translation.

## Development Rules

- **Commits**: Follow [Conventional Commits](https://www.conventionalcommits.org/) (`feat:` / `fix:` / `docs:` / `refactor:` / `test:` / `chore:`)
- **Versioning**: Keep `VERSION` and `.claude-plugin/marketplace.json` in sync. Leave both unchanged in normal PRs; use `./scripts/sync-version.sh bump` only when cutting a release
- **CHANGELOG**: Record changes under `[Unreleased]` in Before/After format. Details: [.claude/rules/changelog.md](.claude/rules/changelog.md)
- **Code style**: Clear names, comments for complex logic, single-responsibility skills/agents
- **Test tampering**: Absolutely prohibited. Details: [.claude/rules/test-quality.md](.claude/rules/test-quality.md)
- [Repository structure](docs/repository-structure.md)
- Contributing guidelines: [CONTRIBUTING.md](CONTRIBUTING.md)


## Development Flow

0. **When editing skills/hooks**: run `/reload-plugins` to refresh runtime cache
1. **Plan**: `/harness-plan` to add tasks to Plans.md
2. **Implement**: `/harness-work` (single task or parallel workers)
3. **Review**: `/harness-review` (runs automatically after work; manual trigger available)
4. **Validate**: `./tests/validate-plugin.sh` and `./scripts/ci/check-consistency.sh`

## Skills

Before starting work, check if a relevant skill exists and launch it with the Skill tool.

| Skill | Purpose |
|-------|---------|
| `harness-plan` | Ideas → Plans.md |
| `harness-work` | Task implementation with parallel workers |
| `breezing` | Full auto-run with Agent Teams |
| `harness-review` | Multi-angle code review |
| `harness-release` | CHANGELOG, tag, GitHub Release |
| `harness-setup` | Project initialization |
| `memory` | SSOT management (decisions.md, patterns.md) |
| `update-changelog` | Generate CHANGELOG entries after version bump |

Full catalog: [docs/CLAUDE-skill-catalog.md](docs/CLAUDE-skill-catalog.md)

## SSOT (Single Source of Truth)

- `.claude/memory/decisions.md` - Decisions (Why)
- `.claude/memory/patterns.md` - Reusable patterns (How)
