#!/bin/bash
# harness-mem wrapper scripts should resolve the sibling repo without hardcoded Desktop paths.

set -euo pipefail
export TMPDIR=/tmp  # Force /tmp for sandboxed execution (sandbox blocks /var/folders)

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
HARNESS_DIR="${ROOT_DIR}/harness"
TMP_DIR="$(mktemp -d "/tmp/harness-test.XXXXXX")"
trap 'rm -rf "${TMP_DIR}"' EXIT

mkdir -p "${TMP_DIR}/claude-code-harness/scripts/lib"
mkdir -p "${TMP_DIR}/claude-code-harness/scripts/hook-handlers"
mkdir -p "${TMP_DIR}/harness-mem/scripts/hook-handlers"

cp "${HARNESS_DIR}/scripts/lib/harness-mem-bridge.sh" "${TMP_DIR}/claude-code-harness/scripts/lib/harness-mem-bridge.sh"
cp "${HARNESS_DIR}/scripts/hook-handlers/memory-session-start.sh" "${TMP_DIR}/claude-code-harness/scripts/hook-handlers/memory-session-start.sh"

cat > "${TMP_DIR}/harness-mem/scripts/hook-handlers/memory-session-start.sh" <<'EOF'
#!/bin/bash
set -euo pipefail
export TMPDIR=/tmp  # Force /tmp for sandboxed execution (sandbox blocks /var/folders)
printf 'memory-session-start-ok\n'
EOF

chmod +x \
  "${TMP_DIR}/claude-code-harness/scripts/hook-handlers/memory-session-start.sh" \
  "${TMP_DIR}/harness-mem/scripts/hook-handlers/memory-session-start.sh"

wrapper_output="$(cd "${TMP_DIR}/claude-code-harness" && ./scripts/hook-handlers/memory-session-start.sh)"

[ "${wrapper_output}" = "memory-session-start-ok" ] || {
  echo "memory-session-start wrapper did not resolve sibling harness-mem repo"
  exit 1
}

echo "OK"
