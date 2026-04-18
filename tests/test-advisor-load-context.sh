#!/usr/bin/env bash
# test-advisor-load-context.sh — end-to-end test for the scoped context
# loader used by the advisor agent (Phase 73.2).
#
# Exercises:
#   - All four sources (trace, git_diff, session_log, patterns)
#   - Missing files produce placeholders, not errors
#   - Per-source 10 KiB cap applies head+tail trim with a visible marker
#   - Unknown source values are skipped with a warning line
#   - Total context across 4 sources stays under 20 KiB (DoD requirement)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
LOADER="$REPO_ROOT/harness/scripts/advisor-load-context.sh"

[[ -f "$LOADER" ]] || { echo "FAIL: loader not found at $LOADER"; exit 1; }

TEST_DIR="$(mktemp -d "${TMPDIR:-/tmp}/advisor-ctx-test.XXXXXX")"
cleanup() { rm -rf "$TEST_DIR"; }
trap cleanup EXIT

cd "$TEST_DIR"

# -- fixture: a real-ish project with each of the 4 sources populated --

git init -q
git config user.email test@example.com
git config user.name test

mkdir -p .claude/state/traces .claude/memory

# Plans.md (not strictly read by loader, but present for realism)
cat > Plans.md <<'EOF'
| 73.4 | integration test | - | - | cc:WIP |
EOF

# Trace file — valid trace.v1 JSONL events
cat > .claude/state/traces/73.4.jsonl <<'EOF'
{"schema":"trace.v1","ts":"2026-04-18T10:00:00Z","task_id":"73.4","event_type":"task_start","payload":{"description":"integration test"}}
{"schema":"trace.v1","ts":"2026-04-18T10:01:00Z","task_id":"73.4","event_type":"tool_call","payload":{"tool":"Edit","args_summary":"file_path=foo.go"}}
{"schema":"trace.v1","ts":"2026-04-18T10:02:00Z","task_id":"73.4","event_type":"error","payload":{"error_signature":"compile error in foo.go","raw_error":"..."}}
{"schema":"trace.v1","ts":"2026-04-18T10:03:00Z","task_id":"73.4","event_type":"fix_attempt","payload":{"approach":"switch to typed API"}}
EOF

# Session log with a mention of the task id + context
cat > .claude/memory/session-log.md <<'EOF'
# Session Log

## Session: 2026-04-17
- Started Phase 73 planning
- Decided on defaults for 73.4 loader

## Session: 2026-04-18
- Worker hit a compile error on 73.4
- Escalated to advisor after 3 retries
- Advisor returned CORRECTION; fix landed
EOF

# Patterns.md with P-headings — only headings should surface
cat > .claude/memory/patterns.md <<'EOF'
# Patterns

## P1: Declarative rule table pattern #guardrails
...
## P10: Per-task execution trace as causal-history layer #observability
...
EOF

# Commit with "(73.4)" in the message — loader uses this as the task-start commit.
git add Plans.md .claude/
git commit -q -m "feat(test): seed fixtures (73.4)"
# Additional change post-start so git_diff has content to show
echo "new work in progress" > work-in-progress.txt
git add work-in-progress.txt
git commit -q -m "progress (73.4): wip"

# -- run the loader --
output=$("$LOADER" --task 73.4 --sources trace,git_diff,session_log,patterns)

fail() { echo "FAIL: $*"; echo "--- output ---"; echo "$output" | head -50; exit 1; }

# All 4 section headings present
for src in trace git_diff session_log patterns; do
    echo "$output" | grep -q "^## source: $src " || fail "missing section for $src"
done

# Trace section contains the task_start event marker
echo "$output" | grep -q 'event_type":"task_start"' || fail "trace content missing task_start"

# git_diff section includes content from commits (file additions)
echo "$output" | grep -q '^+new work in progress' || fail "git_diff missing post-start diff"

# session_log contains the "73.4" match line with context
echo "$output" | grep -q 'compile error on 73.4' || fail "session_log missing task mention context"

# patterns section is just headings — the fixture has literal "..." body
# lines between H2s, which must NOT appear in loader output.
echo "$output" | grep -q 'P10: Per-task execution trace' || fail "patterns missing P10 heading"
patterns_section=$(echo "$output" | awk '/^## source: patterns/,/^## source: (trace|git_diff|session_log)$/' | tail -n +2)
if echo "$patterns_section" | grep -qxF '...'; then
    fail "patterns output leaked body content (literal '...' line from fixture)"
fi

# Total output < 20 KiB (DoD)
total=$(printf '%s' "$output" | wc -c | tr -d ' ')
if (( total >= 20480 )); then
    fail "total output $total bytes exceeds 20 KiB cap"
fi

# -- missing task: should produce placeholders but not error --
output_missing=$("$LOADER" --task 999.99 --sources trace,session_log)
echo "$output_missing" | grep -q "no trace file at" || fail "missing-trace should produce placeholder"
echo "$output_missing" | grep -q "no session-log mentions of 999.99" || fail "missing session_log should produce placeholder"

# -- trim behaviour: trace file larger than 10 KiB --
trace_path="$TEST_DIR/.claude/state/traces/big.jsonl"
python3 -c "
import json
# Use compact separators to match real trace output from the Go writer
# (json.Marshal emits no whitespace). Assertions below grep for compact form.
for i in range(200):
    line = json.dumps({
        'schema': 'trace.v1',
        'ts': '2026-04-18T10:%02d:%02d.000Z' % (i // 60, i % 60),
        'task_id': 'big',
        'event_type': 'tool_call',
        'payload': {'tool': 'Edit', 'args_summary': 'file_path=test.go', 'i': i},
    }, separators=(',', ':'))
    print(line)
" > "$trace_path"
trace_size=$(wc -c < "$trace_path" | tr -d ' ')
(( trace_size > 10240 )) || fail "fixture trace ($trace_size bytes) should be >10KiB; test assumption broken"

output_trim=$("$LOADER" --task big --sources trace)
echo "$output_trim" | grep -q 'trimmed to 10240' || fail "size heading should show trimmed-to value"
echo "$output_trim" | grep -q 'bytes trimmed' || fail "trim marker should appear in body"
# Head should include i=0 (task start); tail should include i=199 (most recent)
echo "$output_trim" | grep -q '"i":0' || fail "head should preserve task-start events"
echo "$output_trim" | grep -q '"i":199' || fail "tail should preserve most-recent events"

# -- unknown source skip --
output_unknown=$("$LOADER" --task 73.4 --sources trace,bogus)
echo "$output_unknown" | grep -q "^## source: bogus (unknown" || fail "unknown source should emit warning section"

# -- missing flags --
if "$LOADER" --sources trace 2>/dev/null; then
    fail "missing --task should exit non-zero"
fi
if "$LOADER" --task 73.4 2>/dev/null; then
    fail "missing --sources should exit non-zero"
fi

echo "PASS test-advisor-load-context.sh"
