package hookhandler

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"bytes"
)

func TestPreCompactSave_BasicOutput(t *testing.T) {
	dir := t.TempDir()
	h := &PreCompactSave{
		RepoRoot:  dir,
		StateDir:  filepath.Join(dir, "state"),
		PlansFile: filepath.Join(dir, "Plans.md"),
	}
	// Plans.md なし
	var buf bytes.Buffer
	if err := h.Handle(strings.NewReader("{}"), &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var resp preCompactResponse
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &resp); err != nil {
		t.Fatalf("invalid JSON output: %v\n%s", err, buf.String())
	}
	if !resp.Continue {
		t.Errorf("expected continue=true, got: %v", resp.Continue)
	}
}

func TestPreCompactSave_SavesArtifact(t *testing.T) {
	dir := t.TempDir()
	stateDir := filepath.Join(dir, "state")
	plansFile := filepath.Join(dir, "Plans.md")

	content := "| 1 | Implement feature | In progress | none | `cc:WIP` |\n" +
		"| 2 | Write tests | Not started | none | `cc:TODO` |\n"
	if err := os.WriteFile(plansFile, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	h := &PreCompactSave{
		RepoRoot:  dir,
		StateDir:  stateDir,
		PlansFile: plansFile,
	}

	var buf bytes.Buffer
	if err := h.Handle(strings.NewReader("{}"), &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// handoff-artifact.json が生成されていることを確認
	artifactPath := filepath.Join(stateDir, "handoff-artifact.json")
	data, err := os.ReadFile(artifactPath)
	if err != nil {
		t.Fatalf("handoff-artifact.json not found: %v", err)
	}

	var artifact handoffArtifact
	if err := json.Unmarshal(data, &artifact); err != nil {
		t.Fatalf("invalid artifact JSON: %v", err)
	}
	if artifact.Version != artifactVersion {
		t.Errorf("expected version %s, got: %s", artifactVersion, artifact.Version)
	}
	if artifact.ArtifactType != "structured-handoff" {
		t.Errorf("expected artifactType=structured-handoff, got: %s", artifact.ArtifactType)
	}
}

func TestPreCompactSave_WIPTasksExtracted(t *testing.T) {
	dir := t.TempDir()
	stateDir := filepath.Join(dir, "state")
	plansFile := filepath.Join(dir, "Plans.md")

	content := "| 1 | Feature X | In progress | none | `cc:WIP` |\n" +
		"| 2 | Feature Y | Not started | none | cc:TODO |\n" +
		"| 3 | Done | Completed | none | cc:完了 |\n"
	if err := os.WriteFile(plansFile, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	h := &PreCompactSave{
		RepoRoot:  dir,
		StateDir:  stateDir,
		PlansFile: plansFile,
	}

	var buf bytes.Buffer
	if err := h.Handle(strings.NewReader("{}"), &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp preCompactResponse
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if !resp.Continue {
		t.Errorf("expected continue=true")
	}
	// レスポンスメッセージに WIP タスク数が含まれることを確認
	if !strings.Contains(resp.Message, "WIP task") {
		t.Errorf("expected WIP task count in message, got: %s", resp.Message)
	}
}

func TestPreCompactSave_SavesLegacySnapshot(t *testing.T) {
	dir := t.TempDir()
	stateDir := filepath.Join(dir, "state")

	h := &PreCompactSave{
		RepoRoot:  dir,
		StateDir:  stateDir,
		PlansFile: filepath.Join(dir, "Plans.md"),
	}

	var buf bytes.Buffer
	if err := h.Handle(strings.NewReader("{}"), &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// precompact-snapshot.json が生成されていることを確認
	snapshotPath := filepath.Join(stateDir, "precompact-snapshot.json")
	data, err := os.ReadFile(snapshotPath)
	if err != nil {
		t.Fatalf("precompact-snapshot.json not found: %v", err)
	}

	var snapshot map[string]interface{}
	if err := json.Unmarshal(data, &snapshot); err != nil {
		t.Fatalf("invalid snapshot JSON: %v", err)
	}
	if snapshot["artifactType"] != "precompact-snapshot" {
		t.Errorf("expected artifactType=precompact-snapshot, got: %v", snapshot["artifactType"])
	}
	if snapshot["version"] != legacySnapshotVersion {
		t.Errorf("expected version=%s, got: %v", legacySnapshotVersion, snapshot["version"])
	}
}

func TestPreCompactSave_SecuritySymlinkCheck(t *testing.T) {
	dir := t.TempDir()
	stateDir := filepath.Join(dir, "state")
	if err := os.MkdirAll(stateDir, 0700); err != nil {
		t.Fatal(err)
	}

	// artifact path をシンボリックリンクにする
	target := filepath.Join(dir, "evil.json")
	if err := os.WriteFile(target, []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	artifactPath := filepath.Join(stateDir, "handoff-artifact.json")
	if err := os.Symlink(target, artifactPath); err != nil {
		t.Skip("symlink creation not supported")
	}

	h := &PreCompactSave{
		RepoRoot:  dir,
		StateDir:  stateDir,
		PlansFile: filepath.Join(dir, "Plans.md"),
	}

	var buf bytes.Buffer
	if err := h.Handle(strings.NewReader("{}"), &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp preCompactResponse
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if !resp.Continue {
		t.Error("should still continue even on symlink skip")
	}
	if !strings.Contains(resp.Message, "symlink") {
		t.Errorf("expected symlink message, got: %s", resp.Message)
	}
}

func TestPreCompactSave_PlanRowParsing(t *testing.T) {
	h := &PreCompactSave{}

	tests := []struct {
		name     string
		content  string
		wantWIP  int
		wantTODO int
	}{
		{
			name:     "WIP and TODO tasks",
			content:  "| 1 | Task A | DoD A | none | `cc:WIP` |\n| 2 | Task B | DoD B | none | cc:TODO |\n",
			wantWIP:  1,
			wantTODO: 1,
		},
		{
			name:    "completed task only",
			content: "| 1 | Task A | DoD A | none | cc:完了 |\n",
			wantWIP: 0,
		},
		{
			name:    "header line ignored",
			content: "| Task | Title | DoD | Depends | Status |\n|---|---|---|---|---|\n| 1 | Feature | - | none | cc:WIP |\n",
			wantWIP: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			plansFile := filepath.Join(dir, "Plans.md")
			if err := os.WriteFile(plansFile, []byte(tt.content), 0600); err != nil {
				t.Fatal(err)
			}
			h.PlansFile = plansFile
			rows := h.getPlanRows(plansFile)
			wipCount := countWIP(rows)
			if wipCount != tt.wantWIP {
				t.Errorf("expected WIP count %d, got %d (rows: %+v)", tt.wantWIP, wipCount, rows)
			}
		})
	}
}

func TestPreCompactSave_NextActionSelection(t *testing.T) {
	h := &PreCompactSave{}

	rows := []planRow{
		{TaskID: "1", Title: "Task A", Tags: planTags{Todo: true}},
		{TaskID: "2", Title: "Task B", Tags: planTags{Wip: true}},
		{TaskID: "3", Title: "Task C", Tags: planTags{Blocked: true}},
	}

	na := h.pickNextAction(rows)
	if na == nil {
		t.Fatal("expected next action, got nil")
	}
	// WIP が最優先
	if na.TaskID != "2" {
		t.Errorf("expected WIP task (ID=2), got: %s", na.TaskID)
	}
	if na.Priority != "high" {
		t.Errorf("expected priority=high, got: %s", na.Priority)
	}
}

func TestPreCompactSave_ContextResetPolicy(t *testing.T) {
	t.Setenv("HARNESS_CONTEXT_RESET_WIP_THRESHOLD", "2")
	t.Setenv("HARNESS_CONTEXT_RESET_MODE", "manual")

	policy := getContextResetPolicy()
	if policy.Mode != "manual" {
		t.Errorf("expected mode=manual, got: %s", policy.Mode)
	}
	if policy.Thresholds.WIPTasks != 2 {
		t.Errorf("expected WIPTasks threshold=2, got: %d", policy.Thresholds.WIPTasks)
	}
}

func TestPreCompactSave_OpenRisks(t *testing.T) {
	h := &PreCompactSave{}
	rows := []planRow{
		{TaskID: "1", Title: "Task A", Tags: planTags{Wip: true}},
		{TaskID: "2", Title: "Task B", Tags: planTags{Blocked: true}},
	}

	risks := h.buildOpenRisks(rows, []string{"file1.go", "file2.go"}, nil)
	if len(risks) == 0 {
		t.Fatal("expected risks, got none")
	}

	// WIP リスクがあることを確認
	found := false
	for _, r := range risks {
		if r.Kind == "continuity" && strings.Contains(r.Summary, "WIP") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected continuity risk for WIP tasks")
	}
}
