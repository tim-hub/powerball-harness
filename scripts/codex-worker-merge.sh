#!/usr/bin/env bash
#
# codex-worker-merge.sh
# Merge integration of Worker artifacts
#
# Usage:
#   ./scripts/codex-worker-merge.sh --worktree PATH --target-branch BRANCH [--squash] [--dry-run]
#

set -euo pipefail

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Load common library
# shellcheck source=lib/codex-worker-common.sh
source "$SCRIPT_DIR/lib/codex-worker-common.sh"

# Update Plans.md
# Note: Use absolute path from repo root to avoid CWD dependency
update_plans() {
    local task_pattern="$1"
    local repo_root
    repo_root=$(get_repo_root) || return 1
    local plans_file="$repo_root/Plans.md"

    if [[ ! -f "$plans_file" ]]; then
        log_warn "Plans.md not found: $plans_file"
        return 1
    fi

    # cc:WIP → cc:done, [ ] → [x]
    if grep -q "$task_pattern" "$plans_file"; then
        sed -i.bak "s/\(.*$task_pattern.*\)cc:WIP/\1cc:done/" "$plans_file"
        sed -i.bak "s/\(.*$task_pattern.*\)\[ \]/\1[x]/" "$plans_file"
        rm -f "$plans_file.bak"
        log_info "Plans.md updated: $task_pattern → cc:done"
        return 0
    fi

    return 1
}

# Cherry-pick merge
do_cherry_pick() {
    local commit_hash="$1"
    local dry_run="$2"

    log_merge "cherry-pick: $commit_hash"

    if [[ "$dry_run" == "true" ]]; then
        log_info "[DRY-RUN] git cherry-pick $commit_hash"
        return 0
    fi

    if git cherry-pick "$commit_hash" 2>/dev/null; then
        return 0
    else
        # Conflict occurred
        git cherry-pick --abort 2>/dev/null || true
        return 1
    fi
}

# Squash merge
do_squash_merge() {
    local worktree="$1"
    local dry_run="$2"

    # Get the worktree branch name
    local branch_name
    branch_name=$(cd "$worktree" && git branch --show-current)

    log_merge "squash merge: $branch_name"

    if [[ "$dry_run" == "true" ]]; then
        log_info "[DRY-RUN] git merge --squash $branch_name"
        return 0
    fi

    if git merge --squash "$branch_name" 2>/dev/null; then
        git commit -m "feat: Merge Worker artifacts ($branch_name)"
        return 0
    else
        git merge --abort 2>/dev/null || true
        return 1
    fi
}

# Main processing
main() {
    check_dependencies

    local worktree=""
    local target_branch=""
    local squash=false
    local dry_run=false
    local force=false

    while [[ $# -gt 0 ]]; do
        case "$1" in
            --worktree)
                if [[ -z "${2:-}" ]]; then
                    log_error "--worktree requires a value"
                    exit 1
                fi
                worktree="$2"; shift 2 ;;
            --target-branch)
                if [[ -z "${2:-}" ]]; then
                    log_error "--target-branch requires a value"
                    exit 1
                fi
                target_branch="$2"; shift 2 ;;
            --squash) squash=true; shift ;;
            --dry-run) dry_run=true; shift ;;
            --force) force=true; shift ;;
            -h|--help) usage; exit 0 ;;
            *) log_error "Unknown option: $1"; exit 1 ;;
        esac
    done

    # Get default branch
    if [[ -z "$target_branch" ]]; then
        target_branch=$(get_default_branch)
    fi

    # Required parameter check
    if [[ -z "$worktree" ]]; then
        log_error "--worktree is required"
        exit 1
    fi

    if [[ ! -d "$worktree" ]]; then
        log_error "Worktree does not exist: $worktree"
        exit 1
    fi

    # Security: verify this is a worktree of the same repository (using common function)
    if ! validate_worktree_path "$worktree"; then
        exit 1
    fi

    # Quality: check if worktree working tree is clean
    local worktree_status
    worktree_status=$(cd "$worktree" && git status --porcelain 2>/dev/null)
    if [[ -n "$worktree_status" ]]; then
        log_warn "Worktree has uncommitted changes:"
        echo "$worktree_status" | head -5
        if [[ "$force" != "true" ]]; then
            log_error "Aborting due to uncommitted changes. Use --force to skip"
            echo '{"status": "blocked", "reason": "uncommitted_changes"}'
            exit 1
        fi
        log_warn "Ignoring uncommitted changes and proceeding with merge"
    fi

    # Security: verify quality gate pass (check centrally managed gate results)
    local require_gate_pass
    require_gate_pass=$(get_config "require_gate_pass_for_merge")

    if [[ "$require_gate_pass" == "true" ]]; then
        # verify_gate_result checks gate results corresponding to the worktree HEAD commit
        # References centrally managed result files that Workers cannot tamper with
        if ! verify_gate_result "$worktree"; then
            log_error "Please pass the quality gate before merging"
            log_error "Use --force to skip, but this is not recommended"

            if [[ "$force" != "true" ]]; then
                echo '{"status": "blocked", "reason": "gate_not_passed"}'
                exit 1
            fi
            log_warn "Force-merging without passing quality gate"
        fi
    fi

    # Get latest commit from worktree
    local commit_hash
    commit_hash=$(cd "$worktree" && git log -1 --format="%H")

    if [[ -z "$commit_hash" ]]; then
        log_error "No commit found"
        echo '{"status": "failed", "commit_hash": null, "conflicts": [], "plans_updated": false}'
        exit 1
    fi

    log_info "Worker commit: $commit_hash"

    # Check current branch
    local current_branch
    current_branch=$(git branch --show-current)

    # Validate target branch
    if ! git check-ref-format --branch "$target_branch" 2>/dev/null; then
        log_error "Invalid branch name: $target_branch"
        echo '{"status": "failed", "commit_hash": null, "conflicts": ["invalid branch name"], "plans_updated": false}'
        exit 1
    fi

    # Switch to target branch
    if [[ "$current_branch" != "$target_branch" ]]; then
        if [[ "$dry_run" == "false" ]]; then
            git switch "$target_branch"
        else
            log_info "[DRY-RUN] git switch $target_branch"
        fi
    fi

    # Execute merge
    local merge_status="merged"
    local conflicts=()

    if [[ "$squash" == "true" ]]; then
        if ! do_squash_merge "$worktree" "$dry_run"; then
            merge_status="conflict"
            conflicts+=("squash merge failed")
        fi
    else
        if ! do_cherry_pick "$commit_hash" "$dry_run"; then
            merge_status="conflict"
            conflicts+=("cherry-pick failed")
        fi
    fi

    # Update Plans.md
    local plans_updated=false
    if [[ "$merge_status" == "merged" ]] && [[ "$dry_run" == "false" ]]; then
        # Infer task ID from worktree name (worker-1 → task-1, etc.)
        local worker_id
        worker_id=$(basename "$worktree" | sed 's/worker-//')

        if update_plans "task-$worker_id\|Task $worker_id" 2>/dev/null; then
            plans_updated=true
        fi
    fi

    # Output results
    local conflicts_json
    conflicts_json=$(printf '%s\n' "${conflicts[@]:-}" | jq -R -s -c 'split("\n") | map(select(length > 0))')

    local result
    result=$(jq -n \
        --arg status "$merge_status" \
        --arg commit_hash "$commit_hash" \
        --argjson conflicts "$conflicts_json" \
        --argjson plans_updated "$plans_updated" \
        '{
            status: $status,
            commit_hash: $commit_hash,
            conflicts: $conflicts,
            plans_updated: $plans_updated
        }')

    echo "$result"

    # Return to original branch (after merge)
    if [[ "$dry_run" == "false" ]] && [[ -n "$current_branch" ]] && [[ "$current_branch" != "$target_branch" ]]; then
        git switch "$current_branch" 2>/dev/null || log_warn "Could not return to original branch: $current_branch"
    fi

    # Exit code
    if [[ "$merge_status" == "merged" ]]; then
        exit 0
    else
        exit 1
    fi
}

# Usage
usage() {
    cat << EOF
Usage: $0 --worktree PATH [OPTIONS]

Options:
  --worktree PATH         Worker worktree path (required)
  --target-branch BRANCH  Target branch for merge (default: main)
  --squash                Use squash merge
  --dry-run               Check only without actually merging
  --force                 Force merge even without passing quality gate (not recommended)
  -h, --help              Show help

Examples:
  $0 --worktree ../worktrees/worker-1
  $0 --worktree ../worktrees/worker-1 --target-branch develop
  $0 --worktree ../worktrees/worker-1 --squash
  $0 --worktree ../worktrees/worker-1 --dry-run
  $0 --worktree ../worktrees/worker-1 --force  # Skip quality gate (use with caution)
EOF
}

main "$@"
