package guardrail

import (
	"strings"
	"testing"

	"github.com/tim-hub/powerball-harness/go/pkg/hookproto"
)

func TestPostTool_NonWriteApproved(t *testing.T) {
	input := hookproto.HookInput{
		ToolName:  "Read",
		ToolInput: map[string]interface{}{"file_path": "/test.txt"},
	}
	result := EvaluatePostTool(input)
	if result.Decision != hookproto.DecisionApprove {
		t.Errorf("expected approve for Read, got %s", result.Decision)
	}
	if result.SystemMessage != "" {
		t.Errorf("expected no systemMessage, got: %s", result.SystemMessage)
	}
}

func TestPostTool_TamperingDetected(t *testing.T) {
	input := hookproto.HookInput{
		ToolName: "Write",
		ToolInput: map[string]interface{}{
			"file_path": "src/utils.test.ts",
			"content":   "describe.skip('should work', () => {});",
		},
	}
	result := EvaluatePostTool(input)
	if result.Decision != hookproto.DecisionApprove {
		t.Errorf("expected approve (warning only), got %s", result.Decision)
	}
	if !strings.Contains(result.SystemMessage, "tampering") {
		t.Errorf("expected tampering warning, got: %s", result.SystemMessage)
	}
}

func TestPostTool_SecurityRiskDetected(t *testing.T) {
	input := hookproto.HookInput{
		ToolName: "Write",
		ToolInput: map[string]interface{}{
			"file_path": "src/main.ts",
			"content":   `password = "super_secret_12345"`,
		},
	}
	result := EvaluatePostTool(input)
	if !strings.Contains(result.SystemMessage, "Security risk") {
		t.Errorf("expected security warning, got: %s", result.SystemMessage)
	}
}

func TestPostTool_CleanWrite(t *testing.T) {
	input := hookproto.HookInput{
		ToolName: "Write",
		ToolInput: map[string]interface{}{
			"file_path": "src/main.ts",
			"content":   "const x = 42;\nconsole.log(x);",
		},
	}
	result := EvaluatePostTool(input)
	if result.SystemMessage != "" {
		t.Errorf("expected no warnings for clean code, got: %s", result.SystemMessage)
	}
}

func TestPostTool_EditNewString(t *testing.T) {
	input := hookproto.HookInput{
		ToolName: "Edit",
		ToolInput: map[string]interface{}{
			"file_path":  "src/app.test.ts",
			"new_string": "it.skip('broken test', () => {});",
		},
	}
	result := EvaluatePostTool(input)
	if !strings.Contains(result.SystemMessage, "tampering") {
		t.Errorf("expected tampering warning for Edit, got: %s", result.SystemMessage)
	}
}

func TestPostTool_CIConfigTampering(t *testing.T) {
	input := hookproto.HookInput{
		ToolName: "Write",
		ToolInput: map[string]interface{}{
			"file_path": ".github/workflows/ci.yml",
			"content":   "continue-on-error: true",
		},
	}
	result := EvaluatePostTool(input)
	if !strings.Contains(result.SystemMessage, "tampering") {
		t.Errorf("expected CI tampering warning, got: %s", result.SystemMessage)
	}
}
