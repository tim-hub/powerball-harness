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
