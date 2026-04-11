/**
 * core/src/guardrails/pre-tool.ts
 * PreToolUse hook evaluation function
 *
 * Receives HookInput, evaluates the declarative guard rule table in rules.ts,
 * and returns an approve / deny / ask HookResult.
 */
import { existsSync } from "node:fs";
import { resolve } from "node:path";
import { evaluateRules } from "./rules.js";
import { HarnessStore } from "../state/store.js";
/** Check if an environment variable has a truthy value ("1", "true", "yes") */
function isTruthy(value) {
    return value === "1" || value === "true" || value === "yes";
}
/**
 * Resolve the SQLite DB path from the project root.
 * Returns null if .harness/state.db does not exist.
 */
function resolveDbPath(projectRoot) {
    const dbPath = resolve(projectRoot, ".harness", "state.db");
    return existsSync(dbPath) ? dbPath : null;
}
/**
 * Build RuleContext from the execution environment.
 * Priority: SQLite work_states > environment variables
 */
function buildContext(input) {
    // cwd is the project root. plugin_root is the plugin's own path, so excluded
    const projectRoot = input.cwd ??
        process.env["HARNESS_PROJECT_ROOT"] ??
        process.env["PROJECT_ROOT"] ??
        process.cwd();
    // Initial values from environment variables
    let workMode = isTruthy(process.env["HARNESS_WORK_MODE"]) ||
        isTruthy(process.env["ULTRAWORK_MODE"]);
    let codexMode = isTruthy(process.env["HARNESS_CODEX_MODE"]);
    // Breezing role: obtained from environment variable
    const breezingRole = process.env["HARNESS_BREEZING_ROLE"] ?? null;
    // Supplement from SQLite work_states (if session_id is available)
    const sessionId = input.session_id;
    if (sessionId) {
        const dbPath = resolveDbPath(projectRoot);
        if (dbPath !== null) {
            try {
                const store = new HarnessStore(dbPath);
                try {
                    const state = store.getWorkState(sessionId);
                    if (state !== null) {
                        // Override env vars with DB values (more reliable)
                        workMode = workMode || state.bypassRmRf || state.bypassGitPush;
                        codexMode = codexMode || state.codexMode;
                    }
                }
                finally {
                    store.close();
                }
            }
            catch {
                // Ignore DB access failures (fall back to environment variables)
            }
        }
    }
    return {
        input,
        projectRoot,
        workMode,
        codexMode,
        breezingRole,
    };
}
/**
 * PreToolUse hook entry point.
 * Receives HookInput, evaluates guard rules, and returns a HookResult.
 */
export function evaluatePreTool(input) {
    const ctx = buildContext(input);
    return evaluateRules(ctx);
}
//# sourceMappingURL=pre-tool.js.map