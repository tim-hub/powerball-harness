# SEO Expert Prompt for Codex

SEO/OGP review prompt for Codex MCP.

## 7-Section Format

### TASK

Analyze SEO optimization and OGP tags, detecting issues affecting search engine ranking and social sharing quality.

### EXPECTED OUTCOME

Report SEO issues in the following format:
- Issue list (Severity: Critical/High/Medium/Low)
- Specific fix proposals
- SEO score (A-F)

### CONTEXT

Review target:
- Changed files: {files}
- Framework: {tech_stack}
- Focus: Meta tags, OGP, robots.txt, structured data

### CONSTRAINTS

- Consider framework-specific SEO features (Next.js Metadata API, etc.)
- Check dynamic pages with representative patterns

### MUST DO

1. **Basic meta tags**:
   - title missing/duplicate/length
   - description missing/length
   - canonical URL
   - viewport

2. **OGP**:
   - og:title, og:description, og:image
   - og:image size (1200x630 recommended)
   - og:url and canonical match

3. **Twitter Card**:
   - twitter:card (summary_large_image recommended)
   - twitter:title, twitter:description, twitter:image

4. **Crawlability**:
   - robots.txt existence/config
   - sitemap.xml existence
   - Residual noindex check

5. **HTTP status**:
   - 404/500 returning 200
   - Redirect chains

### MUST NOT DO

- Do not flag API routes/backend files for SEO
- Do not flag admin/private pages for SEO
- Do not report structured data as required (treat as optional)

### OUTPUT FORMAT

```markdown
## SEO/OGP Review Results

**Score**: [A-F]

### Findings

| # | Severity | File | Issue | Fix |
|---|----------|------|-------|-----|
| 1 | High | app/layout.tsx | Missing viewport meta | Add viewport meta tag |
```
