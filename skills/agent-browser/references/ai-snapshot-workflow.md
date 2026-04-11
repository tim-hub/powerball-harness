# AI Snapshot Workflow

AI agent workflow utilizing the agent-browser `snapshot` command.

---

## Overview

The `snapshot` command retrieves the page's accessibility tree and assigns reference IDs (`@e1`, `@e2`, etc.) to each element. This enables:

1. **No CSS selectors needed**: No dependency on dynamic IDs or class names
2. **Context awareness**: Element roles (button, input, link) are clearly identified
3. **Deterministic operations**: Reliable interaction using references like `@e1`

---

## Basic Workflow

### Step 1: Open a Page

```bash
agent-browser open https://example.com
```

### Step 2: Take a Snapshot

```bash
agent-browser snapshot -i -c
```

**Option descriptions**:
- `-i, --interactive`: Show only interactive elements (buttons, links, input fields, etc.)
- `-c, --compact`: Remove empty structural elements for compact output

**Example output**:
```
✓ Example Domain
  https://example.com/

- link "Home" [ref=e1]
- link "About" [ref=e2]
- button "Login" [ref=e3]
- input "Search" [ref=e4]
- button "Search" [ref=e5]
```

### Step 3: Interact Using Element References

```bash
# Click a link
agent-browser click @e1

# Fill the search form
agent-browser fill @e4 "search query"

# Click the search button
agent-browser click @e5
```

### Step 4: Verify Results

```bash
# Take a new snapshot of the updated state
agent-browser snapshot -i -c
```

---

## Snapshot Option Details

### `-i, --interactive`

Show only interactive elements. Useful for narrowing down actionable targets.

```bash
# Interactive elements only
agent-browser snapshot -i

# All elements (including text nodes)
agent-browser snapshot
```

### `-c, --compact`

Remove empty structural elements (div, span, etc. with no content).

```bash
# Compact output
agent-browser snapshot -c

# Show with structure included
agent-browser snapshot
```

### `-d, --depth <n>`

Limit tree depth. Useful for getting an overview of large pages.

```bash
# Depth limit of 3
agent-browser snapshot -d 3
```

### `-s, --selector <sel>`

Scope to a specific selector.

```bash
# Only within the form
agent-browser snapshot -s "form.login"

# Only within the navigation
agent-browser snapshot -s "nav"
```

### Combinations

```bash
# Recommended: interactive + compact
agent-browser snapshot -i -c

# Only interactive elements within a form
agent-browser snapshot -i -c -s "form"

# Shallow tree for overview
agent-browser snapshot -i -d 2
```

---

## Use Case Workflows

### Login Flow

```bash
# 1. Open login page
agent-browser open https://example.com/login

# 2. Take snapshot
agent-browser snapshot -i -c
# Output:
# - input "Email" [ref=e1]
# - input "Password" [ref=e2]
# - button "Login" [ref=e3]
# - link "Forgot password?" [ref=e4]

# 3. Enter login credentials
agent-browser fill @e1 "user@example.com"
agent-browser fill @e2 "password123"

# 4. Click login button
agent-browser click @e3

# 5. Verify results
agent-browser snapshot -i -c
agent-browser get url
```

### Form Submission

```bash
# 1. Open form page
agent-browser open https://example.com/contact

# 2. Snapshot within the form
agent-browser snapshot -i -c -s "form"
# Output:
# - input "Name" [ref=e1]
# - input "Email" [ref=e2]
# - textarea "Message" [ref=e3]
# - button "Send" [ref=e4]

# 3. Fill the form
agent-browser fill @e1 "John Doe"
agent-browser fill @e2 "john@example.com"
agent-browser fill @e3 "Hello, this is a test message."

# 4. Submit
agent-browser click @e4

# 5. Verify
agent-browser snapshot -i -c
```

### Navigation Exploration

```bash
# 1. Open top page
agent-browser open https://example.com

# 2. Check navigation
agent-browser snapshot -i -c -s "nav"
# Output:
# - link "Home" [ref=e1]
# - link "Products" [ref=e2]
# - link "About" [ref=e3]
# - link "Contact" [ref=e4]

# 3. Go to Products page
agent-browser click @e2

# 4. Check the new page structure
agent-browser snapshot -i -c
```

### Dynamic Content Interaction

```bash
# 1. Open page
agent-browser open https://example.com/dashboard

# 2. Initial snapshot
agent-browser snapshot -i -c

# 3. Open dropdown
agent-browser click @e5

# 4. Wait (for dynamic content to load)
agent-browser wait 500

# 5. New snapshot (dropdown menu is now visible)
agent-browser snapshot -i -c
# New elements appear:
# - menuitem "Option 1" [ref=e10]
# - menuitem "Option 2" [ref=e11]
# - menuitem "Option 3" [ref=e12]

# 6. Select an option
agent-browser click @e11
```

---

## Troubleshooting

### Element Not Found

```bash
# Full snapshot (all elements)
agent-browser snapshot

# Narrow down with specific selector
agent-browser snapshot -s "#target-element"

# Wait and retry
agent-browser wait 2000
agent-browser snapshot -i -c
```

### Dynamic Pages

```bash
# Take snapshot after JavaScript execution
agent-browser eval "document.querySelector('#load-more').click()"
agent-browser wait 1000
agent-browser snapshot -i -c
```

### Elements Inside iframes

```bash
# Main frame snapshot
agent-browser snapshot -i -c

# Elements inside iframes cannot be directly accessed,
# so use eval to interact with iframe content
agent-browser eval "document.querySelector('iframe').contentDocument.querySelector('button').click()"
```

---

## Best Practices

### 1. Always Start with a Snapshot

Always take a snapshot before performing operations to understand the current state.

### 2. Use Interactive + Compact as Default

```bash
agent-browser snapshot -i -c
```

### 3. Verify State After Operations

```bash
agent-browser click @e1
agent-browser snapshot -i -c  # Verify results
```

### 4. Add Appropriate Waits

When dealing with dynamic content, add waits:

```bash
agent-browser click @e1
agent-browser wait 500
agent-browser snapshot -i -c
```

### 5. Leverage Sessions

Use sessions to maintain authentication state:

```bash
agent-browser --session myapp open https://example.com/login
# ... login operations ...
# Continue operations in the same session
agent-browser --session myapp open https://example.com/dashboard
```
