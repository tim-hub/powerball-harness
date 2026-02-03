#!/usr/bin/env bash
#
# codex-worker-engine.sh
# Codex Worker 実行エンジン
#
# Usage: ./scripts/codex-worker-engine.sh --task "タスク内容" [--worktree PATH] [--dry-run]
#

set -euo pipefail

# 色定義
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# デフォルト設定
MAX_RETRIES=3
APPROVAL_POLICY="never"
SANDBOX="workspace-write"

# グローバル変数
TASK=""
WORKTREE_PATH=""
DRY_RUN=false
PROJECT_ROOT=""
AGENTS_HASH=""

# 依存コマンドチェック
check_dependencies() {
    local missing=()

    for cmd in jq shasum git; do
        if ! command -v "$cmd" &> /dev/null; then
            missing+=("$cmd")
        fi
    done

    if [[ ${#missing[@]} -gt 0 ]]; then
        log_error "必須コマンドが見つかりません: ${missing[*]}"
        exit 1
    fi
}

# ヘルパー関数
log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_step() { echo -e "${BLUE}[STEP]${NC} $1"; }

# 使用方法
usage() {
    cat << EOF
Usage: $0 --task "タスク内容" [OPTIONS]

Options:
  --task TEXT       実行するタスク内容（必須）
  --worktree PATH   Worktree パス（省略時はカレントディレクトリ）
  --dry-run         ドライラン（実行せず内容を表示）
  -h, --help        ヘルプ表示

Examples:
  $0 --task "ログイン機能を実装して"
  $0 --task "APIエンドポイントを追加" --worktree ../worktree-task-1
  $0 --task "テストを修正" --dry-run
EOF
}

# 引数解析
parse_args() {
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --task)
                if [[ -z "${2:-}" ]]; then
                    log_error "--task には値が必要です"
                    exit 1
                fi
                TASK="$2"
                shift 2
                ;;
            --worktree)
                if [[ -z "${2:-}" ]]; then
                    log_error "--worktree には値が必要です"
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
        log_error "--task は必須です"
        usage
        exit 1
    fi
}

# プロジェクトルート検出
detect_project_root() {
    if [[ -n "$WORKTREE_PATH" ]]; then
        PROJECT_ROOT="$WORKTREE_PATH"
    else
        PROJECT_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
    fi
    log_info "プロジェクトルート: $PROJECT_ROOT"
}

# AGENTS.md ハッシュ計算
compute_agents_hash() {
    local agents_file="$PROJECT_ROOT/AGENTS.md"

    if [[ ! -f "$agents_file" ]]; then
        log_error "AGENTS.md が見つかりません: $agents_file"
        log_error "AGENTS.md は必須です。Worker 実行を中断します。"
        exit 1
    fi

    # BOM除去、LF正規化、SHA256先頭8文字
    AGENTS_HASH=$(sed '1s/^\xEF\xBB\xBF//' "$agents_file" | tr -d '\r' | shasum -a 256 | cut -c1-8)
    log_info "AGENTS.md ハッシュ: $AGENTS_HASH"
}

# Rules 連結
collect_rules() {
    local rules_dir="$PROJECT_ROOT/.claude/rules"
    local rules_content=""

    if [[ -d "$rules_dir" ]]; then
        for rule_file in "$rules_dir"/*.md; do
            if [[ -f "$rule_file" ]]; then
                rules_content+="# $(basename "$rule_file")"$'\n'
                rules_content+="$(cat "$rule_file")"$'\n\n'
            fi
        done
        log_info "Rules ファイル収集: $(find "$rules_dir" -name "*.md" | wc -l | tr -d ' ') 件"
    else
        log_warn "Rules ディレクトリが見つかりません: $rules_dir"
    fi

    echo "$rules_content"
}

# base-instructions 生成
generate_base_instructions() {
    local rules_content
    rules_content=$(collect_rules)

    cat << EOF
# Codex Worker Instructions

## Rules（プロジェクト固有ルール）

$rules_content

## AGENTS.md 強制読み込み指示

最初に AGENTS.md を読み、以下の形式で証跡を出力してください:

\`\`\`
AGENTS_SUMMARY: <1行要約> | HASH:<SHA256先頭8文字>
\`\`\`

証跡を出力せずに作業を開始しないでください。
証跡のハッシュは AGENTS.md の内容から計算してください。

EOF
}

# prompt 生成
generate_prompt() {
    cat << EOF
$TASK

---

重要: 作業開始前に、以下の形式で AGENTS.md の証跡を出力してください:

AGENTS_SUMMARY: <AGENTS.mdの1行要約> | HASH:<SHA256先頭8文字>

この証跡がない場合、作業は無効とみなされます。
EOF
}

# 証跡検証（Claude Code 内から呼び出す想定、このスクリプト内では使用しない）
# 実際の検証は codex-worker-quality-gate.sh の gate_evidence() で行う
verify_agents_summary() {
    local output="$1"

    # 正規表現でマッチ（大文字小文字両対応）
    if [[ "$output" =~ AGENTS_SUMMARY:[[:space:]]*(.+)[[:space:]]*\|[[:space:]]*HASH:([A-Fa-f0-9]{8}) ]]; then
        local summary="${BASH_REMATCH[1]}"
        local hash="${BASH_REMATCH[2]}"

        if [[ "${hash,,}" == "${AGENTS_HASH,,}" ]]; then
            log_info "証跡検証: OK (ハッシュ一致)"
            return 0
        else
            log_error "証跡検証: NG (ハッシュ不一致: 期待=$AGENTS_HASH, 実際=$hash)"
            return 1
        fi
    else
        log_error "証跡検証: NG (AGENTS_SUMMARY が見つかりません)"
        return 1
    fi
}

# Codex Worker 呼び出し（MCP 経由）
invoke_codex_worker() {
    local base_instructions
    local prompt
    local cwd="$PROJECT_ROOT"

    base_instructions=$(generate_base_instructions)
    prompt=$(generate_prompt)

    if [[ "$DRY_RUN" == true ]]; then
        echo ""
        echo "========================================"
        echo "ドライラン: 以下の内容で Codex を呼び出します"
        echo "========================================"
        echo ""
        echo "--- prompt ---"
        echo "$prompt"
        echo ""
        echo "--- base-instructions (先頭500文字) ---"
        echo "${base_instructions:0:500}..."
        echo ""
        echo "--- パラメータ ---"
        echo "cwd: $cwd"
        echo "approval-policy: $APPROVAL_POLICY"
        echo "sandbox: $SANDBOX"
        echo ""
        return 0
    fi

    log_step "Codex Worker を呼び出し中..."

    # 注: 実際の MCP 呼び出しは Claude Code 内から行う
    # このスクリプトは base-instructions と prompt の生成を担当
    # 出力をファイルに保存して Claude Code が読み取る

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
        }' > "$output_dir/mcp-params.json"

    # 検証用情報を保存
    cat > "$output_dir/verify-info.json" << EOF
{
  "agents_hash": "$AGENTS_HASH",
  "max_retries": $MAX_RETRIES,
  "verify_pattern": "AGENTS_SUMMARY:\\\\s*(.+?)\\\\s*\\\\|\\\\s*HASH:([A-Fa-f0-9]{8})"
}
EOF

    log_info "MCP パラメータを保存: $output_dir/mcp-params.json"
    log_info "検証情報を保存: $output_dir/verify-info.json"
    echo ""
    log_info "次のステップ:"
    log_info "  1. Claude Code から mcp__codex__codex を呼び出す"
    log_info "  2. 出力に AGENTS_SUMMARY 証跡があることを確認"
    log_info "  3. ハッシュが $AGENTS_HASH と一致することを確認"
    log_info "  4. 失敗時は最大 $MAX_RETRIES 回まで再試行"
}

# メイン処理
main() {
    parse_args "$@"

    echo "========================================"
    echo "Codex Worker Engine"
    echo "========================================"
    echo ""

    log_step "1. プロジェクトルート検出"
    detect_project_root

    log_step "2. 依存コマンドチェック"
    check_dependencies

    log_step "3. AGENTS.md ハッシュ計算"
    compute_agents_hash

    log_step "4. Codex Worker 呼び出し準備"
    invoke_codex_worker

    echo ""
    log_info "完了"
}

main "$@"
