package event

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// StalenessThreshold is the reset threshold (in seconds) for the consecutive failure counter.
// The counter is reset when StalenessThreshold or more seconds have elapsed since the last failure.
const StalenessThreshold = 60

// PostToolFailureHandler is the PostToolUseFailure hook handler.
// Counts consecutive tool failures and returns an escalation message after 3 consecutive failures.
//
// Shell equivalent: scripts/hook-handlers/post-tool-failure.sh
type PostToolFailureHandler struct {
	// StateDir is the storage location for counter files.
	// If empty, ResolveStateDir(projectRoot) is used.
	StateDir string
	// nowFunc is a time injection function for testing. If nil, time.Now() is used.
	nowFunc func() time.Time
}

// postToolFailureInput is the stdin JSON for the PostToolUseFailure hook.
type postToolFailureInput struct {
	ToolName string `json:"tool_name"`
	// Also accept the toolName alias
	ToolNameAlt string `json:"toolName,omitempty"`
	Error       string `json:"error,omitempty"`
	Message     string `json:"message,omitempty"`
}

// counterRecord is a record in the counter file.
type counterRecord struct {
	Count     int
	Timestamp int64
}

// Handle reads the PostToolUseFailure payload from stdin and writes a
// systemMessage to stdout based on the consecutive failure count.
func (h *PostToolFailureHandler) Handle(r io.Reader, w io.Writer) error {
	data, err := io.ReadAll(r)
	if err != nil || len(data) == 0 {
		return WriteJSON(w, SystemMessageResponse{})
	}

	var inp postToolFailureInput
	if err := json.Unmarshal(data, &inp); err != nil {
		return WriteJSON(w, SystemMessageResponse{})
	}

	toolName := inp.ToolName
	if toolName == "" {
		toolName = inp.ToolNameAlt
	}
	if toolName == "" {
		toolName = "unknown"
	}

	errorMsg := inp.Error
	if errorMsg == "" {
		errorMsg = inp.Message
	}
	if len(errorMsg) > 200 {
		errorMsg = errorMsg[:200]
	}

	// Ensure state directory
	stateDir := h.StateDir
	if stateDir == "" {
		projectRoot := resolveProjectRoot(data)
		stateDir = ResolveStateDir(projectRoot)
	}

	if err := EnsureStateDir(stateDir); err != nil {
		return WriteJSON(w, SystemMessageResponse{})
	}

	counterFile := filepath.Join(stateDir, "tool-failure-counter.txt")
	if isSymlink(counterFile) {
		return WriteJSON(w, SystemMessageResponse{})
	}

	now := h.now().Unix()
	rec := h.readCounter(counterFile, now)
	rec.Count++
	rec.Timestamp = now

	if err := h.writeCounter(counterFile, rec); err != nil {
		return WriteJSON(w, SystemMessageResponse{})
	}

	if rec.Count >= 3 {
		// 3 consecutive failures: escalate
		h.resetCounter(counterFile)
		msg := fmt.Sprintf(
			"WARNING: %d consecutive tool failures detected (tool: %s). "+
				"Stop retrying the same approach. Diagnose the root cause or try an alternative approach. "+
				"Last error: %s",
			rec.Count, toolName, errorMsg,
		)
		return WriteJSON(w, SystemMessageResponse{SystemMessage: msg})
	}

	// Failures 1-2: warning only
	msg := fmt.Sprintf(
		"Tool failure #%d/3 (tool: %s). Will escalate after 3 consecutive failures.",
		rec.Count, toolName,
	)
	return WriteJSON(w, SystemMessageResponse{SystemMessage: msg})
}

// now returns the current time (injectable for testing).
func (h *PostToolFailureHandler) now() time.Time {
	if h.nowFunc != nil {
		return h.nowFunc()
	}
	return time.Now()
}

// readCounter reads the counter file.
// Returns count=0 when the file is missing or has an invalid format.
// Also resets when StalenessThreshold or more seconds have elapsed since the last failure.
func (h *PostToolFailureHandler) readCounter(path string, now int64) counterRecord {
	data, err := os.ReadFile(path)
	if err != nil {
		return counterRecord{}
	}

	parts := strings.Fields(strings.TrimSpace(string(data)))
	if len(parts) < 2 {
		return counterRecord{}
	}

	count, err1 := strconv.Atoi(parts[0])
	ts, err2 := strconv.ParseInt(parts[1], 10, 64)
	if err1 != nil || err2 != nil {
		return counterRecord{}
	}

	// Reset if stale
	if now-ts > StalenessThreshold {
		return counterRecord{}
	}

	return counterRecord{Count: count, Timestamp: ts}
}

// writeCounter writes the counter file.
func (h *PostToolFailureHandler) writeCounter(path string, rec counterRecord) error {
	if isSymlink(path) {
		return fmt.Errorf("security: symlinked counter file: %s", path)
	}
	content := fmt.Sprintf("%d %d\n", rec.Count, rec.Timestamp)
	return os.WriteFile(path, []byte(content), 0600)
}

// resetCounter resets the counter to 0.
func (h *PostToolFailureHandler) resetCounter(path string) {
	_ = h.writeCounter(path, counterRecord{Count: 0, Timestamp: 0})
}

// resolveProjectRoot infers the project root from the input JSON or environment variables.
func resolveProjectRoot(data []byte) string {
	// Try the CWD field
	var v struct {
		CWD string `json:"cwd"`
	}
	if err := json.Unmarshal(data, &v); err == nil && v.CWD != "" {
		return v.CWD
	}

	// Environment variable fallback
	if r := os.Getenv("HARNESS_PROJECT_ROOT"); r != "" {
		return r
	}
	if r := os.Getenv("PROJECT_ROOT"); r != "" {
		return r
	}

	cwd, _ := os.Getwd()
	return cwd
}
