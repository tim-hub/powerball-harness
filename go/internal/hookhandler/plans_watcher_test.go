package hookhandler

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestHandlePlansWatcher_NoInput(t *testing.T) {
	var out bytes.Buffer
	err := HandlePlansWatcher(strings.NewReader(""), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result postToolOutput
	if jsonErr := json.Unmarshal(out.Bytes(), &result); jsonErr != nil {
		t.Fatalf("invalid JSON output: %v, raw: %s", jsonErr, out.String())
	}
	if result.HookSpecificOutput.HookEventName != "PostToolUse" {
		t.Errorf("expected hookEventName=PostToolUse, got %q", result.HookSpecificOutput.HookEventName)
	}
}

func TestHandlePlansWatcher_NoFilePath(t *testing.T) {
	input := `{"tool_name":"Edit","tool_input":{}}`
	var out bytes.Buffer
	if err := HandlePlansWatcher(strings.NewReader(input), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result postToolOutput
	if jsonErr := json.Unmarshal(out.Bytes(), &result); jsonErr != nil {
		t.Fatalf("invalid JSON output: %v", jsonErr)
	}
	if result.HookSpecificOutput.AdditionalContext != "" {
		t.Errorf("expected empty context for no file_path, got %q", result.HookSpecificOutput.AdditionalContext)
	}
}

func TestHandlePlansWatcher_NonPlansFile(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	// Plans.md を作成しておく
	if err := os.WriteFile("Plans.md", []byte("# Plans\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Plans.md 以外のファイルを編集した場合はスキップ
	input := `{"tool_name":"Write","tool_input":{"file_path":"src/main.go"}}`
	var out bytes.Buffer
	if err := HandlePlansWatcher(strings.NewReader(input), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result postToolOutput
	if jsonErr := json.Unmarshal(out.Bytes(), &result); jsonErr != nil {
		t.Fatalf("invalid JSON output: %v", jsonErr)
	}
	if result.HookSpecificOutput.AdditionalContext != "" {
		t.Errorf("expected empty context for non-Plans.md file, got %q", result.HookSpecificOutput.AdditionalContext)
	}
}

func TestHandlePlansWatcher_NoPlansFile(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	// Plans.md が存在しない場合はスキップ
	input := `{"tool_name":"Write","tool_input":{"file_path":"Plans.md"}}`
	var out bytes.Buffer
	if err := HandlePlansWatcher(strings.NewReader(input), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result postToolOutput
	if jsonErr := json.Unmarshal(out.Bytes(), &result); jsonErr != nil {
		t.Fatalf("invalid JSON output: %v", jsonErr)
	}
	if result.HookSpecificOutput.AdditionalContext != "" {
		t.Errorf("expected empty context when Plans.md not found, got %q", result.HookSpecificOutput.AdditionalContext)
	}
}

func TestHandlePlansWatcher_NewTaskDetected(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	// Plans.md に pm:依頼中 を含む内容を作成
	plansContent := "| Task 1 | 実装A | DoD | - | pm:依頼中 |\n"
	if err := os.WriteFile("Plans.md", []byte(plansContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// 前回の状態（pm_pending=0）を保存
	if err := os.MkdirAll(".claude/state", 0o755); err != nil {
		t.Fatal(err)
	}
	prevState := `{"timestamp":"2026-01-01T00:00:00Z","pm_pending":0,"cc_todo":0,"cc_wip":0,"cc_done":0,"pm_confirmed":0}`
	if err := os.WriteFile(plansStateFile, []byte(prevState), 0o644); err != nil {
		t.Fatal(err)
	}

	input := `{"tool_name":"Edit","tool_input":{"file_path":"Plans.md"}}`
	var out bytes.Buffer
	if err := HandlePlansWatcher(strings.NewReader(input), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result postToolOutput
	if jsonErr := json.Unmarshal(out.Bytes(), &result); jsonErr != nil {
		t.Fatalf("invalid JSON output: %v, raw: %s", jsonErr, out.String())
	}

	// 新規タスクが検出されること
	if !strings.Contains(result.HookSpecificOutput.AdditionalContext, "新規タスク") {
		t.Errorf("expected '新規タスク' in additionalContext, got %q",
			result.HookSpecificOutput.AdditionalContext)
	}

	// pm-notification.md が作成されること
	data, err := os.ReadFile(pmNotificationFile)
	if err != nil {
		t.Fatalf("pm-notification.md not created: %v", err)
	}
	if !strings.Contains(string(data), "新規タスク") {
		t.Errorf("pm-notification.md should contain '新規タスク', got: %s", string(data))
	}
}

func TestHandlePlansWatcher_CompletedTaskDetected(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	// Plans.md に cc:完了 を含む内容を作成
	plansContent := "| Task 1 | 実装A | DoD | - | cc:完了 |\n"
	if err := os.WriteFile("Plans.md", []byte(plansContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// 前回の状態（cc_done=0）を保存
	if err := os.MkdirAll(".claude/state", 0o755); err != nil {
		t.Fatal(err)
	}
	prevState := `{"timestamp":"2026-01-01T00:00:00Z","pm_pending":0,"cc_todo":0,"cc_wip":0,"cc_done":0,"pm_confirmed":0}`
	if err := os.WriteFile(plansStateFile, []byte(prevState), 0o644); err != nil {
		t.Fatal(err)
	}

	input := `{"tool_name":"Edit","tool_input":{"file_path":"Plans.md"}}`
	var out bytes.Buffer
	if err := HandlePlansWatcher(strings.NewReader(input), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result postToolOutput
	if jsonErr := json.Unmarshal(out.Bytes(), &result); jsonErr != nil {
		t.Fatalf("invalid JSON output: %v", jsonErr)
	}

	if !strings.Contains(result.HookSpecificOutput.AdditionalContext, "タスク完了") {
		t.Errorf("expected 'タスク完了' in additionalContext, got %q",
			result.HookSpecificOutput.AdditionalContext)
	}
}

func TestHandlePlansWatcher_NoChange(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	// 変化なし（cc:TODO 1件のまま）
	plansContent := "| Task 1 | 実装A | DoD | - | cc:TODO |\n"
	if err := os.WriteFile("Plans.md", []byte(plansContent), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(".claude/state", 0o755); err != nil {
		t.Fatal(err)
	}
	prevState := `{"timestamp":"2026-01-01T00:00:00Z","pm_pending":0,"cc_todo":1,"cc_wip":0,"cc_done":0,"pm_confirmed":0}`
	if err := os.WriteFile(plansStateFile, []byte(prevState), 0o644); err != nil {
		t.Fatal(err)
	}

	input := `{"tool_name":"Edit","tool_input":{"file_path":"Plans.md"}}`
	var out bytes.Buffer
	if err := HandlePlansWatcher(strings.NewReader(input), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result postToolOutput
	if jsonErr := json.Unmarshal(out.Bytes(), &result); jsonErr != nil {
		t.Fatalf("invalid JSON output: %v", jsonErr)
	}

	// 変化がない場合は通知なし
	if result.HookSpecificOutput.AdditionalContext != "" {
		t.Errorf("expected empty context for no change, got %q", result.HookSpecificOutput.AdditionalContext)
	}
}

func TestHandlePlansWatcher_StatusSummary(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	// 複数マーカーを含む Plans.md
	plansContent := "| Task 1 | A | DoD | - | cc:TODO |\n" +
		"| Task 2 | B | DoD | - | cc:WIP |\n" +
		"| Task 3 | C | DoD | - | cc:完了 |\n" +
		"| Task 4 | D | DoD | - | pm:依頼中 |\n"
	if err := os.WriteFile("Plans.md", []byte(plansContent), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(".claude/state", 0o755); err != nil {
		t.Fatal(err)
	}
	prevState := `{"timestamp":"2026-01-01T00:00:00Z","pm_pending":0,"cc_todo":0,"cc_wip":0,"cc_done":0,"pm_confirmed":0}`
	if err := os.WriteFile(plansStateFile, []byte(prevState), 0o644); err != nil {
		t.Fatal(err)
	}

	input := `{"tool_name":"Edit","tool_input":{"file_path":"Plans.md"}}`
	var out bytes.Buffer
	if err := HandlePlansWatcher(strings.NewReader(input), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result postToolOutput
	if jsonErr := json.Unmarshal(out.Bytes(), &result); jsonErr != nil {
		t.Fatalf("invalid JSON output: %v", jsonErr)
	}

	ctx := result.HookSpecificOutput.AdditionalContext
	// サマリにステータスカウントが含まれること
	if !strings.Contains(ctx, "cc:TODO") {
		t.Errorf("expected 'cc:TODO' in summary, got %q", ctx)
	}
	if !strings.Contains(ctx, "cc:WIP") {
		t.Errorf("expected 'cc:WIP' in summary, got %q", ctx)
	}
	if !strings.Contains(ctx, "cc:完了") {
		t.Errorf("expected 'cc:完了' in summary, got %q", ctx)
	}
}

func TestIsPlansFile(t *testing.T) {
	// isPlansFile は filepath.Clean による厳密一致のみ。
	// projectRoot を加味した比較は isPlansFileWithRoot で行う。
	cases := []struct {
		changed  string
		plans    string
		expected bool
	}{
		// 完全一致（相対パス）
		{"Plans.md", "Plans.md", true},
		// 完全一致（絶対パス）
		{"/home/user/project/Plans.md", "/home/user/project/Plans.md", true},
		// フルパスが異なれば不一致（別ディレクトリの同名ファイルは false）
		{"docs/Plans.md", "Plans.md", false},
		{"/home/user/project/Plans.md", "Plans.md", false},
		{"src/main.go", "Plans.md", false},
		{"NotPlans.md", "Plans.md", false},
	}
	for _, c := range cases {
		got := isPlansFile(c.changed, c.plans)
		if got != c.expected {
			t.Errorf("isPlansFile(%q, %q) = %v, want %v", c.changed, c.plans, got, c.expected)
		}
	}
}

func TestIsPlansFileWithRoot(t *testing.T) {
	// isPlansFileWithRoot は projectRoot を使って相対パスを解決する。
	projectRoot := "/home/user/project"
	cases := []struct {
		changedFile string
		plansFile   string
		expected    bool
		desc        string
	}{
		// 相対パス → projectRoot で解決して一致
		{"Plans.md", "/home/user/project/Plans.md", true, "relative path match"},
		// 絶対パスで一致
		{"/home/user/project/Plans.md", "/home/user/project/Plans.md", true, "absolute path match"},
		// 別ディレクトリの相対パス（Plans.md だが plansFile は projectRoot 直下）
		{"docs/Plans.md", "/home/user/project/Plans.md", false, "subdirectory mismatch"},
		// 別ディレクトリの相対パスが plansFile と一致する場合
		{"docs/Plans.md", "/home/user/project/docs/Plans.md", true, "subdirectory match"},
		// 全く別のファイル
		{"src/main.go", "/home/user/project/Plans.md", false, "non plans file"},
		// 別プロジェクトの同名ファイル（絶対パス）
		{"/tmp/other/Plans.md", "/home/user/project/Plans.md", false, "different project Plans.md"},
	}
	for _, tc := range cases {
		got := isPlansFileWithRoot(tc.changedFile, tc.plansFile, projectRoot)
		if got != tc.expected {
			t.Errorf("[%s] isPlansFileWithRoot(%q, %q, %q) = %v, want %v",
				tc.desc, tc.changedFile, tc.plansFile, projectRoot, got, tc.expected)
		}
	}
}

func TestCountMarker(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	content := "cc:TODO\ncc:TODO\ncc:WIP\ncc:完了\npm:依頼中\n"
	if err := os.WriteFile("Plans.md", []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		marker   string
		expected int
	}{
		{"cc:TODO", 2},
		{"cc:WIP", 1},
		{"cc:完了", 1},
		{"pm:依頼中", 1},
		{"pm:確認済", 0},
	}
	for _, c := range cases {
		got := countMarker("Plans.md", c.marker)
		if got != c.expected {
			t.Errorf("countMarker(Plans.md, %q) = %d, want %d", c.marker, got, c.expected)
		}
	}
}

func TestHandlePlansWatcher_CursorCompatMarker(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	// cursor:依頼中 が pm:依頼中 と同義で扱われること
	plansContent := "| Task 1 | A | DoD | - | cursor:依頼中 |\n"
	if err := os.WriteFile("Plans.md", []byte(plansContent), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(".claude/state", 0o755); err != nil {
		t.Fatal(err)
	}
	prevState := `{"timestamp":"2026-01-01T00:00:00Z","pm_pending":0,"cc_todo":0,"cc_wip":0,"cc_done":0,"pm_confirmed":0}`
	if err := os.WriteFile(plansStateFile, []byte(prevState), 0o644); err != nil {
		t.Fatal(err)
	}

	input := `{"tool_name":"Edit","tool_input":{"file_path":"Plans.md"}}`
	var out bytes.Buffer
	if err := HandlePlansWatcher(strings.NewReader(input), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result postToolOutput
	if jsonErr := json.Unmarshal(out.Bytes(), &result); jsonErr != nil {
		t.Fatalf("invalid JSON output: %v", jsonErr)
	}

	// cursor:依頼中 も新規タスクとして検出されること
	if !strings.Contains(result.HookSpecificOutput.AdditionalContext, "新規タスク") {
		t.Errorf("expected '新規タスク' for cursor:依頼中, got %q",
			result.HookSpecificOutput.AdditionalContext)
	}
}

// TestHandlePlansWatcher_CustomPlansDirectory は plansDirectory 設定があるとき
// カスタムディレクトリの Plans.md を正しく検出することを確認する。
func TestHandlePlansWatcher_CustomPlansDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	// 設定ファイルを作成（plansDirectory: docs）
	configContent := "plansDirectory: docs\n"
	if err := os.WriteFile(harnessConfigFileName, []byte(configContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// docs/Plans.md を作成
	if err := os.MkdirAll("docs", 0o755); err != nil {
		t.Fatal(err)
	}
	plansContent := "| Task 1 | 実装A | DoD | - | pm:依頼中 |\n"
	if err := os.WriteFile("docs/Plans.md", []byte(plansContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// 前回の状態（pm_pending=0）を保存
	if err := os.MkdirAll(".claude/state", 0o755); err != nil {
		t.Fatal(err)
	}
	prevState := `{"timestamp":"2026-01-01T00:00:00Z","pm_pending":0,"cc_todo":0,"cc_wip":0,"cc_done":0,"pm_confirmed":0}`
	if err := os.WriteFile(plansStateFile, []byte(prevState), 0o644); err != nil {
		t.Fatal(err)
	}

	// docs/Plans.md を変更したイベントを送信
	input := `{"tool_name":"Edit","tool_input":{"file_path":"docs/Plans.md"}}`
	var out bytes.Buffer
	if err := HandlePlansWatcher(strings.NewReader(input), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result postToolOutput
	if jsonErr := json.Unmarshal(out.Bytes(), &result); jsonErr != nil {
		t.Fatalf("invalid JSON output: %v, raw: %s", jsonErr, out.String())
	}

	// 新規タスクが検出されること（カスタムパスの Plans.md が認識される）
	if !strings.Contains(result.HookSpecificOutput.AdditionalContext, "新規タスク") {
		t.Errorf("expected '新規タスク' in additionalContext for custom plansDirectory, got %q",
			result.HookSpecificOutput.AdditionalContext)
	}
}

// TestIsPlansFile_CustomPath はカスタムパスの Plans.md に対して
// isPlansFileWithRoot が正しく動作することを確認する（P2修正: 別ディレクトリ同名ファイルを誤マッチしない）。
func TestIsPlansFile_CustomPath(t *testing.T) {
	projectRoot := "/project"
	cases := []struct {
		changedFile string
		plansFile   string
		want        bool
		desc        string
	}{
		// 完全一致（相対パスをprojectRootで解決）
		{"Plans.md", "/project/Plans.md", true, "exact match via projectRoot"},
		// カスタムディレクトリの Plans.md が plansFile と一致
		{"docs/Plans.md", "/project/docs/Plans.md", true, "custom subdir match"},
		// 別ディレクトリの同名ファイルは誤マッチしない（修正の核心）
		{"docs/Plans.md", "/project/Plans.md", false, "subdirectory mismatch - must not match"},
		// 全く別のファイル
		{"src/main.go", "/project/Plans.md", false, "non plans file"},
		{"README.md", "/project/Plans.md", false, "readme not plans"},
		// ファイル名が Plans.md に似ているが別ファイル
		{"Plans.md.bak", "/project/Plans.md", false, "backup file not matched"},
		// 絶対パスで別プロジェクトの Plans.md
		{"/tmp/other/Plans.md", "/project/Plans.md", false, "different project Plans.md must not match"},
	}

	for _, tc := range cases {
		got := isPlansFileWithRoot(tc.changedFile, tc.plansFile, projectRoot)
		if got != tc.want {
			t.Errorf("[%s] isPlansFileWithRoot(%q, %q, %q) = %v, want %v",
				tc.desc, tc.changedFile, tc.plansFile, projectRoot, got, tc.want)
		}
	}
}
