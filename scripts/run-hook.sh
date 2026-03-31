#!/bin/bash
# run-hook.sh — node-free hook entry point
#
# Purpose:
#   Replace `node run-script.js <name>` in hooks.json to eliminate
#   the hard dependency on `node` being in /bin/sh PATH.
#   - .sh scripts → bash direct execution (no node dependency)
#   - .js scripts → resolve node dynamically, then execute
#
# Background:
#   nvm/fnm/volta users have `node` available in interactive shells
#   but NOT in /bin/sh (used by Claude Code hooks). This caused all
#   hooks to fail silently for marketplace plugin users.
#
# Usage (in hooks.json):
#   "command": "bash \"${CLAUDE_PLUGIN_ROOT}/scripts/run-hook.sh\" session-init"
#   "command": "bash \"${CLAUDE_PLUGIN_ROOT}/scripts/run-hook.sh\" hook-handlers/pre-compact-save.js"

set -uo pipefail
# NOTE: set -e is intentionally omitted. resolve_node uses commands that
# may fail (command -v, ls -d glob) and we handle errors explicitly.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# === 引数チェック ===
if [ $# -eq 0 ]; then
  echo "Usage: run-hook.sh <script-name> [args...]" >&2
  exit 1
fi

script_name="$1"; shift

# === Windows 検出 → run-script.js にフォールバック ===
case "$(uname -s)" in
  CYGWIN*|MINGW*|MSYS*)
    # Windows: run-script.js の Windows パス変換が必要
    NODE="$(command -v node 2>/dev/null || true)"
    if [ -n "$NODE" ]; then
      exec "$NODE" "${SCRIPT_DIR}/run-script.js" "$script_name" "$@"
    fi
    echo "harness: node not found on Windows — hooks require Node.js in PATH" >&2
    exit 1
    ;;
esac

# === パストラバーサル防止 ===
case "$script_name" in
  ../*|*/../*|*/..)
    echo "harness: invalid script path: ${script_name}" >&2
    exit 1
    ;;
esac

# === 拡張子補完 ===
if [[ "$script_name" != *.sh && "$script_name" != *.js ]]; then
  script_name="${script_name}.sh"
fi
script_path="${SCRIPT_DIR}/${script_name}"

# === 解決後パスが SCRIPT_DIR 配下にあることを検証 ===
real_script="$(cd "$(dirname "$script_path")" 2>/dev/null && pwd)/$(basename "$script_path")" || true
if [[ -z "$real_script" || "$real_script" != "${SCRIPT_DIR}/"* ]]; then
  echo "harness: script path escapes scripts dir: ${script_name}" >&2
  exit 1
fi

# === スクリプト不在 → silent exit（配布差分に対応）===
if [ ! -f "$script_path" ]; then
  exit 0
fi

# === .sh → bash 直接実行（node 不要）===
if [[ "$script_path" == *.sh ]]; then
  exec bash "$script_path" "$@"
fi

# === .js → node を動的解決して実行 ===
resolve_node() {
  # 1. PATH に node があればそれを使う
  local found
  found="$(command -v node 2>/dev/null || true)"
  if [ -n "$found" ]; then
    printf '%s' "$found"
    return 0
  fi

  # 2. nvm
  if [ -d "${HOME}/.nvm/versions/node" ]; then
    found="$(ls -d "${HOME}/.nvm/versions/node/"*/bin/node 2>/dev/null | tail -1 || true)"
    if [ -n "$found" ] && [ -x "$found" ]; then
      printf '%s' "$found"
      return 0
    fi
  fi

  # 3. fnm
  if [ -d "${HOME}/.local/share/fnm/node-versions" ]; then
    found="$(ls -d "${HOME}/.local/share/fnm/node-versions/"*/installation/bin/node 2>/dev/null | tail -1 || true)"
    if [ -n "$found" ] && [ -x "$found" ]; then
      printf '%s' "$found"
      return 0
    fi
  fi

  # 4. volta
  if [ -x "${HOME}/.volta/bin/node" ]; then
    printf '%s' "${HOME}/.volta/bin/node"
    return 0
  fi

  # 5. homebrew / system
  local candidate
  for candidate in /opt/homebrew/bin/node /usr/local/bin/node /usr/bin/node; do
    if [ -x "$candidate" ]; then
      printf '%s' "$candidate"
      return 0
    fi
  done

  return 1
}

NODE="$(resolve_node)" || true
if [ -z "$NODE" ]; then
  echo "harness: node not found — .js hook skipped: ${script_name}" >&2
  echo "harness: install Node.js or add it to system PATH" >&2
  exit 0
fi

exec "$NODE" "$script_path" "$@"
