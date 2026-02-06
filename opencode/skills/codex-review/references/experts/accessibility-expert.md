# Accessibility Expert Prompt for Codex

Accessibility (a11y) review prompt for Codex MCP.

## 7-Section Format

### TASK

Analyze web accessibility and verify compliance with WCAG 2.1 AA guidelines.

### EXPECTED OUTCOME

Report a11y issues in the following format:
- Issue list (Severity: Critical/High/Medium/Low)
- WCAG criterion reference
- Fix proposals
- Accessibility score (A-F)

### CONTEXT

Review target:
- Changed files: {files}
- Framework: {tech_stack}
- Focus: UI components, forms, images, navigation

### CONSTRAINTS

- WCAG 2.1 AA baseline
- Consider framework-specific patterns (React/Vue/Svelte)

### MUST DO

1. **Semantic HTML**: Heading hierarchy, landmarks, button vs div
2. **Images/media**: alt attributes, decorative images, video captions
3. **Forms**: Labels, error messages, required fields
4. **Keyboard**: Focus management, trap prevention, ESC handling
5. **ARIA**: Remove redundant ARIA, aria-live for dynamic content

### MUST NOT DO

- Do not flag non-UI files (API, utilities) for a11y
- Do not require meaningful alt for decorative images
- Do not report correct aria-hidden usage as an issue

### OUTPUT FORMAT

```markdown
## Accessibility Review Results

**Score**: [A-F]

### Findings

| # | Severity | File | Line | Issue | WCAG | Fix |
|---|----------|------|------|-------|------|-----|
| 1 | High | components/Button.tsx | 12 | div used as button | 4.1.2 | Use <button> element |
```
