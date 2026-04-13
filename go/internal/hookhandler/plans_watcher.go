package hookhandler

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// plansWatcherInput is the stdin JSON passed to plans-watcher.sh.
type plansWatcherInput struct {
	ToolName string `json:"tool_name"`
	CWD      string `json:"cwd"`
	ToolInput struct {
		FilePath string `json:"file_path"`
	} `json:"tool_input"`
	ToolResponse struct {
		FilePath string `json:"filePath"`
	} `json:"tool_response"`
}

// plansStateFile is the path to the file that stores the previous state.
const plansStateFile = ".claude/state/plans-state.json"

// pmNotificationFile is the path to the PM notification file.
const pmNotificationFile = ".claude/state/pm-notification.md"

// cursorNotificationFile is the path to the compatibility cursor notification file.
const cursorNotificationFile = ".claude/state/cursor-notification.md"

// plansState holds the aggregated marker counts from Plans.md.
type plansState struct {
	Timestamp   string `json:"timestamp"`
	PmPending   int    `json:"pm_pending"`
	CcTodo      int    `json:"cc_todo"`
	CcWip       int    `json:"cc_wip"`
	CcDone      int    `json:"cc_done"`
	PmConfirmed int    `json:"pm_confirmed"`
}

// plansFileNames lists the candidate Plans.md file names to search for.
var plansFileNames = []string{"Plans.md", "plans.md", "PLANS.md", "PLANS.MD"}

// HandlePlansWatcher is the Go port of plans-watcher.sh.
//
// Called on PostToolUse Write/Edit events to detect changes to Plans.md.
// Generates an aggregated summary of WIP/TODO/done markers and writes a PM notification file.
// Files other than Plans.md are skipped.
func HandlePlansWatcher(in io.Reader, out io.Writer) error {
	data, err := io.ReadAll(in)
	if err != nil {
		return emptyPostToolOutput(out)
	}

	if len(strings.TrimSpace(string(data))) == 0 {
		return emptyPostToolOutput(out)
	}

	var input plansWatcherInput
	if err := json.Unmarshal(data, &input); err != nil {
		return emptyPostToolOutput(out)
	}

	// get the changed file path
	changedFile := input.ToolInput.FilePath
	if changedFile == "" {
		changedFile = input.ToolResponse.FilePath
	}

	if changedFile == "" {
		return emptyPostToolOutput(out)
	}

	// convert to relative path if CWD is available
	if input.CWD != "" {
		changedFile = makeRelativePath(
			normalizePathSeparators(changedFile),
			normalizePathSeparators(input.CWD),
		)
	}

	// locate the Plans.md file (supports plansDirectory in config)
	// Use input.CWD as projectRoot when available.
	// Fixes an issue where the hook process CWD differs from input.CWD, causing the wrong Plans.md to be referenced.
	projectRoot := input.CWD
	if projectRoot == "" {
		projectRoot = resolveProjectRoot()
	}
	plansFile := resolvePlansPath(projectRoot)
	if plansFile == "" {
		return emptyPostToolOutput(out)
	}

	// skip if the changed file is not Plans.md (strict full-path comparison)
	if !isPlansFileWithRoot(changedFile, plansFile, projectRoot) {
		return emptyPostToolOutput(out)
	}

	// aggregate the current state
	current, err := collectPlansState(plansFile)
	if err != nil {
		return emptyPostToolOutput(out)
	}

	// load the previous state
	prev := loadPrevPlansState()

	// save the current state
	stateDir := filepath.Dir(plansStateFile)
	if mkErr := os.MkdirAll(stateDir, 0o755); mkErr == nil {
		savePlansState(current)
	}

	// determine the type of change
	hasNewTasks := current.PmPending > prev.PmPending
	hasCompletedTasks := current.CcDone > prev.CcDone

	if !hasNewTasks && !hasCompletedTasks {
		return emptyPostToolOutput(out)
	}

	// generate the PM notification file
	if err := writePMNotification(current, hasNewTasks, hasCompletedTasks); err != nil {
		fmt.Fprintf(os.Stderr, "[plans-watcher] write notification: %v\n", err)
	}

	// output the notification summary via systemMessage
	summary := buildSummaryMessage(current, hasNewTasks, hasCompletedTasks)
	o := postToolOutput{}
	o.HookSpecificOutput.HookEventName = "PostToolUse"
	o.HookSpecificOutput.AdditionalContext = summary
	return writeJSON(out, o)
}

// findPlansFile searches the current directory for a Plans.md file.
func findPlansFile() string {
	for _, name := range plansFileNames {
		if _, err := os.Stat(name); err == nil {
			return name
		}
	}
	return ""
}

// isPlansFile returns true if the changed file is Plans.md.
//
// Matching logic:
//  1. Exact match after filepath.Clean (handles both relative and absolute paths)
//  2. If changedFile is a relative path, convert to absolute using projectRoot before comparing
//
// The case-insensitive basename fallback from the old implementation has been removed.
// Basename-only comparison would incorrectly match files with the same name in different
// directories (e.g. /tmp/other/Plans.md), so only strict full-path matching is used.
func isPlansFile(changedFile, plansFile string) bool {
	// normalize with filepath.Clean and check for exact match
	if filepath.Clean(changedFile) == filepath.Clean(plansFile) {
		return true
	}
	return false
}

// isPlansFileWithRoot supplements changedFile with projectRoot when it is a relative path.
// Used when called from HandlePlansWatcher.
func isPlansFileWithRoot(changedFile, plansFile, projectRoot string) bool {
	// if changedFile is absolute, compare directly
	if filepath.IsAbs(changedFile) {
		return isPlansFile(changedFile, plansFile)
	}
	// if relative, convert to absolute using projectRoot
	absChanged := filepath.Join(projectRoot, changedFile)
	return isPlansFile(absChanged, plansFile)
}

// countMarker returns the number of occurrences of the marker string in Plans.md.
func countMarker(plansFile, marker string) int {
	data, err := os.ReadFile(plansFile)
	if err != nil {
		return 0
	}
	re := regexp.MustCompile(regexp.QuoteMeta(marker))
	return len(re.FindAllIndex(data, -1))
}

// collectPlansState aggregates the markers in Plans.md.
func collectPlansState(plansFile string) (plansState, error) {
	if _, err := os.Stat(plansFile); err != nil {
		return plansState{}, fmt.Errorf("plans file not found: %w", err)
	}

	pmPending := countMarker(plansFile, "pm:pending") + countMarker(plansFile, "cursor:pending")
	ccTodo := countMarker(plansFile, "cc:TODO")
	ccWip := countMarker(plansFile, "cc:WIP")
	ccDone := countMarker(plansFile, "cc:done")
	pmConfirmed := countMarker(plansFile, "pm:confirmed") + countMarker(plansFile, "cursor:confirmed")

	return plansState{
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		PmPending:   pmPending,
		CcTodo:      ccTodo,
		CcWip:       ccWip,
		CcDone:      ccDone,
		PmConfirmed: pmConfirmed,
	}, nil
}

// loadPrevPlansState loads the previously saved state. Returns zero value if not found.
func loadPrevPlansState() plansState {
	data, err := os.ReadFile(plansStateFile)
	if err != nil {
		return plansState{}
	}
	var state plansState
	if err := json.Unmarshal(data, &state); err != nil {
		return plansState{}
	}
	return state
}

// savePlansState saves the current state to a file.
func savePlansState(state plansState) {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return
	}
	os.WriteFile(plansStateFile, append(data, '\n'), 0o644)
}

// buildSummaryMessage constructs the notification summary string.
func buildSummaryMessage(state plansState, hasNewTasks, hasCompletedTasks bool) string {
	var sb strings.Builder

	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString("Plans.md update detected\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	if hasNewTasks {
		sb.WriteString("New tasks: requested by PM\n")
		sb.WriteString("   → Check status with /sync-status, then start work with /work\n")
	}

	if hasCompletedTasks {
		sb.WriteString("Tasks completed: ready to report to PM\n")
		sb.WriteString("   → Report with /handoff-to-pm-claude (or /handoff-to-cursor)\n")
	}

	sb.WriteString("\nCurrent status:\n")
	sb.WriteString("   pm:pending      : " + strconv.Itoa(state.PmPending) + "\n")
	sb.WriteString("   cc:TODO        : " + strconv.Itoa(state.CcTodo) + "\n")
	sb.WriteString("   cc:WIP         : " + strconv.Itoa(state.CcWip) + "\n")
	sb.WriteString("   cc:done        : " + strconv.Itoa(state.CcDone) + "\n")
	sb.WriteString("   pm:confirmed      : " + strconv.Itoa(state.PmConfirmed) + "\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	return sb.String()
}

// writePMNotification generates the PM notification file.
func writePMNotification(state plansState, hasNewTasks, hasCompletedTasks bool) error {
	stateDir := filepath.Dir(pmNotificationFile)
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		return fmt.Errorf("mkdir state dir: %w", err)
	}

	ts := time.Now().Format("2006-01-02 15:04:05")

	var sb strings.Builder
	sb.WriteString("# Notification to PM\n\n")
	sb.WriteString("**Generated at**: " + ts + "\n\n")
	sb.WriteString("## Status changes\n\n")

	if hasNewTasks {
		sb.WriteString("### New tasks\n\n")
		sb.WriteString("New tasks have been requested by PM (pm:pending / compat: cursor:pending).\n\n")
	}

	if hasCompletedTasks {
		sb.WriteString("### Completed tasks\n\n")
		sb.WriteString("Impl Claude has completed tasks. Please review (cc:done).\n\n")
	}

	sb.WriteString("---\n\n")
	sb.WriteString("**Next action**: Review in PM Claude and re-delegate if needed (/handoff-to-impl-claude).\n")

	content := []byte(sb.String())
	if err := os.WriteFile(pmNotificationFile, content, 0o644); err != nil {
		return fmt.Errorf("write pm-notification.md: %w", err)
	}

	// compat: also copy to cursor-notification.md
	_ = os.WriteFile(cursorNotificationFile, content, 0o644)

	return nil
}
