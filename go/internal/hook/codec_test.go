package hook

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/Chachamaru127/claude-code-harness/go/pkg/hookproto"
)

func TestReadInput_Valid(t *testing.T) {
	json := `{"tool_name":"Bash","tool_input":{"command":"ls -la"},"session_id":"abc-123","cwd":"/project"}`
	input, err := ReadInput(strings.NewReader(json))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if input.ToolName != "Bash" {
		t.Errorf("expected Bash, got %s", input.ToolName)
	}
	if input.SessionID != "abc-123" {
		t.Errorf("expected abc-123, got %s", input.SessionID)
	}
	if input.CWD != "/project" {
		t.Errorf("expected /project, got %s", input.CWD)
	}
	cmd, ok := input.ToolInput["command"].(string)
	if !ok || cmd != "ls -la" {
		t.Errorf("expected 'ls -la', got %v", input.ToolInput["command"])
	}
}

func TestReadInput_OfficialFields(t *testing.T) {
	json := `{
		"tool_name":"Write",
		"tool_input":{"file_path":"/test.txt","content":"hello"},
		"session_id":"sess-1",
		"transcript_path":"/tmp/transcript.jsonl",
		"cwd":"/project",
		"permission_mode":"auto",
		"hook_event_name":"PreToolUse"
	}`
	input, err := ReadInput(strings.NewReader(json))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if input.TranscriptPath != "/tmp/transcript.jsonl" {
		t.Errorf("expected transcript_path, got %s", input.TranscriptPath)
	}
	if input.PermissionMode != "auto" {
		t.Errorf("expected auto, got %s", input.PermissionMode)
	}
	if input.HookEventName != "PreToolUse" {
		t.Errorf("expected PreToolUse, got %s", input.HookEventName)
	}
}

func TestReadInput_MissingToolName(t *testing.T) {
	json := `{"tool_input":{"command":"ls"}}`
	_, err := ReadInput(strings.NewReader(json))
	if err == nil {
		t.Fatal("expected error for missing tool_name")
	}
}

func TestReadInput_EmptyInput(t *testing.T) {
	_, err := ReadInput(strings.NewReader(""))
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestReadInput_InvalidJSON(t *testing.T) {
	_, err := ReadInput(strings.NewReader("{invalid json}"))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestReadInput_NilToolInput(t *testing.T) {
	json := `{"tool_name":"Read"}`
	input, err := ReadInput(strings.NewReader(json))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if input.ToolInput == nil {
		t.Error("expected ToolInput to be initialized to empty map")
	}
}

func TestWriteResult(t *testing.T) {
	var buf bytes.Buffer
	result := hookproto.HookResult{
		Decision: hookproto.DecisionDeny,
		Reason:   "blocked",
	}
	err := WriteResult(&buf, result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `"decision":"deny"`) {
		t.Errorf("expected deny in output, got: %s", out)
	}
	if !strings.Contains(out, `"reason":"blocked"`) {
		t.Errorf("expected reason in output, got: %s", out)
	}
	if !strings.HasSuffix(out, "\n") {
		t.Error("expected newline at end of output")
	}
}

func TestSafeResult(t *testing.T) {
	result := SafeResult(fmt.Errorf("test error"))
	if result.Decision != hookproto.DecisionApprove {
		t.Errorf("expected approve (safe fallback), got %s", result.Decision)
	}
	if !strings.Contains(result.Reason, "test error") {
		t.Errorf("expected error message in reason, got: %s", result.Reason)
	}
}
