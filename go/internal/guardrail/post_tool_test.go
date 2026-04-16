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

func TestPostTool_StructuredSecretDetection(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		shouldMatch bool
		wantContain string
	}{
		// Positive cases — must trigger warning
		{
			name:        "Anthropic API key",
			content:     `apiKey = "sk-ant-abc123def456ghi789jklmnopqrst"`,
			shouldMatch: true,
			wantContain: "Anthropic API key",
		},
		{
			name:        "OpenAI API key",
			content:     `openaiKey = "sk-abc123def456ghi789jklmnopqrstuvwx"`,
			shouldMatch: true,
			wantContain: "OpenAI/generic API key",
		},
		{
			name:        "AWS access key",
			content:     `awsKey = "AKIAIOSFODNN7EXAMPLE"`,
			shouldMatch: true,
			wantContain: "AWS access key",
		},
		{
			name:        "GitHub personal access token",
			content:     `githubToken = "ghp_1234567890abcdefghijklmnopqrstuvwxyz"`,
			shouldMatch: true,
			wantContain: "GitHub token",
		},
		{
			name:        "Stripe live secret key",
			content:     `stripeKey = "rk_live_abcdefghijklmnopqrst1234567"`,
			shouldMatch: true,
			wantContain: "Stripe live key",
		},
		{
			name:        "JWT token",
			content:     `jwt = "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJ1c2VyMTIzIn0"`,
			shouldMatch: true,
			wantContain: "JWT token",
		},
		// Negative cases — must NOT trigger warning
		{
			name:        "sk-ant- too short",
			content:     `key = "sk-ant-"`,
			shouldMatch: false,
		},
		{
			name:        "AKIA with no suffix",
			content:     `text = "AKIA"`,
			shouldMatch: false,
		},
		{
			name:        "GitHub token too short",
			content:     `token = "ghp_short"`,
			shouldMatch: false,
		},
		{
			name:        "Stripe test key (not live)",
			content:     `key = "sk_test_abc123"`,
			shouldMatch: false,
		},
		{
			name:        "Normal prose",
			content:     "This is a normal comment with no secrets here.",
			shouldMatch: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			input := hookproto.HookInput{
				ToolName: "Write",
				ToolInput: map[string]interface{}{
					"file_path": "src/config.ts",
					"content":   tc.content,
				},
			}
			result := EvaluatePostTool(input)
			hasWarning := strings.Contains(result.SystemMessage, "Security risk")
			if tc.shouldMatch && !hasWarning {
				t.Errorf("expected security warning for %q, got none (message: %q)", tc.name, result.SystemMessage)
			}
			if !tc.shouldMatch && hasWarning {
				t.Errorf("expected no security warning for %q, but got: %s", tc.name, result.SystemMessage)
			}
			if tc.shouldMatch && tc.wantContain != "" && !strings.Contains(result.SystemMessage, tc.wantContain) {
				t.Errorf("expected warning to contain %q for %q, got: %s", tc.wantContain, tc.name, result.SystemMessage)
			}
		})
	}
}
