---
name: auth
description: "Use when implementing authentication, OAuth, sessions, payments, subscriptions, or billing — including route protection, RBAC, and payment webhooks. Do NOT load for: general UI, schema design, or non-auth API endpoints."
allowed-tools: ["Read", "Write", "Edit", "Bash"]
user-invocable: false
---

# Auth Skills

A collection of skills responsible for implementing authentication and payment features.

## Feature Details

| Feature | Details |
|---------|--------|
| **Authentication** | See [references/authentication.md](${CLAUDE_SKILL_DIR}/references/authentication.md) |
| **Payments** | See [references/payments.md](${CLAUDE_SKILL_DIR}/references/payments.md) |

## Execution Steps

1. **Quality Gate** (Step 0)
2. Classify the user's request (authentication or payments)
3. Read the appropriate reference file from "Feature Details" above
4. Implement according to its contents

### Step 0: Quality Gate (Security Checklist)

Authentication and payment features always carry high security risk. Always display the following before starting work:

```markdown
🔐 Security Checklist

This work is security-critical. Please verify the following:

### Authentication
- [ ] Passwords are hashed (bcrypt/argon2)
- [ ] Session management is secure (HTTPOnly Cookie)
- [ ] CSRF protection is implemented
- [ ] Rate limiting (brute-force protection)

### Payments
- [ ] Sensitive information (card numbers, etc.) is not stored on the server
- [ ] Stripe/payment provider SDK is used correctly
- [ ] Webhook signature verification
- [ ] Amount tampering prevention (amounts finalized server-side)

### Common
- [ ] Error messages are not too detailed (prevent information leakage)
- [ ] Sensitive information is not logged
```

### Security Severity Display

```markdown
⚠️ Severity Level: 🔴 High

This feature carries the following risks:
- Credential leakage
- Unauthorized access
- Fraudulent payment operations

Expert review is recommended.
```

### For VibeCoders

```markdown
🔐 Building Login & Payment Features Safely

1. **Hash passwords**
   - Store passwords in an irreversible form
   - Data remains safe even if it leaks

2. **Do not store card information on your server**
   - Delegate to dedicated services like Stripe
   - Store nothing on your own server

3. **Keep error messages vague**
   - Use "Authentication failed" instead of "Wrong password"
   - Do not give hints to malicious actors
```
