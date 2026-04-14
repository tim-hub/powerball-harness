#!/usr/bin/env bash
#
# codex-worker-engine.sh
# Codex Worker execution engine
#
# Usage: ./scripts/codex-worker-engine.sh --task "task description" [--worktree PATH] [--dry-run]
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
MAX_RETRIES=""
APPROVAL_POLICY=""
SANDBOX=""

# Configuration initialization (called after check_dependencies)
init_config() {
    validate_config || {
        log_error "Configuration file is invalid"
        exit 1
    }
    MAX_RETRIES=$(get_config "max_retries")
    APPROVAL_POLICY=$(get_config "approval_policy")
    SANDBOX=$(get_config "sandbox")
}

# Global variables
TASK=""
WORKTREE_PATH=""
DRY_RUN=false
PROJECT_ROOT=""
AGENTS_HASH=""
CONTRACT_TEMPLATE="$SCRIPT_DIR/lib/codex-hardening-contract.txt"

# Usage
usage() {
    cat << EOF
Usage: $0 --task "task description" [OPTIONS]

Options:
  --task TEXT       Task to execute (required)
  --worktree PATH   Worktree path (defaults to current directory)
  --dry-run         Dry run (display content without executing)
  -h, --help        Show help

Examples:
  $0 --task "Implement login feature"
  $0 --task "Add API endpoint" --worktree ../worktree-task-1
  $0 --task "Fix tests" --dry-run
EOF
}

# Argument parsing
parse_args() {
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --task)
                if [[ -z "${2:-}" ]]; then
                    log_error "--task requires a value"
                    exit 1
                fi
                TASK="$2"
                shift 2
                ;;
            --worktree)
                if [[ -z "${2:-}" ]]; then
                    log_error "--worktree requires a value"
                    exit 1
                fi
                WORKTREE_PATH="$2"
                shift 2
                ;;
            --dry-run)
                DRY_RUN=true
                shift
                ;;
            -h|--help)
                usage
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                usage
                exit 1
                ;;
        esac
    done

    if [[ -z "$TASK" ]]; then
        log_error "--task is required"
        usage
        exit 1
    fi
}

# Detect project root
detect_project_root() {
    if [[ -n "$WORKTREE_PATH" ]]; then
        # Security: validate worktree path (outside repo is OK, but verify it is a worktree of the same repo)
        if ! validate_worktree_path "$WORKTREE_PATH"; then
            log_error "Invalid worktree path: $WORKTREE_PATH"
            exit 1
        fi
        PROJECT_ROOT="$WORKTREE_PATH"
    else
        PROJECT_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
    fi
    log_info "Project root: $PROJECT_ROOT"
}

# Compute AGENTS.md hash
compute_agents_hash() {
    local agents_file="$PROJECT_ROOT/AGENTS.md"

    if [[ ! -f "$agents_file" ]]; then
        log_error "AGENTS.md not found: $agents_file"
        log_error "AGENTS.md is required. Aborting Worker execution."
        exit 1
    fi

    # Strip BOM, normalize to LF, take first 8 chars of SHA256 (cross-platform)
    AGENTS_HASH=$(calculate_file_hash "$agents_file" 8)
    log_info "AGENTS.md hash: $AGENTS_HASH"
}

# Concatenate rules (sorted by name for determinism)
collect_rules() {
    local rules_dir="$PROJECT_ROOT/.claude/rules"
    local rules_content=""
    local rules_hash=""

    if [[ -d "$rules_dir" ]]; then
        # Quality: sort by name for determinism before concatenation
        local rule_files
        rule_files=$(find "$rules_dir" -name "*.md" -type f 2>/dev/null | sort)

        if [[ -n "$rule_files" ]]; then
            while IFS= read -r rule_file; do
                if [[ -f "$rule_file" ]]; then
                    rules_content+="# $(basename "$rule_file")"$'\n'
                    rules_content+="$(cat "$rule_file")"$'\n\n'
                fi
            done <<< "$rule_files"

            # Log hash of concatenated result (for debugging)
            rules_hash=$(calculate_sha256 "$rules_content" 8 2>/dev/null || echo "unknown")
            log_info "Rules files collected: $(echo "$rule_files" | wc -l | tr -d ' ') (hash: $rules_hash)"
        fi
    else
        log_warn "Rules directory not found: $rules_dir"
    fi

    echo "$rules_content"
}

generate_hardening_contract() {
    if [[ ! -f "$CONTRACT_TEMPLATE" ]]; then
        log_error "hardening contract template not found: $CONTRACT_TEMPLATE"
        exit 1
    fi
    cat "$CONTRACT_TEMPLATE"
}

prepend_hardening_contract() {
    local body="$1"
    printf '%s\n\n---\n\n%s\n' "$(generate_hardening_contract)" "$body"
}

# Generate base-instructions
generate_base_instructions() {
    local rules_content
    rules_content=$(collect_rules)

    local body
    body=$(cat << EOF
# Codex Worker Instructions

## Rules (project-specific rules)

$rules_content

## AGENTS.md mandatory read instruction

Read AGENTS.md first and output evidence in the following format:

\`\`\`
AGENTS_SUMMARY: <one-line summary> | HASH:<first 8 chars of SHA256>
\`\`\`

Do not start work without outputting evidence.
Calculate the evidence hash from the contents of AGENTS.md.

EOF
)
    prepend_hardening_contract "$body"
}

# Generate prompt
generate_prompt() {
    local body
    body=$(cat << EOF
$TASK

---

Important: Before starting work, output AGENTS.md evidence in the following format:

AGENTS_SUMMARY: <one-line summary of AGENTS.md> | HASH:<first 8 chars of SHA256>

If this evidence is missing, the work is considered invalid.
EOF
)
    prepend_hardening_contract "$body"
}

# Evidence verification (intended to be called from Claude Code; not used in this script)
# Actual verification is performed by gate_evidence() in codex-worker-quality-gate.sh
verify_agents_summary() {
    local output="$1"

    # Match with regex (case-insensitive)
    if [[ "$output" =~ AGENTS_SUMMARY:[[:space:]]*(.+)[[:space:]]*\|[[:space:]]*HASH:([A-Fa-f0-9]{8}) ]]; then
        local summary="${BASH_REMATCH[1]}"
        local hash="${BASH_REMATCH[2]}"

        if [[ "${hash,,}" == "${AGENTS_HASH,,}" ]]; then
            log_info "Evidence verification: OK (hash match, summary: ${summary:0:50}...)"
            return 0
        else
            log_error "Evidence verification: NG (hash mismatch: expected=$AGENTS_HASH, actual=$hash)"
            return 1
        fi
    else
        log_error "Evidence verification: NG (AGENTS_SUMMARY not found)"
        return 1
    fi
}

# Invoke Codex Worker (via CLI)
invoke_codex_worker() {
    local base_instructions
    local prompt
    local cwd="$PROJECT_ROOT"

    base_instructions=$(generate_base_instructions)
    prompt=$(generate_prompt)

    if [[ "$DRY_RUN" == true ]]; then
        echo ""
        echo "========================================"
        echo "Dry run: Codex would be called with the following"
        echo "========================================"
        echo ""
        echo "--- prompt ---"
        echo "$prompt"
        echo ""
        echo "--- base-instructions (first 500 chars) ---"
        echo "${base_instructions:0:500}..."
        echo ""
        echo "--- parameters ---"
        echo "cwd: $cwd"
        echo "approval-policy: $APPROVAL_POLICY"
        echo "sandbox: $SANDBOX"
        echo ""
        return 0
    fi

    log_step "Invoking Codex Worker..."

    # Note: the actual codex exec call is made from within Claude Code
    # This script is responsible for generating base-instructions and prompt
    # Output is saved to files and read by Claude Code

    local output_dir="$PROJECT_ROOT/.claude/state/codex-worker"
    mkdir -p "$output_dir"

    echo "$base_instructions" > "$output_dir/base-instructions.txt"
    echo "$prompt" > "$output_dir/prompt.txt"

    jq -n \
        --arg prompt "$prompt" \
        --arg base_instructions "$base_instructions" \
        --arg cwd "$cwd" \
        --arg approval_policy "$APPROVAL_POLICY" \
        --arg sandbox "$SANDBOX" \
        '{
            "prompt": $prompt,
            "base-instructions": $base_instructions,
            "cwd": $cwd,
            "approval-policy": $approval_policy,
            "sandbox": $sandbox
        }' > "$output_dir/codex-exec-params.json"

    # Save verification info (note: agents_hash is excluded for security)
    # This forces the Worker to actually read AGENTS.md and output evidence
    cat > "$output_dir/verify-info.json" << EOF
{
  "max_retries": $MAX_RETRIES,
  "verify_pattern": "AGENTS_SUMMARY:\\\\s*(.+?)\\\\s*\\\\|\\\\s*HASH:([A-Fa-f0-9]{8})",
  "note": "agents_hash is computed by quality-gate at verification time (prevents leaking to Worker)"
}
EOF

    log_info "Codex CLI parameters saved: $output_dir/codex-exec-params.json"
    log_info "Verification info saved: $output_dir/verify-info.json"
    echo ""
    log_info "Next steps:"
    log_info "  1. Call codex exec from Claude Code"
    log_info "  2. Confirm AGENTS_SUMMARY evidence appears in output"
    log_info "  3. Confirm hash matches $AGENTS_HASH"
    log_info "  4. Retry up to $MAX_RETRIES times on failure"
}

# Main processing
main() {
    parse_args "$@"

    echo "========================================"
    echo "Codex Worker Engine"
    echo "========================================"
    echo ""

    log_step "1. Detect project root"
    detect_project_root

    log_step "2. Check dependencies"
    check_dependencies

    log_step "2.5. Initialize configuration"
    init_config

    log_step "3. Compute AGENTS.md hash"
    compute_agents_hash

    log_step "4. Prepare Codex Worker invocation"
    invoke_codex_worker

    echo ""
    log_info "Done"
}

main "$@"
