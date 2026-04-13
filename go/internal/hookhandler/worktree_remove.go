package hookhandler

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
)

// WorktreeRemoveHandler is the WorktreeRemove hook handler.
// Cleans up worktree-specific temporary files when a Breezing sub-agent exits.
//
// shell version: scripts/hook-handlers/worktree-remove.sh
type WorktreeRemoveHandler struct{}

// worktreeRemoveInput is the stdin JSON for the WorktreeRemove hook.
type worktreeRemoveInput struct {
	SessionID string `json:"session_id,omitempty"`
	CWD       string `json:"cwd,omitempty"`
}

// worktreeRemoveResponse is the response from the WorktreeRemove hook.
type worktreeRemoveResponse struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
}

// Handle reads the WorktreeRemove payload from stdin,
// deletes worktree-specific temporary files, and writes the result to stdout.
func (h *WorktreeRemoveHandler) Handle(r io.Reader, w io.Writer) error {
	data, _ := io.ReadAll(r)

	// skip if payload is empty
	if len(data) == 0 || string(data) == "\n" || string(data) == "\r\n" {
		return writeJSON(w, worktreeRemoveResponse{
			Decision: "approve",
			Reason:   "WorktreeRemove: no payload",
		})
	}

	var inp worktreeRemoveInput
	if err := json.Unmarshal(data, &inp); err != nil {
		return writeJSON(w, worktreeRemoveResponse{
			Decision: "approve",
			Reason:   "WorktreeRemove: no payload",
		})
	}

	if inp.SessionID == "" {
		return writeJSON(w, worktreeRemoveResponse{
			Decision: "approve",
			Reason:   "WorktreeRemove: no session_id",
		})
	}

	// delete Codex prompt temporary files
	removeTmpGlob("/tmp/codex-prompt-*.md")

	// delete Harness Codex logs
	removeTmpGlob("/tmp/harness-codex-*.log")

	// clean up worktree-info.json
	if inp.CWD != "" {
		infoFile := filepath.Join(inp.CWD, ".claude", "state", "worktree-info.json")
		_ = os.Remove(infoFile)
	}

	return writeJSON(w, worktreeRemoveResponse{
		Decision: "approve",
		Reason:   "WorktreeRemove: cleaned up worktree resources",
	})
}

// removeTmpGlob deletes files under /tmp that match the given glob pattern.
func removeTmpGlob(pattern string) {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return
	}
	for _, path := range matches {
		_ = os.Remove(path)
	}
}
