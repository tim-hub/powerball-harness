// Package plans reads and queries the canonical task SSOT,
// .claude/harness/plans.json. It is the Go counterpart to the plans.json query
// helpers in harness/scripts/config-utils.sh, shared by the hook handlers and
// session reporters that previously parsed the legacy Plans.md markdown file.
package plans

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
)

// Plans is the root structure of .claude/harness/plans.json.
type Plans struct {
	Project              string   `json:"project"`
	Meta                 Meta     `json:"meta"`
	Phases               []Phase  `json:"phases"`
	FutureConsiderations []string `json:"futureConsiderations"`
}

// Meta holds project-level release metadata.
type Meta struct {
	LastRelease            string `json:"lastRelease"`
	LastReleaseDate        string `json:"lastReleaseDate"`
	LastReleaseDescription string `json:"lastReleaseDescription"`
}

// Phase is a planning phase containing tasks.
type Phase struct {
	ID         int       `json:"id"`
	Title      string    `json:"title"`
	Created    string    `json:"created"`
	Goal       string    `json:"goal"`
	Status     string    `json:"status"` // "active" | "archived"
	Urgency    string    `json:"urgency"`
	Importance string    `json:"importance"`
	Comments   []Comment `json:"comments"`
	Tasks      []Task    `json:"tasks"`
}

// Task is a single work item within a phase.
type Task struct {
	ID             string       `json:"id"`
	Name           string       `json:"name"`
	Description    string       `json:"description"`
	DoD            string       `json:"dod"`
	Depends        []string     `json:"depends"`
	Status         string       `json:"status"` // cc:TODO | cc:WIP | cc:done | pm:confirmed | pm:requested | blocked
	StatusHash     string       `json:"statusHash,omitempty"`
	BlockedReason  string       `json:"blockedReason,omitempty"`
	Urgency        string       `json:"urgency"`
	Importance     string       `json:"importance"`
	QualityMarkers []string     `json:"qualityMarkers"`
	Ralph          *RalphConfig `json:"ralph,omitempty"`
	Comments       []Comment    `json:"comments"`
}

// RalphConfig holds configuration for [ralph] iterative-loop tasks.
type RalphConfig struct {
	Verify  string `json:"verify"`
	MaxIter int    `json:"maxIter"`
}

// Comment is a timestamped annotation on a phase or task.
type Comment struct {
	ID         string `json:"id"`
	Author     string `json:"author"`
	AuthorName string `json:"authorName"`
	At         string `json:"at"`
	Text       string `json:"text"`
}

// DoneStatuses are the task statuses that count as "done" for dependency and
// completion checks.
var DoneStatuses = map[string]bool{
	"cc:done":      true,
	"pm:confirmed": true,
}

// DefaultRelPath is the canonical plans.json location relative to a plans dir.
const DefaultRelPath = ".claude/harness/plans.json"

// ResolvePath returns the full path to plans.json under projectRoot. plansDir is
// the configured plansDirectory ("" means project root). It returns the path
// only when the file exists; otherwise it returns "".
func ResolvePath(projectRoot, plansDir string) string {
	base := projectRoot
	if plansDir != "" && plansDir != "." {
		base = filepath.Join(projectRoot, plansDir)
	}
	full := filepath.Join(base, ".claude", "harness", "plans.json")
	if _, err := os.Stat(full); err == nil {
		return full
	}
	return ""
}

// Load reads and parses plans.json at path. It returns (nil, nil) when the file
// does not exist so callers can treat "no plan" as a non-error empty state.
func Load(path string) (*Plans, error) {
	if path == "" {
		return nil, nil
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var p Plans
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// LoadFrom resolves plans.json under projectRoot/plansDir and loads it. Returns
// (nil, nil) when no plans.json exists.
func LoadFrom(projectRoot, plansDir string) (*Plans, error) {
	return Load(ResolvePath(projectRoot, plansDir))
}

// AllTasks returns every task across all phases (a flat copy of references).
func (p *Plans) AllTasks() []Task {
	if p == nil {
		return nil
	}
	var out []Task
	for _, ph := range p.Phases {
		out = append(out, ph.Tasks...)
	}
	return out
}

// CountStatus counts tasks whose status equals the given value.
func (p *Plans) CountStatus(status string) int {
	if p == nil {
		return 0
	}
	n := 0
	for _, ph := range p.Phases {
		for _, t := range ph.Tasks {
			if t.Status == status {
				n++
			}
		}
	}
	return n
}

// HasWIP reports whether any task is cc:WIP.
func (p *Plans) HasWIP() bool {
	return p.CountStatus("cc:WIP") > 0
}

// WIPNames returns the names of cc:WIP tasks in phase order.
func (p *Plans) WIPNames() []string {
	if p == nil {
		return nil
	}
	var out []string
	for _, ph := range p.Phases {
		for _, t := range ph.Tasks {
			if t.Status == "cc:WIP" {
				out = append(out, t.Name)
			}
		}
	}
	return out
}

// WIPHasSkipTDD reports whether any cc:WIP task carries the skip:tdd marker.
func (p *Plans) WIPHasSkipTDD() bool {
	if p == nil {
		return false
	}
	for _, ph := range p.Phases {
		for _, t := range ph.Tasks {
			if t.Status != "cc:WIP" {
				continue
			}
			if slices.Contains(t.QualityMarkers, "skip:tdd") {
				return true
			}
		}
	}
	return false
}

// FindTask returns the task with the given id and its owning phase, or
// (nil, nil) if not found.
func (p *Plans) FindTask(id string) (*Task, *Phase) {
	if p == nil {
		return nil, nil
	}
	for pi := range p.Phases {
		for ti := range p.Phases[pi].Tasks {
			if p.Phases[pi].Tasks[ti].ID == id {
				return &p.Phases[pi].Tasks[ti], &p.Phases[pi]
			}
		}
	}
	return nil, nil
}
