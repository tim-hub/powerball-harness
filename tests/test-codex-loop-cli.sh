#!/bin/bash
# test-codex-loop-cli.sh
# Codex-native long-running loop CLI の統合テスト

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
LOOP_SCRIPT="${REPO_ROOT}/scripts/codex-loop.sh"

PASS_COUNT=0
FAIL_COUNT=0

pass() {
  echo "[PASS] $1"
  PASS_COUNT=$((PASS_COUNT + 1))
}

fail() {
  echo "[FAIL] $1" >&2
  FAIL_COUNT=$((FAIL_COUNT + 1))
}

setup_fake_tools() {
  local workdir="$1"
  local mode="$2"
  mkdir -p "${workdir}/bin"

  cat > "${workdir}/bin/fake-generate-contract.sh" <<'EOF'
#!/bin/bash
set -euo pipefail
task_id="$1"
state_dir="${PROJECT_ROOT}/.claude/state/contracts"
mkdir -p "${state_dir}"
path="${state_dir}/${task_id}.sprint-contract.json"
cat > "${path}" <<JSON
{
  "task": { "id": "${task_id}", "title": "Task ${task_id}" },
  "review": { "status": "draft", "reviewer_profile": "static" }
}
JSON
echo "${path}"
EOF

  cat > "${workdir}/bin/fake-enrich-contract.sh" <<'EOF'
#!/bin/bash
set -euo pipefail
echo "$1"
EOF

  cat > "${workdir}/bin/fake-ensure-contract.sh" <<'EOF'
#!/bin/bash
set -euo pipefail
test -f "$1"
echo "$1"
EOF

  cat > "${workdir}/bin/fake-runtime-review.sh" <<'EOF'
#!/bin/bash
set -euo pipefail
contract="$1"
output="$2"
cat > "${output}" <<JSON
{
  "schema_version": "runtime-review.v1",
  "task": { "id": "fake" },
  "reviewer_profile": "static",
  "verdict": "APPROVE",
  "checks": []
}
JSON
echo "${output}"
EOF

  cat > "${workdir}/bin/fake-write-review-result.sh" <<'EOF'
#!/bin/bash
set -euo pipefail
input="$1"
commit_hash="$2"
output="$3"
cp "${input}" "${output}"
echo "${output}"
EOF

  cat > "${workdir}/bin/fake-plateau.sh" <<'EOF'
#!/bin/bash
set -euo pipefail
echo "STATUS: PIVOT_NOT_REQUIRED"
echo "ENTRIES: 3"
echo "JACCARD_AVG: 0.1000"
echo "REASON: progress is fine"
exit 0
EOF

  cat > "${workdir}/bin/fake-checkpoint.sh" <<'EOF'
#!/bin/bash
set -euo pipefail
touch "${PROJECT_ROOT}/.claude/state/checkpoint-called"
EOF

  cat > "${workdir}/bin/fake-validate.sh" <<'EOF'
#!/bin/bash
set -euo pipefail
exit 0
EOF

  cat > "${workdir}/bin/fake-mem.sh" <<'EOF'
#!/bin/bash
set -euo pipefail
echo '{"ok":true,"items":[]}'
EOF

  cat > "${workdir}/bin/fake-companion.sh" <<EOF
#!/bin/bash
set -euo pipefail
STATE_DIR="${workdir}/companion-state"
mkdir -p "\${STATE_DIR}"
JOB_FILE="\${STATE_DIR}/job.json"
COUNTER_FILE="\${STATE_DIR}/status-count"
MODE="${mode}"

read_json() {
  python3 - "\$1" <<'PY'
import json, os, sys
path = sys.argv[1]
if not os.path.exists(path):
    print("{}")
else:
    with open(path, "r", encoding="utf-8") as fh:
        print(json.dumps(json.load(fh)))
PY
}

write_job() {
  python3 - "\$JOB_FILE" "\$1" <<'PY'
import json, sys
path, status = sys.argv[1:3]
payload = {
  "id": "fake-job-1",
  "status": status,
  "title": "Codex Task",
  "phase": "running" if status == "running" else ("done" if status == "completed" else status)
}
with open(path, "w", encoding="utf-8") as fh:
    json.dump(payload, fh)
PY
}

complete_task() {
  python3 - "\${PROJECT_ROOT}/Plans.md" <<'PY'
from pathlib import Path
import sys
path = Path(sys.argv[1])
text = path.read_text(encoding="utf-8")
text = text.replace("cc:TODO", "cc:完了 [fake]", 1)
path.write_text(text, encoding="utf-8")
PY
  git -C "\${PROJECT_ROOT}" add Plans.md >/dev/null 2>&1
  git -C "\${PROJECT_ROOT}" commit -m "codex-loop test completion" >/dev/null 2>&1 || true
}

cmd="\${1:-}"
shift || true
case "\${cmd}" in
  task)
    write_job "queued"
    printf '0' > "\${COUNTER_FILE}"
    printf '{"jobId":"fake-job-1","status":"queued","title":"Codex Task","logFile":"%s"}\n' "\${STATE_DIR}/job.log"
    ;;
  status)
    job_id="\${1:-}"
    shift || true
    count=0
    if [ -f "\${COUNTER_FILE}" ]; then
      count="\$(cat "\${COUNTER_FILE}")"
    fi
    if [ "\${MODE}" = "complete" ]; then
      if [ "\${count}" = "0" ]; then
        write_job "running"
        printf '1' > "\${COUNTER_FILE}"
        printf '{"workspaceRoot":"%s","job":{"id":"fake-job-1","status":"running","phase":"running","title":"Codex Task"}}\n' "\${PROJECT_ROOT}"
      else
        complete_task
        write_job "completed"
        printf '2' > "\${COUNTER_FILE}"
        printf '{"workspaceRoot":"%s","job":{"id":"fake-job-1","status":"completed","phase":"done","title":"Codex Task"}}\n' "\${PROJECT_ROOT}"
      fi
    else
      if [ -f "\${STATE_DIR}/cancelled" ]; then
        write_job "cancelled"
        printf '{"workspaceRoot":"%s","job":{"id":"fake-job-1","status":"cancelled","phase":"cancelled","title":"Codex Task"}}\n' "\${PROJECT_ROOT}"
      else
        write_job "running"
        printf '{"workspaceRoot":"%s","job":{"id":"fake-job-1","status":"running","phase":"running","title":"Codex Task"}}\n' "\${PROJECT_ROOT}"
      fi
    fi
    ;;
  result)
    status="completed"
    if [ -f "\${STATE_DIR}/cancelled" ]; then
      status="cancelled"
    fi
    printf '{"job":{"id":"fake-job-1","status":"%s","title":"Codex Task"},"storedJob":{"status":"%s","result":{"rawOutput":"RESULT: %s\\nsummary"}}}\n' "\${status}" "\${status}" "\$( [ "\${status}" = "completed" ] && echo APPROVED || echo BLOCKED )"
    ;;
  cancel)
    touch "\${STATE_DIR}/cancelled"
    write_job "cancelled"
    printf '{"jobId":"fake-job-1","status":"cancelled","title":"Codex Task"}\n'
    ;;
  *)
    echo "unknown fake companion command: \${cmd}" >&2
    exit 1
    ;;
esac
EOF

  chmod +x "${workdir}"/bin/fake-*.sh
}

setup_repo() {
  local repo="$1"
  mkdir -p "${repo}/.claude/state"
  cat > "${repo}/Plans.md" <<'EOF'
# Plans

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 1 | fake task | done | - | cc:TODO |
EOF
  git -C "${repo}" init >/dev/null 2>&1
  git -C "${repo}" config user.email "loop-test@example.com"
  git -C "${repo}" config user.name "Loop Test"
  git -C "${repo}" add Plans.md
  git -C "${repo}" commit -m "initial" >/dev/null 2>&1
}

poll_for_status() {
  local run_json="$1"
  local expected="$2"
  local tries=0
  while [ "${tries}" -lt 40 ]; do
    if [ -f "${run_json}" ]; then
      local status
      status="$(python3 - "${run_json}" <<'PY'
import json, sys
with open(sys.argv[1], "r", encoding="utf-8") as fh:
    print(json.load(fh).get("status", ""))
PY
)"
      if [ "${status}" = "${expected}" ]; then
        return 0
      fi
    fi
    sleep 0.2
    tries=$((tries + 1))
  done
  return 1
}

run_completion_case() {
  local tmp
  tmp="$(mktemp -d)"
  local repo="${tmp}/repo"
  mkdir -p "${repo}"
  setup_repo "${repo}"
  setup_fake_tools "${tmp}" "complete"

  PROJECT_ROOT="${repo}" \
  CODEX_LOOP_COMPANION="${tmp}/bin/fake-companion.sh" \
  CODEX_LOOP_VALIDATE_SCRIPT="${tmp}/bin/fake-validate.sh" \
  CODEX_LOOP_ENRICH_CONTRACT_SCRIPT="${tmp}/bin/fake-enrich-contract.sh" \
  CODEX_LOOP_ENSURE_CONTRACT_SCRIPT="${tmp}/bin/fake-ensure-contract.sh" \
  CODEX_LOOP_RUNTIME_REVIEW_SCRIPT="${tmp}/bin/fake-runtime-review.sh" \
  CODEX_LOOP_WRITE_REVIEW_RESULT_SCRIPT="${tmp}/bin/fake-write-review-result.sh" \
  CODEX_LOOP_PLATEAU_SCRIPT="${tmp}/bin/fake-plateau.sh" \
  CODEX_LOOP_CHECKPOINT_SCRIPT="${tmp}/bin/fake-checkpoint.sh" \
  CODEX_LOOP_MEM_CLIENT="${tmp}/bin/fake-mem.sh" \
  CODEX_LOOP_GENERATE_CONTRACT_SCRIPT="${tmp}/bin/fake-generate-contract.sh" \
  CODEX_LOOP_POLL_INTERVAL_SEC=1 \
  bash "${LOOP_SCRIPT}" start all --max-cycles 1 --pacing worker >/dev/null

  if poll_for_status "${repo}/.claude/state/codex-loop/run.json" "completed"; then
    pass "completion case: run finished"
  else
    fail "completion case: run did not finish"
  fi

  if grep -q 'cc:完了' "${repo}/Plans.md"; then
    pass "completion case: Plans.md updated"
  else
    fail "completion case: Plans.md was not updated"
  fi

  if [ -f "${repo}/.claude/state/checkpoint-called" ]; then
    pass "completion case: checkpoint executed"
  else
    fail "completion case: checkpoint was not executed"
  fi

  rm -rf "${tmp}"
}

run_stop_case() {
  local tmp
  tmp="$(mktemp -d)"
  local repo="${tmp}/repo"
  mkdir -p "${repo}"
  setup_repo "${repo}"
  setup_fake_tools "${tmp}" "stall"

  PROJECT_ROOT="${repo}" \
  CODEX_LOOP_COMPANION="${tmp}/bin/fake-companion.sh" \
  CODEX_LOOP_VALIDATE_SCRIPT="${tmp}/bin/fake-validate.sh" \
  CODEX_LOOP_ENRICH_CONTRACT_SCRIPT="${tmp}/bin/fake-enrich-contract.sh" \
  CODEX_LOOP_ENSURE_CONTRACT_SCRIPT="${tmp}/bin/fake-ensure-contract.sh" \
  CODEX_LOOP_RUNTIME_REVIEW_SCRIPT="${tmp}/bin/fake-runtime-review.sh" \
  CODEX_LOOP_WRITE_REVIEW_RESULT_SCRIPT="${tmp}/bin/fake-write-review-result.sh" \
  CODEX_LOOP_PLATEAU_SCRIPT="${tmp}/bin/fake-plateau.sh" \
  CODEX_LOOP_CHECKPOINT_SCRIPT="${tmp}/bin/fake-checkpoint.sh" \
  CODEX_LOOP_MEM_CLIENT="${tmp}/bin/fake-mem.sh" \
  CODEX_LOOP_GENERATE_CONTRACT_SCRIPT="${tmp}/bin/fake-generate-contract.sh" \
  CODEX_LOOP_POLL_INTERVAL_SEC=1 \
  bash "${LOOP_SCRIPT}" start all --max-cycles 2 --pacing worker >/dev/null

  sleep 1
  PROJECT_ROOT="${repo}" \
  CODEX_LOOP_COMPANION="${tmp}/bin/fake-companion.sh" \
  bash "${LOOP_SCRIPT}" stop >/dev/null

  if poll_for_status "${repo}/.claude/state/codex-loop/run.json" "stopped"; then
    pass "stop case: run stopped"
  else
    fail "stop case: run did not stop"
  fi

  if [ -f "${tmp}/companion-state/cancelled" ]; then
    pass "stop case: active job was cancelled"
  else
    fail "stop case: cancel was not sent to companion"
  fi

  rm -rf "${tmp}"
}

run_completion_case
run_stop_case

echo "passed=${PASS_COUNT} failed=${FAIL_COUNT}"
[ "${FAIL_COUNT}" -eq 0 ]
