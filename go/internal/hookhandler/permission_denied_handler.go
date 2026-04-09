package hookhandler

// permission_denied_handler.go
// permission-denied-handler.sh の Go 移植。
//
// PermissionDenied イベント（auto mode classifier が拒否した場合）を処理する:
//   - .claude/state/permission-denied-events.jsonl に記録
//   - Worker の場合は {retry: true, systemMessage: ...} を返す
//   - Worker 以外の場合は approve を返す

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// permissionDeniedInput は PermissionDenied フックの stdin JSON。
type permissionDeniedInput struct {
	Tool        string `json:"tool"`
	ToolName    string `json:"tool_name"`
	DeniedReason string `json:"denied_reason"`
	Reason      string `json:"reason"`
	SessionID   string `json:"session_id"`
	AgentID     string `json:"agent_id"`
	AgentType   string `json:"agent_type"`
}

// permissionDeniedLogEntry は permission-denied-events.jsonl の1エントリ。
type permissionDeniedLogEntry struct {
	Event     string `json:"event"`
	Timestamp string `json:"timestamp"`
	SessionID string `json:"session_id"`
	AgentID   string `json:"agent_id"`
	AgentType string `json:"agent_type"`
	Tool      string `json:"tool"`
	Reason    string `json:"reason"`
}

// permissionDeniedRetryResponse は Worker 向けの retry レスポンス。
type permissionDeniedRetryResponse struct {
	Retry         bool   `json:"retry"`
	SystemMessage string `json:"systemMessage"`
}

// permissionDeniedApproveResponse は Worker 以外向けの approve レスポンス。
type permissionDeniedApproveResponse struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
}

// HandlePermissionDenied は permission-denied-handler.sh の Go 移植。
//
// PermissionDenied フックで呼び出され:
//  1. .claude/state/permission-denied-events.jsonl にイベントを記録する
//  2. Worker の場合は {retry: true, systemMessage: ...} を返す
//  3. Worker 以外の場合は approve を返す
func HandlePermissionDenied(in io.Reader, out io.Writer) error {
	data, err := io.ReadAll(in)
	if err != nil || len(strings.TrimSpace(string(data))) == 0 {
		// 入力なし: 正常終了
		return nil
	}

	var input permissionDeniedInput
	if jsonErr := json.Unmarshal(data, &input); jsonErr != nil {
		// パース失敗でも通過
		return writePermissionDeniedApprove(out, "PermissionDenied logged")
	}

	// tool / denied_reason の解決（フォールバック）
	toolName := input.Tool
	if toolName == "" {
		toolName = input.ToolName
	}
	if toolName == "" {
		toolName = "unknown"
	}

	deniedReason := input.DeniedReason
	if deniedReason == "" {
		deniedReason = input.Reason
	}
	if deniedReason == "" {
		deniedReason = "unknown"
	}

	sessionID := input.SessionID
	if sessionID == "" {
		sessionID = "unknown"
	}
	agentID := input.AgentID
	if agentID == "" {
		agentID = "unknown"
	}
	agentType := input.AgentType
	if agentType == "" {
		agentType = "unknown"
	}

	// ステートディレクトリを確保
	stateDir := resolveNotificationStateDir()
	if mkErr := ensureNotificationStateDir(stateDir); mkErr != nil {
		// ディレクトリ作成失敗でも通過
		return writePermissionDeniedApprove(out, "PermissionDenied logged")
	}

	// JSONL に記録
	logFile := filepath.Join(stateDir, "permission-denied-events.jsonl")
	entry := permissionDeniedLogEntry{
		Event:     "permission_denied",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		SessionID: sessionID,
		AgentID:   agentID,
		AgentType: agentType,
		Tool:      toolName,
		Reason:    deniedReason,
	}
	if logErr := appendPermissionDeniedLog(logFile, entry); logErr != nil {
		_ = logErr
	}

	// stderr にデバッグ出力（bash スクリプトと同等）
	fmt.Fprintf(os.Stderr,
		"[PermissionDenied] agent=%s type=%s tool=%s reason=%s\n",
		agentID, agentType, toolName, deniedReason,
	)

	// Worker の場合: retry + systemMessage を返す
	if isWorkerAgentType(agentType) {
		notificationText := fmt.Sprintf(
			"[PermissionDenied] Worker のツール %s が auto mode で拒否されました。理由: %s。代替アプローチを検討するか、必要なら手動承認してください。",
			toolName, deniedReason,
		)

		resp := permissionDeniedRetryResponse{
			Retry:         true,
			SystemMessage: notificationText,
		}
		respData, marshalErr := json.Marshal(resp)
		if marshalErr != nil {
			return writePermissionDeniedApprove(out, "PermissionDenied logged")
		}
		_, writeErr := fmt.Fprintf(out, "%s\n", respData)
		return writeErr
	}

	// Worker 以外: approve を返す
	return writePermissionDeniedApprove(out, "PermissionDenied logged")
}

// isWorkerAgentType は agentType が Worker かどうかを判定する。
// bash の: [ "${AGENT_TYPE}" = "worker" ] || [ "${AGENT_TYPE}" = "task-worker" ] || echo "${AGENT_TYPE}" | grep -qE ':worker$'
// と同等。
func isWorkerAgentType(agentType string) bool {
	if agentType == "worker" || agentType == "task-worker" {
		return true
	}
	return strings.HasSuffix(agentType, ":worker")
}

// writePermissionDeniedApprove は approve レスポンスを書き込む。
func writePermissionDeniedApprove(out io.Writer, reason string) error {
	resp := permissionDeniedApproveResponse{
		Decision: "approve",
		Reason:   reason,
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("marshal approve response: %w", err)
	}
	_, err = fmt.Fprintf(out, "%s\n", data)
	return err
}

// appendPermissionDeniedLog は JSONL ファイルに1エントリ追記し、ローテーションする。
func appendPermissionDeniedLog(logFile string, entry permissionDeniedLogEntry) error {
	// シンボリックリンクチェック（notification_handler.go の isSymlinkNotification を使用）
	if isSymlinkNotification(logFile) {
		return fmt.Errorf("symlinked log file refused: %s", logFile)
	}

	entryJSON, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal log entry: %w", err)
	}

	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}
	defer f.Close()

	if _, writeErr := fmt.Fprintf(f, "%s\n", entryJSON); writeErr != nil {
		return fmt.Errorf("write log entry: %w", writeErr)
	}

	// ローテーション: 500行超なら400行に切り詰め（notification_handler.go の rotateJSONLNotification を使用）
	return rotateJSONLNotification(logFile, 500, 400)
}
