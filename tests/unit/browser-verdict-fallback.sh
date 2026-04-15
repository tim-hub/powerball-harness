#!/bin/bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

cat > "${TMP_DIR}/static-approve.json" <<'EOF'
{
  "verdict": "APPROVE",
  "reviewer_profile": "browser",
  "task": {
    "id": "41.2.2",
    "title": "browser verdict fallback"
  },
  "route": "playwright",
  "browser_mode": "scripted"
}
EOF

cat > "${TMP_DIR}/browser-pending.json" <<'EOF'
{
  "browser_verdict": "PENDING_BROWSER",
  "runner_status": "timeout",
  "note": "browser review timed out"
}
EOF

(cd "$TMP_DIR" && "${PROJECT_ROOT}/scripts/write-review-result.sh" "${TMP_DIR}/static-approve.json" "" "${TMP_DIR}/approve-result.json" --browser-result "${TMP_DIR}/browser-pending.json" >/dev/null)

jq -e '
  .verdict == "APPROVE" and
  .browser_verdict == "PENDING_BROWSER" and
  .reviewer_profile == "browser"
' "${TMP_DIR}/approve-result.json" >/dev/null

cat > "${TMP_DIR}/static-blocking.json" <<'EOF'
{
  "verdict": "REQUEST_CHANGES",
  "reviewer_profile": "browser",
  "task": {
    "id": "41.2.2",
    "title": "browser verdict fallback"
  },
  "route": "playwright",
  "browser_mode": "scripted"
}
EOF

(cd "$TMP_DIR" && "${PROJECT_ROOT}/scripts/write-review-result.sh" "${TMP_DIR}/static-blocking.json" "" "${TMP_DIR}/blocking-result.json" --browser-result "${TMP_DIR}/browser-pending.json" >/dev/null)

jq -e '
  .verdict == "REQUEST_CHANGES" and
  .browser_verdict == "PENDING_BROWSER" and
  .reviewer_profile == "browser"
' "${TMP_DIR}/blocking-result.json" >/dev/null

echo "browser-verdict-fallback: ok"
