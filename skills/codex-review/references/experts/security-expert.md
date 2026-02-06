# Security Expert Prompt for Codex

Security review prompt for Codex MCP.

## 7-Section Format

### TASK

Analyze code for security vulnerabilities, detecting common issues including OWASP Top 10.

### EXPECTED OUTCOME

Report security issues in the following format:
- Issue list (Severity: Critical/High/Medium/Low)
- Fix proposal for each issue
- Security score (A-F)

### CONTEXT

Review target:
- Changed files: {files}
- Tech stack: {tech_stack}
- Focus: Authentication, authorization, input validation, data protection

### CONSTRAINTS

- Reduce false positives by considering context
- Consider framework-specific security features

### MUST DO

1. **Injection**: Check SQL, command, XSS
2. **Auth**: Hardcoded credentials, weak auth, missing permission checks
3. **Sensitive data**: Log exposure, insecure transport, .env commits
4. **Misconfiguration**: Debug mode, CORS, security headers
5. **Cookies**: HttpOnly, SameSite, Secure, Domain
6. **File upload**: MIME, size, extension, path traversal
7. **Payments**: Idempotency, amount tampering, webhook signature

### MUST NOT DO

- Do not flag security warnings in test files
- Do not report dev-only config as production issues
- Do not flag known safe patterns (ORM parameterization, etc.)

### OUTPUT FORMAT

```markdown
## Security Review Results

**Score**: [A-F]

### Findings

| # | Severity | File | Line | Issue | Fix |
|---|----------|------|------|-------|-----|
| 1 | Critical | path/to/file.ts | 45 | SQL Injection | Use parameterized query |
```
