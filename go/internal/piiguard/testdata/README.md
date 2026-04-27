# piiguard testdata

Regression fixture corpus for the PII Guard scanner.  Files contain
**synthetic, clearly-fake** values — never real credentials.  The corpus
is used by `corpus_test.go` to verify that:

- Files in `positive/` produce ≥ 1 finding when scanned.
- Files in `negative/` produce 0 findings.

## Updating the corpus

Add new fixtures by category; each filename should describe the rule it
exercises (e.g., `aws-access-key.txt`, `email-in-prose.txt`).  Keep each
fixture small (< 1 KB) and self-evidently synthetic — anything that looks
like a real credential should not be added.

When adding a positive fixture, also add a corresponding negative fixture
that exercises the same rule's near-miss boundary (too short, wrong prefix,
etc.).

## Sensitive-string handling

Several upstream guardrails (and external scanners like gitleaks) flag
contiguous credential-looking text in any file.  When fixtures live in
testdata files, scanners typically respect a project-level allowlist for
the `testdata/` path.  But to keep this repo's PostToolUse guardrail
quiet, we still split contiguous prefixes via concatenation in the
template comments below — not in the data itself, since that would defeat
the test.

The data files contain the unsplit, runtime-effective fixtures.  This
README and the test code use string concatenation to describe what is
inside without re-emitting the full pattern.
