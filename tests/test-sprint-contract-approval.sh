#!/bin/bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

cat > "${TMP_DIR}/Plans.md" <<'EOF'
| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 32.1.1 | contract を作る | runtime validation を contract に載せる | 32.0.1 | cc:TODO |
EOF

CONTRACT_PATH="${TMP_DIR}/contract.json"
"${PROJECT_ROOT}/scripts/generate-sprint-contract.sh" "32.1.1" "${TMP_DIR}/Plans.md" "$CONTRACT_PATH" >/dev/null

if "${PROJECT_ROOT}/scripts/ensure-sprint-contract-ready.sh" "$CONTRACT_PATH" >/dev/null 2>&1; then
  echo "contract should fail before approval"
  exit 1
fi

"${PROJECT_ROOT}/scripts/enrich-sprint-contract.sh" "$CONTRACT_PATH" \
  --check "DoD を reviewer 観点で再確認" \
  --non-goal "UI polish は今回やらない" \
  --risk "needs-spike" \
  --note "reviewer checked runtime path" \
  --approve >/dev/null

"${PROJECT_ROOT}/scripts/ensure-sprint-contract-ready.sh" "$CONTRACT_PATH" >/dev/null

jq -e '
  .review.status == "approved" and
  (.contract.non_goals | length) == 1 and
  (.contract.risk_flags | index("needs-spike")) != null
' "$CONTRACT_PATH" >/dev/null

echo "test-sprint-contract-approval: ok"
