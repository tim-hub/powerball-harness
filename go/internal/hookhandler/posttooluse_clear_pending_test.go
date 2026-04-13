package hookhandler

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClearPendingHandler_NoPendingDir(t *testing.T) {
	dir := t.TempDir()

	h := &ClearPendingHandler{ProjectRoot: dir}

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(`{}`), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp clearPendingResponse
	if err := json.Unmarshal(bytes.TrimRight(out.Bytes(), "\n"), &resp); err != nil {
		t.Fatalf("invalid JSON: %s", out.String())
	}
	if !resp.Continue {
		t.Errorf("expected continue=true")
	}
}

func TestClearPendingHandler_DeletesPendingFiles(t *testing.T) {
	dir := t.TempDir()

	pendingDir := filepath.Join(dir, ".claude", "state", "pending-skills")
	if err := os.MkdirAll(pendingDir, 0700); err != nil {
		t.Fatal(err)
	}

	// create .pending files
	pendingFiles := []string{"skill-a.pending", "skill-b.pending", "skill-c.pending"}
	for _, name := range pendingFiles {
		if err := os.WriteFile(filepath.Join(pendingDir, name), []byte("pending"), 0600); err != nil {
			t.Fatal(err)
		}
	}

	// non-.pending file (should not be deleted)
	otherFile := filepath.Join(pendingDir, "skill-a.json")
	if err := os.WriteFile(otherFile, []byte(`{}`), 0600); err != nil {
		t.Fatal(err)
	}

	h := &ClearPendingHandler{ProjectRoot: dir}

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(`{}`), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp clearPendingResponse
	if err := json.Unmarshal(bytes.TrimRight(out.Bytes(), "\n"), &resp); err != nil {
		t.Fatalf("invalid JSON: %s", out.String())
	}
	if !resp.Continue {
		t.Errorf("expected continue=true")
	}

	// verify that .pending files were deleted
	for _, name := range pendingFiles {
		if _, err := os.Stat(filepath.Join(pendingDir, name)); err == nil {
			t.Errorf("expected %s to be deleted", name)
		}
	}

	// verify that the .json file is retained
	if _, err := os.Stat(otherFile); err != nil {
		t.Errorf("skill-a.json should not be deleted")
	}
}

func TestClearPendingHandler_EmptyPendingDir(t *testing.T) {
	dir := t.TempDir()

	pendingDir := filepath.Join(dir, ".claude", "state", "pending-skills")
	if err := os.MkdirAll(pendingDir, 0700); err != nil {
		t.Fatal(err)
	}
	// empty directory (no .pending files)

	h := &ClearPendingHandler{ProjectRoot: dir}

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(`{}`), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp clearPendingResponse
	if err := json.Unmarshal(bytes.TrimRight(out.Bytes(), "\n"), &resp); err != nil {
		t.Fatalf("invalid JSON: %s", out.String())
	}
	if !resp.Continue {
		t.Errorf("expected continue=true")
	}
}

func TestClearPendingHandler_EmptyInput(t *testing.T) {
	dir := t.TempDir()

	h := &ClearPendingHandler{ProjectRoot: dir}

	var out bytes.Buffer
	// should not error even with empty stdin
	err := h.Handle(strings.NewReader(""), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp clearPendingResponse
	if err := json.Unmarshal(bytes.TrimRight(out.Bytes(), "\n"), &resp); err != nil {
		t.Fatalf("invalid JSON: %s", out.String())
	}
	if !resp.Continue {
		t.Errorf("expected continue=true")
	}
}

func TestClearPendingHandler_MultipleRuns(t *testing.T) {
	dir := t.TempDir()

	pendingDir := filepath.Join(dir, ".claude", "state", "pending-skills")
	if err := os.MkdirAll(pendingDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pendingDir, "skill-x.pending"), []byte(""), 0600); err != nil {
		t.Fatal(err)
	}

	h := &ClearPendingHandler{ProjectRoot: dir}

	// 1st run: delete the file
	var out1 bytes.Buffer
	if err := h.Handle(strings.NewReader(`{}`), &out1); err != nil {
		t.Fatal(err)
	}

	// 2nd run: already deleted → should not error
	var out2 bytes.Buffer
	if err := h.Handle(strings.NewReader(`{}`), &out2); err != nil {
		t.Fatalf("second run should not error: %v", err)
	}

	var resp clearPendingResponse
	if err := json.Unmarshal(bytes.TrimRight(out2.Bytes(), "\n"), &resp); err != nil {
		t.Fatalf("invalid JSON: %s", out2.String())
	}
	if !resp.Continue {
		t.Errorf("expected continue=true on second run")
	}
}
