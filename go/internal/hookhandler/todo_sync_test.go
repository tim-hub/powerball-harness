package hookhandler

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTodoSyncHandler_EmptyInput(t *testing.T) {
	h := &TodoSyncHandler{ProjectRoot: t.TempDir()}

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

func TestTodoSyncHandler_NotTodoWrite(t *testing.T) {
	h := &TodoSyncHandler{ProjectRoot: t.TempDir()}

	input := `{"tool_name":"Read","tool_input":{"file_path":"Plans.md"}}`

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(input), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// TodoWrite 以外はスキップ
	if out.Len() != 0 {
		t.Errorf("expected no output for non-TodoWrite tool, got %q", out.String())
	}
}

func TestTodoSyncHandler_EmptyTodos(t *testing.T) {
	h := &TodoSyncHandler{ProjectRoot: t.TempDir()}

	input := `{"tool_name":"TodoWrite","tool_input":{"todos":[]}}`

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(input), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 空 todos は出力なし
	if out.Len() != 0 {
		t.Errorf("expected no output for empty todos, got %q", out.String())
	}
}

func TestTodoSyncHandler_CountsInOutput(t *testing.T) {
	dir := t.TempDir()
	// Plans.md が存在しないとスキップされるため作成する
	if err := os.WriteFile(filepath.Join(dir, "Plans.md"), []byte("# Plans\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	h := &TodoSyncHandler{ProjectRoot: dir}

	input := `{"tool_name":"TodoWrite","tool_input":{"todos":[
		{"status":"pending"},
		{"status":"pending"},
		{"status":"in_progress"},
		{"status":"completed"}
	]}}`

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(input), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// JSON 出力の additionalContext を確認
	type hookOutput struct {
		AdditionalContext string `json:"additionalContext"`
	}
	type output struct {
		HookSpecificOutput hookOutput `json:"hookSpecificOutput"`
	}
	var result output
	if err := json.Unmarshal(bytes.TrimRight(out.Bytes(), "\n"), &result); err != nil {
		t.Fatalf("invalid JSON: %s", out.String())
	}

	ctx := result.HookSpecificOutput.AdditionalContext
	if !strings.Contains(ctx, "TODO=2") {
		t.Errorf("expected TODO=2 in context, got %q", ctx)
	}
	if !strings.Contains(ctx, "WIP=1") {
		t.Errorf("expected WIP=1 in context, got %q", ctx)
	}
	if !strings.Contains(ctx, "done=1") {
		t.Errorf("expected done=1 in context, got %q", ctx)
	}
}

func TestTodoSyncHandler_SavesSyncState(t *testing.T) {
	dir := t.TempDir()
	// Plans.md が存在しないとスキップされるため作成する
	if err := os.WriteFile(filepath.Join(dir, "Plans.md"), []byte("# Plans\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	h := &TodoSyncHandler{ProjectRoot: dir}

	input := `{"tool_name":"TodoWrite","tool_input":{"todos":[
		{"status":"pending"},
		{"status":"completed"}
	]}}`

	var out bytes.Buffer
	_ = h.Handle(strings.NewReader(input), &out)

	// 状態ファイルが作成されているか確認
	stateFile := filepath.Join(dir, ".claude", "state", todoSyncStateFile)
	data, err := os.ReadFile(stateFile)
	if err != nil {
		t.Fatalf("sync state file not created: %v", err)
	}

	var state struct {
		SyncedAt string     `json:"synced_at"`
		Todos    []todoItem `json:"todos"`
	}
	if err := json.Unmarshal(data, &state); err != nil {
		t.Fatalf("invalid state file JSON: %s", string(data))
	}
	if state.SyncedAt == "" {
		t.Errorf("expected synced_at to be set")
	}
	if len(state.Todos) != 2 {
		t.Errorf("expected 2 todos in state, got %d", len(state.Todos))
	}
}

func TestTodoSyncHandler_AppendsEventLog(t *testing.T) {
	dir := t.TempDir()
	// Plans.md が存在しないとスキップされるため作成する
	if err := os.WriteFile(filepath.Join(dir, "Plans.md"), []byte("# Plans\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	h := &TodoSyncHandler{ProjectRoot: dir}

	stateDir := filepath.Join(dir, ".claude", "state")
	_ = os.MkdirAll(stateDir, 0700)

	// session.events.jsonl を事前に作成（存在する場合のみ追記）
	eventLog := filepath.Join(stateDir, "session.events.jsonl")
	_ = os.WriteFile(eventLog, []byte(""), 0600)

	input := `{"tool_name":"TodoWrite","tool_input":{"todos":[
		{"status":"pending"},
		{"status":"completed"}
	]}}`

	var out bytes.Buffer
	_ = h.Handle(strings.NewReader(input), &out)

	// イベントログに追記されているか確認
	data, err := os.ReadFile(eventLog)
	if err != nil {
		t.Fatalf("event log not found: %v", err)
	}

	var event struct {
		Type string `json:"type"`
		Data struct {
			Pending    int `json:"pending"`
			Completed  int `json:"completed"`
		} `json:"data"`
	}
	if err := json.Unmarshal(bytes.TrimRight(data, "\n"), &event); err != nil {
		t.Fatalf("invalid event JSON: %s", string(data))
	}
	if event.Type != "todo.sync" {
		t.Errorf("expected event type=todo.sync, got %q", event.Type)
	}
	if event.Data.Pending != 1 {
		t.Errorf("expected pending=1, got %d", event.Data.Pending)
	}
	if event.Data.Completed != 1 {
		t.Errorf("expected completed=1, got %d", event.Data.Completed)
	}
}

func TestTodoSyncHandler_NoEventLog_NoError(t *testing.T) {
	dir := t.TempDir()
	// Plans.md が存在しないとスキップされるため作成する
	if err := os.WriteFile(filepath.Join(dir, "Plans.md"), []byte("# Plans\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	h := &TodoSyncHandler{ProjectRoot: dir}

	// session.events.jsonl が存在しない場合はスキップ（エラーにならない）
	input := `{"tool_name":"TodoWrite","tool_input":{"todos":[
		{"status":"completed"}
	]}}`

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(input), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTodoSyncHandler_WorkModeWarning_AllComplete(t *testing.T) {
	dir := t.TempDir()
	// Plans.md が存在しないとスキップされるため作成する
	if err := os.WriteFile(filepath.Join(dir, "Plans.md"), []byte("# Plans\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	h := &TodoSyncHandler{ProjectRoot: dir}

	stateDir := filepath.Join(dir, ".claude", "state")
	_ = os.MkdirAll(stateDir, 0700)

	// work-active.json を作成（review_status=pending）
	workFile := filepath.Join(stateDir, "work-active.json")
	_ = os.WriteFile(workFile, []byte(`{"review_status":"pending"}`), 0600)

	// 全タスク完了（pending=0, in_progress=0, completed>0）
	input := `{"tool_name":"TodoWrite","tool_input":{"todos":[
		{"status":"completed"},
		{"status":"completed"}
	]}}`

	var out bytes.Buffer
	_ = h.Handle(strings.NewReader(input), &out)

	type hookOutput struct {
		AdditionalContext string `json:"additionalContext"`
	}
	type output struct {
		HookSpecificOutput hookOutput `json:"hookSpecificOutput"`
	}
	var result output
	_ = json.Unmarshal(bytes.TrimRight(out.Bytes(), "\n"), &result)

	ctx := result.HookSpecificOutput.AdditionalContext
	if !strings.Contains(ctx, "harness-review") {
		t.Errorf("expected harness-review warning in context, got %q", ctx)
	}
}

func TestTodoSyncHandler_WorkModeWarning_ReviewPassed(t *testing.T) {
	dir := t.TempDir()
	// Plans.md が存在しないとスキップされるため作成する
	if err := os.WriteFile(filepath.Join(dir, "Plans.md"), []byte("# Plans\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	h := &TodoSyncHandler{ProjectRoot: dir}

	stateDir := filepath.Join(dir, ".claude", "state")
	_ = os.MkdirAll(stateDir, 0700)

	// review_status=passed の場合は警告しない
	workFile := filepath.Join(stateDir, "work-active.json")
	_ = os.WriteFile(workFile, []byte(`{"review_status":"passed"}`), 0600)

	input := `{"tool_name":"TodoWrite","tool_input":{"todos":[
		{"status":"completed"}
	]}}`

	var out bytes.Buffer
	_ = h.Handle(strings.NewReader(input), &out)

	type hookOutput struct {
		AdditionalContext string `json:"additionalContext"`
	}
	type output struct {
		HookSpecificOutput hookOutput `json:"hookSpecificOutput"`
	}
	var result output
	_ = json.Unmarshal(bytes.TrimRight(out.Bytes(), "\n"), &result)

	ctx := result.HookSpecificOutput.AdditionalContext
	if strings.Contains(ctx, "harness-review") {
		t.Errorf("expected no warning when review_status=passed, got %q", ctx)
	}
}

func TestTodoSyncHandler_WorkModeWarning_StillHasPending(t *testing.T) {
	dir := t.TempDir()
	// Plans.md が存在しないとスキップされるため作成する
	if err := os.WriteFile(filepath.Join(dir, "Plans.md"), []byte("# Plans\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	h := &TodoSyncHandler{ProjectRoot: dir}

	stateDir := filepath.Join(dir, ".claude", "state")
	_ = os.MkdirAll(stateDir, 0700)

	workFile := filepath.Join(stateDir, "work-active.json")
	_ = os.WriteFile(workFile, []byte(`{"review_status":"pending"}`), 0600)

	// pending タスクがある場合は警告しない
	input := `{"tool_name":"TodoWrite","tool_input":{"todos":[
		{"status":"pending"},
		{"status":"completed"}
	]}}`

	var out bytes.Buffer
	_ = h.Handle(strings.NewReader(input), &out)

	type hookOutput struct {
		AdditionalContext string `json:"additionalContext"`
	}
	type output struct {
		HookSpecificOutput hookOutput `json:"hookSpecificOutput"`
	}
	var result output
	_ = json.Unmarshal(bytes.TrimRight(out.Bytes(), "\n"), &result)

	ctx := result.HookSpecificOutput.AdditionalContext
	if strings.Contains(ctx, "harness-review") {
		t.Errorf("expected no warning with pending todos, got %q", ctx)
	}
}

// TestTodoSyncHandler_SkipWhenNoPlansFile は Plans.md が存在しない場合に
// 出力なしでスキップされることを確認する（bash 版の動作と一致）。
func TestTodoSyncHandler_SkipWhenNoPlansFile(t *testing.T) {
	dir := t.TempDir()
	// Plans.md を意図的に作成しない
	h := &TodoSyncHandler{ProjectRoot: dir}

	input := `{"tool_name":"TodoWrite","tool_input":{"todos":[
		{"status":"pending"},
		{"status":"completed"}
	]}}`

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(input), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Plans.md がない場合は何も出力せずスキップ
	if out.Len() != 0 {
		t.Errorf("expected no output when Plans.md does not exist, got: %s", out.String())
	}
}

// TestTodoSyncHandler_UsesResolveProjectRoot は ProjectRoot が空のとき
// resolveProjectRoot() が使われ HARNESS_PROJECT_ROOT 環境変数が参照されることを確認する。
// os.Getwd() を使っていた場合、cwd != project root なら .claude/state が見つからず
// 状態ファイルが書き込まれないため、本テストで解決を検証する。
func TestTodoSyncHandler_UsesResolveProjectRoot(t *testing.T) {
	dir := t.TempDir()
	// HARNESS_PROJECT_ROOT を設定（resolveProjectRoot はこれを最優先で使う）
	t.Setenv("HARNESS_PROJECT_ROOT", dir)

	// Plans.md を作成
	if err := os.WriteFile(filepath.Join(dir, "Plans.md"), []byte("# Plans\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// ProjectRoot を空にすることで resolveProjectRoot() を経由させる
	h := &TodoSyncHandler{ProjectRoot: ""}

	input := `{"tool_name":"TodoWrite","tool_input":{"todos":[
		{"status":"pending"},
		{"status":"completed"}
	]}}`

	var out bytes.Buffer
	if err := h.Handle(strings.NewReader(input), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 処理が通り出力があること（resolveProjectRoot で dir が解決されたため）
	if out.Len() == 0 {
		t.Errorf("expected output when HARNESS_PROJECT_ROOT is set, got none")
	}

	// 状態ファイルが HARNESS_PROJECT_ROOT 配下に作成されていること
	stateFile := filepath.Join(dir, ".claude", "state", todoSyncStateFile)
	if _, err := os.Stat(stateFile); err != nil {
		t.Errorf("sync state file should be created at %s: %v", stateFile, err)
	}
}

// TestTodoSyncHandler_CustomPlansDirectory は plansDirectory 設定があるとき
// カスタムディレクトリの Plans.md が存在する場合に処理が通ることを確認する。
func TestTodoSyncHandler_CustomPlansDirectory(t *testing.T) {
	dir := t.TempDir()

	// 設定ファイルを作成（plansDirectory: work）
	configContent := "plansDirectory: work\n"
	if err := os.WriteFile(filepath.Join(dir, harnessConfigFileName), []byte(configContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// work/Plans.md を作成
	workDir := filepath.Join(dir, "work")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(workDir, "Plans.md"), []byte("# Plans\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	h := &TodoSyncHandler{ProjectRoot: dir}

	input := `{"tool_name":"TodoWrite","tool_input":{"todos":[
		{"status":"pending"},
		{"status":"completed"}
	]}}`

	var out bytes.Buffer
	if err := h.Handle(strings.NewReader(input), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 処理が通り出力があること（Plans.md が存在するので処理される）
	if out.Len() == 0 {
		t.Errorf("expected output when custom-dir Plans.md exists, got none")
	}
}
