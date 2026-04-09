package hookhandler

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStopFailureHandler_EmptyInput(t *testing.T) {
	h := &StopFailureHandler{}
	var out bytes.Buffer
	if err := h.Handle(strings.NewReader(""), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 空入力の場合は何も出力しない
	if out.Len() != 0 {
		t.Errorf("expected empty output, got: %s", out.String())
	}
}

func TestStopFailureHandler_LogsEntry(t *testing.T) {
	dir := t.TempDir()
	h := &StopFailureHandler{ProjectRoot: dir}

	payload := `{
		"error": {"message": "service unavailable", "status": "503"},
		"session_id": "sess-abc"
	}`
	var out bytes.Buffer
	if err := h.Handle(strings.NewReader(payload), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	logFile := filepath.Join(dir, ".claude", "state", "stop-failures.jsonl")
	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("log file not created: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "stop_failure") {
		t.Errorf("log missing event: %s", content)
	}
	if !strings.Contains(content, "sess-abc") {
		t.Errorf("log missing session_id: %s", content)
	}
}

func TestStopFailureHandler_RateLimit429_SystemMessage(t *testing.T) {
	dir := t.TempDir()
	h := &StopFailureHandler{ProjectRoot: dir}

	payload := `{
		"error": {"message": "rate limit exceeded", "status": "429"},
		"session_id": "worker-rate-01"
	}`
	var out bytes.Buffer
	if err := h.Handle(strings.NewReader(payload), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 429 の場合は systemMessage が出力される
	output := strings.TrimSpace(out.String())
	if output == "" {
		t.Fatal("expected systemMessage output for 429, got empty")
	}
	var resp map[string]string
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, output)
	}
	if !strings.Contains(resp["systemMessage"], "worker-rate-01") {
		t.Errorf("systemMessage does not contain session_id: %s", resp["systemMessage"])
	}
	if !strings.Contains(resp["systemMessage"], "429") || !strings.Contains(resp["systemMessage"], "Breezing") {
		t.Errorf("systemMessage does not mention rate limit / Breezing: %s", resp["systemMessage"])
	}
}

func TestStopFailureHandler_NonRateLimit_NoSystemMessage(t *testing.T) {
	dir := t.TempDir()
	h := &StopFailureHandler{ProjectRoot: dir}

	payload := `{
		"error": {"message": "internal server error", "status": "500"},
		"session_id": "worker-500"
	}`
	var out bytes.Buffer
	if err := h.Handle(strings.NewReader(payload), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 429 以外は systemMessage を出力しない
	output := strings.TrimSpace(out.String())
	if output != "" {
		t.Errorf("expected no output for non-429, got: %s", output)
	}
}

func TestStopFailureHandler_StringError(t *testing.T) {
	// error フィールドが文字列の場合のテスト
	dir := t.TempDir()
	h := &StopFailureHandler{ProjectRoot: dir}

	payload := `{"error": "rate limit exceeded", "session_id": "str-sess"}`
	var out bytes.Buffer
	if err := h.Handle(strings.NewReader(payload), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	logFile := filepath.Join(dir, ".claude", "state", "stop-failures.jsonl")
	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("log file not created: %v", err)
	}
	if !strings.Contains(string(data), "rate limit exceeded") {
		t.Errorf("log missing raw error message: %s", data)
	}
}

func TestClassifyErrorCode(t *testing.T) {
	tests := []struct {
		rawCode string
		msg     string
		want    string
	}{
		{"429", "rate limit", "429"},
		{"401", "unauthorized", "auth_error"},
		{"403", "forbidden", "auth_error"},
		{"", "rate limit exceeded", "rate_limit"},
		{"", "auth failure", "auth_error"},
		{"", "network error", "network_error"},
		{"", "connection refused", "network_error"},
		{"", "request timeout", "network_error"},
		{"", "something else", "unknown"},
		{"503", "service unavailable", "503"},
	}
	for _, tt := range tests {
		got := classifyErrorCode(tt.rawCode, tt.msg)
		if got != tt.want {
			t.Errorf("classifyErrorCode(%q, %q) = %q, want %q", tt.rawCode, tt.msg, got, tt.want)
		}
	}
}

func TestIsStopFailureLogSymlink(t *testing.T) {
	dir := t.TempDir()
	realFile := filepath.Join(dir, "real.txt")
	if err := os.WriteFile(realFile, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	// 実ファイルは false
	if isStopFailureLogSymlink(realFile) {
		t.Error("real file reported as symlink")
	}

	// 存在しないファイルは false
	if isStopFailureLogSymlink(filepath.Join(dir, "noexist")) {
		t.Error("nonexistent file reported as symlink")
	}

	// シンボリックリンクは true
	linkFile := filepath.Join(dir, "link.txt")
	if err := os.Symlink(realFile, linkFile); err != nil {
		t.Skip("symlink creation not supported:", err)
	}
	if !isStopFailureLogSymlink(linkFile) {
		t.Error("symlink not detected")
	}
}

func TestStopFailureHandler_Idempotent(t *testing.T) {
	// 複数回呼び出しで JSONL に複数行が追記されることを確認
	dir := t.TempDir()
	h := &StopFailureHandler{ProjectRoot: dir}

	payload := `{"error": {"message": "err", "status": "500"}, "session_id": "s1"}`
	for i := 0; i < 3; i++ {
		var out bytes.Buffer
		if err := h.Handle(strings.NewReader(payload), &out); err != nil {
			t.Fatalf("call %d: unexpected error: %v", i+1, err)
		}
	}

	logFile := filepath.Join(dir, ".claude", "state", "stop-failures.jsonl")
	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("log file not created: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 log lines, got %d\n%s", len(lines), string(data))
	}
}
