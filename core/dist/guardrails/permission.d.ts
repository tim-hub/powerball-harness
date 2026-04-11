/**
 * core/src/guardrails/permission.ts
 * PermissionRequest hook evaluation function
 *
 * Full logic ported from permission-request.sh to TypeScript.
 * Auto-approves safe commands (read-only git, test commands, etc.).
 *
 * Reference: scripts/permission-request.sh
 */
import { type HookInput, type HookResult } from "../types.js";
/**
 * PermissionRequest hook evaluation function.
 *
 * Edit/Write are auto-approved (bypassPermissions equivalent).
 * Bash only auto-approves safe command patterns.
 * Others return nothing (default behavior = prompt the user).
 */
export declare function evaluatePermission(input: HookInput): HookResult;
/**
 * Generate stdout output for the PermissionRequest hook.
 * Called from index.ts route() for the "permission" hook type.
 */
export declare function formatPermissionOutput(result: HookResult): string;
//# sourceMappingURL=permission.d.ts.map