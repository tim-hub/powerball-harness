package hookhandler

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// sprintTaskFixture describes a single plans.json task for sprint-contract tests.
type sprintTaskFixture struct {
	ID             string
	Name           string
	DoD            string
	Depends        []string
	Status         string
	QualityMarkers []string
}

// writeSprintPlansJSON writes a plans.json containing the given tasks under
// dir/.claude/harness/plans.json and returns its absolute path.
func writeSprintPlansJSON(t *testing.T, dir string, tasks ...sprintTaskFixture) string {
	t.Helper()
	hdir := filepath.Join(dir, ".claude", "harness")
	if err := os.MkdirAll(hdir, 0o755); err != nil {
		t.Fatal(err)
	}
	type jtask struct {
		ID             string   `json:"id"`
		Name           string   `json:"name"`
		DoD            string   `json:"dod"`
		Depends        []string `json:"depends"`
		Status         string   `json:"status"`
		QualityMarkers []string `json:"qualityMarkers"`
	}
	var jtasks []jtask
	for _, tk := range tasks {
		depends := tk.Depends
		if depends == nil {
			depends = []string{}
		}
		markers := tk.QualityMarkers
		if markers == nil {
			markers = []string{}
		}
		jtasks = append(jtasks, jtask{
			ID:             tk.ID,
			Name:           tk.Name,
			DoD:            tk.DoD,
			Depends:        depends,
			Status:         tk.Status,
			QualityMarkers: markers,
		})
	}
	doc := map[string]interface{}{
		"phases": []map[string]interface{}{
			{"id": 1, "title": "P1", "status": "active", "tasks": jtasks},
		},
	}
	data, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(hdir, "plans.json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestSprintContractGenerator_RuntimeContract(t *testing.T) {
	dir := t.TempDir()
	packageJSONPath := filepath.Join(dir, "package.json")
	plansPath := writeSprintPlansJSON(t, dir, sprintTaskFixture{
		ID: "32.1.1", Name: "create contract", DoD: "put runtime validation in contract",
		Depends: []string{"32.0.1"}, Status: "cc:TODO",
	})
	if err := os.WriteFile(packageJSONPath, []byte(`{"scripts":{"test":"vitest run","test:e2e":"playwright test"},"devDependencies":{"@playwright/test":"^1.52.0"}}`), 0o600); err != nil {
		t.Fatal(err)
	}

	g := &SprintContractGenerator{
		ProjectRoot: dir,
		PlansFile:   plansPath,
		Now:         func() string { return "2026-04-16T00:00:00Z" },
	}
	doc, err := g.Generate("32.1.1")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if doc.SchemaVersion != "sprint-contract.v1" {
		t.Fatalf("unexpected schema version: %s", doc.SchemaVersion)
	}
	if doc.Review.ReviewerProfile != "runtime" {
		t.Fatalf("expected runtime profile, got %s", doc.Review.ReviewerProfile)
	}
	if len(doc.Contract.RuntimeValidation) == 0 || doc.Contract.RuntimeValidation[0].Command != "CI=true npm test" {
		t.Fatalf("unexpected runtime validation: %+v", doc.Contract.RuntimeValidation)
	}
	if !doc.Advisor.Enabled || doc.Advisor.Mode != "on-demand" {
		t.Fatalf("unexpected advisor defaults: %+v", doc.Advisor)
	}
	if doc.Advisor.MaxConsults != 3 || doc.Advisor.RetryThreshold != 2 || !doc.Advisor.PreEscalationConsult {
		t.Fatalf("unexpected advisor thresholds: %+v", doc.Advisor)
	}
	if doc.Advisor.ModelPolicy.ClaudeDefault != "opus" || doc.Advisor.ModelPolicy.CodexDefault != "gpt-5.4" {
		t.Fatalf("unexpected advisor model policy: %+v", doc.Advisor.ModelPolicy)
	}
	if len(doc.Advisor.Triggers) != 0 {
		t.Fatalf("expected no advisor triggers, got %+v", doc.Advisor.Triggers)
	}
}

func TestSprintContractGenerator_UIRubricDefaults(t *testing.T) {
	dir := t.TempDir()
	plansPath := writeSprintPlansJSON(t, dir, sprintTaskFixture{
		ID: "41.3.1", Name: "design-heavy task",
		DoD:     "polish UI layout with design and styling and aesthetic quality",
		Depends: []string{"41.2.1"}, Status: "cc:TODO",
	})

	g := &SprintContractGenerator{ProjectRoot: dir, PlansFile: plansPath}
	doc, err := g.Generate("41.3.1")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if doc.Review.ReviewerProfile != "ui-rubric" {
		t.Fatalf("expected ui-rubric, got %s", doc.Review.ReviewerProfile)
	}
	if doc.Review.MaxIterations != 10 {
		t.Fatalf("expected max_iterations=10, got %d", doc.Review.MaxIterations)
	}
	if doc.Review.RubricTarget == nil || doc.Review.RubricTarget.Design != 6 || doc.Review.RubricTarget.Functionality != 6 {
		t.Fatalf("unexpected rubric target: %+v", doc.Review.RubricTarget)
	}
}

func TestSprintContractGenerator_MaxIterationsHTMLOverride(t *testing.T) {
	dir := t.TempDir()
	plansPath := writeSprintPlansJSON(t, dir, sprintTaskFixture{
		ID: "T-html-comment", Name: "HTML comment task",
		DoD: "<!-- max_iterations: 15 --> specified in DoD", Status: "cc:TODO",
	})

	g := &SprintContractGenerator{ProjectRoot: dir, PlansFile: plansPath}
	doc, err := g.Generate("T-html-comment")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if doc.Review.MaxIterations != 15 {
		t.Fatalf("expected max_iterations=15, got %d", doc.Review.MaxIterations)
	}
}

func TestSprintContractGenerator_BrowserRouteRules(t *testing.T) {
	dir := t.TempDir()
	packageJSONPath := filepath.Join(dir, "package.json")
	plansPath := writeSprintPlansJSON(t, dir,
		sprintTaskFixture{
			ID: "scripted", Name: "add browser evaluator", DoD: "verify UI flow in browser",
			Depends: []string{"32.2.1"}, Status: "cc:TODO",
		},
		sprintTaskFixture{
			ID: "exploratory", Name: "handle browser_mode: exploratory",
			DoD: "prioritize AgentBrowser in exploratory mode", Depends: []string{"32.2.2"}, Status: "cc:TODO",
		},
	)
	if err := os.WriteFile(packageJSONPath, []byte(`{"scripts":{"test:e2e":"playwright test"},"devDependencies":{"@playwright/test":"^1.52.0"}}`), 0o600); err != nil {
		t.Fatal(err)
	}

	g := &SprintContractGenerator{ProjectRoot: dir, PlansFile: plansPath}

	scripted, err := g.Generate("scripted")
	if err != nil {
		t.Fatalf("Generate scripted: %v", err)
	}
	if scripted.Review.ReviewerProfile != "browser" {
		t.Fatalf("expected browser profile, got %s", scripted.Review.ReviewerProfile)
	}
	if scripted.Review.Route == nil || *scripted.Review.Route != "playwright" {
		t.Fatalf("expected scripted route=playwright, got %+v", scripted.Review.Route)
	}

	exploratory, err := g.Generate("exploratory")
	if err != nil {
		t.Fatalf("Generate exploratory: %v", err)
	}
	if exploratory.Review.BrowserMode == nil || *exploratory.Review.BrowserMode != "exploratory" {
		t.Fatalf("expected exploratory browser mode, got %+v", exploratory.Review.BrowserMode)
	}
	if exploratory.Review.Route != nil {
		t.Fatalf("expected exploratory route=nil, got %+v", exploratory.Review.Route)
	}
}

func TestSprintContractGenerator_AdvisorTriggers(t *testing.T) {
	dir := t.TempDir()
	plansPath := writeSprintPlansJSON(t, dir, sprintTaskFixture{
		ID: "43.1.1", Name: "security migration contract",
		DoD:            "verify state migration guard <!-- advisor:required -->",
		Status:         "cc:TODO",
		QualityMarkers: []string{"needs-spike"},
	})

	g := &SprintContractGenerator{ProjectRoot: dir, PlansFile: plansPath}
	doc, err := g.Generate("43.1.1")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	expected := []string{"needs-spike", "security-sensitive", "state-migration", "<!-- advisor:required -->"}
	if len(doc.Advisor.Triggers) != len(expected) {
		t.Fatalf("unexpected advisor triggers length: got=%v want=%v", doc.Advisor.Triggers, expected)
	}
	for i, trigger := range expected {
		if doc.Advisor.Triggers[i] != trigger {
			t.Fatalf("unexpected advisor trigger order: got=%v want=%v", doc.Advisor.Triggers, expected)
		}
	}
}

func TestSprintContractGenerator_WriteRoundTrip(t *testing.T) {
	dir := t.TempDir()
	plansPath := writeSprintPlansJSON(t, dir, sprintTaskFixture{
		ID: "32.1.1", Name: "create contract", DoD: "put runtime validation in contract",
		Depends: []string{"32.0.1"}, Status: "cc:TODO",
	})
	outputPath := filepath.Join(dir, "out", "32.1.1.sprint-contract.json")
	g := &SprintContractGenerator{ProjectRoot: dir, PlansFile: plansPath, OutputFile: outputPath}
	written, err := g.Write("32.1.1")
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if written != outputPath {
		t.Fatalf("unexpected output path: %s", written)
	}
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	var doc sprintContractDoc
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if doc.Task.ID != "32.1.1" {
		t.Fatalf("unexpected task id: %s", doc.Task.ID)
	}
}

// ---------------------------------------------------------------------------
// TDD sprint contract detection
// ---------------------------------------------------------------------------

func TestSprintContract_TDDRequired_Tag(t *testing.T) {
	dir := t.TempDir()
	plansPath := writeSprintPlansJSON(t, dir, sprintTaskFixture{
		ID: "97.1", Name: "Port R14", DoD: "cd go && go test -race ./... passes",
		Status: "cc:TODO", QualityMarkers: []string{"tdd:required"},
	})
	// Simulate a Go project (go.mod present)
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/test\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	g := &SprintContractGenerator{ProjectRoot: dir, PlansFile: plansPath}
	doc, err := g.Generate("97.1")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !doc.Contract.TDDRequired {
		t.Error("expected TDDRequired=true for [tdd:required] tag")
	}
	if doc.Contract.TestFramework != "go-test" {
		t.Errorf("expected TestFramework=go-test, got %q", doc.Contract.TestFramework)
	}
	if len(doc.Contract.TestTodoList) == 0 {
		t.Error("expected non-empty TestTodoList for TDD required")
	}
	if doc.Contract.SkipTDDReason != nil {
		t.Errorf("expected nil SkipTDDReason, got %q", *doc.Contract.SkipTDDReason)
	}
}

func TestSprintContract_TDDSkip_Tag(t *testing.T) {
	dir := t.TempDir()
	plansPath := writeSprintPlansJSON(t, dir, sprintTaskFixture{
		ID: "97.3", Name: "tdd-paths.yaml", DoD: "file exists and parses",
		Status: "cc:TODO", QualityMarkers: []string{"tdd:skip:config-only"},
	})

	g := &SprintContractGenerator{ProjectRoot: dir, PlansFile: plansPath}
	doc, err := g.Generate("97.3")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if doc.Contract.TDDRequired {
		t.Error("expected TDDRequired=false for [tdd:skip] tag")
	}
	if doc.Contract.SkipTDDReason == nil {
		t.Fatal("expected SkipTDDReason to be set for [tdd:skip] tag")
	}
	if *doc.Contract.SkipTDDReason != "config-only" {
		t.Errorf("expected skip reason %q, got %q", "config-only", *doc.Contract.SkipTDDReason)
	}
}

func TestSprintContract_TDD_NotRequired_WhenNoTag(t *testing.T) {
	dir := t.TempDir()
	plansPath := writeSprintPlansJSON(t, dir, sprintTaskFixture{
		ID: "98.1", Name: "Add Step 1.5 memory check", DoD: "harness-plan create.md updated",
		Status: "cc:TODO",
	})

	g := &SprintContractGenerator{ProjectRoot: dir, PlansFile: plansPath}
	doc, err := g.Generate("98.1")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if doc.Contract.TDDRequired {
		t.Error("expected TDDRequired=false when no TDD tag present")
	}
	if doc.Contract.SkipTDDReason != nil {
		t.Errorf("expected nil SkipTDDReason, got %q", *doc.Contract.SkipTDDReason)
	}
}
