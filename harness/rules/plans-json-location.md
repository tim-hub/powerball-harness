# Plans JSON Location Rule

## SSOT Path

The canonical plans file is always:

```
<repo-root>/.claude/harness/plans.json
```

**Never** read from or write to any plugin cache path, such as:

```
~/.claude/plugins/cache/powerball-harness-marketplace/.../.claude/harness/plans.json
```

Plugin caches are frozen snapshots published with each release. They do not reflect the current project state and will be silently overwritten on the next plugin update.

## Why This Matters

During sessions where the plugin cache is also in scope, both paths exist on disk with similar names. The wrong path reads stale data and — because the project hook only guards the repo path — writes to the cache path will silently succeed without persisting to the live plan.

## Rule

| Context | Correct path |
|---------|-------------|
| Reading plans | `harness plan-cli list` or `harness plan-cli get <id>` (resolves repo path automatically) |
| Writing plans | `harness plan-cli add-phase`, `add-task`, `update`, `archive`, `delete-task`, `delete-phase`, `comment` |
| Direct file read (debugging only) | `<repo-root>/.claude/harness/plans.json` |
| Never | `~/.claude/plugins/cache/powerball-harness-marketplace/**/.claude/harness/plans.json` |

Always prefer `harness plan-cli` subcommands over reading the file directly so path resolution is automatic.
