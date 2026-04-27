package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/tim-hub/powerball-harness/go/internal/piiguard"
)

// ---------------------------------------------------------------------------
// pii-prompt handler tests
// ---------------------------------------------------------------------------

func TestPIIPrompt_Clean(t *testing.T) {
	in := strings.NewReader(`{"prompt": "Hello, this is a clean message."}`)
	var out, errOut bytes.Buffer
	exit := piiPromptHandler(in, &out, &errOut)
	if exit != 0 {
		t.Errorf("clean prompt should exit 0, got %d (out=%q)", exit, out.String())
	}
	if out.Len() > 0 {
		t.Errorf("clean prompt should produce no stdout, got %q", out.String())
	}
}

func TestPIIPrompt_WithGitHubToken(t *testing.T) {
	planted := "GH_TOKEN=" + "ghp_" + "abcdefghijklmnopqrstuvwxyz12345678901234"
	in := strings.NewReader(`{"prompt": "Please use ` + planted + ` to deploy"}`)
	var out, errOut bytes.Buffer
	exit := piiPromptHandler(in, &out, &errOut)
	if exit != 1 {
		t.Errorf("planted GitHub token should exit 1, got %d", exit)
	}
	if !strings.Contains(out.String(), `"decision":"block"`) {
		t.Errorf("expected decision:block in stdout, got %q", out.String())
	}
	if !strings.Contains(out.String(), "Privacy Guard") {
		t.Errorf("expected Privacy Guard reason, got %q", out.String())
	}
}

func TestPIIPrompt_DisabledByEnv(t *testing.T) {
	t.Setenv(piiguardDisabledEnvVar, "1")
	planted := "GH_TOKEN=" + "ghp_" + "abcdefghijklmnopqrstuvwxyz12345678901234"
	in := strings.NewReader(`{"prompt": "secret: ` + planted + `"}`)
	var out, errOut bytes.Buffer
	exit := piiPromptHandler(in, &out, &errOut)
	if exit != 0 {
		t.Errorf("kill switch should exit 0, got %d", exit)
	}
	if out.Len() > 0 {
		t.Errorf("kill switch should produce no stdout, got %q", out.String())
	}
}

func TestPIIPrompt_EmptyInput(t *testing.T) {
	in := strings.NewReader("")
	var out, errOut bytes.Buffer
	exit := piiPromptHandler(in, &out, &errOut)
	if exit != 0 {
		t.Errorf("empty stdin should exit 0, got %d", exit)
	}
}

func TestPIIPrompt_InvalidJSON(t *testing.T) {
	in := strings.NewReader(`{not valid json`)
	var out, errOut bytes.Buffer
	exit := piiPromptHandler(in, &out, &errOut)
	if exit != 0 {
		t.Errorf("invalid JSON should fail open (exit 0), got %d", exit)
	}
}

func TestPIIPrompt_MissingPromptField(t *testing.T) {
	in := strings.NewReader(`{"other_field": "no prompt here"}`)
	var out, errOut bytes.Buffer
	exit := piiPromptHandler(in, &out, &errOut)
	if exit != 0 {
		t.Errorf("missing prompt field should exit 0, got %d", exit)
	}
}

// ---------------------------------------------------------------------------
// pii-pretool handler tests
// ---------------------------------------------------------------------------

func TestPIIPreTool_CleanBash(t *testing.T) {
	in := strings.NewReader(`{
		"tool_name": "Bash",
		"tool_input": {"command": "ls -la"}
	}`)
	var out, errOut bytes.Buffer
	exit := piiPreToolHandler(in, &out, &errOut)
	if exit != 0 || out.Len() > 0 {
		t.Errorf("clean Bash should produce no output (got exit=%d, out=%q)", exit, out.String())
	}
}

func TestPIIPreTool_BashWithAWS(t *testing.T) {
	planted := "export AWS_KEY=" + "AKIA" + "IOSFODNN7EXAMPLE"
	in := strings.NewReader(`{
		"tool_name": "Bash",
		"tool_input": {"command": "` + planted + `"}
	}`)
	var out, errOut bytes.Buffer
	exit := piiPreToolHandler(in, &out, &errOut)
	if exit != 0 {
		t.Errorf("pre-tool always exits 0, got %d", exit)
	}
	if !strings.Contains(out.String(), `"permissionDecision":"deny"`) {
		t.Errorf("expected permissionDecision:deny, got %q", out.String())
	}
	if !strings.Contains(out.String(), `"hookEventName":"PreToolUse"`) {
		t.Errorf("expected hookEventName:PreToolUse, got %q", out.String())
	}
}

func TestPIIPreTool_WriteWithSecret(t *testing.T) {
	planted := "ANTHROPIC_KEY=sk-ant-" + "api03-abcdefghijklmnopqrstuvwxyz"
	in := strings.NewReader(`{
		"tool_name": "Write",
		"tool_input": {"file_path": "/tmp/foo.go", "content": "config := \"` + planted + `\""}
	}`)
	var out, errOut bytes.Buffer
	piiPreToolHandler(in, &out, &errOut)
	if !strings.Contains(out.String(), `"permissionDecision":"deny"`) {
		t.Errorf("Write with planted secret should be denied, got %q", out.String())
	}
}

func TestPIIPreTool_MultiEditWithSecret(t *testing.T) {
	planted := "ghp_" + "abcdefghijklmnopqrstuvwxyz12345678901234"
	in := strings.NewReader(`{
		"tool_name": "MultiEdit",
		"tool_input": {
			"file_path": "/tmp/x.go",
			"edits": [
				{"old_string": "OLD", "new_string": "TOKEN=` + planted + `"}
			]
		}
	}`)
	var out, errOut bytes.Buffer
	piiPreToolHandler(in, &out, &errOut)
	if !strings.Contains(out.String(), `"permissionDecision":"deny"`) {
		t.Errorf("MultiEdit with planted secret should be denied, got %q", out.String())
	}
}

func TestPIIPreTool_DisabledByEnv(t *testing.T) {
	t.Setenv(piiguardDisabledEnvVar, "1")
	planted := "AKIA" + "IOSFODNN7EXAMPLE"
	in := strings.NewReader(`{
		"tool_name": "Bash",
		"tool_input": {"command": "echo ` + planted + `"}
	}`)
	var out, errOut bytes.Buffer
	exit := piiPreToolHandler(in, &out, &errOut)
	if exit != 0 || out.Len() > 0 {
		t.Errorf("kill switch should produce no output, got exit=%d out=%q", exit, out.String())
	}
}

// ---------------------------------------------------------------------------
// pii-posttool handler tests
// ---------------------------------------------------------------------------

func TestPIIPostTool_CleanResponse(t *testing.T) {
	in := strings.NewReader(`{
		"tool_name": "Bash",
		"tool_response": {"stdout": "all good\n", "stderr": ""}
	}`)
	var out, errOut bytes.Buffer
	exit := piiPostToolHandler(in, &out, &errOut)
	if exit != 0 || out.Len() > 0 {
		t.Errorf("clean response should produce no output, got exit=%d out=%q", exit, out.String())
	}
}

func TestPIIPostTool_PrivateKeyInOutput(t *testing.T) {
	keyBlob := "-----BEGIN PRIVATE" + " KEY-----\\nMIIEvQIBADAN\\n-----END PRIVATE" + " KEY-----"
	in := strings.NewReader(`{
		"tool_name": "Read",
		"tool_response": {"content": "` + keyBlob + `"}
	}`)
	var out, errOut bytes.Buffer
	exit := piiPostToolHandler(in, &out, &errOut)
	if exit != 0 {
		t.Errorf("post-tool always exits 0, got %d", exit)
	}
	if !strings.Contains(out.String(), `"hookEventName":"PostToolUse"`) {
		t.Errorf("expected hookEventName:PostToolUse, got %q", out.String())
	}
	if !strings.Contains(out.String(), "additionalContext") {
		t.Errorf("expected additionalContext, got %q", out.String())
	}
	if !strings.Contains(out.String(), "Privacy Guard") {
		t.Errorf("expected Privacy Guard message in additionalContext, got %q", out.String())
	}
}

func TestPIIPostTool_StringResponse(t *testing.T) {
	planted := "ghp_" + "abcdefghijklmnopqrstuvwxyz12345678901234"
	in := strings.NewReader(`{
		"tool_name": "Bash",
		"tool_response": "TOKEN=` + planted + `"
	}`)
	var out, errOut bytes.Buffer
	piiPostToolHandler(in, &out, &errOut)
	if !strings.Contains(out.String(), "additionalContext") {
		t.Errorf("string tool_response with secret should produce additionalContext, got %q", out.String())
	}
}

func TestPIIPostTool_DisabledByEnv(t *testing.T) {
	t.Setenv(piiguardDisabledEnvVar, "1")
	planted := "ghp_" + "abcdefghijklmnopqrstuvwxyz12345678901234"
	in := strings.NewReader(`{
		"tool_name": "Bash",
		"tool_response": {"stdout": "TOKEN=` + planted + `"}
	}`)
	var out, errOut bytes.Buffer
	exit := piiPostToolHandler(in, &out, &errOut)
	if exit != 0 || out.Len() > 0 {
		t.Errorf("kill switch should produce no output, got exit=%d out=%q", exit, out.String())
	}
}

func TestPIIPostTool_NoToolResponse(t *testing.T) {
	in := strings.NewReader(`{"tool_name": "Bash", "tool_input": {"command": "ls"}}`)
	var out, errOut bytes.Buffer
	exit := piiPostToolHandler(in, &out, &errOut)
	if exit != 0 || out.Len() > 0 {
		t.Errorf("missing tool_response should be no-op, got exit=%d out=%q", exit, out.String())
	}
}

// ---------------------------------------------------------------------------
// formatter tests
// ---------------------------------------------------------------------------

func TestFormatPromptBlockReason_Shape(t *testing.T) {
	in := strings.NewReader(`{"prompt": "TOKEN=` + "ghp_" + `abcdefghijklmnopqrstuvwxyz12345678901234"}`)
	var out, errOut bytes.Buffer
	piiPromptHandler(in, &out, &errOut)
	body := out.String()
	for _, want := range []string{"Privacy Guard", "Risk Score", "GitHub Token", "Found", "Please remove"} {
		if !strings.Contains(body, want) {
			t.Errorf("expected reason to contain %q, got: %s", want, body)
		}
	}
}

// ---------------------------------------------------------------------------
// helper coverage tests (piiguardFilterDisabled, piiguardEnabled)
// ---------------------------------------------------------------------------

// TestPiiguardFilterDisabled covers the env-var rule disable list parser.
func TestPiiguardFilterDisabled(t *testing.T) {
	makeRules := func() []piiguard.Rule {
		return []piiguard.Rule{
			{ID: "rule-a"}, {ID: "rule-b"}, {ID: "rule-c"}, {ID: "rule-d"},
		}
	}

	t.Run("no env var leaves rules intact", func(t *testing.T) {
		t.Setenv(piiguardRulesEnvVar, "")
		out := piiguardFilterDisabled(makeRules())
		if len(out) != 4 {
			t.Errorf("want 4 rules, got %d", len(out))
		}
	})

	t.Run("single rule disabled", func(t *testing.T) {
		t.Setenv(piiguardRulesEnvVar, "rule-b")
		out := piiguardFilterDisabled(makeRules())
		for _, r := range out {
			if r.ID == "rule-b" {
				t.Errorf("rule-b should have been filtered out")
			}
		}
		if len(out) != 3 {
			t.Errorf("want 3 rules, got %d", len(out))
		}
	})

	t.Run("multiple rules with whitespace tolerance", func(t *testing.T) {
		t.Setenv(piiguardRulesEnvVar, " rule-a , rule-c ")
		out := piiguardFilterDisabled(makeRules())
		ids := map[string]bool{}
		for _, r := range out {
			ids[r.ID] = true
		}
		if !ids["rule-b"] || !ids["rule-d"] {
			t.Errorf("expected rule-b and rule-d to remain, got %v", ids)
		}
		if ids["rule-a"] || ids["rule-c"] {
			t.Errorf("rule-a and rule-c should have been filtered out")
		}
	})

	t.Run("empty entries ignored", func(t *testing.T) {
		t.Setenv(piiguardRulesEnvVar, ",,, ,,")
		out := piiguardFilterDisabled(makeRules())
		if len(out) != 4 {
			t.Errorf("empty/whitespace-only should leave rules intact, got %d", len(out))
		}
	})
}

// TestPiiguardEnabled covers the kill-switch env var parser.
func TestPiiguardEnabled(t *testing.T) {
	cases := []struct {
		val  string
		want bool
	}{
		{"", true},
		{"0", true},
		{"false", true},
		{"1", false},
		{"true", false},
		{"TRUE", false},
		{"yes", false},
		{"YES", false},
		{" 1 ", false},
		{"random-other-value", true},
	}
	for _, tc := range cases {
		t.Run("val="+tc.val, func(t *testing.T) {
			t.Setenv(piiguardDisabledEnvVar, tc.val)
			if got := piiguardEnabled(); got != tc.want {
				t.Errorf("piiguardEnabled() with %q: want %v, got %v", tc.val, tc.want, got)
			}
		})
	}
}
