package hookhandler

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// autoTestRunnerInput は PostToolUse フックから渡される stdin JSON。
type autoTestRunnerInput struct {
	ToolName  string `json:"tool_name"`
	CWD       string `json:"cwd"`
	ToolInput struct {
		FilePath string `json:"file_path"`
	} `json:"tool_input"`
	ToolResponse struct {
		FilePath string `json:"filePath"`
	} `json:"tool_response"`
}

// autoTestResult は .claude/state/test-result.json に書き出す構造体。
type autoTestResult struct {
	Timestamp   string `json:"timestamp"`
	ChangedFile string `json:"changed_file"`
	Command     string `json:"command"`
	Status      string `json:"status"`
	ExitCode    int    `json:"exit_code"`
	Output      string `json:"output"`
}

// autoTestRecommendation は .claude/state/test-recommendation.json に書き出す構造体。
type autoTestRecommendation struct {
	Timestamp    string `json:"timestamp"`
	ChangedFile  string `json:"changed_file"`
	TestCommand  string `json:"test_command"`
	RelatedTest  string `json:"related_test"`
	Recommendation string `json:"recommendation"`
}

// autoTestHookOutput は additionalContext 付きの hookSpecificOutput。
type autoTestHookOutput struct {
	HookSpecificOutput struct {
		HookEventName     string `json:"hookEventName"`
		AdditionalContext string `json:"additionalContext"`
	} `json:"hookSpecificOutput"`
}

// sourceFileExtensions はテスト実行が必要なファイル拡張子。
var sourceFileExtensions = []string{
	".ts", ".tsx", ".js", ".jsx", ".py", ".go", ".rs",
}

// excludedDirs はテスト対象から除外するディレクトリプレフィックス。
var excludedDirs = []string{
	"node_modules/",
	"dist/",
	"build/",
	".next/",
}

// excludedExtensions はテスト対象から除外するファイル拡張子。
var excludedExtensions = []string{
	".md", ".json", ".yml", ".yaml", ".lock",
}

// HandleAutoTestRunner は auto-test-runner.sh の Go 移植。
//
// PostToolUse Write/Edit イベントでソースファイル変更を検出し、
// テストフレームワークを自動検出してテストを実行する。
//
// 動作モード:
//   - HARNESS_AUTO_TEST=run → テストを実際に実行し、additionalContext で通知
//   - デフォルト (recommend) → テスト推奨を .claude/state/ に記録
func HandleAutoTestRunner(in io.Reader, out io.Writer) error {
	data, err := io.ReadAll(in)
	if err != nil || len(bytes.TrimSpace(data)) == 0 {
		return emptyPostToolOutput(out)
	}

	var input autoTestRunnerInput
	if err := json.Unmarshal(data, &input); err != nil {
		return emptyPostToolOutput(out)
	}

	// 変更ファイルを取得
	changedFile := input.ToolInput.FilePath
	if changedFile == "" {
		changedFile = input.ToolResponse.FilePath
	}
	if changedFile == "" {
		return emptyPostToolOutput(out)
	}

	// プロジェクト相対パスへ正規化
	changedFile = normalizePathSeparators(changedFile)
	if input.CWD != "" {
		cwd := normalizePathSeparators(input.CWD)
		changedFile = makeRelativePath(changedFile, cwd)
	}

	// プロジェクトルートを決定（CWD またはカレントディレクトリ）
	projectRoot := input.CWD
	if projectRoot == "" {
		projectRoot, _ = os.Getwd()
	}

	// テスト実行が必要か判定
	if !shouldRunTests(changedFile) {
		return emptyPostToolOutput(out)
	}

	// テストコマンドを検出
	testCmd := detectTestCommand(projectRoot)
	if testCmd == "" {
		return emptyPostToolOutput(out)
	}

	// 関連テストファイルを検出
	relatedTest := findRelatedTests(changedFile)

	stateDir := filepath.Join(projectRoot, ".claude", "state")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		return emptyPostToolOutput(out)
	}

	// HARNESS_AUTO_TEST=run の場合は実際にテストを実行
	if os.Getenv("HARNESS_AUTO_TEST") == "run" {
		return runTestsAndReport(out, projectRoot, stateDir, changedFile, testCmd, relatedTest)
	}

	// デフォルト: recommend モード
	return writeTestRecommendation(out, stateDir, changedFile, testCmd, relatedTest)
}

// shouldRunTests はファイルがテスト実行が必要かどうかを判定する。
func shouldRunTests(file string) bool {
	if file == "" {
		return false
	}

	// 除外ディレクトリチェック
	for _, dir := range excludedDirs {
		if strings.HasPrefix(file, dir) {
			return false
		}
	}

	// 除外拡張子チェック
	for _, ext := range excludedExtensions {
		if strings.HasSuffix(file, ext) {
			return false
		}
	}

	// .gitignore
	if file == ".gitignore" {
		return false
	}

	// テストファイル自体の変更
	if strings.Contains(file, ".test.") || strings.Contains(file, ".spec.") || strings.Contains(file, "__tests__") {
		return true
	}

	// ソースコードファイルの変更
	for _, ext := range sourceFileExtensions {
		if strings.HasSuffix(file, ext) {
			return true
		}
	}

	return false
}

// detectTestCommand はプロジェクトルートからテストコマンドを自動検出する。
//
// 検出優先順:
//  1. vitest.config.* → npx vitest run --reporter=verbose
//  2. jest.config.* or package.json の jest → npx jest --verbose
//  3. pytest.ini or pyproject.toml の [tool.pytest] → pytest -v
//  4. *_test.go → go test ./...
func detectTestCommand(projectRoot string) string {
	// vitest
	vitestConfigs := []string{
		"vitest.config.ts", "vitest.config.js", "vitest.config.mts", "vitest.config.mjs",
	}
	for _, cfg := range vitestConfigs {
		if autoTestFileExists(filepath.Join(projectRoot, cfg)) {
			return "npx vitest run --reporter=verbose"
		}
	}

	// jest
	jestConfigs := []string{
		"jest.config.ts", "jest.config.js", "jest.config.mjs", "jest.config.cjs",
	}
	for _, cfg := range jestConfigs {
		if autoTestFileExists(filepath.Join(projectRoot, cfg)) {
			return "npx jest --verbose"
		}
	}
	// package.json の jest フィールド
	pkgPath := filepath.Join(projectRoot, "package.json")
	if autoTestFileExists(pkgPath) {
		content, err := os.ReadFile(pkgPath)
		if err == nil && bytes.Contains(content, []byte(`"jest"`)) {
			return "npx jest --verbose"
		}
	}

	// pytest.ini
	if autoTestFileExists(filepath.Join(projectRoot, "pytest.ini")) {
		return "pytest -v"
	}
	// pyproject.toml の [tool.pytest]
	pyprojectPath := filepath.Join(projectRoot, "pyproject.toml")
	if autoTestFileExists(pyprojectPath) {
		content, err := os.ReadFile(pyprojectPath)
		if err == nil && bytes.Contains(content, []byte("[tool.pytest")) {
			return "pytest -v"
		}
	}

	// go test: go.mod が存在するか確認
	if autoTestFileExists(filepath.Join(projectRoot, "go.mod")) {
		return "go test ./..."
	}

	return ""
}

// findRelatedTests は変更ファイルに対応するテストファイルを探す。
func findRelatedTests(file string) string {
	ext := filepath.Ext(file)
	basename := strings.TrimSuffix(file, ext)
	dirname := filepath.Dir(file)
	baseName := filepath.Base(basename)

	patterns := []string{
		basename + ".test.ts",
		basename + ".test.tsx",
		basename + ".test.js",
		basename + ".test.jsx",
		basename + ".spec.ts",
		basename + ".spec.tsx",
		basename + ".spec.js",
		basename + ".spec.jsx",
		filepath.Join(dirname, "__tests__", baseName+".test.ts"),
		filepath.Join(dirname, "__tests__", baseName+".test.tsx"),
		filepath.Join(dirname, "test_"+baseName+".py"),
		basename + "_test.go",
	}

	for _, pattern := range patterns {
		if autoTestFileExists(pattern) {
			return pattern
		}
	}
	return ""
}

// runTestsAndReport はテストを実行して結果を記録し、additionalContext で通知する。
func runTestsAndReport(out io.Writer, projectRoot, stateDir, changedFile, testCmd, relatedTest string) error {
	// 実行コマンドを決定
	execCmd := testCmd
	if relatedTest != "" {
		execCmd = testCmd + " -- " + relatedTest
	}

	ts := time.Now().UTC().Format(time.RFC3339)

	// タイムアウト付きでテスト実行（最大 60 秒）
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bash", "-c", execCmd) //nolint:gosec
	cmd.Dir = projectRoot

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	runErr := cmd.Run()
	exitCode := 0
	status := "passed"

	if ctx.Err() == context.DeadlineExceeded {
		exitCode = 124
		status = "timeout"
	} else if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
		status = "failed"
	}

	// 出力を最大 200 行に制限
	output := limitLines(buf.String(), 200)

	// 結果を JSON で書き出す
	resultPath := filepath.Join(stateDir, "test-result.json")
	result := autoTestResult{
		Timestamp:   ts,
		ChangedFile: changedFile,
		Command:     execCmd,
		Status:      status,
		ExitCode:    exitCode,
		Output:      output,
	}
	if err := autoTestWriteJSONFile(resultPath, result); err != nil {
		fmt.Fprintf(os.Stderr, "[auto-test-runner] write result: %v\n", err)
	}

	// additionalContext を構築
	var contextMsg string
	outputSnippet := limitLines(output, 30)

	switch status {
	case "passed":
		contextMsg = fmt.Sprintf("[Auto Test Runner] Tests passed / テスト成功\nCommand: %s\nFile: %s\nStatus: PASSED (exit=0)",
			testCmd, changedFile)
	case "timeout":
		contextMsg = fmt.Sprintf("[Auto Test Runner] Tests timed out / テストがタイムアウトしました (60s)\nCommand: %s\nFile: %s\nStatus: TIMEOUT\n\nOutput:\n%s",
			testCmd, changedFile, outputSnippet)
	default:
		contextMsg = fmt.Sprintf("[Auto Test Runner] Tests failed / テスト失敗\nCommand: %s\nFile: %s\nStatus: FAILED (exit=%d)\n\nOutput:\n%s\n\nFix the implementation to make the tests pass. / テストが通るように実装を修正してください。",
			testCmd, changedFile, exitCode, outputSnippet)
	}

	var hookOut autoTestHookOutput
	hookOut.HookSpecificOutput.HookEventName = "PostToolUse"
	hookOut.HookSpecificOutput.AdditionalContext = contextMsg

	return json.NewEncoder(out).Encode(hookOut)
}

// writeTestRecommendation はテスト推奨を記録する（recommend モード）。
func writeTestRecommendation(out io.Writer, stateDir, changedFile, testCmd, relatedTest string) error {
	ts := time.Now().UTC().Format(time.RFC3339)
	recPath := filepath.Join(stateDir, "test-recommendation.json")
	rec := autoTestRecommendation{
		Timestamp:      ts,
		ChangedFile:    changedFile,
		TestCommand:    testCmd,
		RelatedTest:    relatedTest,
		Recommendation: "テストの実行を推奨します",
	}
	if err := autoTestWriteJSONFile(recPath, rec); err != nil {
		fmt.Fprintf(os.Stderr, "[auto-test-runner] write recommendation: %v\n", err)
	}

	// recommend モードでは空の PostToolUse 出力を返す
	return emptyPostToolOutput(out)
}

// autoTestFileExists はファイルが存在するかどうかを確認する。
func autoTestFileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// limitLines はテキストを最大 n 行に制限する。
func limitLines(text string, n int) string {
	scanner := bufio.NewScanner(strings.NewReader(text))
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) >= n {
			break
		}
	}
	return strings.Join(lines, "\n")
}

// autoTestWriteJSONFile は v を JSON エンコードしてファイルに書き出す。
func autoTestWriteJSONFile(path string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	return nil
}
