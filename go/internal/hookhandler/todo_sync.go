package hookhandler

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

//
type TodoSyncHandler struct {
	ProjectRoot string
}

type todoSyncInput struct {
	ToolName  string        `json:"tool_name"`
	ToolInput todoWriteBody `json:"tool_input"`
}

type todoWriteBody struct {
	Todos []todoItem `json:"todos"`
}

type todoItem struct {
	Status string `json:"status"`
}

const todoSyncStateFile = "todo-sync-state.json"

func (h *TodoSyncHandler) Handle(r io.Reader, w io.Writer) error {
	data, _ := io.ReadAll(r)

	if len(data) == 0 {
		return nil
	}

	var inp todoSyncInput
	if err := json.Unmarshal(data, &inp); err != nil {
		return nil
	}

	if inp.ToolName != "TodoWrite" {
		return nil
	}

	todos := inp.ToolInput.Todos
	if len(todos) == 0 {
		return nil
	}

	projectRoot := h.ProjectRoot
	if projectRoot == "" {
		projectRoot = resolveProjectRoot()
	}

	if resolvePlansPath(projectRoot) == "" {
		return nil
	}

	stateDir := filepath.Join(projectRoot, ".claude", "state")
	_ = os.MkdirAll(stateDir, 0700)

	var pending, inProgress, done int
	for _, t := range todos {
		switch t.Status {
		case "pending":
			pending++
		case "in_progress":
			inProgress++
		case "completed":
			done++
		}
	}

	h.saveSyncState(stateDir, todos)

	h.appendEventLog(stateDir, pending, inProgress, done)

	workWarning := h.checkWorkModeWarning(stateDir, pending, inProgress, done)

	ctx := fmt.Sprintf("[TodoSync] synced with Plans.md: TODO=%d, WIP=%d, done=%d%s",
		pending, inProgress, done, workWarning)

	return writeTodoSyncOutput(w, ctx)
}

func (h *TodoSyncHandler) saveSyncState(stateDir string, todos []todoItem) {
	type syncState struct {
		SyncedAt string     `json:"synced_at"`
		Todos    []todoItem `json:"todos"`
	}
	state := syncState{
		SyncedAt: time.Now().UTC().Format(time.RFC3339),
		Todos:    todos,
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(filepath.Join(stateDir, todoSyncStateFile), data, 0600)
}

func (h *TodoSyncHandler) appendEventLog(stateDir string, pending, inProgress, done int) {
	eventLog := filepath.Join(stateDir, "session.events.jsonl")

	if _, err := os.Stat(eventLog); err != nil {
		return
	}

	type eventData struct {
		Pending    int `json:"pending"`
		InProgress int `json:"in_progress"`
		Completed  int `json:"completed"`
	}
	type event struct {
		Type string    `json:"type"`
		Ts   string    `json:"ts"`
		Data eventData `json:"data"`
	}
	ev := event{
		Type: "todo.sync",
		Ts:   time.Now().UTC().Format(time.RFC3339),
		Data: eventData{
			Pending:    pending,
			InProgress: inProgress,
			Completed:  done,
		},
	}
	line, err := json.Marshal(ev)
	if err != nil {
		return
	}

	f, err := os.OpenFile(eventLog, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = fmt.Fprintf(f, "%s\n", line)
}

func (h *TodoSyncHandler) checkWorkModeWarning(stateDir string, pending, inProgress, done int) string {
	if pending != 0 || inProgress != 0 || done == 0 {
		return ""
	}

	workFile := filepath.Join(stateDir, "work-active.json")
	if _, err := os.Stat(workFile); err != nil {
		workFile = filepath.Join(stateDir, "ultrawork-active.json")
		if _, err2 := os.Stat(workFile); err2 != nil {
			return ""
		}
	}

	data, err := os.ReadFile(workFile)
	if err != nil {
		return ""
	}
	var state struct {
		ReviewStatus string `json:"review_status"`
	}
	if err := json.Unmarshal(data, &state); err != nil {
		return ""
	}

	if state.ReviewStatus == "passed" {
		return ""
	}

	return fmt.Sprintf("\n\n⚠️ **pre-completion check**: review_status=%s\n→ Run /harness-review to get APPROVE before completing",
		state.ReviewStatus)
}

func writeTodoSyncOutput(w io.Writer, ctx string) error {
	type hookOutput struct {
		AdditionalContext string `json:"additionalContext"`
	}
	type output struct {
		HookSpecificOutput hookOutput `json:"hookSpecificOutput"`
	}
	out := output{
		HookSpecificOutput: hookOutput{AdditionalContext: ctx},
	}
	data, err := json.Marshal(out)
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}
	_, err = fmt.Fprintf(w, "%s\n", data)
	return err
}
