// Package state provides SQLite state management for Harness v4.
// Ported from the TypeScript core/src/state/store.ts to Go.
//
// Driver used: modernc.org/sqlite (pure Go, no CGO required)
// WAL mode improves parallel read throughput; busy timeout of 5s reduces lock contention.
package state

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite" // Register SQLite driver (side-effect only)
)

// ============================================================
// Type definitions
// ============================================================

// SessionMode represents the operating mode of a session.
type SessionMode string

const (
	SessionModeNormal   SessionMode = "normal"
	SessionModeWork     SessionMode = "work"
	SessionModeCodex    SessionMode = "codex"
	SessionModeBreezing SessionMode = "breezing"
)

// SessionState represents the state of a session.
type SessionState struct {
	SessionID   string                 `json:"session_id"`
	Mode        SessionMode            `json:"mode"`
	ProjectRoot string                 `json:"project_root"`
	StartedAt   string                 `json:"started_at"` // ISO 8601
	EndedAt     *string                `json:"ended_at,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
}

// Signal represents an inter-agent signal.
type Signal struct {
	ID            int64                  `json:"id,omitempty"`
	Type          string                 `json:"type"`
	FromSessionID string                 `json:"from_session_id"`
	ToSessionID   *string                `json:"to_session_id,omitempty"`
	Payload       map[string]interface{} `json:"payload"`
	Timestamp     string                 `json:"timestamp"` // ISO 8601
}

// TaskFailure represents a task failure event.
type TaskFailure struct {
	TaskID    string  `json:"task_id"`
	Severity  string  `json:"severity"` // warning | error | critical
	Message   string  `json:"message"`
	Detail    *string `json:"detail,omitempty"`
	Timestamp string  `json:"timestamp"` // ISO 8601
	Attempt   int     `json:"attempt"`
}

// WorkState represents the state of work/codex mode.
type WorkState struct {
	SessionID      string `json:"session_id"`
	CodexMode      bool   `json:"codex_mode"`
	BypassRmRf     bool   `json:"bypass_rm_rf"`
	BypassGitPush  bool   `json:"bypass_git_push"`
	WorkMode       bool   `json:"work_mode"`
	ExpiresAt      int64  `json:"expires_at"`
}

// Assumption represents an assumption or precondition recorded by an agent.
type Assumption struct {
	ID          int64    `json:"id,omitempty"`
	SessionID   string   `json:"session_id"`
	TaskID      *string  `json:"task_id,omitempty"`
	Assumption  string   `json:"assumption"`
	Confidence  float64  `json:"confidence"`
	CreatedAt   string   `json:"created_at"`   // ISO 8601
	ValidatedAt *string  `json:"validated_at,omitempty"`
}

// WorkStateOptions is the option argument for SetWorkState.
type WorkStateOptions struct {
	CodexMode     bool
	BypassRmRf    bool
	BypassGitPush bool
	WorkMode      bool
}

// ============================================================
// Path resolution
// ============================================================

// ResolveStatePath resolves the state DB path according to priority order.
// Priority:
//  1. ${CLAUDE_PLUGIN_DATA}/state.db
//  2. ${PROJECT_ROOT}/.harness/state.db
//  3. Fallback: ${PROJECT_ROOT}/.claude/state/state.db (Open is attempted even though read-only is expected)
func ResolveStatePath(projectRoot string) string {
	// 1. Plugin data directory with persistence guarantees (CC v2.1.78+)
	if pluginData := os.Getenv("CLAUDE_PLUGIN_DATA"); pluginData != "" {
		return filepath.Join(pluginData, "state.db")
	}
	// 2. Project-local .harness directory
	if projectRoot != "" {
		return filepath.Join(projectRoot, ".harness", "state.db")
	}
	// 3. Fallback: relative to current directory
	cwd, err := os.Getwd()
	if err != nil {
		return ".harness/state.db"
	}
	return filepath.Join(cwd, ".harness", "state.db")
}

// ============================================================
// HarnessStore
// ============================================================

// HarnessStore is the Harness SQLite state store.
// A 1:1 Go port of the TypeScript HarnessStore class.
type HarnessStore struct {
	db *sql.DB
}

// NewHarnessStore opens a SQLite DB at the given path, initializes the schema,
// and returns a HarnessStore.
// Sets WAL mode and busy timeout of 5s to reduce lock contention.
func NewHarnessStore(dbPath string) (*HarnessStore, error) {
	// Create parent directory for the DB file if it does not exist
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, fmt.Errorf("state: mkdir %s: %w", filepath.Dir(dbPath), err)
	}

	// Open modernc SQLite via database/sql
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("state: open db %s: %w", dbPath, err)
	}

	// Limit connection count to match SQLite file-level locking.
	// WAL mode allows parallel reads, but writes must be serialized.
	db.SetMaxOpenConns(1)

	// PRAGMAs cannot use placeholders, so execute them directly via Exec.
	if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("state: set WAL mode: %w", err)
	}
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("state: enable foreign_keys: %w", err)
	}
	// Set lock wait timeout to 5 seconds (SPEC.md §6)
	if _, err := db.Exec("PRAGMA busy_timeout = 5000"); err != nil {
		db.Close()
		return nil, fmt.Errorf("state: set busy_timeout: %w", err)
	}

	store := &HarnessStore{db: db}
	if err := store.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("state: init schema: %w", err)
	}

	return store, nil
}

// Close closes the DB connection.
func (s *HarnessStore) Close() error {
	return s.db.Close()
}

// ============================================================
// Schema initialization
// ============================================================

// initSchema executes all DDL on first startup and records the version in schema_meta.
// If a version is already recorded, this is a no-op (migrations are a future concern).
func (s *HarnessStore) initSchema() error {
	// Always create schema_meta first
	if _, err := s.db.Exec(createSchemaMeta); err != nil {
		return fmt.Errorf("create schema_meta: %w", err)
	}

	// Version check
	var version string
	err := s.db.QueryRow("SELECT value FROM schema_meta WHERE key = 'version'").Scan(&version)

	switch {
	case err == sql.ErrNoRows:
		// First run: execute all DDL and record the version
		for _, ddl := range allDDL {
			if _, execErr := s.db.Exec(ddl); execErr != nil {
				return fmt.Errorf("exec ddl: %w", execErr)
			}
		}
		if _, execErr := s.db.Exec(
			"INSERT OR REPLACE INTO schema_meta(key, value) VALUES ('version', ?)",
			fmt.Sprintf("%d", SchemaVersion),
		); execErr != nil {
			return fmt.Errorf("set schema version: %w", execErr)
		}
	case err != nil:
		return fmt.Errorf("read schema version: %w", err)
	default:
		// Version already recorded — migrations are a future concern
		_ = version
	}

	return nil
}

// ============================================================
// Session management
// ============================================================

// UpsertSession registers or updates a session.
// Corresponds to TypeScript's upsertSession.
func (s *HarnessStore) UpsertSession(session SessionState) error {
	// Convert ISO 8601 string to Unix timestamp seconds
	startedAt, err := parseISOToUnix(session.StartedAt)
	if err != nil {
		return fmt.Errorf("parse started_at: %w", err)
	}

	contextJSON := "{}"
	if session.Context != nil {
		b, encErr := json.Marshal(session.Context)
		if encErr != nil {
			return fmt.Errorf("marshal context: %w", encErr)
		}
		contextJSON = string(b)
	}

	_, err = s.db.Exec(
		`INSERT INTO sessions(session_id, mode, project_root, started_at, context_json)
         VALUES (?, ?, ?, ?, ?)
         ON CONFLICT(session_id) DO UPDATE SET
           mode = excluded.mode,
           project_root = excluded.project_root,
           context_json = excluded.context_json`,
		session.SessionID,
		string(session.Mode),
		session.ProjectRoot,
		startedAt,
		contextJSON,
	)
	return err
}

// EndSession marks a session as ended.
// Corresponds to TypeScript's endSession.
func (s *HarnessStore) EndSession(sessionID string) error {
	endedAt := time.Now().Unix()
	_, err := s.db.Exec(
		"UPDATE sessions SET ended_at = ? WHERE session_id = ?",
		endedAt, sessionID,
	)
	return err
}

// GetSession retrieves session information. Returns nil if not found.
// Corresponds to TypeScript's getSession.
func (s *HarnessStore) GetSession(sessionID string) (*SessionState, error) {
	var (
		sessionIDOut   string
		mode           string
		projectRoot    string
		startedAt      int64
		endedAt        sql.NullInt64
		contextJSON    string
	)

	err := s.db.QueryRow(
		"SELECT session_id, mode, project_root, started_at, ended_at, context_json FROM sessions WHERE session_id = ?",
		sessionID,
	).Scan(&sessionIDOut, &mode, &projectRoot, &startedAt, &endedAt, &contextJSON)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query session: %w", err)
	}

	var ctx map[string]interface{}
	if err := json.Unmarshal([]byte(contextJSON), &ctx); err != nil {
		ctx = map[string]interface{}{}
	}

	state := &SessionState{
		SessionID:   sessionIDOut,
		Mode:        SessionMode(mode),
		ProjectRoot: projectRoot,
		StartedAt:   unixToISO(startedAt),
		Context:     ctx,
	}
	if endedAt.Valid {
		s := unixToISO(endedAt.Int64)
		state.EndedAt = &s
	}

	return state, nil
}

// ============================================================
// Signal management
// ============================================================

// SendSignal sends a signal and returns the inserted row ID.
// Corresponds to TypeScript's sendSignal.
func (s *HarnessStore) SendSignal(signal Signal) (int64, error) {
	sentAt := time.Now().Unix()

	payloadJSON, err := json.Marshal(signal.Payload)
	if err != nil {
		return 0, fmt.Errorf("marshal payload: %w", err)
	}

	result, err := s.db.Exec(
		`INSERT INTO signals(type, from_session_id, to_session_id, payload_json, sent_at)
         VALUES (?, ?, ?, ?, ?)`,
		signal.Type,
		signal.FromSessionID,
		signal.ToSessionID, // nil is treated as SQL NULL
		string(payloadJSON),
		sentAt,
	)
	if err != nil {
		return 0, fmt.Errorf("insert signal: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last insert id: %w", err)
	}

	return id, nil
}

// ReceiveSignals receives unconsumed signals and marks them as consumed.
// Targets signals addressed to sessionId or broadcast (to_session_id IS NULL).
// Excludes signals sent by the session itself.
// Corresponds to TypeScript's receiveSignals.
func (s *HarnessStore) ReceiveSignals(sessionID string) ([]Signal, error) {
	rows, err := s.db.Query(
		`SELECT id, type, from_session_id, to_session_id, payload_json, sent_at
         FROM signals
         WHERE consumed = 0
           AND (to_session_id = ? OR to_session_id IS NULL)
           AND from_session_id != ?
         ORDER BY sent_at ASC`,
		sessionID, sessionID,
	)
	if err != nil {
		return nil, fmt.Errorf("query signals: %w", err)
	}
	defer rows.Close()

	var signals []Signal
	var ids []interface{}

	for rows.Next() {
		var (
			id            int64
			signalType    string
			fromSession   string
			toSession     sql.NullString
			payloadJSON   string
			sentAt        int64
		)
		if scanErr := rows.Scan(&id, &signalType, &fromSession, &toSession, &payloadJSON, &sentAt); scanErr != nil {
			return nil, fmt.Errorf("scan signal row: %w", scanErr)
		}

		var payload map[string]interface{}
		if jsonErr := json.Unmarshal([]byte(payloadJSON), &payload); jsonErr != nil {
			payload = map[string]interface{}{}
		}

		sig := Signal{
			ID:            id,
			Type:          signalType,
			FromSessionID: fromSession,
			Payload:       payload,
			Timestamp:     unixToISO(sentAt),
		}
		if toSession.Valid {
			sig.ToSessionID = &toSession.String
		}

		signals = append(signals, sig)
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate signal rows: %w", err)
	}

	if len(ids) == 0 {
		return []Signal{}, nil
	}

	// Mark received signals as consumed.
	// database/sql does not natively support variadic placeholders, so build the SQL dynamically.
	placeholders := buildPlaceholders(len(ids))
	_, err = s.db.Exec(
		"UPDATE signals SET consumed = 1 WHERE id IN ("+placeholders+")",
		ids...,
	)
	if err != nil {
		return nil, fmt.Errorf("mark signals consumed: %w", err)
	}

	return signals, nil
}

// ============================================================
// Task failure management
// ============================================================

// RecordFailure records a task failure and returns the inserted row ID.
// Corresponds to TypeScript's recordFailure.
func (s *HarnessStore) RecordFailure(failure TaskFailure, sessionID string) (int64, error) {
	failedAt := time.Now().Unix()

	result, err := s.db.Exec(
		`INSERT INTO task_failures(task_id, session_id, severity, message, detail, failed_at, attempt)
         VALUES (?, ?, ?, ?, ?, ?, ?)`,
		failure.TaskID,
		sessionID,
		failure.Severity,
		failure.Message,
		failure.Detail, // nil is treated as SQL NULL
		failedAt,
		failure.Attempt,
	)
	if err != nil {
		return 0, fmt.Errorf("insert task_failure: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last insert id: %w", err)
	}

	return id, nil
}

// GetFailures retrieves the failure history for a task.
// Corresponds to TypeScript's getFailures.
func (s *HarnessStore) GetFailures(taskID string) ([]TaskFailure, error) {
	rows, err := s.db.Query(
		`SELECT task_id, severity, message, detail, failed_at, attempt
         FROM task_failures
         WHERE task_id = ?
         ORDER BY failed_at ASC`,
		taskID,
	)
	if err != nil {
		return nil, fmt.Errorf("query task_failures: %w", err)
	}
	defer rows.Close()

	var failures []TaskFailure
	for rows.Next() {
		var (
			tid      string
			severity string
			message  string
			detail   sql.NullString
			failedAt int64
			attempt  int
		)
		if scanErr := rows.Scan(&tid, &severity, &message, &detail, &failedAt, &attempt); scanErr != nil {
			return nil, fmt.Errorf("scan failure row: %w", scanErr)
		}

		f := TaskFailure{
			TaskID:    tid,
			Severity:  severity,
			Message:   message,
			Timestamp: unixToISO(failedAt),
			Attempt:   attempt,
		}
		if detail.Valid {
			f.Detail = &detail.String
		}

		failures = append(failures, f)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate failure rows: %w", err)
	}

	if failures == nil {
		return []TaskFailure{}, nil
	}
	return failures, nil
}

// ============================================================
// work_states management
// ============================================================

// SetWorkState registers a work/codex mode state (TTL 24 hours).
// Corresponds to TypeScript's setWorkState.
func (s *HarnessStore) SetWorkState(sessionID string, opts WorkStateOptions) error {
	expiresAt := time.Now().Unix() + 24*3600

	_, err := s.db.Exec(
		`INSERT INTO work_states(session_id, codex_mode, bypass_rm_rf, bypass_git_push, work_mode, expires_at)
         VALUES (?, ?, ?, ?, ?, ?)
         ON CONFLICT(session_id) DO UPDATE SET
           codex_mode = excluded.codex_mode,
           bypass_rm_rf = excluded.bypass_rm_rf,
           bypass_git_push = excluded.bypass_git_push,
           work_mode = excluded.work_mode,
           expires_at = excluded.expires_at`,
		sessionID,
		boolToInt(opts.CodexMode),
		boolToInt(opts.BypassRmRf),
		boolToInt(opts.BypassGitPush),
		boolToInt(opts.WorkMode),
		expiresAt,
	)
	return err
}

// GetWorkState retrieves a valid work_state. Returns nil if expired.
// Corresponds to TypeScript's getWorkState.
func (s *HarnessStore) GetWorkState(sessionID string) (*WorkState, error) {
	now := time.Now().Unix()

	var (
		sid           string
		codexMode     int
		bypassRmRf    int
		bypassGitPush int
		workMode      int
		expiresAt     int64
	)

	err := s.db.QueryRow(
		`SELECT session_id, codex_mode, bypass_rm_rf, bypass_git_push, work_mode, expires_at
         FROM work_states
         WHERE session_id = ? AND expires_at > ?`,
		sessionID, now,
	).Scan(&sid, &codexMode, &bypassRmRf, &bypassGitPush, &workMode, &expiresAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query work_state: %w", err)
	}

	return &WorkState{
		SessionID:     sid,
		CodexMode:     intToBool(codexMode),
		BypassRmRf:    intToBool(bypassRmRf),
		BypassGitPush: intToBool(bypassGitPush),
		WorkMode:      intToBool(workMode),
		ExpiresAt:     expiresAt,
	}, nil
}

// CleanExpiredWorkStates deletes expired work_states and returns the number of rows deleted.
// Corresponds to TypeScript's cleanExpiredWorkStates.
func (s *HarnessStore) CleanExpiredWorkStates() (int64, error) {
	now := time.Now().Unix()
	result, err := s.db.Exec(
		"DELETE FROM work_states WHERE expires_at <= ?",
		now,
	)
	if err != nil {
		return 0, fmt.Errorf("clean expired work_states: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("rows affected: %w", err)
	}
	return affected, nil
}

// ============================================================
// schema_meta key/value management
// ============================================================

// GetMeta retrieves a value from the schema_meta table. Returns "" if not found.
// Corresponds to TypeScript's getMeta.
func (s *HarnessStore) GetMeta(key string) (string, error) {
	var value string
	err := s.db.QueryRow(
		"SELECT value FROM schema_meta WHERE key = ?",
		key,
	).Scan(&value)

	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("query schema_meta: %w", err)
	}

	return value, nil
}

// SetMeta saves a value to the schema_meta table (upsert).
// Corresponds to TypeScript's setMeta.
func (s *HarnessStore) SetMeta(key, value string) error {
	_, err := s.db.Exec(
		`INSERT INTO schema_meta(key, value) VALUES (?, ?)
         ON CONFLICT(key) DO UPDATE SET value = excluded.value`,
		key, value,
	)
	return err
}

// ============================================================
// assumptions management (new table)
// ============================================================

// RecordAssumption saves an assumption or precondition recorded by an agent, and returns the inserted row ID.
func (s *HarnessStore) RecordAssumption(a Assumption) (int64, error) {
	createdAt := time.Now().Unix()

	result, err := s.db.Exec(
		`INSERT INTO assumptions(session_id, task_id, assumption, confidence, created_at)
         VALUES (?, ?, ?, ?, ?)`,
		a.SessionID,
		a.TaskID, // nil is treated as SQL NULL
		a.Assumption,
		a.Confidence,
		createdAt,
	)
	if err != nil {
		return 0, fmt.Errorf("insert assumption: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last insert id: %w", err)
	}

	return id, nil
}

// GetAssumptions returns the list of assumptions associated with a session.
func (s *HarnessStore) GetAssumptions(sessionID string) ([]Assumption, error) {
	rows, err := s.db.Query(
		`SELECT id, session_id, task_id, assumption, confidence, created_at, validated_at
         FROM assumptions
         WHERE session_id = ?
         ORDER BY created_at ASC`,
		sessionID,
	)
	if err != nil {
		return nil, fmt.Errorf("query assumptions: %w", err)
	}
	defer rows.Close()

	var result []Assumption
	for rows.Next() {
		var (
			id          int64
			sid         string
			taskID      sql.NullString
			assumption  string
			confidence  float64
			createdAt   int64
			validatedAt sql.NullInt64
		)
		if scanErr := rows.Scan(&id, &sid, &taskID, &assumption, &confidence, &createdAt, &validatedAt); scanErr != nil {
			return nil, fmt.Errorf("scan assumption row: %w", scanErr)
		}

		a := Assumption{
			ID:         id,
			SessionID:  sid,
			Assumption: assumption,
			Confidence: confidence,
			CreatedAt:  unixToISO(createdAt),
		}
		if taskID.Valid {
			a.TaskID = &taskID.String
		}
		if validatedAt.Valid {
			s := unixToISO(validatedAt.Int64)
			a.ValidatedAt = &s
		}

		result = append(result, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate assumption rows: %w", err)
	}

	if result == nil {
		return []Assumption{}, nil
	}
	return result, nil
}

// ============================================================
// agent_states management
// ============================================================

// AgentStateRecord represents a single row in the agent_states table.
type AgentStateRecord struct {
	AgentID          string  `json:"agent_id"`
	AgentType        string  `json:"agent_type"`
	SessionID        string  `json:"session_id"`
	State            string  `json:"state"`
	StartedAt        string  `json:"started_at"`         // ISO 8601
	StoppedAt        *string `json:"stopped_at,omitempty"` // ISO 8601
	RecoveryAttempts int     `json:"recovery_attempts"`
}

// UpsertAgentState registers or updates an agent state.
// If agent_id already exists, updates state, stopped_at, and recovery_attempts.
func (s *HarnessStore) UpsertAgentState(rec AgentStateRecord) error {
	startedAt, err := parseISOToUnix(rec.StartedAt)
	if err != nil {
		return fmt.Errorf("parse started_at: %w", err)
	}

	var stoppedAt interface{}
	if rec.StoppedAt != nil {
		t, tErr := parseISOToUnix(*rec.StoppedAt)
		if tErr != nil {
			return fmt.Errorf("parse stopped_at: %w", tErr)
		}
		stoppedAt = t
	}

	_, err = s.db.Exec(
		`INSERT INTO agent_states(agent_id, agent_type, session_id, state, started_at, stopped_at, recovery_attempts)
         VALUES (?, ?, ?, ?, ?, ?, ?)
         ON CONFLICT(agent_id) DO UPDATE SET
           state             = excluded.state,
           stopped_at        = excluded.stopped_at,
           recovery_attempts = excluded.recovery_attempts`,
		rec.AgentID,
		rec.AgentType,
		rec.SessionID,
		rec.State,
		startedAt,
		stoppedAt,
		rec.RecoveryAttempts,
	)
	return err
}

// GetAgentState retrieves the agent state for the given agent_id.
// Returns nil if not found.
func (s *HarnessStore) GetAgentState(agentID string) (*AgentStateRecord, error) {
	var (
		agentIDOut       string
		agentType        string
		sessionID        string
		state            string
		startedAt        int64
		stoppedAt        sql.NullInt64
		recoveryAttempts int
	)

	err := s.db.QueryRow(
		`SELECT agent_id, agent_type, session_id, state, started_at, stopped_at, recovery_attempts
         FROM agent_states WHERE agent_id = ?`,
		agentID,
	).Scan(&agentIDOut, &agentType, &sessionID, &state, &startedAt, &stoppedAt, &recoveryAttempts)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query agent_state: %w", err)
	}

	rec := &AgentStateRecord{
		AgentID:          agentIDOut,
		AgentType:        agentType,
		SessionID:        sessionID,
		State:            state,
		StartedAt:        unixToISO(startedAt),
		RecoveryAttempts: recoveryAttempts,
	}
	if stoppedAt.Valid {
		s := unixToISO(stoppedAt.Int64)
		rec.StoppedAt = &s
	}

	return rec, nil
}

// ListAgentStates returns all agent states ordered by started_at ascending.
// If onlyActive is true, only records where stopped_at IS NULL are returned.
func (s *HarnessStore) ListAgentStates(onlyActive bool) ([]AgentStateRecord, error) {
	query := `SELECT agent_id, agent_type, session_id, state, started_at, stopped_at, recovery_attempts
              FROM agent_states`
	if onlyActive {
		query += ` WHERE stopped_at IS NULL`
	}
	query += ` ORDER BY started_at ASC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query agent_states: %w", err)
	}
	defer rows.Close()

	var result []AgentStateRecord
	for rows.Next() {
		var (
			agentID          string
			agentType        string
			sessionID        string
			state            string
			startedAt        int64
			stoppedAt        sql.NullInt64
			recoveryAttempts int
		)
		if scanErr := rows.Scan(&agentID, &agentType, &sessionID, &state, &startedAt, &stoppedAt, &recoveryAttempts); scanErr != nil {
			return nil, fmt.Errorf("scan agent_state row: %w", scanErr)
		}

		rec := AgentStateRecord{
			AgentID:          agentID,
			AgentType:        agentType,
			SessionID:        sessionID,
			State:            state,
			StartedAt:        unixToISO(startedAt),
			RecoveryAttempts: recoveryAttempts,
		}
		if stoppedAt.Valid {
			st := unixToISO(stoppedAt.Int64)
			rec.StoppedAt = &st
		}
		result = append(result, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate agent_state rows: %w", err)
	}

	if result == nil {
		return []AgentStateRecord{}, nil
	}
	return result, nil
}

// ============================================================
// Utilities
// ============================================================

// unixToISO converts a Unix timestamp in seconds to an ISO 8601 UTC string.
func unixToISO(unixSec int64) string {
	return time.Unix(unixSec, 0).UTC().Format(time.RFC3339)
}

// parseISOToUnix converts an ISO 8601 string to a Unix timestamp in seconds.
func parseISOToUnix(iso string) (int64, error) {
	t, err := time.Parse(time.RFC3339, iso)
	if err != nil {
		// Fallback: use current time
		return time.Now().Unix(), fmt.Errorf("parse time %q: %w", iso, err)
	}
	return t.Unix(), nil
}

// boolToInt converts a bool to a SQLite integer (0/1).
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// intToBool converts a SQLite integer (0/1) to a bool.
func intToBool(n int) bool {
	return n != 0
}

// buildPlaceholders returns N "?" placeholders joined by commas.
// Example: buildPlaceholders(3) => "?,?,?"
func buildPlaceholders(n int) string {
	if n <= 0 {
		return ""
	}
	buf := make([]byte, 0, n*2-1)
	for i := 0; i < n; i++ {
		if i > 0 {
			buf = append(buf, ',')
		}
		buf = append(buf, '?')
	}
	return string(buf)
}
