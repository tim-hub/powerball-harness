// Package trace writes per-task execution trace events to JSONL files under
// .claude/state/traces/. The schema is defined in
// .claude/memory/schemas/trace.v1.md — this package is the canonical writer
// implementation for that schema.
//
// Task-level traces (this package) are distinct from session-level traces
// emitted by hookhandler.EmitAgentTrace (which writes .claude/state/agent-trace.jsonl).
// Session traces log every tool call across a session; task traces scope to
// one Plans.md task and capture its full attempt history for replay by the
// Advisor and code-space search.
package trace

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

const (
	// SchemaVersion is the schema identifier stamped on every event.
	SchemaVersion = "trace.v1"

	// HardCapBytes is the rotation threshold per trace.v1 schema.
	HardCapBytes int64 = 50 * 1024 * 1024

	// MaxLineBytes is a defensive upper bound on a single marshaled event.
	// A payload larger than this almost certainly indicates a caller bug
	// (e.g. accidentally embedding file contents). 1 MiB is generous enough
	// that legitimate events never hit it but tight enough that file contents
	// do.
	MaxLineBytes = 1 * 1024 * 1024

	// traceDirName is the subdirectory under .claude/state/.
	traceDirName = "traces"
)

// Event is one line of a trace.v1 JSONL file. Payload is deliberately
// json.RawMessage so this package isn't coupled to every payload schema
// defined in trace.v1.md — callers marshal their own payload types.
type Event struct {
	Schema    string          `json:"schema"`
	TS        string          `json:"ts"`
	TaskID    string          `json:"task_id"`
	EventType string          `json:"event_type"`
	Agent     string          `json:"agent,omitempty"`
	AttemptN  int             `json:"attempt_n,omitempty"`
	Payload   json.RawMessage `json:"payload"`
}

// validEventTypes enumerates the event_type values defined in trace.v1.md.
var validEventTypes = map[string]bool{
	"task_start":  true,
	"tool_call":   true,
	"decision":    true,
	"error":       true,
	"fix_attempt": true,
	"outcome":     true,
}

// writerDeps holds injectable syscalls so tests can substitute flock and the
// clock without touching package-level state. Pattern mirrors plansWatcherDeps
// in go/internal/hookhandler/plans_watcher.go.
type writerDeps struct {
	flock func(fd int, how int) error
	now   func() time.Time
}

func defaultWriterDeps() writerDeps {
	return writerDeps{
		flock: func(fd int, how int) error { return syscall.Flock(fd, how) },
		now:   func() time.Time { return time.Now().UTC() },
	}
}

// Writer appends events to per-task trace files rooted at a project root.
// Safe for concurrent use from multiple goroutines; flock serializes writers
// across processes that share the same filesystem.
type Writer struct {
	root string
	deps writerDeps
}

// NewWriter returns a Writer rooted at the given project root (the directory
// that contains .claude/).
func NewWriter(root string) *Writer {
	return &Writer{root: root, deps: defaultWriterDeps()}
}

// TraceDir returns the absolute directory path where trace files live.
func (w *Writer) TraceDir() string {
	return filepath.Join(w.root, ".claude", "state", traceDirName)
}

// TracePath returns the absolute JSONL file path for the given task id.
// Exposed so callers (e.g. the advisor-context loader) can locate a trace
// file without duplicating the path convention.
func (w *Writer) TracePath(taskID string) (string, error) {
	if err := validateTaskID(taskID); err != nil {
		return "", err
	}
	return filepath.Join(w.TraceDir(), taskID+".jsonl"), nil
}

// validateTaskID rejects ids that could escape the trace dir via path
// traversal. Plans.md task ids are numeric with dots (e.g. "72.1" or
// "72.1.fix"); anything with separators or ".." is suspect.
func validateTaskID(id string) error {
	if id == "" {
		return fmt.Errorf("task_id is required")
	}
	// Traversal is checked before separator since ".." can appear even in
	// path-separator-free ids (e.g. "72..fix") and is the more specific danger.
	if strings.Contains(id, "..") {
		return fmt.Errorf("task_id %q contains traversal sequence", id)
	}
	if strings.ContainsAny(id, `/\`) {
		return fmt.Errorf("task_id %q contains path separator", id)
	}
	return nil
}

// AppendEvent validates the event, fills defaults, and appends one JSONL line
// to the per-task trace file. Flock serializes cross-process writers;
// fsync ensures durability before the append is considered complete.
//
// Privacy note: the caller is responsible for ensuring Payload does not
// contain file contents, environment variable values, or secrets. The writer
// enforces an upper size bound (MaxLineBytes) to catch accidental leaks of
// large payloads.
func (w *Writer) AppendEvent(ev Event) error {
	if ev.Schema == "" {
		ev.Schema = SchemaVersion
	}
	if ev.Schema != SchemaVersion {
		return fmt.Errorf("unsupported schema %q (writer emits %q)", ev.Schema, SchemaVersion)
	}
	if ev.TS == "" {
		ev.TS = w.deps.now().Format(time.RFC3339Nano)
	}
	if ev.EventType == "" {
		return fmt.Errorf("event_type is required")
	}
	if !validEventTypes[ev.EventType] {
		return fmt.Errorf("invalid event_type %q", ev.EventType)
	}
	if len(ev.Payload) == 0 {
		ev.Payload = json.RawMessage("{}")
	}

	path, err := w.TracePath(ev.TaskID)
	if err != nil {
		return err
	}

	line, err := json.Marshal(ev)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}
	if len(line) > MaxLineBytes {
		return fmt.Errorf("event too large: %d bytes exceeds cap %d", len(line), MaxLineBytes)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("mkdir traces: %w", err)
	}

	if err := w.maybeRotate(path); err != nil {
		return fmt.Errorf("rotate: %w", err)
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open trace file: %w", err)
	}
	defer f.Close()

	if err := w.deps.flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		return fmt.Errorf("acquire flock: %w", err)
	}
	defer func() {
		// Unlock is belt-and-suspenders: close(fd) also releases the lock,
		// but explicit unlock makes the intent visible and keeps behavior
		// consistent if we later refactor to keep fds open.
		_ = w.deps.flock(int(f.Fd()), syscall.LOCK_UN)
	}()

	if _, err := f.Write(append(line, '\n')); err != nil {
		return fmt.Errorf("append line: %w", err)
	}
	if err := f.Sync(); err != nil {
		return fmt.Errorf("fsync: %w", err)
	}
	return nil
}

// maybeRotate renames the file to the next <task_id>.N.jsonl generation when
// it has reached HardCapBytes. The new event will land in a fresh file.
// Returns nil if the file doesn't exist yet or is under the cap.
func (w *Writer) maybeRotate(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if info.Size() < HardCapBytes {
		return nil
	}
	base := strings.TrimSuffix(path, ".jsonl")
	for i := 1; i < 10000; i++ {
		rotated := fmt.Sprintf("%s.%d.jsonl", base, i)
		if _, statErr := os.Stat(rotated); os.IsNotExist(statErr) {
			return os.Rename(path, rotated)
		}
	}
	return fmt.Errorf("rotation exhausted 10000 generations for %s", path)
}

// MarshalPayload is a convenience for callers that want to pass a typed
// struct as payload without wrapping in json.RawMessage themselves.
func MarshalPayload(v any) (json.RawMessage, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}
	return json.RawMessage(b), nil
}
