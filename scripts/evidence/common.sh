#!/bin/bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
HARNESS_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
OUTPUT_ROOT="${HARNESS_EVIDENCE_OUT_DIR:-$HARNESS_ROOT/out/evidence/work-all}"

resolve_timeout_cmd() {
  if command -v timeout >/dev/null 2>&1; then
    echo "timeout"
  elif command -v gtimeout >/dev/null 2>&1; then
    echo "gtimeout"
  else
    echo ""
  fi
}

prepare_case_dir() {
  local case_name="$1"
  local case_root="$OUTPUT_ROOT/$case_name"
  local timestamp
  timestamp="$(date +%Y%m%d-%H%M%S)"
  local case_dir

  mkdir -p "$case_root"
  if case_dir="$(mktemp -d "$case_root/$timestamp.XXXXXX" 2>/dev/null)"; then
    :
  else
    case_dir="$case_root/$timestamp.$$"
    mkdir -p "$case_dir"
  fi

  echo "$case_dir"
}

copy_fixture_to_case() {
  local fixture_name="$1"
  local case_dir="$2"
  local fixture_src="$HARNESS_ROOT/tests/fixtures/$fixture_name"
  local worktree_dir="$case_dir/worktree"

  mkdir -p "$worktree_dir"
  cp -R "$fixture_src"/. "$worktree_dir"/
  echo "$worktree_dir"
}

overlay_tree() {
  local src="$1"
  local dest="$2"

  if [ -d "$src" ]; then
    if command -v rsync >/dev/null 2>&1; then
      rsync -a "$src/" "$dest/"
    else
      cp -R "$src"/. "$dest"/
    fi
  fi
}

init_git_repo() {
  local worktree_dir="$1"
  git -C "$worktree_dir" init -q
  git -C "$worktree_dir" config user.name "Harness Evidence"
  git -C "$worktree_dir" config user.email "harness-evidence@example.com"
  git -C "$worktree_dir" add .
  git -C "$worktree_dir" commit -qm "chore: baseline fixture"
}

write_command_preview() {
  local case_dir="$1"
  local worktree_dir="$2"
  cat > "$case_dir/command-preview.txt" <<EOF
cd "$worktree_dir"
claude --plugin-dir "$HARNESS_ROOT" --dangerously-skip-permissions --output-format json --no-session-persistence -p "\$(cat PROMPT.md)"
EOF
}

run_baseline_tests() {
  local worktree_dir="$1"
  local log_file="$2"
  set +e
  (cd "$worktree_dir" && npm test) >"$log_file" 2>&1
  local status=$?
  set -e
  return "$status"
}

run_claude_full() {
  local worktree_dir="$1"
  local stdout_file="$2"
  local stderr_file="$3"
  local elapsed_file="$4"
  local timeout_cmd
  timeout_cmd="$(resolve_timeout_cmd)"

  if ! command -v claude >/dev/null 2>&1; then
    echo "claude CLI is not installed" >"$stderr_file"
    return 127
  fi

  local start_ts end_ts duration status
  start_ts="$(date +%s)"

  set +e
  if [ -n "$timeout_cmd" ]; then
    (cd "$worktree_dir" && "$timeout_cmd" 900 claude --plugin-dir "$HARNESS_ROOT" --dangerously-skip-permissions --output-format json --no-session-persistence -p "$(cat PROMPT.md)") >"$stdout_file" 2>"$stderr_file"
    status=$?
  else
    (cd "$worktree_dir" && claude --plugin-dir "$HARNESS_ROOT" --dangerously-skip-permissions --output-format json --no-session-persistence -p "$(cat PROMPT.md)") >"$stdout_file" 2>"$stderr_file"
    status=$?
  fi
  set -e

  end_ts="$(date +%s)"
  duration="$((end_ts - start_ts))"
  printf '%s\n' "$duration" >"$elapsed_file"

  return "$status"
}

claude_hit_rate_limit() {
  local stderr_file="$1"
  local stdout_file="${2:-}"
  local pattern='hit your limit|rate limit|usage limit|quota'

  if [ -f "$stderr_file" ] && grep -Eqi "$pattern" "$stderr_file"; then
    return 0
  fi

  if [ -n "$stdout_file" ] && [ -f "$stdout_file" ] && grep -Eqi "$pattern" "$stdout_file"; then
    return 0
  fi

  return 1
}

apply_success_replay() {
  local worktree_dir="$1"
  local baseline_commit="$2"
  local replay_log="$3"
  local overlay_dir="$worktree_dir/.evidence-replay/success"
  local commit_message_file="$worktree_dir/.evidence-replay/COMMIT_MESSAGE.txt"
  local commit_message="test: replay success evidence artifact"

  if [ ! -d "$overlay_dir" ]; then
    echo "success replay overlay not found: $overlay_dir" >"$replay_log"
    return 1
  fi

  if [ -f "$commit_message_file" ]; then
    commit_message="$(cat "$commit_message_file")"
  fi

  git -C "$worktree_dir" reset --hard -q "$baseline_commit"
  git -C "$worktree_dir" clean -fdq
  overlay_tree "$overlay_dir" "$worktree_dir"

  git -C "$worktree_dir" add Plans.md src/math.js src/format.js
  git -C "$worktree_dir" commit -qm "$commit_message"

  {
    echo "replay_applied=yes"
    echo "baseline_commit=$baseline_commit"
    echo "overlay_dir=$overlay_dir"
    echo "commit_message=$commit_message"
  } >"$replay_log"
}

collect_git_artifacts() {
  local worktree_dir="$1"
  local case_dir="$2"

  git -C "$worktree_dir" status --short >"$case_dir/git-status.txt"
  git -C "$worktree_dir" diff --stat >"$case_dir/git-diff-stat.txt" || true
  git -C "$worktree_dir" diff >"$case_dir/git-diff.patch" || true
  git -C "$worktree_dir" log --oneline --decorate -5 >"$case_dir/git-log.txt"
  git -C "$worktree_dir" rev-list --count HEAD >"$case_dir/commit-count.txt"
}
