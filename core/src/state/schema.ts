/**
 * core/src/state/schema.ts
 * Harness v3 SQLite table definitions
 *
 * Uses better-sqlite3 to manage session state, inter-agent signals,
 * and task failure events in a single SQLite file.
 */

// ============================================================
// Table creation DDL
// ============================================================

/**
 * sessions table
 * - session_id: Session identifier issued by Claude Code
 * - mode: normal | work | codex | breezing
 * - project_root: Project root associated with the session
 * - started_at: Session start time (Unix timestamp in seconds)
 * - ended_at: Session end time (NULL = active)
 * - context_json: Arbitrary additional information (JSON text)
 */
export const CREATE_SESSIONS = `
  CREATE TABLE IF NOT EXISTS sessions (
    session_id   TEXT    NOT NULL PRIMARY KEY,
    mode         TEXT    NOT NULL CHECK(mode IN ('normal','work','codex','breezing')),
    project_root TEXT    NOT NULL,
    started_at   INTEGER NOT NULL,
    ended_at     INTEGER,
    context_json TEXT    DEFAULT '{}'
  )
` as const;

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
export const CREATE_SIGNALS = `
  CREATE TABLE IF NOT EXISTS signals (
    id              INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    type            TEXT    NOT NULL,
    from_session_id TEXT    NOT NULL,
    to_session_id   TEXT,
    payload_json    TEXT    NOT NULL DEFAULT '{}',
    sent_at         INTEGER NOT NULL,
    consumed        INTEGER NOT NULL DEFAULT 0 CHECK(consumed IN (0,1))
  )
` as const;

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
export const CREATE_TASK_FAILURES = `
  CREATE TABLE IF NOT EXISTS task_failures (
    id         INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    task_id    TEXT    NOT NULL,
    session_id TEXT    NOT NULL,
    severity   TEXT    NOT NULL CHECK(severity IN ('warning','error','critical')),
    message    TEXT    NOT NULL,
    detail     TEXT,
    failed_at  INTEGER NOT NULL,
    attempt    INTEGER NOT NULL DEFAULT 1 CHECK(attempt >= 1)
  )
` as const;

/**
 * work_states table
 * - Successor to work-active.json. Manages state for work/codex/breezing modes
 * - session_id: Associated session ID
 * - codex_mode: Codex mode flag
 * - bypass_rm_rf: rm -rf guard bypass flag
 * - bypass_git_push: git push guard bypass flag
 * - expires_at: Expiration time (Unix timestamp in seconds, 24 hours after creation)
 */
export const CREATE_WORK_STATES = `
  CREATE TABLE IF NOT EXISTS work_states (
    session_id      TEXT    NOT NULL PRIMARY KEY,
    codex_mode      INTEGER NOT NULL DEFAULT 0 CHECK(codex_mode IN (0,1)),
    bypass_rm_rf    INTEGER NOT NULL DEFAULT 0 CHECK(bypass_rm_rf IN (0,1)),
    bypass_git_push INTEGER NOT NULL DEFAULT 0 CHECK(bypass_git_push IN (0,1)),
    expires_at      INTEGER NOT NULL,
    FOREIGN KEY (session_id) REFERENCES sessions(session_id)
  )
` as const;

// ============================================================
// Indexes
// ============================================================

export const CREATE_INDEXES = [
  `CREATE INDEX IF NOT EXISTS idx_signals_to_session
     ON signals(to_session_id, consumed)`,
  `CREATE INDEX IF NOT EXISTS idx_signals_from_session
     ON signals(from_session_id, sent_at)`,
  `CREATE INDEX IF NOT EXISTS idx_task_failures_task
     ON task_failures(task_id, failed_at)`,
  `CREATE INDEX IF NOT EXISTS idx_work_states_expires
     ON work_states(expires_at)`,
] as const;

// ============================================================
// Schema version management
// ============================================================

export const SCHEMA_VERSION = 1;

export const CREATE_SCHEMA_META = `
  CREATE TABLE IF NOT EXISTS schema_meta (
    key   TEXT NOT NULL PRIMARY KEY,
    value TEXT NOT NULL
  )
` as const;

// ============================================================
// Export: DDL list to execute during initialization
// ============================================================

/** Array of DDL statements to execute in order during DB initialization */
export const ALL_DDL: readonly string[] = [
  CREATE_SCHEMA_META,
  CREATE_SESSIONS,
  CREATE_SIGNALS,
  CREATE_TASK_FAILURES,
  CREATE_WORK_STATES,
  ...CREATE_INDEXES,
];
