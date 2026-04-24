package hookhandler

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAutoCleanupHandler_EmptyInput(t *testing.T) {
	h := &AutoCleanupHandler{ProjectRoot: t.TempDir()}

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(""), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 空入力は出力なし
	if out.Len() != 0 {
		t.Errorf("expected no output, got %q", out.String())
	}
}

func TestAutoCleanupHandler_NoFilePath(t *testing.T) {
	h := &AutoCleanupHandler{ProjectRoot: t.TempDir()}

	input := `{"tool_name":"Write","tool_input":{}}`

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(input), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Len() != 0 {
		t.Errorf("expected no output for missing file_path, got %q", out.String())
	}
}

func TestAutoCleanupHandler_UnrelatedFile(t *testing.T) {
	dir := t.TempDir()
	h := &AutoCleanupHandler{ProjectRoot: dir}

	// Plans.md / session-log.md / CLAUDE.md 以外のファイル
	fpath := filepath.Join(dir, "README.md")
	content := strings.Repeat("line\n", 300)
	_ = os.WriteFile(fpath, []byte(content), 0600)

	input := `{"tool_name":"Write","tool_input":{"file_path":"` + fpath + `"}}`

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(input), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Len() != 0 {
		t.Errorf("expected no output for unrelated file, got %q", out.String())
	}
}

func TestAutoCleanupHandler_Plansmd_UnderThreshold(t *testing.T) {
	dir := t.TempDir()
	h := &AutoCleanupHandler{ProjectRoot: dir, PlansMaxLines: 200}

	fpath := filepath.Join(dir, "Plans.md")
	content := strings.Repeat("line\n", 100) // 100 行 < 200
	_ = os.WriteFile(fpath, []byte(content), 0600)

	input := `{"tool_name":"Write","tool_input":{"file_path":"` + fpath + `"},"cwd":"` + dir + `"}`

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(input), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Len() != 0 {
		t.Errorf("expected no output under threshold, got %q", out.String())
	}
}

func TestAutoCleanupHandler_PlansmdOverThreshold(t *testing.T) {
	t.Setenv("CLAUDE_CODE_HARNESS_LANG", "")
	dir := t.TempDir()
	h := &AutoCleanupHandler{ProjectRoot: dir, PlansMaxLines: 200}

	fpath := filepath.Join(dir, "Plans.md")
	content := strings.Repeat("line\n", 250) // 250 行 > 200
	_ = os.WriteFile(fpath, []byte(content), 0600)

	input := `{"tool_name":"Write","tool_input":{"file_path":"` + fpath + `"},"cwd":"` + dir + `"}`

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(input), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.Len() == 0 {
		t.Fatalf("expected warning output, got nothing")
	}

	var result struct {
		HookSpecificOutput struct {
			AdditionalContext string `json:"additionalContext"`
		} `json:"hookSpecificOutput"`
	}
	if err := json.Unmarshal(bytes.TrimRight(out.Bytes(), "\n"), &result); err != nil {
		t.Fatalf("invalid JSON: %s", out.String())
	}

	ctx := result.HookSpecificOutput.AdditionalContext
	if !strings.Contains(ctx, "Plans.md") {
		t.Errorf("expected Plans.md warning, got %q", ctx)
	}
	if !strings.Contains(ctx, "250") {
		t.Errorf("expected line count 250 in warning, got %q", ctx)
	}
}

func TestAutoCleanupHandler_LocaleDefaultEnglish(t *testing.T) {
	t.Setenv("CLAUDE_CODE_HARNESS_LANG", "")
	dir := t.TempDir()
	h := &AutoCleanupHandler{ProjectRoot: dir, PlansMaxLines: 2}

	fpath := filepath.Join(dir, "Plans.md")
	_ = os.WriteFile(fpath, []byte("one\ntwo\nthree\n"), 0600)

	input := `{"tool_name":"Write","tool_input":{"file_path":"` + fpath + `"},"cwd":"` + dir + `"}`
	var out bytes.Buffer
	if err := h.Handle(strings.NewReader(input), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := parseAutoCleanupContext(t, &out)
	if !strings.Contains(ctx, "Warning: Plans.md has 3 lines") {
		t.Fatalf("default additionalContext should be English, got %q", ctx)
	}
	if strings.Contains(ctx, "行です") {
		t.Fatalf("default additionalContext should not be Japanese, got %q", ctx)
	}
}

func TestAutoCleanupHandler_LocaleJapaneseEnv(t *testing.T) {
	t.Setenv("CLAUDE_CODE_HARNESS_LANG", "ja")
	dir := t.TempDir()
	h := &AutoCleanupHandler{ProjectRoot: dir, PlansMaxLines: 2}

	fpath := filepath.Join(dir, "Plans.md")
	_ = os.WriteFile(fpath, []byte("one\ntwo\nthree\n"), 0600)

	input := `{"tool_name":"Write","tool_input":{"file_path":"` + fpath + `"},"cwd":"` + dir + `"}`
	var out bytes.Buffer
	if err := h.Handle(strings.NewReader(input), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := parseAutoCleanupContext(t, &out)
	if !strings.Contains(ctx, "Plans.md が 3 行です") {
		t.Fatalf("ja env additionalContext should be Japanese, got %q", ctx)
	}
}

func TestAutoCleanupHandler_LocaleJapaneseConfig(t *testing.T) {
	t.Setenv("CLAUDE_CODE_HARNESS_LANG", "en")
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, harnessConfigFileName), []byte("i18n:\n  language: ja\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	h := &AutoCleanupHandler{ProjectRoot: dir, SessionLogMaxLines: 2}

	fpath := filepath.Join(dir, "session-log.md")
	_ = os.WriteFile(fpath, []byte("one\ntwo\nthree\n"), 0600)

	input := `{"tool_name":"Write","tool_input":{"file_path":"` + fpath + `"},"cwd":"` + dir + `"}`
	var out bytes.Buffer
	if err := h.Handle(strings.NewReader(input), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := parseAutoCleanupContext(t, &out)
	if !strings.Contains(ctx, "session-log.md が 3 行です") {
		t.Fatalf("config ja additionalContext should be Japanese, got %q", ctx)
	}
}

func parseAutoCleanupContext(t *testing.T, out *bytes.Buffer) string {
	t.Helper()
	if out.Len() == 0 {
		t.Fatal("expected warning output, got nothing")
	}
	var result struct {
		HookSpecificOutput struct {
			HookEventName     string `json:"hookEventName"`
			AdditionalContext string `json:"additionalContext"`
		} `json:"hookSpecificOutput"`
	}
	if err := json.Unmarshal(bytes.TrimRight(out.Bytes(), "\n"), &result); err != nil {
		t.Fatalf("invalid JSON: %s", out.String())
	}
	if result.HookSpecificOutput.HookEventName != "PostToolUse" {
		t.Fatalf("hookEventName = %q, want PostToolUse", result.HookSpecificOutput.HookEventName)
	}
	return result.HookSpecificOutput.AdditionalContext
}

func TestAutoCleanupHandler_SessionLog_OverThreshold(t *testing.T) {
	dir := t.TempDir()
	h := &AutoCleanupHandler{ProjectRoot: dir, SessionLogMaxLines: 500}

	fpath := filepath.Join(dir, "session-log.md")
	content := strings.Repeat("line\n", 600) // 600 行 > 500
	_ = os.WriteFile(fpath, []byte(content), 0600)

	input := `{"tool_name":"Write","tool_input":{"file_path":"` + fpath + `"},"cwd":"` + dir + `"}`

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(input), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.Len() == 0 {
		t.Fatalf("expected warning output")
	}

	var result struct {
		HookSpecificOutput struct {
			AdditionalContext string `json:"additionalContext"`
		} `json:"hookSpecificOutput"`
	}
	_ = json.Unmarshal(bytes.TrimRight(out.Bytes(), "\n"), &result)
	ctx := result.HookSpecificOutput.AdditionalContext
	if !strings.Contains(ctx, "session-log.md") {
		t.Errorf("expected session-log.md warning, got %q", ctx)
	}
}

func TestAutoCleanupHandler_ClaudeMd_OverThreshold(t *testing.T) {
	dir := t.TempDir()
	h := &AutoCleanupHandler{ProjectRoot: dir, ClaudeMdMaxLines: 100}

	fpath := filepath.Join(dir, "CLAUDE.md")
	content := strings.Repeat("line\n", 150) // 150 行 > 100
	_ = os.WriteFile(fpath, []byte(content), 0600)

	input := `{"tool_name":"Write","tool_input":{"file_path":"` + fpath + `"},"cwd":"` + dir + `"}`

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(input), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.Len() == 0 {
		t.Fatalf("expected warning output")
	}

	var result struct {
		HookSpecificOutput struct {
			AdditionalContext string `json:"additionalContext"`
		} `json:"hookSpecificOutput"`
	}
	_ = json.Unmarshal(bytes.TrimRight(out.Bytes(), "\n"), &result)
	ctx := result.HookSpecificOutput.AdditionalContext
	if !strings.Contains(ctx, "CLAUDE.md") {
		t.Errorf("expected CLAUDE.md warning, got %q", ctx)
	}
}

func TestAutoCleanupHandler_PlansmdArchive_WithSSOTFlag(t *testing.T) {
	dir := t.TempDir()
	h := &AutoCleanupHandler{ProjectRoot: dir, PlansMaxLines: 200}

	// SSOT フラグを事前に作成（警告なし）
	stateDir := filepath.Join(dir, ".claude", "state")
	_ = os.MkdirAll(stateDir, 0700)
	_ = os.WriteFile(filepath.Join(stateDir, ".ssot-synced-this-session"), []byte(""), 0600)

	fpath := filepath.Join(dir, "Plans.md")
	content := "## アーカイブ\n" + strings.Repeat("line\n", 10)
	_ = os.WriteFile(fpath, []byte(content), 0600)

	input := `{"tool_name":"Write","tool_input":{"file_path":"` + fpath + `"},"cwd":"` + dir + `"}`

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(input), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// SSOT フラグあり → アーカイブ警告は出ない
	if out.Len() != 0 {
		var result struct {
			HookSpecificOutput struct {
				AdditionalContext string `json:"additionalContext"`
			} `json:"hookSpecificOutput"`
		}
		_ = json.Unmarshal(bytes.TrimRight(out.Bytes(), "\n"), &result)
		if strings.Contains(result.HookSpecificOutput.AdditionalContext, "memory sync") {
			t.Errorf("expected no SSOT warning with flag present, got %q", result.HookSpecificOutput.AdditionalContext)
		}
	}
}

func TestAutoCleanupHandler_PlansmdArchive_NoSSOTFlag(t *testing.T) {
	dir := t.TempDir()
	h := &AutoCleanupHandler{ProjectRoot: dir, PlansMaxLines: 200}

	// SSOT フラグなし
	fpath := filepath.Join(dir, "Plans.md")
	content := "## アーカイブ\n" + strings.Repeat("line\n", 10)
	_ = os.WriteFile(fpath, []byte(content), 0600)

	input := `{"tool_name":"Write","tool_input":{"file_path":"` + fpath + `"},"cwd":"` + dir + `"}`

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(input), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.Len() == 0 {
		t.Fatalf("expected SSOT warning output")
	}

	var result struct {
		HookSpecificOutput struct {
			AdditionalContext string `json:"additionalContext"`
		} `json:"hookSpecificOutput"`
	}
	_ = json.Unmarshal(bytes.TrimRight(out.Bytes(), "\n"), &result)
	ctx := result.HookSpecificOutput.AdditionalContext
	if !strings.Contains(ctx, "memory sync") {
		t.Errorf("expected memory sync warning, got %q", ctx)
	}
}

func TestAutoCleanupHandler_ToolResponseFilePath(t *testing.T) {
	dir := t.TempDir()
	h := &AutoCleanupHandler{ProjectRoot: dir, ClaudeMdMaxLines: 100}

	fpath := filepath.Join(dir, "CLAUDE.md")
	content := strings.Repeat("line\n", 150)
	_ = os.WriteFile(fpath, []byte(content), 0600)

	// tool_response.filePath を使用するケース
	input := `{"tool_name":"Edit","tool_input":{},"tool_response":{"filePath":"` + fpath + `"},"cwd":"` + dir + `"}`

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(input), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.Len() == 0 {
		t.Fatalf("expected warning output via tool_response.filePath")
	}
}

func TestCountLines(t *testing.T) {
	dir := t.TempDir()
	fpath := filepath.Join(dir, "test.txt")

	tests := []struct {
		content string
		want    int
	}{
		{"", 0},
		{"line1\n", 1},
		{"line1\nline2\n", 2},
		{"line1\nline2\nline3", 3}, // 末尾改行なし
	}

	for _, tt := range tests {
		_ = os.WriteFile(fpath, []byte(tt.content), 0600)
		got, err := countLines(fpath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != tt.want {
			t.Errorf("countLines(%q) = %d, want %d", tt.content, got, tt.want)
		}
	}
}

func TestContainsArchiveSection(t *testing.T) {
	dir := t.TempDir()
	fpath := filepath.Join(dir, "test.md")

	tests := []struct {
		content string
		want    bool
	}{
		{"# Tasks\n## TODO\n", false},
		{"## アーカイブ\n", true},
		{"📦 アーカイブ済み\n", true},
		{"## Archive\n", true},
		{"# Normal\nsome text\n", false},
	}

	for _, tt := range tests {
		_ = os.WriteFile(fpath, []byte(tt.content), 0600)
		got := containsArchiveSection(fpath)
		if got != tt.want {
			t.Errorf("containsArchiveSection(%q) = %v, want %v", tt.content, got, tt.want)
		}
	}
}
