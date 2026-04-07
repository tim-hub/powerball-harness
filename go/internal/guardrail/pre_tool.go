package guardrail

import (
	"os"

	"github.com/Chachamaru127/claude-code-harness/go/internal/state"
	"github.com/Chachamaru127/claude-code-harness/go/pkg/hookproto"
)

// isTruthy checks if an env var value is truthy ("1", "true", "yes").
func isTruthy(value string) bool {
	return value == "1" || value == "true" || value == "yes"
}

// BuildContext constructs a RuleContext from a HookInput and environment variables.
// Priority:
//  1. Environment variables (explicit overrides)
//  2. SQLite state DB (session-level state: codex_mode, work_mode)
//  3. Defaults (false / empty)
//
// The SQLite lookup is best-effort: any DB error is silently ignored so that
// the hook fast-path remains available even when the DB is unreachable.
func BuildContext(input hookproto.HookInput) hookproto.RuleContext {
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

	// 環境変数ベースの値（明示的なオーバーライド）
	workMode := isTruthy(os.Getenv("HARNESS_WORK_MODE")) ||
		isTruthy(os.Getenv("ULTRAWORK_MODE"))
	codexMode := isTruthy(os.Getenv("HARNESS_CODEX_MODE"))
	breezingRole := os.Getenv("HARNESS_BREEZING_ROLE")

	// SQLite から work_states を補完する（セッション ID がある場合のみ）
	// フック高速パスの制約（SPEC.md §12）に従い、I/O エラーは無視する。
	if input.SessionID != "" && !workMode && !codexMode {
		dbPath := state.ResolveStatePath(projectRoot)
		if ws, err := loadWorkStateFromDB(dbPath, input.SessionID); err == nil && ws != nil {
			if ws.CodexMode {
				codexMode = true
			}
			if ws.WorkMode {
				workMode = true
			}
		}
	}

	return hookproto.RuleContext{
		Input:        input,
		ProjectRoot:  projectRoot,
		WorkMode:     workMode,
		CodexMode:    codexMode,
		BreezingRole: breezingRole,
	}
}

// loadWorkStateFromDB は指定した DB パスから work_state を取得する。
// DB が存在しない・読み取れない場合は (nil, nil) を返す（エラーを伝播させない）。
// これにより hooks の fast-path がファイルシステムの問題で止まることを防ぐ。
func loadWorkStateFromDB(dbPath, sessionID string) (*state.WorkState, error) {
	// DB ファイルが存在しない場合は開かない（スロースタートの防止）
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil, nil
	}

	store, err := state.NewHarnessStore(dbPath)
	if err != nil {
		return nil, nil //nolint:nilerr // best-effort: DB エラーを伝播させない
	}
	defer store.Close()

	ws, err := store.GetWorkState(sessionID)
	if err != nil {
		return nil, nil //nolint:nilerr // best-effort
	}

	return ws, nil
}

// EvaluatePreTool is the PreToolUse hook entry point.
// It builds the context and evaluates all guard rules.
func EvaluatePreTool(input hookproto.HookInput) hookproto.HookResult {
	ctx := BuildContext(input)
	return EvaluateRules(ctx)
}

// PreToolToOutput converts a HookResult to the official PreToolUse hookSpecificOutput.
func PreToolToOutput(result hookproto.HookResult) *hookproto.PreToolOutput {
	// Only convert deny/ask decisions to hookSpecificOutput.
	// approve with no systemMessage needs no output (exit 0 with empty stdout).
	if result.Decision == hookproto.DecisionApprove && result.SystemMessage == "" {
		return nil
	}

	inner := hookproto.PreToolHookSpecific{
		HookEventName: "PreToolUse",
	}

	switch result.Decision {
	case hookproto.DecisionDeny:
		inner.PermissionDecision = "deny"
		inner.PermissionDecisionReason = result.Reason
	case hookproto.DecisionAsk:
		inner.PermissionDecision = "ask"
		inner.PermissionDecisionReason = result.Reason
	case hookproto.DecisionApprove:
		inner.PermissionDecision = "allow"
		if result.SystemMessage != "" {
			inner.AdditionalContext = result.SystemMessage
		}
	}

	return &hookproto.PreToolOutput{HookSpecificOutput: inner}
}

// FormatPreToolResult converts a HookResult to the appropriate output for PreToolUse.
// Returns (json bytes or nil, exit code).
//   - deny → hookSpecificOutput JSON, exit 2
//   - ask → hookSpecificOutput JSON, exit 0
//   - approve with systemMessage → hookSpecificOutput JSON, exit 0
//   - approve without message → nil, exit 0
func FormatPreToolResult(result hookproto.HookResult) (output interface{}, exitCode int) {
	// deny always blocks
	if result.Decision == hookproto.DecisionDeny {
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

