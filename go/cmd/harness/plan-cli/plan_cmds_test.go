package plancli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Pure helper tests (no file I/O, no os.Exit)
// ---------------------------------------------------------------------------

func TestPhaseMatchesStatus(t *testing.T) {
	active := Phase{Status: "active"}
	archived := Phase{Status: "archived"}
	empty := Phase{Status: ""}

	cases := []struct {
		phase  Phase
		filter string
		want   bool
	}{
		{active, "active", true},
		{active, "", true},
		{active, "all", true},
		{active, "archived", false},
		{active, "cc:done", true}, // task-level filter keeps all phases
		{archived, "archived", true},
		{archived, "active", false},
		{archived, "all", true},
		{empty, "active", true},
		{empty, "", true},
	}
	for _, c := range cases {
		got := phaseMatchesStatus(c.phase, c.filter)
		if got != c.want {
			t.Errorf("phaseMatchesStatus(status=%q, filter=%q) = %v, want %v",
				c.phase.Status, c.filter, got, c.want)
		}
	}
}

func TestTaskMatchesStatus(t *testing.T) {
	todo := Task{Status: "cc:TODO"}
	done := Task{Status: "cc:done"}

	cases := []struct {
		task   Task
		filter string
		want   bool
	}{
		{todo, "all", true},
		{todo, "active", true},
		{todo, "", true},
		{todo, "archived", false},
		{todo, "cc:TODO", true},
		{todo, "cc:done", false},
		{done, "cc:done", true},
		{done, "cc:TODO", false},
	}
	for _, c := range cases {
		got := taskMatchesStatus(c.task, c.filter)
		if got != c.want {
			t.Errorf("taskMatchesStatus(status=%q, filter=%q) = %v, want %v",
				c.task.Status, c.filter, got, c.want)
		}
	}
}

func TestTaskHasMarker(t *testing.T) {
	task := Task{QualityMarkers: []string{"feature:security", "ralph"}}

	if !taskHasMarker(task, "ralph") {
		t.Error("expected taskHasMarker to find 'ralph'")
	}
	if !taskHasMarker(task, "feature:security") {
		t.Error("expected taskHasMarker to find 'feature:security'")
	}
	if taskHasMarker(task, "skip:tdd") {
		t.Error("expected taskHasMarker to NOT find 'skip:tdd'")
	}
	if taskHasMarker(Task{}, "ralph") {
		t.Error("expected taskHasMarker to return false for empty markers")
	}
}

func TestUrgencyDot(t *testing.T) {
	if urgencyDot("high") != "🔴" {
		t.Errorf("urgencyDot(high) unexpected: %s", urgencyDot("high"))
	}
	if urgencyDot("low") != "🟢" {
		t.Errorf("urgencyDot(low) unexpected: %s", urgencyDot("low"))
	}
	if urgencyDot("medium") != "🟡" {
		t.Errorf("urgencyDot(medium) unexpected: %s", urgencyDot("medium"))
	}
	if urgencyDot("") != "🟡" {
		t.Errorf("urgencyDot('') should default to medium dot")
	}
}

func TestImportanceStar(t *testing.T) {
	if importanceStar("high") != "★★" {
		t.Errorf("importanceStar(high) unexpected: %s", importanceStar("high"))
	}
	if importanceStar("low") != "☆" {
		t.Errorf("importanceStar(low) unexpected: %s", importanceStar("low"))
	}
	if importanceStar("medium") != "★" {
		t.Errorf("importanceStar(medium) unexpected: %s", importanceStar("medium"))
	}
}

// ---------------------------------------------------------------------------
// Command integration tests via t.Chdir
// ---------------------------------------------------------------------------

// setupPlanDir creates a temp dir with .claude/harness/ and writes the given
// Plans. It also calls t.Chdir so resolveProjectRoot picks up the temp dir.
func setupPlanDir(t *testing.T, p *Plans) string {
	t.Helper()
	dir := t.TempDir()
	t.Chdir(dir)
	path := filepath.Join(dir, ".claude", "harness", "plans.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := SavePlans(path, p); err != nil {
		t.Fatalf("SavePlans: %v", err)
	}
	return dir
}

func samplePlans() *Plans {
	return &Plans{
		Project: "test",
		Phases: []Phase{
			{
				ID:         1,
				Title:      "Alpha",
				Goal:       "do alpha",
				Status:     "active",
				Urgency:    "high",
				Importance: "medium",
				Comments:   []Comment{},
				Tasks: []Task{
					{
						ID:             "1.1",
						Name:           "first task",
						DoD:            "thing done",
						Depends:        []string{},
						Status:         "cc:TODO",
						Urgency:        "medium",
						Importance:     "medium",
						QualityMarkers: []string{},
						Comments:       []Comment{},
					},
				},
			},
		},
	}
}

func TestRunPlanAddPhase(t *testing.T) {
	dir := setupPlanDir(t, &Plans{Project: "test"})

	runPlanAddPhase([]string{
		"--title", "New Phase",
		"--goal", "achieve something",
		"--urgency", "high",
		"--importance", "low",
	})

	p, err := LoadPlans(filepath.Join(dir, ".claude", "harness", "plans.json"))
	if err != nil {
		t.Fatalf("LoadPlans: %v", err)
	}
	if len(p.Phases) != 1 {
		t.Fatalf("expected 1 phase, got %d", len(p.Phases))
	}
	ph := p.Phases[0]
	if ph.Title != "New Phase" {
		t.Errorf("title: got %q want %q", ph.Title, "New Phase")
	}
	if ph.Goal != "achieve something" {
		t.Errorf("goal mismatch")
	}
	if ph.Urgency != "high" {
		t.Errorf("urgency: got %q want high", ph.Urgency)
	}
	if ph.Importance != "low" {
		t.Errorf("importance: got %q want low", ph.Importance)
	}
	if ph.Status != "active" {
		t.Errorf("status: got %q want active", ph.Status)
	}
	if ph.ID != 1 {
		t.Errorf("id: got %d want 1", ph.ID)
	}
}

func TestRunPlanAddPhase_IncrementsID(t *testing.T) {
	initial := &Plans{
		Project: "test",
		Phases: []Phase{
			{ID: 3, Title: "Existing", Status: "active", Comments: []Comment{}, Tasks: []Task{}},
		},
	}
	dir := setupPlanDir(t, initial)

	runPlanAddPhase([]string{"--title", "Second", "--goal", "g"})

	p, _ := LoadPlans(filepath.Join(dir, ".claude", "harness", "plans.json"))
	if len(p.Phases) != 2 {
		t.Fatalf("expected 2 phases, got %d", len(p.Phases))
	}
	// newest phase is at index 0 (newest-first ordering)
	if p.Phases[0].ID != 4 {
		t.Errorf("new phase id: got %d want 4", p.Phases[0].ID)
	}
}

func TestRunPlanAddTask(t *testing.T) {
	dir := setupPlanDir(t, samplePlans())

	runPlanAddTask([]string{
		"1",
		"--name", "second task",
		"--dod", "tests pass",
		"--description", "do the thing",
		"--depends", "1.1",
		"--urgency", "high",
		"--importance", "high",
		"--marker", "feature:security,ralph",
	})

	p, _ := LoadPlans(filepath.Join(dir, ".claude", "harness", "plans.json"))
	tasks := p.Phases[0].Tasks
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}
	t2 := tasks[1]
	if t2.ID != "1.2" {
		t.Errorf("task id: got %q want 1.2", t2.ID)
	}
	if t2.Name != "second task" {
		t.Errorf("name: %q", t2.Name)
	}
	if t2.Status != "cc:TODO" {
		t.Errorf("status: %q", t2.Status)
	}
	if len(t2.Depends) != 1 || t2.Depends[0] != "1.1" {
		t.Errorf("depends: %v", t2.Depends)
	}
	if len(t2.QualityMarkers) != 2 {
		t.Errorf("markers: %v", t2.QualityMarkers)
	}
}

func TestRunPlanUpdate_Status(t *testing.T) {
	dir := setupPlanDir(t, samplePlans())

	runPlanUpdate([]string{"1.1", "--status", "cc:done", "--hash", "abc1234"})

	p, _ := LoadPlans(filepath.Join(dir, ".claude", "harness", "plans.json"))
	task := p.Phases[0].Tasks[0]
	if task.Status != "cc:done" {
		t.Errorf("status: got %q want cc:done", task.Status)
	}
	if task.StatusHash != "abc1234" {
		t.Errorf("hash: got %q want abc1234", task.StatusHash)
	}
}

func TestRunPlanUpdate_UrgencyImportance(t *testing.T) {
	dir := setupPlanDir(t, samplePlans())

	runPlanUpdate([]string{"1.1", "--urgency", "low", "--importance", "high"})

	p, _ := LoadPlans(filepath.Join(dir, ".claude", "harness", "plans.json"))
	task := p.Phases[0].Tasks[0]
	if task.Urgency != "low" {
		t.Errorf("urgency: got %q want low", task.Urgency)
	}
	if task.Importance != "high" {
		t.Errorf("importance: got %q want high", task.Importance)
	}
}

func TestRunPlanArchive(t *testing.T) {
	dir := setupPlanDir(t, samplePlans())

	runPlanArchive([]string{"1"})

	p, _ := LoadPlans(filepath.Join(dir, ".claude", "harness", "plans.json"))
	if p.Phases[0].Status != "archived" {
		t.Errorf("status: got %q want archived", p.Phases[0].Status)
	}
}

func TestRunPlanCommentOnPhase(t *testing.T) {
	dir := setupPlanDir(t, samplePlans())

	runPlanComment([]string{"1", "--text", "phase comment", "--author", "agent", "--author-name", "TestBot"})

	p, _ := LoadPlans(filepath.Join(dir, ".claude", "harness", "plans.json"))
	comments := p.Phases[0].Comments
	if len(comments) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(comments))
	}
	c := comments[0]
	if c.Text != "phase comment" {
		t.Errorf("text: %q", c.Text)
	}
	if c.Author != "agent" {
		t.Errorf("author: %q", c.Author)
	}
	if c.AuthorName != "TestBot" {
		t.Errorf("authorName: %q", c.AuthorName)
	}
	if c.ID == "" {
		t.Error("comment ID should be non-empty (UUID)")
	}
	if !strings.Contains(c.ID, "-") {
		t.Errorf("comment ID doesn't look like a UUID: %q", c.ID)
	}
}

func TestRunPlanCommentOnTask(t *testing.T) {
	dir := setupPlanDir(t, samplePlans())

	runPlanComment([]string{"1.1", "--text", "task comment"})

	p, _ := LoadPlans(filepath.Join(dir, ".claude", "harness", "plans.json"))
	comments := p.Phases[0].Tasks[0].Comments
	if len(comments) != 1 {
		t.Fatalf("expected 1 task comment, got %d", len(comments))
	}
	if comments[0].Text != "task comment" {
		t.Errorf("text: %q", comments[0].Text)
	}
	if comments[0].Author != "human" {
		t.Errorf("default author: %q", comments[0].Author)
	}
}

func TestRunPlanList_FilterByStatus(t *testing.T) {
	plans := &Plans{
		Project: "test",
		Phases: []Phase{
			{ID: 1, Title: "Active", Status: "active", Comments: []Comment{}, Tasks: []Task{}},
			{ID: 2, Title: "Archived", Status: "archived", Comments: []Comment{}, Tasks: []Task{}},
		},
	}
	dir := setupPlanDir(t, plans)
	_ = dir

	// Just verify the command doesn't panic/exit for valid inputs.
	// Output capture would require os.Stdout redirection.
	runPlanList([]string{"--status", "active"})
	runPlanList([]string{"--status", "archived"})
	runPlanList([]string{"--status", "all"})
}
