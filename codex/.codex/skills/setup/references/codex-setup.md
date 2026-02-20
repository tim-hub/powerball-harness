# Codex CLI Setup Reference

Setup project for Codex CLI compatibility.

## Deliverables

- `.codex/skills/` - Harness skills for Codex
- `.codex/rules/` - Guardrails
- `AGENTS.md` - Project instructions (project mode only)
- Optional: `.codex/config.toml` (MCP template)
- Required defaults: `features.multi_agent=true` and `[agents.*]` role declarations
- Legacy cleanup: `_archived` and `*.backup.*` under skills are moved to backup directory

---

## Execution Flow

### Step 1: Confirmation

```text
Codex CLI 用の設定を入れますか？
- skills/rules
- (project mode only) AGENTS.md
- config.toml defaults (multi_agent + roles)
```

### Step 2: MCP Template Decision

```text
MCP テンプレート（config.toml）をコピーしますか？
- yes -> --with-mcp
- no  -> --skip-mcp
```

### Step 3: Run Setup Script

```bash
bash "${CLAUDE_PLUGIN_ROOT}/scripts/codex-setup-local.sh" --with-mcp
```

or

```bash
bash "${CLAUDE_PLUGIN_ROOT}/scripts/codex-setup-local.sh" --skip-mcp
```

### Step 4: Verify

```bash
ls -la .codex/skills
ls -la .codex/rules
[ -f .codex/config.toml ] && rg -n "multi_agent|\[agents" .codex/config.toml
```

### Step 5: Completion Message

```text
Codex CLI setup complete.
- multi_agent defaults applied
- harness role defaults applied
```

---

## Runtime Expectations

- `$work` / `$breezing` default to Codex native multi-agent execution.
- `--claude` routes both implementation and review to Claude.
- `--claude + --codex-review` is invalid.

## State Path

Use `${CODEX_HOME:-~/.codex}/state/harness/` for runtime state.

## Duplicate Listing Guard

- Setup scripts skip `_archived` and `*.backup.*` when syncing skills.
- Existing legacy skill folders are moved to `${CODEX_HOME:-~/.codex}/backups/*` (or `.codex/backups/*` in project mode).
