package session

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
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

func TestMonitorHandler_AdvisorDrift(t *testing.T) {
	// fixedNow is the reference time used in all table cases.
	fixedNow := time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC)

	cases := []struct {
		name        string
		events      string       // raw JSONL content for session.events.jsonl
		ttlSeconds  int          // 0 = use default (600)
		nowOffset   time.Duration // offset from fixedNow for h.now
		wantOutput  bool
		wantSubstr  string
	}{
		{
			name: "miss_responded_within_ttl",
			events: `{"event_type":"advisor-request.v1","request_id":"req-001","timestamp":"2026-04-19T10:00:00Z"}
{"event_type":"advisor-response.v1","request_id":"req-001","timestamp":"2026-04-19T10:05:00Z"}`,
			nowOffset:  6 * time.Minute,
			wantOutput: false,
		},
		{
			name: "hit_unanswered_past_ttl",
			events: `{"event_type":"advisor-request.v1","request_id":"req-002","timestamp":"2026-04-19T08:00:00Z"}`,
			nowOffset:  2 * time.Hour, // 7200s elapsed, TTL=600
			wantOutput: true,
			wantSubstr: "advisor drift: request_id=req-002",
		},
		{
			name: "config_override_low_ttl",
			events: `{"event_type":"advisor-request.v1","request_id":"req-003","timestamp":"2026-04-19T09:59:30Z"}`,
			ttlSeconds: 5,
			nowOffset:  30 * time.Second, // 30s elapsed, TTL=5 → triggers
			wantOutput: true,
			wantSubstr: "advisor drift: request_id=req-003",
		},
		{
			name:       "file_missing_graceful",
			events:     "", // no file will be written
			nowOffset:  0,
			wantOutput: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			stateDir := filepath.Join(dir, "state")
			if err := os.MkdirAll(stateDir, 0700); err != nil {
				t.Fatal(err)
			}

			// Write events file only when events content is provided
			if tc.events != "" {
				eventsFile := filepath.Join(stateDir, "session.events.jsonl")
				if err := os.WriteFile(eventsFile, []byte(tc.events), 0600); err != nil {
					t.Fatal(err)
				}
			}

			// Write config file if custom TTL is specified
			if tc.ttlSeconds > 0 {
				configDir := filepath.Join(dir, "harness")
				if err := os.MkdirAll(configDir, 0755); err != nil {
					t.Fatal(err)
				}
				configContent := "orchestration:\n  advisor_ttl_seconds: " + strconv.Itoa(tc.ttlSeconds) + "\n"
				if err := os.WriteFile(filepath.Join(configDir, ".claude-code-harness.config.yaml"), []byte(configContent), 0644); err != nil {
					t.Fatal(err)
				}
			}

			h := &MonitorHandler{
				now: func() time.Time { return fixedNow.Add(tc.nowOffset) },
			}

			var buf bytes.Buffer
			h.CheckAdvisorDrift(&buf, stateDir, dir)

			got := buf.String()
			if tc.wantOutput {
				if got == "" {
					t.Errorf("expected advisor drift warning but got no output")
					return
				}
				if tc.wantSubstr != "" && !strings.Contains(got, tc.wantSubstr) {
					t.Errorf("expected output to contain %q, got: %q", tc.wantSubstr, got)
				}
			} else {
				if got != "" {
					t.Errorf("expected no output, got: %q", got)
				}
			}
		})
	}
}

func TestMonitorHandler_PlansDrift(t *testing.T) {
	// fixedNow is the reference time used in all table cases.
	fixedNow := time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC)

	cases := []struct {
		name           string
		wipCount       int
		staleHours     int // hours since last modification
		wipThreshold   int // 0 = use default (5)
		staleThreshold int // 0 = use default (24)
		wantOutput     bool
		wantWIP        int
		wantStale      int
	}{
		{
			name:       "miss",
			wipCount:   3,
			staleHours: 2,
			wantOutput: false,
		},
		{
			name:       "wip_hit",
			wipCount:   6,
			staleHours: 0,
			wantOutput: true,
			wantWIP:    6,
			wantStale:  0,
		},
		{
			name:       "stale_hit",
			wipCount:   1,
			staleHours: 30,
			wantOutput: true,
			wantWIP:    1,
			wantStale:  30,
		},
		{
			name:           "config_override",
			wipCount:       2,
			staleHours:     0,
			wipThreshold:   2,
			staleThreshold: 24,
			wantOutput:     true,
			wantWIP:        2,
			wantStale:      0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()

			// Build a Plans.md path (file doesn't need to exist for this test)
			plansFile := filepath.Join(dir, "Plans.md")

			// Compute LastModified from staleHours
			lastMod := fixedNow.Add(-time.Duration(tc.staleHours) * time.Hour).Unix()

			plans := plansStateJSON{
				Exists:       true,
				LastModified: lastMod,
				WIPTasks:     tc.wipCount,
			}

			// If custom thresholds are needed, write a config file
			configPath := filepath.Join(dir, "harness", ".claude-code-harness.config.yaml")
			if tc.wipThreshold > 0 || tc.staleThreshold > 0 {
				if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
					t.Fatal(err)
				}
				wip := tc.wipThreshold
				if wip == 0 {
					wip = 5
				}
				stale := tc.staleThreshold
				if stale == 0 {
					stale = 24
				}
				cfg := strings.Join([]string{
					"monitor:",
					"  plans_drift:",
					"    wip_threshold: " + strconv.Itoa(wip),
					"    stale_hours: " + strconv.Itoa(stale),
				}, "\n") + "\n"
				if err := os.WriteFile(configPath, []byte(cfg), 0644); err != nil {
					t.Fatal(err)
				}
			}

			h := &MonitorHandler{
				now: func() time.Time { return fixedNow },
			}

			var buf bytes.Buffer
			h.CheckPlansDrift(&buf, plans, plansFile, dir)

			got := buf.String()
			if tc.wantOutput {
				if got == "" {
					t.Errorf("expected drift warning but got no output")
					return
				}
				wantLine := "⚠️ plans drift: WIP=" + strconv.Itoa(tc.wantWIP) + ", stale_for=" + strconv.Itoa(tc.wantStale) + "h"
				if !strings.Contains(got, wantLine) {
					t.Errorf("expected output to contain %q, got: %q", wantLine, got)
				}
			} else {
				if got != "" {
					t.Errorf("expected no output, got: %q", got)
				}
			}
		})
	}
}

