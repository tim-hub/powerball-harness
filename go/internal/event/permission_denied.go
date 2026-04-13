package event

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// PermissionDeniedHandler is the PermissionDenied hook handler (v2.1.89+).
// Fires when the auto mode classifier rejects a command.
// Records the denial event in telemetry and returns a Lead notification with retry hint in Worker mode.
//
// Shell equivalent: scripts/hook-handlers/permission-denied-handler.sh
type PermissionDeniedHandler struct {
	// StateDir is the storage location for log files.
	// If empty, ResolveStateDir(projectRoot) is used.
	StateDir string
}

// permissionDeniedInput is the stdin JSON for the PermissionDenied hook.
type permissionDeniedInput struct {
	Tool         string `json:"tool,omitempty"`
	ToolName     string `json:"tool_name,omitempty"`
	DeniedReason string `json:"denied_reason,omitempty"`
	Reason       string `json:"reason,omitempty"`
	SessionID    string `json:"session_id,omitempty"`
	AgentID      string `json:"agent_id,omitempty"`
	AgentType    string `json:"agent_type,omitempty"`
	CWD          string `json:"cwd,omitempty"`
}

// permissionDeniedLogEntry is an entry written to permission-denied.jsonl.
type permissionDeniedLogEntry struct {
	Event     string `json:"event"`
	Timestamp string `json:"timestamp"`
	SessionID string `json:"session_id"`
	AgentID   string `json:"agent_id"`
	AgentType string `json:"agent_type"`
	Tool      string `json:"tool"`
	Reason    string `json:"reason"`
}

// Handle reads the PermissionDenied payload from stdin, records it in the log,
// and returns retry + systemMessage for Workers; returns approve for non-Workers.
func (h *PermissionDeniedHandler) Handle(r io.Reader, w io.Writer) error {
	data, err := io.ReadAll(r)
	if err != nil || len(data) == 0 {
		return nil
	}

	var inp permissionDeniedInput
	if err := json.Unmarshal(data, &inp); err != nil {
		return nil
	}

	// Normalize fields
	toolName := inp.Tool
	if toolName == "" {
		toolName = inp.ToolName
	}
	if toolName == "" {
		toolName = "unknown"
	}

	deniedReason := inp.DeniedReason
	if deniedReason == "" {
		deniedReason = inp.Reason
	}
	if deniedReason == "" {
		deniedReason = "unknown"
	}

	sessionID := inp.SessionID
	if sessionID == "" {
		sessionID = "unknown"
	}
	agentID := inp.AgentID
	if agentID == "" {
		agentID = "unknown"
	}
	agentType := inp.AgentType
	if agentType == "" {
		agentType = "unknown"
	}

	// Determine state directory
	projectRoot := inp.CWD
	if projectRoot == "" {
		projectRoot = resolveProjectRoot(data)
	}

	stateDir := h.StateDir
	if stateDir == "" {
		stateDir = ResolveStateDir(projectRoot)
	}

	if err := EnsureStateDir(stateDir); err != nil {
		return nil
	}

	logFile := filepath.Join(stateDir, "permission-denied.jsonl")
	if isSymlink(logFile) {
		return nil
	}

	// Record in the log
	entry := permissionDeniedLogEntry{
		Event:     "permission_denied",
		Timestamp: Now(),
		SessionID: sessionID,
		AgentID:   agentID,
		AgentType: agentType,
		Tool:      toolName,
		Reason:    deniedReason,
	}
	h.appendLog(logFile, entry)

	// Debug output to stderr
	fmt.Fprintf(os.Stderr,
		"[PermissionDenied] agent=%s type=%s tool=%s reason=%s\n",
		agentID, agentType, toolName, deniedReason,
	)

	// Worker: return retry + systemMessage
	if h.isWorker(agentType) {
		notificationText := fmt.Sprintf(
			"[PermissionDenied] Worker tool %s was denied by auto mode. "+
				"Reason: %s. Consider an alternative approach or request manual approval.",
			toolName, deniedReason,
		)
		return WriteJSON(w, RetryResponse{
			Retry:         true,
			SystemMessage: notificationText,
		})
	}

	// Non-Worker: return approve
	return WriteJSON(w, ApproveResponse{
		Decision: "approve",
		Reason:   "PermissionDenied logged",
	})
}

// isWorker reports whether agentType is a Worker type.
// Returns true for "worker", "task-worker", or any type ending with ":worker".
func (h *PermissionDeniedHandler) isWorker(agentType string) bool {
	return agentType == "worker" ||
		agentType == "task-worker" ||
		strings.HasSuffix(agentType, ":worker")
}

// appendLog appends one entry to the permission-denied log.
func (h *PermissionDeniedHandler) appendLog(path string, entry permissionDeniedLogEntry) {
	logData, err := json.Marshal(entry)
	if err != nil {
		return
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = fmt.Fprintf(f, "%s\n", logData)

	RotateJSONL(path)
}
