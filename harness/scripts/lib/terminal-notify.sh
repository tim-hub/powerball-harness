#!/bin/bash
# terminal-notify.sh
# Shared helpers for building the CC 2.1.141+ hook JSON output `terminalSequence` field.
# Opt-in via the HARNESS_TERMINAL_NOTIFY env variable.
#
# Usage: source this file, then call build_terminal_sequence "<title>" "<body>"
#        to print the OSC sequence to stdout. Returns empty string when env is unset.
#
# Env: HARNESS_TERMINAL_NOTIFY (optional)
#   unset / "0" : no sequence output
#   "1" / "bell" : BEL (\x07)
#   "title"     : OSC 0 window title update
#   "osc9"      : OSC 9 macOS / iTerm notification
#   "notify"    : OSC 777 KDE/GNOME desktop notification
#
# Security:
#   - Control characters are stripped from title/body to prevent terminal corruption
#   - Printable non-ASCII characters pass through; ESC/BEL/ST and similar are excluded

set -euo pipefail

# Strip control characters (0x00-0x1F, 0x7F) from input.
# Args:
#   $1: input string
# Stdout: input with control characters removed
_terminal_notify_sanitize() {
  printf '%s' "${1:-}" | tr -d '\000-\037\177' 2>/dev/null || true
}

# Build a terminal escape sequence.
# Args:
#   $1: title (e.g. "Build complete")
#   $2: body (optional, used by OSC 777 notify mode only)
# Stdout: constructed sequence string (raw bytes)
build_terminal_sequence() {
  local mode="${HARNESS_TERMINAL_NOTIFY:-}"
  local title body
  title="$(_terminal_notify_sanitize "${1:-}")"
  body="$(_terminal_notify_sanitize "${2:-}")"

  # Opt-in: no env means no output.
  case "${mode}" in
    ''|0) return 0 ;;
  esac

  # Bell does not need a title — it fires even with an empty title.
  # All other modes require a non-empty title.
  if [ "${mode}" != "1" ] && [ "${mode}" != "bell" ] && [ -z "${title}" ]; then
    return 0
  fi

  local ESC BEL
  ESC=$'\x1b'
  BEL=$'\x07'

  case "${mode}" in
    1|bell)
      printf '%s' "${BEL}"
      ;;
    title)
      printf '%s]0;%s%s' "${ESC}" "${title}" "${BEL}"
      ;;
    osc9)
      printf '%s]9;%s%s' "${ESC}" "${title}" "${BEL}"
      ;;
    notify)
      # OSC 777;notify;<title>;<body><BEL>
      if [ -n "${body}" ]; then
        printf '%s]777;notify;%s;%s%s' "${ESC}" "${title}" "${body}" "${BEL}"
      else
        printf '%s]777;notify;%s%s' "${ESC}" "${title}" "${BEL}"
      fi
      ;;
    *)
      # Unknown values are silently ignored (documented range only).
      ;;
  esac
}

# Encode a built sequence as a JSON string literal (including surrounding quotes).
# Primary path uses jq for correct encoding; fallback for environments without jq.
# Note: the fallback produces best-effort output only — use jq in production.
# Args:
#   $1: sequence (raw bytes)
# Stdout: JSON string literal (with quotes) ready for embedding in a JSON object
encode_terminal_sequence_json() {
  local seq="${1:-}"
  if [ -z "${seq}" ]; then
    return 0
  fi
  if command -v jq >/dev/null 2>&1; then
    # jq -Rs encodes raw input as a JSON string (with surrounding quotes).
    printf '%s' "${seq}" | jq -Rs . 2>/dev/null || printf '""'
  else
    # Fallback: escape backslashes and quotes, then substitute ESC/BEL visually.
    # This is a best-effort encoding for environments without jq.
    local out
    out="$(printf '%s' "${seq}" \
      | sed -e 's/\\/\\\\/g' \
            -e 's/"/\\"/g' \
            -e $'s/\x1b/\\\\^[/g' \
            -e $'s/\x07/\\\\^G/g')"
    printf '"%s"' "${out}"
  fi
}
