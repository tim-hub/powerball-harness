#!/bin/bash
#
# setup-codex.sh
#
# Setup Harness for Codex CLI.
#
# Usage:
#   ./scripts/setup-codex.sh [--user|--project] [--with-mcp|--skip-mcp]
#

set -euo pipefail
IFS=$'\n\t'

HARNESS_REPO="https://github.com/Chachamaru127/claude-code-harness.git"
HARNESS_BRANCH="main"
TEMP_DIR=$(mktemp -d)
PROJECT_DIR=$(pwd)
CODEX_HOME_DIR="${CODEX_HOME:-$HOME/.codex}"
TARGET_MODE="user"
WITH_MCP="auto"

cleanup() {
    rm -rf "$TEMP_DIR"
}
trap cleanup EXIT

log_info() { echo "[INFO] $1"; }
log_warn() { echo "[WARN] $1"; }
log_ok() { echo "[OK]   $1"; }
log_err() { echo "[ERR]  $1" >&2; }

usage() {
    cat <<USAGE
Usage: $0 [--user|--project] [--with-mcp|--skip-mcp]

Modes:
  --user      Install to CODEX_HOME (default: $CODEX_HOME_DIR)
  --project   Install to current project (.codex/ + AGENTS.md)

MCP:
  --with-mcp  Copy config.toml template to target root
  --skip-mcp  Do not copy config.toml template
USAGE
}

parse_args() {
    while [ $# -gt 0 ]; do
        case "$1" in
            --user)
                TARGET_MODE="user"
                ;;
            --project)
                TARGET_MODE="project"
                ;;
            --with-mcp)
                WITH_MCP=true
                ;;
            --skip-mcp)
                WITH_MCP=false
                ;;
            -h|--help)
                usage
                exit 0
                ;;
            *)
                log_err "Unknown argument: $1"
                usage
                exit 1
                ;;
        esac
        shift
    done
}

check_requirements() {
    log_info "Checking requirements..."
    if ! command -v git >/dev/null 2>&1; then
        log_err "git is required but not installed"
        exit 1
    fi
    log_ok "All requirements met"
}

clone_harness() {
    log_info "Downloading Harness..."
    git clone --depth 1 --branch "$HARNESS_BRANCH" "$HARNESS_REPO" "$TEMP_DIR/harness" 2>/dev/null || {
        log_err "Failed to clone Harness repository"
        exit 1
    }
    log_ok "Harness downloaded"
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
        log_warn "Backed up $target to $dst"
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

    [ -d "$src_dir" ] || {
        log_err "$label source not found: $src_dir"
        exit 1
    }

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

    log_ok "$label synced to $dst_dir ($copied items, $skipped skipped)"
}

copy_project_agents() {
    local backup_root="$1"
    local src="$TEMP_DIR/harness/codex/AGENTS.md"
    local dst="$PROJECT_DIR/AGENTS.md"

    if [ ! -f "$src" ]; then
        log_err "codex/AGENTS.md not found in Harness"
        exit 1
    fi

    if [ -f "$dst" ]; then
        backup_path "$dst" "$backup_root"
    fi

    cp "$src" "$dst"
    log_ok "AGENTS.md copied to project root"
}

resolve_target_root() {
    if [ "$TARGET_MODE" = "user" ]; then
        echo "$CODEX_HOME_DIR"
    else
        echo "$PROJECT_DIR/.codex"
    fi
}

resolve_backup_root() {
    local target_root="$1"
    if [ "$TARGET_MODE" = "user" ]; then
        echo "$CODEX_HOME_DIR/backups/setup-codex"
    else
        echo "$target_root/backups/setup-codex"
    fi
}

setup_mcp_template() {
    local target_root="$1"

    if [ "$WITH_MCP" = "auto" ]; then
        echo ""
        echo "Setup MCP server config template? (optional)"
        read -r -p "Setup MCP template? (y/N): " setup_mcp_answer
        if [[ "$setup_mcp_answer" =~ ^[Yy]$ ]]; then
            WITH_MCP=true
        else
            WITH_MCP=false
        fi
    fi

    [ "$WITH_MCP" = true ] || return 0

    local src="$TEMP_DIR/harness/codex/.codex/config.toml"
    local dst="$target_root/config.toml"

    if [ -f "$dst" ]; then
        log_warn "$dst already exists, skipping"
        return
    fi

    if [ -f "$src" ]; then
        mkdir -p "$target_root"
        cp "$src" "$dst"
        log_ok "config.toml copied to $dst"
        log_warn "Edit config.toml to set the correct MCP server/notify paths"
    else
        log_warn "codex/.codex/config.toml not found in Harness"
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
        log_ok "Created $cfg with multi_agent + harness role defaults"
        return
    fi

    if ! grep -q '^[[:space:]]*multi_agent[[:space:]]*=' "$cfg"; then
        cat >> "$cfg" <<'CFG'

[features]
multi_agent = true
CFG
        log_ok "Enabled features.multi_agent in $cfg"
    fi

    if ! grep -q '^\[agents\]' "$cfg"; then
        cat >> "$cfg" <<'CFG'

[agents]
max_threads = 8
CFG
        log_ok "Added [agents] defaults to $cfg"
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

print_success() {
    local target_root="$1"
    local backup_root="$2"

    echo ""
    echo "============================================"
    echo "Harness for Codex CLI setup complete."
    echo "============================================"
    echo ""
    echo "Mode: $TARGET_MODE"
    echo "Target: $target_root"
    echo ""
    echo "Created/updated:"
    echo "  $target_root/skills/  - Harness skills"
    echo "  $target_root/rules/   - Guardrails"
    echo "  $backup_root/ - Setup backups (outside skill scan path)"
    [ -f "$target_root/config.toml" ] && echo "  $target_root/config.toml - MCP template"

    if [ "$TARGET_MODE" = "project" ]; then
        echo "  $PROJECT_DIR/AGENTS.md - Project instructions"
    else
        echo "  (project AGENTS.md unchanged in user mode)"
    fi

    echo ""
    echo "Next steps:"
    echo "  1. Restart Codex"
    echo "  2. Use \$skill-name to invoke skills (example: \$work)"
    echo ""
}

main() {
    parse_args "$@"

    echo ""
    log_info "Setting up Harness for Codex CLI"
    log_info "Mode: $TARGET_MODE"
    if [ "$TARGET_MODE" = "user" ]; then
        log_info "Target: $CODEX_HOME_DIR"
    else
        log_info "Target: $PROJECT_DIR/.codex"
    fi
    echo ""

    check_requirements
    clone_harness

    local target_root
    target_root="$(resolve_target_root)"
    local backup_root
    backup_root="$(resolve_backup_root "$target_root")"

    cleanup_legacy_skill_entries "$target_root/skills" "$backup_root"
    sync_named_children "$TEMP_DIR/harness/codex/.codex/skills" "$target_root/skills" "Skills" "$backup_root"
    sync_named_children "$TEMP_DIR/harness/codex/.codex/rules" "$target_root/rules" "Rules" "$backup_root"

    if [ "$TARGET_MODE" = "project" ]; then
        copy_project_agents "$backup_root"
    else
        log_info "User mode: project AGENTS.md is unchanged"
    fi

    setup_mcp_template "$target_root"
    ensure_multi_agent_defaults "$target_root"
    print_success "$target_root" "$backup_root"
}

main "$@"
