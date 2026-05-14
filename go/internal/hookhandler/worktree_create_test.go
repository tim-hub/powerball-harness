package hookhandler

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// WorktreeCreate is a notification hook — the handler writes nothing to stdout.
// Tests verify side effects (state dir, worktree-info.json) only.

func TestHandleWorktreeCreate_EmptyInput(t *testing.T) {
	var out bytes.Buffer
	if err := HandleWorktreeCreate(strings.NewReader(""), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Len() != 0 {
		t.Errorf("expected no stdout output, got: %s", out.String())
	}
}

func TestHandleWorktreeCreate_InvalidJSON(t *testing.T) {
	var out bytes.Buffer
	if err := HandleWorktreeCreate(strings.NewReader("not json"), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Len() != 0 {
		t.Errorf("expected no stdout output, got: %s", out.String())
	}
}

func TestHandleWorktreeCreate_NoCWD(t *testing.T) {
	var out bytes.Buffer
	if err := HandleWorktreeCreate(strings.NewReader(`{"session_id":"s1","cwd":""}`), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Len() != 0 {
		t.Errorf("expected no stdout output, got: %s", out.String())
	}
}

func TestHandleWorktreeCreate_CreatesStateDir(t *testing.T) {
	dir := t.TempDir()

	var out bytes.Buffer
	payload := `{"session_id":"worker-123","cwd":"` + dir + `"}`
	if err := HandleWorktreeCreate(strings.NewReader(payload), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Len() != 0 {
		t.Errorf("expected no stdout output, got: %s", out.String())
	}

	stateDir := filepath.Join(dir, ".claude", "state")
	if info, err := os.Stat(stateDir); err != nil || !info.IsDir() {
		t.Errorf(".claude/state/ was not created at %s", stateDir)
	}
}

func TestHandleWorktreeCreate_WritesWorktreeInfo(t *testing.T) {
	dir := t.TempDir()

	var out bytes.Buffer
	payload := `{"session_id":"worker-xyz","cwd":"` + dir + `"}`
	if err := HandleWorktreeCreate(strings.NewReader(payload), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	infoPath := filepath.Join(dir, ".claude", "state", "worktree-info.json")
	data, err := os.ReadFile(infoPath)
	if err != nil {
		t.Fatalf("worktree-info.json not created: %v", err)
	}

	var info worktreeInfo
	if err := json.Unmarshal(bytes.TrimSpace(data), &info); err != nil {
		t.Fatalf("worktree-info.json is not valid JSON: %v\n%s", err, data)
	}

	if info.WorkerID != "worker-xyz" {
		t.Errorf("WorkerID = %q, want worker-xyz", info.WorkerID)
	}
	if info.CWD != dir {
		t.Errorf("CWD = %q, want %q", info.CWD, dir)
	}
	if info.CreatedAt == "" {
		t.Error("CreatedAt is empty")
	}
}

func TestHandleWorktreeCreate_Idempotent(t *testing.T) {
	dir := t.TempDir()

	for i := range 2 {
		var out bytes.Buffer
		payload := `{"session_id":"s","cwd":"` + dir + `"}`
		if err := HandleWorktreeCreate(strings.NewReader(payload), &out); err != nil {
			t.Fatalf("call %d: unexpected error: %v", i+1, err)
		}
		if out.Len() != 0 {
			t.Errorf("call %d: expected no stdout output, got: %s", i+1, out.String())
		}
	}
}
