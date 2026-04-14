# Contributing to powerball-harness

Thank you for your interest in contributing to **powerball-harness**! This document provides guidelines for contributing.

## How to Contribute

### Reporting Issues

1. Check if the issue already exists in [GitHub Issues](https://github.com/tim-hub/powerball-harness/issues)
2. If not, create a new issue with:
   - Clear title describing the problem
   - Steps to reproduce
   - Expected vs actual behavior
   - Claude Code version and OS

### Submitting Changes

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/your-feature-name`
3. Make your changes
4. Test your changes locally
5. Commit with clear messages: `git commit -m "feat: add new feature"`
6. Push to your fork: `git push origin feature/your-feature-name`
7. Open a Pull Request

### Commit Message Format

We follow [Conventional Commits](https://www.conventionalcommits.org/). **Every commit message must start with one of these prefixes** — no exceptions, including for small or simplification commits:

- `feat:` - New features
- `fix:` - Bug fixes
- `docs:` - Documentation changes
- `refactor:` - Code refactoring (no behavior change)
- `test:` - Adding or updating tests
- `chore:` - Maintenance tasks (build config, tooling, dependency bumps)
- `perf:` - Performance improvements
- `ci:` - CI/CD pipeline changes

Optional scope in parentheses: `refactor(skills): split codex sections out of harness-review`.

Past commits without prefixes are not rewritten — the rule applies going forward.

### Code Style

- Use clear, descriptive names
- Add comments for complex logic
- Keep commands/agents/skills focused on single responsibilities

## Plugin Structure

```
powerball-harness/
├── .claude-plugin/
│   └── marketplace.json # Marketplace config (source: ./harness/)
├── harness/             # Harness plugin root
│   ├── skills/          # Skills (primary - v2.17.0+)
│   ├── agents/          # Subagents
│   ├── hooks/           # Lifecycle hooks
│   ├── templates/       # Template files
│   ├── scripts/         # Shell scripts
│   ├── output-styles/   # Output style definitions
│   ├── VERSION          # Plugin version
│   └── harness.toml     # Plugin TOML config
├── docs/                # Documentation
├── go/                  # Go guardrail engine source
├── tests/               # Validation scripts
├── README.md
├── LICENSE
└── CONTRIBUTING.md
```

> **Note**: As of v2.17.0, commands have been migrated to skills. Skills are the recommended approach for new functionality. As of v4.4.0, all plugin-specific files live under `harness/` to support multi-plugin marketplaces.

### Adding a New Skill (Recommended)

1. Create `harness/skills/your-skill/SKILL.md` with YAML frontmatter:
   ```yaml
   ---
   name: your-skill
   description: "Description with trigger phrases. Use when... Do NOT load for..."
   allowed-tools: ["Read", "Write", "Edit", "Bash"]
   ---
   ```
2. Add supporting files to `harness/skills/your-skill/references/` if needed
3. Update README.md with the new skill

### Adding a New Agent

1. Create `harness/agents/your-agent.md`
2. Define the agent with YAML frontmatter
3. Update README.md (recommended)

> Note: agents/ are auto-discovered by Claude Code. You typically do not need to manually enumerate them in `marketplace.json`.

## Version Management

Version is defined in two places that must stay in sync:
- `harness/VERSION` - Source of truth
- `.claude-plugin/marketplace.json` - Used by plugin system

Normal feature/docs PRs should leave both files unchanged and record user-facing changes in `CHANGELOG.md` under `[Unreleased]`.
Use a version bump only when you are intentionally cutting a release.

### Version Scripts

```bash
# Check if versions are in sync
./harness/scripts/sync-version.sh check

# Sync marketplace.json to VERSION
./harness/scripts/sync-version.sh sync

# Bump patch version for a release (e.g., 2.0.0 → 2.0.1)
./harness/scripts/sync-version.sh bump
```

### Release-only Versioning Policy

- Normal PRs: do not edit `VERSION` or `.claude-plugin/marketplace.json`; add notes under `[Unreleased]`
- Release work: run `./harness/scripts/sync-version.sh bump`, add a versioned `CHANGELOG.md` entry, then create the tag / GitHub Release
- The repo pre-commit hook only syncs `marketplace.json` to `VERSION` when you intentionally edit release metadata; it does not auto-bump patch versions

### Version Consistency Checks

- **Local (recommended)**: run `./harness/scripts/sync-version.sh check` before committing
- **CI (recommended)**: run `./tests/validate-plugin.sh` and `./.claude/scripts/check-consistency.sh` on PRs

## CHANGELOG Rules (Required)

**Follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/) format**

Use the following sections for each version entry:

```markdown
## [X.Y.Z] - YYYY-MM-DD

### Added
- New features

### Changed
- Changes to existing functionality

### Deprecated
- Features soon to be removed

### Removed
- Removed features

### Fixed
- Bug fixes

### Security
- Security-related fixes

#### Before/After (for major changes only)

| Before | After |
|--------|-------|
| Previous state | New state |
```

**Section guidelines**:

| Section | When to use |
|---------|-------------|
| Added | When adding completely new functionality |
| Changed | When changing behavior or UX of existing features |
| Deprecated | When announcing features slated for future removal |
| Removed | When removing features or commands |
| Fixed | When fixing bugs or issues |
| Security | When making security-related fixes |

**Before/After table**: Only add when there is a significant experience change (command deprecation/consolidation, workflow changes, breaking changes). Omit for minor fixes.

**Version comparison links**: Add to the end of CHANGELOG.md in the format `[X.Y.Z]: https://github.com/.../compare/vPREV...vX.Y.Z`

---

## Testing

Before submitting:

1. Validate plugin structure and consistency:

   ```bash
   ./tests/validate-plugin.sh
   ./.claude/scripts/check-consistency.sh
   ```

2. (Recommended) Enable pre-commit hooks (keep release metadata in sync without auto-bumping):

   ```bash
   ./harness/scripts/install-git-hooks.sh
   ```

   **Windows users**: Git hooks require [Git for Windows](https://gitforwindows.org/) which includes Git Bash. The hooks run automatically via Git Bash regardless of your shell (PowerShell, CMD, etc.).

3. Test locally in a separate project using `--plugin-dir`:

   ```bash
   cd /path/to/your-project
   claude --plugin-dir /path/to/powerball-harness
   ```

3. Verify commands work as expected (`/help`), and the core loop runs:

   - `/harness-init`
   - `/plan-with-agent`
   - `/work`
   - `/harness-review`

## Questions?

- Open an issue for questions
- Check existing documentation

Thank you for contributing!
