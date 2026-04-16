#!/bin/bash
# test-harness-loop-flow.sh
# harness-loop flow.md の contract_path / reviewer_profile 分岐の回帰テスト

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "${SCRIPT_DIR}")"
FLOW_FILE="${PROJECT_ROOT}/skills/harness-loop/references/flow.md"

fail() {
  echo "FAIL: $1" >&2
  exit 1
}

[ -f "${FLOW_FILE}" ] || fail "flow.md が見つかりません"

grep -q 'CONTRACT_PATH=".claude/state/contracts/${task_id}.sprint-contract.json"' "${FLOW_FILE}" \
  || fail "Step 2 に CONTRACT_PATH 初期化がありません"

if grep -q 'task_contract_path' "${FLOW_FILE}"; then
  fail "flow.md に削除済みの task_contract_path 参照が残っています"
fi

grep -q 'REVIEWER_PROFILE=$(jq -r '\''\.review\.reviewer_profile // "static"'\'' "${CONTRACT_PATH}"' "${FLOW_FILE}" \
  || fail "reviewer_profile 読み取りが CONTRACT_PATH を参照していません"

grep -q 'generate-browser-review-artifact.sh "${CONTRACT_PATH}"' "${FLOW_FILE}" \
  || fail "browser profile 分岐が CONTRACT_PATH を使っていません"

echo "test-harness-loop-flow: ok"
