#!/usr/bin/env bash
set -euo pipefail

# run-benchmarks.sh
# Runs vanilla and breezing experiments, then converts and analyzes results
#
# Usage:
#   ./scripts/run-benchmarks.sh                    # Run both experiments
#   ./scripts/run-benchmarks.sh --mode vanilla      # Vanilla only
#   ./scripts/run-benchmarks.sh --mode breezing     # Breezing only
#   ./scripts/run-benchmarks.sh --dry               # Dry run (preview)
#   ./scripts/run-benchmarks.sh --task task-01      # Single task

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
AGENT_EVAL_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
RESULTS_DIR="$AGENT_EVAL_DIR/results"
CONVERTED_DIR="$RESULTS_DIR/converted"

MODE="both"
DRY_RUN=""
TASK_FILTER=""

while [[ $# -gt 0 ]]; do
    case "$1" in
        --mode)
            MODE="$2"
            shift 2
            ;;
        --dry)
            DRY_RUN="--dry"
            shift
            ;;
        --task)
            TASK_FILTER="--evals $2"
            shift 2
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: $0 [--mode vanilla|breezing|both] [--dry] [--task task-XX]"
            exit 1
            ;;
    esac
done

cd "$AGENT_EVAL_DIR"

# Install dependencies if needed
if [[ ! -d "node_modules" ]]; then
    echo "Installing dependencies..."
    npm install
fi

TIMESTAMP=$(date -u +"%Y-%m-%dT%H-%M-%SZ")

run_experiment() {
    local experiment="$1"
    echo ""
    echo "========================================"
    echo "Running $experiment experiment..."
    echo "========================================"
    echo ""

    # shellcheck disable=SC2086
    npx @vercel/agent-eval "$experiment" $DRY_RUN $TASK_FILTER || {
        echo "WARNING: $experiment experiment had failures (continuing...)"
    }
}

# Run experiments
if [[ "$MODE" == "both" || "$MODE" == "vanilla" ]]; then
    run_experiment "vanilla"
fi

if [[ "$MODE" == "both" || "$MODE" == "breezing" ]]; then
    run_experiment "breezing"
fi

if [[ -n "$DRY_RUN" ]]; then
    echo ""
    echo "Dry run complete. No results to analyze."
    exit 0
fi

# Find latest result directories
echo ""
echo "========================================"
echo "Converting results..."
echo "========================================"

VANILLA_LATEST=""
BREEZING_LATEST=""

if [[ -d "$RESULTS_DIR/vanilla" ]]; then
    VANILLA_LATEST=$(ls -1d "$RESULTS_DIR/vanilla/"* 2>/dev/null | sort | tail -1)
fi

if [[ -d "$RESULTS_DIR/breezing" ]]; then
    BREEZING_LATEST=$(ls -1d "$RESULTS_DIR/breezing/"* 2>/dev/null | sort | tail -1)
fi

CONVERT_ARGS="--output-dir $CONVERTED_DIR/$TIMESTAMP"

if [[ -n "$VANILLA_LATEST" ]]; then
    CONVERT_ARGS="$CONVERT_ARGS --vanilla-dir $VANILLA_LATEST"
fi

if [[ -n "$BREEZING_LATEST" ]]; then
    CONVERT_ARGS="$CONVERT_ARGS --breezing-dir $BREEZING_LATEST"
fi

# shellcheck disable=SC2086
python3 "$AGENT_EVAL_DIR/analyzer/convert-results.py" $CONVERT_ARGS

# Run analysis
echo ""
echo "========================================"
echo "Generating analysis report..."
echo "========================================"

REPORT_FILE="$RESULTS_DIR/report-$TIMESTAMP.md"
python3 "$AGENT_EVAL_DIR/analyzer/analyzer.py" \
    --results-dir "$CONVERTED_DIR/$TIMESTAMP" \
    --output "$REPORT_FILE"

echo ""
echo "========================================"
echo "Done!"
echo "========================================"
echo "Results:  $CONVERTED_DIR/$TIMESTAMP/"
echo "Report:   $REPORT_FILE"
