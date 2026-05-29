package plancli_test

import (
	"os"
	"path/filepath"
	"testing"

	plancli "github.com/tim-hub/powerball-harness/go/cmd/harness/plan-cli"
)

func TestPlanRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "plans.json")

	original := &plancli.Plans{
		Project: "test-project",
		Meta: plancli.PlansMeta{
			LastRelease:            "v1.0.0",
			LastReleaseDate:        "2026-01-01",
			LastReleaseDescription: "initial release",
		},
		FutureConsiderations: []string{"consider A", "consider B"},
		Phases: []plancli.Phase{
			{
				ID:         1,
				Title:      "Phase One",
				Created:    "2026-01-01",
				Goal:       "accomplish something",
				Status:     "active",
				Urgency:    "high",
				Importance: "medium",
				Comments: []plancli.Comment{
					{ID: "c1", Author: "human", AuthorName: "Alice", At: "2026-01-01T10:00:00Z", Text: "looks good"},
				},
				Tasks: []plancli.Task{
					{
						ID:             "1.1",
						Name:           "first task",
						Description:    "do the thing",
						DoD:            "thing is done",
						Depends:        []string{},
						Status:         "cc:done",
						StatusHash:     "abc1234",
						Urgency:        "high",
						Importance:     "high",
						QualityMarkers: []string{"feature:security"},
						Comments:       []plancli.Comment{},
					},
					{
						ID:             "1.2",
						Name:           "ralph task",
						Description:    "iterate [ralph]",
						DoD:            "tests pass",
						Depends:        []string{"1.1"},
						Status:         "cc:TODO",
						Urgency:        "medium",
						Importance:     "medium",
						QualityMarkers: []string{"ralph"},
						Ralph:          &plancli.RalphConfig{Verify: "npm test", MaxIter: 10},
						Comments:       []plancli.Comment{},
					},
				},
			},
		},
	}

	if err := plancli.SavePlans(path, original); err != nil {
		t.Fatalf("SavePlans: %v", err)
	}

	loaded, err := plancli.LoadPlans(path)
	if err != nil {
		t.Fatalf("LoadPlans: %v", err)
	}

	// Project-level fields
	if loaded.Project != original.Project {
		t.Errorf("project: got %q want %q", loaded.Project, original.Project)
	}
	if loaded.Meta.LastRelease != original.Meta.LastRelease {
		t.Errorf("meta.lastRelease: got %q want %q", loaded.Meta.LastRelease, original.Meta.LastRelease)
	}
	if len(loaded.FutureConsiderations) != len(original.FutureConsiderations) {
		t.Errorf("futureConsiderations len: got %d want %d", len(loaded.FutureConsiderations), len(original.FutureConsiderations))
	}

	// Phase fields
	if len(loaded.Phases) != 1 {
		t.Fatalf("phases len: got %d want 1", len(loaded.Phases))
	}
	ph := loaded.Phases[0]
	if ph.ID != 1 || ph.Title != "Phase One" || ph.Status != "active" {
		t.Errorf("phase fields mismatch: %+v", ph)
	}
	if len(ph.Comments) != 1 || ph.Comments[0].Text != "looks good" {
		t.Errorf("phase comments: %+v", ph.Comments)
	}

	// Task fields
	if len(ph.Tasks) != 2 {
		t.Fatalf("tasks len: got %d want 2", len(ph.Tasks))
	}
	t1 := ph.Tasks[0]
	if t1.ID != "1.1" || t1.StatusHash != "abc1234" {
		t.Errorf("task[0] fields: %+v", t1)
	}
	if len(t1.QualityMarkers) != 1 || t1.QualityMarkers[0] != "feature:security" {
		t.Errorf("task[0] qualityMarkers: %+v", t1.QualityMarkers)
	}

	t2 := ph.Tasks[1]
	if t2.Ralph == nil || t2.Ralph.Verify != "npm test" || t2.Ralph.MaxIter != 10 {
		t.Errorf("task[1] ralph: %+v", t2.Ralph)
	}
}

func TestLoadPlans_NotExist(t *testing.T) {
	p, err := plancli.LoadPlans("/nonexistent/path/plans.json")
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
	if p == nil {
		t.Fatal("expected non-nil Plans for missing file")
	}
}

func TestEnsurePlansDir(t *testing.T) {
	dir := t.TempDir()
	path, err := plancli.EnsurePlansDir(dir)
	if err != nil {
		t.Fatalf("EnsurePlansDir: %v", err)
	}
	expected := filepath.Join(dir, ".claude", "harness", "plans.json")
	if path != expected {
		t.Errorf("path: got %q want %q", path, expected)
	}
	// Dir must exist
	if _, err := os.Stat(filepath.Dir(path)); err != nil {
		t.Errorf(".claude/harness dir not created: %v", err)
	}
	// Idempotent second call
	if _, err := plancli.EnsurePlansDir(dir); err != nil {
		t.Errorf("second EnsurePlansDir: %v", err)
	}
}

func TestSavePlans_Atomic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "plans.json")

	p := &plancli.Plans{Project: "atomic-test"}
	if err := plancli.SavePlans(path, p); err != nil {
		t.Fatalf("SavePlans: %v", err)
	}
	// .tmp file must not remain
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Error(".tmp file should not exist after successful save")
	}
}
