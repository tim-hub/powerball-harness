# Typical Workflow Examples

Actual flow of the 2-agent workflow.

---

## Example 1: New Feature Development

### Phase 1: PM (Cursor) Creates the Task

```markdown
# Plans.md

## Not Started Tasks

- [ ] User profile editing feature `pm:requested`
  - Edit name, email, and avatar image
  - With validation
  - Save change history
```

**PM's statement**: "Please ask Claude Code to handle the profile editing feature"

---

### Phase 2: Claude Code Starts Work

```bash
# Run in Claude Code
harness-work
```

**Claude Code's work**:
1. Read Plans.md
2. Detect `pm:requested` task
3. Update marker to `cc:WIP`
4. Begin implementation
5. Quality review with `harness-review`
6. If issues found, fix and re-review (loop, max 3 times)
7. Review OK -> Auto-commit

```markdown
# Plans.md (after update)

## In Progress Tasks

- [ ] User profile editing feature `cc:WIP`
  - Edit name, email, and avatar image
  - With validation
  - Save change history
  - Related files:
    - `src/components/ProfileForm.tsx`
    - `src/lib/api/profile.ts`
```

---

### Phase 3: Claude Code Reports Completion (2-Agent only)

After Review OK and Auto-commit complete, in 2-Agent mode run `cc-cursor-cc` to report to the PM.

> **In Solo mode, handoff is not needed** -- Review OK -> Auto-commit completes `harness-work`.

```bash
# Run in Claude Code (2-Agent mode only)
cc-cursor-cc
```

**Generated report**:

```markdown
## Completion Report: User Profile Editing Feature

### Implementation Details
- Created ProfileForm component
- Profile API endpoint
- Validation with Zod
- Added change history table

### Changed Files
- src/components/ProfileForm.tsx (+150 lines)
- src/lib/api/profile.ts (+80 lines)
- src/lib/validations/profile.ts (+25 lines)
- prisma/schema.prisma (+10 lines)

### Review Results
harness-review APPROVE (No Critical/High issues)

### Test Results
All tests passed (12/12)

### Next Actions
- [ ] Verify behavior in staging environment
- [ ] Design review
```

---

### Phase 4: PM Confirms

```markdown
# Plans.md (after PM update)

## Completed Tasks

- [x] User profile editing feature `pm:confirmed` (2024-01-15)
```

---

## Example 2: Emergency Bug Fix

### Emergency Request from PM

```markdown
## Not Started Tasks

- [ ] [URGENT] Fix login error `pm:requested`
  - Symptom: Certain users cannot log in
  - Error: "Invalid token format"
  - Priority: Highest
```

### Claude Code's Response

1. Start with `harness-work`
2. Investigate error logs
3. Identify cause and fix
4. Add tests
5. Review with `harness-review` (fix and re-review if issues found)
6. Review OK -> Auto-commit
7. Report completion with `cc-cursor-cc` (2-Agent only; skip in Solo)

---

## Example 3: Auto-fix on CI Failure

### CI Fails

```
GitHub Actions: Build failed
- TypeScript error in src/utils/date.ts:45
```

### Claude Code's Automatic Response

1. Detect error
2. Fix type error
3. Re-commit and push

**If it fails 3 times**:

```markdown
## CI Escalation

Attempted 3 fixes but could not resolve the issue.

### Attempted Fixes
1. Added type annotations -> Failed
2. Updated type definition files -> Failed
3. Adjusted tsconfig -> Failed

### Estimated Cause
Type definitions for the external library may be outdated

### Recommended Actions
- [ ] Update @types/xxx to the latest version
- [ ] Check the library version itself
```

---

## Example 4: Parallel Task Execution

### When There Are Multiple Tasks

```markdown
## Not Started Tasks

- [ ] Refactor header component `cc:TODO`
- [ ] Refactor footer component `cc:TODO`
- [ ] Add tests: utility functions `cc:TODO`
```

### When `harness-work` Is Executed

Claude Code determines if parallel execution is possible:
- Independent tasks -> Parallel execution
- Dependencies exist -> Sequential execution

```
Parallel execution started
|- Agent 1: Header refactoring
|- Agent 2: Footer refactoring
|- Agent 3: Add tests
```
