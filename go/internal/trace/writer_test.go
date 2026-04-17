package trace

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

// newTestWriter returns a Writer whose root is t.TempDir() and whose clock
// returns a fixed time unless the caller overrides it. Using a fixed clock
// removes flakiness from time-dependent assertions.
func newTestWriter(t *testing.T) *Writer {
	t.Helper()
	w := NewWriter(t.TempDir())
	w.deps.now = func() time.Time {
		return time.Date(2026, 4, 17, 10, 30, 0, 0, time.UTC)
	}
	return w
}

func TestAppendEvent_WritesValidJSONLine(t *testing.T) {
	w := newTestWriter(t)
	ev := Event{
		TaskID:    "72.2",
		EventType: "task_start",
		Payload:   json.RawMessage(`{"description":"test"}`),
	}
	if err := w.AppendEvent(ev); err != nil {
		t.Fatalf("AppendEvent: %v", err)
	}

	path, err := w.TracePath("72.2")
	if err != nil {
		t.Fatalf("TracePath: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read trace file: %v", err)
	}
	if !bytes.HasSuffix(data, []byte("\n")) {
		t.Errorf("trace file must end with newline, got %q", string(data))
	}
	var got Event
	if err := json.Unmarshal(bytes.TrimSpace(data), &got); err != nil {
		t.Fatalf("invalid JSON: %v (raw: %s)", err, string(data))
	}
	if got.Schema != SchemaVersion {
		t.Errorf("Schema=%q, want %q", got.Schema, SchemaVersion)
	}
	if got.TaskID != "72.2" {
		t.Errorf("TaskID=%q, want %q", got.TaskID, "72.2")
	}
	if got.TS == "" {
		t.Errorf("TS must be filled when empty; got empty")
	}
}

func TestAppendEvent_CreatesTraceDir(t *testing.T) {
	w := newTestWriter(t)
	if _, err := os.Stat(w.TraceDir()); !os.IsNotExist(err) {
		t.Fatalf("trace dir should not exist before first write")
	}
	ev := Event{TaskID: "a", EventType: "task_start", Payload: json.RawMessage(`{}`)}
	if err := w.AppendEvent(ev); err != nil {
		t.Fatalf("AppendEvent: %v", err)
	}
	info, err := os.Stat(w.TraceDir())
	if err != nil {
		t.Fatalf("stat trace dir: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("trace dir is not a directory")
	}
	// 0700 mode check — 0600 on the file, 0700 on the dir.
	if info.Mode().Perm() != 0o700 {
		t.Errorf("trace dir perms=%o, want 0700", info.Mode().Perm())
	}
}

func TestAppendEvent_Validation(t *testing.T) {
	cases := []struct {
		name    string
		ev      Event
		wantErr string
	}{
		{
			name:    "missing event_type",
			ev:      Event{TaskID: "1"},
			wantErr: "event_type is required",
		},
		{
			name:    "invalid event_type",
			ev:      Event{TaskID: "1", EventType: "frobnicate"},
			wantErr: "invalid event_type",
		},
		{
			name:    "empty task id",
			ev:      Event{EventType: "task_start"},
			wantErr: "task_id is required",
		},
		{
			name:    "task id with separator",
			ev:      Event{TaskID: "../escape", EventType: "task_start"},
			wantErr: "traversal",
		},
		{
			name:    "task id with slash",
			ev:      Event{TaskID: "a/b", EventType: "task_start"},
			wantErr: "separator",
		},
		{
			name:    "wrong schema",
			ev:      Event{TaskID: "1", EventType: "task_start", Schema: "trace.v99"},
			wantErr: "unsupported schema",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			w := newTestWriter(t)
			err := w.AppendEvent(tc.ev)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("error=%q, want substring %q", err.Error(), tc.wantErr)
			}
		})
	}
}

func TestAppendEvent_FillsDefaults(t *testing.T) {
	w := newTestWriter(t)
	ev := Event{TaskID: "d", EventType: "outcome"}
	if err := w.AppendEvent(ev); err != nil {
		t.Fatalf("AppendEvent: %v", err)
	}
	path, _ := w.TracePath("d")
	data, _ := os.ReadFile(path)
	var got Event
	if err := json.Unmarshal(bytes.TrimSpace(data), &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if got.Schema != SchemaVersion {
		t.Errorf("Schema default not applied: %q", got.Schema)
	}
	if got.TS != "2026-04-17T10:30:00Z" {
		t.Errorf("TS default not applied: %q", got.TS)
	}
	// Empty payload must serialize as empty object, not null/missing,
	// so readers can always index into it.
	if string(got.Payload) != "{}" {
		t.Errorf("Payload default=%s, want {}", string(got.Payload))
	}
}

func TestAppendEvent_RejectsOversizedEvent(t *testing.T) {
	w := newTestWriter(t)
	// Build a payload just over MaxLineBytes to verify the cap fires.
	huge := bytes.Repeat([]byte("x"), MaxLineBytes+1024)
	ev := Event{
		TaskID:    "big",
		EventType: "decision",
		Payload:   json.RawMessage(fmt.Sprintf(`{"blob":%q}`, string(huge))),
	}
	err := w.AppendEvent(ev)
	if err == nil {
		t.Fatalf("expected size cap error, got nil")
	}
	if !strings.Contains(err.Error(), "too large") {
		t.Errorf("error=%q, want containing 'too large'", err.Error())
	}
	// File must NOT exist — cap check precedes file creation.
	path, _ := w.TracePath("big")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("trace file should not exist after rejected event: %v", err)
	}
}

// TestAppendEvent_ConcurrentDifferentFiles is the DoD test from Plans.md 72.2:
// 10 goroutines writing to 10 different task files produce 10 valid JSONL
// files with no corruption.
func TestAppendEvent_ConcurrentDifferentFiles(t *testing.T) {
	w := NewWriter(t.TempDir())
	const N = 10
	var wg sync.WaitGroup
	errs := make(chan error, N)
	for i := range N {
		wg.Add(1)
		go func(taskID string) {
			defer wg.Done()
			ev := Event{
				TaskID:    taskID,
				EventType: "task_start",
				Agent:     "worker",
				Payload:   json.RawMessage(`{"description":"concurrent test"}`),
			}
			if err := w.AppendEvent(ev); err != nil {
				errs <- fmt.Errorf("%s: %w", taskID, err)
			}
		}(fmt.Sprintf("cc.%d", i))
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatalf("append failed: %v", err)
	}
	for i := range N {
		taskID := fmt.Sprintf("cc.%d", i)
		path := filepath.Join(w.TraceDir(), taskID+".jsonl")
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
		if len(lines) != 1 {
			t.Errorf("file %s has %d lines, want 1", path, len(lines))
		}
		var ev Event
		if err := json.Unmarshal([]byte(lines[0]), &ev); err != nil {
			t.Errorf("file %s line invalid JSON: %v", path, err)
			continue
		}
		if ev.TaskID != taskID {
			t.Errorf("file %s event TaskID=%q, want %q", path, ev.TaskID, taskID)
		}
	}
}

// TestAppendEvent_ConcurrentSameFile exercises flock serialization: many
// goroutines appending to one file must each produce a complete line with no
// interleaving or truncation.
func TestAppendEvent_ConcurrentSameFile(t *testing.T) {
	w := NewWriter(t.TempDir())
	const N = 50
	var wg sync.WaitGroup
	errs := make(chan error, N)
	for i := range N {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ev := Event{
				TaskID:    "shared",
				EventType: "tool_call",
				Payload:   json.RawMessage(fmt.Sprintf(`{"goroutine":%d}`, i)),
			}
			if err := w.AppendEvent(ev); err != nil {
				errs <- fmt.Errorf("g=%d: %w", i, err)
			}
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatalf("append failed: %v", err)
	}

	path := filepath.Join(w.TraceDir(), "shared.jsonl")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) != N {
		t.Fatalf("got %d lines, want %d", len(lines), N)
	}
	// Every line must be valid JSON and claim the shared task_id. Also
	// collect goroutine ids to confirm no duplicates or losses.
	seenG := make(map[int]bool, N)
	for i, line := range lines {
		var ev Event
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			t.Errorf("line %d invalid JSON: %v (content: %q)", i, err, line)
			continue
		}
		if ev.TaskID != "shared" {
			t.Errorf("line %d TaskID=%q, want %q", i, ev.TaskID, "shared")
		}
		var p struct {
			Goroutine int `json:"goroutine"`
		}
		if err := json.Unmarshal(ev.Payload, &p); err != nil {
			t.Errorf("line %d payload invalid: %v", i, err)
			continue
		}
		if seenG[p.Goroutine] {
			t.Errorf("goroutine %d appears more than once", p.Goroutine)
		}
		seenG[p.Goroutine] = true
	}
	if len(seenG) != N {
		t.Errorf("saw %d distinct goroutines, want %d", len(seenG), N)
	}
}

func TestMaybeRotate(t *testing.T) {
	w := newTestWriter(t)
	path, err := w.TracePath("r")
	if err != nil {
		t.Fatalf("TracePath: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// File doesn't exist — no-op.
	if err := w.maybeRotate(path); err != nil {
		t.Errorf("rotate on missing file: %v", err)
	}

	// Write a file under the cap — no-op.
	if err := os.WriteFile(path, []byte("small\n"), 0o600); err != nil {
		t.Fatalf("write small: %v", err)
	}
	if err := w.maybeRotate(path); err != nil {
		t.Errorf("rotate under cap: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("under-cap file should still exist: %v", err)
	}

	// Force the file to appear over-cap via truncate — rotation must move it.
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := f.Truncate(HardCapBytes + 1); err != nil {
		t.Fatalf("truncate: %v", err)
	}
	_ = f.Close()
	if err := w.maybeRotate(path); err != nil {
		t.Fatalf("rotate over cap: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("original path should be renamed away: %v", err)
	}
	rotated := strings.TrimSuffix(path, ".jsonl") + ".1.jsonl"
	if _, err := os.Stat(rotated); err != nil {
		t.Errorf("rotated file missing at %s: %v", rotated, err)
	}
}

func TestAppendEvent_FlockFailureIsSurfaced(t *testing.T) {
	w := NewWriter(t.TempDir())
	// Inject a flock that always fails to confirm the error path surfaces.
	w.deps.flock = func(fd int, how int) error {
		if how == syscall.LOCK_UN {
			return nil // allow unlock so defer is quiet
		}
		return fmt.Errorf("injected flock failure")
	}
	ev := Event{TaskID: "f", EventType: "task_start", Payload: json.RawMessage(`{}`)}
	err := w.AppendEvent(ev)
	if err == nil {
		t.Fatalf("expected flock error, got nil")
	}
	if !strings.Contains(err.Error(), "acquire flock") {
		t.Errorf("error=%q, want containing 'acquire flock'", err.Error())
	}
}

func TestMarshalPayload(t *testing.T) {
	raw, err := MarshalPayload(struct {
		Tool string `json:"tool"`
	}{Tool: "Edit"})
	if err != nil {
		t.Fatalf("MarshalPayload: %v", err)
	}
	if string(raw) != `{"tool":"Edit"}` {
		t.Errorf("got %s, want %s", string(raw), `{"tool":"Edit"}`)
	}
}
