package hookhandler

// permission_denied_handler.go
// Go port of permission-denied-handler.sh.
//
// Handles PermissionDenied events (when the auto mode classifier denies a request):
//   - Records the event in .claude/state/permission-denied-events.jsonl
//   - Returns {retry: true, systemMessage: ...} for Workers
//   - Returns approve for non-Workers

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// permissionDeniedInput is the stdin JSON for the PermissionDenied hook.
type permissionDeniedInput struct {
	Tool        string `json:"tool"`
	ToolName    string `json:"tool_name"`
	DeniedReason string `json:"denied_reason"`
	Reason      string `json:"reason"`
	SessionID   string `json:"session_id"`
	AgentID     string `json:"agent_id"`
	AgentType   string `json:"agent_type"`
}

// permissionDeniedLogEntry is a single entry in permission-denied-events.jsonl.
type permissionDeniedLogEntry struct {
	Event     string `json:"event"`
	Timestamp string `json:"timestamp"`
	SessionID string `json:"session_id"`
	AgentID   string `json:"agent_id"`
	AgentType string `json:"agent_type"`
	Tool      string `json:"tool"`
	Reason    string `json:"reason"`
}

// permissionDeniedRetryResponse is the retry response for Workers.
type permissionDeniedRetryResponse struct {
	Retry         bool   `json:"retry"`
	SystemMessage string `json:"systemMessage"`
}

// permissionDeniedApproveResponse is the approve response for non-Workers.
type permissionDeniedApproveResponse struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
}

// HandlePermissionDenied is the Go port of permission-denied-handler.sh.
//
// Called on the PermissionDenied hook:
//  1. Records the event in .claude/state/permission-denied-events.jsonl
//  2. Returns {retry: true, systemMessage: ...} for Workers
//  3. Returns approve for non-Workers
func HandlePermissionDenied(in io.Reader, out io.Writer) error {
	data, err := io.ReadAll(in)
	if err != nil || len(strings.TrimSpace(string(data))) == 0 {
		// no input: exit normally
		return nil
	}

	var input permissionDeniedInput
	if jsonErr := json.Unmarshal(data, &input); jsonErr != nil {
		// pass through even on parse failure
		return writePermissionDeniedApprove(out, "PermissionDenied logged")
	}

	// resolve tool / denied_reason (with fallbacks)
	toolName := input.Tool
	if toolName == "" {
		toolName = input.ToolName
	}
	if toolName == "" {
		toolName = "unknown"
	}

	deniedReason := input.DeniedReason
	if deniedReason == "" {
		deniedReason = input.Reason
	}
	if deniedReason == "" {
		deniedReason = "unknown"
	}

	sessionID := input.SessionID
	if sessionID == "" {
		sessionID = "unknown"
	}
	agentID := input.AgentID
	if agentID == "" {
		agentID = "unknown"
	}
	agentType := input.AgentType
	if agentType == "" {
		agentType = "unknown"
	}

	// ensure state directory exists
	stateDir := resolveNotificationStateDir()
	if mkErr := ensureNotificationStateDir(stateDir); mkErr != nil {
		// pass through even on directory creation failure
		return writePermissionDeniedApprove(out, "PermissionDenied logged")
	}

	// record to JSONL
	logFile := filepath.Join(stateDir, "permission-denied-events.jsonl")
	entry := permissionDeniedLogEntry{
		Event:     "permission_denied",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		SessionID: sessionID,
		AgentID:   agentID,
		AgentType: agentType,
		Tool:      toolName,
		Reason:    deniedReason,
	}
	if logErr := appendPermissionDeniedLog(logFile, entry); logErr != nil {
		_ = logErr
	}

	// debug output to stderr (equivalent to the bash script)
	fmt.Fprintf(os.Stderr,
		"[PermissionDenied] agent=%s type=%s tool=%s reason=%s\n",
		agentID, agentType, toolName, deniedReason,
	)

	// for Workers: return retry + systemMessage
	if isWorkerAgentType(agentType) {
		notificationText := fmt.Sprintf(
			"[PermissionDenied] Worker tool %s was denied by auto mode. Reason: %s. Consider an alternative approach or manually approve if needed.",
			toolName, deniedReason,
		)

		resp := permissionDeniedRetryResponse{
			Retry:         true,
			SystemMessage: notificationText,
		}
		respData, marshalErr := json.Marshal(resp)
		if marshalErr != nil {
			return writePermissionDeniedApprove(out, "PermissionDenied logged")
		}
		_, writeErr := fmt.Fprintf(out, "%s\n", respData)
		return writeErr
	}

	// non-Workers: return approve
	return writePermissionDeniedApprove(out, "PermissionDenied logged")
}

// isWorkerAgentType returns true if agentType is a Worker type.
// Equivalent to the bash: [ "${AGENT_TYPE}" = "worker" ] || [ "${AGENT_TYPE}" = "task-worker" ] || echo "${AGENT_TYPE}" | grep -qE ':worker$'
func isWorkerAgentType(agentType string) bool {
	if agentType == "worker" || agentType == "task-worker" {
		return true
	}
	return strings.HasSuffix(agentType, ":worker")
}

// writePermissionDeniedApprove writes an approve response.
func writePermissionDeniedApprove(out io.Writer, reason string) error {
	resp := permissionDeniedApproveResponse{
		Decision: "approve",
		Reason:   reason,
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("marshal approve response: %w", err)
	}
	_, err = fmt.Fprintf(out, "%s\n", data)
	return err
}

// appendPermissionDeniedLog appends a single entry to the JSONL file and rotates it.
func appendPermissionDeniedLog(logFile string, entry permissionDeniedLogEntry) error {
	// check for symlink
	if isSymlink(logFile) {
		return fmt.Errorf("symlinked log file refused: %s", logFile)
	}

	entryJSON, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal log entry: %w", err)
	}

	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}
	defer f.Close()

	if _, writeErr := fmt.Fprintf(f, "%s\n", entryJSON); writeErr != nil {
		return fmt.Errorf("write log entry: %w", writeErr)
	}

	// rotate: trim to 400 lines when exceeding 500
	return rotateJSONL(logFile, 500, 400)
}
