#!/bin/bash
# release-preflight.sh
# Vendor-neutral pre-release verification for Harness release flow.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEFAULT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
PROJECT_ROOT="${HARNESS_RELEASE_PROJECT_ROOT:-$DEFAULT_ROOT}"

if [ "${1:-}" = "--help" ]; then
  cat <<'EOF'
Usage: scripts/release-preflight.sh [--root PATH] [--dry-run]

Checks:
  - git worktree cleanliness
  - CHANGELOG.md / [Unreleased]
  - env parity / healthcheck
  - runtime residual scan
  - CI status when available
EOF
  exit 0
fi

while [ "$#" -gt 0 ]; do
  case "$1" in
    --root)
      if [ "${2:-}" = "" ]; then
        echo "error: --root requires a path" >&2
        exit 2
      fi
      PROJECT_ROOT="$2"
      shift 2
      ;;
    --dry-run)
      shift
      ;;
    *)
      echo "error: unknown argument: $1" >&2
      exit 2
      ;;
  esac
done

if [ ! -d "$PROJECT_ROOT" ]; then
  echo "error: project root not found: $PROJECT_ROOT" >&2
  exit 1
fi

cd "$PROJECT_ROOT"

GIT_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || echo "$PROJECT_ROOT")"

PASS_COUNT=0
WARN_COUNT=0
FAIL_COUNT=0

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

pass() {
  echo -e "[PASS] $1"
  PASS_COUNT=$((PASS_COUNT + 1))
}

warn() {
  echo -e "[WARN] $1"
  WARN_COUNT=$((WARN_COUNT + 1))
}

fail() {
  echo -e "[FAIL] $1"
  FAIL_COUNT=$((FAIL_COUNT + 1))
}

run_optional_command() {
  local label="$1"
  local command_text="$2"
  local output_file
  output_file="$(mktemp /tmp/harness-tmp.XXXXXX)"

  if bash -lc "$command_text" >"$output_file" 2>&1; then
    pass "$label"
  else
    fail "$label"
    sed 's/^/  /' "$output_file"
  fi

  rm -f "$output_file"
}

extract_env_keys() {
  local file="$1"
  awk -F= '
    /^[[:space:]]*#/ { next }
    /^[[:space:]]*$/ { next }
    {
      line=$0
      sub(/^[[:space:]]*export[[:space:]]+/, "", line)
      split(line, parts, "=")
      key=parts[1]
      gsub(/[[:space:]]+$/, "", key)
      if (key ~ /^[A-Za-z_][A-Za-z0-9_]*$/) {
        print key
      }
    }
  ' "$file" | sort -u
}

check_git_clean() {
  if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    fail "git worktree"
    return
  fi

  local status
  status="$(git status --porcelain --untracked-files=normal)"
  if [ -n "$status" ]; then
    fail "working tree clean"
    printf '%s\n' "$status" | sed 's/^/  /'
  else
    pass "working tree clean"
  fi
}

check_changelog() {
  local changelog="$GIT_ROOT/CHANGELOG.md"
  if [ ! -f "$changelog" ]; then
    fail "CHANGELOG.md not found"
    return
  fi

  if grep -q '^\## \[Unreleased\]' "$changelog"; then
    pass "CHANGELOG.md has [Unreleased]"
  else
    fail "CHANGELOG.md has [Unreleased]"
  fi
}

check_env_and_healthcheck() {
  local env_ok=1

  if [ -f .env.example ]; then
    if [ ! -f .env ]; then
      warn ".env missing for .env.example"
      env_ok=0
    else
      local missing
      missing="$(
        comm -23 \
          <(extract_env_keys .env.example) \
          <(extract_env_keys .env)
      )"
      if [ -n "$missing" ]; then
        fail ".env matches .env.example"
        printf '%s\n' "$missing" | sed 's/^/  missing: /'
        env_ok=0
      else
        pass ".env matches .env.example"
      fi
    fi
  else
    warn ".env.example not found; env parity skipped"
  fi

  if [ -n "${HARNESS_RELEASE_HEALTHCHECK_CMD:-}" ]; then
    run_optional_command "healthcheck command" "$HARNESS_RELEASE_HEALTHCHECK_CMD"
    return
  fi

  if [ -f package.json ]; then
    local has_healthcheck=0
    local has_preflight=0

    if node -e 'const fs=require("fs"); const pkg=JSON.parse(fs.readFileSync("package.json","utf8")); process.exit(pkg.scripts && pkg.scripts.healthcheck ? 0 : 1)' >/dev/null 2>&1; then
      has_healthcheck=1
    fi

    if node -e 'const fs=require("fs"); const pkg=JSON.parse(fs.readFileSync("package.json","utf8")); process.exit(pkg.scripts && pkg.scripts.preflight ? 0 : 1)' >/dev/null 2>&1; then
      has_preflight=1
    fi

    if [ "$has_healthcheck" -eq 1 ]; then
      run_optional_command "healthcheck command" "npm run healthcheck --silent"
      return
    fi

    if [ "$has_preflight" -eq 1 ]; then
      run_optional_command "healthcheck command" "npm run preflight --silent"
      return
    fi
  fi

  if [ "$env_ok" -eq 1 ]; then
    warn "healthcheck command not configured"
  elif [ ! -f .env ] && [ -f .env.example ]; then
    warn "healthcheck command not configured"
  fi
}

check_runtime_residuals() {
  local residual_patterns="${HARNESS_RELEASE_RESIDUAL_PATTERNS:-mockData|dummy|fakeData|localhost|TODO|FIXME|test\\.skip|describe\\.skip|it\\.skip}"
  local files=()
  local file

  while IFS= read -r -d '' file; do
    case "$file" in
      agents/*|core/*|hooks/*|scripts/*)
        if [ "$file" = "scripts/release-preflight.sh" ]; then
          continue
        fi
        files+=("$file")
        ;;
      *)
        continue
        ;;
    esac
  done < <(git ls-files -z)

  if [ "${#files[@]}" -eq 0 ]; then
    warn "runtime residual scan skipped"
    return
  fi

  local matches
  if command -v rg >/dev/null 2>&1; then
    matches="$(rg -n -I --no-heading -e "$residual_patterns" -- "${files[@]}" 2>/dev/null || true)"
  else
    matches="$(grep -nIH -E "$residual_patterns" -- "${files[@]}" 2>/dev/null || true)"
  fi
  if [ -n "$matches" ]; then
    warn "runtime residual scan"
    printf '%s\n' "$matches" | head -20 | sed 's/^/  /'
  else
    pass "runtime residual scan"
  fi
}

check_ci_status() {
  if [ -n "${HARNESS_RELEASE_CI_STATUS_CMD:-}" ]; then
    run_optional_command "CI status" "$HARNESS_RELEASE_CI_STATUS_CMD"
    return
  fi

  if ! command -v gh >/dev/null 2>&1; then
    warn "CI status unavailable (gh not installed)"
    return
  fi

  local branch
  branch="$(git branch --show-current 2>/dev/null || true)"
  if [ -z "$branch" ] || [ "$branch" = "HEAD" ]; then
    warn "CI status unavailable (detached HEAD)"
    return
  fi

  local gh_output
  gh_output="$(gh run list --branch "$branch" --limit 1 --json status,conclusion 2>/dev/null || true)"
  if [ -z "$gh_output" ]; then
    warn "CI status unavailable (no GitHub Actions data)"
    return
  fi

  local status
  local conclusion
  if command -v jq >/dev/null 2>&1; then
    status="$(printf '%s' "$gh_output" | jq -r '.[0].status // empty')"
    conclusion="$(printf '%s' "$gh_output" | jq -r '.[0].conclusion // empty')"
  else
    status="$(printf '%s' "$gh_output" | sed -n 's/.*"status":"\([^"]*\)".*/\1/p' | head -n 1)"
    conclusion="$(printf '%s' "$gh_output" | sed -n 's/.*"conclusion":"\([^"]*\)".*/\1/p' | head -n 1)"
  fi

  if [ "$status" = "completed" ] && [ "$conclusion" = "success" ]; then
    pass "CI status"
  else
    fail "CI status"
    printf '  latest run status=%s conclusion=%s\n' "${status:-unknown}" "${conclusion:-unknown}"
  fi
}

printf 'Release preflight: %s\n' "$PROJECT_ROOT"
echo "----------------------------------------"

check_git_clean
check_changelog
check_env_and_healthcheck
check_runtime_residuals
check_ci_status

echo "----------------------------------------"
printf 'Summary: %d passed, %d warnings, %d failed\n' "$PASS_COUNT" "$WARN_COUNT" "$FAIL_COUNT"

if [ "$FAIL_COUNT" -gt 0 ]; then
  exit 1
fi

exit 0
