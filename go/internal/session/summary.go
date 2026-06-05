package session

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SummaryHandler is the Stop hook handler (session end summary).
// Records end info in session.json, logs the stop event, and archives session
// files when the session ends.
//
// shell version: scripts/session-summary.sh
type SummaryHandler struct {
	// StateDir is the path to the state directory. Inferred from cwd if empty.
	StateDir string
	// PlansFile is the path to plans.json. Defaults to the canonical
	// projectRoot/.claude/harness/plans.json if empty.
	PlansFile string
	// now is the injected current time function (for testing). Uses time.Now() if nil.
	now func() time.Time
}

// summaryInput is the stdin JSON for the Stop hook.
type summaryInput struct {
	CWD       string `json:"cwd,omitempty"`
	SessionID string `json:"session_id,omitempty"`
}

// summarySessionData is the data read from session.json.
type summarySessionData struct {
	SessionID     string  `json:"session_id"`
	State         string  `json:"state"`
	StartedAt     string  `json:"started_at"`
	ProjectName   string  `json:"project_name"`
	GitBranch     string  `json:"-"` // nested
	MemoryLogged  bool    `json:"memory_logged"`
	EventSeq      int     `json:"event_seq"`
	ChangesCount  int     `json:"-"` // computed value

	// raw JSON (for field additions)
	raw map[string]interface{}
}

// Handle reads the Stop payload from stdin and finalizes the session: it logs
// the stop event, records end info in session.json, and archives session files.
func (h *SummaryHandler) Handle(r io.Reader, w io.Writer) error {
	data, _ := io.ReadAll(r)

	var inp summaryInput
	if len(data) > 0 {
		_ = json.Unmarshal(data, &inp)
	}

	projectRoot := resolveProjectRoot(inp.CWD)

	stateDir := h.StateDir
	if stateDir == "" {
		stateDir = filepath.Join(projectRoot, ".claude", "state")
	}

	sessionFile := filepath.Join(stateDir, "session.json")
	eventLogFile := filepath.Join(stateDir, "session.events.jsonl")
	archiveDir := filepath.Join(stateDir, "sessions")

	now := h.currentTime()
	nowStr := now.UTC().Format(time.RFC3339)

	// Skip if session.json does not exist
	if _, err := os.Stat(sessionFile); err != nil {
		return nil
	}

	// Read session.json
	sessData := h.readSessionData(sessionFile)

	// Prevent double-execution (skip if memory_logged is true)
	if sessData.MemoryLogged {
		return nil
	}

	// Calculate session duration
	durationMinutes := h.calcDurationMinutes(sessData.StartedAt, now)

	// Get WIP tasks from plans.json (used to gate the terminal summary)
	wipTasks := h.readWIPTasks(h.PlansFile, projectRoot)

	// Get changed file information
	changedFiles, _ := h.readChangedFiles(sessData.raw)

	// Log session.stop event
	h.appendEvent(eventLogFile, sessionFile, "session.stop", "stopped", nowStr)

	// Update session.json with end information
	h.finalizeSessionFile(sessionFile, nowStr, durationMinutes)

	// Archive session files
	h.archiveSession(sessionFile, eventLogFile, archiveDir, sessData.SessionID)

	// Write summary to stdout (only if there are changes)
	if len(changedFiles) > 0 || len(wipTasks) > 0 {
		h.writeSummaryOutput(w, sessData, durationMinutes, changedFiles, wipTasks)
	}

	_ = w
	return nil
}

// readSessionData reads the necessary information from session.json.
func (h *SummaryHandler) readSessionData(sessionFile string) summarySessionData {
	data, err := os.ReadFile(sessionFile)
	if err != nil {
		return summarySessionData{}
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return summarySessionData{}
	}

	sess := summarySessionData{raw: raw}
	sess.SessionID = stringField(raw, "session_id", "unknown")
	sess.State = stringField(raw, "state", "")
	sess.StartedAt = stringField(raw, "started_at", "")
	sess.ProjectName = stringField(raw, "project_name", "")
	sess.MemoryLogged = boolField(raw, "memory_logged", false)
	sess.EventSeq = intField(raw, "event_seq", 0)

	// Changed file count
	if changes, ok := raw["changes_this_session"].([]interface{}); ok {
		sess.ChangesCount = len(changes)
	}

	return sess
}

// readChangedFiles reads the list of changed files from session.json.
func (h *SummaryHandler) readChangedFiles(raw map[string]interface{}) ([]string, []string) {
	changes, ok := raw["changes_this_session"].([]interface{})
	if !ok {
		return nil, nil
	}

	seen := map[string]bool{}
	var changedFiles, importantFiles []string

	for _, item := range changes {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		file := stringField(m, "file", "")
		if file == "" || seen[file] {
			continue
		}
		seen[file] = true
		changedFiles = append(changedFiles, file)
		if boolField(m, "important", false) {
			importantFiles = append(importantFiles, file)
		}
	}
	return changedFiles, importantFiles
}

// readWIPTasks reads WIP/in-progress tasks from plans.json. Each returned line
// summarizes one task as "<id> <name> (<status>)". Up to 20 tasks are returned.
func (h *SummaryHandler) readWIPTasks(plansFile, projectRoot string) []string {
	p, _ := loadPlansJSON(plansFile, projectRoot)
	if p == nil {
		return nil
	}

	var tasks []string
	for _, t := range p.AllTasks() {
		if t.Status != "cc:WIP" && t.Status != "pm:requested" {
			continue
		}
		line := strings.TrimSpace(fmt.Sprintf("%s %s (%s)", t.ID, t.Name, t.Status))
		tasks = append(tasks, line)
		if len(tasks) >= 20 {
			break
		}
	}
	return tasks
}

// appendEvent appends one entry to the event log and updates EventSeq in session.json.
func (h *SummaryHandler) appendEvent(eventLogFile, sessionFile, eventType, state, ts string) {
	if isSymlink(eventLogFile) || isSymlink(sessionFile) {
		return
	}

	// Read EventSeq from session.json and increment
	seq := 0
	if data, err := os.ReadFile(sessionFile); err == nil {
		var raw map[string]interface{}
		if json.Unmarshal(data, &raw) == nil {
			seq = intField(raw, "event_seq", 0)
		}
	}
	seq++
	eventID := fmt.Sprintf("event-%06d", seq)

	entry := fmt.Sprintf(`{"id":%q,"type":%q,"ts":%q,"state":%q}`, eventID, eventType, ts, state)
	appendLine(eventLogFile, entry)
}

// finalizeSessionFile records end information in session.json.
func (h *SummaryHandler) finalizeSessionFile(sessionFile, endedAt string, durationMinutes int) {
	if isSymlink(sessionFile) {
		return
	}

	data, err := os.ReadFile(sessionFile)
	if err != nil {
		return
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return
	}

	raw["ended_at"] = endedAt
	raw["duration_minutes"] = durationMinutes
	raw["memory_logged"] = true
	raw["state"] = "stopped"

	out, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return
	}

	_ = writeFileAtomic(sessionFile, append(out, '\n'), 0600)
}

// archiveSession copies session files to the archive directory.
func (h *SummaryHandler) archiveSession(sessionFile, eventLogFile, archiveDir, sessionID string) {
	if sessionID == "" {
		return
	}
	if err := os.MkdirAll(archiveDir, 0700); err != nil {
		return
	}

	archivePath := filepath.Join(archiveDir, sessionID+".json")
	if !isSymlink(archivePath) {
		_ = copyFile(sessionFile, archivePath, 0600)
	}

	archiveEvents := filepath.Join(archiveDir, sessionID+".events.jsonl")
	if !isSymlink(archiveEvents) {
		_ = copyFile(eventLogFile, archiveEvents, 0600)
	}
}

// writeSummaryOutput outputs the session end summary.
func (h *SummaryHandler) writeSummaryOutput(w io.Writer, sess summarySessionData, durationMinutes int, changedFiles, wipTasks []string) {
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Session Summary")
	fmt.Fprintln(w, strings.Repeat("─", 32))

	if sess.ProjectName != "" {
		fmt.Fprintf(w, "Project: %s\n", sess.ProjectName)
	}
	fmt.Fprintf(w, "Changed files: %d\n", len(changedFiles))
	if durationMinutes > 0 {
		fmt.Fprintf(w, "Session duration: %d min\n", durationMinutes)
	}

	fmt.Fprintln(w, strings.Repeat("─", 32))
	fmt.Fprintln(w, "")
}

// calcDurationMinutes calculates the number of minutes from session start time to now.
func (h *SummaryHandler) calcDurationMinutes(startedAt string, now time.Time) int {
	if startedAt == "" || startedAt == "null" {
		return 0
	}
	t, err := time.Parse(time.RFC3339, startedAt)
	if err != nil {
		return 0
	}
	return int(now.Sub(t).Minutes())
}

// currentTime returns the current time.
func (h *SummaryHandler) currentTime() time.Time {
	if h.now != nil {
		return h.now()
	}
	return time.Now()
}

// ---------------------------------------------------------------------------
// Helper functions
// ---------------------------------------------------------------------------

// stringField safely retrieves a string field from a map.
func stringField(m map[string]interface{}, key, defaultVal string) string {
	v, ok := m[key]
	if !ok {
		return defaultVal
	}
	s, ok := v.(string)
	if !ok {
		return defaultVal
	}
	return s
}

// boolField safely retrieves a bool field from a map.
func boolField(m map[string]interface{}, key string, defaultVal bool) bool {
	v, ok := m[key]
	if !ok {
		return defaultVal
	}
	b, ok := v.(bool)
	if !ok {
		return defaultVal
	}
	return b
}

// intField safely retrieves an int field from a map (JSON numbers are float64).
func intField(m map[string]interface{}, key string, defaultVal int) int {
	v, ok := m[key]
	if !ok {
		return defaultVal
	}
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	}
	return defaultVal
}

// appendLine appends a text line to path.
func appendLine(path, line string) {
	if isSymlink(path) {
		return
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = fmt.Fprintln(f, line)
}

// copyFile copies a file.
func copyFile(src, dst string, perm os.FileMode) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, perm)
}

// newLineScanner is a wrapper around bufio.Scanner (to avoid test dependencies).
func newLineScanner(r io.Reader) *lineScanner {
	return &lineScanner{inner: &bufioScanner{r: r}}
}

type lineScanner struct {
	inner interface {
		Scan() bool
		Text() string
	}
}

func (ls *lineScanner) Scan() bool {
	return ls.inner.Scan()
}

func (ls *lineScanner) Text() string {
	return ls.inner.Text()
}

// bufioScanner wraps io.Reader to provide line scanning.
type bufioScanner struct {
	r    io.Reader
	buf  []byte
	pos  int
	end  int
	done bool
	cur  string
}

func (bs *bufioScanner) Scan() bool {
	if bs.done {
		return false
	}
	// Initialize read buffer
	if bs.buf == nil {
		data, err := io.ReadAll(bs.r)
		if err != nil {
			bs.done = true
			return false
		}
		bs.buf = data
	}
	if bs.pos >= len(bs.buf) {
		bs.done = true
		return false
	}
	// Find the next newline
	end := bs.pos
	for end < len(bs.buf) && bs.buf[end] != '\n' {
		end++
	}
	bs.cur = string(bs.buf[bs.pos:end])
	if end < len(bs.buf) {
		bs.pos = end + 1
	} else {
		bs.pos = end
		bs.done = true
	}
	return true
}

func (bs *bufioScanner) Text() string {
	return bs.cur
}
