package event

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// PermissionDeniedHandler は PermissionDenied フックハンドラ (v2.1.89+)。
// auto mode classifier がコマンドを拒否した際に発火する。
// 拒否イベントを telemetry に記録し、Worker モードでは Lead への通知と retry ヒントを返す。
//
// shell 版: scripts/hook-handlers/permission-denied-handler.sh
type PermissionDeniedHandler struct {
	// StateDir はログファイルの保存先。
	// 空の場合は ResolveStateDir(projectRoot) を使う。
	StateDir string
}

// permissionDeniedInput は PermissionDenied フックの stdin JSON。
type permissionDeniedInput struct {
	Tool         string `json:"tool,omitempty"`
	ToolName     string `json:"tool_name,omitempty"`
	DeniedReason string `json:"denied_reason,omitempty"`
	Reason       string `json:"reason,omitempty"`
	SessionID    string `json:"session_id,omitempty"`
	AgentID      string `json:"agent_id,omitempty"`
	AgentType    string `json:"agent_type,omitempty"`
	CWD          string `json:"cwd,omitempty"`
}

// permissionDeniedLogEntry は permission-denied.jsonl に書き出すエントリ。
type permissionDeniedLogEntry struct {
	Event     string `json:"event"`
	Timestamp string `json:"timestamp"`
	SessionID string `json:"session_id"`
	AgentID   string `json:"agent_id"`
	AgentType string `json:"agent_type"`
	Tool      string `json:"tool"`
	Reason    string `json:"reason"`
}

// Handle は stdin から PermissionDenied ペイロードを読み取り、
// ログに記録した後、Worker の場合は retry + systemMessage を返す。
// Worker 以外の場合は approve を返す。
func (h *PermissionDeniedHandler) Handle(r io.Reader, w io.Writer) error {
	data, err := io.ReadAll(r)
	if err != nil || len(data) == 0 {
		return nil
	}

	var inp permissionDeniedInput
	if err := json.Unmarshal(data, &inp); err != nil {
		return nil
	}

	// フィールドの正規化
	toolName := inp.Tool
	if toolName == "" {
		toolName = inp.ToolName
	}
	if toolName == "" {
		toolName = "unknown"
	}

	deniedReason := inp.DeniedReason
	if deniedReason == "" {
		deniedReason = inp.Reason
	}
	if deniedReason == "" {
		deniedReason = "unknown"
	}

	sessionID := inp.SessionID
	if sessionID == "" {
		sessionID = "unknown"
	}
	agentID := inp.AgentID
	if agentID == "" {
		agentID = "unknown"
	}
	agentType := inp.AgentType
	if agentType == "" {
		agentType = "unknown"
	}

	// ステートディレクトリを決定
	projectRoot := inp.CWD
	if projectRoot == "" {
		projectRoot = resolveProjectRoot(data)
	}

	stateDir := h.StateDir
	if stateDir == "" {
		stateDir = ResolveStateDir(projectRoot)
	}

	if err := EnsureStateDir(stateDir); err != nil {
		return nil
	}

	logFile := filepath.Join(stateDir, "permission-denied.jsonl")
	if isSymlink(logFile) {
		return nil
	}

	// ログに記録
	entry := permissionDeniedLogEntry{
		Event:     "permission_denied",
		Timestamp: Now(),
		SessionID: sessionID,
		AgentID:   agentID,
		AgentType: agentType,
		Tool:      toolName,
		Reason:    deniedReason,
	}
	h.appendLog(logFile, entry)

	// stderr にデバッグ出力
	fmt.Fprintf(os.Stderr,
		"[PermissionDenied] agent=%s type=%s tool=%s reason=%s\n",
		agentID, agentType, toolName, deniedReason,
	)

	// Worker の場合: retry + systemMessage を返す
	if h.isWorker(agentType) {
		notificationText := fmt.Sprintf(
			"[PermissionDenied] Worker のツール %s が auto mode で拒否されました。"+
				"理由: %s。代替アプローチを検討するか、必要なら手動承認してください。",
			toolName, deniedReason,
		)
		return WriteJSON(w, RetryResponse{
			Retry:         true,
			SystemMessage: notificationText,
		})
	}

	// Worker 以外: approve を返す
	return WriteJSON(w, ApproveResponse{
		Decision: "approve",
		Reason:   "PermissionDenied logged",
	})
}

// isWorker は agentType が Worker であるかを判定する。
// "worker", "task-worker", または ":worker" で終わる場合に true を返す。
func (h *PermissionDeniedHandler) isWorker(agentType string) bool {
	return agentType == "worker" ||
		agentType == "task-worker" ||
		strings.HasSuffix(agentType, ":worker")
}

// appendLog は permission-denied ログに 1 エントリ追記する。
func (h *PermissionDeniedHandler) appendLog(path string, entry permissionDeniedLogEntry) {
	logData, err := json.Marshal(entry)
	if err != nil {
		return
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = fmt.Fprintf(f, "%s\n", logData)

	RotateJSONL(path)
}
