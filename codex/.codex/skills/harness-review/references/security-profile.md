# Security Reviewer Profile

A security-specific review profile launched with `harness-review --security`.
Comprehensively checks authentication, authorization, secrets, and dependency vulnerabilities based on the OWASP Top 10.

> **Read-only constraint**: The reviewer operating under this profile uses
> only Read / Grep / Glob / Bash (read-only commands).
> Write / Edit / write-mode Bash are never executed.

---

## Security Review Flow

### Step 1: Identify Target Scope

```bash
# Collect changed files (BASE_REF is inherited from the caller)
CHANGED_FILES="$(git diff --name-only --diff-filter=ACMR "${BASE_REF:-HEAD~1}")"
git diff "${BASE_REF:-HEAD~1}" -- ${CHANGED_FILES}
```

### Step 2: OWASP Top 10 Check

Verify each of the following items against the **change diff** and **related files**.

#### A01: Broken Access Control

| Check Item | Verification Method |
|-----------|-------------------|
| Missing authorization check | Whether auth middleware is applied to route/endpoint definitions |
| Horizontal privilege escalation | Whether resource retrieval filters by `userId` or similar |
| Vertical privilege escalation | Whether role checks (admin/user/guest, etc.) are properly implemented |
| IDOR | Whether IDs in URL params or request body are accepted without authorization |
| Directory traversal | Whether path operations containing `../` are sanitized |

**Detection patterns (verify with Grep)**:
```bash
# Candidate routes without authentication
grep -rn "app\.\(get\|post\|put\|delete\|patch\)" --include="*.ts" --include="*.js"
# DB queries without userId
grep -rn "findById\|findOne\|select.*where" --include="*.ts"
```

#### A02: Cryptographic Failures

| Check Item | Verification Method |
|-----------|-------------------|
| Plaintext storage of sensitive data | Whether passwords, tokens, PII are stored in plaintext in DB/logs |
| Weak hash algorithms | Whether MD5 / SHA1 are used for password hashing |
| Insecure randomness | Whether `Math.random()` is used for auth token generation |
| TLS strength | Whether sensitive data is transmitted over HTTP (non-HTTPS) |
| Hardcoded keys | Whether encryption keys/IVs are embedded as constants |

**Detection patterns**:
```bash
grep -rn "md5\|sha1\|Math\.random\(\)" --include="*.ts" --include="*.js"
grep -rn "createHash.*md5\|createHash.*sha1" --include="*.ts"
grep -rn "http://" --include="*.ts" --include="*.js" --include="*.env*"
```

#### A03: Injection

| Check Item | Verification Method |
|-----------|-------------------|
| SQL injection | Whether user input is concatenated directly into SQL strings |
| NoSQL injection | Whether `$where` or input values are used as operators in MongoDB, etc. |
| Command injection | Whether user input is passed to `exec()` / `spawn()` |
| LDAP injection | Whether unsanitized input is used in LDAP queries |
| Template injection | Whether user input is passed directly to template engines |

**Detection patterns**:
```bash
grep -rn "exec\|execSync\|spawn" --include="*.ts" --include="*.js"
grep -rn "\`SELECT\|\"SELECT\|'SELECT" --include="*.ts" --include="*.js"
grep -rn "\$where\|\$\[" --include="*.ts" --include="*.js"
```

#### A04: Insecure Design

| Check Item | Verification Method |
|-----------|-------------------|
| Lack of rate limiting | Whether rate limiting is implemented on auth endpoints |
| TOCTOU race condition | Whether state changes between check and use can be exploited |
| Business logic flaws | Whether state transitions can be executed in invalid order |

#### A05: Security Misconfiguration

| Check Item | Verification Method |
|-----------|-------------------|
| Default credentials | Whether default passwords/usernames are left in use |
| Verbose error messages | Whether stack traces or internal info are returned to clients in production |
| Unnecessary features enabled | Whether debug endpoints/admin panels are enabled in production |
| HTTP security headers | Whether HSTS, CSP, X-Frame-Options, etc. are configured |
| CORS configuration | Whether `Access-Control-Allow-Origin: *` is set in production |

**Detection patterns**:
```bash
grep -rn "cors.*origin.*\*\|allowedOrigins.*\*" --include="*.ts" --include="*.js"
grep -rn "debug.*true\|NODE_ENV.*development" --include="*.ts"
grep -rn "console\.log.*password\|console\.log.*token\|console\.log.*secret" --include="*.ts"
```

#### A06: Vulnerable and Outdated Components

| Check Item | Verification Method |
|-----------|-------------------|
| Packages with known vulnerabilities | Whether any versions in `package.json` dependencies have reported CVEs |
| `npm audit` results | Whether high / critical vulnerabilities are left unaddressed |
| Lock file consistency | Whether `package-lock.json` / `yarn.lock` is up to date |

**Verification commands**:
```bash
# Check package.json dependencies (read-only)
cat package.json | grep -E '"dependencies"|"devDependencies"' -A 50 | head -60
# Check lock file existence
ls -la package-lock.json yarn.lock pnpm-lock.yaml 2>/dev/null
```

#### A07: Identification and Authentication Failures

| Check Item | Verification Method |
|-----------|-------------------|
| Brute force protection | Whether login attempt limits / account lockout are implemented |
| Weak password policy | Whether minimum length / complexity requirements are configured |
| Session fixation attack | Whether session ID is regenerated after login |
| Session expiration | Whether long-lived sessions/tokens expire appropriately |
| JWT validation | Whether `alg: none` or weak key signatures are accepted |

**Detection patterns**:
```bash
grep -rn "jwt\.verify\|jwt\.sign" --include="*.ts" --include="*.js"
grep -rn "expiresIn.*\|expire.*" --include="*.ts"
grep -rn "algorithm.*none\|alg.*none" --include="*.ts" --include="*.js"
```

#### A08: Software and Data Integrity Failures

| Check Item | Verification Method |
|-----------|-------------------|
| Code execution from untrusted sources | Whether scripts are dynamically loaded from external CDNs / URLs |
| Deserialization | Whether untrusted data is passed directly to `eval()` / `Function()` |
| CI/CD pipeline protection | Whether build scripts execute external input without validation |

**Detection patterns**:
```bash
grep -rn "eval(\|new Function(" --include="*.ts" --include="*.js"
grep -rn "require(.*\$\|import(.*\$" --include="*.ts" --include="*.js"
```

#### A09: Security Logging and Monitoring Failures

| Check Item | Verification Method |
|-----------|-------------------|
| Logging of auth failures | Whether login failures / permission errors are logged |
| Sensitive data in logs | Whether passwords / tokens / PII are included in logs |
| Log injection | Whether user input is written directly to logs (CRLF injection) |

#### A10: Server-Side Request Forgery (SSRF)

| Check Item | Verification Method |
|-----------|-------------------|
| Requests to user-specified URLs | Whether user-input URLs can access internal networks |
| URL validation | Whether allowed domain lists or IP filtering are implemented |
| Redirect following | Whether the request library follows redirects to internal addresses |

**Detection patterns**:
```bash
grep -rn "fetch(\|axios\.\|got(\|request(" --include="*.ts" --include="*.js"
```

---

## Authentication & Authorization Review Points

### Authentication Flow

```
1. Input validation -> Are type/length/format checks present?
2. Authentication processing -> Is timing attack protection (constantTimeCompare, etc.) present?
3. Token issuance -> Is there sufficient entropy (crypto.randomBytes, etc.)?
4. Token storage -> httpOnly + Secure + SameSite Cookie, or LocalStorage?
5. Token validation -> Are signature/expiration/revocation checks complete?
6. Logout -> Is server-side token invalidation implemented?
```

### Authorization Flow

```
1. Are required roles explicitly defined per endpoint?
2. Are checks performed in both middleware and route handlers (defense in depth)?
3. Is there no reliance solely on frontend hiding (backend enforcement required)?
4. Is resource ownership verification not missing?
```

---

## Secret Handling

### Hardcoded Secret Detection

```bash
# Patterns resembling API keys / secrets
grep -rn "api[_-]key\s*=\s*['\"][^'\"]\|secret\s*=\s*['\"][^'\"]" \
  --include="*.ts" --include="*.js" --include="*.sh"

# AWS / GCP / Azure credentials
grep -rn "AKIA\|sk-[a-zA-Z0-9]\{20\}\|AIza" --include="*.ts" --include="*.js"

# Hardcoded JWT signing keys
grep -rn "jwt.*secret.*=\s*['\"][^'\"]\{8,\}" --include="*.ts" --include="*.js"

# .env files committed
git diff "${BASE_REF:-HEAD~1}" -- .env .env.local .env.production
```

### Proper Use of Environment Variables

| Good Pattern | Bad Pattern |
|-------------|------------|
| `process.env.DATABASE_URL` | `"postgresql://user:pass@localhost/db"` |
| `process.env.JWT_SECRET` | `const JWT_SECRET = "my-super-secret"` |
| `process.env.API_KEY` | `const API_KEY = "sk-abc123..."` |

### .env File Management

- Does `.env.example` contain dummy values?
- Are `.env` / `.env.local` included in `.gitignore`?
- Are production secrets not committed in `.env.production`?

```bash
# Check .gitignore
grep -n "\.env" .gitignore 2>/dev/null
# Check whether .env files are included in the repository
git diff "${BASE_REF:-HEAD~1}" --name-only | grep "\.env"
```

---

## Known Vulnerability Check for Dependencies

### package.json Verification Steps

1. Read the changed `package.json`
2. Identify newly added or version-upgraded packages
3. Cross-reference with known CVE databases (NVD, Snyk, GitHub Advisory) recommended

```bash
# Check changed packages
git diff "${BASE_REF:-HEAD~1}" -- package.json package-lock.json

# Check current dependency versions
cat package.json | python3 -c "import json,sys; d=json.load(sys.stdin); [print(k,v) for d2 in [d.get('dependencies',{}),d.get('devDependencies',{})] for k,v in d2.items()]" 2>/dev/null
```

### High-Risk Package Categories

| Category | Notes |
|---------|-------|
| Authentication libraries | passport, jsonwebtoken, bcrypt -- many version-dependent vulnerabilities |
| HTTP clients | axios, node-fetch, got -- verify default settings for SSRF protection |
| Template engines | handlebars, ejs, pug -- past RCE vulnerability cases |
| XML parsers | xml2js, fast-xml-parser -- watch for XXE attacks |
| Serialization | serialize-javascript, node-serialize -- RCE risk |
| Image processing | sharp, imagemagick -- buffer overflow vulnerabilities |

---

## Security Review Output Format

Uses the same JSON schema as the standard Code Review, but with `reviewer_profile: "security"` set.

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
|-------|-------------|
| `owasp` | Applicable OWASP Top 10 category (e.g., `A01:2021 - Broken Access Control`) |
| `cwe` | Applicable CWE number (e.g., `CWE-89`) |
| `cvss_estimate` | Estimated CVSS score (Critical: 9.0+, High: 7.0-8.9, Medium: 4.0-6.9) |

### Verdict Criteria (Security Mode)

Security mode applies stricter criteria than normal.

| Severity | Definition | Verdict |
|----------|-----------|---------|
| **critical** | RCE, auth bypass, direct exposure of sensitive data, SQLi/CMDi | REQUEST_CHANGES on even 1 occurrence |
| **major** | Insufficient authorization checks, hardcoded secrets, weak cryptography | REQUEST_CHANGES on even 1 occurrence |
| **minor** | Missing security headers, excessive error info, minor misconfigurations | APPROVE (with fix recommendations) |
| **recommendation** | Security best practice suggestions | APPROVE |
