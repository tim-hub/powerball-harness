#!/usr/bin/env bash
#
# codex-worker-merge.sh
# Worker 成果物のマージ統合
#
# Usage:
#   ./scripts/codex-worker-merge.sh --worktree PATH --target-branch BRANCH [--squash] [--dry-run]
#

set -euo pipefail

# 色定義
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# ヘルパー関数
log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_merge() { echo -e "${BLUE}[MERGE]${NC} $1"; }

# 依存チェック
check_dependencies() {
    for cmd in git jq sed; do
        if ! command -v "$cmd" &> /dev/null; then
            log_error "必須コマンドが見つかりません: $cmd"
            exit 1
        fi
    done
}

# Plans.md 更新
update_plans() {
    local task_pattern="$1"
    local plans_file="Plans.md"

    if [[ ! -f "$plans_file" ]]; then
        log_warn "Plans.md が見つかりません"
        return 1
    fi

    # cc:WIP → cc:done, [ ] → [x]
    if grep -q "$task_pattern" "$plans_file"; then
        sed -i.bak "s/\(.*$task_pattern.*\)cc:WIP/\1cc:done/" "$plans_file"
        sed -i.bak "s/\(.*$task_pattern.*\)\[ \]/\1[x]/" "$plans_file"
        rm -f "$plans_file.bak"
        log_info "Plans.md 更新: $task_pattern → cc:done"
        return 0
    fi

    return 1
}

# cherry-pick マージ
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
        # 競合発生
        git cherry-pick --abort 2>/dev/null || true
        return 1
    fi
}

# squash マージ
do_squash_merge() {
    local worktree="$1"
    local dry_run="$2"

    # worktree のブランチ名を取得
    local branch_name
    branch_name=$(cd "$worktree" && git branch --show-current)

    log_merge "squash merge: $branch_name"

    if [[ "$dry_run" == "true" ]]; then
        log_info "[DRY-RUN] git merge --squash $branch_name"
        return 0
    fi

    if git merge --squash "$branch_name" 2>/dev/null; then
        git commit -m "feat: Worker 成果物のマージ ($branch_name)"
        return 0
    else
        git merge --abort 2>/dev/null || true
        return 1
    fi
}

# メイン処理
main() {
    check_dependencies

    local worktree=""
    local target_branch="main"
    local squash=false
    local dry_run=false

    while [[ $# -gt 0 ]]; do
        case "$1" in
            --worktree)
                if [[ -z "${2:-}" ]]; then
                    log_error "--worktree には値が必要です"
                    exit 1
                fi
                worktree="$2"; shift 2 ;;
            --target-branch)
                if [[ -z "${2:-}" ]]; then
                    log_error "--target-branch には値が必要です"
                    exit 1
                fi
                target_branch="$2"; shift 2 ;;
            --squash) squash=true; shift ;;
            --dry-run) dry_run=true; shift ;;
            -h|--help) usage; exit 0 ;;
            *) log_error "Unknown option: $1"; exit 1 ;;
        esac
    done

    # 必須パラメータチェック
    if [[ -z "$worktree" ]]; then
        log_error "--worktree は必須です"
        exit 1
    fi

    if [[ ! -d "$worktree" ]]; then
        log_error "Worktree が存在しません: $worktree"
        exit 1
    fi

    # worktree の最新コミット取得
    local commit_hash
    commit_hash=$(cd "$worktree" && git log -1 --format="%H")

    if [[ -z "$commit_hash" ]]; then
        log_error "コミットが見つかりません"
        echo '{"status": "failed", "commit_hash": null, "conflicts": [], "plans_updated": false}'
        exit 1
    fi

    log_info "Worker コミット: $commit_hash"

    # 現在のブランチを確認
    local current_branch
    current_branch=$(git branch --show-current)

    # ターゲットブランチの検証
    if ! git check-ref-format --branch "$target_branch" 2>/dev/null; then
        log_error "無効なブランチ名: $target_branch"
        echo '{"status": "failed", "commit_hash": null, "conflicts": ["invalid branch name"], "plans_updated": false}'
        exit 1
    fi

    # ターゲットブランチに切り替え
    if [[ "$current_branch" != "$target_branch" ]]; then
        if [[ "$dry_run" == "false" ]]; then
            git checkout -- "$target_branch"
        else
            log_info "[DRY-RUN] git checkout -- $target_branch"
        fi
    fi

    # マージ実行
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

    # Plans.md 更新
    local plans_updated=false
    if [[ "$merge_status" == "merged" ]] && [[ "$dry_run" == "false" ]]; then
        # worktree 名からタスク ID を推測（worker-1 → task-1 など）
        local worker_id
        worker_id=$(basename "$worktree" | sed 's/worker-//')

        if update_plans "task-$worker_id\|Task $worker_id" 2>/dev/null; then
            plans_updated=true
        fi
    fi

    # 結果出力
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

    # 終了コード
    if [[ "$merge_status" == "merged" ]]; then
        exit 0
    else
        exit 1
    fi
}

# 使用方法
usage() {
    cat << EOF
Usage: $0 --worktree PATH [OPTIONS]

Options:
  --worktree PATH         Worker の worktree パス（必須）
  --target-branch BRANCH  マージ先ブランチ（デフォルト: main）
  --squash                squash merge を使用
  --dry-run               実際にマージせず確認のみ
  -h, --help              ヘルプ表示

Examples:
  $0 --worktree ../worktrees/worker-1
  $0 --worktree ../worktrees/worker-1 --target-branch develop
  $0 --worktree ../worktrees/worker-1 --squash
  $0 --worktree ../worktrees/worker-1 --dry-run
EOF
}

main "$@"
