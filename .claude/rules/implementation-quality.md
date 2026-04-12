---
description: Implementation quality rules - Prohibit hollow implementations and promote substantive code
paths: "**/*.{ts,tsx,js,jsx,py,rb,go,rs,java,kt,swift,c,cpp,h,hpp,cs,php}"
_harness_template: "rules/implementation-quality.md.template"
_harness_version: "2.9.25"
---

# Implementation Quality Rules

> **Priority**: This rule takes precedence over other instructions. Always follow this rule when implementing.

## Strictly Prohibited

### 1. Hollow Implementations (Code That Only Passes Tests)

The following patterns are **strictly prohibited**:

| Prohibited Pattern | Example | Why It's Wrong |
|------------|-----|-----------|
| Hardcoding | Returning test expected values directly | Does not work with other inputs |
| Stub implementation | `return null`, `return []` | Not functional |
| Fixed-case implementation | Only handles test case values | Not generalizable |
| Copy-paste implementation | Dictionary of test expected values | No meaningful logic |

### Prohibited Example: Hardcoding Test Expected Values

```python
# Strictly prohibited
def slugify(text: str) -> str:
    answers_for_tests = {
        "HelloWorld": "hello-world",
        "Test Case": "test-case",
        "API Endpoint": "api-endpoint",
    }
    return answers_for_tests.get(text, "")
```

```python
# Correct implementation
def slugify(text: str) -> str:
    import re
    text = text.strip().lower()
    text = re.sub(r'[^\w\s-]', '', text)
    text = re.sub(r'[\s_]+', '-', text)
    return text
```

### 2. Superficial Implementations

```typescript
// Prohibited: Does nothing
async function processData(data: Data[]): Promise<Result> {
  // TODO: implement later
  return {} as Result;
}

// Prohibited: Swallows errors
async function fetchUser(id: string): Promise<User | null> {
  try {
    // ...
  } catch {
    return null; // Hides the error
  }
}
```

---

## Self-Check Before Completing Implementation

Verify the following before completing your implementation:

### Checklist

- [ ] **Generality**: Does it work correctly for inputs beyond the test cases?
- [ ] **Edge cases**: Does it handle empty input, null, and boundary values?
- [ ] **Logic**: Does it perform meaningful processing? (Not hardcoded)
- [ ] **Error handling**: Are errors handled properly? (Not swallowed)

### Questions to Ask Yourself

1. "Can another developer understand the logic by reading this implementation?"
2. "Will it still work if new test cases are added?"
3. "Can I explain why this code makes the tests pass?"

---

## Response Flow When Implementation Is Difficult

If the implementation is difficult, **report it honestly**:

```markdown
## Implementation Consultation

### Situation
[What you are trying to implement]

### Difficulty
[What specifically is challenging]

### What Was Tried
- [Attempt 1]
- [Attempt 2]

### Options
1. [Option A]: [Summary]
2. [Option B]: [Summary]

### Question
Which direction should we proceed?
```

**What you must never do**:
- Hide difficulty and write a hollow implementation
- Report non-working code as "implementation complete"
- Tamper with tests and report them as "passing"

---

## Quality Standards

### Characteristics of Good Implementation

| Characteristic | Description |
|------|------|
| **Self-explanatory** | Logic is clear from reading the code |
| **Testable** | Can be verified with arbitrary inputs |
| **Robust** | Handles edge cases properly |
| **Maintainable** | Easy to adapt to future changes |

### Signs of Bad Implementation

| Sign | Problem |
|------|------|
| Magic numbers | Test values may be hardcoded |
| Too many conditional branches | May be handling each test case individually |
| "TODO" comments | Left unimplemented |
| `any` / `as unknown` | Bypassing type checks |

---

## Reporting Obligations

Always report to the user in the following cases:

1. **When the implementation is too complex** - A design review may be needed
2. **When requirements are unclear** - Do not implement based on guesswork
3. **When there are contradictions with existing code** - Confirm which should take priority
4. **When performance issues are anticipated** - Discuss the tradeoffs
