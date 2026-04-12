package hookhandler

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandlePostToolUseQualityPack_EmptyInput(t *testing.T) {
	var out bytes.Buffer
	err := HandlePostToolUseQualityPack(strings.NewReader(""), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Len() != 0 {
		t.Errorf("expected no output for empty input, got %q", out.String())
	}
}

func TestHandlePostToolUseQualityPack_NonWriteEdit(t *testing.T) {
	// Read ツールは対象外
	input := `{"tool_name":"Read","tool_input":{"file_path":"src/foo.ts"},"cwd":"/tmp"}`
	var out bytes.Buffer
	if err := HandlePostToolUseQualityPack(strings.NewReader(input), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Len() != 0 {
		t.Errorf("expected no output for non-Write/Edit tool, got %q", out.String())
	}
}

func TestHandlePostToolUseQualityPack_NoFilePath(t *testing.T) {
	input := `{"tool_name":"Write","tool_input":{},"cwd":"/tmp"}`
	var out bytes.Buffer
	if err := HandlePostToolUseQualityPack(strings.NewReader(input), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Len() != 0 {
		t.Errorf("expected no output for missing file_path, got %q", out.String())
	}
}

func TestHandlePostToolUseQualityPack_NonTSFile(t *testing.T) {
	// .go ファイルはスキップ
	input := `{"tool_name":"Write","tool_input":{"file_path":"src/main.go"},"cwd":"/tmp"}`
	var out bytes.Buffer
	if err := HandlePostToolUseQualityPack(strings.NewReader(input), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Len() != 0 {
		t.Errorf("expected no output for non-JS/TS file, got %q", out.String())
	}
}

func TestHandlePostToolUseQualityPack_ExcludedPath(t *testing.T) {
	// node_modules は除外
	input := `{"tool_name":"Write","tool_input":{"file_path":"node_modules/foo/bar.ts"},"cwd":"/tmp"}`
	var out bytes.Buffer
	if err := HandlePostToolUseQualityPack(strings.NewReader(input), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Len() != 0 {
		t.Errorf("expected no output for excluded path, got %q", out.String())
	}
}

func TestHandlePostToolUseQualityPack_DisabledByDefault(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	// config ファイルなし → disabled
	input := `{"tool_name":"Write","tool_input":{"file_path":"src/foo.ts"}}`
	var out bytes.Buffer
	if err := HandlePostToolUseQualityPack(strings.NewReader(input), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Len() != 0 {
		t.Errorf("expected no output when disabled (no config), got %q", out.String())
	}
}

func TestHandlePostToolUseQualityPack_EnabledWarnMode(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	// 設定ファイルを作成（enabled, warn モード）
	config := `quality_pack:
  enabled: true
  mode: warn
  prettier: true
  tsc: true
  console_log: true
`
	if err := os.WriteFile(".claude-code-harness.config.yaml", []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}

	// TS ファイルを作成（console.log を含む）
	tsFile := filepath.Join(tmpDir, "src", "foo.ts")
	if err := os.MkdirAll(filepath.Dir(tsFile), 0o755); err != nil {
		t.Fatal(err)
	}
	tsContent := `const x = 1;
console.log("debug1");
console.log("debug2");
export { x };
`
	if err := os.WriteFile(tsFile, []byte(tsContent), 0o644); err != nil {
		t.Fatal(err)
	}

	input := `{"tool_name":"Write","tool_input":{"file_path":"src/foo.ts"}}`
	var out bytes.Buffer
	if err := HandlePostToolUseQualityPack(strings.NewReader(input), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.Len() == 0 {
		t.Fatal("expected output when quality pack is enabled")
	}

	var result postToolOutput
	if jsonErr := json.Unmarshal(out.Bytes(), &result); jsonErr != nil {
		t.Fatalf("invalid JSON output: %v, raw: %s", jsonErr, out.String())
	}

	ctx := result.HookSpecificOutput.AdditionalContext
	if !strings.Contains(ctx, "Quality Pack") {
		t.Errorf("expected 'Quality Pack' in additionalContext, got %q", ctx)
	}
	// warn モード: prettier は推奨メッセージ
	if !strings.Contains(ctx, "Prettier") {
		t.Errorf("expected 'Prettier' in additionalContext, got %q", ctx)
	}
	// warn モード: tsc は推奨メッセージ
	if !strings.Contains(ctx, "tsc") {
		t.Errorf("expected 'tsc' in additionalContext, got %q", ctx)
	}
	// console.log が2件検出されること
	if !strings.Contains(ctx, "console.log") {
		t.Errorf("expected console.log detection in additionalContext, got %q", ctx)
	}
	if !strings.Contains(ctx, "2") {
		t.Errorf("expected count '2' for console.log in additionalContext, got %q", ctx)
	}
}

func TestHandlePostToolUseQualityPack_NoConsoleLog(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	config := `quality_pack:
  enabled: true
  mode: warn
  prettier: false
  tsc: false
  console_log: true
`
	if err := os.WriteFile(".claude-code-harness.config.yaml", []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}

	// console.log がないファイル
	tsFile := filepath.Join(tmpDir, "src", "clean.ts")
	if err := os.MkdirAll(filepath.Dir(tsFile), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tsFile, []byte("const x = 1;\nexport { x };\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	input := `{"tool_name":"Edit","tool_input":{"file_path":"src/clean.ts"}}`
	var out bytes.Buffer
	if err := HandlePostToolUseQualityPack(strings.NewReader(input), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// console.log がない + prettier/tsc 無効 → フィードバックなし → 出力なし
	if out.Len() != 0 {
		t.Errorf("expected no output when no issues found, got %q", out.String())
	}
}

func TestHandlePostToolUseQualityPack_CWDRelativePath(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	config := `quality_pack:
  enabled: true
  mode: warn
  prettier: true
  tsc: false
  console_log: false
`
	if err := os.WriteFile(".claude-code-harness.config.yaml", []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}

	// CWD プレフィックスの除去確認（絶対パス → 相対パス変換）
	input := `{"tool_name":"Write","tool_input":{"file_path":"` + tmpDir + `/src/bar.ts"},"cwd":"` + tmpDir + `"}`
	var out bytes.Buffer
	if err := HandlePostToolUseQualityPack(strings.NewReader(input), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.Len() == 0 {
		t.Fatal("expected output (prettier warn mode) for TS file")
	}

	var result postToolOutput
	if jsonErr := json.Unmarshal(out.Bytes(), &result); jsonErr != nil {
		t.Fatalf("invalid JSON output: %v, raw: %s", jsonErr, out.String())
	}
	if !strings.Contains(result.HookSpecificOutput.AdditionalContext, "Prettier") {
		t.Errorf("expected Prettier in output, got %q", result.HookSpecificOutput.AdditionalContext)
	}
}

func TestReadQualityPackConfig_Defaults(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	// ファイルなし → デフォルト値
	cfg := readQualityPackConfig(".claude-code-harness.config.yaml")
	if cfg.Enabled {
		t.Error("expected enabled=false by default")
	}
	if cfg.Mode != "warn" {
		t.Errorf("expected mode=warn by default, got %q", cfg.Mode)
	}
	if !cfg.Prettier {
		t.Error("expected prettier=true by default")
	}
	if !cfg.TSC {
		t.Error("expected tsc=true by default")
	}
	if !cfg.ConsoleLog {
		t.Error("expected console_log=true by default")
	}
}

func TestIsJSTSFile(t *testing.T) {
	cases := []struct {
		path     string
		expected bool
	}{
		{"src/foo.ts", true},
		{"src/foo.tsx", true},
		{"src/foo.js", true},
		{"src/foo.jsx", true},
		{"src/foo.go", false},
		{"src/foo.py", false},
		{"src/foo.md", false},
		{"src/FOO.TS", true},  // 大文字小文字
	}

	for _, tc := range cases {
		got := isJSTSFile(tc.path)
		if got != tc.expected {
			t.Errorf("isJSTSFile(%q) = %v, want %v", tc.path, got, tc.expected)
		}
	}
}

func TestIsExcludedPath(t *testing.T) {
	cases := []struct {
		path     string
		excluded bool
	}{
		{".claude/foo.ts", true},
		{"docs/guide.ts", true},
		{"templates/bar.ts", true},
		{"benchmarks/bench.ts", true},
		{"node_modules/lib/foo.ts", true},
		{".git/hooks/pre-commit", true},
		{"src/foo.ts", false},
		{"lib/bar.ts", false},
	}

	for _, tc := range cases {
		got := isExcludedPath(tc.path)
		if got != tc.excluded {
			t.Errorf("isExcludedPath(%q) = %v, want %v", tc.path, got, tc.excluded)
		}
	}
}
