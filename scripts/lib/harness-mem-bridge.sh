#!/bin/bash
# harness-mem-bridge.sh
# Local-only sibling repo bridge for harness-mem wrapper scripts.

resolve_harness_mem_root() {
  if [ -n "${HARNESS_MEM_ROOT:-}" ] && [ -d "${HARNESS_MEM_ROOT}" ]; then
    cd "${HARNESS_MEM_ROOT}" && pwd
    return 0
  fi

  local bridge_dir repo_root repo_parent candidate
  bridge_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
  repo_root="$(cd "${bridge_dir}/../.." && pwd)"
  repo_parent="$(cd "${repo_root}/.." && pwd)"

  for candidate in \
    "${HOME}/.harness-mem/runtime/harness-mem" \
    "${repo_parent}/harness-mem" \
    "${HOME}/LocalWork/Code/CC-harness/harness-mem" \
    "${HOME}/Desktop/Code/CC-harness/harness-mem"
  do
    if [ -d "${candidate}" ]; then
      cd "${candidate}" && pwd
      return 0
    fi
  done

  return 1
}

exec_harness_mem_script() {
  local relative_path="$1"
  shift

  local harness_mem_root target_path
  if ! harness_mem_root="$(resolve_harness_mem_root)"; then
    echo "[claude-code-harness] harness-mem repo not found" >&2
    exit 0
  fi

  target_path="${harness_mem_root}/${relative_path}"
  if [ ! -x "${target_path}" ]; then
    echo "[claude-code-harness] harness-mem target missing: ${target_path}" >&2
    exit 0
  fi

  exec "${target_path}" "$@"
}
