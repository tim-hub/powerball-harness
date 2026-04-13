package hookhandler

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandleNotification_EmptyInput(t *testing.T) {
	var out bytes.Buffer
	err := HandleNotification(strings.NewReader(""), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Len() != 0 {
		t.Errorf("expected no output for empty input, got %q", out.String())
	}
}

func TestHandleNotification_InvalidJSON(t *testing.T) {
	var out bytes.Buffer
	err := HandleNotification(strings.NewReader("{invalid json}"), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Len() != 0 {
		t.Errorf("expected no output for invalid JSON, got %q", out.String())
	}
}

func TestHandleNotification_BasicEvent(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	t.Setenv("PROJECT_ROOT", tmpDir)
	t.Setenv("CLAUDE_PLUGIN_DATA", "")

	input := `{"notification_type":"auth_success","session_id":"sess-001","agent_type":"worker"}`
	var out bytes.Buffer
	if err := HandleNotification(strings.NewReader(input), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	logFile := filepath.Join(tmpDir, ".claude", "state", "notification-events.jsonl")
	data, readErr := os.ReadFile(logFile)
	if readErr != nil {
		t.Fatalf("log file not created: %v", readErr)
	}

	var entry notificationLogEntry
	if jsonErr := json.Unmarshal(bytes.TrimSpace(data), &entry); jsonErr != nil {
		t.Fatalf("invalid JSONL entry: %v, raw: %s", jsonErr, string(data))
	}

	if entry.Event != "notification" {
		t.Errorf("expected event=notification, got %q", entry.Event)
	}
	if entry.NotificationType != "auth_success" {
		t.Errorf("expected notification_type=auth_success, got %q", entry.NotificationType)
	}
	if entry.SessionID != "sess-001" {
		t.Errorf("expected session_id=sess-001, got %q", entry.SessionID)
	}
	if entry.AgentType != "worker" {
		t.Errorf("expected agent_type=worker, got %q", entry.AgentType)
	}
	if entry.Timestamp == "" {
		t.Error("expected non-empty timestamp")
	}
}

func TestHandleNotification_TypeFallback(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	t.Setenv("PROJECT_ROOT", tmpDir)
	t.Setenv("CLAUDE_PLUGIN_DATA", "")

	input := `{"type":"idle_prompt","session_id":"sess-002"}`
	var out bytes.Buffer
	if err := HandleNotification(strings.NewReader(input), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	logFile := filepath.Join(tmpDir, ".claude", "state", "notification-events.jsonl")
	data, readErr := os.ReadFile(logFile)
	if readErr != nil {
		t.Fatalf("log file not created: %v", readErr)
	}

	var entry notificationLogEntry
	if jsonErr := json.Unmarshal(bytes.TrimSpace(data), &entry); jsonErr != nil {
		t.Fatalf("invalid JSONL entry: %v", jsonErr)
	}
	if entry.NotificationType != "idle_prompt" {
		t.Errorf("expected notification_type=idle_prompt (from type field), got %q", entry.NotificationType)
	}
}

func TestHandleNotification_MatcherFallback(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	t.Setenv("PROJECT_ROOT", tmpDir)
	t.Setenv("CLAUDE_PLUGIN_DATA", "")

	input := `{"matcher":"permission_prompt","session_id":"sess-003","agent_type":"task-worker"}`
	var out bytes.Buffer
	if err := HandleNotification(strings.NewReader(input), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	logFile := filepath.Join(tmpDir, ".claude", "state", "notification-events.jsonl")
	data, readErr := os.ReadFile(logFile)
	if readErr != nil {
		t.Fatalf("log file not created: %v", readErr)
	}

	var entry notificationLogEntry
	if jsonErr := json.Unmarshal(bytes.TrimSpace(data), &entry); jsonErr != nil {
		t.Fatalf("invalid JSONL entry: %v", jsonErr)
	}
	if entry.NotificationType != "permission_prompt" {
		t.Errorf("expected notification_type=permission_prompt (from matcher), got %q", entry.NotificationType)
	}
}

func TestHandleNotification_RotatesAtLimit(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	t.Setenv("PROJECT_ROOT", tmpDir)
	t.Setenv("CLAUDE_PLUGIN_DATA", "")

	stateDir := filepath.Join(tmpDir, ".claude", "state")
	if err := os.MkdirAll(stateDir, 0o700); err != nil {
		t.Fatal(err)
	}
	logFile := filepath.Join(stateDir, "notification-events.jsonl")
	f, err := os.Create(logFile)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 501; i++ {
		f.WriteString(`{"event":"notification","notification_type":"dummy"}` + "\n")
	}
	f.Close()

	input := `{"notification_type":"auth_success","session_id":"sess-rotate"}`
	var out bytes.Buffer
	if err := HandleNotification(strings.NewReader(input), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, readErr := os.ReadFile(logFile)
	if readErr != nil {
		t.Fatalf("log file not found: %v", readErr)
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) > 401 {
		t.Errorf("expected ≤401 lines after rotation, got %d", len(lines))
	}
}

func TestHandleNotification_NoOutput(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	t.Setenv("PROJECT_ROOT", tmpDir)
	t.Setenv("CLAUDE_PLUGIN_DATA", "")

	input := `{"notification_type":"auth_success","session_id":"sess-out"}`
	var out bytes.Buffer
	if err := HandleNotification(strings.NewReader(input), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Len() != 0 {
		t.Errorf("notification handler should produce no stdout output, got %q", out.String())
	}
}
