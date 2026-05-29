package plans

import (
	"os"
	"path/filepath"
	"testing"
)

const sample = `{
  "phases": [
    {"id": 1, "title": "P1", "status": "active", "tasks": [
      {"id": "1.1", "name": "WIP task", "status": "cc:WIP", "qualityMarkers": ["skip:tdd"], "depends": []},
      {"id": "1.2", "name": "Todo task", "status": "cc:TODO", "qualityMarkers": [], "depends": ["1.1"]}
    ]},
    {"id": 2, "title": "P2", "status": "archived", "tasks": [
      {"id": "2.1", "name": "Done task", "status": "cc:done", "qualityMarkers": [], "depends": []}
    ]}
  ]
}`

func writeSample(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, ".claude", "harness")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "plans.json"), []byte(sample), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestResolveAndLoad(t *testing.T) {
	root := writeSample(t)
	path := ResolvePath(root, "")
	if path == "" {
		t.Fatal("ResolvePath returned empty for existing plans.json")
	}
	p, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if p == nil || len(p.Phases) != 2 {
		t.Fatalf("expected 2 phases, got %#v", p)
	}
}

func TestResolveMissing(t *testing.T) {
	if got := ResolvePath(t.TempDir(), ""); got != "" {
		t.Fatalf("expected empty path for missing plans.json, got %q", got)
	}
	// Load on empty path is a non-error empty state.
	p, err := Load("")
	if err != nil || p != nil {
		t.Fatalf("Load(\"\") = (%v, %v), want (nil, nil)", p, err)
	}
}

func TestQueries(t *testing.T) {
	p, err := Load(ResolvePath(writeSample(t), ""))
	if err != nil {
		t.Fatal(err)
	}
	if got := p.CountStatus("cc:WIP"); got != 1 {
		t.Errorf("CountStatus(cc:WIP) = %d, want 1", got)
	}
	if got := p.CountStatus("cc:TODO"); got != 1 {
		t.Errorf("CountStatus(cc:TODO) = %d, want 1", got)
	}
	if got := p.CountStatus("cc:done"); got != 1 {
		t.Errorf("CountStatus(cc:done) = %d, want 1", got)
	}
	if !p.HasWIP() {
		t.Error("HasWIP() = false, want true")
	}
	if names := p.WIPNames(); len(names) != 1 || names[0] != "WIP task" {
		t.Errorf("WIPNames() = %v, want [WIP task]", names)
	}
	if !p.WIPHasSkipTDD() {
		t.Error("WIPHasSkipTDD() = false, want true")
	}
	if len(p.AllTasks()) != 3 {
		t.Errorf("AllTasks() len = %d, want 3", len(p.AllTasks()))
	}
	task, phase := p.FindTask("2.1")
	if task == nil || phase == nil || task.Name != "Done task" || phase.ID != 2 {
		t.Errorf("FindTask(2.1) = %#v / %#v", task, phase)
	}
	if task, _ := p.FindTask("nope"); task != nil {
		t.Error("FindTask(nope) should be nil")
	}
}
