/**
 * core/src/state/store.ts
 * Harness v3 SQLite store
 *
 * Wrapper class that operates on sessions / signals / task_failures / work_states
 * tables using better-sqlite3.
 * Leverages the synchronous API (a characteristic of better-sqlite3) for a simple
 * and robust implementation.
 */
import type { Signal, SessionState, TaskFailure } from "../types.js";
export declare class HarnessStore {
    private readonly db;
    constructor(dbPath: string);
    private initSchema;
    /** Register or update a session */
    upsertSession(session: SessionState): void;
    /** Mark a session as ended */
    endSession(sessionId: string): void;
    /** Retrieve session information */
    getSession(sessionId: string): SessionState | null;
    /** Send a signal */
    sendSignal(signal: Omit<Signal, "timestamp">): number;
    /** Receive unconsumed signals (destination = sessionId or broadcast) */
    receiveSignals(sessionId: string): Signal[];
    /** Record a task failure */
    recordFailure(failure: Omit<TaskFailure, "timestamp">, sessionId: string): number;
    /** Retrieve failure history for a task */
    getFailures(taskId: string): TaskFailure[];
    /** Register a work/codex mode (TTL 24 hours) */
    setWorkState(sessionId: string, options?: {
        codexMode?: boolean;
        bypassRmRf?: boolean;
        bypassGitPush?: boolean;
    }): void;
    /** Get a valid work_state (returns null if expired) */
    getWorkState(sessionId: string): {
        codexMode: boolean;
        bypassRmRf: boolean;
        bypassGitPush: boolean;
    } | null;
    /** Delete expired work_states */
    cleanExpiredWorkStates(): number;
    /** Get a value from the schema_meta table (returns null if not found) */
    getMeta(key: string): string | null;
    /** Save a value to the schema_meta table (upsert) */
    setMeta(key: string, value: string): void;
    close(): void;
}
//# sourceMappingURL=store.d.ts.map