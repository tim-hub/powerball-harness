#!/usr/bin/env bash
# merge-gitignore.sh — Merge Harness-managed block into a project's .gitignore
#
# Usage:
#   bash merge-gitignore.sh [target-gitignore-path]
#
# Defaults to .gitignore in the current directory when no argument is given.
# Idempotent: skips if the marker block is already present.
#
# The managed block is delimited by:
#   # ---- Harness managed begin ----
#   # ---- Harness managed end ----

set -euo pipefail

TARGET="${1:-.gitignore}"
BEGIN_MARKER="# ---- Harness managed begin ----"
END_MARKER="# ---- Harness managed end ----"

if [ -f "$TARGET" ] && grep -qF "$BEGIN_MARKER" "$TARGET"; then
  echo "harness-managed block already present in $TARGET — skipping"
  exit 0
fi

cat >> "$TARGET" <<'BLOCK'

# ---- Harness managed begin ----
# Claude Code / Harness runtime files
.claude/sessions/
.claude/logs/
.claude/state/
.claude/worktrees/
.claude/settings.local.json


# Force-track Harness configuration (do not ignore these)
!.claude/memory/
!.claude/output-styles/
!.claude/rules/
!.claude/scripts/
!.claude/skills/
!.claude/settings.json
# ---- Harness managed end ----
BLOCK

echo "harness-managed block appended to $TARGET"
