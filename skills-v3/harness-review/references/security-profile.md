# Security Reviewer Profile

Security-dedicated review profile activated by `harness-review --security`.
Comprehensively checks authentication, authorization, secrets, and dependency package vulnerabilities based on OWASP Top 10.

> **Read-only constraint**: The reviewer operating under this profile
> uses only Read / Grep / Glob / Bash (read-only commands).
> No Write / Edit / write Bash commands are executed.

---

## Security Review Flow

### Step 1: Identify Target Scope

```bash
# Collect changed files (BASE_REF inherited from caller)
CHANGED_FILES="$(git diff --name-only --diff-filter=ACMR "${BASE_REF:-HEAD~1}")"
git diff "${BASE_REF:-HEAD~1}" -- ${CHANGED_FILES}
```

### Step 2: OWASP Top 10 Check

Check each item below against the **change diff** and **related files**.

#### A01: Broken Access Control

| Check Item | Verification Method |
|------------|---------|
| Missing authorization checks | Are auth middlewares applied to route/endpoint definitions? |
| Horizontal privilege escalation | Is resource retrieval filtered by `userId` or similar? |
| Vertical privilege escalation | Are role checks (admin/user/guest etc.) properly implemented? |
| IDOR | Are IDs from URL parameters or request bodies accepted without authorization? |
| Directory traversal | Are path operations containing `../` sanitized? |

**Detection patterns (verify with Grep)**:
```bash
# Routes potentially missing auth
grep -rn "app\.\(get\|post\|put\|delete\|patch\)" --include="*.ts" --include="*.js"
# DB retrieval without userId
grep -rn "findById\|findOne\|select.*where" --include="*.ts"
```

#### A02: Cryptographic Failures

| Check Item | Verification Method |
|------------|---------|
| Plaintext sensitive data storage | Are passwords, tokens, PII stored in plaintext in DB/logs? |
| Weak hash algorithms | Is MD5 / SHA1 used for password hashing? |
| Insecure random numbers | Is `Math.random()` used for auth token generation? |
| TLS strength | Is sensitive data sent/received over HTTP (non-HTTPS)? |
| Hardcoded keys | Are encryption keys/IVs embedded as constants? |

**Detection patterns**:
```bash
grep -rn "md5\|sha1\|Math\.random\(\)" --include="*.ts" --include="*.js"
grep -rn "createHash.*md5\|createHash.*sha1" --include="*.ts"
grep -rn "http://" --include="*.ts" --include="*.js" --include="*.env*"
```

#### A03: Injection

| Check Item | Verification Method |
|------------|---------|
| SQL injection | Is user input concatenated into SQL strings? |
| NoSQL injection | In MongoDB etc., is `$where` or input values used as operators? |
| Command injection | Is user input passed to `exec()` / `spawn()`? |
| LDAP injection | Is unsanitized input used in LDAP queries? |
| Template injection | Is user input passed directly to template engines? |

**Detection patterns**:
```bash
grep -rn "exec\|execSync\|spawn" --include="*.ts" --include="*.js"
grep -rn "\`SELECT\|\"SELECT\|'SELECT" --include="*.ts" --include="*.js"
grep -rn "\$where\|\$\[" --include="*.ts" --include="*.js"
```

#### A04: Insecure Design

| Check Item | Verification Method |
|------------|---------|
| Missing rate limiting | Are rate limits implemented on authentication endpoints? |
| TOCTOU race conditions | Can state changes between check and use be exploited? |
| Business logic flaws | Can state transitions be executed in invalid order? |

#### A05: Security Misconfiguration

| Check Item | Verification Method |
|------------|---------|
| Default credentials | Are default passwords/usernames still in use? |
| Verbose error messages | Are stack traces or internal info returned to clients in production? |
| Unnecessary features enabled | Are debug endpoints or admin panels enabled in production? |
| HTTP security headers | Are HSTS, CSP, X-Frame-Options etc. configured? |
| CORS configuration | Is `Access-Control-Allow-Origin: *` set in production? |

**Detection patterns**:
```bash
grep -rn "cors.*origin.*\*\|allowedOrigins.*\*" --include="*.ts" --include="*.js"
grep -rn "debug.*true\|NODE_ENV.*development" --include="*.ts"
grep -rn "console\.log.*password\|console\.log.*token\|console\.log.*secret" --include="*.ts"
```

#### A06: Vulnerable and Outdated Components

| Check Item | Verification Method |
|------------|---------|
| Packages with known vulnerabilities | Are there `package.json` dependencies with reported CVEs? |
| `npm audit` results | Are high / critical vulnerabilities left unaddressed? |
| Lock file consistency | Is `package-lock.json` / `yarn.lock` up to date? |

**Verification commands**:
```bash
# Check package.json dependencies (read-only)
cat package.json | grep -E '"dependencies"|"devDependencies"' -A 50 | head -60
# Check lock file existence
ls -la package-lock.json yarn.lock pnpm-lock.yaml 2>/dev/null
```

#### A07: Identification and Authentication Failures

| Check Item | Verification Method |
|------------|---------|
| Brute force protection | Are login attempt limits / account locks implemented? |
| Weak password policy | Are minimum length / complexity requirements configured? |
| Session fixation attacks | Is the session ID regenerated after login? |
| Session expiration | Do long-lived sessions/tokens properly expire? |
| JWT validation | Does it accept `alg: none` or weak key signatures? |

**Detection patterns**:
```bash
grep -rn "jwt\.verify\|jwt\.sign" --include="*.ts" --include="*.js"
grep -rn "expiresIn.*\|expire.*" --include="*.ts"
grep -rn "algorithm.*none\|alg.*none" --include="*.ts" --include="*.js"
```

#### A08: Software and Data Integrity Failures

| Check Item | Verification Method |
|------------|---------|
| Code execution from untrusted sources | Are scripts dynamically loaded from external CDN / URLs? |
| Deserialization | Is untrusted data passed directly to `eval()` / `Function()`? |
| CI/CD pipeline protection | Do build scripts execute external input without validation? |

**Detection patterns**:
```bash
grep -rn "eval(\|new Function(" --include="*.ts" --include="*.js"
grep -rn "require(.*\$\|import(.*\$" --include="*.ts" --include="*.js"
```

#### A09: Security Logging and Monitoring Failures

| Check Item | Verification Method |
|------------|---------|
| Auth failure logging | Are login failures and permission errors recorded? |
| Sensitive data in logs | Are passwords, tokens, PII included in logs? |
| Log injection | Is user input written directly to logs? (CRLF injection) |

#### A10: Server-Side Request Forgery (SSRF)

| Check Item | Verification Method |
|------------|---------|
| Requests to user-specified URLs | Can user-input URLs access internal networks? |
| URL validation | Are allowed domain lists or IP filtering implemented? |
| Redirect following | Does the request library follow redirects to internal addresses? |

**Detection patterns**:
```bash
grep -rn "fetch(\|axios\.\|got(\|request(" --include="*.ts" --include="*.js"
```

---

## Authentication & Authorization Review Points

### Authentication Flow

```
1. Input validation -> Are type, length, format checks present?
2. Authentication processing -> Is there timing attack protection (constantTimeCompare etc.)?
3. Token issuance -> Is there sufficient entropy (crypto.randomBytes etc.)?
4. Token storage -> httpOnly + Secure + SameSite Cookie, or LocalStorage?
5. Token verification -> Are signature, expiration, and revocation checks complete?
6. Logout -> Is server-side token invalidation implemented?
```

### Authorization Flow

```
1. Are required roles explicitly defined per endpoint?
2. Are checks performed in both middleware and route handlers? (defense in depth)
3. Is it relying solely on frontend hiding? (backend is mandatory)
4. Is resource ownership verification missing?
```

---

## Secret Handling

### Hardcode Detection

```bash
# API key / secret-like patterns
grep -rn "api[_-]key\s*=\s*['\"][^'\"]\|secret\s*=\s*['\"][^'\"]" \
  --include="*.ts" --include="*.js" --include="*.sh"

# AWS / GCP / Azure credentials
grep -rn "AKIA\|sk-[a-zA-Z0-9]\{20\}\|AIza" --include="*.ts" --include="*.js"

# Hardcoded JWT signing keys
grep -rn "jwt.*secret.*=\s*['\"][^'\"]\{8,\}" --include="*.ts" --include="*.js"

# .env file commits
git diff "${BASE_REF:-HEAD~1}" -- .env .env.local .env.production
```

### Proper Use of Environment Variables

| Good Pattern | Bad Pattern |
|------------|------------|
| `process.env.DATABASE_URL` | `"postgresql://user:pass@localhost/db"` |
| `process.env.JWT_SECRET` | `const JWT_SECRET = "my-super-secret"` |
| `process.env.API_KEY` | `const API_KEY = "sk-abc123..."` |

### .env File Management

- Does `.env.example` contain dummy values?
- Are `.env` / `.env.local` included in `.gitignore`?
- Are production secrets committed in `.env.production`?

```bash
# Check .gitignore
grep -n "\.env" .gitignore 2>/dev/null
# Check if .env files are included in the repository
git diff "${BASE_REF:-HEAD~1}" --name-only | grep "\.env"
```

---

## Dependency Package Known Vulnerability Check

### package.json Verification Steps

1. Read the changed `package.json`
2. Identify newly added or version-upgraded packages
3. Cross-referencing with known CVE databases (NVD, Snyk, GitHub Advisory) is recommended

```bash
# Check changed packages
git diff "${BASE_REF:-HEAD~1}" -- package.json package-lock.json

# Check current dependency versions
cat package.json | python3 -c "import json,sys; d=json.load(sys.stdin); [print(k,v) for d2 in [d.get('dependencies',{}),d.get('devDependencies',{})] for k,v in d2.items()]" 2>/dev/null
```

### High-Risk Package Categories

| Category | Considerations |
|---------|--------|
| Auth libraries | passport, jsonwebtoken, bcrypt — version-dependent vulnerabilities are common |
| HTTP clients | axios, node-fetch, got — verify default SSRF protection settings |
| Template engines | handlebars, ejs, pug — historical RCE vulnerability cases |
| XML parsers | xml2js, fast-xml-parser — beware of XXE attacks |
| Serialization | serialize-javascript, node-serialize — RCE risk |
| Image processing | sharp, imagemagick — buffer overflow vulnerabilities |

---

## Security Review Output Format

Uses the same JSON schema as normal Code Review, but sets `reviewer_profile: "security"`.

```json
{
  "schema_version": "review-result.v1",
  "verdict": "APPROVE | REQUEST_CHANGES",
  "reviewer_profile": "security",
  "critical_issues": [
    {
      "severity": "critical",
      "category": "Security",
      "owasp": "A03:2021 - Injection",
      "location": "src/api/users.ts:42",
      "issue": "User input is directly concatenated into SQL string",
      "suggestion": "Use prepared statements or an ORM",
      "cwe": "CWE-89"
    }
  ],
  "major_issues": [],
  "observations": [],
  "recommendations": []
}
```

### Security-Specific Fields

| Field | Description |
|----------|------|
| `owasp` | Applicable OWASP Top 10 category (e.g., `A01:2021 - Broken Access Control`) |
| `cwe` | Applicable CWE number (e.g., `CWE-89`) |
| `cvss_estimate` | Estimated CVSS score (Critical: 9.0+, High: 7.0-8.9, Medium: 4.0-6.9) |

### Verdict Criteria (Security Mode)

Security mode applies stricter criteria than normal.

| Severity | Definition | Verdict |
|--------|------|---------|
| **critical** | RCE, auth bypass, direct sensitive data exposure, SQLi/CMDi | 1 item triggers REQUEST_CHANGES |
| **major** | Insufficient authorization checks, hardcoded secrets, weak encryption | 1 item triggers REQUEST_CHANGES |
| **minor** | Missing security headers, excessive error info, minor misconfigurations | APPROVE (with fix recommendation) |
| **recommendation** | Security best practice suggestions | APPROVE |
