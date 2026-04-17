package hook

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// TestRunPostToolBatch_Smoke verifies that RunPostToolBatch completes without
// error when given a minimal PostToolUse payload.
func TestRunPostToolBatch_Smoke(t *testing.T) {
	payload := `{"tool_name":"Write","tool_input":{"file_path":"/tmp/test.go"},"tool_response":{},"cwd":"/tmp","hook_event_name":"PostToolUse"}`

	var out bytes.Buffer
	if err := RunPostToolBatch(strings.NewReader(payload), &out); err != nil {
		t.Fatalf("RunPostToolBatch returned error: %v", err)
	}

	// Must produce valid JSON output.
	outStr := strings.TrimSpace(out.String())
	if outStr == "" {
		t.Fatal("expected non-empty output, got empty")
	}
	var resp map[string]interface{}
	if err := json.Unmarshal([]byte(outStr), &resp); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, outStr)
	}
}

// TestRunPostToolBatch_EmptyInput verifies that an empty stdin is handled
// gracefully without panicking or returning a hard error.
func TestRunPostToolBatch_EmptyInput(t *testing.T) {
	var out bytes.Buffer
	if err := RunPostToolBatch(strings.NewReader(""), &out); err != nil {
		t.Fatalf("RunPostToolBatch should not error on empty input: %v", err)
	}
	// Should still produce valid JSON (approve fallback).
	outStr := strings.TrimSpace(out.String())
	if outStr == "" {
		t.Fatal("expected non-empty output, got empty")
	}
	var resp map[string]interface{}
	if err := json.Unmarshal([]byte(outStr), &resp); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, outStr)
	}
}

// TestMergePostBatchOutputs_AllEmpty verifies that when all hooks produce no
// output, the merge function returns an approve response.
func TestMergePostBatchOutputs_AllEmpty(t *testing.T) {
	results := []BatchResult{
		{Name: "hook-a", Output: nil, Err: nil},
		{Name: "hook-b", Output: []byte(""), Err: nil},
		{Name: "hook-c", Output: []byte("   "), Err: nil},
	}

	var out bytes.Buffer
	if err := mergePostBatchOutputs(results, &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp map[string]string
	if err := json.Unmarshal([]byte(strings.TrimSpace(out.String())), &resp); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if resp["decision"] != "approve" {
		t.Errorf("decision = %q, want approve", resp["decision"])
	}
}

// TestMergePostBatchOutputs_FirstOutputWins verifies that the first non-empty
// valid JSON output is used as the final response.
func TestMergePostBatchOutputs_FirstOutputWins(t *testing.T) {
	firstJSON := `{"decision":"approve","reason":"hook-a output"}`
	results := []BatchResult{
		{Name: "hook-a", Output: []byte(firstJSON), Err: nil},
		{Name: "hook-b", Output: []byte(`{"decision":"approve","reason":"hook-b output"}`), Err: nil},
	}

	var out bytes.Buffer
	if err := mergePostBatchOutputs(results, &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	outStr := strings.TrimSpace(out.String())
	var resp map[string]string
	if err := json.Unmarshal([]byte(outStr), &resp); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if resp["reason"] != "hook-a output" {
		t.Errorf("expected first hook output to win, got reason=%q", resp["reason"])
	}
}

// TestMergePostBatchOutputs_OneErrorOthersRun verifies that when one hook
// errors, the others still run and their outputs are considered.
func TestMergePostBatchOutputs_OneErrorOthersRun(t *testing.T) {
	successJSON := `{"decision":"approve","reason":"hook-b success"}`
	results := []BatchResult{
		{Name: "hook-a", Output: nil, Err: errFake("hook-a failed")},
		{Name: "hook-b", Output: []byte(successJSON), Err: nil},
		{Name: "hook-c", Output: nil, Err: nil},
	}

	var out bytes.Buffer
	// Should NOT return an error even though hook-a errored.
	if err := mergePostBatchOutputs(results, &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	outStr := strings.TrimSpace(out.String())
	var resp map[string]string
	if err := json.Unmarshal([]byte(outStr), &resp); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if resp["decision"] != "approve" {
		t.Errorf("decision = %q, want approve", resp["decision"])
	}
}

// TestMergePostBatchOutputs_InvalidJSONSkipped verifies that non-JSON output
// from a hook is skipped and the fallback approve is used.
func TestMergePostBatchOutputs_InvalidJSONSkipped(t *testing.T) {
	results := []BatchResult{
		{Name: "hook-a", Output: []byte("not json at all"), Err: nil},
	}

	var out bytes.Buffer
	if err := mergePostBatchOutputs(results, &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp map[string]string
	if err := json.Unmarshal([]byte(strings.TrimSpace(out.String())), &resp); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if resp["decision"] != "approve" {
		t.Errorf("decision = %q, want approve after invalid JSON skip", resp["decision"])
	}
}

// TestRunPostToolBatch_Concurrent verifies no data races under -race flag.
func TestRunPostToolBatch_Concurrent(t *testing.T) {
	payload := `{"tool_name":"Edit","tool_input":{"file_path":"/tmp/test.go"},"tool_response":{},"cwd":"/tmp","hook_event_name":"PostToolUse"}`

	const goroutines = 5
	errs := make(chan error, goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			var out bytes.Buffer
			errs <- RunPostToolBatch(strings.NewReader(payload), &out)
		}()
	}
	for i := 0; i < goroutines; i++ {
		if err := <-errs; err != nil {
			t.Errorf("concurrent RunPostToolBatch error: %v", err)
		}
	}
}

// errFake is a simple error type for tests.
type errFake string

func (e errFake) Error() string { return string(e) }
