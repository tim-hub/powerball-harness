package hookhandler

// notification_handler.go
// notification-handler.sh の Go 移植。
//
// Notification イベント (permission_prompt, idle_prompt, auth_success 等) を
// .claude/state/notification-events.jsonl に記録する。
// 通知ハンドラはブロックしない（常に approve）。

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// notificationInput は Notification フックの stdin JSON。
type notificationInput struct {
	NotificationType string `json:"notification_type"`
	Type             string `json:"type"`
	Matcher          string `json:"matcher"`
	SessionID        string `json:"session_id"`
	AgentType        string `json:"agent_type"`
}

// notificationLogEntry は notification-events.jsonl の1エントリ。
type notificationLogEntry struct {
	Event            string `json:"event"`
	NotificationType string `json:"notification_type"`
	SessionID        string `json:"session_id"`
	AgentType        string `json:"agent_type"`
	Timestamp        string `json:"timestamp"`
}

// HandleNotification は notification-handler.sh の Go 移植。
//
// Notification フックで呼び出され、通知イベントを
// .claude/state/notification-events.jsonl に記録する。
// 通知ハンドラは常に approve を返す（ブロックしない）。
func HandleNotification(in io.Reader, out io.Writer) error {
	data, err := io.ReadAll(in)
	if err != nil || len(strings.TrimSpace(string(data))) == 0 {
		// 入力なし: 正常終了（exit 0 相当）
		return nil
	}

	var input notificationInput
	if jsonErr := json.Unmarshal(data, &input); jsonErr != nil {
		// パース失敗でも通過（通知ハンドラはブロックしない）
		return nil
	}

	// notification_type の解決（type / matcher でフォールバック）
	notificationType := input.NotificationType
	if notificationType == "" {
		notificationType = input.Type
	}
	if notificationType == "" {
		notificationType = input.Matcher
	}

	// ステートディレクトリを確保
	stateDir := resolveNotificationStateDir()
	if mkErr := ensureNotificationStateDir(stateDir); mkErr != nil {
		// ディレクトリ作成失敗でも通過
		return nil
	}

	// JSONL に記録
	logFile := filepath.Join(stateDir, "notification-events.jsonl")
	entry := notificationLogEntry{
		Event:            "notification",
		NotificationType: notificationType,
		SessionID:        input.SessionID,
		AgentType:        input.AgentType,
		Timestamp:        time.Now().UTC().Format(time.RFC3339),
	}
	if logErr := appendNotificationLog(logFile, entry); logErr != nil {
		// ログ書き込み失敗は無視
		_ = logErr
	}

	// Breezing バックグラウンド Worker に関する重要通知を stderr に出力（デバッグ用）
	if notificationType == "permission_prompt" && input.AgentType != "" {
		fmt.Fprintf(os.Stderr, "Notification: permission_prompt for agent_type=%s\n", input.AgentType)
	}
	if notificationType == "elicitation_dialog" && input.AgentType != "" {
		fmt.Fprintf(os.Stderr,
			"Notification: elicitation_dialog for agent_type=%s (auto-skipped in background)\n",
			input.AgentType)
	}

	// 通知ハンドラは常に正常終了（approve）
	return nil
}

// resolveNotificationStateDir は環境変数を考慮してステートディレクトリを返す。
// CLAUDE_PLUGIN_DATA が設定されている場合はプロジェクトスコープに切り替える。
func resolveNotificationStateDir() string {
	pluginData := os.Getenv("CLAUDE_PLUGIN_DATA")
	if pluginData != "" {
		projectRoot := os.Getenv("PROJECT_ROOT")
		if projectRoot == "" {
			cwd, err := os.Getwd()
			if err == nil {
				projectRoot = cwd
			}
		}
		hash := shortHashNotification(projectRoot)
		return filepath.Join(pluginData, "projects", hash)
	}

	projectRoot := os.Getenv("PROJECT_ROOT")
	if projectRoot == "" {
		cwd, err := os.Getwd()
		if err == nil {
			projectRoot = cwd
		}
	}
	return filepath.Join(projectRoot, ".claude", "state")
}

// ensureNotificationStateDir はディレクトリを作成し、シンボリックリンクを拒否する。
func ensureNotificationStateDir(stateDir string) error {
	parent := filepath.Dir(stateDir)

	// シンボリックリンクチェック（セキュリティ）
	if isSymlinkNotification(parent) || isSymlinkNotification(stateDir) {
		return fmt.Errorf("symlinked state path refused: %s", stateDir)
	}

	if mkErr := os.MkdirAll(stateDir, 0o700); mkErr != nil {
		return fmt.Errorf("mkdir state dir: %w", mkErr)
	}

	// 作成後も検証
	info, statErr := os.Lstat(stateDir)
	if statErr != nil {
		return fmt.Errorf("stat state dir: %w", statErr)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("state dir is symlink: %s", stateDir)
	}
	return nil
}

// appendNotificationLog は JSONL ファイルに1エントリ追記し、ローテーションする。
func appendNotificationLog(logFile string, entry notificationLogEntry) error {
	// シンボリックリンクチェック
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

	// ローテーション: 500行超なら400行に切り詰め
	return rotateJSONLNotification(logFile, 500, 400)
}

// rotateJSONLNotification は JSONL ファイルが maxLines を超えたら keepLines に切り詰める。
// notification_handler.go のローカル実装（パッケージ内の重複を避けるため _Notification サフィックスを付与）。
func rotateJSONLNotification(path string, maxLines, keepLines int) error {
	// シンボリックリンクチェック
	if isSymlinkNotification(path) || isSymlinkNotification(path+".tmp") {
		return fmt.Errorf("symlinked file refused for rotation")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil // ファイルが存在しない場合は無視
	}

	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) <= maxLines {
		return nil
	}

	// 末尾 keepLines 行を残す
	start := len(lines) - keepLines
	if start < 0 {
		start = 0
	}
	trimmed := strings.Join(lines[start:], "\n") + "\n"

	tmpPath := path + ".tmp"
	if writeErr := os.WriteFile(tmpPath, []byte(trimmed), 0o644); writeErr != nil {
		return fmt.Errorf("write tmp file: %w", writeErr)
	}
	return os.Rename(tmpPath, path)
}

// isSymlinkNotification はパスがシンボリックリンクかどうかを返す（存在しない場合は false）。
func isSymlinkNotification(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeSymlink != 0
}

// shortHashNotification はプロジェクトルートパスの短縮ハッシュ（12文字）を返す。
// bash の shasum -a 256 | cut -c1-12 と同等。
func shortHashNotification(input string) string {
	if input == "" {
		return "default"
	}
	// 簡易ハッシュ: FNV-1a ベース（外部依存なし）
	var h uint64 = 14695981039346656037
	for i := 0; i < len(input); i++ {
		h ^= uint64(input[i])
		h *= 1099511628211
	}
	return fmt.Sprintf("%012x", h&0xffffffffffff)
}
