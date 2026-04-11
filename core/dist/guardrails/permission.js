/**
 * core/src/guardrails/permission.ts
 * PermissionRequest hook evaluation function
 *
 * Full logic ported from permission-request.sh to TypeScript.
 * Auto-approves safe commands (read-only git, test commands, etc.).
 *
 * Reference: scripts/permission-request.sh
 */
import { existsSync, readFileSync } from "node:fs";
import { join } from "node:path";
function makeAllow() {
    return {
        hookSpecificOutput: {
            hookEventName: "PermissionRequest",
            decision: { behavior: "allow" },
        },
    };
}
// ============================================================
// Package manager auto-approval allowlist
// ============================================================
/**
 * If .claude/config/allowed-pkg-managers.json exists and allowed is true,
 * auto-approve npm/pnpm/yarn test/build/lint commands.
 */
function isPkgManagerAllowed(cwd) {
    const allowlistPath = join(cwd, ".claude", "config", "allowed-pkg-managers.json");
    if (!existsSync(allowlistPath))
        return false;
    try {
        const raw = readFileSync(allowlistPath, "utf-8");
        const data = JSON.parse(raw);
        if (typeof data === "object" && data !== null && "allowed" in data) {
            return data["allowed"] === true;
        }
    }
    catch {
        // Treat JSON parse errors as disallowed
    }
    return false;
}
// ============================================================
// Safe command detection
// ============================================================
/**
 * Determine whether a command string can be auto-approved.
 *
 * Security hardening:
 * - Reject commands containing pipes, redirects, variable expansion, or command substitution (conservative)
 * - Only auto-approve simple commands
 */
function isSafeCommand(command, cwd) {
    // Reject multi-line commands
    if (command.includes("\n") || command.includes("\r"))
        return false;
    // Reject if contains shell special characters (pipes, redirects, variable expansion, command substitution)
    if (/[;&|<>`$]/.test(command))
        return false;
    // Read-only git commands are always safe
    if (/^git\s+(status|diff|log|branch|rev-parse|show|ls-files)(\s|$)/i.test(command)) {
        return true;
    }
    // JS/TS test/validation commands check the package manager allowlist
    if (/^(npm|pnpm|yarn)\s+(test|run\s+(test|lint|typecheck|build|validate)|lint|typecheck|build)(\s|$)/i.test(command)) {
        return isPkgManagerAllowed(cwd);
    }
    // Python tests (no package.json risk)
    if (/^(pytest|python\s+-m\s+pytest)(\s|$)/i.test(command))
        return true;
    // Go / Rust tests
    if (/^(go\s+test|cargo\s+test)(\s|$)/i.test(command))
        return true;
    return false;
}
// ============================================================
// evaluatePermission: main export
// ============================================================
/**
 * PermissionRequest hook evaluation function.
 *
 * Edit/Write are auto-approved (bypassPermissions equivalent).
 * Bash only auto-approves safe command patterns.
 * Others return nothing (default behavior = prompt the user).
 */
export function evaluatePermission(input) {
    const toolName = input.tool_name;
    const cwd = input.cwd ?? process.cwd();
    // Auto-approve Edit / Write (bypassPermissions mode complement)
    if (toolName === "Edit" || toolName === "Write" || toolName === "MultiEdit") {
        return _permissionResponseToHookResult(makeAllow());
    }
    // Non-Bash tools use default behavior (pass through)
    if (toolName !== "Bash") {
        return { decision: "approve" };
    }
    // Bash: get the command and check safety
    const command = input.tool_input["command"];
    if (typeof command !== "string" || command.trim() === "") {
        return { decision: "approve" };
    }
    if (isSafeCommand(command, cwd)) {
        return _permissionResponseToHookResult(makeAllow());
    }
    // Unsafe commands use default behavior (defer to user confirmation)
    return { decision: "approve" };
}
/**
 * Convert a PermissionResponse to a HookResult.
 *
 * The PermissionRequest hook uses a different output format than regular HookResult,
 * but internally we treat it as a HookResult. During stdout output in index.ts route(),
 * formatPermissionOutput() converts it to the correct format.
 */
function _permissionResponseToHookResult(response) {
    return {
        decision: "approve",
        systemMessage: JSON.stringify(response),
    };
}
/**
 * Generate stdout output for the PermissionRequest hook.
 * Called from index.ts route() for the "permission" hook type.
 */
export function formatPermissionOutput(result) {
    // If systemMessage contains PermissionResponse JSON, use that preferentially
    if (result.systemMessage !== undefined) {
        try {
            const parsed = JSON.parse(result.systemMessage);
            if (typeof parsed === "object" &&
                parsed !== null &&
                "hookSpecificOutput" in parsed) {
                return JSON.stringify(parsed);
            }
        }
        catch {
            // On parse failure, output as a regular HookResult
        }
    }
    return JSON.stringify(result);
}
//# sourceMappingURL=permission.js.map