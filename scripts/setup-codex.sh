#!/bin/bash
#
# setup-codex.sh
#
# Setup Harness for Codex CLI
#
# Usage:
#   ./scripts/setup-codex.sh
#

set -e

HARNESS_REPO="https://github.com/Chachamaru127/claude-code-harness.git"
HARNESS_BRANCH="main"
TEMP_DIR=$(mktemp -d)
PROJECT_DIR=$(pwd)

cleanup() {
    rm -rf "$TEMP_DIR"
}
trap cleanup EXIT

log_info() { echo "[INFO] $1"; }
log_warn() { echo "[WARN] $1"; }
log_ok() { echo "[OK]   $1"; }
log_err() { echo "[ERR]  $1" >&2; }

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
    git clone --depth 1 --branch "$HARNESS_BRANCH" "$HARNESS_REPO" "$TEMP_DIR/harness" 2>/dev/null || \
        { log_err "Failed to clone Harness repository"; exit 1; }
    log_ok "Harness downloaded"
}

backup_dir() {
    local target="$1"
    if [ -d "$target" ] && [ "$(ls -A "$target" 2>/dev/null)" ]; then
        local ts
        ts=$(date +%Y%m%d%H%M%S)
        mv "$target" "${target}.backup.${ts}"
        log_warn "Backed up $target to ${target}.backup.${ts}"
    fi
}

copy_codex_files() {
    log_info "Setting up Codex files..."

    mkdir -p "$PROJECT_DIR/.codex"

    backup_dir "$PROJECT_DIR/.codex/skills"
    backup_dir "$PROJECT_DIR/.codex/rules"

    if [ -d "$TEMP_DIR/harness/codex/.codex/skills" ]; then
        cp -r "$TEMP_DIR/harness/codex/.codex/skills" "$PROJECT_DIR/.codex/"
        log_ok "Skills copied to .codex/skills"
    else
        log_err "codex/.codex/skills not found in Harness"
        exit 1
    fi

    if [ -d "$TEMP_DIR/harness/codex/.codex/rules" ]; then
        cp -r "$TEMP_DIR/harness/codex/.codex/rules" "$PROJECT_DIR/.codex/"
        log_ok "Rules copied to .codex/rules"
    else
        log_err "codex/.codex/rules not found in Harness"
        exit 1
    fi

    if [ -f "$PROJECT_DIR/AGENTS.md" ]; then
        local ts
        ts=$(date +%Y%m%d%H%M%S)
        mv "$PROJECT_DIR/AGENTS.md" "$PROJECT_DIR/AGENTS.md.backup.${ts}"
        log_warn "Backed up existing AGENTS.md to AGENTS.md.backup.${ts}"
    fi

    if [ -f "$TEMP_DIR/harness/codex/AGENTS.md" ]; then
        cp "$TEMP_DIR/harness/codex/AGENTS.md" "$PROJECT_DIR/AGENTS.md"
        log_ok "AGENTS.md copied"
    else
        log_err "codex/AGENTS.md not found in Harness"
        exit 1
    fi
}

setup_mcp() {
    echo ""
    echo "Setup MCP server config? (optional)"
    read -r -p "Setup MCP? (y/N): " setup_mcp_answer

    if [[ "$setup_mcp_answer" =~ ^[Yy]$ ]]; then
        if [ -f "$PROJECT_DIR/.codex/config.toml" ]; then
            log_warn ".codex/config.toml already exists, skipping"
            return
        fi

        if [ -f "$TEMP_DIR/harness/codex/.codex/config.toml" ]; then
            cp "$TEMP_DIR/harness/codex/.codex/config.toml" "$PROJECT_DIR/.codex/config.toml"
            log_ok "config.toml copied"
            log_warn "Edit .codex/config.toml to set the correct MCP server path"
        else
            log_warn "codex/.codex/config.toml not found in Harness"
        fi
    fi
}

print_success() {
    echo ""
    echo "============================================"
    echo "Harness for Codex CLI setup complete."
    echo "============================================"
    echo ""
    echo "Created/updated:"
    echo "  .codex/skills/    - Harness skills"
    echo "  .codex/rules/     - Temporary guardrails"
    [ -f "$PROJECT_DIR/.codex/config.toml" ] && echo "  .codex/config.toml - MCP configuration"
    echo "  AGENTS.md         - Project instructions"
    echo ""
    echo "Next steps:"
    echo "  1. Start Codex in your project"
    echo "  2. Use \\$skill-name to invoke skills (example: \\$work)"
    echo ""
}

main() {
    echo ""
    log_info "Setting up Harness for Codex CLI in: $PROJECT_DIR"
    echo ""

    check_requirements
    clone_harness
    copy_codex_files
    setup_mcp
    print_success
}

main "$@"
