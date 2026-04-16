package hookhandler

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSprintContractGenerator_RuntimeContract(t *testing.T) {
	dir := t.TempDir()
	plansPath := filepath.Join(dir, "Plans.md")
	packageJSONPath := filepath.Join(dir, "package.json")
	if err := os.WriteFile(plansPath, []byte(`| Task | Content | DoD | Depends | Status |
|------|---------|-----|---------|--------|
| 32.1.1 | create contract | put runtime validation in contract | 32.0.1 | cc:TODO |
`), 0o600); err != nil {
		t.Fatal(err)
	}
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
	plansPath := filepath.Join(dir, "Plans.md")
	if err := os.WriteFile(plansPath, []byte(`| Task | Content | DoD | Depends | Status |
|------|---------|-----|---------|--------|
| 41.3.1 | design-heavy task | polish UI layout with design and styling and aesthetic quality | 41.2.1 | cc:TODO |
`), 0o600); err != nil {
		t.Fatal(err)
	}

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
	plansPath := filepath.Join(dir, "Plans.md")
	if err := os.WriteFile(plansPath, []byte(`| Task | Content | DoD | Depends | Status |
|------|---------|-----|---------|--------|
| T-html-comment | HTML comment task | <!-- max_iterations: 15 --> specified in DoD | - | cc:TODO |
`), 0o600); err != nil {
		t.Fatal(err)
	}

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
	plansPath := filepath.Join(dir, "Plans.md")
	packageJSONPath := filepath.Join(dir, "package.json")
	if err := os.WriteFile(plansPath, []byte(`| Task | Content | DoD | Depends | Status |
|------|---------|-----|---------|--------|
| scripted | add browser evaluator | verify UI flow in browser | 32.2.1 | cc:TODO |
| exploratory | handle browser_mode: exploratory | prioritize AgentBrowser in exploratory mode | 32.2.2 | cc:TODO |
`), 0o600); err != nil {
		t.Fatal(err)
	}
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
	plansPath := filepath.Join(dir, "Plans.md")
	if err := os.WriteFile(plansPath, []byte(`| Task | Content | DoD | Depends | Status |
|------|---------|-----|---------|--------|
| 43.1.1 | [needs-spike] security migration contract | verify state migration guard <!-- advisor:required --> | - | cc:TODO |
`), 0o600); err != nil {
		t.Fatal(err)
	}

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
	plansPath := filepath.Join(dir, "Plans.md")
	if err := os.WriteFile(plansPath, []byte(`| Task | Content | DoD | Depends | Status |
|------|---------|-----|---------|--------|
| 32.1.1 | create contract | put runtime validation in contract | 32.0.1 | cc:TODO |
`), 0o600); err != nil {
		t.Fatal(err)
	}
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
