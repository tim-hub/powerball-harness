package plancli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Plans is the root structure for .claude/harness/plans.json.
type Plans struct {
	Project              string      `json:"project"`
	Meta                 PlansMeta   `json:"meta"`
	Phases               []Phase     `json:"phases"`
	FutureConsiderations []string    `json:"futureConsiderations"`
}

// PlansMeta holds project-level release metadata.
type PlansMeta struct {
	LastRelease            string `json:"lastRelease"`
	LastReleaseDate        string `json:"lastReleaseDate"`
	LastReleaseDescription string `json:"lastReleaseDescription"`
}

// Phase represents a planning phase containing tasks.
type Phase struct {
	ID         int       `json:"id"`
	Title      string    `json:"title"`
	Created    string    `json:"created"`
	Goal       string    `json:"goal"`
	Status     string    `json:"status"` // "active" | "archived"
	Urgency    string    `json:"urgency"`    // "low" | "medium" | "high"
	Importance string    `json:"importance"` // "low" | "medium" | "high"
	Comments   []Comment `json:"comments"`
	Tasks      []Task    `json:"tasks"`
}

// Task represents a single work item within a phase.
type Task struct {
	ID             string      `json:"id"`
	Name           string      `json:"name"`
	Description    string      `json:"description"`
	DoD            string      `json:"dod"`
	Depends        []string    `json:"depends"`
	Status         string      `json:"status"` // "cc:TODO" | "cc:WIP" | "cc:done" | "pm:confirmed" | "pm:requested" | "blocked"
	StatusHash     string      `json:"statusHash,omitempty"`
	BlockedReason  string      `json:"blockedReason,omitempty"`
	Urgency        string      `json:"urgency"`    // "low" | "medium" | "high"
	Importance     string      `json:"importance"` // "low" | "medium" | "high"
	QualityMarkers []string    `json:"qualityMarkers"`
	Ralph          *RalphConfig `json:"ralph,omitempty"`
	Comments       []Comment   `json:"comments"`
}

// UnmarshalJSON normalises nil slice fields to empty slices so that
// json.Marshal always emits [] instead of null, and nil checks are safe.
func (t *Task) UnmarshalJSON(data []byte) error {
	type taskAlias Task // break infinite recursion
	var a taskAlias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*t = Task(a)
	if t.Depends == nil {
		t.Depends = []string{}
	}
	if t.QualityMarkers == nil {
		t.QualityMarkers = []string{}
	}
	if t.Comments == nil {
		t.Comments = []Comment{}
	}
	return nil
}

// RalphConfig holds configuration for [ralph] iterative loop tasks.
type RalphConfig struct {
	Verify  string `json:"verify"`
	MaxIter int    `json:"maxIter"`
}

// Comment is a timestamped annotation on a phase or task.
type Comment struct {
	ID         string `json:"id"`
	Author     string `json:"author"`     // "human" | "agent"
	AuthorName string `json:"authorName"`
	At         string `json:"at"` // ISO-8601
	Text       string `json:"text"`
}

// DefaultPlansPath returns the canonical path to plans.json for the given
// project root. The file lives at <root>/.claude/harness/plans.json.
func DefaultPlansPath(projectRoot string) string {
	return filepath.Join(projectRoot, ".claude", "harness", "plans.json")
}

// EnsurePlansDir creates .claude/harness/ under projectRoot if absent and
// returns the path to plans.json.
func EnsurePlansDir(projectRoot string) (string, error) {
	dir := filepath.Join(projectRoot, ".claude", "harness")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create plans dir: %w", err)
	}
	return filepath.Join(dir, "plans.json"), nil
}

// LoadPlans reads and parses plans.json at the given path.
// Returns an empty Plans struct if the file does not exist.
func LoadPlans(path string) (*Plans, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Plans{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read plans.json: %w", err)
	}
	var p Plans
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("parse plans.json: %w", err)
	}
	normalizePlans(&p)
	return &p, nil
}

// normalizePlans ensures that slice fields on every Task are non-nil so that
// json.Marshal always emits [] instead of null and callers can range safely.
func normalizePlans(p *Plans) {
	for i := range p.Phases {
		for j := range p.Phases[i].Tasks {
			t := &p.Phases[i].Tasks[j]
			if t.Depends == nil {
				t.Depends = []string{}
			}
			if t.QualityMarkers == nil {
				t.QualityMarkers = []string{}
			}
			if t.Comments == nil {
				t.Comments = []Comment{}
			}
		}
	}
}

// SavePlans writes p to path atomically: it writes to a .tmp file then
// renames it to path so readers never observe a partial write.
func SavePlans(path string, p *Plans) error {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal plans: %w", err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("write plans tmp: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("rename plans tmp: %w", err)
	}
	return nil
}
