package plancli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ---------------------------------------------------------------------------
// Migrate: Plans.md → plans.json
// ---------------------------------------------------------------------------

func runPlanMigrate(args []string) {
	var fromFile string
	var dryRun bool

	fromFile = "Plans.md"
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--from":
			i++
			if i < len(args) {
				fromFile = args[i]
			}
		case "--dry-run":
			dryRun = true
		}
	}

	data, err := os.ReadFile(fromFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "migrate: cannot read %s: %v\n", fromFile, err)
		os.Exit(1)
	}

	plans, err := parsePlansMD(string(data))
	if err != nil {
		fmt.Fprintf(os.Stderr, "migrate: parse error: %v\n", err)
		os.Exit(1)
	}

	out, err := json.MarshalIndent(plans, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "migrate: marshal error: %v\n", err)
		os.Exit(1)
	}

	if dryRun {
		fmt.Println(string(out))
		return
	}

	destPath, err := planPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "migrate: %v\n", err)
		os.Exit(1)
	}
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "migrate: mkdir: %v\n", err)
		os.Exit(1)
	}
	if err := SavePlans(destPath, plans); err != nil {
		fmt.Fprintf(os.Stderr, "migrate: save: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("migrated %s → %s (%d phases)\n", fromFile, destPath, len(plans.Phases))
}

// ---------------------------------------------------------------------------
// Parser
// ---------------------------------------------------------------------------

var (
	rePhaseHeader   = regexp.MustCompile(`(?m)^## Phase (\d+):\s*(.+)$`)
	reCreated       = regexp.MustCompile(`(?m)^Created:\s*(.+)$`)
	reGoal          = regexp.MustCompile(`(?m)\*\*Goal\*\*:\s*(.+)`)
	reLastRelease   = regexp.MustCompile(`(?m)^Last release:\s*(.+)`)
	reProjectHeader = regexp.MustCompile(`(?m)^# (.+)\s*—\s*Plans\.md`)
	reStatusHash    = regexp.MustCompile(`cc:done\s+\[([^\]]+)\]`)
	reBlocked       = regexp.MustCompile(`blocked\s*\(([^)]*)\)`)
	reFuture        = regexp.MustCompile(`(?ms)^## Future Considerations\s*\n(.+?)(?:^## |\z)`)
	reRalphVerify   = regexp.MustCompile(`(?i)Verify:\s*(.+)`)
	reRalphMaxIter  = regexp.MustCompile(`(?i)MaxIter:\s*(\d+)`)
	reMarker        = regexp.MustCompile(`\[(feature:[^\]]+|ralph|skip:[^\]]+|needs-spike)\]`)
)

func parsePlansMD(content string) (*Plans, error) {
	plans := &Plans{
		FutureConsiderations: []string{},
	}

	// Extract project name from header
	if m := reProjectHeader.FindStringSubmatch(content); len(m) > 1 {
		plans.Project = strings.TrimSpace(m[1])
	}

	// Extract meta from "Last release:" line
	if m := reLastRelease.FindStringSubmatch(content); len(m) > 1 {
		plans.Meta = parseLastReleaseLine(strings.TrimSpace(m[1]))
	}

	// Extract future considerations
	if m := reFuture.FindStringSubmatch(content); len(m) > 1 {
		for _, line := range strings.Split(m[1], "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			// Strip leading list marker
			line = strings.TrimPrefix(line, "- ")
			line = strings.TrimPrefix(line, "* ")
			if line != "" {
				plans.FutureConsiderations = append(plans.FutureConsiderations, line)
			}
		}
	}

	// Split content into phase sections
	phaseMatches := rePhaseHeader.FindAllStringIndex(content, -1)
	for idx, loc := range phaseMatches {
		start := loc[0]
		var end int
		if idx+1 < len(phaseMatches) {
			end = phaseMatches[idx+1][0]
		} else {
			end = len(content)
		}
		phaseBlock := content[start:end]
		ph, err := parsePhaseBlock(phaseBlock)
		if err != nil {
			return nil, fmt.Errorf("phase block: %w", err)
		}
		plans.Phases = append(plans.Phases, ph)
	}

	return plans, nil
}

func parseLastReleaseLine(line string) PlansMeta {
	// Format: "v5.8.0 on 2026-05-26 (description)"
	meta := PlansMeta{}
	parts := strings.SplitN(line, " on ", 2)
	if len(parts) == 2 {
		meta.LastRelease = strings.TrimSpace(parts[0])
		rest := strings.TrimSpace(parts[1])
		// Rest: "2026-05-26 (description)" or "2026-05-26"
		datePart := rest
		desc := ""
		if lparen := strings.Index(rest, "("); lparen != -1 {
			datePart = strings.TrimSpace(rest[:lparen])
			desc = strings.Trim(rest[lparen:], "()")
			desc = strings.TrimSpace(desc)
		}
		meta.LastReleaseDate = datePart
		meta.LastReleaseDescription = desc
	}
	return meta
}

func parsePhaseBlock(block string) (Phase, error) {
	ph := Phase{
		Status:     "active",
		Urgency:    "medium",
		Importance: "medium",
		Comments:   []Comment{},
		Tasks:      []Task{},
	}

	// Extract phase ID and title from first line
	m := rePhaseHeader.FindStringSubmatch(block)
	if len(m) < 3 {
		return ph, fmt.Errorf("cannot parse phase header")
	}
	id, err := strconv.Atoi(m[1])
	if err != nil {
		return ph, fmt.Errorf("phase id: %w", err)
	}
	ph.ID = id
	ph.Title = strings.TrimSpace(m[2])

	// Detect archived phases (check for archive note before this section)
	// We treat all phases in Plans.md as "active" unless explicitly noted
	// (archived phases are typically moved out of Plans.md entirely)

	// Created date
	if cm := reCreated.FindStringSubmatch(block); len(cm) > 1 {
		ph.Created = strings.TrimSpace(cm[1])
	} else {
		ph.Created = time.Now().Format("2006-01-02")
	}

	// Goal
	if gm := reGoal.FindStringSubmatch(block); len(gm) > 1 {
		ph.Goal = strings.TrimSpace(gm[1])
	}

	// Parse task table
	lines := strings.Split(block, "\n")
	inTable := false
	for _, line := range lines {
		line = strings.TrimRight(line, " \t")
		if strings.HasPrefix(line, "| Task ") || strings.HasPrefix(line, "|---") {
			inTable = true
			continue
		}
		if inTable && strings.HasPrefix(line, "|") {
			task, ok := parseTaskRow(line)
			if ok {
				ph.Tasks = append(ph.Tasks, task)
			}
		} else if inTable && !strings.HasPrefix(line, "|") {
			// End of table
			inTable = false
		}
	}

	return ph, nil
}

func parseTaskRow(line string) (Task, bool) {
	// | 108.1 | description | DoD | depends | status |
	cols := splitTableRow(line)
	if len(cols) < 5 {
		return Task{}, false
	}

	idStr := strings.TrimSpace(cols[0])
	// Validate task ID format N.M
	if !regexp.MustCompile(`^\d+\.\d+$`).MatchString(idStr) {
		return Task{}, false
	}

	descRaw := strings.TrimSpace(cols[1])
	dodRaw := strings.TrimSpace(cols[2])
	dependsRaw := strings.TrimSpace(cols[3])
	statusRaw := strings.TrimSpace(cols[4])

	// Strip inline status markers from description (e.g., "[cc:WIP]" at start)
	descRaw = strings.TrimPrefix(descRaw, "[cc:WIP] ")
	descRaw = strings.TrimPrefix(descRaw, "[cc:TODO] ")
	descRaw = strings.TrimPrefix(descRaw, "[pm:requested] ")

	// Parse quality markers from description. Strip backtick spans first
	// so that references like `[ralph]` in text are not treated as markers.
	descForMarkers := regexp.MustCompile("`[^`]*`").ReplaceAllString(descRaw, "")
	markers := []string{}
	for _, mm := range reMarker.FindAllStringSubmatch(descForMarkers, -1) {
		markers = append(markers, mm[1])
	}
	// Remove marker annotations from description for cleanliness
	cleanDesc := reMarker.ReplaceAllString(descRaw, "")
	cleanDesc = strings.Join(strings.Fields(cleanDesc), " ")

	// Parse status
	status, hash, blockedReason := parseStatus(statusRaw)

	// Parse depends
	var depends []string
	if dependsRaw != "" && dependsRaw != "-" {
		for _, d := range strings.Split(dependsRaw, ",") {
			d = strings.TrimSpace(d)
			if d != "" && d != "-" {
				depends = append(depends, d)
			}
		}
	}
	if depends == nil {
		depends = []string{}
	}

	// Extract Ralph config from description
	var ralph *RalphConfig
	if strings.Contains(descRaw, "[ralph]") || strings.Contains(strings.ToLower(descRaw), "verify:") {
		rc := &RalphConfig{MaxIter: 10}
		if vm := reRalphVerify.FindStringSubmatch(descRaw); len(vm) > 1 {
			rc.Verify = strings.TrimSpace(vm[1])
		}
		if mm := reRalphMaxIter.FindStringSubmatch(descRaw); len(mm) > 1 {
			n, _ := strconv.Atoi(mm[1])
			if n > 0 {
				rc.MaxIter = n
			}
		}
		if rc.Verify != "" {
			ralph = rc
		}
	}

	t := Task{
		ID:             idStr,
		Name:           extractTaskName(cleanDesc),
		Description:    cleanDesc,
		DoD:            dodRaw,
		Depends:        depends,
		Status:         status,
		StatusHash:     hash,
		BlockedReason:  blockedReason,
		Urgency:        "medium",
		Importance:     "medium",
		QualityMarkers: markers,
		Ralph:          ralph,
		Comments:       []Comment{},
	}
	return t, true
}

// extractTaskName pulls the first sentence or first 80 chars as a short name.
func extractTaskName(desc string) string {
	if len(desc) <= 80 {
		return desc
	}
	// Try first sentence
	if idx := strings.Index(desc, ". "); idx > 0 && idx < 100 {
		return desc[:idx]
	}
	// Try first clause
	if idx := strings.Index(desc, ": "); idx > 0 && idx < 60 {
		return desc[:idx]
	}
	return desc[:80] + "…"
}

// parseStatus returns (status, hash, blockedReason) from a Plans.md status cell.
func parseStatus(raw string) (string, string, string) {
	raw = strings.TrimSpace(raw)
	if m := reStatusHash.FindStringSubmatch(raw); len(m) > 1 {
		return "cc:done", m[1], ""
	}
	if strings.HasPrefix(raw, "blocked") {
		if m := reBlocked.FindStringSubmatch(raw); len(m) > 1 {
			return "blocked", "", m[1]
		}
		return "blocked", "", ""
	}
	// Clean up any trailing text after valid statuses
	for _, s := range []string{"pm:confirmed", "pm:requested", "cc:done", "cc:WIP", "cc:TODO"} {
		if strings.HasPrefix(raw, s) {
			return s, "", ""
		}
	}
	if raw == "" || raw == "-" {
		return "cc:TODO", "", ""
	}
	return raw, "", ""
}

// splitTableRow splits a Markdown table row into columns, respecting
// backtick spans so that pipes inside `...` are not treated as delimiters.
// Input: "| col1 | `a|b` | col3 |"
// Returns: ["col1", "`a|b`", "col3"]
func splitTableRow(line string) []string {
	// Strip leading/trailing whitespace and outer pipes
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "|") {
		line = line[1:]
	}
	if strings.HasSuffix(line, "|") {
		line = line[:len(line)-1]
	}

	var cols []string
	var cur strings.Builder
	inBacktick := false
	backtickChar := byte('`')

	for i := 0; i < len(line); i++ {
		c := line[i]
		if c == backtickChar {
			inBacktick = !inBacktick
			cur.WriteByte(c)
		} else if c == '|' && !inBacktick {
			cols = append(cols, strings.TrimSpace(cur.String()))
			cur.Reset()
		} else {
			cur.WriteByte(c)
		}
	}
	// Last column
	if cur.Len() > 0 {
		cols = append(cols, strings.TrimSpace(cur.String()))
	}
	return cols
}
