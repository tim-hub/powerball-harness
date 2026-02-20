# Codex CLI Setup Reference

Setup for Codex CLI compatibility (user-based by default).

## Quick Reference

- "**Codex CLI でも使いたい**" → Codex setup
- "**Codex の設定を入れて**" → Codex setup
- "**.codex を入れて**" → Codex setup

## Deliverables (default: user mode)

- `${CODEX_HOME:-~/.codex}/skills/` - Harness skills for Codex
- `${CODEX_HOME:-~/.codex}/rules/` - Guardrails
- Optional: `${CODEX_HOME:-~/.codex}/config.toml` (MCP template)
- Project mode only: `AGENTS.md`

---

## Execution Flow

### Step 1: Confirmation

> Codex CLI 用の設定を**ユーザーベース**で入れますか？
>
> - `${CODEX_HOME:-~/.codex}/skills/`
> - `${CODEX_HOME:-~/.codex}/rules/`
> - (optional) `${CODEX_HOME:-~/.codex}/config.toml`
>
> 続行しますか？ (y/n)

**Wait for response**

### Step 2: MCP Template Decision

> MCP テンプレート（`config.toml`）もコピーしますか？
> - yes → `--with-mcp`
> - no  → `--skip-mcp`

**Wait for response**

### Step 3: Run Setup Script (user mode)

```bash
bash "${CLAUDE_PLUGIN_ROOT}/scripts/codex-setup-local.sh" --user --with-mcp
```

or:

```bash
bash "${CLAUDE_PLUGIN_ROOT}/scripts/codex-setup-local.sh" --user --skip-mcp
```

> Note: If `CLAUDE_PLUGIN_ROOT` is unavailable, run from plugin repo root:
>
> ```bash
> bash ./scripts/codex-setup-local.sh --user --with-mcp
> ```
>
> or:
>
> ```bash
> bash ./scripts/codex-setup-local.sh --user --skip-mcp
> ```

### Step 4: Verify Install

```bash
CODEX_HOME="${CODEX_HOME:-$HOME/.codex}"
ls -la "$CODEX_HOME/skills"
ls -la "$CODEX_HOME/rules"
[ -f "$CODEX_HOME/config.toml" ] && echo "config.toml: found"
```

### Step 5: Completion Message

> Codex CLI setup complete! (user mode)
>
> **Generated/updated:**
> - `${CODEX_HOME:-~/.codex}/skills/`
> - `${CODEX_HOME:-~/.codex}/rules/`
> - (optional) `${CODEX_HOME:-~/.codex}/config.toml`
>
> **Usage:**
> - Restart Codex
> - Use `$plan-with-agent`, `$work`, `$harness-review`

---

## Optional: Project-local mode

If you explicitly need project-local install:

```bash
bash "${CLAUDE_PLUGIN_ROOT}/scripts/codex-setup-local.sh" --project --skip-mcp
```

Project mode adds/updates:
- `.codex/skills/`
- `.codex/rules/`
- `AGENTS.md`

## Notes

- Existing target items are backed up with timestamp suffixes
- User mode keeps project `AGENTS.md` unchanged
- MCP template is optional and not overwritten if already present

## Related: Codex MCP Review Integration

If you want to register Codex as Claude Code MCP server for second-opinion reviews, see codex-review:

- [codex-mcp-setup.md](../../codex-review/references/codex-mcp-setup.md) - Codex MCP registration
- [codex-review-integration.md](../../codex-review/references/codex-review-integration.md) - Review integration
