#\!/usr/bin/env bash
# enable-1h-cache.sh — opt-in 1-hour Anthropic prompt cache for long sessions.
#
# Usage:
#   source harness/scripts/enable-1h-cache.sh
#
# Effect: sets ANTHROPIC_CACHE_CONTROL=max-age=3600 and exports it so that
# all Claude Code subprocesses (Workers, Reviewers, hooks) inherit the setting.
#
# When to use: source this at the start of a breezing session expected to
# exceed 30 minutes. The 1-hour TTL keeps the system prompt cached across
# all spawned agents, reducing token cost and latency significantly.
#
# To revert: unset ANTHROPIC_CACHE_CONTROL

export ANTHROPIC_CACHE_CONTROL="max-age=3600"
echo "[harness] 1-hour prompt cache enabled (ANTHROPIC_CACHE_CONTROL=max-age=3600)"
