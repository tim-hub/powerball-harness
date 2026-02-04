# Codex CLI Setup Reference

Setup project for Codex CLI compatibility.

## Quick Reference

- "**Codex CLI でも使いたい**" → Codex setup
- "**Codex の設定を入れて**" → Codex setup
- "**.codex を入れて**" → Codex setup

## Deliverables

- `.codex/skills/` - Harness skills for Codex
- `.codex/rules/` - Temporary guardrails
- `AGENTS.md` - Codex rules file
- Optional: `.codex/config.toml` (MCP template)

---

## Execution Flow

### Step 1: Confirmation

> Codex CLI 用の設定を入れますか？
>
> - `.codex/skills/`
> - `.codex/rules/`
> - `AGENTS.md`
> - (optional) `.codex/config.toml`
>
> 続行しますか？ (y/n)

**Wait for response**

### Step 2: MCP Template Decision

> MCP テンプレート（`.codex/config.toml`）もコピーしますか？
> - yes → `--with-mcp`
> - no  → `--skip-mcp`

**Wait for response**

### Step 3: Run Setup Script

```bash
bash "${CLAUDE_PLUGIN_ROOT}/scripts/codex-setup-local.sh" --with-mcp
```

or:

```bash
bash "${CLAUDE_PLUGIN_ROOT}/scripts/codex-setup-local.sh" --skip-mcp
```

> Note: If `CLAUDE_PLUGIN_ROOT` is unavailable, run from plugin repo root:
>
> ```bash
> bash ./scripts/codex-setup-local.sh --with-mcp
> ```
>
> or:
>
> ```bash
> bash ./scripts/codex-setup-local.sh --skip-mcp
> ```

### Step 4: Verify Copy

```bash
ls -la .codex/skills
ls -la .codex/rules
ls -la AGENTS.md
```

### Step 5: Completion Message

> Codex CLI setup complete!
>
> **Generated/updated:**
> - `.codex/skills/`
> - `.codex/rules/`
> - `AGENTS.md`
> - (optional) `.codex/config.toml`
>
> **Usage:**
> - Start Codex in the project
> - Use `$plan-with-agent`, `$work`, `$harness-review`

---

## Notes

- If `.codex/skills` or `.codex/rules` exists, the script creates a timestamp backup
- `AGENTS.md` is backed up before overwrite
- MCP template is optional and not overwritten if already present
