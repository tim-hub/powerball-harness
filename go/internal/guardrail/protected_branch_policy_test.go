package guardrail

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tim-hub/powerball-harness/go/pkg/hookproto"
)

func TestNormalizeProtectedBranchPushPolicy(t *testing.T) {
	cases := map[string]string{
		"":           "ask",
		"ask":        "ask",
		"confirm":    "ask",
		"DENY":       "deny",
		"block":      "deny",
		"allow":      "allow",
		"approve":    "allow",
		"unexpected": "ask",
	}
	for input, want := range cases {
		if got := normalizeProtectedBranchPushPolicy(input); got != want {
			t.Errorf("normalizeProtectedBranchPushPolicy(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestBuildContextProtectedBranchPushPolicyFromEnv(t *testing.T) {
	t.Setenv("HARNESS_PROTECTED_BRANCH_PUSH_POLICY", "deny")

	ctx := BuildContext(hookproto.HookInput{
		CWD:       t.TempDir(),
		ToolName:  "Bash",
		ToolInput: map[string]interface{}{"command": "git push origin main"},
	})

	if ctx.ProtectedBranchPushPolicy != "deny" {
		t.Fatalf("ProtectedBranchPushPolicy = %q, want deny", ctx.ProtectedBranchPushPolicy)
	}
}

func TestBuildContextProtectedBranchPushPolicyFromProjectYAML(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".claude-code-harness.config.yaml")
	if err := os.WriteFile(configPath, []byte("safety:\n  protected_branch_push: allow\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	ctx := BuildContext(hookproto.HookInput{
		CWD:       dir,
		ToolName:  "Bash",
		ToolInput: map[string]interface{}{"command": "git push origin main"},
	})

	if ctx.ProtectedBranchPushPolicy != "allow" {
		t.Fatalf("ProtectedBranchPushPolicy = %q, want allow", ctx.ProtectedBranchPushPolicy)
	}
}

func TestBuildContextProtectedBranchPushPolicyFromHarnessTOML(t *testing.T) {
	dir := t.TempDir()
	tomlPath := filepath.Join(dir, "harness.toml")
	data := []byte(`
[project]
name = "test"

[safety.permissions]
protectedBranchPush = "deny"
`)
	if err := os.WriteFile(tomlPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	ctx := BuildContext(hookproto.HookInput{
		CWD:       dir,
		ToolName:  "Bash",
		ToolInput: map[string]interface{}{"command": "git push origin main"},
	})

	if ctx.ProtectedBranchPushPolicy != "deny" {
		t.Fatalf("ProtectedBranchPushPolicy = %q, want deny", ctx.ProtectedBranchPushPolicy)
	}
}

func TestBuildContextProtectedBranchPushPolicyDefaultAsk(t *testing.T) {
	ctx := BuildContext(hookproto.HookInput{
		CWD:       t.TempDir(),
		ToolName:  "Bash",
		ToolInput: map[string]interface{}{"command": "git push origin main"},
	})

	if ctx.ProtectedBranchPushPolicy != "ask" {
		t.Fatalf("ProtectedBranchPushPolicy = %q, want ask", ctx.ProtectedBranchPushPolicy)
	}
}
