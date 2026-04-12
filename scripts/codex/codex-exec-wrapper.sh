#!/bin/bash
# codex-exec-wrapper.sh
# Wrapper automating pre-processing (rule injection) and post-processing (result recording/marker extraction) for codex exec
#
# Usage: ./scripts/codex/codex-exec-wrapper.sh <prompt_file> [timeout_seconds]
#   prompt_file      : Path to prompt file for codex exec
#   timeout_seconds  : Timeout in seconds (default: 120)
#
# Environment variables:
#   HARNESS_CODEX_NO_SYNC : Set to 1 to skip sync-rules-to-agents.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PLUGIN_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
EXECUTION_ROOT="${HARNESS_CODEX_EXECUTION_ROOT:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
CONTRACT_TEMPLATE="${SCRIPT_DIR}/../lib/codex-hardening-contract.txt"
HARDENING_MARKER="HARNESS_HARDENING_CONTRACT_V1"

PROMPT_FILE="${1:-}"
TIMEOUT_SEC="${2:-120}"

# === Argument check ===
if [ -z "${PROMPT_FILE}" ]; then
  echo "Usage: $0 <prompt_file> [timeout_seconds]" >&2
  exit 1
fi

if [ ! -f "${PROMPT_FILE}" ]; then
  echo "Error: prompt file not found: ${PROMPT_FILE}" >&2
  exit 1
fi

# === Detect timeout command (macOS compat) ===
TIMEOUT=$(command -v timeout || command -v gtimeout || echo "")

# === Pre-processing: Verify AGENTS.md is up to date ===
SYNC_SCRIPT="${SCRIPT_DIR}/sync-rules-to-agents.sh"
if [ "${HARNESS_CODEX_NO_SYNC:-}" != "1" ] && [ -f "${SYNC_SCRIPT}" ]; then
  echo "[codex-exec-wrapper] Running sync-rules-to-agents.sh..." >&2
  bash "${SYNC_SCRIPT}" >&2 || {
    echo "[codex-exec-wrapper] Warning: sync-rules-to-agents.sh failed (continuing)" >&2
  }
fi

# === Hardening contract ===
generate_hardening_contract() {
  if [ ! -f "${CONTRACT_TEMPLATE}" ]; then
    echo "[codex-exec-wrapper] Error: hardening contract template not found: ${CONTRACT_TEMPLATE}" >&2
    exit 1
  fi
  cat "${CONTRACT_TEMPLATE}"
}

# Generate the injected contract once so the prompt, base instructions, and state artifact stay aligned.
build_hardening_contract_artifact() {
  local output_dir="$1"
  mkdir -p "$output_dir"
  generate_hardening_contract > "$output_dir/hardening-contract.txt"
}

prepend_hardening_contract_if_missing() {
  local file_path="$1"
  local tmp_file=""
  if [ ! -f "${file_path}" ]; then
    return 0
  fi
  if grep -Fq "${HARDENING_MARKER}" "${file_path}" 2>/dev/null; then
    return 0
  fi
  tmp_file="$(mktemp /tmp/codex-contract-sync.XXXXXX)"
  {
    generate_hardening_contract
    printf '\n---\n\n'
    cat "${file_path}"
  } > "${tmp_file}"
  mv "${tmp_file}" "${file_path}"
}

# === Create injected prompt ===
CODEX_STATE_DIR="${HARNESS_CODEX_STATE_DIR:-${EXECUTION_ROOT}/.claude/state/codex-worker}"
TMP_PROMPT="$(mktemp /tmp/codex-exec-prompt.XXXXXX)"
build_hardening_contract_artifact "$CODEX_STATE_DIR"
prepend_hardening_contract_if_missing "${CODEX_STATE_DIR}/base-instructions.txt"
prepend_hardening_contract_if_missing "${CODEX_STATE_DIR}/prompt.txt"
if grep -Fq "${HARDENING_MARKER}" "${PROMPT_FILE}" 2>/dev/null; then
  cp "${PROMPT_FILE}" "${TMP_PROMPT}"
else
  {
    generate_hardening_contract
    printf '\n---\n\n'
    cat "${PROMPT_FILE}"
  } > "${TMP_PROMPT}"
fi

# === Prepare temporary files ===
TMP_OUT="$(mktemp /tmp/codex-exec-out.XXXXXX)"
TMP_LEARNING="$(mktemp /tmp/codex-learning.XXXXXX)"
trap 'rm -f "${TMP_OUT}" "${TMP_LEARNING}" "${TMP_PROMPT}"' EXIT

# === Main: Execute codex exec ===
echo "[codex-exec-wrapper] Running codex exec (timeout=${TIMEOUT_SEC}s)..." >&2

EXIT_CODE=0
# Pass prompt via stdin (avoid ARG_MAX overflow)
# "-" is the official stdin input specification for codex exec
if [ -n "${TIMEOUT}" ]; then
  cat "${TMP_PROMPT}" | ${TIMEOUT} "${TIMEOUT_SEC}" codex exec - --full-auto > "${TMP_OUT}" 2>>/tmp/harness-codex-$$.log || EXIT_CODE=$?
else
  cat "${TMP_PROMPT}" | codex exec - --full-auto > "${TMP_OUT}" 2>>/tmp/harness-codex-$$.log || EXIT_CODE=$?
fi

# Output log even on timeout (exit 124)
if [ "${EXIT_CODE}" -eq 124 ]; then
  echo "[codex-exec-wrapper] Warning: codex exec timed out (${TIMEOUT_SEC}s)" >&2
fi

# === Post-processing: Extract [HARNESS-LEARNING] marker lines ===
# NOTE: Structured JSON output is possible via Codex CLI --output-schema option.
# Migration from marker grep to --output-schema is considered for the future (requires schema definition).
# Extract only lines starting with `[HARNESS-LEARNING]` from stdout and strip marker
LEARNING_COUNT=0
if grep -q '^\[HARNESS-LEARNING\]' "${TMP_OUT}" 2>/dev/null; then
  grep '^\[HARNESS-LEARNING\]' "${TMP_OUT}" | sed 's/^\[HARNESS-LEARNING\] *//' > "${TMP_LEARNING}"
  LEARNING_COUNT="$(wc -l < "${TMP_LEARNING}" | tr -d ' ')"
  echo "[codex-exec-wrapper] ${LEARNING_COUNT} learning markers detected" >&2

  # === Secret filter ===
  # Remove lines containing token/key/password/secret/credential/api_key (case insensitive)
  TMP_FILTERED="$(mktemp /tmp/codex-filtered.XXXXXX)"
  trap 'rm -f "${TMP_OUT}" "${TMP_LEARNING}" "${TMP_FILTERED}"' EXIT
  grep -viE '(token|key|password|secret|credential|api_key)' "${TMP_LEARNING}" > "${TMP_FILTERED}" 2>/dev/null || true
  FILTERED_COUNT="$(wc -l < "${TMP_FILTERED}" | tr -d ' ')"
  REMOVED=$((LEARNING_COUNT - FILTERED_COUNT))
  if [ "${REMOVED}" -gt 0 ]; then
    echo "[codex-exec-wrapper] Warning: Removed ${REMOVED} lines containing potential secrets" >&2
  fi

  # === Atomic append to codex-learnings.md (mkdir lock, macOS compat) ===
  MEMORY_DIR="${HARNESS_CODEX_MEMORY_DIR:-${EXECUTION_ROOT}/.claude/memory}"
  mkdir -p "${MEMORY_DIR}"
  LEARNINGS_FILE="${MEMORY_DIR}/codex-learnings.md"
  LOCK_DIR="${MEMORY_DIR}/.codex-learnings.lock"
  TS="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
  DATE_ONLY="$(date -u +"%Y-%m-%d")"
  PROMPT_BASENAME="$(basename "${PROMPT_FILE}")"

  # Acquire lock (max 10 second wait)
  _lock_acquired=0
  for _i in $(seq 1 20); do
    if mkdir "${LOCK_DIR}" 2>/dev/null; then
      _lock_acquired=1
      break
    fi
    sleep 0.5
  done

  if [ "${_lock_acquired}" -eq 1 ]; then
    # Create header if file does not exist
    if [ ! -f "${LEARNINGS_FILE}" ]; then
      printf '# codex-learnings.md\n\nRecord of learnings extracted from codex exec.\n\n' > "${LEARNINGS_FILE}"
    fi

    # Append with section header
    if [ "${FILTERED_COUNT}" -gt 0 ]; then
      {
        printf '\n## %s %s\n\n' "${DATE_ONLY}" "${PROMPT_BASENAME}"
        while IFS= read -r line; do
          printf '- %s\n' "${line}"
        done < "${TMP_FILTERED}"
      } >> "${LEARNINGS_FILE}" 2>/dev/null || true
    fi

    # Release lock
    rmdir "${LOCK_DIR}" 2>/dev/null || true
  else
    echo "[codex-exec-wrapper] Warning: Lock acquisition timeout, skipping codex-learnings.md append" >&2
  fi

  # Also save learnings as JSONL in state directory (backward compat)
  STATE_DIR="${HARNESS_CODEX_GENERAL_STATE_DIR:-${EXECUTION_ROOT}/.claude/state}"
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

# === Pass through stdout ===
cat "${TMP_OUT}"

# === Propagate exit code ===
exit "${EXIT_CODE}"
