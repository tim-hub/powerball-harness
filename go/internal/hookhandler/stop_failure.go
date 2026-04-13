package hookhandler

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// stopFailureInput is the stdin JSON payload for the StopFailure hook.
type stopFailureInput struct {
	Error     stopFailureError `json:"error"`
	SessionID string           `json:"session_id"`
}

// stopFailureError handles both struct and string variants of the error field.
type stopFailureError struct {
	Message string `json:"message"`
	Status  string `json:"status"`
	Code    string `json:"code"`
	Raw     string // set when error is a plain string
}

// UnmarshalJSON supports both string and object variants of the error field.
func (e *stopFailureError) UnmarshalJSON(data []byte) error {
	// Try as string first
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		e.Raw = s
		return nil
	}
	// Then try as object
	type plain struct {
		Message string `json:"message"`
		Status  string `json:"status"`
		Code    string `json:"code"`
	}
	var p plain
	if err := json.Unmarshal(data, &p); err != nil {
		return err
	}
	e.Message = p.Message
	e.Status = p.Status
	e.Code = p.Code
	return nil
}

// stopFailureLogEntry is the entry recorded in stop-failures.jsonl.
type stopFailureLogEntry struct {
	Event     string `json:"event"`
	Timestamp string `json:"timestamp"`
	SessionID string `json:"session_id"`
	ErrorCode string `json:"error_code"`
	Message   string `json:"message"`
}

// stopFailureSystemMessage is the systemMessage response for 429 rate-limit events.
type stopFailureSystemMessage struct {
	SystemMessage string `json:"systemMessage"`
}

// StopFailureHandler is the Go port of scripts/hook-handlers/stop-failure.sh.
//
// Handles StopFailure events (when session stop fails due to an API error):
//   - Records error information in .claude/state/stop-failures.jsonl
//   - Classifies error type (rate_limit, auth_error, network_error, unknown)
//   - Notifies the Lead via systemMessage on 429 rate limit
type StopFailureHandler struct {
	// ProjectRoot is the project root path. Resolved from env vars/CWD when empty.
	ProjectRoot string
}

// Handle processes the StopFailure hook.
func (h *StopFailureHandler) Handle(in io.Reader, out io.Writer) error {
	data, err := io.ReadAll(in)
	if err != nil || len(strings.TrimSpace(string(data))) == 0 {
		// No payload: nothing to log, no output.
		return nil
	}

	// Resolve project root.
	projectRoot := h.ProjectRoot
	if projectRoot == "" {
		projectRoot = resolveProjectRoot()
	}
	stateDir := resolveStopFailureStateDir(projectRoot)

	// Ensure state directory exists.
	if mkErr := os.MkdirAll(stateDir, 0o700); mkErr != nil {
		// Failed to create directory: print to stderr and exit.
		fmt.Fprintf(os.Stderr, "[StopFailure] mkdir %s: %v\n", stateDir, mkErr)
		return nil
	}

	logFile := stateDir + "/stop-failures.jsonl"

	// Symlink check (security).
	if isStopFailureLogSymlink(logFile) {
		fmt.Fprintf(os.Stderr, "[StopFailure] symlink detected at %s, aborting\n", logFile)
		return nil
	}

	// Parse JSON.
	var input stopFailureInput
	_ = json.Unmarshal(data, &input)

	// Normalize error information.
	errorMsg, errorCode := normalizeStopFailureError(input.Error)
	sessionID := input.SessionID
	if sessionID == "" {
		sessionID = "unknown"
	}

	ts := time.Now().UTC().Format(time.RFC3339)

	// Record JSONL log entry.
	entry := stopFailureLogEntry{
		Event:     "stop_failure",
		Timestamp: ts,
		SessionID: sessionID,
		ErrorCode: errorCode,
		Message:   errorMsg,
	}
	if lineData, merr := json.Marshal(entry); merr == nil {
		f, ferr := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if ferr == nil {
			fmt.Fprintf(f, "%s\n", lineData)
			f.Close()
			_ = rotateJSONL(logFile, 500, 400)
		}
	}

	// 429 rate limit: notify Lead via systemMessage.
	if errorCode == "429" || errorCode == "rate_limit" {
		msg := fmt.Sprintf(
			"[StopFailure] Worker %s stopped due to rate limit (429). Breezing Lead should attempt auto-resume after exponential backoff.",
			sessionID,
		)
		if err := writeJSON(out, stopFailureSystemMessage{SystemMessage: msg}); err != nil {
			fmt.Fprintf(os.Stderr, "[StopFailure] write systemMessage: %v\n", err)
		}
	}

	// Debug output to stderr.
	fmt.Fprintf(os.Stderr, "[StopFailure] session=%s code=%s msg=%s\n", sessionID, errorCode, errorMsg)

	return nil
}

// normalizeStopFailureError normalizes error information and returns the message and code.
// Error classification:
//   - "429" / message contains "rate" → rate_limit
//   - "401", "403" / message contains "auth" → auth_error
//   - message contains "network", "connection", "timeout" → network_error
//   - otherwise → unknown
func normalizeStopFailureError(e stopFailureError) (msg, code string) {
	// Prefer Raw (string error field).
	if e.Raw != "" {
		msg = e.Raw
		code = classifyErrorCode("", msg)
		return
	}

	msg = e.Message
	if msg == "" {
		msg = "unknown"
	}

	rawCode := firstNonEmpty(e.Status, e.Code)
	if rawCode == "" {
		rawCode = "unknown"
	}

	code = classifyErrorCode(rawCode, msg)
	return
}

// classifyErrorCode classifies the error type from the raw code and message.
func classifyErrorCode(rawCode, msg string) string {
	// Classify by HTTP status code.
	switch rawCode {
	case "429":
		return "429"
	case "401", "403":
		return "auth_error"
	}

	// Classify by message (also applies when code is "unknown").
	lower := strings.ToLower(msg)
	if strings.Contains(lower, "rate") || strings.Contains(lower, "429") {
		return "rate_limit"
	}
	if strings.Contains(lower, "auth") || strings.Contains(lower, "unauthorized") || strings.Contains(lower, "forbidden") {
		return "auth_error"
	}
	if strings.Contains(lower, "network") || strings.Contains(lower, "connection") || strings.Contains(lower, "timeout") {
		return "network_error"
	}

	if rawCode != "" && rawCode != "unknown" {
		return rawCode
	}
	return "unknown"
}

// isStopFailureLogSymlink reports whether the log file is a symlink.
// Used for security checks. isSymlink is defined in userprompt_track_command.go.
func isStopFailureLogSymlink(path string) bool {
	return isSymlink(path)
}

// resolveStopFailureStateDir returns the directory where stop-failures.jsonl is stored.
// Matches the behavior of bash stop-failure.sh:
//   - When CLAUDE_PLUGIN_DATA is set: ${CLAUDE_PLUGIN_DATA}/projects/<hash>
//     where <hash> is the first 12 characters of the CWD's SHA-256
//   - Otherwise: ${projectRoot}/.claude/state
func resolveStopFailureStateDir(projectRoot string) string {
	pluginData := os.Getenv("CLAUDE_PLUGIN_DATA")
	if pluginData == "" {
		return projectRoot + "/.claude/state"
	}

	// Use the first 12 characters of the CWD's SHA-256 as the project hash.
	hash := sha256.Sum256([]byte(projectRoot))
	hashStr := fmt.Sprintf("%x", hash)
	if len(hashStr) > 12 {
		hashStr = hashStr[:12]
	}
	if hashStr == "" {
		hashStr = "default"
	}
	return pluginData + "/projects/" + hashStr
}
