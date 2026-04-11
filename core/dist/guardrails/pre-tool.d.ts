/**
 * core/src/guardrails/pre-tool.ts
 * PreToolUse hook evaluation function
 *
 * Receives HookInput, evaluates the declarative guard rule table in rules.ts,
 * and returns an approve / deny / ask HookResult.
 */
import { type HookInput, type HookResult } from "../types.js";
/**
 * PreToolUse hook entry point.
 * Receives HookInput, evaluates guard rules, and returns a HookResult.
 */
export declare function evaluatePreTool(input: HookInput): HookResult;
//# sourceMappingURL=pre-tool.d.ts.map