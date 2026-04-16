package guardrail

import (
	"testing"

	"github.com/tim-hub/powerball-harness/go/pkg/hookproto"
)

// BenchmarkEvaluatePreTool measures pre-tool check latency on a representative Bash command.
// This is a safe, typical command that will pass all guardrail rules.
func BenchmarkEvaluatePreTool(b *testing.B) {
	input := hookproto.HookInput{
		ToolName: "Bash",
		ToolInput: map[string]interface{}{
			"command": "go test ./...",
		},
		SessionID: "bench-session-001",
		CWD:       "/tmp/bench-project",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EvaluatePreTool(input)
	}
}

// BenchmarkEvaluatePostToolNonTest measures post-tool latency on a regular Go source file.
// Non-test files skip the tampering pattern scan (T01–T12), so this measures only security scanning.
func BenchmarkEvaluatePostToolNonTest(b *testing.B) {
	input := hookproto.HookInput{
		ToolName: "Write",
		ToolInput: map[string]interface{}{
			"file_path": "cmd/server/main.go",
			"content": `package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, World!")
	})
	http.ListenAndServe(":8080", nil)
}
`,
		},
		SessionID: "bench-session-001",
		CWD:       "/tmp/bench-project",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EvaluatePostTool(input)
	}
}

// BenchmarkEvaluatePostToolTestFile measures post-tool latency on a test file.
// Test files trigger the tampering pattern scan (T01–T12) in addition to security scanning.
func BenchmarkEvaluatePostToolTestFile(b *testing.B) {
	input := hookproto.HookInput{
		ToolName: "Write",
		ToolInput: map[string]interface{}{
			"file_path": "internal/handler/handler_test.go",
			"content": `package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	HandleRequest(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestHandleError(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/invalid", nil)
	w := httptest.NewRecorder()
	HandleRequest(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}
`,
		},
		SessionID: "bench-session-001",
		CWD:       "/tmp/bench-project",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EvaluatePostTool(input)
	}
}
