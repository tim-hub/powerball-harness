package guardrail

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/tim-hub/powerball-harness/go/pkg/hookproto"
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
// Backslash-escape attack detection (CC 2.1.98 bypass mitigation)
// ---------------------------------------------------------------------------

// backslashEscapePattern matches backslash followed by a flag character or space.
// This detects "git\ status", "git\--force", "rm\ -rf" attack vectors.
var backslashEscapePattern = regexp.MustCompile(`\\[\-\s]`)

// hasBackslashEscape returns true if the command contains backslash-escaped
// flag characters or spaces, which are used to bypass pattern matching.
func hasBackslashEscape(cmd string) bool {
	return backslashEscapePattern.MatchString(cmd)
}

// ---------------------------------------------------------------------------
// Env-var prefix bypass detection (CC 2.1.98 bypass mitigation)
// ---------------------------------------------------------------------------

// knownSafeEnvVars is the allowlist of environment variable names that are
// permitted as prefix assignments (e.g. "LANG=C git status").
// Variables prefixed with LC_ (LC_ALL, LC_CTYPE, etc.) are also permitted.
var knownSafeEnvVars = map[string]bool{
	"LANG":        true,
	"LANGUAGE":    true,
	"TZ":          true,
	"NO_COLOR":    true,
	"FORCE_COLOR": true,
}

// isValidEnvVarName checks if s is a valid shell identifier (ASCII letters,
// digits, underscore, must start with letter or underscore).
func isValidEnvVarName(s string) bool {
	if len(s) == 0 {
		return false
	}
	for i, c := range s {
		if i == 0 {
			if !unicode.IsLetter(c) && c != '_' {
				return false
			}
		} else {
			if !unicode.IsLetter(c) && !unicode.IsDigit(c) && c != '_' {
				return false
			}
		}
	}
	return true
}

// stripSafeEnvPrefix strips known-safe "VAR=value" prefix tokens from the
// command and returns the remaining command. Returns ("", false) if any
// unknown environment variable is found in the prefix.
//
// Examples:
//   - "LANG=C git status"       → ("git status", true)
//   - "EVIL=x git status"       → ("", false)
//   - "LANG=C EVIL=x git log"   → ("", false)
//   - "git status"               → ("git status", true)
func stripSafeEnvPrefix(cmd string) (string, bool) {
	tokens := strings.Fields(cmd)
	i := 0
	for i < len(tokens) {
		token := tokens[i]
		eqIdx := strings.IndexByte(token, '=')
		if eqIdx <= 0 {
			// Not a VAR=value token; stop consuming prefix
			break
		}
		varName := token[:eqIdx]
		varValue := token[eqIdx+1:]

		// Reject empty values (VAR= without value)
		if varValue == "" {
			return "", false
		}

		// Reject invalid env var names (must be ASCII identifier)
		if !isValidEnvVarName(varName) {
			return "", false
		}

		// Check against allowlist: known safe vars or LC_ prefix
		if !knownSafeEnvVars[varName] && !strings.HasPrefix(varName, "LC_") {
			return "", false
		}
		i++
	}

	if i == 0 {
		// No prefix tokens consumed; return cmd as-is
		return cmd, true
	}

	// Rejoin remaining tokens
	if i >= len(tokens) {
		// All tokens were env var prefixes, no actual command remains
		return "", false
	}
	return strings.Join(tokens[i:], " "), true
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
	// Backslash-escaped flags/spaces are an attack vector; reject immediately.
	if hasBackslashEscape(command) {
		return false
	}

	// Strip known-safe env-var prefixes. Unknown prefixes (e.g. EVIL=x) are rejected.
	stripped, ok := stripSafeEnvPrefix(command)
	if !ok {
		return false
	}
	// Evaluate the rest of the command against safe patterns.
	command = stripped

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
func EvaluatePermission(input hookproto.HookInput) (hookproto.HookResult, *hookproto.PermissionOutput) {
	toolName := input.ToolName
	cwd := input.CWD
	if cwd == "" {
		cwd, _ = os.Getwd()
	}

	// Edit/Write/MultiEdit are auto-allowed
	if matchesWriteEditMultiEdit(toolName) {
		return hookproto.HookResult{Decision: hookproto.DecisionApprove},
			makePermissionAllow()
	}

	// Only Bash gets further evaluation
	if toolName != "Bash" {
		return hookproto.HookResult{Decision: hookproto.DecisionApprove}, nil
	}

	command, ok := input.ToolInput["command"].(string)
	if !ok || command == "" {
		return hookproto.HookResult{Decision: hookproto.DecisionApprove}, nil
	}

	if isSafeCommand(command, cwd) {
		return hookproto.HookResult{Decision: hookproto.DecisionApprove},
			makePermissionAllow()
	}

	// Unsafe command — pass through to user
	return hookproto.HookResult{Decision: hookproto.DecisionApprove}, nil
}

func makePermissionAllow() *hookproto.PermissionOutput {
	return &hookproto.PermissionOutput{
		HookSpecificOutput: hookproto.PermissionHookSpecific{
			HookEventName: "PermissionRequest",
			Decision: hookproto.PermissionDecisionBehavior{
				Behavior: "allow",
			},
		},
	}
}
