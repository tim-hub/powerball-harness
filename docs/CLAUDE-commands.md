# Key Command Reference

Commands and handoffs used during Claude harness development.

## Key Commands (for development)

| Command | Purpose |
|---------|---------|
| `/plan-with-agent` | Add improvement tasks to Plans.md |
| `/work` | Implement tasks (auto-scope detection, --codex support) |
| `/breezing` | Full team parallel run with Agent Teams (--codex support) |
| `/reload-plugins` | Instant refresh after editing skills/hooks (no restart needed) |
| `/harness-review` | Review changes |
| `/validate` | Validate plugin |
| `/remember` | Record learnings |

## Handoffs

| Command | Purpose |
|---------|---------|
| `/handoff-to-cursor` | Completion report for Cursor operations |

**Skills (auto-triggered in conversation)**:
- `handoff-to-impl` - "Hand off to implementer" -> PM to Impl delegation
- `handoff-to-pm` - "Report to PM" -> Impl to PM completion report

## Related Documentation

- [CLAUDE.md](../CLAUDE.md) - Project development guide
- [docs/CLAUDE-skill-catalog.md](./CLAUDE-skill-catalog.md) - Skill catalog
- [docs/CLAUDE-feature-table.md](./CLAUDE-feature-table.md) - Feature utilization table
