# Credits & References

## Credits

| Contributor | Contribution |
|-------------|-------------|
| [@Chachamaru127](https://github.com/Chachamaru127/claude-code-harness) | Original Claude Code Harness — the Plan → Work → Review workflow structure this project builds on |
| [@datumbrain](https://github.com/datumbrain/claude-privacy-guard) | PII pattern catalog that seeds the Go-native PII Guard (`go/internal/piiguard/`) |
| [@affaan-m](https://github.com/affaan-m/everything-claude-code) | `strategic-compact` skill that inspired `harness-compact` |

## Academic References

| Paper | Relevance |
|-------|-----------|
| [Meta-Harness: End-to-End Optimization of Model Harnesses](https://arxiv.org/abs/2603.28052) | Drives the `harness-review` eval loop — compressed feedback loses causal signal, so agents work from concrete review findings rather than summaries |
| [Natural-Language Agent Harnesses](https://arxiv.org/abs/2603.25723) | Named failure modes drive recovery strategies — the Failure Taxonomy (`FT-*` IDs) in [`harness/rules/failure-taxonomy.md`](../harness/rules/failure-taxonomy.md) is a direct implementation |

---

## `.claude/` Folder Layout

When this plugin is installed and `/harness-setup` is run, the following directories are created under your project's `.claude/` folder.

### Committed to git (tracked)

| Path | Purpose |
|------|---------|
| `.claude/agents/` | Local agent definitions (project-specific overrides, e.g. `releaser.md`) |
| `.claude/harness/plans.json` | Task plan — the SSOT for phases and tasks; written only via `harness plan-cli` |
| `.claude/memory/` | Entire memory folder: decisions, patterns, session logs, and any other notes |
| `.claude/output-styles/` | Output formatting preferences |
| `.claude/rules/` | Project-specific rules loaded by Claude on every session |
| `.claude/scripts/` | Helper scripts invoked by hooks or skills |
| `.claude/skills/` | Local (project-owned) skills, e.g. `release-this` |
| `.claude/settings.json` | Claude Code permission and hook configuration |

### Git-ignored (generated / volatile)

These are created at runtime and should not be committed.

| Path | Purpose |
|------|---------|
| `.claude/sessions/` | Claude Code session transcripts |
| `.claude/logs/` | Hook and script logs |
| `.claude/state/` | Runtime state: review results, lock files |
| `.claude/worktrees/` | Temporary git worktrees used by breezing workers |

> The `.gitignore` shipped with this plugin already contains the ignore rules for the volatile paths above.
> The tracked paths are force-included with `!.claude/harness/`, `!.claude/memory/`, etc. to survive
> any global `.gitignore` rules that might otherwise exclude `.claude/`.
