package event

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPostCompactHandler_EmptyInput(t *testing.T) {
	dir := t.TempDir()
	h := &PostCompactHandler{StateDir: dir}

	var buf bytes.Buffer
	if err := h.Handle(strings.NewReader(""), &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp ApproveResponse
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &resp); err != nil {
		t.Fatalf("invalid JSON output: %v\n%s", err, buf.String())
	}
	if resp.Decision != "approve" {
		t.Errorf("expected approve, got: %s", resp.Decision)
	}
}

func TestPostCompactHandler_NoWIPTasks(t *testing.T) {
	dir := t.TempDir()
	plansFile := filepath.Join(dir, "Plans.md")
	// WIP タスクなし
	if err := os.WriteFile(plansFile, []byte("# Plans\n\n| 1 | Task A | Done | cc:完了 |\n"), 0600); err != nil {
		t.Fatal(err)
	}

	h := &PostCompactHandler{StateDir: dir, PlansFile: plansFile}
	var buf bytes.Buffer
	if err := h.Handle(strings.NewReader(`{"cwd":"/tmp"}`), &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp ApproveResponse
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &resp); err != nil {
		t.Fatalf("invalid JSON output: %v\n%s", err, buf.String())
	}
	if resp.Decision != "approve" {
		t.Errorf("expected approve, got: %s", resp.Decision)
	}
	if resp.AdditionalContext != "" {
		t.Errorf("expected no additionalContext, got: %s", resp.AdditionalContext)
	}
}

func TestPostCompactHandler_WithWIPTasks(t *testing.T) {
	dir := t.TempDir()
	plansFile := filepath.Join(dir, "Plans.md")
	content := `# Plans

| 1 | Implement feature X | In progress | cc:WIP |
| 2 | Write tests | Not started | cc:TODO |
| 3 | Done task | Done | cc:完了 |
`
	if err := os.WriteFile(plansFile, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	h := &PostCompactHandler{StateDir: dir, PlansFile: plansFile}
	var buf bytes.Buffer
	if err := h.Handle(strings.NewReader(`{"cwd":"/tmp"}`), &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp ApproveResponse
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &resp); err != nil {
		t.Fatalf("invalid JSON output: %v\n%s", err, buf.String())
	}
	if resp.Decision != "approve" {
		t.Errorf("expected approve, got: %s", resp.Decision)
	}
	if resp.AdditionalContext == "" {
		t.Fatal("expected additionalContext with WIP tasks")
	}
	if !strings.Contains(resp.AdditionalContext, "PostCompact Re-injection") {
		t.Errorf("expected re-injection header, got: %s", resp.AdditionalContext)
	}
	if !strings.Contains(resp.AdditionalContext, "cc:WIP") {
		t.Errorf("expected WIP task in context, got: %s", resp.AdditionalContext)
	}
}

func TestPostCompactHandler_WithPrecompactSnapshot(t *testing.T) {
	dir := t.TempDir()

	// precompact-snapshot.json を作成
	snapshot := `{
		"wipTasks": ["35.3.1 hook-handlers Go migration", "35.3.2 session actions"],
		"recentEdits": ["go/internal/event/event.go", "go/internal/event/session_env.go"]
	}`
	if err := os.WriteFile(filepath.Join(dir, "precompact-snapshot.json"), []byte(snapshot), 0600); err != nil {
		t.Fatal(err)
	}

	h := &PostCompactHandler{StateDir: dir}
	var buf bytes.Buffer
	if err := h.Handle(strings.NewReader(`{"cwd":"/tmp"}`), &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp ApproveResponse
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &resp); err != nil {
		t.Fatalf("invalid JSON output: %v\n%s", err, buf.String())
	}
	if resp.AdditionalContext == "" {
		t.Fatal("expected additionalContext from snapshot")
	}
	if !strings.Contains(resp.AdditionalContext, "hook-handlers Go migration") {
		t.Errorf("expected WIP task from snapshot, got: %s", resp.AdditionalContext)
	}
}

func TestPostCompactHandler_WritesCompactionLog(t *testing.T) {
	dir := t.TempDir()
	plansFile := filepath.Join(dir, "Plans.md")
	if err := os.WriteFile(plansFile, []byte("# Plans\n| 1 | Task | WIP | cc:WIP |\n"), 0600); err != nil {
		t.Fatal(err)
	}

	h := &PostCompactHandler{StateDir: dir, PlansFile: plansFile}
	var buf bytes.Buffer
	if err := h.Handle(strings.NewReader(`{}`), &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	logFile := filepath.Join(dir, "compaction-events.jsonl")
	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("expected compaction log, got error: %v", err)
	}

	var entry compactionLogEntry
	if err := json.Unmarshal(bytes.TrimSpace(data), &entry); err != nil {
		t.Fatalf("invalid log entry: %v\n%s", err, string(data))
	}
	if entry.Event != "post_compact" {
		t.Errorf("expected event=post_compact, got: %s", entry.Event)
	}
	if !entry.HasWIP {
		t.Error("expected has_wip=true")
	}
}
