#!/usr/bin/env bash
#
# codex-worker-quality-gate.sh
# Orchestrator による Worker 成果物の品質検証
#
# Usage:
#   ./scripts/codex-worker-quality-gate.sh --worktree PATH [--skip-gate GATE --reason TEXT]
#

set -euo pipefail

# 色定義
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 設定
GATE_SKIP_LOG=".claude/state/gate-skips.log"
AGENTS_SUMMARY_PATTERN='AGENTS_SUMMARY:[[:space:]]*(.+?)[[:space:]]*\|[[:space:]]*HASH:([A-Fa-f0-9]{8})'

# ヘルパー関数
log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_gate() { echo -e "${BLUE}[GATE]${NC} $1"; }

# 依存チェック
check_dependencies() {
    for cmd in jq shasum git npm; do
        if ! command -v "$cmd" &> /dev/null; then
            log_error "必須コマンドが見つかりません: $cmd"
            exit 1
        fi
    done
}

# ISO8601 UTC 現在時刻
now_utc() {
    date -u +%Y-%m-%dT%H:%M:%SZ
}

# スキップログ記録
log_skip() {
    local gate="$1"
    local reason="$2"
    local user="${USER:-unknown}"

    mkdir -p "$(dirname "$GATE_SKIP_LOG")"
    printf '%s\t%s\t%s\t%s\n' "$(now_utc)" "$gate" "$reason" "$user" >> "$GATE_SKIP_LOG"
}

# Gate 1: 証跡検証
gate_evidence() {
    local worktree="$1"
    local output_file="$2"

    log_gate "Gate 1: 証跡検証 (evidence)"

    # AGENTS.md のハッシュ計算
    local agents_file="$worktree/AGENTS.md"
    if [[ ! -f "$agents_file" ]]; then
        echo '{"status": "critical", "details": "AGENTS.md not found"}' > "$output_file"
        return 2  # Critical: AGENTS.md 必須
    fi

    # ハッシュ計算（engine と完全同一: BOM除去、CR除去、SHA256先頭8文字）
    local expected_hash
    expected_hash=$(sed '1s/^\xEF\xBB\xBF//' "$agents_file" | tr -d '\r' | shasum -a 256 | cut -c1-8)

    # Worker 出力から AGENTS_SUMMARY を検索
    # 1. Worker 出力ログ（優先）
    # 2. 最新のコミットメッセージ（フォールバック）
    local worker_output_log="$worktree/.claude/state/worker-output.log"
    local search_content=""
    local found_in_log=false

    if [[ -f "$worker_output_log" ]]; then
        search_content=$(cat "$worker_output_log")
        if echo "$search_content" | grep -qE "$AGENTS_SUMMARY_PATTERN"; then
            found_in_log=true
        fi
    fi

    # ログで未検出の場合はコミットメッセージをフォールバック検索
    if [[ "$found_in_log" == "false" ]]; then
        local commit_msg
        commit_msg=$(cd "$worktree" && git log -1 --pretty=%B 2>/dev/null || echo "")
        if echo "$commit_msg" | grep -qE "$AGENTS_SUMMARY_PATTERN"; then
            search_content="$commit_msg"
        fi
    fi

    # パターンマッチ
    if echo "$search_content" | grep -qE "$AGENTS_SUMMARY_PATTERN"; then
        local found_hash
        found_hash=$(echo "$search_content" | grep -oE 'HASH:[A-Fa-f0-9]{8}' | head -1 | cut -d: -f2)

        if [[ "${found_hash,,}" == "${expected_hash,,}" ]]; then
            echo '{"status": "passed", "details": "証跡確認OK"}' > "$output_file"
            return 0
        else
            echo "{\"status\": \"failed\", \"details\": \"ハッシュ不一致: expected=$expected_hash, found=$found_hash\"}" > "$output_file"
            return 1  # High: ハッシュ不一致は再試行可能
        fi
    else
        echo '{"status": "critical", "details": "AGENTS_SUMMARY 証跡が見つかりません（即失敗）"}' > "$output_file"
        return 2  # Critical: 証跡欠落は即失敗
    fi
}

# Gate 2: 構造チェック
gate_structure() {
    local worktree="$1"
    local output_file="$2"

    log_gate "Gate 2: 構造チェック (structure)"

    local lint_result=0
    local type_result=0
    local details=""

    # package.json が存在しない場合はスキップ扱い
    if [[ ! -f "$worktree/package.json" ]]; then
        echo '{"status": "passed", "details": "package.json なし（スキップ）"}' > "$output_file"
        return 0
    fi

    # jq で scripts キーを正確に判定
    # lint チェック
    if jq -e '.scripts.lint' "$worktree/package.json" > /dev/null 2>&1; then
        if ! (cd "$worktree" && npm run lint --silent 2>&1); then
            lint_result=1
            details="lint エラー"
        fi
    fi

    # type-check
    if jq -e '.scripts["type-check"]' "$worktree/package.json" > /dev/null 2>&1; then
        if ! (cd "$worktree" && npm run type-check --silent 2>&1); then
            type_result=1
            details="${details:+$details, }type エラー"
        fi
    fi

    if [[ $lint_result -eq 0 ]] && [[ $type_result -eq 0 ]]; then
        echo '{"status": "passed", "details": "構造チェックOK"}' > "$output_file"
        return 0
    else
        echo "{\"status\": \"failed\", \"details\": \"$details\"}" > "$output_file"
        return 1
    fi
}

# Gate 3: テスト
gate_test() {
    local worktree="$1"
    local output_file="$2"

    log_gate "Gate 3: テスト (test)"

    # 改ざん検出パターン（追加行用）
    local tamper_add_patterns=(
        'it\.skip\s*\('
        'test\.skip\s*\('
        'describe\.skip\s*\('
        'eslint-disable'
        'expect\([^)]*\)\.toBe\((true|false|null|undefined|0|1)\)'
    )

    # 削除行用パターン（アサーション削除検出）
    local tamper_remove_patterns=(
        'expect\s*\('
        'assert\s*\('
        '\.should\s*\('
        '\.to\.\w+'
    )

    # 差分ベースの改ざん検出（main ブランチからの全変更）
    local merge_base
    merge_base=$(cd "$worktree" && git merge-base HEAD origin/main 2>/dev/null || git rev-parse HEAD~1 2>/dev/null || echo "")

    if [[ -n "$merge_base" ]]; then
        # 追加行の検出（+で始まる行、+++ ヘッダのみ除外）
        local added_lines
        added_lines=$(cd "$worktree" && git diff "$merge_base"..HEAD --unified=0 -- '*.ts' '*.tsx' '*.js' '*.jsx' '*.spec.*' '*.test.*' 2>/dev/null | grep '^+' | grep -v '^+++ ' || echo "")

        for pattern in "${tamper_add_patterns[@]}"; do
            if echo "$added_lines" | grep -qE "$pattern"; then
                echo "{\"status\": \"critical\", \"details\": \"改ざん検出: 追加行に '$pattern' パターン\"}" > "$output_file"
                return 2  # Critical: 改ざん検出
            fi
        done

        # 削除行の検出（-で始まる行）- テストファイルのみ
        local removed_lines
        removed_lines=$(cd "$worktree" && git diff "$merge_base"..HEAD --unified=0 -- '*.spec.*' '*.test.*' 2>/dev/null | grep '^-' | grep -v '^---' || echo "")

        for pattern in "${tamper_remove_patterns[@]}"; do
            local removed_count
            removed_count=$(echo "$removed_lines" | grep -cE "$pattern" 2>/dev/null || echo 0)

            if [[ "$removed_count" -gt 2 ]]; then
                echo "{\"status\": \"critical\", \"details\": \"改ざん検出: テストから '$pattern' が $removed_count 件削除\"}" > "$output_file"
                return 2  # Critical: アサーション大量削除
            fi
        done
    fi

    # テスト実行
    if [[ -f "$worktree/package.json" ]]; then
        if jq -e '.scripts.test' "$worktree/package.json" > /dev/null 2>&1; then
            if ! (cd "$worktree" && npm test --silent 2>&1); then
                echo '{"status": "failed", "details": "テスト失敗"}' > "$output_file"
                return 1
            fi
        fi
    fi

    echo '{"status": "passed", "details": "テストOK"}' > "$output_file"
    return 0
}

# メイン処理
main() {
    check_dependencies

    local worktree=""
    local skip_gates=()
    local skip_reason=""

    while [[ $# -gt 0 ]]; do
        case "$1" in
            --worktree)
                if [[ -z "${2:-}" ]]; then
                    log_error "--worktree には値が必要です"
                    exit 1
                fi
                worktree="$2"; shift 2 ;;
            --skip-gate)
                if [[ -z "${2:-}" ]]; then
                    log_error "--skip-gate には値が必要です"
                    exit 1
                fi
                skip_gates+=("$2"); shift 2 ;;
            --reason)
                if [[ -z "${2:-}" ]]; then
                    log_error "--reason には値が必要です"
                    exit 1
                fi
                skip_reason="$2"; shift 2 ;;
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

    # スキップ時は理由必須
    if [[ ${#skip_gates[@]} -gt 0 ]] && [[ -z "$skip_reason" ]]; then
        log_error "--skip-gate 使用時は --reason が必須です"
        exit 1
    fi

    # 一時ディレクトリ
    local tmp_dir
    tmp_dir=$(mktemp -d)
    trap 'rm -rf "$tmp_dir"' EXIT

    # 結果格納
    local overall_status="passed"
    local gates_json="{}"
    local skipped_json="[]"
    local errors_json="[]"

    # Gate 1: 証跡検証
    if [[ " ${skip_gates[*]} " =~ " evidence " ]]; then
        log_warn "Gate 1 (evidence) をスキップ"
        log_skip "evidence" "$skip_reason"
        skipped_json=$(echo "$skipped_json" | jq '. + ["evidence"]')
        gates_json=$(echo "$gates_json" | jq --arg reason "$skip_reason" '.evidence = {"status": "skipped", "details": $reason}')
    else
        local evidence_exit_code=0
        gate_evidence "$worktree" "$tmp_dir/evidence.json" || evidence_exit_code=$?

        if [[ $evidence_exit_code -eq 0 ]]; then
            gates_json=$(echo "$gates_json" | jq --slurpfile g "$tmp_dir/evidence.json" '.evidence = $g[0]')
        elif [[ $evidence_exit_code -eq 2 ]]; then
            # Critical: 証跡欠落
            overall_status="critical"
            gates_json=$(echo "$gates_json" | jq --slurpfile g "$tmp_dir/evidence.json" '.evidence = $g[0]')
            errors_json=$(echo "$errors_json" | jq '. + ["CRITICAL: 証跡欠落"]')
        else
            overall_status="failed"
            gates_json=$(echo "$gates_json" | jq --slurpfile g "$tmp_dir/evidence.json" '.evidence = $g[0]')
            errors_json=$(echo "$errors_json" | jq '. + ["Gate 1 failed: ハッシュ不一致"]')
        fi
    fi

    # Gate 2: 構造チェック
    if [[ " ${skip_gates[*]} " =~ " structure " ]]; then
        log_warn "Gate 2 (structure) をスキップ"
        log_skip "structure" "$skip_reason"
        skipped_json=$(echo "$skipped_json" | jq '. + ["structure"]')
        gates_json=$(echo "$gates_json" | jq --arg reason "$skip_reason" '.structure = {"status": "skipped", "details": $reason}')
    else
        if gate_structure "$worktree" "$tmp_dir/structure.json"; then
            gates_json=$(echo "$gates_json" | jq --slurpfile g "$tmp_dir/structure.json" '.structure = $g[0]')
        else
            overall_status="failed"
            gates_json=$(echo "$gates_json" | jq --slurpfile g "$tmp_dir/structure.json" '.structure = $g[0]')
            errors_json=$(echo "$errors_json" | jq '. + ["Gate 2 failed"]')
        fi
    fi

    # Gate 3: テスト
    if [[ " ${skip_gates[*]} " =~ " test " ]]; then
        log_warn "Gate 3 (test) をスキップ"
        log_skip "test" "$skip_reason"
        skipped_json=$(echo "$skipped_json" | jq '. + ["test"]')
        gates_json=$(echo "$gates_json" | jq --arg reason "$skip_reason" '.test = {"status": "skipped", "details": $reason}')
    else
        local test_exit_code=0
        gate_test "$worktree" "$tmp_dir/test.json" || test_exit_code=$?

        if [[ $test_exit_code -eq 0 ]]; then
            gates_json=$(echo "$gates_json" | jq --slurpfile g "$tmp_dir/test.json" '.test = $g[0]')
        elif [[ $test_exit_code -eq 2 ]]; then
            # Critical: 改ざん検出
            overall_status="critical"
            gates_json=$(echo "$gates_json" | jq --slurpfile g "$tmp_dir/test.json" '.test = $g[0]')
            errors_json=$(echo "$errors_json" | jq '. + ["CRITICAL: 改ざん検出"]')
        else
            overall_status="failed"
            gates_json=$(echo "$gates_json" | jq --slurpfile g "$tmp_dir/test.json" '.test = $g[0]')
            errors_json=$(echo "$errors_json" | jq '. + ["Gate 3 failed"]')
        fi
    fi

    # 最終結果出力
    local result
    result=$(jq -n \
        --arg status "$overall_status" \
        --argjson gates "$gates_json" \
        --argjson skipped "$skipped_json" \
        --argjson errors "$errors_json" \
        '{
            status: $status,
            gates: $gates,
            skipped: $skipped,
            errors: $errors
        }')

    echo "$result"

    # 終了コード
    case "$overall_status" in
        passed) exit 0 ;;
        failed) exit 1 ;;
        critical) exit 2 ;;
    esac
}

# 使用方法
usage() {
    cat << EOF
Usage: $0 --worktree PATH [OPTIONS]

Options:
  --worktree PATH       検査対象の worktree（必須）
  --skip-gate GATE      特定ゲートをスキップ (evidence, structure, test)
  --reason TEXT         スキップ理由（--skip-gate と併用、必須）
  -h, --help            ヘルプ表示

Gates:
  evidence   - AGENTS_SUMMARY 証跡検証
  structure  - lint, type-check
  test       - テスト実行、改ざん検出

Examples:
  $0 --worktree ../worktrees/worker-1
  $0 --worktree ../worktrees/worker-1 --skip-gate test --reason "テスト環境未構築"
EOF
}

main "$@"
