/**
 * core/src/guardrails/tampering.ts
 * Test tampering detection engine
 *
 * All patterns from posttooluse-tampering-detector.sh ported to TypeScript.
 * After test files or CI configuration are modified by Write / Edit / MultiEdit tools,
 * detects tampering patterns and returns warnings (does not block).
 */
import type { HookInput, HookResult } from "../types.js";
/**
 * Detect test tampering in a PostToolUse hook and return warnings.
 * Even when tampering is detected, the decision is "approve" (does not block).
 * Warnings are passed to Claude as systemMessage.
 */
export declare function detectTestTampering(input: HookInput): HookResult;
//# sourceMappingURL=tampering.d.ts.map