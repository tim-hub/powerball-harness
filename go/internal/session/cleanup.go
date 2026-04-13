package session

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// CleanupHandler is the SessionEnd hook handler.
// Removes temporary files at session end.
//
// Shell equivalent: scripts/session-cleanup.sh
type CleanupHandler struct {
	// StateDir is the state directory path. If empty, it is inferred from cwd.
	StateDir string
}

// cleanupInput is the stdin JSON for the SessionEnd hook.
type cleanupInput struct {
	CWD string `json:"cwd,omitempty"`
}

// cleanupResponse is the cleanup result response.
type cleanupResponse struct {
	Continue bool   `json:"continue"`
	Message  string `json:"message"`
}

// Handle reads the SessionEnd payload from stdin, removes temporary files,
// and writes the result to stdout.
func (h *CleanupHandler) Handle(r io.Reader, w io.Writer) error {
	data, _ := io.ReadAll(r)

	var inp cleanupInput
	if len(data) > 0 {
		_ = json.Unmarshal(data, &inp)
	}

	// Determine state directory
	stateDir := h.StateDir
	if stateDir == "" {
		projectRoot := resolveProjectRoot(inp.CWD)
		stateDir = filepath.Join(projectRoot, ".claude", "state")
	}

	// Early return if state directory does not exist
	if _, err := os.Stat(stateDir); err != nil {
		return writeJSON(w, cleanupResponse{Continue: true, Message: "No state directory"})
	}

	// Symlink check (security)
	if isSymlink(stateDir) {
		return writeJSON(w, cleanupResponse{Continue: true, Message: "State directory is symlink, skipping"})
	}

	// Delete fixed temporary files
	tempFiles := []string{
		"pending-skill.json",
		"current-operation.json",
	}
	for _, name := range tempFiles {
		path := filepath.Join(stateDir, name)
		if isRegularFile(path) {
			_ = os.Remove(path)
		}
	}

	// Clean up inbox-*.tmp files
	h.cleanupGlob(stateDir, "inbox-*.tmp")

	return writeJSON(w, cleanupResponse{Continue: true, Message: "Session cleanup completed"})
}

// cleanupGlob removes files matching the glob pattern in the state directory.
func (h *CleanupHandler) cleanupGlob(stateDir, pattern string) {
	matches, err := filepath.Glob(filepath.Join(stateDir, pattern))
	if err != nil {
		return
	}
	for _, path := range matches {
		if isRegularFile(path) {
			_ = os.Remove(path)
		}
	}
}

// isRegularFile reports whether path is a regular file (excluding symbolic links).
func isRegularFile(path string) bool {
	fi, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return fi.Mode().IsRegular()
}

// cleanupFilenameGlobMatch reports whether name glob-matches pattern.
// Uses filepath.Match but only supports patterns without path separators.
func cleanupFilenameGlobMatch(pattern, name string) bool {
	matched, err := filepath.Match(pattern, name)
	if err != nil {
		return false
	}
	return matched
}

// buildCleanupSummary builds a log-ready list of files to be cleaned up (for debugging).
func buildCleanupSummary(files []string) string {
	if len(files) == 0 {
		return "none"
	}
	return strings.Join(files, ", ")
}

// formatCleanupResult returns a JSON cleanup result string (for error display).
func formatCleanupResult(deleted int, err error) string {
	if err != nil {
		return fmt.Sprintf(`{"continue":true,"message":"cleanup partial: %d files removed, error: %v"}`, deleted, err)
	}
	return fmt.Sprintf(`{"continue":true,"message":"Session cleanup completed: %d files removed"}`, deleted)
}
