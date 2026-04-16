package hook

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// TestRunPreToolBatch_Smoke verifies that RunPreToolBatch completes without
// error when given a minimal PreToolUse payload.
func TestRunPreToolBatch_Smoke(t *testing.T) {
	payload := `{"tool_name":"Write","tool_input":{"file_path":"/tmp/test.go"},"cwd":"/tmp","hook_event_name":"PreToolUse"}`

	var out bytes.Buffer
	if err := RunPreToolBatch(strings.NewReader(payload), &out); err != nil {
		t.Fatalf("RunPreToolBatch returned error: %v", err)
	}

	outStr := strings.TrimSpace(out.String())
	if outStr == "" {
		t.Fatal("expected non-empty output, got empty")
	}
	var resp map[string]interface{}
	if err := json.Unmarshal([]byte(outStr), &resp); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, outStr)
	}
}

// TestRunPreToolBatch_EmptyInput verifies graceful handling of empty stdin.
func TestRunPreToolBatch_EmptyInput(t *testing.T) {
	var out bytes.Buffer
	if err := RunPreToolBatch(strings.NewReader(""), &out); err != nil {
		t.Fatalf("RunPreToolBatch should not error on empty input: %v", err)
	}
	outStr := strings.TrimSpace(out.String())
	if outStr == "" {
		t.Fatal("expected non-empty output, got empty")
	}
	var resp map[string]interface{}
	if err := json.Unmarshal([]byte(outStr), &resp); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, outStr)
	}
}

// TestMergePreBatchOutputs_DenyWins verifies that a deny response takes
// precedence over an approve.
func TestMergePreBatchOutputs_DenyWins(t *testing.T) {
	denyJSON := `{"decision":"deny","reason":"hook denied"}`
	results := []PreBatchResult{
		{Name: "hook-a", Output: []byte(`{"decision":"approve"}`), Err: nil},
		{Name: "hook-b", Output: []byte(denyJSON), Err: nil},
	}

	var out bytes.Buffer
	if err := mergePreBatchOutputs(results, &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	outStr := strings.TrimSpace(out.String())
	var resp map[string]string
	if err := json.Unmarshal([]byte(outStr), &resp); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if resp["decision"] != "deny" {
		t.Errorf("decision = %q, want deny (deny should win over approve)", resp["decision"])
	}
}

// TestMergePreBatchOutputs_AllEmpty returns approve when all hooks produce no output.
func TestMergePreBatchOutputs_AllEmpty(t *testing.T) {
	results := []PreBatchResult{
		{Name: "hook-a", Output: nil, Err: nil},
		{Name: "hook-b", Output: []byte(""), Err: nil},
	}

	var out bytes.Buffer
	if err := mergePreBatchOutputs(results, &out); err != nil {
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

// TestMergePreBatchOutputs_OneErrorOthersRun verifies that hook errors
// don't stop the batch — other hooks still contribute their output.
func TestMergePreBatchOutputs_OneErrorOthersRun(t *testing.T) {
	successJSON := `{"decision":"approve","reason":"hook-b ok"}`
	results := []PreBatchResult{
		{Name: "hook-a", Output: nil, Err: errFake("hook-a failed")},
		{Name: "hook-b", Output: []byte(successJSON), Err: nil},
	}

	var out bytes.Buffer
	if err := mergePreBatchOutputs(results, &out); err != nil {
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

// TestRunPreToolBatch_Concurrent verifies no data races under -race flag.
func TestRunPreToolBatch_Concurrent(t *testing.T) {
	payload := `{"tool_name":"Edit","tool_input":{"file_path":"/tmp/test.go"},"cwd":"/tmp","hook_event_name":"PreToolUse"}`

	const goroutines = 5
	errs := make(chan error, goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			var out bytes.Buffer
			errs <- RunPreToolBatch(strings.NewReader(payload), &out)
		}()
	}
	for i := 0; i < goroutines; i++ {
		if err := <-errs; err != nil {
			t.Errorf("concurrent RunPreToolBatch error: %v", err)
		}
	}
}
