#!/usr/bin/env bash
#
# codex-worker-lock.sh
# Task ownership and lock mechanism
#
# Usage:
#   ./scripts/codex-worker-lock.sh acquire --path PATH --worker WORKER_ID
#   ./scripts/codex-worker-lock.sh release --path PATH --worker WORKER_ID
#   ./scripts/codex-worker-lock.sh heartbeat --path PATH --worker WORKER_ID
#   ./scripts/codex-worker-lock.sh check --path PATH
#   ./scripts/codex-worker-lock.sh cleanup
#

set -euo pipefail

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Load common library
# shellcheck source=lib/codex-worker-common.sh
source "$SCRIPT_DIR/lib/codex-worker-common.sh"

# ============================================
# Local configuration (initialized in main)
# ============================================
TTL_MINUTES=""
HEARTBEAT_MINUTES=""
LOCK_DIR=""  # Initialized as absolute path
LOCK_LOG=""  # Initialized as absolute path

# Configuration initialization (called after check_dependencies)
init_lock_config() {
    validate_config || {
        log_error "Configuration file is invalid"
        exit 1
    }
    TTL_MINUTES=$(get_config "ttl_minutes")
    HEARTBEAT_MINUTES=$(get_config "heartbeat_minutes")

    # Security: fix as absolute path (eliminate CWD dependency)
    local repo_root
    repo_root=$(get_repo_root) || exit 1
    LOCK_DIR="$repo_root/.claude/state/locks"
    LOCK_LOG="$repo_root/.claude/state/locks.log"
}

# ============================================
# Lock-specific functions
# ============================================

# Validate and initialize lock directory
# Security: prevent symlink attacks (including parent directories)
# Note: LOCK_DIR is already set to an absolute path by init_lock_config()
init_lock_dir() {
    local repo_root
    repo_root=$(get_repo_root) || exit 1

    # Resolve repository root
    local real_repo_root
    real_repo_root=$(realpath "$repo_root" 2>/dev/null) || {
        log_error "Cannot resolve repository root: $repo_root"
        exit 1
    }

    # LOCK_DIR is already an absolute path (set by init_lock_config)
    local full_lock_dir="$LOCK_DIR"

    # Symlink check on parent directories (Security: verify each level)
    local check_path="$repo_root"
    for segment in .claude state locks; do
        check_path="$check_path/$segment"
        if [[ -L "$check_path" ]]; then
            log_error "Path hierarchy contains a symlink (forbidden for security): $check_path"
            exit 1
        fi
    done

    # If directory exists, confirm it is inside the repository
    if [[ -e "$full_lock_dir" ]]; then
        local real_lock_dir
        real_lock_dir=$(realpath "$full_lock_dir" 2>/dev/null) || {
            log_error "Cannot resolve lock directory: $full_lock_dir"
            exit 1
        }

        # Security: distinguish /repo from /repo2
        if [[ "$real_lock_dir" != "$real_repo_root" && "$real_lock_dir" != "$real_repo_root/"* ]]; then
            log_error "Lock directory is outside the repository: $real_lock_dir"
            exit 1
        fi
    fi

    # Create directory (Security: 700 permissions)
    mkdir -p "$full_lock_dir"
    chmod 700 "$full_lock_dir"
}

# Generate lock key (first 8 chars of SHA256)
generate_lock_key() {
    local path="$1"
    local normalized
    normalized=$(normalize_path "$path")
    calculate_sha256 "$normalized" 8
}

# Get lock file path
get_lock_file() {
    local path="$1"
    local key
    key=$(generate_lock_key "$path")
    printf '%s/%s.lock.json' "$LOCK_DIR" "$key"
}

# Get multiple lock file fields in one call (performance optimization)
# Usage: read_lock_fields "$lock_file" worker heartbeat path
# Returns: tab-separated values
read_lock_fields() {
    local lock_file="$1"
    shift
    local fields=("$@")

    if [[ ! -f "$lock_file" ]]; then
        return 1
    fi

    # Retrieve multiple fields with a single jq call
    local jq_filter
    jq_filter=$(printf '.%s, ' "${fields[@]}")
    jq_filter="${jq_filter%, }"  # Remove trailing comma

    jq -r "[$jq_filter] | @tsv" "$lock_file"
}

# Record log event
log_event() {
    local event="$1"
    local path="$2"
    local worker="$3"
    mkdir -p "$(dirname "$LOCK_LOG")"
    printf '%s\t%s\t%s\t%s\n' "$(now_utc)" "$event" "$path" "$worker" >> "$LOCK_LOG"
}

# Acquire lock
cmd_acquire() {
    local path=""
    local worker=""

    while [[ $# -gt 0 ]]; do
        case "$1" in
            --path)
                if [[ -z "${2:-}" ]]; then
                    log_error "--path requires a value"
                    exit 1
                fi
                path="$2"; shift 2 ;;
            --worker)
                if [[ -z "${2:-}" ]]; then
                    log_error "--worker requires a value"
                    exit 1
                fi
                worker="$2"; shift 2 ;;
            *) log_error "Unknown option: $1"; exit 1 ;;
        esac
    done

    if [[ -z "$path" ]] || [[ -z "$worker" ]]; then
        log_error "--path and --worker are required"
        exit 1
    fi

    # Security: path validation
    if ! validate_repo_path "$path"; then
        exit 1
    fi

    # Security: validate lock directory
    init_lock_dir

    local lock_file
    lock_file=$(get_lock_file "$path")
    local normalized_path
    normalized_path=$(normalize_path "$path")

    # Check existing lock (Performance: get multiple fields with one jq call)
    if [[ -f "$lock_file" ]]; then
        local lock_data
        lock_data=$(read_lock_fields "$lock_file" worker heartbeat) || {
            log_warn "Failed to read lock file: $lock_file"
            rm -f "$lock_file"
        }

        local existing_worker
        local heartbeat
        existing_worker=$(echo "$lock_data" | cut -f1)
        heartbeat=$(echo "$lock_data" | cut -f2)

        # TTL check
        local heartbeat_epoch
        local now_epoch
        local ttl_seconds=$((TTL_MINUTES * 60))

        heartbeat_epoch=$(parse_utc_to_epoch "$heartbeat")
        now_epoch=$(date "+%s")

        if (( now_epoch - heartbeat_epoch > ttl_seconds )); then
            log_warn "TTL exceeded: releasing existing lock (worker=$existing_worker)"
            log_event "expired" "$normalized_path" "$existing_worker"
            rm -f "$lock_file"
        else
            log_error "Lock acquisition failed: $normalized_path is locked by $existing_worker"
            exit 1
        fi
    fi

    # Create new lock (atomic creation)
    local now
    now=$(now_utc)

    # Security: create with owner-only read/write permissions
    local tmp_file
    tmp_file=$(mktemp "$LOCK_DIR/tmp.XXXXXX")
    jq -n \
        --arg path "$normalized_path" \
        --arg worker "$worker" \
        --arg acquired "$now" \
        --arg heartbeat "$now" \
        '{
            path: $path,
            worker: $worker,
            acquired: $acquired,
            heartbeat: $heartbeat
        }' > "$tmp_file"
    chmod 600 "$tmp_file"

    # Atomic placement via ln (fails if file already exists)
    if ! ln "$tmp_file" "$lock_file" 2>/dev/null; then
        rm -f "$tmp_file"
        log_error "Lock acquisition failed: $normalized_path is locked by another Worker (race condition)"
        exit 1
    fi

    chmod 600 "$lock_file"
    rm -f "$tmp_file"
    log_event "acquire" "$normalized_path" "$worker"
    log_info "Lock acquired: $normalized_path (worker=$worker)"
}

# Release lock
cmd_release() {
    local path=""
    local worker=""

    while [[ $# -gt 0 ]]; do
        case "$1" in
            --path)
                if [[ -z "${2:-}" ]]; then
                    log_error "--path requires a value"
                    exit 1
                fi
                path="$2"; shift 2 ;;
            --worker)
                if [[ -z "${2:-}" ]]; then
                    log_error "--worker requires a value"
                    exit 1
                fi
                worker="$2"; shift 2 ;;
            *) log_error "Unknown option: $1"; exit 1 ;;
        esac
    done

    if [[ -z "$path" ]] || [[ -z "$worker" ]]; then
        log_error "--path and --worker are required"
        exit 1
    fi

    # Security: path validation
    if ! validate_repo_path "$path"; then
        exit 1
    fi

    local lock_file
    lock_file=$(get_lock_file "$path")
    local normalized_path
    normalized_path=$(normalize_path "$path")

    if [[ ! -f "$lock_file" ]]; then
        log_warn "Lock does not exist: $normalized_path"
        exit 0
    fi

    local existing_worker
    existing_worker=$(jq -r '.worker' "$lock_file")

    if [[ "$existing_worker" != "$worker" ]]; then
        log_error "Lock release failed: $normalized_path is locked by $existing_worker"
        exit 1
    fi

    rm -f "$lock_file"
    log_event "release" "$normalized_path" "$worker"
    log_info "Lock released: $normalized_path (worker=$worker)"
}

# Update heartbeat
cmd_heartbeat() {
    local path=""
    local worker=""

    while [[ $# -gt 0 ]]; do
        case "$1" in
            --path)
                if [[ -z "${2:-}" ]]; then
                    log_error "--path requires a value"
                    exit 1
                fi
                path="$2"; shift 2 ;;
            --worker)
                if [[ -z "${2:-}" ]]; then
                    log_error "--worker requires a value"
                    exit 1
                fi
                worker="$2"; shift 2 ;;
            *) log_error "Unknown option: $1"; exit 1 ;;
        esac
    done

    if [[ -z "$path" ]] || [[ -z "$worker" ]]; then
        log_error "--path and --worker are required"
        exit 1
    fi

    # Security: path validation
    if ! validate_repo_path "$path"; then
        exit 1
    fi

    local lock_file
    lock_file=$(get_lock_file "$path")
    local normalized_path
    normalized_path=$(normalize_path "$path")

    if [[ ! -f "$lock_file" ]]; then
        log_error "Lock does not exist: $normalized_path"
        exit 1
    fi

    local existing_worker
    existing_worker=$(jq -r '.worker' "$lock_file")

    if [[ "$existing_worker" != "$worker" ]]; then
        log_error "Heartbeat update failed: $normalized_path is locked by $existing_worker"
        exit 1
    fi

    local now
    now=$(now_utc)

    jq --arg heartbeat "$now" '.heartbeat = $heartbeat' "$lock_file" > "$lock_file.tmp"
    # Security: preserve permissions
    chmod 600 "$lock_file.tmp"
    mv "$lock_file.tmp" "$lock_file"

    log_info "Heartbeat updated: $normalized_path (worker=$worker)"
}

# Check lock status
cmd_check() {
    local path=""

    while [[ $# -gt 0 ]]; do
        case "$1" in
            --path)
                if [[ -z "${2:-}" ]]; then
                    log_error "--path requires a value"
                    exit 1
                fi
                path="$2"; shift 2 ;;
            *) log_error "Unknown option: $1"; exit 1 ;;
        esac
    done

    if [[ -z "$path" ]]; then
        log_error "--path is required"
        exit 1
    fi

    # Security: path validation
    if ! validate_repo_path "$path"; then
        exit 1
    fi

    local lock_file
    lock_file=$(get_lock_file "$path")
    local normalized_path
    normalized_path=$(normalize_path "$path")

    if [[ ! -f "$lock_file" ]]; then
        echo '{"locked": false}'
        exit 0
    fi

    local worker
    local heartbeat
    worker=$(jq -r '.worker' "$lock_file")
    heartbeat=$(jq -r '.heartbeat' "$lock_file")

    # TTL check
    local heartbeat_epoch
    local now_epoch
    local ttl_seconds=$((TTL_MINUTES * 60))

    heartbeat_epoch=$(parse_utc_to_epoch "$heartbeat")
    now_epoch=$(date "+%s")

    if (( now_epoch - heartbeat_epoch > ttl_seconds )); then
        # TTL exceeded: read-only (deletion is done by acquire/cleanup)
        echo '{"locked": false, "expired": true, "hint": "run cleanup or acquire to release"}'
    else
        jq -c '. + {locked: true}' "$lock_file"
    fi
}

# Clean up expired locks
cmd_cleanup() {
    # Security: validate lock directory
    init_lock_dir

    local cleaned=0
    local now_epoch
    now_epoch=$(date "+%s")
    local ttl_seconds=$((TTL_MINUTES * 60))

    for lock_file in "$LOCK_DIR"/*.lock.json; do
        [[ -f "$lock_file" ]] || continue

        # Performance: get multiple fields with one jq call
        local lock_data
        lock_data=$(read_lock_fields "$lock_file" heartbeat worker path) || continue

        local heartbeat
        local worker
        local path
        heartbeat=$(echo "$lock_data" | cut -f1)
        worker=$(echo "$lock_data" | cut -f2)
        path=$(echo "$lock_data" | cut -f3)

        local heartbeat_epoch
        heartbeat_epoch=$(parse_utc_to_epoch "$heartbeat")

        if (( now_epoch - heartbeat_epoch > ttl_seconds )); then
            log_warn "TTL exceeded: $path (worker=$worker)"
            log_event "expired" "$path" "$worker"
            rm -f "$lock_file"
            cleaned=$((cleaned + 1))
        fi
    done

    log_info "Cleanup complete: released $cleaned lock(s)"
}

# Usage
usage() {
    cat << EOF
Usage: $0 COMMAND [OPTIONS]

Commands:
  acquire   --path PATH --worker WORKER_ID   Acquire lock
  release   --path PATH --worker WORKER_ID   Release lock
  heartbeat --path PATH --worker WORKER_ID   Update heartbeat
  check     --path PATH                      Check lock status
  cleanup                                    Clean up expired locks

Options:
  --path PATH       Target file path
  --worker WORKER_ID Worker identifier

Settings:
  TTL: $TTL_MINUTES min
  Heartbeat interval: $HEARTBEAT_MINUTES min
  Lock directory: $LOCK_DIR

Examples:
  $0 acquire --path src/auth/login.ts --worker worker-1
  $0 heartbeat --path src/auth/login.ts --worker worker-1
  $0 release --path src/auth/login.ts --worker worker-1
  $0 check --path src/auth/login.ts
  $0 cleanup
EOF
}

# Main processing
main() {
    check_dependencies
    init_lock_config

    if [[ $# -eq 0 ]]; then
        usage
        exit 0
    fi

    local cmd="$1"
    shift

    case "$cmd" in
        acquire)   cmd_acquire "$@" ;;
        release)   cmd_release "$@" ;;
        heartbeat) cmd_heartbeat "$@" ;;
        check)     cmd_check "$@" ;;
        cleanup)   cmd_cleanup ;;
        -h|--help) usage; exit 0 ;;
        *)
            log_error "Unknown command: $cmd"
            usage
            exit 1
            ;;
    esac
}

main "$@"
