# Sandbox Allowlist Recipe

How to configure the Claude Code sandbox for projects that need external network access (web scraping, APIs, package registries).

> **TL;DR**: The CC sandbox defaults to **empty allowlist = deny all outbound**. Add `sandbox.network.allowedDomains` to your `~/.claude/settings.json` manually — the AI cannot write to this file (blocked by self-audit guardrail).

## Symptoms

External network calls from `WebFetch`, `Bash(curl)`, or MCP tools fail with:

```
HTTP/2 403
x-deny-reason: host_not_allowed
```

or

```
curl: (6) Could not resolve host: api.example.com
```

## Cause

The Claude Code sandbox (macOS Seatbelt / Linux bubblewrap) uses an **allowlist-default** network policy. Without explicit `allowedDomains`, all outbound traffic is blocked.

## Fix: Merge sandbox settings into `~/.claude/settings.json`

**Important**: Check whether a `sandbox` key already exists before editing — overwriting it silently drops existing guardrails like `failIfUnavailable` and `deniedDomains`.

### Step 0: Check for existing sandbox config

```bash
jq 'has("sandbox")' ~/.claude/settings.json
# false → Case A (new addition)
# true  → Case B (inner merge)
```

### Case A: No existing `sandbox` key

Add a top-level `sandbox` key alongside existing `permissions` / `hooks` / `enabledPlugins` / `mcpServers`. Do not touch existing keys:

```json
{
  "permissions": { /* preserve existing */ },
  "hooks": { /* preserve existing */ },

  "sandbox": {
    "enabled": true,
    "autoAllowBashIfSandboxed": true,
    "excludedCommands": [
      "docker", "docker-compose", "watchman",
      "systemctl", "launchctl", "brew services"
    ],
    "network": {
      "allowedDomains": [
        "github.com", "api.github.com", "raw.githubusercontent.com",
        "codeload.github.com", "objects.githubusercontent.com",
        "registry.npmjs.org", "api.anthropic.com",
        "pypi.org", "files.pythonhosted.org",
        "proxy.golang.org", "sum.golang.org",
        "crates.io", "static.crates.io", "rubygems.org"
      ],
      "deniedDomains": [
        "169.254.169.254", "metadata.google.internal", "metadata.azure.com",
        "pastebin.com", "transfer.sh", "0x0.st",
        "paste.ee", "termbin.com", "ix.io"
      ]
    }
  }
}
```

### Case B: `sandbox` key already exists

Use `jq` to merge only the `allowedDomains` array without overwriting the rest:

```bash
# Dry run — preview merged result
jq '.sandbox.network.allowedDomains += ["api.example.com"]' ~/.claude/settings.json

# Apply
tmp=$(mktemp)
jq '.sandbox.network.allowedDomains += ["api.example.com"]' ~/.claude/settings.json > "$tmp"
mv "$tmp" ~/.claude/settings.json
```

## Allowlist Domain Groups

The `harness/templates/sandbox-settings.json.template` organizes allowed domains into three tiers:

### Tier 1: Development Core (14 domains)

```
github.com, api.github.com, raw.githubusercontent.com,
codeload.github.com, objects.githubusercontent.com
registry.npmjs.org
api.anthropic.com
pypi.org, files.pythonhosted.org
proxy.golang.org, sum.golang.org
crates.io, static.crates.io
rubygems.org
```

Always include these if you do any development work.

### Tier 2: Web Scraping / Firecrawl (2 domains)

```
api.firecrawl.dev, firecrawl.dev
```

Add when using Firecrawl-based skills.

### Tier 3: Tech Blog Scraping targets (13 domains)

```
techblog.zozo.com, note.com, assets.st-note.com,
zenn.dev, qiita.com, dev.to, medium.com,
cdn-ak.f.st-hatena.com,
engineering.dena.com, developers.cyberagent.co.jp,
tech.uzabase.com, engineer.crowdworks.jp, tech.smarthr.jp
```

Add selectively based on which sites you actually need to access.

## Deny List (Always Keep)

Block SSRF targets and data-exfiltration sinks — these should stay in `deniedDomains` regardless of other settings:

```json
"deniedDomains": [
  "169.254.169.254",       // AWS metadata
  "metadata.google.internal", // GCP metadata
  "metadata.azure.com",    // Azure metadata
  "pastebin.com",          // exfil sink
  "transfer.sh",           // exfil sink
  "0x0.st",               // exfil sink
  "paste.ee",              // exfil sink
  "termbin.com",           // exfil sink
  "ix.io"                  // exfil sink
]
```

`deniedDomains` takes precedence over `allowedDomains` — a domain on both lists is denied.

## Template File

`harness/templates/sandbox-settings.json.template` contains the full recommended configuration (29 allowed + 9 denied + 6 excludedCommands). Use it as a starting point:

```bash
# Review the template
cat harness/templates/sandbox-settings.json.template

# Apply (Case A — no existing sandbox key)
jq --slurpfile tmpl harness/templates/sandbox-settings.json.template \
  '. + {sandbox: $tmpl[0].sandbox}' ~/.claude/settings.json > /tmp/merged.json
mv /tmp/merged.json ~/.claude/settings.json
```

## Verification

After updating, restart Claude Code and test with:

```bash
curl -I https://api.example.com  # should return 200/301, not 403
```

Or via Claude:

```
Use Bash to run: curl -I https://api.example.com
```

A `403` with `x-deny-reason: host_not_allowed` means the domain is still blocked. Double-check the `allowedDomains` list and ensure you saved the correct file (`~/.claude/settings.json`, not a project-level `.claude/settings.json`).

## Security Notes

- **Never allowlist all domains** — the allowlist default exists to prevent credential exfiltration and SSRF
- **Verify `deniedDomains`** persists through any merge — cloud metadata endpoints are the highest-priority entries
- **Project vs user settings**: `~/.claude/settings.json` (user-global) is preferred for sandbox config; project-level `.claude/settings.json` also works but affects all contributors

## Related

- `harness/templates/sandbox-settings.json.template` — full recommended config template
- Claude Code sandbox documentation: https://docs.anthropic.com/en/docs/claude-code/settings#sandbox
