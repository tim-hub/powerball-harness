/**
 * core/src/guardrails/rules.ts
 * Harness v3 declarative guard rule table
 *
 * All rules from pretooluse-guard.sh ported as a type-safe declarative table.
 * Each GuardRule is a pair of condition (toolPattern + evaluate) and action (HookResult).
 */
import type { GuardRule, HookResult, RuleContext } from "../types.js";
export declare const GUARD_RULES: readonly GuardRule[];
/**
 * Evaluate all rules in order and return the HookResult of the first match.
 * Returns approve if no rules match.
 */
export declare function evaluateRules(ctx: RuleContext): HookResult;
//# sourceMappingURL=rules.d.ts.map