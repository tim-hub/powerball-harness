#!/bin/bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

cat > "${TMP_DIR}/playwright-artifact.json" <<'EOF'
{
  "schema_version": "browser-review.v1",
  "task": {
    "id": "41.2.2",
    "title": "playwright default route"
  },
  "reviewer_profile": "browser",
  "browser_mode": "scripted",
  "route": "playwright",
  "tool_matcher": "mcp__playwright__*",
  "required_artifacts": ["trace", "screenshot", "ui-flow-log"],
  "execution_instructions": ["Use Playwright MCP for browser review."]
}
EOF

mkdir -p "${TMP_DIR}/playwright"
cat > "${TMP_DIR}/playwright/package.json" <<'EOF'
{
  "scripts": {
    "test:e2e": "echo playwright default"
  }
}
EOF

PLAYWRIGHT_BIN="${TMP_DIR}/playwright-bin"
mkdir -p "$PLAYWRIGHT_BIN"
cat > "${PLAYWRIGHT_BIN}/npm" <<'EOF'
#!/bin/sh
if [ "$1" = "run" ] && [ "$2" = "test:e2e" ]; then
  cat <<'JSON'
{"browser_verdict":"APPROVE","note":"playwright default path"}
JSON
  exit 0
fi
echo "unexpected npm invocation: $*" >&2
exit 1
EOF
chmod +x "${PLAYWRIGHT_BIN}/npm"

(cd "${TMP_DIR}/playwright" && PATH="${PLAYWRIGHT_BIN}:${PATH}" "${PROJECT_ROOT}/scripts/browser-review-runner.sh" "${TMP_DIR}/playwright-artifact.json" "${TMP_DIR}/playwright-result.json" >/dev/null)

jq -e '
  .browser_verdict == "APPROVE" and
  .verdict == "APPROVE" and
  .runner_status == "ok" and
  .note == "browser review command completed"
' "${TMP_DIR}/playwright-result.json" >/dev/null

cat > "${TMP_DIR}/agent-browser-artifact.json" <<'EOF'
{
  "schema_version": "browser-review.v1",
  "task": {
    "id": "41.2.2",
    "title": "agent-browser default route"
  },
  "reviewer_profile": "browser",
  "browser_mode": "exploratory",
  "route": "agent-browser",
  "tool_matcher": "agent-browser|bash agent-browser",
  "required_artifacts": ["snapshot", "ui-flow-log"],
  "execution_instructions": ["Use agent-browser CLI for browser review."]
}
EOF

AGENT_BROWSER_BIN="${TMP_DIR}/agent-browser-bin"
mkdir -p "$AGENT_BROWSER_BIN"
cat > "${AGENT_BROWSER_BIN}/agent-browser" <<'EOF'
#!/bin/sh
cat <<'JSON'
{"browser_verdict":"REQUEST_CHANGES","note":"agent-browser default path"}
JSON
EOF
chmod +x "${AGENT_BROWSER_BIN}/agent-browser"

(cd "$TMP_DIR" && PATH="${AGENT_BROWSER_BIN}:${PATH}" "${PROJECT_ROOT}/scripts/browser-review-runner.sh" "${TMP_DIR}/agent-browser-artifact.json" "${TMP_DIR}/agent-browser-result.json" >/dev/null)

jq -e '
  .browser_verdict == "REQUEST_CHANGES" and
  .verdict == "REQUEST_CHANGES" and
  .runner_status == "ok" and
  .note == "browser review command completed"
' "${TMP_DIR}/agent-browser-result.json" >/dev/null

cat > "${TMP_DIR}/chrome-devtools-artifact.json" <<'EOF'
{
  "schema_version": "browser-review.v1",
  "task": {
    "id": "41.2.2",
    "title": "chrome devtools pending route"
  },
  "reviewer_profile": "browser",
  "browser_mode": "scripted",
  "route": "chrome-devtools",
  "tool_matcher": "mcp__chrome-devtools__*",
  "required_artifacts": ["screenshot", "ui-flow-log"],
  "execution_instructions": ["Use Chrome DevTools MCP to capture screenshot and UI flow log."]
}
EOF

(cd "$TMP_DIR" && "${PROJECT_ROOT}/scripts/browser-review-runner.sh" "${TMP_DIR}/chrome-devtools-artifact.json" "${TMP_DIR}/chrome-devtools-result.json" >/dev/null)

jq -e '
  .browser_verdict == "PENDING_BROWSER" and
  .verdict == "PENDING_BROWSER" and
  .runner_status == "unavailable" and
  (.note | contains("chrome-devtools has no shell-executable default"))
' "${TMP_DIR}/chrome-devtools-result.json" >/dev/null

echo "browser-review-runner-defaults: ok"
