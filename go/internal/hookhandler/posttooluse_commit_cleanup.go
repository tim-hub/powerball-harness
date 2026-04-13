package hookhandler

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// CommitCleanupHandler is a PostToolUse hook handler (cleanup after git commit).
// Deletes the review approval state file after a successful git commit command.
//
// shell version: scripts/posttooluse-commit-cleanup.sh
type CommitCleanupHandler struct {
	// ProjectRoot is the path to the project root. Uses cwd if empty.
	ProjectRoot string
}

// commitCleanupInput is the stdin JSON for the PostToolUse hook.
type commitCleanupInput struct {
	ToolName   string                 `json:"tool_name,omitempty"`
	ToolInput  map[string]interface{} `json:"tool_input,omitempty"`
	ToolResult interface{}            `json:"tool_result,omitempty"`
}

// Handle reads the PostToolUse payload from stdin and deletes the review
// approval state file if a git commit command succeeded.
// This handler writes only log messages to stdout (no JSON needed).
func (h *CommitCleanupHandler) Handle(r io.Reader, w io.Writer) error {
	data, _ := io.ReadAll(r)

	if len(data) == 0 {
		return nil
	}

	var inp commitCleanupInput
	if err := json.Unmarshal(data, &inp); err != nil {
		return nil
	}

	// skip tools other than Bash
	if inp.ToolName != "Bash" {
		return nil
	}

	// get the command
	command := ""
	if v, ok := inp.ToolInput["command"]; ok {
		if s, ok := v.(string); ok {
			command = s
		}
	}
	if command == "" {
		return nil
	}

	// check whether this is a git commit command (case-insensitive)
	if !isGitCommitCommand(command) {
		return nil
	}

	// convert tool result to string
	toolResult := ""
	switch v := inp.ToolResult.(type) {
	case string:
		toolResult = v
	case map[string]interface{}:
		if b, err := json.Marshal(v); err == nil {
			toolResult = string(b)
		}
	}

	// skip if the result contains error indicators
	if containsErrorIndicator(toolResult) {
		return nil
	}

	// delete the review approval state files
	projectRoot := h.ProjectRoot
	if projectRoot == "" {
		projectRoot, _ = os.Getwd()
	}

	reviewStateFile := projectRoot + "/.claude/state/review-approved.json"
	reviewResultFile := projectRoot + "/.claude/state/review-result.json"

	stateFileExists := fileExists(reviewStateFile)
	resultFileExists := fileExists(reviewResultFile)

	if stateFileExists || resultFileExists {
		_ = os.Remove(reviewStateFile)
		_ = os.Remove(reviewResultFile)

		_, _ = fmt.Fprintf(w, "[Commit Guard] Cleared review approval state. Please run an independent review again before your next commit.\n")
	}

	return nil
}

// isGitCommitCommand returns true if the command string contains a git commit invocation.
// Equivalent to bash grep -Eiq: '(^|[[:space:]])git[[:space:]]+commit([[:space:]]|$)'
func isGitCommitCommand(command string) bool {
	lower := strings.ToLower(command)
	// sequentially search for the "git commit" pattern
	searchFrom := 0
	for searchFrom < len(lower) {
		idx := strings.Index(lower[searchFrom:], "git")
		if idx < 0 {
			break
		}
		absIdx := searchFrom + idx

		// "git" must be preceded by start-of-string or whitespace
		if absIdx > 0 && !isWordBoundaryBefore(lower[absIdx-1]) {
			searchFrom = absIdx + 1
			continue
		}

		// "git" must be followed by whitespace
		afterGit := absIdx + 3
		if afterGit >= len(lower) || !isWordBoundaryBefore(lower[afterGit]) {
			searchFrom = absIdx + 1
			continue
		}

		// skip whitespace and look for "commit"
		i := afterGit
		for i < len(lower) && isWordBoundaryBefore(lower[i]) {
			i++
		}
		if strings.HasPrefix(lower[i:], "commit") {
			after := i + 6
			if after >= len(lower) || isWordBoundaryBefore(lower[after]) {
				return true
			}
		}
		searchFrom = absIdx + 1
	}
	return false
}

// isWordBoundaryBefore returns true if c is a whitespace character (word boundary).
func isWordBoundaryBefore(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

// containsErrorIndicator returns true if the tool result contains error indicators.
func containsErrorIndicator(result string) bool {
	lower := strings.ToLower(result)
	for _, indicator := range []string{"error", "fatal", "failed", "nothing to commit"} {
		if strings.Contains(lower, indicator) {
			return true
		}
	}
	return false
}
