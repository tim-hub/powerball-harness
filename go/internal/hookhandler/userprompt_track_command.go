package hookhandler

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// TrackCommandHandler is a UserPromptSubmit hook handler (slash command tracking).
// Detects /slash commands from the user prompt and records usage counts.
// Also creates pending-skills marker files for required commands.
//
// shell version: scripts/userprompt-track-command.sh
type TrackCommandHandler struct {
	// ProjectRoot is the path to the project root. Uses cwd if empty.
	ProjectRoot string
}

// trackCommandInput is the input for the UserPromptSubmit hook.
type trackCommandInput struct {
	Prompt string `json:"prompt"`
}

// trackCommandResponse is the response from the TrackCommand hook.
type trackCommandResponse struct {
	Continue bool `json:"continue"`
}

// pendingEntry is the content of a pending file.
type pendingEntry struct {
	Command       string `json:"command"`
	StartedAt     string `json:"started_at"`
	PromptPreview string `json:"prompt_preview"`
}

// skillRequiredCommands is the list of commands that require a pending marker to be created.
var skillRequiredCommands = map[string]bool{
	"work":            true,
	"harness-review":  true,
	"validate":        true,
	"plan-with-agent": true,
}

// slashCommandRe is the regular expression to detect a /slash-command at the beginning of a line.
var slashCommandRe = regexp.MustCompile(`^/([a-zA-Z0-9_:/-]+)`)

// Handle reads the payload from stdin and detects and records slash commands.
func (h *TrackCommandHandler) Handle(r io.Reader, w io.Writer) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return writeTrackJSON(w, trackCommandResponse{Continue: true})
	}

	if len(data) == 0 {
		return writeTrackJSON(w, trackCommandResponse{Continue: true})
	}

	var input trackCommandInput
	if err := json.Unmarshal(data, &input); err != nil {
		return writeTrackJSON(w, trackCommandResponse{Continue: true})
	}

	if input.Prompt == "" {
		return writeTrackJSON(w, trackCommandResponse{Continue: true})
	}

	// check the first line only
	firstLine := strings.SplitN(input.Prompt, "\n", 2)[0]
	matches := slashCommandRe.FindStringSubmatch(firstLine)
	if matches == nil {
		return writeTrackJSON(w, trackCommandResponse{Continue: true})
	}

	rawCommand := matches[1]

	// normalize command name (strip plugin prefix)
	// claude-code-harness:xxx:yyy → yyy (last segment)
	commandName := rawCommand
	if strings.HasPrefix(commandName, "claude-code-harness:") || strings.HasPrefix(commandName, "claude-code-harness/") {
		// extract the last segment (after : or /)
		parts := regexp.MustCompile(`[:/]`).Split(commandName, -1)
		commandName = parts[len(parts)-1]
	}

	if commandName == "" {
		return writeTrackJSON(w, trackCommandResponse{Continue: true})
	}

	projectRoot := h.ProjectRoot
	if projectRoot == "" {
		projectRoot, _ = os.Getwd()
	}

	stateDir := filepath.Join(projectRoot, ".claude", "state")
	pendingDir := filepath.Join(stateDir, "pending-skills")

	// check if this is a skill-required command
	if skillRequiredCommands[commandName] {
		if err := h.createPendingMarker(pendingDir, commandName, input.Prompt); err != nil {
			// ignore pending file creation failure and continue the hook
			_, _ = fmt.Fprintf(os.Stderr, "[track-command] Warning: failed to create pending marker: %v\n", err)
		}
	}

	return writeTrackJSON(w, trackCommandResponse{Continue: true})
}

// createPendingMarker creates a pending marker file.
// Validates each path to prevent path traversal via symbolic links.
func (h *TrackCommandHandler) createPendingMarker(pendingDir, commandName, prompt string) error {
	// check for symlinks (pendingDir and its parent)
	parentDir := filepath.Dir(pendingDir)
	if isSymlink(parentDir) || isSymlink(pendingDir) {
		return fmt.Errorf("symlink detected in state path, skipping")
	}

	// create directory (owner-only permissions)
	if err := os.MkdirAll(pendingDir, 0700); err != nil {
		return fmt.Errorf("mkdir pending dir: %w", err)
	}

	pendingFile := filepath.Join(pendingDir, commandName+".pending")

	// verify the pending file itself is not a symlink
	if isSymlink(pendingFile) {
		return fmt.Errorf("symlink detected at %s, skipping", pendingFile)
	}

	// prompt preview (up to 200 rune characters, newlines converted to spaces)
	preview := strings.ReplaceAll(prompt, "\n", " ")
	runes := []rune(preview)
	if len(runes) > 200 {
		preview = string(runes[:200])
	}

	entry := pendingEntry{
		Command:       commandName,
		StartedAt:     time.Now().UTC().Format(time.RFC3339),
		PromptPreview: preview,
	}

	entryData, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling pending entry: %w", err)
	}

	// write with owner-only permissions
	return os.WriteFile(pendingFile, entryData, 0600)
}

// writeTrackJSON writes v as JSON to w.
func writeTrackJSON(w io.Writer, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}
	_, err = fmt.Fprintf(w, "%s\n", data)
	return err
}
