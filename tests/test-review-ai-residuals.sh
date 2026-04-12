#!/bin/bash
# Minimal regression test for review-ai-residuals.sh

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SCRIPT_PATH="${ROOT_DIR}/scripts/review-ai-residuals.sh"
FIXTURE_DIR="${ROOT_DIR}/tests/fixtures/review-ai-residuals"

command -v jq >/dev/null 2>&1 || {
  echo "jq is required for tests/test-review-ai-residuals.sh"
  exit 1
}

TMP_DIR="$(mktemp -d)"
cleanup() {
  rm -rf "${TMP_DIR}"
}
trap cleanup EXIT

mkdir -p "${TMP_DIR}/src"
cp "${FIXTURE_DIR}/clean.ts" "${TMP_DIR}/src/clean.ts"
cp "${FIXTURE_DIR}/major.ts" "${TMP_DIR}/src/major.ts"
cp "${FIXTURE_DIR}/minor.ts" "${TMP_DIR}/src/minor.ts"
cp "${FIXTURE_DIR}/recommendation.ts" "${TMP_DIR}/src/recommendation.ts"
cp "${FIXTURE_DIR}/README.md" "${TMP_DIR}/README.md"

file_output="$(
  cd "${TMP_DIR}" && \
  bash "${SCRIPT_PATH}" src/clean.ts src/major.ts src/minor.ts src/recommendation.ts README.md
)"

echo "${file_output}" | jq -e '
  .summary.verdict == "REQUEST_CHANGES" and
  .summary.major >= 3 and
  .summary.minor >= 4 and
  .summary.recommendation >= 1 and
  .summary.total == (.observations | length) and
  (.files_scanned | length) == 4 and
  ([.observations[].rule] | index("hardcoded-secret")) != null and
  ([.observations[].rule] | index("localhost-reference")) != null and
  ([.observations[].rule] | index("test-skip")) != null and
  ([.observations[].rule] | index("dummy-value")) != null and
  ([.observations[].rule] | index("todo-fixme")) != null and
  (([.observations[].match] | join(" ")) | contains("localhost:3000")) and
  (([.observations[].match] | join(" ")) | contains("<redacted>"))
' >/dev/null || {
  echo "explicit file scan did not return the expected JSON summary"
  exit 1
}

cd "${TMP_DIR}"
git init -q
git config user.name "Harness Test"
git config user.email "harness-test@example.com"
cp "${FIXTURE_DIR}/clean.ts" "${TMP_DIR}/src/major.ts"
cp "${FIXTURE_DIR}/clean.ts" "${TMP_DIR}/src/minor.ts"
git add src/clean.ts src/major.ts src/minor.ts
git commit -qm "chore: baseline"

cp "${FIXTURE_DIR}/major.ts" "${TMP_DIR}/src/major.ts"
cp "${FIXTURE_DIR}/minor.ts" "${TMP_DIR}/src/minor.ts"

diff_output="$(bash "${SCRIPT_PATH}" --base-ref HEAD)"

echo "${diff_output}" | jq -e '
  .scan_mode == "diff" and
  .base_ref == "HEAD" and
  .summary.verdict == "REQUEST_CHANGES" and
  .summary.major >= 3 and
  .summary.minor >= 1 and
  (.files_scanned | sort) == ["src/major.ts", "src/minor.ts"]
' >/dev/null || {
  echo "diff scan did not return the expected files or severities"
  exit 1
}

echo "OK"
