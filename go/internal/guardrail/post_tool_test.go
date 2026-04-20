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

func TestHasSuspiciousContent(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{name: "empty string", content: "", want: false},
		{name: "normal go code", content: "func main() {\n\tfmt.Println(\"hello\")\n}\n", want: false},
		{name: "Anthropic API key prefix", content: `apiKey = "sk-ant-api03-xxx123"`, want: true},
		{name: "JWT token prefix", content: `eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJ1c2VyIn0`, want: true},
		{name: "password assignment", content: `password = "hunter2"`, want: true},
		{name: "AWS access key prefix", content: `awsKey = "AKIA1234567890ABCDEF"`, want: true},
		{name: "process.env reference", content: `const key = process.env.API_KEY`, want: true},
		{name: "innerHTML assignment", content: `el.innerHTML = userInput`, want: true},
		{name: "eval call", content: `eval(userInput)`, want: true},
		{name: "exec call", content: "exec(`cmd ${arg}`)", want: true},
		{name: "GitHub token prefix", content: `token = "ghp_abc123"`, want: true},
		{name: "Stripe live key", content: `sk_live_somekey`, want: true},
		{name: "eyJ jwt fragment", content: `"eyJzdWIiOiJ1c2VyIn0"`, want: true},
		{name: "plain struct definition", content: "type Config struct {\n\tPort int\n}\n", want: false},
		{name: "innocuous comment", content: "// This function processes data without side effects", want: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := hasSuspiciousContent(tc.content)
			if got != tc.want {
				t.Errorf("hasSuspiciousContent(%q) = %v, want %v", tc.content, got, tc.want)
			}
		})
	}
}

func TestPostTool_TamperingWarningIncludesTaxonomyID(t *testing.T) {
	input := hookproto.HookInput{
		ToolName: "Write",
		ToolInput: map[string]interface{}{
			"file_path": "src/utils.test.ts",
			"content":   "it.skip('broken test', () => {});",
		},
	}
	result := EvaluatePostTool(input)
	if !strings.Contains(result.SystemMessage, "FT-TAMPER-01") {
		t.Errorf("expected FT-TAMPER-01 in warning, got: %s", result.SystemMessage)
	}
	if !strings.Contains(result.SystemMessage, `"taxonomy_ids"`) {
		t.Errorf("expected taxonomy_ids JSON field in warning, got: %s", result.SystemMessage)
	}
}

func TestDetectTampering_TaxonomyIDPopulated(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		isTest     bool
		wantFTID   string
		wantExists bool
	}{
		{"it.skip → FT-TAMPER-01", "it.skip('x', () => {})", true, "FT-TAMPER-01", true},
		{"xdescribe → FT-TAMPER-02", "xdescribe('x', () => {})", true, "FT-TAMPER-02", true},
		{"pytest.mark.skip → FT-TAMPER-03", "@pytest.mark.skip", true, "FT-TAMPER-03", true},
		{"t.Skip → FT-TAMPER-04", "t.Skip()", true, "FT-TAMPER-04", true},
		{"//expect( → FT-TAMPER-05", "// expect(x)", true, "FT-TAMPER-05", true},
		{"//assert( → FT-TAMPER-06", "// assertEqual(a, b)", true, "FT-TAMPER-06", true},
		{"eslint-disable → FT-TAMPER-08", "// eslint-disable", false, "FT-TAMPER-08", true},
		{"continue-on-error → FT-TAMPER-09", "continue-on-error: true", false, "FT-TAMPER-09", true},
		{"no tampering", "const x = 1;", true, "", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			warnings := detectTampering(tc.content, tc.isTest)
			if !tc.wantExists {
				if len(warnings) != 0 {
					t.Errorf("expected no warnings, got %d", len(warnings))
				}
				return
			}
			found := false
			for _, w := range warnings {
				if w.TaxonomyID == tc.wantFTID {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected TaxonomyID %s in warnings %+v", tc.wantFTID, warnings)
			}
		})
	}
}

func TestDetectSecurityRisksSkippedForCleanContent(t *testing.T) {
	// Clean content with no suspicious strings should get immediate approve
	// with no security warning — verifying the pre-screen path via EvaluatePostTool.
	input := hookproto.HookInput{
		ToolName: "Write",
		ToolInput: map[string]interface{}{
			"file_path": "pkg/util/math.go",
			"content":   "package util\n\nfunc Add(a, b int) int { return a + b }\n",
		},
	}
	result := EvaluatePostTool(input)
	if result.SystemMessage != "" {
		t.Errorf("expected no warnings for clean content, got: %s", result.SystemMessage)
	}
	if result.Decision != hookproto.DecisionApprove {
		t.Errorf("expected approve for clean content, got: %s", result.Decision)
	}
}

func TestPostTool_StructuredSecretDetection(t *testing.T) {
	// Synthetic test values built from parts to avoid GitHub Push Protection false positives.
	// These are NOT real credentials — they are pattern fixtures for unit testing only.
	testAntKey    := `apiKey = "` + "sk-ant-" + `abc123def456ghi789jklmnopqrst"`
	testOAIKey    := `openaiKey = "` + "sk-" + `abc123def456ghi789jklmnopqrstuvwx"`
	testGHPToken  := `githubToken = "` + "ghp_" + `1234567890abcdefghijklmnopqrstuvwxyz"`
	testStripeKey := `stripeKey = "` + "sk_live" + `_abcdefghijklmnopqrst1234567"`

	tests := []struct {
		name        string
		content     string
		shouldMatch bool
		wantContain string
	}{
		// Positive cases — must trigger warning
		{
			name:        "Anthropic API key",
			content:     testAntKey,
			shouldMatch: true,
			wantContain: "Anthropic API key",
		},
		{
			name:        "OpenAI API key",
			content:     testOAIKey,
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
			content:     testGHPToken,
			shouldMatch: true,
			wantContain: "GitHub token",
		},
		{
			name:        "Stripe live secret key",
			content:     testStripeKey,
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
