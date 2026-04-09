package hookhandler

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestElicitationResultHandler_EmptyInput(t *testing.T) {
	h := &ElicitationResultHandler{}
	var out bytes.Buffer
	if err := h.Handle(strings.NewReader(""), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertElicitationDecision(t, out.String(), "approve", "no payload")
}

func TestElicitationResultHandler_InvalidJSON(t *testing.T) {
	h := &ElicitationResultHandler{}
	var out bytes.Buffer
	if err := h.Handle(strings.NewReader("{bad json"), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertElicitationDecision(t, out.String(), "approve", "no payload")
}

func TestElicitationResultHandler_AlwaysApprove(t *testing.T) {
	dir := t.TempDir()
	h := &ElicitationResultHandler{ProjectRoot: dir}

	payload := `{
		"mcp_server_name": "result-mcp",
		"elicitation_id": "res-001",
		"result_status": "submitted"
	}`
	var out bytes.Buffer
	if err := h.Handle(strings.NewReader(payload), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertElicitationDecision(t, out.String(), "approve", "ElicitationResult tracked")
}

func TestElicitationResultHandler_LogWritten(t *testing.T) {
	dir := t.TempDir()
	h := &ElicitationResultHandler{ProjectRoot: dir}

	payload := `{
		"mcp_server_name": "log-mcp",
		"elicitation_id": "res-log-01",
		"result_status": "cancelled"
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
	content := string(data)
	if !strings.Contains(content, "elicitation_result") {
		t.Errorf("log missing event field: %s", content)
	}
	if !strings.Contains(content, "log-mcp") {
		t.Errorf("log missing mcp_server: %s", content)
	}
	if !strings.Contains(content, "res-log-01") {
		t.Errorf("log missing elicitation_id: %s", content)
	}
	if !strings.Contains(content, "cancelled") {
		t.Errorf("log missing result_status: %s", content)
	}
}

func TestElicitationResultHandler_FallbackFields(t *testing.T) {
	dir := t.TempDir()
	h := &ElicitationResultHandler{ProjectRoot: dir}

	// server_name と status フォールバック
	payload := `{"server_name": "fb-mcp", "id": "fb-res-01", "status": "ok"}`
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
	if !strings.Contains(string(data), "fb-mcp") {
		t.Errorf("log missing fallback mcp_server: %s", data)
	}
}

func TestElicitationResultHandler_SharedLogWithHandler(t *testing.T) {
	// ElicitationHandler と ElicitationResultHandler は同じ JSONL ファイルを共有する
	dir := t.TempDir()

	t.Setenv("HARNESS_BREEZING_SESSION_ID", "")

	h1 := &ElicitationHandler{ProjectRoot: dir}
	h2 := &ElicitationResultHandler{ProjectRoot: dir}

	// ElicitationHandler でログ記録
	p1 := `{"mcp_server_name":"shared-mcp","elicitation_id":"shared-01","message":"q"}`
	var out1 bytes.Buffer
	_ = h1.Handle(strings.NewReader(p1), &out1)

	// ElicitationResultHandler でログ記録
	p2 := `{"mcp_server_name":"shared-mcp","elicitation_id":"shared-01","result_status":"done"}`
	var out2 bytes.Buffer
	_ = h2.Handle(strings.NewReader(p2), &out2)

	logFile := filepath.Join(dir, ".claude", "state", "elicitation-events.jsonl")
	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("log file not created: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 log entries, got %d\n%s", len(lines), string(data))
	}
}
