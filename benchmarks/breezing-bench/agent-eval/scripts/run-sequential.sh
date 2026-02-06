#!/usr/bin/env bash
set -euo pipefail

# run-sequential.sh
# Runs experiments sequentially (1 task at a time) to avoid Docker resource exhaustion
#
# Usage:
#   ./scripts/run-sequential.sh vanilla     # Run vanilla for all tasks
#   ./scripts/run-sequential.sh breezing    # Run breezing for all tasks
#   ./scripts/run-sequential.sh both        # Run both (default)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
AGENT_EVAL_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
EXPERIMENTS_DIR="$AGENT_EVAL_DIR/experiments"

MODE="${1:-both}"

cd "$AGENT_EVAL_DIR"

run_experiment_sequential() {
    local experiment="$1"
    local base_config="$EXPERIMENTS_DIR/${experiment}.ts"

    if [[ ! -f "$base_config" ]]; then
        echo "ERROR: Config not found: $base_config"
        return 1
    fi

    echo ""
    echo "========================================"
    echo "Running $experiment experiment (sequential)"
    echo "========================================"

    for task_num in $(seq -w 1 10); do
        local task_id="task-${task_num}"
        local temp_config="$EXPERIMENTS_DIR/_seq_${experiment}_${task_id}.ts"

        echo ""
        echo "--- $experiment: $task_id ---"

        # Create temporary single-task config by modifying evals array
        sed "s/evals: \[/evals: [\"${task_id}\"],\/\/ SEQUENTIAL: /;s/\"task-[0-9]*\",\?//g" \
            "$base_config" > "$temp_config"

        # Actually, sed is fragile for this. Use node instead.
        node -e "
import { readFileSync, writeFileSync } from 'fs';
const src = readFileSync('$base_config', 'utf-8');
const modified = src.replace(
  /evals:\s*\[[\s\S]*?\]/,
  'evals: [\"${task_id}\"]'
);
writeFileSync('$temp_config', modified);
"

        npx @vercel/agent-eval "_seq_${experiment}_${task_id}" || {
            echo "WARNING: $experiment $task_id had failures (continuing...)"
        }

        rm -f "$temp_config"
    done
}

if [[ "$MODE" == "both" || "$MODE" == "vanilla" ]]; then
    run_experiment_sequential "vanilla"
fi

if [[ "$MODE" == "both" || "$MODE" == "breezing" ]]; then
    run_experiment_sequential "breezing"
fi

echo ""
echo "========================================"
echo "Sequential execution complete!"
echo "Results in: $AGENT_EVAL_DIR/results/"
echo "========================================"
