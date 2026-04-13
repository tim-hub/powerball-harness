package hookhandler

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// ClearPendingHandler is a PostToolUse hook handler (clear pending-skills).
// Deletes .claude/state/pending-skills/*.pending files after a Skill tool is executed.
// The Skill invocation is treated as evidence that the quality gate has been satisfied,
// resolving the pending state.
//
// shell version: scripts/posttooluse-clear-pending.sh
type ClearPendingHandler struct {
	// ProjectRoot is the path to the project root. Uses cwd if empty.
	ProjectRoot string
}

// clearPendingResponse is the response from the ClearPending hook.
type clearPendingResponse struct {
	Continue bool `json:"continue"`
}

// Handle reads and discards the payload from stdin (not used),
// then deletes all *.pending files in the pending-skills directory.
func (h *ClearPendingHandler) Handle(r io.Reader, w io.Writer) error {
	// discard stdin (this handler does not use the input)
	_, _ = io.ReadAll(r)

	projectRoot := h.ProjectRoot
	if projectRoot == "" {
		projectRoot, _ = os.Getwd()
	}

	pendingDir := filepath.Join(projectRoot, ".claude", "state", "pending-skills")

	// skip if the pending directory does not exist
	if _, err := os.Stat(pendingDir); os.IsNotExist(err) {
		return writePendingJSON(w, clearPendingResponse{Continue: true})
	}

	// delete all *.pending files
	matches, err := filepath.Glob(filepath.Join(pendingDir, "*.pending"))
	if err == nil {
		for _, path := range matches {
			_ = os.Remove(path)
		}
	}

	return writePendingJSON(w, clearPendingResponse{Continue: true})
}

// writePendingJSON writes v as JSON to w.
func writePendingJSON(w io.Writer, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}
	_, err = fmt.Fprintf(w, "%s\n", data)
	return err
}
