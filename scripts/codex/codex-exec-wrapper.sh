#!/bin/bash
# codex-exec-wrapper.sh
# codex exec の前処理（ルール注入）と後処理（結果記録・マーカー抽出）を自動化するラッパー
#
# Usage: ./scripts/codex/codex-exec-wrapper.sh <prompt_file> [timeout_seconds]
#   prompt_file      : codex exec に渡すプロンプトファイルのパス
#   timeout_seconds  : タイムアウト秒数（デフォルト: 120）
#
# 環境変数:
#   HARNESS_CODEX_NO_SYNC : 1 を指定すると sync-rules-to-agents.sh をスキップ

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

PROMPT_FILE="${1:-}"
TIMEOUT_SEC="${2:-120}"

# === 引数チェック ===
if [ -z "${PROMPT_FILE}" ]; then
  echo "Usage: $0 <prompt_file> [timeout_seconds]" >&2
  exit 1
fi

if [ ! -f "${PROMPT_FILE}" ]; then
  echo "Error: prompt file not found: ${PROMPT_FILE}" >&2
  exit 1
fi

# === timeout コマンド検出（macOS 対応）===
TIMEOUT=$(command -v timeout || command -v gtimeout || echo "")

# === 前処理: AGENTS.md が最新であることを確認 ===
SYNC_SCRIPT="${SCRIPT_DIR}/sync-rules-to-agents.sh"
if [ "${HARNESS_CODEX_NO_SYNC:-}" != "1" ] && [ -f "${SYNC_SCRIPT}" ]; then
  echo "[codex-exec-wrapper] sync-rules-to-agents.sh を実行中..." >&2
  bash "${SYNC_SCRIPT}" >&2 || {
    echo "[codex-exec-wrapper] Warning: sync-rules-to-agents.sh が失敗しました（続行）" >&2
  }
fi

# === 一時ファイルの準備 ===
TMP_OUT="$(mktemp /tmp/codex-exec-out.XXXXXX)"
TMP_LEARNING="$(mktemp /tmp/codex-learning.XXXXXX)"
trap 'rm -f "${TMP_OUT}" "${TMP_LEARNING}"' EXIT

# === 本体: codex exec を実行 ===
echo "[codex-exec-wrapper] codex exec 実行中（timeout=${TIMEOUT_SEC}s）..." >&2

EXIT_CODE=0
if [ -n "${TIMEOUT}" ]; then
  ${TIMEOUT} "${TIMEOUT_SEC}" codex exec "$(cat "${PROMPT_FILE}")" > "${TMP_OUT}" 2>/dev/null || EXIT_CODE=$?
else
  codex exec "$(cat "${PROMPT_FILE}")" > "${TMP_OUT}" 2>/dev/null || EXIT_CODE=$?
fi

# タイムアウト（exit 124）の場合もログを出力
if [ "${EXIT_CODE}" -eq 124 ]; then
  echo "[codex-exec-wrapper] Warning: codex exec がタイムアウトしました（${TIMEOUT_SEC}s）" >&2
fi

# === 後処理: [HARNESS-LEARNING] マーカー行の抽出 ===
# stdout から `[HARNESS-LEARNING]` で始まる行のみを抽出してマーカーを除去
LEARNING_COUNT=0
if grep -q '^\[HARNESS-LEARNING\]' "${TMP_OUT}" 2>/dev/null; then
  grep '^\[HARNESS-LEARNING\]' "${TMP_OUT}" | sed 's/^\[HARNESS-LEARNING\] *//' > "${TMP_LEARNING}"
  LEARNING_COUNT="$(wc -l < "${TMP_LEARNING}" | tr -d ' ')"
  echo "[codex-exec-wrapper] ${LEARNING_COUNT} 件の学習マーカーを検出しました" >&2

  # === シークレットフィルタ ===
  # token/key/password/secret/credential/api_key を含む行を除去（大文字小文字無視）
  TMP_FILTERED="$(mktemp /tmp/codex-filtered.XXXXXX)"
  trap 'rm -f "${TMP_OUT}" "${TMP_LEARNING}" "${TMP_FILTERED}"' EXIT
  grep -viE '(token|key|password|secret|credential|api_key)' "${TMP_LEARNING}" > "${TMP_FILTERED}" 2>/dev/null || true
  FILTERED_COUNT="$(wc -l < "${TMP_FILTERED}" | tr -d ' ')"
  REMOVED=$((LEARNING_COUNT - FILTERED_COUNT))
  if [ "${REMOVED}" -gt 0 ]; then
    echo "[codex-exec-wrapper] Warning: シークレット候補 ${REMOVED} 行を除去しました" >&2
  fi

  # === codex-learnings.md にアトミック追記（mkdir ロック方式、macOS 対応）===
  MEMORY_DIR="${PROJECT_ROOT}/.claude/memory"
  mkdir -p "${MEMORY_DIR}"
  LEARNINGS_FILE="${MEMORY_DIR}/codex-learnings.md"
  LOCK_DIR="${MEMORY_DIR}/.codex-learnings.lock"
  TS="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
  DATE_ONLY="$(date -u +"%Y-%m-%d")"
  PROMPT_BASENAME="$(basename "${PROMPT_FILE}")"

  # ロック取得（最大 10 秒待機）
  _lock_acquired=0
  for _i in $(seq 1 20); do
    if mkdir "${LOCK_DIR}" 2>/dev/null; then
      _lock_acquired=1
      break
    fi
    sleep 0.5
  done

  if [ "${_lock_acquired}" -eq 1 ]; then
    # ファイルが存在しない場合はヘッダーを作成
    if [ ! -f "${LEARNINGS_FILE}" ]; then
      printf '# codex-learnings.md\n\ncodex exec から抽出した学習内容の記録。\n\n' > "${LEARNINGS_FILE}"
    fi

    # セクションヘッダーを付与して追記
    if [ "${FILTERED_COUNT}" -gt 0 ]; then
      {
        printf '\n## %s %s\n\n' "${DATE_ONLY}" "${PROMPT_BASENAME}"
        while IFS= read -r line; do
          printf '- %s\n' "${line}"
        done < "${TMP_FILTERED}"
      } >> "${LEARNINGS_FILE}" 2>/dev/null || true
    fi

    # ロック解放
    rmdir "${LOCK_DIR}" 2>/dev/null || true
  else
    echo "[codex-exec-wrapper] Warning: ロック取得タイムアウト、codex-learnings.md への追記をスキップ" >&2
  fi

  # 学習内容を state ディレクトリにも JSONL 保存（既存互換）
  STATE_DIR="${PROJECT_ROOT}/.claude/state"
  mkdir -p "${STATE_DIR}"
  LEARNING_FILE="${STATE_DIR}/codex-learning.jsonl"

  while IFS= read -r line; do
    if command -v jq >/dev/null 2>&1; then
      jq -nc \
        --arg ts "${TS}" \
        --arg prompt_file "${PROMPT_FILE}" \
        --arg content "${line}" \
        '{timestamp:$ts, prompt_file:$prompt_file, content:$content}' \
        >> "${LEARNING_FILE}" 2>/dev/null || true
    else
      printf '{"timestamp":"%s","prompt_file":"%s","content":"%s"}\n' \
        "${TS}" "${PROMPT_FILE}" "${line//\"/\\\"}" \
        >> "${LEARNING_FILE}" 2>/dev/null || true
    fi
  done < "${TMP_FILTERED}"
fi

# === stdout を通過させる ===
cat "${TMP_OUT}"

# === exit code を伝播 ===
exit "${EXIT_CODE}"
