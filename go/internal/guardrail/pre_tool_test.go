package guardrail

import (
	"encoding/json"
	"testing"

	"github.com/tim-hub/powerball-harness/go/pkg/hookproto"
)

// ---------------------------------------------------------------------------
// PreToolToOutput — output conversion tests
// ---------------------------------------------------------------------------

func TestPreToolToOutput_Deny(t *testing.T) {
	result := hookproto.HookResult{
		Decision: hookproto.DecisionDeny,
		Reason:   "forbidden command",
	}
	out := PreToolToOutput(result)
	if out == nil {
		t.Fatal("expected non-nil output for deny")
	}
	if out.HookSpecificOutput.PermissionDecision != "deny" {
		t.Errorf("expected permissionDecision=deny, got %s", out.HookSpecificOutput.PermissionDecision)
	}
	if out.HookSpecificOutput.PermissionDecisionReason != "forbidden command" {
		t.Errorf("expected reason to be 'forbidden command', got %s", out.HookSpecificOutput.PermissionDecisionReason)
	}
}

func TestPreToolToOutput_Ask(t *testing.T) {
	result := hookproto.HookResult{
		Decision: hookproto.DecisionAsk,
		Reason:   "confirm action",
	}
	out := PreToolToOutput(result)
	if out == nil {
		t.Fatal("expected non-nil output for ask")
	}
	if out.HookSpecificOutput.PermissionDecision != "ask" {
		t.Errorf("expected permissionDecision=ask, got %s", out.HookSpecificOutput.PermissionDecision)
	}
}

func TestPreToolToOutput_ApproveNoMessage(t *testing.T) {
	result := hookproto.HookResult{
		Decision: hookproto.DecisionApprove,
	}
	out := PreToolToOutput(result)
	// Pure approve with no message → nil (empty output, exit 0)
	if out != nil {
		t.Error("expected nil output for pure approve with no message")
	}
}

func TestPreToolToOutput_ApproveWithSystemMessage(t *testing.T) {
	result := hookproto.HookResult{
		Decision:      hookproto.DecisionApprove,
		SystemMessage: "Warning: reading a sensitive file",
	}
	out := PreToolToOutput(result)
	if out == nil {
		t.Fatal("expected non-nil output for approve with system message")
	}
	if out.HookSpecificOutput.PermissionDecision != "allow" {
		t.Errorf("expected permissionDecision=allow, got %s", out.HookSpecificOutput.PermissionDecision)
	}
	if out.HookSpecificOutput.AdditionalContext == "" {
		t.Error("expected AdditionalContext to be set for systemMessage")
	}
}

// ---------------------------------------------------------------------------
// Task 38.0.2: DecisionDefer switch case (CC 2.1.89)
// ---------------------------------------------------------------------------

func TestPreToolToOutput_Defer(t *testing.T) {
	result := hookproto.HookResult{
		Decision: hookproto.DecisionDefer,
		Reason:   "requires human review",
	}
	out := PreToolToOutput(result)
	if out == nil {
		t.Fatal("expected non-nil output for defer")
	}
	if out.HookSpecificOutput.PermissionDecision != "defer" {
		t.Errorf("expected permissionDecision=defer, got %s", out.HookSpecificOutput.PermissionDecision)
	}
	if out.HookSpecificOutput.PermissionDecisionReason != "requires human review" {
		t.Errorf("expected reason to be 'requires human review', got %s", out.HookSpecificOutput.PermissionDecisionReason)
	}
}

func TestPreToolToOutput_DeferJSON(t *testing.T) {
	// Verify the JSON serialization includes the correct fields
	result := hookproto.HookResult{
		Decision: hookproto.DecisionDefer,
		Reason:   "requires human review",
	}
	out := PreToolToOutput(result)
	if out == nil {
		t.Fatal("expected non-nil output for defer")
	}

	jsonBytes, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("failed to marshal output: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &decoded); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	hookOutput, ok := decoded["hookSpecificOutput"].(map[string]interface{})
	if !ok {
		t.Fatal("expected hookSpecificOutput in JSON")
	}

	if hookOutput["permissionDecision"] != "defer" {
		t.Errorf("expected permissionDecision=defer in JSON, got %v", hookOutput["permissionDecision"])
	}
	if hookOutput["permissionDecisionReason"] != "requires human review" {
		t.Errorf("expected permissionDecisionReason='requires human review' in JSON, got %v", hookOutput["permissionDecisionReason"])
	}
}

// ---------------------------------------------------------------------------
// FormatPreToolResult — exit code tests
// ---------------------------------------------------------------------------

func TestFormatPreToolResult_DenyExitCode2(t *testing.T) {
	result := hookproto.HookResult{
		Decision: hookproto.DecisionDeny,
		Reason:   "blocked",
	}
	out, code := FormatPreToolResult(result)
	if code != 2 {
		t.Errorf("expected exit code 2 for deny, got %d", code)
	}
	if out == nil {
		t.Error("expected non-nil output for deny")
	}
}

func TestFormatPreToolResult_ApproveExitCode0(t *testing.T) {
	result := hookproto.HookResult{
		Decision: hookproto.DecisionApprove,
	}
	out, code := FormatPreToolResult(result)
	if code != 0 {
		t.Errorf("expected exit code 0 for approve, got %d", code)
	}
	if out != nil {
		t.Error("expected nil output for pure approve")
	}
}
