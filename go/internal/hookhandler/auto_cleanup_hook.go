package hookhandler

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// AutoCleanupHandler is the PostToolUse hook handler (automatic size check).
// It checks the line count of files written by the Write/Edit tool and emits
// a systemMessage warning when Plans.md / CLAUDE.md exceed the threshold.
type AutoCleanupHandler struct {
	// ProjectRoot is the project root path. Falls back to cwd when empty.
	ProjectRoot string

	// Thresholds (use default values when 0).
	PlansMaxLines    int
	ClaudeMdMaxLines int
}

const (
	defaultPlansMaxLines    = 200
	defaultClaudeMdMaxLines = 100
)

// autoCleanupInput is the stdin JSON for the PostToolUse hook.
type autoCleanupInput struct {
	ToolInput    autoCleanupToolInput    `json:"tool_input"`
	ToolResponse autoCleanupToolResponse `json:"tool_response"`
	CWD          string                  `json:"cwd"`
}

type autoCleanupToolInput struct {
	FilePath string `json:"file_path"`
}

type autoCleanupToolResponse struct {
	FilePath string `json:"filePath"`
}

// Handle reads the PostToolUse payload from stdin and checks file sizes.
func (h *AutoCleanupHandler) Handle(r io.Reader, w io.Writer) error {
	data, _ := io.ReadAll(r)

	if len(data) == 0 {
		return nil
	}

	var inp autoCleanupInput
	if err := json.Unmarshal(data, &inp); err != nil {
		return nil
	}

	filePath := inp.ToolInput.FilePath
	if filePath == "" {
		filePath = inp.ToolResponse.FilePath
	}
	if filePath == "" {
		return nil
	}

	cwd := inp.CWD
	if cwd == "" {
		if h.ProjectRoot != "" {
			cwd = h.ProjectRoot
		} else {
			cwd, _ = os.Getwd()
		}
	}

	// Normalize to a project-relative path.
	if strings.HasPrefix(filePath, cwd+"/") {
		filePath = filePath[len(cwd)+1:]
	}

	// Determine thresholds.
	plansMax := h.PlansMaxLines
	if plansMax == 0 {
		plansMax = h.envInt("PLANS_MAX_LINES", defaultPlansMaxLines)
	}
	claudeMax := h.ClaudeMdMaxLines
	if claudeMax == 0 {
		claudeMax = h.envInt("CLAUDE_MD_MAX_LINES", defaultClaudeMdMaxLines)
	}

	// Resolve the absolute path (used for file existence checks).
	absPath := filePath
	if !filepath.IsAbs(absPath) {
		absPath = filepath.Join(cwd, filePath)
	}

	feedback := h.checkFile(filePath, absPath, plansMax, claudeMax)
	if feedback == "" {
		return nil
	}

	return writeCleanupOutput(w, feedback)
}

// checkFile identifies the file and performs a size check, returning a feedback string.
func (h *AutoCleanupHandler) checkFile(relPath, absPath string, plansMax, claudeMax int) string {
	lower := strings.ToLower(relPath)
	var feedback string

	switch {
	case strings.Contains(lower, "plans.md"):
		feedback = h.checkPlans(absPath, plansMax)
	case strings.Contains(lower, "claude.md"):
		feedback = h.checkClaudeMd(absPath, claudeMax)
	}

	return feedback
}

// checkPlans checks the line count of Plans.md.
func (h *AutoCleanupHandler) checkPlans(absPath string, maxLines int) string {
	lines, err := countLines(absPath)
	if err != nil {
		return ""
	}

	var feedback string
	if lines > maxLines {
		feedback = fmt.Sprintf("⚠️ Plans.md has %d lines (limit: %d). It is recommended to archive old tasks with /harness-plan archive.", lines, maxLines)
	}

	return feedback
}

// checkClaudeMd checks the line count of CLAUDE.md.
func (h *AutoCleanupHandler) checkClaudeMd(absPath string, maxLines int) string {
	lines, err := countLines(absPath)
	if err != nil {
		return ""
	}
	if lines > maxLines {
		return fmt.Sprintf("⚠️ CLAUDE.md has %d lines. Consider splitting it into .claude/rules/ or moving content to docs/ and referencing it with @docs/filename.md.", lines)
	}
	return ""
}

// countLines counts the number of lines in a file.
func countLines(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	count := 0
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		count++
	}
	return count, sc.Err()
}

// envInt reads an environment variable as an integer, returning defaultVal when unset or unparseable.
func (h *AutoCleanupHandler) envInt(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	var n int
	if _, err := fmt.Sscanf(val, "%d", &n); err != nil {
		return defaultVal
	}
	return n
}

// writeCleanupOutput writes feedback as additionalContext in JSON format.
// Matches the simple JSON string format used by the bash counterpart.
func writeCleanupOutput(w io.Writer, feedback string) error {
	type hookOutput struct {
		HookEventName     string `json:"hookEventName"`
		AdditionalContext string `json:"additionalContext"`
	}
	type output struct {
		HookSpecificOutput hookOutput `json:"hookSpecificOutput"`
	}
	out := output{
		HookSpecificOutput: hookOutput{
			HookEventName:     "PostToolUse",
			AdditionalContext: feedback,
		},
	}
	data, err := json.Marshal(out)
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}
	_, err = fmt.Fprintf(w, "%s\n", data)
	return err
}
