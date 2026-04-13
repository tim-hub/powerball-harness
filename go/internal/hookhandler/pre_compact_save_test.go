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
	// no Plans.md
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

	// verify handoff-artifact.json is generated
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
		"| 3 | Done | Completed | none | cc:done |\n"
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
	// verify response message contains WIP task count
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

	// verify precompact-snapshot.json is generated
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

	// make artifact path a symbolic link
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
			content: "| 1 | Task A | DoD A | none | cc:done |\n",
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
	// WIP takes highest priority
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

	risks := h.buildOpenRisks(rows, []string{"file1.go", "file2.go"}, nil, nil)
	if len(risks) == 0 {
		t.Fatal("expected risks, got none")
	}

	// verify WIP risk is present
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

// TestBuildOpenRisks_SessionMetricsQualityRisk verifies that failed_checks in
// session-metrics.json is converted to a quality risk in open_risks (fix for issue #3).
func TestBuildOpenRisks_SessionMetricsQualityRisk(t *testing.T) {
	h := &PreCompactSave{}
	rows := []planRow{
		{TaskID: "1", Title: "Task A", Tags: planTags{Wip: true}},
	}

	// case with 2 failed_checks in session-metrics
	metrics := map[string]interface{}{
		"failed_checks": []interface{}{
			"lint check failed",
			"type check failed",
		},
	}

	risks := h.buildOpenRisks(rows, nil, nil, metrics)

	// verify quality risk is included
	found := false
	for _, r := range risks {
		if r.Kind == "quality" && strings.Contains(r.Summary, "failed check") {
			found = true
			if r.Severity != "high" {
				t.Errorf("expected severity=high for quality risk, got %q", r.Severity)
			}
			break
		}
	}
	if !found {
		t.Error("expected quality risk from session-metrics failed_checks, got none")
	}
}

// TestBuildOpenRisks_SessionMetrics_failedChecks_Field verifies that the
// failedChecks (camelCase) field in session-metrics is also converted to a quality risk.
func TestBuildOpenRisks_SessionMetrics_CamelCase(t *testing.T) {
	h := &PreCompactSave{}

	metrics := map[string]interface{}{
		"failedChecks": []interface{}{"test failed"},
	}

	risks := h.buildOpenRisks(nil, nil, nil, metrics)

	found := false
	for _, r := range risks {
		if r.Kind == "quality" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected quality risk from failedChecks (camelCase), got none")
	}
}

// TestBuildOpenRisks_NoMetrics verifies that no quality risk is added when metrics is nil.
func TestBuildOpenRisks_NoMetrics(t *testing.T) {
	h := &PreCompactSave{}
	rows := []planRow{
		{TaskID: "1", Title: "Task A", Tags: planTags{Wip: true}},
	}

	risks := h.buildOpenRisks(rows, nil, nil, nil)

	for _, r := range risks {
		if r.Kind == "quality" {
			t.Errorf("expected no quality risk when metrics is nil, but got: %+v", r)
		}
	}
}

// TestBuildOpenRisks_EmptyMetricsFailedChecks verifies that no quality risk is added
// when failed_checks is an empty array.
func TestBuildOpenRisks_EmptyMetricsFailedChecks(t *testing.T) {
	h := &PreCompactSave{}
	rows := []planRow{
		{TaskID: "1", Title: "Task A", Tags: planTags{Wip: true}},
	}

	metrics := map[string]interface{}{
		"failed_checks": []interface{}{},
	}

	risks := h.buildOpenRisks(rows, nil, nil, metrics)

	for _, r := range risks {
		if r.Kind == "quality" {
			t.Errorf("expected no quality risk when failed_checks is empty, but got: %+v", r)
		}
	}
}

// TestPreCompactSave_SessionMetricsOpenRisks verifies end-to-end that failures from
// session-metrics.json appear as quality risks in handoff-artifact.json open_risks
// (integration test for fix #3).
func TestPreCompactSave_SessionMetricsOpenRisks(t *testing.T) {
	dir := t.TempDir()
	stateDir := filepath.Join(dir, "state")
	plansFile := filepath.Join(dir, "Plans.md")

	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatal(err)
	}

	content := "| 1 | Feature X | In progress | none | `cc:WIP` |\n"
	if err := os.WriteFile(plansFile, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	// set failed_checks in session-metrics.json
	// getSessionMetrics reads from repoRoot/.claude/state/session-metrics.json
	claudeStateDir := filepath.Join(dir, ".claude", "state")
	if err := os.MkdirAll(claudeStateDir, 0o755); err != nil {
		t.Fatal(err)
	}
	metricsContent := `{"failed_checks":["lint failed","type error"],"session_id":"test-session"}`
	if err := os.WriteFile(filepath.Join(claudeStateDir, "session-metrics.json"), []byte(metricsContent), 0600); err != nil {
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

	artifactPath := filepath.Join(stateDir, "handoff-artifact.json")
	data, err := os.ReadFile(artifactPath)
	if err != nil {
		t.Fatalf("handoff-artifact.json not found: %v", err)
	}

	var artifact handoffArtifact
	if err := json.Unmarshal(data, &artifact); err != nil {
		t.Fatalf("invalid artifact JSON: %v", err)
	}

	// verify quality risk is included in open_risks
	found := false
	for _, r := range artifact.OpenRisks {
		if r.Kind == "quality" && strings.Contains(r.Summary, "failed check") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected quality risk in open_risks from session-metrics, got risks: %+v", artifact.OpenRisks)
	}
}

// TestBuildOpenRisks_ReviewStatusPassed verifies that no risk is added when
// review_status="passed" (symmetric with bash/JS version).
func TestBuildOpenRisks_ReviewStatusPassed(t *testing.T) {
	h := &PreCompactSave{}

	workState := map[string]interface{}{
		"review_status": "passed",
	}

	risks := h.buildOpenRisks(nil, nil, workState, nil)

	for _, r := range risks {
		if r.Kind == "review" {
			t.Errorf("expected no review risk when review_status=passed, got: %+v", r)
		}
	}
}

// TestBuildOpenRisks_ReviewStatusFailed verifies that a high-severity review risk
// is added when review_status="failed".
func TestBuildOpenRisks_ReviewStatusFailed(t *testing.T) {
	h := &PreCompactSave{}

	workState := map[string]interface{}{
		"review_status":  "failed",
		"failure_reason": "critical issue found",
	}

	risks := h.buildOpenRisks(nil, nil, workState, nil)

	found := false
	for _, r := range risks {
		if r.Kind == "review" && r.Severity == "high" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected high severity review risk when review_status=failed, got risks: %+v", risks)
	}
}

// TestBuildOpenRisks_ReviewStatusPending verifies that a medium-severity review risk
// is added when review_status="pending".
func TestBuildOpenRisks_ReviewStatusPending(t *testing.T) {
	h := &PreCompactSave{}

	workState := map[string]interface{}{
		"review_status": "pending",
	}

	risks := h.buildOpenRisks(nil, nil, workState, nil)

	found := false
	for _, r := range risks {
		if r.Kind == "review" && r.Severity == "medium" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected medium severity review risk when review_status=pending, got risks: %+v", risks)
	}
}

// TestBuildFailedChecks_SingleObjectFailedChecks verifies that when failed_checks is
// a single map object, it is wrapped in an array for processing (symmetric with JS version).
func TestBuildFailedChecks_SingleObjectFailedChecks(t *testing.T) {
	h := &PreCompactSave{}

	workState := map[string]interface{}{
		"failed_checks": map[string]interface{}{
			"check":  "type-check",
			"status": "failed",
			"detail": "type mismatch at line 42",
		},
	}

	checks := h.buildFailedChecks(workState, nil)

	if len(checks) != 1 {
		t.Fatalf("expected 1 check, got %d: %+v", len(checks), checks)
	}
	if checks[0].Check != "type-check" {
		t.Errorf("expected check=type-check, got: %s", checks[0].Check)
	}
	if checks[0].Status != "failed" {
		t.Errorf("expected status=failed, got: %s", checks[0].Status)
	}
	if checks[0].Detail != "type mismatch at line 42" {
		t.Errorf("expected detail to be set, got: %s", checks[0].Detail)
	}
}

// TestBuildFailedChecks_ArrayFailedChecks verifies that when failed_checks is an array,
// all entries are processed as before (regression test).
func TestBuildFailedChecks_ArrayFailedChecks(t *testing.T) {
	h := &PreCompactSave{}

	workState := map[string]interface{}{
		"failed_checks": []interface{}{
			map[string]interface{}{
				"check":  "lint",
				"status": "failed",
			},
			map[string]interface{}{
				"check":  "test",
				"status": "failed",
			},
		},
	}

	checks := h.buildFailedChecks(workState, nil)

	if len(checks) != 2 {
		t.Fatalf("expected 2 checks, got %d: %+v", len(checks), checks)
	}
}

// TestPreCompactSave_CustomPlansDirectory verifies that when plansDirectory is configured,
// Plans.md in the custom directory is read via resolvePlansPath (P1 fix).
func TestPreCompactSave_CustomPlansDirectory(t *testing.T) {
	dir := t.TempDir()

	// set plansDirectory: workspace in the config file
	configContent := "plansDirectory: workspace\n"
	configPath := filepath.Join(dir, harnessConfigFileName)
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// place Plans.md in the workspace/ directory
	workspaceDir := filepath.Join(dir, "workspace")
	if err := os.MkdirAll(workspaceDir, 0o755); err != nil {
		t.Fatal(err)
	}
	plansContent := "| 1 | Custom Dir Task | DoD | none | `cc:WIP` |\n"
	if err := os.WriteFile(filepath.Join(workspaceDir, "Plans.md"), []byte(plansContent), 0o644); err != nil {
		t.Fatal(err)
	}

	stateDir := filepath.Join(dir, ".claude", "state")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// leave PlansFile empty to delegate to resolvePlansPath (PlansFile="" → use config)
	h := &PreCompactSave{
		RepoRoot: dir,
		StateDir: stateDir,
		// intentionally not setting PlansFile
	}

	var buf bytes.Buffer
	if err := h.Handle(strings.NewReader("{}"), &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	artifactPath := filepath.Join(stateDir, "handoff-artifact.json")
	data, err := os.ReadFile(artifactPath)
	if err != nil {
		t.Fatalf("handoff-artifact.json not found: %v", err)
	}

	var artifact handoffArtifact
	if err := json.Unmarshal(data, &artifact); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// verify WIP tasks are detected (proof that Plans.md in the custom directory was read)
	if len(artifact.WIPTasks) == 0 {
		t.Error("expected WIP tasks from custom plansDirectory, got none")
	}
}
