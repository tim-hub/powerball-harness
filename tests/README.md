# Test Suite

This directory contains tests to ensure the quality of the claude-code-harness plugin.

## Tests for VibeCoders

These are not enterprise-level complex tests, but simple tests that allow a **solo VibeCoder working on client projects** to quickly verify that the plugin is functioning correctly.

## How to Run Tests

### Plugin Structure Validation

Validates that the basic plugin structure is correct:

```bash
./tests/validate-plugin.sh
./tests/validate-plugin-v3.sh
./scripts/ci/check-consistency.sh
```

### Unified Memory Validation

Validates basic operation of the shared memory daemon:

```bash
./tests/test-memory-daemon.sh
```

Loop-validates that no zombie processes remain:

```bash
./tests/test-memory-daemon-zombie.sh 100
```

Validates search quality (hybrid ranking / privacy filter / API routing):

```bash
./tests/test-memory-search-quality.sh
```

These validations check the following:

1. **Plugin structure**: plugin.json existence and validity
2. **Commands**: Existence of registered command files
3. **Skills**: Skill definitions existence and basic quality
4. **Agents**: Agent definitions existence
5. **Hooks**: hooks.json validity
6. **Scripts**: Automation script existence and execute permissions
7. **Documentation**: Required documents like README

### Expected Output

```
==========================================
Claude harness - Plugin Validation Test
==========================================

1. Plugin structure validation
----------------------------------------
✓ plugin.json exists
✓ plugin.json is valid JSON
✓ plugin.json has 'name' field
✓ plugin.json has 'version' field
...

==========================================
Test Results Summary
==========================================
Passed: 25
Warnings: 1
Failed: 0

✓ All tests passed!
```

## Adding Tests

When adding new commands or skills, run these tests to confirm the structure is correct.

## CI/CD Usage

GitHub Actions runs the following in `.github/workflows/validate-plugin.yml`:

- `./tests/validate-plugin.sh`
- `./scripts/ci/check-consistency.sh`
- `./tests/test-codex-package.sh`
- `cd core && npm test`

The `/harness-work all` success / failure fixtures are managed separately for smoke / full runs. See [docs/evidence/work-all.md](../docs/evidence/work-all.md) for details.

## Troubleshooting

### jq command not found

Test scripts use the `jq` command. If not installed:

```bash
# macOS
brew install jq

# Ubuntu/Debian
sudo apt-get install jq

# Windows (WSL)
sudo apt-get install jq
```

### When tests fail

1. Check the error message
2. Verify the referenced file exists
3. Check JSON files for syntax errors

## Key Points for VibeCoders

- **Simple**: No complex test frameworks needed
- **Practical**: Detects structural errors that actually cause problems
- **Fast**: Completes in seconds
- **Clear**: Results are obvious at a glance

These tests are for quickly checking "is anything broken?" after modifying the plugin.
