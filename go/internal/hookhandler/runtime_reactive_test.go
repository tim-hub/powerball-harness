package hookhandler

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// assertReactiveApprove は approve レスポンスを検証するヘルパー。
func assertReactiveApprove(t *testing.T, output, wantReason string) {
	t.Helper()
	output = strings.TrimSpace(output)
	if output == "" {
		t.Fatal("expected JSON output, got empty")
	}
	var resp map[string]string
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, output)
	}
	if resp["decision"] != "approve" {
		t.Errorf("decision = %q, want approve", resp["decision"])
	}
	if wantReason != "" && !strings.Contains(resp["reason"], wantReason) {
		t.Errorf("reason = %q, want to contain %q", resp["reason"], wantReason)
	}
}

// assertReactiveHookOutput は hookSpecificOutput レスポンスを検証するヘルパー。
func assertReactiveHookOutput(t *testing.T, output, wantEvent, wantContextSubstr string) {
	t.Helper()
	output = strings.TrimSpace(output)
	if output == "" {
		t.Fatal("expected JSON output, got empty")
	}
	var resp map[string]interface{}
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, output)
	}
	hookOut, ok := resp["hookSpecificOutput"].(map[string]interface{})
	if !ok {
		t.Fatalf("missing hookSpecificOutput in: %s", output)
	}
	if hookOut["hookEventName"] != wantEvent {
		t.Errorf("hookEventName = %q, want %q", hookOut["hookEventName"], wantEvent)
	}
	ctx, _ := hookOut["additionalContext"].(string)
	if wantContextSubstr != "" && !strings.Contains(ctx, wantContextSubstr) {
		t.Errorf("additionalContext = %q, want to contain %q", ctx, wantContextSubstr)
	}
}

func TestHandleRuntimeReactive_EmptyInput(t *testing.T) {
	var out bytes.Buffer
	if err := HandleRuntimeReactive(strings.NewReader(""), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertReactiveApprove(t, out.String(), "no payload")
}

func TestHandleRuntimeReactive_InvalidJSON(t *testing.T) {
	var out bytes.Buffer
	if err := HandleRuntimeReactive(strings.NewReader("not json"), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertReactiveApprove(t, out.String(), "no payload")
}

func TestHandleRuntimeReactive_FileChanged_Plansmd(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PROJECT_ROOT", dir)

	payload := `{"hook_event_name":"FileChanged","file_path":"Plans.md","cwd":"` + dir + `"}`
	var out bytes.Buffer
	if err := HandleRuntimeReactive(strings.NewReader(payload), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertReactiveHookOutput(t, out.String(), "FileChanged", "Plans.md")
}

func TestHandleRuntimeReactive_FileChanged_PlansmdWithPath(t *testing.T) {
	dir := t.TempDir()

	// プロジェクトルートを含む絶対パスの場合
	payload := `{"hook_event_name":"FileChanged","file_path":"` + dir + `/Plans.md","cwd":"` + dir + `"}`
	var out bytes.Buffer
	if err := HandleRuntimeReactive(strings.NewReader(payload), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertReactiveHookOutput(t, out.String(), "FileChanged", "Plans.md")
}

func TestHandleRuntimeReactive_FileChanged_AgentsMd(t *testing.T) {
	dir := t.TempDir()

	payload := `{"hook_event_name":"FileChanged","file_path":"AGENTS.md","cwd":"` + dir + `"}`
	var out bytes.Buffer
	if err := HandleRuntimeReactive(strings.NewReader(payload), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertReactiveHookOutput(t, out.String(), "FileChanged", "Harness 設定")
}

func TestHandleRuntimeReactive_FileChanged_ClaudeRules(t *testing.T) {
	dir := t.TempDir()

	payload := `{"hook_event_name":"FileChanged","file_path":".claude/rules/test.md","cwd":"` + dir + `"}`
	var out bytes.Buffer
	if err := HandleRuntimeReactive(strings.NewReader(payload), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertReactiveHookOutput(t, out.String(), "FileChanged", "Harness 設定")
}

func TestHandleRuntimeReactive_FileChanged_HooksJson(t *testing.T) {
	dir := t.TempDir()

	payload := `{"hook_event_name":"FileChanged","file_path":"hooks/hooks.json","cwd":"` + dir + `"}`
	var out bytes.Buffer
	if err := HandleRuntimeReactive(strings.NewReader(payload), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertReactiveHookOutput(t, out.String(), "FileChanged", "Harness 設定")
}

func TestHandleRuntimeReactive_FileChanged_SettingsJson(t *testing.T) {
	dir := t.TempDir()

	payload := `{"hook_event_name":"FileChanged","file_path":".claude-plugin/settings.json","cwd":"` + dir + `"}`
	var out bytes.Buffer
	if err := HandleRuntimeReactive(strings.NewReader(payload), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertReactiveHookOutput(t, out.String(), "FileChanged", "Harness 設定")
}

func TestHandleRuntimeReactive_FileChanged_UnrelatedFile(t *testing.T) {
	dir := t.TempDir()

	payload := `{"hook_event_name":"FileChanged","file_path":"src/main.go","cwd":"` + dir + `"}`
	var out bytes.Buffer
	if err := HandleRuntimeReactive(strings.NewReader(payload), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 無関係ファイルの場合は approve を返す
	assertReactiveApprove(t, out.String(), "Reactive hook tracked")
}

func TestHandleRuntimeReactive_CwdChanged(t *testing.T) {
	dir := t.TempDir()

	payload := `{"hook_event_name":"CwdChanged","previous_cwd":"/old/path","cwd":"` + dir + `"}`
	var out bytes.Buffer
	if err := HandleRuntimeReactive(strings.NewReader(payload), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertReactiveHookOutput(t, out.String(), "CwdChanged", "作業ディレクトリが切り替わりました")
}

func TestHandleRuntimeReactive_TaskCreated(t *testing.T) {
	dir := t.TempDir()

	payload := `{"hook_event_name":"TaskCreated","task_id":"t1","task_title":"test task","cwd":"` + dir + `"}`
	var out bytes.Buffer
	if err := HandleRuntimeReactive(strings.NewReader(payload), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// TaskCreated はログのみ → approve を返す
	assertReactiveApprove(t, out.String(), "Reactive hook tracked")
}

func TestHandleRuntimeReactive_WritesLogFile(t *testing.T) {
	dir := t.TempDir()

	payload := `{"hook_event_name":"FileChanged","file_path":"Plans.md","session_id":"sess-1","cwd":"` + dir + `"}`
	var out bytes.Buffer
	if err := HandleRuntimeReactive(strings.NewReader(payload), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// ログファイルが作成されているか確認
	logFile := filepath.Join(dir, ".claude", "state", "runtime-reactive.jsonl")
	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("log file not created: %v", err)
	}

	if !strings.Contains(string(data), "FileChanged") {
		t.Errorf("log file does not contain FileChanged event\ncontents: %s", data)
	}
	if !strings.Contains(string(data), "sess-1") {
		t.Errorf("log file does not contain session_id\ncontents: %s", data)
	}
}

func TestHandleRuntimeReactive_AlternativeFieldNames(t *testing.T) {
	dir := t.TempDir()

	// event_name と path の代替フィールドを使う
	payload := `{"event_name":"FileChanged","path":"CLAUDE.md","project_root":"` + dir + `"}`
	var out bytes.Buffer
	if err := HandleRuntimeReactive(strings.NewReader(payload), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertReactiveHookOutput(t, out.String(), "FileChanged", "Harness 設定")
}

func TestNormalizeReactivePath(t *testing.T) {
	tests := []struct {
		rawPath    string
		projectRoot string
		want       string
	}{
		{"Plans.md", "", "Plans.md"},
		{"", "/some/root", ""},
		{"/root/Plans.md", "/root", "Plans.md"},
		{"/root", "/root", "."},
		{"./Plans.md", "", "Plans.md"},
		{"/root/subdir/file.md", "/root", "subdir/file.md"},
	}

	for _, tt := range tests {
		got := normalizeReactivePath(tt.rawPath, tt.projectRoot)
		if got != tt.want {
			t.Errorf("normalizeReactivePath(%q, %q) = %q, want %q", tt.rawPath, tt.projectRoot, got, tt.want)
		}
	}
}

func TestIsRuleOrConfigFile(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"AGENTS.md", true},
		{"sub/AGENTS.md", true},
		{"CLAUDE.md", true},
		{".claude/rules/test.md", true},
		{"sub/.claude/rules/foo.md", true},
		{"hooks/hooks.json", true},
		{"sub/hooks/hooks.json", true},
		{".claude-plugin/settings.json", true},
		{"sub/.claude-plugin/settings.json", true},
		{"Plans.md", false},
		{"src/main.go", false},
		{"README.md", false},
	}

	for _, tt := range tests {
		got := isRuleOrConfigFile(tt.path)
		if got != tt.want {
			t.Errorf("isRuleOrConfigFile(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}
