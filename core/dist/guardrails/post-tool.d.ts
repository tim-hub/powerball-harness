/**
 * core/src/guardrails/post-tool.ts
 * PostToolUse hook integrated evaluation function
 *
 * Runs the following PostToolUse script equivalents in parallel via Promise.allSettled
 * and aggregates the results into a single HookResult:
 *
 * 1. tampering-detector: Test tampering detection (warning only)
 * 2. security-review: Security pattern detection (warning only)
 *
 * Others (log-toolname, commit-cleanup, etc.) are side-effect-only and do not affect
 * the HookResult, so they maintain their design as separate hooks.json entries.
 */
import type { HookInput, HookResult } from "../types.js";
/**
 * PostToolUse hook entry point.
 * Runs multiple detectors in parallel and returns aggregated warnings.
 */
export declare function evaluatePostTool(input: HookInput): Promise<HookResult>;
//# sourceMappingURL=post-tool.d.ts.map