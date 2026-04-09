package hookhandler

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
