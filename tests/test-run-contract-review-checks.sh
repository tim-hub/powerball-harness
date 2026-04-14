#!/bin/bash

set -euo pipefail
export TMPDIR=/tmp  # Force /tmp for sandboxed execution (sandbox blocks /var/folders)

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
TMP_DIR="$(mktemp -d "/tmp/harness-test.XXXXXX")"
trap 'rm -rf "$TMP_DIR"' EXIT

cat > "${TMP_DIR}/runtime-contract.json" <<'EOF'
{
  "task": {
    "id": "32.2.1",
    "title": "runtime profile"
  },
  "review": {
    "reviewer_profile": "runtime"
  },
  "contract": {
    "runtime_validation": [
      { "label": "pass-check", "command": "echo ok" },
      { "label": "fail-check", "command": "exit 1" }
    ]
  }
}
EOF

OUTPUT="${TMP_DIR}/runtime-review.json"
(cd "$TMP_DIR" && "${PROJECT_ROOT}/harness/scripts/run-contract-review-checks.sh" "${TMP_DIR}/runtime-contract.json" "$OUTPUT" >/dev/null)

jq -e '
  .reviewer_profile == "runtime" and
  .verdict == "REQUEST_CHANGES" and
  (.checks | length) == 2 and
  ([.checks[] | .status] | index("failed")) != null
' "$OUTPUT" >/dev/null

cat > "${TMP_DIR}/static-contract.json" <<'EOF'
{
  "task": {
    "id": "32.2.1",
    "title": "static profile"
  },
  "review": {
    "reviewer_profile": "static"
  },
  "contract": {
    "runtime_validation": []
  }
}
EOF

STATIC_OUTPUT="${TMP_DIR}/static-review.json"
(cd "$TMP_DIR" && "${PROJECT_ROOT}/harness/scripts/run-contract-review-checks.sh" "${TMP_DIR}/static-contract.json" "$STATIC_OUTPUT" >/dev/null)

jq -e '
  .reviewer_profile == "static" and
  .verdict == "SKIPPED"
' "$STATIC_OUTPUT" >/dev/null

echo "test-run-contract-review-checks: ok"
