#!/usr/bin/env bash
set -euo pipefail

# generate-benchmark-fixtures.sh
# Converts tasks/task-XX/ into agent-benchmark fixture directories
#
# Usage: ./scripts/generate-evals.sh [task_number]
#   task_number: 1-10 (omit for all tasks)
#
# SECURITY NOTE: execFileSync calls in generated EVAL.ts run only inside
# agent-eval Docker sandbox with hardcoded arguments (no user input).

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BENCH_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
TASKS_DIR="$BENCH_DIR/tasks"
OUTPUT_DIR="$SCRIPT_DIR/../evals"

generate_fixture() {
    local task_num="$1"
    local task_id
    task_id=$(printf "task-%02d" "$task_num")
    local task_dir="$TASKS_DIR/$task_id"
    local fixture_dir="$OUTPUT_DIR/$task_id"

    if [[ ! -d "$task_dir" ]]; then
        echo "SKIP: $task_dir does not exist"
        return
    fi

    echo "Generating fixture for $task_id..."
    mkdir -p "$fixture_dir/src"

    # 1. PROMPT.md from task.yaml
    python3 -c "
import yaml
with open('$task_dir/task.yaml', 'r') as f:
    data = yaml.safe_load(f)
print(data.get('prompt', ''))
" > "$fixture_dir/PROMPT.md"

    # 2. Copy base configs
    cp "$TASKS_DIR/base-package.json" "$fixture_dir/package.json"
    cp "$TASKS_DIR/base-tsconfig.json" "$fixture_dir/tsconfig.json"

    # 3. Copy setup/src/ files
    if [[ -d "$task_dir/setup/src" ]]; then
        cp -r "$task_dir/setup/src/"* "$fixture_dir/src/"
    fi

    # 4. Generate FIXTURE.ts from hidden tests
    local hidden_test
    hidden_test=$(find "$task_dir/hidden-tests/" -name "*.test.ts" | head -1)
    if [[ -z "$hidden_test" ]]; then
        echo "  WARNING: No hidden tests found for $task_id"
        return
    fi

    local test_filename
    test_filename=$(basename "$hidden_test")

    # Read and escape hidden test content for template literal
    python3 << PYGEN > "$fixture_dir/EVAL.ts"
import sys

with open("$hidden_test", "r") as f:
    content = f.read()

# Escape backticks and dollar signs for JS template literal
escaped = content.replace("\\\\", "\\\\\\\\").replace("\`", "\\\\\`").replace("\$", "\\\\\$")

print('''import { test, expect } from "vitest";
import { execFileSync } from "child_process";
import { writeFileSync, mkdirSync, existsSync } from "fs";
import { join } from "path";

const HIDDEN_TEST = \`
''' + escaped + '''
\`;

test("hidden tests pass", () => {
  const hiddenDir = join(process.cwd(), "src", "__hidden_tests__");
  if (!existsSync(hiddenDir)) {
    mkdirSync(hiddenDir, { recursive: true });
  }
  writeFileSync(join(hiddenDir, "$test_filename"), HIDDEN_TEST);

  if (!existsSync(join(process.cwd(), "node_modules"))) {
    execFileSync("npm", ["install"], { stdio: "pipe" });
  }

  const stdout = execFileSync(
    "npx", ["vitest", "run", "--reporter=json", "src/__hidden_tests__/"],
    { encoding: "utf-8", stdio: "pipe" }
  );

  const report = JSON.parse(stdout);
  expect(report.numPassedTests).toBe(report.numTotalTests);
  expect(report.numTotalTests).toBeGreaterThanOrEqual(5);
});

test("typecheck passes", () => {
  try {
    execFileSync("npx", ["tsc", "--noEmit"], {
      encoding: "utf-8",
      stdio: "pipe",
    });
  } catch (e: any) {
    const output = (e.stdout || "") + (e.stderr || "");
    const tsErrors = output.match(/error TS/g) || [];
    expect(tsErrors.length).toBe(0);
  }
});''')
PYGEN

    echo "  Done: $fixture_dir"
}

# Main
if [[ "${1:-}" =~ ^[0-9]+$ ]]; then
    generate_fixture "$1"
else
    for i in $(seq 1 10); do
        generate_fixture "$i"
    done
fi

echo ""
echo "All fixtures generated successfully."
