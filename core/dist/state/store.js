/**
 * core/src/state/store.ts
 * Harness v3 SQLite store
 *
 * Wrapper class that operates on sessions / signals / task_failures / work_states
 * tables using better-sqlite3.
 * Leverages the synchronous API (a characteristic of better-sqlite3) for a simple
 * and robust implementation.
 */
import { createRequire } from "node:module";
import { ALL_DDL, SCHEMA_VERSION, CREATE_SCHEMA_META } from "./schema.js";
// Standard pattern for loading CommonJS native addons from ESM
const require = createRequire(import.meta.url);
const Database = require("better-sqlite3");
// ============================================================
// HarnessStore class
// ============================================================
export class HarnessStore {
    db;
    constructor(dbPath) {
        this.db = new Database(dbPath);
        // WAL mode improves concurrent read performance
        this.db.pragma("journal_mode = WAL");
        this.db.pragma("foreign_keys = ON");
        this.initSchema();
    }
    // ============================================================
    // Schema initialization
    // ============================================================
    initSchema() {
        this.db.exec(CREATE_SCHEMA_META);
        const versionRow = this.db
            .prepare("SELECT value FROM schema_meta WHERE key = 'version'")
            .get();
        if (versionRow === undefined) {
            // First run: execute all DDL and record the version
            for (const ddl of ALL_DDL) {
                this.db.exec(ddl);
            }
            this.db
                .prepare("INSERT OR REPLACE INTO schema_meta(key, value) VALUES ('version', ?)")
                .run(String(SCHEMA_VERSION));
        }
        // Migrations are handled by migration.ts in the future
    }
    // ============================================================
    // Session management
    // ============================================================
    /** Register or update a session */
    upsertSession(session) {
        const startedAt = Math.floor(new Date(session.started_at).getTime() / 1000);
        const contextJson = JSON.stringify(session.context ?? {});
        this.db
            .prepare(`INSERT INTO sessions(session_id, mode, project_root, started_at, context_json)
         VALUES (?, ?, ?, ?, ?)
         ON CONFLICT(session_id) DO UPDATE SET
           mode = excluded.mode,
           project_root = excluded.project_root,
           context_json = excluded.context_json`)
            .run(session.session_id, session.mode, session.project_root, startedAt, contextJson);
    }
    /** Mark a session as ended */
    endSession(sessionId) {
        const endedAt = Math.floor(Date.now() / 1000);
        this.db
            .prepare("UPDATE sessions SET ended_at = ? WHERE session_id = ?")
            .run(endedAt, sessionId);
    }
    /** Retrieve session information */
    getSession(sessionId) {
        const row = this.db
            .prepare("SELECT * FROM sessions WHERE session_id = ?")
            .get(sessionId);
        if (row === undefined)
            return null;
        return {
            session_id: row.session_id,
            mode: row.mode,
            project_root: row.project_root,
            started_at: new Date(row.started_at * 1000).toISOString(),
            context: JSON.parse(row.context_json),
        };
    }
    // ============================================================
    // Signal management
    // ============================================================
    /** Send a signal */
    sendSignal(signal) {
        const sentAt = Math.floor(Date.now() / 1000);
        const payloadJson = JSON.stringify(signal.payload);
        const result = this.db
            .prepare(`INSERT INTO signals(type, from_session_id, to_session_id, payload_json, sent_at)
         VALUES (?, ?, ?, ?, ?)`)
            .run(signal.type, signal.from_session_id, signal.to_session_id ?? null, payloadJson, sentAt);
        return result.lastInsertRowid;
    }
    /** Receive unconsumed signals (destination = sessionId or broadcast) */
    receiveSignals(sessionId) {
        const rows = this.db
            .prepare(`SELECT * FROM signals
         WHERE consumed = 0
           AND (to_session_id = ? OR to_session_id IS NULL)
           AND from_session_id != ?
         ORDER BY sent_at ASC`)
            .all(sessionId, sessionId);
        if (rows.length === 0)
            return [];
        // Mark as consumed
        const ids = rows.map((r) => r.id);
        const placeholders = ids.map(() => "?").join(",");
        this.db
            .prepare(`UPDATE signals SET consumed = 1 WHERE id IN (${placeholders})`)
            .run(...ids);
        return rows.map((r) => {
            const signal = {
                type: r.type,
                from_session_id: r.from_session_id,
                payload: JSON.parse(r.payload_json),
                timestamp: new Date(r.sent_at * 1000).toISOString(),
            };
            if (r.to_session_id !== null) {
                signal.to_session_id = r.to_session_id;
            }
            return signal;
        });
    }
    // ============================================================
    // Task failure management
    // ============================================================
    /** Record a task failure */
    recordFailure(failure, sessionId) {
        const failedAt = Math.floor(Date.now() / 1000);
        const result = this.db
            .prepare(`INSERT INTO task_failures(task_id, session_id, severity, message, detail, failed_at, attempt)
         VALUES (?, ?, ?, ?, ?, ?, ?)`)
            .run(failure.task_id, sessionId, failure.severity, failure.message, failure.detail ?? null, failedAt, failure.attempt);
        return result.lastInsertRowid;
    }
    /** Retrieve failure history for a task */
    getFailures(taskId) {
        const rows = this.db
            .prepare("SELECT task_id, severity, message, detail, failed_at, attempt FROM task_failures WHERE task_id = ? ORDER BY failed_at ASC")
            .all(taskId);
        return rows.map((r) => {
            const failure = {
                task_id: r.task_id,
                severity: r.severity,
                message: r.message,
                timestamp: new Date(r.failed_at * 1000).toISOString(),
                attempt: r.attempt,
            };
            if (r.detail !== null) {
                failure.detail = r.detail;
            }
            return failure;
        });
    }
    // ============================================================
    // work_states management
    // ============================================================
    /** Register a work/codex mode (TTL 24 hours) */
    setWorkState(sessionId, options = {}) {
        const expiresAt = Math.floor(Date.now() / 1000) + 24 * 3600;
        this.db
            .prepare(`INSERT INTO work_states(session_id, codex_mode, bypass_rm_rf, bypass_git_push, expires_at)
         VALUES (?, ?, ?, ?, ?)
         ON CONFLICT(session_id) DO UPDATE SET
           codex_mode = excluded.codex_mode,
           bypass_rm_rf = excluded.bypass_rm_rf,
           bypass_git_push = excluded.bypass_git_push,
           expires_at = excluded.expires_at`)
            .run(sessionId, options.codexMode ? 1 : 0, options.bypassRmRf ? 1 : 0, options.bypassGitPush ? 1 : 0, expiresAt);
    }
    /** Get a valid work_state (returns null if expired) */
    getWorkState(sessionId) {
        const now = Math.floor(Date.now() / 1000);
        const row = this.db
            .prepare("SELECT * FROM work_states WHERE session_id = ? AND expires_at > ?")
            .get(sessionId, now);
        if (row === undefined)
            return null;
        return {
            codexMode: row.codex_mode === 1,
            bypassRmRf: row.bypass_rm_rf === 1,
            bypassGitPush: row.bypass_git_push === 1,
        };
    }
    /** Delete expired work_states */
    cleanExpiredWorkStates() {
        const now = Math.floor(Date.now() / 1000);
        const result = this.db
            .prepare("DELETE FROM work_states WHERE expires_at <= ?")
            .run(now);
        return result.changes;
    }
    // ============================================================
    // schema_meta key/value management
    // ============================================================
    /** Get a value from the schema_meta table (returns null if not found) */
    getMeta(key) {
        const row = this.db
            .prepare("SELECT value FROM schema_meta WHERE key = ?")
            .get(key);
        return row?.value ?? null;
    }
    /** Save a value to the schema_meta table (upsert) */
    setMeta(key, value) {
        this.db
            .prepare(`INSERT INTO schema_meta(key, value) VALUES (?, ?)
         ON CONFLICT(key) DO UPDATE SET value = excluded.value`)
            .run(key, value);
    }
    // ============================================================
    // Close
    // ============================================================
    close() {
        this.db.close();
    }
}
//# sourceMappingURL=store.js.map