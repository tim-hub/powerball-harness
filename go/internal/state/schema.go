// Package state provides SQLite state management for Harness v4.
// Ported from the TypeScript core/src/state/schema.ts to Go.
package state

// SchemaVersion is the current schema version number.
// Increment whenever a migration is needed.
const SchemaVersion = 1

// ============================================================
// DDL definitions
// ============================================================

// createSchemaMeta is the DDL for the schema version management table.
// Must be created before all other tables.
const createSchemaMeta = `
CREATE TABLE IF NOT EXISTS schema_meta (
  key   TEXT NOT NULL PRIMARY KEY,
  value TEXT NOT NULL
)`

// createSessions is the DDL for the sessions table.
// session_id: session identifier issued by Claude Code
// mode: normal | work | codex | breezing
// project_root: project root the session is associated with
// started_at: session start time (Unix timestamp seconds)
// ended_at: session end time (NULL = active)
// context_json: arbitrary additional information (JSON text)
const createSessions = `
CREATE TABLE IF NOT EXISTS sessions (
  session_id   TEXT    NOT NULL PRIMARY KEY,
  mode         TEXT    NOT NULL CHECK(mode IN ('normal','work','codex','breezing')),
  project_root TEXT    NOT NULL,
  started_at   INTEGER NOT NULL,
  ended_at     INTEGER,
  context_json TEXT    NOT NULL DEFAULT '{}'
)`

// createSignals is the DDL for the signals table.
// id: auto-increment PK
// type: signal type
// from_session_id: sending session
// to_session_id: destination session (NULL = broadcast)
// payload_json: payload (JSON text)
// sent_at: send time (Unix timestamp seconds)
// consumed: received flag (0 = unconsumed, 1 = consumed)
const createSignals = `
CREATE TABLE IF NOT EXISTS signals (
  id              INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  type            TEXT    NOT NULL,
  from_session_id TEXT    NOT NULL,
  to_session_id   TEXT,
  payload_json    TEXT    NOT NULL DEFAULT '{}',
  sent_at         INTEGER NOT NULL,
  consumed        INTEGER NOT NULL DEFAULT 0 CHECK(consumed IN (0,1))
)`

// createTaskFailures is the DDL for the task_failures table.
// id: auto-increment PK
// task_id: identifier of the failed task
// session_id: session that was executing the task
// severity: warning | error | critical
// message: description of the failure
// detail: stack trace or other details (nullable)
// failed_at: failure time (Unix timestamp seconds)
// attempt: attempt number (1-based)
const createTaskFailures = `
CREATE TABLE IF NOT EXISTS task_failures (
  id         INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  task_id    TEXT    NOT NULL,
  session_id TEXT    NOT NULL,
  severity   TEXT    NOT NULL CHECK(severity IN ('warning','error','critical')),
  message    TEXT    NOT NULL,
  detail     TEXT,
  failed_at  INTEGER NOT NULL,
  attempt    INTEGER NOT NULL DEFAULT 1 CHECK(attempt >= 1)
)`

// createWorkStates is the DDL for the work_states table.
// Successor to work-active.json. Manages state for work/codex/breezing modes.
// session_id: associated session ID (PK)
// codex_mode: codex mode flag (0/1)
// bypass_rm_rf: rm -rf guard bypass flag (0/1)
// bypass_git_push: git push guard bypass flag (0/1)
// expires_at: expiry time (Unix timestamp seconds, 24 hours from now)
// work_mode: work mode flag (0/1)
const createWorkStates = `
CREATE TABLE IF NOT EXISTS work_states (
  session_id      TEXT    NOT NULL PRIMARY KEY,
  codex_mode      INTEGER NOT NULL DEFAULT 0 CHECK(codex_mode IN (0,1)),
  bypass_rm_rf    INTEGER NOT NULL DEFAULT 0 CHECK(bypass_rm_rf IN (0,1)),
  bypass_git_push INTEGER NOT NULL DEFAULT 0 CHECK(bypass_git_push IN (0,1)),
  work_mode       INTEGER NOT NULL DEFAULT 0 CHECK(work_mode IN (0,1)),
  expires_at      INTEGER NOT NULL,
  FOREIGN KEY (session_id) REFERENCES sessions(session_id)
)`

// createAssumptions is the DDL for the assumptions table (new table).
// A table for tracking assumptions and preconditions made by agents.
// id: auto-increment PK
// session_id: session that recorded the assumption
// task_id: associated task identifier (nullable)
// assumption: assumption content (text)
// confidence: confidence level (0.0 to 1.0)
// created_at: record time (Unix timestamp seconds)
// validated_at: validation time (NULL = not yet validated)
const createAssumptions = `
CREATE TABLE IF NOT EXISTS assumptions (
  id           INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  session_id   TEXT    NOT NULL,
  task_id      TEXT,
  assumption   TEXT    NOT NULL,
  confidence   REAL    NOT NULL DEFAULT 1.0 CHECK(confidence >= 0.0 AND confidence <= 1.0),
  created_at   INTEGER NOT NULL,
  validated_at INTEGER
)`

// createAgentStates is the DDL for the table that persists agent lifecycle state.
// Recorded by SubagentStart/Stop hooks and displayed by harness status.
// agent_id: agent identifier issued by CC (PK)
// agent_type: worker | reviewer | scaffolder, etc.
// session_id: parent session identifier
// state: SPAWNING | RUNNING | REVIEWING | APPROVED | COMMITTED | FAILED |
//         CANCELLED | STALE | RECOVERING | ABORTED
// started_at: start time (Unix timestamp seconds)
// stopped_at: stop time (NULL = running)
// recovery_attempts: number of recovery attempts
const createAgentStates = `
CREATE TABLE IF NOT EXISTS agent_states (
  agent_id          TEXT    NOT NULL PRIMARY KEY,
  agent_type        TEXT    NOT NULL DEFAULT '',
  session_id        TEXT    NOT NULL DEFAULT '',
  state             TEXT    NOT NULL DEFAULT 'SPAWNING',
  started_at        INTEGER NOT NULL,
  stopped_at        INTEGER,
  recovery_attempts INTEGER NOT NULL DEFAULT 0 CHECK(recovery_attempts >= 0)
)`

// ============================================================
// Index definitions
// ============================================================

// createIndexes is the set of indexes for improving query performance.
var createIndexes = []string{
	`CREATE INDEX IF NOT EXISTS idx_signals_to_session
     ON signals(to_session_id, consumed)`,
	`CREATE INDEX IF NOT EXISTS idx_signals_from_session
     ON signals(from_session_id, sent_at)`,
	`CREATE INDEX IF NOT EXISTS idx_task_failures_task
     ON task_failures(task_id, failed_at)`,
	`CREATE INDEX IF NOT EXISTS idx_work_states_expires
     ON work_states(expires_at)`,
	`CREATE INDEX IF NOT EXISTS idx_assumptions_session
     ON assumptions(session_id, created_at)`,
	`CREATE INDEX IF NOT EXISTS idx_assumptions_task
     ON assumptions(task_id, created_at)`,
	`CREATE INDEX IF NOT EXISTS idx_agent_states_session
     ON agent_states(session_id, started_at)`,
	`CREATE INDEX IF NOT EXISTS idx_agent_states_state
     ON agent_states(state, started_at)`,
}

// ============================================================
// Initialization DDL list
// ============================================================

// allDDL is the array of DDL statements executed in order during DB initialization.
// Creates schema_meta first, then each table, and finally indexes.
var allDDL []string

func init() {
	allDDL = append(allDDL,
		createSchemaMeta,
		createSessions,
		createSignals,
		createTaskFailures,
		createWorkStates,
		createAssumptions,
		createAgentStates,
	)
	allDDL = append(allDDL, createIndexes...)
}
