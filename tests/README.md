# Test Suite

This directory contains tests that ensure the quality of the claude-code-harness plugin.

## Tests for VibeCoder

Simple tests designed so that a **VibeCoder working solo on client projects** can quickly confirm the plugin is working correctly — no enterprise-level complexity required.

## How to Run Tests

### Plugin Structure Validation

Validates that the basic plugin structure is correct:

```bash
./tests/validate-plugin.sh
./tests/validate-plugin-v3.sh
./local-scripts/check-consistency.sh
```

### Unified Memory Validation

Validates basic operation of the shared memory daemon:

```bash
./tests/test-memory-daemon.sh
```

Loop validation to check that no zombie processes remain:

```bash
./tests/test-memory-daemon-zombie.sh 100
```

Validates search quality (hybrid ranking / privacy filter / API routing):

```bash
./tests/test-memory-search-quality.sh
```

These validations check:

1. **Plugin structure**: plugin.json existence and validity
2. **Commands**: existence of registered command files
3. **Skills**: skill definition existence and basic quality
4. **Agents**: agent definition existence
5. **Hooks**: hooks.json validity
6. **Scripts**: automation script existence and execute permissions
7. **Documentation**: required documentation such as README

### Expected Output

```
==========================================
Claude harness - Plugin validation test
==========================================

1. Plugin structure validation
----------------------------------------
✓ plugin.json exists
✓ plugin.json is valid JSON
✓ plugin.json has name field
✓ plugin.json has version field
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

When adding new commands or skills, run this test to confirm the structure is correct.

## Usage in CI/CD

GitHub Actions runs the following via `.github/workflows/validate-plugin.yml`:

- `./tests/validate-plugin.sh`
- `./local-scripts/check-consistency.sh`
- `./tests/test-codex-package.sh`
- `cd core && npm test`

The `/harness-work all` success / failure fixtures are managed separately for smoke / full. See [docs/evidence/work-all.md](../docs/evidence/work-all.md) for details.

## Troubleshooting

### jq command not found

Test scripts use the `jq` command. If it is not installed:

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
2. Verify that the relevant file exists
3. Check for JSON syntax errors

## Key Points for VibeCoder

- **Simple**: No complex test frameworks needed
- **Practical**: Detects structural errors that actually cause problems
- **Fast**: Completes in seconds
- **Clear**: Results are easy to read at a glance

These tests are for quickly checking "is anything broken?" after modifying the plugin.
