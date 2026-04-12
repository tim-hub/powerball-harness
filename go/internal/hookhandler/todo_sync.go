package hookhandler

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// TodoSyncHandler は PostToolUse フックハンドラ（TodoWrite と Plans.md の同期）。
// TodoWrite の内容を解析し、状態カウントを記録・イベントログに追記する。
// work-mode で全タスクが完了した場合は追加警告を出力する。
//
// shell 版: scripts/todo-sync.sh
type TodoSyncHandler struct {
	// ProjectRoot はプロジェクトルートのパス。空の場合は cwd を使用する。
	ProjectRoot string
}

// todoSyncInput は PostToolUse フックの stdin JSON。
type todoSyncInput struct {
	ToolName  string        `json:"tool_name"`
	ToolInput todoWriteBody `json:"tool_input"`
}

// todoWriteBody は TodoWrite ツールの tool_input。
type todoWriteBody struct {
	Todos []todoItem `json:"todos"`
}

// todoItem は TodoWrite の 1 件エントリ。
type todoItem struct {
	Status string `json:"status"`
}

// todoSyncStateFile は同期状態の保存先ファイル名。
const todoSyncStateFile = "todo-sync-state.json"

// Handle は stdin からペイロードを読み取り、TodoWrite の状態を記録・通知する。
func (h *TodoSyncHandler) Handle(r io.Reader, w io.Writer) error {
	data, _ := io.ReadAll(r)

	if len(data) == 0 {
		return nil
	}

	var inp todoSyncInput
	if err := json.Unmarshal(data, &inp); err != nil {
		return nil
	}

	// TodoWrite 以外はスキップ
	if inp.ToolName != "TodoWrite" {
		return nil
	}

	todos := inp.ToolInput.Todos
	if len(todos) == 0 {
		return nil
	}

	// プロジェクトルートを決定。
	// os.Getwd() ではなく resolveProjectRoot() を使うことで monorepo の
	// サブディレクトリから実行した場合でも .claude/state が正しく解決される
	// （git rev-parse --show-toplevel 対応）。
	projectRoot := h.ProjectRoot
	if projectRoot == "" {
		projectRoot = resolveProjectRoot()
	}

	// Plans.md が存在しない場合はスキップ（bash 版の動作と一致）
	if resolvePlansPath(projectRoot) == "" {
		return nil
	}

	stateDir := filepath.Join(projectRoot, ".claude", "state")
	_ = os.MkdirAll(stateDir, 0700)

	// カウント集計
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

	// 同期状態ファイルに保存
	h.saveSyncState(stateDir, todos)

	// イベントログに追記
	h.appendEventLog(stateDir, pending, inProgress, done)

	// work-mode 警告チェック
	workWarning := h.checkWorkModeWarning(stateDir, pending, inProgress, done)

	// additionalContext として同期情報を出力
	ctx := fmt.Sprintf("[TodoSync] Plans.md と同期: TODO=%d, WIP=%d, done=%d%s",
		pending, inProgress, done, workWarning)

	return writeTodoSyncOutput(w, ctx)
}

// saveSyncState は todos の状態を JSON ファイルに保存する。
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

// appendEventLog はイベントログ JSONL ファイルに同期イベントを追記する。
// session.events.jsonl が存在する場合のみ追記する（bash の動作と一致）。
func (h *TodoSyncHandler) appendEventLog(stateDir string, pending, inProgress, done int) {
	eventLog := filepath.Join(stateDir, "session.events.jsonl")

	// bash と同様に、ファイルが存在する場合のみ追記
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

// checkWorkModeWarning は全完了かつ work-mode が有効な場合に警告文字列を返す。
func (h *TodoSyncHandler) checkWorkModeWarning(stateDir string, pending, inProgress, done int) string {
	// 全タスク完了（pending=0, in_progress=0, completed>0）でなければスキップ
	if pending != 0 || inProgress != 0 || done == 0 {
		return ""
	}

	// work-active.json または ultrawork-active.json の存在を確認
	workFile := filepath.Join(stateDir, "work-active.json")
	if _, err := os.Stat(workFile); err != nil {
		workFile = filepath.Join(stateDir, "ultrawork-active.json")
		if _, err2 := os.Stat(workFile); err2 != nil {
			return ""
		}
	}

	// review_status を確認
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

	return fmt.Sprintf("\n\n⚠️ **work 完了前チェック**: review_status=%s\n→ 完了処理の前に /harness-review で APPROVE を取得してください",
		state.ReviewStatus)
}

// writeTodoSyncOutput は additionalContext を JSON として w に書き出す。
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
