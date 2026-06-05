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
	content := strings.Repeat("line\n", 100)
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
	dir := t.TempDir()
	h := &AutoCleanupHandler{ProjectRoot: dir, PlansMaxLines: 200}

	fpath := filepath.Join(dir, "Plans.md")
	content := strings.Repeat("line\n", 250)
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

func TestAutoCleanupHandler_ClaudeMd_OverThreshold(t *testing.T) {
	dir := t.TempDir()
	h := &AutoCleanupHandler{ProjectRoot: dir, ClaudeMdMaxLines: 100}

	fpath := filepath.Join(dir, "CLAUDE.md")
	content := strings.Repeat("line\n", 150)
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

func TestAutoCleanupHandler_ToolResponseFilePath(t *testing.T) {
	dir := t.TempDir()
	h := &AutoCleanupHandler{ProjectRoot: dir, ClaudeMdMaxLines: 100}

	fpath := filepath.Join(dir, "CLAUDE.md")
	content := strings.Repeat("line\n", 150)
	_ = os.WriteFile(fpath, []byte(content), 0600)

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
		{"line1\nline2\nline3", 3},
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

