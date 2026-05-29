package plancli

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// splitTableRow
// ---------------------------------------------------------------------------

func TestSplitTableRow_Simple(t *testing.T) {
	cols := splitTableRow("| col1 | col2 | col3 |")
	if len(cols) != 3 {
		t.Fatalf("want 3 cols, got %d: %v", len(cols), cols)
	}
	if cols[0] != "col1" || cols[1] != "col2" || cols[2] != "col3" {
		t.Errorf("cols: %v", cols)
	}
}

func TestSplitTableRow_PipeInBacktick(t *testing.T) {
	// Pipes inside backtick spans must not split the column.
	cols := splitTableRow("| id | `a | b` | status |")
	if len(cols) != 3 {
		t.Fatalf("want 3 cols, got %d: %v", len(cols), cols)
	}
	if cols[1] != "`a | b`" {
		t.Errorf("col[1]: got %q, want `a | b`", cols[1])
	}
}

// ---------------------------------------------------------------------------
// parseStatus
// ---------------------------------------------------------------------------

func TestParseStatus(t *testing.T) {
	cases := []struct {
		raw         string
		wantStatus  string
		wantHash    string
		wantBlocked string
	}{
		{"cc:TODO", "cc:TODO", "", ""},
		{"cc:WIP", "cc:WIP", "", ""},
		{"cc:done", "cc:done", "", ""},
		{"cc:done [abc1234]", "cc:done", "abc1234", ""},
		{"pm:confirmed", "pm:confirmed", "", ""},
		{"blocked (waiting on CI)", "blocked", "", "waiting on CI"},
		{"blocked", "blocked", "", ""},
		{"-", "cc:TODO", "", ""},
		{"", "cc:TODO", "", ""},
	}
	for _, c := range cases {
		s, h, b := parseStatus(c.raw)
		if s != c.wantStatus {
			t.Errorf("parseStatus(%q) status = %q, want %q", c.raw, s, c.wantStatus)
		}
		if h != c.wantHash {
			t.Errorf("parseStatus(%q) hash = %q, want %q", c.raw, h, c.wantHash)
		}
		if b != c.wantBlocked {
			t.Errorf("parseStatus(%q) blocked = %q, want %q", c.raw, b, c.wantBlocked)
		}
	}
}

// ---------------------------------------------------------------------------
// parseLastReleaseLine
// ---------------------------------------------------------------------------

func TestParseLastReleaseLine(t *testing.T) {
	meta := parseLastReleaseLine("v5.8.0 on 2026-05-26 (skill consolidation shipped)")
	if meta.LastRelease != "v5.8.0" {
		t.Errorf("release: %q", meta.LastRelease)
	}
	if meta.LastReleaseDate != "2026-05-26" {
		t.Errorf("date: %q", meta.LastReleaseDate)
	}
	if !strings.Contains(meta.LastReleaseDescription, "skill consolidation") {
		t.Errorf("desc: %q", meta.LastReleaseDescription)
	}
}

// ---------------------------------------------------------------------------
// parsePlansMD — minimal fixture
// ---------------------------------------------------------------------------

var minimalPlansMD = `# My Project — Plans.md

Last release: v1.0.0 on 2026-01-01 (initial)

---

## Phase 2: Second Phase

Created: 2026-02-01

**Goal**: accomplish second thing

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 2.1 | Do alpha | thing done | - | cc:TODO |
| 2.2 | Do beta [skip:tdd] | beta done | 2.1 | cc:done [abc1234] |
| 2.3 | Do gamma | gamma done | 2.1,2.2 | blocked (need approval) |

---

## Phase 1: First Phase

Created: 2026-01-15

**Goal**: accomplish something

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 1.1 | First task | done | - | pm:confirmed |
`

func TestParsePlansMD_Minimal(t *testing.T) {
	plans, err := parsePlansMD(minimalPlansMD)
	if err != nil {
		t.Fatalf("parsePlansMD: %v", err)
	}

	if plans.Project != "My Project" {
		t.Errorf("project: %q", plans.Project)
	}
	if plans.Meta.LastRelease != "v1.0.0" {
		t.Errorf("lastRelease: %q", plans.Meta.LastRelease)
	}

	if len(plans.Phases) != 2 {
		t.Fatalf("phases: got %d, want 2", len(plans.Phases))
	}

	ph2 := plans.Phases[0]
	if ph2.ID != 2 {
		t.Errorf("phase[0].id: %d", ph2.ID)
	}
	if ph2.Title != "Second Phase" {
		t.Errorf("phase[0].title: %q", ph2.Title)
	}
	if ph2.Created != "2026-02-01" {
		t.Errorf("phase[0].created: %q", ph2.Created)
	}
	if ph2.Goal == "" {
		t.Error("phase[0].goal should not be empty")
	}
	if len(ph2.Tasks) != 3 {
		t.Fatalf("phase[0] tasks: got %d, want 3", len(ph2.Tasks))
	}

	t1 := ph2.Tasks[0]
	if t1.ID != "2.1" || t1.Status != "cc:TODO" {
		t.Errorf("task 2.1: id=%q status=%q", t1.ID, t1.Status)
	}
	if len(t1.Depends) != 0 {
		t.Errorf("task 2.1 depends: %v (want empty)", t1.Depends)
	}

	t2 := ph2.Tasks[1]
	if t2.ID != "2.2" || t2.Status != "cc:done" || t2.StatusHash != "abc1234" {
		t.Errorf("task 2.2: status=%q hash=%q", t2.Status, t2.StatusHash)
	}
	if len(t2.QualityMarkers) != 1 || t2.QualityMarkers[0] != "skip:tdd" {
		t.Errorf("task 2.2 markers: %v", t2.QualityMarkers)
	}
	if len(t2.Depends) != 1 || t2.Depends[0] != "2.1" {
		t.Errorf("task 2.2 depends: %v", t2.Depends)
	}

	t3 := ph2.Tasks[2]
	if t3.ID != "2.3" || t3.Status != "blocked" {
		t.Errorf("task 2.3: status=%q", t3.Status)
	}
	if !strings.Contains(t3.BlockedReason, "need approval") {
		t.Errorf("task 2.3 blockedReason: %q", t3.BlockedReason)
	}
	if len(t3.Depends) != 2 {
		t.Errorf("task 2.3 depends: %v (want 2)", t3.Depends)
	}

	ph1 := plans.Phases[1]
	if ph1.ID != 1 || len(ph1.Tasks) != 1 {
		t.Errorf("phase[1]: id=%d tasks=%d", ph1.ID, len(ph1.Tasks))
	}
	if ph1.Tasks[0].Status != "pm:confirmed" {
		t.Errorf("task 1.1 status: %q", ph1.Tasks[0].Status)
	}
}

func TestParsePlansMD_BacktickPipeInDod(t *testing.T) {
	// Ensure pipes inside backtick spans in DoD column don't corrupt parsing.
	content := `# Test — Plans.md

## Phase 1: Test

Created: 2026-01-01

**Goal**: test

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 1.1 | Do thing | ` + "`" + `jq '.x | length'` + "`" + ` returns 5 | - | cc:TODO |
`
	plans, err := parsePlansMD(content)
	if err != nil {
		t.Fatalf("parsePlansMD: %v", err)
	}
	if len(plans.Phases) != 1 || len(plans.Phases[0].Tasks) != 1 {
		t.Fatalf("expected 1 phase with 1 task")
	}
	task := plans.Phases[0].Tasks[0]
	if task.ID != "1.1" {
		t.Errorf("task id: %q", task.ID)
	}
	if task.Status != "cc:TODO" {
		t.Errorf("task status: %q", task.Status)
	}
	if !strings.Contains(task.DoD, "jq") {
		t.Errorf("DoD should contain jq reference: %q", task.DoD)
	}
}

// ---------------------------------------------------------------------------
// Integration test against actual Plans.md
// ---------------------------------------------------------------------------

func TestActualPlansJSON(t *testing.T) {
	// Plans.md was migrated to plans.json in Phase 108; validate the SSOT.
	// This package is at go/cmd/harness/plan-cli/; plans.json is at .claude/harness/plans.json.
	data, err := os.ReadFile("../../../../.claude/harness/plans.json")
	if err != nil {
		t.Fatalf("cannot read .claude/harness/plans.json: %v", err)
	}

	var plans Plans
	if err := json.Unmarshal(data, &plans); err != nil {
		t.Fatalf("json.Unmarshal plans.json: %v", err)
	}

	if len(plans.Phases) == 0 {
		t.Fatal("should have at least one phase")
	}

	// All tasks must have valid IDs and non-empty statuses.
	totalTasks := 0
	for _, ph := range plans.Phases {
		if ph.ID == 0 {
			t.Errorf("phase has zero ID: %q", ph.Title)
		}
		if ph.Status != "active" && ph.Status != "archived" {
			t.Errorf("phase %d has unexpected status: %q", ph.ID, ph.Status)
		}
		for _, task := range ph.Tasks {
			totalTasks++
			if task.ID == "" {
				t.Errorf("task has empty ID in phase %d", ph.ID)
			}
			if task.Status == "" {
				t.Errorf("task %s has empty status", task.ID)
			}
			if task.Depends == nil {
				t.Errorf("task %s has nil depends (want [])", task.ID)
			}
			if task.QualityMarkers == nil {
				t.Errorf("task %s has nil qualityMarkers (want [])", task.ID)
			}
			if task.Comments == nil {
				t.Errorf("task %s has nil comments (want [])", task.ID)
			}
		}
	}
	if totalTasks == 0 {
		t.Error("expected at least one task across all phases")
	}
	t.Logf("parsed %d phases, %d total tasks", len(plans.Phases), totalTasks)
}
