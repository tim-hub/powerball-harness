#!/bin/bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

"$SCRIPT_DIR/run-work-all-success.sh" --smoke
"$SCRIPT_DIR/run-work-all-failure.sh" --smoke

echo "All smoke checks passed."
