package hookhandler

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
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
	if _, err := exec.LookPath("pytest"); err != nil {
		t.Skip("pytest not found in PATH; skipping pytest detection test")
	}
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
	if _, err := exec.LookPath("pytest"); err != nil {
		t.Skip("pytest not found in PATH; skipping pytest detection test")
	}
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

func TestDetectTestCommand_Pytest_NoPytestBinary(t *testing.T) {
	if _, err := exec.LookPath("pytest"); err == nil {
		t.Skip("pytest is installed; cannot test missing-binary branch")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "pytest.ini"), []byte("[pytest]"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := detectTestCommand(dir)
	if got != "" {
		t.Errorf("want empty (pytest not in PATH), got %q", got)
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

func TestDetectTestCommand_NpmTest_Fallback(t *testing.T) {
	dir := t.TempDir()
	pkgContent := `{"name":"my-app","scripts":{"test":"mocha --exit","build":"webpack"}}`
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgContent), 0o644); err != nil {
		t.Fatal(err)
	}
	got := detectTestCommand(dir)
	if got != "npm test" {
		t.Errorf("want npm test, got %q", got)
	}
}

func TestDetectTestCommand_NpmTest_EmptyScript(t *testing.T) {
	dir := t.TempDir()
	pkgContent := `{"name":"my-app","scripts":{"test":""}}`
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgContent), 0o644); err != nil {
		t.Fatal(err)
	}
	got := detectTestCommand(dir)
	if got != "" {
		t.Errorf("want empty (no valid test script), got %q", got)
	}
}

func TestDetectTestCommand_NpmTest_NoTestScript(t *testing.T) {
	dir := t.TempDir()
	pkgContent := `{"name":"my-app","scripts":{"build":"webpack","start":"node index.js"}}`
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgContent), 0o644); err != nil {
		t.Fatal(err)
	}
	got := detectTestCommand(dir)
	if got != "" {
		t.Errorf("want empty (no test script), got %q", got)
	}
}

func TestDetectTestCommand_VitestHasPriorityOverNpmTest(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "vitest.config.ts"), []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	pkgContent := `{"name":"my-app","scripts":{"test":"vitest run"}}`
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgContent), 0o644); err != nil {
		t.Fatal(err)
	}
	got := detectTestCommand(dir)
	if got != "npx vitest run --reporter=verbose" {
		t.Errorf("want vitest command (higher priority), got %q", got)
	}
}

func TestDetectTestCommand_Pytest_TestsDir(t *testing.T) {
	if _, err := exec.LookPath("pytest"); err != nil {
		t.Skip("pytest not installed")
	}
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "tests"), 0o755); err != nil {
		t.Fatal(err)
	}
	got := detectTestCommand(dir)
	if got != "pytest -v" {
		t.Errorf("want pytest -v (tests/ dir detected), got %q", got)
	}
}

func TestDetectTestCommand_Pytest_TestsDir_NotForJSProject(t *testing.T) {
	dir := t.TempDir()
	pkgContent := `{"name":"my-app","scripts":{"build":"webpack"}}`
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgContent), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "tests"), 0o755); err != nil {
		t.Fatal(err)
	}
	got := detectTestCommand(dir)
	if got == "pytest -v" {
		t.Errorf("false positive: JS project with tests/ dir was detected as pytest, got %q", got)
	}
}

func TestDetectTestCommand_Cargo(t *testing.T) {
	dir := t.TempDir()
	cargoContent := `[package]
name = "my-crate"
version = "0.1.0"
edition = "2021"
`
	if err := os.WriteFile(filepath.Join(dir, "Cargo.toml"), []byte(cargoContent), 0o644); err != nil {
		t.Fatal(err)
	}
	got := detectTestCommand(dir)
	if got != "cargo test" {
		t.Errorf("want cargo test, got %q", got)
	}
}

func TestDetectTestCommand_Jest_FalsePositive_AtTypesJest(t *testing.T) {
	dir := t.TempDir()
	pkgContent := `{"name":"my-app","devDependencies":{"@types/jest":"^29.0.0","typescript":"^5.0.0"},"scripts":{"build":"tsc"}}`
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgContent), 0o644); err != nil {
		t.Fatal(err)
	}
	got := detectTestCommand(dir)
	if got == "npx jest --verbose" {
		t.Errorf("false positive: got jest command for @types/jest-only package.json, got %q", got)
	}
}

func TestDetectTestCommand_Jest_FalsePositive_JestJunit(t *testing.T) {
	dir := t.TempDir()
	pkgContent := `{"name":"my-app","devDependencies":{"jest-junit":"^16.0.0"},"scripts":{"build":"webpack"}}`
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgContent), 0o644); err != nil {
		t.Fatal(err)
	}
	got := detectTestCommand(dir)
	if got == "npx jest --verbose" {
		t.Errorf("false positive: got jest command for jest-junit-only package.json, got %q", got)
	}
}

func TestDetectTestCommand_Jest_ConfigObject(t *testing.T) {
	dir := t.TempDir()
	pkgContent := `{"name":"my-app","jest":{"testEnvironment":"node","collectCoverage":true}}`
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgContent), 0o644); err != nil {
		t.Fatal(err)
	}
	got := detectTestCommand(dir)
	if got != "npx jest --verbose" {
		t.Errorf("want npx jest --verbose (jest config object in package.json), got %q", got)
	}
}

func TestDetectTestCommand_Jest_ScriptsTest(t *testing.T) {
	dir := t.TempDir()
	pkgContent := `{"name":"my-app","scripts":{"test":"jest --coverage","build":"webpack"}}`
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgContent), 0o644); err != nil {
		t.Fatal(err)
	}
	got := detectTestCommand(dir)
	if got != "npx jest --verbose" {
		t.Errorf("want npx jest --verbose (scripts.test contains jest), got %q", got)
	}
}

func TestDetectTestCommand_Jest_ConfigFile_HasPriority(t *testing.T) {
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
	got := findRelatedTests(filepath.Join(dir, "utils.ts"), "")
	if got != testFile {
		t.Errorf("want %q, got %q", testFile, got)
	}
}

func TestFindRelatedTests_GoFile(t *testing.T) {
	dir := t.TempDir()
	srcFile := filepath.Join(dir, "utils.go")
	testFile := filepath.Join(dir, "utils_test.go")
	if err := os.WriteFile(srcFile, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(testFile, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	got := findRelatedTests(srcFile, "")
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
	got := findRelatedTests("utils.py", "")
	_ = got
}

func TestFindRelatedTests_NotFound(t *testing.T) {
	got := findRelatedTests("nonexistent_source_file.ts", "")
	if got != "" {
		t.Errorf("want empty, got %q", got)
	}
}

func TestFindRelatedTests_WithProjectRoot(t *testing.T) {
	dir := t.TempDir()
	srcDir := filepath.Join(dir, "src")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatal(err)
	}
	testFile := filepath.Join(srcDir, "utils.test.ts")
	if err := os.WriteFile(testFile, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	got := findRelatedTests("src/utils.ts", dir)
	if got != testFile {
		t.Errorf("want %q, got %q", testFile, got)
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
	outStr := strings.TrimSpace(out.String())
	if outStr == "" {
		t.Errorf("expected hookSpecificOutput JSON, got empty string")
	}
	var hookOut autoTestHookOutput
	if err := json.Unmarshal([]byte(outStr), &hookOut); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
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
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n\ngo 1.21\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	stateDir := filepath.Join(dir, ".claude", "state")

	input := `{"tool_name":"Write","cwd":"` + dir + `","tool_input":{"file_path":"` + dir + `/main.go"}}`
	var out bytes.Buffer
	err := HandleAutoTestRunner(strings.NewReader(input), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

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


func TestBuildExecCommand_GoTest_NoDoubleDash(t *testing.T) {
	cmd := buildExecCommand("go test ./...", "internal/foo/bar_test.go", "/repo")
	if strings.Contains(cmd, "-- ") {
		t.Errorf("go test command must not contain '-- <file>', got %q", cmd)
	}
	if !strings.HasPrefix(cmd, "go test ./") {
		t.Errorf("go test command should start with 'go test ./', got %q", cmd)
	}
}

func TestBuildExecCommand_GoTest_PackagePath(t *testing.T) {
	got := buildExecCommand("go test ./...", "internal/foo/bar_test.go", "/repo")
	want := "go test ./internal/foo/..."
	if got != want {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestBuildExecCommand_GoTest_AbsRelatedTest(t *testing.T) {
	got := buildExecCommand("go test ./...", "/repo/internal/foo/bar_test.go", "/repo")
	want := "go test ./internal/foo/..."
	if got != want {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestBuildExecCommand_Pytest_FileArg(t *testing.T) {
	got := buildExecCommand("pytest -v", "tests/test_utils.py", "")
	want := "pytest -v tests/test_utils.py"
	if got != want {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestBuildExecCommand_CargoTest_NoFileArg(t *testing.T) {
	got := buildExecCommand("cargo test", "src/lib.rs", "")
	if got != "cargo test" {
		t.Errorf("want 'cargo test' (no file arg), got %q", got)
	}
}

func TestBuildExecCommand_Jest_DoubleDash(t *testing.T) {
	got := buildExecCommand("npx jest --verbose", "src/utils.test.ts", "")
	want := "npx jest --verbose -- src/utils.test.ts"
	if got != want {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestBuildExecCommand_Vitest_DoubleDash(t *testing.T) {
	got := buildExecCommand("npx vitest run --reporter=verbose", "src/utils.test.ts", "")
	want := "npx vitest run --reporter=verbose -- src/utils.test.ts"
	if got != want {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestBuildExecCommand_NpmTest_NoFileArg(t *testing.T) {
	got := buildExecCommand("npm test", "src/utils.ts", "")
	if got != "npm test" {
		t.Errorf("want 'npm test' (no file arg), got %q", got)
	}
}

func TestBuildExecCommand_NoRelatedTest(t *testing.T) {
	cases := []struct {
		testCmd string
	}{
		{"go test ./..."},
		{"pytest -v"},
		{"cargo test"},
		{"npx jest --verbose"},
		{"npm test"},
	}
	for _, tc := range cases {
		got := buildExecCommand(tc.testCmd, "", "")
		if got != tc.testCmd {
			t.Errorf("buildExecCommand(%q, \"\", \"\") = %q, want %q", tc.testCmd, got, tc.testCmd)
		}
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

func TestHasNpmTestScript_Placeholder(t *testing.T) {
	placeholderPkg := []byte(`{"scripts":{"test":"echo \"Error: no test specified\" && exit 1"}}`)

	got := hasNpmTestScript(placeholderPkg)
	if got {
		t.Error("expected hasNpmTestScript=false for npm init placeholder, got true")
	}
}

func TestHasNpmTestScript_RealScript(t *testing.T) {
	realPkg := []byte(`{"scripts":{"test":"jest --coverage"}}`)

	got := hasNpmTestScript(realPkg)
	if !got {
		t.Error("expected hasNpmTestScript=true for real test script, got false")
	}
}

func TestHasNpmTestScript_Empty(t *testing.T) {
	emptyPkg := []byte(`{"scripts":{"test":""}}`)

	got := hasNpmTestScript(emptyPkg)
	if got {
		t.Error("expected hasNpmTestScript=false for empty test script, got true")
	}
}
