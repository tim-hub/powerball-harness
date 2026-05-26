#!/usr/bin/env bash
# opencode-companion.sh — Proxy to official opencode-plugin-cc companion
#
# Dynamically discovers and calls opencode-companion.mjs from the official plugin
# tasict/opencode-plugin-cc. Harness skills and agents call OpenCode through this
# proxy rather than invoking raw `opencode run` directly.
#
# Usage:
#   bash scripts/opencode-companion.sh task "Fix the bug"
#   bash scripts/opencode-companion.sh review --base HEAD~3
#   bash scripts/opencode-companion.sh setup --json
#   bash scripts/opencode-companion.sh status
#   bash scripts/opencode-companion.sh result <job-id>
#   bash scripts/opencode-companion.sh cancel <job-id>
#
# Subcommands: task, review, adversarial-review, setup, status, result, cancel
#
# Effort propagation:
#   When the task subcommand runs, effort is calculated via calculate-effort.sh
#   and passed to the companion with the --effort flag. If calculate-effort.sh
#   is absent, falls back to the OPENCODE_EFFORT environment variable (default: medium).

set -euo pipefail

# Search for the official plugin companion
# Looks in the Claude plugin directory only (no ~/.opencode/ equivalent).
PLUGIN_DIRS=()
[ -d "${HOME}/.claude/plugins" ] && PLUGIN_DIRS+=("${HOME}/.claude/plugins")

COMPANION=""
if [ "${#PLUGIN_DIRS[@]}" -gt 0 ]; then
  # Extract version segment from path and compare numerically (macOS BSD sort compatible)
  COMPANION=$(find "${PLUGIN_DIRS[@]}" -name "opencode-companion.mjs" \
    \( -path "*/tasict-opencode-plugin-cc/*" \
       -o -path "*/opencode-plugin-cc/*" \
       -o -path "*/plugins/opencode/*" \) \
    2>/dev/null \
    | awk -F/ '{version="0.0.0"; for(i=1;i<=NF;i++){if($i~/^[0-9]+\.[0-9]+(\.[0-9]+)?$/){version=$i}} print version,$0}' \
    | sort -t. -k1,1n -k2,2n -k3,3n \
    | tail -1 \
    | cut -d' ' -f2-)
fi

if [ -z "$COMPANION" ]; then
  echo "ERROR: opencode-plugin-cc not found." >&2
  echo "Install: plugin marketplace add tasict/opencode-plugin-cc" >&2
  echo "Or run: /opencode:setup" >&2
  exit 1
fi

# ---- Effort propagation (task subcommand only) ----
# For the task subcommand, calculate effort from the task description and pass
# it via the --effort flag. Falls back to OPENCODE_EFFORT env var (default: medium)
# if calculate-effort.sh is not present.
SUBCOMMAND="${1:-}"
if [ "$SUBCOMMAND" = "task" ]; then
  SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
  EFFORT_SCRIPT="${SCRIPT_DIR}/calculate-effort.sh"

  # Skip if --effort flag already specified, or if --resume-last is present
  # --resume-last carries a continuation prompt (e.g. "continue from where you left off"),
  # which makes effort calculation inaccurate
  EFFORT_ALREADY_SET=0
  for arg in "$@"; do
    if [ "$arg" = "--effort" ] || echo "$arg" | grep -qE '^--effort='; then
      EFFORT_ALREADY_SET=1
      break
    fi
    if [ "$arg" = "--resume-last" ] || [ "$arg" = "--resume" ]; then
      EFFORT_ALREADY_SET=1
      break
    fi
  done

  if [ "$EFFORT_ALREADY_SET" -eq 0 ]; then
    # Extract task description from arguments (last non-flag argument)
    # Boolean flags (no value): --write, --resume-last, --json, --full-auto, --ephemeral, --oss, --skip-git-repo-check
    # Value flags (consume next arg): --base, --effort, --model, -m, -i, --image, -c, --config, -C, --cd, --add-dir, --output-schema, -o, --output-last-message, --color, --enable, --disable, --local-provider
    # Unknown --* flags → treated conservatively as value flags (consume next arg)
    TASK_DESC=""
    EXPECT_VALUE=""
    for arg in "${@:2}"; do
      if [ -n "$EXPECT_VALUE" ]; then
        # Value for the previous flag — skip
        EXPECT_VALUE=""
        continue
      fi
      case "$arg" in
        --write|--resume-last|--json|--full-auto|--ephemeral|--oss|--skip-git-repo-check|--dangerously-bypass-approvals-and-sandbox|--background|--resume|--fresh)
          # Boolean flag with no value — skip
          ;;
        --base|--effort|--model|-m|-i|--image|-c|--config|-C|--cd|--add-dir|--output-schema|-o|--output-last-message|--color|--enable|--disable|--local-provider)
          # Explicitly value-bearing flag
          EXPECT_VALUE="$arg"
          ;;
        --*)
          # Unknown flag — treated conservatively as value-bearing (avoid using next arg as TASK_DESC)
          EXPECT_VALUE="$arg"
          ;;
        *)
          # Non-flag argument = task description
          TASK_DESC="$arg"
          ;;
      esac
    done

    # Calculate effort
    COMPUTED_EFFORT=""
    if [ -f "$EFFORT_SCRIPT" ]; then
      if [ -n "$TASK_DESC" ]; then
        COMPUTED_EFFORT=$(bash "$EFFORT_SCRIPT" "$TASK_DESC" 2>/dev/null || true)
      elif [ ! -t 0 ]; then
        # stdin available (pipe): read content and calculate effort
        STDIN_CONTENT=$(cat)
        if [ -n "$STDIN_CONTENT" ]; then
          COMPUTED_EFFORT=$(echo "$STDIN_CONTENT" | bash "$EFFORT_SCRIPT" 2>/dev/null || true)
          # Re-setup stdin (pass to companion via here-string)
          exec node "$COMPANION" "$@" --effort "${COMPUTED_EFFORT:-medium}" <<< "$STDIN_CONTENT"
        fi
        # If stdin is empty (e.g. </dev/null), fall through to normal flow
      fi
    fi

    # Fallback: OPENCODE_EFFORT env var → medium
    if [ -z "$COMPUTED_EFFORT" ]; then
      COMPUTED_EFFORT="${OPENCODE_EFFORT:-medium}"
    fi

    # Pass only effort levels supported by the companion
    case "$COMPUTED_EFFORT" in
      none|minimal|low|medium|high|xhigh) ;;
      *) COMPUTED_EFFORT="medium" ;;
    esac

    exec node "$COMPANION" "$@" --effort "$COMPUTED_EFFORT"
  fi
fi

exec node "$COMPANION" "$@"
