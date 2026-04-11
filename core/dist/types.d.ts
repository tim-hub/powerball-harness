/**
 * core/src/types.ts
 * Harness v3 common type definitions
 *
 * Defines the stdin/stdout JSON schema for Claude Code Hooks
 * and internal types for the guardrail engine.
 */
/** Input to PreToolUse / PostToolUse hooks */
export interface HookInput {
    /** Tool name about to be executed (e.g. "Bash", "Write") */
    tool_name: string;
    /** Input parameters for the tool */
    tool_input: Record<string, unknown>;
    /** Session ID (set by Claude Code) */
    session_id?: string;
    /** Current working directory */
    cwd?: string;
    /** Plugin root directory */
    plugin_root?: string;
}
/** Action returned by a hook */
export type HookDecision = "approve" | "deny" | "ask";
/** Hook output (Claude Code Hooks protocol) */
export interface HookResult {
    /** Whether to allow or deny execution */
    decision: HookDecision;
    /** Explanation message for the user */
    reason?: string;
    /** Additional context for Claude (systemMessage) */
    systemMessage?: string;
}
/** Evaluation context for a guard rule */
export interface RuleContext {
    input: HookInput;
    projectRoot: string;
    workMode: boolean;
    codexMode: boolean;
    breezingRole: string | null;
}
/** Definition of a single guard rule */
export interface GuardRule {
    /** Rule identifier (for logging and debugging) */
    id: string;
    /** Tool name pattern this rule applies to (regex) */
    toolPattern: RegExp;
    /** Function to evaluate the rule. Returns null if no match */
    evaluate: (ctx: RuleContext) => HookResult | null;
}
/** Types of signals exchanged between agents */
export type SignalType = "task_completed" | "task_failed" | "teammate_idle" | "session_start" | "session_end" | "stop_failure" | "request_review";
/** Inter-agent signal */
export interface Signal {
    type: SignalType;
    /** Source session ID */
    from_session_id: string;
    /** Destination session ID (omit for broadcast) */
    to_session_id?: string;
    /** Signal payload */
    payload: Record<string, unknown>;
    /** Timestamp (ISO 8601) */
    timestamp: string;
}
/** Severity of a task failure */
export type FailureSeverity = "warning" | "error" | "critical";
/** Task failure event */
export interface TaskFailure {
    /** Identifier of the failed task */
    task_id: string;
    /** Failure severity */
    severity: FailureSeverity;
    /** Failure description */
    message: string;
    /** Stack trace or detailed information */
    detail?: string;
    /** Failure timestamp (ISO 8601) */
    timestamp: string;
    /** Attempt number */
    attempt: number;
}
/** Session execution mode */
export type SessionMode = "normal" | "work" | "codex" | "breezing";
/** Session state */
export interface SessionState {
    session_id: string;
    mode: SessionMode;
    project_root: string;
    started_at: string;
    /** Context information for work/breezing mode */
    context?: Record<string, unknown>;
}
//# sourceMappingURL=types.d.ts.map