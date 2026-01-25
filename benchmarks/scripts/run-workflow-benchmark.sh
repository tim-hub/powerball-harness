#!/bin/bash
# ======================================
# Workflow ベンチマーク実行スクリプト
# ======================================
#
# Plan → Work → Review の3ステップワークフローを
# with-plugin / no-plugin で実行し、結果を比較します。
#
# harness推奨コマンド運用の効果を測定するための評価スイート。

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
BENCHMARK_DIR="$(dirname "$SCRIPT_DIR")"
PLUGIN_ROOT="$(dirname "$BENCHMARK_DIR")"
TASKS_DIR="$BENCHMARK_DIR/tasks"
RESULTS_DIR="$BENCHMARK_DIR/results"
TEST_PROJECT="$BENCHMARK_DIR/test-project"

# デフォルト設定
TASK=""
WITH_PLUGIN=false
PLUGIN_PATH=""
ITERATION=1
ITERATIONS=1
TRACE_MODE=true
API_KEY="${ANTHROPIC_API_KEY:-}"
API_KEY_FILE=""
TIMEOUT=300
USE_TEMP_HOME=true  # CI用。ローカルでは false にして既存セッションを使用

# ヘルプ表示
show_help() {
  cat << EOF
Workflow ベンチマーク実行（Plan → Work → Review 3ステップ）

harness推奨コマンド運用の効果を測定するための評価スイート。
with-plugin は CI専用コマンドを使用、no-plugin は同等の手動プロンプトを使用。

Usage: $0 [OPTIONS]

OPTIONS:
  --task <name>       実行するタスク名（必須）
  --with-plugin       プラグインを有効化してテスト（CI専用コマンド使用）
  --plugin-path <p>   プラグインのパス（デフォルト: このリポジトリ）
  --iterations <n>    試行回数（デフォルト: 1）。N回実行してN個の結果JSONを出力
  --iteration <n>     [内部用] イテレーション番号（デフォルト: 1）
  --api-key <key>     ANTHROPIC_API_KEY
  --api-key-file <p>  APIキーをファイルから読み込む
  --timeout <sec>     各ステップのタイムアウト秒数（デフォルト: 300）
  --no-temp-home      一時HOMEを使わず既存セッション（OAuth）を使用（ローカル用）
  --no-trace          trace を無効化
  --help              このヘルプを表示

EXAMPLES:
  # プラグインなし（ベースライン）
  $0 --task plan-feature

  # プラグインあり（harness推奨コマンド）
  $0 --task plan-feature --with-plugin

  # 比較実行
  $0 --task plan-feature && $0 --task plan-feature --with-plugin
EOF
}

# 引数解析
while [[ $# -gt 0 ]]; do
  case "$1" in
    --task)
      TASK="$2"
      shift 2
      ;;
    --with-plugin)
      WITH_PLUGIN=true
      shift
      ;;
    --plugin-path)
      PLUGIN_PATH="$2"
      shift 2
      ;;
    --iterations)
      ITERATIONS="$2"
      shift 2
      ;;
    --iteration)
      ITERATION="$2"
      shift 2
      ;;
    --api-key)
      API_KEY="$2"
      shift 2
      ;;
    --api-key-file)
      API_KEY_FILE="$2"
      shift 2
      ;;
    --timeout)
      TIMEOUT="$2"
      shift 2
      ;;
    --no-trace)
      TRACE_MODE=false
      shift
      ;;
    --no-temp-home)
      USE_TEMP_HOME=false
      shift
      ;;
    --help)
      show_help
      exit 0
      ;;
    *)
      echo "Unknown option: $1"
      show_help
      exit 1
      ;;
  esac
done

# APIキーをファイルから読み込む
if [[ -n "$API_KEY_FILE" ]]; then
  if [[ ! -f "$API_KEY_FILE" ]]; then
    echo "Error: --api-key-file のファイルが見つかりません: $API_KEY_FILE" >&2
    exit 1
  fi
  API_KEY="$(head -n 1 "$API_KEY_FILE" | tr -d '\r\n' | xargs)"
fi

# バリデーション
if [[ -z "$TASK" ]]; then
  echo "Error: --task を指定してください"
  show_help
  exit 1
fi

# デフォルトのプラグインパス
if [[ "$WITH_PLUGIN" == "true" && -z "$PLUGIN_PATH" ]]; then
  PLUGIN_PATH="$PLUGIN_ROOT"
fi

# タスクのプロンプトを取得
get_task_prompt() {
  local task="$1"
  local task_file="$TASKS_DIR/$task.md"

  if [[ ! -f "$task_file" ]]; then
    echo "Error: タスクファイルが見つかりません: $task_file" >&2
    exit 1
  fi

  sed -n '/^## プロンプト/,/^## /p' "$task_file" | \
    sed -n '/^```$/,/^```$/p' | \
    sed '1d;$d'
}

# タイムアウトコマンドを取得
get_timeout_cmd() {
  if command -v gtimeout &> /dev/null; then
    echo "gtimeout $TIMEOUT"
  elif command -v timeout &> /dev/null; then
    echo "timeout $TIMEOUT"
  else
    echo ""
  fi
}

# bc出力をJSON有効な数値に正規化（.5 → 0.5, -.5 → -0.5）
normalize_number() {
  local num="$1"
  # 空や不正な場合は 0 を返す
  if [[ -z "$num" || "$num" == "." || "$num" == "-" ]]; then
    echo "0"
    return
  fi
  # 先頭が . の場合は 0 を付ける（例: .5 → 0.5）
  if [[ "${num:0:1}" == "." ]]; then
    num="0$num"
  # 先頭が -. の場合は -0. に変換（例: -.5 → -0.5）
  elif [[ "${num:0:2}" == "-." ]]; then
    num="-0${num:1}"
  fi
  echo "$num"
}

# 単一ステップを実行
run_step() {
  local step_name="$1"
  local prompt="$2"
  local output_file="$3"
  local trace_file="$4"
  local temp_home="$5"

  local claude_args=(--print --dangerously-skip-permissions)

  if [[ "$WITH_PLUGIN" == "true" && -n "$PLUGIN_PATH" ]]; then
    claude_args+=(--plugin-dir "$PLUGIN_PATH")
  fi

  if [[ "$TRACE_MODE" == "true" ]]; then
    claude_args+=(--output-format stream-json --verbose)
  fi

  local start_time=$(date +%s.%N)
  local exit_code=0
  local timeout_cmd=$(get_timeout_cmd)

  echo "  実行中: $step_name" >&2

  if [[ -n "$timeout_cmd" ]]; then
    if HOME="$temp_home" $timeout_cmd claude "${claude_args[@]}" "$prompt" > "$output_file" 2>&1; then
      exit_code=0
    else
      exit_code=$?
    fi
  else
    if HOME="$temp_home" claude "${claude_args[@]}" "$prompt" > "$output_file" 2>&1; then
      exit_code=0
    else
      exit_code=$?
    fi
  fi

  local end_time=$(date +%s.%N)
  local duration_raw=$(echo "$end_time - $start_time" | bc)
  local duration=$(normalize_number "$duration_raw")

  # trace ファイルをコピー
  if [[ "$TRACE_MODE" == "true" ]]; then
    cp -f "$output_file" "$trace_file" 2>/dev/null || true
  fi

  # メトリクス抽出
  local input_tokens=0
  local output_tokens=0

  if [[ "$TRACE_MODE" == "true" && -f "$trace_file" ]] && command -v jq &> /dev/null; then
    input_tokens=$(grep '"type":"result"' "$trace_file" 2>/dev/null | jq -r '.usage.input_tokens // 0' 2>/dev/null | awk '{sum+=$1} END {print sum+0}')
    output_tokens=$(grep '"type":"result"' "$trace_file" 2>/dev/null | jq -r '.usage.output_tokens // 0' 2>/dev/null | awk '{sum+=$1} END {print sum+0}')
  fi

  # 結果を返す（JSON形式）
  cat << EOF
{
  "step": "$step_name",
  "duration": $duration,
  "exit_code": $exit_code,
  "success": $([ $exit_code -eq 0 ] && echo "true" || echo "false"),
  "input_tokens": $input_tokens,
  "output_tokens": $output_tokens
}
EOF
}

# ワークフローを実行
run_workflow() {
  local version_label
  if [[ "$WITH_PLUGIN" == "true" ]]; then
    version_label="isolated-with-plugin"
  else
    version_label="isolated-no-plugin"
  fi

  local timestamp=$(date +%Y%m%d-%H%M%S)
  local result_prefix="$RESULTS_DIR/${TASK}_${version_label}_${ITERATION}_${timestamp}"
  local result_file="${result_prefix}.json"

  echo "========================================"
  echo "Workflow ベンチマーク"
  echo "========================================"
  echo "Task: $TASK"
  echo "Mode: $version_label"
  echo "Iteration: $ITERATION"
  echo "========================================"

  # HOME ディレクトリ設定
  local temp_home
  local cleanup_home=false

  if [[ "$USE_TEMP_HOME" == "true" ]]; then
    # CI用: 一時 HOME を作成
    temp_home=$(mktemp -d)
    mkdir -p "$temp_home/.claude"
    cleanup_home=true
    echo "一時 HOME: $temp_home"

    # 認証設定
    if [[ -n "$API_KEY" ]]; then
      export ANTHROPIC_API_KEY="$API_KEY"
    elif [[ -f "$HOME/.claude/.credentials.json" ]]; then
      cp "$HOME/.claude/.credentials.json" "$temp_home/.claude/"
    fi
  else
    # ローカル用: 既存の HOME（OAuth セッション）を使用
    temp_home="$HOME"
    echo "既存 HOME を使用（OAuth セッション）"
  fi

  # テストプロジェクトをリセット
  "$SCRIPT_DIR/setup-test-project.sh" > /dev/null 2>&1
  cd "$TEST_PROJECT"

  # タスクプロンプト取得
  local task_prompt=$(get_task_prompt "$TASK")

  # 3ステップのプロンプトを構築
  local plan_prompt work_prompt review_prompt

  if [[ "$WITH_PLUGIN" == "true" ]]; then
    # with-plugin: CIモードを使用（完全修飾名で指定）
    plan_prompt="/claude-code-harness:core:plan-with-agent --ci $task_prompt"
    work_prompt="/claude-code-harness:core:work --ci"
    review_prompt="/claude-code-harness:core:harness-review --ci"
  else
    # no-plugin: CI コマンドと同等の詳細プロンプト（公平な比較のため）
    # 違いは「プラグイン基盤（hooks/skills/agents）の有無」のみ

    plan_prompt="## 計画作成タスク（CI用・非対話）

### 制約
- AskUserQuestion 禁止: 質問せずに進める
- WebSearch 禁止: 外部検索なしで進める
- 確認プロンプト禁止: 自動で完了まで進める

### 入力要件
$task_prompt

### 実行手順
1. 要件の解析: 上記の要件を抽出
2. タスク分解: 実装可能な単位に分解（3-7個程度）
3. Plans.md 生成: Plans.md に書き込み
4. 完了出力: 生成したタスク数を報告

### 出力形式
Plans.md を以下の形式で生成:

\`\`\`markdown
## タスク一覧

- [ ] タスク1の説明 \`cc:TODO\`
- [ ] タスク2の説明 \`cc:TODO\`
- [ ] タスク3の説明 \`cc:TODO\`
\`\`\`

### 成功基準
- Plans.md が存在する
- 3つ以上のタスクが cc:TODO マーカー付きで記載されている
- 各タスクが実装可能な具体的な内容である"

    work_prompt="## 実装実行タスク（CI用・非対話）

### 制約
- AskUserQuestion 禁止: 質問せずに進める
- WebSearch 禁止: 外部検索なしで進める
- 確認プロンプト禁止: 自動で完了まで進める
- ビルド検証: 可能なら npm test / npm run build を実行

### 実行手順
1. Plans.md 読み込み: cc:TODO マーカーのタスクを抽出
2. 順次実装: 各タスクを実装（ファイル作成/編集）
3. マーカー更新: 完了したタスクを cc:完了 に変更
4. ビルド検証: npm test または npm run build を実行（可能な場合）
5. 完了出力: 実装結果のサマリーを報告

### 出力形式
\`\`\`
## 実装結果

### 完了タスク
- [x] タスク1 \`cc:完了\`
- [x] タスク2 \`cc:完了\`

### 作成/変更ファイル
- src/utils/helper.ts (新規)
- src/index.ts (変更)

### ビルド結果
- テスト: X/X passed
- ビルド: success/failure
\`\`\`

### 成功基準
- 1つ以上のタスクが cc:完了 になっている
- 実装したファイルが存在する"

    review_prompt="## レビュー実行タスク（CI用・非対話）

### 制約
- AskUserQuestion 禁止: 質問せずに進める
- WebSearch 禁止: 外部検索なしで進める
- 確認プロンプト禁止: 自動で完了まで進める
- 修正適用禁止: レビュー結果の報告のみ（修正は行わない）

### 実行手順
変更ファイルを検出してレビューする。

### 出力形式（必須: 各指摘に Severity 行を含める）

\`\`\`
## Review Result

### Summary
- Files Reviewed: 5
- Total Issues: 3
- Critical: 0
- High: 1
- Medium: 2
- Low: 0

### Issues

#### Issue 1
- File: src/utils/helper.ts
- Line: 25
- Severity: High
- Category: Security
- Description: 問題の説明
- Suggestion: 修正案

#### Issue 2
- File: src/index.ts
- Line: 42
- Severity: Medium
- Category: Quality
- Description: 問題の説明
- Suggestion: 修正案

### Pass/Fail
- Result: PASS または FAIL
- Reason: 判定理由
\`\`\`

### Severity 定義
| Severity | 基準 |
|----------|------|
| Critical | セキュリティ脆弱性、データ損失リスク |
| High | 重大なバグ、パフォーマンス問題 |
| Medium | コード品質、ベストプラクティス違反 |
| Low | スタイル、軽微な改善点 |

### 成功基準
- レビュー結果が出力されている
- 各指摘に Severity が付与されている
- Summary セクションに集計がある
- Pass/Fail 判定がある"
  fi

  # ワークフロー開始時刻
  local workflow_start=$(date +%s.%N)

  # Step 1: Plan
  echo ""
  echo "Step 1/3: Plan"
  local plan_output="${result_prefix}_plan.output.txt"
  local plan_trace="${result_prefix}_plan.trace.jsonl"
  local plan_result=$(run_step "plan" "$plan_prompt" "$plan_output" "$plan_trace" "$temp_home")

  # Step 2: Work
  echo "Step 2/3: Work"
  local work_output="${result_prefix}_work.output.txt"
  local work_trace="${result_prefix}_work.trace.jsonl"
  local work_result=$(run_step "work" "$work_prompt" "$work_output" "$work_trace" "$temp_home")

  # Step 3: Review
  echo "Step 3/3: Review"
  local review_output="${result_prefix}_review.output.txt"
  local review_trace="${result_prefix}_review.trace.jsonl"
  local review_result=$(run_step "review" "$review_prompt" "$review_output" "$review_trace" "$temp_home")

  # ワークフロー終了時刻
  local workflow_end=$(date +%s.%N)
  local total_duration_raw=$(echo "$workflow_end - $workflow_start" | bc)
  local total_duration=$(normalize_number "$total_duration_raw")

  # 合計トークン・コスト計算
  local total_input_tokens=0
  local total_output_tokens=0
  local estimated_cost="0.0000"

  if command -v jq &> /dev/null; then
    total_input_tokens=$(printf '%s\n%s\n%s\n' "$plan_result" "$work_result" "$review_result" | jq -s '[.[].input_tokens] | add // 0')
    total_output_tokens=$(printf '%s\n%s\n%s\n' "$plan_result" "$work_result" "$review_result" | jq -s '[.[].output_tokens] | add // 0')

    if [[ "$total_input_tokens" -gt 0 || "$total_output_tokens" -gt 0 ]]; then
      estimated_cost_raw=$(echo "($total_input_tokens * 0.000003) + ($total_output_tokens * 0.000015)" | bc 2>/dev/null || echo "0")
      estimated_cost=$(echo "$estimated_cost_raw" | awk '{printf "%.4f", $1}')
    fi
  fi

  # グレード計算
  local grade_json="null"
  if [[ -f "$SCRIPT_DIR/grade-task.sh" ]]; then
    grade_json=$(bash "$SCRIPT_DIR/grade-task.sh" \
      --task "$TASK" \
      --project-dir "$TEST_PROJECT" \
      --output-file "$review_output" \
      --trace-file "$review_trace" \
      --workflow-mode \
      2>/dev/null || echo "null")
  fi

  # CI専用コマンドが使われたか検出（with-plugin の transcript grader）
  local used_ci_commands=false
  if [[ "$WITH_PLUGIN" == "true" && -f "$plan_trace" ]]; then
    # 完全修飾名またはショートネームで検出
    if grep -qE 'claude-code-harness:core:(plan-with-agent|work|harness-review)-ci|/(plan-with-agent|work|harness-review)-ci' "$plan_trace" "$work_trace" "$review_trace" 2>/dev/null; then
      used_ci_commands=true
    fi
  fi

  # 結果 JSON を生成
  cat > "$result_file" << EOF
{
  "task": "$TASK",
  "version": "$version_label",
  "iteration": $ITERATION,
  "timestamp": "$timestamp",
  "suite_id": "workflow-v1",
  "workflow_mode": true,
  "duration_seconds": $total_duration,
  "success": true,
  "with_plugin": $WITH_PLUGIN,
  "plugin_path": $([ "$WITH_PLUGIN" == "true" ] && echo "\"$PLUGIN_PATH\"" || echo "null"),
  "used_ci_commands": $used_ci_commands,
  "input_tokens": $total_input_tokens,
  "output_tokens": $total_output_tokens,
  "estimated_cost_usd": $estimated_cost,
  "cost_assumption": "sonnet_3_5_input_3_per_mtok_output_15_per_mtok",
  "steps": {
    "plan": $plan_result,
    "work": $work_result,
    "review": $review_result
  },
  "trace_files": [
    "$(basename "$plan_trace")",
    "$(basename "$work_trace")",
    "$(basename "$review_trace")"
  ],
  "grade": $grade_json
}
EOF

  echo ""
  echo "========================================"
  echo "結果: $result_file"
  echo "所要時間: ${total_duration}s"
  echo "推定コスト: \$${estimated_cost}"
  echo "========================================"

  # クリーンアップ（一時HOMEの場合のみ）
  if [[ "$cleanup_home" == "true" ]]; then
    rm -rf "$temp_home"
  fi
}

# メイン実行
mkdir -p "$RESULTS_DIR"

# --iterations が指定されていれば、N回ループ実行
if [[ "$ITERATIONS" -gt 1 ]]; then
  echo "========================================"
  echo "複数試行モード: $ITERATIONS 回実行"
  echo "========================================"
  for i in $(seq 1 "$ITERATIONS"); do
    ITERATION=$i
    echo ""
    echo "=== Trial $i / $ITERATIONS ==="
    run_workflow
  done
  echo ""
  echo "========================================"
  echo "全 $ITERATIONS 試行完了"
  echo "========================================"
else
  run_workflow
fi
