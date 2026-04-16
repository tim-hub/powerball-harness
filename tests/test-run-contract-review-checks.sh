#!/bin/bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
TMP_DIR="$(mktemp -d)"
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
(cd "$TMP_DIR" && "${PROJECT_ROOT}/scripts/run-contract-review-checks.sh" "${TMP_DIR}/runtime-contract.json" "$OUTPUT" >/dev/null)

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
(cd "$TMP_DIR" && "${PROJECT_ROOT}/scripts/run-contract-review-checks.sh" "${TMP_DIR}/static-contract.json" "$STATIC_OUTPUT" >/dev/null)

jq -e '
  .reviewer_profile == "static" and
  .verdict == "SKIPPED"
' "$STATIC_OUTPUT" >/dev/null

cat > "${TMP_DIR}/browser-contract.json" <<'EOF'
{
  "task": {
    "id": "32.2.2",
    "title": "browser profile"
  },
  "review": {
    "reviewer_profile": "browser"
  },
  "contract": {
    "browser_validation": [
      { "label": "browser-check", "command": "echo ok" }
    ]
  }
}
EOF

cat > "${TMP_DIR}/browser-approve.sh" <<'EOF'
#!/bin/sh
cat <<'JSON'
{"browser_verdict":"APPROVE","note":"browser ok"}
JSON
EOF
chmod +x "${TMP_DIR}/browser-approve.sh"

BROWSER_APPROVE_OUTPUT="${TMP_DIR}/browser-review-approve.json"
(cd "$TMP_DIR" && HARNESS_BROWSER_REVIEW_COMMAND="${TMP_DIR}/browser-approve.sh" "${PROJECT_ROOT}/scripts/run-contract-review-checks.sh" "${TMP_DIR}/browser-contract.json" "$BROWSER_APPROVE_OUTPUT" >/dev/null)

jq -e '
  .reviewer_profile == "browser" and
  .verdict == "APPROVE" and
  .browser_verdict == "APPROVE" and
  (.checks | length) == 1 and
  (.browser_artifact_path | endswith(".browser-review.json")) and
  (.browser_result_path | endswith(".browser-result.json"))
' "$BROWSER_APPROVE_OUTPUT" >/dev/null

cat > "${TMP_DIR}/browser-request-changes.sh" <<'EOF'
#!/bin/sh
cat <<'JSON'
{"browser_verdict":"REQUEST_CHANGES","note":"needs fixes"}
JSON
EOF
chmod +x "${TMP_DIR}/browser-request-changes.sh"

BROWSER_REQUEST_OUTPUT="${TMP_DIR}/browser-review-request.json"
(cd "$TMP_DIR" && HARNESS_BROWSER_REVIEW_COMMAND="${TMP_DIR}/browser-request-changes.sh" "${PROJECT_ROOT}/scripts/run-contract-review-checks.sh" "${TMP_DIR}/browser-contract.json" "$BROWSER_REQUEST_OUTPUT" >/dev/null)

jq -e '
  .reviewer_profile == "browser" and
  .verdict == "REQUEST_CHANGES" and
  .browser_verdict == "REQUEST_CHANGES"
' "$BROWSER_REQUEST_OUTPUT" >/dev/null

cat > "${TMP_DIR}/browser-timeout.sh" <<'EOF'
#!/bin/sh
sleep 2
EOF
chmod +x "${TMP_DIR}/browser-timeout.sh"

BROWSER_TIMEOUT_OUTPUT="${TMP_DIR}/browser-review-timeout.json"
(cd "$TMP_DIR" && HARNESS_BROWSER_REVIEW_TIMEOUT_SECONDS=1 HARNESS_BROWSER_REVIEW_COMMAND="${TMP_DIR}/browser-timeout.sh" "${PROJECT_ROOT}/scripts/run-contract-review-checks.sh" "${TMP_DIR}/browser-contract.json" "$BROWSER_TIMEOUT_OUTPUT" >/dev/null)

jq -e '
  .reviewer_profile == "browser" and
  .verdict == "PENDING_BROWSER" and
  .browser_verdict == "PENDING_BROWSER"
' "$BROWSER_TIMEOUT_OUTPUT" >/dev/null

TIMEOUT_BROWSER_RESULT="$(jq -r '.browser_result_path' "$BROWSER_TIMEOUT_OUTPUT")"
jq -e '
  .runner_status == "timeout" and
  .browser_verdict == "PENDING_BROWSER"
' "$TIMEOUT_BROWSER_RESULT" >/dev/null

echo "test-run-contract-review-checks: ok"
