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

// assertSetupOutput is a helper that validates the response from the Setup hook.
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
	// already initialized or a message is returned
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
	// set a temporary directory as the working directory
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

	// verify that .claude/state/ was created
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

	// pre-create the state directory
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

	// create old session files
	sessionsDir := filepath.Join(dir, ".claude", "state", "sessions")
	if err := os.MkdirAll(sessionsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	oldFile := filepath.Join(sessionsDir, "session-old.json")
	if err := os.WriteFile(oldFile, []byte(`{}`), 0o644); err != nil {
		t.Fatal(err)
	}

	// set the file mtime to 8 days ago
	eightDaysAgo := time.Now().AddDate(0, 0, -8)
	if err := os.Chtimes(oldFile, eightDaysAgo, eightDaysAgo); err != nil {
		t.Fatal(err)
	}

	// also create a new session file (should not be deleted)
	newFile := filepath.Join(sessionsDir, "session-new.json")
	if err := os.WriteFile(newFile, []byte(`{}`), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := HandleSetupHookMaintenance(strings.NewReader(""), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertSetupOutput(t, out.String(), "[Setup:maintenance]")

	// verify that the old file was deleted
	if _, err := os.Stat(oldFile); err == nil {
		t.Error("old session file should have been deleted")
	}
	// verify that the new file still exists
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

	// create .tmp files in the state directory
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

	// verify that the .tmp file was deleted
	if _, err := os.Stat(tmpFile); err == nil {
		t.Error(".tmp file should have been deleted")
	}
}

func TestHandleSetupHook_UnknownMode(t *testing.T) {
	var out bytes.Buffer
	// send an unknown mode via JSON payload
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
	if !strings.Contains(ctx, "Unknown mode") {
		t.Errorf("expected 'Unknown mode' in %q", ctx)
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

	// create .tmp files
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

// TestResolveSetupScriptDir_CLAUDE_PLUGIN_ROOT verifies that CLAUDE_PLUGIN_ROOT takes priority.
func TestResolveSetupScriptDir_CLAUDE_PLUGIN_ROOT(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLAUDE_PLUGIN_ROOT", dir)
	// also set HARNESS_SCRIPT_DIR to verify priority ordering
	t.Setenv("HARNESS_SCRIPT_DIR", "/should/not/be/used")

	got := resolveSetupScriptDir()
	want := filepath.Join(dir, "scripts")
	if got != want {
		t.Errorf("resolveSetupScriptDir() = %q, want %q", got, want)
	}
}

// TestResolveSetupScriptDir_HARNESS_SCRIPT_DIR verifies that HARNESS_SCRIPT_DIR is used
// when CLAUDE_PLUGIN_ROOT is not set.
func TestResolveSetupScriptDir_HARNESS_SCRIPT_DIR(t *testing.T) {
	dir := t.TempDir()
	os.Unsetenv("CLAUDE_PLUGIN_ROOT")
	t.Setenv("HARNESS_SCRIPT_DIR", dir)

	got := resolveSetupScriptDir()
	if got != dir {
		t.Errorf("resolveSetupScriptDir() = %q, want %q", got, dir)
	}
}

// TestResolveSetupScriptDir_CWDFallback verifies that CWD/scripts is returned when neither environment variable is set.
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

	// On macOS, os.Getwd() may return the real path via /private/var,
	// while t.TempDir() may return a path via /var (symlink).
	// Verify by checking that the returned path ends with "/scripts".
	if filepath.Base(got) != "scripts" {
		t.Errorf("resolveSetupScriptDir() = %q, want path ending in 'scripts'", got)
	}
	// the return value is os.Getwd() + "/scripts", so its parent should match CWD
	// (compared after symlink resolution)
	cwd, _ := os.Getwd()
	gotDir := filepath.Dir(got)
	gotDirReal, _ := filepath.EvalSymlinks(gotDir)
	cwdReal, _ := filepath.EvalSymlinks(cwd)
	if gotDirReal != cwdReal {
		t.Errorf("resolveSetupScriptDir() parent = %q (real: %q), want CWD = %q (real: %q)",
			gotDir, gotDirReal, cwd, cwdReal)
	}
}

// import time package for use in setup_hook_test.go
var _ = time.Now
