package hookhandler

// task_completed.go - task-completed.sh の Go 移植 (エントリポイント)
//
// TaskCompleted イベント（チームモード）ハンドラ。
// タスク完了をタイムラインに記録し、Breezing 状態管理・
// テスト失敗エスカレーション・harness-mem finalize を担う。
//
// 元スクリプト: scripts/hook-handlers/task-completed.sh

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// taskCompletedInput は TaskCompleted フックの stdin JSON。
type taskCompletedInput struct {
	TeammateName    string      `json:"teammate_name"`
	AgentName       string      `json:"agent_name"`
	TaskID          string      `json:"task_id"`
	TaskSubject     string      `json:"task_subject"`
	Subject         string      `json:"subject"`
	TaskDescription string      `json:"task_description"`
	Description     string      `json:"description"`
	AgentID         string      `json:"agent_id"`
	AgentType       string      `json:"agent_type"`
	Continue        *bool       `json:"continue"`
	StopReason      string      `json:"stopReason"`
	StopReasonSnake string      `json:"stop_reason"`
	CWD             string      `json:"cwd"`
	ProjectRoot     string      `json:"project_root"`
}

// taskCompletedHandler は task-completed の全状態を保持する。
type taskCompletedHandler struct {
	projectRoot    string
	stateDir       string
	timelineFile   string
	pendingFixFile string
	finalizeMarker string
}

// HandleTaskCompleted は task-completed.sh の Go 移植エントリポイント。
func HandleTaskCompleted(in io.Reader, out io.Writer) error {
	data, err := io.ReadAll(in)
	if err != nil || len(data) == 0 {
		return writeJSON(out, approveResponse("TaskCompleted: no payload"))
	}

	var input taskCompletedInput
	if err := json.Unmarshal(data, &input); err != nil {
		return writeJSON(out, approveResponse("TaskCompleted: invalid payload"))
	}

	// プロジェクトルートを決定
	projectRoot := input.ProjectRoot
	if projectRoot == "" {
		projectRoot = input.CWD
	}
	if projectRoot == "" {
		projectRoot, _ = os.Getwd()
	}

	h := &taskCompletedHandler{
		projectRoot:    projectRoot,
		stateDir:       filepath.Join(projectRoot, ".claude", "state"),
		timelineFile:   filepath.Join(projectRoot, ".claude", "state", "breezing-timeline.jsonl"),
		pendingFixFile: filepath.Join(projectRoot, ".claude", "state", "pending-fix-proposals.jsonl"),
		finalizeMarker: filepath.Join(projectRoot, ".claude", "state", "harness-mem-finalize-work-completed.json"),
	}

	return h.handle(input, data, out)
}

func (h *taskCompletedHandler) handle(input taskCompletedInput, rawData []byte, out io.Writer) error {
	// フィールド正規化
	teammateName := firstNonEmpty(input.TeammateName, input.AgentName)
	taskID := input.TaskID
	taskSubject := firstNonEmpty(input.TaskSubject, input.Subject)
	taskDesc := input.TaskDescription
	if taskDesc == "" {
		taskDesc = input.Description
	}
	if len(taskDesc) > 100 {
		taskDesc = taskDesc[:100]
	}
	agentID := input.AgentID
	agentType := input.AgentType

	stopReason := firstNonEmpty(input.StopReason, input.StopReasonSnake)
	requestContinue := true // デフォルトは続行
	if input.Continue != nil {
		requestContinue = *input.Continue
	}

	ts := utcNow()

	// 状態ディレクトリの作成
	if err := os.MkdirAll(h.stateDir, 0o700); err != nil {
		fmt.Fprintf(os.Stderr, "[task-completed] mkdir: %v\n", err)
	}

	// タイムラインへ記録
	h.appendTimeline(timelineEntry{
		Event:       "task_completed",
		Teammate:    teammateName,
		TaskID:      taskID,
		Subject:     taskSubject,
		Description: taskDesc,
		AgentID:     agentID,
		AgentType:   agentType,
		Timestamp:   ts,
	})

	// Breezing シグナル生成
	totalTasks, completedCount := h.updateBreezingSignals(taskID, ts)

	// テスト結果チェック
	testOK, failCount := h.checkTestResultAndEscalate(taskID, taskSubject, teammateName, ts)
	if !testOK {
		if failCount >= 3 {
			// 3-strike エスカレーション
			return h.emitEscalationResponse(out, taskID, taskSubject, failCount)
		}
		return writeJSON(out, map[string]string{
			"decision": "block",
			"reason":   "TaskCompleted: test result shows failure - escalation required",
		})
	}

	// Webhook 通知（同期、5秒タイムアウト）
	h.fireWebhook(rawData)

	// 停止判定
	if !requestContinue || stopReason != "" {
		finalReason := stopReason
		if finalReason == "" {
			finalReason = "TaskCompleted requested stop"
		}
		return writeJSON(out, map[string]interface{}{
			"continue":   false,
			"stopReason": finalReason,
		})
	}

	// 全タスク完了判定
	if totalTasks > 0 && completedCount >= totalTasks {
		h.maybeFinalizeHarnessMem(ts)
		return writeJSON(out, map[string]interface{}{
			"continue":   false,
			"stopReason": "all_tasks_completed",
		})
	}

	// プログレスサマリー付き承認レスポンス
	if totalTasks > 0 && taskSubject != "" {
		progressMsg := fmt.Sprintf("Progress: Task %d/%d 完了 — %q", completedCount, totalTasks, taskSubject)
		return writeJSON(out, map[string]string{
			"decision":      "approve",
			"reason":        "TaskCompleted tracked",
			"systemMessage": progressMsg,
		})
	}

	return writeJSON(out, approveResponse("TaskCompleted tracked"))
}

// approveResponse は標準承認レスポンスを返す。
func approveResponse(reason string) map[string]string {
	return map[string]string{
		"decision": "approve",
		"reason":   reason,
	}
}
