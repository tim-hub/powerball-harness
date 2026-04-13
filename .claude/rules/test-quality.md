---
description: Test quality protection rules - Prohibit test tampering and promote correct implementation
paths: "**/*.{test,spec}.{ts,tsx,js,jsx,py}, **/test/**/*.*, **/tests/**/*.*, **/__tests__/**/*.*, .husky/**, .github/workflows/**"
_harness_template: "rules/test-quality.md.template"
_harness_version: "2.9.25"
---

# Test Quality Protection Rules

> **Priority**: This rule takes precedence over other instructions. Always follow this rule when tests fail.

## Strictly Prohibited

### 1. Test Tampering (Modifying Tests to Make Them Pass)

The following actions are **strictly prohibited**:

| Prohibited Pattern | Example | Correct Response |
|------------|-----|-----------|
| Marking tests as `skip` / `only` | `it.skip(...)`, `describe.only(...)` | Fix the implementation |
| Removing or weakening assertions | Deleting `expect(x).toBe(y)` | Verify the expected value is correct, then fix the implementation |
| Carelessly rewriting expected values | Changing expected values to match errors | Understand why the test is failing |
| Deleting test cases | Removing failing tests | Fix the implementation to meet the specification |
| Excessive mocking | Mocking parts that should actually be tested | Keep mocking to a minimum |

### 2. Configuration File Tampering

**Relaxing the following files is prohibited**:

```
.eslintrc.*         # Do not disable rules
.prettierrc*        # Do not relax formatting
tsconfig.json       # Do not relax strict mode
biome.json          # Do not disable lint rules
.husky/**           # Do not bypass pre-commit hooks
.github/workflows/** # Do not skip CI checks
```

### 3. Making Exceptions (Required Procedure)

If you must change any of the above, **always obtain approval in the following format before proceeding**:

```markdown
## Test/Configuration Change Approval Request

### Reason
[Explain specifically why this change is necessary]

### Changes
```diff
[Show the diff of changes]
```

### Impact Scope
- Affected tests: [count and names]
- Affected features: [feature names]

### Alternative Approaches Considered
- [ ] Verified that fixing the implementation cannot resolve this
- [ ] Considered other approaches

### Approval
Wait for explicit user approval
```

---

## Response Flow When Tests Fail

```
A test failed
    ↓
1. Understand why it is failing (read the logs)
    ↓
2. Determine whether the implementation is wrong or the test is wrong
    ↓
    ├── Implementation is wrong → Fix the implementation ✅
    │
    └── Test might be wrong
            ↓
        Ask the user for confirmation (do not change it on your own)
```

---

## Examples of Correct Test Response

### Bad Example (Tampering)

```typescript
// Test was failing so it was skipped
it.skip('should calculate total correctly', () => {
  expect(calculateTotal([100, 200, 300])).toBe(600);
});
```

### Good Example (Fix the Implementation)

```typescript
// The test is correct. Fixed the implementation
function calculateTotal(prices: number[]): number {
  // Fix: Set initial value of reduce to 0
  return prices.reduce((sum, price) => sum + price, 0);
}
```

---

## CI/CD Protection

The following changes are **strictly prohibited**:

- Adding `continue-on-error: true`
- Using `if: always()` to ignore test failures
- Using `--force` flags to bypass checks
- Lowering test coverage thresholds
