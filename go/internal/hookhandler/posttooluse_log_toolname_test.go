package hookhandler

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestPostToolUseLogToolName_AlwaysContinue verifies that continue=true is always returned.
func TestPostToolUseLogToolName_AlwaysContinue(t *testing.T) {
	dir := t.TempDir()
	h := &PostToolUseLogToolNameHandler{ProjectRoot: dir}

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(`{"tool_name":"Read"}`), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp map[string]bool
	if err := json.Unmarshal(bytes.TrimRight(out.Bytes(), "\n"), &resp); err != nil {
		t.Fatalf("invalid JSON: %s", out.String())
	}
	if !resp["continue"] {
		t.Errorf("expected continue=true")
	}
}

// TestPostToolUseLogToolName_EmptyInput verifies that continue=true is returned even for empty input.
func TestPostToolUseLogToolName_EmptyInput(t *testing.T) {
	dir := t.TempDir()
	h := &PostToolUseLogToolNameHandler{ProjectRoot: dir}

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(""), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp map[string]bool
	if err := json.Unmarshal(bytes.TrimRight(out.Bytes(), "\n"), &resp); err != nil {
		t.Fatalf("invalid JSON: %s", out.String())
	}
	if !resp["continue"] {
		t.Errorf("expected continue=true for empty input")
	}
}

// TestPostToolUseLogToolName_Phase0Log verifies that events are recorded in tool-events.jsonl when CC_HARNESS_PHASE0_LOG=1.
func TestPostToolUseLogToolName_Phase0Log(t *testing.T) {
	dir := t.TempDir()
	stateDir := filepath.Join(dir, ".claude", "state")
	if err := os.MkdirAll(stateDir, 0700); err != nil {
		t.Fatal(err)
	}

	t.Setenv("CC_HARNESS_PHASE0_LOG", "1")

	h := &PostToolUseLogToolNameHandler{ProjectRoot: dir}

	input := `{"tool_name":"Read","session_id":"test-session"}`
	var out bytes.Buffer
	err := h.Handle(strings.NewReader(input), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// verify tool-events.jsonl is created
	logFile := filepath.Join(stateDir, toolEventsFile)
	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("expected tool-events.jsonl to be created: %v", err)
	}

	// verify JSONL entry
	var entry toolEventEntry
	if err := json.Unmarshal(bytes.TrimRight(data, "\n"), &entry); err != nil {
		t.Fatalf("invalid JSONL entry: %s", string(data))
	}

	if entry.ToolName != "Read" {
		t.Errorf("expected tool_name=Read, got %s", entry.ToolName)
	}
	if entry.SessionID != "test-session" {
		t.Errorf("expected session_id=test-session, got %s", entry.SessionID)
	}
	if entry.HookEventName != "PostToolUse" {
		t.Errorf("expected hook_event_name=PostToolUse, got %s", entry.HookEventName)
	}
	if entry.V != 1 {
		t.Errorf("expected v=1, got %d", entry.V)
	}
	if entry.Ts == "" {
		t.Errorf("expected non-empty timestamp")
	}
}

// TestPostToolUseLogToolName_Phase0LogDisabled verifies that no recording occurs when CC_HARNESS_PHASE0_LOG is not set.
func TestPostToolUseLogToolName_Phase0LogDisabled(t *testing.T) {
	dir := t.TempDir()
	stateDir := filepath.Join(dir, ".claude", "state")
	if err := os.MkdirAll(stateDir, 0700); err != nil {
		t.Fatal(err)
	}

	// disable CC_HARNESS_PHASE0_LOG
	t.Setenv("CC_HARNESS_PHASE0_LOG", "0")

	h := &PostToolUseLogToolNameHandler{ProjectRoot: dir}
	var out bytes.Buffer
	_ = h.Handle(strings.NewReader(`{"tool_name":"Read"}`), &out)

	// verify tool-events.jsonl is NOT created
	logFile := filepath.Join(stateDir, toolEventsFile)
	if _, err := os.Stat(logFile); err == nil {
		t.Errorf("expected tool-events.jsonl to NOT be created when Phase0 log is disabled")
	}
}

func TestPostToolUseLogToolName_LSPTracking(t *testing.T) {
	dir := t.TempDir()
	stateDir := filepath.Join(dir, ".claude", "state")
	if err := os.MkdirAll(stateDir, 0700); err != nil {
		t.Fatal(err)
	}

	sessionData := `{"prompt_seq": 3}`
	if err := os.WriteFile(filepath.Join(stateDir, "session.json"), []byte(sessionData), 0600); err != nil {
		t.Fatal(err)
	}
	policyData := `{"lsp": {"available": true, "used_since_last_prompt": false}, "skills": {}}`
	if err := os.WriteFile(filepath.Join(stateDir, "tooling-policy.json"), []byte(policyData), 0600); err != nil {
		t.Fatal(err)
	}

	h := &PostToolUseLogToolNameHandler{ProjectRoot: dir}

	input := `{"tool_name":"harness_lsp_definition","session_id":"s1"}`
	var out bytes.Buffer
	err := h.Handle(strings.NewReader(input), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rawPolicy, err := os.ReadFile(filepath.Join(stateDir, "tooling-policy.json"))
	if err != nil {
		t.Fatalf("failed to read tooling-policy.json: %v", err)
	}
	var policy map[string]interface{}
	if err := json.Unmarshal(rawPolicy, &policy); err != nil {
		t.Fatalf("invalid tooling-policy.json: %s", string(rawPolicy))
	}

	lsp, ok := policy["lsp"].(map[string]interface{})
	if !ok {
		t.Fatal("expected lsp object in policy")
	}
	if used, _ := lsp["used_since_last_prompt"].(bool); !used {
		t.Errorf("expected used_since_last_prompt=true after LSP tool use")
	}
	if name, _ := lsp["last_used_tool_name"].(string); name != "harness_lsp_definition" {
		t.Errorf("expected last_used_tool_name=harness_lsp_definition, got %s", name)
	}
	if seq, _ := lsp["last_used_prompt_seq"].(float64); int(seq) != 3 {
		t.Errorf("expected last_used_prompt_seq=3, got %v", lsp["last_used_prompt_seq"])
	}
}

func TestPostToolUseLogToolName_NonLSPToolNoPolicy(t *testing.T) {
	dir := t.TempDir()
	stateDir := filepath.Join(dir, ".claude", "state")
	if err := os.MkdirAll(stateDir, 0700); err != nil {
		t.Fatal(err)
	}

	policyData := `{"lsp": {"used_since_last_prompt": false}}`
	policyPath := filepath.Join(stateDir, "tooling-policy.json")
	if err := os.WriteFile(policyPath, []byte(policyData), 0600); err != nil {
		t.Fatal(err)
	}

	h := &PostToolUseLogToolNameHandler{ProjectRoot: dir}
	var out bytes.Buffer
	_ = h.Handle(strings.NewReader(`{"tool_name":"Bash"}`), &out)

	rawAfter, _ := os.ReadFile(policyPath)
	if string(rawAfter) != policyData {
		t.Errorf("expected policy to be unchanged for non-LSP tool, got: %s", string(rawAfter))
	}
}

func TestPostToolUseLogToolName_SessionEventLog(t *testing.T) {
	dir := t.TempDir()
	stateDir := filepath.Join(dir, ".claude", "state")
	if err := os.MkdirAll(stateDir, 0700); err != nil {
		t.Fatal(err)
	}

	sessionData := `{"prompt_seq": 1, "event_seq": 0, "state": "executing"}`
	if err := os.WriteFile(filepath.Join(stateDir, "session.json"), []byte(sessionData), 0600); err != nil {
		t.Fatal(err)
	}

	h := &PostToolUseLogToolNameHandler{ProjectRoot: dir}

	input := `{"tool_name":"Write","session_id":"s1","tool_input":{"file_path":"/foo/bar.go"}}`
	var out bytes.Buffer
	err := h.Handle(strings.NewReader(input), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	eventFile := filepath.Join(stateDir, sessionEventsFile)
	data, err := os.ReadFile(eventFile)
	if err != nil {
		t.Fatalf("expected session-events.jsonl to be created: %v", err)
	}

	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) == 0 || lines[0] == "" {
		t.Fatal("expected at least one event in session-events.jsonl")
	}

	var event map[string]interface{}
	if err := json.Unmarshal([]byte(lines[0]), &event); err != nil {
		t.Fatalf("invalid event JSON: %s", lines[0])
	}
	if event["type"] != "tool.write" {
		t.Errorf("expected type=tool.write, got %v", event["type"])
	}
	if event["id"] == "" {
		t.Errorf("expected non-empty event id")
	}
}

func TestPostToolUseLogToolName_SkillTracking(t *testing.T) {
	dir := t.TempDir()
	stateDir := filepath.Join(dir, ".claude", "state")
	if err := os.MkdirAll(stateDir, 0700); err != nil {
		t.Fatal(err)
	}

	h := &PostToolUseLogToolNameHandler{ProjectRoot: dir}

	input := `{"tool_name":"Skill","tool_input":{"skill":"harness-review"}}`
	var out bytes.Buffer
	err := h.Handle(strings.NewReader(input), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	skillsFile := filepath.Join(stateDir, skillsUsedFile)
	rawData, err := os.ReadFile(skillsFile)
	if err != nil {
		t.Fatalf("expected session-skills-used.json to be created: %v", err)
	}

	var state skillsUsedState
	if err := json.Unmarshal(rawData, &state); err != nil {
		t.Fatalf("invalid skills JSON: %s", string(rawData))
	}
	if len(state.Used) != 1 {
		t.Fatalf("expected 1 skill used, got %d", len(state.Used))
	}
	if state.Used[0] != "harness-review" {
		t.Errorf("expected harness-review in used skills, got %s", state.Used[0])
	}
	if state.LastUsed == "" {
		t.Errorf("expected last_used to be set")
	}
}

func TestPostToolUseLogToolName_SkillTrackingMultiple(t *testing.T) {
	dir := t.TempDir()
	stateDir := filepath.Join(dir, ".claude", "state")
	if err := os.MkdirAll(stateDir, 0700); err != nil {
		t.Fatal(err)
	}

	h := &PostToolUseLogToolNameHandler{ProjectRoot: dir}

	skills := []string{"harness-review", "harness-work", "harness-remember"}
	for _, s := range skills {
		input, _ := json.Marshal(map[string]interface{}{
			"tool_name":  "Skill",
			"tool_input": map[string]string{"skill": s},
		})
		var out bytes.Buffer
		_ = h.Handle(bytes.NewReader(input), &out)
	}

	skillsFile := filepath.Join(stateDir, skillsUsedFile)
	rawData, _ := os.ReadFile(skillsFile)
	var state skillsUsedState
	_ = json.Unmarshal(rawData, &state)

	if len(state.Used) != 3 {
		t.Errorf("expected 3 skills used, got %d", len(state.Used))
	}
}

func TestIsImportantTool(t *testing.T) {
	important := []string{"Write", "Edit", "Bash", "Task", "Skill", "SlashCommand"}
	notImportant := []string{"Read", "Glob", "Grep", "harness_lsp_definition", "unknown"}

	for _, tool := range important {
		if !isImportantTool(tool) {
			t.Errorf("expected %s to be important", tool)
		}
	}
	for _, tool := range notImportant {
		if isImportantTool(tool) {
			t.Errorf("expected %s to NOT be important", tool)
		}
	}
}

func TestNeedsRotation(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.jsonl")

	if needsRotation(path, 100, 10) {
		t.Error("expected false for non-existent file")
	}

	if err := os.WriteFile(path, []byte("line1\nline2\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if needsRotation(path, 100, 10) {
		t.Error("expected false for small file")
	}

	if needsRotation(path, 5, 10) == false {
		t.Error("expected true when file exceeds size limit")
	}
}

func TestRotateLog(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.jsonl")

	if err := os.WriteFile(path, []byte("original content\n"), 0600); err != nil {
		t.Fatal(err)
	}

	rotateLog(path, 5)

	if _, err := os.Stat(path + ".1"); err != nil {
		t.Errorf("expected %s.1 to exist after rotation", path)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected new %s to exist after rotation", path)
	}

	data1, _ := os.ReadFile(path + ".1")
	if !strings.Contains(string(data1), "original content") {
		t.Errorf("expected original content in .1 file")
	}
}

func TestBuildEventData(t *testing.T) {
	tests := []struct {
		inp      logToolNameInput
		wantKey  string
		wantVal  string
	}{
		{
			inp:     logToolNameInput{ToolInput: logToolNameToolInput{FilePath: "/foo/bar.go"}},
			wantKey: "file_path",
			wantVal: "/foo/bar.go",
		},
		{
			inp:     logToolNameInput{ToolInput: logToolNameToolInput{Command: "go test ./..."}},
			wantKey: "command",
			wantVal: "go test ./...",
		},
		{
			inp:     logToolNameInput{},
			wantKey: "",
		},
	}

	for _, tc := range tests {
		result := buildEventData(tc.inp)
		if tc.wantKey == "" {
			if result != "" {
				t.Errorf("expected empty result for empty input, got %s", result)
			}
			continue
		}
		if !strings.Contains(result, tc.wantKey) {
			t.Errorf("expected result to contain key %q, got %s", tc.wantKey, result)
		}
		if !strings.Contains(result, tc.wantVal) {
			t.Errorf("expected result to contain value %q, got %s", tc.wantVal, result)
		}
	}
}
