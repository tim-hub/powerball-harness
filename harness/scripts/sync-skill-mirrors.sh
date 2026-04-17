#!/bin/bash
# sync-skill-mirrors.sh
# Sync skills from harness/skills/ (SSOT) to Codex/OpenCode mirrors.
#
# Why:
#   Windows checkout with core.symlinks=false turns repository symlinks into
#   plain text files. Claude Code ignores those files when building the slash
#   command list, so the harness-* entry skills disappear before SessionStart
#   repair hooks can run.
#
#   This script keeps skills as real directories (not symlinks) in:
#     - harness/templates/codex-skills/
#     - harness/templates/opencode/skills/
#
# Source of truth:
#   - harness/skills/ for all shared skills
#
# Sync scope:
#   - All skill directories that exist in BOTH the SSOT and a mirror root
#   - New skills added only to harness/skills/ are NOT auto-propagated (add manually)
#   - routing-rules.md is synced if present in both
#
# Usage:
#   ./harness/scripts/sync-skill-mirrors.sh          # overwrite mirrors from harness/skills/
#   ./harness/scripts/sync-skill-mirrors.sh --check  # verify mirrors match harness/skills/

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"  # skill-local: this scripts/ directory
PLUGIN_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"               # plugin-local: harness plugin root

SHARED_SSOT_DIR="${PLUGIN_ROOT}/skills"

# Mirror roots (relative to PLUGIN_ROOT)
MIRROR_ROOTS=(
  "templates/codex-skills"
  "templates/opencode/skills"
)

MODE="sync"
if [ "${1:-}" = "--check" ]; then
  MODE="check"
elif [ -n "${1:-}" ]; then
  echo "Usage: harness/scripts/sync-skill-mirrors.sh [--check]" >&2
  exit 2
fi

resolve_src_dir() {
  local skill="$1"
  printf '%s\n' "${SHARED_SSOT_DIR}/${skill}"
}

sync_skill() {
  local skill="$1"
  local mirror_root="$2"
  local src
  src="$(resolve_src_dir "$skill")"
  local dst_root="${PLUGIN_ROOT}/${mirror_root}"
  local dst="${dst_root}/${skill}"

  mkdir -p "$dst_root"
  rm -rf "$dst"
  cp -R "$src" "$dst"
  echo "synced ${mirror_root}/${skill}"
}

check_skill() {
  local skill="$1"
  local mirror_root="$2"
  local src
  src="$(resolve_src_dir "$skill")"
  local dst="${PLUGIN_ROOT}/${mirror_root}/${skill}"

  if [ ! -d "$dst" ]; then
    echo "missing ${mirror_root}/${skill}" >&2
    return 1
  fi

  if [ -L "$dst" ]; then
    echo "symlink ${mirror_root}/${skill}" >&2
    return 1
  fi

  if ! diff -qr --exclude='.DS_Store' --exclude='.claude' "$src" "$dst" >/dev/null; then
    echo "drift ${mirror_root}/${skill}" >&2
    return 1
  fi

  echo "ok ${mirror_root}/${skill}"
}

FAILURES=0
for mirror_root in "${MIRROR_ROOTS[@]}"; do
  mirror_dir="${PLUGIN_ROOT}/${mirror_root}"
  [ -d "$mirror_dir" ] || continue

  # Discovery: sync every skill directory that exists in both SSOT and mirror
  for entry in "${mirror_dir}"/*/; do
    [ -d "$entry" ] || continue
    skill="$(basename "$entry")"

    # Skip non-skill entries
    [ "$skill" = "node_modules" ] && continue
    [ "$skill" = ".git" ] && continue

    # Only sync if the SSOT has this skill
    if [ ! -d "$(resolve_src_dir "$skill")" ]; then
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
  if [ -f "${SHARED_SSOT_DIR}/routing-rules.md" ] && [ -f "${mirror_dir}/routing-rules.md" ]; then
    if [ "$MODE" = "sync" ]; then
      cp "${SHARED_SSOT_DIR}/routing-rules.md" "${mirror_dir}/routing-rules.md"
      echo "synced ${mirror_root}/routing-rules.md"
    else
      if ! diff -q "${SHARED_SSOT_DIR}/routing-rules.md" "${mirror_dir}/routing-rules.md" >/dev/null; then
        echo "drift ${mirror_root}/routing-rules.md" >&2
        FAILURES=$((FAILURES + 1))
      else
        echo "ok ${mirror_root}/routing-rules.md"
      fi
    fi
  fi
done

if [ "$MODE" = "check" ] && [ "$FAILURES" -gt 0 ]; then
  exit 1
fi
