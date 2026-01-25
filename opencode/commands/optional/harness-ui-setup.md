---
description: "[Deprecated] Set up harness-ui dashboard - Use /harness-ui instead"
user-invocable: false
---

# /harness-ui-setup - [Deprecated]

> **This command has been deprecated and merged into `/harness-ui`.**
>
> `/harness-ui` now automatically detects whether setup is needed and enters setup mode when license key is not configured.
>
> **Use `/harness-ui` instead.**

## Migration Guide

| Old Command | New Command |
|-------------|-------------|
| `/harness-ui-setup` | `/harness-ui` (auto-setup mode) |
| `/harness-ui-setup YOUR-KEY` | `/harness-ui YOUR-KEY` |
| `/harness-ui-setup --force` | `/harness-ui --force` |

## Why Deprecated?

1. **Improved UX**: Users no longer need to remember two separate commands
2. **Auto-detection**: `/harness-ui` automatically enters setup mode when needed
3. **Unified interface**: Single command for both setup and dashboard access

## See Also

- `/harness-ui` - Unified dashboard display and setup command
