#!/bin/bash
# test-guardrails-r01-r13.sh
# Guardrail rule table R01-R13 の Go テストをまとめて実行するスモーク。

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
GO_DIR="$PROJECT_ROOT/go"

if [ ! -d "$GO_DIR" ]; then
  echo "go ディレクトリが見つかりません: $GO_DIR" >&2
  exit 1
fi

echo "Running guardrail R01-R13 tests..."
(
  cd "$GO_DIR"
  go test ./internal/guardrail -run '^TestR(0[1-9]|1[0-3])_'
)

echo "Guardrail R01-R13 tests passed."
