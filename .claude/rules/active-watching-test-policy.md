# Active-Watching Test Policy

Quality standards for features that probe an external daemon or optional dependency.
Introduced in Phase 80 (v4.11.x) after the tri-state health fix to `go/cmd/harness/mem.go`.

## Purpose

When a feature probes an optional external daemon (e.g., a memory backend, language server, or socket service),
three distinct runtime states exist. Each state must be covered by a dedicated test so that regressions
cannot silently collapse multiple states into one.

This policy codifies the naming convention (`_NotConfigured`, `_Unreachable`, `_Corrupted`) and the
injection-variable pattern that enables deterministic unit testing without a live daemon.

---

## (a) Three Mandatory Test States

Every feature that probes an external daemon **must** have tests covering all three states below.
Missing any one state is a `major` review finding (triggers `REQUEST_CHANGES`).

| Test suffix | Condition | Expected result |
|-------------|-----------|-----------------|
| `_NotConfigured` | Dependency not installed / directory absent | `healthy:true`, no warning emitted, daemon probe **NOT** called |
| `_Unreachable` | Dependency installed but daemon unreachable (e.g., socket refused) | `healthy:false`, warning emitted |
| `_Corrupted` | Dependency installed but state is malformed (e.g., bad JSON, missing keys) | `healthy:false`, warning emitted |

### Naming Convention

Use the suffix directly on the test function name:

```go
func TestMemHealth_NotConfigured(t *testing.T) { ... }
func TestMemHealth_Unreachable(t *testing.T)   { ... }
func TestMemHealth_Corrupted(t *testing.T)     { ... }
```

The suffix must be one of the three canonical names above — do not abbreviate or rename them.
This makes `go test -run _NotConfigured` filtering reliable across all features.

---

## (b) Injection Hook Requirement

Probes that call external daemons **must** be defined as package-level `var` functions so that
tests can inject stubs without starting a real daemon.

### Pattern (from `go/cmd/harness/mem.go`)

```go
// daemonProbe is a package-level variable for test injection.
// Production code uses net.DialTimeout; tests swap this out.
var daemonProbe = func(addr string, timeout time.Duration) error {
    conn, err := net.DialTimeout("tcp", addr, timeout)
    if err != nil {
        return err
    }
    conn.Close()
    return nil
}
```

### Test Stub Pattern

Tests inject stubs via a save-and-restore idiom:

```go
orig := daemonProbe
defer func() { daemonProbe = orig }()

daemonProbe = func(addr string, timeout time.Duration) error {
    return fmt.Errorf("connection refused")   // simulate _Unreachable
}
```

### Rules

1. The `var` declaration must be at package level, not inside a function.
2. The variable name must end in `Probe` (e.g., `daemonProbe`, `socketProbe`, `lspProbe`) so that
   code search can locate all injection points: `grep -r 'Probe = func'`.
3. The production implementation must be the default value of the `var` — not assigned in `init()`.
4. Never call the daemon directly inside business logic; always call it through the `var` function.

---

## (c) Exit-Code Contract

`_NotConfigured` **must not** exit non-zero.

An absent optional dependency is not a failure. The health response must be:

```json
{ "healthy": true, "reason": "not-configured" }
```

and the process must exit 0.

**Why**: Operators run `harness doctor` in CI pipelines. A missing optional integration must not
break CI — it must be a no-op until the integration is actually configured.

### Corollary for `_Unreachable` and `_Corrupted`

These states **must** exit non-zero (or set `healthy: false` in the structured response) so that
monitoring tools can detect daemon failures automatically.

---

## Enforcement

| Gate | Check |
|------|-------|
| Pre-merge review | Reviewer verifies all three test suffixes exist for any new daemon probe |
| `harness-review` | Flags missing `_NotConfigured` / `_Unreachable` / `_Corrupted` as a `major` finding |
| `tests/validate-plugin.sh` | Future section (planned): grep for `daemonProbe` without matching test suffixes |

---

## Reference Implementation

- `go/cmd/harness/mem.go` — canonical `daemonProbe` var and tri-state health logic
- `go/cmd/harness/mem_test.go` — `TestMemHealth_NotConfigured`, `TestMemHealth_Unreachable`, `TestMemHealth_Corrupted`
