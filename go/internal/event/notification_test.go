package event

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNotificationHandler_EmptyInput(t *testing.T) {
	dir := t.TempDir()
	h := &NotificationHandler{StateDir: dir}

	var buf bytes.Buffer
	if err := h.Handle(strings.NewReader(""), &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Notification は stdout に何も出さない
	if buf.Len() != 0 {
		t.Errorf("expected empty output, got: %s", buf.String())
	}
}

func TestNotificationHandler_LogsEvent(t *testing.T) {
	dir := t.TempDir()
	h := &NotificationHandler{StateDir: dir}

	input := `{
		"notification_type": "permission_prompt",
		"session_id": "sess-abc",
		"agent_type": "reviewer"
	}`
	var buf bytes.Buffer
	if err := h.Handle(strings.NewReader(input), &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// ログファイルを確認
	logFile := filepath.Join(dir, "notification-events.jsonl")
	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("expected log file, got error: %v", err)
	}

	var entry notificationLogEntry
	if err := json.Unmarshal(bytes.TrimSpace(data), &entry); err != nil {
		t.Fatalf("invalid log entry: %v\n%s", err, string(data))
	}
	if entry.Event != "notification" {
		t.Errorf("expected event=notification, got: %s", entry.Event)
	}
	if entry.NotificationType != "permission_prompt" {
		t.Errorf("expected notification_type=permission_prompt, got: %s", entry.NotificationType)
	}
	if entry.SessionID != "sess-abc" {
		t.Errorf("expected session_id=sess-abc, got: %s", entry.SessionID)
	}
	if entry.AgentType != "reviewer" {
		t.Errorf("expected agent_type=reviewer, got: %s", entry.AgentType)
	}
}

func TestNotificationHandler_TypeFallback(t *testing.T) {
	dir := t.TempDir()
	h := &NotificationHandler{StateDir: dir}

	// notification_type ではなく type フィールドを使用
	input := `{"type": "idle_prompt", "session_id": "s1"}`
	var buf bytes.Buffer
	if err := h.Handle(strings.NewReader(input), &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	logFile := filepath.Join(dir, "notification-events.jsonl")
	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("expected log file, got error: %v", err)
	}

	var entry notificationLogEntry
	_ = json.Unmarshal(bytes.TrimSpace(data), &entry)
	if entry.NotificationType != "idle_prompt" {
		t.Errorf("expected notification_type=idle_prompt from 'type' field, got: %s", entry.NotificationType)
	}
}

func TestNotificationHandler_MatcherFallback(t *testing.T) {
	dir := t.TempDir()
	h := &NotificationHandler{StateDir: dir}

	// matcher フィールドを使用
	input := `{"matcher": "auth_success", "session_id": "s2"}`
	var buf bytes.Buffer
	if err := h.Handle(strings.NewReader(input), &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	logFile := filepath.Join(dir, "notification-events.jsonl")
	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("expected log file, got error: %v", err)
	}

	var entry notificationLogEntry
	_ = json.Unmarshal(bytes.TrimSpace(data), &entry)
	if entry.NotificationType != "auth_success" {
		t.Errorf("expected notification_type=auth_success from 'matcher' field, got: %s", entry.NotificationType)
	}
}

func TestNotificationHandler_NoOutput(t *testing.T) {
	dir := t.TempDir()
	h := &NotificationHandler{StateDir: dir}

	input := `{"notification_type": "idle_prompt", "session_id": "s3"}`
	var buf bytes.Buffer
	if err := h.Handle(strings.NewReader(input), &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Notification フックは stdout に何も出さない
	if buf.Len() != 0 {
		t.Errorf("expected no stdout output, got: %s", buf.String())
	}
}

func TestNotificationHandler_LogRotation(t *testing.T) {
	dir := t.TempDir()
	h := &NotificationHandler{StateDir: dir}

	logFile := filepath.Join(dir, "notification-events.jsonl")

	// 510 行のダミーログを作成（ローテーション閾値 500 行を超える）
	var existing strings.Builder
	for i := 0; i < 510; i++ {
		existing.WriteString(`{"event":"notification","notification_type":"old","session_id":"x","agent_type":"","timestamp":"2026-01-01T00:00:00Z"}`)
		existing.WriteString("\n")
	}
	if err := os.WriteFile(logFile, []byte(existing.String()), 0600); err != nil {
		t.Fatal(err)
	}

	input := `{"notification_type": "new_event", "session_id": "s4"}`
	var buf bytes.Buffer
	if err := h.Handle(strings.NewReader(input), &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	// ローテーション後は 401 行以下（400 + 新規 1）
	if len(lines) > 401 {
		t.Errorf("expected rotation to ~401 lines, got %d", len(lines))
	}
}
