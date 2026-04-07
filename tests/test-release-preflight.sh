#!/bin/bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

fail() {
  echo "FAIL: $1" >&2
  exit 1
}

assert_contains() {
  local file="$1"
  local pattern="$2"
  if ! grep -q "$pattern" "$file"; then
    fail "expected pattern '$pattern' in $file"
  fi
}

setup_repo() {
  local repo="$1"
  mkdir -p "$repo/scripts"
  git init -q "$repo"
  git -C "$repo" config user.name "Test User"
  git -C "$repo" config user.email "test@example.com"

  cat > "$repo/CHANGELOG.md" <<'EOF'
# Changelog

## [Unreleased]

### Added
- Initial release preflight fixture
EOF

  cat > "$repo/.env.example" <<'EOF'
API_URL=https://example.com
API_KEY=
EOF

  cat > "$repo/.env" <<'EOF'
API_URL=https://example.com
API_KEY=secret
EOF

  cat > "$repo/package.json" <<'EOF'
{
  "name": "release-preflight-fixture",
  "private": true,
  "scripts": {
    "healthcheck": "node -e \"process.exit(0)\""
  }
}
EOF

  cat > "$repo/scripts/app.sh" <<'EOF'
#!/bin/bash
echo "ready"
EOF

  cat > "$repo/scripts/release-preflight.sh" <<'EOF'
#!/bin/bash
local residual_patterns="${HARNESS_RELEASE_RESIDUAL_PATTERNS:-mockData|dummy|fakeData|localhost|TODO|FIXME}"
EOF

  git -C "$repo" add .
  git -C "$repo" commit -qm "initial"
}

test_skill_mentions_preflight() {
  assert_contains "$PROJECT_ROOT/skills/harness-release/SKILL.md" "release-preflight.sh"
  assert_contains "$PROJECT_ROOT/skills/harness-release/SKILL.md" "HARNESS_RELEASE_HEALTHCHECK_CMD"
  assert_contains "$PROJECT_ROOT/skills/harness-release/SKILL.md" "dry-run"
}

test_doc_mentions_overrides() {
  assert_contains "$PROJECT_ROOT/docs/release-preflight.md" "HARNESS_RELEASE_PROJECT_ROOT"
  assert_contains "$PROJECT_ROOT/docs/release-preflight.md" "HARNESS_RELEASE_CI_STATUS_CMD"
}

test_preflight_pass_and_fail() {
  local repo="$TMP_DIR/release-preflight-repo"
  setup_repo "$repo"

  local success_output="$TMP_DIR/success.txt"
  HARNESS_RELEASE_PROJECT_ROOT="$repo" \
  HARNESS_RELEASE_HEALTHCHECK_CMD='true' \
  HARNESS_RELEASE_CI_STATUS_CMD='true' \
    "$PROJECT_ROOT/scripts/release-preflight.sh" >"$success_output"

  assert_contains "$success_output" "\\[PASS\\] working tree clean"
  assert_contains "$success_output" "\\[PASS\\] CHANGELOG.md has \\[Unreleased\\]"
  assert_contains "$success_output" "\\[PASS\\] .env matches .env.example"
  assert_contains "$success_output" "\\[PASS\\] healthcheck command"
  assert_contains "$success_output" "\\[PASS\\] runtime residual scan"
  assert_contains "$success_output" "\\[PASS\\] CI status"
  assert_contains "$success_output" "Summary: "

  printf 'BROKEN\n' >> "$repo/scripts/app.sh"
  local failure_output="$TMP_DIR/failure.txt"
  if HARNESS_RELEASE_PROJECT_ROOT="$repo" \
    HARNESS_RELEASE_HEALTHCHECK_CMD='true' \
    HARNESS_RELEASE_CI_STATUS_CMD='true' \
      "$PROJECT_ROOT/scripts/release-preflight.sh" >"$failure_output" 2>&1; then
    fail "preflight should fail on dirty tree"
  fi

  assert_contains "$failure_output" "\\[FAIL\\] working tree clean"
}

test_preflight_warns_when_env_is_managed_elsewhere() {
  local repo="$TMP_DIR/release-preflight-managed-secrets"
  setup_repo "$repo"

  rm -f "$repo/.env"
  git -C "$repo" add -u
  git -C "$repo" commit -qm "remove local env"

  local output="$TMP_DIR/managed-secrets.txt"
  HARNESS_RELEASE_PROJECT_ROOT="$repo" \
  HARNESS_RELEASE_HEALTHCHECK_CMD='true' \
  HARNESS_RELEASE_CI_STATUS_CMD='true' \
    "$PROJECT_ROOT/scripts/release-preflight.sh" >"$output"

  assert_contains "$output" "\\[WARN\\] .env missing for .env.example"
  assert_contains "$output" "\\[PASS\\] healthcheck command"
  assert_contains "$output" "\\[PASS\\] CI status"
}

test_skill_mentions_preflight
test_doc_mentions_overrides
test_preflight_pass_and_fail
test_preflight_warns_when_env_is_managed_elsewhere

echo "test-release-preflight: ok"
