#!/bin/bash
# sync-v3-skill-mirrors.sh
# Public v3 skill mirrors for Claude/Codex/OpenCode.
#
# Why:
#   Windows checkout with core.symlinks=false turns repository symlinks into
#   plain text files. Claude Code ignores those files when building the slash
#   command list, so the harness-* entry skills disappear before SessionStart
#   repair hooks can run.
#
# This script keeps the public harness-* skills as real directories in:
#   - skills/
#   - codex/.codex/skills/
#   - opencode/skills/
#
# Usage:
#   ./scripts/sync-v3-skill-mirrors.sh          # overwrite mirrors from skills
#   ./scripts/sync-v3-skill-mirrors.sh --check  # verify mirrors match skills

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PLUGIN_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

V3_SKILLS=(
  "harness-plan"
  "harness-sync"
  "harness-work"
  "harness-review"
  "harness-release"
  "harness-setup"
)

ALIAS_SKILLS=(
  "breezing"
)

MIRROR_ROOTS=(
  "skills"
  "codex/.codex/skills"
  "opencode/skills"
)

MODE="sync"
if [ "${1:-}" = "--check" ]; then
  MODE="check"
elif [ -n "${1:-}" ]; then
  echo "Usage: $0 [--check]" >&2
  exit 2
fi

sync_skill() {
  local skill="$1"
  local mirror_root="$2"
  local src="$PLUGIN_ROOT/skills/$skill"
  local dst_root="$PLUGIN_ROOT/$mirror_root"
  local dst="$dst_root/$skill"

  mkdir -p "$dst_root"
  rm -rf "$dst"
  cp -R "$src" "$dst"
  echo "synced $mirror_root/$skill"
}

check_skill() {
  local skill="$1"
  local mirror_root="$2"
  local src="$PLUGIN_ROOT/skills/$skill"
  local dst="$PLUGIN_ROOT/$mirror_root/$skill"

  if [ ! -d "$dst" ]; then
    echo "missing $mirror_root/$skill" >&2
    return 1
  fi

  if [ -L "$dst" ]; then
    echo "symlink $mirror_root/$skill" >&2
    return 1
  fi

  if ! diff -qr "$src" "$dst" >/dev/null; then
    echo "drift $mirror_root/$skill" >&2
    return 1
  fi

  echo "ok $mirror_root/$skill"
}

sync_alias_skill() {
  local skill="$1"
  local mirror_root="$2"
  local src="$PLUGIN_ROOT/skills/$skill"
  local dst_root="$PLUGIN_ROOT/$mirror_root"
  local dst="$dst_root/$skill"

  mkdir -p "$dst_root"
  rm -rf "$dst"
  cp -R "$src" "$dst"
  echo "synced $mirror_root/$skill"
}

check_alias_skill() {
  local skill="$1"
  local mirror_root="$2"
  local src="$PLUGIN_ROOT/skills/$skill"
  local dst="$PLUGIN_ROOT/$mirror_root/$skill"

  if [ ! -d "$dst" ]; then
    echo "missing $mirror_root/$skill" >&2
    return 1
  fi

  if [ -L "$dst" ]; then
    echo "symlink $mirror_root/$skill" >&2
    return 1
  fi

  if diff -qr "$src" "$dst" >/dev/null 2>&1; then
    echo "ok $mirror_root/$skill"
    return 0
  fi

  echo "drift $mirror_root/$skill" >&2
  return 1
}

FAILURES=0
for mirror_root in "${MIRROR_ROOTS[@]}"; do
  for skill in "${V3_SKILLS[@]}"; do
    if [ "$MODE" = "sync" ]; then
      sync_skill "$skill" "$mirror_root"
    else
      if ! check_skill "$skill" "$mirror_root"; then
        FAILURES=$((FAILURES + 1))
      fi
    fi
  done
done

ALIAS_MIRROR_ROOTS=(
  "skills"
  "codex/.codex/skills"
)

for skill in "${ALIAS_SKILLS[@]}"; do
  for mirror_root in "${ALIAS_MIRROR_ROOTS[@]}"; do
    if [ "$MODE" = "sync" ]; then
      sync_alias_skill "$skill" "$mirror_root"
    else
      check_alias_skill "$skill" "$mirror_root"
    fi
  done
done

if [ "$MODE" = "check" ] && [ "$FAILURES" -gt 0 ]; then
  exit 1
fi
