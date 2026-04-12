package hookhandler

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// assertSetupOutput は Setup フックのレスポンスを検証するヘルパー。
func assertSetupOutput(t *testing.T, output, wantSubstr string) {
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
	if hookOut["hookEventName"] != "Setup" {
		t.Errorf("hookEventName = %q, want Setup", hookOut["hookEventName"])
	}
	ctx, _ := hookOut["additionalContext"].(string)
	if wantSubstr != "" && !strings.Contains(ctx, wantSubstr) {
		t.Errorf("additionalContext = %q, want to contain %q", ctx, wantSubstr)
	}
}

func TestHandleSetupHookInit_EmptyInput(t *testing.T) {
	var out bytes.Buffer
	err := HandleSetupHookInit(strings.NewReader(""), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 既に初期化済みか、何かメッセージが返る
	output := strings.TrimSpace(out.String())
	if output == "" {
		t.Fatal("expected JSON output")
	}
	var resp map[string]interface{}
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, output)
	}
	hookOut, ok := resp["hookSpecificOutput"].(map[string]interface{})
	if !ok {
		t.Fatalf("missing hookSpecificOutput")
	}
	if hookOut["hookEventName"] != "Setup" {
		t.Errorf("hookEventName = %q, want Setup", hookOut["hookEventName"])
	}
}

func TestHandleSetupHookInit_CreatesStateDir(t *testing.T) {
	// 一時ディレクトリをカレントに設定
	dir := t.TempDir()
	origWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origWD)

	var out bytes.Buffer
	if err := HandleSetupHookInit(strings.NewReader(""), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// .claude/state/ が作成されているか確認
	stateDir := filepath.Join(dir, ".claude", "state")
	if info, err := os.Stat(stateDir); err != nil || !info.IsDir() {
		t.Errorf(".claude/state/ was not created at %s", stateDir)
	}
}

func TestHandleSetupHookInit_AlreadyInitialized(t *testing.T) {
	dir := t.TempDir()
	origWD, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origWD)

	// 事前に状態ディレクトリを作成
	if err := os.MkdirAll(filepath.Join(dir, ".claude", "state"), 0o755); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := HandleSetupHookInit(strings.NewReader(""), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertSetupOutput(t, out.String(), "[Setup:init]")
}

func TestHandleSetupHookMaintenance_EmptyInput(t *testing.T) {
	var out bytes.Buffer
	err := HandleSetupHookMaintenance(strings.NewReader(""), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertSetupOutput(t, out.String(), "[Setup:maintenance]")
}

func TestHandleSetupHookMaintenance_CleansOldSessions(t *testing.T) {
	dir := t.TempDir()
	origWD, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origWD)

	// 古いセッションファイルを作成
	sessionsDir := filepath.Join(dir, ".claude", "state", "sessions")
	if err := os.MkdirAll(sessionsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	oldFile := filepath.Join(sessionsDir, "session-old.json")
	if err := os.WriteFile(oldFile, []byte(`{}`), 0o644); err != nil {
		t.Fatal(err)
	}

	// ファイルの mtime を8日前に設定
	eightDaysAgo := time.Now().AddDate(0, 0, -8)
	if err := os.Chtimes(oldFile, eightDaysAgo, eightDaysAgo); err != nil {
		t.Fatal(err)
	}

	// 新しいセッションファイルも作成（削除されないはず）
	newFile := filepath.Join(sessionsDir, "session-new.json")
	if err := os.WriteFile(newFile, []byte(`{}`), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := HandleSetupHookMaintenance(strings.NewReader(""), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertSetupOutput(t, out.String(), "[Setup:maintenance]")

	// 古いファイルが削除されているか確認
	if _, err := os.Stat(oldFile); err == nil {
		t.Error("old session file should have been deleted")
	}
	// 新しいファイルが残っているか確認
	if _, err := os.Stat(newFile); err != nil {
		t.Error("new session file should still exist")
	}
}

func TestHandleSetupHookMaintenance_CleansTmpFiles(t *testing.T) {
	dir := t.TempDir()
	origWD, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origWD)

	// 状態ディレクトリに .tmp ファイルを作成
	stateDir := filepath.Join(dir, ".claude", "state")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatal(err)
	}
	tmpFile := filepath.Join(stateDir, "test.tmp")
	if err := os.WriteFile(tmpFile, []byte("temp"), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := HandleSetupHookMaintenance(strings.NewReader(""), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// .tmp ファイルが削除されているか確認
	if _, err := os.Stat(tmpFile); err == nil {
		t.Error(".tmp file should have been deleted")
	}
}

func TestHandleSetupHook_UnknownMode(t *testing.T) {
	var out bytes.Buffer
	// JSON ペイロードで不明なモードを送信
	payload := `{"mode":"unknown"}`
	if err := handleSetupHook(strings.NewReader(payload), &out, "unknown"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := strings.TrimSpace(out.String())
	var resp map[string]interface{}
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	hookOut := resp["hookSpecificOutput"].(map[string]interface{})
	ctx := hookOut["additionalContext"].(string)
	if !strings.Contains(ctx, "不明なモード") {
		t.Errorf("expected 不明なモード in %q", ctx)
	}
}

func TestIsSimpleMode(t *testing.T) {
	tests := []struct {
		envVal string
		want   bool
	}{
		{"1", true},
		{"true", true},
		{"TRUE", true},
		{"yes", true},
		{"YES", true},
		{"false", false},
		{"0", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.envVal, func(t *testing.T) {
			if tt.envVal != "" {
				t.Setenv("CLAUDE_CODE_SIMPLE", tt.envVal)
			} else {
				os.Unsetenv("CLAUDE_CODE_SIMPLE")
			}
			got := isSimpleMode()
			if got != tt.want {
				t.Errorf("isSimpleMode() = %v, want %v (env=%q)", got, tt.want, tt.envVal)
			}
		})
	}
}

func TestRemoveTmpFiles(t *testing.T) {
	dir := t.TempDir()

	// .tmp ファイルを作成
	tmpFile1 := filepath.Join(dir, "a.tmp")
	tmpFile2 := filepath.Join(dir, "subdir", "b.tmp")
	normalFile := filepath.Join(dir, "normal.json")

	_ = os.MkdirAll(filepath.Join(dir, "subdir"), 0o755)
	_ = os.WriteFile(tmpFile1, []byte("tmp1"), 0o644)
	_ = os.WriteFile(tmpFile2, []byte("tmp2"), 0o644)
	_ = os.WriteFile(normalFile, []byte("{}"), 0o644)

	removeTmpFiles(dir)

	if _, err := os.Stat(tmpFile1); err == nil {
		t.Error("a.tmp should have been deleted")
	}
	if _, err := os.Stat(tmpFile2); err == nil {
		t.Error("subdir/b.tmp should have been deleted")
	}
	if _, err := os.Stat(normalFile); err != nil {
		t.Error("normal.json should still exist")
	}
}

func TestCopyFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")

	content := []byte("hello world")
	if err := os.WriteFile(src, content, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := copyFile(src, dst); err != nil {
		t.Fatalf("copyFile failed: %v", err)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(content) {
		t.Errorf("copied content = %q, want %q", got, content)
	}
}

// TestResolveSetupScriptDir_CLAUDE_PLUGIN_ROOT は CLAUDE_PLUGIN_ROOT が優先されることを確認する。
func TestResolveSetupScriptDir_CLAUDE_PLUGIN_ROOT(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLAUDE_PLUGIN_ROOT", dir)
	// HARNESS_SCRIPT_DIR も設定して、優先順位の確認
	t.Setenv("HARNESS_SCRIPT_DIR", "/should/not/be/used")

	got := resolveSetupScriptDir()
	want := filepath.Join(dir, "scripts")
	if got != want {
		t.Errorf("resolveSetupScriptDir() = %q, want %q", got, want)
	}
}

// TestResolveSetupScriptDir_HARNESS_SCRIPT_DIR は CLAUDE_PLUGIN_ROOT がない場合に
// HARNESS_SCRIPT_DIR が使われることを確認する。
func TestResolveSetupScriptDir_HARNESS_SCRIPT_DIR(t *testing.T) {
	dir := t.TempDir()
	os.Unsetenv("CLAUDE_PLUGIN_ROOT")
	t.Setenv("HARNESS_SCRIPT_DIR", dir)

	got := resolveSetupScriptDir()
	if got != dir {
		t.Errorf("resolveSetupScriptDir() = %q, want %q", got, dir)
	}
}

// TestResolveSetupScriptDir_CWDFallback は両環境変数がない場合に CWD/scripts が返ることを確認する。
func TestResolveSetupScriptDir_CWDFallback(t *testing.T) {
	dir := t.TempDir()
	origWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origWD) //nolint:errcheck

	os.Unsetenv("CLAUDE_PLUGIN_ROOT")
	os.Unsetenv("HARNESS_SCRIPT_DIR")

	got := resolveSetupScriptDir()

	// macOS では os.Getwd() が /private/var 経由の実パスを返すが、
	// t.TempDir() は /var 経由のパスを返すことがある (symlink)。
	// パスの末尾が "/scripts" になっているかを確認することで対応する。
	if filepath.Base(got) != "scripts" {
		t.Errorf("resolveSetupScriptDir() = %q, want path ending in 'scripts'", got)
	}
	// 戻り値は os.Getwd() + "/scripts" なので、ディレクトリ部分は CWD と一致する
	// (symlink 解決後の比較)
	cwd, _ := os.Getwd()
	gotDir := filepath.Dir(got)
	gotDirReal, _ := filepath.EvalSymlinks(gotDir)
	cwdReal, _ := filepath.EvalSymlinks(cwd)
	if gotDirReal != cwdReal {
		t.Errorf("resolveSetupScriptDir() parent = %q (real: %q), want CWD = %q (real: %q)",
			gotDir, gotDirReal, cwd, cwdReal)
	}
}

// time パッケージを setup_hook_test.go でも使用するため
var _ = time.Now
