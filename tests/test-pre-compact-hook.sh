#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "${TMP_DIR}"' EXIT
HARNESS_BIN="${TMP_DIR}/harness"

(cd "${ROOT_DIR}/go" && go build -o "${HARNESS_BIN}" ./cmd/harness)

run_pre_compact() {
  local workdir="$1"
  local session_id="$2"
  local agent_type="${3:-}"
  local input
  input=$(printf '{"session_id":"%s","cwd":"%s","agent_type":"%s"}' "$session_id" "$workdir" "$agent_type")
  PRE_COMPACT_OUTPUT=""
  PRE_COMPACT_STATUS=0
  set +e
  PRE_COMPACT_OUTPUT=$(cd "${ROOT_DIR}" && HARNESS_SESSION_ROLE="${agent_type}" CLAUDE_SESSION_ID="${session_id}" "${HARNESS_BIN}" pre-compact <<<"${input}" 2>&1)
  PRE_COMPACT_STATUS=$?
  set -e
}

assert_blocked() {
  local label="$1"
  local workdir="$2"
  local session_id="$3"
  local agent_type="${4:-}"
  run_pre_compact "${workdir}" "${session_id}" "${agent_type}"
  if [[ "${PRE_COMPACT_STATUS}" -ne 2 ]]; then
    echo "[${label}] expected exit 2, got ${PRE_COMPACT_STATUS}"
    echo "${PRE_COMPACT_OUTPUT}"
    exit 1
  fi
  if ! grep -q '"decision":"block"' <<<"${PRE_COMPACT_OUTPUT}"; then
    echo "[${label}] expected block decision JSON"
    echo "${PRE_COMPACT_OUTPUT}"
    exit 1
  fi
}

assert_allowed() {
  local label="$1"
  local workdir="$2"
  local session_id="$3"
  local agent_type="${4:-}"
  run_pre_compact "${workdir}" "${session_id}" "${agent_type}"
  if [[ "${PRE_COMPACT_STATUS}" -ne 0 ]]; then
    echo "[${label}] expected exit 0, got ${PRE_COMPACT_STATUS}"
    echo "${PRE_COMPACT_OUTPUT}"
    exit 1
  fi
  if [[ -n "${PRE_COMPACT_OUTPUT}" ]]; then
    echo "[${label}] expected no output on allow"
    echo "${PRE_COMPACT_OUTPUT}"
    exit 1
  fi
}

worker_repo="${TMP_DIR}/worker"
mkdir -p "${worker_repo}/.claude/state/locks/loop-session.lock.d"
cat > "${worker_repo}/.claude/state/locks/loop-session.lock.d/meta.json" <<'JSON'
{"session_id":"sess-worker"}
JSON
assert_blocked "worker-lock" "${worker_repo}" "sess-worker" "worker"

reviewer_repo="${TMP_DIR}/reviewer"
mkdir -p "${reviewer_repo}"
assert_allowed "reviewer-allow" "${reviewer_repo}" "sess-reviewer" "reviewer"

dirty_repo="${TMP_DIR}/dirty"
mkdir -p "${dirty_repo}"
git -C "${dirty_repo}" init >/dev/null 2>&1
git -C "${dirty_repo}" config user.name "Harness Test"
git -C "${dirty_repo}" config user.email "harness@example.com"
cat > "${dirty_repo}/Plans.md" <<'EOF_PLANS'
| 44.2.1 | pre-compact | done | none | cc:TODO |
EOF_PLANS
git -C "${dirty_repo}" add Plans.md
git -C "${dirty_repo}" commit -m "test: add plans" >/dev/null 2>&1
cat >> "${dirty_repo}/Plans.md" <<'EOF_PLANS'
| 44.2.2 | monitors | done | 44.2.1 | cc:TODO |
EOF_PLANS
assert_blocked "plans-dirty" "${dirty_repo}" "sess-main"

echo "OK"
