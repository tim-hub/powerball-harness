package guard

import (
	"os"

	"github.com/Chachamaru127/claude-code-harness/go/pkg/protocol"
)

// isTruthy checks if an env var value is truthy ("1", "true", "yes").
func isTruthy(value string) bool {
	return value == "1" || value == "true" || value == "yes"
}

// BuildContext constructs a RuleContext from a HookInput and environment variables.
// Priority: environment variables (SQLite integration deferred to Phase 1).
func BuildContext(input protocol.HookInput) protocol.RuleContext {
	projectRoot := input.CWD
	if projectRoot == "" {
		projectRoot = os.Getenv("HARNESS_PROJECT_ROOT")
	}
	if projectRoot == "" {
		projectRoot = os.Getenv("PROJECT_ROOT")
	}
	if projectRoot == "" {
		projectRoot, _ = os.Getwd()
	}

	workMode := isTruthy(os.Getenv("HARNESS_WORK_MODE")) ||
		isTruthy(os.Getenv("ULTRAWORK_MODE"))
	codexMode := isTruthy(os.Getenv("HARNESS_CODEX_MODE"))
	breezingRole := os.Getenv("HARNESS_BREEZING_ROLE")

	return protocol.RuleContext{
		Input:        input,
		ProjectRoot:  projectRoot,
		WorkMode:     workMode,
		CodexMode:    codexMode,
		BreezingRole: breezingRole,
	}
}

// EvaluatePreTool is the PreToolUse hook entry point.
// It builds the context and evaluates all guard rules.
func EvaluatePreTool(input protocol.HookInput) protocol.HookResult {
	ctx := BuildContext(input)
	return EvaluateRules(ctx)
}

// PreToolToOutput converts a HookResult to the official PreToolUse hookSpecificOutput.
func PreToolToOutput(result protocol.HookResult) *protocol.PreToolOutput {
	// Only convert deny/ask decisions to hookSpecificOutput.
	// approve with no systemMessage needs no output (exit 0 with empty stdout).
	if result.Decision == protocol.DecisionApprove && result.SystemMessage == "" {
		return nil
	}

	out := &protocol.PreToolOutput{
		HookEventName: "PreToolUse",
	}

	switch result.Decision {
	case protocol.DecisionDeny:
		out.PermissionDecision = "deny"
		out.PermissionDecisionReason = result.Reason
	case protocol.DecisionAsk:
		out.PermissionDecision = "ask"
		out.PermissionDecisionReason = result.Reason
	case protocol.DecisionApprove:
		out.PermissionDecision = "allow"
		if result.SystemMessage != "" {
			out.AdditionalContext = result.SystemMessage
		}
	}

	return out
}

// FormatPreToolResult converts a HookResult to the appropriate output for PreToolUse.
// Returns (json bytes or nil, exit code).
//   - deny → hookSpecificOutput JSON, exit 2
//   - ask → hookSpecificOutput JSON, exit 0
//   - approve with systemMessage → hookSpecificOutput JSON, exit 0
//   - approve without message → nil, exit 0
func FormatPreToolResult(result protocol.HookResult) (output interface{}, exitCode int) {
	// deny always blocks
	if result.Decision == protocol.DecisionDeny {
		return PreToolToOutput(result), 2
	}

	out := PreToolToOutput(result)
	if out != nil {
		return out, 0
	}

	// Pure approve — empty output, exit 0
	return nil, 0
}

// matchesWriteEditMultiEdit checks if tool name is Write, Edit, or MultiEdit.
func matchesWriteEditMultiEdit(toolName string) bool {
	return toolName == "Write" || toolName == "Edit" || toolName == "MultiEdit"
}

// getStringField safely extracts a string field from tool_input.
func getStringField(input map[string]interface{}, key string) (string, bool) {
	v, ok := input[key]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok && s != ""
}

// getChangedContent extracts the changed content from Write (content) or Edit (new_string).
func getChangedContent(input map[string]interface{}) string {
	if content, ok := getStringField(input, "content"); ok {
		return content
	}
	if newStr, ok := getStringField(input, "new_string"); ok {
		return newStr
	}
	return ""
}

