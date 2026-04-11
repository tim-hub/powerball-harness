/**
 * core/src/state/migration.ts
 * Harness v2 JSON / JSONL → v3 SQLite migration script
 *
 * Imports v2 state files into the v3 SQLite DB.
 * Migration targets:
 *   .claude/state/session.json      → sessions table
 *   .claude/state/session.events.jsonl → signals table (task_completed, etc.)
 *   .claude/work-active.json         → work_states table
 *
 * Idempotent design: safe to re-run if migration is already complete.
 */
import { readFileSync, existsSync, renameSync } from "node:fs";
import { resolve } from "node:path";
import { HarnessStore } from "./store.js";
// ============================================================
// Helper functions
// ============================================================
/** Normalize an ISO date string or Unix timestamp to an ISO string */
function toIsoString(value) {
    if (value === null || value === undefined) {
        return new Date().toISOString();
    }
    if (typeof value === "number") {
        // Determine if Unix timestamp is in seconds or milliseconds
        const ms = value > 1e10 ? value : value * 1000;
        return new Date(ms).toISOString();
    }
    // Already an ISO string — return as-is
    return value;
}
/** Normalize a v2 mode string to a v3 mode */
function normalizeMode(mode) {
    switch (mode) {
        case "work":
        case "codex":
        case "breezing":
            return mode;
        default:
            return "normal";
    }
}
/** Check if a string is a valid SignalType */
function normalizeSignalType(type) {
    // Valid SignalType list (synchronized with SignalType in types.ts)
    const valid = [
        "task_completed", "task_failed", "teammate_idle",
        "session_start", "session_end", "stop_failure", "request_review",
    ];
    if (type && valid.includes(type))
        return type;
    return "task_completed"; // Fallback for unknown types
}
// ============================================================
// JSON file reading utilities
// ============================================================
/** Safely read a JSON file. Returns null if it does not exist */
function readJsonFile(filePath) {
    if (!existsSync(filePath))
        return null;
    try {
        const content = readFileSync(filePath, "utf8");
        return JSON.parse(content);
    }
    catch {
        return null;
    }
}
/** Safely read a JSONL file (one JSON per line). Returns [] if it does not exist */
function readJsonlFile(filePath) {
    if (!existsSync(filePath))
        return [];
    try {
        const content = readFileSync(filePath, "utf8");
        return content
            .split("\n")
            .filter((line) => line.trim().length > 0)
            .map((line) => JSON.parse(line));
    }
    catch {
        return [];
    }
}
/**
 * Migrate v2 JSON/JSONL state files to the v3 SQLite DB.
 *
 * @param projectRoot - Project root path (default: process.cwd())
 * @param dbPath - SQLite DB path (default: <projectRoot>/.harness/state.db)
 * @returns Migration result
 */
export function migrate(projectRoot = process.cwd(), dbPath) {
    const stateDir = resolve(projectRoot, ".claude", "state");
    const resolvedDbPath = dbPath ?? resolve(projectRoot, ".harness", "state.db");
    const result = {
        sessions: 0,
        signals: 0,
        workStates: 0,
        skipped: false,
        errors: [],
    };
    const store = new HarnessStore(resolvedDbPath);
    try {
        // Skip if already migrated: check for migration_done in schema_meta
        const migrationDone = store.getMeta("migration_v1_done");
        if (migrationDone === "1") {
            result.skipped = true;
            return result;
        }
        // ------------------------------------------------
        // 1. session.json → sessions table
        // ------------------------------------------------
        const sessionFile = resolve(stateDir, "session.json");
        const v2Session = readJsonFile(sessionFile);
        if (v2Session !== null) {
            const sessionId = v2Session.session_id ?? v2Session.id ?? "migrated-session";
            try {
                store.upsertSession({
                    session_id: sessionId,
                    mode: normalizeMode(v2Session.mode),
                    project_root: v2Session.project_root ?? projectRoot,
                    started_at: toIsoString(v2Session.started_at),
                });
                if (v2Session.ended_at !== null && v2Session.ended_at !== undefined) {
                    store.endSession(sessionId);
                }
                result.sessions++;
            }
            catch (err) {
                result.errors.push(`session migration failed: ${err}`);
            }
        }
        // ------------------------------------------------
        // 2. session.events.jsonl → signals table
        // ------------------------------------------------
        const eventsFile = resolve(stateDir, "session.events.jsonl");
        const v2Events = readJsonlFile(eventsFile);
        for (const event of v2Events) {
            const type = normalizeSignalType(event.type ?? event.event);
            const fromSessionId = event.from_session_id ?? event.session_id ?? "unknown";
            const payload = event.payload ?? event.data ?? {};
            try {
                const signal = {
                    type,
                    from_session_id: fromSessionId,
                    payload,
                };
                if (event.to_session_id) {
                    signal.to_session_id = event.to_session_id;
                }
                store.sendSignal(signal);
                result.signals++;
            }
            catch (err) {
                result.errors.push(`signal migration failed (type=${type}): ${err}`);
            }
        }
        // ------------------------------------------------
        // 3. work-active.json → work_states table
        // ------------------------------------------------
        const workActiveFile = resolve(projectRoot, ".claude", "work-active.json");
        const v2WorkActive = readJsonFile(workActiveFile);
        if (v2WorkActive !== null) {
            const sessionId = v2WorkActive.session_id ?? "migrated-work-session";
            try {
                // Register a placeholder session to satisfy the FK constraint
                store.upsertSession({
                    session_id: sessionId,
                    mode: normalizeMode(v2WorkActive.mode ?? "work"),
                    project_root: projectRoot,
                    started_at: new Date().toISOString(),
                });
                store.setWorkState(sessionId, {
                    codexMode: v2WorkActive.codex_mode ?? false,
                    bypassRmRf: v2WorkActive.bypass_rm_rf ?? false,
                    bypassGitPush: v2WorkActive.bypass_git_push ?? false,
                });
                result.workStates++;
            }
            catch (err) {
                result.errors.push(`work_state migration failed: ${err}`);
            }
        }
        // ------------------------------------------------
        // 4. Record migration completion marker
        // ------------------------------------------------
        store.setMeta("migration_v1_done", "1");
        // ------------------------------------------------
        // 5. Backup original files (do not delete)
        // ------------------------------------------------
        if (v2Session !== null && existsSync(sessionFile)) {
            try {
                renameSync(sessionFile, `${sessionFile}.v2.bak`);
            }
            catch {
                // Ignore backup failures (migration itself is already complete)
            }
        }
    }
    finally {
        store.close();
    }
    return result;
}
// ============================================================
// CLI entry point (when executed directly via node)
// ============================================================
// In ESM, import.meta.url can determine "direct execution"
// After compiling to dist/, invoke with `node dist/state/migration.js`
const isMain = process.argv[1]?.endsWith("migration.js");
if (isMain) {
    const projectRoot = process.argv[2] ?? process.cwd();
    const dbPath = process.argv[3];
    const result = migrate(projectRoot, dbPath);
    if (result.skipped) {
        console.log("Migration already completed. Skipped.");
        process.exit(0);
    }
    if (result.errors.length > 0) {
        console.error("Migration completed with errors:");
        for (const err of result.errors) {
            console.error(`  - ${err}`);
        }
    }
    console.log(`Migration done: ${result.sessions} sessions, ${result.signals} signals, ${result.workStates} work_states`);
    process.exit(result.errors.length > 0 ? 1 : 0);
}
//# sourceMappingURL=migration.js.map