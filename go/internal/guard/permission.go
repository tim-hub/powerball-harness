package guard

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"

	"github.com/Chachamaru127/claude-code-harness/go/pkg/protocol"
)

// ---------------------------------------------------------------------------
// Package manager allowlist
// ---------------------------------------------------------------------------

func isPkgManagerAllowed(cwd string) bool {
	allowlistPath := filepath.Join(cwd, ".claude", "config", "allowed-pkg-managers.json")
	data, err := os.ReadFile(allowlistPath)
	if err != nil {
		return false
	}
	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return false
	}
	allowed, ok := obj["allowed"]
	if !ok {
		return false
	}
	b, ok := allowed.(bool)
	return ok && b
}

// ---------------------------------------------------------------------------
// Safe command patterns
// ---------------------------------------------------------------------------

var (
	safeGitPattern  = regexp.MustCompile(`(?i)^git\s+(status|diff|log|branch|rev-parse|show|ls-files)(\s|$)`)
	safePkgPattern  = regexp.MustCompile(`(?i)^(npm|pnpm|yarn)\s+(test|run\s+(test|lint|typecheck|build|validate)|lint|typecheck|build)(\s|$)`)
	safePytest      = regexp.MustCompile(`(?i)^(pytest|python\s+-m\s+pytest)(\s|$)`)
	safeGoRustTests = regexp.MustCompile(`(?i)^(go\s+test|cargo\s+test)(\s|$)`)
	shellSpecials   = regexp.MustCompile(`[;&|<>` + "`" + `$]`)
)

func isSafeCommand(command, cwd string) bool {
	// Multiline commands are not safe
	for _, c := range command {
		if c == '\n' || c == '\r' {
			return false
		}
	}

	// Shell special chars disqualify
	if shellSpecials.MatchString(command) {
		return false
	}

	// Read-only git is always safe
	if safeGitPattern.MatchString(command) {
		return true
	}

	// JS/TS test/build commands require package manager allowlist
	if safePkgPattern.MatchString(command) {
		return isPkgManagerAllowed(cwd)
	}

	// Python test
	if safePytest.MatchString(command) {
		return true
	}

	// Go / Rust test
	if safeGoRustTests.MatchString(command) {
		return true
	}

	return false
}

// ---------------------------------------------------------------------------
// EvaluatePermission — PermissionRequest hook entry point
// ---------------------------------------------------------------------------

// EvaluatePermission evaluates a PermissionRequest hook.
//
//   - Edit/Write/MultiEdit → auto-allow (bypassPermissions complement)
//   - Bash safe commands → auto-allow
//   - Everything else → pass through (user prompted by CC)
func EvaluatePermission(input protocol.HookInput) (protocol.HookResult, *protocol.PermissionOutput) {
	toolName := input.ToolName
	cwd := input.CWD
	if cwd == "" {
		cwd, _ = os.Getwd()
	}

	// Edit/Write/MultiEdit are auto-allowed
	if matchesWriteEditMultiEdit(toolName) {
		return protocol.HookResult{Decision: protocol.DecisionApprove},
			makePermissionAllow()
	}

	// Only Bash gets further evaluation
	if toolName != "Bash" {
		return protocol.HookResult{Decision: protocol.DecisionApprove}, nil
	}

	command, ok := input.ToolInput["command"].(string)
	if !ok || command == "" {
		return protocol.HookResult{Decision: protocol.DecisionApprove}, nil
	}

	if isSafeCommand(command, cwd) {
		return protocol.HookResult{Decision: protocol.DecisionApprove},
			makePermissionAllow()
	}

	// Unsafe command — pass through to user
	return protocol.HookResult{Decision: protocol.DecisionApprove}, nil
}

func makePermissionAllow() *protocol.PermissionOutput {
	return &protocol.PermissionOutput{
		HookSpecificOutput: protocol.PermissionHookSpecific{
			HookEventName: "PermissionRequest",
			Decision: protocol.PermissionDecisionBehavior{
				Behavior: "allow",
			},
		},
	}
}
