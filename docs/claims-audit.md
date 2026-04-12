# Claims Audit

Last updated: 2026-03-06

This document is an audit memo that classifies public-facing claims by whether they are "currently proven" or "need additional evidence."
When updating README or release copy, review this table first.

## Current Classification

| Claim | Status | Current evidence | Before stronger wording |
|------|--------|------------------|-------------------------|
| Harness is built around **5 verb skills** | Proven now | `skills-v3/`, `README`, `validate-plugin-v3.sh` | None |
| Harness uses a **TypeScript guardrail engine** | Proven now | `core/`, `core npm test`, `hooks/` | None |
| README / docs / Plans no longer contradict each other on version and missing links | Proven now | `README*`, `docs/CLAUDE_CODE_COMPATIBILITY.md`, `docs/CURSOR_INTEGRATION.md`, `check-consistency.sh` | Continue simultaneous updates when changing documentation |
| `commands/` and `mcp-server/` are intentionally retained with clear boundaries | Proven now | `docs/distribution-scope.md`, `.gitignore`, `Plans.md` wording repair | Update scope table simultaneously when boundaries change |
| `/harness-work all` has a rerunnable success/failure contract | Proven now | `docs/evidence/work-all.md`, fixture smoke, failure contract, success replay-fallback artifact | If strict-live success artifact becomes available, live proof can also be added |
| `/harness-work all` can be trusted as a default production path | Not yet safe to claim strongly | README now avoids this wording | Stable reproduction of a full success run; add CI or captured artifact if needed |
| Codex setup and path-based loading are aligned with current package layout | Proven now | `codex/README.md`, `tests/test-codex-package.sh`, setup script fixes | Continue verifying path-based loading on real devices |
| Cursor 2-agent workflow is documented | Proven as documentation | `docs/CURSOR_INTEGRATION.md` | Can be strengthened with real environment screenshots or smoke logs |
| README includes a dated feature matrix against popular GitHub harness plugins | Proven as dated snapshot | `docs/github-harness-plugin-benchmark.md`, linked GitHub repos, README / README_ja comparison table | Update stars and comparison targets before release |

## Notes

- The 2026-03-06 success full runner was modified to automatically fall back to replay overlay when Claude Code usage limits (`You've hit your limit · resets 12pm (Asia/Tokyo)`) are detected.
- As a result, artifact generation itself is not blocked by quota. However, **evidence that only live Claude runs completed the entire process** requires a separate `--strict-live` success artifact.
- The failure path is structured to make it easy to verify the "do not commit while red" contract.
