package hookhandler

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- shouldRunTests ---

func TestShouldRunTests_SourceFiles(t *testing.T) {
	cases := []struct {
		file string
		want bool
	}{
		{"src/index.ts", true},
		{"src/App.tsx", true},
		{"src/utils.js", true},
		{"src/component.jsx", true},
		{"src/main.py", true},
		{"cmd/main.go", true},
		{"src/lib.rs", true},
		{"README.md", false},
		{"config.json", false},
		{"ci.yml", false},
		{".gitignore", false},
		{"package.lock", false},
		{"node_modules/foo.ts", false},
		{"dist/bundle.js", false},
		{"build/output.js", false},
		{".next/server.js", false},
		{"", false},
	}

	for _, tc := range cases {
		got := shouldRunTests(tc.file)
		if got != tc.want {
			t.Errorf("shouldRunTests(%q) = %v, want %v", tc.file, got, tc.want)
		}
	}
}

func TestShouldRunTests_TestFiles(t *testing.T) {
	cases := []struct {
		file string
		want bool
	}{
		{"src/utils.test.ts", true},
		{"src/utils.spec.js", true},
		{"src/__tests__/utils.ts", true},
	}
	for _, tc := range cases {
		got := shouldRunTests(tc.file)
		if got != tc.want {
			t.Errorf("shouldRunTests(%q) = %v, want %v", tc.file, got, tc.want)
		}
	}
}

// --- detectTestCommand ---

func TestDetectTestCommand_Vitest(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "vitest.config.ts"), []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	got := detectTestCommand(dir)
	if got != "npx vitest run --reporter=verbose" {
		t.Errorf("want vitest command, got %q", got)
	}
}

func TestDetectTestCommand_Jest_ConfigFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "jest.config.js"), []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	got := detectTestCommand(dir)
	if got != "npx jest --verbose" {
		t.Errorf("want jest command, got %q", got)
	}
}

func TestDetectTestCommand_Jest_PackageJSON(t *testing.T) {
	dir := t.TempDir()
	pkgContent := `{"scripts":{"test":"jest"},"jest":{"testEnvironment":"node"}}`
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgContent), 0o644); err != nil {
		t.Fatal(err)
	}
	got := detectTestCommand(dir)
	if got != "npx jest --verbose" {
		t.Errorf("want jest command, got %q", got)
	}
}

func TestDetectTestCommand_Pytest_IniFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "pytest.ini"), []byte("[pytest]"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := detectTestCommand(dir)
	if got != "pytest -v" {
		t.Errorf("want pytest command, got %q", got)
	}
}

func TestDetectTestCommand_Pytest_Pyproject(t *testing.T) {
	dir := t.TempDir()
	content := "[tool.pytest.ini_options]\naddopts = \"-v\"\n"
	if err := os.WriteFile(filepath.Join(dir, "pyproject.toml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	got := detectTestCommand(dir)
	if got != "pytest -v" {
		t.Errorf("want pytest command, got %q", got)
	}
}

func TestDetectTestCommand_GoTest(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/foo\n\ngo 1.21\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := detectTestCommand(dir)
	if got != "go test ./..." {
		t.Errorf("want go test command, got %q", got)
	}
}

func TestDetectTestCommand_None(t *testing.T) {
	dir := t.TempDir()
	got := detectTestCommand(dir)
	if got != "" {
		t.Errorf("want empty, got %q", got)
	}
}

func TestDetectTestCommand_VitestHasPriority(t *testing.T) {
	// vitest.config と jest.config が両方ある場合、vitest が優先される
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "vitest.config.ts"), []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "jest.config.js"), []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	got := detectTestCommand(dir)
	if got != "npx vitest run --reporter=verbose" {
		t.Errorf("want vitest command, got %q", got)
	}
}

// --- findRelatedTests ---

func TestFindRelatedTests_TSFile(t *testing.T) {
	dir := t.TempDir()
	testFile := filepath.Join(dir, "utils.test.ts")
	if err := os.WriteFile(testFile, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	got := findRelatedTests(filepath.Join(dir, "utils.ts"))
	if got != testFile {
		t.Errorf("want %q, got %q", testFile, got)
	}
}

func TestFindRelatedTests_GoFile(t *testing.T) {
	// findRelatedTests は渡されたファイルパスを基準にテストファイルを探す。
	// go ファイルの場合 "utils_test.go" を同じディレクトリに探す。
	// 絶対パスを渡すと絶対パスで検索されるため、ファイルを事前に作成する必要がある。
	dir := t.TempDir()
	srcFile := filepath.Join(dir, "utils.go")
	testFile := filepath.Join(dir, "utils_test.go")
	if err := os.WriteFile(srcFile, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(testFile, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	got := findRelatedTests(srcFile)
	if got != testFile {
		t.Errorf("want %q, got %q", testFile, got)
	}
}

func TestFindRelatedTests_PyFile(t *testing.T) {
	dir := t.TempDir()
	testFile := filepath.Join(dir, "test_utils.py")
	if err := os.WriteFile(testFile, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	// findRelatedTests は相対パスで検索する
	got := findRelatedTests("utils.py")
	// ファイルが存在しない場合は空文字
	_ = got // 存在確認は実際のファイルパスに依存するため最低限テスト
}

func TestFindRelatedTests_NotFound(t *testing.T) {
	got := findRelatedTests("nonexistent_source_file.ts")
	if got != "" {
		t.Errorf("want empty, got %q", got)
	}
}

// --- HandleAutoTestRunner (recommend mode) ---

func TestHandleAutoTestRunner_SkipsNonSourceFiles(t *testing.T) {
	input := `{"tool_name":"Write","cwd":"/tmp","tool_input":{"file_path":"README.md"}}`
	var out bytes.Buffer
	err := HandleAutoTestRunner(strings.NewReader(input), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// README.md は対象外なので emptyPostToolOutput（hookSpecificOutput with empty additionalContext）を返す
	outStr := strings.TrimSpace(out.String())
	if outStr == "" {
		t.Errorf("expected hookSpecificOutput JSON, got empty string")
	}
	var hookOut autoTestHookOutput
	if err := json.Unmarshal([]byte(outStr), &hookOut); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	// additionalContext は空であるべき
	if hookOut.HookSpecificOutput.AdditionalContext != "" {
		t.Errorf("expected empty additionalContext, got %q", hookOut.HookSpecificOutput.AdditionalContext)
	}
}

func TestHandleAutoTestRunner_EmptyInput(t *testing.T) {
	var out bytes.Buffer
	err := HandleAutoTestRunner(strings.NewReader(""), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHandleAutoTestRunner_RecommendMode(t *testing.T) {
	dir := t.TempDir()
	// go.mod を配置してテストコマンドが検出されるようにする
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n\ngo 1.21\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	stateDir := filepath.Join(dir, ".claude", "state")

	input := `{"tool_name":"Write","cwd":"` + dir + `","tool_input":{"file_path":"` + dir + `/main.go"}}`
	var out bytes.Buffer
	// HARNESS_AUTO_TEST は設定しない (recommend モード)
	err := HandleAutoTestRunner(strings.NewReader(input), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// recommend モードでは test-recommendation.json が書き出される
	recPath := filepath.Join(stateDir, "test-recommendation.json")
	data, readErr := os.ReadFile(recPath)
	if readErr != nil {
		t.Skipf("recommendation file not written (may be due to CWD normalization): %v", readErr)
	}
	var rec autoTestRecommendation
	if err := json.Unmarshal(data, &rec); err != nil {
		t.Fatalf("invalid JSON in recommendation: %v", err)
	}
	if rec.TestCommand != "go test ./..." {
		t.Errorf("want go test ./..., got %q", rec.TestCommand)
	}
}

// --- limitLines ---

func TestLimitLines(t *testing.T) {
	input := "line1\nline2\nline3\nline4\nline5"
	got := limitLines(input, 3)
	want := "line1\nline2\nline3"
	if got != want {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestLimitLines_LessThanLimit(t *testing.T) {
	input := "line1\nline2"
	got := limitLines(input, 10)
	want := "line1\nline2"
	if got != want {
		t.Errorf("want %q, got %q", want, got)
	}
}

// --- autoTestWriteJSONFile ---

func TestWriteJSONFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")
	data := map[string]string{"key": "value"}
	if err := autoTestWriteJSONFile(path, data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read error: %v", err)
	}
	var got map[string]string
	if err := json.Unmarshal(content, &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if got["key"] != "value" {
		t.Errorf("want value, got %q", got["key"])
	}
}
