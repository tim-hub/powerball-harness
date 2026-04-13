package hookhandler

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// stopSessionInput is the stdin JSON payload for the Stop hook.
// CC 2.1.47+ includes last_assistant_message.
type stopSessionInput struct {
	StopHookActive       bool   `json:"stop_hook_active"`
	TranscriptPath       string `json:"transcript_path"`
	LastAssistantMessage string `json:"last_assistant_message"`
}

// stopSessionResponse is the response for the Stop hook.
type stopSessionResponse struct {
	OK            bool   `json:"ok"`
	Reason        string `json:"reason,omitempty"`
	SystemMessage string `json:"systemMessage,omitempty"`
}

// StopSessionEvaluatorHandler is the Go port of scripts/hook-handlers/stop-session-evaluator.sh.
//
// Evaluates session state on the Stop event:
//   - Records last_assistant_message as length + hash (first 16 chars of SHA-256) in session.json
//   - Emits a systemMessage warning when WIP tasks remain (does not block)
//   - Always permits stop (ok: true)
type StopSessionEvaluatorHandler struct {
	// ProjectRoot is the project root path. Resolved from env vars/CWD when empty.
	ProjectRoot string
}

// Handle processes the Stop hook.
func (h *StopSessionEvaluatorHandler) Handle(in io.Reader, out io.Writer) error {
	// Resolve project root.
	projectRoot := h.ProjectRoot
	if projectRoot == "" {
		projectRoot = resolveProjectRoot()
	}
	stateFile := projectRoot + "/.claude/state/session.json"

	// Read stdin (size limit: 64 KiB).
	var payload []byte
	limited := io.LimitReader(in, 65536)
	payload, _ = io.ReadAll(limited)

	// Record last_assistant_message metadata in session.json.
	if len(payload) > 0 {
		var input stopSessionInput
		if jsonErr := json.Unmarshal(payload, &input); jsonErr == nil {
			if input.LastAssistantMessage != "" {
				h.recordLastMessage(stateFile, input.LastAssistantMessage)
			}
		}
	}

	// Default to ok when session.json does not exist.
	if _, err := os.Stat(stateFile); os.IsNotExist(err) {
		return writeJSON(out, stopSessionResponse{OK: true})
	}

	// Read session state.
	sessionData, err := os.ReadFile(stateFile)
	if err != nil {
		return writeJSON(out, stopSessionResponse{OK: true})
	}

	var sessionMap map[string]interface{}
	if jsonErr := json.Unmarshal(sessionData, &sessionMap); jsonErr != nil {
		return writeJSON(out, stopSessionResponse{OK: true})
	}

	// Immediately ok if already in stopped state.
	if state, ok := sessionMap["state"].(string); ok && state == "stopped" {
		return writeJSON(out, stopSessionResponse{OK: true})
	}

	// WIP task check: find Plans.md and count cc:WIP markers.
	wipCount := h.countWIPTasks(projectRoot)
	if wipCount > 0 {
		msg := fmt.Sprintf(
			"[StopSession] %d WIP task(s) remain. Please check Plans.md.",
			wipCount,
		)
		return writeJSON(out, stopSessionResponse{
			OK:            true,
			SystemMessage: msg,
		})
	}

	return writeJSON(out, stopSessionResponse{OK: true})
}

// recordLastMessage records last_message_length and last_message_hash in session.json.
// The plaintext content is not saved (privacy protection).
func (h *StopSessionEvaluatorHandler) recordLastMessage(stateFile, msg string) {
	// Skip if the file does not exist (matches bash counterpart behavior).
	sessionData, err := os.ReadFile(stateFile)
	if err != nil {
		return
	}

	var sessionMap map[string]interface{}
	if jsonErr := json.Unmarshal(sessionData, &sessionMap); jsonErr != nil {
		return
	}

	msgLen := len(msg)
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(msg)))[:16]

	sessionMap["last_message_length"] = msgLen
	sessionMap["last_message_hash"] = hash

	newData, err := json.Marshal(sessionMap)
	if err != nil {
		return
	}

	// Atomic write: temp file + rename.
	stateDir := stateFile[:strings.LastIndex(stateFile, "/")]
	tmpFile, err := os.CreateTemp(stateDir, "session.json.*")
	if err != nil {
		return
	}
	tmpPath := tmpFile.Name()
	defer func() {
		// Cleanup if rename fails.
		os.Remove(tmpPath)
	}()

	if _, err := tmpFile.Write(append(newData, '\n')); err != nil {
		tmpFile.Close()
		return
	}
	tmpFile.Close()

	_ = os.Rename(tmpPath, stateFile)
}

// countWIPTasks finds Plans.md under projectRoot and returns the count of cc:WIP markers.
func (h *StopSessionEvaluatorHandler) countWIPTasks(projectRoot string) int {
	for _, name := range plansFileNames {
		path := projectRoot + "/" + name
		f, err := os.Open(path)
		if err != nil {
			continue
		}
		count := 0
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			if strings.Contains(scanner.Text(), "cc:WIP") {
				count++
			}
		}
		f.Close()
		return count
	}
	return 0
}
