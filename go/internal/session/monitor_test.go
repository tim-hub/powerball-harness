package session

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMonitorHandler_GeneratesSessionFile(t *testing.T) {
	dir := t.TempDir()
	stateDir := filepath.Join(dir, "state")
	plansFile := filepath.Join(dir, "Plans.md")

	// Create Plans.md
	plans := "| t1 | cc:WIP |\n| t2 | cc:TODO |\n"
	if err := os.WriteFile(plansFile, []byte(plans), 0644); err != nil {
		t.Fatal(err)
	}

	h := &MonitorHandler{
		StateDir:  stateDir,
		PlansFile: plansFile,
		now:       func() time.Time { return time.Date(2026, 4, 5, 12, 0, 0, 0, time.UTC) },
	}

	inp := `{"cwd":"` + dir + `"}`
	var out bytes.Buffer
	if err := h.Handle(strings.NewReader(inp), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify session.json was created
	sessionFile := filepath.Join(stateDir, "session.json")
	data, err := os.ReadFile(sessionFile)
	if err != nil {
		t.Fatalf("session.json not created: %v", err)
	}

	var sess sessionStateJSON
	if err := json.Unmarshal(data, &sess); err != nil {
		t.Fatalf("invalid session.json: %v", err)
	}

	if sess.State != "initialized" {
		t.Errorf("expected state=initialized, got %q", sess.State)
	}
	if sess.SessionID == "" {
		t.Errorf("expected non-empty session_id")
	}
	if sess.Plans.WIPTasks != 1 {
		t.Errorf("expected wip_tasks=1, got %d", sess.Plans.WIPTasks)
	}
	if sess.Plans.TODOTasks != 1 {
		t.Errorf("expected todo_tasks=1, got %d", sess.Plans.TODOTasks)
	}
}

func TestMonitorHandler_GeneratesToolingPolicy(t *testing.T) {
	dir := t.TempDir()
	stateDir := filepath.Join(dir, "state")

	h := &MonitorHandler{
		StateDir:  stateDir,
		PlansFile: filepath.Join(dir, "Plans.md"),
	}

	var out bytes.Buffer
	if err := h.Handle(strings.NewReader(`{"cwd":"`+dir+`"}`), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	policyFile := filepath.Join(stateDir, "tooling-policy.json")
	data, err := os.ReadFile(policyFile)
	if err != nil {
		t.Fatalf("tooling-policy.json not created: %v", err)
	}

	var policy toolingPolicyJSON
	if err := json.Unmarshal(data, &policy); err != nil {
		t.Fatalf("invalid tooling-policy.json: %v\nraw: %s", err, data)
	}

	if policy.LSP.Available {
		t.Errorf("expected lsp.available=false")
	}
	if policy.Skills.DecisionRequired {
		t.Errorf("expected skills.decision_required=false")
	}
}

func TestMonitorHandler_ResumesExistingSession(t *testing.T) {
	dir := t.TempDir()
	stateDir := filepath.Join(dir, "state")
	if err := os.MkdirAll(stateDir, 0700); err != nil {
		t.Fatal(err)
	}

	// Create existing session
	existingSession := sessionStateJSON{
		SessionID:   "session-existing",
		State:       "running",
		StateVersion: 1,
		StartedAt:   "2026-04-05T10:00:00Z",
		UpdatedAt:   "2026-04-05T10:00:00Z",
		ResumeToken: "resume-token",
		EventSeq:    5,
		Plans:       plansStateJSON{Exists: false},
		Git:         gitStateJSON{Branch: "main"},
		ChangesThisSession: []interface{}{},
	}
	existingData, _ := json.MarshalIndent(existingSession, "", "  ")
	sessionFile := filepath.Join(stateDir, "session.json")
	if err := os.WriteFile(sessionFile, existingData, 0600); err != nil {
		t.Fatal(err)
	}

	h := &MonitorHandler{
		StateDir:  stateDir,
		PlansFile: filepath.Join(dir, "Plans.md"),
	}

	var out bytes.Buffer
	if err := h.Handle(strings.NewReader(`{"cwd":"`+dir+`"}`), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Resumed session should retain the same session_id
	data, _ := os.ReadFile(sessionFile)
	var sess sessionStateJSON
	if err := json.Unmarshal(data, &sess); err != nil {
		t.Fatal(err)
	}

	if sess.SessionID != "session-existing" {
		t.Errorf("expected session_id=session-existing (resume), got %q", sess.SessionID)
	}
	if sess.ResumeToken != "resume-token" {
		t.Errorf("expected resume_token preserved, got %q", sess.ResumeToken)
	}
}

func TestMonitorHandler_NewSessionOnStopped(t *testing.T) {
	dir := t.TempDir()
	stateDir := filepath.Join(dir, "state")
	if err := os.MkdirAll(stateDir, 0700); err != nil {
		t.Fatal(err)
	}

	// Create stopped session
	existingSession := map[string]interface{}{
		"session_id": "session-old",
		"state":      "stopped",
		"started_at": "2026-04-04T10:00:00Z",
	}
	existingData, _ := json.MarshalIndent(existingSession, "", "  ")
	sessionFile := filepath.Join(stateDir, "session.json")
	if err := os.WriteFile(sessionFile, existingData, 0600); err != nil {
		t.Fatal(err)
	}

	h := &MonitorHandler{
		StateDir:  stateDir,
		PlansFile: filepath.Join(dir, "Plans.md"),
	}

	var out bytes.Buffer
	if err := h.Handle(strings.NewReader(`{"cwd":"`+dir+`"}`), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(sessionFile)
	var sess sessionStateJSON
	if err := json.Unmarshal(data, &sess); err != nil {
		t.Fatal(err)
	}

	// A new session_id should have been generated
	if sess.SessionID == "session-old" {
		t.Errorf("expected new session_id, got session-old")
	}
	if sess.State != "initialized" {
		t.Errorf("expected state=initialized, got %q", sess.State)
	}
}

func TestMonitorHandler_SymlinkStateDir(t *testing.T) {
	dir := t.TempDir()
	realDir := filepath.Join(dir, "real-state")
	if err := os.MkdirAll(realDir, 0700); err != nil {
		t.Fatal(err)
	}
	linkDir := filepath.Join(dir, "link-state")
	if err := os.Symlink(realDir, linkDir); err != nil {
		t.Skip("symlink creation not supported")
	}

	h := &MonitorHandler{StateDir: linkDir}
	var out bytes.Buffer
	// Should not return an error (early return)
	err := h.Handle(strings.NewReader(`{}`), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMonitorHandler_ReadGitBranch(t *testing.T) {
	dir := t.TempDir()
	gitDir := filepath.Join(dir, ".git")
	if err := os.MkdirAll(gitDir, 0700); err != nil {
		t.Fatal(err)
	}

	// Create HEAD file
	if err := os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte("ref: refs/heads/feat/test\n"), 0600); err != nil {
		t.Fatal(err)
	}

	h := &MonitorHandler{}
	branch := h.readGitBranch(dir)
	if branch != "feat/test" {
		t.Errorf("expected branch=feat/test, got %q", branch)
	}
}

func TestMonitorHandler_WriteSummary(t *testing.T) {
	h := &MonitorHandler{}
	var out bytes.Buffer
	h.writeSummary(&out, "my-project", gitStateJSON{Branch: "main"}, plansStateJSON{
		Exists:   true,
		WIPTasks: 2,
		TODOTasks: 3,
	})

	s := out.String()
	if !strings.Contains(s, "my-project") {
		t.Errorf("expected project name in summary, got:\n%s", s)
	}
	if !strings.Contains(s, "main") {
		t.Errorf("expected branch in summary, got:\n%s", s)
	}
	if !strings.Contains(s, "WIP 2") {
		t.Errorf("expected WIP count in summary, got:\n%s", s)
	}
}
