package hookhandler

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// memoryBridgeEvent represents a dispatched memory bridge event written to the
// event log. Actual MCP calls are deferred to future implementation.
type memoryBridgeEvent struct {
	Event     string `json:"event"`
	SessionID string `json:"session_id,omitempty"`
	CWD       string `json:"cwd,omitempty"`
	Timestamp string `json:"timestamp"`
}

// memoryBridgeInput is the minimal JSON payload expected on stdin.
type memoryBridgeInput struct {
	SessionID     string `json:"session_id"`
	CWD           string `json:"cwd"`
	HookEventName string `json:"hook_event_name"`
}

// validTargets lists the recognised dispatch targets.
var validTargets = map[string]bool{
	"session-start":  true,
	"user-prompt":    true,
	"post-tool-use":  true,
	"stop":           true,
	"codex-notify":   true,
}

// HandleMemoryBridge ports scripts/hook-handlers/memory-bridge.sh.
//
// Dispatches one of the four known event targets (session-start, user-prompt,
// post-tool-use, stop). The real MCP call (harness-mem-bridge) is deferred;
// for now only an event log entry is written.
//
// Usage: the target is read from the JSON field "hook_event_name" which is
// populated by the hooks dispatcher.  Unknown targets exit 0 (fail-open).
func HandleMemoryBridge(in io.Reader, out io.Writer) error {
	data, err := io.ReadAll(in)
	if err != nil {
		return approveMemoryBridge(out, "")
	}

	var input memoryBridgeInput
	if jsonErr := json.Unmarshal(data, &input); jsonErr != nil {
		return approveMemoryBridge(out, "")
	}

	target := input.HookEventName
	if !validTargets[target] {
		// Unknown target: log to stderr (matching bash `echo ... >&2; exit 0`)
		fmt.Fprintf(os.Stderr, "[claude-code-harness] unknown memory bridge target: %s\n", target)
		return approveMemoryBridge(out, target)
	}

	// Log the event (MCP call to be implemented in the future).
	if logErr := logMemoryBridgeEvent(target, input.SessionID, input.CWD); logErr != nil {
		// Non-fatal: the hook must not block on log failures.
		fmt.Fprintf(os.Stderr, "[claude-code-harness] memory-bridge log error: %v\n", logErr)
	}

	return approveMemoryBridge(out, target)
}

// approveMemoryBridge writes the standard approve response.
func approveMemoryBridge(out io.Writer, target string) error {
	reason := "memory-bridge: ok"
	if target != "" && validTargets[target] {
		reason = fmt.Sprintf("memory-bridge: %s dispatched", target)
	}
	resp := map[string]string{"decision": "approve", "reason": reason}
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "%s\n", data)
	return err
}

// logMemoryBridgeEvent appends an event entry to the memory bridge event log.
// The log file is created at .claude/state/memory-bridge-events.jsonl.
func logMemoryBridgeEvent(event, sessionID, cwd string) error {
	projectRoot := resolveProjectRoot()
	stateDir := projectRoot + "/.claude/state"
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		return err
	}

	entry := memoryBridgeEvent{
		Event:     event,
		SessionID: sessionID,
		CWD:       cwd,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	logPath := stateDir + "/memory-bridge-events.jsonl"
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = fmt.Fprintf(f, "%s\n", data)
	return err
}
