#!/bin/bash
# test-breezing-advisor-protocol.sh
# breezing / harness-work 系スキルの advisor protocol と mirror 同期を固定する

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "${SCRIPT_DIR}")"

fail() {
  echo "FAIL: $1" >&2
  exit 1
}

assert_contains() {
  local file="$1"
  local pattern="$2"
  local label="$3"
  grep -q "$pattern" "$file" || fail "${label}: ${pattern}"
}

SHARED_WORK="${PROJECT_ROOT}/skills/harness-work/SKILL.md"
SHARED_BREEZING="${PROJECT_ROOT}/skills/breezing/SKILL.md"
CODEX_WORK="${PROJECT_ROOT}/skills-codex/harness-work/SKILL.md"
CODEX_BREEZING="${PROJECT_ROOT}/skills-codex/breezing/SKILL.md"
CODEX_MIRROR_WORK="${PROJECT_ROOT}/codex/.codex/skills/harness-work/SKILL.md"
CODEX_MIRROR_BREEZING="${PROJECT_ROOT}/codex/.codex/skills/breezing/SKILL.md"
OPENCODE_MIRROR_WORK="${PROJECT_ROOT}/opencode/skills/harness-work/SKILL.md"

for file in \
  "${SHARED_WORK}" "${SHARED_BREEZING}" \
  "${CODEX_WORK}" "${CODEX_BREEZING}" \
  "${CODEX_MIRROR_WORK}" "${CODEX_MIRROR_BREEZING}" "${OPENCODE_MIRROR_WORK}"
do
  [ -f "${file}" ] || fail "missing file: ${file}"
done

assert_contains "${SHARED_WORK}" 'advisor-request.v1' "shared harness-work"
assert_contains "${SHARED_WORK}" 'advisor-response.v1' "shared harness-work"
assert_contains "${SHARED_WORK}" 'task ごとの相談回数は最大 3 回' "shared harness-work"
assert_contains "${SHARED_BREEZING}" 'Worker → `advisor-request.v1`' "shared breezing"
assert_contains "${SHARED_BREEZING}" 'task ごとの相談回数は最大 3 回' "shared breezing"

assert_contains "${CODEX_WORK}" 'advisor-request.v1' "codex harness-work"
assert_contains "${CODEX_WORK}" 'Advisor Protocol' "codex harness-work"
assert_contains "${CODEX_BREEZING}" 'advisor-response.v1' "codex breezing"
assert_contains "${CODEX_BREEZING}" 'PIVOT_REQUIRED' "codex breezing"

diff -q "${SHARED_WORK}" "${OPENCODE_MIRROR_WORK}" >/dev/null \
  || fail "opencode mirror の harness-work が SSOT と不一致です"
diff -q "${CODEX_WORK}" "${CODEX_MIRROR_WORK}" >/dev/null \
  || fail "codex mirror の harness-work が SSOT と不一致です"
diff -q "${CODEX_BREEZING}" "${CODEX_MIRROR_BREEZING}" >/dev/null \
  || fail "codex mirror の breezing が SSOT と不一致です"

echo "test-breezing-advisor-protocol: ok"
