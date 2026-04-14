#!/bin/bash
# SubagentStop hook sample script
# This script evaluates the output of a subagent and logs it.

# Read input from stdin (JSON containing subagent information)
INPUT=$(cat)

# Extract relevant fields
AGENT_ID=$(echo "$INPUT" | jq -r '.agent_id // "unknown"')
TRANSCRIPT_PATH=$(echo "$INPUT" | jq -r '.transcript_path // "unknown"')
TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')

# Log directory
LOG_DIR="${CLAUDE_PROJECT_DIR}/.claude/subagent_logs"
mkdir -p "$LOG_DIR"

# Log file
LOG_FILE="$LOG_DIR/subagent_completions.log"

# Log the completion
echo "=== Subagent Completed ===" >> "$LOG_FILE"
echo "Timestamp: $TIMESTAMP" >> "$LOG_FILE"
echo "Agent ID: $AGENT_ID" >> "$LOG_FILE"
echo "Transcript: $TRANSCRIPT_PATH" >> "$LOG_FILE"
echo "" >> "$LOG_FILE"

# Optional: Extract the last message from the transcript
if [ -f "$TRANSCRIPT_PATH" ]; then
    LAST_MESSAGE=$(tail -n 1 "$TRANSCRIPT_PATH" | jq -r '.content // "N/A"')
    echo "Last Message: $LAST_MESSAGE" >> "$LOG_FILE"
    echo "" >> "$LOG_FILE"
fi

# Exit with success
exit 0
