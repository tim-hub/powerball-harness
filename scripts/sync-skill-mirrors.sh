#!/bin/bash
# sync-skill-mirrors.sh
# Sync skills from skills/ (SSOT) to Codex/OpenCode mirrors.
#
# Why:
#   Windows checkout with core.symlinks=false turns repository symlinks into
#   plain text files. Claude Code ignores those files when building the slash
#   command list, so the harness-* entry skills disappear before SessionStart
#   repair hooks can run.
#
# This script keeps skills as real directories in:
#   - codex/.codex/skills/
#   - opencode/skills/
#
# Source of truth:
#   - skills/ for shared skills
#   - skills-codex/ for Codex-only variants
#
# Sync scope:
#   - All skill directories that exist in BOTH the resolved SSOT and a mirror root
#   - New skills added only to skills/ are NOT auto-propagated (add manually)
#   - routing-rules.md is synced if present in both
#
# Usage:
#   ./scripts/sync-skill-mirrors.sh          # overwrite mirrors from skills/
#   ./scripts/sync-skill-mirrors.sh --check  # verify mirrors match skills/

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PLUGIN_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

SHARED_SSOT_DIR="$PLUGIN_ROOT/skills"
CODEX_SSOT_DIR="$PLUGIN_ROOT/skills-codex"

# Mirror roots (codex and opencode only — skills/ is the SSOT, not a mirror)
MIRROR_ROOTS=(
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
  local src
  src="$(resolve_src_dir "$skill" "$mirror_root")"
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
  local src
  src="$(resolve_src_dir "$skill" "$mirror_root")"
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

resolve_src_dir() {
  local skill="$1"
  local mirror_root="$2"

  if [ "$mirror_root" = "codex/.codex/skills" ] && [ -d "$CODEX_SSOT_DIR/$skill" ]; then
    printf '%s\n' "$CODEX_SSOT_DIR/$skill"
    return 0
  fi

  printf '%s\n' "$SHARED_SSOT_DIR/$skill"
}

FAILURES=0
for mirror_root in "${MIRROR_ROOTS[@]}"; do
  mirror_dir="$PLUGIN_ROOT/$mirror_root"
  [ -d "$mirror_dir" ] || continue

  # Discovery: sync every skill directory that exists in both SSOT and mirror
  for entry in "$mirror_dir"/*/; do
    [ -d "$entry" ] || continue
    skill="$(basename "$entry")"

    # Skip non-skill entries
    [ "$skill" = "node_modules" ] && continue
    [ "$skill" = ".git" ] && continue

    # Only sync if the resolved SSOT has this skill
    if [ ! -d "$(resolve_src_dir "$skill" "$mirror_root")" ]; then
      continue
    fi

    if [ "$MODE" = "sync" ]; then
      sync_skill "$skill" "$mirror_root"
    else
      if ! check_skill "$skill" "$mirror_root"; then
        FAILURES=$((FAILURES + 1))
      fi
    fi
  done

  # Also sync routing-rules.md if present in both
  if [ -f "$SHARED_SSOT_DIR/routing-rules.md" ] && [ -f "$mirror_dir/routing-rules.md" ]; then
    if [ "$MODE" = "sync" ]; then
      cp "$SHARED_SSOT_DIR/routing-rules.md" "$mirror_dir/routing-rules.md"
      echo "synced $mirror_root/routing-rules.md"
    else
      if ! diff -q "$SHARED_SSOT_DIR/routing-rules.md" "$mirror_dir/routing-rules.md" >/dev/null; then
        echo "drift $mirror_root/routing-rules.md" >&2
        FAILURES=$((FAILURES + 1))
      else
        echo "ok $mirror_root/routing-rules.md"
      fi
    fi
  fi
done

if [ "$MODE" = "check" ] && [ "$FAILURES" -gt 0 ]; then
  exit 1
fi
