#!/bin/bash
# test-team-composition-advisor.sh
# Advisor を含む役割分担ドキュメントの整合性テスト

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "${SCRIPT_DIR}")"

ADVISOR_FILE="${PROJECT_ROOT}/agents/advisor.md"
TEAM_FILE="${PROJECT_ROOT}/agents/team-composition.md"
WORKER_FILE="${PROJECT_ROOT}/agents/worker.md"
REVIEWER_FILE="${PROJECT_ROOT}/agents/reviewer.md"

fail() {
  echo "FAIL: $1" >&2
  exit 1
}

for file in "${ADVISOR_FILE}" "${TEAM_FILE}" "${WORKER_FILE}" "${REVIEWER_FILE}"; do
  [ -f "${file}" ] || fail "missing file: ${file}"
done

grep -q 'advisor-response.v1' "${ADVISOR_FILE}" \
  || fail "advisor.md に advisor-response.v1 がありません"

grep -q 'PLAN / CORRECTION / STOP' "${ADVISOR_FILE}" \
  || fail "advisor.md に decision 3 種がありません"

grep -q 'コードを書かない' "${ADVISOR_FILE}" \
  || fail "advisor.md に非実装ルールがありません"

grep -q 'Harness の4エージェント構成' "${TEAM_FILE}" \
  || fail "team-composition.md に 4 役構成の説明がありません"

grep -q 'advisor-request.v1' "${WORKER_FILE}" \
  || fail "worker.md に advisor-request.v1 がありません"

grep -q 'Advisor は別ロールであり、Reviewer の代替ではない' "${REVIEWER_FILE}" \
  || fail "reviewer.md に advisor 非代替の明記がありません"

echo "test-team-composition-advisor: ok"
