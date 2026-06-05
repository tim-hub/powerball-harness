package hookhandler

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/tim-hub/powerball-harness/go/internal/plans"
)

// SprintContractGenerator generates sprint-contracts from plans.json.
type SprintContractGenerator struct {
	ProjectRoot string
	// PlansFile is the path to plans.json. When empty, the canonical
	// .claude/harness/plans.json under ProjectRoot is resolved.
	PlansFile  string
	OutputFile string
	Now        func() string
}

type sprintTaskRow struct {
	TaskID  string
	Title   string
	DoD     string
	Depends string
	Status  string
}

type sprintContractDoc struct {
	SchemaVersion string                `json:"schema_version"`
	GeneratedAt   string                `json:"generated_at"`
	Source        sprintContractSource  `json:"source"`
	Task          sprintContractTask    `json:"task"`
	Contract      sprintContractBody    `json:"contract"`
	Review        sprintContractReview  `json:"review"`
}

type sprintContractSource struct {
	PlansFile string `json:"plans_file"`
	TaskID    string `json:"task_id"`
}

type sprintContractTask struct {
	ID                 string   `json:"id"`
	Title              string   `json:"title"`
	DefinitionOfDone   string   `json:"definition_of_done"`
	DependsOn          []string `json:"depends_on"`
	StatusAtGeneration string   `json:"status_at_generation"`
}

type sprintContractBody struct {
	Checks            []sprintCheck      `json:"checks"`
	NonGoals          []string           `json:"non_goals"`
	RuntimeValidation []sprintValidation `json:"runtime_validation"`
	BrowserValidation []sprintValidation `json:"browser_validation"`
	RiskFlags         []string           `json:"risk_flags"`
	TDDRequired       bool               `json:"tdd_required"`
	TestFramework     string             `json:"test_framework,omitempty"`
	TestTodoList      []string           `json:"test_todo_list,omitempty"`
	SkipTDDReason     *string            `json:"skip_tdd_reason,omitempty"`
}

type sprintCheck struct {
	ID          string `json:"id"`
	Source      string `json:"source"`
	Description string `json:"description"`
}

type sprintValidation struct {
	ID                string   `json:"id"`
	Label             string   `json:"label,omitempty"`
	Command           string   `json:"command,omitempty"`
	Description       string   `json:"description,omitempty"`
	RequiredArtifacts []string `json:"required_artifacts,omitempty"`
}

type sprintContractReview struct {
	Status          string          `json:"status"`
	ReviewerProfile string          `json:"reviewer_profile"`
	MaxIterations   int             `json:"max_iterations"`
	RubricTarget    *uiRubricTarget `json:"rubric_target"`
	BrowserMode     *string         `json:"browser_mode"`
	Route           *string         `json:"route"`
	ReviewerNotes   []string        `json:"reviewer_notes"`
	ApprovedAt      *string         `json:"approved_at"`
	Gaps            []string        `json:"gaps"`
	Followups       []string        `json:"followups"`
}

type uiRubricTarget struct {
	Design        int `json:"design"`
	Originality   int `json:"originality"`
	Craft         int `json:"craft"`
	Functionality int `json:"functionality"`
}

var (
	// uiRubricRe matches tasks involving UI design quality evaluation.
	uiRubricRe = regexp.MustCompile(`(?i)\bui-rubric\b|\bdesign\b|styling|aesthetic|visual polish|design-heavy|design quality|originality|craft|functionality`)
	// uiWithDesignRe matches tasks mentioning "ui".
	uiWithDesignRe = regexp.MustCompile(`(?i)\bui\b`)
	// layoutWithDesignRe matches tasks mentioning "layout".
	layoutWithDesignRe = regexp.MustCompile(`(?i)\blayout\b`)
	// uiDesignHintRe matches tasks with design/visual hints.
	uiDesignHintRe = regexp.MustCompile(`(?i)design|styling|aesthetic|layout|visual|polish`)
	// browserProfileRe matches tasks that require browser-based review.
	browserProfileRe = regexp.MustCompile(`(?i)browser|chrome|playwright|\bui\b|layout|responsive|screenshot|screen|web app|webapp`)
	// runtimeProfileRe matches tasks that require runtime validation commands.
	runtimeProfileRe = regexp.MustCompile(`(?i)runtime|typecheck|lint|test|api|probe|integration|e2e|validation command`)
	// maxIterationsRe parses HTML comment overrides for max_iterations.
	maxIterationsRe = regexp.MustCompile(`(?is)<!--\s*max_iterations:\s*(\d+)\s*-->`)
	// exploratoryModeRe detects exploratory browser mode hints.
	exploratoryModeRe = regexp.MustCompile(`(?i)(browser_mode\s*:\s*exploratory|\bexploratory\b)`)
	// scriptedModeRe detects scripted browser mode hints.
	scriptedModeRe = regexp.MustCompile(`(?i)(browser_mode\s*:\s*scripted|\bscripted\b)`)
	// explicitRouteRe parses explicit browser route declarations.
	explicitRouteRe = regexp.MustCompile(`(?i)(?:browser_)?route\s*:\s*(playwright|agent-browser|chrome-devtools)`)
	// securitySensitiveRe detects security-related tasks.
	securitySensitiveRe = regexp.MustCompile(`(?i)security|auth|permission|secret|guardrail`)
	// stateMigrationRe detects state or schema migration tasks.
	stateMigrationRe = regexp.MustCompile(`(?i)migration|schema|state|resume|session|artifact`)
	// uxRegressionRe detects tasks with UI/browser regression risk.
	uxRegressionRe = regexp.MustCompile(`(?i)browser|ui|layout|responsive|playwright|chrome|screen`)
	// tddRequiredTagRe matches the [tdd:required] task tag.
	tddRequiredTagRe = regexp.MustCompile(`(?i)\[tdd:required\]`)
	// tddSkipTagRe matches the [tdd:skip:<reason>] task tag and captures the reason.
	tddSkipTagRe = regexp.MustCompile(`(?i)\[tdd:skip:([^\]]+)\]`)
)

var profileMaxIterations = map[string]int{
	"static":    3,
	"runtime":   3,
	"browser":   5,
	"ui-rubric": 10,
}

var defaultUIRubricTarget = &uiRubricTarget{
	Design:        6,
	Originality:   6,
	Craft:         6,
	Functionality: 6,
}

// sprintTDDContract holds the TDD configuration inferred for a task.
type sprintTDDContract struct {
	Required      bool
	SkipReason    *string
	TestFramework string
	TestTodoList  []string
}

// detectSprintTDD inspects task tags and project files to infer TDD requirements.
// Tag priority: [tdd:skip:reason] > [tdd:required] > path inference.
func detectSprintTDD(root string, task *sprintTaskRow) sprintTDDContract {
	taskText := fmt.Sprintf("%s %s", task.Title, task.DoD)

	// [tdd:skip] always wins — explicit opt-out.
	if m := tddSkipTagRe.FindStringSubmatch(taskText); len(m) >= 2 {
		reason := strings.TrimSpace(m[1])
		return sprintTDDContract{Required: false, SkipReason: &reason}
	}

	// [tdd:required] explicit opt-in.
	if tddRequiredTagRe.MatchString(taskText) {
		fw := detectSprintTestFramework(root)
		return sprintTDDContract{
			Required:      true,
			TestFramework: fw,
			TestTodoList:  defaultSprintTDDTodoList(fw),
		}
	}

	// No explicit tag → not required.
	return sprintTDDContract{Required: false}
}

// detectSprintTestFramework returns the test framework name by probing project files.
func detectSprintTestFramework(root string) string {
	probes := []struct {
		file      string
		framework string
	}{
		{"vitest.config.ts", "vitest"},
		{"vitest.config.js", "vitest"},
		{"vitest.config.mjs", "vitest"},
		{"jest.config.js", "jest"},
		{"jest.config.ts", "jest"},
		{"pytest.ini", "pytest"},
		{"pyproject.toml", "pytest"},
		{"Cargo.toml", "cargo-test"},
		{"go.mod", "go-test"},
	}
	for _, p := range probes {
		if _, err := os.Stat(filepath.Join(root, p.file)); err == nil {
			return p.framework
		}
	}
	return ""
}

// defaultSprintTDDTodoList returns a minimal Red-Green-Refactor todo list for the framework.
func defaultSprintTDDTodoList(framework string) []string {
	switch framework {
	case "vitest", "jest":
		return []string{"Write failing test (Red)", "Implement minimal code (Green)", "Refactor + re-run"}
	case "pytest":
		return []string{"Write failing pytest (Red)", "Implement minimal code (Green)", "Refactor + re-run"}
	case "go-test":
		return []string{"Write failing _test.go (Red)", "Implement minimal code (Green)", "Refactor + go test ./..."}
	case "cargo-test":
		return []string{"Write failing #[test] (Red)", "Implement minimal code (Green)", "Refactor + cargo test"}
	default:
		return []string{"Write failing test (Red)", "Implement minimal code (Green)", "Refactor"}
	}
}

// Generate produces a sprint-contract document for the specified task ID.
func (g *SprintContractGenerator) Generate(taskID string) (*sprintContractDoc, error) {
	projectRoot := g.ProjectRoot
	if projectRoot == "" {
		projectRoot = resolveProjectRoot()
	}

	plansFile := g.PlansFile
	var p *plans.Plans
	if plansFile != "" {
		loaded, err := plans.Load(plansFile)
		if err != nil {
			return nil, fmt.Errorf("read plans file: %w", err)
		}
		p = loaded
	} else {
		loaded, err := resolvePlansJSON(projectRoot)
		if err != nil {
			return nil, fmt.Errorf("read plans file: %w", err)
		}
		p = loaded
		plansFile = plans.ResolvePath(projectRoot, readPlansDirectoryFromConfig(projectRoot))
	}

	if p == nil {
		return nil, fmt.Errorf("plans.json not found: %s", plansFile)
	}

	row, err := sprintTaskRowFromPlans(p, taskID)
	if err != nil {
		return nil, err
	}

	reviewerProfile := detectSprintProfile(row)
	maxIterations := detectSprintMaxIterations(reviewerProfile, row)
	runtimeValidation := pickRuntimeCommands(projectRoot)
	riskFlags := detectSprintRiskFlags(row)
	tddContract := detectSprintTDD(projectRoot, row)

	var browserMode *string
	var route *string
	if reviewerProfile == "browser" {
		mode := detectSprintBrowserMode(row)
		browserMode = &mode
		if detected := detectSprintBrowserRoute(row, projectRoot, mode); detected != nil {
			route = detected
		}
	}

	var rubricTarget *uiRubricTarget
	if reviewerProfile == "ui-rubric" {
		copy := *defaultUIRubricTarget
		rubricTarget = &copy
	}

	browserValidation := []sprintValidation{}
	if reviewerProfile == "browser" && browserMode != nil {
		requiredArtifacts := []string{"trace", "screenshot", "ui-flow-log"}
		if *browserMode == "exploratory" {
			requiredArtifacts = []string{"snapshot", "ui-flow-log"}
		}
		browserValidation = []sprintValidation{
			{
				ID:                "browser-smoke",
				Description:       row.DoD,
				RequiredArtifacts: requiredArtifacts,
			},
		}
	}

	now := time.Now().UTC().Format(time.RFC3339)
	if g.Now != nil {
		now = g.Now()
	}

	relPlans := plansFile
	if rel, err := filepath.Rel(projectRoot, plansFile); err == nil && rel != "" {
		relPlans = rel
	}

	doc := &sprintContractDoc{
		SchemaVersion: "sprint-contract.v1",
		GeneratedAt:   now,
		Source: sprintContractSource{
			PlansFile: relPlans,
			TaskID:    row.TaskID,
		},
		Task: sprintContractTask{
			ID:                 row.TaskID,
			Title:              row.Title,
			DefinitionOfDone:   row.DoD,
			DependsOn:          sprintToList(row.Depends),
			StatusAtGeneration: row.Status,
		},
		Contract: sprintContractBody{
			Checks: []sprintCheck{
				{
					ID:          "dod-primary",
					Source:      "plans.json.dod",
					Description: row.DoD,
				},
			},
			NonGoals:          []string{},
			RuntimeValidation: runtimeValidation,
			BrowserValidation: browserValidation,
			RiskFlags:         riskFlags,
			TDDRequired:       tddContract.Required,
			TestFramework:     tddContract.TestFramework,
			TestTodoList:      tddContract.TestTodoList,
			SkipTDDReason:     tddContract.SkipReason,
		},
		Review: sprintContractReview{
			Status:          "draft",
			ReviewerProfile: reviewerProfile,
			MaxIterations:   maxIterations,
			RubricTarget:    rubricTarget,
			BrowserMode:     browserMode,
			Route:           route,
			ReviewerNotes:   []string{},
			ApprovedAt:      nil,
			Gaps:            []string{},
			Followups:       []string{},
		},
	}

	return doc, nil
}

// Write generates a contract and writes it as JSON to the output file.
func (g *SprintContractGenerator) Write(taskID string) (string, error) {
	doc, err := g.Generate(taskID)
	if err != nil {
		return "", err
	}

	projectRoot := g.ProjectRoot
	if projectRoot == "" {
		projectRoot = resolveProjectRoot()
	}

	outputFile := g.OutputFile
	if outputFile == "" {
		outputFile = filepath.Join(projectRoot, ".claude", "state", "contracts", fmt.Sprintf("%s.sprint-contract.json", taskID))
	}

	if err := os.MkdirAll(filepath.Dir(outputFile), 0o755); err != nil {
		return "", fmt.Errorf("create output dir: %w", err)
	}

	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal contract: %w", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(outputFile, data, 0o644); err != nil {
		return "", fmt.Errorf("write contract: %w", err)
	}
	return outputFile, nil
}

// sprintTaskRowFromPlans builds a sprintTaskRow from a plans.json task. Quality
// markers are appended to the title text as `[marker]` tokens so the existing
// text-based profile, risk-flag, and TDD detectors continue to fire.
func sprintTaskRowFromPlans(p *plans.Plans, targetTaskID string) (*sprintTaskRow, error) {
	task, _ := p.FindTask(targetTaskID)
	if task == nil {
		return nil, fmt.Errorf("task row not found in plans.json: %s", targetTaskID)
	}

	title := task.Name
	for _, marker := range task.QualityMarkers {
		marker = strings.TrimSpace(marker)
		if marker == "" {
			continue
		}
		title = strings.TrimSpace(title + " [" + marker + "]")
	}

	return &sprintTaskRow{
		TaskID:  task.ID,
		Title:   title,
		DoD:     task.DoD,
		Depends: strings.Join(task.Depends, ", "),
		Status:  task.Status,
	}, nil
}

func sprintToList(value string) []string {
	if value == "" || value == "-" {
		return []string{}
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func detectSprintProfile(task *sprintTaskRow) string {
	text := strings.ToLower(fmt.Sprintf("%s %s", task.Title, task.DoD))
	hasUIHints := uiRubricRe.MatchString(text) ||
		(uiWithDesignRe.MatchString(text) && uiDesignHintRe.MatchString(text)) ||
		(layoutWithDesignRe.MatchString(text) && uiDesignHintRe.MatchString(text))
	if hasUIHints {
		return "ui-rubric"
	}
	if browserProfileRe.MatchString(text) {
		return "browser"
	}
	if runtimeProfileRe.MatchString(text) {
		return "runtime"
	}
	return "static"
}

func detectSprintMaxIterations(profile string, task *sprintTaskRow) int {
	defaultValue := 3
	if value, ok := profileMaxIterations[profile]; ok {
		defaultValue = value
	}
	text := fmt.Sprintf("%s\n%s", task.Title, task.DoD)
	match := maxIterationsRe.FindStringSubmatch(text)
	if len(match) >= 2 {
		value, err := strconv.Atoi(match[1])
		if err == nil && value >= 1 && value <= 30 {
			return value
		}
		if err == nil {
			fmt.Fprintf(os.Stderr, "[warn] max_iterations=%d out of range (1-30), falling back to default %d\n", value, defaultValue)
		}
	}
	return defaultValue
}

func detectSprintBrowserMode(task *sprintTaskRow) string {
	text := strings.ToLower(fmt.Sprintf("%s %s", task.Title, task.DoD))
	if exploratoryModeRe.MatchString(text) {
		return "exploratory"
	}
	if scriptedModeRe.MatchString(text) {
		return "scripted"
	}
	return "scripted"
}

func detectSprintRiskFlags(task *sprintTaskRow) []string {
	text := strings.ToLower(fmt.Sprintf("%s %s", task.Title, task.DoD))
	flags := []string{}
	if strings.Contains(task.Title, "[needs-spike]") || strings.Contains(task.DoD, "[needs-spike]") {
		flags = append(flags, "needs-spike")
	}
	if securitySensitiveRe.MatchString(text) {
		flags = append(flags, "security-sensitive")
	}
	if stateMigrationRe.MatchString(text) {
		flags = append(flags, "state-migration")
	}
	if uxRegressionRe.MatchString(text) {
		flags = append(flags, "ux-regression")
	}
	seen := map[string]struct{}{}
	unique := make([]string, 0, len(flags))
	for _, flag := range flags {
		if _, ok := seen[flag]; ok {
			continue
		}
		seen[flag] = struct{}{}
		unique = append(unique, flag)
	}
	return unique
}

func detectSprintBrowserRoute(task *sprintTaskRow, root, browserMode string) *string {
	text := fmt.Sprintf("%s\n%s", task.Title, task.DoD)
	match := explicitRouteRe.FindStringSubmatch(text)
	if len(match) >= 2 {
		value := strings.ToLower(match[1])
		return &value
	}
	if browserMode == "exploratory" {
		return nil
	}
	if hasPlaywrightBasis(root) {
		value := "playwright"
		return &value
	}
	return nil
}

func hasPlaywrightBasis(root string) bool {
	if os.Getenv("HARNESS_BROWSER_REVIEW_DISABLE_PLAYWRIGHT") != "" {
		return false
	}
	packageJSONPath := filepath.Join(root, "package.json")
	data, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return false
	}

	var pkg struct {
		Scripts         map[string]string `json:"scripts"`
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return false
	}
	if _, ok := pkg.Scripts["test:e2e"]; ok {
		return true
	}
	if _, ok := pkg.Dependencies["playwright"]; ok {
		return true
	}
	if _, ok := pkg.DevDependencies["playwright"]; ok {
		return true
	}
	if _, ok := pkg.Dependencies["@playwright/test"]; ok {
		return true
	}
	if _, ok := pkg.DevDependencies["@playwright/test"]; ok {
		return true
	}
	return false
}

func pickRuntimeCommands(root string) []sprintValidation {
	commands := []sprintValidation{}

	packageJSONPath := filepath.Join(root, "package.json")
	if data, err := os.ReadFile(packageJSONPath); err == nil {
		var pkg struct {
			Scripts map[string]string `json:"scripts"`
		}
		if err := json.Unmarshal(data, &pkg); err == nil {
			if _, ok := pkg.Scripts["test"]; ok {
				commands = append(commands, sprintValidation{Label: "package-test", Command: "CI=true npm test"})
			}
			if _, ok := pkg.Scripts["lint"]; ok {
				commands = append(commands, sprintValidation{Label: "package-lint", Command: "npm run lint"})
			}
			if _, ok := pkg.Scripts["typecheck"]; ok {
				commands = append(commands, sprintValidation{Label: "package-typecheck", Command: "npm run typecheck"})
			}
			if _, ok := pkg.Scripts["test:e2e"]; ok {
				commands = append(commands, sprintValidation{Label: "package-e2e", Command: "npm run test:e2e"})
			}
		} else {
			commands = append(commands, sprintValidation{
				Label:   "package-parse-error",
				Command: fmt.Sprintf("echo \"ERROR: package.json parse failed: %s\" >&2; exit 1", strings.ReplaceAll(err.Error(), `"`, `\"`)),
			})
		}
	}

	if len(commands) == 0 {
		fallbacks := []struct {
			Marker  string
			Label   string
			Command string
		}{
			{Marker: "pnpm-lock.yaml", Label: "pnpm-test", Command: "pnpm test"},
			{Marker: "bun.lock", Label: "bun-test", Command: "bun test"},
			{Marker: "go.mod", Label: "go-test", Command: "go test ./..."},
			{Marker: "Cargo.toml", Label: "cargo-test", Command: "cargo test"},
		}
		for _, fallback := range fallbacks {
			if _, err := os.Stat(filepath.Join(root, fallback.Marker)); err == nil {
				commands = append(commands, sprintValidation{Label: fallback.Label, Command: fallback.Command})
				break
			}
		}
	}

	if len(commands) == 0 {
		shellFallbacks := []struct {
			Path    string
			Label   string
			Command string
		}{
			{Path: "tests/validate-plugin.sh", Label: "validate-plugin", Command: "./tests/validate-plugin.sh"},
			{Path: ".claude/skills/release-this/scripts/check-consistency.sh", Label: "check-consistency", Command: "./.claude/skills/release-this/scripts/check-consistency.sh"},
		}
		for _, fallback := range shellFallbacks {
			if _, err := os.Stat(filepath.Join(root, fallback.Path)); err == nil {
				commands = append(commands, sprintValidation{Label: fallback.Label, Command: fallback.Command})
			}
		}
	}

	return commands
}
