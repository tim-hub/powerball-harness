#!/bin/bash
#
# opencode-setup-local.sh
#
# Copy opencode templates from the installed Harness plugin.
#
# Usage:
#   ./scripts/opencode-setup-local.sh
#
set -euo pipefail
IFS=$'\n\t'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(pwd)"

fail() {
  echo "Error: $1" >&2
  exit 1
}

pick_latest_version_dir() {
  local base_dir="$1"
  if [ ! -d "$base_dir" ]; then
    return 1
  fi

  local latest
  latest="$(ls -1 "$base_dir" 2>/dev/null | sort -V | tail -n 1)"
  if [ -z "$latest" ]; then
    return 1
  fi
  echo "$base_dir/$latest"
}

resolve_plugin_dir() {
  local repo_root
  repo_root="$(cd "$SCRIPT_DIR/.." && pwd)"

  local marketplace_dir="$HOME/.claude/plugins/marketplaces/claude-code-harness-marketplace"
  local cache_root="$HOME/.claude/plugins/cache/claude-code-harness-marketplace/claude-code-harness"
  local cache_dir
  cache_dir="$(pick_latest_version_dir "$cache_root" || true)"

  local candidates=(
    "${CLAUDE_PLUGIN_ROOT:-}"
    "$repo_root"
    "$marketplace_dir"
    "$cache_dir"
  )

  local fallback=""
  local candidate
  for candidate in "${candidates[@]}"; do
    [ -n "$candidate" ] || continue
    if [ -d "$candidate/opencode/commands" ]; then
      [ -z "$fallback" ] && fallback="$candidate"
      if [ -d "$candidate/opencode/skills" ]; then
        echo "$candidate"
        return 0
      fi
    fi
  done

  if [ -n "$fallback" ]; then
    echo "$fallback"
    return 0
  fi

  return 1
}

PLUGIN_DIR="$(resolve_plugin_dir || true)"
if [ -z "$PLUGIN_DIR" ]; then
  fail "Harness plugin directory not found. Set CLAUDE_PLUGIN_ROOT or install the plugin."
fi

echo "Using Harness plugin: $PLUGIN_DIR"

mkdir -p "$PROJECT_DIR/.opencode/commands/core"
mkdir -p "$PROJECT_DIR/.opencode/commands/optional"
mkdir -p "$PROJECT_DIR/.opencode/commands/pm"
mkdir -p "$PROJECT_DIR/.opencode/commands/handoff"
mkdir -p "$PROJECT_DIR/.claude/skills"

if [ -d "$PROJECT_DIR/.claude/skills" ] && [ "$(ls -A "$PROJECT_DIR/.claude/skills" 2>/dev/null)" ]; then
  backup_dir="$PROJECT_DIR/.claude/skills.backup.$(date +%Y%m%d%H%M%S)"
  mv "$PROJECT_DIR/.claude/skills" "$backup_dir"
  mkdir -p "$PROJECT_DIR/.claude/skills"
  echo "Backed up existing .claude/skills to: $backup_dir"
fi

cp -r "$PLUGIN_DIR/opencode/commands/"* "$PROJECT_DIR/.opencode/commands/"
if [ -d "$PLUGIN_DIR/opencode/skills" ]; then
  cp -r "$PLUGIN_DIR/opencode/skills/"* "$PROJECT_DIR/.claude/skills/"
else
  echo "Warning: opencode/skills not found in plugin source."
fi

if [ -f "$PROJECT_DIR/AGENTS.md" ]; then
  backup_agents="$PROJECT_DIR/AGENTS.md.backup.$(date +%Y%m%d%H%M%S)"
  mv "$PROJECT_DIR/AGENTS.md" "$backup_agents"
  echo "Backed up existing AGENTS.md to: $backup_agents"
fi

if [ -f "$PLUGIN_DIR/opencode/AGENTS.md" ]; then
  cp "$PLUGIN_DIR/opencode/AGENTS.md" "$PROJECT_DIR/AGENTS.md"
fi

echo "Copied opencode commands, skills, and AGENTS.md."
