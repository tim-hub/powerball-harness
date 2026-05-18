package guardrail

import (
	"strings"
	"testing"

	"github.com/tim-hub/powerball-harness/go/pkg/hookproto"
)

// helper to build a RuleContext for testing
func makeCtx(toolName string, toolInput map[string]any) hookproto.RuleContext {
	return hookproto.RuleContext{
		Input: hookproto.HookInput{
			ToolName:  toolName,
			ToolInput: toolInput,
		},
		ProjectRoot:  "/project",
		WorkMode:     false,
		CodexMode:    false,
		BreezingRole: "",
	}
}

// ---------------------------------------------------------------------------
// R01: sudo block
// ---------------------------------------------------------------------------

func TestR01_SudoBlocked(t *testing.T) {
	ctx := makeCtx("Bash", map[string]any{"command": "sudo rm -rf /"})
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionDeny {
		t.Errorf("expected deny, got %s", result.Decision)
	}
}

func TestR01_SudoInMiddle(t *testing.T) {
	ctx := makeCtx("Bash", map[string]any{"command": "echo hello && sudo apt install"})
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionDeny {
		t.Errorf("expected deny, got %s", result.Decision)
	}
}

func TestR01_NoSudo(t *testing.T) {
	ctx := makeCtx("Bash", map[string]any{"command": "ls -la"})
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionApprove {
		t.Errorf("expected approve, got %s", result.Decision)
	}
}

// ---------------------------------------------------------------------------
// R02: protected path write block
// ---------------------------------------------------------------------------

func TestR02_WriteToEnv(t *testing.T) {
	ctx := makeCtx("Write", map[string]any{"file_path": ".env"})
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionDeny {
		t.Errorf("expected deny, got %s", result.Decision)
	}
}

func TestR02_WriteToGitDir(t *testing.T) {
	ctx := makeCtx("Edit", map[string]any{"file_path": ".git/config"})
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionDeny {
		t.Errorf("expected deny, got %s", result.Decision)
	}
}

func TestR02_WriteToIdRsa(t *testing.T) {
	ctx := makeCtx("Write", map[string]any{"file_path": "/home/user/.ssh/id_rsa"})
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionDeny {
		t.Errorf("expected deny, got %s", result.Decision)
	}
}

func TestR02_WriteToNormalFile(t *testing.T) {
	ctx := makeCtx("Write", map[string]any{"file_path": "/project/src/main.ts"})
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionApprove {
		t.Errorf("expected approve, got %s", result.Decision)
	}
}

func TestR02_WriteToPemFile(t *testing.T) {
	ctx := makeCtx("Write", map[string]any{"file_path": "certs/server.pem"})
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionDeny {
		t.Errorf("expected deny, got %s", result.Decision)
	}
}

// ---------------------------------------------------------------------------
// R03: Bash write to protected paths
// ---------------------------------------------------------------------------

func TestR03_EchoToEnv(t *testing.T) {
	ctx := makeCtx("Bash", map[string]any{"command": "echo SECRET=foo > .env"})
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionDeny {
		t.Errorf("expected deny, got %s", result.Decision)
	}
}

func TestR03_TeeToGit(t *testing.T) {
	ctx := makeCtx("Bash", map[string]any{"command": "echo test | tee .git/config"})
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionDeny {
		t.Errorf("expected deny, got %s", result.Decision)
	}
}

func TestR03_NormalBash(t *testing.T) {
	ctx := makeCtx("Bash", map[string]any{"command": "echo hello > output.txt"})
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionApprove {
		t.Errorf("expected approve, got %s", result.Decision)
	}
}

// ---------------------------------------------------------------------------
// R04: write outside project root
// ---------------------------------------------------------------------------

func TestR04_WriteOutsideProject(t *testing.T) {
	ctx := makeCtx("Write", map[string]any{"file_path": "/tmp/malicious.sh"})
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionAsk {
		t.Errorf("expected ask, got %s", result.Decision)
	}
}

func TestR04_WriteInsideProject(t *testing.T) {
	ctx := makeCtx("Write", map[string]any{"file_path": "/project/src/index.ts"})
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionApprove {
		t.Errorf("expected approve, got %s", result.Decision)
	}
}

func TestR04_RelativePath(t *testing.T) {
	ctx := makeCtx("Write", map[string]any{"file_path": "src/index.ts"})
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionApprove {
		t.Errorf("expected approve, got %s", result.Decision)
	}
}

func TestR04_WorkModeBypass(t *testing.T) {
	ctx := makeCtx("Write", map[string]any{"file_path": "/tmp/file.txt"})
	ctx.WorkMode = true
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionApprove {
		t.Errorf("expected approve in work mode, got %s", result.Decision)
	}
}

// ---------------------------------------------------------------------------
// R05: rm -rf confirmation
// ---------------------------------------------------------------------------

func TestR05_RmRfBlocked(t *testing.T) {
	ctx := makeCtx("Bash", map[string]any{"command": "rm -rf /var/data"})
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionAsk {
		t.Errorf("expected ask, got %s", result.Decision)
	}
}

func TestR05_RmRfWorkMode(t *testing.T) {
	ctx := makeCtx("Bash", map[string]any{"command": "rm -rf ./dist"})
	ctx.WorkMode = true
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionApprove {
		t.Errorf("expected approve in work mode, got %s", result.Decision)
	}
}

func TestR05_RmFOnly(t *testing.T) {
	// rm -f (without -r) should NOT trigger R05
	ctx := makeCtx("Bash", map[string]any{"command": "rm -f temp.txt"})
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionApprove {
		t.Errorf("expected approve for rm -f (no -r), got %s", result.Decision)
	}
}

func TestR05_RmRecursive(t *testing.T) {
	ctx := makeCtx("Bash", map[string]any{"command": "rm --recursive ./dir"})
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionAsk {
		t.Errorf("expected ask, got %s", result.Decision)
	}
}

// ---------------------------------------------------------------------------
// R06: force push block
// ---------------------------------------------------------------------------

func TestR06_ForcePush(t *testing.T) {
	ctx := makeCtx("Bash", map[string]any{"command": "git push --force origin main"})
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionDeny {
		t.Errorf("expected deny, got %s", result.Decision)
	}
}

func TestR06_ForceWithLease(t *testing.T) {
	ctx := makeCtx("Bash", map[string]any{"command": "git push --force-with-lease origin feature"})
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionDeny {
		t.Errorf("expected deny, got %s", result.Decision)
	}
}

func TestR06_ShortForce(t *testing.T) {
	ctx := makeCtx("Bash", map[string]any{"command": "git push -f origin main"})
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionDeny {
		t.Errorf("expected deny, got %s", result.Decision)
	}
}

func TestR06_NormalPush(t *testing.T) {
	ctx := makeCtx("Bash", map[string]any{"command": "git push origin feature"})
	result := EvaluateRules(ctx)
	// R12 might trigger a warning for protected branch, but this is "feature"
	if result.Decision == hookproto.DecisionDeny {
		t.Errorf("expected non-deny for normal push, got deny")
	}
}

// ---------------------------------------------------------------------------
// R07: Codex mode write block
// ---------------------------------------------------------------------------

func TestR07_CodexModeWrite(t *testing.T) {
	ctx := makeCtx("Write", map[string]any{"file_path": "/project/src/main.ts"})
	ctx.CodexMode = true
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionDeny {
		t.Errorf("expected deny in codex mode, got %s", result.Decision)
	}
}

func TestR07_CodexModeEdit(t *testing.T) {
	ctx := makeCtx("Edit", map[string]any{"file_path": "/project/src/main.ts"})
	ctx.CodexMode = true
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionDeny {
		t.Errorf("expected deny in codex mode, got %s", result.Decision)
	}
}

func TestR07_CodexModeBash(t *testing.T) {
	// Bash is NOT blocked by R07 (only Write/Edit/MultiEdit)
	ctx := makeCtx("Bash", map[string]any{"command": "echo hello"})
	ctx.CodexMode = true
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionApprove {
		t.Errorf("expected approve for Bash in codex mode, got %s", result.Decision)
	}
}

func TestR07_NormalModeWrite(t *testing.T) {
	ctx := makeCtx("Write", map[string]any{"file_path": "/project/src/main.ts"})
	ctx.CodexMode = false
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionApprove {
		t.Errorf("expected approve in normal mode, got %s", result.Decision)
	}
}

// ---------------------------------------------------------------------------
// R08: Breezing reviewer write block
// ---------------------------------------------------------------------------

func TestR08_ReviewerWrite(t *testing.T) {
	ctx := makeCtx("Write", map[string]any{"file_path": "/project/src/main.ts"})
	ctx.BreezingRole = "reviewer"
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionDeny {
		t.Errorf("expected deny for reviewer write, got %s", result.Decision)
	}
}

func TestR08_ReviewerBashGitCommit(t *testing.T) {
	ctx := makeCtx("Bash", map[string]any{"command": "git commit -m 'test'"})
	ctx.BreezingRole = "reviewer"
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionDeny {
		t.Errorf("expected deny for reviewer git commit, got %s", result.Decision)
	}
}

func TestR08_ReviewerBashReadOnly(t *testing.T) {
	ctx := makeCtx("Bash", map[string]any{"command": "cat README.md"})
	ctx.BreezingRole = "reviewer"
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionApprove {
		t.Errorf("expected approve for reviewer read-only bash, got %s", result.Decision)
	}
}

func TestR08_WorkerWrite(t *testing.T) {
	ctx := makeCtx("Write", map[string]any{"file_path": "/project/src/main.ts"})
	ctx.BreezingRole = "worker"
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionApprove {
		t.Errorf("expected approve for worker write, got %s", result.Decision)
	}
}

// ---------------------------------------------------------------------------
// R09: secret file read warning
// ---------------------------------------------------------------------------

func TestR09_ReadEnv(t *testing.T) {
	ctx := makeCtx("Read", map[string]any{"file_path": "/project/.env"})
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionApprove {
		t.Errorf("expected approve (warning only), got %s", result.Decision)
	}
	if result.SystemMessage == "" {
		t.Error("expected a warning systemMessage")
	}
}

func TestR09_ReadIdRsa(t *testing.T) {
	ctx := makeCtx("Read", map[string]any{"file_path": "/home/user/.ssh/id_rsa"})
	result := EvaluateRules(ctx)
	if result.SystemMessage == "" {
		t.Error("expected a warning systemMessage for id_rsa")
	}
}

func TestR09_ReadNormalFile(t *testing.T) {
	ctx := makeCtx("Read", map[string]any{"file_path": "/project/README.md"})
	result := EvaluateRules(ctx)
	if result.SystemMessage != "" {
		t.Errorf("expected no warning for normal file, got: %s", result.SystemMessage)
	}
}

// ---------------------------------------------------------------------------
// R10: --no-verify / --no-gpg-sign block
// ---------------------------------------------------------------------------

func TestR10_NoVerify(t *testing.T) {
	ctx := makeCtx("Bash", map[string]any{"command": "git commit --no-verify -m 'test'"})
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionDeny {
		t.Errorf("expected deny, got %s", result.Decision)
	}
}

func TestR10_NoGpgSign(t *testing.T) {
	ctx := makeCtx("Bash", map[string]any{"command": "git commit --no-gpg-sign -m 'test'"})
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionDeny {
		t.Errorf("expected deny, got %s", result.Decision)
	}
}

func TestR10_NormalCommit(t *testing.T) {
	ctx := makeCtx("Bash", map[string]any{"command": "git commit -m 'test'"})
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionApprove {
		t.Errorf("expected approve, got %s", result.Decision)
	}
}

// ---------------------------------------------------------------------------
// R11: protected branch reset --hard
// ---------------------------------------------------------------------------

func TestR11_ResetHardMain(t *testing.T) {
	ctx := makeCtx("Bash", map[string]any{"command": "git reset --hard origin/main"})
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionDeny {
		t.Errorf("expected deny, got %s", result.Decision)
	}
}

func TestR11_ResetHardMaster(t *testing.T) {
	ctx := makeCtx("Bash", map[string]any{"command": "git reset --hard master"})
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionDeny {
		t.Errorf("expected deny, got %s", result.Decision)
	}
}

func TestR11_ResetHardFeature(t *testing.T) {
	ctx := makeCtx("Bash", map[string]any{"command": "git reset --hard origin/feature"})
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionApprove {
		t.Errorf("expected approve for non-protected branch, got %s", result.Decision)
	}
}

func TestR11_ResetSoftMain(t *testing.T) {
	ctx := makeCtx("Bash", map[string]any{"command": "git reset --soft main"})
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionApprove {
		t.Errorf("expected approve for soft reset, got %s", result.Decision)
	}
}

// ---------------------------------------------------------------------------
// R12: configurable direct push policy for protected branch
// ---------------------------------------------------------------------------

func TestR12_PushToMain(t *testing.T) {
	ctx := makeCtx("Bash", map[string]any{"command": "git push origin main"})
	result := EvaluateRules(ctx)
	// Default policy is ask → approve+warn
	if result.Decision != hookproto.DecisionApprove {
		t.Errorf("expected approve (warn), got %s", result.Decision)
	}
	if result.SystemMessage == "" {
		t.Error("expected warning SystemMessage for push to main")
	}
}

func TestR12_PushToFeature(t *testing.T) {
	ctx := makeCtx("Bash", map[string]any{"command": "git push origin feature-branch"})
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionApprove {
		t.Errorf("expected approve for feature branch push, got %s", result.Decision)
	}
	if result.SystemMessage != "" {
		t.Errorf("expected no warning for feature branch push, got: %s", result.SystemMessage)
	}
}

func TestR12_PushRefspecToMain(t *testing.T) {
	ctx := makeCtx("Bash", map[string]any{"command": "git push origin HEAD:main"})
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionApprove {
		t.Errorf("expected approve (warn), got %s", result.Decision)
	}
	if result.SystemMessage == "" {
		t.Error("expected warning SystemMessage for refspec push to main")
	}
}

func TestR12_PushToMainPolicyDeny(t *testing.T) {
	ctx := makeCtx("Bash", map[string]any{"command": "git push origin main"})
	ctx.ProtectedBranchPushPolicy = "deny"
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionDeny {
		t.Errorf("expected deny, got %s", result.Decision)
	}
}

func TestR12_PushToMainPolicyAllow(t *testing.T) {
	ctx := makeCtx("Bash", map[string]any{"command": "git push origin main"})
	ctx.ProtectedBranchPushPolicy = "allow"
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionApprove {
		t.Errorf("expected approve, got %s", result.Decision)
	}
	if result.SystemMessage != "" {
		t.Errorf("expected no warning when allow policy, got: %s", result.SystemMessage)
	}
}

func TestR12_PushToMainPolicyConfirmAliasesAsk(t *testing.T) {
	ctx := makeCtx("Bash", map[string]any{"command": "git push origin main"})
	ctx.ProtectedBranchPushPolicy = "confirm"
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionApprove {
		t.Errorf("expected approve (warn) for confirm alias, got %s", result.Decision)
	}
	if result.SystemMessage == "" {
		t.Error("expected warning SystemMessage for confirm alias")
	}
}

func TestR12_PushToMainInvalidPolicyDefaultsAsk(t *testing.T) {
	ctx := makeCtx("Bash", map[string]any{"command": "git push origin main"})
	ctx.ProtectedBranchPushPolicy = "unexpected"
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionApprove {
		t.Errorf("expected approve (warn) for invalid policy, got %s", result.Decision)
	}
}

// ---------------------------------------------------------------------------
// R13: protected review paths warning
// ---------------------------------------------------------------------------

func TestR13_WritePackageJson(t *testing.T) {
	ctx := makeCtx("Write", map[string]any{"file_path": "/project/package.json"})
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionApprove {
		t.Errorf("expected approve (warning only), got %s", result.Decision)
	}
	if result.SystemMessage == "" {
		t.Error("expected warning systemMessage for package.json")
	}
}

func TestR13_WriteDockerfile(t *testing.T) {
	ctx := makeCtx("Edit", map[string]any{"file_path": "Dockerfile"})
	result := EvaluateRules(ctx)
	if result.SystemMessage == "" {
		t.Error("expected warning for Dockerfile")
	}
}

func TestR13_WriteGitHubWorkflow(t *testing.T) {
	ctx := makeCtx("Write", map[string]any{"file_path": ".github/workflows/ci.yml"})
	result := EvaluateRules(ctx)
	if result.SystemMessage == "" {
		t.Error("expected warning for GitHub workflow")
	}
}

func TestR13_WriteNormalFile(t *testing.T) {
	ctx := makeCtx("Write", map[string]any{"file_path": "/project/src/utils.ts"})
	result := EvaluateRules(ctx)
	if result.SystemMessage != "" {
		t.Errorf("expected no warning for normal file, got: %s", result.SystemMessage)
	}
}

// ---------------------------------------------------------------------------
// Rule evaluation order: first match wins
// ---------------------------------------------------------------------------

func TestFirstMatchWins(t *testing.T) {
	// sudo rm -rf should be caught by R01 (sudo) before R05 (rm -rf)
	ctx := makeCtx("Bash", map[string]any{"command": "sudo rm -rf /"})
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionDeny {
		t.Errorf("expected deny, got %s", result.Decision)
	}
	// R01 reason mentions sudo
	if result.Reason == "" || result.Reason[0:4] != "sudo" {
		// Check that the reason is about sudo, not rm -rf
		if result.Reason != "Use of sudo is prohibited. If necessary, ask the user to run it manually." {
			t.Errorf("expected sudo reason, got: %s", result.Reason)
		}
	}
}

func TestUnknownToolApproved(t *testing.T) {
	ctx := makeCtx("UnknownTool", map[string]any{})
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionApprove {
		t.Errorf("expected approve for unknown tool, got %s", result.Decision)
	}
}

// ---------------------------------------------------------------------------
// Task 38.1.2: R06 whitespace normalization tests (CC 2.1.98)
// ---------------------------------------------------------------------------

func TestR06_PushForceMultipleSpaces(t *testing.T) {
	// "git  push  --force" with multiple spaces should still be denied
	ctx := makeCtx("Bash", map[string]any{"command": "git  push  --force"})
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionDeny {
		t.Errorf("expected deny for multi-space force push, got %s", result.Decision)
	}
}

func TestR06_PushForceTabs(t *testing.T) {
	// "git\tpush\t-f" with tab separators should still be denied
	ctx := makeCtx("Bash", map[string]any{"command": "git\tpush\t-f"})
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionDeny {
		t.Errorf("expected deny for tab-separated force push, got %s", result.Decision)
	}
}

func TestR06_PushForceWithLeaseSpaces(t *testing.T) {
	// "git push   --force-with-lease" with extra spaces should still be denied
	ctx := makeCtx("Bash", map[string]any{"command": "git push   --force-with-lease"})
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionDeny {
		t.Errorf("expected deny for force-with-lease with extra spaces, got %s", result.Decision)
	}
}

// Regression: compound commands where -f appears in an UNRELATED subcommand
// (e.g. `rm -f`) must NOT trigger the force-push detector. The `.*` between
// `git push` and `-f` previously crossed shell separators and false-positively
// matched `rm -f` after `&&`/`;`/`|`.
func TestR06_CompoundCommand_RmDashF_AfterPush(t *testing.T) {
	ctx := makeCtx("Bash", map[string]any{"command": "git push origin master && rm -f /tmp/foo"})
	result := EvaluateRules(ctx)
	if result.Decision == hookproto.DecisionDeny {
		t.Errorf("expected approve for `git push && rm -f` (rm -f is unrelated), got deny: %s", result.Reason)
	}
}

func TestR06_CompoundCommand_RmDashF_Semicolon(t *testing.T) {
	ctx := makeCtx("Bash", map[string]any{"command": "git push origin master; rm -f /tmp/foo"})
	result := EvaluateRules(ctx)
	if result.Decision == hookproto.DecisionDeny {
		t.Errorf("expected approve for `git push; rm -f` (rm -f is unrelated), got deny: %s", result.Reason)
	}
}

func TestR06_CompoundCommand_RealisticReleaseCleanup(t *testing.T) {
	// The actual command that triggered the false positive in production:
	// release flow's completion-commit cleanup.
	cmd := `git push origin master && rm -f .claude/state/harness-release-wrapper.lock && echo "Done"`
	ctx := makeCtx("Bash", map[string]any{"command": cmd})
	result := EvaluateRules(ctx)
	if result.Decision == hookproto.DecisionDeny {
		t.Errorf("expected approve for release cleanup compound, got deny: %s", result.Reason)
	}
}

// True-positive guard: even when the force push is in the SECOND subcommand
// after a benign `rm -f`, the detector must still catch it.
func TestR06_CompoundCommand_ForcePushInSecondSegment(t *testing.T) {
	ctx := makeCtx("Bash", map[string]any{"command": "rm -f /tmp/foo && git push --force origin main"})
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionDeny {
		t.Errorf("expected deny for force push in second segment, got %s", result.Decision)
	}
}

// ---------------------------------------------------------------------------
// Phase 95.3: wrapper bypass coverage for R06/R11/R12
//
// CC 2.1.113 hardened deny rules to match env/sudo/watch wrappers.
// These tests verify that Harness guardrails behave the same for the three
// commonly-used wrappers, so that `env FOO=1 git push --force` (etc.) cannot
// be used to slip past force-push / reset-hard / protected-push deny.
//
// Note: when wrapped by `sudo`, R01 (sudo block) fires first and the test
// observes Deny via R01 rather than the target rule. That is still the
// intended security posture — the deny chain catches the wrapper.
// ---------------------------------------------------------------------------

// R06 wrapper bypass: env / sudo / watch must not let force-push through.
func TestR06_WrappedByEnv(t *testing.T) {
	ctx := makeCtx("Bash", map[string]any{
		"command": "env GIT_SSL_NO_VERIFY=1 git push --force origin main",
	})
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionDeny {
		t.Errorf("expected deny for env-wrapped force push, got %s", result.Decision)
	}
}

func TestR06_WrappedBySudo(t *testing.T) {
	// R01 fires first for sudo, but the deny posture is enforced.
	ctx := makeCtx("Bash", map[string]any{
		"command": "sudo git push --force origin main",
	})
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionDeny {
		t.Errorf("expected deny for sudo-wrapped force push, got %s", result.Decision)
	}
}

func TestR06_WrappedByWatch(t *testing.T) {
	ctx := makeCtx("Bash", map[string]any{
		"command": "watch -n 5 git push --force origin main",
	})
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionDeny {
		t.Errorf("expected deny for watch-wrapped force push, got %s", result.Decision)
	}
}

// R11 wrapper bypass: env / sudo / watch must not let reset --hard main through.
func TestR11_WrappedByEnv(t *testing.T) {
	ctx := makeCtx("Bash", map[string]any{
		"command": "env GIT_DIR=.git git reset --hard main",
	})
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionDeny {
		t.Errorf("expected deny for env-wrapped reset --hard main, got %s", result.Decision)
	}
}

func TestR11_WrappedBySudo(t *testing.T) {
	// R01 fires first for sudo; the deny posture is enforced.
	ctx := makeCtx("Bash", map[string]any{
		"command": "sudo git reset --hard main",
	})
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionDeny {
		t.Errorf("expected deny for sudo-wrapped reset --hard main, got %s", result.Decision)
	}
}

func TestR11_WrappedByWatch(t *testing.T) {
	ctx := makeCtx("Bash", map[string]any{
		"command": "watch -n 30 git reset --hard origin/main",
	})
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionDeny {
		t.Errorf("expected deny for watch-wrapped reset --hard origin/main, got %s", result.Decision)
	}
}

// R12 wrapper bypass: env / sudo / watch must not let direct push to main through.
// Default policy is `confirm`; for env / watch tests we set `deny` policy explicitly
// so that the wrapper bypass coverage is exercised against the strongest
// configuration. For the sudo case R01 fires first.
func TestR12_WrappedByEnv(t *testing.T) {
	ctx := makeCtx("Bash", map[string]any{
		"command": "env LANG=C git push origin main",
	})
	ctx.ProtectedBranchPushPolicy = "deny"
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionDeny {
		t.Errorf("expected deny for env-wrapped push to main with deny policy, got %s", result.Decision)
	}
}

func TestR12_WrappedBySudo(t *testing.T) {
	// R01 fires first for sudo; the deny posture is enforced regardless of policy.
	ctx := makeCtx("Bash", map[string]any{
		"command": "sudo git push origin main",
	})
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionDeny {
		t.Errorf("expected deny for sudo-wrapped push to main, got %s", result.Decision)
	}
}

func TestR12_WrappedByWatch(t *testing.T) {
	ctx := makeCtx("Bash", map[string]any{
		"command": "watch -n 60 git push origin main",
	})
	ctx.ProtectedBranchPushPolicy = "deny"
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionDeny {
		t.Errorf("expected deny for watch-wrapped push to main with deny policy, got %s", result.Decision)
	}
}

// ---------------------------------------------------------------------------
// R14: TDD enforcement — local-trial registration (non-blocking stub)
// ---------------------------------------------------------------------------

func TestR14_EnforceLevelOff_SrcFile(t *testing.T) {
	// Default (off): R14 must be completely inert regardless of the file.
	ctx := makeCtx("Write", map[string]any{"file_path": "/project/src/app.ts"})
	ctx.TddEnforceLevel = "off"
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionApprove {
		t.Errorf("expected approve when TDD enforce level is off, got %s", result.Decision)
	}
	if result.SystemMessage != "" {
		t.Errorf("expected no system message when TDD enforce level is off, got: %s", result.SystemMessage)
	}
}

func TestR14_EnforceLevelMax_SrcFile_LocalTrial(t *testing.T) {
	// max + src file: local-trial stub returns nil → approve (non-blocking)
	ctx := makeCtx("Write", map[string]any{"file_path": "/project/src/service.ts"})
	ctx.TddEnforceLevel = "max"
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionApprove {
		t.Errorf("expected approve from local-trial stub, got %s", result.Decision)
	}
}

func TestR14_EnforceLevelMax_TestFile_Skipped(t *testing.T) {
	// max + test file: R14 must not fire on test files themselves
	ctx := makeCtx("Write", map[string]any{"file_path": "/project/src/service.test.ts"})
	ctx.TddEnforceLevel = "max"
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionApprove {
		t.Errorf("expected approve for test file (R14 must not target test files), got %s", result.Decision)
	}
	if result.SystemMessage != "" {
		t.Errorf("expected no system message for test file, got: %s", result.SystemMessage)
	}
}

func TestR14_EnforceLevelMax_GoTestFile_Skipped(t *testing.T) {
	// max + _test.go: must be excluded
	ctx := makeCtx("Write", map[string]any{"file_path": "/project/src/server_test.go"})
	ctx.TddEnforceLevel = "max"
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionApprove {
		t.Errorf("expected approve for _test.go file, got %s", result.Decision)
	}
}

func TestR14_EnforceLevelMax_Bypass(t *testing.T) {
	// max + bypass=true: returns approve with bypass SystemMessage
	ctx := makeCtx("Edit", map[string]any{"file_path": "/project/src/handler.go"})
	ctx.TddEnforceLevel = "max"
	ctx.TddBypass = true
	ctx.TddBypassReason = "hotfix: production incident"
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionApprove {
		t.Errorf("expected approve for bypass, got %s", result.Decision)
	}
	if result.SystemMessage == "" {
		t.Error("expected bypass SystemMessage to be set")
	}
}

func TestR14_EnforceLevelCentral_SrcFile_Inert(t *testing.T) {
	// central (not max): R14 must be inert
	ctx := makeCtx("Write", map[string]any{"file_path": "/project/src/main.py"})
	ctx.TddEnforceLevel = "central"
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionApprove {
		t.Errorf("expected approve when TDD enforce level is central, got %s", result.Decision)
	}
}

func TestR14_NonWriteTool_Inert(t *testing.T) {
	// R14 only fires on Write/Edit/MultiEdit; Bash must be ignored
	ctx := makeCtx("Bash", map[string]any{"command": "cat src/main.ts"})
	ctx.TddEnforceLevel = "max"
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionApprove {
		t.Errorf("expected approve for Bash tool (R14 scope is Write/Edit/MultiEdit only), got %s", result.Decision)
	}
}

func TestR14_NonSrcPath_Inert(t *testing.T) {
	// max + file outside src_patterns: R14 must not fire
	ctx := makeCtx("Write", map[string]any{"file_path": "/project/docs/readme.md"})
	ctx.TddEnforceLevel = "max"
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionApprove {
		t.Errorf("expected approve for docs file (outside src_patterns), got %s", result.Decision)
	}
}

// ---------------------------------------------------------------------------
// R03: break-glass ask-list (ProtectedPathAskList)
// ---------------------------------------------------------------------------

func makeCtxWithAskList(command string, entries []hookproto.ProtectedPathAskEntry) hookproto.RuleContext {
	ctx := makeCtx("Bash", map[string]any{"command": command})
	ctx.ProtectedPathAskList = entries
	return ctx
}

func TestR03_EnvAskListReturnsAsk(t *testing.T) {
	ctx := makeCtxWithAskList(
		"printf 'SECRET=foo\\n' > .env",
		[]hookproto.ProtectedPathAskEntry{{Path: ".env", Reason: "customer deploy env update"}},
	)
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionAsk {
		t.Errorf("expected ask for .env with ask-list, got %s: %s", result.Decision, result.Reason)
	}
	if !strings.Contains(result.Reason, "R03") {
		t.Errorf("reason should mention R03, got: %s", result.Reason)
	}
	if !strings.Contains(result.Reason, ".env") {
		t.Errorf("reason should mention .env, got: %s", result.Reason)
	}
	if !strings.Contains(result.Reason, "harness.toml") {
		t.Errorf("reason should mention harness.toml, got: %s", result.Reason)
	}
	if !strings.Contains(result.Reason, "customer deploy env update") {
		t.Errorf("reason should include configured reason, got: %s", result.Reason)
	}
	// Must NOT echo back the command content
	if strings.Contains(result.Reason, "SECRET=foo") {
		t.Errorf("reason must not echo command content, got: %s", result.Reason)
	}
}

func TestR03_EnvAskListRequiresNonEmptyReason(t *testing.T) {
	ctx := makeCtxWithAskList(
		"echo x > .env",
		[]hookproto.ProtectedPathAskEntry{{Path: ".env", Reason: "   "}},
	)
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionDeny {
		t.Errorf("expected deny when reason is whitespace-only, got %s", result.Decision)
	}
}

func TestR03_EnvAskListRequiresExactPath(t *testing.T) {
	// config has .env, command writes .env.production — must NOT match
	ctx := makeCtxWithAskList(
		"echo x > .env.production",
		[]hookproto.ProtectedPathAskEntry{{Path: ".env", Reason: "deploy env"}},
	)
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionDeny {
		t.Errorf("expected deny for .env.production when ask-list only has .env, got %s", result.Decision)
	}
}

func TestR03_EnvAskListCanMatchNarrowEnvVariant(t *testing.T) {
	// config explicitly lists .env.production
	ctx := makeCtxWithAskList(
		"echo x > .env.production",
		[]hookproto.ProtectedPathAskEntry{{Path: ".env.production", Reason: "prod deploy"}},
	)
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionAsk {
		t.Errorf("expected ask for .env.production with exact ask-list entry, got %s", result.Decision)
	}
}

func TestR03_EnvAskListDoesNotBypassOutsideProjectPath(t *testing.T) {
	for _, cmd := range []string{
		"echo x > ../other/.env",
		"echo x > /tmp/.env",
	} {
		ctx := makeCtxWithAskList(cmd, []hookproto.ProtectedPathAskEntry{{Path: ".env", Reason: "deploy env"}})
		result := EvaluateRules(ctx)
		if result.Decision != hookproto.DecisionDeny {
			t.Errorf("expected deny for out-of-project path %q, got %s", cmd, result.Decision)
		}
	}
}

func TestR03_EnvAskListDoesNotBypassHardDenyPaths(t *testing.T) {
	hardDeny := []string{
		".git/config",
		"secrets/.env",
		"certs/server.pem",
		".env.key",
		"authorized_keys",
		".zshrc",
		".claude/hooks/pre-tool.sh",
		".husky/pre-commit",
	}
	for _, target := range hardDeny {
		ctx := makeCtxWithAskList(
			"echo x > "+target,
			[]hookproto.ProtectedPathAskEntry{{Path: target, Reason: "testing"}},
		)
		result := EvaluateRules(ctx)
		if result.Decision != hookproto.DecisionDeny {
			t.Errorf("expected deny for hard-deny path %q even with ask-list, got %s", target, result.Decision)
		}
	}
}

func TestR03_EnvAskListDoesNotHideMixedHardDenyTarget(t *testing.T) {
	// command writes to both a safe target and a hard-deny target — must deny
	ctx := makeCtxWithAskList(
		"echo safe > .env && echo hard > .git/config",
		[]hookproto.ProtectedPathAskEntry{{Path: ".env", Reason: "deploy env"}},
	)
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionDeny {
		t.Errorf("expected deny for mixed command containing hard-deny target, got %s", result.Decision)
	}
}

func TestR03_SedInPlaceEnvWriteOutOfScope(t *testing.T) {
	// sed -i in-place does not use redirection — R03 break-glass must not intercept
	ctx := makeCtxWithAskList(
		"sed -i '' 's/foo/bar/' .env",
		[]hookproto.ProtectedPathAskEntry{{Path: ".env", Reason: "deploy env"}},
	)
	result := EvaluateRules(ctx)
	// sed -i is out of scope for break-glass; R03 pattern-based deny also does not match sed
	if result.Decision != hookproto.DecisionApprove {
		t.Errorf("expected approve for sed -i (out of R03 scope), got %s", result.Decision)
	}
}

func TestR02_EnvAskListDoesNotBypassWriteDeny(t *testing.T) {
	// The Write tool hitting .env must still be denied — break-glass is Bash R03 only
	ctx := makeCtx("Write", map[string]any{"file_path": ".env"})
	ctx.ProtectedPathAskList = []hookproto.ProtectedPathAskEntry{{Path: ".env", Reason: "deploy env"}}
	result := EvaluateRules(ctx)
	if result.Decision != hookproto.DecisionDeny {
		t.Errorf("expected deny for Write tool to .env regardless of ask-list, got %s", result.Decision)
	}
}
