# Hardening Parity

Last updated: 2026-03-25

This document organizes the shared policy for how far Harness provides the same level of safety across both **Claude Code** and **Codex CLI**.

The two key points are:

- What is standardized is the **policy** of "what is considered dangerous"
- Implementation is separated to match platform differences

Claude Code can stop execution immediately before and after via hooks.
Codex CLI lacks the same hooks, so it achieves similar enforcement through pre-execution instructions injection, post-execution quality gates, and pre-merge verification.

## Policy Matrix

| Policy | Examples | Severity | Claude Code | Codex CLI |
|--------|----|----------|-------------|-----------|
| No verification bypass | `git commit --no-verify`, `git commit --no-gpg-sign` | Deny | PreToolUse deny | Prohibited in instructions + quality gate fail |
| Protected branch destructive reset | `git reset --hard origin/main`, `git reset --hard main` | Deny | PreToolUse deny | Prohibited in instructions + quality gate fail |
| Direct push to protected branch | `git push origin main` | Warn | PreToolUse approve + warning | Prohibited in instructions, merge gate required |
| Force push | `git push --force`, `git push -f` | Deny | PreToolUse deny | Prohibited in instructions, merge gate required |
| Protected files editing | `package.json`, `Dockerfile`, `.github/workflows/*`, `schema.prisma`, etc. | Warn | PreToolUse approve + warning | quality gate fail (stricter than Claude) |
| Pre-push secrets scan | hardcoded secret, DB URL, private IP, token-like string | Deny | deny or fail before push-equivalent Bash | quality gate fail |

## Protected Files Profile

Default protected files are limited to those that "have wide impact when broken but are not modified in every implementation task."

- `package.json`
- `Dockerfile`
- `docker-compose.yml`
- `.github/workflows/*.yml`
- `.github/workflows/*.yaml`
- `schema.prisma`
- `wrangler.toml`
- `index.html`

Design intent:

- **Default to warn, not deny**
  Legitimate changes exist, so intent confirmation is prioritized first
- **Clearly sensitive/dangerous files like `.env` or private keys are denied by separate rules**
  This is the responsibility of existing protected path rules, not protected files
- **Codex CLI merge gate currently treats protected files as fail**
  Codex side cannot confirm interactively before execution, so protected files are stopped more aggressively via post-inspection

## Runtime Mapping

### Claude Code

Claude Code prioritizes runtime enforcement.

- **PreToolUse**
  Deny / ask / warn dangerous commands before execution
- **PostToolUse**
  Warn about tampering and security patterns after writes
- **PermissionRequest**
  Auto-approve only safe read-only / test commands

### Codex CLI

Codex CLI lacks runtime hooks, so it uses a 3-layer approximate enforcement:

1. **Pre-execution contract injection**
   Explicitly state prohibitions in instructions passed to `codex exec`, and save the same contract to state artifacts
2. **Post-exec quality gate**
   Inspect Worker output by diff / file / content basis
3. **Merge gate**
   Outputs that do not pass the quality gate are not integrated into main

## Known Asymmetry

This is important. The two are not fully equivalent.

| Item | Claude Code | Codex CLI |
|------|-------------|-----------|
| Pre-execution interruption | Possible | Not directly possible |
| Post-execution warning | Possible | Approximated via quality gate |
| Per-command deny | Strong | Instructions-dependent + post-check |
| Pre-merge blocking | Possible | Possible |
| Protected files | Warn-centric | Fail-centric |
| Direct push / force push | Detectable at runtime | Not detectable at runtime, handled via merge gate |

In summary:

- **Claude Code excels at stopping operations in real-time**
- **Codex CLI protects by not letting outputs through**

## Operator Guidance

- For safety-critical work, prefer the Claude Code path
- Use Codex CLI as an implementation/review assistant, always passing through quality gate before main integration
- For work touching protected files or release-related operations, assume Codex side will fail rather than warn

## Validation Surface

At minimum, aim for the following 4 points to be verifiable via `validate-plugin`:

- Shared policy document exists
- Claude Code guardrail has the target rules
- Codex wrapper injects the hardening contract
- Codex quality gate has parity checks
