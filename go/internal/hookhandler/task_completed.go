package hookhandler

// task_completed.go - Go port of task-completed.sh (entry point)
//
// Handler for TaskCompleted events (team mode).
// Records task completion to the timeline and handles Breezing state management,
// test-failure escalation, and harness-mem finalization.
//
// Original script: scripts/hook-handlers/task-completed.sh

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// taskCompletedInput is the stdin JSON for the TaskCompleted hook.
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

// taskCompletedHandler holds all state for task-completed.
type taskCompletedHandler struct {
	projectRoot    string
	stateDir       string
	timelineFile   string
	pendingFixFile string
	finalizeMarker string
	// plansPath is the resolved path to Plans.md (respects plansDirectory from config).
	// Empty string when the file does not exist.
	plansPath string
}

// HandleTaskCompleted is the Go port entry point for task-completed.sh.
func HandleTaskCompleted(in io.Reader, out io.Writer) error {
	data, err := io.ReadAll(in)
	if err != nil || len(data) == 0 {
		return writeJSON(out, approveResponse("TaskCompleted: no payload"))
	}

	var input taskCompletedInput
	if err := json.Unmarshal(data, &input); err != nil {
		return writeJSON(out, approveResponse("TaskCompleted: invalid payload"))
	}

	// Determine the project root.
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
		// Resolve the Plans.md path, respecting plansDirectory in the config file.
		plansPath: resolvePlansPath(projectRoot),
	}

	return h.handle(input, data, out)
}

func (h *taskCompletedHandler) handle(input taskCompletedInput, rawData []byte, out io.Writer) error {
	// Normalize fields.
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
	requestContinue := true // Default: continue.
	if input.Continue != nil {
		requestContinue = *input.Continue
	}

	ts := utcNow()

	// Create the state directory.
	if err := os.MkdirAll(h.stateDir, 0o700); err != nil {
		fmt.Fprintf(os.Stderr, "[task-completed] mkdir: %v\n", err)
	}

	// Append to the timeline.
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

	// Generate Breezing signals.
	totalTasks, completedCount := h.updateBreezingSignals(taskID, ts)

	// Check test results.
	testOK, failCount := h.checkTestResultAndEscalate(taskID, taskSubject, teammateName, ts)
	if !testOK {
		if failCount >= 3 {
			// 3-strike escalation.
			return h.emitEscalationResponse(out, taskID, taskSubject, failCount)
		}
		return writeJSON(out, map[string]string{
			"decision": "block",
			"reason":   "TaskCompleted: test result shows failure - escalation required",
		})
	}

	// Webhook notification (synchronous, 5-second timeout).
	h.fireWebhook(rawData)

	// Determine whether to stop.
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

	// Check whether all tasks are complete.
	if totalTasks > 0 && completedCount >= totalTasks {
		h.maybeFinalizeHarnessMem(ts)
		return writeJSON(out, map[string]interface{}{
			"continue":   false,
			"stopReason": "all_tasks_completed",
		})
	}

	// Approval response with progress summary.
	if totalTasks > 0 && taskSubject != "" {
		progressMsg := fmt.Sprintf("Progress: Task %d/%d completed — %q", completedCount, totalTasks, taskSubject)
		return writeJSON(out, map[string]string{
			"decision":      "approve",
			"reason":        "TaskCompleted tracked",
			"systemMessage": progressMsg,
		})
	}

	return writeJSON(out, approveResponse("TaskCompleted tracked"))
}

// approveResponse returns a standard approval response.
func approveResponse(reason string) map[string]string {
	return map[string]string{
		"decision": "approve",
		"reason":   reason,
	}
}
