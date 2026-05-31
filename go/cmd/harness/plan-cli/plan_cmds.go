package plancli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// list / ls
// ---------------------------------------------------------------------------

func runPlanList(args []string) {
	var (
		filterStatus     string
		filterUrgency    string
		filterImportance string
		filterMarker     string
		filterPhase      int
		pretty           bool
	)
	// Default to "all" so a bare `list` shows every phase regardless of status.
	// This keeps an empty result unambiguous: {"phases": null} means no plan
	// exists, never "all phases were filtered out by the default status".
	filterStatus = "all"
	filterPhase = -1

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--status":
			i++
			if i < len(args) {
				filterStatus = args[i]
			}
		case "--urgency":
			i++
			if i < len(args) {
				filterUrgency = args[i]
			}
		case "--importance":
			i++
			if i < len(args) {
				filterImportance = args[i]
			}
		case "--marker":
			i++
			if i < len(args) {
				filterMarker = args[i]
			}
		case "--phase":
			i++
			if i < len(args) {
				id, err := strconv.Atoi(args[i])
				if err != nil {
					fmt.Fprintf(os.Stderr, "--phase must be an integer, got %q\n", args[i])
					os.Exit(1)
				}
				filterPhase = id
			}
		case "--pretty":
			pretty = true
		}
	}

	p, _, err := planLoad()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	type result struct {
		Phases []Phase `json:"phases"`
	}
	out := result{}

	for _, ph := range p.Phases {
		if filterPhase != -1 && ph.ID != filterPhase {
			continue
		}
		if !phaseMatchesStatus(ph, filterStatus) {
			continue
		}
		if filterUrgency != "" && ph.Urgency != filterUrgency {
			continue
		}
		if filterImportance != "" && ph.Importance != filterImportance {
			continue
		}

		filtered := ph
		filtered.Tasks = nil
		for _, t := range ph.Tasks {
			if !taskMatchesStatus(t, filterStatus) {
				continue
			}
			if filterUrgency != "" && t.Urgency != filterUrgency {
				continue
			}
			if filterImportance != "" && t.Importance != filterImportance {
				continue
			}
			if filterMarker != "" && !taskHasMarker(t, filterMarker) {
				continue
			}
			filtered.Tasks = append(filtered.Tasks, t)
		}
		out.Phases = append(out.Phases, filtered)
	}

	if pretty {
		printPlanTable(out.Phases)
		return
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(out)
}

func phaseMatchesStatus(ph Phase, filter string) bool {
	switch filter {
	case "all":
		return true
	case "active", "":
		return ph.Status == "active" || ph.Status == ""
	case "archived":
		return ph.Status == "archived"
	default:
		// task-level status filter: keep all phases that may have matching tasks
		return true
	}
}

func taskMatchesStatus(t Task, filter string) bool {
	switch filter {
	case "all", "active", "":
		return true
	case "archived":
		return false
	default:
		return t.Status == filter
	}
}

func taskHasMarker(t Task, marker string) bool {
	for _, m := range t.QualityMarkers {
		if m == marker {
			return true
		}
	}
	return false
}

func printPlanTable(phases []Phase) {
	for _, ph := range phases {
		urgencyDotStr := urgencyDot(ph.Urgency)
		importanceStarStr := importanceStar(ph.Importance)
		fmt.Printf("## Phase %d: %s  %s %s  [%s]\n", ph.ID, ph.Title, urgencyDotStr, importanceStarStr, ph.Status)
		if len(ph.Tasks) == 0 {
			fmt.Println("  (no tasks)")
			continue
		}
		fmt.Printf("  %-8s %-10s %-6s %-6s  %s\n", "Task", "Status", "Urgency", "Imprt", "Name")
		fmt.Printf("  %-8s %-10s %-6s %-6s  %s\n", "--------", "----------", "------", "------", "----")
		for _, t := range ph.Tasks {
			fmt.Printf("  %-8s %-10s %-6s %-6s  %s\n",
				t.ID, t.Status, t.Urgency, t.Importance, t.Name)
		}
		fmt.Println()
	}
}

func urgencyDot(u string) string {
	switch u {
	case "high":
		return "🔴"
	case "low":
		return "🟢"
	default:
		return "🟡"
	}
}

func importanceStar(i string) string {
	switch i {
	case "high":
		return "★★"
	case "low":
		return "☆"
	default:
		return "★"
	}
}

// ---------------------------------------------------------------------------
// get
// ---------------------------------------------------------------------------

func runPlanGet(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: harness plan-cli get <phase-id|task-id>")
		os.Exit(1)
	}
	id := args[0]
	p, _, err := planLoad()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")

	// task-id: contains "."
	if strings.Contains(id, ".") {
		for _, ph := range p.Phases {
			for _, t := range ph.Tasks {
				if t.ID == id {
					enc.Encode(t)
					return
				}
			}
		}
		fmt.Fprintf(os.Stderr, "task %q not found\n", id)
		os.Exit(1)
	}

	// phase-id: integer
	phID, err := strconv.Atoi(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid id %q (expected integer or N.M)\n", id)
		os.Exit(1)
	}
	for _, ph := range p.Phases {
		if ph.ID == phID {
			enc.Encode(ph)
			return
		}
	}
	fmt.Fprintf(os.Stderr, "phase %d not found\n", phID)
	os.Exit(1)
}

// ---------------------------------------------------------------------------
// add-phase
// ---------------------------------------------------------------------------

func runPlanAddPhase(args []string) {
	var title, goal, urgency, importance string
	urgency = "medium"
	importance = "medium"

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--title":
			i++
			if i < len(args) {
				title = args[i]
			}
		case "--goal":
			i++
			if i < len(args) {
				goal = args[i]
			}
		case "--urgency":
			i++
			if i < len(args) {
				urgency = args[i]
			}
		case "--importance":
			i++
			if i < len(args) {
				importance = args[i]
			}
		}
	}
	if title == "" || goal == "" {
		fmt.Fprintln(os.Stderr, "--title and --goal are required")
		os.Exit(1)
	}

	path, err := planPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	p, err := LoadPlans(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	maxID := 0
	for _, ph := range p.Phases {
		if ph.ID > maxID {
			maxID = ph.ID
		}
	}
	newPhase := Phase{
		ID:         maxID + 1,
		Title:      title,
		Created:    time.Now().Format("2006-01-02"),
		Goal:       goal,
		Status:     "active",
		Urgency:    urgency,
		Importance: importance,
		Comments:   []Comment{},
		Tasks:      []Task{},
	}
	// Insert at front (newest phase on top convention)
	p.Phases = append([]Phase{newPhase}, p.Phases...)

	if err := SavePlans(path, p); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("added phase %d: %s\n", newPhase.ID, newPhase.Title)
}

// ---------------------------------------------------------------------------
// add-task
// ---------------------------------------------------------------------------

func runPlanAddTask(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: harness plan-cli add-task <phase-id> [flags]")
		os.Exit(1)
	}
	phID, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "phase-id must be integer, got %q\n", args[0])
		os.Exit(1)
	}
	args = args[1:]

	var name, description, dod, dependsStr, urgency, importance, markerStr string
	urgency = "medium"
	importance = "medium"

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--name":
			i++
			if i < len(args) {
				name = args[i]
			}
		case "--description":
			i++
			if i < len(args) {
				description = args[i]
			}
		case "--dod":
			i++
			if i < len(args) {
				dod = args[i]
			}
		case "--depends":
			i++
			if i < len(args) {
				dependsStr = args[i]
			}
		case "--urgency":
			i++
			if i < len(args) {
				urgency = args[i]
			}
		case "--importance":
			i++
			if i < len(args) {
				importance = args[i]
			}
		case "--marker":
			i++
			if i < len(args) {
				markerStr = args[i]
			}
		}
	}
	if name == "" || dod == "" {
		fmt.Fprintln(os.Stderr, "--name and --dod are required")
		os.Exit(1)
	}

	p, path, err := planLoad()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	phIdx := -1
	for i, ph := range p.Phases {
		if ph.ID == phID {
			phIdx = i
			break
		}
	}
	if phIdx == -1 {
		fmt.Fprintf(os.Stderr, "phase %d not found\n", phID)
		os.Exit(1)
	}

	seq := len(p.Phases[phIdx].Tasks) + 1
	taskID := fmt.Sprintf("%d.%d", phID, seq)

	var depends []string
	if dependsStr != "" {
		for _, d := range strings.Split(dependsStr, ",") {
			depends = append(depends, strings.TrimSpace(d))
		}
	}
	markers := []string{}
	if markerStr != "" {
		for _, m := range strings.Split(markerStr, ",") {
			markers = append(markers, strings.TrimSpace(m))
		}
	}

	t := Task{
		ID:             taskID,
		Name:           name,
		Description:    description,
		DoD:            dod,
		Depends:        depends,
		Status:         "cc:TODO",
		Urgency:        urgency,
		Importance:     importance,
		QualityMarkers: markers,
		Comments:       []Comment{},
	}
	p.Phases[phIdx].Tasks = append(p.Phases[phIdx].Tasks, t)

	if err := SavePlans(path, p); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("added task %s: %s\n", taskID, name)
}

// ---------------------------------------------------------------------------
// update
// ---------------------------------------------------------------------------

func runPlanUpdate(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: harness plan-cli update <task-id> [flags]")
		os.Exit(1)
	}
	taskID := args[0]
	args = args[1:]

	var status, hash, reason, urgency, importance string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--status":
			i++
			if i < len(args) {
				status = args[i]
			}
		case "--hash":
			i++
			if i < len(args) {
				hash = args[i]
			}
		case "--reason":
			i++
			if i < len(args) {
				reason = args[i]
			}
		case "--urgency":
			i++
			if i < len(args) {
				urgency = args[i]
			}
		case "--importance":
			i++
			if i < len(args) {
				importance = args[i]
			}
		}
	}

	p, path, err := planLoad()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	found := false
	for pi := range p.Phases {
		for ti := range p.Phases[pi].Tasks {
			if p.Phases[pi].Tasks[ti].ID != taskID {
				continue
			}
			t := &p.Phases[pi].Tasks[ti]
			if status != "" {
				t.Status = status
				if status == "cc:done" {
					t.StatusHash = hash
				}
				if status == "blocked" {
					t.BlockedReason = reason
				}
			}
			if urgency != "" {
				t.Urgency = urgency
			}
			if importance != "" {
				t.Importance = importance
			}
			found = true
		}
	}
	if !found {
		fmt.Fprintf(os.Stderr, "task %q not found\n", taskID)
		os.Exit(1)
	}

	if err := SavePlans(path, p); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("updated task %s\n", taskID)
}

// ---------------------------------------------------------------------------
// archive
// ---------------------------------------------------------------------------

func runPlanArchive(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: harness plan-cli archive <phase-id>")
		os.Exit(1)
	}
	phID, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "phase-id must be integer, got %q\n", args[0])
		os.Exit(1)
	}

	p, path, err := planLoad()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	found := false
	for i := range p.Phases {
		if p.Phases[i].ID == phID {
			p.Phases[i].Status = "archived"
			found = true
		}
	}
	if !found {
		fmt.Fprintf(os.Stderr, "phase %d not found\n", phID)
		os.Exit(1)
	}

	if err := SavePlans(path, p); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("archived phase %d\n", phID)
}

// ---------------------------------------------------------------------------
// comment
// ---------------------------------------------------------------------------

func runPlanComment(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: harness plan-cli comment <phase-id|task-id> [flags]")
		os.Exit(1)
	}
	targetID := args[0]
	args = args[1:]

	var text, author, authorName string
	author = "human"

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--text":
			i++
			if i < len(args) {
				text = args[i]
			}
		case "--author":
			i++
			if i < len(args) {
				author = args[i]
			}
		case "--author-name":
			i++
			if i < len(args) {
				authorName = args[i]
			}
		}
	}
	if text == "" {
		fmt.Fprintln(os.Stderr, "--text is required")
		os.Exit(1)
	}

	p, path, err := planLoad()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	c := Comment{
		ID:         uuid.New().String(),
		Author:     author,
		AuthorName: authorName,
		At:         time.Now().UTC().Format(time.RFC3339),
		Text:       text,
	}

	found := false
	if strings.Contains(targetID, ".") {
		// task
		for pi := range p.Phases {
			for ti := range p.Phases[pi].Tasks {
				if p.Phases[pi].Tasks[ti].ID == targetID {
					p.Phases[pi].Tasks[ti].Comments = append(p.Phases[pi].Tasks[ti].Comments, c)
					found = true
				}
			}
		}
	} else {
		// phase
		phID, err2 := strconv.Atoi(targetID)
		if err2 != nil {
			fmt.Fprintf(os.Stderr, "invalid id %q\n", targetID)
			os.Exit(1)
		}
		for i := range p.Phases {
			if p.Phases[i].ID == phID {
				p.Phases[i].Comments = append(p.Phases[i].Comments, c)
				found = true
			}
		}
	}

	if !found {
		fmt.Fprintf(os.Stderr, "target %q not found\n", targetID)
		os.Exit(1)
	}

	if err := SavePlans(path, p); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("added comment %s to %s\n", c.ID, targetID)
}

// ---------------------------------------------------------------------------
// update-phase
// ---------------------------------------------------------------------------

func runPlanUpdatePhase(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: harness plan-cli update-phase <phase-id> [flags]")
		os.Exit(1)
	}
	phID, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "phase-id must be integer, got %q\n", args[0])
		os.Exit(1)
	}
	args = args[1:]

	// Pointers distinguish "flag absent" from "flag set to empty string".
	var title, goal, urgency, importance, status *string
	capture := func(i *int) *string {
		*i++
		if *i < len(args) {
			v := args[*i]
			return &v
		}
		return nil
	}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--title":
			title = capture(&i)
		case "--goal":
			goal = capture(&i)
		case "--urgency":
			urgency = capture(&i)
		case "--importance":
			importance = capture(&i)
		case "--status":
			status = capture(&i)
		}
	}

	p, path, err := planLoad()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	found := false
	for i := range p.Phases {
		if p.Phases[i].ID != phID {
			continue
		}
		if title != nil {
			p.Phases[i].Title = *title
		}
		if goal != nil {
			p.Phases[i].Goal = *goal
		}
		if urgency != nil {
			p.Phases[i].Urgency = *urgency
		}
		if importance != nil {
			p.Phases[i].Importance = *importance
		}
		if status != nil {
			p.Phases[i].Status = *status
		}
		found = true
	}
	if !found {
		fmt.Fprintf(os.Stderr, "phase %d not found\n", phID)
		os.Exit(1)
	}

	if err := SavePlans(path, p); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("updated phase %d\n", phID)
}

// ---------------------------------------------------------------------------
// delete-task
// ---------------------------------------------------------------------------

func runPlanDeleteTask(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: harness plan-cli delete-task <task-id>")
		os.Exit(1)
	}
	taskID := args[0]

	p, path, err := planLoad()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	found := false
	for pi := range p.Phases {
		tasks := p.Phases[pi].Tasks
		for ti := range tasks {
			if tasks[ti].ID == taskID {
				p.Phases[pi].Tasks = append(tasks[:ti], tasks[ti+1:]...)
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	if !found {
		fmt.Fprintf(os.Stderr, "task %q not found\n", taskID)
		os.Exit(1)
	}

	if err := SavePlans(path, p); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("deleted task %s\n", taskID)
}

// ---------------------------------------------------------------------------
// delete-phase
// ---------------------------------------------------------------------------

func runPlanDeletePhase(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: harness plan-cli delete-phase <phase-id>")
		os.Exit(1)
	}
	phID, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "phase-id must be integer, got %q\n", args[0])
		os.Exit(1)
	}

	p, path, err := planLoad()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	idx := -1
	for i := range p.Phases {
		if p.Phases[i].ID == phID {
			idx = i
			break
		}
	}
	if idx == -1 {
		fmt.Fprintf(os.Stderr, "phase %d not found\n", phID)
		os.Exit(1)
	}
	p.Phases = append(p.Phases[:idx], p.Phases[idx+1:]...)

	if err := SavePlans(path, p); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("deleted phase %d\n", phID)
}

// runPlanServe is implemented in plan_serve.go.
