package event

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// NotificationHandler は Notification フックハンドラ。
// Claude Code が通知を発行する際に発火し、イベントを JSONL ログに記録する。
//
// shell 版: scripts/hook-handlers/notification-handler.sh
type NotificationHandler struct {
	// StateDir はログファイルの保存先。
	// 空の場合は ResolveStateDir(projectRoot) を使う。
	StateDir string
}

// notificationInput は Notification フックの stdin JSON。
type notificationInput struct {
	NotificationType string `json:"notification_type,omitempty"`
	Type             string `json:"type,omitempty"`
	Matcher          string `json:"matcher,omitempty"`
	SessionID        string `json:"session_id,omitempty"`
	AgentType        string `json:"agent_type,omitempty"`
	CWD              string `json:"cwd,omitempty"`
}

// notificationLogEntry は notification-events.jsonl に書き出すエントリ。
type notificationLogEntry struct {
	Event            string `json:"event"`
	NotificationType string `json:"notification_type"`
	SessionID        string `json:"session_id"`
	AgentType        string `json:"agent_type"`
	Timestamp        string `json:"timestamp"`
}

// Handle は stdin から Notification ペイロードを読み取り、
// ログに記録して stdout には何も返さない。
// エラーは無視して処理を継続する（Breezing のバックグラウンド動作に影響しないため）。
func (h *NotificationHandler) Handle(r io.Reader, w io.Writer) error {
	data, err := io.ReadAll(r)
	if err != nil || len(data) == 0 {
		return nil
	}

	var inp notificationInput
	if err := json.Unmarshal(data, &inp); err != nil {
		return nil
	}

	// notification_type を正規化（複数フィールドに同じ情報が入る場合がある）
	notificationType := inp.NotificationType
	if notificationType == "" {
		notificationType = inp.Type
	}
	if notificationType == "" {
		notificationType = inp.Matcher
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

	logFile := filepath.Join(stateDir, "notification-events.jsonl")
	if isSymlink(logFile) {
		return nil
	}

	entry := notificationLogEntry{
		Event:            "notification",
		NotificationType: notificationType,
		SessionID:        inp.SessionID,
		AgentType:        inp.AgentType,
		Timestamp:        Now(),
	}

	h.appendLog(logFile, entry)

	// Breezing のバックグラウンド Worker への重要通知検出
	// stdout への出力は行わない（ログへの記録のみ）
	h.handleBreezingNotifications(notificationType, inp.AgentType)

	// Notification フックは出力不要（approve は暗黙）
	_ = w
	return nil
}

// appendLog は通知ログに 1 エントリ追記する。
func (h *NotificationHandler) appendLog(path string, entry notificationLogEntry) {
	data, err := json.Marshal(entry)
	if err != nil {
		return
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = fmt.Fprintf(f, "%s\n", data)

	RotateJSONL(path)
}

// handleBreezingNotifications は Breezing バックグラウンド Worker への
// 重要通知を stderr に出力する（デバッグ・観測性のため）。
func (h *NotificationHandler) handleBreezingNotifications(notificationType, agentType string) {
	if agentType == "" {
		return
	}

	switch notificationType {
	case "permission_prompt":
		// Worker が権限ダイアログに応答できない
		fmt.Fprintf(os.Stderr,
			"Notification: permission_prompt for agent_type=%s\n", agentType)
	case "elicitation_dialog":
		// MCP サーバーからの入力要求（v2.1.76+）
		// バックグラウンド Worker では Elicitation フォームに応答不能
		fmt.Fprintf(os.Stderr,
			"Notification: elicitation_dialog for agent_type=%s (auto-skipped in background)\n",
			agentType)
	}
}
