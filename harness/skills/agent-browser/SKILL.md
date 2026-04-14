---
name: agent-browser
description: "Use when automating browser workflows — visiting URLs, clicking, filling forms, scraping, screenshots, UI testing, or web data extraction. Do NOT load for: embedding URLs in code, reading local screenshots, or non-browser file ops."
allowed-tools: ["Bash", "Read"]
user-invocable: false
context: fork
argument-hint: "[url] [--headless]"
---

# Agent Browser Skill

A skill for browser automation. Uses the agent-browser CLI to perform UI debugging, verification, and automated operations.

---

## Trigger Phrases

This skill is automatically triggered by the following phrases:

- "open the page", "check the URL"
- "click on", "type into", "fill the form"
- "take a screenshot"
- "check the UI", "test the screen"
- "open this page", "click on", "fill the form", "screenshot"

---

## Feature Details

| Feature | Details |
|---------|--------|
| **Browser Automation** | See [references/browser-automation.md](${CLAUDE_SKILL_DIR}/references/browser-automation.md) |
| **AI Snapshot Workflow** | See [references/ai-snapshot-workflow.md](${CLAUDE_SKILL_DIR}/references/ai-snapshot-workflow.md) |

## Execution Steps

### Step 0: Verify agent-browser

```bash
# Check installation
which agent-browser

# If not installed
npm install -g agent-browser
agent-browser install
```

### Step 1: Classify the User's Request

| Request Type | Corresponding Action |
|-------------|---------------------|
| Open a URL | `agent-browser open <url>` |
| Click an element | Snapshot → `agent-browser click @ref` |
| Fill a form | Snapshot → `agent-browser fill @ref "text"` |
| Check state | `agent-browser snapshot -i -c` |
| Screenshot | `agent-browser screenshot <path>` |
| Debug | `agent-browser --headed open <url>` |

### Step 2: AI Snapshot Workflow (Recommended)

For most operations, first **take a snapshot** and then interact using element references:

```bash
# 1. Open the page
agent-browser open https://example.com

# 2. Take a snapshot (AI-friendly, interactive elements only)
agent-browser snapshot -i -c

# Example output:
# - link "Home" [ref=e1]
# - button "Login" [ref=e2]
# - input "Email" [ref=e3]
# - input "Password" [ref=e4]
# - button "Submit" [ref=e5]

# 3. Interact using element references
agent-browser click @e2           # Click the Login button
agent-browser fill @e3 "user@example.com"
agent-browser fill @e4 "password123"
agent-browser click @e5           # Submit
```

### Step 3: Verify Results

```bash
# Check current state with a snapshot
agent-browser snapshot -i -c

# Or check the URL
agent-browser get url

# Take a screenshot
agent-browser screenshot result.png
```

---

## Quick Reference

### Basic Operations

| Command | Description |
|---------|-------------|
| `open <url>` | Open a URL |
| `snapshot -i -c` | AI-friendly snapshot |
| `click @e1` | Click an element |
| `fill @e1 "text"` | Fill a form field |
| `type @e1 "text"` | Type text |
| `press Enter` | Press a key |
| `screenshot [path]` | Take a screenshot |
| `close` | Close the browser |

### Navigation

| Command | Description |
|---------|-------------|
| `back` | Go back |
| `forward` | Go forward |
| `reload` | Reload |

### Information Retrieval

| Command | Description |
|---------|-------------|
| `get text @e1` | Get text |
| `get html @e1` | Get HTML |
| `get url` | Current URL |
| `get title` | Page title |

### Waiting

| Command | Description |
|---------|-------------|
| `wait @e1` | Wait for an element |
| `wait 1000` | Wait 1 second |

### Debugging

| Command | Description |
|---------|-------------|
| `--headed` | Show the browser |
| `console` | Console logs |
| `errors` | Page errors |
| `highlight @e1` | Highlight an element |

---

## Session Management

Manage multiple tabs/sessions in parallel:

```bash
# Specify a session
agent-browser --session admin open https://admin.example.com
agent-browser --session user open https://example.com

# List sessions
agent-browser session list

# Operate within a specific session
agent-browser --session admin snapshot -i -c
```

---

## Choosing Between MCP Browser Tools

| Tool | Recommendation | Use Case |
|------|---------------|----------|
| **agent-browser** | ★★★ | First choice. Powerful AI-friendly snapshots |
| chrome-devtools MCP | ★★☆ | When Chrome is already open |
| playwright MCP | ★★☆ | Complex E2E testing |

**Principle**: Try agent-browser first; use MCP tools only if it does not work.

---

## Notes

- agent-browser defaults to headless mode
- Use the `--headed` option to display the browser
- Sessions persist until explicitly closed with `close`
- Use sessions for sites that require authentication
