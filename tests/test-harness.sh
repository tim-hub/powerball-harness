#\!/usr/bin/env bash
# test-harness.sh — run all test-*.sh scripts in harness/tests/
# These tests cover harness-internal scripts (harness/scripts/)

set -euo pipefail
export TMPDIR=/tmp  # Force /tmp for sandboxed execution (sandbox blocks /var/folders)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
HARNESS_TESTS_DIR="$(cd "$SCRIPT_DIR/../harness/tests" && pwd)"

passed=0
failed=0
failures=""

for script in "$HARNESS_TESTS_DIR"/test-*.sh; do
  printf "  %-55s" "$(basename "$script")"
  if bash "$script" >/dev/null 2>&1; then
    echo "PASS"
    passed=$((passed + 1))
  else
    echo "FAIL"
    failed=$((failed + 1))
    failures="$failures\n    $script"
  fi
done

echo ""
echo "Results: $passed passed, $failed failed"

if [ "$failed" -gt 0 ]; then
  printf "Failed scripts:%s\n" "$failures"
  exit 1
fi
