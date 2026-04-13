package hookhandler

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// elicitationInput is the stdin JSON payload for the Elicitation hook.
type elicitationInput struct {
	MCPServerName  string `json:"mcp_server_name"`
	ServerName     string `json:"server_name"`
	Matcher        string `json:"matcher"`
	ElicitationID  string `json:"elicitation_id"`
	ID             string `json:"id"`
	Message        string `json:"message"`
}

// elicitationLogEntry is the entry written to elicitation-events.jsonl.
type elicitationLogEntry struct {
	Event           string `json:"event"`
	MCPServer       string `json:"mcp_server"`
	ElicitationID   string `json:"elicitation_id"`
	Message         string `json:"message"`
	BreezingSession string `json:"breezing_session"`
	Timestamp       string `json:"timestamp"`
}

// elicitationDecision is the response for the Elicitation hook.
type elicitationDecision struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
}

// ElicitationHandler is the Go port of scripts/hook-handlers/elicitation-handler.sh.
//
// On Elicitation events it logs the MCP elicitation request,
// automatically skips (deny) for Breezing Workers (background, no UI),
// and passes through (allow) for normal sessions.
//
// Logs are recorded in .claude/state/elicitation-events.jsonl.
type ElicitationHandler struct {
	// ProjectRoot is the project root path. Resolved from env vars/CWD when empty.
	ProjectRoot string
}

// Handle processes the Elicitation hook.
func (h *ElicitationHandler) Handle(in io.Reader, out io.Writer) error {
	data, err := io.ReadAll(in)
	if err != nil || len(strings.TrimSpace(string(data))) == 0 {
		return writeJSON(out, elicitationDecision{
			Decision: "approve",
			Reason:   "Elicitation: no payload",
		})
	}

	var input elicitationInput
	if jsonErr := json.Unmarshal(data, &input); jsonErr != nil {
		return writeJSON(out, elicitationDecision{
			Decision: "approve",
			Reason:   "Elicitation: no payload",
		})
	}

	// Normalize fields (equivalent to // fallback in the bash version).
	mcpServer := firstNonEmpty(input.MCPServerName, input.ServerName, input.Matcher)
	elicitationID := firstNonEmpty(input.ElicitationID, input.ID)
	message := input.Message

	// Resolve project root.
	projectRoot := h.ProjectRoot
	if projectRoot == "" {
		projectRoot = resolveProjectRoot()
	}
	stateDir := projectRoot + "/.claude/state"
	logFile := stateDir + "/elicitation-events.jsonl"

	// Record log entry.
	if err := os.MkdirAll(stateDir, 0o700); err == nil {
		ts := time.Now().UTC().Format(time.RFC3339)
		breezingSession := os.Getenv("HARNESS_BREEZING_SESSION_ID")
		entry := elicitationLogEntry{
			Event:           "elicitation",
			MCPServer:       mcpServer,
			ElicitationID:   elicitationID,
			Message:         message,
			BreezingSession: breezingSession,
			Timestamp:       ts,
		}
		if lineData, merr := json.Marshal(entry); merr == nil {
			f, ferr := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
			if ferr == nil {
				fmt.Fprintf(f, "%s\n", lineData)
				f.Close()
				_ = rotateJSONL(logFile, 500, 400)
			}
		}
	}

	// During a Breezing session: auto-skip (background Worker cannot interact with UI).
	breezingSession := os.Getenv("HARNESS_BREEZING_SESSION_ID")
	if breezingSession != "" {
		reason := fmt.Sprintf(
			"Breezing session (%s): background agent cannot interact with elicitation UI",
			breezingSession,
		)
		return writeJSON(out, elicitationDecision{
			Decision: "deny",
			Reason:   reason,
		})
	}

	// Normal session: pass through (user responds interactively).
	return writeJSON(out, elicitationDecision{
		Decision: "approve",
		Reason:   "Elicitation: forwarding to user",
	})
}
