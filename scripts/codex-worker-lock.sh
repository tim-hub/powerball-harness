#!/usr/bin/env bash
#
# codex-worker-lock.sh
# タスク所有権・ロック機構
#
# Usage:
#   ./scripts/codex-worker-lock.sh acquire --path PATH --worker WORKER_ID
#   ./scripts/codex-worker-lock.sh release --path PATH --worker WORKER_ID
#   ./scripts/codex-worker-lock.sh heartbeat --path PATH --worker WORKER_ID
#   ./scripts/codex-worker-lock.sh check --path PATH
#   ./scripts/codex-worker-lock.sh cleanup
#

set -euo pipefail

# 色定義
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 設定
TTL_MINUTES=30
HEARTBEAT_MINUTES=10
LOCK_DIR=".claude/state/locks"
LOCK_LOG=".claude/state/locks.log"

# ヘルパー関数
log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# 依存チェック
check_dependencies() {
    for cmd in jq shasum date; do
        if ! command -v "$cmd" &> /dev/null; then
            log_error "必須コマンドが見つかりません: $cmd"
            exit 1
        fi
    done
}

# パス正規化
normalize_path() {
    local path="$1"
    # ./ 除去、/ 統一
    path="${path#./}"
    path="${path//\\//}"
    printf '%s' "$path"
}

# ロックキー生成（SHA256 先頭8文字）
generate_lock_key() {
    local path="$1"
    local normalized
    normalized=$(normalize_path "$path")
    printf '%s' "$normalized" | shasum -a 256 | cut -c1-8
}

# ロックファイルパス取得
get_lock_file() {
    local path="$1"
    local key
    key=$(generate_lock_key "$path")
    printf '%s/%s.lock.json' "$LOCK_DIR" "$key"
}

# ISO8601 UTC 現在時刻
now_utc() {
    date -u +%Y-%m-%dT%H:%M:%SZ
}

# ISO8601 UTC をエポック秒に変換（macOS/Linux 互換）
parse_utc_to_epoch() {
    local ts="$1"
    # Z サフィックスを除去して -u フラグで UTC として解釈
    local ts_no_z="${ts%Z}"
    date -u -j -f "%Y-%m-%dT%H:%M:%S" "$ts_no_z" "+%s" 2>/dev/null || \
    date -u -d "$ts" "+%s" 2>/dev/null || \
    echo 0
}

# ログ記録
log_event() {
    local event="$1"
    local path="$2"
    local worker="$3"
    mkdir -p "$(dirname "$LOCK_LOG")"
    printf '%s\t%s\t%s\t%s\n' "$(now_utc)" "$event" "$path" "$worker" >> "$LOCK_LOG"
}

# ロック取得
cmd_acquire() {
    local path=""
    local worker=""

    while [[ $# -gt 0 ]]; do
        case "$1" in
            --path)
                if [[ -z "${2:-}" ]]; then
                    log_error "--path には値が必要です"
                    exit 1
                fi
                path="$2"; shift 2 ;;
            --worker)
                if [[ -z "${2:-}" ]]; then
                    log_error "--worker には値が必要です"
                    exit 1
                fi
                worker="$2"; shift 2 ;;
            *) log_error "Unknown option: $1"; exit 1 ;;
        esac
    done

    if [[ -z "$path" ]] || [[ -z "$worker" ]]; then
        log_error "--path と --worker は必須です"
        exit 1
    fi

    mkdir -p "$LOCK_DIR"

    local lock_file
    lock_file=$(get_lock_file "$path")
    local normalized_path
    normalized_path=$(normalize_path "$path")

    # 既存ロックチェック
    if [[ -f "$lock_file" ]]; then
        local existing_worker
        local heartbeat
        existing_worker=$(jq -r '.worker' "$lock_file")
        heartbeat=$(jq -r '.heartbeat' "$lock_file")

        # TTL チェック
        local heartbeat_epoch
        local now_epoch
        local ttl_seconds=$((TTL_MINUTES * 60))

        heartbeat_epoch=$(parse_utc_to_epoch "$heartbeat")
        now_epoch=$(date "+%s")

        if (( now_epoch - heartbeat_epoch > ttl_seconds )); then
            log_warn "TTL 超過: 既存ロックを解放 (worker=$existing_worker)"
            log_event "expired" "$normalized_path" "$existing_worker"
            rm -f "$lock_file"
        else
            log_error "ロック取得失敗: $normalized_path は $existing_worker がロック中"
            exit 1
        fi
    fi

    # 新規ロック作成（原子的作成）
    local now
    now=$(now_utc)

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

    # ln で原子的配置（既存ファイルがあれば失敗）
    if ! ln "$tmp_file" "$lock_file" 2>/dev/null; then
        rm -f "$tmp_file"
        log_error "ロック取得失敗: $normalized_path は他の Worker がロック中（競合）"
        exit 1
    fi

    rm -f "$tmp_file"
    log_event "acquire" "$normalized_path" "$worker"
    log_info "ロック取得: $normalized_path (worker=$worker)"
}

# ロック解放
cmd_release() {
    local path=""
    local worker=""

    while [[ $# -gt 0 ]]; do
        case "$1" in
            --path)
                if [[ -z "${2:-}" ]]; then
                    log_error "--path には値が必要です"
                    exit 1
                fi
                path="$2"; shift 2 ;;
            --worker)
                if [[ -z "${2:-}" ]]; then
                    log_error "--worker には値が必要です"
                    exit 1
                fi
                worker="$2"; shift 2 ;;
            *) log_error "Unknown option: $1"; exit 1 ;;
        esac
    done

    if [[ -z "$path" ]] || [[ -z "$worker" ]]; then
        log_error "--path と --worker は必須です"
        exit 1
    fi

    local lock_file
    lock_file=$(get_lock_file "$path")
    local normalized_path
    normalized_path=$(normalize_path "$path")

    if [[ ! -f "$lock_file" ]]; then
        log_warn "ロックが存在しません: $normalized_path"
        exit 0
    fi

    local existing_worker
    existing_worker=$(jq -r '.worker' "$lock_file")

    if [[ "$existing_worker" != "$worker" ]]; then
        log_error "ロック解放失敗: $normalized_path は $existing_worker のロックです"
        exit 1
    fi

    rm -f "$lock_file"
    log_event "release" "$normalized_path" "$worker"
    log_info "ロック解放: $normalized_path (worker=$worker)"
}

# heartbeat 更新
cmd_heartbeat() {
    local path=""
    local worker=""

    while [[ $# -gt 0 ]]; do
        case "$1" in
            --path)
                if [[ -z "${2:-}" ]]; then
                    log_error "--path には値が必要です"
                    exit 1
                fi
                path="$2"; shift 2 ;;
            --worker)
                if [[ -z "${2:-}" ]]; then
                    log_error "--worker には値が必要です"
                    exit 1
                fi
                worker="$2"; shift 2 ;;
            *) log_error "Unknown option: $1"; exit 1 ;;
        esac
    done

    if [[ -z "$path" ]] || [[ -z "$worker" ]]; then
        log_error "--path と --worker は必須です"
        exit 1
    fi

    local lock_file
    lock_file=$(get_lock_file "$path")
    local normalized_path
    normalized_path=$(normalize_path "$path")

    if [[ ! -f "$lock_file" ]]; then
        log_error "ロックが存在しません: $normalized_path"
        exit 1
    fi

    local existing_worker
    existing_worker=$(jq -r '.worker' "$lock_file")

    if [[ "$existing_worker" != "$worker" ]]; then
        log_error "heartbeat 更新失敗: $normalized_path は $existing_worker のロックです"
        exit 1
    fi

    local now
    now=$(now_utc)

    jq --arg heartbeat "$now" '.heartbeat = $heartbeat' "$lock_file" > "$lock_file.tmp"
    mv "$lock_file.tmp" "$lock_file"

    log_info "heartbeat 更新: $normalized_path (worker=$worker)"
}

# ロック状態確認
cmd_check() {
    local path=""

    while [[ $# -gt 0 ]]; do
        case "$1" in
            --path)
                if [[ -z "${2:-}" ]]; then
                    log_error "--path には値が必要です"
                    exit 1
                fi
                path="$2"; shift 2 ;;
            *) log_error "Unknown option: $1"; exit 1 ;;
        esac
    done

    if [[ -z "$path" ]]; then
        log_error "--path は必須です"
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

    # TTL チェック
    local heartbeat_epoch
    local now_epoch
    local ttl_seconds=$((TTL_MINUTES * 60))

    heartbeat_epoch=$(parse_utc_to_epoch "$heartbeat")
    now_epoch=$(date "+%s")

    if (( now_epoch - heartbeat_epoch > ttl_seconds )); then
        # TTL 超過: 読み取り専用（削除は acquire/cleanup で行う）
        echo '{"locked": false, "expired": true, "hint": "run cleanup or acquire to release"}'
    else
        jq -c '. + {locked: true}' "$lock_file"
    fi
}

# 期限切れロックのクリーンアップ
cmd_cleanup() {
    mkdir -p "$LOCK_DIR"

    local cleaned=0
    local now_epoch
    now_epoch=$(date "+%s")
    local ttl_seconds=$((TTL_MINUTES * 60))

    for lock_file in "$LOCK_DIR"/*.lock.json; do
        [[ -f "$lock_file" ]] || continue

        local heartbeat
        local worker
        local path
        heartbeat=$(jq -r '.heartbeat' "$lock_file")
        worker=$(jq -r '.worker' "$lock_file")
        path=$(jq -r '.path' "$lock_file")

        local heartbeat_epoch
        heartbeat_epoch=$(parse_utc_to_epoch "$heartbeat")

        if (( now_epoch - heartbeat_epoch > ttl_seconds )); then
            log_warn "TTL 超過: $path (worker=$worker)"
            log_event "expired" "$path" "$worker"
            rm -f "$lock_file"
            cleaned=$((cleaned + 1))
        fi
    done

    log_info "クリーンアップ完了: $cleaned 件のロックを解放"
}

# 使用方法
usage() {
    cat << EOF
Usage: $0 COMMAND [OPTIONS]

Commands:
  acquire   --path PATH --worker WORKER_ID   ロック取得
  release   --path PATH --worker WORKER_ID   ロック解放
  heartbeat --path PATH --worker WORKER_ID   heartbeat 更新
  check     --path PATH                      ロック状態確認
  cleanup                                    期限切れロックのクリーンアップ

Options:
  --path PATH       対象ファイルパス
  --worker WORKER_ID Worker 識別子

Settings:
  TTL: $TTL_MINUTES 分
  Heartbeat 間隔: $HEARTBEAT_MINUTES 分
  ロックディレクトリ: $LOCK_DIR

Examples:
  $0 acquire --path src/auth/login.ts --worker worker-1
  $0 heartbeat --path src/auth/login.ts --worker worker-1
  $0 release --path src/auth/login.ts --worker worker-1
  $0 check --path src/auth/login.ts
  $0 cleanup
EOF
}

# メイン処理
main() {
    check_dependencies

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
