/**
 * core/src/state/schema.ts
 * Harness v3 SQLite table definitions
 *
 * Uses better-sqlite3 to manage session state, inter-agent signals,
 * and task failure events in a single SQLite file.
 */
/**
 * sessions table
 * - session_id: Session identifier issued by Claude Code
 * - mode: normal | work | codex | breezing
 * - project_root: Project root associated with the session
 * - started_at: Session start time (Unix timestamp in seconds)
 * - ended_at: Session end time (NULL = active)
 * - context_json: Arbitrary additional information (JSON text)
 */
export declare const CREATE_SESSIONS: "\n  CREATE TABLE IF NOT EXISTS sessions (\n    session_id   TEXT    NOT NULL PRIMARY KEY,\n    mode         TEXT    NOT NULL CHECK(mode IN ('normal','work','codex','breezing')),\n    project_root TEXT    NOT NULL,\n    started_at   INTEGER NOT NULL,\n    ended_at     INTEGER,\n    context_json TEXT    DEFAULT '{}'\n  )\n";
/**
 * signals table
 * - id: Auto-increment PK
 * - type: Signal type (SignalType)
 * - from_session_id: Source session
 * - to_session_id: Destination session (NULL = broadcast)
 * - payload_json: Payload (JSON text)
 * - sent_at: Send time (Unix timestamp in seconds)
 * - consumed: Consumed flag
 */
export declare const CREATE_SIGNALS: "\n  CREATE TABLE IF NOT EXISTS signals (\n    id              INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,\n    type            TEXT    NOT NULL,\n    from_session_id TEXT    NOT NULL,\n    to_session_id   TEXT,\n    payload_json    TEXT    NOT NULL DEFAULT '{}',\n    sent_at         INTEGER NOT NULL,\n    consumed        INTEGER NOT NULL DEFAULT 0 CHECK(consumed IN (0,1))\n  )\n";
/**
 * task_failures table
 * - id: Auto-increment PK
 * - task_id: Identifier of the failed task
 * - session_id: Session that was executing the task (foreign reference)
 * - severity: warning | error | critical
 * - message: Failure description
 * - detail: Stack trace or detailed information (nullable)
 * - failed_at: Failure time (Unix timestamp in seconds)
 * - attempt: Attempt count (1-based)
 */
export declare const CREATE_TASK_FAILURES: "\n  CREATE TABLE IF NOT EXISTS task_failures (\n    id         INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,\n    task_id    TEXT    NOT NULL,\n    session_id TEXT    NOT NULL,\n    severity   TEXT    NOT NULL CHECK(severity IN ('warning','error','critical')),\n    message    TEXT    NOT NULL,\n    detail     TEXT,\n    failed_at  INTEGER NOT NULL,\n    attempt    INTEGER NOT NULL DEFAULT 1 CHECK(attempt >= 1)\n  )\n";
/**
 * work_states table
 * - Successor to work-active.json. Manages state for work/codex/breezing modes
 * - session_id: Associated session ID
 * - codex_mode: Codex mode flag
 * - bypass_rm_rf: rm -rf guard bypass flag
 * - bypass_git_push: git push guard bypass flag
 * - expires_at: Expiration time (Unix timestamp in seconds, 24 hours after creation)
 */
export declare const CREATE_WORK_STATES: "\n  CREATE TABLE IF NOT EXISTS work_states (\n    session_id      TEXT    NOT NULL PRIMARY KEY,\n    codex_mode      INTEGER NOT NULL DEFAULT 0 CHECK(codex_mode IN (0,1)),\n    bypass_rm_rf    INTEGER NOT NULL DEFAULT 0 CHECK(bypass_rm_rf IN (0,1)),\n    bypass_git_push INTEGER NOT NULL DEFAULT 0 CHECK(bypass_git_push IN (0,1)),\n    expires_at      INTEGER NOT NULL,\n    FOREIGN KEY (session_id) REFERENCES sessions(session_id)\n  )\n";
export declare const CREATE_INDEXES: readonly ["CREATE INDEX IF NOT EXISTS idx_signals_to_session\n     ON signals(to_session_id, consumed)", "CREATE INDEX IF NOT EXISTS idx_signals_from_session\n     ON signals(from_session_id, sent_at)", "CREATE INDEX IF NOT EXISTS idx_task_failures_task\n     ON task_failures(task_id, failed_at)", "CREATE INDEX IF NOT EXISTS idx_work_states_expires\n     ON work_states(expires_at)"];
export declare const SCHEMA_VERSION = 1;
export declare const CREATE_SCHEMA_META: "\n  CREATE TABLE IF NOT EXISTS schema_meta (\n    key   TEXT NOT NULL PRIMARY KEY,\n    value TEXT NOT NULL\n  )\n";
/** Array of DDL statements to execute in order during DB initialization */
export declare const ALL_DDL: readonly string[];
//# sourceMappingURL=schema.d.ts.map