#!/bin/bash
#
# codex-setup-local.sh
#
# Copy Codex CLI templates from the installed Harness plugin.
#
# Usage:
#   ./scripts/codex-setup-local.sh [--user|--project] [--with-mcp|--skip-mcp]
#
set -euo pipefail
IFS=$'\n\t'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(pwd)"
CODEX_HOME_DIR="${CODEX_HOME:-$HOME/.codex}"
WITH_MCP="auto"
TARGET_MODE="user"

while [ $# -gt 0 ]; do
  case "$1" in
    --with-mcp)
      WITH_MCP=true
      ;;
    --skip-mcp)
      WITH_MCP=false
      ;;
    --user)
      TARGET_MODE="user"
      ;;
    --project)
      TARGET_MODE="project"
      ;;
    *)
      echo "Unknown argument: $1" >&2
      echo "Usage: $0 [--user|--project] [--with-mcp|--skip-mcp]" >&2
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

backup_path() {
  local target="$1"
  local backup_root="$2"
  if [ -e "$target" ]; then
    local ts
    local base
    local dst
    ts=$(date +%Y%m%d%H%M%S)
    base="$(basename "$target")"
    mkdir -p "$backup_root"
    dst="$backup_root/${base}.${ts}.$$"
    mv "$target" "$dst"
    echo "Backed up $target to $dst"
  fi
}

should_skip_sync_entry() {
  local name="$1"
  case "$name" in
    _archived|*.backup.*)
      return 0
      ;;
  esac
  return 1
}

cleanup_legacy_skill_entries() {
  local dst_dir="$1"
  local backup_root="$2"
  [ -d "$dst_dir" ] || return 0

  local legacy_path
  for legacy_path in "$dst_dir"/_archived "$dst_dir"/*.backup.*; do
    [ -e "$legacy_path" ] || continue
    backup_path "$legacy_path" "$backup_root"
  done
}

sync_named_children() {
  local src_dir="$1"
  local dst_dir="$2"
  local label="$3"
  local backup_root="$4"

  [ -d "$src_dir" ] || fail "$label source not found: $src_dir"
  mkdir -p "$dst_dir"

  local copied=0
  local skipped=0
  local entry
  for entry in "$src_dir"/*; do
    [ -e "$entry" ] || continue
    local name
    name="$(basename "$entry")"
    if should_skip_sync_entry "$name"; then
      skipped=$((skipped + 1))
      continue
    fi
    local dst_path="$dst_dir/$name"

    if [ -e "$dst_path" ]; then
      backup_path "$dst_path" "$backup_root"
    fi

    cp -R "$entry" "$dst_dir/"
    copied=$((copied + 1))
  done

  echo "$label synced to $dst_dir ($copied items, $skipped skipped)"
}

copy_project_agents() {
  local plugin_dir="$1"
  local backup_root="$2"
  local agents_src="$plugin_dir/codex/AGENTS.md"
  local agents_dst="$PROJECT_DIR/AGENTS.md"

  [ -f "$agents_src" ] || fail "codex/AGENTS.md not found in plugin source"

  if [ -f "$agents_dst" ]; then
    backup_path "$agents_dst" "$backup_root"
  fi

  cp "$agents_src" "$agents_dst"
  echo "AGENTS.md copied to project root"
}

setup_mcp_template() {
  local plugin_dir="$1"
  local target_root="$2"

  [ "$WITH_MCP" = true ] || return 0

  local src="$plugin_dir/codex/.codex/config.toml"
  local dst="$target_root/config.toml"

  if [ -f "$dst" ]; then
    echo "Warning: $dst already exists, skipping"
    return 0
  fi

  if [ -f "$src" ]; then
    mkdir -p "$target_root"
    cp "$src" "$dst"
    echo "config.toml copied to: $dst"
    echo "Edit MCP server/notify paths for your environment"
  else
    echo "Warning: codex/.codex/config.toml not found in plugin source"
  fi
}

ensure_multi_agent_defaults() {
  local target_root="$1"
  local cfg="$target_root/config.toml"

  mkdir -p "$target_root"

  if [ ! -f "$cfg" ]; then
    cat > "$cfg" <<'CFG'
[features]
multi_agent = true

[agents]
max_threads = 8

[agents.implementer]
description = "Codex implementation worker for harness task execution"

[agents.reviewer]
description = "Codex reviewer worker for harness review and retake loops"

[agents.claude_implementer]
description = "Claude CLI delegated implementation worker (used when --claude)"

[agents.claude_reviewer]
description = "Claude CLI delegated reviewer worker (used when --claude)"
CFG
    echo "Created $cfg with multi_agent + harness role defaults"
    return
  fi

  if ! grep -q '^[[:space:]]*multi_agent[[:space:]]*=' "$cfg"; then
    cat >> "$cfg" <<'CFG'

[features]
multi_agent = true
CFG
    echo "Enabled features.multi_agent in $cfg"
  fi

  if ! grep -q '^\[agents\]' "$cfg"; then
    cat >> "$cfg" <<'CFG'

[agents]
max_threads = 8
CFG
  fi

  if ! grep -q '^\[agents\.implementer\]' "$cfg"; then
    cat >> "$cfg" <<'CFG'

[agents.implementer]
description = "Codex implementation worker for harness task execution"
CFG
  fi

  if ! grep -q '^\[agents\.reviewer\]' "$cfg"; then
    cat >> "$cfg" <<'CFG'

[agents.reviewer]
description = "Codex reviewer worker for harness review and retake loops"
CFG
  fi

  if ! grep -q '^\[agents\.claude_implementer\]' "$cfg"; then
    cat >> "$cfg" <<'CFG'

[agents.claude_implementer]
description = "Claude CLI delegated implementation worker (used when --claude)"
CFG
  fi

  if ! grep -q '^\[agents\.claude_reviewer\]' "$cfg"; then
    cat >> "$cfg" <<'CFG'

[agents.claude_reviewer]
description = "Claude CLI delegated reviewer worker (used when --claude)"
CFG
  fi
}

PLUGIN_DIR="$(resolve_plugin_dir || true)"
if [ -z "$PLUGIN_DIR" ]; then
  fail "Harness plugin directory not found. Set CLAUDE_PLUGIN_ROOT or install the plugin."
fi

echo "Using Harness plugin: $PLUGIN_DIR"

target_root=""
backup_root=""
if [ "$TARGET_MODE" = "user" ]; then
  target_root="$CODEX_HOME_DIR"
  backup_root="$CODEX_HOME_DIR/backups/codex-setup-local"
  echo "Install mode: user (target: $target_root)"
else
  target_root="$PROJECT_DIR/.codex"
  backup_root="$target_root/backups/codex-setup-local"
  echo "Install mode: project (target: $target_root)"
fi

cleanup_legacy_skill_entries "$target_root/skills" "$backup_root"
sync_named_children "$PLUGIN_DIR/codex/.codex/skills" "$target_root/skills" "Skills" "$backup_root"
sync_named_children "$PLUGIN_DIR/codex/.codex/rules" "$target_root/rules" "Rules" "$backup_root"

if [ "$TARGET_MODE" = "project" ]; then
  copy_project_agents "$PLUGIN_DIR" "$backup_root"
else
  echo "User mode: project AGENTS.md is unchanged"
fi

setup_mcp_template "$PLUGIN_DIR" "$target_root"
ensure_multi_agent_defaults "$target_root"

echo "Codex CLI setup complete."
echo "Backups are stored under: $backup_root (outside skill scan path)"
if [ "$TARGET_MODE" = "user" ]; then
  echo "Restart Codex to reload user-level skills/rules if needed."
fi
