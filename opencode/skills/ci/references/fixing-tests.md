---
name: ci-fix-failing-tests
description: "Guide for fixing tests that failed in CI. Use after CI failure causes have been identified to attempt automatic fixes."
allowed-tools: ["Read", "Edit", "Bash"]
---

# CI Fix Failing Tests

A skill for fixing tests that failed in CI.
Performs fixes to either the test code or the production code.

---

## Input

- **Failing test information**: Test name, error message
- **Test file**: Source of the failing test
- **Code under test**: Implementation being tested

---

## Output

- **Fixed code**: Fixes to tests or implementation
- **Test pass confirmation**

---

## Execution Steps

### Step 1: Identify Failing Tests

```bash
# Run tests locally
npm test 2>&1 | tail -50

# Run tests for a specific file
npm test -- {{test-file}}
```

### Step 2: Classify Error Type

#### Type A: Assertion Failure

```
Expected: "expected value"
Received: "actual value"
```

-> Implementation differs from expectation, or the test's expected value is wrong

#### Type B: Timeout

```
Timeout - Async callback was not invoked within the 5000ms timeout
```

-> Async operation not completing, or taking too long

#### Type C: Type Error

```
TypeError: Cannot read properties of undefined
```

-> Accessing null/undefined, or initialization issue

#### Type D: Mock-Related

```
expected mockFn to have been called
```

-> Insufficient mock setup, or the call was not made

### Step 3: Determine Fix Strategy

```markdown
## Fix Strategy Decision

1. **If the test is correct** -> Fix the implementation
2. **If the implementation is correct** -> Fix the test
3. **If both need fixes** -> Prioritize the implementation

Decision criteria:
- Which is correct according to specs/requirements
- What changed recently
- Impact on other tests
```

### Step 4: Implement the Fix

#### Fixing Assertion Failures

```typescript
// When the test's expected value is wrong
it('calculates correctly', () => {
  // Before fix
  expect(calculate(2, 3)).toBe(5)
  // After fix (when spec calls for multiplication)
  expect(calculate(2, 3)).toBe(6)
})

// When the implementation is wrong
// -> Fix the implementation file
```

#### Fixing Timeouts

```typescript
// Extend timeout
it('fetches data', async () => {
  // ...
}, 10000)  // Extended to 10 seconds

// Or use async/await correctly
it('fetches data', async () => {
  await waitFor(() => {
    expect(screen.getByText('Data')).toBeInTheDocument()
  })
})
```

#### Fixing Mock-Related Issues

```typescript
// Add mock setup
vi.mock('../api', () => ({
  fetchData: vi.fn().mockResolvedValue({ data: 'mock' })
}))

// Reset in beforeEach
beforeEach(() => {
  vi.clearAllMocks()
})
```

### Step 5: Verify After Fix

```bash
# Re-run failing tests
npm test -- {{test-file}}

# Run all tests (regression check)
npm test
```

---

## Fix Pattern Collection

### Snapshot Update

```bash
# Update snapshots
npm test -- -u

# Specific test only
npm test -- {{test-file}} -u
```

### Fixing Async Tests

```typescript
// Use findBy (auto-waits)
const element = await screen.findByText('Text')

// Use waitFor
await waitFor(() => {
  expect(mockFn).toHaveBeenCalled()
})
```

### Updating Mock Data

```typescript
// Update mocks to match implementation changes
const mockData = {
  id: 1,
  name: 'Test',
  createdAt: new Date().toISOString()  // New field
}
```

---

## Post-Fix Checklist

- [ ] Previously failing tests now pass
- [ ] No other tests are broken
- [ ] Consistent with the implementation intent
- [ ] Tests are not overly lenient

---

## Completion Report Format

```markdown
## ✅ Test Fix Complete

### Fix Details

| Test | Issue | Fix |
|------|-------|-----|
| `{{test name}}` | {{issue}} | {{fix details}} |

### Verification Results

```
Tests: {{passed}} passed, {{total}} total
```

### Next Actions

"Commit" or "Re-run CI"
```

---

## Notes

- **Do not delete tests**: Deletion is a last resort
- **Skip is temporary**: Permanent skips are prohibited
- **Identify the root cause**: Avoid superficial fixes
