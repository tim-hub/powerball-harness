package hookhandler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// plansWatcherDeps holds injectable dependencies so tests can substitute
// syscall.Flock, time.Sleep, and os.Exit without touching package-level state.
// This eliminates the global-mutation pattern that breaks t.Parallel().
type plansWatcherDeps struct {
	flock  func(fd int, how int) error
	sleep  func(time.Duration)
	exitFn func(msg string) // called on the fail-closed path instead of os.Exit(1)
}

func defaultPlansWatcherDeps() plansWatcherDeps {
	return plansWatcherDeps{
		flock: func(fd int, how int) error { return syscall.Flock(fd, how) },
		sleep: time.Sleep,
		exitFn: func(msg string) {
			fmt.Fprintf(os.Stderr, "[plans-watcher] fail-closed exit: %s\n", msg)
			os.Exit(1)
		},
	}
}

// plansWatcher is the stateless handler; deps is the only mutable surface.
type plansWatcher struct {
	deps plansWatcherDeps
}

// plansWatcherInput is the stdin JSON passed to the plans-watcher hook.
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

// plansLockFile is the flock file path used for exclusive access to plans-state.json.
// Semantically equivalent to the 3-tier fallback in the shell version scripts/plans-watcher.sh.
const plansLockFile = ".claude/state/locks/plans.flock"

// plansLockDirSuffix is the suffix for the mkdir-based fallback lock used when flock is unavailable.
const plansLockDirSuffix = ".dir"

// plansLockMaxRetries is the maximum number of lock acquisition retries.
const plansLockMaxRetries = 3

// plansLockHandle represents either a flock or mkdir-based fallback lock.
type plansLockHandle struct {
	file    *os.File
	lockDir string
	mode    string
}

// acquireLock acquires an exclusive lock protecting plans-state.json.
// Normally uses flock; falls back to mkdir-based atomic lock when flock is unavailable
// (e.g. shared/network storage).
func (w *plansWatcher) acquireLock(lockPath string) (*plansLockHandle, error) {
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		return nil, fmt.Errorf("mkdir for plans lock: %w", err)
	}
	for attempt := 1; attempt <= plansLockMaxRetries; attempt++ {
		f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o644)
		if err != nil {
			return nil, fmt.Errorf("open plans lock file: %w", err)
		}

		if err := w.deps.flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err == nil {
			return &plansLockHandle{
				file:    f,
				lockDir: lockPath + plansLockDirSuffix,
				mode:    "flock",
			}, nil
		} else if !isPlansLockBusy(err) {
			f.Close()
			return w.acquireMkdirLock(lockPath + plansLockDirSuffix)
		}

		f.Close()
		if attempt < plansLockMaxRetries {
			w.deps.sleep(1 * time.Second)
		}
	}
	return nil, fmt.Errorf("failed to acquire plans lock after %d retries", plansLockMaxRetries)
}

func (w *plansWatcher) acquireMkdirLock(lockDir string) (*plansLockHandle, error) {
	for attempt := 1; attempt <= plansLockMaxRetries; attempt++ {
		if err := os.Mkdir(lockDir, 0o755); err == nil {
			return &plansLockHandle{
				lockDir: lockDir,
				mode:    "mkdir",
			}, nil
		} else if !errors.Is(err, os.ErrExist) {
			return nil, fmt.Errorf("mkdir fallback lock: %w", err)
		}

		if attempt < plansLockMaxRetries {
			w.deps.sleep(1 * time.Second)
		}
	}
	return nil, fmt.Errorf("failed to acquire mkdir fallback lock after %d retries", plansLockMaxRetries)
}

func isPlansLockBusy(err error) bool {
	return errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EAGAIN)
}

// releaseLock releases the lock and closes any open file handles.
func (w *plansWatcher) releaseLock(lock *plansLockHandle) {
	if lock == nil {
		return
	}
	switch lock.mode {
	case "mkdir":
		os.Remove(lock.lockDir) //nolint:errcheck
	default:
		if lock.file == nil {
			return
		}
		w.deps.flock(int(lock.file.Fd()), syscall.LOCK_UN) //nolint:errcheck
		lock.file.Close()
	}
}

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

// HandlePlansWatcher is the public entry point used by the post-tool-batch hook.
// It constructs a plansWatcher with production defaults and delegates to handle().
func HandlePlansWatcher(in io.Reader, out io.Writer) error {
	w := &plansWatcher{deps: defaultPlansWatcherDeps()}
	return w.handle(in, out)
}

// handle is the Go port of plans-watcher.sh.
//
// Called on PostToolUse Write/Edit events to detect changes to Plans.md.
// Generates an aggregated summary of WIP/TODO/done markers and writes a PM notification file.
// Files other than Plans.md are skipped.
func (w *plansWatcher) handle(in io.Reader, out io.Writer) error {
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

	// Determine the effective CWD for deriving lock and state file paths.
	// By using the same CWD for both lock and state paths, we ensure that
	// two hooks running in different worktrees (CWD A / CWD B) use:
	//   - same CWD → same lock + same state (correct serialization)
	//   - different CWD → different lock + different state (correct isolation)
	cwd := input.CWD
	if cwd == "" {
		var cwdErr error
		cwd, cwdErr = os.Getwd()
		if cwdErr != nil {
			cwd = ""
		}
	}

	// derive lock and state file paths from the same cwd
	lockPath := plansLockFile
	stateFilePath := plansStateFile
	if cwd != "" {
		lockPath = filepath.Join(cwd, plansLockFile)
		// plansStateFile is a relative path constant; join with cwd to get an absolute path
		stateFilePath = filepath.Join(cwd, plansStateFile)
	}

	// Protect the plans-state.json read-modify-write with flock.
	// Fail-closed: if lock acquisition fails, abort rather than silently dropping the update.
	// Per the PostToolUse hook spec, exit code 1 signals an error to the hook framework.
	// Returning emptyPostToolOutput (= empty success) would cause the framework to treat
	// a lock failure as success, resulting in lost updates. exitFailClosed exits with code 1
	// to explicitly signal the error.
	lockFile, lockErr := w.acquireLock(lockPath)
	if lockErr != nil {
		fmt.Fprintf(os.Stderr, "[plans-watcher] lock acquisition failed (fail-closed): %v\n", lockErr)
		w.deps.exitFn("lock acquisition timed out (3 retries exhausted)")
		// exitFn normally calls os.Exit(1), but if mocked in tests to be a no-op,
		// fall back to an empty response.
		return emptyPostToolOutput(out)
	}
	defer w.releaseLock(lockFile)

	// aggregate the current state
	current, err := collectPlansState(plansFile)
	if err != nil {
		return emptyPostToolOutput(out)
	}

	// load the previous state (using CWD-based absolute path)
	prev := loadPrevPlansState(stateFilePath)

	// save the current state (using CWD-based absolute path)
	stateDir := filepath.Dir(stateFilePath)
	if mkErr := os.MkdirAll(stateDir, 0o755); mkErr == nil {
		savePlansState(stateFilePath, current)
	}

	// determine the type of change
	hasNewTasks := current.PmPending > prev.PmPending
	hasCompletedTasks := current.CcDone > prev.CcDone

	if !hasNewTasks && !hasCompletedTasks {
		return emptyPostToolOutput(out)
	}

	// generate the PM notification file
	if err := writePMNotification(cwd, current, hasNewTasks, hasCompletedTasks); err != nil {
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
// stateFilePath accepts both absolute and relative paths.
func loadPrevPlansState(stateFilePath string) plansState {
	data, err := os.ReadFile(stateFilePath)
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
// stateFilePath accepts both absolute and relative paths.
func savePlansState(stateFilePath string, state plansState) {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return
	}
	os.WriteFile(stateFilePath, append(data, '\n'), 0o644) //nolint:errcheck
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
// cwd is used to resolve the notification file paths to absolute paths.
func writePMNotification(cwd string, state plansState, hasNewTasks, hasCompletedTasks bool) error {
	pmPath := pmNotificationFile
	cursorPath := cursorNotificationFile
	if cwd != "" {
		pmPath = filepath.Join(cwd, pmNotificationFile)
		cursorPath = filepath.Join(cwd, cursorNotificationFile)
	}

	stateDir := filepath.Dir(pmPath)
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
	if err := os.WriteFile(pmPath, content, 0o644); err != nil {
		return fmt.Errorf("write pm-notification.md: %w", err)
	}

	// compat: also copy to cursor-notification.md
	_ = os.WriteFile(cursorPath, content, 0o644)

	return nil
}
