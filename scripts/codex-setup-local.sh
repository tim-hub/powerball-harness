#!/bin/bash
#
# codex-setup-local.sh
#
# Copy Codex CLI templates from the installed Harness plugin.
#
# Usage:
#   ./scripts/codex-setup-local.sh [--with-mcp|--skip-mcp]
#
set -euo pipefail
IFS=$'\n\t'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(pwd)"
WITH_MCP="auto"

while [ $# -gt 0 ]; do
  case "$1" in
    --with-mcp)
      WITH_MCP=true
      ;;
    --skip-mcp)
      WITH_MCP=false
      ;;
    *)
      echo "Unknown argument: $1" >&2
      exit 1
      ;;
  esac
  shift
done

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

  local candidate
  for candidate in "${candidates[@]}"; do
    [ -n "$candidate" ] || continue
    if [ -d "$candidate/codex/.codex/skills" ]; then
      echo "$candidate"
      return 0
    fi
  done

  return 1
}

backup_dir() {
  local target="$1"
  if [ -d "$target" ] && [ "$(ls -A "$target" 2>/dev/null)" ]; then
    local ts
    ts=$(date +%Y%m%d%H%M%S)
    mv "$target" "${target}.backup.${ts}"
    echo "Backed up $target to ${target}.backup.${ts}"
  fi
}

PLUGIN_DIR="$(resolve_plugin_dir || true)"
if [ -z "$PLUGIN_DIR" ]; then
  fail "Harness plugin directory not found. Set CLAUDE_PLUGIN_ROOT or install the plugin."
fi

echo "Using Harness plugin: $PLUGIN_DIR"

mkdir -p "$PROJECT_DIR/.codex"

backup_dir "$PROJECT_DIR/.codex/skills"
backup_dir "$PROJECT_DIR/.codex/rules"

if [ -d "$PLUGIN_DIR/codex/.codex/skills" ]; then
  cp -r "$PLUGIN_DIR/codex/.codex/skills" "$PROJECT_DIR/.codex/"
  echo "Skills copied to .codex/skills"
else
  fail "codex/.codex/skills not found in plugin source"
fi

if [ -d "$PLUGIN_DIR/codex/.codex/rules" ]; then
  cp -r "$PLUGIN_DIR/codex/.codex/rules" "$PROJECT_DIR/.codex/"
  echo "Rules copied to .codex/rules"
else
  fail "codex/.codex/rules not found in plugin source"
fi

if [ -f "$PROJECT_DIR/AGENTS.md" ]; then
  backup_agents="$PROJECT_DIR/AGENTS.md.backup.$(date +%Y%m%d%H%M%S)"
  mv "$PROJECT_DIR/AGENTS.md" "$backup_agents"
  echo "Backed up existing AGENTS.md to: $backup_agents"
fi

if [ -f "$PLUGIN_DIR/codex/AGENTS.md" ]; then
  cp "$PLUGIN_DIR/codex/AGENTS.md" "$PROJECT_DIR/AGENTS.md"
fi

if [ "$WITH_MCP" = true ]; then
  if [ -f "$PROJECT_DIR/.codex/config.toml" ]; then
    echo "Warning: .codex/config.toml already exists, skipping"
  elif [ -f "$PLUGIN_DIR/codex/.codex/config.toml" ]; then
    cp "$PLUGIN_DIR/codex/.codex/config.toml" "$PROJECT_DIR/.codex/config.toml"
    echo "config.toml copied (edit MCP server path)"
  else
    echo "Warning: codex/.codex/config.toml not found in plugin source"
  fi
fi

echo "Codex CLI setup complete."
