#!/bin/bash
# session-cleanup.sh
# SessionEnd hook for cleaning up temporary files
#
# Usage: ./scripts/session-cleanup.sh
# Runs on session complete termination (not on every response)

set -euo pipefail

# Find repo root
REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
STATE_DIR="${REPO_ROOT}/.claude/state"

# Skip if state directory doesn't exist
if [[ ! -d "$STATE_DIR" ]]; then
  echo '{"continue": true, "message": "No state directory"}'
  exit 0
fi

# Security: Verify state directory is not a symlink
if [[ -L "$STATE_DIR" ]]; then
  echo '{"continue": true, "message": "State directory is symlink, skipping"}'
  exit 0
fi

# Clean up temporary files
cleanup_files=(
  "pending-skill.json"
  "current-operation.json"
)

for filename in "${cleanup_files[@]}"; do
  file="${STATE_DIR}/${filename}"
  if [[ -f "$file" && ! -L "$file" ]]; then
    rm -f "$file" 2>/dev/null || true
  fi
done

# Clean up inbox temp files (glob pattern)
for file in "${STATE_DIR}"/inbox-*.tmp; do
  if [[ -f "$file" && ! -L "$file" ]]; then
    rm -f "$file" 2>/dev/null || true
  fi
done

# Output for hook feedback
echo '{"continue": true, "message": "Session cleanup completed"}'
