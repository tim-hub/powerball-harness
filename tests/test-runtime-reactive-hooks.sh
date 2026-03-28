#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "${TMP_DIR}"' EXIT

mkdir -p "${TMP_DIR}/.claude/state"
SCRIPT="${ROOT_DIR}/scripts/hook-handlers/runtime-reactive.sh"

task_output="$(
  printf '%s' '{"hook_event_name":"TaskCreated","session_id":"sess-1","cwd":"'"${TMP_DIR}"'","task_id":"task-7","task_title":"Implement sync"}' \
  | bash "${SCRIPT}"
)"

grep -q 'Reactive hook tracked: TaskCreated' <<<"${task_output}" || {
  echo "TaskCreated should return a tracking response"
  exit 1
}

grep -q '"event":"TaskCreated"' "${TMP_DIR}/.claude/state/runtime-reactive.jsonl" || {
  echo "TaskCreated event was not logged"
  exit 1
}

grep -q '"task_id":"task-7"' "${TMP_DIR}/.claude/state/runtime-reactive.jsonl" || {
  echo "TaskCreated task_id was not logged"
  exit 1
}

file_output="$(
  printf '%s' '{"hook_event_name":"FileChanged","session_id":"sess-2","cwd":"'"${TMP_DIR}"'","file_path":"'"${TMP_DIR}"'/Plans.md"}' \
  | bash "${SCRIPT}"
)"

grep -q '"hookEventName":"FileChanged"' <<<"${file_output}" || {
  echo "FileChanged should return hookSpecificOutput"
  exit 1
}

grep -q 'Plans.md' <<<"${file_output}" || {
  echo "FileChanged should mention Plans.md re-read guidance"
  exit 1
}

settings_output="$(
  printf '%s' '{"hook_event_name":"FileChanged","session_id":"sess-4","cwd":"'"${TMP_DIR}"'","file_path":".claude-plugin/settings.json"}' \
  | bash "${SCRIPT}"
)"

grep -q '"hookEventName":"FileChanged"' <<<"${settings_output}" || {
  echo "Relative settings path should return hookSpecificOutput"
  exit 1
}

grep -q 'Harness 設定が更新されました' <<<"${settings_output}" || {
  echo "Relative settings path should emit rule/settings guidance"
  exit 1
}

rules_output="$(
  printf '%s' '{"hook_event_name":"FileChanged","session_id":"sess-5","cwd":"'"${TMP_DIR}"'","file_path":".claude/rules/testing.md"}' \
  | bash "${SCRIPT}"
)"

grep -q '"hookEventName":"FileChanged"' <<<"${rules_output}" || {
  echo "Relative rules path should return hookSpecificOutput"
  exit 1
}

grep -q 'Harness 設定が更新されました' <<<"${rules_output}" || {
  echo "Relative rules path should emit rule/settings guidance"
  exit 1
}

cwd_output="$(
  printf '%s' '{"hook_event_name":"CwdChanged","session_id":"sess-3","cwd":"'"${TMP_DIR}"'/worktree","previous_cwd":"'"${TMP_DIR}"'"}' \
  | bash "${SCRIPT}"
)"

grep -q '"hookEventName":"CwdChanged"' <<<"${cwd_output}" || {
  echo "CwdChanged should return hookSpecificOutput"
  exit 1
}

grep -q '作業ディレクトリが切り替わりました' <<<"${cwd_output}" || {
  echo "CwdChanged should emit context guidance"
  exit 1
}

echo "OK"
