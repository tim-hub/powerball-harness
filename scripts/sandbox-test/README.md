# Sandbox Test

> Test directory for `/work --full` verification

## Purpose

This directory was created to verify the `/work --full` command and `task-worker` agent added in Claude harness v2.9.0.

## File Structure

| File | Description |
|---------|------|
| `greeting.ts` | Test utility functions |
| `greeting.test.ts` | Unit tests (Vitest) |
| `README.md` | This file |

## Running Tests

```bash
# If Vitest is installed
npx vitest run scripts/sandbox-test/

# Or
bun test scripts/sandbox-test/
```

## /work --full Test Results

This directory was generated with the following command:

```bash
/work --full --parallel 3
```

### Expected Behavior

1. **Phase 1**: 3 task-workers launched in parallel
   - task-worker #1: Create `greeting.ts`
   - task-worker #2: Create `greeting.test.ts`
   - task-worker #3: Create `README.md`

2. **Phase 2**: Codex 8-parallel cross-review (optional)

3. **Phase 3**: Conflict resolution -> Commit

## Related Documentation

- [/work --full Documentation](../../docs/PARALLEL_FULL_CYCLE.md)
- [Worker Agent](../../agents/worker.md)
