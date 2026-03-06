#!/bin/bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=scripts/evidence/common.sh
source "$SCRIPT_DIR/common.sh"

usage() {
  cat <<'EOF'
Usage:
  scripts/evidence/run-work-all-success.sh [--smoke|--full] [--strict-live]

Options:
  --smoke       Validate fixture shape only (default)
  --full        Run live Claude execution first, then replay the expected success overlay if Claude is rate-limited
  --strict-live Fail instead of replaying if Claude cannot complete the live run
EOF
}

MODE="--smoke"
STRICT_LIVE=0

while [ $# -gt 0 ]; do
  case "$1" in
    --smoke|--full)
      MODE="$1"
      shift
      ;;
    --strict-live)
      STRICT_LIVE=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown mode: $1" >&2
      usage >&2
      exit 1
      ;;
  esac
done

CASE_DIR="$(prepare_case_dir success)"
WORKTREE_DIR="$(copy_fixture_to_case work-all-success "$CASE_DIR")"

init_git_repo "$WORKTREE_DIR"
write_command_preview "$CASE_DIR" "$WORKTREE_DIR"
BASELINE_COMMIT="$(git -C "$WORKTREE_DIR" rev-parse HEAD)"

if run_baseline_tests "$WORKTREE_DIR" "$CASE_DIR/baseline-test.log"; then
  echo "Expected baseline tests to fail for success fixture" >&2
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
EXECUTION_MODE="live"
FALLBACK_REASON="none"
RATE_LIMIT_DETECTED="no"
set +e
run_claude_full "$WORKTREE_DIR" "$CASE_DIR/claude-stdout.json" "$CASE_DIR/claude-stderr.log" "$CASE_DIR/elapsed-seconds.txt"
CLAUDE_STATUS=$?
set -e

if claude_hit_rate_limit "$CASE_DIR/claude-stderr.log" "$CASE_DIR/claude-stdout.json"; then
  RATE_LIMIT_DETECTED="yes"
fi

if [ "$RATE_LIMIT_DETECTED" = "yes" ] && [ "$STRICT_LIVE" -eq 0 ]; then
  if apply_success_replay "$WORKTREE_DIR" "$BASELINE_COMMIT" "$CASE_DIR/replay.log"; then
    EXECUTION_MODE="replay-after-rate-limit"
    FALLBACK_REASON="claude-rate-limit"
  fi
fi

printf '%s\n' "$EXECUTION_MODE" >"$CASE_DIR/execution-mode.txt"
printf '%s\n' "$FALLBACK_REASON" >"$CASE_DIR/fallback-reason.txt"
printf '%s\n' "$RATE_LIMIT_DETECTED" >"$CASE_DIR/rate-limit-detected.txt"

collect_git_artifacts "$WORKTREE_DIR" "$CASE_DIR"

set +e
(cd "$WORKTREE_DIR" && npm test) >"$CASE_DIR/post-test.log" 2>&1
POST_STATUS=$?
set -e

BASELINE_COMMITS=1
FINAL_COMMITS="$(cat "$CASE_DIR/commit-count.txt")"

cat > "$CASE_DIR/result.txt" <<EOF
scenario=success
claude_status=$CLAUDE_STATUS
execution_mode=$EXECUTION_MODE
fallback_reason=$FALLBACK_REASON
rate_limit_detected=$RATE_LIMIT_DETECTED
post_test_status=$POST_STATUS
baseline_commits=$BASELINE_COMMITS
final_commits=$FINAL_COMMITS
EOF

if [ "$CLAUDE_STATUS" -ne 0 ] && [ "$EXECUTION_MODE" = "live" ]; then
  echo "Claude execution failed for success scenario. See $CASE_DIR" >&2
  exit 1
fi

if [ "$RATE_LIMIT_DETECTED" = "yes" ] && [ "$STRICT_LIVE" -eq 1 ]; then
  echo "Claude execution hit a rate limit in strict-live mode. See $CASE_DIR" >&2
  exit 1
fi

if [ "$POST_STATUS" -ne 0 ]; then
  echo "Success scenario still fails tests. See $CASE_DIR" >&2
  exit 1
fi

if [ "$FINAL_COMMITS" -le "$BASELINE_COMMITS" ]; then
  echo "Success scenario did not produce a follow-up commit. See $CASE_DIR" >&2
  exit 1
fi

if [ "$EXECUTION_MODE" = "replay-after-rate-limit" ]; then
  echo "Full evidence passed via replay fallback: $CASE_DIR"
else
  echo "Full evidence passed: $CASE_DIR"
fi
