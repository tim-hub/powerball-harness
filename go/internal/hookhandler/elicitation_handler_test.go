package hookhandler

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// assertElicitationDecision は出力 JSON の decision と reason を検証するヘルパー。
func assertElicitationDecision(t *testing.T, output, wantDecision, wantReasonContains string) {
	t.Helper()
	output = strings.TrimSpace(output)
	if output == "" {
		t.Fatal("expected JSON output, got empty")
	}
	var resp map[string]string
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatalf("invalid JSON output: %v\noutput: %s", err, output)
	}
	if resp["decision"] != wantDecision {
		t.Errorf("decision = %q, want %q", resp["decision"], wantDecision)
	}
	if wantReasonContains != "" && !strings.Contains(resp["reason"], wantReasonContains) {
		t.Errorf("reason = %q, want to contain %q", resp["reason"], wantReasonContains)
	}
}

func TestElicitationHandler_EmptyInput(t *testing.T) {
	h := &ElicitationHandler{}
	var out bytes.Buffer
	if err := h.Handle(strings.NewReader(""), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertElicitationDecision(t, out.String(), "approve", "no payload")
}

func TestElicitationHandler_InvalidJSON(t *testing.T) {
	h := &ElicitationHandler{}
	var out bytes.Buffer
	if err := h.Handle(strings.NewReader("not json"), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertElicitationDecision(t, out.String(), "approve", "no payload")
}

func TestElicitationHandler_NormalSession_Approve(t *testing.T) {
	// HARNESS_BREEZING_SESSION_ID が未設定 → 通常セッション → approve
	t.Setenv("HARNESS_BREEZING_SESSION_ID", "")

	dir := t.TempDir()
	h := &ElicitationHandler{ProjectRoot: dir}

	payload := `{
		"mcp_server_name": "my-mcp",
		"elicitation_id": "elic-001",
		"message": "Which repo?"
	}`
	var out bytes.Buffer
	if err := h.Handle(strings.NewReader(payload), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertElicitationDecision(t, out.String(), "approve", "forwarding to user")
}

func TestElicitationHandler_BreezingSession_Deny(t *testing.T) {
	// HARNESS_BREEZING_SESSION_ID が設定されている → Breezing → deny
	t.Setenv("HARNESS_BREEZING_SESSION_ID", "session-breezing-42")

	dir := t.TempDir()
	h := &ElicitationHandler{ProjectRoot: dir}

	payload := `{
		"mcp_server_name": "some-mcp",
		"elicitation_id": "elic-002",
		"message": "Enter value"
	}`
	var out bytes.Buffer
	if err := h.Handle(strings.NewReader(payload), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertElicitationDecision(t, out.String(), "deny", "Breezing session")
}

func TestElicitationHandler_LogWritten(t *testing.T) {
	t.Setenv("HARNESS_BREEZING_SESSION_ID", "")

	dir := t.TempDir()
	h := &ElicitationHandler{ProjectRoot: dir}

	payload := `{
		"mcp_server_name": "log-mcp",
		"elicitation_id": "elic-log-01",
		"message": "test message"
	}`
	var out bytes.Buffer
	if err := h.Handle(strings.NewReader(payload), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	logFile := filepath.Join(dir, ".claude", "state", "elicitation-events.jsonl")
	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("log file not created: %v", err)
	}
	logContent := string(data)
	if !strings.Contains(logContent, "elicitation") {
		t.Errorf("log does not contain 'elicitation': %s", logContent)
	}
	if !strings.Contains(logContent, "log-mcp") {
		t.Errorf("log does not contain mcp_server: %s", logContent)
	}
	if !strings.Contains(logContent, "elic-log-01") {
		t.Errorf("log does not contain elicitation_id: %s", logContent)
	}
}

func TestElicitationHandler_FallbackFields(t *testing.T) {
	// server_name と id フォールバックのテスト
	t.Setenv("HARNESS_BREEZING_SESSION_ID", "")

	dir := t.TempDir()
	h := &ElicitationHandler{ProjectRoot: dir}

	// mcp_server_name なし → server_name を使う
	payload := `{"server_name": "fallback-mcp", "id": "fb-001", "message": "hi"}`
	var out bytes.Buffer
	if err := h.Handle(strings.NewReader(payload), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertElicitationDecision(t, out.String(), "approve", "")

	logFile := filepath.Join(dir, ".claude", "state", "elicitation-events.jsonl")
	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("log file not created: %v", err)
	}
	if !strings.Contains(string(data), "fallback-mcp") {
		t.Errorf("log does not contain 'fallback-mcp': %s", data)
	}
}

func TestFirstNonEmpty(t *testing.T) {
	tests := []struct {
		vals []string
		want string
	}{
		{[]string{"", "", "c"}, "c"},
		{[]string{"a", "b", "c"}, "a"},
		{[]string{"", "b", ""}, "b"},
		{[]string{"", "", ""}, ""},
		{[]string{}, ""},
	}
	for _, tt := range tests {
		got := firstNonEmpty(tt.vals...)
		if got != tt.want {
			t.Errorf("firstNonEmpty(%v) = %q, want %q", tt.vals, got, tt.want)
		}
	}
}
