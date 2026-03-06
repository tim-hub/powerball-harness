#!/bin/bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=scripts/evidence/common.sh
source "$SCRIPT_DIR/common.sh"

MODE="${1:---smoke}"
CASE_DIR="$(prepare_case_dir failure)"
WORKTREE_DIR="$(copy_fixture_to_case work-all-failure "$CASE_DIR")"

init_git_repo "$WORKTREE_DIR"
write_command_preview "$CASE_DIR" "$WORKTREE_DIR"

if run_baseline_tests "$WORKTREE_DIR" "$CASE_DIR/baseline-test.log"; then
  echo "Expected baseline tests to fail for failure fixture" >&2
  exit 1
fi

if [ "$MODE" = "--smoke" ]; then
  echo "smoke-ok" >"$CASE_DIR/result.txt"
  echo "Smoke validation passed: $CASE_DIR"
  exit 0
fi

if [ "$MODE" != "--full" ]; then
  echo "Unknown mode: $MODE" >&2
  exit 1
fi

CLAUDE_STATUS=0
set +e
run_claude_full "$WORKTREE_DIR" "$CASE_DIR/claude-stdout.json" "$CASE_DIR/claude-stderr.log" "$CASE_DIR/elapsed-seconds.txt"
CLAUDE_STATUS=$?
set -e

collect_git_artifacts "$WORKTREE_DIR" "$CASE_DIR"

set +e
(cd "$WORKTREE_DIR" && npm test) >"$CASE_DIR/post-test.log" 2>&1
POST_STATUS=$?
set -e

BASELINE_COMMITS=1
FINAL_COMMITS="$(cat "$CASE_DIR/commit-count.txt")"

cat > "$CASE_DIR/result.txt" <<EOF
scenario=failure
claude_status=$CLAUDE_STATUS
post_test_status=$POST_STATUS
baseline_commits=$BASELINE_COMMITS
final_commits=$FINAL_COMMITS
EOF

if [ "$POST_STATUS" -eq 0 ]; then
  echo "Failure scenario unexpectedly passed tests. See $CASE_DIR" >&2
  exit 1
fi

if [ "$FINAL_COMMITS" -gt "$BASELINE_COMMITS" ]; then
  echo "Failure scenario created a commit despite unresolved failure. See $CASE_DIR" >&2
  exit 1
fi

echo "Full failure evidence passed: $CASE_DIR"
