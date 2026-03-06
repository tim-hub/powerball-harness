#!/bin/bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
DEFAULT_OUTPUT_ROOT="$REPO_ROOT/out/evidence/work-all"

usage() {
  cat <<'EOF'
Usage:
  scripts/evidence/run-work-all-case.sh --case <work-all-success|work-all-failure> [--dry-run|--full] [--output-dir PATH]

Modes:
  --dry-run   Materialize fixture, prompt, and artifact layout only. Does not require claude CLI.
  --full      Run baseline tests, execute /harness-work all through Claude CLI, then capture artifacts.

If no mode is specified, --dry-run is used.
EOF
}

log() {
  printf '[work-all-evidence] %s\n' "$*"
}

die() {
  printf '[work-all-evidence] ERROR: %s\n' "$*" >&2
  exit 1
}

copy_tree() {
  local src="$1"
  local dest="$2"

  if command -v rsync >/dev/null 2>&1; then
    rsync -a --delete "$src/" "$dest/"
  else
    rm -rf "$dest"
    mkdir -p "$dest"
    cp -R "$src/." "$dest/"
  fi
}

write_summary() {
  local summary_file="$1"
  local mode="$2"
  local observed_outcome="$3"
  local baseline_exit="$4"
  local final_exit="$5"
  local claude_exit="$6"
  local commit_delta="$7"
  local changed_ok="$8"
  local workspace_dir="$9"
  local artifacts_dir="${10}"
  local expectation="${11}"

  cat >"$summary_file" <<EOF
# Work-All Evidence Summary

- Case: $CASE_NAME
- Mode: $mode
- Expected outcome: $expectation
- Observed outcome: $observed_outcome
- Baseline test exit: $baseline_exit
- Final test exit: $final_exit
- Claude exit: $claude_exit
- New commits: $commit_delta
- Changed files within contract: $changed_ok
- Workspace: $workspace_dir
- Artifacts: $artifacts_dir
EOF
}

compare_allowed_changes() {
  local changed_file_list="$1"
  local allowed_list="$2"
  local violation_file="$3"

  : >"$violation_file"

  while IFS= read -r changed; do
    [ -z "$changed" ] && continue
    local allowed=0
    for expected in $allowed_list; do
      if [ "$changed" = "$expected" ]; then
        allowed=1
        break
      fi
    done
    if [ "$allowed" -eq 0 ]; then
      echo "$changed" >>"$violation_file"
    fi
  done <"$changed_file_list"

  if [ -s "$violation_file" ]; then
    return 1
  fi
  return 0
}

CASE_NAME=""
MODE="dry-run"
OUTPUT_ROOT="$DEFAULT_OUTPUT_ROOT"

while [ $# -gt 0 ]; do
  case "$1" in
    --case)
      CASE_NAME="${2:-}"
      shift 2
      ;;
    --dry-run)
      MODE="dry-run"
      shift
      ;;
    --full)
      MODE="full"
      shift
      ;;
    --output-dir)
      OUTPUT_ROOT="${2:-}"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      die "unknown argument: $1"
      ;;
  esac
done

[ -n "$CASE_NAME" ] || die "--case is required"

CASE_DIR="$REPO_ROOT/tests/fixtures/$CASE_NAME"
[ -d "$CASE_DIR" ] || die "fixture not found: $CASE_DIR"

FIXTURE_ENV="$CASE_DIR/fixture.env"
[ -f "$FIXTURE_ENV" ] || die "fixture metadata missing: $FIXTURE_ENV"

# shellcheck disable=SC1090
source "$FIXTURE_ENV"

RUN_ID="$(date +"%Y%m%d-%H%M%S")"
ARTIFACTS_DIR="$OUTPUT_ROOT/$CASE_NAME/$RUN_ID"
WORKSPACE_DIR="$ARTIFACTS_DIR/workspace"
mkdir -p "$ARTIFACTS_DIR"

copy_tree "$CASE_DIR" "$WORKSPACE_DIR"

PROMPT_PATH="$WORKSPACE_DIR/PROMPT.md"
[ -f "$PROMPT_PATH" ] || die "PROMPT.md missing in workspace"

git -C "$WORKSPACE_DIR" init -q
git -C "$WORKSPACE_DIR" config user.name "Harness Evidence"
git -C "$WORKSPACE_DIR" config user.email "evidence@local.invalid"
git -C "$WORKSPACE_DIR" add .
git -C "$WORKSPACE_DIR" commit -q -m "chore: evidence baseline"
BASELINE_COMMIT="$(git -C "$WORKSPACE_DIR" rev-parse HEAD)"
BASELINE_COMMIT_COUNT="$(git -C "$WORKSPACE_DIR" rev-list --count HEAD)"

cp "$FIXTURE_ENV" "$ARTIFACTS_DIR/fixture.env"
cp "$PROMPT_PATH" "$ARTIFACTS_DIR/prompt.txt"

cat >"$ARTIFACTS_DIR/paths.txt" <<EOF
output_root=$OUTPUT_ROOT
artifacts_dir=$ARTIFACTS_DIR
workspace_dir=$WORKSPACE_DIR
baseline_commit=$BASELINE_COMMIT
EOF

if [ "$MODE" = "dry-run" ]; then
  write_summary \
    "$ARTIFACTS_DIR/summary.md" \
    "$MODE" \
    "fixture-inspection-only" \
    "not-run" \
    "not-run" \
    "not-run" \
    "0" \
    "not-checked" \
    "$WORKSPACE_DIR" \
    "$ARTIFACTS_DIR" \
    "$EXPECTED_OUTCOME"
  log "dry-run completed: $ARTIFACTS_DIR"
  exit 0
fi

command -v claude >/dev/null 2>&1 || die "claude CLI not found. Use --dry-run for fixture inspection."
command -v npm >/dev/null 2>&1 || die "npm not found. Use --dry-run for fixture inspection."

BASELINE_TEST_EXIT=0
FINAL_TEST_EXIT=0
CLAUDE_EXIT=0
OBSERVED_OUTCOME="unknown"
CHANGED_OK="yes"

set +e
(cd "$WORKSPACE_DIR" && npm test) >"$ARTIFACTS_DIR/test-before.log" 2>&1
BASELINE_TEST_EXIT=$?
set -e

if [ "$BASELINE_TEST_EXIT" -ne "$EXPECT_BASELINE_TEST_EXIT" ]; then
  die "baseline test exit mismatch: expected $EXPECT_BASELINE_TEST_EXIT, got $BASELINE_TEST_EXIT"
fi

START_TS="$(date +%s)"
set +e
(cd "$WORKSPACE_DIR" && claude --plugin-dir "$REPO_ROOT" --dangerously-skip-permissions --output-format json --no-session-persistence -p "$(cat "$PROMPT_PATH")") \
  >"$ARTIFACTS_DIR/claude.stdout.log" 2>"$ARTIFACTS_DIR/claude.stderr.log"
CLAUDE_EXIT=$?
set -e
END_TS="$(date +%s)"
ELAPSED_SECONDS=$((END_TS - START_TS))
echo "$ELAPSED_SECONDS" >"$ARTIFACTS_DIR/elapsed-seconds.txt"

set +e
(cd "$WORKSPACE_DIR" && npm test) >"$ARTIFACTS_DIR/test-after.log" 2>&1
FINAL_TEST_EXIT=$?
set -e

FINAL_COMMIT_COUNT="$(git -C "$WORKSPACE_DIR" rev-list --count HEAD)"
COMMIT_DELTA=$((FINAL_COMMIT_COUNT - BASELINE_COMMIT_COUNT))

if [ "$COMMIT_DELTA" -gt 0 ]; then
  git -C "$WORKSPACE_DIR" diff --stat "$BASELINE_COMMIT"..HEAD >"$ARTIFACTS_DIR/git-diff-stat.txt"
  git -C "$WORKSPACE_DIR" diff "$BASELINE_COMMIT"..HEAD >"$ARTIFACTS_DIR/git-diff.patch"
  git -C "$WORKSPACE_DIR" diff --name-only "$BASELINE_COMMIT"..HEAD >"$ARTIFACTS_DIR/changed-files.txt"
else
  git -C "$WORKSPACE_DIR" diff --stat "$BASELINE_COMMIT" >"$ARTIFACTS_DIR/git-diff-stat.txt" || true
  git -C "$WORKSPACE_DIR" diff "$BASELINE_COMMIT" >"$ARTIFACTS_DIR/git-diff.patch" || true
  git -C "$WORKSPACE_DIR" diff --name-only "$BASELINE_COMMIT" >"$ARTIFACTS_DIR/changed-files.txt" || true
fi

git -C "$WORKSPACE_DIR" status --short >"$ARTIFACTS_DIR/git-status.txt"
git -C "$WORKSPACE_DIR" log --oneline --decorate -5 >"$ARTIFACTS_DIR/git-log.txt"

if ! compare_allowed_changes "$ARTIFACTS_DIR/changed-files.txt" "$ALLOWED_CHANGED_FILES" "$ARTIFACTS_DIR/contract-violations.txt"; then
  CHANGED_OK="no"
fi

if [ "$EXPECTED_OUTCOME" = "success" ] && [ "$FINAL_TEST_EXIT" -eq 0 ] && [ "$CHANGED_OK" = "yes" ]; then
  OBSERVED_OUTCOME="success"
elif [ "$EXPECTED_OUTCOME" = "failure" ] && [ "$FINAL_TEST_EXIT" -ne 0 ] && [ "$COMMIT_DELTA" -eq 0 ] && [ "$CHANGED_OK" = "yes" ]; then
  OBSERVED_OUTCOME="failure"
else
  OBSERVED_OUTCOME="mismatch"
fi

write_summary \
  "$ARTIFACTS_DIR/summary.md" \
  "$MODE" \
  "$OBSERVED_OUTCOME" \
  "$BASELINE_TEST_EXIT" \
  "$FINAL_TEST_EXIT" \
  "$CLAUDE_EXIT" \
  "$COMMIT_DELTA" \
  "$CHANGED_OK" \
  "$WORKSPACE_DIR" \
  "$ARTIFACTS_DIR" \
  "$EXPECTED_OUTCOME"

if [ "$OBSERVED_OUTCOME" = "mismatch" ]; then
  die "observed outcome did not match expectation. See $ARTIFACTS_DIR/summary.md"
fi

log "full run completed: $ARTIFACTS_DIR"
