# Source attribution

`pii-regex.json` is ported verbatim from:

- **Upstream**: [datumbrain/claude-privacy-guard](https://github.com/datumbrain/claude-privacy-guard)
- **Path**: `data/regex_list_1.json`
- **License**: MIT
- **Original work**: Copyright (c) Datum Brain

This file is loaded at runtime via `//go:embed` from `external.go` and
filtered through `isCodingSecretPattern` (keyword filter) before each pattern
is compiled by Go's RE2 engine. Patterns that RE2 cannot compile (lookahead,
backreferences, etc.) are silently skipped with a stderr warning.

## Updating the catalog

To pull a newer revision from upstream:

```bash
curl -fsSL https://raw.githubusercontent.com/datumbrain/claude-privacy-guard/main/data/regex_list_1.json \
  -o go/internal/piiguard/data/pii-regex.json
go test ./go/internal/piiguard/...
```

Run the tests after updating — the count and filter assertions will catch
any structural drift.
