package hookhandler

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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

	ledgerFile := filepath.Join(dir, ".claude", "state", "elicitation", "events.jsonl")
	ledgerData, err := os.ReadFile(ledgerFile)
	if err != nil {
		t.Fatalf("ledger file not created: %v", err)
	}
	var event ElicitationEvent
	if err := json.Unmarshal(bytes.TrimSpace(ledgerData), &event); err != nil {
		t.Fatalf("ledger entry is not valid JSON: %v\n%s", err, ledgerData)
	}
	if event.SchemaVersion != "elicitation-event.v1" {
		t.Errorf("schema_version = %q, want elicitation-event.v1", event.SchemaVersion)
	}
	if event.EventKind != "capability_probe" {
		t.Errorf("event_kind = %q, want capability_probe", event.EventKind)
	}
	if event.RunID == "" {
		t.Error("run_id is empty")
	}
	if len(event.PrivacyTags) != 1 || event.PrivacyTags[0] != "do_not_train" {
		t.Errorf("privacy_tags = %v, want [do_not_train]", event.PrivacyTags)
	}
}

func TestElicitationHandler_FallbackFields(t *testing.T) {
	t.Setenv("HARNESS_BREEZING_SESSION_ID", "")

	dir := t.TempDir()
	h := &ElicitationHandler{ProjectRoot: dir}

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

func TestElicitationHandler_LedgerAppendOnly(t *testing.T) {
	t.Setenv("HARNESS_BREEZING_SESSION_ID", "")

	dir := t.TempDir()
	ledgerDir := filepath.Join(dir, ".claude", "state", "elicitation")
	if err := os.MkdirAll(ledgerDir, 0o700); err != nil {
		t.Fatal(err)
	}
	ledgerFile := filepath.Join(ledgerDir, "events.jsonl")
	firstLine := `{"schema_version":"elicitation-event.v1","event_kind":"weak_label","run_id":"preexisting","privacy_tags":["do_not_train"],"evidence_refs":[],"source":"test","timestamp":"2026-05-06T00:00:00Z"}`
	if err := os.WriteFile(ledgerFile, []byte(firstLine+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	h := &ElicitationHandler{ProjectRoot: dir}
	payload := `{"mcp_server_name":"append-mcp","elicitation_id":"append-01","message":"append"}`
	var out bytes.Buffer
	if err := h.Handle(strings.NewReader(payload), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(ledgerFile)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 ledger entries, got %d\n%s", len(lines), string(data))
	}
	if lines[0] != firstLine {
		t.Fatalf("first line was modified; append-only contract broken\nwant: %s\ngot:  %s", firstLine, lines[0])
	}
}

func TestElicitationHandler_PrivacyTagValidationFallsBackToDecision(t *testing.T) {
	t.Setenv("HARNESS_ELICITATION_PRIVACY_TAGS", "unknown_tag")
	t.Setenv("HARNESS_BREEZING_SESSION_ID", "")

	dir := t.TempDir()
	h := &ElicitationHandler{ProjectRoot: dir}
	var out bytes.Buffer
	if err := h.Handle(strings.NewReader(`{"mcp_server_name":"m","elicitation_id":"e"}`), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertElicitationDecision(t, out.String(), "approve", "forwarding to user")

	ledgerFile := filepath.Join(dir, ".claude", "state", "elicitation", "events.jsonl")
	if _, err := os.Stat(ledgerFile); !os.IsNotExist(err) {
		t.Fatalf("invalid privacy tag should not create ledger entry, stat err=%v", err)
	}
}

func TestAppendElicitationEvent_ValidatesSchemaFields(t *testing.T) {
	dir := t.TempDir()
	score := 1.2
	tests := []struct {
		name  string
		event ElicitationEvent
	}{
		{
			name: "bad event kind",
			event: ElicitationEvent{
				EventKind:    "not-a-kind",
				RunID:        "run",
				PrivacyTags:  []string{"do_not_train"},
				EvidenceRefs: []string{},
				Source:       "test",
			},
		},
		{
			name: "bad verdict",
			event: ElicitationEvent{
				EventKind:    "eval_result",
				RunID:        "run",
				Verdict:      "MAYBE",
				PrivacyTags:  []string{"do_not_train"},
				EvidenceRefs: []string{},
				Source:       "test",
			},
		},
		{
			name: "bad reward score",
			event: ElicitationEvent{
				EventKind:    "eval_result",
				RunID:        "run",
				RewardScore:  &score,
				PrivacyTags:  []string{"do_not_train"},
				EvidenceRefs: []string{},
				Source:       "test",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := appendElicitationEvent(dir, tt.event); err == nil {
				t.Fatal("appendElicitationEvent should reject invalid schema field")
			}
		})
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
