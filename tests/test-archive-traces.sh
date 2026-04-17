#!/usr/bin/env bash
# test-archive-traces.sh — end-to-end test for
# harness/skills/maintenance/scripts/archive-traces.sh.
#
# Exercises:
#   - Archival of cc:Done + aged traces
#   - Retention of cc:Done + recent traces
#   - Retention of cc:WIP / cc:TODO traces regardless of age
#   - Idempotency (second run is a no-op)
#   - JSONL validity survives the move

set -euo pipefail

# project-root: derive from this test's location, not cwd
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
ARCHIVE_SCRIPT="${REPO_ROOT}/harness/skills/maintenance/scripts/archive-traces.sh"

if [[ ! -x "$ARCHIVE_SCRIPT" && ! -f "$ARCHIVE_SCRIPT" ]]; then
    echo "FAIL: archive script not found at $ARCHIVE_SCRIPT"
    exit 1
fi

TEST_DIR="$(mktemp -d "${TMPDIR:-/tmp}/archive-traces-test.XXXXXX")"
cleanup() { rm -rf "$TEST_DIR"; }
trap cleanup EXIT

# -- setup fake project --
cd "$TEST_DIR"
git init -q
git config user.email test@example.com
git config user.name test
mkdir -p .claude/state/traces .claude/memory

cat > Plans.md <<'EOF'
# Test Plans

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 1.1 | Old done task | | | cc:Done [abc1234] |
| 1.2 | Recent done task | | | cc:Done [def5678] |
| 1.3 | In progress | | | cc:WIP |
| 1.4 | Not started | | | cc:TODO |
| 2.1.1 | Rotated case | | | cc:Done [999888a] |
EOF

# Write trace content — each valid JSONL per trace.v1
write_trace() {
    local path="$1"
    printf '%s\n' \
        '{"schema":"trace.v1","ts":"2026-01-01T00:00:00Z","task_id":"sample","event_type":"task_start","payload":{}}' \
        > "$path"
}

write_trace .claude/state/traces/1.1.jsonl
write_trace .claude/state/traces/1.2.jsonl
write_trace .claude/state/traces/1.3.jsonl
write_trace .claude/state/traces/1.4.jsonl
# Rotated form: 2.1 is the base task id, 2.1.1 is the rotation.
# Plans.md has 2.1.1 as cc:Done, so this hits the resolve_task_id branch.
write_trace .claude/state/traces/2.1.jsonl
write_trace .claude/state/traces/2.1.1.jsonl

# Age the "old done" files past the default 30-day retention.
# touch -t uses YYYYMMDDhhmm; Jan 1 2026 is >3 months before test date.
touch -t 202601010000 .claude/state/traces/1.1.jsonl
touch -t 202601010000 .claude/state/traces/2.1.jsonl
touch -t 202601010000 .claude/state/traces/2.1.1.jsonl

# -- run the archive script --
bash "$ARCHIVE_SCRIPT" > /tmp/archive-traces-run1.log 2>&1

# -- assertions --
fail() { echo "FAIL: $*"; cat /tmp/archive-traces-run1.log; exit 1; }

# 1.1: cc:Done + old -> archived
[[ ! -f .claude/state/traces/1.1.jsonl ]] || fail "1.1 should have been archived"
ARCHIVED_1_1="$(find .claude/memory/archive/traces -name 1.1.jsonl 2>/dev/null | head -n 1 || true)"
[[ -n "$ARCHIVED_1_1" ]] || fail "no archived 1.1.jsonl under .claude/memory/archive/traces/"

# Validate JSONL still parses after move
python3 -c "
import json, sys
for line in open(sys.argv[1]):
    line = line.strip()
    if not line: continue
    obj = json.loads(line)
    assert obj.get('schema') == 'trace.v1', 'bad schema: %s' % obj
" "$ARCHIVED_1_1" || fail "archived 1.1.jsonl is not valid JSONL"

# 1.2: cc:Done + recent -> kept
[[ -f .claude/state/traces/1.2.jsonl ]] || fail "1.2 should have stayed (recent)"

# 1.3: cc:WIP -> kept regardless of age
[[ -f .claude/state/traces/1.3.jsonl ]] || fail "1.3 should have stayed (WIP)"

# 1.4: cc:TODO (no cc:Done marker) -> kept
[[ -f .claude/state/traces/1.4.jsonl ]] || fail "1.4 should have stayed (TODO)"

# 2.1.1 (rotated file, base task 2.1 is cc:Done via the row "2.1.1 | ... cc:Done")
# Wait — Plans.md has 2.1.1 as cc:Done, not 2.1. So 2.1.jsonl's task_is_done check
# for "2.1" would FAIL (no such Plans.md row). And 2.1.1.jsonl's check for "2.1.1"
# would succeed directly.
#
# Expected behavior:
#   2.1.jsonl   -> task_id="2.1" -> task_is_done("2.1") fails -> KEPT
#   2.1.1.jsonl -> task_id="2.1.1" via resolve_task_id direct match -> ARCHIVED
[[ -f .claude/state/traces/2.1.jsonl ]] || fail "2.1 should have stayed (no Plans.md row)"
[[ ! -f .claude/state/traces/2.1.1.jsonl ]] || fail "2.1.1 should have been archived"
ARCHIVED_2_1_1="$(find .claude/memory/archive/traces -name 2.1.1.jsonl 2>/dev/null | head -n 1 || true)"
[[ -n "$ARCHIVED_2_1_1" ]] || fail "no archived 2.1.1.jsonl"

# -- idempotency: second run must be a no-op --
before_state="$(find .claude -type f | sort)"
bash "$ARCHIVE_SCRIPT" > /tmp/archive-traces-run2.log 2>&1
after_state="$(find .claude -type f | sort)"
if [[ "$before_state" != "$after_state" ]]; then
    echo "FAIL: second run changed filesystem state"
    diff <(echo "$before_state") <(echo "$after_state") || true
    exit 1
fi

# -- DRY_RUN flag does not move files --
# Re-age 1.2 so it would be eligible, but run with DRY_RUN.
touch -t 202601010000 .claude/state/traces/1.2.jsonl
DRY_RUN=1 bash "$ARCHIVE_SCRIPT" > /tmp/archive-traces-dryrun.log 2>&1
[[ -f .claude/state/traces/1.2.jsonl ]] || fail "DRY_RUN must not move files"
grep -q 'DRY-RUN would move' /tmp/archive-traces-dryrun.log || fail "DRY_RUN did not announce action"

echo "PASS test-archive-traces.sh"
