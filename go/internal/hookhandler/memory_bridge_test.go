package hookhandler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// --- Existing tests (unchanged) ---

func TestHandleMemoryBridge_EmptyInput(t *testing.T) {
	var out bytes.Buffer
	if err := HandleMemoryBridge(strings.NewReader(""), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertApprove(t, out.String())
}

func TestHandleMemoryBridge_InvalidJSON(t *testing.T) {
	var out bytes.Buffer
	if err := HandleMemoryBridge(strings.NewReader("not json"), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertApprove(t, out.String())
}

func TestHandleMemoryBridge_UnknownTarget(t *testing.T) {
	var out bytes.Buffer
	payload := `{"hook_event_name":"unknown-event","session_id":"s1","cwd":"/tmp"}`
	if err := HandleMemoryBridge(strings.NewReader(payload), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertApprove(t, out.String())
}

func TestHandleMemoryBridge_ValidTargets(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HARNESS_PROJECT_ROOT", dir)

	targets := []string{"session-start", "user-prompt", "post-tool-use", "stop", "codex-notify"}
	for _, target := range targets {
		t.Run(target, func(t *testing.T) {
			var out bytes.Buffer
			payload := `{"hook_event_name":"` + target + `","session_id":"s1","cwd":"` + dir + `"}`
			if err := HandleMemoryBridge(strings.NewReader(payload), &out); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			assertApprove(t, out.String())

			// Event log must contain the dispatched event.
			logPath := filepath.Join(dir, ".claude", "state", "memory-bridge-events.jsonl")
			logData, err := os.ReadFile(logPath)
			if err != nil {
				t.Fatalf("event log not created: %v", err)
			}
			if !strings.Contains(string(logData), target) {
				t.Errorf("event log does not contain target %q: %s", target, string(logData))
			}
		})
	}
}

func TestHandleMemoryBridge_LogEntry_Format(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HARNESS_PROJECT_ROOT", dir)

	var out bytes.Buffer
	payload := `{"hook_event_name":"session-start","session_id":"sess-abc","cwd":"` + dir + `"}`
	if err := HandleMemoryBridge(strings.NewReader(payload), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	logPath := filepath.Join(dir, ".claude", "state", "memory-bridge-events.jsonl")
	logData, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("event log not created: %v", err)
	}

	var entry memoryBridgeEvent
	if err := json.Unmarshal(bytes.TrimSpace(logData), &entry); err != nil {
		t.Fatalf("log entry is not valid JSON: %v\n%s", err, logData)
	}
	if entry.Event != "session-start" {
		t.Errorf("entry.Event = %q, want session-start", entry.Event)
	}
	if entry.SessionID != "sess-abc" {
		t.Errorf("entry.SessionID = %q, want sess-abc", entry.SessionID)
	}
	if entry.Timestamp == "" {
		t.Error("entry.Timestamp is empty")
	}
}

// TestHandleMemoryBridge_PascalCaseNormalization verifies that Claude Code's
// PascalCase hook_event_name values (e.g. "SessionStart") are correctly mapped
// to their kebab-case internal targets and dispatched properly.
func TestHandleMemoryBridge_PascalCaseNormalization(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HARNESS_PROJECT_ROOT", dir)

	cases := []struct {
		ccEventName    string // CC sends this PascalCase name
		wantLogTarget  string // kebab-case target expected in the JSONL log
	}{
		{"SessionStart", "session-start"},
		{"UserPromptSubmit", "user-prompt"},
		{"PostToolUse", "post-tool-use"},
		{"Stop", "stop"},
	}

	for _, tc := range cases {
		t.Run(tc.ccEventName, func(t *testing.T) {
			var out bytes.Buffer
			payload := `{"hook_event_name":"` + tc.ccEventName + `","session_id":"s1","cwd":"` + dir + `"}`
			if err := HandleMemoryBridge(strings.NewReader(payload), &out); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			assertApprove(t, out.String())

			// For "stop" there is no JSONL log entry (finalize-only path skips log for stop)
			if tc.ccEventName == "Stop" {
				return
			}
			logPath := filepath.Join(dir, ".claude", "state", "memory-bridge-events.jsonl")
			logData, err := os.ReadFile(logPath)
			if err != nil {
				t.Fatalf("event log not created: %v", err)
			}
			if !strings.Contains(string(logData), tc.wantLogTarget) {
				t.Errorf("event log does not contain %q: %s", tc.wantLogTarget, string(logData))
			}
		})
	}
}

// --- New tests for harness-mem HTTP integration ---

func TestMemoryBridgeClient_PostEvents(t *testing.T) {
	var mu sync.Mutex
	received := make(map[string][]byte) // path -> body

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		mu.Lock()
		received[r.URL.Path] = body
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	dir := t.TempDir()
	t.Setenv("HARNESS_PROJECT_ROOT", dir)

	tests := []struct {
		target    string
		wantPath  string
		wantType  string // event_type in the request body
		isFinalize bool
	}{
		{"session-start", "/v1/events/record", "session_start", false},
		{"user-prompt", "/v1/events/record", "user_prompt", false},
		{"post-tool-use", "/v1/events/record", "tool_use", false},
		{"codex-notify", "/v1/events/record", "checkpoint", false},
		{"stop", "/v1/sessions/finalize", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.target, func(t *testing.T) {
			mu.Lock()
			received = make(map[string][]byte) // reset
			mu.Unlock()

			c := &MemoryBridgeClient{
				HTTPClient: server.Client(),
				BaseURL:    server.URL,
			}
			var out bytes.Buffer
			payload := `{"hook_event_name":"` + tt.target + `","session_id":"sess-42","cwd":"` + dir + `/myproject"}`
			if err := c.Handle(strings.NewReader(payload), &out); err != nil {
				t.Fatalf("Handle error: %v", err)
			}
			assertApprove(t, out.String())

			mu.Lock()
			body, ok := received[tt.wantPath]
			mu.Unlock()

			if !ok {
				t.Fatalf("no POST received at %s; received paths: %v", tt.wantPath, keys(received))
			}

			if tt.isFinalize {
				var req harnessMemFinalizeRequest
				if err := json.Unmarshal(body, &req); err != nil {
					t.Fatalf("finalize body parse error: %v\nbody: %s", err, body)
				}
				if req.SessionID != "sess-42" {
					t.Errorf("SessionID = %q, want sess-42", req.SessionID)
				}
				if req.Platform != "claude" {
					t.Errorf("Platform = %q, want claude", req.Platform)
				}
				if req.Project != "myproject" {
					t.Errorf("Project = %q, want myproject", req.Project)
				}
			} else {
				var req harnessMemRecordRequest
				if err := json.Unmarshal(body, &req); err != nil {
					t.Fatalf("record body parse error: %v\nbody: %s", err, body)
				}
				if req.Event.EventType != tt.wantType {
					t.Errorf("EventType = %q, want %q", req.Event.EventType, tt.wantType)
				}
				if req.Event.SessionID != "sess-42" {
					t.Errorf("SessionID = %q, want sess-42", req.Event.SessionID)
				}
				if req.Event.Platform != "claude" {
					t.Errorf("Platform = %q, want claude", req.Event.Platform)
				}
				if req.Event.Project != "myproject" {
					t.Errorf("Project = %q, want myproject", req.Event.Project)
				}
				if req.Event.TS == "" {
					t.Error("TS is empty")
				}
			}
		})
	}
}

func TestMemoryBridgeClient_StopUsesFinalize(t *testing.T) {
	var receivedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	dir := t.TempDir()
	t.Setenv("HARNESS_PROJECT_ROOT", dir)

	c := &MemoryBridgeClient{
		HTTPClient: server.Client(),
		BaseURL:    server.URL,
	}
	var out bytes.Buffer
	payload := `{"hook_event_name":"stop","session_id":"sess-99","cwd":"` + dir + `"}`
	if err := c.Handle(strings.NewReader(payload), &out); err != nil {
		t.Fatalf("Handle error: %v", err)
	}

	if receivedPath != "/v1/sessions/finalize" {
		t.Errorf("stop target routed to %q, want /v1/sessions/finalize", receivedPath)
	}
}

func TestMemoryBridgeClient_ServerDown_NoError(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HARNESS_PROJECT_ROOT", dir)

	// Use a client with very short timeout pointing at a closed port.
	c := &MemoryBridgeClient{
		HTTPClient: &http.Client{Timeout: 100 * time.Millisecond},
		BaseURL:    "http://127.0.0.1:1", // port 1: connection refused
	}
	var out bytes.Buffer
	payload := `{"hook_event_name":"user-prompt","session_id":"s1","cwd":"` + dir + `"}`
	if err := c.Handle(strings.NewReader(payload), &out); err != nil {
		t.Fatalf("Handle should not error when harness-mem is down: %v", err)
	}
	assertApprove(t, out.String())

	// JSONL log must still be written even when harness-mem is unreachable.
	logPath := filepath.Join(dir, ".claude", "state", "memory-bridge-events.jsonl")
	if _, err := os.Stat(logPath); err != nil {
		t.Fatal("JSONL log should still be written when harness-mem is down")
	}
}

func TestMemoryBridgeClient_BearerToken(t *testing.T) {
	var receivedAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	dir := t.TempDir()
	t.Setenv("HARNESS_PROJECT_ROOT", dir)
	t.Setenv("HARNESS_MEM_ADMIN_TOKEN", "secret-token-xyz")

	c := &MemoryBridgeClient{
		HTTPClient: server.Client(),
		BaseURL:    server.URL,
	}
	var out bytes.Buffer
	payload := `{"hook_event_name":"session-start","session_id":"s1","cwd":"` + dir + `"}`
	if err := c.Handle(strings.NewReader(payload), &out); err != nil {
		t.Fatalf("Handle error: %v", err)
	}

	if receivedAuth != "Bearer secret-token-xyz" {
		t.Errorf("Authorization = %q, want %q", receivedAuth, "Bearer secret-token-xyz")
	}
}

func TestValidateBridgeInput(t *testing.T) {
	tests := []struct {
		name    string
		input   memoryBridgeInput
		wantErr bool
	}{
		{"valid", memoryBridgeInput{SessionID: "s1", CWD: "/tmp", HookEventName: "session-start"}, false},
		{"empty session_id", memoryBridgeInput{SessionID: "", CWD: "/tmp", HookEventName: "session-start"}, true},
		{"empty cwd", memoryBridgeInput{SessionID: "s1", CWD: "", HookEventName: "session-start"}, true},
		{"session_id too long", memoryBridgeInput{SessionID: strings.Repeat("x", 257), CWD: "/tmp", HookEventName: "session-start"}, true},
		{"session_id at limit", memoryBridgeInput{SessionID: strings.Repeat("x", 256), CWD: "/tmp", HookEventName: "session-start"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateBridgeInput(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateBridgeInput() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMemoryBridgeClient_ValidationBlocksPost(t *testing.T) {
	postReceived := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postReceived = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	dir := t.TempDir()
	t.Setenv("HARNESS_PROJECT_ROOT", dir)

	c := &MemoryBridgeClient{HTTPClient: server.Client(), BaseURL: server.URL}
	var out bytes.Buffer
	// Empty session_id should fail validation and NOT post to server.
	payload := `{"hook_event_name":"session-start","session_id":"","cwd":"` + dir + `"}`
	if err := c.Handle(strings.NewReader(payload), &out); err != nil {
		t.Fatalf("Handle error: %v", err)
	}
	assertApprove(t, out.String())
	if postReceived {
		t.Error("validation should have blocked POST to harness-mem, but server received a request")
	}
}

// --- helpers ---

// assertApprove verifies the output is a valid JSON approve response.
func assertApprove(t *testing.T, output string) {
	t.Helper()
	output = strings.TrimSpace(output)
	if output == "" {
		t.Fatal("expected approve JSON, got empty output")
	}
	var resp map[string]string
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, output)
	}
	if resp["decision"] != "approve" {
		t.Errorf("decision = %q, want approve", resp["decision"])
	}
}

// keys returns the keys of a map for error messages.
func keys(m map[string][]byte) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	return ks
}
