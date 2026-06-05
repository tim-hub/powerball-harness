package session

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSummaryHandler_NoSessionFile(t *testing.T) {
	dir := t.TempDir()
	h := &SummaryHandler{StateDir: filepath.Join(dir, "nonexistent")}

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(`{}`), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// No output (early return)
	if out.Len() != 0 {
		t.Errorf("expected no output, got %q", out.String())
	}
}

func TestSummaryHandler_AlreadyLogged(t *testing.T) {
	dir := t.TempDir()
	stateDir := filepath.Join(dir, "state")
	if err := os.MkdirAll(stateDir, 0700); err != nil {
		t.Fatal(err)
	}

	// Create a session file with memory_logged=true
	sess := map[string]interface{}{
		"session_id":    "sess-001",
		"state":         "stopped",
		"started_at":    "2026-04-05T10:00:00Z",
		"memory_logged": true,
	}
	data, _ := json.MarshalIndent(sess, "", "  ")
	sessionFile := filepath.Join(stateDir, "session.json")
	if err := os.WriteFile(sessionFile, data, 0600); err != nil {
		t.Fatal(err)
	}

	h := &SummaryHandler{StateDir: stateDir}
	var out bytes.Buffer
	if err := h.Handle(strings.NewReader(`{}`), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// No output to prevent double execution
	if out.Len() != 0 {
		t.Errorf("expected no output for already-logged session, got %q", out.String())
	}
}

func TestSummaryHandler_FinalizesSession(t *testing.T) {
	dir := t.TempDir()
	stateDir := filepath.Join(dir, "state")
	if err := os.MkdirAll(stateDir, 0700); err != nil {
		t.Fatal(err)
	}

	fixedTime := time.Date(2026, 4, 5, 12, 30, 0, 0, time.UTC)

	// Create session file
	sess := map[string]interface{}{
		"session_id":   "sess-test-001",
		"state":        "running",
		"started_at":   "2026-04-05T12:00:00Z",
		"project_name": "test-project",
		"git": map[string]interface{}{
			"branch": "feat/test",
		},
		"memory_logged": false,
		"event_seq":     3,
		"changes_this_session": []interface{}{
			map[string]interface{}{"file": "src/foo.go", "important": false},
			map[string]interface{}{"file": "src/bar.go", "important": true},
		},
	}
	data, _ := json.MarshalIndent(sess, "", "  ")
	sessionFile := filepath.Join(stateDir, "session.json")
	if err := os.WriteFile(sessionFile, data, 0600); err != nil {
		t.Fatal(err)
	}

	h := &SummaryHandler{
		StateDir: stateDir,
		now:      func() time.Time { return fixedTime },
	}

	var out bytes.Buffer
	if err := h.Handle(strings.NewReader(`{"cwd":"`+dir+`"}`), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify session.json was finalized: memory_logged guard + stopped state
	updatedData, _ := os.ReadFile(sessionFile)
	var updatedSess map[string]interface{}
	if err := json.Unmarshal(updatedData, &updatedSess); err != nil {
		t.Fatal(err)
	}
	if !boolField(updatedSess, "memory_logged", false) {
		t.Errorf("expected memory_logged=true in session.json")
	}
	if stringField(updatedSess, "state", "") != "stopped" {
		t.Errorf("expected state=stopped in session.json")
	}

	// Verify the terminal summary surfaced the changed files
	if !strings.Contains(out.String(), "Session Summary") {
		t.Errorf("expected session summary output, got %q", out.String())
	}
}

func TestSummaryHandler_WIPTasksGateSummary(t *testing.T) {
	dir := t.TempDir()
	stateDir := filepath.Join(dir, "state")
	if err := os.MkdirAll(stateDir, 0700); err != nil {
		t.Fatal(err)
	}

	// Create plans.json: 1 WIP, 1 pm:requested, 1 TODO
	plansFile := writeSessionPlansJSON(t, dir, "cc:WIP", "pm:requested", "cc:TODO")

	// No changed files: the terminal summary should still surface because
	// WIP/requested tasks remain (the wipTasks gate in Handle).
	sess := map[string]interface{}{
		"session_id":           "sess-002",
		"state":                "running",
		"started_at":           "2026-04-05T10:00:00Z",
		"memory_logged":        false,
		"changes_this_session": []interface{}{},
	}
	data, _ := json.MarshalIndent(sess, "", "  ")
	if err := os.WriteFile(filepath.Join(stateDir, "session.json"), data, 0600); err != nil {
		t.Fatal(err)
	}

	fixedTime := time.Date(2026, 4, 5, 11, 0, 0, 0, time.UTC)
	h := &SummaryHandler{
		StateDir:  stateDir,
		PlansFile: plansFile,
		now:       func() time.Time { return fixedTime },
	}

	var out bytes.Buffer
	if err := h.Handle(strings.NewReader(`{}`), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out.String(), "Session Summary") {
		t.Errorf("expected summary output gated by WIP tasks, got %q", out.String())
	}
}

func TestSummaryHandler_ArchivesSession(t *testing.T) {
	dir := t.TempDir()
	stateDir := filepath.Join(dir, "state")
	if err := os.MkdirAll(stateDir, 0700); err != nil {
		t.Fatal(err)
	}

	sess := map[string]interface{}{
		"session_id":   "sess-archive-001",
		"state":        "running",
		"started_at":   "2026-04-05T10:00:00Z",
		"memory_logged": false,
		"changes_this_session": []interface{}{},
	}
	data, _ := json.MarshalIndent(sess, "", "  ")
	if err := os.WriteFile(filepath.Join(stateDir, "session.json"), data, 0600); err != nil {
		t.Fatal(err)
	}

	fixedTime := time.Date(2026, 4, 5, 11, 0, 0, 0, time.UTC)
	h := &SummaryHandler{
		StateDir: stateDir,
		now:      func() time.Time { return fixedTime },
	}

	var out bytes.Buffer
	if err := h.Handle(strings.NewReader(`{}`), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify archive file was created
	archiveFile := filepath.Join(stateDir, "sessions", "sess-archive-001.json")
	if _, err := os.Stat(archiveFile); err != nil {
		t.Errorf("archive file not created: %v", err)
	}
}

func TestSummaryHandler_CalcDurationMinutes(t *testing.T) {
	now := time.Date(2026, 4, 5, 12, 30, 0, 0, time.UTC)
	h := &SummaryHandler{now: func() time.Time { return now }}

	tests := []struct {
		startedAt string
		want      int
	}{
		{"2026-04-05T12:00:00Z", 30},
		{"2026-04-05T12:29:00Z", 1},
		{"", 0},
		{"null", 0},
		{"invalid", 0},
	}

	for _, tt := range tests {
		got := h.calcDurationMinutes(tt.startedAt, now)
		if got != tt.want {
			t.Errorf("calcDurationMinutes(%q) = %d, want %d", tt.startedAt, got, tt.want)
		}
	}
}

func TestSummaryHandler_EmptyInput(t *testing.T) {
	dir := t.TempDir()
	h := &SummaryHandler{StateDir: filepath.Join(dir, "state")}

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(""), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStringField(t *testing.T) {
	m := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	}
	if got := stringField(m, "key1", "default"); got != "value1" {
		t.Errorf("expected value1, got %q", got)
	}
	if got := stringField(m, "key2", "default"); got != "default" {
		t.Errorf("expected default (non-string), got %q", got)
	}
	if got := stringField(m, "missing", "fallback"); got != "fallback" {
		t.Errorf("expected fallback, got %q", got)
	}
}

func TestIntField(t *testing.T) {
	m := map[string]interface{}{
		"n1": float64(42),
		"n2": "not-int",
	}
	if got := intField(m, "n1", 0); got != 42 {
		t.Errorf("expected 42, got %d", got)
	}
	if got := intField(m, "n2", 99); got != 99 {
		t.Errorf("expected 99, got %d", got)
	}
	if got := intField(m, "missing", 7); got != 7 {
		t.Errorf("expected 7, got %d", got)
	}
}

func TestAppendLine(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.jsonl")

	appendLine(path, `{"a":1}`)
	appendLine(path, `{"b":2}`)

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d: %q", len(lines), string(data))
	}
	if lines[0] != `{"a":1}` {
		t.Errorf("expected first line={\"a\":1}, got %q", lines[0])
	}
}

func TestCopyFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")

	if err := os.WriteFile(src, []byte("content"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := copyFile(src, dst, 0600); err != nil {
		t.Fatalf("copyFile failed: %v", err)
	}
	data, _ := os.ReadFile(dst)
	if string(data) != "content" {
		t.Errorf("expected 'content', got %q", data)
	}
}

// suppress unused import warning
var _ = fmt.Sprintf
