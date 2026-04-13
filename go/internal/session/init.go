// Package session implements session lifecycle handlers for Claude Code Harness.
//
// Each handler corresponds to a shell script that was previously used:
//   - Init      → scripts/session-init.sh
//   - Cleanup   → scripts/session-cleanup.sh
//   - Monitor   → scripts/session-monitor.sh
//   - Summary   → scripts/session-summary.sh
//
// Handlers read hook JSON from stdin and write the appropriate response to stdout.
package session

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ---------------------------------------------------------------------------
// InitHandler
// ---------------------------------------------------------------------------

// InitHandler is the SessionStart hook handler.
// Ports the main functionality of session-init.sh to Go:
//  1. Lightweight initialization for subagents
//  2. Session JSON initialization (session.json)
//  3. Plans.md task counting
//  4. JSON response including additionalContext
//
// shell version: scripts/session-init.sh
type InitHandler struct {
	// StateDir is the path to the .claude/state directory. Inferred from cwd if empty.
	StateDir string
	// PlansFile is the path to Plans.md. Defaults to projectRoot/Plans.md if empty.
	PlansFile string
}

// initInput is the stdin JSON for the SessionStart hook.
type initInput struct {
	SessionID string `json:"session_id,omitempty"`
	AgentType string `json:"agent_type,omitempty"`
	CWD       string `json:"cwd,omitempty"`
}

// sessionJSON is the minimal schema for session.json.
type sessionJSON struct {
	SessionID  string `json:"session_id"`
	State      string `json:"state"`
	StartedAt  string `json:"started_at"`
	UpdatedAt  string `json:"updated_at"`
	EventSeq   int    `json:"event_seq"`
	LastEventID string `json:"last_event_id"`
}

// initResponse is the JSON output for the SessionStart hook.
type initResponse struct {
	HookSpecificOutput initHookOutput `json:"hookSpecificOutput"`
}

type initHookOutput struct {
	HookEventName     string `json:"hookEventName"`
	AdditionalContext string `json:"additionalContext"`
}

// Handle reads the SessionStart payload from stdin, initializes the session,
// and writes JSON including additionalContext to stdout.
func (h *InitHandler) Handle(r io.Reader, w io.Writer) error {
	data, _ := io.ReadAll(r)

	var inp initInput
	if len(data) > 0 {
		_ = json.Unmarshal(data, &inp)
	}

	// Lightweight initialization for subagents (skip session.json operations)
	if inp.AgentType == "subagent" {
		return writeJSON(w, initResponse{
			HookSpecificOutput: initHookOutput{
				HookEventName:     "SessionStart",
				AdditionalContext: "[subagent] lightweight initialization complete",
			},
		})
	}

	// Determine project root and state directory
	projectRoot := resolveProjectRoot(inp.CWD)
	stateDir := h.StateDir
	if stateDir == "" {
		stateDir = filepath.Join(projectRoot, ".claude", "state")
	}

	// Create state directory (with symlink check)
	if err := ensureStateDir(stateDir); err != nil {
		// Continue even on error (still output banner and Plans info)
		_ = err
	}

	// Initialize session.json (when it does not exist or is in a stopped state)
	_ = h.initSessionFile(stateDir)

	// Reset session-skills-used.json
	skillsUsedFile := filepath.Join(stateDir, "session-skills-used.json")
	now := time.Now().UTC().Format(time.RFC3339)
	_ = writeFileAtomic(skillsUsedFile, []byte(fmt.Sprintf(`{"used":[],"session_start":%q}`, now)+"\n"), 0600)

	// Clear SSOT sync flag
	_ = os.Remove(filepath.Join(stateDir, ".ssot-synced-this-session"))
	// Clear work review warning flags
	_ = os.Remove(filepath.Join(stateDir, ".work-review-warned"))
	_ = os.Remove(filepath.Join(stateDir, ".ultrawork-review-warned"))

	// Count Plans.md tasks
	plansFile := h.PlansFile
	if plansFile == "" {
		plansFile = filepath.Join(projectRoot, "Plans.md")
	}
	plansInfo := buildPlansInfo(plansFile)

	// Append marker legend
	context := buildAdditionalContext(plansInfo)

	return writeJSON(w, initResponse{
		HookSpecificOutput: initHookOutput{
			HookEventName:     "SessionStart",
			AdditionalContext: context,
		},
	})
}

// initSessionFile initializes session.json.
// Does nothing if the existing file is in an active state (initialized/running/working).
func (h *InitHandler) initSessionFile(stateDir string) error {
	sessionFile := filepath.Join(stateDir, "session.json")

	if isSymlink(sessionFile) {
		return fmt.Errorf("security: symlinked session file: %s", sessionFile)
	}

	// Check the state of the existing file
	if data, err := os.ReadFile(sessionFile); err == nil {
		var s sessionJSON
		if json.Unmarshal(data, &s) == nil {
			// Keep as-is for states other than stopped/completed/failed
			switch s.State {
			case "stopped", "completed", "failed":
				// New initialization required
			default:
				return nil
			}
		}
	}

	// New session initialization
	now := time.Now().UTC().Format(time.RFC3339)
	sessionID := fmt.Sprintf("session-%d", time.Now().Unix())
	s := sessionJSON{
		SessionID:  sessionID,
		State:      "initialized",
		StartedAt:  now,
		UpdatedAt:  now,
		EventSeq:   0,
		LastEventID: "",
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return writeFileAtomic(sessionFile, append(data, '\n'), 0600)
}

// buildPlansInfo reads Plans.md and returns an info string with WIP/TODO counts.
func buildPlansInfo(plansFile string) string {
	if _, err := os.Stat(plansFile); err != nil {
		return "Plans.md: not found"
	}

	wipCount := countMatches(plansFile, "cc:WIP", "pm:pending", "cursor:pending")
	todoCount := countMatches(plansFile, "cc:TODO")

	return fmt.Sprintf("Plans.md: in-progress %d / todo %d", wipCount, todoCount)
}

// buildAdditionalContext builds the additionalContext for session initialization.
func buildAdditionalContext(plansInfo string) string {
	var sb strings.Builder
	sb.WriteString("# [claude-code-harness] Session Initialization\n\n")
	sb.WriteString(plansInfo + "\n")
	sb.WriteString("\n## Marker Legend\n")
	sb.WriteString("| Marker | Status | Description |\n")
	sb.WriteString("|---------|------|------|\n")
	sb.WriteString("| `cc:TODO` | Not started | Scheduled for execution by Impl (Claude Code) |\n")
	sb.WriteString("| `cc:WIP` | In progress | Being implemented by Impl |\n")
	sb.WriteString("| `cc:blocked` | Blocked | Waiting for dependency task |\n")
	sb.WriteString("| `pm:pending` | Requested by PM | Used in 2-Agent setup |\n")
	sb.WriteString("\n> **Compatibility**: `cursor:pending` / `cursor:confirmed` are treated as synonyms for `pm:*`.\n")
	return sb.String()
}

// ---------------------------------------------------------------------------
// Utilities (package-private)
// ---------------------------------------------------------------------------

// writeJSON writes v as JSON to w.
func writeJSON(w io.Writer, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}
	_, err = fmt.Fprintf(w, "%s\n", data)
	return err
}

// resolveProjectRoot infers the project root from the CWD field or environment variables.
func resolveProjectRoot(cwd string) string {
	if cwd != "" {
		return cwd
	}
	if r := os.Getenv("HARNESS_PROJECT_ROOT"); r != "" {
		return r
	}
	if r := os.Getenv("PROJECT_ROOT"); r != "" {
		return r
	}
	root, _ := os.Getwd()
	return root
}

// ensureStateDir creates the state directory.
// Returns an error if the path is a symbolic link.
func ensureStateDir(stateDir string) error {
	parent := filepath.Dir(stateDir)
	if isSymlink(parent) || isSymlink(stateDir) {
		return fmt.Errorf("security: symlinked state path refused: %s", stateDir)
	}
	return os.MkdirAll(stateDir, 0700)
}

// isSymlink returns whether the path is a symbolic link.
func isSymlink(path string) bool {
	fi, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeSymlink != 0
}

// countMatches returns the total number of lines containing any of the patterns.
func countMatches(filePath string, patterns ...string) int {
	f, err := os.Open(filePath)
	if err != nil {
		return 0
	}
	defer f.Close()

	count := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		for _, p := range patterns {
			if strings.Contains(line, p) {
				count++
				break
			}
		}
	}
	return count
}

// writeFileAtomic atomically writes a file via a temporary file.
func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	if isSymlink(path) {
		return fmt.Errorf("security: symlinked file refused: %s", path)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, perm); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
