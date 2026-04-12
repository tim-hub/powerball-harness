#!/usr/bin/env bash
#
# codex-worker-setup.sh
# Codex Worker setup script
#
# Usage: ./scripts/codex-worker-setup.sh [--check-only]
#
# Options:
#   --check-only  Check installation status only (no changes)
#

set -euo pipefail

# Color definitions
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Minimum version requirements
MIN_CODEX_VERSION="0.107.0"
MIN_GIT_VERSION="2.5.0"

# Global variables
CHECK_ONLY=false
ERRORS=()
WARNINGS=()
CODEX_CLI_OK=false
CODEX_EXEC_OK=false

# Helper functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
    WARNINGS+=("$1")
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
    ERRORS+=("$1")
}

# Version comparison (semver)
version_gte() {
    local v1="$1"
    local v2="$2"

    # Convert version string to numeric array
    IFS='.' read -ra V1 <<< "$v1"
    IFS='.' read -ra V2 <<< "$v2"

    # Compare each segment
    for i in 0 1 2; do
        local n1="${V1[$i]:-0}"
        local n2="${V2[$i]:-0}"

        if (( n1 > n2 )); then
            return 0
        elif (( n1 < n2 )); then
            return 1
        fi
    done

    return 0
}

# Check Codex CLI
check_codex_cli() {
    log_info "Checking Codex CLI..."

    if ! command -v codex &> /dev/null; then
        log_error "Codex CLI not found"
        log_info "Install with: npm install -g @openai/codex"
        return 1
    fi

    local version
    version=$(codex --version 2>/dev/null | head -1 | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' || echo "0.0.0")

    if version_gte "$version" "$MIN_CODEX_VERSION"; then
        log_info "Codex CLI v$version (>= $MIN_CODEX_VERSION)"
        CODEX_CLI_OK=true
        return 0
    else
        log_error "Codex CLI v$version is outdated (>= $MIN_CODEX_VERSION required)"
        return 1
    fi
}

# Check Codex auth
check_codex_auth() {
    log_info "Checking Codex auth..."

    if [[ "$CODEX_CLI_OK" != true ]]; then
        log_warn "Skipping: Codex CLI not installed or version too old"
        return 1
    fi

    if codex login status &> /dev/null; then
        log_info "Codex auth: OK"
        return 0
    else
        log_warn "Codex not authenticated: run 'codex login'"
        return 1
    fi
}

# Check Git version (worktree support)
check_git_version() {
    log_info "Checking Git version..."

    if ! command -v git &> /dev/null; then
        log_error "Git not found"
        return 1
    fi

    local version
    version=$(git --version | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' || echo "0.0.0")

    if version_gte "$version" "$MIN_GIT_VERSION"; then
        log_info "Git v$version (>= $MIN_GIT_VERSION, worktree support)"
        return 0
    else
        log_error "Git v$version is outdated (>= $MIN_GIT_VERSION required for worktree support)"
        return 1
    fi
}

# Codex CLI execution check (CLI-only)
check_codex_exec() {
    log_info "Checking Codex CLI execution..."

    if [[ "$CODEX_CLI_OK" != true ]]; then
        log_warn "Skipping: Codex CLI not installed or version too old"
        return 1
    fi

    local timeout_cmd=""
    if command -v timeout &> /dev/null; then
        timeout_cmd="timeout"
    elif command -v gtimeout &> /dev/null; then
        timeout_cmd="gtimeout"
    fi

    if [[ -z "$timeout_cmd" ]]; then
        log_warn "timeout/gtimeout not found (skipping Codex CLI execution check)"
        return 1
    fi

    if "$timeout_cmd" 15 codex exec "echo test" >/dev/null 2>&1; then
        log_info "Codex CLI execution: OK"
        CODEX_EXEC_OK=true
        return 0
    else
        log_warn "Codex CLI execution check failed (check auth/connection/timeout)"
        return 1
    fi
}

# Generate configuration file
generate_config() {
    local config_dir=".claude/state"
    local config_file="$config_dir/codex-worker-config.json"

    log_info "Generating config file..."

    if [[ "$CHECK_ONLY" == true ]]; then
        if [[ -f "$config_file" ]]; then
            log_info "Config file: exists"
        else
            log_warn "Config file: not created"
        fi
        return 0
    fi

    # Create directory
    mkdir -p "$config_dir"

    # Get Codex version
    local codex_version="unknown"
    if [[ "$CODEX_CLI_OK" == true ]]; then
        codex_version=$(codex --version 2>/dev/null | head -1 | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' || echo "unknown")
    fi

    # Codex exec status
    local codex_exec_ready="false"
    if [[ "$CODEX_EXEC_OK" == true ]]; then
        codex_exec_ready="true"
    fi

    # Generate configuration file (key names match common.sh get_config)
    cat > "$config_file" << EOF
{
  "codex_version": "$codex_version",
  "codex_exec_ready": $codex_exec_ready,
  "setup_date": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "approval_policy": "never",
  "sandbox": "workspace-write",
  "ttl_minutes": 30,
  "heartbeat_minutes": 10,
  "max_retries": 3,
  "base_branch": "",
  "gate_skip_allowlist": [],
  "require_gate_pass_for_merge": true,
  "parallel": {
    "enabled": true,
    "max_workers": 3,
    "worktree_base": "../worktrees"
  }
}
EOF

    # Security: owner read/write only
    chmod 600 "$config_file"
    log_info "Config file generated: $config_file"
}

# Main processing
main() {
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --check-only)
                CHECK_ONLY=true
                shift
                ;;
            *)
                echo "Unknown option: $1"
                exit 1
                ;;
        esac
    done

    echo "========================================"
    echo "Codex Worker Setup"
    echo "========================================"
    echo ""

    # Run each check
    check_codex_cli || true
    check_codex_auth || true
    check_git_version || true
    check_codex_exec || true
    generate_config || true

    echo ""
    echo "========================================"
    echo "Results Summary"
    echo "========================================"

    if [[ ${#ERRORS[@]} -eq 0 ]] && [[ ${#WARNINGS[@]} -eq 0 ]]; then
        echo -e "${GREEN}All checks passed${NC}"
        exit 0
    fi

    if [[ ${#WARNINGS[@]} -gt 0 ]]; then
        echo -e "${YELLOW}Warnings (${#WARNINGS[@]}):${NC}"
        for w in "${WARNINGS[@]}"; do
            echo "  - $w"
        done
    fi

    if [[ ${#ERRORS[@]} -gt 0 ]]; then
        echo -e "${RED}Errors (${#ERRORS[@]}):${NC}"
        for e in "${ERRORS[@]}"; do
            echo "  - $e"
        done
        exit 1
    fi

    exit 0
}

main "$@"
