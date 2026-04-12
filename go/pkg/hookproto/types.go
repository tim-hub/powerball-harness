// Package protocol defines the Claude Code hooks protocol types.
//
// These types conform to the official Claude Code plugin hooks specification:
// https://code.claude.com/docs/en/hooks
package hookproto

import "encoding/json"

// ---------------------------------------------------------------------------
// Hook Input (stdin JSON from Claude Code)
// ---------------------------------------------------------------------------

// HookInput represents the JSON payload sent to hooks via stdin.
// Fields match the official Claude Code hooks hookproto.
type HookInput struct {
	SessionID      string                 `json:"session_id,omitempty"`
	TranscriptPath string                 `json:"transcript_path,omitempty"`
	CWD            string                 `json:"cwd,omitempty"`
	PermissionMode string                 `json:"permission_mode,omitempty"`
	HookEventName  string                 `json:"hook_event_name,omitempty"`
	ToolName       string                 `json:"tool_name"`
	ToolInput      map[string]interface{} `json:"tool_input"`

	// Harness extension fields
	PluginRoot string `json:"plugin_root,omitempty"`
}

// ---------------------------------------------------------------------------
// Hook Output — generic
// ---------------------------------------------------------------------------

// HookDecision represents an action returned by a hook.
type HookDecision string

const (
	DecisionApprove HookDecision = "approve"
	DecisionDeny    HookDecision = "deny"
	DecisionAsk     HookDecision = "ask"
	DecisionDefer   HookDecision = "defer"
)

// HookResult is the generic hook output (legacy format).
type HookResult struct {
	Decision      HookDecision `json:"decision"`
	Reason        string       `json:"reason,omitempty"`
	SystemMessage string       `json:"systemMessage,omitempty"`
}

// ---------------------------------------------------------------------------
// PreToolUse hookSpecificOutput (official protocol)
// ---------------------------------------------------------------------------

// PreToolHookSpecific is the inner hookSpecificOutput for PreToolUse events.
type PreToolHookSpecific struct {
	HookEventName            string          `json:"hookEventName"`
	PermissionDecision       string          `json:"permissionDecision"`
	PermissionDecisionReason string          `json:"permissionDecisionReason,omitempty"`
	UpdatedInput             json.RawMessage `json:"updatedInput,omitempty"`
	AdditionalContext        string          `json:"additionalContext,omitempty"`
}

// PreToolOutput wraps PreToolHookSpecific with the required hookSpecificOutput key.
type PreToolOutput struct {
	HookSpecificOutput PreToolHookSpecific `json:"hookSpecificOutput"`
}

// ---------------------------------------------------------------------------
// PostToolUse hookSpecificOutput (documented fields only)
// ---------------------------------------------------------------------------

// PostToolHookSpecific is the inner hookSpecificOutput for PostToolUse events.
type PostToolHookSpecific struct {
	HookEventName    string `json:"hookEventName"`
	AdditionalContext string `json:"additionalContext,omitempty"`
}

// PostToolOutput wraps PostToolHookSpecific with the required hookSpecificOutput key.
type PostToolOutput struct {
	HookSpecificOutput PostToolHookSpecific `json:"hookSpecificOutput"`
}

// ---------------------------------------------------------------------------
// PermissionRequest hookSpecificOutput (official protocol)
// ---------------------------------------------------------------------------

// PermissionDecisionBehavior is the behavior field in PermissionRequest output.
type PermissionDecisionBehavior struct {
	Behavior string `json:"behavior"` // "allow" or "deny"
}

// PermissionHookSpecific is the hookSpecificOutput for PermissionRequest events.
type PermissionHookSpecific struct {
	HookEventName string                     `json:"hookEventName"`
	Decision      PermissionDecisionBehavior `json:"decision"`
}

// PermissionOutput wraps PermissionHookSpecific for the full response.
type PermissionOutput struct {
	HookSpecificOutput PermissionHookSpecific `json:"hookSpecificOutput"`
}

// ---------------------------------------------------------------------------
// Guard rule context (used by the guardrail engine)
// ---------------------------------------------------------------------------

// RuleContext carries the evaluation context for guard rules.
type RuleContext struct {
	Input        HookInput
	ProjectRoot  string
	WorkMode     bool
	CodexMode    bool
	BreezingRole string // "" means not in breezing mode
}
