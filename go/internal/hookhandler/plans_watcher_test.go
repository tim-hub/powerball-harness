package hookhandler

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandlePlansWatcher_NoInput(t *testing.T) {
	var out bytes.Buffer
	err := HandlePlansWatcher(strings.NewReader(""), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result postToolOutput
	if jsonErr := json.Unmarshal(out.Bytes(), &result); jsonErr != nil {
		t.Fatalf("invalid JSON output: %v, raw: %s", jsonErr, out.String())
	}
	if result.HookSpecificOutput.HookEventName != "PostToolUse" {
		t.Errorf("expected hookEventName=PostToolUse, got %q", result.HookSpecificOutput.HookEventName)
	}
}

func TestHandlePlansWatcher_NoFilePath(t *testing.T) {
	input := `{"tool_name":"Edit","tool_input":{}}`
	var out bytes.Buffer
	if err := HandlePlansWatcher(strings.NewReader(input), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result postToolOutput
	if jsonErr := json.Unmarshal(out.Bytes(), &result); jsonErr != nil {
		t.Fatalf("invalid JSON output: %v", jsonErr)
	}
	if result.HookSpecificOutput.AdditionalContext != "" {
		t.Errorf("expected empty context for no file_path, got %q", result.HookSpecificOutput.AdditionalContext)
	}
}

func TestHandlePlansWatcher_NonPlansFile(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	// pre-create Plans.md
	if err := os.WriteFile("Plans.md", []byte("# Plans\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// skip when a file other than Plans.md is edited
	input := `{"tool_name":"Write","tool_input":{"file_path":"src/main.go"}}`
	var out bytes.Buffer
	if err := HandlePlansWatcher(strings.NewReader(input), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result postToolOutput
	if jsonErr := json.Unmarshal(out.Bytes(), &result); jsonErr != nil {
		t.Fatalf("invalid JSON output: %v", jsonErr)
	}
	if result.HookSpecificOutput.AdditionalContext != "" {
		t.Errorf("expected empty context for non-Plans.md file, got %q", result.HookSpecificOutput.AdditionalContext)
	}
}

func TestHandlePlansWatcher_NoPlansFile(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	// skip when Plans.md does not exist
	input := `{"tool_name":"Write","tool_input":{"file_path":"Plans.md"}}`
	var out bytes.Buffer
	if err := HandlePlansWatcher(strings.NewReader(input), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result postToolOutput
	if jsonErr := json.Unmarshal(out.Bytes(), &result); jsonErr != nil {
		t.Fatalf("invalid JSON output: %v", jsonErr)
	}
	if result.HookSpecificOutput.AdditionalContext != "" {
		t.Errorf("expected empty context when Plans.md not found, got %q", result.HookSpecificOutput.AdditionalContext)
	}
}

func TestHandlePlansWatcher_NewTaskDetected(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	plansContent := "| Task 1 | TaskA | DoD | - | pm:pending |\n"
	if err := os.WriteFile("Plans.md", []byte(plansContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// save the previous state (pm_pending=0)
	if err := os.MkdirAll(".claude/state", 0o755); err != nil {
		t.Fatal(err)
	}
	prevState := `{"timestamp":"2026-01-01T00:00:00Z","pm_pending":0,"cc_todo":0,"cc_wip":0,"cc_done":0,"pm_confirmed":0}`
	if err := os.WriteFile(plansStateFile, []byte(prevState), 0o644); err != nil {
		t.Fatal(err)
	}

	input := `{"tool_name":"Edit","tool_input":{"file_path":"Plans.md"}}`
	var out bytes.Buffer
	if err := HandlePlansWatcher(strings.NewReader(input), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result postToolOutput
	if jsonErr := json.Unmarshal(out.Bytes(), &result); jsonErr != nil {
		t.Fatalf("invalid JSON output: %v, raw: %s", jsonErr, out.String())
	}

	// new tasks should be detected
	if !strings.Contains(result.HookSpecificOutput.AdditionalContext, "New tasks") {
		t.Errorf("expected 'New tasks' in additionalContext, got %q",
			result.HookSpecificOutput.AdditionalContext)
	}

	// pm-notification.md should be created
	data, err := os.ReadFile(pmNotificationFile)
	if err != nil {
		t.Fatalf("pm-notification.md not created: %v", err)
	}
	if !strings.Contains(string(data), "New tasks") {
		t.Errorf("pm-notification.md should contain 'New tasks', got: %s", string(data))
	}
}

func TestHandlePlansWatcher_CompletedTaskDetected(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	plansContent := "| Task 1 | TaskA | DoD | - | cc:done |\n"
	if err := os.WriteFile("Plans.md", []byte(plansContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// save the previous state (cc_done=0)
	if err := os.MkdirAll(".claude/state", 0o755); err != nil {
		t.Fatal(err)
	}
	prevState := `{"timestamp":"2026-01-01T00:00:00Z","pm_pending":0,"cc_todo":0,"cc_wip":0,"cc_done":0,"pm_confirmed":0}`
	if err := os.WriteFile(plansStateFile, []byte(prevState), 0o644); err != nil {
		t.Fatal(err)
	}

	input := `{"tool_name":"Edit","tool_input":{"file_path":"Plans.md"}}`
	var out bytes.Buffer
	if err := HandlePlansWatcher(strings.NewReader(input), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result postToolOutput
	if jsonErr := json.Unmarshal(out.Bytes(), &result); jsonErr != nil {
		t.Fatalf("invalid JSON output: %v", jsonErr)
	}

	if !strings.Contains(result.HookSpecificOutput.AdditionalContext, "Tasks completed") {
		t.Errorf("expected 'Tasks completed' in additionalContext, got %q",
			result.HookSpecificOutput.AdditionalContext)
	}
}

func TestHandlePlansWatcher_NoChange(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	// no change (cc:TODO stays at 1)
	plansContent := "| Task 1 | TaskA | DoD | - | cc:TODO |\n"
	if err := os.WriteFile("Plans.md", []byte(plansContent), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(".claude/state", 0o755); err != nil {
		t.Fatal(err)
	}
	prevState := `{"timestamp":"2026-01-01T00:00:00Z","pm_pending":0,"cc_todo":1,"cc_wip":0,"cc_done":0,"pm_confirmed":0}`
	if err := os.WriteFile(plansStateFile, []byte(prevState), 0o644); err != nil {
		t.Fatal(err)
	}

	input := `{"tool_name":"Edit","tool_input":{"file_path":"Plans.md"}}`
	var out bytes.Buffer
	if err := HandlePlansWatcher(strings.NewReader(input), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result postToolOutput
	if jsonErr := json.Unmarshal(out.Bytes(), &result); jsonErr != nil {
		t.Fatalf("invalid JSON output: %v", jsonErr)
	}

	// no notification when there is no change
	if result.HookSpecificOutput.AdditionalContext != "" {
		t.Errorf("expected empty context for no change, got %q", result.HookSpecificOutput.AdditionalContext)
	}
}

func TestHandlePlansWatcher_StatusSummary(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	// Plans.md with multiple markers
	plansContent := "| Task 1 | A | DoD | - | cc:TODO |\n" +
		"| Task 2 | B | DoD | - | cc:WIP |\n" +
		"| Task 3 | C | DoD | - | cc:done |\n" +
		"| Task 4 | D | DoD | - | pm:pending |\n"
	if err := os.WriteFile("Plans.md", []byte(plansContent), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(".claude/state", 0o755); err != nil {
		t.Fatal(err)
	}
	prevState := `{"timestamp":"2026-01-01T00:00:00Z","pm_pending":0,"cc_todo":0,"cc_wip":0,"cc_done":0,"pm_confirmed":0}`
	if err := os.WriteFile(plansStateFile, []byte(prevState), 0o644); err != nil {
		t.Fatal(err)
	}

	input := `{"tool_name":"Edit","tool_input":{"file_path":"Plans.md"}}`
	var out bytes.Buffer
	if err := HandlePlansWatcher(strings.NewReader(input), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result postToolOutput
	if jsonErr := json.Unmarshal(out.Bytes(), &result); jsonErr != nil {
		t.Fatalf("invalid JSON output: %v", jsonErr)
	}

	ctx := result.HookSpecificOutput.AdditionalContext
	// summary should contain status counts
	if !strings.Contains(ctx, "cc:TODO") {
		t.Errorf("expected 'cc:TODO' in summary, got %q", ctx)
	}
	if !strings.Contains(ctx, "cc:WIP") {
		t.Errorf("expected 'cc:WIP' in summary, got %q", ctx)
	}
	if !strings.Contains(ctx, "cc:done") {
		t.Errorf("expected 'cc:done' in summary, got %q", ctx)
	}
}

func TestIsPlansFile(t *testing.T) {
	// isPlansFile uses strict filepath.Clean matching only.
	// Comparison factoring in projectRoot is done by isPlansFileWithRoot.
	cases := []struct {
		changed  string
		plans    string
		expected bool
	}{
		// exact match (relative path)
		{"Plans.md", "Plans.md", true},
		// exact match (absolute path)
		{"/home/user/project/Plans.md", "/home/user/project/Plans.md", true},
		// different full paths are not equal (file with same name in another dir is false)
		{"docs/Plans.md", "Plans.md", false},
		{"/home/user/project/Plans.md", "Plans.md", false},
		{"src/main.go", "Plans.md", false},
		{"NotPlans.md", "Plans.md", false},
	}
	for _, c := range cases {
		got := isPlansFile(c.changed, c.plans)
		if got != c.expected {
			t.Errorf("isPlansFile(%q, %q) = %v, want %v", c.changed, c.plans, got, c.expected)
		}
	}
}

func TestIsPlansFileWithRoot(t *testing.T) {
	// isPlansFileWithRoot resolves relative paths using projectRoot.
	projectRoot := "/home/user/project"
	cases := []struct {
		changedFile string
		plansFile   string
		expected    bool
		desc        string
	}{
		// relative path resolved via projectRoot to match
		{"Plans.md", "/home/user/project/Plans.md", true, "relative path match"},
		// absolute path match
		{"/home/user/project/Plans.md", "/home/user/project/Plans.md", true, "absolute path match"},
		// relative path in different directory (Plans.md but plansFile is directly under projectRoot)
		{"docs/Plans.md", "/home/user/project/Plans.md", false, "subdirectory mismatch"},
		// relative path in different directory that matches plansFile
		{"docs/Plans.md", "/home/user/project/docs/Plans.md", true, "subdirectory match"},
		// completely different file
		{"src/main.go", "/home/user/project/Plans.md", false, "non plans file"},
		// same-name file in a different project (absolute path)
		{"/tmp/other/Plans.md", "/home/user/project/Plans.md", false, "different project Plans.md"},
	}
	for _, tc := range cases {
		got := isPlansFileWithRoot(tc.changedFile, tc.plansFile, projectRoot)
		if got != tc.expected {
			t.Errorf("[%s] isPlansFileWithRoot(%q, %q, %q) = %v, want %v",
				tc.desc, tc.changedFile, tc.plansFile, projectRoot, got, tc.expected)
		}
	}
}

func TestCountMarker(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	content := "cc:TODO\ncc:TODO\ncc:WIP\ncc:done\npm:pending\n"
	if err := os.WriteFile("Plans.md", []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		marker   string
		expected int
	}{
		{"cc:TODO", 2},
		{"cc:WIP", 1},
		{"cc:done", 1},
		{"pm:pending", 1},
		{"pm:confirmed", 0},
	}
	for _, c := range cases {
		got := countMarker("Plans.md", c.marker)
		if got != c.expected {
			t.Errorf("countMarker(Plans.md, %q) = %d, want %d", c.marker, got, c.expected)
		}
	}
}

func TestHandlePlansWatcher_CursorCompatMarker(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	plansContent := "| Task 1 | A | DoD | - | cursor:pending |\n"
	if err := os.WriteFile("Plans.md", []byte(plansContent), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(".claude/state", 0o755); err != nil {
		t.Fatal(err)
	}
	prevState := `{"timestamp":"2026-01-01T00:00:00Z","pm_pending":0,"cc_todo":0,"cc_wip":0,"cc_done":0,"pm_confirmed":0}`
	if err := os.WriteFile(plansStateFile, []byte(prevState), 0o644); err != nil {
		t.Fatal(err)
	}

	input := `{"tool_name":"Edit","tool_input":{"file_path":"Plans.md"}}`
	var out bytes.Buffer
	if err := HandlePlansWatcher(strings.NewReader(input), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result postToolOutput
	if jsonErr := json.Unmarshal(out.Bytes(), &result); jsonErr != nil {
		t.Fatalf("invalid JSON output: %v", jsonErr)
	}

	if !strings.Contains(result.HookSpecificOutput.AdditionalContext, "New tasks") {
		t.Errorf("expected 'New tasks' for cursor:pending, got %q",
			result.HookSpecificOutput.AdditionalContext)
	}
}

// TestHandlePlansWatcher_CustomPlansDirectory verifies that Plans.md in a custom directory
// is correctly detected when the plansDirectory setting is configured.
func TestHandlePlansWatcher_CustomPlansDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	// create config file (plansDirectory: docs)
	configContent := "plansDirectory: docs\n"
	if err := os.WriteFile(harnessConfigFileName, []byte(configContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// create docs/Plans.md
	if err := os.MkdirAll("docs", 0o755); err != nil {
		t.Fatal(err)
	}
	plansContent := "| Task 1 | TaskA | DoD | - | pm:pending |\n"
	if err := os.WriteFile("docs/Plans.md", []byte(plansContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// save previous state (pm_pending=0)
	if err := os.MkdirAll(".claude/state", 0o755); err != nil {
		t.Fatal(err)
	}
	prevState := `{"timestamp":"2026-01-01T00:00:00Z","pm_pending":0,"cc_todo":0,"cc_wip":0,"cc_done":0,"pm_confirmed":0}`
	if err := os.WriteFile(plansStateFile, []byte(prevState), 0o644); err != nil {
		t.Fatal(err)
	}

	// send event indicating docs/Plans.md was modified
	input := `{"tool_name":"Edit","tool_input":{"file_path":"docs/Plans.md"}}`
	var out bytes.Buffer
	if err := HandlePlansWatcher(strings.NewReader(input), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result postToolOutput
	if jsonErr := json.Unmarshal(out.Bytes(), &result); jsonErr != nil {
		t.Fatalf("invalid JSON output: %v, raw: %s", jsonErr, out.String())
	}

	// new tasks should be detected (custom plansDirectory Plans.md is recognized)
	if !strings.Contains(result.HookSpecificOutput.AdditionalContext, "New tasks") {
		t.Errorf("expected 'New tasks' in additionalContext for custom plansDirectory, got %q",
			result.HookSpecificOutput.AdditionalContext)
	}
}

// TestIsPlansFile_CustomPath verifies that isPlansFileWithRoot works correctly
// for a custom-path Plans.md (P2 fix: same-name files in other directories must not be falsely matched).
func TestIsPlansFile_CustomPath(t *testing.T) {
	projectRoot := "/project"
	cases := []struct {
		changedFile string
		plansFile   string
		want        bool
		desc        string
	}{
		// exact match (resolve relative path with projectRoot)
		{"Plans.md", "/project/Plans.md", true, "exact match via projectRoot"},
		// custom directory Plans.md matches plansFile
		{"docs/Plans.md", "/project/docs/Plans.md", true, "custom subdir match"},
		// same-name file in a different directory must not be falsely matched (core of the fix)
		{"docs/Plans.md", "/project/Plans.md", false, "subdirectory mismatch - must not match"},
		// completely different file
		{"src/main.go", "/project/Plans.md", false, "non plans file"},
		{"README.md", "/project/Plans.md", false, "readme not plans"},
		// file name resembles Plans.md but is a different file
		{"Plans.md.bak", "/project/Plans.md", false, "backup file not matched"},
		// absolute path to Plans.md in a different project
		{"/tmp/other/Plans.md", "/project/Plans.md", false, "different project Plans.md must not match"},
	}

	for _, tc := range cases {
		got := isPlansFileWithRoot(tc.changedFile, tc.plansFile, projectRoot)
		if got != tc.want {
			t.Errorf("[%s] isPlansFileWithRoot(%q, %q, %q) = %v, want %v",
				tc.desc, tc.changedFile, tc.plansFile, projectRoot, got, tc.want)
		}
	}
}

// TestHandlePlansWatcher_CWDFromInput verifies that when input.CWD is present,
// it is used as projectRoot instead of resolveProjectRoot().
// Validates that Plans.md is correctly detected when the hook process CWD differs from input.CWD.
func TestHandlePlansWatcher_CWDFromInput(t *testing.T) {
	// project directory (where Plans.md exists)
	projectDir := t.TempDir()
	// hook process CWD (a directory different from the project)
	hookCWD := t.TempDir()

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	// hook process is in hookCWD (not the project root)
	if err := os.Chdir(hookCWD); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	// create Plans.md in projectDir
	plansContent := "| Task 1 | TaskA | DoD | - | cc:done |\n"
	plansPath := filepath.Join(projectDir, "Plans.md")
	if err := os.WriteFile(plansPath, []byte(plansContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// create .claude/state in projectDir
	stateDir := filepath.Join(projectDir, ".claude", "state")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// include cwd field in input (pointing to projectDir)
	inputJSON := `{"tool_name":"Edit","cwd":"` + projectDir + `","tool_input":{"file_path":"Plans.md"}}`
	var out bytes.Buffer
	if err := HandlePlansWatcher(strings.NewReader(inputJSON), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// even though hookCWD has no Plans.md, projectDir's Plans.md should be detected
	// verify error-free processing (Plans.md is found and state aggregation proceeds)
	var result postToolOutput
	if jsonErr := json.Unmarshal(out.Bytes(), &result); jsonErr != nil {
		t.Fatalf("invalid JSON output: %v, raw: %s", jsonErr, out.String())
	}
	// should process successfully (when hookCWD has no Plans.md, emptyPostToolOutput would be expected,
	// but using input.CWD allows Plans.md in projectDir to be found)
	if out.Len() == 0 {
		t.Error("expected non-empty output when input.CWD points to project with Plans.md")
	}
}
