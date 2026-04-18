#!/usr/bin/env bash
#
# reenter-worktree.sh
# CC 2.1.105 で追加された EnterWorktree path 引数を使い、既存 worktree へ再入するための
# エージェント spawn 用ヘルパー。Breezing の修正ループ（REQUEST_CHANGES 後の amend）で
# Worker が一度離脱した worktree に再度入る場合に使用する。
#
# Usage: ./scripts/reenter-worktree.sh --path <worktree-path> [--task-id <id>]
#
# 前提:
#   - git worktree list に <worktree-path> が存在すること
#   - CC 2.1.105 以上 (EnterWorktree path 引数サポート)
#
# 出力 (JSON):
#   {"decision":"approve","worktree_path":"<path>","task_id":"<id>"}
#   または
#   {"decision":"deny","reason":"<message>"}

set -euo pipefail

WORKTREE_PATH=""
TASK_ID=""

canonicalize_path() {
    local target="$1"
    if [[ -d "$target" ]]; then
        (
            cd "$target" >/dev/null 2>&1 && pwd -P
        )
        return
    fi
    printf '%s\n' "$target"
}

usage() {
    echo "Usage: $0 --path <worktree-path> [--task-id <id>]" >&2
    exit 1
}

# 引数パース
while [[ $# -gt 0 ]]; do
    case "$1" in
        --path)
            [[ $# -lt 2 ]] && { echo "Error: --path requires a value" >&2; usage; }
            WORKTREE_PATH="$2"
            shift 2
            ;;
        --task-id)
            [[ $# -lt 2 ]] && { echo "Error: --task-id requires a value" >&2; usage; }
            TASK_ID="$2"
            shift 2
            ;;
        -h|--help)
            usage
            ;;
        *)
            echo "Error: Unknown argument: $1" >&2
            usage
            ;;
    esac
done

if [[ -z "$WORKTREE_PATH" ]]; then
    echo "Error: --path is required" >&2
    usage
fi

# worktree パスの存在確認
if [[ ! -d "$WORKTREE_PATH" ]]; then
    printf '{"decision":"deny","reason":"worktree path does not exist: %s"}\n' "$WORKTREE_PATH"
    exit 1
fi

CANONICAL_WORKTREE_PATH="$(canonicalize_path "$WORKTREE_PATH")"

# git worktree list で登録確認
if ! git worktree list --porcelain 2>/dev/null | awk '/^worktree / {sub(/^worktree /, ""); print}' | while IFS= read -r listed_path; do
    [[ "$(canonicalize_path "$listed_path")" == "$CANONICAL_WORKTREE_PATH" ]] && exit 0
done; then
    printf '{"decision":"deny","reason":"path is not a registered git worktree: %s"}\n' "$WORKTREE_PATH"
    exit 1
fi

# worktree 内の .claude/state/worktree-info.json を確認（Breezing Worker が作成）
WORKTREE_INFO="${WORKTREE_PATH}/.claude/state/worktree-info.json"
if [[ -f "$WORKTREE_INFO" ]] && command -v jq >/dev/null 2>&1; then
    REGISTERED_WORKER_ID="$(jq -r '.worker_id // ""' "$WORKTREE_INFO" 2>/dev/null || echo "")"
else
    REGISTERED_WORKER_ID=""
fi

print_guidance() {
    cat >&2 <<EOF
# EnterWorktree path 再入確認

## worktree 情報
- path:      $WORKTREE_PATH
- task_id:   ${TASK_ID:-"(未指定)"}
- worker_id: ${REGISTERED_WORKER_ID:-"(取得不可)"}

## CC 2.1.105 以降: エージェント定義での利用方法

Agent tool の isolation フィールドで既存 worktree を指定するには、
EnterWorktree の path パラメータを以下のように渡す:

  isolation: "worktree"
  worktreePath: "$WORKTREE_PATH"

Harness breezing の Lead が SendMessage でワーカーを resume する際、
同じ worktree に再入するのに使用する。

## スクリプトからの検証
- git worktree list: OK (path 登録確認済み)
- ディレクトリ存在: OK
EOF
}

print_guidance

# JSON 出力
if command -v jq >/dev/null 2>&1; then
    jq -nc \
        --arg decision "approve" \
        --arg worktree_path "$WORKTREE_PATH" \
        --arg task_id "${TASK_ID:-""}" \
        --arg worker_id "${REGISTERED_WORKER_ID:-""}" \
        '{"decision":$decision,"worktree_path":$worktree_path,"task_id":$task_id,"worker_id":$worker_id}'
else
    printf '{"decision":"approve","worktree_path":"%s","task_id":"%s"}\n' \
        "${WORKTREE_PATH//\"/\\\"}" \
        "${TASK_ID//\"/\\\"}"
fi
