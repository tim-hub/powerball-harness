# Release Preflight

`scripts/release-preflight.sh` is a read-only check to proactively determine "is it safe to release now" before going public.
It assumes vendor-neutrality, so it does not depend on AWS or any specific deployment platform.

## What It Checks

- Whether the working tree is clean
- Whether `CHANGELOG.md` contains `[Unreleased]`
- Whether `.env.example` and `.env` are significantly out of sync. For repos without `.env`, it only warns to avoid blocking managed-secrets workflows
- Whether existing `healthcheck` / `preflight` commands pass
- Whether residuals like `mockData` / `dummy` / `localhost` / `TODO` / `FIXME` remain in shipped surfaces (`agents/` / `core/` / `hooks/` / `scripts/`)
- Whether the latest CI status is passing, when retrievable

## Usage

```bash
scripts/release-preflight.sh
scripts/release-preflight.sh --root /path/to/other/repo
```

## Environment Variables

- `HARNESS_RELEASE_PROJECT_ROOT`: Root path when checking a different repo
- `HARNESS_RELEASE_HEALTHCHECK_CMD`: Custom healthcheck command for repo-specific checks
- `HARNESS_RELEASE_CI_STATUS_CMD`: Command to override CI status retrieval

## Relationship with dry-run

`/release --dry-run` always runs preflight.
dry-run means "do not perform public operations," while preflight means "verify that the state is safe for public release."
These are separate concepts, so preflight is not skipped even in dry-run mode.
